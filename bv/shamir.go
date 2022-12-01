package bv

import (
	"math/big"
	"strings"
)

type ShamirContext struct {
	tau int
	N   int
	mod *big.Int
}

type ShamirShare struct {
	x *big.Int
	y *big.Int
}

type ShamirPoly struct {
	coeffs []*big.Int
}

// -----------------------------------------------------------------------------

func NewShamirContext(tau, N int, mod *big.Int) ShamirContext {
	return ShamirContext{tau, N, mod}
}

func (ctx *ShamirContext) NewShamirPoly(secret *big.Int) ShamirPoly {
	var poly ShamirPoly
	poly.coeffs = make([]*big.Int, ctx.tau)
	poly.coeffs[0] = secret
	for i := 1; i < ctx.tau; i++ {
		poly.coeffs[i] = RandomScalar(ctx.mod)
	}
	return poly
}

func (ctx *ShamirContext) evaluate(poly *ShamirPoly, x *big.Int) *big.Int {
	ret := big.NewInt(0)
	xPow := big.NewInt(1)

	for i := 0; i < len(poly.coeffs); i++ {
		term := new(big.Int).Mul(poly.coeffs[i], xPow)
		ret.Add(ret, term)
		ret.Mod(ret, ctx.mod)
		xPow.Mul(xPow, x)
		xPow.Mod(xPow, ctx.mod)
	}
	return ret

}

func (ctx *ShamirContext) NewSharing(secret *big.Int) (ShamirPoly, []ShamirShare) {
	shares := make([]ShamirShare, ctx.N)
	poly := ctx.NewShamirPoly(secret)

	for i := 0; i < ctx.N; i++ {
		// shares[i].x = RandomBnd(ctx.q);
		shares[i].x = big.NewInt(int64(i + 1))
		shares[i].y = ctx.evaluate(&poly, shares[i].x)
		// shares[i].ctx = *ctx
	}

	return poly, shares
}

func (ctx *ShamirContext) AddShares(shares []ShamirShare) ShamirShare {
	ret := shares[0]
	for i := 1; i < len(shares); i++ {
		Assert(shares[i].x.Cmp(ret.x) == 0)
		// Assert(shares[i].x == ret.x)
		// Assert(shares[i].ctx.N == ret.ctx.N)
		// Assert(shares[i].ctx.tau == ret.ctx.tau)

		ret.y.Add(ret.y, shares[i].y)
		ret.y.Mod(ret.y, ctx.mod)
	}
	return ret
}

// func (ctx *ShamirContext)

// -----------------------------------------------------------------------------

func (share *ShamirShare) String() string {
	return share.x.Text(16) + "," + share.y.Text(16)
}

func ShamirShareFrom(s string) ShamirShare {
	strs := strings.Split(s, ",")
	// fmt.Println("strs", strs)
	var share ShamirShare
	share.x, _ = big.NewInt(0).SetString(strs[0], 16)
	share.y, _ = big.NewInt(0).SetString(strs[1], 16)

	return share
}
