package main

import (
	"math/big"
)

type NM_Context struct {
	nRepeat int
	mode    int
	pc      PC_Context
}

type NM_Prover struct {
	alpha     *big.Int
	x         string
	Hx        DHElement
	H1        DHElement
	H2        DHElement
	r         *big.Int
	s         *big.Int
	t         *big.Int
	rnd_r     *big.Int
	rnd_s     *big.Int
	rnd_t     *big.Int
	rnd_alpha *big.Int
}

type NM_Verifier struct {
	x  string
	Hx DHElement
	H1 DHElement
	H2 DHElement
}

type NM_MSG_1P struct {
	L         DHElement
	P1        DHElement
	P2        DHElement
	R         DHElement
	S         DHElement
	T         DHElement
	com_alpha PC_Commitment
	com_r     PC_Commitment
	com_s     PC_Commitment
	com_t     PC_Commitment
}

type NM_MSG_1V struct {
	coin int64
}

type NM_MSG_2P_0 struct {
	r     *big.Int
	s     *big.Int
	t     *big.Int
	rnd_r *big.Int
	rnd_s *big.Int
	rnd_t *big.Int
}

type NM_MSG_2P_Not0 struct {
	aPrime     *big.Int
	rnd_aPrime *big.Int
}

// -----------------------------------------------------------------------------

func NewNM_Context(nRepeat, mode int) NM_Context {
	return NM_Context{nRepeat, mode, NewPC_Context()}
}

func (p *NM_Prover) Init(ctx *NM_Context, x string, H1, H2 DHElement) {
	p.alpha = ctx.pc.PC_RandomScalar()
	p.H1 = H1
	p.H2 = H2
	p.x = x
	p.Hx = ctx.pc.ctxDH.HashToCurve(p.x)
}

func (p *NM_Prover) One(ctx *NM_Context) NM_MSG_1P {
	// defer timer(time.Now(), "PR1")
	p.r = ctx.pc.PC_RandomScalar()
	p.s = ctx.pc.PC_RandomScalar()
	p.t = ctx.pc.PC_RandomScalar()

	var com_alpha, com_r, com_s, com_t PC_Commitment

	p.rnd_alpha, com_alpha = ctx.pc.PC_Commit(p.alpha)
	p.rnd_r, com_r = ctx.pc.PC_Commit(p.r)
	p.rnd_s, com_s = ctx.pc.PC_Commit(p.s)
	p.rnd_t, com_t = ctx.pc.PC_Commit(p.t)

	R := ctx.pc.ctxDH.EC_BaseMultiply(p.r)
	S := ctx.pc.ctxDH.EC_Multiply(p.s, p.Hx)
	T := ctx.pc.ctxDH.EC_Multiply(p.t, p.Hx)

	L := ctx.pc.ctxDH.EC_BaseMultiply(p.alpha)
	P1 := ctx.pc.ctxDH.EC_Multiply(p.alpha, p.H1)
	P2 := ctx.pc.ctxDH.EC_Multiply(p.alpha, p.H2)

	return NM_MSG_1P{L, P1, P2, R, S, T, com_alpha, com_r, com_s, com_t}
}

func (p *NM_Prover) Two(ctx *NM_Context, msg NM_MSG_1V) (*NM_MSG_2P_0, *NM_MSG_2P_Not0) {
	// defer timer(time.Now(), "PR2")
	aPrime := new(big.Int)
	rnd_aPrime := new(big.Int)

	switch msg.coin {
	case 0:
		return &NM_MSG_2P_0{p.r, p.s, p.t, p.rnd_r, p.rnd_s, p.rnd_t}, nil
	case 1:
		aPrime.Add(p.alpha, p.r)
		rnd_aPrime.Add(p.rnd_alpha, p.rnd_r)
	case 2:
		aPrime.Add(p.alpha, p.s)
		rnd_aPrime.Add(p.rnd_alpha, p.rnd_s)
	case 3:
		aPrime.Add(p.alpha, p.t)
		rnd_aPrime.Add(p.rnd_alpha, p.rnd_t)
	}

	return nil, &NM_MSG_2P_Not0{aPrime, rnd_aPrime}
}

