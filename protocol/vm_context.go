package protocol

import (
	"errors"
	"math/big"
)

type Context struct {
	Account
	changes []Change
	FundsTx
}

type Change struct {
	index int
	value big.Int
}

func NewChange(index int, value big.Int) Change {
	return Change{index, value}
}

func (c *Change) GetChange() (int, big.Int) {
	return c.index, c.value
}

func NewContext(account Account, fundsTx FundsTx) *Context {
	newContext := Context{
		Account: account,
		changes: []Change{},
		FundsTx: fundsTx,
	}
	return &newContext
}

func (c *Context) GetContract() []byte {
	return c.Contract
}

func (c *Context) GetContractVariable(index int) (big.Int, error) {
	if index >= len(c.ContractVariables) {
		return big.Int{}, errors.New("Index out of bounds")
	}
	return c.ContractVariables[index], nil
}

func (c *Context) SetContractVariable(index int, value big.Int) error {
	if len(c.ContractVariables) <= index {
		return errors.New("Index out of bounds")
	}
	change := NewChange(index, value)
	c.changes = append(c.changes, change)
	return nil
}

func (c *Context) PersistChanges() {
	for _, change := range c.changes {
		i, value := change.GetChange()
		c.ContractVariables[i] = value
	}
}

func (c *Context) GetAddress() [64]byte {
	return c.Address
}

func (c *Context) GetIssuer() [32]byte {
	return c.Issuer
}

func (c *Context) GetBalance() uint64 {
	return c.Balance
}

func (c *Context) GetSender() [32]byte {
	return c.From
}

func (c *Context) GetAmount() uint64 {
	return c.Amount
}

func (c *Context) GetTransactionData() []byte {
	return c.Data
}

func (c *Context) GetFee() uint64 {
	return c.Fee
}

func (c *Context) GetSig1() [64]byte {
	return c.Sig1
}
