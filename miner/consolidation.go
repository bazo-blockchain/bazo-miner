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
		case p2p.CONSOLIDATIONTX_RES:
			var consolidationTx *protocol.ConsolidationTx
			consolidationTx = consolidationTx.Decode(payload)
			tx = consolidationTx
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
		tx := reqTx(p2p.STAKETX_REQ, txHash)
		fmt.Println(tx)
		stakeTx := tx.(*protocol.StakeTx)
		if _, exists := state[stakeTx.Account]; !exists {
			initialiseConsAccount(state, stakeTx.Account)
		}
		state[stakeTx.Account].Staking = stakeTx.IsStaking
		state[block.Beneficiary].Balance += stakeTx.Fee
	}
}

func processConsolidationTx(state map[[32]byte]*protocol.ConsolidatedAccount, block *protocol.Block) {
	for _, txHash := range block.ConsolidationTxData {
		tx := reqTx(p2p.CONSOLIDATIONTX_REQ, txHash)
		fmt.Println(tx)
		consolidationTx := tx.(*protocol.ConsolidationTx)

		for i := 0; i < len(consolidationTx.Accounts); i++ {
			acc := consolidationTx.Accounts[i]
			address := acc.Account
			if _, exists := state[address]; !exists {
				initialiseConsAccount(state, address)
			}
			state[address].Balance = acc.Balance
			state[address].Staking = acc.Staking
		}
	}
}

func GetConsolidationTx(lastHash [32]byte) (tx *protocol.ConsolidationTx, err error) {
	blockList := GetFullChainFromBlock(lastHash)
	return GetConsolidationTxFromChain(blockList)
}


func GetFullChainFromBlock(lastHash [32]byte)(chain []*protocol.Block) {
	fmt.Println("Adding consolidationtx")
	// Create a snapshot of the current state
	var blockList []*protocol.Block
	// Go back X blocks
	prevHash := lastHash
	for prevHash != [32]byte{} {
		prevBlock := storage.ReadClosedBlock(prevHash)
		if prevBlock != nil {
			blockList = append(blockList, prevBlock)
			prevHash = prevBlock.PrevHash
			continue
		}

		//Fetch the block we apparently missed from the network
		err := p2p.BlockReq(prevHash)
		if err != nil {
			fmt.Println(err)
		}
		prevBlock = storage.ReadOpenBlock(prevHash)
		blockList = append(blockList, prevBlock)
		prevHash = prevBlock.PrevHash
	}
	fmt.Printf("len chain %d", len(blockList))
	return blockList
}
func GetConsolidationTxFromChain(chain []*protocol.Block) (tx *protocol.ConsolidationTx, err error) {
	// Create a snapshot of the current state
	state := make(protocol.StateAccounts)

	//Process all the blocks in the chain
	for i := len(chain) - 1; i >= 0; i-- {
		block := chain[i]
		// Skip empty blocks
		if block == nil {
			continue
		}
		if _, exists := state[block.Beneficiary]; !exists {
			initialiseConsAccount(state, block.Beneficiary)
		}

		processFundsTx(state, block)
		processStakeTx(state, block)
		processConsolidationTx(state, block)
		// process other TX types
	}

	// The consolidation tx creates a snapshot of the system till a certain
	// block of which we have to keep track of
	lastBlockHash := chain[0].Hash
	consTx, err := protocol.ConstrConsolidationTx(0x01, state, lastBlockHash)
	if err != nil {
		errors.New(fmt.Sprintf("Error creating the ConstrConsolidationTx."))
	}
	return consTx, nil
}