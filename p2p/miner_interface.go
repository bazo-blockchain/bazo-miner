package p2p

import (
	"github.com/bazo-blockchain/bazo-miner/protocol"
)

var (
	//Block from the network, to the miner
	BlockIn = make(chan []byte)
	//Block from the miner, to the network
	BlockOut = make(chan []byte)

	//State transition from the miner to the network
	StateTransitionOut = make(chan []byte)

	//State transition from the network to the miner
	StateTransitionIn = make(chan []byte)

	//EpochBlock from the network, to the miner
	EpochBlockIn = make(chan []byte)
	//EpochBlock from the miner, to the network
	EpochBlockOut = make(chan []byte)

	//BlockHeader from the miner, to the clients
	BlockHeaderOut = make(chan []byte)

	VerifiedTxsOut = make(chan []byte)

	//Data requested by miner, to allow parallelism, we have a chan for every tx type.
	FundsTxChan  = make(chan *protocol.FundsTx)
	ContractTxChan    = make(chan *protocol.ContractTx)
	ConfigTxChan = make(chan *protocol.ConfigTx)
	StakeTxChan  = make(chan *protocol.StakeTx)

	BlockReqChan 	= make(chan []byte)
	StateTransitionShardReqChan 	= make(chan []byte)
	StateTransitionShardOut 		= make(chan []byte)

	GenesisReqChan 	= make(chan []byte)
	FirstEpochBlockReqChan 	= make(chan []byte)
	EpochBlockReqChan 	= make(chan []byte)
	LastEpochBlockReqChan 	= make(chan []byte)

	ValidatorShardMapReq 	= make(chan []byte)

	receivedTXStash = make([]*protocol.FundsTx, 0)
)

//This is for blocks and txs that the miner successfully validated.
func forwardBlockBrdcstToMiner() {
	for {
		block := <-BlockOut
		FileLogger.Printf("Building block broadcast packet\n")
		toBrdcst := BuildPacket(BLOCK_BRDCST, block)
		minerBrdcstMsg <- toBrdcst
	}
}

func forwardStateTransitionShardToMiner(){
	for {
		st := <- StateTransitionShardOut
		FileLogger.Printf("Building state transition request packet\n")
		toBrdcst := BuildPacket(STATE_TRANSITION_REQ, st)
		minerBrdcstMsg <- toBrdcst
	}
}

func forwardStateTransitionBrdcstToMiner()  {
	for {
		st := <-StateTransitionOut
		toBrdcst := BuildPacket(STATE_TRANSITION_BRDCST, st)
		minerBrdcstMsg <- toBrdcst
	}
}

func forwardEpochBlockBrdcstToMiner() {
	for {
		epochBlock := <-EpochBlockOut
		toBrdcst := BuildPacket(EPOCH_BLOCK_BRDCST, epochBlock)
		FileLogger.Printf("Build Epoch Block Brdcst Packet...\n")
		minerBrdcstMsg <- toBrdcst
	}
}

func forwardBlockHeaderBrdcstToMiner() {
	for {
		blockHeader := <- BlockHeaderOut
		clientBrdcstMsg <- BuildPacket(BLOCK_HEADER_BRDCST, blockHeader)
	}
}

func forwardVerifiedTxsToMiner() {
	for {
		verifiedTxs := <- VerifiedTxsOut
		clientBrdcstMsg <- BuildPacket(VERIFIEDTX_BRDCST, verifiedTxs)
	}
}

func forwardBlockToMiner(p *peer, payload []byte) {
	BlockIn <- payload
}

func txAlreadyInStash(slice []*protocol.FundsTx, newTXHash [32]byte) bool {
	for _, txInStash := range slice {
		if txInStash.Hash() == newTXHash {
			return true
		}
	}
	return false
}

//These are transactions the miner specifically requested.
func forwardTxReqToMiner(p *peer, payload []byte, txType uint8) {
	if payload == nil {
		return
	}

	switch txType {
	case FUNDSTX_RES:
		var fundsTx *protocol.FundsTx
		fundsTx = fundsTx.Decode(payload)
		if fundsTx == nil {
			return
		}
		/*My version: fundstx_res*/
		//FundsTxChan <- fundsTx

		// If TX is not received with the last 1000 Transaction, send it through the channel to the TX_FETCH.
		// Otherwise send nothing. This means, that the TX was sent before and we ensure, that only one TX per Broadcast
		// request is going through to the FETCH Request. This should prevent the "Received txHash did not correspond to
		// our request." error
		if !txAlreadyInStash(receivedTXStash, fundsTx.Hash()) {
			receivedTXStash = append(receivedTXStash, fundsTx)
			FundsTxChan <- fundsTx
			if len(receivedTXStash) > 1000 {
				receivedTXStash = append(receivedTXStash[:0], receivedTXStash[1:]...)
			}
		}
	case CONTRACTTX_RES:
		var contractTx *protocol.ContractTx
		contractTx = contractTx.Decode(payload)
		if contractTx == nil {
			return
		}
		ContractTxChan <- contractTx
	case CONFIGTX_RES:
		var configTx *protocol.ConfigTx
		configTx = configTx.Decode(payload)
		if configTx == nil {
			return
		}
		ConfigTxChan <- configTx
	case STAKETX_RES:
		var stakeTx *protocol.StakeTx
		stakeTx = stakeTx.Decode(payload)
		if stakeTx == nil {
			return
		}
		StakeTxChan <- stakeTx
	}
}

func forwardBlockReqToMiner(p *peer, payload []byte) {
	BlockReqChan <- payload
}

func forwardStateTransitionShardReqToMiner(p *peer, payload []byte) {
	FileLogger.Printf("received state transition response..\n")
	StateTransitionShardReqChan <- payload
}

func forwardGenesisReqToMiner(p *peer, payload []byte) {
	GenesisReqChan <- payload
}

func forwardFirstEpochBlockToMiner(p *peer, payload []byte) {
	FirstEpochBlockReqChan <- payload
}

func forwardEpochBlockToMiner(p *peer, payload []byte) {
	EpochBlockReqChan <- payload
}

func forwardEpochBlockToMinerIn(p *peer, payload []byte) {
	FileLogger.Printf("Writing Epoch block to channel EpochBlockIn.\n")
	EpochBlockIn <- payload
}

func forwardStateTransitionToMiner(p *peer, payload []byte) () {
	StateTransitionIn <- payload
}


func forwardLastEpochBlockToMiner(p *peer, payload []byte)  {
	LastEpochBlockReqChan <- payload
}

func ReadSystemTime() int64 {
	return systemTime
}
