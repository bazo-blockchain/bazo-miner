package protocol

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"golang.org/x/crypto/sha3"
	"testing"
)

func TestBuildMerkleTree3N(t *testing.T) {

	var hashSlice [][32]byte
	var tx *FundsTx

	//Generating a private key and prepare data
	privA, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)

	for i := 0; i < 3; i++ {
		tx, _ = ConstrFundsTx(0, 10, 1, uint32(i), [32]byte{'1'}, [32]byte{'2'}, privA)
		hashSlice = append(hashSlice, tx.Hash())
	}

	b := Block{
		FundsTxData: hashSlice,
	}

	concat12 := append(hashSlice[0][:], hashSlice[1][:]...)
	hash12 := sha3.Sum256(concat12)

	concat33 := append(hashSlice[2][:], hashSlice[2][:]...)
	hash33 := sha3.Sum256(concat33)

	concat1233 := append(hash12[:], hash33[:]...)
	hash1233 := sha3.Sum256(concat1233)

	m := BuildMerkleTree(&b)
	if hash1233 != m.MerkleRoot() {
		t.Errorf("Hashes don't match: %x != %x\n", hash1233, m.MerkleRoot())
	}
}

func TestBuildMerkleTree2N(t *testing.T) {

	var hashSlice [][32]byte
	var tx *FundsTx

	//Generating a private key and prepare data
	privA, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)

	for i := 0; i < 2; i++ {
		tx, _ = ConstrFundsTx(0, 10, 1, uint32(i), [32]byte{'1'}, [32]byte{'2'}, privA)
		hashSlice = append(hashSlice, tx.Hash())
	}

	b := Block{
		FundsTxData: hashSlice,
	}

	concat12 := append(hashSlice[0][:], hashSlice[1][:]...)
	hash12 := sha3.Sum256(concat12)

	m := BuildMerkleTree(&b)
	if hash12 != m.MerkleRoot() {
		t.Errorf("Hashes don't match: %x != %x\n", hash12, m.MerkleRoot())
	}

}

func TestBuildMerkleTree4N(t *testing.T) {

	var hashSlice [][32]byte
	var tx *FundsTx

	//Generating a private key and prepare data
	privA, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)

	for i := 0; i < 4; i++ {
		tx, _ = ConstrFundsTx(0, 10, 1, uint32(i), [32]byte{'1'}, [32]byte{'2'}, privA)
		hashSlice = append(hashSlice, tx.Hash())
	}

	b := Block{
		FundsTxData: hashSlice,
	}

	concat12 := append(hashSlice[0][:], hashSlice[1][:]...)
	hash12 := sha3.Sum256(concat12)

	concat34 := append(hashSlice[2][:], hashSlice[3][:]...)
	hash34 := sha3.Sum256(concat34)

	//

	concat1234 := append(hash12[:], hash34[:]...)
	hash1234 := sha3.Sum256(concat1234)

	m := BuildMerkleTree(&b)
	if hash1234 != m.MerkleRoot() {
		t.Errorf("Hashes don't match: %x != %x\n", hash1234, m.MerkleRoot())
	}

}

func TestBuildMerkleTree6N(t *testing.T) {

	var hashSlice [][32]byte
	var tx *FundsTx

	//Generating a private key and prepare data
	privA, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)

	for i := 0; i < 6; i++ {
		tx, _ = ConstrFundsTx(0, 10, 1, uint32(i), [32]byte{'1'}, [32]byte{'2'}, privA)
		hashSlice = append(hashSlice, tx.Hash())
	}

	b := Block{
		FundsTxData: hashSlice,
	}

	concat12 := append(hashSlice[0][:], hashSlice[1][:]...)
	hash12 := sha3.Sum256(concat12)

	concat34 := append(hashSlice[2][:], hashSlice[3][:]...)
	hash34 := sha3.Sum256(concat34)

	//

	concat1234 := append(hash12[:], hash34[:]...)
	hash1234 := sha3.Sum256(concat1234)

	concat56 := append(hashSlice[4][:], hashSlice[5][:]...)
	hash56 := sha3.Sum256(concat56)

	//

	concat123456 := append(hash1234[:], hash56[:]...)
	hash123456 := sha3.Sum256(concat123456)

	m := BuildMerkleTree(&b)
	if hash123456 != m.MerkleRoot() {
		t.Errorf("Hashes don't match: %x != %x\n", hash123456, m.MerkleRoot())
	}

}

