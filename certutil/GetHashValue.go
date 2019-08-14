package certutil

import (
	"crypto/sha1"
)

func GetHashValue(cert2byte []byte) ([]byte) {
	/*
	Md5Inst := md5.New()
	//Md5Inst.Write([]byte(TestString))
	Md5Inst.Write(cert2bytes)
	Result := Md5Inst.Sum([]byte(""))
	fmt.Printf("%x\n\n", Result)
	*/

	Sha1Inst := sha1.New()
	Sha1Inst.Write(cert2byte)
	Result := Sha1Inst.Sum([]byte(""))
	return Result
}
