package main

import (
	"bytes"
	"fmt"
	"math/big"
	"runtime"
	"sort"
	"strings"

	"github.com/cloudflare/bn256"
	// "go.dedis.ch/kyber/v3/pairing/bn256"
)

type BVContext struct {
	N          int
	tau        int
	aggSeed    DHScalar
	dh         DHContext
	shamir     ShamirContext
	tblsShamir ShamirContext
	bls        BLS_Context
}

type BVPData struct {
	L DHElement
	P []DHElement
}

type BVServer struct {
	ctx   BVContext
	X     []string
	alpha DHScalar
	pdata BVPData
	Xfreq map[string][]int
}

type BVGroup struct {
	ctx   BVContext
	id    int
	Xi    []string
	sk    BLS_SecretKey
	aggPk BLS_PublicKey
}

// -----------------------------------------------------------------------------

type BVMSG_Header struct {
	id    int
	round string
}

type BVMSG_Internal struct {
	header BVMSG_Header
	m1gs   *BVMSG_1_GS
	m1gg   *BVMSG_1_GG
	m2s    *BVMSG_2_S
	m3gg   *BVMSG_3_GG
	// m2bgg  *BVMSG_2BGG
	m3gs *BVMSG_3_GS
	// m3gs  *BVMSG_3GS
}

// n
type BVMSG_1_GS struct {
	seedPRF *big.Int // 32 bytes
	Xi      string 
}

// Total = 64 * N^2 bytes
type BVMSG_1_GG struct {
	seedPRF  *big.Int // 32 bytes
	shareBLS *big.Int // 32 bytes
}

// Total = (97 + (65 * |X|)) * N bytes
type BVMSG_2_S struct {
	seedPRF    *big.Int // 32 bytes
	aggSeedPRF *big.Int // 32 bytes
	pdata      string   // 33 bytes * (|X| + 1)
	X          string   // |X| * 32 bytes
}

// Total = 161 * N^2 bytes
type BVMSG_3_GG struct {
	aggSeedPRF *big.Int // 32 bytes
	pk         BLS_PublicKey // 129 bytes
}

// type BVMSG_2BGG struct {
// 	pk BLS_PublicKey
// }

// Total = ((97 * |X|) + 129) * N
type BVMSG_3_GS struct {
	aggPk BLS_PublicKey // 129 bytes
	Q     string // |X| * 33 bytes
	ct    string // |X| * 64 bytes
}

// Total comm. = (((97 * |X|) + 129) * N) + (161 * N^2) + ((97 + (65 * |X|)) * N) + (64 * N^2) = 225 N^2 + N (65 |X| + 97) + N (97 |X| + 129)

// type BVMSG_3GS struct {
// 	aggPk BLS_PublicKey
// }

// -----------------------------------------------------------------------------

func (v BVMSG_Header) MarshalBinary() ([]byte, error) {
	var b bytes.Buffer
	_, err := fmt.Fprintf(&b, "%d %s", v.id, v.round)
	return b.Bytes(), err
}

func (v *BVMSG_Header) UnmarshalBinary(data []byte) error {
	b := bytes.NewBuffer(data)
	_, err := fmt.Fscanf(b, "%d %s", &v.id, &v.round)
	return err
}

func (v BVMSG_1_GS) MarshalBinary() ([]byte, error) {
	var b bytes.Buffer
	_, err := fmt.Fprintf(&b, "%s", v.Xi)
	return append(v.seedPRF.Bytes(), b.Bytes()...), err
}

func (v *BVMSG_1_GS) UnmarshalBinary(data []byte) error {
	v.seedPRF = new(big.Int)
	v.seedPRF.SetBytes(data[:32])
	b := bytes.NewBuffer(data[32:])
	_, err := fmt.Fscanf(b, "%s", &v.Xi)
	return err
}

