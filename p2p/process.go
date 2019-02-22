package p2p

import (
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/bazo-blockchain/bazo-miner/protocol"
	"github.com/bazo-blockchain/bazo-miner/storage"
	"strconv"
	"time"
)

var(
	writtenTXCount 				= 0
)

//Process tx broadcasts from other miners. We can't broadcast incoming messages directly, first check if
//the tx has already been broadcast before, whether it is a valid tx etc.
func processTxBrdcst(p *peer, payload []byte, brdcstType uint8) {
	var tx protocol.Transaction
	//Make sure the transaction can be properly decoded, verification is done at a later stage to reduce latency
	switch brdcstType {
	case FUNDSTX_BRDCST:
		var fTx *protocol.FundsTx
		fTx = fTx.Decode(payload)
		if fTx == nil {
			return
		}
		tx = fTx
	case ACCTX_BRDCST:
		var aTx *protocol.ContractTx
		aTx = aTx.Decode(payload)
		if aTx == nil {
			return
		}
		tx = aTx
	case CONFIGTX_BRDCST:
		var cTx *protocol.ConfigTx
		cTx = cTx.Decode(payload)
		if cTx == nil {
			return
		}
		tx = cTx
	case STAKETX_BRDCST:
		var sTx *protocol.StakeTx
		sTx = sTx.Decode(payload)
		if sTx == nil {
			return
		}
		tx = sTx
	}

	//Response tx acknowledgment if the peer is a client
	if !peers.minerConns[p] {
		packet := BuildPacket(TX_BRDCST_ACK, nil)
		sendData(p, packet)
	}

	if storage.ReadOpenTx(tx.Hash()) != nil {
		logger.Printf("Received transaction (%x) already in the mempool.\n", tx.Hash())
		FileLogger.Printf("Received transaction (%x) already in the mempool.\n", tx.Hash())
		return
	}
	if storage.ReadClosedTx(tx.Hash()) != nil {
		logger.Printf("Received transaction (%x) already validated.\n", tx.Hash())
		FileLogger.Printf("Received transaction (%x) already validated.\n", tx.Hash())
		return
	}

	//Write to mempool and rebroadcast
	logger.Printf("Writing transaction (%x) in the mempool.\n", tx.Hash())
	FileLogger.Printf("Writing transaction (%x) in the mempool.\n", tx.Hash())
	logger.Printf("Writing transaction at time: %d\n", time.Now().Unix())
	FileLogger.Printf("Writing transaction at time: %d\n", time.Now().Unix())

	writtenTXCount += 1

	logger.Printf("Written tx count: %d\n", writtenTXCount)
	FileLogger.Printf("Written tx count: %d\n", writtenTXCount)

	storage.WriteOpenTx(tx)

	//for p := range peers.minerConns {
	//	if err := SendTx(p.getIPPort(), tx, FUNDSTX_BRDCST); err != nil {
	//		return
	//	}
	//}

	//toBrdcst := BuildPacket(brdcstType, payload)
	//minerBrdcstMsg <- toBrdcst
}

func SendTx(dial string, tx protocol.Transaction, typeID uint8) (err error) {
	if conn := Connect(dial); conn != nil {
		packet := BuildPacket(typeID, tx.Encode())
		conn.Write(packet)

		header, payload, err := RcvData_(conn)
		if err != nil || header.TypeID == NOT_FOUND {
			err = errors.New(string(payload[:]))
		}
		conn.Close()

		return err
	}

	txHash := tx.Hash()
	return errors.New(fmt.Sprintf("Sending tx %x failed.", txHash[:8]))
}

func processTimeRes(p *peer, payload []byte) {
	time := int64(binary.BigEndian.Uint64(payload))
	//Concurrent writes need to be protected.
	//We use the same peer lock to prevent concurrent writes (on the network). It would be more efficient to use
	//different locks but the speedup is so marginal that it's not worth it.

	p.l.Lock()
	defer p.l.Unlock()
	p.time = time
}

func processNeighborRes(p *peer, payload []byte) {
	//Parse the incoming ipv4 addresses.
	ipportList := _processNeighborRes(payload)

	for _, ipportIter := range ipportList {
		//logger.Printf("IP/Port received: %v\n", ipportIter)
		//FileConnectionsLog.WriteString(fmt.Sprintf("IP/Port received: %v\n", ipportIter))
		//iplistChan is a buffered channel to handle ips asynchronously.
		iplistChan <- ipportIter
	}
}

func processValMappingRes(p *peer, payload []byte) {
	ValidatorShardMapReq <- payload
}

//Split the processNeighborRes function in two for cleaner testing.
func _processNeighborRes(payload []byte) (ipportList []string) {
	index := 0

	for cnt := 0; cnt < len(payload)/(IPV4ADDR_SIZE+PORT_SIZE); cnt++ {
		var addr string
		for singleAddr := index; singleAddr < index+IPV4ADDR_SIZE; singleAddr++ {
			tmp := int(payload[singleAddr])
			addr += strconv.Itoa(tmp)
			addr += "."
		}
		//Remove trailing dot.
		addr = addr[:len(addr)-1]
		addr += ":"
		//Extract port number.
		addr += strconv.Itoa(int(binary.BigEndian.Uint16(payload[index+4 : index+6])))

		ipportList = append(ipportList, addr)
		index += IPV4ADDR_SIZE + PORT_SIZE
	}

	return ipportList
}
