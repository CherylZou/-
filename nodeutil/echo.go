package nodeutil

import (
	"io/ioutil"
	"log"
	"math/big"

	//"os"
	//"strconv"

	"github.com/gogo/protobuf/proto"
	"github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/peer"

	"github.com/google/uuid"
	pb "github.com/wasmerio/go-ext-wasm/nodeutil/pb"
	"github.com/wasmerio/go-ext-wasm/test"
)

// pattern: /protocol-name/request-or-response-message/version
const echoRequest = "/echo/echoreq/0.0.1"
const echoResponse = "/echo/echoresp/0.0.1"
const echoCert = "/echo/echocert/0.0.1"

type EchoProtocol struct {
	Node     *Node                      // local host
	requests map[string]*pb.EchoRequest // used to access request data from response handlers
	ResCh    chan int
	CertCh   chan bool
}

func NewEchoProtocol(node *Node, resCh chan int, cerCh chan bool) *EchoProtocol { //, done chan bool
	e := EchoProtocol{Node: node, requests: make(map[string]*pb.EchoRequest), ResCh: resCh, CertCh: cerCh}
	node.SetStreamHandler(echoRequest, e.onEchoRequest)
	node.SetStreamHandler(echoResponse, e.onEchoResponse)
	node.SetStreamHandler(echoCert, e.onEchoCert)
	return &e
}

// remote peer cert handler
func (e *EchoProtocol) onEchoCert(s network.Stream) {

	// get cert data
	data := &pb.EchoCert{}
	buf, err := ioutil.ReadAll(s)
	if err != nil {
		s.Reset()
		log.Println(err)
		return
	}
	s.Close()

	// unmarshal it
	proto.Unmarshal(buf, data)
	if err != nil {
		log.Println(err)
		return
	}

	//authenticate message
	valid := e.Node.authenticateMessage(data, data.MessageData)
	if !valid {
		log.Println("onEchoCert Failed to authenticate message")
		return
	}

	//print
	log.Printf("Receive cert\n")
	log.Printf("	MessageID: %s\n", data.MessageData.Id)
	log.Printf("	SourcePeer: %s\n", s.Conn().RemotePeer())
	log.Printf("	TargetPeer: %s\n", s.Conn().LocalPeer())
	log.Printf("	PubKey elliptic curve X: %s\n", data.X)
	log.Printf("	PubKey elliptic curve Y: %s\n", data.Y)
	log.Printf("	Content of the certificate: %s\n", string(data.Cert))
	log.Printf("	Sign R: %s\n", data.R)
	log.Printf("	Sign S: %s\n", data.S)

	rr := &big.Int{}
	rr, ok := rr.SetString(data.X, 10)
	if !ok {
		panic("type change error occurred!")
	}
	//log.Print(rr)

	ss := &big.Int{}
	ss, ok2 := ss.SetString(data.Y, 10)
	if !ok2 {
		panic("type change error occurred!")
	}
	//log.Print(ss)

	flag := test.VerifySignature(rr, ss, data.Cert, data.R, data.S)
	log.Print(flag)
	e.CertCh <- flag
	if flag == true {
		log.Print("证书验证通过")
	} else {
		log.Print("证书验证未通过")
	}
	//certRes, ok := <- e.CertCh
	//log.Print(ok)
	//log.Print(certRes)
}

