package integration

import (
	"testing"
	"time"
	"github.com/bazo-blockchain/bazo-client/client"
	"github.com/bazo-blockchain/bazo-miner/storage"
	"github.com/bazo-blockchain/bazo-miner/p2p"
	"github.com/bazo-blockchain/bazo-miner/miner"
	"github.com/stretchr/testify/assert"
	"fmt"
	"os"
	"strconv"
)

var (
	waitTimeSeconds     = time.Duration(20) * time.Second
	cmdCreateMiner      = []string{"accTx", "0", "1", RootKey, MinerKey}
	cmdFundMiner        = []string{"fundsTx", "0", "1000", "1", "0", RootKey, MinerKey, MultisigKey}
	cmdFundMiner2       = []string{"fundsTx", "0", "123", "2", "1", RootKey, MinerKey, MultisigKey}
	cmdStakeMiner       = []string{"stakeTx", "0", "5", "1", MinerKey, MinerKey}
	cmdUnstakeMiner     = []string{"stakeTx", "1", "5", "0", MinerKey, MinerKey}
	newConsInterval	    = "11"
	cmdConsConfig       = []string{"configTx", "0", "11", newConsInterval, "1", "0", RootKey}
	minerIpPort         = "127.0.0.1:8002"
	minerDbName         = "miner_test.db"
	minerSeedFileName   = "miner_seed_test.json"
)

func TestMiner (t *testing.T){
	// At this point a bootstrap miner is already running in the background
	// We want to create a new miner from scratch and run it
	client.AutoRefresh = false
	client.Init()
	createMiner(t)
	sendConsConfig(t)
	fundMiner(t)
	fundMiner2(t)
	stakeMiner(t)
	unstakeMiner(t)
	//..start miner and check that everything is ok
	startMiner(t)
}

func createMiner(t *testing.T) {
	client.Process(cmdCreateMiner)
	time.Sleep(waitTimeSeconds)

	MinerPubKey, _, _ := storage.ExtractKeyFromFile(MinerKey)
	MinerAccAddress := storage.GetAddressFromPubKey(&MinerPubKey)
	client.InitState()

	acc, _, err := client.GetAccount(MinerAccAddress)
	if err != nil {
		assert.NoError(t, err)
		t.Fatal()
	}
	assert.Equal(t, uint64(0), acc.Balance, "non zero balance")
	assert.False(t, acc.IsRoot, "account shouldn't be root")
}

func sendConsConfig(t *testing.T) {
	client.Process(cmdConsConfig)
	time.Sleep(waitTimeSeconds)
}

func fundMiner(t *testing.T) {
	client.Process(cmdFundMiner)
	time.Sleep(waitTimeSeconds)

	MinerPubKey, _, _ := storage.ExtractKeyFromFile(MinerKey)
	MinerAccAddress := storage.GetAddressFromPubKey(&MinerPubKey)
	client.InitState()

	acc, _, err := client.GetAccount(MinerAccAddress)
	assert.NoError(t, err)
	assert.Equal(t, uint64(1000), acc.Balance, "incorrect balance")
}

func fundMiner2(t *testing.T) {
	client.Process(cmdFundMiner2)
	time.Sleep(waitTimeSeconds)

	MinerPubKey, _, _ := storage.ExtractKeyFromFile(MinerKey)
	MinerAccAddress := storage.GetAddressFromPubKey(&MinerPubKey)
	client.InitState()

	acc, _, err := client.GetAccount(MinerAccAddress)
	assert.NoError(t, err)
	assert.Equal(t, uint64(1123), acc.Balance, "incorrect balance")
}


func stakeMiner(t *testing.T) {
	client.Process(cmdStakeMiner)
	time.Sleep(waitTimeSeconds)

	MinerPubKey, _, _ := storage.ExtractKeyFromFile(MinerKey)
	MinerAccAddress := storage.GetAddressFromPubKey(&MinerPubKey)

	client.InitState()
	acc, _, err := client.GetAccount(MinerAccAddress)
	assert.NoError(t, err)
	fmt.Println(acc)
	// TODO: staking is not implemented in client so these are not correct
	//assert.Equal(t, uint64(995), acc.Balance, "incorrect balance:\n%v", acc)
	//assert.True(t, acc.isStaking, "miner is not staking")
}


func unstakeMiner(t *testing.T) {
	client.Process(cmdUnstakeMiner)
	time.Sleep(waitTimeSeconds)

	MinerPubKey, _, _ := storage.ExtractKeyFromFile(MinerKey)
	MinerAccAddress := storage.GetAddressFromPubKey(&MinerPubKey)

	client.InitState()
	acc, _, err := client.GetAccount(MinerAccAddress)
	assert.NoError(t, err)
	fmt.Println(acc)
	// TODO: staking is not implemented in client so these are not correct
	//assert.Equal(t, uint64(995), acc.Balance, "incorrect balance:\n%v", acc)
	//assert.False(t, acc.isStaking, "miner is staking")
}

func _startMiner() {
	// TODO: find out why MinerKey cannot be used. How would the miner receive the root key safely?
	minerPubKey, _, _ := storage.ExtractKeyFromFile(RootKey)
	multisigPubKey, _, _ := storage.ExtractKeyFromFile(MultisigKey)
	fmt.Printf("\n\n\nStarting test miner\n")
	miner.Init(&minerPubKey, &multisigPubKey, minerSeedFileName, false)
}

func startMiner(t *testing.T) {
	os.Remove(minerDbName)
	miner.InitialRootBalance = InitRootBalance
	multisigPubKey, _, _ := storage.ExtractKeyFromFile(RootKey)
	minerPubKey, _, _ := storage.ExtractKeyFromFile(MinerKey)
	storage.INITROOTPUBKEY1 = multisigPubKey.X.Text(16)
	storage.INITROOTPUBKEY2 = multisigPubKey.Y.Text(16)
	storage.Init(minerDbName, minerIpPort)
	p2p.Init(minerIpPort)
	go _startMiner()
	time.Sleep(waitTimeSeconds)

	// Check expected status
	assert.Equal(t, 2, len(storage.State))

	// Check root status
	validatorAccAddress := storage.GetAddressFromPubKey(&multisigPubKey)
	hashAddress := storage.SerializeHashContent(validatorAccAddress)
	acc := storage.GetAccount(hashAddress)
	assert.Equal(t, true, acc.Balance > 0)
	assert.Equal(t, uint32(2), acc.TxCnt)

	// Check miner status
	minerAccAddress := storage.GetAddressFromPubKey(&minerPubKey)
	minerHashAddress := storage.SerializeHashContent(minerAccAddress)
	minerAcc := storage.GetAccount(minerHashAddress)
	fmt.Printf("State:\n%v\n", storage.GetState())
	// 1000 + 123
	assert.Equal(t, uint64(1123), minerAcc.Balance)
	assert.Equal(t, uint32(0), minerAcc.TxCnt)
	assert.False(t, minerAcc.IsStaking)


	parameters := miner.GetActiveParameters()
	expectedConsInterval, _ := strconv.Atoi(newConsInterval)
	assert.Equal(t, uint64(expectedConsInterval), parameters.Consolidation_interval)


}

