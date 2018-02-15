package protocol

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"github.com/willf/bloom"
	"golang.org/x/crypto/sha3"
)

const (
	HASH_LEN                = 32
	MIN_BLOCKHEADER_SIZE    = 185
	BLOOM_FILTER_ERROR_RATE = 0.1
)

type Block struct {
	//Header
	Hash           [32]byte
	PrevHash       [32]byte
	Nonce          [8]byte
	Timestamp      int64
	MerkleRoot     [32]byte
	Beneficiary    [32]byte
	NrFundsTx      uint16
	NrAccTx        uint16
	NrConfigTx     uint8
	NrStakeTx      uint16
	SlashedAddress [32]byte
	NrElementsBF   uint16
	BloomFilter    *bloom.BloomFilter

	//Body
	StateCopy             map[[32]byte]*Account //won't be serialized, just keeping track of local state changes
	Seed                  [32]byte
	Height                uint32
	HashedSeed            [32]byte
	ConflictingBlockHash1 [32]byte
	ConflictingBlockHash2 [32]byte
	AccTxData             [][32]byte
	FundsTxData           [][32]byte
	ConfigTxData          [][32]byte
	StakeTxData           [][32]byte
}

//TODO Block hashing for PoS changes
//Just Hash() conflicts with struct field
func (b *Block) HashBlock() (hash [32]byte) {

	var buf bytes.Buffer

	blockToHash := struct {
		prevHash    [32]byte
		timestamp   int64
		merkleRoot  [32]byte
		beneficiary [32]byte
	}{
		b.PrevHash,
		b.Timestamp,
		b.MerkleRoot,
		b.Beneficiary,
	}

	binary.Write(&buf, binary.BigEndian, blockToHash)
	return sha3.Sum256(buf.Bytes())
}

func (b *Block) InitBloomFilter(txPubKeys [][32]byte) {
	b.NrElementsBF = uint16(len(txPubKeys))

	m, k := calculateBloomFilterParams(float64(len(txPubKeys)), BLOOM_FILTER_ERROR_RATE)
	filter := bloom.New(m, k)
	for _, txPubKey := range txPubKeys {
		filter.Add(txPubKey[:])
	}

	b.BloomFilter = filter
}

func (b *Block) GetSize() uint64 {
	size :=
		MIN_BLOCKHEADER_SIZE +
			int(b.NrAccTx)*HASH_LEN +
			int(b.NrFundsTx)*HASH_LEN +
			int(b.NrConfigTx)*HASH_LEN +
			int(b.NrStakeTx)*HASH_LEN + 128 + 4

	if b.BloomFilter != nil {
		encodedBF, _ := b.BloomFilter.GobEncode()
		size += len(encodedBF)
	}

	return uint64(size)
}

func (b *Block) Encode() (encodedBlock []byte) {
	if b == nil {
		return nil
	}

	//Making byte array of all non-byte data
	var timeStamp [8]byte
	var nrFundsTx, nrAccTx, nrStakeTx, nrElementsBF [2]byte
	var height [4]byte

	binary.BigEndian.PutUint64(timeStamp[:], uint64(b.Timestamp))
	binary.BigEndian.PutUint16(nrFundsTx[:], b.NrFundsTx)
	binary.BigEndian.PutUint16(nrAccTx[:], b.NrAccTx)
	binary.BigEndian.PutUint16(nrStakeTx[:], b.NrStakeTx)
	binary.BigEndian.PutUint16(nrElementsBF[:], b.NrElementsBF)
	binary.BigEndian.PutUint32(height[:], b.Height)

	//Allocate memory
	encodedBlock = make([]byte, b.GetSize())

	//Serialize header

	copy(encodedBlock[0:32], b.Hash[:])
	copy(encodedBlock[32:64], b.PrevHash[:])
	copy(encodedBlock[64:72], b.Nonce[:])
	copy(encodedBlock[72:80], timeStamp[:])
	copy(encodedBlock[80:112], b.MerkleRoot[:])
	copy(encodedBlock[112:144], b.Beneficiary[:])
	copy(encodedBlock[144:146], nrFundsTx[:])
	copy(encodedBlock[146:148], nrAccTx[:])
	encodedBlock[148] = byte(b.NrConfigTx)
	copy(encodedBlock[149:151], nrStakeTx[:])
	copy(encodedBlock[151:183], b.SlashedAddress[:])
	copy(encodedBlock[183:185], nrElementsBF[:])

	index := MIN_BLOCKHEADER_SIZE

	if b.BloomFilter != nil {
		//Encode BloomFilter
		encodedBF, _ := b.BloomFilter.GobEncode()

		encodedBFSize := len(encodedBF)

		//Serialize BloomFilter
		copy(encodedBlock[index:index+encodedBFSize], encodedBF)
		index += encodedBFSize
	}
	//Serialize body
	copy(encodedBlock[index:index+HASH_LEN], b.Seed[:])
	index += HASH_LEN
	copy(encodedBlock[index:index+4], height[:])
	index += 4
	copy(encodedBlock[index:index+HASH_LEN], b.HashedSeed[:])
	index += HASH_LEN
	copy(encodedBlock[index:index+HASH_LEN], b.ConflictingBlockHash1[:])
	index += HASH_LEN
	copy(encodedBlock[index:index+HASH_LEN], b.ConflictingBlockHash2[:])
	index += HASH_LEN

	//Serialize all tx hashes
	for _, txHash := range b.FundsTxData {
		copy(encodedBlock[index:index+HASH_LEN], txHash[:])
		index += HASH_LEN
	}

	for _, txHash := range b.AccTxData {
		copy(encodedBlock[index:index+HASH_LEN], txHash[:])
		index += HASH_LEN
	}

	for _, txHash := range b.ConfigTxData {
		copy(encodedBlock[index:index+HASH_LEN], txHash[:])
		index += HASH_LEN
	}

	for _, txHash := range b.StakeTxData {
		copy(encodedBlock[index:index+HASH_LEN], txHash[:])
		index += HASH_LEN
	}

	return encodedBlock
}

