package protocol

import (
	"fmt"
)

const (
	CONSOLIDATIONTX_SIZE = 213
)

type ConsolidatedAccount struct {
	Account [32]byte
	Balance uint64
	Staking bool
}
type StateAccounts map[[32]byte]*ConsolidatedAccount

//when we broadcast transactions we need a way to distinguish with a type
type ConsolidationTx struct {
	// Header
	Header    byte
	LastBlock [32]byte
	NumAccounts int // not needed
	// TODO: length of the body should be included somewhere because I need to know how much body i need to read

	// Body
	Accounts []ConsolidatedAccount
}

func ConstrConsolidationTx(header byte, state StateAccounts, lastBlockHash [32]byte) (tx *ConsolidationTx, err error) {
	tx = new(ConsolidationTx)
	tx.Header = header
	tx.LastBlock = lastBlockHash

	tx.NumAccounts = len(state)
	totalBalance := uint64(0)
	for hash, cons := range state {
		consAccount := new(ConsolidatedAccount)
		consAccount.Account = hash
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
		NumAccounts int

	}{
		tx.Header,
		tx.NumAccounts,
	}

	return SerializeHashContent(txHash)
}

//when we serialize the struct with binary.Write, unexported field get serialized as well, undesired
//behavior. Therefore, writing own encoder/decoder
func (tx *ConsolidationTx) Encode() (encodedTx []byte) {
	if tx == nil {
		return nil
	}

	encodedTx = make([]byte, CONSOLIDATIONTX_SIZE)
	encodedTx[0] = tx.Header

	return encodedTx
}

func (*ConsolidationTx) Decode(encodedTx []byte) (tx *ConsolidationTx) {
	tx = new(ConsolidationTx)

	tx.Header = encodedTx[0]

	return tx
}

func (tx *ConsolidationTx) Size() uint64  { return CONSOLIDATIONTX_SIZE }

func (tx ConsolidationTx) String() string {
	status := fmt.Sprintf(
		"\nHeader: %v\n" +
		"lastBlockHash: %v\n" +
		"numAccounts: %v\n",
		tx.Header,
		tx.LastBlock,
		tx.NumAccounts,
	)
	mapping := ""
	for _, cons := range tx.Accounts {
		mapping += fmt.Sprintf("ConsolidatedAccount: %v\n", cons)
	}

	return fmt.Sprintf("%v\n%v", status, mapping)
}
