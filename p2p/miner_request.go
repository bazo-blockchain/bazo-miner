package p2p

import (
	"errors"
	"strconv"
)

//Both block and tx requests are handled asymmetricaly, using channels as inter-communication
//All the request in this file are specifically initiated by the miner package
func BlockReq(hash [32]byte) error {
	// Block Request with a Broadcast request. This does rise the possibility of a valid answer.
	for p := range peers.minerConns {
		//Write to the channel, which the peerBroadcast(*peer) running in a seperate goroutine consumes right away.

		if p == nil {
			return errors.New("Couldn't get a connection, request not transmitted.")
		}

		FileLogger.Printf("Building block request packet for miner: %v\n",p.getIPPort())

		packet := BuildPacket(BLOCK_REQ, hash[:])
		sendData(p, packet)
	}

	return nil
}

func StateTransitionReqShard(shardID int,height int) {
	strShardID := strconv.Itoa(shardID)
	strHeight := strconv.Itoa(height)

	strRequest := ""
	strRequest += strShardID
	strRequest += ":"
	strRequest += strHeight

	FileLogger.Printf("strRequest: %s\n",strRequest)

	StateTransitionShardOut <- []byte(strRequest)
}

func LastBlockReq() error {

	p := peers.getRandomPeer(PEERTYPE_MINER)
	if p == nil {
		return errors.New("Couldn't get a connection, request not transmitted.")
	}

	packet := BuildPacket(BLOCK_REQ, nil)
	sendData(p, packet)
	return nil
}

func GenesisReq() error {
	p := peers.getRandomPeer(PEERTYPE_MINER)
	if p == nil {
		return errors.New("Couldn't get a connection, request not transmitted.")
	}

	packet := BuildPacket(GENESIS_REQ, nil)
	sendData(p, packet)
	return nil
}

func FirstEpochBlockReq() error {
	p := peers.getRandomPeer(PEERTYPE_MINER)
	if p == nil {
		return errors.New("Couldn't get a connection, request not transmitted.")
	}

	packet := BuildPacket(FIRST_EPOCH_BLOCK_REQ, nil)
	sendData(p, packet)
	return nil
}

func LastEpochBlockReq() error {
	p := peers.getRandomPeer(PEERTYPE_MINER)
	if p == nil {
		return errors.New("Couldn't get a connection, request not transmitted.")
	}

	packet := BuildPacket(LAST_EPOCH_BLOCK_REQ, nil)
	sendData(p, packet)
	return nil
}

func EpochBlockReq(hash [32]byte) error {
	p := peers.getRandomPeer(PEERTYPE_MINER)
	if p == nil {
		return errors.New("Couldn't get a connection, request not transmitted.")
	}

	packet := BuildPacket(EPOCH_BLOCK_REQ, hash[:])
	sendData(p, packet)
	return nil
}

func TxReq(hash [32]byte, reqType uint8) error {
	// Tx Request also as broadcast so that teh possibility of an answer is higher.
	for p := range peers.minerConns {
		//Write to the channel, which the peerBroadcast(*peer) running in a seperate goroutine consumes right away.

		if p == nil {
			return errors.New("Couldn't get a connection, request not transmitted.")
		}
		packet := BuildPacket(reqType, hash[:])
		sendData(p, packet)
	}

	return nil
}