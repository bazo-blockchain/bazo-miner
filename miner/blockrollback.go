package miner

import (
	"errors"
	"github.com/bazo-blockchain/bazo-miner/protocol"
	"github.com/bazo-blockchain/bazo-miner/storage"
)

//Already validated block but not part of the current longest chain.
//No need for an additional state mutex, because this function is called while the blockValidation mutex is actively held
func rollback(b *protocol.Block) error {
	accTxSlice, fundsTxSlice, configTxSlice, stakeTxSlice, err := preValidateRollback(b)
	if err != nil {
		return err
	}

	data := blockData{accTxSlice, fundsTxSlice, configTxSlice, stakeTxSlice, b}

	//Going back to pre-block system parameters before the state is rolled back
	configStateChangeRollback(data.configTxSlice, b.Hash)

	//TODO Does not throw error but crashes
	validateStateRollback(data)

	postValidateRollback(data)

	return nil
}

func preValidateRollback(b *protocol.Block) (accTxSlice []*protocol.AccTx, fundsTxSlice []*protocol.FundsTx, configTxSlice []*protocol.ConfigTx, stakeTxSlice []*protocol.StakeTx, err error) {
	//Fetch all transactions from closed storage
	for _, hash := range b.AccTxData {
		var accTx *protocol.AccTx
		tx := storage.ReadClosedTx(hash)
		if tx == nil {
			//This should never happen, because all validated transactions are in closed storage
			return nil, nil, nil, nil, errors.New("CRITICAL: Validated accTx was not in the confirmed tx storage")
		} else {
			accTx = tx.(*protocol.AccTx)
		}
		accTxSlice = append(accTxSlice, accTx)
	}

	for _, hash := range b.FundsTxData {
		var fundsTx *protocol.FundsTx
		tx := storage.ReadClosedTx(hash)
		if tx == nil {
			return nil, nil, nil, nil, errors.New("CRITICAL: Validated fundsTx was not in the confirmed tx storage")
		} else {
			fundsTx = tx.(*protocol.FundsTx)
		}
		fundsTxSlice = append(fundsTxSlice, fundsTx)
	}

	for _, hash := range b.ConfigTxData {
		var configTx *protocol.ConfigTx
		tx := storage.ReadClosedTx(hash)
		if tx == nil {
			return nil, nil, nil, nil, errors.New("CRITICAL: Validated configTx was not in the confirmed tx storage")
		} else {
			configTx = tx.(*protocol.ConfigTx)
		}
		configTxSlice = append(configTxSlice, configTx)
	}

	for _, hash := range b.StakeTxData {
		var stakeTx *protocol.StakeTx
		tx := storage.ReadClosedTx(hash)
		if tx == nil {
			return nil, nil, nil, nil, errors.New("CRITICAL: Validated stakeTx was not in the confirmed tx storage")
		} else {
			stakeTx = tx.(*protocol.StakeTx)
		}
		stakeTxSlice = append(stakeTxSlice, stakeTx)
	}

	return accTxSlice, fundsTxSlice, configTxSlice, stakeTxSlice, nil
}

func validateStateRollback(data blockData) {
	collectSlashRewardRollback(activeParameters.Slash_reward, data.block)
	collectBlockRewardRollback(activeParameters.Block_reward, data.block.Beneficiary)
	collectTxFeesRollback(data.accTxSlice, data.fundsTxSlice, data.configTxSlice, data.stakeTxSlice, data.block.Beneficiary)
	stakeStateChangeRollback(data.stakeTxSlice)
	fundsStateChangeRollback(data.fundsTxSlice)
	accStateChangeRollback(data.accTxSlice)
}

func postValidateRollback(data blockData) {
	//Put all validated txs into invalidated state
	for _, tx := range data.accTxSlice {
		storage.WriteOpenTx(tx)
		storage.DeleteClosedTx(tx)
	}

	for _, tx := range data.fundsTxSlice {
		storage.WriteOpenTx(tx)
		storage.DeleteClosedTx(tx)
	}

	for _, tx := range data.configTxSlice {
		storage.WriteOpenTx(tx)
		storage.DeleteClosedTx(tx)
	}

	for _, tx := range data.stakeTxSlice {
		storage.WriteOpenTx(tx)
		storage.DeleteClosedTx(tx)
	}

	collectStatisticsRollback(data.block)

	//For transactions we switch from closed to open. However, we do not write back blocks
	//to open storage, because in case of rollback the chain they belonged to is likely to starve
	storage.DeleteClosedBlock(data.block.Hash)

	//Save the previous block as the last closed block
	storage.DeleteAllLastClosedBlock()
	storage.WriteLastClosedBlock(storage.ReadClosedBlock(data.block.PrevHash))
}
