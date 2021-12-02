package main

import (
	"math/big"
)

type PC_Context struct {
	ctxDH DHContext
	H     DHElement
}

type PC_Commitment struct {
	C DHElement
}

func NewPC_Context() PC_Context {
	var ctx PC_Context
	ctx.ctxDH = NewDHContext()
	ctx.H = ctx.ctxDH.HashToCurve("This is a generator...")
	return ctx
}

func (ctx *PC_Context) PC_RandomScalar() *big.Int {
	return RandomScalar(ctx.ctxDH.Curve.Params().P)
}

func (ctx *PC_Context) PC_ComputeLC(x *big.Int, r *big.Int) DHElement {
	return ctx.ctxDH.EC_Add(ctx.ctxDH.EC_BaseMultiply(x), ctx.ctxDH.EC_Multiply(r, ctx.H))
}

func (ctx *PC_Context) PC_Commit(x *big.Int) (*big.Int, PC_Commitment) {
	r := RandomScalar(ctx.ctxDH.Curve.Params().P)
	return r, PC_Commitment{ctx.PC_ComputeLC(x, r)}
}

func (ctx *PC_Context) PC_Decommit(x *big.Int, r *big.Int, com PC_Commitment) bool {
	C := ctx.PC_ComputeLC(x, r)
	// fmt.Println(C.x, C.y)
	// fmt.Println(com.C.x, com.C.y)
	return (C.x.Cmp(com.C.x) == 0) && (C.y.Cmp(com.C.y) == 0)
}

func (ctx *PC_Context) PC_Add(com1, com2 PC_Commitment) PC_Commitment {
	return PC_Commitment{ctx.ctxDH.EC_Add(com1.C, com2.C)}
}

func (ctx *PC_Context) PC_AddInts(x, y *big.Int) *big.Int {

	ret := new(big.Int).Add(x, y)
	// ret = ret.Mod(ret, ctx.ctxDH.Curve.Params().P)
	// fmt.Println(x, y, ret)
	return ret
}
