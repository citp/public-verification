package nm

import (
	"testing"

	"github.com/citp/pvphm/bv"
)

// For 128 bit security, need log_2(6) * 128 ~ 331 iterations
var NM_ROUNDS_128 int = 331
var NM_STRING_LEN int = 15

func BenchmarkNM(bench *testing.B) {
	var prover NM_Prover
	var verifier NM_Verifier

	for i := 0; i < bench.N; i++ {
		ctx := NewNM_Context(NM_ROUNDS_128)

		bench.Run("Init", func(b *testing.B) {
			HelperNMInit(b, &ctx, &prover, &verifier)
		})

		bench.Run("Proof", func(b *testing.B) {
			for j := 0; j < NM_ROUNDS_128; j++ {
				HelperNMInner(b, &ctx, &prover, &verifier)
			}
		})
	}
}

// #############################################################################

func HelperNMInit(bench *testing.B, ctx *NM_Context, p *NM_Prover, v *NM_Verifier) {
	x := string(RandomBytes(NM_STRING_LEN))
	H1 := ctx.pc.ctxDH.HashToCurve(string(RandomBytes(NM_STRING_LEN)))
	H2 := ctx.pc.ctxDH.HashToCurve(string(RandomBytes(NM_STRING_LEN)))

	bench.Run("Prove", func(b *testing.B) {
		p.Init(ctx, x, H1, H2)
	})

	bench.Run("Verify", func(b *testing.B) {
		v.Init(ctx, x, H1, H2)
	})
}

func HelperNMInner(bench *testing.B, ctx *NM_Context, p *NM_Prover, v *NM_Verifier) {

	var msg1V NM_MSG_1V
	var msg1P NM_MSG_1P
	var msg2P_0 *NM_MSG_2P_0
	var msg2P_Not0 *NM_MSG_2P_Not0

	msg1V = v.One(ctx)
	msg1P = p.One(ctx)
	msg2P_0, msg2P_Not0 = p.Two(ctx, msg1V)

	if msg2P_0 != nil {
		v.Two(ctx, msg1V, msg1P, *msg2P_0)
	} else if msg2P_Not0 != nil {
		v.Two(ctx, msg1V, msg1P, *msg2P_Not0)
	}
}

// #############################################################################

func HelperNMFSInit(bench *testing.B, ctx *NM_Context, p *NMFS_Prover, v *NM_Verifier) {
	x := string(RandomBytes(NM_STRING_LEN))
	H1 := ctx.pc.ctxDH.HashToCurve(string(RandomBytes(NM_STRING_LEN)))
	H2 := ctx.pc.ctxDH.HashToCurve(string(RandomBytes(NM_STRING_LEN)))

	bench.Run("Prove", func(b *testing.B) {
		p.Init(ctx, x, H1, H2)
	})

	bench.Run("Verify", func(b *testing.B) {
		v.Init(ctx, x, H1, H2)
	})
}

func BenchmarkNMFS(bench *testing.B) {
	var prover NMFS_Prover
	var verifier NM_Verifier
	var msg NMFS_MSG_P

	ctx := NewNM_Context(NM_ROUNDS_128)

	bench.Run("Init", func(b *testing.B) {
		HelperNMFSInit(b, &ctx, &prover, &verifier)
	})

	for i := 0; i < bench.N; i++ {

		bench.Run("NIZK Proof Gen", func(b *testing.B) {
			msg = prover.FiatShamir(&ctx)
		})

		bench.Run("NIZK Proof Ver", func(b *testing.B) {
			Assert(verifier.FiatShamir(&ctx, msg))
		})
	}

}
