package main

/*
#include<stdlib.h>
#include<stdio.h>
extern int32_t getValuefromNode1(void *context, int32_t x);
extern int32_t getValuefromNode2(void *context, int32_t x);
extern int32_t getValuefromNode(void *context, int32_t nodeid, int32_t fruitid);
*/
import "C"
import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	mrand "math/rand"
	"os"
	"strconv"
	"strings"
	"unsafe"

	"github.com/google/uuid"
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p-core/crypto"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p-core/peerstore"

	golog "github.com/ipfs/go-log"
	//	quic "github.com/libp2p/go-libp2p-quic-transport"
	ma "github.com/multiformats/go-multiaddr"
	pb "github.com/wasmerio/go-ext-wasm/nodeutil/pb"

	"github.com/wasmerio/go-ext-wasm/certutil"
	gologging "github.com/whyrusleeping/go-logging"
)
import wasm "github.com/wasmerio/go-ext-wasm/wasmer"

import "github.com/wasmerio/go-ext-wasm/nodeutil"

type Message struct {
	FunctionName string
	Parameter    string
}

//type Certicifate struct {
//	certbyte []byte
//	r *big.Int
//	s *big.Int
//}

//export getValuefromNode
func getValuefromNode(context unsafe.Pointer, nodeid int32, fruitid int32) int32 {
	golog.SetAllLoggers(gologging.INFO) //change to debug for extra info
	ipaddress := "127.0.0.1"
	listenPort := 10003 //!!!!!!!!!!!!!!
	var seed int64
	seed = 2 //!!!!!!!!!

	resCh := make(chan int, 2)
	certCh := make(chan bool, 2)
	//make a host
	ha := makeNode(ipaddress, listenPort, seed, resCh, certCh)

	//set a stream handler on host A
	// /echo/1.0.0 is a user-defined protocol name
	//ha.SetStreamHandler("/echo/1.0.0", func(s network.Stream) {
	//	log.Println("Got a new stream!")
	//	if err := doEcho(s); err != nil {
	//		log.Println(err)
	//		s.Reset()
	//	} else {
	//		s.Close()
	//	}
	//})

	log.Println("listening for connections")
	var target string
	if nodeid == 1 {
		target = "/ip4/127.0.0.1/tcp/10005/ipfs/Qma3emmGx7zTvTstNyppztsp7FBSmXZdCTAoPGnsjJ61cr" ///ip4/127.0.0.1/udp/10000/quic/ipfs/QmexAnfpHrhMmAC5UNQVS8iBuUUgDrMbMY17Cck2gKrqeX"
	} else {
		target = "/ip4/127.0.0.1/tcp/10005/ipfs/Qma3emmGx7zTvTstNyppztsp7FBSmXZdCTAoPGnsjJ61cr" ///ip4/127.0.0.1/udp/10000/quic/ipfs/QmexAnfpHrhMmAC5UNQVS8iBuUUgDrMbMY17Cck2gKrqeX"
	}
	otherID := ha.StorePeer(target)
	log.Printf("This is a conversation between %s and %s\n", ha.ID(), otherID)

	var parameter string
	switch fruitid {
	case 1:
		parameter = "banana"
	case 2:
		parameter = "apple"
	}

	//msg := &Message{
	//	FunctionName: "getValue",
	//	Parameter:    parameter,
	//}
	//res := conact(ha,target,msg)
	//return int32(res)
	req := &pb.EchoRequest{
		MessageData: ha.Node.NewMessageData(
			uuid.New().String(),
			ha.Node.ID(),
			otherID),
		Parameter: parameter,
		FuncName:  "getValue"}
	// add the signature to the message
	signature, err := ha.Node.SignProtoMessage(req)
	if err != nil {
		log.Println("failed to sign message")
		log.Println(err)
		return -1
	}
	req.MessageData.Sign = signature

	privateKey := certutil.GenPrivateKey() // CA's privateKey
	publicKey := &privateKey.PublicKey     // CA's publicKey

	// save CA's privateKey and publickey's parameters(ps: big.int to string)
	x := publicKey.X.String() // x and y will transfer to peers first
	y := publicKey.Y.String()
	d := privateKey.D.String()

	// certs for 3 peers(ps: able to transfer in type []byte)
	cert2byte1, _ := certutil.File2Bytes("/home/lirunnan/version-3.0/go-ext-wasm-master3.0/test/cert1/cert.pem")
	//cert2byte2, _ := test.File2Bytes("/home/lirunnan/version-3.0/go-ext-wasm-master3.0/test/cert2/cert.pem")
	//cert2byte3, _ := test.File2Bytes("/home/lirunnan/version-3.0/go-ext-wasm-master3.0/test/cert3/cert.pem")

	// signatures for 3 peers(Unable to tranfer in type *big.Int, we need to change the type for r, s)
	r1, s1, _ := certutil.GetCASignature(cert2byte1, d)
	//r2, s2, _ := test.GetCASignature(cert2byte2, d)
	//r3, s3, _ := test.GetCASignature(cert2byte3, d)

	R1 := r1.String()
	S1 := s1.String()

	//此处应加上返回值
	ha.SendCert(otherID, R1, S1, cert2byte1, x, y) //!!!!!!!!!!!!!!!!!!!!!!!
	log.Print("pass sendcert")
	//ok := false
	//var certRes bool
	//for ok == false {
	//	log.Print(ok)
	//	certRes, ok = <-certCh
	//}
	//
	//if certRes == false {
	//	log.Println("Certificate validation error")
	//	return -1
	//} else {
	//	log.Println("Certificate validation correctly")
	ha.Echo(otherID, req)
	//}
		//valueRes, err = ReadAll("/home/pixie/tmp/value.txt")
		//if err != nil {
		//	log.Println(err)
		//	return -1
		//}
		//log.Println(valueRes) }

	ok := false
	var result int
	for ok == false {
		result, ok = <-resCh
	}
	return int32(result)
}

