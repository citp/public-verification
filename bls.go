package main

import (
	"math/big"
	"runtime"

	"github.com/cloudflare/bn256"
)

type BLS_Element1 *bn256.G1
type BLS_Element2 *bn256.G2
type BLS_SecretKey *big.Int
type BLS_PublicKey BLS_Element2
type BLS_Signature BLS_Element1

type BLS_Context struct {
	// tau int
	// N   int
	order *big.Int
	// g     BLSElement2
	// h     BLSElement2
}

type BLS_PublicKey_Share struct {
	x *big.Int
	y BLS_PublicKey
}

type BLS_Signature_Share struct {
	x *big.Int
	y BLS_Signature
}

// -----------------------------------------------------------------------------

func NewBLS_Context() BLS_Context {
	// _, g, err := bn256.RandomG2(rand.Reader)
	// Check(err)
	return BLS_Context{bn256.Order}
}

func (ctx *BLS_Context) BLS_SecKeygen() BLS_SecretKey {
	return RandomScalar(ctx.order)
}

// func (ctx *BLS_Context) BLS_AggSK(sk []BLS_SecretKey) BLS_SecretKey {
// 	AggregateFieldElems(sk, ctx.order)
// }

func (ctx *BLS_Context) BLS_PubKeygen(sk BLS_SecretKey) BLS_PublicKey {
	return new(bn256.G2).ScalarBaseMult(sk)
}

func (ctx *BLS_Context) BLS_Sign(sk BLS_SecretKey, msg []byte) BLS_Signature {
	h := bn256.HashG1(msg, []byte("BLS_Sign"))
	h.ScalarMult(h, sk)
	return h
}

func (ctx *BLS_Context) BLS_Verify(pk BLS_PublicKey, msg []byte, sign BLS_Signature) bool {
	lhs := bn256.Pair(sign, new(bn256.G2).ScalarBaseMult(big.NewInt(1)))
	rhs := bn256.Pair(bn256.HashG1(msg, []byte("BLS_Sign")), pk)
	// if lhs.String() != rhs.String() {
	// 	fmt.Println(lhs, rhs)
	// }
	return (lhs.String() == rhs.String())
}

func (ctx *BLS_Context) BLS_Verify_Worker(pk BLS_PublicKey, outChan chan bool, inChan chan BLS_Verify_Input) {
	for inp := range inChan {
		// msg []byte, sign BLS_Signature,
		outChan <- ctx.BLS_Verify(pk, inp.msg, inp.sign)
	}

}

type BLS_Verify_Input struct {
	msg  []byte
	sign BLS_Signature
}

func (ctx *BLS_Context) BLS_Verify_Parallel(pk BLS_PublicKey, msgs [][]byte, signs []BLS_Signature) bool {
	// defer timer(time.Now(), "BLS_Verify_Parallel")

	outChan := make(chan bool, len(signs))
	inChan := make(chan BLS_Verify_Input, len(signs))

	for i := 0; i < len(signs); i++ {
		inChan <- BLS_Verify_Input{msgs[i], signs[i]}
	}

	for i := 0; i < runtime.NumCPU(); i++ {
		go ctx.BLS_Verify_Worker(pk, outChan, inChan)
	}

	close(inChan)

	for i := 0; i < len(signs); i++ {
		out := <-outChan
		if !out {
			return false
		}
	}
	return true
}

func (ctx *BLS_Context) BLS_AggPK(ctxS *ShamirContext, shares []BLS_PublicKey_Share) BLS_PublicKey {
	// t1 := new(BLS_PublicKey)
	// t2 := new(BLS_PublicKey)
	secret := new(bn256.G2)
	// secret.Set(shares[0].y)

	if len(shares) < ctxS.tau {
		return *new(BLS_PublicKey)
	}

	for j := 0; j < ctxS.tau; j++ {
		// Integer prod(1);
		prod := big.NewInt(1)
		for m := 0; m < ctxS.tau; m++ {
			if m != j {
				// invDiff := big.NewInt(int64((m + 1) - (j + 1)))
				invDiff := new(big.Int).Sub(shares[m].x, shares[j].x)
				invDiff.Mod(invDiff, ctxS.mod)
				invDiff.ModInverse(invDiff, ctxS.mod)
				// Integer invDiff = InvMod(Integer((m + 1) - (j + 1)) % ctxS.q, ctxS.q);
				prod.Mul(prod, invDiff)
				// prod.Mul(prod, big.NewInt(int64(m+1)))
				prod.Mul(prod, shares[m].x)
				prod.Mod(prod, ctxS.mod)
				// prod = (prod * ((m + 1) * invDiff)) % ctxS.q
			}
		}
		t2 := (*shares[j].y)
		t2.ScalarMult(&t2, prod)
		secret.Add(secret, &t2)
		// fromZZ(ctxB, prod, t1)
		// element_pow_zn(t2, shares[j].data, t1)
		// element_mul(secret.data, secret.data, t2)
	}

	return secret
}

func (ctx *BLS_Context) BLS_AggSign(ctxS *ShamirContext, shares []BLS_Signature_Share) BLS_Signature {
	secret := new(bn256.G1)
	// secret.Set(shares[0].y)

	if len(shares) < ctxS.tau {
		return *new(BLS_Signature)
	}

	for j := 0; j < ctxS.tau; j++ {
		prod := big.NewInt(1)
		for m := 0; m < ctxS.tau; m++ {
			if m != j {
				// invDiff := big.NewInt(int64((m + 1) - (j + 1)))
				invDiff := new(big.Int).Sub(shares[m].x, shares[j].x)
				invDiff.Mod(invDiff, ctxS.mod)
				invDiff.ModInverse(invDiff, ctxS.mod)
				prod.Mul(prod, invDiff)
				// prod.Mul(prod, big.NewInt(int64(m+1)))
				prod.Mul(prod, shares[m].x)
				prod.Mod(prod, ctxS.mod)
			}
		}
		t2 := (*shares[j].y)
		t2.ScalarMult(&t2, prod)
		secret.Add(secret, &t2)
	}

	return secret
}

// -----------------------------------------------------------------------------

// func Init() {
// 	s := suites.MustFind("Ed25519")
// 	x := s.Scalar().Zero()
// 	fmt.Println(x)
// }

// func (ctx *BLS_Context) Sign(sh *share.PriShare, msg []byte) {
// 	tbls.Sign(pairing.NewSuiteBn256(), sh, msg)
// }

// func (ctx *BLS_Context) Verify(sh *share.PriShare, msg []byte) {
// 	tbls.Sign(pairing.NewSuiteBn256(), sh, msg)
// }
