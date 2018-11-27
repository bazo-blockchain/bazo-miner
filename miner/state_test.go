package miner

import (
	"github.com/bazo-blockchain/bazo-miner/crypto"
	"math/rand"
	"reflect"
	"testing"
	"time"

	"github.com/bazo-blockchain/bazo-miner/protocol"
	"github.com/bazo-blockchain/bazo-miner/storage"
)

//Testing state change, rollback and fee collection
func TestFundsTxStateChange(t *testing.T) {
	cleanAndPrepare()

	randVar := rand.New(rand.NewSource(time.Now().Unix()))

	var testSize uint32
	testSize = 1000

	b := newBlock([32]byte{}, [crypto.COMM_PROOF_LENGTH]byte{}, 1)
	var funds []*protocol.FundsTx

	var feeA, feeB uint64

	//we're testing an overflowing balance in another test, this is that no interference occurs
	accA.Balance = 12343478374563434
	accB.Balance = 2947939489348234
	balanceA := accA.Balance
	balanceB := accB.Balance
	minerBal := validatorAcc.Balance

	loopMax := int(randVar.Uint32()%testSize + 1)
	for i := 0; i < loopMax+1; i++ {
		ftx, _ := protocol.ConstrFundsTx(0x01, randVar.Uint64()%1000000+1, randVar.Uint64()%100+1, uint32(i), accA.Address, accB.Address, PrivKeyAccA, nil, nil)
		if addTx(b, ftx) == nil {
			funds = append(funds, ftx)
			balanceA -= ftx.Amount
			feeA += ftx.Fee

			balanceB += ftx.Amount
		}

		ftx2, _ := protocol.ConstrFundsTx(0x01, randVar.Uint64()%1000+1, randVar.Uint64()%100+1, uint32(i), accA.Address, accB.Address, PrivKeyAccB, nil, nil)
		if addTx(b, ftx2) == nil {
			funds = append(funds, ftx2)
			balanceB -= ftx2.Amount
			feeB += ftx2.Fee

			balanceA += ftx2.Amount
		}
	}

	fundsStateChange(funds)

	if accA.Balance != balanceA || accB.Balance != balanceB {
		t.Errorf("State update failed: %v != %v or %v != %v\n", accA.Balance, balanceA, accB.Balance, balanceB)
	}

	collectTxFees(nil, funds, nil, nil, validatorAcc.Address)
	if feeA+feeB != validatorAcc.Balance-minerBal {
		t.Error("Fee Collection failed!")
	}

	t.Log(activeParameters)
	balBeforeRew := validatorAcc.Balance
	collectBlockReward(activeParameters.Block_reward, validatorAcc.Address)
	if validatorAcc.Balance != balBeforeRew+activeParameters.Block_reward {
		t.Error("Block reward collection failed!")
	}
}

func TestAccountOverflow(t *testing.T) {
	cleanAndPrepare()

	var accSlice []*protocol.FundsTx

	accA.Balance = MAX_MONEY
	accA.TxCnt = 0
	tx, err := protocol.ConstrFundsTx(0x01, 1, 1, 0, accB.Address, accA.Address, PrivKeyAccB, PrivKeyMultiSig, nil)
	if !verifyFundsTx(tx) || err != nil {
		t.Error("Failed to create reasonable fundsTx\n")
		return
	}
	accSlice = append(accSlice, tx)
	err = fundsStateChange(accSlice)

	//Err shouldn't be nil, because the tx can't have been successful
	//Also, the balance of A shouldn't have changed
	if err == nil || accA.Balance != MAX_MONEY {
		t.Errorf("Failed to block overflowing transaction to account with balance: %v\n", accA.Balance)
	}
}

func TestAccTxNewAccsStateChange(t *testing.T) {
	cleanAndPrepare()

	randVar := rand.New(rand.NewSource(time.Now().Unix()))

	var testSize uint32
	testSize = 1000

	var accTxs []*protocol.AccTx

	loopMax := int(randVar.Uint32()%testSize) + 1
	for i := 0; i < loopMax; i++ {
		address := [64]byte{}
		rand.Read(address[:])

		tx, _, _ := protocol.ConstrAccTx(0, randVar.Uint64()%1000, address, PrivKeyRoot, nil, nil)
		accTxs = append(accTxs, tx)
	}

	accStateChange(accTxs)

	for _, accTx := range accTxs {
		acc := storage.State[accTx.PubKey]
		//make sure the previously created acc is in the state
		if acc == nil {
			t.Errorf("Account State failed to update for the following account: %v\n", acc)
		}
	}
}

