package miner

const (
	//How many blocks can we verify dynamically (e.g., proper time check) until we're too far behind
	//that this dynamic check is not possible anymore
	DELAYED_BLOCKS = 10

	//After requesting a tx/block, timeout after this amount of seconds
	TXFETCH_TIMEOUT    = 5
	BLOCKFETCH_TIMEOUT = 40

	INITIALINITROOTBALANCE = 1000

	//Some prominent programming languages (e.g., Java) have not unsigned integer types
	//Neglecting MSB simplifies compatibility
	MAX_MONEY = 9223372036854775807 //(2^63)-1

	//Default Block params
	BLOCKHASH_SIZE       = 32      //Byte
	FEE_MINIMUM          = 1       //Coins
	BLOCK_SIZE           = 5000000 //Byte
	//TODO @simibac How can I assure that only every min. a block will be mined?
	DIFF_INTERVAL        = 2016
	BLOCK_INTERVAL       = 60  //Sec
	BLOCK_REWARD         = 0   //Coins
	STAKING_MINIMUM      = 5   //Coins
	WAITING_MINIMUM      = 0   //Blocks
	ACCEPTED_TIME_DIFF   = 60  //Sec
	SLASHING_WINDOW_SIZE = 100 //Blocks
	SLASH_REWARD         = 2   //Coins
	NUM_INCL_PREV_SEEDS  = 5   //Number of previous seeds included in the PoS condition
)
