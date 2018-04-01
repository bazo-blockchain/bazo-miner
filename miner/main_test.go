package miner

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"fmt"
	"github.com/sfontanach/bazo-miner/p2p"
	"github.com/sfontanach/bazo-miner/protocol"
	"github.com/sfontanach/bazo-miner/storage"
	"io/ioutil"
	"log"
	"math/big"
	"os"
	"testing"
)

//Some user accounts for testing
const (
	PubA1 = "c2be9abbeaec39a066c2a09cee23bb9ab2a0b88f2880b1e785b4d317adf0dc7c"
	PubA2 = "8ce020fde838d9c443f6c93345dafe7fd74f091c4d2f30b37e2453679a257ed5"
	PrivA = "ba127fa8f802b008b9cdb58f4e44809d48f1b000cff750dda9cd6b312395c1c5"
	PubB1 = "5d7eefd58e3d2f309471928ab4bbd104e52973372c159fa652b8ca6b57ff68b8"
	PubB2 = "ab301a6a77b201c416ddc13a2d33fdf200a5302f6f687e0ea09085debaf8a1d9"
	PrivB = "7a0a9babcc97ea7991ed67ed7f800f70c5e04e99718960ad8efab2ca052f00c7"
)

//Root account for testing
const (
	RootPub1 = "6323cc034597195ae69bcfb628ecdffa5989c7503154c566bab4a87f3e9910ac"
	RootPub2 = "f6115b77a15852764c609c6a5c1739e698ebc6e49bf14617c561b9110039cec7"
	RootPriv = "277ed539f56122c25a6fc115d07d632b47e71416c9aebf1beb54ee704f11842c"
)

//Validator account (multisig)
var (
	VerPub1 = "d5a0c62eeaf699eeba121f92e08becd38577f57b83eba981dc057e92fde1ad22"
	VerPub2 = "a480e4ee6ff8b4edbf9470631ec27d3b1eb27f210d5a994a7cbcffa3bfce958e"
	VerPriv = "b8d1fa3cc7476eafca970ea222676647da1817d1d9dc602e9446290454ffe1a4"
)

//Globally accessible values for all other tests, (root)account-related
var (
	accA, accB, minerAcc,
	validatorAcc                     *protocol.Account
	PrivKeyA, PrivKeyB, MinerPrivKey ecdsa.PrivateKey

	PubKeyA, PubKeyB, multiSignTest  ecdsa.PublicKey
	RootPrivKey, multiSignPrivKeyA   ecdsa.PrivateKey
	GenesisBlock                     *protocol.Block
	rootHash                         [32]byte
)

//Create some accounts that are used by the tests
func addTestingAccounts() {

	accA, accB, minerAcc, validatorAcc = new(protocol.Account), new(protocol.Account), new(protocol.Account), new(protocol.Account)

	puba1, _ := new(big.Int).SetString(PubA1, 16)
	puba2, _ := new(big.Int).SetString(PubA2, 16)
	priva, _ := new(big.Int).SetString(PrivA, 16)
	PubKeyA = ecdsa.PublicKey{
		elliptic.P256(),
		puba1,
		puba2,
	}
	PrivKeyA = ecdsa.PrivateKey{
		PubKeyA,
		priva,
	}

	pubb1, _ := new(big.Int).SetString(PubB1, 16)
	pubb2, _ := new(big.Int).SetString(PubB2, 16)
	privb, _ := new(big.Int).SetString(PrivB, 16)
	PubKeyB = ecdsa.PublicKey{
		elliptic.P256(),
		pubb1,
		pubb2,
	}
	PrivKeyB = ecdsa.PrivateKey{
		PubKeyB,
		privb,
	}

	copy(accA.Address[0:32], PrivKeyA.PublicKey.X.Bytes())
	copy(accA.Address[32:64], PrivKeyA.PublicKey.Y.Bytes())
	hashA := protocol.SerializeHashContent(accA.Address)

	//This one is just for testing purposes
	copy(accB.Address[0:32], PrivKeyB.PublicKey.X.Bytes())
	copy(accB.Address[32:64], PrivKeyB.PublicKey.Y.Bytes())
	hashB := protocol.SerializeHashContent(accB.Address)

	// Another pubkey to simulate multisig (Validator)
	multiSignTest, _ = storage.GetPubKeyFromString(VerPub1, VerPub2)

	multisigPubKey = &multiSignTest
	multisignpriv, _ := new(big.Int).SetString(VerPriv, 16)
	multiSignPrivKeyA = ecdsa.PrivateKey{
		multiSignTest,
		multisignpriv,
	}
	copy(validatorAcc.Address[0:32], multiSignPrivKeyA.PublicKey.X.Bytes())
	copy(validatorAcc.Address[32:64], multiSignPrivKeyA.PublicKey.Y.Bytes())
	copy(validatorAccAddress[0:32], multiSignPrivKeyA.PublicKey.X.Bytes())
	copy(validatorAccAddress[32:64], multiSignPrivKeyA.PublicKey.Y.Bytes())
	validatorHash := protocol.SerializeHashContent(validatorAcc.Address)
	storage.State[validatorHash] = validatorAcc
	//create and store an initial seed for the validator account
	seed := protocol.CreateRandomSeed()
	hashedSeed := protocol.SerializeHashContent(seed)
	_ = storage.AppendNewSeed(storage.SEED_FILE_NAME, storage.SeedJson{fmt.Sprintf("%x", string(hashedSeed[:])), string(seed[:])})
	validatorAcc.HashedSeed = hashedSeed
	validatorAcc.Balance = 100000
	validatorAcc.IsStaking = true

	//just to bootstrap
	storage.State[hashA] = accA
	storage.State[hashB] = accB

	minerPrivKey, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	var pubKey [64]byte
	var shortMiner [8]byte
	copy(pubKey[:32], minerPrivKey.X.Bytes())
	copy(pubKey[32:], minerPrivKey.Y.Bytes())
	minerHash := protocol.SerializeHashContent(pubKey)
	copy(shortMiner[:], minerHash[0:8])
	minerAcc.Address = pubKey
	storage.State[minerHash] = minerAcc
}

