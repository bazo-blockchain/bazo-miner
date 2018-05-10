package p2p

import (
	"../storage"
)

var (
	logMapping map[uint8]string
)

func initLogger() {
	logger = storage.InitLogger()

	//Instead of logging just the integer, we log the corresponding semantic meaning, makes scrolling through
	//the log file more comfortable
	logMapping = make(map[uint8]string)

	logMapping[1] = "FUNDSTX_BRDCST"
	logMapping[2] = "ACCTX_BRDCST"
	logMapping[3] = "CONFIGTX_BRDCST"
	logMapping[4] = "BLOCK_BRDCST"
	logMapping[5] = "TX_BRDCST_ACK"

	logMapping[10] = "FUNDSTX_REQ"
	logMapping[11] = "ACCTX_REQ"
	logMapping[12] = "CONFIGTX_REQ"
	logMapping[13] = "BLOCK_REQ"
	logMapping[14] = "ACC_REQ"
	logMapping[15] = "BLOCK_HEADER_REQ"

	logMapping[20] = "FUNDSTX_RES"
	logMapping[21] = "ACCTX_RES"
	logMapping[22] = "CONFIGTX_RES"
	logMapping[23] = "BLOCK_RES"
	logMapping[24] = "ACC_RES"
	logMapping[24] = "BlOCK_HEADER_RES"

	logMapping[30] = "NEIGHBOR_REQ"

	logMapping[40] = "NEIGHBOR_RES"

	logMapping[50] = "TIME_BRDCST"

	logMapping[100] = "MINER_PING"
	logMapping[101] = "MINER_PONG"

	logMapping[110] = "NOT_FOUND"
}
