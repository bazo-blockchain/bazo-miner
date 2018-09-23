package miner

import (
	"math/rand"
	"reflect"
	"testing"
	"time"

	"github.com/bazo-blockchain/bazo-miner/protocol"
	"github.com/bazo-blockchain/bazo-miner/storage"
)

//Rollback tests for all tx types
func TestFundsStateChangeRollback(t *testing.T) {
	cleanAndPrepare()

	randVar := rand.New(rand.NewSource(time.Now().Unix()))

	accAHash := protocol.SerializeHashContent(accA.Address)
	accBHash := protocol.SerializeHashContent(accB.Address)
	minerAccHash := protocol.SerializeHashContent(validatorAcc.Address)

	var testSize uint32
	testSize = 1000

	b := newBlock([32]byte{}, [32]byte{}, [32]byte{}, 1)
	var funds []*protocol.FundsTx

	var feeA, feeB uint64

	//State snapshot
	rollBackA := accA.Balance
	rollBackB := accB.Balance

	//Record transaction amounts in this variables
	balanceA := accA.Balance
	balanceB := accB.Balance

	loopMax := int(randVar.Uint32()%testSize + 1)
	for i := 0; i < loopMax+1; i++ {
		ftx, _ := protocol.ConstrFundsTx(0x01, randVar.Uint64()%1000000+1, randVar.Uint64()%100+1, uint32(i), accAHash, accBHash, &PrivKeyAccA, &PrivKeyMultiSig, nil)
		if addTx(b, ftx) == nil {
			funds = append(funds, ftx)
			balanceA -= ftx.Amount
			feeA += ftx.Fee

			balanceB += ftx.Amount
		} else {
			t.Errorf("Block rejected a valid transaction: %v\n", ftx)
		}

		ftx2, _ := protocol.ConstrFundsTx(0x01, randVar.Uint64()%1000+1, randVar.Uint64()%100+1, uint32(i), accBHash, accAHash, &PrivKeyAccB, &PrivKeyMultiSig, nil)
		if addTx(b, ftx2) == nil {
			funds = append(funds, ftx2)
			balanceB -= ftx2.Amount
			feeB += ftx2.Fee

			balanceA += ftx2.Amount
		} else {
			t.Errorf("Block rejected a valid transaction: %v\n", ftx2)
		}
	}
	fundsStateChange(funds)
	if accA.Balance != balanceA || accB.Balance != balanceB {
		t.Error("State update failed!")
	}
	fundsStateChangeRollback(funds)
	if accA.Balance != rollBackA || accB.Balance != rollBackB {
		t.Error("Rollback failed!")
	}

	//collectTxFees is checked below in its own test (to additionally cover overflow scenario)
	balBeforeRew := validatorAcc.Balance
	reward := 5
	collectBlockReward(uint64(reward), minerAccHash)
	if validatorAcc.Balance != balBeforeRew+uint64(reward) {
		t.Error("Block reward collection failed!")
	}
	collectBlockRewardRollback(uint64(reward), minerAccHash)
	if validatorAcc.Balance != balBeforeRew {
		t.Error("Block reward collection rollback failed!")
	}
}

func TestAccStateChangeRollback(t *testing.T) {
	cleanAndPrepare()

	randVar := rand.New(rand.NewSource(time.Now().Unix()))

	var testSize uint32
	testSize = 1000

	var accs []*protocol.AccTx

	//Store accs that are to be changed and rolled back in a accTx slice
	nullAddress := [64]byte{}
	loopMax := int(randVar.Uint32()%testSize) + 1
	for i := 0; i < loopMax; i++ {
		tx, _, _ := protocol.ConstrAccTx(0, randVar.Uint64()%1000, nullAddress, &PrivKeyRoot, nil, nil)
		accs = append(accs, tx)
	}

	accStateChange(accs)

	for _, acc := range accs {
		accHash := protocol.SerializeHashContent(acc.PubKey)
		acc := storage.State[accHash]
		if acc == nil {
			t.Errorf("Account State failed to update for the following account: %v\n", acc)
		}
	}

	accStateChangeRollback(accs)

	for _, acc := range accs {
		accHash := protocol.SerializeHashContent(acc.PubKey)
		acc := storage.State[accHash]
		if acc != nil {
			t.Errorf("Account State failed to rollback the following account: %v\n", acc)
		}
	}
}

