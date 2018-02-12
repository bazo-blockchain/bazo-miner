package p2p

//Requests that the p2p package issues requesting data from other miners.
//Tx and block requests are processed in the miner_interface.go file, because
//this involves inter-communication between the two packages
func neighborReq() {

	p := peers.getRandomPeer()
	if p == nil {
		logger.Print("Could not fetch a random peer.\n")
		return
	}

	packet := BuildPacket(NEIGHBOR_REQ, nil)
	sendData(p, packet)
}
