package p2p

import (
	"testing"
	"time"
)

//Testing handshake and initiating new miner connections with the broadcast server (both run locally)
func TestInitiateNewMinerConnection(t *testing.T) {

	//wait until connections are safely opened
	time.Sleep(time.Second)

	p, err := initiateNewMinerConnection(MINER_IPPORT)
	if err != nil {
		t.Errorf("Could not establish connection to the boostrap server\n")
	}

	//Check that self-connection is not allowed
	_, err = initiateNewMinerConnection("127.0.0.1:9000")
	if err == nil {
		t.Errorf("Self-connection was not prevented\n")
	}

	//We initiate a miner connection so we can test whether already established connections are recognized
	go peerConn(p)
	time.Sleep(time.Second)
	//Check that already established connections are recognized
	_, err = initiateNewMinerConnection("127.0.0.1:8000")
	if err == nil {
		t.Errorf("Connecting to already established connection was not prevented\n")
	}
}

func TestPrepareHandshake(t *testing.T) {

	packet, err := PrepareHandshake(MINER_PING, 9000)

	if err != nil ||
		packet[0] != 0x00 ||
		packet[1] != 0x00 ||
		packet[2] != 0x00 ||
		packet[3] != 0x02 || //payload size is 2 bytes, listener port
		packet[4] != 0x64 || //dec(0x64) == 100, MINER_PING
		packet[5] != 0x23 ||
		packet[6] != 0x28 {
		t.Errorf("Building MINER_PING packet failed")
	}
}
