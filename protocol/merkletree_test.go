package protocol

import (
	"golang.org/x/crypto/sha3"
	"math/rand"
	"testing"
	"time"
)

func TestBuildMerkleTree3N(t *testing.T) {

	var hashSlice [][32]byte

	for i := 0; i < 3; i++ {
		var txHash [32]byte
		rand.Read(txHash[:])
		hashSlice = append(hashSlice, txHash)
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

	m := b.BuildMerkleTree()
	if hash1233 != m.MerkleRoot() {
		t.Errorf("Hashes don't match: %x != %x\n", hash1233, m.MerkleRoot())
	}
}

func TestBuildMerkleTree2N(t *testing.T) {

	var hashSlice [][32]byte

	for i := 0; i < 2; i++ {
		var txHash [32]byte
		rand.Read(txHash[:])
		hashSlice = append(hashSlice, txHash)
	}

	b := Block{
		FundsTxData: hashSlice,
	}

	concat12 := append(hashSlice[0][:], hashSlice[1][:]...)
	hash12 := sha3.Sum256(concat12)

	m := b.BuildMerkleTree()
	if hash12 != m.MerkleRoot() {
		t.Errorf("Hashes don't match: %x != %x\n", hash12, m.MerkleRoot())
	}

}

func TestBuildMerkleTree4N(t *testing.T) {

	var hashSlice [][32]byte

	for i := 0; i < 4; i++ {
		var txHash [32]byte
		rand.Read(txHash[:])
		hashSlice = append(hashSlice, txHash)
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

	m := b.BuildMerkleTree()
	if hash1234 != m.MerkleRoot() {
		t.Errorf("Hashes don't match: %x != %x\n", hash1234, m.MerkleRoot())
	}

}

func TestBuildMerkleTree6N(t *testing.T) {

	var hashSlice [][32]byte

	for i := 0; i < 6; i++ {
		var txHash [32]byte
		rand.Read(txHash[:])
		hashSlice = append(hashSlice, txHash)
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

	m := b.BuildMerkleTree()
	if hash123456 != m.MerkleRoot() {
		t.Errorf("Hashes don't match: %x != %x\n", hash123456, m.MerkleRoot())
	}

}

func TestBuildMerkleTree8N(t *testing.T) {

	var hashSlice [][32]byte

	for i := 0; i < 8; i++ {
		var txHash [32]byte
		rand.Read(txHash[:])
		hashSlice = append(hashSlice, txHash)
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

	m := b.BuildMerkleTree()
	if hash12345678 != m.MerkleRoot() {
		t.Errorf("Hashes don't match: %x != %x\n", hash12345678, m.MerkleRoot())
	}

}

func TestBuildMerkleTree10N(t *testing.T) {

	var hashSlice [][32]byte

	for i := 0; i < 10; i++ {
		var txHash [32]byte
		rand.Read(txHash[:])
		hashSlice = append(hashSlice, txHash)
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

	m := b.BuildMerkleTree()
	if hash12345678910 != m.MerkleRoot() {
		t.Errorf("Hashes don't match: %x != %x\n", hash12345678910, m.MerkleRoot())
	}

}

func TestBuildMerkleTree11N(t *testing.T) {

	var hashSlice [][32]byte

	for i := 0; i < 11; i++ {
		var txHash [32]byte
		rand.Read(txHash[:])
		hashSlice = append(hashSlice, txHash)
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

	m := b.BuildMerkleTree()
	if hash123456789101111 != m.MerkleRoot() {
		t.Errorf("Hashes don't match: %x != %x\n", hash123456789101111, m.MerkleRoot())
	}

}

func TestGetIntermediate(t *testing.T) {

	var hashSlice [][32]byte

	for i := 0; i < 11; i++ {
		var txHash [32]byte
		rand.Read(txHash[:])
		hashSlice = append(hashSlice, txHash)
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

	merkleTree := b.BuildMerkleTree()

	intermediates, _ := GetIntermediate(merkleTree.GetLeaf(hashSlice[9]))

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



func TestMerkleProof(t *testing.T) {
	randVal := rand.New(rand.NewSource(time.Now().Unix()))

	var hash1, hash2, hash3, hash4 [32]byte

	randVal.Read(hash1[:])
	randVal.Read(hash2[:])
	randVal.Read(hash3[:])
	randVal.Read(hash4[:])

	var hashSlice [][32]byte
	hashSlice = append(hashSlice, hash1, hash2, hash3, hash4)

	b := Block{
		FundsTxData: hashSlice,
	}

	m := b.BuildMerkleTree()

	hash12 := MTHash(append(hash1[:], hash2[:]...))
	hash34 := MTHash(append(hash3[:], hash4[:]...))
	hash1234 := MTHash(append(hash12[:], hash34[:]...))
	if hash1234 != m.MerkleRoot() {
		t.Errorf("Root hashes don't match: %x != %x\n", hash1234, m.MerkleRoot())
	}

	hashes, err := m.MerkleProof(hash1)
	if err != nil {
		t.Error(err)
	}

	if len(hashes) != 2 {
		t.Errorf("Merkle proof returns invalid amount of hashes")
	}

	var mhash [32]byte
	var leftOrRight [1]byte
	copy(leftOrRight[:], hashes[0][0:1])
	copy(mhash[:], hashes[0][1:33])

	if leftOrRight != [1]byte{'r'} {
		t.Errorf("invalid left/right byte: is left but should be right")
	}

	if hash2 != mhash {
		t.Errorf("invalid Merkle proof hash at index 0")
	}

	copy(leftOrRight[:], hashes[1][0:1])
	copy(mhash[:], hashes[1][1:33])

	if leftOrRight != [1]byte{'r'} {
		t.Errorf("invalid left/right byte: is left but should be right")
	}

	if hash34 != mhash {
		t.Errorf("invalid Merkle proof hash at index 1")
	}
}



func TestMerkleProofWithVerification(t *testing.T) {
	randVal := rand.New(rand.NewSource(time.Now().Unix()))
	var hashSlice [][32]byte
	nofHashes := int(randVal.Uint32() % 1000) + 1
	for i := 0; i < nofHashes; i++ {
		var txHash [32]byte
		randVal.Read(txHash[:])
		hashSlice = append(hashSlice, txHash)
	}

	b := Block{
		FundsTxData: hashSlice,
	}

	m := b.BuildMerkleTree()


	randomIndex := randVal.Uint32() % uint32(nofHashes)
	leafHash := hashSlice[randomIndex]
	hashes, err := m.MerkleProof(leafHash)
	if err != nil {
		t.Error(err)
	}

	if !m.VerifyMerkleProof(leafHash, hashes) {
		t.Errorf("Merkle proof verification failed for hash %x", leafHash)
	}
}

func TestVerifyMerkleProof(t *testing.T) {
	randVal := rand.New(rand.NewSource(time.Now().Unix()))

	var hash1, hash2, hash3, hash4 [32]byte

	randVal.Read(hash1[:])
	randVal.Read(hash2[:])
	randVal.Read(hash3[:])
	randVal.Read(hash4[:])

	var hashSlice [][32]byte
	hashSlice = append(hashSlice, hash1, hash2, hash3, hash4)

	b := Block{
		FundsTxData: hashSlice,
	}

	m := b.BuildMerkleTree()

	hash12 := MTHash(append(hash1[:], hash2[:]...))
	hash34 := MTHash(append(hash3[:], hash4[:]...))
	hash1234 := MTHash(append(hash12[:], hash34[:]...))
	if hash1234 != m.MerkleRoot() {
		t.Errorf("Root hashes don't match: %x != %x\n", hash1234, m.MerkleRoot())
	}

	hashes, err := m.MerkleProof(hash1)
	if err != nil {
		t.Error(err)
	}

	if !m.VerifyMerkleProof(hash1, hashes) {
		t.Errorf("Merkle proof verification failed for hash %x", hash1)
	}
}