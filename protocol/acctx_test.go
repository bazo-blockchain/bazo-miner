package protocol

import (
	"crypto/ecdsa"
	"reflect"
	"testing"
)

func TestAccTxCreation(t *testing.T) {
	header := byte(0)
	fee := uint64(1)
	tx, newKey, _ := ConstrAccTx(header, fee, accA.Address, RootPrivKey, nil, nil)

	if !reflect.DeepEqual(tx.Header, header) {
		t.Errorf("Header does not match the given one: %x vs. %x\n", tx.Header, header)
	}

	if !reflect.DeepEqual(tx.Header, header) {
		t.Errorf("Fee does not match the given one: %x vs. %x\n", tx.Fee, fee)
	}

	if !reflect.DeepEqual(tx.PubKey, accA.Address) {
		t.Errorf("Public key does not match the given one: %x vs. %x\n", tx.PubKey, accA.Address)
	}

	if !reflect.DeepEqual(tx.Issuer, SerializeHashContent(getAddressFromPubKey(&RootPrivKey.PublicKey))) {
		t.Errorf("Issuer does not match the given root key: %x vs. %x\n", tx.Issuer, RootPrivKey)
	}

	var nilPointer *ecdsa.PrivateKey
	if !reflect.DeepEqual(newKey, nilPointer) {
		t.Errorf("New key pointer should be nil.")
	}

	if reflect.DeepEqual(tx.Sig, [64]byte{}) {
		t.Errorf("Signature should not be empty.")
	}

	header = byte(1)
	fee = uint64(2)
	tx, newKey, _ = ConstrAccTx(header, fee, [64]byte{}, RootPrivKey, nil, nil)

	if !reflect.DeepEqual(tx.Header, header) {
		t.Errorf("Header does not match the given one: %x vs. %x\n", tx.Header, header)
	}

	if !reflect.DeepEqual(tx.Header, header) {
		t.Errorf("Fee does not match the given one: %x vs. %x\n", tx.Fee, fee)
	}

	if reflect.DeepEqual(tx.PubKey, [64]byte{}) {
		t.Errorf("Public key should not be empty.")
	}

	if !reflect.DeepEqual(tx.Issuer, SerializeHashContent(getAddressFromPubKey(&RootPrivKey.PublicKey))) {
		t.Errorf("Issuer does not match the given root key: %x vs. %x\n", tx.Issuer, RootPrivKey)
	}

	if reflect.DeepEqual(newKey, nilPointer) {
		t.Errorf("New key pointer should not be nil.")
	}

	if reflect.DeepEqual(tx.Sig, [64]byte{}) {
		t.Errorf("Signature should not be empty.")
	}
}

func TestAccTxHash(t *testing.T) {
	header := byte(0)
	fee := uint64(1)
	tx, _, _ := ConstrAccTx(header, fee, accA.Address, RootPrivKey, nil, nil)

	hash1 := tx.Hash()

	if !reflect.DeepEqual(hash1, tx.Hash()) {
		t.Errorf("AccTx hashing failed!")
	}

	header = byte(1)
	fee = uint64(2)
	tx, _, _ = ConstrAccTx(header, fee, [64]byte{}, RootPrivKey, nil, nil)

	hash2 := tx.Hash()

	if !reflect.DeepEqual(hash2, tx.Hash()) {
		t.Errorf("AccTx hashing failed!")
	}
}

func TestAccTxSerialization(t *testing.T) {
	header := byte(0)
	fee := uint64(1)
	tx, _, _ := ConstrAccTx(header, fee, accA.Address, RootPrivKey, nil, nil)

	var decodedTx *AccTx
	encodedTx := tx.Encode()
	decodedTx = decodedTx.Decode(encodedTx)

	if !reflect.DeepEqual(tx, decodedTx) {
		t.Errorf("AccTx serialization failed: %v vs. %v\n", tx, decodedTx)
	}

	header = byte(1)
	fee = uint64(2)
	tx, _, _ = ConstrAccTx(header, fee, accA.Address, RootPrivKey, nil, nil)

	encodedTx = tx.Encode()
	decodedTx = decodedTx.Decode(encodedTx)

	if !reflect.DeepEqual(tx, decodedTx) {
		t.Errorf("AccTx serialization failed: %v vs. %v\n", tx, decodedTx)
	}
}

func getAddressFromPubKey(pubKey *ecdsa.PublicKey) (address [64]byte) {
	copy(address[:32], pubKey.X.Bytes())
	copy(address[32:], pubKey.Y.Bytes())

	return address
}