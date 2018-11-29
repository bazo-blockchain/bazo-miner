package protocol

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"github.com/bazo-blockchain/bazo-miner/crypto"
	"github.com/willf/bloom"
	"reflect"
)

const (
	HASH_LEN                = 32
	HEIGHT_LEN				= 4
	//All fixed sizes form the Block struct are 254
	MIN_BLOCKSIZE           = 254 + crypto.COMM_PROOF_LENGTH
	MIN_BLOCKHEADER_SIZE    = 104
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
	CommitmentProof       [crypto.COMM_PROOF_LENGTH]byte
	ConflictingBlockHash1 [32]byte
	ConflictingBlockHash2 [32]byte
	StateCopy             map[[32]byte]*Account //won't be serialized, just keeping track of local state changes

	AccTxData    [][32]byte
	FundsTxData  [][32]byte
	ConfigTxData [][32]byte
	StakeTxData  [][32]byte
}

func NewBlock(prevHash [32]byte, height uint32) *Block {
	newBlock := Block{
		PrevHash:   prevHash,
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
		commitmentProof       [crypto.COMM_PROOF_LENGTH]byte
		slashedAddress        [32]byte
		conflictingBlockHash1 [32]byte
		conflictingBlockHash2 [32]byte
	}{
		block.PrevHash,
		block.Timestamp,
		block.MerkleRoot,
		block.Beneficiary,
		block.CommitmentProof,
		block.SlashedAddress,
		block.ConflictingBlockHash1,
		block.ConflictingBlockHash2,
	}
	return SerializeHashContent(blockHash)
}

func (block *Block) InitBloomFilter(txPubKeys [][32]byte) {
	block.NrElementsBF = uint16(len(txPubKeys))

	m, k := calculateBloomFilterParams(float64(len(txPubKeys)), BLOOM_FILTER_ERROR_RATE)
	filter := bloom.New(m, k)
	for _, txPubKey := range txPubKeys {
		filter.Add(txPubKey[:])
	}

	block.BloomFilter = filter
}

func (block *Block) GetSize() uint64 {
	size :=
		MIN_BLOCKSIZE + int(block.GetTxDataSize())

	if block.BloomFilter != nil {
		encodedBF, _ := block.BloomFilter.GobEncode()
		size += len(encodedBF)
	}

	return uint64(size)
}

func (block *Block) GetHeaderSize() uint64 {
	size := int(reflect.TypeOf(block.Header).Size() +
		reflect.TypeOf(block.Hash).Size() +
		reflect.TypeOf(block.PrevHash).Size() +
		reflect.TypeOf(block.NrConfigTx).Size() +
		reflect.TypeOf(block.NrElementsBF).Size() +
		reflect.TypeOf(block.Height).Size() +
		reflect.TypeOf(block.Beneficiary).Size())

	size += int(block.GetBloomFilterSize())

	return uint64(size)
}

func (block *Block) GetBodySize() uint64 {
	size := int(reflect.TypeOf(block.Nonce).Size() +
		reflect.TypeOf(block.Timestamp).Size() +
		reflect.TypeOf(block.MerkleRoot).Size() +
		reflect.TypeOf(block.NrAccTx).Size() +
		reflect.TypeOf(block.NrFundsTx).Size() +
		reflect.TypeOf(block.NrStakeTx).Size() +
		reflect.TypeOf(block.SlashedAddress).Size() +
		reflect.TypeOf(block.CommitmentProof).Size() +
		reflect.TypeOf(block.ConflictingBlockHash1).Size() +
		reflect.TypeOf(block.ConflictingBlockHash2).Size()) +
		int(block.GetTxDataSize())

	size += int(block.GetBloomFilterSize())

	return uint64(size)
}

func (block *Block) GetTxDataSize() uint64 {
	size := int(block.NrAccTx)*HASH_LEN +
		int(block.NrFundsTx)*HASH_LEN +
		int(block.NrConfigTx)*HASH_LEN +
		int(block.NrStakeTx)*HASH_LEN

	return uint64(size)
}

