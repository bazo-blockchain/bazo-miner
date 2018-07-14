package protocol

import (
	"fmt"
	"encoding/binary"
	"github.com/bazo-blockchain/bazo-miner/conf"
)

const (
	CONSOLIDATIONTX_SIZE = 50
	PARAMETERS_SIZE = 11*8 + 32
	CONS_ACCOUNT_SIZE = 97
)

type ConsolidatedAccount struct {
	Account [32]byte
	Balance uint64
	TxCnt   uint32
	Staking bool
}
type StateAccounts map[[32]byte]*ConsolidatedAccount

//when we broadcast transactions we need a way to distinguish with a type
type ConsolidationTx struct {
	// Header
	Header     byte
	Fee        uint64
	LastBlock  [32]byte
	// TODO: length of the body should be included somewhere because I need to know how much body i need to read

	// Body
	Accounts []ConsolidatedAccount
	ActiveParameters conf.Parameters
}

func ConstrConsolidationTx(header byte, state StateAccounts, lastBlockHash [32]byte, params conf.Parameters) (tx *ConsolidationTx, err error) {
	tx = new(ConsolidationTx)
	tx.Header = header
	tx.LastBlock = lastBlockHash
	tx.Fee = 1
	totalBalance := uint64(0)
	for hash, cons := range state {
		consAccount := new(ConsolidatedAccount)
		consAccount.Account = hash
		consAccount.TxCnt = cons.TxCnt
		consAccount.Balance = cons.Balance
		consAccount.Staking = cons.Staking
		totalBalance += cons.Balance
		tx.Accounts = append(tx.Accounts, *consAccount)
	}
	tx.ActiveParameters = params
	return tx, nil
}


func (tx *ConsolidationTx) encodeActiveParameters() (encodedTx []byte) {
	encodedTx = make([]byte, PARAMETERS_SIZE)
	offset := 0
	params := tx.ActiveParameters
	var fee_minimum, block_size, diff_interval, block_interval [8]byte
	var staking_minimum, waiting_minimum, accepted_time_diff, slashing_window_size [8]byte
	var slash_reward, num_included_prev_seeds, consolidation_interval  [8]byte
	binary.BigEndian.PutUint64(fee_minimum[:], uint64(params.Fee_minimum))
	binary.BigEndian.PutUint64(block_size[:], uint64(params.Block_size))
	binary.BigEndian.PutUint64(diff_interval[:], uint64(params.Diff_interval))
	binary.BigEndian.PutUint64(block_interval[:], uint64(params.Block_interval))
	binary.BigEndian.PutUint64(staking_minimum[:], uint64(params.Staking_minimum))
	binary.BigEndian.PutUint64(waiting_minimum[:], uint64(params.Waiting_minimum))
	binary.BigEndian.PutUint64(accepted_time_diff[:], uint64(params.Accepted_time_diff))
	binary.BigEndian.PutUint64(slashing_window_size[:], uint64(params.Slashing_window_size))
	binary.BigEndian.PutUint64(slash_reward[:], uint64(params.Slash_reward))
	binary.BigEndian.PutUint64(num_included_prev_seeds[:], uint64(params.Num_included_prev_seeds))
	binary.BigEndian.PutUint64(consolidation_interval[:], uint64(params.Consolidation_interval))
	copy(encodedTx[offset:offset+32], params.BlockHash[:])
	offset += 32
	copy(encodedTx[offset:offset+8], fee_minimum[:])
	offset += 8
	copy(encodedTx[offset:offset+8], block_size[:])
	offset += 8
	copy(encodedTx[offset:offset+8], diff_interval[:])
	offset += 8
	copy(encodedTx[offset:offset+8], block_interval[:])
	offset += 8
	copy(encodedTx[offset:offset+8], staking_minimum[:])
	offset += 8
	copy(encodedTx[offset:offset+8], waiting_minimum[:])
	offset += 8
	copy(encodedTx[offset:offset+8], accepted_time_diff[:])
	offset += 8
	copy(encodedTx[offset:offset+8], slashing_window_size[:])
	offset += 8
	copy(encodedTx[offset:offset+8], slash_reward[:])
	offset += 8
	copy(encodedTx[offset:offset+8], num_included_prev_seeds[:])
	offset += 8
	copy(encodedTx[offset:offset+8], consolidation_interval[:])

	return encodedTx
}


