package p2p

import (
	"bufio"
	"errors"
	"fmt"
	"github.com/bazo-blockchain/bazo-miner/protocol"
	"github.com/bazo-blockchain/bazo-miner/storage"
	"net"
	"time"
)

func SendVerifiedTxs(receivedTxs []*protocol.FundsTx) {
	if conn := connect(storage.BOOTSTRAP_SERVER_IP + ":8002"); conn != nil {
		encodedReceivedTxs := make([]byte, len(receivedTxs)*protocol.FUNDSTX_SIZE)
		index := 0

		for _, tx := range receivedTxs {
			copy(encodedReceivedTxs[index:index+protocol.FUNDSTX_SIZE], tx.Encode())
			index += protocol.FUNDSTX_SIZE
		}

		packet := BuildPacket(RECEIVEDTX_BRDCST, encodedReceivedTxs)
		conn.Write(packet)

		_, _, err := rcvData2(conn)
		if err != nil {
			logger.Printf("Could not send the verified transactions.\n")
		}

		conn.Close()
	}
}

func connect(connectionString string) *net.TCPConn {
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

func rcvData2(c net.Conn) (header *Header, payload []byte, err error) {
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
