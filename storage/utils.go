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
	txPubKeys = append(txPubKeys, block.Beneficiary)

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