//Create some root accounts that are used by the tests
func addRootAccounts() {

	var pubKey [64]byte

	pub1, _ := new(big.Int).SetString(RootPub1, 16)
	pub2, _ := new(big.Int).SetString(RootPub2, 16)
	priv, _ := new(big.Int).SetString(RootPriv, 16)
	PubKeyA = ecdsa.PublicKey{
		elliptic.P256(),
		pub1,
		pub2,
	}
	RootPrivKey = ecdsa.PrivateKey{
		PubKeyA,
		priv,
	}

	copy(pubKey[32-len(pub1.Bytes()):32], pub1.Bytes())
	copy(pubKey[64-len(pub2.Bytes()):], pub2.Bytes())

	rootHash := protocol.SerializeHashContent(pubKey)

	rootAcc := protocol.Account{Address: pubKey}

	//create root file
	file, _ := os.Create(storage.DEFAULT_KEY_FILE_NAME)
	_, _ = file.WriteString(RootPub1 + "\n")
	_, _ = file.WriteString(RootPub2 + "\n")
	_, _ = file.WriteString(RootPriv + "\n")

	var hashedSeed [32]byte


	//create and store an initial seed for the root account
	seed := protocol.CreateRandomSeed()
	hashedSeed = protocol.SerializeHashContent(seed)
	_ = storage.AppendNewSeed(storage.SEED_FILE_NAME, storage.SeedJson{fmt.Sprintf("%x", string(hashedSeed[:])), string(seed[:])})

	rootAcc.HashedSeed = hashedSeed

	//set funds of root account in order to avoid zero division for PoS
	rootAcc.Balance = 10000000000
	rootAcc.IsStaking = true

	storage.State[rootHash] = &rootAcc
	storage.RootKeys[rootHash] = &rootAcc
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

	GenesisBlock = newBlock([32]byte{}, [32]byte{}, [32]byte{}, 0)

	var genesisSeedSlice [32]byte
	copy(genesisSeedSlice[:], storage.GENESIS_SEED)
	GenesisBlock.Seed = genesisSeedSlice

	collectStatistics(GenesisBlock)
	if err := storage.WriteClosedBlock(GenesisBlock); err != nil {
		fmt.Printf("Error: %v\n", err)
	}
	if err := storage.WriteLastClosedBlock(GenesisBlock); err != nil {
		fmt.Printf("Error: %v\n", err)
	}

	seedFile = "seed.json"
	addTestingAccounts()
	addRootAccounts()

	//Some meaningful balance to simplify testing
	minerAcc.Balance = 0
	accA.Balance = 123232345678
	accB.Balance = 823237654321
	accA.TxCnt = 0
	accB.TxCnt = 0
}

func TestMain(m *testing.M) {
	storage.Init("test.db" , "127.0.0.1:8000")
	p2p.Init("127.0.0.1:8000")

	addTestingAccounts()
	addRootAccounts()
	//We don't want logging msgs when testing, we have designated messages
	logger = log.New(nil, "", 0)
	logger.SetOutput(ioutil.Discard)
	os.Exit(m.Run())

	storage.TearDown()
}
