package miner

import (
	"crypto/ecdsa"
	"crypto/rsa"
	"github.com/bazo-blockchain/bazo-miner/crypto"
	"log"
	"sync"
	"time"

	"github.com/bazo-blockchain/bazo-miner/protocol"
	"github.com/bazo-blockchain/bazo-miner/storage"
)

var (
	logger              *log.Logger
	blockValidation     = &sync.Mutex{}
	parameterSlice      []Parameters
	activeParameters    *Parameters
	uptodate            bool
	slashingDict        = make(map[[64]byte]SlashingProof)
	validatorAccAddress [64]byte
	commPrivKey         *rsa.PrivateKey
)

//Miner entry point
func Init(wallet *ecdsa.PublicKey, commitment *rsa.PrivateKey) error {
	validatorAccAddress = crypto.GetAddressFromPubKey(wallet)
	commPrivKey = commitment

	//Set up logger.
	logger = storage.InitLogger()

	parameterSlice = append(parameterSlice, NewDefaultParameters())
	activeParameters = &parameterSlice[0]

	currentTargetTime = new(timerange)
	target = append(target, 15)

	initialBlock, err := initState()
	if err != nil {
		return err
	}

	logger.Printf("Active config params:%v", activeParameters)

	//Start to listen to network inputs (txs and blocks).
	go incomingData()
	mining(initialBlock)

	return nil
}

func InitFirstStart(wallet *ecdsa.PublicKey, commitment *rsa.PrivateKey) error {
	rootAddress := crypto.GetAddressFromPubKey(wallet)

	var rootCommitment [crypto.COMM_KEY_LENGTH]byte
	copy(rootCommitment[:], commitment.N.Bytes())

	genesis := protocol.NewGenesis(rootAddress, rootCommitment)
	storage.WriteGenesis(&genesis)
	return Init(wallet, commitment)
}

//Mining is a constant process, trying to come up with a successful PoW.
func mining(initialBlock *protocol.Block) {
	currentBlock := newBlock(initialBlock.Hash, [crypto.COMM_PROOF_LENGTH]byte{}, initialBlock.Height+1)

	for {
		_, err := storage.ReadAccount(validatorAccAddress)
		if err != nil {
			logger.Printf("%v\n", err)
			time.Sleep(10 * time.Second)
			continue
		}

		err = finalizeBlock(currentBlock)
		if err != nil {
			logger.Printf("%v\n", err)
		} else {
			logger.Printf("Block mined (%x)\n", currentBlock.Hash[0:8])
		}

		if err == nil {
			broadcastBlock(currentBlock)
			err := validate(currentBlock, false)
			if err == nil {
				logger.Printf("Validated block: %vState:\n%v", currentBlock, getState())
			} else {
				logger.Printf("Received block (%x) could not be validated: %v\n", currentBlock.Hash[0:8], err)
			}
		}

		//This is the same mutex that is claimed at the beginning of a block validation. The reason we do this is
		//that before start mining a new block we empty the mempool which contains tx data that is likely to be
		//validated with block validation, so we wait in order to not work on tx data that is already validated
		//when we finish the block.
		blockValidation.Lock()
		nextBlock := newBlock(lastBlock.Hash, [crypto.COMM_PROOF_LENGTH]byte{}, lastBlock.Height+1)
		currentBlock = nextBlock
		prepareBlock(currentBlock)
		blockValidation.Unlock()
	}
}
