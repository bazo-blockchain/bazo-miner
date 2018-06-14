package p2p

import (
	"time"
)

//This is not accessed concurrently, one single goroutine. However, the "peers" are accessed concurrently, therefore the
//thread-safe implementation
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
		//Broadcasting all messages
		case msg := <-brdcstMsg:
			for p := range peers.peerConns {
				//Write to the channel, which the peerBroadcast(*peer) running in a seperate goroutine consumes right away
				p.ch <- msg
			}

		}
	}
}

//Belongs to the broadcast service
func peerBroadcast(p *peer) {
	for msg := range p.ch {
		sendData(p, msg)
	}
}

//Single goroutine that makes sure the system is well connected
func checkHealthService() {

	for {
		//Periodically check if we are well-connected
		if len(peers.peerConns) >= MIN_MINERS {
			time.Sleep(2 * HEALTH_CHECK_INTERVAL * time.Minute)
			continue
		} else {
			//This delay is needed to prevent sending neighbor requests like a maniac
			time.Sleep(HEALTH_CHECK_INTERVAL * time.Second)
		}

		//The only goto in the code (I promise), but best solution here IMHO
	RETRY:
		select {
		//iplistChan gets filled with every incoming neighborRes, they're consumed here
		case ipaddr := <-iplistChan:
			p, err := initiateNewMinerConnection(ipaddr)
			if err != nil {
				logger.Printf("%v\n", err)
			}
			if p == nil || err != nil {
				goto RETRY
			}
			go minerConn(p)
			break
		default:
			//In case we don't have any ip addresses in the channel left, make a request to the network
			neighborReq()
			break
		}
	}
}

//Calculates periodically system time from available sources and broadcasts the time to all connected peers
func timeService() {
	//Initialize system time
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
		brdcstMsg <- packet
	}
}
