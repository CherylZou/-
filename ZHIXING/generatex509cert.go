//证书生成

package main

import (
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"net"
	"os"
	"time"
	"test"
)

type PublicKey struct {
	elliptic.Curve
	X, Y *big.Int
}

type PrivateKey struct {
	PublicKey
	D *big.Int
}

func main(){
	max := new(big.Int).Lsh(big.NewInt(1),128)  //把 1 左移 128 位，返回给 big.Int
	serialNumber, _ := rand.Int(rand.Reader, max)   //返回在 [0, max) 区间均匀随机分布的一个随机值
	subject := pkix.Name{   //Name代表一个X.509识别名。只包含识别名的公共属性，额外的属性被忽略。
		Organization:       []string{"Hyper chain Co."},
		OrganizationalUnit: []string{"fbz"},
		CommonName:         "CA from lx",
	}
	template := x509.Certificate{
		SerialNumber:   serialNumber, // SerialNumber 是 CA 颁布的唯一序列号，在此使用一个大随机数来代表它
		Subject:        subject,
		NotBefore:      time.Now(),
		NotAfter:       time.Now().Add(365 * 24 *time.Hour),
		KeyUsage:       x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature, //KeyUsage 与 ExtKeyUsage 用来表明该证书是用来做服务器认证的
		ExtKeyUsage:    []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth}, // 密钥扩展用途的序列
		IPAddresses:    []net.IP{net.ParseIP("127.0.0.1")},
	}

	pk := test.GenPrivateKey()

	//CreateCertificate基于模板创建一个新的证书
	//第二个第三个参数相同，则证书是自签名的
	//返回的切片是DER编码的证书
	derBytes, _ := x509.CreateCertificate(rand.Reader, &template, &template, &pk.PublicKey, pk) //DER 格式
	certOut, _ := os.Create("/home/shorel/go/src/testProject/fabric-ca/test/cert2/cert.pem")
	pem.Encode(certOut,&pem.Block{Type:"CERTIFICATE", Bytes: derBytes})
	certOut.Close()
	keyOut, _ := os.Create("/home/shorel/go/src/testProject/fabric-ca/test/cert2/key.pem")
	x509bytes, _ := x509.MarshalECPrivateKey(pk)
	pem.Encode(keyOut, &pem.Block{Type: "EC PRIVATE KEY", Bytes: x509bytes})
	keyOut.Close()

}
