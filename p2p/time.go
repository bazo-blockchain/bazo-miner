package p2p

import (
	"encoding/binary"
	"sort"
	"time"
)

var (
	systemTime int64
)

//Get current local time
func getTime() []byte {

	var buf [8]byte
	time := time.Now().Unix()
	binary.BigEndian.PutUint64(buf[:], uint64(time))
	return buf[:]
}

func writeSystemTime() {
	peerTimes := peers.getMinerTimes()

	//Add our own time as well
	peerTimes = append(peerTimes, time.Now().Unix())

	var ipeerTimes []int
	//Remove all 0s and cast to int (needed to leverage sort.Ints)
	for _, time := range peerTimes {
		if time != 0 {
			ipeerTimes = append(ipeerTimes, int(time))
		}
	}

	//If we don't have at least MIN_PEERS_FOR_TIME different time values, we take our own system time for reference
	if len(ipeerTimes) < MIN_PEERS_FOR_TIME {
		systemTime = time.Now().Unix()
		return
	}

	systemTime = calcMedian(ipeerTimes)
}

//To protect against outliers, get the median
func calcMedian(ipeerTimes []int) (median int64) {

	sort.Ints(ipeerTimes)
	//odd number of entries
	if len(ipeerTimes)%2 == 1 {
		return int64(ipeerTimes[len(ipeerTimes)/2])
	} else {
		//even number of entries
		low := int64(ipeerTimes[len(ipeerTimes)/2])
		high := int64(ipeerTimes[len(ipeerTimes)/2+1])

		return (high + low) / 2
	}
}
