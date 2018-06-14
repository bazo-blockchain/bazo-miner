package miner

import (
	"crypto/ecdsa"
	"errors"
	"fmt"
	"github.com/bazo-blockchain/bazo-miner/p2p"
	"github.com/bazo-blockchain/bazo-miner/protocol"
	"github.com/bazo-blockchain/bazo-miner/storage"
	"log"
	"sync"
	"time"
)

var (
	logger              *log.Logger
	blockValidation     = &sync.Mutex{}
	parameterSlice      []Parameters
	activeParameters    *Parameters
	uptodate            bool
	slashingDict        map[[32]byte]SlashingProof
	validatorAccAddress [64]byte
	multisigPubKey      *ecdsa.PublicKey
	seedFile            string
)

//Miner entry point
func Init(validatorPubKey, multisig *ecdsa.PublicKey, seedFileName string, isBootstrap bool) {
	var err error
	var hashedSeed [32]byte

	validatorAccAddress = storage.GetAddressFromPubKey(validatorPubKey)
	multisigPubKey = multisig

	seedFile = seedFileName

	//Set up logger
	logger = storage.InitLogger()

	//Initialize root key
	//the hashedSeed is necessary since it must be included in the initial block
	hashedSeed, err = initRootKey()
	if err != nil {
		logger.Printf("Could not create a root account.\n")
	}

	parameterSlice = append(parameterSlice, NewDefaultParameters())
	activeParameters = &parameterSlice[0]

	currentTargetTime = new(timerange)
	target = append(target, 15)

	// Get the last closed block from DB or create genesis
	initialBlock, err := SetUpInitialState(hashedSeed)
	if err != nil {
		logger.Printf("Could not set up initial state: %v.\n", err)
		return
	}


	}

	logger.Printf("Active config params:%v", activeParameters)

	//Start to listen to network inputs (txs and blocks)
	go incomingData()
	mining(initialBlock)
}

//Mining is a constant process, trying to come up with a successful PoW
func mining(initialBlock *protocol.Block) {
	currentBlock := newBlock(initialBlock.Hash, [32]byte{}, [32]byte{}, initialBlock.Height+1)

	for {
		err := finalizeBlock(currentBlock)
		if err != nil {
			fmt.Printf("%v\n", err)
		} else {
			fmt.Println("Block mined")
		}
		//else a block was received meanwhile that was added to the chain, all the effort was in vain :(
		//wait for lock here only
		if err == nil {
			broadcastBlock(currentBlock)
			err := validateBlock(currentBlock)
			if err != nil {
				logger.Printf("Received block (%x) could not be validated: %v\n", currentBlock.Hash[0:8], err)
			}
		}

		//This is the same mutex that is claimed at the beginning of a block validation. The reason we do this is
		//that before start mining a new block we empty the mempool which contains tx data that is likely to be
		//validated with block validation, so we wait in order to not work on tx data that is already validated
		//when we finish the block
		blockValidation.Lock()
		nextBlock := newBlock(lastBlock.Hash, [32]byte{}, [32]byte{}, lastBlock.Height+1)
		currentBlock = nextBlock
		prepareBlock(currentBlock)
		blockValidation.Unlock()
	}
}

//At least one root key needs to be set which is allowed to create new accounts
func initRootKey() ([32]byte, error) {
	address, addressHash := storage.GetInitRootPubKey()

	var seed [32]byte
	copy(seed[:], storage.INIT_ROOT_SEED[:])

	//Create the hash of the seed which will be included in the transaction
	hashedSeed := protocol.SerializeHashContent(seed)

	err := storage.AppendNewSeed(seedFile, storage.SeedJson{fmt.Sprintf("%x", string(hashedSeed[:])), string(seed[:])})
	if err != nil {
		return hashedSeed, errors.New(fmt.Sprintf("Error creating the seed file."))
	}

	//Balance must be greater staking minimum
	rootAcc := protocol.NewAccount(address, INITIALINITROOTBALANCE, true, hashedSeed)
	storage.State[addressHash] = &rootAcc
	storage.RootKeys[addressHash] = &rootAcc

	return hashedSeed, nil
}
