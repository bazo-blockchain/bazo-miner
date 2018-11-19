package protocol

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"github.com/bazo-blockchain/bazo-miner/crypto"
	"github.com/willf/bloom"
	"golang.org/x/crypto/sha3"
)

const (
	TXHASH_LEN              = 32
	HEIGHT_LEN              = 4
	MIN_BLOCKHEADER_SIZE    = 136
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

	//Body
	Nonce                 [8]byte
	Timestamp             int64
	MerkleRoot            [32]byte
	NrAccTx               uint16
	NrFundsTx             uint16
	NrStakeTx             uint16
	SlashedAddress        [64]byte
	CommitmentProof       [crypto.COMM_PROOF_LENGTH]byte
	ConflictingBlockHash1 [32]byte
	ConflictingBlockHash2 [32]byte
	StateCopy             map[[64]byte]*Account //won't be serialized, just keeping track of local state changes

	AccTxData    [][32]byte
	FundsTxData  [][32]byte
	ConfigTxData [][32]byte
	StakeTxData  [][32]byte
}

//Just Hash() conflicts with struct field
func (b *Block) HashBlock() (hash [32]byte) {

	var buf bytes.Buffer

	blockToHash := struct {
		prevHash              [32]byte
		timestamp             int64
		merkleRoot            [32]byte
		beneficiary           [64]byte
		commitmentProof       [crypto.COMM_PROOF_LENGTH]byte
		slashedAddress        [64]byte
		conflictingBlockHash1 [32]byte
		conflictingBlockHash2 [32]byte
	}{
		b.PrevHash,
		b.Timestamp,
		b.MerkleRoot,
		b.Beneficiary,
		b.CommitmentProof,
		b.SlashedAddress,
		b.ConflictingBlockHash1,
		b.ConflictingBlockHash2,
	}

	binary.Write(&buf, binary.BigEndian, blockToHash)
	return sha3.Sum256(buf.Bytes())
}

func (b *Block) InitBloomFilter(txPubKeys [][64]byte) {
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
			int(b.NrAccTx)*TXHASH_LEN +
			int(b.NrFundsTx)*TXHASH_LEN +
			int(b.NrConfigTx)*TXHASH_LEN +
			int(b.NrStakeTx)*TXHASH_LEN

	if b.BloomFilter != nil {
		encodedBF, _ := b.BloomFilter.GobEncode()
		size += len(encodedBF)
	}

	return uint64(size)
}

func (b *Block) GetHeaderSize() uint64 {
	size := MIN_BLOCKHEADER_SIZE

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

	encodedBlock[0] = b.Header
	copy(encodedBlock[1:33], b.Hash[:])
	copy(encodedBlock[33:65], b.PrevHash[:])
	copy(encodedBlock[65:73], b.Nonce[:])
	copy(encodedBlock[73:81], timeStamp[:])
	copy(encodedBlock[81:113], b.MerkleRoot[:])
	copy(encodedBlock[113:177], b.Beneficiary[:])
	copy(encodedBlock[177:179], nrFundsTx[:])
	copy(encodedBlock[179:181], nrAccTx[:])
	encodedBlock[181] = byte(b.NrConfigTx)
	copy(encodedBlock[182:184], nrStakeTx[:])
	copy(encodedBlock[184:248], b.SlashedAddress[:])
	copy(encodedBlock[248:250], nrElementsBF[:])

	index := 250

	if b.BloomFilter != nil {
		//Encode BloomFilter
		encodedBF, _ := b.BloomFilter.GobEncode()

		encodedBFSize := len(encodedBF)

		//Serialize BloomFilter
		copy(encodedBlock[index:index+encodedBFSize], encodedBF)
		index += encodedBFSize
	}

	copy(encodedBlock[index:index+HEIGHT_LEN], height[:])
	index += HEIGHT_LEN
	copy(encodedBlock[index:index+crypto.COMM_PROOF_LENGTH], b.CommitmentProof[:])
	index += crypto.COMM_PROOF_LENGTH
	copy(encodedBlock[index:index+TXHASH_LEN], b.ConflictingBlockHash1[:])
	index += TXHASH_LEN
	copy(encodedBlock[index:index+TXHASH_LEN], b.ConflictingBlockHash2[:])
	index += TXHASH_LEN

	//Serialize all tx hashes
	for _, txHash := range b.FundsTxData {
		copy(encodedBlock[index:index+TXHASH_LEN], txHash[:])
		index += TXHASH_LEN
	}

	for _, txHash := range b.AccTxData {
		copy(encodedBlock[index:index+TXHASH_LEN], txHash[:])
		index += TXHASH_LEN
	}

	for _, txHash := range b.ConfigTxData {
		copy(encodedBlock[index:index+TXHASH_LEN], txHash[:])
		index += TXHASH_LEN
	}

	for _, txHash := range b.StakeTxData {
		copy(encodedBlock[index:index+TXHASH_LEN], txHash[:])
		index += TXHASH_LEN
	}

	return encodedBlock
}

