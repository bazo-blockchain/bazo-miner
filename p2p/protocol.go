package p2p

import "fmt"

const HEADER_LEN = 5

//Mapping constants, used to parse incoming messages
const (
	FUNDSTX_BRDCST      = 1
	ACCTX_BRDCST        = 2
	CONFIGTX_BRDCST     = 3
	STAKETX_BRDCST      = 4
	VERIFIEDTX_BRDCST   = 5
	BLOCK_BRDCST        = 6
	BLOCK_HEADER_BRDCST = 7
	TX_BRDCST_ACK       = 8

	FUNDSTX_REQ            = 10
	ACCTX_REQ              = 11
	CONFIGTX_REQ           = 12
	STAKETX_REQ            = 13
	BLOCK_REQ              = 14
	BLOCK_HEADER_REQ       = 15
	ACC_REQ                = 16
	ROOTACC_REQ            = 17
	INTERMEDIATE_NODES_REQ = 18

	FUNDSTX_RES            = 20
	ACCTX_RES              = 21
	CONFIGTX_RES           = 22
	STAKETX_RES            = 23
	BLOCK_RES              = 24
	BlOCK_HEADER_RES       = 25
	ACC_RES                = 26
	ROOTACC_RES            = 27
	INTERMEDIATE_NODES_RES = 28

	NEIGHBOR_REQ = 30
	NEIGHBOR_RES = 40

	TIME_BRDCST = 50

	MINER_PING  = 100
	MINER_PONG  = 101
	CLIENT_PING = 102
	CLIENT_PONG = 103

	//Used to signal error
	NOT_FOUND = 110
)

type Header struct {
	Len    uint32
	TypeID uint8
}

func GetPongID(pingID uint8) uint8 {
	switch pingID {
	case MINER_PING:
		return MINER_PONG
	case CLIENT_PING:
		return CLIENT_PONG
	default:
		return 0
	}
}

func (header Header) String() string {
	return fmt.Sprintf(
		"Length: %v\n"+
			"TypeID: %v\n",
		header.Len,
		header.TypeID,
	)
}
