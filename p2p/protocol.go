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
	CONTRACTTX_REQ              = 11
	CONFIGTX_REQ           = 12
	STAKETX_REQ            = 13
	BLOCK_REQ              = 14
	BLOCK_HEADER_REQ       = 15
	ACC_REQ                = 16
	ROOTACC_REQ            = 17
	INTERMEDIATE_NODES_REQ = 18
	GENESIS_REQ			   = 19

	FUNDSTX_RES            = 20
	CONTRACTTX_RES              = 21
	CONFIGTX_RES           = 22
	STAKETX_RES            = 23
	BLOCK_RES              = 24
	BlOCK_HEADER_RES       = 25
	ACC_RES                = 26
	ROOTACC_RES            = 27
	INTERMEDIATE_NODES_RES = 28
	GENESIS_RES			   = 29

	NEIGHBOR_REQ = 30
	NEIGHBOR_RES = 40

	TIME_BRDCST = 50

	MINER_PING  = 100
	MINER_PONG  = 101
	CLIENT_PING = 102
	CLIENT_PONG = 103

	//Used to signal error
	NOT_FOUND = 110

	STATE_REQ = 120
	STATE_RES = 121

	FIRST_EPOCH_BLOCK_REQ = 122
	FIRST_EPOCH_BLOCK_RES = 123
	EPOCH_BLOCK_REQ = 124
	EPOCH_BLOCK_RES = 125
	VALIDATOR_SHARD_BRDCST = 126
	VALIDATOR_SHARD_REQ = 127
	VALIDATOR_SHARD_RES = 128
	EPOCH_BLOCK_BRDCST        = 129
	LAST_EPOCH_BLOCK_REQ = 130
	LAST_EPOCH_BLOCK_RES = 131
	TX_PAYLOAD_BRDCST = 132
	STATE_TRANSITION_BRDCST = 133
	BLOCK_REQ_SHARD = 134
	BLOCK_RES_SHARD = 135
	STATE_TRANSITION_REQ = 136
	STATE_TRANSITION_RES = 137
)

type Header struct {
	Len    uint32
	TypeID uint8
}

func (header Header) String() string {
	return fmt.Sprintf(
		"Length: %v\n"+
			"TypeID: %v\n",
		header.Len,
		header.TypeID,
	)
}