func (v BVMSG_1_GG) MarshalBinary() ([]byte, error) {
	return append(v.seedPRF.Bytes(), v.shareBLS.Bytes()...), nil

	// fmt.Println(len(v.shareBLS.Bytes()), len(v.seedPRF.Bytes()))
	// var b bytes.Buffer
	// _, err := fmt.Fprintf(&b, "%s", v.seedPRF)
	// return b.Bytes(), err
}

func (v *BVMSG_1_GG) UnmarshalBinary(data []byte) error {
	v.seedPRF = new(big.Int)
	v.shareBLS = new(big.Int)
	v.seedPRF.SetBytes(data[:32])
	v.shareBLS.SetBytes(data[32:])
	return nil

	// b := bytes.NewBuffer(data)
	// _, err := fmt.Fscanf(b, "%s", &v.seedPRF)
	// return err
}

func (v BVMSG_2_S) MarshalBinary() ([]byte, error) {
	var b bytes.Buffer
	_, err := fmt.Fprintf(&b, "%s %s", v.pdata, v.X)
	return append(append(v.seedPRF.Bytes(), v.aggSeedPRF.Bytes()...), b.Bytes()...), err
}

func (v *BVMSG_2_S) UnmarshalBinary(data []byte) error {
	v.seedPRF = new(big.Int)
	v.aggSeedPRF = new(big.Int)
	v.seedPRF.SetBytes(data[:32])
	v.aggSeedPRF.SetBytes(data[32:64])

	b := bytes.NewBuffer(data[64:])
	_, err := fmt.Fscanf(b, "%s %s", &v.pdata, &v.X)
	return err
}

func (v BVMSG_3_GG) MarshalBinary() ([]byte, error) {
	// var b bytes.Buffer
	// _, err := fmt.Fprintf(&b, "%s %s", v.aggSeed, v.share)
	// return b.Bytes(), err
	return append(v.aggSeedPRF.Bytes(), (*v.pk).Marshal()...), nil
}

func (v *BVMSG_3_GG) UnmarshalBinary(data []byte) error {
	v.aggSeedPRF = new(big.Int)
	v.aggSeedPRF.SetBytes(data[:32])
	v.pk = new(bn256.G2)
	_, err := (*v.pk).Unmarshal(data[32:])
	return err

	// b := bytes.NewBuffer(data)
	// _, err := fmt.Fscanf(b, "%s %s", &v.aggSeed, &v.share)
	// return err
}

// func (v BVMSG_2BGG) MarshalBinary() ([]byte, error) {

// 	return (*v.pk).Marshal(), nil
// }

// func (v *BVMSG_2BGG) UnmarshalBinary(data []byte) error {
// 	v.pk = new(bn256.G2)
// 	_, err := (*v.pk).Unmarshal(data)
// 	return err
// }

func (v BVMSG_3_GS) MarshalBinary() ([]byte, error) {
	var b bytes.Buffer
	// size 129
	// fmt.Println("aggpk size", len((*v.aggPk).Marshal()))
	_, err := fmt.Fprintf(&b, "%s %s", v.ct, v.Q)
	return append((*v.aggPk).Marshal(), b.Bytes()...), err
}

func (v *BVMSG_3_GS) UnmarshalBinary(data []byte) error {
	b := bytes.NewBuffer(data[129:])
	_, err := fmt.Fscanf(b, "%s %s", &v.ct, &v.Q)
	if err != nil {
		return err
	}
	v.aggPk = new(bn256.G2)
	_, err = (*v.aggPk).Unmarshal(data[:129])
	return err
}

// func (v BVMSG_3GS) MarshalBinary() ([]byte, error) {

// 	return (*v.aggPk).Marshal(), nil
// }

// func (v *BVMSG_3GS) UnmarshalBinary(data []byte) error {
// 	v.aggPk = new(bn256.G2)
// 	_, err := (*v.aggPk).Unmarshal(data)
// 	return err
// }

// #############################################################################
// #############################################################################

