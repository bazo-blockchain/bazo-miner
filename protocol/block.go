package protocol

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"github.com/bazo-blockchain/bazo-miner/crypto"
	"github.com/willf/bloom"
)

const (
	TXHASH_LEN              = 32
	HEIGHT_LEN              = 4
	MIN_BLOCKHEADER_SIZE    = 144
	MIN_BLOCKSIZE           = 184 + MIN_BLOCKHEADER_SIZE + crypto.COMM_PROOF_LENGTH
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
	Beneficiary  [64]byte
	TotalFees	 uint64

	//Body
	Nonce                 [8]byte
	Timestamp             int64
	MerkleRoot            [32]byte
	NrContractTx          uint16
	NrFundsTx             uint16
	NrStakeTx             uint16
	NrTxBucket			  uint16
	SlashedAddress        [64]byte
	CommitmentProof       [crypto.COMM_PROOF_LENGTH]byte
	ConflictingBlockHash1 [32]byte
	ConflictingBlockHash2 [32]byte
	StateCopy             map[[64]byte]*Account 	// won't be serialized, just keeping track of local state changes
	TxBuckets			  txBucketMap 				// won't be serialized, just keeping track of transaction buckets

	ContractTxData  [][32]byte
	FundsTxData  	[][32]byte
	ConfigTxData 	[][32]byte
	StakeTxData  	[][32]byte
	TxBucketData	[][32]byte
}

func NewBlock(prevHash [32]byte, height uint32) *Block {
	newBlock := Block{
		PrevHash: prevHash,
		Height:   height,
	}

	newBlock.StateCopy = make(map[[64]byte]*Account)
	newBlock.TxBuckets = NewTxBucketMap()

	return &newBlock
}

func (block *Block) AddFundsTx(tx *FundsTx) {
	bucket := block.createBucketIfNotExists(tx.From)
	bucket.AddFundsTx(tx)

	bucket = block.createBucketIfNotExists(tx.To)
	bucket.AddFundsTx(tx)

	block.FundsTxData = append(block.FundsTxData, tx.Hash())
}

func (block *Block) createBucketIfNotExists(address AddressType) *TxBucket {
	bucket, exists := block.TxBuckets[address]
	if !exists {
		bucket = NewTxBucket(address)
		block.TxBuckets[address] = bucket
	}
	return bucket
}

func (block *Block) BuildMerkleTree() *MerkleTree {
	var hashes HashArray

	if block == nil {
		return nil
	}

	// Note that the Merkle roots of the buckets are
	// the leaves of the Merkle tree, not the FundsTx itselves
	/*for _, bucket := range block.TxBuckets {
		hashes = append(hashes, bucket.Hash())
	}*/

	if block.TxBucketData != nil {
		for _, hash := range block.TxBucketData {
			hashes = append(hashes, hash)
		}
	}
	if block.ContractTxData != nil {
		for _, txHash := range block.ContractTxData {
			hashes = append(hashes, txHash)
		}
	}
	if block.ConfigTxData != nil {
		for _, txHash := range block.ConfigTxData {
			hashes = append(hashes, txHash)
		}
	}

	if block.StakeTxData != nil {
		for _, txHash := range block.StakeTxData {
			hashes = append(hashes, txHash)
		}
	}

	//Merkle root for no transactions is 0 hash
	if len(hashes) == 0 {
		return nil
	}

	m, _ := NewMerkleTree(hashes)

	return m
}

func (block *Block) HashBlock() [32]byte {
	if block == nil {
		return [32]byte{}
	}

	blockHash := struct {
		prevHash              [32]byte
		timestamp             int64
		merkleRoot            [32]byte
		beneficiary           [64]byte
		totalFees			  uint64
		commitmentProof       [crypto.COMM_PROOF_LENGTH]byte
		slashedAddress        [64]byte
		conflictingBlockHash1 [32]byte
		conflictingBlockHash2 [32]byte
	}{
		block.PrevHash,
		block.Timestamp,
		block.MerkleRoot,
		block.Beneficiary,
		block.TotalFees,
		block.CommitmentProof,
		block.SlashedAddress,
		block.ConflictingBlockHash1,
		block.ConflictingBlockHash2,
	}
	return SerializeHashContent(blockHash)
}

func (block *Block) InitBloomFilter(txPubKeys [][64]byte) {
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
		MIN_BLOCKSIZE +
			int(block.NrContractTx)*TXHASH_LEN +
			int(block.NrFundsTx)*TXHASH_LEN +
			int(block.NrConfigTx)*TXHASH_LEN +
			int(block.NrStakeTx)*TXHASH_LEN +
			int(block.NrTxBucket)*TXHASH_LEN

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
		TotalFees:			   block.TotalFees,
		Beneficiary:           block.Beneficiary,
		NrContractTx:          block.NrContractTx,
		NrFundsTx:             block.NrFundsTx,
		NrConfigTx:            block.NrConfigTx,
		NrStakeTx:             block.NrStakeTx,
		NrTxBucket:			   block.NrTxBucket,
		NrElementsBF:          block.NrElementsBF,
		BloomFilter:           block.BloomFilter,
		SlashedAddress:        block.SlashedAddress,
		Height:                block.Height,
		CommitmentProof:       block.CommitmentProof,
		ConflictingBlockHash1: block.ConflictingBlockHash1,
		ConflictingBlockHash2: block.ConflictingBlockHash2,

		ContractTxData:    block.ContractTxData,
		FundsTxData:  block.FundsTxData,
		ConfigTxData: block.ConfigTxData,
		StakeTxData:  block.StakeTxData,
		TxBucketData: block.TxBucketData,
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
		TotalFees:	  block.TotalFees,
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

	decoded.StateCopy = make(map[[64]byte]*Account)
	decoded.TxBuckets = NewTxBucketMap()

	return &decoded
}

func (block Block) String() string {
	return fmt.Sprintf("\nHash: %x\n"+
		"Previous Hash: %x\n"+
		"Nonce: %x\n"+
		"Timestamp: %v\n"+
		"MerkleRoot: %x\n"+
		"Beneficiary: %x\n"+
		"Total Fees: %v\n"+
		"Amount of fundsTx: %v\n"+
		"Amount of contractTx: %v\n"+
		"Amount of configTx: %v\n"+
		"Amount of stakeTx: %v\n"+
		"Amount of txBuckets: %v\n"+
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
		block.TotalFees,
		block.NrFundsTx,
		block.NrContractTx,
		block.NrConfigTx,
		block.NrStakeTx,
		block.NrTxBucket,
		block.Height,
		block.CommitmentProof[0:8],
		block.SlashedAddress[0:8],
		block.ConflictingBlockHash1[0:8],
		block.ConflictingBlockHash2[0:8],
	)
}