func (b *Block) EncodeHeader() (encodedHeader []byte) {
	if b == nil {
		return nil
	}

	//Making byte array of all non-byte data
	var nrElementsBF [2]byte
	var height [HEIGHT_LEN]byte

	binary.BigEndian.PutUint16(nrElementsBF[:], b.NrElementsBF)
	binary.BigEndian.PutUint32(height[:], b.Height)

	//Allocate memory
	encodedHeader = make([]byte, b.GetHeaderSize())

	encodedHeader[0] = b.Header
	copy(encodedHeader[1:33], b.Hash[:])
	copy(encodedHeader[33:65], b.PrevHash[:])
	encodedHeader[65] = byte(b.NrConfigTx)
	copy(encodedHeader[66:68], nrElementsBF[:])
	copy(encodedHeader[68:72], height[:])
	copy(encodedHeader[72:136], b.Beneficiary[:])

	index := MIN_BLOCKHEADER_SIZE

	if b.BloomFilter != nil {
		//Encode BloomFilter
		encodedBF, _ := b.BloomFilter.GobEncode()

		encodedBFSize := len(encodedBF)

		//Serialize BloomFilter
		copy(encodedHeader[index:index+encodedBFSize], encodedBF)
		index += encodedBFSize
	}

	return encodedHeader
}

func (*Block) Decode(encodedBlock []byte) (b *Block) {
	b = new(Block)

	if len(encodedBlock) < MIN_BLOCKSIZE {
		return nil
	}

	timeStampTmp := binary.BigEndian.Uint64(encodedBlock[73:81])
	timeStamp := int64(timeStampTmp)

	b.Header = encodedBlock[0]
	copy(b.Hash[:], encodedBlock[1:33])
	copy(b.PrevHash[:], encodedBlock[33:65])
	copy(b.Nonce[:], encodedBlock[65:73])
	b.Timestamp = timeStamp
	copy(b.MerkleRoot[:], encodedBlock[81:113])
	copy(b.Beneficiary[:], encodedBlock[113:177])
	b.NrFundsTx = binary.BigEndian.Uint16(encodedBlock[177:179])
	b.NrAccTx = binary.BigEndian.Uint16(encodedBlock[179:181])
	b.NrConfigTx = uint8(encodedBlock[181])
	b.NrStakeTx = binary.BigEndian.Uint16(encodedBlock[182:184])
	copy(b.SlashedAddress[:], encodedBlock[184:248])
	b.NrElementsBF = binary.BigEndian.Uint16(encodedBlock[248:250])

	index := 250

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

	b.Height = binary.BigEndian.Uint32(encodedBlock[index : index+HEIGHT_LEN])
	index += HEIGHT_LEN
	copy(b.CommitmentProof[:], encodedBlock[index:index+crypto.COMM_PROOF_LENGTH])
	index += crypto.COMM_PROOF_LENGTH
	copy(b.ConflictingBlockHash1[:], encodedBlock[index:index+TXHASH_LEN])
	index += TXHASH_LEN
	copy(b.ConflictingBlockHash2[:], encodedBlock[index:index+TXHASH_LEN])
	index += TXHASH_LEN

	//Deserialize all tx hashes
	var hash [32]byte
	for cnt := 0; cnt < int(b.NrFundsTx); cnt++ {
		copy(hash[:], encodedBlock[index:index+TXHASH_LEN])
		b.FundsTxData = append(b.FundsTxData, hash)
		index += TXHASH_LEN
	}

	for cnt := 0; cnt < int(b.NrAccTx); cnt++ {
		copy(hash[:], encodedBlock[index:index+TXHASH_LEN])
		b.AccTxData = append(b.AccTxData, hash)
		index += TXHASH_LEN
	}

	for cnt := 0; cnt < int(b.NrConfigTx); cnt++ {
		copy(hash[:], encodedBlock[index:index+TXHASH_LEN])
		b.ConfigTxData = append(b.ConfigTxData, hash)
		index += TXHASH_LEN
	}

	for cnt := 0; cnt < int(b.NrStakeTx); cnt++ {
		copy(hash[:], encodedBlock[index:index+TXHASH_LEN])
		b.StakeTxData = append(b.StakeTxData, hash)
		index += TXHASH_LEN
	}

	b.StateCopy = make(map[[64]byte]*Account)

	return b
}

func (*Block) DecodeHeader(encodedHeader []byte) (b *Block) {

	b = new(Block)

	if len(encodedHeader) < MIN_BLOCKHEADER_SIZE {
		return nil
	}

	b.Header = encodedHeader[0]
	copy(b.Hash[:], encodedHeader[1:33])
	copy(b.PrevHash[:], encodedHeader[33:65])
	b.NrConfigTx = uint8(encodedHeader[65])
	b.NrElementsBF = binary.BigEndian.Uint16(encodedHeader[66:68])
	b.Height = binary.BigEndian.Uint32(encodedHeader[68:72])
	copy(b.Beneficiary[:], encodedHeader[72:136])

	index := MIN_BLOCKHEADER_SIZE

	if b.NrElementsBF > 0 {
		m, k := calculateBloomFilterParams(float64(b.NrElementsBF), BLOOM_FILTER_ERROR_RATE)

		//Initialize BloomFilter
		b.BloomFilter = bloom.New(m, k)

		//Encode BloomFilter
		encodedBF, _ := b.BloomFilter.GobEncode()

		encodedBFSize := len(encodedBF)

		//Deserialize BloomFilter
		b.BloomFilter.GobDecode(encodedHeader[index : index+encodedBFSize])
		index += encodedBFSize
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
		"Height: %d\n"+
		"Commitment Proof: %x\n"+
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
		b.Height,
		b.CommitmentProof[0:8],
		b.SlashedAddress[0:8],
		b.ConflictingBlockHash1[0:8],
		b.ConflictingBlockHash2[0:8],
	)
}
