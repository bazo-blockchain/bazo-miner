package miner

import (
	"fmt"
	"github.com/bazo-blockchain/bazo-miner/protocol"
	"github.com/bazo-blockchain/bazo-miner/storage"
)

func accStateChangeRollback(txSlice []*protocol.AccTx) error {
	for _, accTx := range txSlice {
		err := storage.DeleteAccount(accTx.PubKey)
		if err != nil {
			logger.Fatal(fmt.Sprintf("account state change rollback failed for account %x", accTx.PubKey[0:8]))
			return err
		}
	}
	return nil
}

func fundsStateChangeRollback(txSlice []*protocol.FundsTx) {
	//Rollback in reverse order than original state change
	for cnt := len(txSlice) - 1; cnt >= 0; cnt-- {
		tx := txSlice[cnt]

		accSender, _ := storage.GetAccount(tx.From)
		accReceiver, _ := storage.GetAccount(tx.To)

		accSender.TxCnt -= 1
		accSender.Balance += tx.Amount
		accReceiver.Balance -= tx.Amount

		//If new coins were issued, revert
		if rootAcc, _ := storage.GetRootAccount(tx.From); rootAcc != nil {
			rootAcc.Balance -= tx.Amount
			rootAcc.Balance -= tx.Fee
		}
	}
}

func configStateChangeRollback(txSlice []*protocol.ConfigTx, blockHash [32]byte) {
	if len(txSlice) == 0 {
		return
	}

	//Only rollback if the config changes lead to a parameterChange
	//there might be the case that the client is not running the latest version, it's still confirming
	//the transaction but does not understand the ID and thus is not changing the state
	if parameterSlice[len(parameterSlice)-1].BlockHash != blockHash {
		return
	}

	//remove the latest entry in the parameters slice$
	parameterSlice = parameterSlice[:len(parameterSlice)-1]
	activeParameters = &parameterSlice[len(parameterSlice)-1]
	logger.Printf("Config parameters rolled back. New configuration: %v", *activeParameters)
}

func stakeStateChangeRollback(txSlice []*protocol.StakeTx) {
	//Rollback in reverse order than original state change
	for cnt := len(txSlice) - 1; cnt >= 0; cnt-- {
		tx := txSlice[cnt]

		accSender, _ := storage.GetAccount(tx.Account)
		//Rolling back stakingBlockHeight not needed
		accSender.IsStaking = !accSender.IsStaking
	}
}

func collectTxFeesRollback(accTx []*protocol.AccTx, fundsTx []*protocol.FundsTx, configTx []*protocol.ConfigTx, stakeTx []*protocol.StakeTx, minerAddress [64]byte) {
	minerAcc, _ := storage.GetAccount(minerAddress)

	//Subtract fees from sender (check if that is allowed has already been done in the block validation)
	for _, tx := range accTx {
		//Money was created out of thin air, no need to write back
		minerAcc.Balance -= tx.Fee
	}

	for _, tx := range fundsTx {
		minerAcc.Balance -= tx.Fee

		senderAcc, _ := storage.GetAccount(tx.From)
		senderAcc.Balance += tx.Fee
	}

	for _, tx := range configTx {
		//Money was created out of thin air, no need to write back
		minerAcc.Balance -= tx.Fee
	}

	for _, tx := range stakeTx {
		minerAcc.Balance -= tx.Fee

		senderAcc, _ := storage.GetAccount(tx.Account)
		senderAcc.Balance += tx.Fee
	}
}

func collectBlockRewardRollback(reward uint64, minerAddress [64]byte) {
	minerAcc, _ := storage.GetAccount(minerAddress)
	minerAcc.Balance -= reward
}

func collectSlashRewardRollback(reward uint64, block *protocol.Block) {
	if block.SlashedAddress != [64]byte{} || block.ConflictingBlockHash1 != [32]byte{} || block.ConflictingBlockHash2 != [32]byte{} {
		minerAcc, _ := storage.GetAccount(block.Beneficiary)
		slashedAcc, _ := storage.GetAccount(block.SlashedAddress)

		minerAcc.Balance -= reward
		slashedAcc.Balance += activeParameters.Staking_minimum
		slashedAcc.IsStaking = true
	}
}
