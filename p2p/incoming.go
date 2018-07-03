package p2p

import "fmt"

//All incoming messages are processed here and acted upon accordingly
func processIncomingMsg(p *Peer, header *Header, payload []byte) {

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
	case TIME_BRDCST:
		processTimeRes(p, payload)

		//Requests
	case FUNDSTX_REQ:
		txRes(p, payload, FUNDSTX_REQ)
	case ACCTX_REQ:
		txRes(p, payload, ACCTX_REQ)
	case CONFIGTX_REQ:
		txRes(p, payload, CONFIGTX_REQ)
	case STAKETX_REQ:
		txRes(p, payload, STAKETX_REQ)
	case BLOCK_REQ:
		//TODO Remove
		fmt.Printf("Receiving request block(hash): %x\n", payload)
		blockRes(p, payload)
	case ACC_REQ:
		accRes(p, payload)
	case ROOTACC_REQ:
		rootAccRes(p, payload)
	case BLOCK_HEADER_REQ:
		blockHeaderRes(p, payload)
	case MINER_PING:
		pongRes(p, payload, MINER_PING)
	case CLIENT_PING:
		pongRes(p, payload, CLIENT_PING)
	case NEIGHBOR_REQ:
		neighborRes(p)
	case INTERMEDIATE_NODES_REQ:
		intermediateNodesRes(p, payload)

	//Responses
	case NEIGHBOR_RES:
		processNeighborRes(p, payload)
	case BLOCK_RES:
		//TODO Remove
		fmt.Printf("Receiving BLOCK_RES\n")
		forwardBlockReqToMiner(p, payload)
	case FUNDSTX_RES:
		forwardTxReqToMiner(p, payload, FUNDSTX_RES)
	case ACCTX_RES:
		forwardTxReqToMiner(p, payload, ACCTX_RES)
	case CONFIGTX_RES:
		forwardTxReqToMiner(p, payload, CONFIGTX_RES)
	case STAKETX_RES:
		forwardTxReqToMiner(p, payload, STAKETX_RES)
	default:
		//Send default NOT_FOUND
		packet := BuildPacket(NOT_FOUND, nil)
		sendData(p, packet)
	}
}
