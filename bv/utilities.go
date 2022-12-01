package bv

import (
	"bufio"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"math/big"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
)

// -----------------------------------------------------------------------------

type stackTracer interface {
	StackTrace() errors.StackTrace
}

func PrintTrace(e error) {
	eST, ok := errors.Cause(e).(stackTracer)
	if !ok {
		panic("wtf: error in error handling")
	}

	st := eST.StackTrace()
	fmt.Printf("error -> %s\n", e.Error())
	fmt.Printf("%+v\n", st[1:3])
}

func Assert(v bool) {
	if !v {
		cause := errors.New("Assertion failed")
		e := errors.WithStack(cause)
		PrintTrace(e)
	}
}

func Check(err error) {
	if err != nil {
		cause := errors.New(err.Error())
		e := errors.WithStack(cause)
		PrintTrace(e)
	}
}

// -----------------------------------------------------------------------------

func ReadFile(fpath string) []string {
	file, err := os.Open(fpath)
	Check(err)
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var list []string
	for scanner.Scan() {
		list = append(list, scanner.Text())
	}
	err = scanner.Err()
	Check(err)

	return list
}

func IDfromIP(ip string) int {
	id, err := strconv.Atoi(strings.Split(ip, ".")[3])
	Check(err)
	return id
}

func RandomBytes(n int) []byte {
	ret := make([]byte, n)
	nR, err := io.ReadFull(rand.Reader, ret)
	Assert(nR == n)
	Check(err)
	return ret
}

func AES_Encrypt(sk, pt []byte) string {
	block, err := aes.NewCipher(sk)
	Check(err)
	aesgcm, err := cipher.NewGCM(block)
	Check(err)
	nonce := RandomBytes(12)
	ct := aesgcm.Seal(nil, nonce, pt, nil)
	return hex.EncodeToString(ct) + "!" + hex.EncodeToString(nonce)
}

func AES_Decrypt(sk []byte, ctStr string) ([]byte, error) {
	ctStrs := strings.Split(ctStr, "!")
	ct, err := hex.DecodeString(ctStrs[0])
	Check(err)
	nonce, err := hex.DecodeString(ctStrs[1])
	Check(err)

	block, err := aes.NewCipher(sk)
	Check(err)
	aesgcm, err := cipher.NewGCM(block)
	Check(err)
	return aesgcm.Open(nil, nonce, ct, nil)

}

// -----------------------------------------------------------------------------

func AggregateFieldElems(arr []*big.Int, mod *big.Int) *big.Int {
	ret := big.NewInt(int64(0))
	for _, r := range arr {
		ret.Add(ret, r)
		ret.Mod(ret, mod)
	}
	return ret
}

func VerifySignatures(ctx BLS_Context, pk BLS_PublicKey, msgs [][]byte, sigs []BLS_Signature) bool {
	defer timer(time.Now(), "VerifySignatures")
	Assert(len(sigs) == len(msgs))
	for i := 0; i < len(sigs); i++ {
		if !ctx.BLS_Verify(pk, msgs[i], sigs[i]) {
			return false
		}
	}
	return true
}

func timer(start time.Time, name string) {
	elapsed := time.Since(start)
	log.Printf("%s took %s", name, elapsed)
}

// func ScalarFrom(dh *DHContext, s string) kyber.Scalar {
// 	ret := dh.Curve.Scalar()
// 	retInt, ok := big.NewInt(0).SetString(s, 16)
// 	if !ok {
// 		panic("Could not deserialize " + s)
// 	}
// 	ret.UnmarshalBinary(retInt.Bytes())
// 	return ret
// }