func (block *Block) GetBloomFilterSize() uint64 {
	size := 0
	if block.BloomFilter != nil {
		encodedBF, _ := block.BloomFilter.GobEncode()
		size += len(encodedBF)
	}

	return uint64(size)
}

func (block *Block) Encode() []byte {
	if block == nil {
		return nil
	}

	encoded := Block{
		Header:                block.Header,
		Hash:                  block.Hash,
		PrevHash:              block.PrevHash,
		Nonce:                 block.Nonce,
		Timestamp:             block.Timestamp,
		MerkleRoot:            block.MerkleRoot,
		Beneficiary:           block.Beneficiary,
		NrAccTx:               block.NrAccTx,
		NrFundsTx:             block.NrFundsTx,
		NrConfigTx:            block.NrConfigTx,
		NrStakeTx:             block.NrStakeTx,
		NrElementsBF:          block.NrElementsBF,
		BloomFilter:           block.BloomFilter,
		SlashedAddress:        block.SlashedAddress,
		Height:                block.Height,
		CommitmentProof:	   block.CommitmentProof,
		ConflictingBlockHash1: block.ConflictingBlockHash1,
		ConflictingBlockHash2: block.ConflictingBlockHash2,

		AccTxData:    block.AccTxData,
		FundsTxData:  block.FundsTxData,
		ConfigTxData: block.ConfigTxData,
		StakeTxData:  block.StakeTxData,
	}

	buffer := new(bytes.Buffer)
	gob.NewEncoder(buffer).Encode(encoded)
	return buffer.Bytes()
}

func (block *Block) EncodeHeader() []byte {
	if block == nil {
		return nil
	}

	encoded := Block{
		Header:       block.Header,
		Hash:         block.Hash,
		PrevHash:     block.PrevHash,
		NrConfigTx:   block.NrConfigTx,
		NrElementsBF: block.NrElementsBF,
		BloomFilter:  block.BloomFilter,
		Height:       block.Height,
		Beneficiary:  block.Beneficiary,
	}

	buffer := new(bytes.Buffer)
	gob.NewEncoder(buffer).Encode(encoded)
	return buffer.Bytes()
}

func (block *Block) Decode(encoded []byte) (b *Block) {
	if encoded == nil {
		return nil
	}

	var decoded Block
	buffer := bytes.NewBuffer(encoded)
	decoder := gob.NewDecoder(buffer)
	decoder.Decode(&decoded)
	return &decoded
}

func (block Block) String() string {
	return fmt.Sprintf("\nHash: %x\n"+
		"Previous Hash: %x\n"+
		"Nonce: %x\n"+
		"Timestamp: %v\n"+
		"MerkleRoot: %x\n"+
		"Beneficiary: %x\n"+
		"Amount of fundsTx: %v --> %x\n"+
		"Amount of accTx: %v --> %x\n"+
		"Amount of configTx: %v --> %x\n"+
		"Amount of stakeTx: %v --> %x\n"+
		"Total Transactions in this block: %v\n"+
		"Height: %d\n"+
		"Commitment Proof: %x\n"+
		"Slashed Address:%x\n"+
		"Conflicted Block Hash 1:%x\n"+
		"Conflicted Block Hash 2:%x\n",
		block.Hash[0:8],
		block.PrevHash[0:8],
		block.Nonce,
		block.Timestamp,
		block.MerkleRoot[0:8],
		block.Beneficiary[0:8],
		block.NrFundsTx, block.FundsTxData,
		block.NrAccTx, block.AccTxData,
		block.NrConfigTx, block.ConfigTxData,
		block.NrStakeTx, block.StakeTxData,
		uint16(block.NrFundsTx) + uint16(block.NrAccTx) + uint16(block.NrConfigTx) + uint16(block.NrStakeTx),
		block.Height,
		block.CommitmentProof[0:8],
		block.SlashedAddress[0:8],
		block.ConflictingBlockHash1[0:8],
		block.ConflictingBlockHash2[0:8],
	)
}