// remote peer requests handler
func (e *EchoProtocol) onEchoRequest(s network.Stream) {
	ok1 := false
	var certRes bool
	for ok1 == false {
		certRes, ok1 = <-e.CertCh
	}
	log.Print("onEchoRequest")
	// get request data
	data := &pb.EchoRequest{}
	buf, err := ioutil.ReadAll(s)
	if err != nil {
		s.Reset()
		log.Println(err)
		return
	}
	s.Close()

	// unmarshal it
	proto.Unmarshal(buf, data)
	if err != nil {
		log.Println(err)
		return
	}

	//authenticate message
	valid := e.Node.authenticateMessage(data, data.MessageData)
	if !valid {
		log.Println("onEchoRequest Failed to authenticate message")
		return
	}

	if certRes == false {
		resp := &pb.EchoResponse{
			MessageData: e.Node.NewMessageData(
				data.MessageData.Id,
				s.Conn().LocalPeer(),
				s.Conn().RemotePeer()),
			Message: -1}

		// sign the data
		signature, err := e.Node.SignProtoMessage(resp)
		if err != nil {
			log.Println("failed to sign response")
			return
		}
		// add the signature to the message
		resp.MessageData.Sign = signature

		ok := e.Node.SendProtoMessage(s.Conn().RemotePeer(), echoResponse, resp)
		if ok {
			log.Println("The response has been sented")
		} else {
			log.Println("Failed to sent the response")
		}
		return
	}

	//print
	log.Printf("Receive request\n")
	log.Printf("	MessageID: %s\n", data.MessageData.Id)
	log.Printf("	SourcePeer: %s\n", s.Conn().RemotePeer())
	log.Printf("	TargetPeer: %s\n", s.Conn().LocalPeer())
	log.Printf("	FunctionName: %s\n", data.FuncName)
	log.Printf("	Parameteer: %s\n", data.Parameter)

	// send response to the request using the message string he provided
	log.Printf("%s: Sending echo response to %s. Message id: %s...", s.Conn().LocalPeer(), s.Conn().RemotePeer(), data.MessageData.Id)
	msg, err := e.Node.Call(data.FuncName, data.Parameter)
	if err != nil {
		log.Println(err)
		return
	}
	resp := &pb.EchoResponse{
		MessageData: e.Node.NewMessageData(
			data.MessageData.Id,
			s.Conn().LocalPeer(),
			s.Conn().RemotePeer()),
		Message: msg[0].Int()}

	// sign the data
	signature, err := e.Node.SignProtoMessage(resp)
	if err != nil {
		log.Println("failed to sign response")
		return
	}
	// add the signature to the message
	resp.MessageData.Sign = signature

	ok := e.Node.SendProtoMessage(s.Conn().RemotePeer(), echoResponse, resp)
	if ok {
		log.Println("The response has been sented")
	} else {
		log.Println("Failed to sent the response")
	}
}

// remote echo response handler
func (e *EchoProtocol) onEchoResponse(s network.Stream) {

	// get response data
	data := &pb.EchoResponse{}
	buf, err := ioutil.ReadAll(s)
	if err != nil {
		s.Reset()
		log.Println(err)
		return
	}
	s.Close()

	// unmarshal it
	proto.Unmarshal(buf, data)
	if err != nil {
		log.Println(err)
		return
	}

	log.Printf("Receive response\n")
	log.Printf("	MessageID: %s\n", data.MessageData.Id)
	log.Printf("	SourcePeer: %s\n", s.Conn().RemotePeer())
	log.Printf("	TargetPeer: %s\n", s.Conn().LocalPeer())
	log.Printf("	Message: %d\n", data.Message)

	// authenticate message content
	valid := e.Node.authenticateMessage(data, data.MessageData)
	if !valid {
		log.Println("onEchoResponse Failed to authenticate message")
		return
	}

	e.ResCh <- int(data.Message)
	//save to file
	//f, err := os.Create("/home/pixie/tmp/value.txt")
	//if err != nil {
	//	log.Println(err)
	//	log.Println("create file error")
	//	return
	//}
	//_, err = f.WriteString(strconv.FormatInt(data.Message, 10))
	//if err != nil {
	//	log.Println(err)
	//	f.Close()
	//	return
	//}

	// locate request data and remove it if found
	_, ok := e.requests[data.MessageData.Id]
	if ok {
		// remove request from map as we have processed it here
		log.Println("Successed to locate request data boject for response")
		delete(e.requests, data.MessageData.Id)
	} else {
		log.Println("Failed to locate request data boject for response")
		return
	}
}

//send request
func (e *EchoProtocol) Echo(otherId peer.ID, req *pb.EchoRequest) bool {
	log.Printf("%s: Sending echo to: %s....", e.Node.ID(), otherId)

	ok := e.Node.SendProtoMessage(otherId, echoRequest, req)
	if !ok {
		return false
	}
	log.Print("pass echo")
	e.requests[req.MessageData.Id] = req

	return true
}

//send cert
func (e *EchoProtocol) SendCert(otherId peer.ID, x string, y string, cert2bytes []byte, r string, s string) bool { //host host.Host
	log.Printf("%s: Sending cert to: %s....", e.Node.ID(), otherId)

	// create message data
	certificate := &pb.EchoCert{
		MessageData: e.Node.NewMessageData(
			uuid.New().String(),
			e.Node.ID(),
			otherId),
		X:    x,
		Y:    y,
		Cert: cert2bytes,
		R:    r,
		S:    s,
	}

	// add the signature to the message
	signature, err := e.Node.SignProtoMessage(certificate)
	if err != nil {
		log.Println("failed to sign message")
		return false
	}
	certificate.MessageData.Sign = signature

	ok := e.Node.SendProtoMessage(otherId, echoCert, certificate)
	if !ok {
		return false
	}

	return true
}
