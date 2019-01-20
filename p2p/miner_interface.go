package p2p

import (
	"github.com/bazo-blockchain/bazo-miner/protocol"
)

var (
	//Block from the network, to the miner
	BlockIn = make(chan []byte)
	//Block from the miner, to the network
	BlockOut = make(chan []byte)

	//TX hashes of validated TXs to the network
	TxPayloadOut = make(chan []byte)

	//TX hashes of validated TXs from the network
	TxPayloadIn = make(chan []byte)

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
	GenesisReqChan 	= make(chan []byte)
	FirstEpochBlockReqChan 	= make(chan []byte)
	EpochBlockReqChan 	= make(chan []byte)
	LastEpochBlockReqChan 	= make(chan []byte)

	ValidatorShardMapReq 	= make(chan []byte)

	ValidatorShardMapChanOut = make(chan []byte) // validator assignment out to the miners
	ValidatorShardMapChanIn = make(chan []byte) // validator assignment received from the bootstrapping node
)

//This is for blocks and txs that the miner successfully validated.
func forwardBlockBrdcstToMiner() {
	for {
		block := <-BlockOut
		toBrdcst := BuildPacket(BLOCK_BRDCST, block)
		minerBrdcstMsg <- toBrdcst
	}
}

func forwardTXPayloadBrdcstToMiner() {
	for {
		txPayload := <-TxPayloadOut
		toBrdcst := BuildPacket(TX_PAYLOAD_BRDCST, txPayload)
		minerBrdcstMsg <- toBrdcst
	}
}

func forwardEpochBlockBrdcstToMiner() {
	for {
		epochBlock := <-EpochBlockOut
		toBrdcst := BuildPacket(EPOCH_BLOCK_BRDCST, epochBlock)
		minerBrdcstMsg <- toBrdcst
	}
}

func forwardValidatorShardMappingToMiner() {
	for {
		mapping := <- ValidatorShardMapChanOut
		toBrdcst := BuildPacket(VALIDATOR_SHARD_BRDCST, mapping)
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

func forwardAssignmentToMiner(p *peer, payload []byte) {
	ValidatorShardMapChanIn <- payload
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
		FundsTxChan <- fundsTx
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
	EpochBlockIn <- payload
}

func forwardTxPayloadToMiner(p *peer, payload []byte) {
	TxPayloadIn <- payload
}

func forwardLastEpochBlockToMiner(p *peer, payload []byte)  {
	LastEpochBlockReqChan <- payload
}

func ReadSystemTime() int64 {
	return systemTime
}
