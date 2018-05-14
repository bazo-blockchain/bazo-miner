package storage

import (
	"bufio"
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/bazo-blockchain/bazo-miner/protocol"
	"golang.org/x/crypto/sha3"
	"log"
	"math/big"
	"os"
	"strings"
)

func InitLogger() *log.Logger {
	return log.New(os.Stdout, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile)
}

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

func GetInitRootPubKey() (address [64]byte, addressHash [32]byte) {
	pubKey, _ := GetPubKeyFromString(INITROOTPUBKEY1, INITROOTPUBKEY2)
	address = GetAddressFromPubKey(&pubKey)

	return address, protocol.SerializeHashContent(address)
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

func GetAllAccounts() map[[32]byte]*protocol.Account {
	return State
}

func GetState() (state string) {
	for _, acc := range State {
		state += fmt.Sprintf("Is root: %v, %v\n", IsRootKey(acc.Hash()), acc)
	}
	return state
}

func ExtractKeyFromFile(filename string) (pubKey ecdsa.PublicKey, privKey ecdsa.PrivateKey, err error) {
	filehandle, err := os.Open(filename)
	if err != nil {
		return pubKey, privKey, errors.New(fmt.Sprintf("%v", err))
	}

	reader := bufio.NewReader(filehandle)

	//Public Key
	pub1, err := reader.ReadString('\n')
	pub2, err := reader.ReadString('\n')
	//Private Key
	priv, err2 := reader.ReadString('\n')
	if err != nil || err2 != nil {
		return pubKey, privKey, errors.New(fmt.Sprintf("Could not read key from file: %v", err))
	}

	pubKey, err = GetPubKeyFromString(strings.Split(pub1, "\n")[0], strings.Split(pub2, "\n")[0])
	if err != nil {
		return pubKey, privKey, errors.New(fmt.Sprintf("%v", err))
	}

	//File consists of public & private key
	if err2 == nil {
		privInt, b := new(big.Int).SetString(strings.Split(priv, "\n")[0], 16)
		if !b {
			return pubKey, privKey, errors.New("Failed to convert the key strings to big.Int.")
		}

		privKey = ecdsa.PrivateKey{
			pubKey,
			privInt,
		}
	}

	//Make sure the key being used is a valid one, that can sign and verify hashes/transactions
	hashed := []byte("testing")
	r, s, err := ecdsa.Sign(rand.Reader, &privKey, hashed)
	if err != nil {
		return  pubKey, privKey, errors.New("the ecdsa key you provided is invalid and cannot sign hashes")
	}

	if !ecdsa.Verify(&pubKey, hashed, r, s) {
		return pubKey, privKey, errors.New("the ecdsa key you provided is invalid and cannot verify hashes")
	}

	return pubKey, privKey, nil
}

func GetAddressFromPubKey(pubKey *ecdsa.PublicKey) (address [64]byte) {
	copy(address[:32], pubKey.X.Bytes())
	copy(address[32:], pubKey.Y.Bytes())

	return address
}

func GetPubKeyFromString(pub1, pub2 string) (pubKey ecdsa.PublicKey, err error) {
	pub1Int, b := new(big.Int).SetString(pub1, 16)
	pub2Int, b := new(big.Int).SetString(pub2, 16)
	if !b {
		return pubKey, errors.New("Failed to convert the key strings to big.Int.")
	}

	pubKey = ecdsa.PublicKey{
		elliptic.P256(),
		pub1Int,
		pub2Int,
	}

	return pubKey, nil
}
