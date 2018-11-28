package p2p

import (
	"github.com/bazo-blockchain/bazo-miner/protocol"
)

var (
	//Block from the network, to the miner
	BlockIn chan []byte = make(chan []byte)
	//Block from the miner, to the network
	BlockOut       chan []byte = make(chan []byte)
	//BlockHeader from the miner, to the clients
	BlockHeaderOut chan []byte = make(chan []byte)

	VerifiedTxsOut chan []byte = make(chan []byte)

	//Data requested by miner, to allow parallelism, we have a chan for every tx type.
	FundsTxChan  = make(chan *protocol.FundsTx)
	ContractTxChan    = make(chan *protocol.ContractTx)
	ConfigTxChan = make(chan *protocol.ConfigTx)
	StakeTxChan  = make(chan *protocol.StakeTx)

	BlockReqChan 	= make(chan []byte)
	GenesisReqChan 	= make(chan []byte)
)

//This is for blocks and txs that the miner successfully validated.
func forwardBlockBrdcstToMiner() {
	for {
		block := <-BlockOut
		toBrdcst := BuildPacket(BLOCK_BRDCST, block)
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
	case ACCTX_RES:
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

func ReadSystemTime() int64 {
	return systemTime
}
