package vm

type StateData struct {
	data []byte
}

type Context struct {
	TransactionSender [32]byte
	TransactionData   []byte
	MaxGasAmount      uint64
	ContractAccount   ContractAccount
	ContractTx        ContractTx
	/*
		stateData StateData


		blockHeader []byte*/
}

func NewContext() *Context {
	//data := map[int][]byte{}

	return &Context{
		TransactionSender: [32]byte{},
		TransactionData:   []byte{},
		MaxGasAmount:      100000,
		ContractAccount:   ContractAccount{},
		ContractTx:        ContractTx{},
	}
}
