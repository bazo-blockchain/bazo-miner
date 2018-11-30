package protocol

import (
	"crypto/ecdsa"
	"reflect"
	"testing"
)

func TestContractTxCreation(t *testing.T) {
	header := byte(0)
	fee := uint64(1)
	tx, newKey, _ := ConstrContractTx(header, fee, RootPrivKey, nil, nil)

	if !reflect.DeepEqual(tx.Header, header) {
		t.Errorf("Header does not match the given one: %x vs. %x\n", tx.Header, header)
	}

	if !reflect.DeepEqual(tx.Header, header) {
		t.Errorf("Fee does not match the given one: %x vs. %x\n", tx.Fee, fee)
	}

	if !reflect.DeepEqual(tx.PubKey, getAddressFromPubKey(&newKey.PublicKey)) {
		t.Errorf("Public key does not match the given one: %x vs. %x\n", tx.PubKey, accA.Address)
	}

	if !reflect.DeepEqual(tx.Issuer, getAddressFromPubKey(&RootPrivKey.PublicKey)) {
		t.Errorf("Issuer does not match the given root key: %x vs. %x\n", tx.Issuer, RootPrivKey)
	}

	if reflect.DeepEqual(tx.Sig, [64]byte{}) {
		t.Errorf("Signature should not be empty.")
	}
}

func TestContractTxHash(t *testing.T) {
	header := byte(0)
	fee := uint64(1)
	tx, _, _ := ConstrContractTx(header, fee, RootPrivKey, nil, nil)

	hash1 := tx.Hash()

	if !reflect.DeepEqual(hash1, tx.Hash()) {
		t.Errorf("ContractTx hashing failed!")
	}

	header = byte(1)
	fee = uint64(2)
	tx, _, _ = ConstrContractTx(header, fee, RootPrivKey, nil, nil)

	hash2 := tx.Hash()

	if !reflect.DeepEqual(hash2, tx.Hash()) {
		t.Errorf("ContractTx hashing failed!")
	}
}

func TestContractTxSerialization(t *testing.T) {
	header := byte(0)
	fee := uint64(1)
	tx, _, _ := ConstrContractTx(header, fee, RootPrivKey, nil, nil)

	var decodedTx *ContractTx
	encodedTx := tx.Encode()
	decodedTx = decodedTx.Decode(encodedTx)

	if !reflect.DeepEqual(tx, decodedTx) {
		t.Errorf("ContractTx serialization failed: %v vs. %v\n", tx, decodedTx)
	}

	header = byte(1)
	fee = uint64(2)
	tx, _, _ = ConstrContractTx(header, fee, RootPrivKey, nil, nil)

	encodedTx = tx.Encode()
	decodedTx = decodedTx.Decode(encodedTx)

	if !reflect.DeepEqual(tx, decodedTx) {
		t.Errorf("ContractTx serialization failed: %v vs. %v\n", tx, decodedTx)
	}
}

func getAddressFromPubKey(pubKey *ecdsa.PublicKey) (address [64]byte) {
	copy(address[:32], pubKey.X.Bytes())
	copy(address[32:], pubKey.Y.Bytes())

	return address
}