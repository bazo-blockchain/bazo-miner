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
