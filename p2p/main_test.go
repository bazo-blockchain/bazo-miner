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
	//Used for some tests, the bootstarp server is listening at 8000 at the same time
	Ipport = "127.0.0.1:9000"
	InitLogging()

	peers.minerConns = make(map[*peer]bool)
	peers.clientConns = make(map[*peer]bool)

	BlockIn = make(chan []byte)
	BlockOut = make(chan []byte)

	iplistChan = make(chan string, MIN_MINERS)
	minerBrdcstMsg = make(chan []byte)
	clientBrdcstMsg = make(chan []byte)
	register = make(chan *peer)
	disconnect = make(chan *peer)

	go broadcastService()
	go checkHealthService()
	go timeService()
	go forwardBlockBrdcstToMiner()
	go peerService()

	//Bootstrap server
	go listener("127.0.0.1:8000")

	os.Exit(m.Run())
}