func TestConfigStateChangeRollback(t *testing.T) {
	cleanAndPrepare()

	var configSlice []*protocol.ConfigTx

	tx, _ := protocol.ConstrConfigTx(uint8(rand.Uint32()%256), 1, 1000, rand.Uint64(), 0, &PrivKeyRoot)
	tx2, _ := protocol.ConstrConfigTx(uint8(rand.Uint32()%256), 2, 2000, rand.Uint64(), 0, &PrivKeyRoot)
	tx3, _ := protocol.ConstrConfigTx(uint8(rand.Uint32()%256), 3, 3000, rand.Uint64(), 0, &PrivKeyRoot)
	tx4, _ := protocol.ConstrConfigTx(uint8(rand.Uint32()%256), 4, 4000, rand.Uint64(), 0, &PrivKeyRoot)
	tx5, _ := protocol.ConstrConfigTx(uint8(rand.Uint32()%256), 5, 5000, rand.Uint64(), 0, &PrivKeyRoot)

	configSlice = append(configSlice, tx)
	configSlice = append(configSlice, tx2)
	configSlice = append(configSlice, tx3)
	configSlice = append(configSlice, tx4)
	configSlice = append(configSlice, tx5)

	before := *activeParameters
	configStateChange(configSlice, [32]byte{'0', '1', '2'})
	if reflect.DeepEqual(before, *activeParameters) {
		t.Error("No config state change.")
	}
	configStateChangeRollback(configSlice, [32]byte{'0', '1', '2'})
	if !reflect.DeepEqual(before, *activeParameters) {
		t.Error("Config state rollback failed.")
	}
}

func TestCollectTxFeesRollback(t *testing.T) {
	cleanAndPrepare()

	randVar := rand.New(rand.NewSource(time.Now().Unix()))

	var funds, funds2 []*protocol.FundsTx

	accAHash := protocol.SerializeHashContent(accA.Address)
	accBHash := protocol.SerializeHashContent(accB.Address)
	minerHash := protocol.SerializeHashContent(validatorAcc.Address)

	minerBal := validatorAcc.Balance
	//Rollback everything
	var fee uint64
	loopMax := int(randVar.Uint64() % 1000)
	for i := 0; i < loopMax+1; i++ {
		tx, _ := protocol.ConstrFundsTx(0x01, randVar.Uint64()%1000000+1, randVar.Uint64()%100+1, uint32(i), accAHash, accBHash, &PrivKeyAccA, nil, nil)

		funds = append(funds, tx)
		fee += tx.Fee
	}

	collectTxFees(nil, funds, nil, nil, minerHash)
	if minerBal+fee != validatorAcc.Balance {
		t.Errorf("%v + %v != %v\n", minerBal, fee, validatorAcc.Balance)
	}
	collectTxFeesRollback(nil, funds, nil, nil, minerHash)
	if minerBal != validatorAcc.Balance {
		t.Errorf("Tx fees rollback failed: %v != %v\n", minerBal, validatorAcc.Balance)
	}

	validatorAcc.Balance = MAX_MONEY - 100
	var fee2 uint64
	minerBal = validatorAcc.Balance
	//Miner gets fees, the miner account balance will overflow at some point
	for i := 2; i < 100; i++ {
		tx, _ := protocol.ConstrFundsTx(0x01, randVar.Uint64()%1000000+1, uint64(i), uint32(i), accAHash, accBHash, &PrivKeyAccA, nil, nil)
		funds2 = append(funds2, tx)
		fee2 += tx.Fee
	}

	accABal := accA.Balance
	accBBal := accB.Balance
	//Should throw an error and result in a rollback, because of acc balance overflow
	tmpBlock := newBlock([32]byte{}, [32]byte{}, [32]byte{}, 1)
	tmpBlock.Beneficiary = minerHash
	data := blockData{nil, funds2, nil, nil, tmpBlock}
	if err := validateState(data); err == nil ||
		minerBal != validatorAcc.Balance ||
		accA.Balance != accABal ||
		accB.Balance != accBBal {
		t.Errorf("No rollback resulted, %v != %v\n", minerBal, validatorAcc.Balance)
	}
}