func (*Block) Decode(encodedBlock []byte) (b *Block) {

	b = new(Block)

	if len(encodedBlock) < MIN_BLOCKHEADER_SIZE {
		return nil
	}

	//Deserialize header

	timeStampTmp := binary.BigEndian.Uint64(encodedBlock[72:80])
	timeStamp := int64(timeStampTmp)

	copy(b.Hash[:], encodedBlock[0:32])
	copy(b.PrevHash[:], encodedBlock[32:64])
	copy(b.Nonce[:], encodedBlock[64:72])
	b.Timestamp = timeStamp
	copy(b.MerkleRoot[:], encodedBlock[80:112])
	copy(b.Beneficiary[:], encodedBlock[112:144])
	b.NrFundsTx = binary.BigEndian.Uint16(encodedBlock[144:146])
	b.NrAccTx = binary.BigEndian.Uint16(encodedBlock[146:148])
	b.NrConfigTx = uint8(encodedBlock[148])
	b.NrStakeTx = binary.BigEndian.Uint16(encodedBlock[149:151])
	copy(b.SlashedAddress[:], encodedBlock[151:183])
	b.NrElementsBF = binary.BigEndian.Uint16(encodedBlock[183:185])

	index := MIN_BLOCKHEADER_SIZE

	if b.NrElementsBF > 0 {
		m, k := calculateBloomFilterParams(float64(b.NrElementsBF), BLOOM_FILTER_ERROR_RATE)

		//Initialize BloomFilter
		b.BloomFilter = bloom.New(m, k)

		//Encode BloomFilter
		encodedBF, _ := b.BloomFilter.GobEncode()

		encodedBFSize := len(encodedBF)

		//Deserialize BloomFilter
		b.BloomFilter.GobDecode(encodedBlock[index : index+encodedBFSize])
		index += encodedBFSize
	}

	//Deserialize body

	copy(b.Seed[:], encodedBlock[index:index+HASH_LEN])
	index += HASH_LEN
	b.Height = binary.BigEndian.Uint32(encodedBlock[index : index+4])
	index += 4
	copy(b.HashedSeed[:], encodedBlock[index:index+HASH_LEN])
	index += HASH_LEN
	copy(b.ConflictingBlockHash1[:], encodedBlock[index:index+HASH_LEN])
	index += HASH_LEN
	copy(b.ConflictingBlockHash2[:], encodedBlock[index:index+HASH_LEN])
	index += HASH_LEN

	//Deserialize all tx hashes
	var hash [32]byte
	for cnt := 0; cnt < int(b.NrFundsTx); cnt++ {
		copy(hash[:], encodedBlock[index:index+HASH_LEN])
		b.FundsTxData = append(b.FundsTxData, hash)
		index += HASH_LEN
	}

	for cnt := 0; cnt < int(b.NrAccTx); cnt++ {
		copy(hash[:], encodedBlock[index:index+HASH_LEN])
		b.AccTxData = append(b.AccTxData, hash)
		index += HASH_LEN
	}

	for cnt := 0; cnt < int(b.NrConfigTx); cnt++ {
		copy(hash[:], encodedBlock[index:index+HASH_LEN])
		b.ConfigTxData = append(b.ConfigTxData, hash)
		index += HASH_LEN
	}

	for cnt := 0; cnt < int(b.NrStakeTx); cnt++ {
		copy(hash[:], encodedBlock[index:index+HASH_LEN])
		b.StakeTxData = append(b.StakeTxData, hash)
		index += HASH_LEN
	}

	return b
}

func (b Block) String() string {
	return fmt.Sprintf("\nHash: %x\n"+
		"Previous Hash: %x\n"+
		"Nonce: %x\n"+
		"Timestamp: %v\n"+
		"MerkleRoot: %x\n"+
		"Beneficiary: %x\n"+
		"Amount of fundsTx: %v\n"+
		"Amount of accTx: %v\n"+
		"Amount of configTx: %v\n"+
		"Amount of stakeTx: %v\n"+
		"Seed: %x\n"+
		"Height: %d\n"+
		"Hashed Seed: %x\n"+
		"Slashed Address:%x\n"+
		"Conflicted Block Hash 1:%x\n"+
		"Conflicted Block Hash 2:%x\n",
		b.Hash[0:8],
		b.PrevHash[0:8],
		b.Nonce,
		b.Timestamp,
		b.MerkleRoot[0:8],
		b.Beneficiary[0:8],
		b.NrFundsTx,
		b.NrAccTx,
		b.NrConfigTx,
		b.NrStakeTx,
		b.Seed[0:8],
		b.Height,
		b.HashedSeed[0:8],
		b.SlashedAddress[0:8],
		b.ConflictingBlockHash1[0:8],
		b.ConflictingBlockHash2[0:8],
	)
}
