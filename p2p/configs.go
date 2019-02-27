package p2p

//Package-wide constants and configuration parameters
const (
	//MIN_MINERS is a lower bound of connections. If there are less, the system actively requests miner peers
	//from neighbors and establishes connections to them
	MIN_MINERS = 2
	//MAX_MINERS is an upper bound of connections. Miner handshakes are rejected if the amount of connections
	//grows above this number. Client connections are always accepted
	MAX_MINERS = 20
	//In order to get a reasonable system time, there needs to be a minimal amount of times available from other peers
	MIN_PEERS_FOR_TIME = 5
	//Interval to check system health in seconds
	HEALTH_CHECK_INTERVAL = 8
	//Broadcast local time to the network in seconds
	TIME_BRDCST_INTERVAL = 60
	//Calculate system time every UPDATE_SYS_TIME seconds
	UPDATE_SYS_TIME = 90

	//Protocol constants
	IPV4ADDR_SIZE = 4
	PORT_SIZE     = 2
)
