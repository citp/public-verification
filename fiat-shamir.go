package main

import (
	"crypto/sha512"
	"math/big"
)

type NMFS_Prover_Rounds struct {
	r         *big.Int
	s         *big.Int
	t         *big.Int
	rnd_r     *big.Int
	rnd_s     *big.Int
	rnd_t     *big.Int
	rnd_alpha *big.Int
}

type NMFS_Prover struct {
	alpha  *big.Int
	x      string
	Hx     DHElement
	H1     DHElement
	H2     DHElement
	Rounds []NMFS_Prover_Rounds
}

type NMFS_MSG_P_Rounds struct {
	R         DHElement
	S         DHElement
	T         DHElement
	com_alpha PC_Commitment
	com_r     PC_Commitment
	com_s     PC_Commitment
	com_t     PC_Commitment
	Response  interface{}
}

type NMFS_MSG_P struct {
	L      DHElement
	P1     DHElement
	P2     DHElement
	Rounds []NMFS_MSG_P_Rounds
}

// -----------------------------------------------------------------------------

func (p *NMFS_Prover) Init(ctx *NM_Context, x string, H1, H2 DHElement) {
	p.alpha = ctx.pc.PC_RandomScalar()
	p.H1 = H1
	p.H2 = H2
	p.x = x
	p.Hx = ctx.pc.ctxDH.HashToCurve(p.x)
	p.Rounds = make([]NMFS_Prover_Rounds, ctx.nRepeat)
}

func (p *NMFS_Prover) FiatShamir(ctx *NM_Context) NMFS_MSG_P {
	// defer timer(time.Now(), "FiatShamir - Prove")

	var msg NMFS_MSG_P
	msg.L = ctx.pc.ctxDH.EC_BaseMultiply(p.alpha)
	msg.P1 = ctx.pc.ctxDH.EC_Multiply(p.alpha, p.H1)
	msg.P2 = ctx.pc.ctxDH.EC_Multiply(p.alpha, p.H2)
	msg.Rounds = make([]NMFS_MSG_P_Rounds, len(p.Rounds))

	payload := ctx.pc.ctxDH.G.String()

	for i := 0; i < len(p.Rounds); i++ {
		p.Rounds[i].r = ctx.pc.PC_RandomScalar()
		p.Rounds[i].s = ctx.pc.PC_RandomScalar()
		p.Rounds[i].t = ctx.pc.PC_RandomScalar()

		p.Rounds[i].rnd_alpha, msg.Rounds[i].com_alpha = ctx.pc.PC_Commit(p.alpha)
		p.Rounds[i].rnd_r, msg.Rounds[i].com_r = ctx.pc.PC_Commit(p.Rounds[i].r)
		p.Rounds[i].rnd_s, msg.Rounds[i].com_s = ctx.pc.PC_Commit(p.Rounds[i].s)
		p.Rounds[i].rnd_t, msg.Rounds[i].com_t = ctx.pc.PC_Commit(p.Rounds[i].t)

		msg.Rounds[i].R = ctx.pc.ctxDH.EC_BaseMultiply(p.Rounds[i].r)
		msg.Rounds[i].S = ctx.pc.ctxDH.EC_Multiply(p.Rounds[i].s, p.Hx)
		msg.Rounds[i].T = ctx.pc.ctxDH.EC_Multiply(p.Rounds[i].t, p.Hx)

		payload += (msg.Rounds[i].R.String() + msg.Rounds[i].S.String() + msg.Rounds[i].T.String())
	}

	// Compute Fiat-Shamir Challenge
	payload += (p.Hx.String() + msg.L.String() + msg.P1.String() + msg.P2.String())
	c := FSHash(payload, len(p.Rounds))

	// Compute response to each challenge
	for i := 0; i < len(p.Rounds); i++ {
		aPrime := new(big.Int)
		rnd_aPrime := new(big.Int)

		switch c[i] {
		case 0:
			msg.Rounds[i].Response = NM_MSG_2P_0{p.Rounds[i].r, p.Rounds[i].s, p.Rounds[i].t, p.Rounds[i].rnd_r, p.Rounds[i].rnd_s, p.Rounds[i].rnd_t}
			break
		case 1:
			aPrime.Add(p.alpha, p.Rounds[i].r)
			rnd_aPrime.Add(p.Rounds[i].rnd_alpha, p.Rounds[i].rnd_r)
			break
		case 2:
			aPrime.Add(p.alpha, p.Rounds[i].s)
			rnd_aPrime.Add(p.Rounds[i].rnd_alpha, p.Rounds[i].rnd_s)
			break
		case 3:
			aPrime.Add(p.alpha, p.Rounds[i].t)
			rnd_aPrime.Add(p.Rounds[i].rnd_alpha, p.Rounds[i].rnd_t)
			break
		}

		if c[i] != 0 {
			msg.Rounds[i].Response = NM_MSG_2P_Not0{aPrime, rnd_aPrime}
		}
	}

	return msg
}