func (tx *ConsolidationTx) decodeActiveParameters(encodedTx []byte) {
	offset := 0
	copy(tx.ActiveParameters.BlockHash[:], encodedTx[offset:offset+32])
	offset += 32
	tx.ActiveParameters.Fee_minimum = binary.BigEndian.Uint64(encodedTx[offset:offset+8])
	offset += 8
	tx.ActiveParameters.Block_size = binary.BigEndian.Uint64(encodedTx[offset:offset+8])
	offset += 8
	tx.ActiveParameters.Diff_interval = binary.BigEndian.Uint64(encodedTx[offset:offset+8])
	offset += 8
	tx.ActiveParameters.Staking_minimum = binary.BigEndian.Uint64(encodedTx[offset:offset+8])
	offset += 8
	tx.ActiveParameters.Waiting_minimum = binary.BigEndian.Uint64(encodedTx[offset:offset+8])
	offset += 8
	tx.ActiveParameters.Accepted_time_diff = binary.BigEndian.Uint64(encodedTx[offset:offset+8])
	offset += 8
	tx.ActiveParameters.Slashing_window_size = binary.BigEndian.Uint64(encodedTx[offset:offset+8])
	offset += 8
	tx.ActiveParameters.Slash_reward = binary.BigEndian.Uint64(encodedTx[offset:offset+8])
	offset += 8
	tx.ActiveParameters.Num_included_prev_seeds = int(binary.BigEndian.Uint64(encodedTx[offset:offset+8]))
	offset += 8
	tx.ActiveParameters.Consolidation_interval = binary.BigEndian.Uint64(encodedTx[offset:offset+8])
}

// TODO: calculate proper hash
func (tx *ConsolidationTx) Hash() (hash [32]byte) {
	if tx == nil {
		return [32]byte{}
	}

	txHash := struct {
		Header byte
		LastBlock   [32]byte
	}{
		tx.Header,
		tx.LastBlock,
	}

	return SerializeHashContent(txHash)
}

//when we serialize the struct with binary.Write, unexported field get serialized as well, undesired
//behavior. Therefore, writing own encoder/decoder
func (tx *ConsolidationTx) Encode() (encodedTx []byte) {
	if tx == nil {
		return nil
	}
	var fee, numberAccounts, balance, txCnt [8]byte
	var isStaking byte

	binary.BigEndian.PutUint64(fee[:], tx.Fee)
	binary.BigEndian.PutUint64(numberAccounts[:], uint64(len(tx.Accounts)))

	encodedTx = make([]byte, tx.Size())

	encodedTx[0] = tx.Header
	copy(encodedTx[1:9], fee[:])
	copy(encodedTx[9:41], tx.LastBlock[:])
	copy(encodedTx[41:49], numberAccounts[:])
	encodedTx[49] = 1  // TODO: always include config parameters?
	encodedParams := tx.encodeActiveParameters()
	offset := 50
	copy(encodedTx[offset:offset+PARAMETERS_SIZE], encodedParams[:])
	offset = CONSOLIDATIONTX_SIZE+PARAMETERS_SIZE

	fmt.Printf("encode last block %v\n ", tx.LastBlock)
	fmt.Printf("encode numaccounts %v\n ", len(tx.Accounts))
	for i := 0; i < len(tx.Accounts); i++ {
		acc := tx.Accounts[i]
		if acc.Staking == true {
			isStaking = 1
		} else {
			isStaking = 0
		}
		fmt.Printf("encode acc %v address %v\n ", i, acc.Account)
		fmt.Printf("encode acc %v staking %v\n ", i, acc.Staking)
		fmt.Printf("encode acc %v balance %v\n ", i, acc.Balance)
		fmt.Printf("encode acc %v txcnt %v\n ", i, acc.TxCnt)
		offset = CONSOLIDATIONTX_SIZE+PARAMETERS_SIZE +i*CONS_ACCOUNT_SIZE
		copy(encodedTx[offset:offset+32], acc.Account[:])
		offset += 32
		encodedTx[offset] = isStaking
		offset += 1
		binary.BigEndian.PutUint64(balance[:], uint64(acc.Balance))
		copy(encodedTx[offset:offset+32], balance[:])
		offset += 32
		binary.BigEndian.PutUint32(txCnt[:], uint32(acc.TxCnt))
		copy(encodedTx[offset:offset+32], txCnt[:])
	}
	return encodedTx
}

