package p2p

import (
	"os"
	"testing"
)

var (
	MINER_IPPORT = "127.0.0.1:8000"
)

//Corresponds largely to server.go -> Init(...)
func TestMain(m *testing.M) {

	initLogging()

	//Used for some tests, the bootstarp server is listening at 8000 at the same time
	localConn = "127.0.0.1:9000"

	BlockIn = make(chan []byte)
	BlockOut = make(chan []byte)

	//channels for specific miner requests

	brdcstMsg = make(chan []byte)
	register = make(chan *peer)
	disconnect = make(chan *peer)

	iplistChan = make(chan string, MIN_MINERS)

	go broadcastService()
	go timeService()
	go checkHealthService()
	go receiveBlockFromMiner()

	//Bootstrap server
	go listener("127.0.0.1:8000")

	os.Exit(m.Run())
}
