package p2p

import (
	"bytes"
	"encoding/binary"
	"github.com/bazo-blockchain/bazo-miner/protocol"
	"github.com/bazo-blockchain/bazo-miner/storage"
	"strconv"
	"strings"
)

//This file responds to incoming requests from miners in a synchronous fashion
func txRes(p *peer, payload []byte, txKind uint8) {
	var txHash [32]byte
	copy(txHash[:], payload[0:32])

	var tx protocol.Transaction
	//Check closed and open storage if the tx is available
	openTx := storage.ReadOpenTx(txHash)
	closedTx := storage.ReadClosedTx(txHash)

	if openTx != nil {
		tx = openTx
	} else if closedTx != nil {
		tx = closedTx
	}

	//In case it was not found, send a corresponding message back
	if tx == nil {
		packet := BuildPacket(NOT_FOUND, nil)
		sendData(p, packet)
		return
	}

	var packet []byte
	switch txKind {
	case FUNDSTX_REQ:
		packet = BuildPacket(FUNDSTX_RES, tx.Encode())
	case ACCTX_REQ:
		packet = BuildPacket(ACCTX_RES, tx.Encode())
	case CONFIGTX_REQ:
		packet = BuildPacket(CONFIGTX_RES, tx.Encode())
	case STAKETX_REQ:
		packet = BuildPacket(STAKETX_RES, tx.Encode())
	}

	sendData(p, packet)
}

//Here as well, checking open and closed block storage
func blockRes(p *peer, payload []byte) {
	var packet []byte
	var block *protocol.Block
	var blockHash [32]byte

	//If no specific block is requested, send latest
	if len(payload) > 0 {
		copy(blockHash[:], payload[:32])
		if block = storage.ReadClosedBlock(blockHash); block == nil {
			block = storage.ReadOpenBlock(blockHash)
		}
	} else {
		block = storage.ReadLastClosedBlock()
	}

	if block != nil {
		packet = BuildPacket(BLOCK_RES, block.Encode())
		logger.Printf("BLOCK_RES to miner %v with block (%x)", p.getIPPort(), block.Hash[0:8])
	} else {
		packet = BuildPacket(NOT_FOUND, nil)
		logger.Printf("BLOCK_RES to miner %v -> Block (%x) not found --> Not in my chain", p.getIPPort(), blockHash[0:8])
	}

	sendData(p, packet)
}

//Response the requested block SPV header
func blockHeaderRes(p *peer, payload []byte) {
	var encodedHeader, packet []byte

	//If no specific header is requested, send latest
	if len(payload) > 0 {
		var blockHash [32]byte
		copy(blockHash[:], payload[:32])
		if block := storage.ReadClosedBlock(blockHash); block != nil {
			block.InitBloomFilter(append(storage.GetTxPubKeys(block)))
			encodedHeader = block.EncodeHeader()
		}
	} else {
		if block := storage.ReadLastClosedBlock(); block != nil {
			block.InitBloomFilter(append(storage.GetTxPubKeys(block)))
			encodedHeader = block.EncodeHeader()
		}
	}

	if len(encodedHeader) > 0 {
		packet = BuildPacket(BlOCK_HEADER_RES, encodedHeader)
	} else {
		packet = BuildPacket(NOT_FOUND, nil)
	}

	sendData(p, packet)
}

//Responds to an account request from another miner
func accRes(p *peer, payload []byte) {
	var packet []byte
	var hash [32]byte
	copy(hash[:], payload[0:32])

	acc, _ := storage.GetAccount(hash)
	packet = BuildPacket(ACC_RES, acc.Encode())

	sendData(p, packet)
}

func rootAccRes(p *peer, payload []byte) {
	var packet []byte
	var hash [32]byte
	copy(hash[:], payload[0:32])

	acc, _ := storage.GetRootAccount(hash)
	packet = BuildPacket(ROOTACC_RES, acc.Encode())

	sendData(p, packet)
}

//Completes the handshake with another miner.
func pongRes(p *peer, payload []byte, peerType uint) {
	//Payload consists of a 2 bytes array (port number [big endian encoded]).
	port := _pongRes(payload)

	if port != "" {
		p.listenerPort = port
	} else {
		p.conn.Close()
		return
	}

	//Restrict amount of connected miners
	if peers.len(PEERTYPE_MINER) >= MAX_MINERS {
		return
	}

	//Complete handshake
	var packet []byte
	if peerType == MINER_PING {
		p.peerType = PEERTYPE_MINER
		packet = BuildPacket(MINER_PONG, nil)
	} else if peerType == CLIENT_PING {
		p.peerType = PEERTYPE_CLIENT
		packet = BuildPacket(CLIENT_PONG, nil)
	}

	go peerConn(p)

	sendData(p, packet)
}

//Decouple the function for testing.
func _pongRes(payload []byte) string {
	if len(payload) == PORT_SIZE {
		return strconv.Itoa(int(binary.BigEndian.Uint16(payload[0:PORT_SIZE])))
	} else {
		return ""
	}
}

func neighborRes(p *peer) {
	//only supporting ipv4 addresses for now, makes fixed-size structure easier
	//in the future following structure is possible:
	//1) nr of ipv4 addresses, 2) nr of ipv6 addresses, followed by list of both
	var packet []byte
	var ipportList []string
	peerList := peers.getAllPeers(PEERTYPE_MINER)

	for _, p := range peerList {
		ipportList = append(ipportList, p.getIPPort())
	}

	packet = BuildPacket(NEIGHBOR_RES, _neighborRes(ipportList))
	sendData(p, packet)
}

//Decouple functionality to facilitate testing
func _neighborRes(ipportList []string) (payload []byte) {

	payload = make([]byte, len(ipportList)*6) //6 = size of ipv4 address + port
	index := 0
	for _, ipportIter := range ipportList {
		ipport := strings.Split(ipportIter, ":")
		split := strings.Split(ipport[0], ".")

		//Serializing IP:Port addr tuples
		for ipv4addr := 0; ipv4addr < 4; ipv4addr++ {
			addrPart, err := strconv.Atoi(split[ipv4addr])
			if err != nil {
				return nil
			}
			payload[index] = byte(addrPart)
			index++
		}

		port, _ := strconv.ParseUint(ipport[1], 10, 16)

		//serialize port number
		var buf bytes.Buffer
		binary.Write(&buf, binary.BigEndian, port)
		payload[index] = buf.Bytes()[len(buf.Bytes())-2]
		index++
		payload[index] = buf.Bytes()[len(buf.Bytes())-1]
		index++
	}

	return payload
}

func intermediateNodesRes(p *peer, payload []byte) {
	var blockHash, txHash [32]byte
	var nodeHashes [][]byte
	var packet []byte

	copy(blockHash[:], payload[:32])
	copy(txHash[:], payload[32:64])

	merkleTree := protocol.BuildMerkleTree(storage.ReadClosedBlock(blockHash))

	if intermediates, _ := protocol.GetIntermediate(protocol.GetLeaf(merkleTree, txHash)); intermediates != nil {
		for _, node := range intermediates {
			nodeHashes = append(nodeHashes, node.Hash[:])
		}

		packet = BuildPacket(INTERMEDIATE_NODES_RES, protocol.Encode(nodeHashes, 32))
	} else {
		packet = BuildPacket(NOT_FOUND, nil)
	}

	sendData(p, packet)
}