func NewBVContext(N, tau int) BVContext {
	var ctx BVContext
	ctx.N = N
	ctx.tau = tau
	ctx.dh = NewDHContext()
	ctx.aggSeed = big.NewInt(0)
	ctx.bls = NewBLS_Context()
	ctx.shamir = NewShamirContext(tau, N, ctx.dh.Curve.Params().P)
	ctx.tblsShamir = NewShamirContext(tau, N, ctx.bls.order)
	return ctx
}

func NewBVGroup(ctx *BVContext, id int, Xi []string) BVGroup {
	return BVGroup{*ctx, id, Xi, big.NewInt(0), nil}
}

func NewBVServer(ctx *BVContext) BVServer {
	var srv BVServer
	srv.ctx = *ctx
	srv.X = make([]string, 0)
	srv.alpha = RandomScalar(ctx.dh.Curve.Params().P)
	return srv
}

// #############################################################################
// #############################################################################

func (grp *BVGroup) SignElements(pdata BVPData, X []string) []BLS_Signature {
	// defer timer(time.Now(), "SignElements")
	Xi := make(map[string]int)
	ret := make([]BLS_Signature, len(X))

	for _, x := range grp.Xi {
		Xi[x] = 1
	}

	for i, x := range X {
		_, ok := Xi[x]
		if ok {
			ret[i] = grp.ctx.bls.BLS_Sign(grp.sk, []byte(pdata.P[i].String()))
		}
	}
	return ret
}

// type EncKeyGetterIn struct {
// 	idx int
// 	P DHElement
// }

// type EncKeyGetterOut struct {
// 	idx int
// 	ret []byte
// 	q   string
// }

func (grp *BVGroup) GetEncKeys(pdata BVPData, X []string) ([][]byte, []string) {
	// defer timer(time.Now(), "GetEncKeys")
	ret := make([][]byte, len(pdata.P))
	Q := make([]string, len(pdata.P))

	H := make([]DHElement, len(pdata.P))
	for j := 0; j < len(pdata.P); j++ {
		H[j] = grp.ctx.dh.HashToCurve(X[j])
	}
	Qout, Sout := grp.ctx.dh.DH_Reduce_Parallel(pdata.P, pdata.L, H, runtime.NumCPU())

	for j := 0; j < len(pdata.P); j++ {
		// _, _, q, S := grp.ctx.dh.DH_Reduce(pdata.P[j], pdata.L, grp.ctx.dh.HashToCurve(X[j]))
		ret[j] = Blake2b([]byte(Sout[j].String()))
		Q[j] = Qout[j].String()
	}

	return ret, Q
}

// returns hex encoded signatures
func (grp *BVGroup) EncyptSigns(signs []BLS_Signature, keys [][]byte) []string {
	// defer timer(time.Now(), "EncyptSigns")
	Assert(len(signs) == len(keys))
	ret := make([]string, len(keys))

	for i := 0; i < len(signs); i++ {
		if signs[i] != nil {
			ret[i] = AES_Encrypt(keys[i], (*signs[i]).Marshal())
		}
	}
	return ret
}

// -----------------------------------------------------------------------------

func (grp *BVGroup) One() ([]BVMSG_1_GG, BVMSG_1_GS) {
	// defer timer(time.Now(), "Group (Round 1)")

	// ret := grp.ctx.dh.Curve.Scalar()
	// ret.Pick(grp.ctx.dh.Curve.RandomStream())
	seedPRF := RandomScalar(grp.ctx.dh.Curve.Params().P)
	skSeed := grp.ctx.bls.BLS_SecKeygen()
	_, shares := grp.ctx.tblsShamir.NewSharing(skSeed)

	msgG := make([]BVMSG_1_GG, grp.ctx.N)
	for i := 0; i < len(msgG); i++ {
		msgG[i] = BVMSG_1_GG{seedPRF, shares[i].y}
	}

	msgS := BVMSG_1_GS{seedPRF, strings.Join(grp.Xi, ",")}

	return msgG, msgS
}

