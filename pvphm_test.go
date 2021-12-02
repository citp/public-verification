package main

import (
	"crypto/rand"
	"fmt"
	"strconv"
	"testing"
)

// #############################################################################

var N int = 3
var tau int = 2

// var N = *flag.Int("groups", 3, "Blocklist Verification: N")
// var tau = *flag.Int("tau", 2, "Blocklist Verification: tau")

// #############################################################################

func TestQuickVerifier(bench *testing.T) {
	ctx := NewBLS_Context()
	sk := ctx.BLS_SecKeygen()
	pk := ctx.BLS_PubKeygen(sk)

	size := 1 << 20

	msg := make([]byte, size*33)
	rand.Read(msg)

	sig := ctx.BLS_Sign(sk, msg)

	// bench.ResetTimer()
	Assert(ctx.BLS_Verify(pk, msg, sig))
}

func BenchmarkBV(bench *testing.B) {
	// fmt.Println(tau, N)
	ctx := NewBVContext(N, tau)

	var G []BVGroup
	var S BVServer
	var aggPk BLS_PublicKey
	var pdataStrs [][]byte
	var signs []BLS_Signature

	bench.Run("Init", func(b *testing.B) {
		G = HelperBVGroupInit(b, &ctx)
	})

	bench.Run("Init/Server", func(b *testing.B) {
		S = NewBVServer(&ctx)
	})

	m1G, m1S := HelperBVGroupRound1(bench, &ctx, &G)

	var m2S BVMSG_2_S
	bench.Run("2/Server", func(b *testing.B) {
		m2S = S.Two(m1S)
	})

	m3S := HelperBVGroupRound2(bench, &ctx, &G, m1G, m2S)

	bench.Run("3/Server", func(b *testing.B) {
		aggPk, pdataStrs, signs = S.Three(m3S)
	})

	fmt.Println("Signs:", len(signs))
	bench.Run("Verifier", func(b *testing.B) {
		Assert(ctx.bls.BLS_Verify_Parallel(aggPk, pdataStrs, signs))
	})
}

// #############################################################################

func HelperBVGroupInit(bench *testing.B, ctx *BVContext) []BVGroup {
	G := make([]BVGroup, ctx.N)
	for i := 0; i < ctx.N; i++ {
		bench.Run("Group", func(b *testing.B) {
			fpath := "data/" + strconv.Itoa(i+1) + ".dat"
			G[i] = NewBVGroup(ctx, i+1, ReadFile(fpath))
		})
	}
	return G
}

func HelperBVGroupRound1(bench *testing.B, ctx *BVContext, G *[]BVGroup) ([][]BVMSG_1_GG, []BVMSG_1_GS) {
	mG := make([][]BVMSG_1_GG, ctx.N)
	for i := 0; i < ctx.N; i++ {
		mG[i] = make([]BVMSG_1_GG, ctx.N)
	}
	mS := make([]BVMSG_1_GS, ctx.N)

	for i := 0; i < ctx.N; i++ {
		var mg []BVMSG_1_GG
		var ms BVMSG_1_GS
		bench.Run("1/Group", func(b *testing.B) {
			mg, ms = (*G)[i].One()
		})
		for j := 0; j < ctx.N; j++ {
			mG[j][i] = mg[j]
		}
		mS[i] = ms
	}

	return mG, mS
}

func HelperBVGroupRound2(bench *testing.B, ctx *BVContext, G *[]BVGroup, mG [][]BVMSG_1_GG, mS BVMSG_2_S) []BVMSG_3_GS {
	m3G := make([][]BVMSG_3_GG, ctx.N)
	for i := 0; i < ctx.N; i++ {
		m3G[i] = make([]BVMSG_3_GG, ctx.N)
	}
	m2S := make([]BVMSG_3_GS, ctx.N)

	var X []string
	var pdata BVPData

	for i := 0; i < ctx.N; i++ {
		bench.Run("2/Group", func(b *testing.B) {
			X, pdata = (*G)[i].Two(mG[i], mS)
		})

		var mg []BVMSG_3_GG
		bench.Run("3A/Group", func(b *testing.B) {
			mg = (*G)[i].ThreeA()
			for j := 0; j < ctx.N; j++ {
				m3G[j][i] = mg[j]
			}
		})
	}

	for i := 0; i < ctx.N; i++ {
		bench.Run("3B/Group", func(b *testing.B) {
			m2S[i] = (*G)[i].ThreeB(m3G[i], X, pdata)
		})
	}

	// m3_gg_s := G.ThreeA()
	return m2S
}

// #############################################################################

func BenchmarkNM(bench *testing.B) {
	var prover NM_Prover
	var verifier NM_Verifier

	for i := 0; i < bench.N; i++ {
		ctx := NewNM_Context(384, 1)
		bench.Run("Init", func(b *testing.B) {
			HelperNMInit(b, &ctx, &prover, &verifier)
		})

		bench.Run("Rounds", func(b *testing.B) {
			HelperNMInner(b, &ctx, &prover, &verifier)
		})
	}
}

// #############################################################################

func HelperNMInit(bench *testing.B, ctx *NM_Context, p *NM_Prover, v *NM_Verifier) {
	x := string(RandomBytes(15))
	H1 := ctx.pc.ctxDH.HashToCurve(string(RandomBytes(15)))
	H2 := ctx.pc.ctxDH.HashToCurve(string(RandomBytes(15)))

	bench.Run("Prover", func(b *testing.B) {
		p.Init(ctx, x, H1, H2)
	})

	bench.Run("Verifier", func(b *testing.B) {
		v.Init(ctx, x, H1, H2)
	})
}

func HelperNMInner(bench *testing.B, ctx *NM_Context, p *NM_Prover, v *NM_Verifier) {

	var msg1V NM_MSG_1V
	var msg1P NM_MSG_1P
	var msg2P_0 *NM_MSG_2P_0
	var msg2P_Not0 *NM_MSG_2P_Not0

	bench.Run("1/Verifier", func(b *testing.B) {
		msg1V = v.One(ctx)
	})

	bench.Run("1/Prover", func(b *testing.B) {
		msg1P = p.One(ctx)
	})

	bench.Run("2/Prover", func(b *testing.B) {
		msg2P_0, msg2P_Not0 = p.Two(ctx, msg1V)
	})

	bench.Run("2/Verifier", func(b *testing.B) {
		if msg2P_0 != nil {
			v.Two(ctx, msg1V, msg1P, *msg2P_0)
		} else if msg2P_Not0 != nil {
			v.Two(ctx, msg1V, msg1P, *msg2P_Not0)
		}
	})
}
