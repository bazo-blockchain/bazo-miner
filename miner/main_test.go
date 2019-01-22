package miner

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"fmt"
	"io/ioutil"
	"log"
	"math/big"
	"os"
	"testing"

	"github.com/bazo-blockchain/bazo-miner/p2p"
	"github.com/bazo-blockchain/bazo-miner/protocol"
	"github.com/bazo-blockchain/bazo-miner/storage"
)

const (
	TestDBFileName   = "test.db"
	TestIpPort       = "127.0.0.1:8000"
	TestSeedFileName = "test_seed.json"
	TestKeyFileName  = "test_root"
)

const (
	PubA1 = "c2be9abbeaec39a066c2a09cee23bb9ab2a0b88f2880b1e785b4d317adf0dc7c"
	PubA2 = "8ce020fde838d9c443f6c93345dafe7fd74f091c4d2f30b37e2453679a257ed5"
	PrivA = "ba127fa8f802b008b9cdb58f4e44809d48f1b000cff750dda9cd6b312395c1c5"
	PubB1 = "5d7eefd58e3d2f309471928ab4bbd104e52973372c159fa652b8ca6b57ff68b8"
	PubB2 = "ab301a6a77b201c416ddc13a2d33fdf200a5302f6f687e0ea09085debaf8a1d9"
	PrivB = "7a0a9babcc97ea7991ed67ed7f800f70c5e04e99718960ad8efab2ca052f00c7"

	//Root account for testing
	PubRoot1 = "6323cc034597195ae69bcfb628ecdffa5989c7503154c566bab4a87f3e9910ac"
	PubRoot2 = "f6115b77a15852764c609c6a5c1739e698ebc6e49bf14617c561b9110039cec7"
	PrivRoot = "277ed539f56122c25a6fc115d07d632b47e71416c9aebf1beb54ee704f11842c"

	//Multisig account for testing
	MultiSigPub1 = "d5a0c62eeaf699eeba121f92e08becd38577f57b83eba981dc057e92fde1ad22"
	MultiSigPub2 = "a480e4ee6ff8b4edbf9470631ec27d3b1eb27f210d5a994a7cbcffa3bfce958e"
	MultiSigPriv = "b8d1fa3cc7476eafca970ea222676647da1817d1d9dc602e9446290454ffe1a4"
)

//Globally accessible values for all other tests, (root)account-related
var (
	accA, accB, validatorAcc, multiSigAcc, rootAcc         *protocol.Account
	PrivKeyAccA, PrivKeyAccB, PrivKeyMultiSig, PrivKeyRoot ecdsa.PrivateKey
	genesisBlock                                           *protocol.Block
)

//Create some accounts that are used by the tests
func addTestingAccounts() {
	accA, accB, validatorAcc, multiSigAcc = new(protocol.Account), new(protocol.Account), new(protocol.Account), new(protocol.Account)

	pubAccA1, _ := new(big.Int).SetString(PubA1, 16)
	pubAccA2, _ := new(big.Int).SetString(PubA2, 16)
	privAccA, _ := new(big.Int).SetString(PrivA, 16)
	pubKeyAccA := ecdsa.PublicKey{
		elliptic.P256(),
		pubAccA1,
		pubAccA2,
	}
	PrivKeyAccA = ecdsa.PrivateKey{
		pubKeyAccA,
		privAccA,
	}

	copy(accA.Address[0:32], PrivKeyAccA.PublicKey.X.Bytes())
	copy(accA.Address[32:64], PrivKeyAccA.PublicKey.Y.Bytes())
	hashAccA := protocol.SerializeHashContent(accA.Address)

	pubAccB1, _ := new(big.Int).SetString(PubB1, 16)
	pubAccB2, _ := new(big.Int).SetString(PubB2, 16)
	privAccB, _ := new(big.Int).SetString(PrivB, 16)
	pubKeyAccB := ecdsa.PublicKey{
		elliptic.P256(),
		pubAccB1,
		pubAccB2,
	}
	PrivKeyAccB = ecdsa.PrivateKey{
		pubKeyAccB,
		privAccB,
	}

	copy(accB.Address[0:32], PrivKeyAccB.PublicKey.X.Bytes())
	copy(accB.Address[32:64], PrivKeyAccB.PublicKey.Y.Bytes())
	hashAccB := protocol.SerializeHashContent(accB.Address)

	privMultiSig, _ := new(big.Int).SetString(MultiSigPriv, 16)
	pubKeyMultiSig := storage.GetPubKeyFromString(MultiSigPub1, MultiSigPub2)
	PrivKeyMultiSig = ecdsa.PrivateKey{
		pubKeyMultiSig,
		privMultiSig,
	}

	copy(multiSigAcc.Address[0:32], PrivKeyMultiSig.PublicKey.X.Bytes())
	copy(multiSigAcc.Address[32:64], PrivKeyMultiSig.PublicKey.Y.Bytes())
	hashMultiSig := protocol.SerializeHashContent(multiSigAcc.Address)

	//Set the global variable in blockchain.go
	multisigPubKey = &pubKeyMultiSig

	privKeyValidator, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)

	copy(validatorAcc.Address[:32], privKeyValidator.X.Bytes())
	copy(validatorAcc.Address[32:64], privKeyValidator.Y.Bytes())
	hashValidator := protocol.SerializeHashContent(validatorAcc.Address)

	//Create and store an initial seed for the validator account.
	seed := protocol.CreateRandomSeed()
	hashedSeed := protocol.SerializeHashContent(seed)
	storage.AppendNewSeed(TestSeedFileName, storage.SeedJson{fmt.Sprintf("%x", string(hashedSeed[:])), string(seed[:])})

	validatorAcc.HashedSeed = hashedSeed
	validatorAcc.Balance = activeParameters.Staking_minimum
	validatorAcc.IsStaking = true

	//Set the global variable in blockchain.go
	validatorAccAddress = validatorAcc.Address

	storage.State[hashAccA] = accA
	storage.State[hashAccB] = accB
	storage.State[hashMultiSig] = multiSigAcc
	storage.State[hashValidator] = validatorAcc
}

