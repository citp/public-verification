package nm

import (
	"testing"

	"github.com/citp/pvphm/bv"
)

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
	x := string(bv.RandomBytes(15))
	H1 := ctx.pc.ctxDH.HashToCurve(string(bv.RandomBytes(15)))
	H2 := ctx.pc.ctxDH.HashToCurve(string(bv.RandomBytes(15)))

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
