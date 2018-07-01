package miner

import (
	"fmt"
	"errors"
	"github.com/bazo-blockchain/bazo-miner/protocol"
	"github.com/bazo-blockchain/bazo-miner/p2p"
	"github.com/bazo-blockchain/bazo-miner/storage"
)


func initialiseConsAccount(state map[[32]byte]*protocol.ConsolidatedAccount, account [32]byte) {
	state[account] = &protocol.ConsolidatedAccount{account, 0, false, }
}


func reqTx(txType uint8, txHash [32]byte) (tx protocol.Transaction) {
	if conn := p2p.Connect(storage.BOOTSTRAP_SERVER); conn != nil {
		packet := p2p.BuildPacket(txType, txHash[:])
		conn.Write(packet)

		header, payload, err := p2p.RcvData(conn)
		if err != nil {
			logger.Printf("Requesting tx failed.")
			return
		}

		switch header.TypeID {
		case p2p.ACCTX_RES:
			var accTx *protocol.AccTx
			accTx = accTx.Decode(payload)
			tx = accTx
		case p2p.CONFIGTX_RES:
			var configTx *protocol.ConfigTx
			configTx = configTx.Decode(payload)
			tx = configTx
		case p2p.FUNDSTX_RES:
			var fundsTx *protocol.FundsTx
			fundsTx = fundsTx.Decode(payload)
			tx = fundsTx
		case p2p.STAKETX_RES:
			var stakeTx *protocol.StakeTx
			stakeTx = stakeTx.Decode(payload)
			tx = stakeTx
		}

		conn.Close()
	}

	return tx
}

func processFundsTx(state map[[32]byte]*protocol.ConsolidatedAccount, block *protocol.Block) {
	for _, txHash := range block.FundsTxData {
		tx := reqTx(p2p.FUNDSTX_REQ, txHash)
		fundsTx := tx.(*protocol.FundsTx)
		source := fundsTx.From
		dest := fundsTx.To
		// Add accounts in the map if they don't exist
		if _, exists := state[source]; !exists {
			initialiseConsAccount(state, source)
		}
		if _, exists := state[dest]; !exists {
			initialiseConsAccount(state, dest)
		}
		state[source].Balance = state[source].Balance - fundsTx.Fee - fundsTx.Amount
		state[dest].Balance += fundsTx.Amount
		state[block.Beneficiary].Balance += fundsTx.Fee
	}
}

func processStakeTx(state map[[32]byte]*protocol.ConsolidatedAccount, block *protocol.Block) {
	for _, txHash := range block.StakeTxData {
		tx := reqTx(p2p.CONFIGTX_REQ, txHash)
		stakeTx := tx.(*protocol.StakeTx)
		state[stakeTx.Account].Staking = stakeTx.IsStaking
		state[block.Beneficiary].Balance += stakeTx.Fee
	}
}

func GetConsolidationTxFromChain(chain []*protocol.Block) (tx *protocol.ConsolidationTx, err error) {
	// Create a snapshot of the current state
	state := make(protocol.StateAccounts)

	// Process all the blocks in the chain
	for _, block := range chain {
		// Skip empty blocks
		if block == nil {
			continue
		}
		if _, exists := state[block.Beneficiary]; !exists {
			initialiseConsAccount(state, block.Beneficiary)
		}

		processFundsTx(state, block)
		processStakeTx(state, block)
		// process other TX types
	}

	// The consolidation tx creates a snapshot of the system till a certain
	// block of which we have to keep track of
	lastBlockHash := chain[len(chain)-1].Hash
	consTx, err := protocol.ConstrConsolidationTx(0, state, lastBlockHash)
	if err != nil {
		errors.New(fmt.Sprintf("Error creating the seed file."))
		}
	return consTx, nil
}