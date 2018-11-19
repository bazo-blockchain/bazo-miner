package storage

import (
	"errors"
	"fmt"
	"github.com/bazo-blockchain/bazo-miner/protocol"
	"log"
	"os"
)

func InitLogger() *log.Logger {
	return log.New(os.Stdout, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile)
}

//Needed by miner and p2p package
func GetAccount(pubKey [64]byte) (acc *protocol.Account, err error) {
	if acc = State[pubKey]; acc != nil {
		return acc, nil
	} else {
		return nil, errors.New(fmt.Sprintf("Acc (%x) not in the state.", pubKey[0:8]))
	}
}

func GetRootAccount(pubKey [64]byte) (acc *protocol.Account, err error) {
	if IsRootKey(pubKey) {
		acc, err = GetAccount(pubKey)
		return acc, err
	}

	return nil, err
}

func IsRootKey(pubKey [64]byte) bool {
	_, exists := RootKeys[pubKey]
	return exists
}

//Get all pubKeys involved in AccTx, FundsTx of a given block
func GetTxPubKeys(block *protocol.Block) (txPubKeys [][64]byte) {
	txPubKeys = GetAccTxPubKeys(block.AccTxData)
	txPubKeys = append(txPubKeys, GetFundsTxPubKeys(block.FundsTxData)...)

	return txPubKeys
}

//Get all pubKey involved in AccTx
func GetAccTxPubKeys(accTxData [][32]byte) (accTxPubKeys [][64]byte) {
	for _, txHash := range accTxData {
		var tx protocol.Transaction
		var accTx *protocol.AccTx

		tx = ReadClosedTx(txHash)
		if tx == nil {
			tx = ReadOpenTx(txHash)
		}

		accTx = tx.(*protocol.AccTx)
		accTxPubKeys = append(accTxPubKeys, accTx.Issuer)
		accTxPubKeys = append(accTxPubKeys, accTx.PubKey)
	}

	return accTxPubKeys
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
