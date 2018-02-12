package p2p

import (
	"net"
	"reflect"
	"testing"
)

func TestBuildPacket(t *testing.T) {

	payloadLen := 10
	packet := BuildPacket(BLOCK_BRDCST, []byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9})

	if packet[0] != 0x00 ||
		packet[1] != 0x00 ||
		packet[2] != 0x00 ||
		packet[3] != 0x0a ||
		packet[4] != BLOCK_BRDCST {
		t.Error("Header not correctly constructed\n")
	}

	for cnt := HEADER_LEN; cnt < HEADER_LEN+payloadLen; cnt++ {
		if packet[cnt] != byte(cnt-5) {
			t.Error("Payload not correctly constructed\n")
		}
	}
}

func TestExtractHeader(t *testing.T) {

	payloadLen := 10
	packet := BuildPacket(BLOCK_BRDCST, []byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9})

	header := extractHeader(packet)

	if header.Len != uint32(payloadLen) ||
		header.TypeID != BLOCK_BRDCST {
		t.Errorf("Header not correctly extracted: %v\n", header)
	}
}

func TestRcvData(t *testing.T) {

	payloadLen := 10
	//net.Pipe gives a very comfortable way of testing connections
	conn1, conn2 := net.Pipe()
	p1 := peer{conn: conn1}
	packet := BuildPacket(BLOCK_BRDCST, []byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9})
	go conn2.Write(packet)

	header, payload, err := rcvData(&p1)
	if err != nil ||
		header.Len != uint32(payloadLen) ||
		header.TypeID != BLOCK_BRDCST ||
		!reflect.DeepEqual([]byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}, payload) {
		t.Error("Receiving data routine failed\n")
	}
}