//read file
func ReadAll(filePth string) (string, error) {
	f, err := os.Open(filePth)
	if err != nil {
		return "", err
	}
	b, err := ioutil.ReadAll(f)
	return string(b), err
}

//export getValuefromNode1
func getValuefromNode1(context unsafe.Pointer, x int32) int32 {
	golog.SetAllLoggers(gologging.INFO) //change to debug for extra info
	ipaddress := "10.1.19.178"
	listenPort := 10001 //!!!!!!!!!!!!!!
	var seed int64
	seed = 2 //!!!!!!!!!

	//make a host
	ha, err := makeBasicHost(ipaddress, listenPort, seed)
	if err != nil {
		log.Fatal(err)
	}

	//set a stream handler on host A
	// /echo/1.0.0 is a user-defined protocol name
	ha.SetStreamHandler("/echo/1.0.0", func(s network.Stream) {
		log.Println("Got a new stream!")
		if err := doEcho(s); err != nil {
			log.Println(err)
			s.Reset()
		} else {
			s.Close()
		}
	})

	log.Println("listening for connections")

	//for {
	//	//get message
	//	var str string
	//	var target string
	//	n, err := fmt.Scan(&str, &target)
	//	if err != nil {
	//		panic(err)
	//	}
	//	//detemine whether to the end
	//	if n == 2 {
	//		conact(ha, target, str)
	//		break
	//	}
	//}
	//target := "/ip4/10.1.19.65/tcp/10000/ipfs/Qmeq45rCLjFt573aFKgLrcAmAMSmYy9WXTuetDsELM2r8m"
	target2 := "/ip4/127.0.0.1/tcp/10000/ipfs/QmexAnfpHrhMmAC5UNQVS8iBuUUgDrMbMY17Cck2gKrqeX"
	msg := &Message{
		"getValue",
		"banana",
	}
	res := conact(ha, target2, msg)
	//log.Println(res)
	return int32(res)
}

//export getValuefromNode2
func getValuefromNode2(context unsafe.Pointer, x int32) int32 {
	return x
	//向节点2发送请求,变量a的值
	//network.send(Node2Address,"x")
	//network.recv()
	//makeBasicHost获取一个主机对象
	//connect(host,"1","x")
}

func makeNode(ipaddress string, port int, randseed int64, resCh chan int, certCh chan bool) *nodeutil.Node {
	//get privateKey
	var r io.Reader
	r = mrand.New(mrand.NewSource(randseed))
	//priv, _, err := crypto.GenerateKeyPairWithReader(crypto.RSA, 2048, r)
	//priv, _, err := crypto.GenerateKeyPair(crypto.Secp256k1, 256)
	priv, _, err := crypto.GenerateRSAKeyPair(2048, r)
	if err != nil {
		log.Println(err)
		return nil
	}

	//	//create transport
	//	trans, err := quic.NewTransport(priv)
	//	if err != nil {
	//		log.Println(err)
	//		return nil
	//	}

	//create host
	listenAddr, _ := ma.NewMultiaddr(fmt.Sprintf("/ip4/%s/tcp/%d", ipaddress, port))
	host, err := libp2p.New(
		context.Background(),
		libp2p.ListenAddrs(listenAddr),
		libp2p.Identity(priv),
		//		libp2p.Transport(trans),
	)
	if err != nil {
		log.Println(err)
		return nil
	}

	//build host multiaddress
	hostAddr, _ := ma.NewMultiaddr(fmt.Sprintf("/ipfs/%s", host.ID().Pretty())) //
	//build a full multiaddress to reach this host by encapsulating both address
	addr := host.Addrs()[0]
	fullAddr := addr.Encapsulate(hostAddr)
	log.Printf("I am %s\n", fullAddr)

	return nodeutil.NewNode(host, resCh, certCh)
}

