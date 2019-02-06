package miner

const (
	//How many blocks can we verify dynamically (e.g. proper time check) until we are too far behind
	//that this dynamic check is not possible anymore?!
	DELAYED_BLOCKS = 10

	TXFETCH_TIMEOUT    		= 5  //Sec
	BLOCKFETCH_TIMEOUT 		= 40 //Sec
	GENESISFETCH_TIMEOUT 	= 40 //Sec
	EPOCHBLOCKFETCH_TIMEOUT 	= 20 //Sec

	//Some prominent programming languages (e.g., Java) have not unsigned integer types
	//Neglecting MSB simplifies compatibility
	MAX_MONEY = 9223372036854775807 //(2^63)-1

	//Default Block params
	//TODO @simibac How can I assure that only every min. a block will be mined?
	BLOCKHASH_SIZE       = 32      //Byte
	FEE_MINIMUM          = 0       //Coins
	BLOCK_SIZE           = 20000 //Byte
	DIFF_INTERVAL        = 15    //Blocks
	BLOCK_INTERVAL       = 15      //Sec
	BLOCK_REWARD         = 0       //Coins
	STAKING_MINIMUM      = 1000    //Coins
	WAITING_MINIMUM      = 0       //Blocks
	ACCEPTED_TIME_DIFF   = 60      //Sec
	SLASHING_WINDOW_SIZE = 100     //Blocks
	SLASH_REWARD         = 0       //Coins
	NUM_INCL_PREV_PROOFS = 5       //Number of previous proofs included in the PoS condition
	//Parameters for sharding concept
	EPOCH_LENGTH         = 9 //blocks
	VALIDATORS_PER_SHARD = 2 //validators
)
