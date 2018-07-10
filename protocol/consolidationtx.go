package protocol

import (
	"fmt"
	"encoding/binary"
)

const (
	CONSOLIDATIONTX_SIZE = 49
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
}

func ConstrConsolidationTx(header byte, state StateAccounts, lastBlockHash [32]byte) (tx *ConsolidationTx, err error) {
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
	return tx, nil
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
	binary.BigEndian.PutUint64(fee[:], tx.Fee)
	binary.BigEndian.PutUint64(numberAccounts[:], uint64(len(tx.Accounts)))

	encodedTx = make([]byte, CONSOLIDATIONTX_SIZE+CONS_ACCOUNT_SIZE*len(tx.Accounts))

	encodedTx[0] = tx.Header
	copy(encodedTx[1:9], fee[:])
	copy(encodedTx[9:41], tx.LastBlock[:])
	copy(encodedTx[41:49], numberAccounts[:])
	var isStaking byte
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
		offset := 49 + i*(CONS_ACCOUNT_SIZE)
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
	tx = new(ConsolidationTx)
	fmt.Printf("DencodedTx %v\n ", encodedTx)
	tx.Header = encodedTx[0]
	tx.Fee = binary.BigEndian.Uint64(encodedTx[1:9])
	fmt.Printf("decode txFee %v\n ", tx.Fee)
	copy(tx.LastBlock[:], encodedTx[9:41])
	var numAccounts int
	numAccounts = int(binary.BigEndian.Uint64(encodedTx[41:49]))
	fmt.Printf("decode last block %v\n ", tx.LastBlock)
	fmt.Printf("decode numaccounts %v\n ", numAccounts)
	for i := 0; i < numAccounts; i++ {
		offset := CONSOLIDATIONTX_SIZE + CONS_ACCOUNT_SIZE*i
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
	return CONSOLIDATIONTX_SIZE + CONS_ACCOUNT_SIZE*uint64(len(tx.Accounts))
}

func (tx ConsolidationTx) String() string {
	status := fmt.Sprintf(
		"\ntxhash: %v\n" +
		"Header: %v\n" +
		"LastBlockHash: %v\n" +
		"Consolidated accounts: %v\n",
		tx.Hash(),
		tx.Header,
		tx.LastBlock,
		len(tx.Accounts),
	)
	mapping := ""
	for _, cons := range tx.Accounts {
		mapping += fmt.Sprintf("ConsolidatedAccount: %v\n", cons)
	}

	return fmt.Sprintf("%v\n%v", status, mapping)
}
