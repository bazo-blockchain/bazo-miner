package storage

import (
	"bufio"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"errors"
	"fmt"
	"github.com/bazo-blockchain/bazo-miner/protocol"
	"log"
	"math/big"
	"os"
	"strings"
)

func InitLogger() *log.Logger {
	return log.New(os.Stdout, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile)
}



//Needed by miner and p2p package
func GetAccount(hash [32]byte) (acc *protocol.Account, err error) {
	if acc = State[hash]; acc != nil {
		return acc, nil
	} else {
		return nil, errors.New(fmt.Sprintf("Acc (%x) not in the state.", hash[0:8]))
	}
}

func GetRootAccount(hash [32]byte) (acc *protocol.Account, err error) {
	if IsRootKey(hash) {
		acc, err = GetAccount(hash)
		return acc, err
	}

	return nil, err
}

func GetInitRootPubKey() (address [64]byte, addressHash [32]byte) {
	pubKey, _ := GetPubKeyFromString(INIT_ROOT_PUB_KEY1, INIT_ROOT_PUB_KEY2)
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
		var tx protocol.Transaction
		var accTx *protocol.AccTx

		tx = ReadClosedTx(txHash)
		if tx == nil {
			tx = ReadOpenTx(txHash)
		}

		accTx = tx.(*protocol.AccTx)
		accTxPubKeys = append(accTxPubKeys, accTx.Issuer)
		accTxPubKeys = append(accTxPubKeys, protocol.SerializeHashContent(accTx.PubKey))
	}

	return accTxPubKeys
}

//Get all pubKey involved in FundsTx
func GetFundsTxPubKeys(fundsTxData [][32]byte) (fundsTxPubKeys [][32]byte) {
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

func ExtractECDSAKeyFromFile(filename string) (pubKey ecdsa.PublicKey, privKey ecdsa.PrivateKey, err error) {
	filehandle, err := os.Open(filename)
	if err != nil {
		return pubKey, privKey, errors.New(fmt.Sprintf("%v", err))
	}
	defer filehandle.Close()

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
		return pubKey, privKey, errors.New("the ecdsa key you provided is invalid and cannot sign hashes")
	}

	if !ecdsa.Verify(&pubKey, hashed, r, s) {
		return pubKey, privKey, errors.New("the ecdsa key you provided is invalid and cannot verify hashes")
	}

	return pubKey, privKey, nil
}

func ReadFile(filename string) (lines []string) {
	file, err := os.Open(filename)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}

	return lines
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