func (grp *BVGroup) Two(msgG []BVMSG_1_GG, msgS BVMSG_2_S) ([]string, BVPData) {
	// defer timer(time.Now(), "Group (Round 2)")

	skShares := []*big.Int{}
	seedShares := []*big.Int{msgS.seedPRF}
	for _, m := range msgG {
		skShares = append(skShares, m.shareBLS)
		seedShares = append(seedShares, m.seedPRF)
	}
	grp.sk = AggregateFieldElems(skShares, grp.ctx.bls.order)
	grp.ctx.aggSeed = AggregateFieldElems(seedShares, grp.ctx.dh.Curve.Params().P)

	// AggregateFieldElems(msgG[0].shareBLS)
	// grp.ctx.aggSeed = msgS.seedPRF
	// // grp.ctx.aggSeed = BigIntFrom(msgS.seed)
	// for _, m := range msgs {
	// 	// other := BigIntFrom(m.seed)
	// 	(*grp.ctx.aggSeed).Add(grp.ctx.aggSeed, m.seedPRF)
	// 	(*grp.ctx.aggSeed).Mod(grp.ctx.aggSeed, grp.ctx.dh.Curve.Params().P)
	// 	(*grp.sk).Add(grp.sk, m.shareBLS)
	// 	(*grp.sk).Mod(grp.sk, grp.ctx.bls.order)
	// }

	// fmt.Println("Aggregate Seed:", (*grp.ctx.aggSeed).Text(16))
	// fmt.Println("sk", (*grp.sk).Text(16))

	X := strings.Split(msgS.X, ",")
	pdataStrs := strings.Split(msgS.pdata, "|")
	var pdata BVPData
	pdata.L = DHElementFrom(pdataStrs[0])
	pdata.P = make([]DHElement, len(pdataStrs)-1)
	for i := 1; i < len(pdataStrs); i++ {
		pdata.P[i-1] = DHElementFrom(pdataStrs[i])
	}
	// fmt.Println("|X| = ", len(X))
	// fmt.Println("|PData.P| = ", len(pdataStrs))
	Assert(len(X) == len(pdataStrs)-1)
	Assert((*msgS.aggSeedPRF).Cmp(grp.ctx.aggSeed) == 0)

	return X, pdata
}

func (grp *BVGroup) ThreeA() []BVMSG_3_GG {
	// fmt.Println("Group: 3A")
	// defer timer(time.Now(), "Group (Round 3A)")
	msgs := make([]BVMSG_3_GG, grp.ctx.N)
	// secret := RandomScalar(grp.ctx.shamir.mod)

	// _, shares := grp.ctx.shamir.NewSharing(secret)
	for i := 0; i < grp.ctx.N; i++ {
		msgs[i] = BVMSG_3_GG{grp.ctx.aggSeed, grp.ctx.bls.BLS_PubKeygen(grp.sk)}
	}

	return msgs
}

func (grp *BVGroup) ThreeB(msgs []BVMSG_3_GG, X []string, pdata BVPData) BVMSG_3_GS {
	// defer timer(time.Now(), "Group (Round 3B)")
	shares := make([]BLS_PublicKey_Share, grp.ctx.N)
	for i := 0; i < grp.ctx.N; i++ {
		Assert((*msgs[i].aggSeedPRF).Text(16) == (*grp.ctx.aggSeed).Text(16))
		shares[i].x = big.NewInt(int64(i + 1))
		shares[i].y = msgs[i].pk
	}

	grp.aggPk = grp.ctx.bls.BLS_AggPK(&grp.ctx.tblsShamir, shares)
	signs := grp.SignElements(pdata, X)
	keys, Q := grp.GetEncKeys(pdata, X)
	ct := grp.EncyptSigns(signs, keys)

	return BVMSG_3_GS{grp.aggPk, strings.Join(Q, "|"), strings.Join(ct, "|")}
}

