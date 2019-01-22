package protocol

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"github.com/willf/bloom"
)

const (
	HASH_LEN                = 32
	MIN_BLOCKSIZE           = 318
	BLOOM_FILTER_ERROR_RATE = 0.1
)

type Block struct {
	//Header
	Header       byte
	Hash         [32]byte
	PrevHash     [32]byte
	NrConfigTx   uint8
	NrElementsBF uint16
	BloomFilter  *bloom.BloomFilter
	Height       uint32
	Beneficiary  [32]byte

	//Body
	Nonce                 [8]byte
	Timestamp             int64
	MerkleRoot            [32]byte
	NrAccTx               uint16
	NrFundsTx             uint16
	NrStakeTx             uint16
	SlashedAddress        [32]byte
	Seed                  [32]byte
	HashedSeed            [32]byte
	ConflictingBlockHash1 [32]byte
	ConflictingBlockHash2 [32]byte
	StateCopy             map[[32]byte]*Account //won't be serialized, just keeping track of local state changes

	AccTxData    [][32]byte
	FundsTxData  [][32]byte
	ConfigTxData [][32]byte
	StakeTxData  [][32]byte
}

func NewBlock(prevHash [32]byte, seed [32]byte, hashedSeed [32]byte, height uint32) *Block {
	newBlock := Block{
		PrevHash:   prevHash,
		Seed:       seed,
		HashedSeed: hashedSeed,
		Height:     height,
	}

	newBlock.StateCopy = make(map[[32]byte]*Account)

	return &newBlock
}

func (block *Block) HashBlock() [32]byte {
	if block == nil {
		return [32]byte{}
	}

	blockHash := struct {
		prevHash              [32]byte
		timestamp             int64
		merkleRoot            [32]byte
		beneficiary           [32]byte
		hashedSeed            [32]byte
		seed                  [32]byte
		slashedAddress        [32]byte
		conflictingBlockHash1 [32]byte
		conflictingBlockHash2 [32]byte
	}{
		block.PrevHash,
		block.Timestamp,
		block.MerkleRoot,
		block.Beneficiary,
		block.HashedSeed,
		block.Seed,
		block.SlashedAddress,
		block.ConflictingBlockHash1,
		block.ConflictingBlockHash2,
	}

	return SerializeHashContent(blockHash)
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
		MIN_BLOCKSIZE +
			int(b.NrAccTx)*HASH_LEN +
			int(b.NrFundsTx)*HASH_LEN +
			int(b.NrConfigTx)*HASH_LEN +
			int(b.NrStakeTx)*HASH_LEN

	if b.BloomFilter != nil {
		encodedBF, _ := b.BloomFilter.GobEncode()
		size += len(encodedBF)
	}

	return uint64(size)
}

func (b *Block) Encode() []byte {
	if b == nil {
		return nil
	}

	encoded := Block{
		Header:                b.Header,
		Hash:                  b.Hash,
		PrevHash:              b.PrevHash,
		Nonce:                 b.Nonce,
		Timestamp:             b.Timestamp,
		MerkleRoot:            b.MerkleRoot,
		Beneficiary:           b.Beneficiary,
		NrAccTx:               b.NrAccTx,
		NrFundsTx:             b.NrFundsTx,
		NrConfigTx:            b.NrConfigTx,
		NrStakeTx:             b.NrStakeTx,
		NrElementsBF:          b.NrElementsBF,
		BloomFilter:           b.BloomFilter,
		SlashedAddress:        b.SlashedAddress,
		Seed:                  b.Seed,
		Height:                b.Height,
		HashedSeed:            b.HashedSeed,
		ConflictingBlockHash1: b.ConflictingBlockHash1,
		ConflictingBlockHash2: b.ConflictingBlockHash2,

		AccTxData:    b.AccTxData,
		FundsTxData:  b.FundsTxData,
		ConfigTxData: b.ConfigTxData,
		StakeTxData:  b.StakeTxData,
	}

	buffer := new(bytes.Buffer)
	gob.NewEncoder(buffer).Encode(encoded)
	return buffer.Bytes()
}

func (b *Block) EncodeHeader() []byte {
	if b == nil {
		return nil
	}

	encoded := Block{
		Header:       b.Header,
		Hash:         b.Hash,
		PrevHash:     b.PrevHash,
		NrConfigTx:   b.NrConfigTx,
		NrElementsBF: b.NrElementsBF,
		BloomFilter:  b.BloomFilter,
		Height:       b.Height,
		Beneficiary:  b.Beneficiary,
	}

	buffer := new(bytes.Buffer)
	gob.NewEncoder(buffer).Encode(encoded)
	return buffer.Bytes()
}

func (*Block) Decode(encoded []byte) (b *Block) {
	var decoded Block
	buffer := bytes.NewBuffer(encoded)
	decoder := gob.NewDecoder(buffer)
	decoder.Decode(&decoded)
	return &decoded
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
		"Slashed Address: %x\n"+
		"Conflicted Block Hash 1: %x\n"+
		"Conflicted Block Hash 2: %x\n",
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
