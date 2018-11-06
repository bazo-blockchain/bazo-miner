package miner

import (
	"crypto/ecdsa"
	"crypto/rsa"
	"github.com/bazo-blockchain/bazo-miner/crypto"
	"log"
	"sync"

	"github.com/bazo-blockchain/bazo-miner/protocol"
	"github.com/bazo-blockchain/bazo-miner/storage"
)

var (
	logger              *log.Logger
	blockValidation     = &sync.Mutex{}
	parameterSlice      []Parameters
	activeParameters    *Parameters
	uptodate            bool
	slashingDict        = make(map[[32]byte]SlashingProof)
	validatorAccAddress [64]byte
	multisigPubKey      *ecdsa.PublicKey
	commPrivKey			*rsa.PrivateKey
	rootCommPrivKey		*rsa.PrivateKey
)

//Miner entry point
func Init(validatorPubKey, multisig *ecdsa.PublicKey, commitmentPrivKey *rsa.PrivateKey) {
	var err error

	validatorAccAddress = crypto.GetAddressFromPubKey(validatorPubKey)
	multisigPubKey = multisig
	commPrivKey = commitmentPrivKey

	//Set up logger.
	logger = storage.InitLogger()

	parameterSlice = append(parameterSlice, NewDefaultParameters())
	activeParameters = &parameterSlice[0]

	//Initialize root key.
	initRootKey()
	if err != nil {
		logger.Printf("Could not create a root account.\n")
	}

	currentTargetTime = new(timerange)
	target = append(target, 15)

	initialBlock, err := initState()
	if err != nil {
		logger.Printf("Could not set up initial state: %v.\n", err)
		return
	}

	logger.Printf("Active config params:%v", activeParameters)

	//Start to listen to network inputs (txs and blocks).
	go incomingData()
	mining(initialBlock)
}

//Mining is a constant process, trying to come up with a successful PoW.
func mining(initialBlock *protocol.Block) {
	currentBlock := newBlock(initialBlock.Hash, [crypto.COMM_PROOF_LENGTH]byte{}, initialBlock.Height+1)

	for {
		err := finalizeBlock(currentBlock)
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

//At least one root key needs to be set which is allowed to create new accounts.
func initRootKey() error {
	address, addressHash := storage.GetInitRootPubKey()

	rootComm, err := crypto.CreateRSAPrivKeyFromBase64(
		storage.INIT_ROOT_COMM_PUB_KEY,
		storage.INIT_ROOT_COMM_PRIV_KEY, []string {
			storage.INIT_ROOT_COMM_PRIME1,
			storage.INIT_ROOT_COMM_PRIME2,
		})

	if err != nil {
		return err
	}

	rootCommPrivKey = rootComm

	var commPubKey [crypto.COMM_KEY_LENGTH]byte
	copy(commPubKey[:], rootComm.N.Bytes())

	rootAcc := protocol.NewAccount(address, [32]byte{}, activeParameters.Staking_minimum, true, commPubKey, nil, nil)
	storage.State[addressHash] = &rootAcc
	storage.RootKeys[addressHash] = &rootAcc

	return nil
}
