package p2p

import (
	"encoding/binary"
	"strconv"
	"testing"
)

//Test serialization of request/responses
func Test_NeighborRes(t *testing.T) {

	ipportList := []string{
		"127.0.0.1:8000",
		"127.0.0.1:8005",
		"127.0.0.1:40000",
	}

	payload := _neighborRes(ipportList)

	//Check for correct deserialization
	index := 0
	if payload[index] != 127 || payload[index+1] != 0 || payload[index+2] != 0 || payload[index+3] != 1 ||
		strconv.Itoa(int(binary.BigEndian.Uint16(payload[index+4:index+6]))) != "8000" {
		t.Error("IP/Port Deserialization failed.")
	}

	index += 6
	if payload[index] != 127 || payload[index+1] != 0 || payload[index+2] != 0 || payload[index+3] != 1 ||
		strconv.Itoa(int(binary.BigEndian.Uint16(payload[index+4:index+6]))) != "8005" {
		t.Error("IP/Port Deserialization failed.")
	}

	index += 6
	if payload[index] != 127 || payload[index+1] != 0 || payload[index+2] != 0 || payload[index+3] != 1 ||
		strconv.Itoa(int(binary.BigEndian.Uint16(payload[index+4:index+6]))) != "40000" {
		t.Error("IP/Port Deserialization failed.")
	}
}

func Test_PongRes(t *testing.T) {

	//This corresponds to the IP:Port 8.8.8.8:8000
	ipport := []byte{
		31, 64,
	}

	//The IP address from the sender is 9.9.9.9:8000
	ipportRet := _pongRes(ipport)

	//A remote miner has the opportunity to send an additional IP:Port if he wishes to receive connection on this tuple
	if ipportRet != "8000" {
		t.Errorf("Failed to extract IP:Port: (%v) vs. (%v)\n", "8000", ipportRet)
	}

	ipport = []byte{
		31, 64,
	}

	ipportRet = _pongRes(ipport)
	if ipportRet != "8000" {
		t.Errorf("Failed to extract IP:Port: (%v) vs. (%v)\n", "8000", ipportRet)
	}
}
