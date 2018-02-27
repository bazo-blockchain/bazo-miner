package p2p

import (
	"bufio"
	"errors"
	"fmt"
	"github.com/bazo-blockchain/bazo-miner/protocol"
	"github.com/bazo-blockchain/bazo-miner/storage"
	"log"
	"net"
	"time"
)

func SendClosedTxHashes(closedTxHashes [][32]byte) {
	conn := connect(storage.BOOTSTRAP_SERVER_IP + ":8002")

	packet := BuildPacket(CLOSEDTX_BRDCST, protocol.SerializeSlice32(closedTxHashes))
	conn.Write(packet)

	_, _, err := rcvData2(conn)
	if err != nil {
		logger.Printf("sending closedtx failed: %v\n", err)
		return
	}
}

//func reqTx(txType uint8, txHash [32]byte) (tx protocol.Transaction) {
//
//	conn := Connect(storage.BOOTSTRAP_SERVER)
//
//	packet := p2p.BuildPacket(txType, txHash[:])
//	conn.Write(packet)
//
//	header, payload, err := RcvData(conn)
//	if err != nil {
//		logger.Printf("Disconnected: %v\n", err)
//		return
//	}
//
//	switch header.TypeID {
//	case p2p.ACCTX_RES:
//		var accTx *protocol.AccTx
//		accTx = accTx.Decode(payload)
//		tx = accTx
//	case p2p.CONFIGTX_RES:
//		var configTx *protocol.ConfigTx
//		configTx = configTx.Decode(payload)
//		tx = configTx
//	case p2p.FUNDSTX_RES:
//		var fundsTx *protocol.FundsTx
//		fundsTx = fundsTx.Decode(payload)
//		tx = fundsTx
//	case p2p.STAKETX_RES:
//		var stakeTx *protocol.StakeTx
//		stakeTx = stakeTx.Decode(payload)
//		tx = stakeTx
//	}
//
//	conn.Close()
//
//	return tx
//}

func connect(connectionString string) *net.TCPConn {
	tcpAddr, err := net.ResolveTCPAddr("tcp", connectionString)
	conn, err := net.DialTCP("tcp", nil, tcpAddr)
	if err != nil {
		log.Fatal(err)
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
