package p2p

import (
	"bufio"
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/bazo-blockchain/bazo-miner/storage"
	"net"
	"time"
)

func Connect(connectionString string) *net.TCPConn {
	logger = storage.InitLogger()
	tcpAddr, err := net.ResolveTCPAddr("tcp", connectionString)
	conn, err := net.DialTCP("tcp", nil, tcpAddr)
	if err != nil {
		logger.Printf("Connection to %v failed.\n", connectionString)
		return nil
	}

	conn.SetLinger(0)
	conn.SetDeadline(time.Now().Add(20 * time.Second))

	return conn
}

func rcvData(p *peer) (header *Header, payload []byte, err error) {
	reader := bufio.NewReader(p.conn)
	header, err = ReadHeader(reader)
	if err != nil {
		p.conn.Close()
		return nil, nil, errors.New(fmt.Sprintf("Connection to %v aborted: %v", p.getIPPort(), err))
	}
	payload = make([]byte, header.Len)

	for cnt := 0; cnt < int(header.Len); cnt++ {
		payload[cnt], err = reader.ReadByte()
		if err != nil {
			p.conn.Close()
			return nil, nil, errors.New(fmt.Sprintf("Connection to %v aborted: %v", p.getIPPort(), err))
		}
	}

	//logger.Printf("Receive message:\nSender: %v\nType: %v\nPayload length: %v\n", p.getIPPort(), logMapping[header.TypeID], len(payload))
	return header, payload, nil
}

func RcvData(c net.Conn) (header *Header, payload []byte, err error) {
	reader := bufio.NewReader(c)
	header, err = ReadHeader(reader)
	if err != nil {
		c.Close()
		return nil, nil, errors.New(fmt.Sprintf("Connection to aborted: (%v)\n", err))
	}
	payload = make([]byte, header.Len)

	for cnt := 0; cnt < int(header.Len); cnt++ {
		payload[cnt], err = reader.ReadByte()
		if err != nil {
			c.Close()
			return nil, nil, errors.New(fmt.Sprintf("Connection to aborted: %v\n", err))
		}
	}

	return header, payload, nil
}

func sendData(p *peer, payload []byte) {
	//logger.Printf("Send message:\nReceiver: %v\nType: %v\nPayload length: %v\n", p.getIPPort(), logMapping[payload[4]], len(payload)-HEADER_LEN)
	p.l.Lock()
	p.conn.Write(payload)
	p.l.Unlock()
}

//Tested in server_test.go
func peerExists(newIpport string) bool {

	peerList := peers.getAllPeers()

	for _, p := range peerList {
		ipport := p.getIPPort()
		if ipport == newIpport {
			return true
		}
	}

	return false
}

//Tested in server_test.go
func peerSelfConn(newIpport string) bool {
	return newIpport == Ipport
}

func BuildPacket(typeID uint8, payload []byte) (packet []byte) {

	var payloadLen [4]byte
	packet = make([]byte, HEADER_LEN+len(payload))
	binary.BigEndian.PutUint32(payloadLen[:], uint32(len(payload)))
	copy(packet[0:4], payloadLen[:])
	packet[4] = byte(typeID)
	copy(packet[5:], payload)
	return packet
}

func ReadHeader(reader *bufio.Reader) (*Header, error) {
	//the first four bytes of any incoming messages is the length of the payload
	//error catching after every read is necessary to avoid panicking
	var headerArr [HEADER_LEN]byte
	//reading byte by byte is surprisingly fast and works a lot better for concurrent connections
	for i := range headerArr {
		extr, err := reader.ReadByte()
		if err != nil {
			return nil, err
		}
		headerArr[i] = extr
	}

	header := extractHeader(headerArr[:])
	return header, nil
}

//Decoupled functionality for testing reasons
func extractHeader(headerData []byte) *Header {

	header := new(Header)

	lenBuf := [4]byte{headerData[0], headerData[1], headerData[2], headerData[3]}
	packetLen := binary.BigEndian.Uint32(lenBuf[:])

	header.Len = packetLen
	header.TypeID = uint8(headerData[4])
	return header
}
