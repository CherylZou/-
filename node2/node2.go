// main
package main

import (
	//"bufio"
	"context"
	//"encoding/json"
	//"strconv"

	//"crypto/rand"
	//"flag"
	"fmt"
	"io"
	//"io/ioutil"
	"log"
	mrand "math/rand"

	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p-core/crypto"
	//"github.com/libp2p/go-libp2p-core/host"
	//"github.com/libp2p/go-libp2p-core/network"
	//"github.com/libp2p/go-libp2p-core/peer"
	//"github.com/libp2p/go-libp2p-core/peerstore"

	golog "github.com/ipfs/go-log"
	ma "github.com/multiformats/go-multiaddr"
	gologging "github.com/whyrusleeping/go-logging"

	quic "github.com/libp2p/go-libp2p-quic-transport"
	"github.com/wasmerio/go-ext-wasm/nodeutil"
)

type Message struct {
	FunctionName string
	Parameter    string
}

//var banana  = &entity.Fruit{}
//
//var apple  = &entity.Fruit{}

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

	//create transport
	trans, err := quic.NewTransport(priv)
	if err != nil {
		log.Println(err)
		return nil
	}

	//create host
	listenAddr, _ := ma.NewMultiaddr(fmt.Sprintf("/ip4/%s/udp/%d/quic", ipaddress, port))
	host, err := libp2p.New(
		context.Background(),
		libp2p.ListenAddrs(listenAddr),
		libp2p.Identity(priv),
		libp2p.Transport(trans),
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

/*
//create host with a random peer ID on given multiaddress.
//It won't encrypt the connection if insecure is true
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
*/

func main() {
	//control the verbosity level for all loggers with
	golog.SetAllLoggers(gologging.INFO) //change to debug for extra info

	//parse option from command line
	//listenF := flag.Int("l", 0, "wait for incoming connections")
	//target := flag.String("d", "", "target peer to dial")
	//insecure := flag.Bool("indecure", false, "use an unencrypted connection")
	//seed := flag.Int64("seed", 0, "set random seed for id generation")
	//flag.Parse()

	/*if *listenF == 0 {
		log.Fatal("Please provide a port to bind on with -l")
	}*/

	ipaddress := "127.0.0.1"
	listenPort := 10005 //!!!!!!!!!!!!!!
	var seed int64
	seed = 3 //!!!!!!!!!

	resCh := make(chan int, 2)
	certCh := make(chan bool, 2)
	_ = makeNode(ipaddress, listenPort, seed, resCh, certCh)
	log.Println("listening for connections...")

	select {} //hang forever
}

//	//make a host
//	ha, err := makeBasicHost(ipaddress, listenPort, seed)
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	banana.SetFruitName("banana")
//	banana.SetAmount(20)
//
//	apple.SetFruitName("apple")
//	apple.SetAmount(30)
//
//	//set a stream handler on host A
//	// /echo/1.0.0 is a user-defined protocol name
//	ha.SetStreamHandler("/echo/1.0.0", func(s network.Stream) {
//		log.Println("Got a new stream!")
//		if err := doEcho(s); err != nil {
//			log.Println(err)
//			s.Reset()
//		} else {
//			s.Close()
//		}
//	})
//
//	log.Println("listening for connections")
//	for {
//		//get message
//		var str string
//		var target string
//		n, err := fmt.Scan(&str, &target)
//		if err != nil {
//			panic(err)
//		}
//		//detemine whether to the end
//		if n == 2 {
//			conact(ha, target, str)
//			break
//		}
//	}
//	//target := "/ip4/10.1.19.65/tcp/10000/ipfs/Qmeq45rCLjFt573aFKgLrcAmAMSmYy9WXTuetDsELM2r8m"
//	////msg := "helloNode1"
//	//msg := &Message{
//	//	functionName: "getValue",
//	//	parameter:    "x",
//	//}
//	//conact(ha,target,*msg)
//}
//
//func conact(ha host.Host, target string, str string) {
//	//get remote peer multiaddr
//	ipfsaddr, err := ma.NewMultiaddr(target)
//	if err != nil {
//		log.Fatalln(err)
//	}
//
//	//get encrypted peer id
//	pid, err := ipfsaddr.ValueForProtocol(ma.P_IPFS)
//	if err != nil {
//		log.Fatalln(err)
//	}
//	//decode peer id
//	peerid, err := peer.IDB58Decode(pid)
//	if err != nil {
//		log.Fatalln(err)
//	}
//
//	//ipfsaddr : ip4/<a.b.c.d>/ipfs/<peerID>
//	//targetPeerAddr : /ipfs/<peerID>
//	//targetAddr : ip4/<a.b.c.d>
//	targetPeerAddr, _ := ma.NewMultiaddr(
//		fmt.Sprintf("/ipfs/%s", peer.IDB58Encode(peerid)))
//	targetAddr := ipfsaddr.Decapsulate(targetPeerAddr)
//
//	//store id and targetAddr
//	ha.Peerstore().AddAddr(peerid, targetAddr, peerstore.PermanentAddrTTL)
//
//	log.Println("opening stream")
//	//make new stream from B to A
//	//it should be handled on host A by the handler set above
//	//because use the same protocol
//	s, err := ha.NewStream(context.Background(), peerid, "/echo/1.0.0")
//	if err != nil {
//		log.Fatalln(err, peerid)
//	}
//
//	_, err = s.Write([]byte(str + "\n"))
//	if err != nil {
//		log.Fatalln(err)
//	}
//
//	buf := bufio.NewReader(s)
//	readStr, err := buf.ReadString('\n')
//	if err != nil {
//		log.Fatalln(err)
//	}
//
//	//out, err := ioutil.ReadAll(s)
//	//if err != nil {
//	//	log.Fatalln(err)
//	//}
//
//	log.Printf("read reply: %s\n", readStr)
//}
//
////read a line of data a stream and wirte it back
//func doEcho(s network.Stream) error {
//	log.Printf("entering to doEcho")
//	buf := bufio.NewReader(s)
//	readStr,_ := buf.ReadString('\n')
//	log.Printf("read: %s", readStr)
//	msg2 := &Message{}
//	err := json.Unmarshal([]byte(readStr),&msg2)
//	log.Printf(msg2.FunctionName)
//	log.Printf(msg2.Parameter)
//
//		//paramatervalue,_ :=strconv.Atoi(msg2.Parameter)
//	if msg2.FunctionName == "getValue" {
//		//returnvalue := int(getValue(int32(paramatervalue)))
//		//log.Println(returnvalue)
//		//repStr := strconv.Itoa(returnvalue)
//		//log.Println(repStr)
//		//_, err = s.Write([]byte(repStr + "\n"))
//		if msg2.Parameter == "banana" {
//			returnvalue := banana.GetAmount()
//			log.Print(returnvalue)
//			repStr := strconv.Itoa(returnvalue)
//			//log.Println(repStr)
//			_, err = s.Write([]byte(repStr + "\n"))
//		} else if msg2.Parameter == "apple" {
//			returnvalue := apple.GetAmount()
//			log.Print(returnvalue)
//			repStr := strconv.Itoa(returnvalue)
//			//log.Println(repStr)
//			_, err = s.Write([]byte(repStr + "\n"))
//		} else {
//			log.Print("没有此水果")
//		}
//	}
//
//	if err != nil {
//		return err
//	}
//	return err
//}
