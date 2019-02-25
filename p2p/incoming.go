package p2p

//All incoming messages are processed here and acted upon accordingly
func processIncomingMsg(p *peer, header *Header, payload []byte) {

	switch header.TypeID {
	//BROADCASTING
	case FUNDSTX_BRDCST:
		processTxBrdcst(p, payload, FUNDSTX_BRDCST)
	case ACCTX_BRDCST:
		processTxBrdcst(p, payload, ACCTX_BRDCST)
	case CONFIGTX_BRDCST:
		processTxBrdcst(p, payload, CONFIGTX_BRDCST)
	case STAKETX_BRDCST:
		processTxBrdcst(p, payload, STAKETX_BRDCST)
	case BLOCK_BRDCST:
		forwardBlockToMiner(p, payload)
	case EPOCH_BLOCK_BRDCST:
		FileLogger.Printf("Incoming Epoch block broadcast.\n")
		forwardEpochBlockToMinerIn(p, payload)
	case STATE_TRANSITION_BRDCST:
		forwardStateTransitionToMiner(p,payload)
	case TIME_BRDCST:
		processTimeRes(p, payload)

		//REQUESTS
	case FUNDSTX_REQ:
		txRes(p, payload, FUNDSTX_REQ)
	case CONTRACTTX_REQ:
		txRes(p, payload, CONTRACTTX_REQ)
	case CONFIGTX_REQ:
		txRes(p, payload, CONFIGTX_REQ)
	case STAKETX_REQ:
		txRes(p, payload, STAKETX_REQ)
	case BLOCK_REQ:
		blockRes(p, payload)
	case STATE_TRANSITION_REQ:
		FileLogger.Printf("Incoming state transition request.\n")
		stateTransitionRes(p,payload)
	case BLOCK_HEADER_REQ:
		blockHeaderRes(p, payload)
	case ACC_REQ:
		accRes(p, payload)
	case ROOTACC_REQ:
		rootAccRes(p, payload)
	case MINER_PING:
		pongRes(p, payload, MINER_PING)
	case CLIENT_PING:
		pongRes(p, payload, CLIENT_PING)
	case NEIGHBOR_REQ:
		neighborRes(p)
	case INTERMEDIATE_NODES_REQ:
		intermediateNodesRes(p, payload)
	case GENESIS_REQ:
		genesisRes(p, payload)
	case FIRST_EPOCH_BLOCK_REQ:
		FirstEpochBlockRes(p,payload)
	case EPOCH_BLOCK_REQ:
		EpochBlockRes(p,payload)
	case LAST_EPOCH_BLOCK_REQ:
		LastEpochBlockRes(p,payload)

		//RESPONSES
	case VALIDATOR_SHARD_RES:
		processValMappingRes(p, payload)
	case NEIGHBOR_RES:
		processNeighborRes(p, payload)
	case BLOCK_RES:
		FileLogger.Printf("Writing following data into channel BlockReqChan: - %v\n",payload)
		forwardBlockReqToMiner(p, payload)
	case STATE_TRANSITION_RES:
		FileLogger.Printf("Incoming state transition response.\n")
		forwardStateTransitionShardReqToMiner(p,payload)
	case FUNDSTX_RES:
		forwardTxReqToMiner(p, payload, FUNDSTX_RES)
	case CONTRACTTX_RES:
		forwardTxReqToMiner(p, payload, CONTRACTTX_RES)
	case CONFIGTX_RES:
		forwardTxReqToMiner(p, payload, CONFIGTX_RES)
	case STAKETX_RES:
		forwardTxReqToMiner(p, payload, STAKETX_RES)
	case GENESIS_RES:
		forwardGenesisReqToMiner(p, payload)
	case FIRST_EPOCH_BLOCK_RES:
		forwardFirstEpochBlockToMiner(p,payload)
	case EPOCH_BLOCK_RES:
		forwardEpochBlockToMiner(p,payload)
	case LAST_EPOCH_BLOCK_RES:
		forwardLastEpochBlockToMiner(p,payload)
	default:
		FileLogger.Printf("Incoming message with unrecognized header Type ID: %d - Payload Len: %d\n",header.TypeID,len(payload))
	}
}
