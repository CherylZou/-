package main

import (
	"fmt"
	"github.com/wasmerio/go-ext-wasm/certutil"
)

func main() {

	privateKey := certutil.GenPrivateKey() // CA's privateKey
	publicKey := &privateKey.PublicKey     // CA's publicKey

	// save CA's privateKey and publickey's parameters(ps: big.int to string)
	x := publicKey.X.String() // x and y will transfer to peers first
	y := publicKey.Y.String()
	d := privateKey.D.String()

    // certs for 3 peers(ps: able to transfer in type []byte)
	cert2byte1, _ := certutil.File2Bytes("/home/shorel/go/src/testProject/fabric-ca/test/cert1/cert.pem")
	cert2byte2, _ := certutil.File2Bytes("/home/shorel/go/src/testProject/fabric-ca/test/cert2/cert.pem")
	cert2byte3, _ := certutil.File2Bytes("/home/shorel/go/src/testProject/fabric-ca/test/cert3/cert.pem")

    // signatures for 3 peers(Unable to tranfer in type *big.Int, we need to change the type for r, s)
	r1, s1, _ := certutil.GetCASignature(cert2byte1, d)
	r2, s2, _ := certutil.GetCASignature(cert2byte2, d)
	r3, s3, _ := certutil.GetCASignature(cert2byte3, d)
/*
	// Change the type and now can be transferred
	R1 := r1.String(); S1 := s1.String()
	R2 := r2.String(); S2 := s2.String()
	R3 := r3.String(); S3 := s3.String()
*/

/*
	// reduce the type to *big.Int for the r, s
	r := &big.Int{}
	r, ok := n.SetString(R1, 10)
	if !ok {
		panic("type change error occurred!")
	}
*/
    // 进入验证
	flag1 := certutil.VerifySignature(r1, s1, cert2byte1, x, y)
	flag2 := certutil.VerifySignature(r2, s2, cert2byte2, x, y)
	flag3 := certutil.VerifySignature(r3, s3, cert2byte3, x, y)

	if flag1 {
		fmt.Println("证书认证通过")
	} else {
		fmt.Println("证书认证不通过")
	}

	if flag2 {
		fmt.Println("证书认证通过")
	} else {
		fmt.Println("证书认证不通过")
	}

	if flag3 {
		fmt.Println("证书认证通过")
	} else {
		fmt.Println("证书认证不通过")
	}

}