/*
Parses a byte array into 2-bit challenges (0, 1, 2, 3)
*/
func ParseIntoChallenge(buf []byte) []uint8 {
	b := big.NewInt(0).SetBytes(buf)
	ret := make([]uint8, b.BitLen()/2)

	for i := 0; i < len(ret); i++ {
		ret[i] = uint8(b.Bit(2*i) + (b.Bit(2*i+1) * 2))
	}

	return ret
}

/* Computes Fiat Shamir challenges */
func FSHash(payload string, rounds int) []uint8 {
	o := sha512.Sum512([]byte(payload))
	o1 := sha512.Sum512(o[:32])
	o2 := sha512.Sum512(o[32:])

	// Parse into round challenges
	c := append(ParseIntoChallenge(o1[:]), ParseIntoChallenge(o2[:])...)
	Assert(rounds < len(c))
	return c[:rounds]
}

// -----------------------------------------------------------------------------

func (v *NM_Verifier) FiatShamir(ctx *NM_Context, msgP NMFS_MSG_P) bool {
	// defer timer(time.Now(), "FiatShamir - Verify")
	H_x := ctx.pc.ctxDH.HashToCurve(v.x)
	aPrime := new(big.Int)
	rnd_aPrime := new(big.Int)

	payload := ctx.pc.ctxDH.G.String()

	for i := 0; i < len(msgP.Rounds); i++ {
		payload += (msgP.Rounds[i].R.String() + msgP.Rounds[i].S.String() + msgP.Rounds[i].T.String())
	}

	// Compute Fiat-Shamir Challenge
	payload += (H_x.String() + msgP.L.String() + msgP.P1.String() + msgP.P2.String())
	c := FSHash(payload, ctx.nRepeat)

	for i := 0; i < len(c); i++ {
		switch c[i] {
		case 0:
			msg := msgP.Rounds[i].Response.(NM_MSG_2P_0)
			Assert(ctx.pc.PC_Decommit(msg.r, msg.rnd_r, msgP.Rounds[i].com_r))
			Assert(ctx.pc.PC_Decommit(msg.s, msg.rnd_s, msgP.Rounds[i].com_s))
			Assert(ctx.pc.PC_Decommit(msg.t, msg.rnd_t, msgP.Rounds[i].com_t))

			lhs_R := ctx.pc.ctxDH.EC_BaseMultiply(msg.r)
			lhs_S := ctx.pc.ctxDH.EC_Multiply(msg.s, H_x)
			lhs_T := ctx.pc.ctxDH.EC_Multiply(msg.t, H_x)
			Assert(msgP.Rounds[i].R.x.Cmp(lhs_R.x) == 0 && msgP.Rounds[i].R.y.Cmp(lhs_R.y) == 0)
			Assert(msgP.Rounds[i].S.x.Cmp(lhs_S.x) == 0 && msgP.Rounds[i].S.y.Cmp(lhs_S.y) == 0)
			Assert(msgP.Rounds[i].T.x.Cmp(lhs_T.x) == 0 && msgP.Rounds[i].T.y.Cmp(lhs_T.y) == 0)
			break
		default:
			msg := msgP.Rounds[i].Response.(NM_MSG_2P_Not0)
			aPrime.Set(msg.aPrime)
			rnd_aPrime.Set(msg.rnd_aPrime)
		}
		switch c[i] {
		case 1:
			com_ar := ctx.pc.PC_Add(msgP.Rounds[i].com_alpha, msgP.Rounds[i].com_r)
			Assert(ctx.pc.PC_Decommit(aPrime, rnd_aPrime, com_ar))
			lhs := ctx.pc.ctxDH.EC_BaseMultiply(aPrime)
			rhs := ctx.pc.ctxDH.EC_Add(msgP.L, msgP.Rounds[i].R)
			Assert(lhs.x.Cmp(rhs.x) == 0 && lhs.y.Cmp(rhs.y) == 0)
			break
		case 2:
			com_as := ctx.pc.PC_Add(msgP.Rounds[i].com_alpha, msgP.Rounds[i].com_s)
			Assert(ctx.pc.PC_Decommit(aPrime, rnd_aPrime, com_as))
			lhs := ctx.pc.ctxDH.EC_Multiply(aPrime, H_x)
			rhs := ctx.pc.ctxDH.EC_Add(msgP.P1, msgP.Rounds[i].S)
			Assert(lhs.x.Cmp(rhs.x) != 0 || lhs.y.Cmp(rhs.y) != 0)
			break
		case 3:
			com_at := ctx.pc.PC_Add(msgP.Rounds[i].com_alpha, msgP.Rounds[i].com_t)
			Assert(ctx.pc.PC_Decommit(aPrime, rnd_aPrime, com_at))
			lhs := ctx.pc.ctxDH.EC_Multiply(aPrime, H_x)
			rhs := ctx.pc.ctxDH.EC_Add(msgP.P2, msgP.Rounds[i].T)
			Assert(lhs.x.Cmp(rhs.x) != 0 || lhs.y.Cmp(rhs.y) != 0)
			break
		}
	}

	return true
}
