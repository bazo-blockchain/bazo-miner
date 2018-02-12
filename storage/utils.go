package storage

import (
	"github.com/mchetelat/bazo_miner/protocol"
	"math/big"
	"bufio"
	"strings"
	"fmt"
	"errors"
	"os"
	"bytes"
	"encoding/binary"
	"golang.org/x/crypto/sha3"
)

//Serializes the input in big endian and returns the sha3 hash function applied on ths input
func SerializeHashContent(data interface{}) (hash [32]byte) {
	var buf bytes.Buffer
	binary.Write(&buf, binary.BigEndian, data)

	return sha3.Sum256(buf.Bytes())
}

//Needed by miner and p2p package
func GetAccount(hash [32]byte) *protocol.Account {
	return State[hash]
}

func GetRootAccount(hash [32]byte) *protocol.Account {
	if IsRootKey(hash) {
		return GetAccount(hash)
	}

	return nil
}

func GetInitRootPubKey() (pubKey [64]byte, pubKeyHash [32]byte) {
	pub1, _ := new(big.Int).SetString(INITROOTKEY1, 16)
	pub2, _ := new(big.Int).SetString(INITROOTKEY2, 16)

	copy(pubKey[:32], pub1.Bytes())
	copy(pubKey[32:], pub2.Bytes())

	return pubKey, protocol.SerializeHashContent(pubKey)
}

func IsRootKey(hash [32]byte) bool {
	_, exists := RootKeys[hash]
	return exists
}

//Get all pubKeys involved in AccTx, FundsTx of a given block
func GetTxPubKeys(block *protocol.Block) (txPubKeys [][32]byte) {
	txPubKeys = GetAccTxPubKeys(block.AccTxData)
	txPubKeys = append(txPubKeys, GetFundsTxPubKeys(block.FundsTxData)...)

	return txPubKeys
}

//Get all pubKey involved in AccTx
func GetAccTxPubKeys(accTxData [][32]byte) (accTxPubKeys [][32]byte) {
	for _, txHash := range accTxData {
		var accTx *protocol.AccTx
		closedTx := ReadClosedTx(txHash)
		accTx = closedTx.(*protocol.AccTx)
		accTxPubKeys = append(accTxPubKeys, accTx.Issuer)
		accTxPubKeys = append(accTxPubKeys, protocol.SerializeHashContent(accTx.PubKey))
	}

	return accTxPubKeys
}

//Get all pubKey involved in FundsTx
func GetFundsTxPubKeys(fundsTxData [][32]byte) (fundsTxPubKeys [][32]byte) {
	for _, txHash := range fundsTxData {
		var fundsTx *protocol.FundsTx
		closedTx := ReadClosedTx(txHash)
		fundsTx = closedTx.(*protocol.FundsTx)
		fundsTxPubKeys = append(fundsTxPubKeys, fundsTx.From)
		fundsTxPubKeys = append(fundsTxPubKeys, fundsTx.To)
	}

	return fundsTxPubKeys
}

func GetMyAccount(filename string)(*protocol.Account, error){
	var (
		myAcc			*protocol.Account
		fromPubKey 		[64]byte
		err 			error
	)

	hashFromFile, err := os.Open(filename)
	if err != nil {
		return myAcc, err
	}

	reader := bufio.NewReader(hashFromFile)

	//We only need the public key
	pub1, err := reader.ReadString('\n')
	pub2, err2 := reader.ReadString('\n')
	if err != nil || err2 != nil {
		return myAcc, err
	}

	pub1Int, _ := new(big.Int).SetString(strings.Split(pub1, "\n")[0], 16)
	pub2Int, _ := new(big.Int).SetString(strings.Split(pub2, "\n")[0], 16)
	copy(fromPubKey[0:32], pub1Int.Bytes())
	copy(fromPubKey[32:64], pub2Int.Bytes())

	myAcc = GetAccount(protocol.SerializeHashContent(fromPubKey))

	if myAcc == nil{
		return myAcc, errors.New(fmt.Sprintf("Could not find such an account"))
	}


	return myAcc, err
}

func GetAllAccounts()map[[32]byte]*protocol.Account{
	return State
}

func GetState() (state string) {
	for _, acc := range State {
		state += fmt.Sprintf("Is root: %v, %v\n", IsRootKey(acc.Hash()), acc)
	}
	return state
}