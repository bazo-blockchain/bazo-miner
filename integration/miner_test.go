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
)


var (
	waitTimeSeconds     = time.Duration(20) * time.Second
	cmdCreateMiner      = []string{"accTx", "0", "1", RootKey, MinerKey}
	cmdFundMiner        = []string{"fundsTx", "0", "1000", "1", "0", RootKey, MinerKey, MultisigKey}
	cmdStakeMiner       = []string{"stakeTx", "0", "5", "1", MinerKey, MinerKey}
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
	fundMiner(t)
	stakeMiner(t)
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
	assert.NoError(t, err)
	assert.Equal(t, uint64(0), acc.Balance, "non zero balance")
	assert.False(t, acc.IsRoot, "account shouldn't be root")
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

func _startMiner() {
	// TODO: find out why MinerKey cannot be used. How would the miner receive the root key safely?
	minerPubKey, _, _ := storage.ExtractKeyFromFile(RootKey)
	multisigPubKey, _, _ := storage.ExtractKeyFromFile(MultisigKey)
	miner.Init(&minerPubKey, &multisigPubKey, minerSeedFileName, false)
}

func startMiner(t *testing.T) {
	os.Remove(minerDbName)
	miner.InitialRootBalance = InitRootBalance
	multisigPubKey, _, _ := storage.ExtractKeyFromFile(RootKey)
	storage.INITROOTPUBKEY1 = multisigPubKey.X.Text(16)
	storage.INITROOTPUBKEY2 = multisigPubKey.Y.Text(16)
	storage.Init(minerDbName, minerIpPort)
	p2p.Init(minerIpPort)

	go _startMiner()
	time.Sleep(waitTimeSeconds)

	// We expect the miner to download the blocks and get the current state
	assert.Equal(t, 2, len(storage.State))
	// TODO: check balances and other fields to make sure the calculation was correct
}