// func (grp *BVGroup) TwoC(msgs []BVMSG_2BGG, X []string, pdata BVPData) BVMSG_3_GS {
// 	var shares []BLS_PublicKey_Share
// 	for i := 0; i < grp.ctx.N; i++ {
// 		shares = append(shares, BLS_PublicKey_Share{big.NewInt(int64(i + 1)), msgs[i].pk})
// 	}
// 	grp.aggPk = grp.ctx.bls.BLS_AggPK(&grp.ctx.shamir, shares)
// signs := grp.SignElements(X)
// keys, Q := grp.GetEncKeys(pdata, X)
// ct := grp.EncyptSigns(signs, keys)

// 	var msg BVMSG_3_GS
// 	msg.ct = strings.Join(ct, "|")
// 	msg.Q = strings.Join(Q, "|")
// 	msg.pk = grp.aggPk

// 	// fmt.Println("pk", (*grp.aggPk).String())

// 	return msg
// }

// #############################################################################
// #############################################################################

func (srv *BVServer) GenerateX(Xi [][]string) {
	srv.Xfreq = make(map[string][]int)
	for i := 0; i < len(Xi); i++ {
		for j := 0; j < len(Xi[i]); j++ {
			if val, ok := srv.Xfreq[Xi[i][j]]; ok {
				srv.Xfreq[Xi[i][j]] = append(val, i+1)
			} else {
				srv.Xfreq[Xi[i][j]] = []int{i + 1}
			}
		}
	}

	for k, v := range srv.Xfreq {
		if len(v) >= srv.ctx.tau {
			srv.X = append(srv.X, k)
		}

		sort.Slice(v, func(i, j int) bool { return v[i] < v[j] })
		srv.Xfreq[k] = v
	}
}

func (srv *BVServer) AggregateSigns(msgs []BVMSG_3_GS) []BLS_Signature {
	shares := make([][]BLS_Signature_Share, len(srv.X))
	for i := 0; i < len(msgs); i++ {
		Qstr := strings.Split(msgs[i].Q, "|")
		ct := strings.Split(msgs[i].ct, "|")
		Assert(len(Qstr) == len(srv.X))
		for j := 0; j < len(Qstr); j++ {
			if len(Qstr[j]) > 0 {
				S := srv.ctx.dh.EC_Multiply(srv.alpha, DHElementFrom(Qstr[j]))
				sk := Blake2b([]byte(S.String()))
				if len(ct[j]) > 0 {
					pt, err := AES_Decrypt(sk, ct[j])
					Check(err)
					share := BLS_Signature_Share{big.NewInt(int64(i + 1)), new(bn256.G1)}
					_, err = (*share.y).Unmarshal(pt)
					Check(err)
					shares[j] = append(shares[j], share)
				}
			}
		}
	}

	signs := make([]BLS_Signature, len(srv.X))
	for i := 0; i < len(signs); i++ {
		Assert(len(shares[i]) >= srv.ctx.tau)
		// shares[i] =
		// sort.Slice(shares[i], func (x, y int) { return shares[x] < shares[y] })
		// fmt.Println("share", i)
		// for j := 0; j < len(shares[i]); j++ {
		// 	fmt.Println(shares[i][j].x)
		// }
		signs[i] = srv.ctx.bls.BLS_AggSign(&srv.ctx.tblsShamir, shares[i])
	}
	return signs
}

func (srv *BVServer) AggregatePublicKeys(pk []BLS_PublicKey) []BLS_PublicKey {
	keys := make([]BLS_PublicKey, len(srv.X))

	for i, x := range srv.X {
		groups := srv.Xfreq[x]
		shares := make([]BLS_PublicKey_Share, len(groups))
		for j := 0; j < len(groups); j++ {
			shares[j].x = big.NewInt(int64(groups[j]))
			shares[j].y = pk[groups[j]-1]
		}
		keys[i] = srv.ctx.bls.BLS_AggPK(&srv.ctx.tblsShamir, shares)
		// fmt.Println(groups, (*keys[i]).String())
	}
	return keys
}

