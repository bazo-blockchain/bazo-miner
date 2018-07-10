package p2p

import (
	"github.com/bazo-blockchain/bazo-miner/protocol"
	"github.com/bazo-blockchain/bazo-miner/storage"
)

func SendVerifiedTxs(txs []*protocol.FundsTx) {
	if conn := Connect(storage.BOOTSTRAP_SERVER_IP + ":8002"); conn != nil {
		var verifiedTxs [][]byte

		for _, tx := range txs {
			verifiedTxs = append(verifiedTxs, tx.Encode()[:])
		}

		packet := BuildPacket(VERIFIEDTX_BRDCST, protocol.Encode(verifiedTxs, protocol.FUNDSTX_SIZE))
		conn.Write(packet)

		header, _, err := RcvData_(conn)
		if err != nil || header.TypeID != TX_BRDCST_ACK {
			logger.Printf("Sending verified tx failed.")
		}

		conn.Close()
	}
}