func TestFundsTxNewAccsStateChange(t *testing.T) {
	cleanAndPrepare()

	randVar := rand.New(rand.NewSource(time.Now().Unix()))

	var testSize uint32
	testSize = 1000

	var fundsTxs []*protocol.FundsTx

	loopMax := int(randVar.Uint32()%testSize) + 1
	for i := 0; i < loopMax; i++ {
		fromAddress := [64]byte{}
		toAddress := [64]byte{}

		rand.Read(fromAddress[:])
		rand.Read(toAddress[:])

		tx, _ := protocol.ConstrFundsTx(0, 0, 0, 0, fromAddress, toAddress, PrivKeyRoot, PrivKeyMultiSig, nil)
		fundsTxs = append(fundsTxs, tx)
	}

	err := fundsStateChange(fundsTxs)
	if err != nil {
		t.Error(err)
	}

	for _, fundsTx := range fundsTxs {
		fromAcc := storage.State[fundsTx.From]
		toAcc := storage.State[fundsTx.To]
		//make sure the previously created acc is in the state
		if fromAcc == nil {
			t.Errorf("Account State failed to update for the following account: %v\n", fromAcc)
		}

		if toAcc == nil {
			t.Errorf("Account State failed to update for the following account: %v\n", toAcc)
		}
	}
}

func TestConfigTxStateChange(t *testing.T) {
	cleanAndPrepare()

	randVar := rand.New(rand.NewSource(time.Now().Unix()))
	var testSize uint32
	testSize = 1000
	var configs []*protocol.ConfigTx

	loopMax := int(randVar.Uint32()%testSize) + 1
	for i := 0; i < loopMax; i++ {
		tx, err := protocol.ConstrConfigTx(uint8(randVar.Uint32()%256), uint8(randVar.Uint32()%5+1), randVar.Uint64()%10000000, randVar.Uint64(), uint8(i), PrivKeyRoot)
		if err != nil {
			t.Errorf("ConfigTx Creation failed (%v)\n", err)
		}
		if verifyConfigTx(tx) {
			configs = append(configs, tx)
		}
	}

	parameterSet := *activeParameters
	tmpLen := len(parameterSlice)
	configStateChange(configs, [32]byte{'0', '1'})
	parameterSet2 := *activeParameters
	if tmpLen != len(parameterSlice)-1 || reflect.DeepEqual(parameterSet, parameterSet2) {
		t.Errorf("Config State Change malfunctioned: %v != %v\n", tmpLen, len(parameterSlice)-1)
	}

	cleanAndPrepare()
	var configs2 []*protocol.ConfigTx
	//test the inner workings of configStateChange as well...
	tx, _ := protocol.ConstrConfigTx(uint8(randVar.Uint32()%256), 1, 1000, randVar.Uint64(), 0, PrivKeyRoot)
	tx2, _ := protocol.ConstrConfigTx(uint8(randVar.Uint32()%256), 2, 2000, randVar.Uint64(), 0, PrivKeyRoot)
	tx3, _ := protocol.ConstrConfigTx(uint8(randVar.Uint32()%256), 3, 3000, randVar.Uint64(), 0, PrivKeyRoot)
	tx4, _ := protocol.ConstrConfigTx(uint8(randVar.Uint32()%256), 4, 4000, randVar.Uint64(), 0, PrivKeyRoot)
	tx5, _ := protocol.ConstrConfigTx(uint8(randVar.Uint32()%256), 5, 5000, randVar.Uint64(), 0, PrivKeyRoot)
	tx6, _ := protocol.ConstrConfigTx(uint8(randVar.Uint32()%256), 6, 6000, randVar.Uint64(), 0, PrivKeyRoot)
	tx7, _ := protocol.ConstrConfigTx(uint8(randVar.Uint32()%256), 7, 7, randVar.Uint64(), 0, PrivKeyRoot)
	tx8, _ := protocol.ConstrConfigTx(uint8(randVar.Uint32()%256), 8, 8, randVar.Uint64(), 0, PrivKeyRoot)
	tx9, _ := protocol.ConstrConfigTx(uint8(randVar.Uint32()%256), 9, 9000, randVar.Uint64(), 0, PrivKeyRoot)
	tx10, _ := protocol.ConstrConfigTx(uint8(randVar.Uint32()%256), 10, 10000, randVar.Uint64(), 0, PrivKeyRoot)

	configs2 = append(configs2, tx)
	configs2 = append(configs2, tx2)
	configs2 = append(configs2, tx3)
	configs2 = append(configs2, tx4)
	configs2 = append(configs2, tx5)
	configs2 = append(configs2, tx6)
	configs2 = append(configs2, tx7)
	configs2 = append(configs2, tx8)
	configs2 = append(configs2, tx9)
	configs2 = append(configs2, tx10)

	configStateChange(configs2, [32]byte{})
	if activeParameters.Block_size != 1000 ||
		activeParameters.Diff_interval != 2000 ||
		activeParameters.Fee_minimum != 3000 ||
		activeParameters.Block_interval != 4000 ||
		activeParameters.Block_reward != 5000 ||
		activeParameters.Staking_minimum != 6000 ||
		activeParameters.Waiting_minimum != 7 ||
		activeParameters.Accepted_time_diff != 8 ||
		activeParameters.Slashing_window_size != 9000 ||
		activeParameters.Slash_reward != 10000 {
		t.Error("Config StateChanged didn't set the correct parameters!", activeParameters)
	}
}

