package p2p

import (
	"math/rand"
	"net"
	"strings"
	"sync"
)

//The reason we use an additional listener port is because the port the miner connected to this peer
//is not the same as the one it listens to for new connections. When we are queried for neighbors
//we send the IP address in p.conn.RemotAddr() with the listenerPort
type peer struct {
	conn         net.Conn
	ch           chan []byte
	l            sync.Mutex
	listenerPort string
	time         int64
}

//peerStruct is a thread-safe map that supports all necessary map operations needed by the server
type peersStruct struct {
	peerConns map[*peer]bool
	peerMutex sync.Mutex
}

func (p *peer) getIPPort() string {
	ip := strings.Split(p.conn.RemoteAddr().String(), ":")
	//Cut off original port
	port := p.listenerPort

	return ip[0] + ":" + port
}

func (peers peersStruct) add(p *peer) {
	peers.peerMutex.Lock()
	defer peers.peerMutex.Unlock()
	peers.peerConns[p] = true
}

func (peers peersStruct) delete(p *peer) {
	peers.peerMutex.Lock()
	defer peers.peerMutex.Unlock()
	delete(peers.peerConns, p)
}

func (peers peersStruct) len() int {
	//no locking needed
	return len(peers.peerConns)
}

func (peers peersStruct) getRandomPeer() (p *peer) {
	//Acquire list before locking, otherwise deadlock
	peerList := peers.getAllPeers()

	if len(peerList) == 0 {
		return nil
	} else {
		return peerList[int(rand.Uint32())%len(peerList)]
	}
}

func (peers peersStruct) getAllPeers() []*peer {
	peers.peerMutex.Lock()
	defer peers.peerMutex.Unlock()

	var peerList []*peer

	for p := range peers.peerConns {
		peerList = append(peerList, p)
	}

	return peerList
}

func (peers peersStruct) getPeerTimes() (peerTimes []int64) {
	peers.peerMutex.Lock()
	defer peers.peerMutex.Unlock()

	for p := range peers.peerConns {
		p.l.Lock()
		peerTimes = append(peerTimes, p.time)
		//Concurrent writes need to protected. We set the time to 0 again as an indicator that the value has been consumed
		p.time = 0
		p.l.Unlock()
	}

	return peerTimes
}
