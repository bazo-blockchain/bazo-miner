package storage

import (
	"github.com/bazo-blockchain/bazo-miner/protocol"
	"log"
	"os"
)

func InitLogger() *log.Logger {
	return log.New(os.Stdout, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile)
}

func IsRootKey(pubKey [64]byte) bool {
	_, exists := RootKeys[pubKey]
	return exists
}

//Get all pubKeys involved in ContractTx, FundsTx of a given block
func GetTxPubKeys(block *protocol.Block) (txPubKeys [][64]byte) {
	txPubKeys = GetContractTxPubKeys(block.ContractTxData)
	txPubKeys = append(txPubKeys, GetFundsTxPubKeys(block.FundsTxData)...)

	return txPubKeys
}

//Get all pubKey involved in ContractTx
func GetContractTxPubKeys(contractTxData [][32]byte) (contractTxPubKeys [][64]byte) {
	for _, txHash := range contractTxData {
		var tx protocol.Transaction
		var contractTx *protocol.ContractTx

		tx = ReadClosedTx(txHash)
		if tx == nil {
			tx = ReadOpenTx(txHash)
		}

		contractTx = tx.(*protocol.ContractTx)
		contractTxPubKeys = append(contractTxPubKeys, contractTx.Issuer)
		contractTxPubKeys = append(contractTxPubKeys, contractTx.PubKey)
	}

	return contractTxPubKeys
}

//Get all pubKey involved in FundsTx
func GetFundsTxPubKeys(fundsTxData [][32]byte) (fundsTxPubKeys [][64]byte) {
	for _, txHash := range fundsTxData {
		var tx protocol.Transaction
		var fundsTx *protocol.FundsTx

		tx = ReadClosedTx(txHash)
		if tx == nil {
			tx = ReadOpenTx(txHash)
		}

		fundsTx = tx.(*protocol.FundsTx)
		fundsTxPubKeys = append(fundsTxPubKeys, fundsTx.From)
		fundsTxPubKeys = append(fundsTxPubKeys, fundsTx.To)
	}

	return fundsTxPubKeys
}

func GetRelativeState(statePrev map[[64]byte]protocol.Account, stateNow map[[64]byte]*protocol.Account) (stateRel map[[64]byte]*protocol.RelativeAccount) {
	var stateRelative = make(map[[64]byte]*protocol.RelativeAccount)

	for know, _ := range stateNow {
		//In case account was newly created during block validation
		if _, ok := statePrev[know]; !ok {
			accNow := stateNow[know]
			accNewRel := protocol.NewRelativeAccount(know,[64]byte{},int64(accNow.Balance),accNow.IsStaking,accNow.CommitmentKey,accNow.Contract,accNow.ContractVariables)
			accNewRel.TxCnt = int32(accNow.TxCnt)
			accNewRel.StakingBlockHeight = int32(accNow.StakingBlockHeight)
			stateRelative[know] = &accNewRel
		} else {
			//Get account as in the version before block validation
			accPrev := statePrev[know]
			accNew := stateNow[know]

			//account with relative adjustments of the fields, will be  applied by the other shards
			accTransition := protocol.NewRelativeAccount(know,[64]byte{},int64(accNew.Balance-accPrev.Balance),accNew.IsStaking,accNew.CommitmentKey,accNew.Contract,accNew.ContractVariables)
			accTransition.TxCnt = int32(accNew.TxCnt - accPrev.TxCnt)
			accTransition.StakingBlockHeight = int32(accNew.StakingBlockHeight - accPrev.StakingBlockHeight)
			stateRelative[know] = &accTransition
		}
	}
	return stateRelative
}

func ApplyRelativeState(statePrev map[[64]byte]*protocol.Account, stateRel map[[64]byte]*protocol.RelativeAccount) (stateUpdated map[[64]byte]*protocol.Account) {
	for krel, _ := range stateRel {
		if _, ok := statePrev[krel]; !ok {
			accNewRel := stateRel[krel]
			accNew := protocol.NewAccount(krel,[64]byte{},uint64(accNewRel.Balance),accNewRel.IsStaking,accNewRel.CommitmentKey,accNewRel.Contract,accNewRel.ContractVariables)
			accNew.TxCnt = uint32(accNewRel.TxCnt)
			accNew.StakingBlockHeight = uint32(accNewRel.StakingBlockHeight)
			statePrev[krel] = &accNew
		} else {
			accPrev := statePrev[krel]
			accRel := stateRel[krel]

			//Adjust the account information
			accPrev.Balance = accPrev.Balance + uint64(accRel.Balance)
			accPrev.TxCnt = accPrev.TxCnt + uint32(accRel.TxCnt)
			accPrev.StakingBlockHeight = accPrev.StakingBlockHeight + uint32(accRel.StakingBlockHeight)
		}
	}
	return statePrev
}
