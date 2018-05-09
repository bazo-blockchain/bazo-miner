package protocol

import (
	"github.com/bazo-blockchain/bazo-miner/protocol"
	"math/big"
)


type Context struct {
	transactionSender [32]byte
	transactionData   []byte
	account protocol.Account
	changes []Change
}

type Change struct {
	index int
	value big.Int
}

func NewChange(index int, value big.Int) Change{
	return Change{index, value}
}

func (c * Change) GetChange() (int, big.Int){
	return c.index, c.value
}

func NewVMContext(from [32]byte, data  []byte, account protocol.Account){
	c := Context{}
	c.transactionSender = from
	c.transactionData = data
	c.account = account
	c.changes = []Change{}
}

func (c * Context) GetContract(){

}

func (c * Context) GetContractVariable(index int) big.Int {
	return big.Int{}
}

func (c * Context) GetTransactionSender() [32]byte {
	return [32]byte{}
}

func (c * Context) GetTransactionData() []byte {
	return []byte{}
}

func (c * Context) SetContractVariable(index int, value big.Int){

}



func (c * Context) GetBalance(){

}