func makeBasicHost(ipaddress string, listenPort int, randseed int64) (host.Host, error) { //, insecure bool
	//if seed is 0, use real cryptographic randomness.
	//Otherwise, use a deterministic randomness source to make generated keys stay the same across multuple runs
	var r io.Reader
	//if randseed == 0 {
	//	r = rand.Reader
	//} else {
	//	r = mrand.New(mrand.NewSource(randseed))
	//}
	r = mrand.New(mrand.NewSource(randseed)) //

	//get PrivateKey by r
	priv, _, err := crypto.GenerateKeyPairWithReader(crypto.RSA, 2048, r)
	if err != nil {
		return nil, err
	}

	opts := []libp2p.Option{
		libp2p.ListenAddrStrings(fmt.Sprintf("/ip4/%s/tcp/%d", ipaddress, listenPort)),
		libp2p.Identity(priv),
		libp2p.DisableRelay(),
	}

	//if insecure {
	//	opts = append(opts, libp2p.NoSecurity)
	//}

	basicHost, err := libp2p.New(context.Background(), opts...)
	if err != nil {
		return nil, err
	}

	//build host multiaddress
	hostAddr, _ := ma.NewMultiaddr(fmt.Sprintf("/ipfs/%s", basicHost.ID().Pretty()))

	//build a full multiaddress to reach this host by encapsulating both address
	addr := basicHost.Addrs()[0]
	fullAddr := addr.Encapsulate(hostAddr)
	log.Printf("I am %s\n", fullAddr)

	return basicHost, nil
}

func conact(ha host.Host, target string, message *Message) int {
	//get remote peer multiaddr
	log.Println("entering to conact")
	ipfsaddr, err := ma.NewMultiaddr(target)
	if err != nil {
		log.Fatalln(err)
	}

	//get encrypted peer id
	pid, err := ipfsaddr.ValueForProtocol(ma.P_IPFS)
	if err != nil {
		log.Fatalln(err)
	}
	//decode peer id
	peerid, err := peer.IDB58Decode(pid)
	if err != nil {
		log.Fatalln(err)
	}

	//ipfsaddr : ip4/<a.b.c.d>/ipfs/<peerID>
	//targetPeerAddr : /ipfs/<peerID>
	//targetAddr : ip4/<a.b.c.d>
	targetPeerAddr, _ := ma.NewMultiaddr(
		fmt.Sprintf("/ipfs/%s", peer.IDB58Encode(peerid)))
	targetAddr := ipfsaddr.Decapsulate(targetPeerAddr)

	//store id and targetAddr
	ha.Peerstore().AddAddr(peerid, targetAddr, peerstore.PermanentAddrTTL)

	log.Println("opening stream")
	//make new stream from B to A
	//it should be handled on host A by the handler set above
	//because use the same protocol
	s, err := ha.NewStream(context.Background(), peerid, "/echo/1.0.0")
	if err != nil {
		log.Fatalln(err, peerid)
	}

	log.Println("new stream created")

	//json marshal
	marshalmsg, err := json.Marshal(message)

	log.Println(string(marshalmsg))

	_, err = s.Write([]byte(string(marshalmsg) + "\n"))
	if err != nil {
		log.Fatalln(err)
	}

	buf := bufio.NewReader(s)
	readStr, err := buf.ReadString('\n')
	if err != nil {
		log.Fatalln(err)
	}

	//out, err := ioutil.ReadAll(s)
	//if err != nil {
	//	log.Fatalln(err)
	//}
	fmt.Printf("read reply: %s", readStr)
	readStr = strings.Replace(readStr, "\n", "", -1)
	res, err := strconv.Atoi(readStr)
	if err != nil {
		log.Fatalln(err)
	}
	return res
}

//read a line of data a stream and wirte it back
func doEcho(s network.Stream) error {
	buf := bufio.NewReader(s)
	readStr, err := buf.ReadString('\n')
	if err != nil {
		return err
	}

	log.Printf("read: %s\n", readStr)

	var repStr = "5\n"
	_, err = s.Write([]byte(repStr))
	return err
}

func main() {
	//fmt.Println("6")

	//control the verbosity level for all loggers with

	//parse option from command line
	//listenF := flag.Int("l", 0, "wait for incoming connections")
	//target := flag.String("d", "", "target peer to dial")
	//insecure := flag.Bool("indecure", false, "use an unencrypted connection")
	//seed := flag.Int64("seed", 0, "set random seed for id generation")
	//flag.Parse()

	/*if *listenF == 0 {
		log.Fatal("Please provide a port to bind on with -l")
	}*/

	imports, err := wasm.NewImports().Append("getValuefromNode1", getValuefromNode1, C.getValuefromNode1)

	if err != nil {
		panic(err)
	}

	imports.Append("getValuefromNode2", getValuefromNode2, C.getValuefromNode2)

	imports.Append("getValuefromNode", getValuefromNode, C.getValuefromNode)

	bytes, err := wasm.ReadBytes("/home/lirunnan/version-3.0/go-ext-wasm-master3.0/node1/addsum2.wasm")
	if err != nil {
		panic(err)
	}

	instance, err := wasm.NewInstanceWithImports(bytes, imports)
	if err != nil {
		panic(err)
	}
	defer instance.Close()

	//四个参数分别为：nodeid1,fruitid1,nodeid2,fruitid2
	results, err := instance.Exports["addsum2"](1, 1, 1, 2)
	if err != nil {
		panic(err)
	}
	log.Printf("计算结果为: %v", results)
	select {} //hang forever
}
