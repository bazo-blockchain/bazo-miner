package p2p

import (
	"testing"
)

//Test the parsing of serialized ip addresses
func TestProcessNeighborRes(t *testing.T) {

	//Build some ip addresses
	payload := []byte{
		127, 0, 0, 1, 31, 64,
		23, 24, 122, 66, 31, 69,
		0, 0, 0, 0, 0, 0,
		255, 255, 255, 255, 156, 64,
	}

	ipportList := _processNeighborRes(payload)

	if ipportList[0] != "127.0.0.1:8000" {
		t.Errorf("Parsing IP address failed: %v\n", ipportList[0])
	}
	if ipportList[1] != "23.24.122.66:8005" {
		t.Errorf("Parsing IP address failed: %v\n", ipportList[1])
	}
	if ipportList[2] != "0.0.0.0:0" {
		t.Errorf("Parsing IP address failed: %v\n", ipportList[2])
	}
	if ipportList[3] != "255.255.255.255:40000" {
		t.Errorf("Parsing IP address failed: %v\n", ipportList[3])
	}
}