//If we parse configTxs which are unknown, we don't change parameter datastructure
func TestConfigTxStateChangeUnknown(t *testing.T) {
	cleanAndPrepare()

	//Issuing configTxs with unknown Id
	var configs []*protocol.ConfigTx
	tx, _ := protocol.ConstrConfigTx(uint8(rand.Uint32()%256), 11, 1000, rand.Uint64(), 0, PrivKeyRoot)
	tx2, _ := protocol.ConstrConfigTx(uint8(rand.Uint32()%256), 11, 2000, rand.Uint64(), 0, PrivKeyRoot)
	tx3, _ := protocol.ConstrConfigTx(uint8(rand.Uint32()%256), 11, 3000, rand.Uint64(), 0, PrivKeyRoot)

	//save parameter state
	tmpParameter := parameterSlice[len(parameterSlice)-1]

	configs = append(configs, tx)
	configs = append(configs, tx2)
	configs = append(configs, tx3)

	configStateChange(configs, [32]byte{'0', '1'})

	if !reflect.DeepEqual(tmpParameter, *activeParameters) {
		t.Error("Parameter state changed even though it shouldn't have.")
	}

	configStateChangeRollback(configs, [32]byte{'0', '1'})

	if !reflect.DeepEqual(tmpParameter, *activeParameters) {
		t.Error("Parameter state changed even though it shouldn't have.")
	}

	//Adding a tx that changes state
	tx4, _ := protocol.ConstrConfigTx(uint8(rand.Uint32()%256), 2, 3000, rand.Uint64(), 0, PrivKeyRoot)
	configs = append(configs, tx4)

	configStateChange(configs, [32]byte{'0', '1'})

	if reflect.DeepEqual(tmpParameter, *activeParameters) {
		t.Error("Parameter state changed even though it shouldn't have.")
	}

	configStateChangeRollback(configs, [32]byte{'0', '1'})

	if !reflect.DeepEqual(tmpParameter, *activeParameters) {
		t.Error("Parameter state changed even though it shouldn't have.")
	}

	configStateChange(configs, [32]byte{'0', '1'})
	configStateChangeRollback(configs, [32]byte{'0'})
	//Only change if block hashes match
	if reflect.DeepEqual(tmpParameter, *activeParameters) {
		t.Error("Parameter state changed even though it shouldn't have.")
	}
}

//Testing state change, rollback and fee collection
func TestStakeTxStateChange(t *testing.T) {
	cleanAndPrepare()

	randVar := rand.New(rand.NewSource(time.Now().Unix()))

	b := newBlock([32]byte{}, [crypto.COMM_PROOF_LENGTH]byte{}, 1)
	var stake, stake2 []*protocol.StakeTx

	accA.IsStaking = false
	stakingA := accA.IsStaking

	stx, _ := protocol.ConstrStakeTx(0x01, randVar.Uint64()%100+1, true, accA.Address, PrivKeyAccA, &CommPrivKeyAccA.PublicKey)
	if addTx(b, stx) == nil {
		stakingA = true
		stake = append(stake, stx)
	}
	stakeStateChange(stake, 0)

	if accA.IsStaking != stakingA {
		t.Errorf("State update failed: %v != %v", accA.IsStaking, stakingA)
	}

	stx2, _ := protocol.ConstrStakeTx(0x01, randVar.Uint64()%100+1, false, accA.Address, PrivKeyAccA, &CommPrivKeyAccA.PublicKey)
	if addTx(b, stx) == nil {
		stakingA = false
		stake2 = append(stake2, stx2)
	}
	stakeStateChange(stake2, 0)

	if accA.IsStaking != stakingA {
		t.Errorf("State update failed: %v != %v", accA.IsStaking, stakingA)
	}

}
