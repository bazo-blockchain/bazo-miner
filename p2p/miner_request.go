package p2p

import (
	"errors"
)

//Both block and tx requests are handled asymmetricaly, using channels as inter-communication
//All the request in this file are specifically initiated by the miner package
func BlockReq(hash [32]byte) error {

	p := peers.getRandomPeer(PEERTYPE_MINER)
	logger.Printf("BLOCKREQ to miner %v with IP-Port: %v", p, p.getIPPort())

	if p == nil {
		return errors.New("Couldn't get a connection, request not transmitted.")
	}

	packet := BuildPacket(BLOCK_REQ, hash[:])
	sendData(p, packet)
	return nil
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

//Request specific transaction
func TxReq(hash [32]byte, reqType uint8) error {

	p := peers.getRandomPeer(PEERTYPE_MINER)
	if p == nil {
		return errors.New("Couldn't get a connection, request not transmitted.")
	}

	packet := BuildPacket(reqType, hash[:])
	sendData(p, packet)
	return nil
}