func TestBuildMerkleTree8N(t *testing.T) {

	var hashSlice [][32]byte
	var tx *FundsTx

	//Generating a private key and prepare data
	privA, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)

	for i := 0; i < 8; i++ {
		tx, _ = ConstrFundsTx(0, 10, 1, uint32(i), [32]byte{'1'}, [32]byte{'2'}, privA)
		hashSlice = append(hashSlice, tx.Hash())
	}

	b := Block{
		FundsTxData: hashSlice,
	}

	concat12 := append(hashSlice[0][:], hashSlice[1][:]...)
	hash12 := sha3.Sum256(concat12)

	concat34 := append(hashSlice[2][:], hashSlice[3][:]...)
	hash34 := sha3.Sum256(concat34)

	concat56 := append(hashSlice[4][:], hashSlice[5][:]...)
	hash56 := sha3.Sum256(concat56)

	concat78 := append(hashSlice[6][:], hashSlice[7][:]...)
	hash78 := sha3.Sum256(concat78)

	//

	concat1234 := append(hash12[:], hash34[:]...)
	hash1234 := sha3.Sum256(concat1234)

	concat5678 := append(hash56[:], hash78[:]...)
	hash5678 := sha3.Sum256(concat5678)

	//

	concat12345678 := append(hash1234[:], hash5678[:]...)
	hash12345678 := sha3.Sum256(concat12345678)

	m := BuildMerkleTree(&b)
	if hash12345678 != m.MerkleRoot() {
		t.Errorf("Hashes don't match: %x != %x\n", hash12345678, m.MerkleRoot())
	}

}

func TestBuildMerkleTree10N(t *testing.T) {

	var hashSlice [][32]byte
	var tx *FundsTx

	//Generating a private key and prepare data
	privA, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)

	for i := 0; i < 10; i++ {
		tx, _ = ConstrFundsTx(0, 10, 1, uint32(i), [32]byte{'1'}, [32]byte{'2'}, privA)
		hashSlice = append(hashSlice, tx.Hash())
	}

	b := Block{
		FundsTxData: hashSlice,
	}

	concat12 := append(hashSlice[0][:], hashSlice[1][:]...)
	hash12 := sha3.Sum256(concat12)

	concat34 := append(hashSlice[2][:], hashSlice[3][:]...)
	hash34 := sha3.Sum256(concat34)

	concat56 := append(hashSlice[4][:], hashSlice[5][:]...)
	hash56 := sha3.Sum256(concat56)

	concat78 := append(hashSlice[6][:], hashSlice[7][:]...)
	hash78 := sha3.Sum256(concat78)

	//

	concat1234 := append(hash12[:], hash34[:]...)
	hash1234 := sha3.Sum256(concat1234)

	concat5678 := append(hash56[:], hash78[:]...)
	hash5678 := sha3.Sum256(concat5678)

	//

	concat12345678 := append(hash1234[:], hash5678[:]...)
	hash12345678 := sha3.Sum256(concat12345678)

	concat910 := append(hashSlice[8][:], hashSlice[9][:]...)
	hash910 := sha3.Sum256(concat910)

	//

	concat12345678910 := append(hash12345678[:], hash910[:]...)
	hash12345678910 := sha3.Sum256(concat12345678910)

	m := BuildMerkleTree(&b)
	if hash12345678910 != m.MerkleRoot() {
		t.Errorf("Hashes don't match: %x != %x\n", hash12345678910, m.MerkleRoot())
	}

}

