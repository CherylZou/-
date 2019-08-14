package test

import (
	"crypto/ecdsa"
	"crypto/sha1"
	"math/big"
	"crypto/elliptic"
)


func VerifySignature(r, s *big.Int, cert2byte []byte, x string, y string) bool{

	publicKey := &ecdsa.PublicKey{}
	publicKey.Curve = elliptic.P384()

	xx := &big.Int{}
	xx.SetString(x, 10)
	yy := &big.Int{}
	yy.SetString(y, 10)

	publicKey.X = xx
	publicKey.Y = yy

	Sha1Inst := sha1.New()
	Sha1Inst.Write(cert2byte)
	Result := Sha1Inst.Sum([]byte(""))
	// 进入验证
	return ecdsa.Verify(publicKey, Result, r, s)
}

