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

func NewVMContext(from [32]byte, data  []byte, account protocol.Account){
	c := Context{}
	c.transactionSender = from
	c.transactionData = data
	c.account = account
	c.changes = []Change{}
}

func (c * Context) GetTransactionSender() [32]byte {

}

func (c * Context) GetTransactionData() []byte {

}

func (c * Context) GetContractVariable(index int) big.Int {

}

func (c * Context) SetContractVariable(index int, value big.Int){

}

func (c * Context) GetContract(){

}

func (c * Context) GetBalance(){

}


