package vm

import (
	"github.com/bazo-blockchain/bazo-miner/protocol"
)

type MockContext struct {
	protocol.Context
	transactionData []byte
}

func NewMockContext(byteCode []byte) *MockContext {
	mc := MockContext{}
	mc.SetContract(byteCode)
	mc.Fee = 50
	return &mc
}

func (mc *MockContext) SetContract(contract []byte) {
	mc.Contract = contract
}

func (mc *MockContext) GetTransactionData() []byte {
	return mc.transactionData
}
