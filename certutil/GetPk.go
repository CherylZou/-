//证书生成

package certutil

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	//"crypto/x509"
	//"crypto/x509/pkix"
	//"encoding/pem"
	"math/big"
	//"net"
	//"os"
	//"time"
)


//var CApk, _ = ecdsa.GenerateKey(elliptic.P384(), rand.Reader)

//func init() {

//}

type PublicKey struct {
	elliptic.Curve
	X, Y *big.Int
}

type PrivateKey struct {
	PublicKey
	D *big.Int
}

func GenPrivateKey() *ecdsa.PrivateKey {
	var pk, _ = ecdsa.GenerateKey(elliptic.P384(), rand.Reader)
	return pk
}