func TestBuildMerkleTree11N(t *testing.T) {

	var hashSlice [][32]byte
	var tx *FundsTx

	//Generating a private key and prepare data
	privA, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)

	for i := 0; i < 11; i++ {
		tx, _ = ConstrFundsTx(0, 10, 1, uint32(i), [32]byte{'1'}, [32]byte{'2'}, privA)
		hashSlice = append(hashSlice, tx.Hash())
	}

	b := Block{
		FundsTxData: hashSlice,
	}

	concat12 := append(hashSlice[0][:], hashSlice[1][:]...)
	hash12 := sha3.Sum256(concat12)

	concat34 := append(hashSlice[2][:], hashSlice[3][:]...)
	hash34 := sha3.Sum256(concat34)

	concat56 := append(hashSlice[4][:], hashSlice[5][:]...)
	hash56 := sha3.Sum256(concat56)

	concat78 := append(hashSlice[6][:], hashSlice[7][:]...)
	hash78 := sha3.Sum256(concat78)

	//

	concat1234 := append(hash12[:], hash34[:]...)
	hash1234 := sha3.Sum256(concat1234)

	concat5678 := append(hash56[:], hash78[:]...)
	hash5678 := sha3.Sum256(concat5678)

	concat910 := append(hashSlice[8][:], hashSlice[9][:]...)
	hash910 := sha3.Sum256(concat910)

	concat1111 := append(hashSlice[10][:], hashSlice[10][:]...)
	hash1111 := sha3.Sum256(concat1111)

	//

	concat12345678 := append(hash1234[:], hash5678[:]...)
	hash12345678 := sha3.Sum256(concat12345678)

	concat9101111 := append(hash910[:], hash1111[:]...)
	hash9101111 := sha3.Sum256(concat9101111)

	//

	concat123456789101111 := append(hash12345678[:], hash9101111[:]...)
	hash123456789101111 := sha3.Sum256(concat123456789101111)

	m := BuildMerkleTree(&b)
	if hash123456789101111 != m.MerkleRoot() {
		t.Errorf("Hashes don't match: %x != %x\n", hash123456789101111, m.MerkleRoot())
	}

}

func TestGetIntermediate(t *testing.T) {

	var hashSlice [][32]byte
	var tx *FundsTx

	//Generating a private key and prepare data
	privA, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)

	for i := 0; i < 11; i++ {
		tx, _ = ConstrFundsTx(0, 10, 1, uint32(i), [32]byte{'1'}, [32]byte{'2'}, privA)
		hashSlice = append(hashSlice, tx.Hash())
	}

	b := Block{
		FundsTxData: hashSlice,
	}

	concat12 := append(hashSlice[0][:], hashSlice[1][:]...)
	hash12 := sha3.Sum256(concat12)

	concat34 := append(hashSlice[2][:], hashSlice[3][:]...)
	hash34 := sha3.Sum256(concat34)

	concat56 := append(hashSlice[4][:], hashSlice[5][:]...)
	hash56 := sha3.Sum256(concat56)

	concat78 := append(hashSlice[6][:], hashSlice[7][:]...)
	hash78 := sha3.Sum256(concat78)

	//

	concat1234 := append(hash12[:], hash34[:]...)
	hash1234 := sha3.Sum256(concat1234)

	concat5678 := append(hash56[:], hash78[:]...)
	hash5678 := sha3.Sum256(concat5678)

	concat910 := append(hashSlice[8][:], hashSlice[9][:]...)
	hash910 := sha3.Sum256(concat910)

	concat1111 := append(hashSlice[10][:], hashSlice[10][:]...)
	hash1111 := sha3.Sum256(concat1111)

	//

	concat9101111 := append(hash910[:], hash1111[:]...)
	hash9101111 := sha3.Sum256(concat9101111)

	//

	concat12345678 := append(hash1234[:], hash5678[:]...)
	hash12345678 := sha3.Sum256(concat12345678)

	merkleTree := BuildMerkleTree(&b)

	intermediates, _ := GetIntermediate(GetLeaf(merkleTree, hashSlice[9]))

	if intermediates[0].Hash != hashSlice[8] {
		t.Errorf("Hashes don't match: %x != %x\n", intermediates[0].Hash, hashSlice[8])
	}

	if intermediates[1].Hash != hash910 {
		t.Errorf("Hashes don't match: %x != %x\n", intermediates[1].Hash, hash910)
	}

	if intermediates[2].Hash != hash1111 {
		t.Errorf("Hashes don't match: %x != %x\n", intermediates[2].Hash, hash1111)
	}

	if intermediates[3].Hash != hash9101111 {
		t.Errorf("Hashes don't match: %x != %x\n", intermediates[3].Hash, hash9101111)
	}

	if intermediates[4].Hash != hash12345678 {
		t.Errorf("Hashes don't match: %x != %x\n", intermediates[4].Hash, hash12345678)
	}

}
