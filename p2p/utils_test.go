package p2p

import (
	"github.com/bazo-blockchain/bazo-miner/protocol"
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
		t.Errorf("Header for Block Brdcst not correctly extracted: %v\n", header)
	}

	epochBlock := protocol.NewEpochBlock([][32]byte{},20)

	ebPacket := BuildPacket(EPOCH_BLOCK_BRDCST,epochBlock.Encode())
	headerEB := extractHeader(ebPacket)

	if headerEB.TypeID != EPOCH_BLOCK_BRDCST {
		t.Errorf("Header for Epoch Block Brdcst not correctly extracted: \n%v\n", headerEB)
	}
}

func TestRcvData(t *testing.T) {

	payloadLen := 10
	//net.Pipe gives a very comfortable way of testing connections
	conn1, conn2 := net.Pipe()
	p1 := peer{conn: conn1}
	packet := BuildPacket(BLOCK_BRDCST, []byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9})
	go conn2.Write(packet)

	header, payload, err := RcvData(&p1)
	if err != nil ||
		header.Len != uint32(payloadLen) ||
		header.TypeID != BLOCK_BRDCST ||
		!reflect.DeepEqual([]byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}, payload) {
		t.Error("Receiving BLOCK_BRDCST failed\n")
	}

	//epochBlock := protocol.NewEpochBlock([][32]byte{},10)
	//acc1 := protocol.NewAccount([64]byte{'0'},[64]byte{},4000,true,[256]byte{},nil,nil)
	//acc2 := protocol.NewAccount([64]byte{'1'},[64]byte{},4000,true,[256]byte{},nil,nil)
	//storage.State[[64]byte{'0'}] = &acc1
	//storage.State[[64]byte{'0'}] = &acc2
	//epochBlock.State = storage.State
	//
	//ebPacket := BuildPacket(EPOCH_BLOCK_BRDCST, epochBlock.Encode())
	//go conn2.Write(ebPacket)
	//
	//header, payload, err = RcvData(&p1)
	//if err != nil ||
	//	header.Len != uint32(len(ebPacket)) ||
	//	header.TypeID != EPOCH_BLOCK_BRDCST ||
	//	!reflect.DeepEqual(epochBlock.Encode(), payload) {
	//	t.Log(header.Len)
	//	t.Log(header.TypeID)
	//	t.Error("Receiving EPOCH_BLOCK_BRDCST failed\n")
	//}
}
