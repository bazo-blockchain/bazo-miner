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
	MIN_BLOCKSIZE           = 318
	MIN_BLOCKHEADER_SIZE    = 68
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

	//Body
	Nonce                 [8]byte
	Timestamp             int64
	MerkleRoot            [32]byte
	Beneficiary           [32]byte
	NrAccTx               uint16
	NrFundsTx             uint16
	NrStakeTx             uint16
	SlashedAddress        [32]byte
	Seed                  [32]byte
	Height                uint32
	HashedSeed            [32]byte
	ConflictingBlockHash1 [32]byte
	ConflictingBlockHash2 [32]byte
	StateCopy             map[[32]byte]*Account //won't be serialized, just keeping track of local state changes

	AccTxData    [][32]byte
	FundsTxData  [][32]byte
	ConfigTxData [][32]byte
	StakeTxData  [][32]byte
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
	copy(encodedBlock[113:145], b.Beneficiary[:])
	copy(encodedBlock[145:147], nrFundsTx[:])
	copy(encodedBlock[147:149], nrAccTx[:])
	encodedBlock[149] = byte(b.NrConfigTx)
	copy(encodedBlock[150:152], nrStakeTx[:])
	copy(encodedBlock[152:184], b.SlashedAddress[:])
	copy(encodedBlock[184:186], nrElementsBF[:])

	index := 186

	if b.BloomFilter != nil {
		//Encode BloomFilter
		encodedBF, _ := b.BloomFilter.GobEncode()

		encodedBFSize := len(encodedBF)

		//Serialize BloomFilter
		copy(encodedBlock[index:index+encodedBFSize], encodedBF)
		index += encodedBFSize
	}

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

func (b *Block) EncodeHeader() (encodedHeader []byte) {
	if b == nil {
		return nil
	}

	//Making byte array of all non-byte data
	var nrElementsBF [2]byte

	binary.BigEndian.PutUint16(nrElementsBF[:], b.NrElementsBF)

	//Allocate memory
	encodedHeader = make([]byte, b.GetHeaderSize())

	encodedHeader[0] = b.Header
	copy(encodedHeader[1:33], b.Hash[:])
	copy(encodedHeader[33:65], b.PrevHash[:])
	encodedHeader[65] = byte(b.NrConfigTx)
	copy(encodedHeader[66:68], nrElementsBF[:])

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
	copy(b.Beneficiary[:], encodedBlock[113:145])
	b.NrFundsTx = binary.BigEndian.Uint16(encodedBlock[145:147])
	b.NrAccTx = binary.BigEndian.Uint16(encodedBlock[147:149])
	b.NrConfigTx = uint8(encodedBlock[149])
	b.NrStakeTx = binary.BigEndian.Uint16(encodedBlock[150:152])
	copy(b.SlashedAddress[:], encodedBlock[152:184])
	b.NrElementsBF = binary.BigEndian.Uint16(encodedBlock[184:186])

	index := 186

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
