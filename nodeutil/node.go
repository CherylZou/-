package nodeutil

import (
	"context"
	"errors"
	"fmt"
	"log"
	"reflect"
	"time"

	"github.com/libp2p/go-libp2p-core/crypto"
	"github.com/libp2p/go-libp2p-core/helpers"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p-core/peerstore"
	"github.com/libp2p/go-libp2p-core/protocol"

	p2p "github.com/wasmerio/go-ext-wasm/nodeutil/pb"

	ggio "github.com/gogo/protobuf/io"
	proto "github.com/gogo/protobuf/proto"
	ma "github.com/multiformats/go-multiaddr"
	"github.com/wasmerio/go-ext-wasm/entity"
)

var banana = &entity.Fruit{
	"banana",
	20,
}

var apple = &entity.Fruit{
	"apple",
	30,
}

// node client version
const clientVersion = "go-p2p-node/0.0.1"

// Node type - a p2p host implementing one or more p2p protocols
type Node struct {
	host.Host     // lib-p2p host
	*EchoProtocol // echo protocol impl
	// add other protocols here...
	funcs map[string]interface{}
}

// Create a new node with its implemented protocols
func NewNode(host host.Host, resCh chan int, certCh chan bool) *Node {
	node := &Node{Host: host}
	node.EchoProtocol = NewEchoProtocol(node, resCh, certCh)
	node.funcs = map[string]interface{}{
		"getValue": getValue,
		"getSum":   getSum,
	}
	return node
}

func getValue(fruitname string) int {
	if fruitname == "banana" {
		return banana.GetAmount()
	} else if fruitname == "apple" {
		return apple.GetAmount()
	} else {
		return 0
	}
}

func getSum(a int, b int) int {
	return a + b
}

// helper method - generate message data shared between all node's p2p protocols
// messageId: unique for requests, copied from request for responses
func (n *Node) NewMessageData(messageId string, sourcePeer peer.ID, targetPeer peer.ID) *p2p.MessageData { //, gossip bool
	// Add protobufs bin data for message author public key
	// this is useful for authenticating  messages forwarded by a node authored by another node

	nodePubKey, err := n.Peerstore().PubKey(n.ID()).Bytes()
	if err != nil {
		panic("Failed to get public key for sender from local peer store.")
	}

	return &p2p.MessageData{
		Id:         messageId,
		SourcePeer: sourcePeer.String(),
		TargetPeer: targetPeer.String(),
		NodePubKey: nodePubKey,
		Timestamp:  time.Now().Unix(),
	}
}

// helper method - writes a protobuf go data object to a network stream
// data: reference of protobuf go data object to send (not the object itself)
// s: network stream to write the data to
func (n *Node) SendProtoMessage(id peer.ID, p protocol.ID, data proto.Message) bool {
	//create stream
	s, err := n.NewStream(context.Background(), id, p)
	if err != nil {
		log.Fatalln(err)
		return false
	}
	//write message
	writer := ggio.NewFullWriter(s)
	err = writer.WriteMsg(data)
	if err != nil {
		log.Println(err)
		s.Reset()
		return false
	}

	//closes the stream and waits for the other side to close their half
	err = helpers.FullClose(s)
	if err != nil {
		log.Println(err)
		s.Reset()
		return false
	}
	return true
}

//call function
func (n *Node) Call(funcName string, params ...interface{}) (result []reflect.Value, err error) {

	//get function
	f := reflect.ValueOf(n.funcs[funcName])
	if len(params) != f.Type().NumIn() {
		err = errors.New("The number of params is not adapted.")
		return nil, err
	}

	in := make([]reflect.Value, len(params))
	for k, param := range params {
		in[k] = reflect.ValueOf(param)
	}
	result = f.Call(in)
	return result, nil
}

func (n *Node) StorePeer(target string) peer.ID {
	//get remote peer multiaddr
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
	n.Peerstore().AddAddr(peerid, targetAddr, peerstore.PermanentAddrTTL)
	return peerid
}

// sign an outgoing p2p message payload
func (n *Node) SignProtoMessage(message proto.Message) ([]byte, error) {
	data, err := proto.Marshal(message)
	if err != nil {
		return nil, err
	}
	return n.signData(data)
}

// sign binary data using the local node's private key
func (n *Node) signData(data []byte) ([]byte, error) {
	key := n.Peerstore().PrivKey(n.ID())
	res, err := key.Sign(data)
	return res, err
}

// Authenticate incoming p2p message
// message: a protobufs go data object
// data: common p2p message data
func (n *Node) authenticateMessage(message proto.Message, messageData *p2p.MessageData) bool {
	// store a temp ref to signature and remove it from message data
	// sign is a string to allow easy reset to zero-value (empty string)
	sign := messageData.Sign
	messageData.Sign = nil

	// marshall data without the signature to protobufs3 binary format
	bin, err := proto.Marshal(message)
	if err != nil {
		log.Println(err, "failed to marshal pb message")
		return false
	}

	// restore sig in message data (for possible future use)
	messageData.Sign = sign

	// restore peer id binary format from base58 encoded node id data
	peerId, err := peer.IDB58Decode(messageData.SourcePeer)
	if err != nil {
		log.Println(err, "Failed to decode node id from base58")
		return false
	}

	// verify the data was authored by the signing peer identified by the public key
	// and signature included in the message
	return n.verifyData(bin, []byte(sign), peerId, messageData.NodePubKey)
}

// Verify incoming p2p message data integrity
// data: data to verify
// signature: author signature provided in the message payload
// peerId: author peer id from the message payload
// pubKeyData: author public key from the message payload
func (n *Node) verifyData(data []byte, signature []byte, peerId peer.ID, pubKeyData []byte) bool {
	key, err := crypto.UnmarshalPublicKey(pubKeyData)
	if err != nil {
		log.Println(err, "Failed to extract key from message key data")
		return false
	}

	// extract node id from the provided public key
	idFromKey, err := peer.IDFromPublicKey(key)

	if err != nil {
		log.Println(err, "Failed to extract peer id from public key")
		return false
	}

	// verify that message author node id matches the provided node public key
	if idFromKey != peerId {
		log.Println(err, "Node id and provided public key mismatch")
		return false
	}

	res, err := key.Verify(data, signature)
	if err != nil {
		log.Println(err, "Error authenticating data")
		return false
	}

	return res
}
