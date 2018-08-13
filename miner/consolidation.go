package miner

import (
	"fmt"
	"errors"
	"github.com/bazo-blockchain/bazo-miner/protocol"
	"github.com/bazo-blockchain/bazo-miner/p2p"
	"github.com/bazo-blockchain/bazo-miner/storage"
	"github.com/bazo-blockchain/bazo-miner/conf"
)


func initialiseConsAccount(state map[[32]byte]*protocol.ConsolidatedAccount, pubKey [64]byte) {
	address := protocol.SerializeHashContent(pubKey)
	isStaking := false
	if _, exists := storage.State[address]; exists {
		isStaking = storage.State[address].IsStaking
	}

	// TODO: is the root account always staking?
	if _, isRoot := storage.RootKeys[address]; isRoot {
		isStaking = true
	}
	state[address] = &protocol.ConsolidatedAccount{address, pubKey, 0, 0, isStaking, 0}
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


func processAccTx(state map[[32]byte]*protocol.ConsolidatedAccount, block *protocol.Block) {
	for _, txHash := range block.AccTxData {
		tx := getTransaction(p2p.ACCTX_REQ, txHash)
		accTx := tx.(*protocol.AccTx)
		address := protocol.SerializeHashContent(accTx.PubKey)
		// Add accounts in the status
		if _, exists := state[address]; !exists {
			initialiseConsAccount(state, accTx.PubKey)
		} else {
			fmt.Println("This shouldn't happen")  // TODO: better error
		}

	}
}


func processFundsTx(state map[[32]byte]*protocol.ConsolidatedAccount, block *protocol.Block) {
	for _, txHash := range block.FundsTxData {
		tx := getTransaction(p2p.FUNDSTX_REQ, txHash)
		fundsTx := tx.(*protocol.FundsTx)
		source := fundsTx.From
		dest := fundsTx.To
		// Add accounts in the map if they don't exist
		if _, exists := state[source]; !exists {
			sourceadd := storage.GetAccount(source)
			initialiseConsAccount(state, sourceadd.Address)
		}
		if _, exists := state[dest]; !exists {
			destadd := storage.GetAccount(dest)
			initialiseConsAccount(state, destadd.Address)
		}
		state[source].Balance = state[source].Balance - fundsTx.Fee - fundsTx.Amount
		state[source].TxCnt += 1
		state[dest].Balance += fundsTx.Amount
		state[block.Beneficiary].Balance += fundsTx.Fee
	}
}

func processStakeTx(state map[[32]byte]*protocol.ConsolidatedAccount, block *protocol.Block) {
	for _, txHash := range block.StakeTxData {
		tx := getTransaction(p2p.STAKETX_REQ, txHash)
		stakeTx := tx.(*protocol.StakeTx)
		addr := stakeTx.Account
		if _, exists := state[addr]; !exists {
			acc := storage.GetAccount(addr)
			initialiseConsAccount(state, acc.Address)
		}
		state[addr].Staking = stakeTx.IsStaking
		fmt.Printf("state[addr].StakingBlockHeight = block.Height: %v = %v\n", state[addr].StakingBlockHeight, block.Height)

		state[addr].StakingBlockHeight = block.Height
		state[block.Beneficiary].Balance += stakeTx.Fee
	}
}

func processConfigTx(params *conf.Parameters, block *protocol.Block) {
	for _, txHash := range block.ConfigTxData {
		tx := getTransaction(p2p.CONFIGTX_REQ, txHash)
		configTx := tx.(*protocol.ConfigTx)
		ConfigTxChange(params, configTx)
	}
}

// Helper function to copy configuration parameters
func copyParameter(dest *conf.Parameters, origin *conf.Parameters) {
	dest.BlockHash = origin.BlockHash
	dest.Fee_minimum = origin.Fee_minimum
	dest.Block_size = origin.Block_size
	dest.Diff_interval = origin.Diff_interval
	dest.Block_interval = origin.Block_interval
	dest.Block_reward = origin.Block_reward
	dest.Staking_minimum = origin.Staking_minimum
	dest.Waiting_minimum = origin.Waiting_minimum
	dest.Accepted_time_diff = origin.Accepted_time_diff
	dest.Slashing_window_size = origin.Slashing_window_size
	dest.Slash_reward = origin.Slash_reward
	dest.Num_included_prev_seeds = origin.Num_included_prev_seeds
	dest.Consolidation_interval = origin.Consolidation_interval
}

func processConsolidationTx(params *conf.Parameters, state map[[32]byte]*protocol.ConsolidatedAccount, block *protocol.Block) {
	for _, txHash := range block.ConsolidationTxData {
		tx := getTransaction(p2p.CONSOLIDATIONTX_REQ, txHash)
		consolidationTx := tx.(*protocol.ConsolidationTx)
		// Copy parameters
		copyParameter(params, &consolidationTx.ActiveParameters)

		// Copy consolidated accounts
		for i := 0; i < len(consolidationTx.Accounts); i++ {
			acc := consolidationTx.Accounts[i]
			address := acc.Account
			if _, exists := state[address]; !exists {
				initialiseConsAccount(state, acc.Address)
			}
			state[address].Balance += acc.Balance
			state[address].TxCnt += acc.TxCnt
			state[address].Staking = acc.Staking
			fmt.Printf("state[address].StakingBlockHeight = acc.StakingBlockHeight: %v = %v\n", state[address].StakingBlockHeight, acc.StakingBlockHeight)
			state[address].StakingBlockHeight = acc.StakingBlockHeight
		}
	}
}

func GetConsolidationTx(lastHash [32]byte) (tx *protocol.ConsolidationTx, err error) {
	blockList := GetFullChainFromBlock(lastHash)
	return GetConsolidationTxFromChain(blockList)
}

// Create the list of blocks needed to create the consolidationTx
func GetFullChainFromBlock(lastHash [32]byte)(chain []*protocol.Block) {
	var blockList []*protocol.Block

	prevHash := lastHash
	// Go back till previous consolidationTx or genesis block
	for prevHash != [32]byte{} {
		// Previous block should be available since blocks between current and
		// previous consolidation are not deleted
		prevBlock := storage.ReadClosedBlock(prevHash)
		if prevBlock != nil {
			blockList = append(blockList, prevBlock)
			prevHash = prevBlock.PrevHash
			// Reached Genesis block (end of chain)
			if prevHash == [32]byte{} {
				prevBlock := storage.ReadClosedBlock(prevHash)
				if prevBlock == nil {
					fmt.Printf("Could not find Genesis Block: %x\n", prevHash)
				}
				blockList = append(blockList, prevBlock)
				break
			}
			// Stop at the first consolidation Tx
			if prevBlock.NrConsolidationTx > 0 {
				fmt.Printf("Found cons block: %x\n", prevBlock.Hash)
				break
			} else {
				continue
			}
		}

		//Fetch the block we apparently missed from the network
		err := p2p.BlockReq(prevHash)
		if err != nil {
			fmt.Println(err)
		}
		prevBlock = storage.ReadClosedBlock(prevHash)
		blockList = append(blockList, prevBlock)
		prevHash = prevBlock.PrevHash
	}
	return blockList
}

func GetConsolidationTxFromChain(chain []*protocol.Block) (tx *protocol.ConsolidationTx, err error) {
	// Create a snapshot of the current state
	state := make(protocol.StateAccounts)
	params := NewDefaultParameters()
	prevConsolidationBlock := findPreviousConsolidationBlock(chain)

	// Process all the previous consolidationTxs
	processConsolidationTx(&params, state, prevConsolidationBlock)

	// Process the other blocks/transactions
	for i := 0; i < len(chain); i++ {
		block := chain[i]

		// Skip empty blocks
		if block == nil || block.Hash == [32]byte{} {
			continue
		}
		if _, exists := state[block.Beneficiary]; !exists {
			add := storage.GetAccount(block.Beneficiary)
			initialiseConsAccount(state, add.Address)
		}

		processAccTx(state, block)
		processFundsTx(state, block)
		processStakeTx(state, block)
		processConfigTx(&params, block)

		state[block.Beneficiary].StakingBlockHeight += 1
		// Stop at the first consolidation block
		if block.Hash == prevConsolidationBlock.Hash {
			break
		}
	}
	fmt.Printf("ConsolidationState\n")
	for _, acc1 := range state {
		fmt.Printf("Address %x Balance %v\n", acc1.Address, acc1.Balance)
	}

	fmt.Printf("Previous consolidation hash selected: %x\n", prevConsolidationBlock.Hash)
	consTx, err := protocol.ConstrConsolidationTx(0x01, state, prevConsolidationBlock.Hash, params)
	if err != nil {
		errors.New(fmt.Sprintf("Error creating the ConstrConsolidationTx."))
	}
	return consTx, nil
}

func findPreviousConsolidationBlock(chain []*protocol.Block) (*protocol.Block) {
	for i := len(chain) - 1; i >= 0; i-- {
		block := chain[i]
		if block.NrConsolidationTx > 0 || block.Hash == [32]byte{} {
			fmt.Printf("Previous Consolidation Block: \n%v\n", block)
			return block
		}
	}
	return nil
}