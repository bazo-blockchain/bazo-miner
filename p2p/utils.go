package p2p

import (
	"bufio"
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/bazo-blockchain/bazo-miner/protocol"
	"github.com/bazo-blockchain/bazo-miner/storage"
	"net"
	"strings"
	"time"
)

func Connect(connectionString string) *net.TCPConn {
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

func RcvData(p *peer) (header *Header, payload []byte, err error) {
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

	//logger.Printf("Receive message:\nSender: %v\nType: %v\nPayload length: %v\n", p.getIPPort(), LogMapping[header.TypeID], len(payload))
	return header, payload, nil
}

func RcvData_(c net.Conn) (header *Header, payload []byte, err error) {
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
	p.l.Lock()
	p.conn.Write(payload)
	p.l.Unlock()
}

//Tested in server_test.go
func peerExists(newIpport string) bool {
	peerList := peers.getAllPeers(PEERTYPE_MINER)

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
	FileLogger.Printf("BuildPacket: typeID - %d ¦ payload length - %d\n",typeID,len(payload))
	var payloadLen [4]byte

	packet = make([]byte, HEADER_LEN+len(payload))
	binary.BigEndian.PutUint32(payloadLen[:], uint32(len(payload)))
	copy(packet[0:4], payloadLen[:])
	packet[4] = byte(typeID)
	copy(packet[5:], payload)

	return packet
}

func ReadHeader(reader *bufio.Reader) (*Header, error) {
	//The first four bytes of any incoming messages is the length of the payload.
	//Error catching after every read is necessary to avoid panicking.

	var headerArr [HEADER_LEN]byte

	//FileLogger.Printf("Reader Size: %d",reader.Size())
	//
	//buf := new(bytes.Buffer)
	//buf.ReadFrom(reader)
	//s := buf.String() // Does a complete copy of the bytes in the buffer.
	//FileLogger.Printf("Reader to String: %v",s)
	//
	//if b, err := ioutil.ReadAll(reader); err == nil {
	//	FileLogger.Printf("Whole reader to String: %v",b)
	//}

	//var buf bytes.Buffer
	//tee := io.TeeReader(reader, &buf)
	//teeString, _ := ioutil.ReadAll(tee)
	//FileLogger.Printf("Reader to String: %x\n",string(teeString))


	//Reading byte by byte is surprisingly fast and works a lot better for concurrent connections.
	for i := range headerArr {
		extr, err := reader.ReadByte()
		//FileLogger.Printf("Byte read at position %d: %x",i,string(extr))
		if err != nil {
			return nil, err
		}

		headerArr[i] = extr
	}

	header := extractHeader(headerArr[:])
	//Check if the type is registered in the protocol.
	if LogMapping[header.TypeID] == "" {
		FileLogger.Printf("Header Length: %d -- Header TypeID: %d\n",header.Len,header.TypeID)
		return nil, errors.New("Header: TypeID not found.")
	}

	//Check if the payload length does not exceed the MAX_BLOCK_SIZE defined in configtx.go
	if header.Len > protocol.MAX_BLOCK_SIZE {
		return nil, errors.New("Header: Payload exceeds MAX_BLOCK_SIZE")
	}

	return header, nil
}

//Decoupled functionality for testing reasons.
func extractHeader(headerData []byte) *Header {
	FileLogger.Printf("Header Data: %v",headerData)
	header := new(Header)

	lenBuf := [4]byte{headerData[0], headerData[1], headerData[2], headerData[3]}
	packetLen := binary.BigEndian.Uint32(lenBuf[:])

	header.Len = packetLen
	header.TypeID = uint8(headerData[4])

	return header
}

func IsBootstrap() bool {
	//Set thisPort global, this will be the listening port for incoming connection
	bootstrapPort := strings.Split(storage.BootstrapServer, ":")[1]
	thisPort := strings.Split(Ipport, ":")[1]
	if thisPort == bootstrapPort {
		return true
	}
	return false
}