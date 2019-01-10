package p2p

import (
	"time"
	"github.com/bazo-blockchain/bazo-miner/storage"
)

//This is not accessed concurrently, one single goroutine. However, the "peers" are accessed concurrently, therefore the
//Thread-safe implementation.
func peerService() {
	for {
		select {
		case p := <-register:
			peers.add(p)
		case p := <-disconnect:
			peers.delete(p)
			close(p.ch)
		}
	}
}

func broadcastService() {
	for {
		select {
		//Broadcasting all messages.
		case msg := <-minerBrdcstMsg:
			for p := range peers.minerConns {
				//Write to the channel, which the peerBroadcast(*peer) running in a seperate goroutine consumes right away.
				logger.Printf("BLOCK_MSG Block Broadcast to %v", p.getIPPort())
				p.ch <- msg
			}
		case msg := <-clientBrdcstMsg:
			for p := range peers.clientConns {
				p.ch <- msg
			}
		}
	}
}

//Belongs to the broadcast service.
func peerBroadcast(p *peer) {
	for msg := range p.ch {
		sendData(p, msg)
	}
}

//Single goroutine that makes sure the system is well connected.
func checkHealthService() {
	for {
		time.Sleep(HEALTH_CHECK_INTERVAL * time.Second)

		if Ipport != storage.Bootstrap_Server && !peers.contains(storage.Bootstrap_Server, PEERTYPE_MINER) {
			p, err := initiateNewMinerConnection(storage.Bootstrap_Server)
			if p == nil || err != nil {
				logger.Printf("%v\n", err)
			} else {
				go peerConn(p)
			}
		}

		//Periodically check if we are well-connected
		if len(peers.minerConns) >= MIN_MINERS {
			continue
		}

		//The only goto in the code (I promise), but best solution here IMHO.
	RETRY:
		select {
		//iplistChan gets filled with every incoming neighborRes, they're consumed here.
		case ipaddr := <-iplistChan:
			p, err := initiateNewMinerConnection(ipaddr)
			if err != nil {
				logger.Printf("%v\n", err)
			}
			if p == nil || err != nil {
				goto RETRY
			}
			go peerConn(p)
			break
		default:
			//In case we don't have any ip addresses in the channel left, make a request to the network.
			neighborReq()
			break
		}
	}
}

//Calculates periodically system time from available sources and broadcasts the time to all connected peers.
func timeService() {
	//Initialize system time.
	systemTime = time.Now().Unix()
	go func() {
		for {
			time.Sleep(UPDATE_SYS_TIME * time.Second)
			writeSystemTime()
		}
	}()

	for {
		time.Sleep(TIME_BRDCST_INTERVAL * time.Second)
		packet := BuildPacket(TIME_BRDCST, getTime())
		minerBrdcstMsg <- packet
	}
}
