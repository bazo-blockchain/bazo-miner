package miner

import (
	"crypto/ecdsa"
	"errors"
	"fmt"
	"github.com/bazo-blockchain/bazo-miner/protocol"
	"github.com/bazo-blockchain/bazo-miner/storage"
	"github.com/bazo-blockchain/bazo-miner/p2p"
	"log"
	"sync"
	"github.com/bazo-blockchain/bazo-miner/conf"
)

var (
	logger              *log.Logger
	blockValidation     = &sync.Mutex{}
	parameterSlice      []conf.Parameters
	activeParameters    *conf.Parameters
	uptodate            bool
	slashingDict        map[[32]byte]SlashingProof
	validatorAccAddress [64]byte
	multisigPubKey      *ecdsa.PublicKey
	seedFile		 	string
)

//Miner entry point
func Init(validatorPubKey, multisig *ecdsa.PublicKey, seedFileName string, isBootstrap bool,) {
	var hashedSeed [32]byte
	var blockToMine *protocol.Block

	//Set up logger
	logger = storage.InitLogger()

	// Initialise variables
	validatorAccAddress = storage.GetAddressFromPubKey(validatorPubKey)
	multisigPubKey = multisig
	seedFile = seedFileName
	parameterSlice = append(parameterSlice, NewDefaultParameters())
	activeParameters = &parameterSlice[0]
	currentTargetTime = new(timerange)
	target = append(target, 15)

	logger.Printf("Active config params:%v", activeParameters)

	//We must first update the state before we can start mining.
	// In order to make PoS we must know our balance in the state
	// Get the last closed block from DB or create genesis
	initialBlock, err := SetUpInitialState(hashedSeed)
	blockToMine = initialBlock
	if err != nil {
		logger.Printf("Could not set up initial state: %v.\n", err)
		return
	}

	//Initialize root key
	//the hashedSeed is necessary since it must be included in the initial block
	if hashedSeed, err = initRootKey(); err != nil {
		logger.Printf("Could not create a root account.\n")
		return
	}

	validatorAccount := storage.GetAccount(protocol.SerializeHashContent(validatorAccAddress))
	if validatorAccount == nil {
		fmt.Printf("Error: Validator address not found in state!\n"+
			"This means that you are trying to bootstrap with a key that is not part of the state.\n"+
			"Validator address expected: %x\n", validatorAccAddress)
		return
	}
	if !isBootstrap {
		payload := <-p2p.BlockIn
		processBlock(payload)
		blockToMine = lastBlock
		logger.Println("############################start mining############################")
	}

	// Listen for incoming blocks
	go incomingData()
	// Start the mining
	mining(blockToMine)
}

//Mining is a constant process, trying to come up with a successful PoW
func mining(initialBlock *protocol.Block) {
	currentBlock := newBlock(initialBlock.Hash, [32]byte{}, [32]byte{}, initialBlock.Height+1)

	for {
		err := finalizeBlock(currentBlock)
		if err != nil {
			fmt.Printf("%v\n", err)
		} else {
			fmt.Printf("Block mined: %v\n", currentBlock.Height)
			//else a block was received meanwhile that was added to the chain, all the effort was in vain :(
			//wait for lock here only
			broadcastBlock(currentBlock)
			err := validateBlock(currentBlock)
			if err != nil {
				logger.Printf("Received block (%x) could not be validated: %v\n", currentBlock.Hash[0:8], err)
			}

			// Block has been finalized and validated.
			// If a consolidation happened at this point we can remove the old blocks.
			// TODO: move to postvalidation?
			if currentBlock.NrConsolidationTx != 0 {
				removeOldBlocks(currentBlock)
			}
		}

		//This is the same mutex that is claimed at the beginning of a block validation. The reason we do this is
		//that before start mining a new block we empty the mempool which contains tx data that is likely to be
		//validated with block validation, so we wait in order to not work on tx data that is already validated
		//when we finish the block
		blockValidation.Lock()
		nextBlock := newBlock(lastBlock.Hash, [32]byte{}, [32]byte{}, lastBlock.Height+1)
		currentBlock = nextBlock

		// Create ConsolidationTx
		if currentBlock.Height > 0 && currentBlock.Height % uint32(activeParameters.Consolidation_interval) == 0 {
			fmt.Printf("Creating Consolidation tx for height %v\n", currentBlock.Height)
			consolidationTx, err := GetConsolidationTx(currentBlock.PrevHash)
			if err != nil {
				fmt.Println("Error while adding consolidation tx")
			}
			// ConsolidationTx will be added in finalizeBlock
			storage.WriteOpenTx(consolidationTx)
		}
		prepareBlock(currentBlock)
		blockValidation.Unlock()
	}
}

func GetActiveParameters() (parameter *conf.Parameters) {
	return activeParameters
}

func removeOldBlocks(b *protocol.Block) {
	for _, txHash := range b.ConsolidationTxData {
		closedTx := storage.ReadClosedTx(txHash)

		if closedTx == nil {
			fmt.Printf("ERROR: nil closed tx")
			return
		}
		consTx := closedTx.(*protocol.ConsolidationTx)
		// Deletion start a this point and continues till there are no more blocks to delete
		blockHashToDelete := consTx.PreviousConsHash

		for ;blockHashToDelete != [32]byte{}; {
			blockToDelete := storage.ReadClosedBlock(blockHashToDelete)

			if blockToDelete == nil {
				fmt.Printf("No more blocks to delete, last one was: %v\n", blockHashToDelete)
				return
			}
			if blockToDelete.NrConsolidationTx > 0 {
				fmt.Printf("Not Deleting Block: %v -- %v\n", blockToDelete.Height, blockHashToDelete)
			} else {
				fmt.Printf("Deleting Block: %v -- %v\n", blockToDelete.Height, blockHashToDelete)
				storage.DeleteClosedBlock(blockHashToDelete)
				// TODO: delete transactions?
			}
			blockHashToDelete = blockToDelete.PrevHash
		}
	}
}
//At least one root key needs to be set which is allowed to create new accounts
func initRootKey() ([32]byte, error) {
	address, addressHash := storage.GetInitRootPubKey()

	var seed [32]byte
	copy(seed[:], storage.INIT_ROOT_SEED[:])

	//Create the hash of the seed which will be included in the transaction
	hashedSeed := protocol.SerializeHashContent(seed)

	newSeed := storage.SeedJson{
		HashedSeed: fmt.Sprintf("%x", string(hashedSeed[:])),
		Seed: string(seed[:])}

	if err := storage.AppendNewSeed(seedFile, newSeed); err != nil {
		return hashedSeed, errors.New(fmt.Sprintf("Error creating the seed file."))
	}

	//Balance must be greater staking minimum
	rootAcc := protocol.NewAccount(address, InitialRootBalance, true, hashedSeed)
	//Add root key to the state
	storage.State[addressHash] = &rootAcc
	storage.RootKeys[addressHash] = &rootAcc

	return hashedSeed, nil
}