// -----------------------------------------------------------------------------

func (v *NM_Verifier) Init(ctx *NM_Context, x string, H1, H2 DHElement) {
	v.H1 = H1
	v.H2 = H2
	v.x = x
	v.Hx = ctx.pc.ctxDH.HashToCurve(v.x)
}

func (v *NM_Verifier) One(ctx *NM_Context) NM_MSG_1V {
	// defer timer(time.Now(), "VR1")
	coin := RandomScalar(big.NewInt(int64(2))).Int64()
	if coin == 1 {
		coin += RandomScalar(big.NewInt(int64(3))).Int64()
	}

	return NM_MSG_1V{coin}
}

func (v *NM_Verifier) Two(ctx *NM_Context, msgV NM_MSG_1V, msg1P NM_MSG_1P, msg2P interface{}) bool {
	// defer timer(time.Now(), "VR2")
	H_x := ctx.pc.ctxDH.HashToCurve(v.x)
	aPrime := new(big.Int)
	rnd_aPrime := new(big.Int)

	switch msgV.coin {
	case 0:
		msg := msg2P.(NM_MSG_2P_0)
		Assert(ctx.pc.PC_Decommit(msg.r, msg.rnd_r, msg1P.com_r))
		Assert(ctx.pc.PC_Decommit(msg.s, msg.rnd_s, msg1P.com_s))
		Assert(ctx.pc.PC_Decommit(msg.t, msg.rnd_t, msg1P.com_t))

		lhs_R := ctx.pc.ctxDH.EC_BaseMultiply(msg.r)
		lhs_S := ctx.pc.ctxDH.EC_Multiply(msg.s, H_x)
		lhs_T := ctx.pc.ctxDH.EC_Multiply(msg.t, H_x)
		Assert(msg1P.R.x.Cmp(lhs_R.x) == 0 && msg1P.R.y.Cmp(lhs_R.y) == 0)
		Assert(msg1P.S.x.Cmp(lhs_S.x) == 0 && msg1P.S.y.Cmp(lhs_S.y) == 0)
		Assert(msg1P.T.x.Cmp(lhs_T.x) == 0 && msg1P.T.y.Cmp(lhs_T.y) == 0)
		return true
	default:
		msg := msg2P.(NM_MSG_2P_Not0)
		aPrime.Set(msg.aPrime)
		rnd_aPrime.Set(msg.rnd_aPrime)
	}

	switch msgV.coin {
	case 1:
		com_ar := ctx.pc.PC_Add(msg1P.com_alpha, msg1P.com_r)
		Assert(ctx.pc.PC_Decommit(aPrime, rnd_aPrime, com_ar))
		lhs := ctx.pc.ctxDH.EC_BaseMultiply(aPrime)
		rhs := ctx.pc.ctxDH.EC_Add(msg1P.L, msg1P.R)
		Assert(lhs.x.Cmp(rhs.x) == 0 && lhs.y.Cmp(rhs.y) == 0)
	case 2:
		com_as := ctx.pc.PC_Add(msg1P.com_alpha, msg1P.com_s)
		Assert(ctx.pc.PC_Decommit(aPrime, rnd_aPrime, com_as))
		lhs := ctx.pc.ctxDH.EC_Multiply(aPrime, H_x)
		rhs := ctx.pc.ctxDH.EC_Add(msg1P.P1, msg1P.S)
		Assert(lhs.x.Cmp(rhs.x) != 0 || lhs.y.Cmp(rhs.y) != 0)
	case 3:
		com_at := ctx.pc.PC_Add(msg1P.com_alpha, msg1P.com_t)
		Assert(ctx.pc.PC_Decommit(aPrime, rnd_aPrime, com_at))
		lhs := ctx.pc.ctxDH.EC_Multiply(aPrime, H_x)
		rhs := ctx.pc.ctxDH.EC_Add(msg1P.P2, msg1P.T)
		Assert(lhs.x.Cmp(rhs.x) != 0 || lhs.y.Cmp(rhs.y) != 0)
	}

	return true
}
