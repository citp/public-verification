package main

import (
	"crypto/elliptic"
	"crypto/rand"
	"math/big"
	"strings"

	"golang.org/x/crypto/blake2b"
)

type DHContext struct {
	Curve elliptic.Curve
	G     DHElement
}

type DHScalar *big.Int
type DHElement struct {
	x *big.Int
	y *big.Int
}

// -----------------------------------------------------------------------------

func Blake2b(buf []byte) []byte {
	// hash, err := blake2b.New256(nil)
	// Check(err)
	ret := blake2b.Sum256(buf)
	// return hash.Sum(buf)
	return ret[:]
}

func RandomScalar(mod *big.Int) *big.Int {
	ret, err := rand.Int(rand.Reader, mod)
	Check(err)
	return ret
}

// -----------------------------------------------------------------------------

func NewDHContext() DHContext {
	var ret DHContext
	ret.Curve = elliptic.P256()
	ret.G = DHElement{ret.Curve.Params().Gx, ret.Curve.Params().Gy}
	return ret
}

func NewDHElement() DHElement {
	return DHElement{big.NewInt(0), big.NewInt(0)}
}

// -----------------------------------------------------------------------------

func (ctx *DHContext) EC_BaseMultiply(s DHScalar) DHElement {
	var ret DHElement
	ret.x, ret.y = ctx.Curve.ScalarBaseMult((*s).Bytes())
	return ret
}

func (ctx *DHContext) EC_Multiply(s DHScalar, p DHElement) DHElement {
	var ret DHElement
	ret.x, ret.y = ctx.Curve.ScalarMult(p.x, p.y, (*s).Bytes())
	return ret
}

func (ctx *DHContext) EC_Add(a, b DHElement) DHElement {
	var ret DHElement
	ret.x, ret.y = ctx.Curve.Add(a.x, a.y, b.x, b.y)
	return ret
}

func (ctx *DHContext) DH_Reduce(P, L, H DHElement) (DHScalar, DHScalar, DHElement, DHElement) {
	beta := RandomScalar(ctx.Curve.Params().P)
	gamma := RandomScalar(ctx.Curve.Params().P)
	Q := ctx.EC_Add(ctx.EC_Multiply(beta, H), ctx.EC_Multiply(gamma, ctx.G))
	S := ctx.EC_Add(ctx.EC_Multiply(beta, P), ctx.EC_Multiply(gamma, L))

	return beta, gamma, Q, S
}

type DH_Reduce_Worker_Input struct {
	idx     int
	P, L, H DHElement
}

type DH_Reduce_Worker_Output struct {
	idx  int
	Q, S DHElement
}

func (ctx *DHContext) DH_Reduce_Worker(inChan chan DH_Reduce_Worker_Input, outChan chan DH_Reduce_Worker_Output) {
	for inp := range inChan {
		// if inp.idx == 1 {
		// 	break
		// }
		_, _, Q, S := ctx.DH_Reduce(inp.P, inp.L, inp.H)
		outChan <- DH_Reduce_Worker_Output{inp.idx, Q, S}
		// fmt.Println(inp.idx)
	}
}

func (ctx *DHContext) DH_Reduce_Parallel(P []DHElement, L DHElement, H []DHElement, nRoutines int) ([]DHElement, []DHElement) {
	// defer timer(time.Now(), "DH_Reduce_Parallel")
	inChan := make(chan DH_Reduce_Worker_Input, len(P))
	outChan := make(chan DH_Reduce_Worker_Output, len(P))

	for i := 0; i < len(P); i++ {
		inChan <- DH_Reduce_Worker_Input{i, P[i], L, H[i]}
	}
	close(inChan)

	for i := 0; i < nRoutines; i++ {
		go ctx.DH_Reduce_Worker(inChan, outChan)
	}

	// inChan <- DH_Reduce_Worker_Input{-1, P[0], L, H[0]}

	Q := make([]DHElement, len(P))
	S := make([]DHElement, len(P))
	for i := 0; i < len(P); i++ {
		out := <-outChan
		Q[out.idx] = out.Q
		S[out.idx] = out.S
	}
	return Q, S
}

// -----------------------------------------------------------------------------

func (ctx *DHContext) LegendreSym(z *big.Int) *big.Int {
	p := ctx.Curve.Params().P
	exp := new(big.Int).Sub(p, big.NewInt(1))
	exp = exp.Div(exp, big.NewInt(2))
	return new(big.Int).Exp(z, exp, p)
}

func (ctx *DHContext) SquareRootModP(z *big.Int) *big.Int {
	p := ctx.Curve.Params().P
	exp := new(big.Int).Add(p, big.NewInt(1))
	exp = exp.Div(exp, big.NewInt(4))
	return new(big.Int).Exp(z, exp, p)
}

func (ctx *DHContext) YfromX(x *big.Int) *big.Int {
	p := ctx.Curve.Params().P
	three := big.NewInt(3)
	x3 := new(big.Int).Exp(x, three, p)
	y2 := new(big.Int).Sub(x3, new(big.Int).Mul(three, x))
	y2 = y2.Add(y2, ctx.Curve.Params().B)
	y2 = y2.Mod(y2, p)
	return ctx.SquareRootModP(y2)
}

func (ctx *DHContext) HashToCurve(s string) DHElement {
	// fmt.Println("modulo 4", new(big.Int).Mod(ctx.Curve.Params().P, big.NewInt(4)))
	buf := []byte(s)
	p := ctx.Curve.Params().P
	for {
		bufHash := Blake2b(buf)
		x := new(big.Int).SetBytes(bufHash)
		x = x.Mod(x, p)
		y := ctx.YfromX(x)
		if ctx.Curve.IsOnCurve(x, y) {
			return DHElement{x, y}
		}
		buf = Blake2b(bufHash)
	}
}

// -----------------------------------------------------------------------------

func (p *DHElement) String() string {
	return p.x.Text(16) + "," + p.y.Text(16)
}

func BigIntFrom(s string) *big.Int {
	ret := big.NewInt(0)
	ret, ok := ret.SetString(s, 16)
	if !ok {
		panic("Could not deserialize")
	}

	return ret
}

func DHElementFrom(s string) DHElement {
	strs := strings.Split(s, ",")
	return DHElement{BigIntFrom(strs[0]), BigIntFrom(strs[1])}
}
