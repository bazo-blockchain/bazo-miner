package protocol

import (
	"fmt"
)

const (
	CONSOLIDATIONTX_SIZE = 213
)

type ConsolidatedAccount struct {
	account [32]byte
	balance uint64
	staking bool
}

//when we broadcast transactions we need a way to distinguish with a type
type ConsolidationTx struct {
	// Header
	Header    byte
	NumAccounts int
	TotalBalance uint64

	// Body
	Accounts [128]ConsolidatedAccount

}

func ConstrConsolidationTx(header byte, state map[[32]byte]uint64) (tx *ConsolidationTx, err error) {
	tx = new(ConsolidationTx)
	tx.Header = header
	tx.NumAccounts = len(state)
	totalBalance := uint64(0)
	for hash, balance := range state {
		consAccount := new(ConsolidatedAccount)
		consAccount.account = hash
		consAccount.balance = balance
		totalBalance += balance
	}
	tx.TotalBalance = totalBalance
	return tx, nil
}

func (tx *ConsolidationTx) Hash() (hash [32]byte) {
	if tx == nil {
		return [32]byte{}
	}

	txHash := struct {
		Header byte

	}{
		tx.Header,
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
	return fmt.Sprintf(
		"\nHeader: %v\n" +
		"numAccounts: %v\n" +
		"balance: %v\n",
		tx.Header,
		tx.NumAccounts,
		tx.TotalBalance,
	)
}
