package certutil

import (
	"crypto/ecdsa"
	"crypto/rand"
	//"testProject/fabric-ca/test"
	"math/big"
	"crypto/elliptic"
)

func GetCASignature(cert2byte []byte, d string) (r, s *big.Int, err error){

	privateKey := &ecdsa.PrivateKey{}
	privateKey.Curve = elliptic.P384()
	dd := &big.Int{}
	dd.SetString(d, 10)
	privateKey.D = dd

	summary:= GetHashValue(cert2byte)

	// 进入签名操作
	return ecdsa.Sign(rand.Reader, privateKey, summary)

}
