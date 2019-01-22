package storage

import (
	"fmt"
	"github.com/bazo-blockchain/bazo-miner/protocol"
	"math/big"
	"reflect"
	"strings"
	"testing"
)

func TestSerializeHashContent(t *testing.T) {
	var data []byte
	pubKeyInt, _ := new(big.Int).SetString(PubA1+PubA2, 16)
	copy(data, pubKeyInt.Bytes())

	hash := SerializeHashContent(data)

	if fmt.Sprintf("%x", hash) != "a7ffc6f8bf1ed76651c14756a061d662f580ff4de43b49fa82d80a4b80f8434a" {
		t.Errorf("Error serializing: %x != %v\n", hash, "a7ffc6f8bf1ed76651c14756a061d662f580ff4de43b49fa82d80a4b80f8434a")
	}
}

func TestGetAccount(t *testing.T) {
	accAHash := protocol.SerializeHashContent(accA.Address)

	acc, err := GetAccount(accAHash)

	if acc != accA && err == nil {
		t.Errorf("Error fetching account from state: %x\n", accAHash)
	}

	if acc == accB && err == nil {
		t.Errorf("Error fetching account from state: %x\n", accAHash)
	}

	var nilHash [32]byte
	acc, err = GetAccount(nilHash)

	if acc != nil || err.Error() != fmt.Sprintf("Acc (%x) not in the state.", nilHash[0:8]) {
		t.Errorf("Error fetching account from state: %x\n", nilHash)
	}
}

func TestGetRootAccount(t *testing.T) {
	rootHash := protocol.SerializeHashContent(rootAcc.Address)

	root, err := GetRootAccount(rootHash)

	if root == nil || err != nil {
		t.Errorf("Error fetching root account from state: %x\n", rootHash)
	}

	var nilHash [32]byte
	root, err = GetRootAccount(nilHash)

	if root != nil {
		t.Errorf("Error fetching account from state: %x\n", nilHash)
	}
}

func TestGetFundsTxPubKeys(t *testing.T) {

}

func TestExtractKeyFromFile(t *testing.T) {
	pubKey, privKey, _ := ExtractKeyFromFile(TestKeyFileName)
	expectedPubKey := GetPubKeyFromString(PubRoot1, PubRoot2)
	privInt, _ := new(big.Int).SetString(strings.Split(PrivRoot, "\n")[0], 16)

	if !reflect.DeepEqual(pubKey, expectedPubKey) {
		t.Errorf("Extracting public key from file %v failed: %v%v vs. %v%v\n", TestKeyFileName, pubKey.X, pubKey.Y, expectedPubKey.X, expectedPubKey.Y)
	}
	if !reflect.DeepEqual(privKey.D, privInt) {
		t.Errorf("Extracting private key from file %v failed: %v vs. %v\n", TestKeyFileName, privKey.D, privInt)
	}

}

func TestGetInitRootAddress(t *testing.T) {
	address, addressHash := GetInitRootAddress()
	pubKey := GetPubKeyFromString(INITROOTPUBKEY1, INITROOTPUBKEY2)
	expected := GetAddressFromPubKey(&pubKey)
	expectedHash := SerializeHashContent(expected)

	if !reflect.DeepEqual(address, expected) {
		t.Errorf("Getting address for init root failed: %x vs. %x\n", address, expected)
	}

	if !reflect.DeepEqual(addressHash, expectedHash) {
		t.Errorf("Getting addressHash for init root failed: %x vs. %x\n", addressHash, expected)
	}
}