//Create some root accounts that are used by the tests
func addRootAccounts() {
	rootAcc = new(protocol.Account)

	pubRoot1, _ := new(big.Int).SetString(PubRoot1, 16)
	pubRoot2, _ := new(big.Int).SetString(PubRoot2, 16)
	privRoot, _ := new(big.Int).SetString(PrivRoot, 16)
	pubKeyRoot := ecdsa.PublicKey{
		elliptic.P256(),
		pubRoot1,
		pubRoot2,
	}
	PrivKeyRoot = ecdsa.PrivateKey{
		pubKeyRoot,
		privRoot,
	}

	copy(rootAcc.Address[:32], PrivKeyRoot.X.Bytes())
	copy(rootAcc.Address[32:64], PrivKeyRoot.Y.Bytes())
	hashRoot := protocol.SerializeHashContent(rootAcc.Address)

	//Create root file
	file, _ := os.Create(TestKeyFileName)
	_, _ = file.WriteString(PubRoot1 + "\n")
	_, _ = file.WriteString(PubRoot2 + "\n")
	_, _ = file.WriteString(PrivRoot + "\n")

	var hashedSeed [32]byte

	//Create and store an initial seed for the root account.
	seed := protocol.CreateRandomSeed()
	hashedSeed = protocol.SerializeHashContent(seed)

	rootAcc.HashedSeed = hashedSeed
	rootAcc.Balance = activeParameters.Staking_minimum
	rootAcc.IsStaking = true

	storage.State[hashRoot] = rootAcc
	storage.RootKeys[hashRoot] = rootAcc
}

//The state changes (accounts, funds, system parameters etc.) need to be reverted before any new test starts
//So every test has the same view on the blockchain
func cleanAndPrepare() {
	storage.DeleteAll()

	tmpState := make(map[[32]byte]*protocol.Account)
	tmpRootKeys := make(map[[32]byte]*protocol.Account)

	storage.State = tmpState
	storage.RootKeys = tmpRootKeys

	lastBlock = nil

	globalBlockCount = -1
	localBlockCount = -1

	//Prepare system parameters
	targetTimes = []timerange{}
	currentTargetTime = new(timerange)
	target = append(target, 8)

	var tmpSlice []Parameters
	tmpSlice = append(tmpSlice, NewDefaultParameters())

	slashingDict = make(map[[32]byte]SlashingProof)

	parameterSlice = tmpSlice
	activeParameters = &tmpSlice[0]

	slashingDict = make(map[[32]byte]SlashingProof)

	genesisBlock = newBlock([32]byte{}, [32]byte{}, [32]byte{}, 0)
	copy(genesisBlock.Seed[:], storage.GENESIS_SEED)

	collectStatistics(genesisBlock)
	if err := storage.WriteClosedBlock(genesisBlock); err != nil {
		fmt.Printf("Error: %v\n", err)
	}
	if err := storage.WriteLastClosedBlock(genesisBlock); err != nil {
		fmt.Printf("Error: %v\n", err)
	}

	seedFile = TestSeedFileName

	//Override some params to ensure tests work correctly.
	activeParameters.num_included_prev_seeds = 0
	activeParameters.Block_reward = 1
	activeParameters.Slash_reward = 1

	addTestingAccounts()
	addRootAccounts()

	//Some meaningful balance to simplify testing
	//validatorAcc.Balance = 1000
	accA.Balance = 123232345678
	accB.Balance = 823237654321
	accA.TxCnt = 0
	accB.TxCnt = 0
}

func TestMain(m *testing.M) {
	storage.Init(TestDBFileName, TestIpPort)
	p2p.Init(TestIpPort)

	cleanAndPrepare()

	//We don't want logging msgs when testing, we have designated messages
	logger = log.New(nil, "", 0)
	logger.SetOutput(ioutil.Discard)
	retCode := m.Run()

	//Teardown
	storage.TearDown()
	os.Remove(TestDBFileName)
	os.Remove(TestKeyFileName)
	os.Remove(TestSeedFileName)
	os.Exit(retCode)
}