func (*ConsolidationTx) Decode(encodedTx []byte) (tx *ConsolidationTx) {
	var numAccounts int
	tx = new(ConsolidationTx)
	tx.Header = encodedTx[0]
	tx.Fee = binary.BigEndian.Uint64(encodedTx[1:9])
	copy(tx.LastBlock[:], encodedTx[9:41])

	numAccounts = int(binary.BigEndian.Uint64(encodedTx[41:49]))
	activeParams := encodedTx[49] == 1
	tx.decodeActiveParameters(encodedTx[CONSOLIDATIONTX_SIZE:CONSOLIDATIONTX_SIZE+PARAMETERS_SIZE])
	offset := PARAMETERS_SIZE + CONSOLIDATIONTX_SIZE

	fmt.Printf("decode last block %v\n ", tx.LastBlock)
	fmt.Printf("decode active params %v\n ", activeParams)
	fmt.Printf("decode numaccounts %v\n ", numAccounts)
	fmt.Printf("decode active params %v\n ", activeParams)
	for i := 0; i < numAccounts; i++ {
		offset =  PARAMETERS_SIZE + CONSOLIDATIONTX_SIZE+CONS_ACCOUNT_SIZE*i
		consAccount := new(ConsolidatedAccount)
		copy(consAccount.Account[:], encodedTx[offset:offset+32])
		offset += 32
		isStakingAsByte := encodedTx[offset]
		consAccount.Staking = isStakingAsByte == 1
		offset += 1
		consAccount.Balance = binary.BigEndian.Uint64(encodedTx[offset:offset+32])
		offset += 32
		consAccount.TxCnt = binary.BigEndian.Uint32(encodedTx[offset:offset+32])
		tx.Accounts = append(tx.Accounts, *consAccount)
		fmt.Printf("decode acc %v address %v\n ", i, consAccount.Account)
		fmt.Printf("decode acc %v staking %v\n ", i, consAccount.Staking)
		fmt.Printf("decode acc %v balance %v\n ", i, consAccount.Balance)
		fmt.Printf("decode acc %v txcount %v\n ", i, consAccount.TxCnt)
	}

	return tx
}

func (tx *ConsolidationTx) TxFee() uint64 { return tx.Fee }
func (tx *ConsolidationTx) Size() uint64 {
	return CONSOLIDATIONTX_SIZE + PARAMETERS_SIZE+CONS_ACCOUNT_SIZE*uint64(len(tx.Accounts))
}

func (tx ConsolidationTx) String() string {
	status := fmt.Sprintf(
		"\ntxhash: %v\n" +
		"Header: %v\n" +
		"LastBlockHash: %v\n" +
		"ActiveParameters: %v\n" +
		"Consolidated accounts: %v\n",
		tx.Hash(),
		tx.Header,
		tx.LastBlock,
		tx.ActiveParameters,
		len(tx.Accounts),
	)
	mapping := ""
	for _, cons := range tx.Accounts {
		mapping += fmt.Sprintf("ConsolidatedAccount: %v\n", cons)
	}

	return fmt.Sprintf("%v\n%v", status, mapping)
}