// -----------------------------------------------------------------------------

func (srv *BVServer) Two(msgs []BVMSG_1_GS) BVMSG_2_S {
	// fmt.Println("Server: 2")
	// defer timer(time.Now(), "Server (Round 2)")
	seedPRF := RandomScalar(srv.ctx.dh.Curve.Params().P)
	srv.ctx.aggSeed = new(big.Int)
	(*srv.ctx.aggSeed).Set(seedPRF)
	var Xi [][]string

	for _, m := range msgs {
		// other := BigIntFrom(m.seed)
		srv.ctx.aggSeed = (*srv.ctx.aggSeed).Add(srv.ctx.aggSeed, m.seedPRF)
		srv.ctx.aggSeed = (*srv.ctx.aggSeed).Mod(srv.ctx.aggSeed, srv.ctx.dh.Curve.Params().P)
		X := strings.Split(m.Xi, ",")
		Xi = append(Xi, X)
	}
	srv.GenerateX(Xi)
	// fmt.Println("|srv.X| = ", len(srv.X))
	srv.pdata.P = make([]DHElement, len(srv.X))
	srv.pdata.L = srv.ctx.dh.EC_BaseMultiply(srv.alpha)
	pdataStrs := make([]string, len(srv.X)+1)
	pdataStrs[0] = srv.pdata.L.String()
	for i, x := range srv.X {
		// fmt.Println(i, x)
		srv.pdata.P[i] = srv.ctx.dh.EC_Multiply(srv.alpha, srv.ctx.dh.HashToCurve(x))
		pdataStrs[i+1] = srv.pdata.P[i].String()
	}

	return BVMSG_2_S{seedPRF, srv.ctx.aggSeed, strings.Join(pdataStrs, "|"), strings.Join(srv.X, ",")}
}

func (srv *BVServer) Three(msgs []BVMSG_3_GS) (BLS_PublicKey, [][]byte, []BLS_Signature) {
	// fmt.Println("Server: 3")
	// defer timer(time.Now(), "Server (Round 3)")
	Assert(len(msgs) == srv.ctx.N)
	aggPk := msgs[0].aggPk
	for i := 1; i < srv.ctx.N; i++ {
		Assert((*msgs[i].aggPk).String() == (*aggPk).String())
	}

	signs := srv.AggregateSigns(msgs)
	Assert(len(signs) == len(srv.X))
	// pkShares := make([]BLS_PublicKey_Share, srv.ctx.N)
	// for i := 0; i < srv.ctx.N; i++ {
	// 	pkShares[i].x = big.NewInt(int64(i + 1))
	// 	pkShares[i].y = msgs[i].pk
	// }
	// aggPk := srv.ctx.bls.BLS_AggPK(&srv.ctx.tblsShamir, pkShares)

	// aggPk := srv.AggregatePublicKeys(pk)
	// for i := 0; i < len(srv.X); i++ {
	// 	Assert(signs[i] != nil)
	// 	// 	// fmt.Println((*aggPk[i]).String())
	// 	res := srv.ctx.bls.BLS_Verify(aggPk, []byte(srv.pdata.P[i].String()), signs[i])
	// 	// fmt.Println(i, res, len(srv.Xfreq[srv.X[i]]))
	// 	Assert(res)
	// }
	pdataStrs := make([][]byte, len(srv.X))
	for i := 0; i < len(srv.X); i++ {
		pdataStrs[i] = []byte(srv.pdata.P[i].String())
	}

	// VerifySignatures(srv.ctx.bls, aggPk, pdataStrs, signs)
	// Assert(srv.ctx.bls.BLS_Verify_Parallel(aggPk, pdataStrs, signs))
	return aggPk, pdataStrs, signs
}
