package vm

const (
	PUSH = iota
	DUP
	ROLL
	POP
	ADD
	SUB
	MULT
	DIV
	MOD
	NEG
	EQ
	NEQ
	LT
	GT
	LTE
	GTE
	SHIFTL
	SHIFTR
	NOP
	JMP
	JMPIF
	CALL
	CALLIF
	CALLEXT
	RET
	SIZE
	STORE
	SSTORE
	LOAD
	SLOAD
	ADDRESS // Address of account
	ISSUER  // Owner of smart contract account
	BALANCE // Balance of account
	CALLER
	CALLVAL  // Amount of bazo coins transacted in transaction
	CALLDATA // Parameters and function signature hash
	NEWMAP
	MAPHASKEY
	MAPPUSH
	MAPGETVAL
	MAPSETVAL
	MAPREMOVE
	NEWARR
	ARRAPPEND
	ARRINSERT
	ARRREMOVE
	ARRAT
	SHA3
	CHECKSIG
	ERRHALT
	HALT
	//	MAPCONTAINSKEY
)

const (
	BYTES = iota + 1
	BYTE
	LABEL
	ADDR
)

type OpCode struct {
	code      byte
	Name      string
	Nargs     int
	ArgTypes  []int
	gasPrice  uint64
	gasFactor uint64
}

var OpCodes = []OpCode{
	{PUSH, "push", 1, []int{BYTES}, 1, 1},
	{DUP, "dup", 0, nil, 1, 2},
	{ROLL, "roll", 1, []int{BYTE}, 1, 2},
	{POP, "pop", 0, nil, 1, 1},
	{ADD, "add", 0, nil, 1, 2},
	{SUB, "sub", 0, nil, 1, 2},
	{MULT, "mult", 0, nil, 1, 2},
	{DIV, "div", 0, nil, 1, 2},
	{MOD, "mod", 0, nil, 1, 2},
	{NEG, "neg", 0, nil, 1, 2},
	{EQ, "eq", 0, nil, 1, 2},
	{NEQ, "neq", 0, nil, 1, 2},
	{LT, "lt", 0, nil, 1, 2},
	{GT, "gt", 0, nil, 1, 2},
	{LTE, "lte", 0, nil, 1, 2},
	{GTE, "gte", 0, nil, 1, 2},
	{SHIFTL, "shiftl", 1, []int{BYTE}, 1, 2},
	{SHIFTR, "shiftl", 1, []int{BYTE}, 1, 2},
	{NOP, "nop", 0, nil, 1, 1},
	{JMP, "jmp", 1, []int{LABEL}, 1, 1},
	{JMPIF, "jmpif", 1, []int{LABEL}, 1, 1},
	{CALL, "call", 2, []int{LABEL, BYTE}, 1, 1},
	{CALLIF, "callif", 2, []int{LABEL, BYTE}, 1, 1},
	{CALLEXT, "callext", 3, []int{ADDR, BYTE, BYTE, BYTE, BYTE, BYTE}, 1000, 2},
	{RET, "ret", 0, nil, 1, 1},
	{SIZE, "size", 0, nil, 1, 1},
	{STORE, "store", 0, nil, 1, 2},
	{SSTORE, "sstore", 1, []int{BYTE}, 1000, 2},
	{LOAD, "load", 1, []int{BYTE}, 1, 2},
	{SLOAD, "sload", 1, []int{BYTE}, 10, 2},
	{ADDRESS, "address", 0, nil, 1, 1},
	{ISSUER, "issuer", 0, nil, 1, 1},
	{BALANCE, "balance", 0, nil, 1, 1},
	{CALLER, "caller", 0, nil, 1, 1},
	{CALLVAL, "callval", 0, nil, 1, 1},
	{CALLDATA, "calldata", 0, nil, 1, 1},
	{NEWMAP, "newmap", 0, nil, 1, 2},
	{MAPHASKEY, "maphaskey", 0, nil, 1, 2},
	{MAPPUSH, "mappush", 0, nil, 1, 2},
	{MAPGETVAL, "mapgetval", 0, nil, 1, 2},
	{MAPSETVAL, "mapsetval", 0, nil, 1, 2},
	{MAPREMOVE, "mapremove", 0, nil, 1, 2},
	{NEWARR, "newarr", 0, nil, 1, 2},
	{ARRAPPEND, "arrappend", 0, nil, 1, 2},
	{ARRINSERT, "arrinsert", 0, nil, 1, 2},
	{ARRREMOVE, "arrremove", 0, nil, 1, 2},
	{ARRAT, "arrat", 0, nil, 1, 2},
	{SHA3, "sha3", 0, nil, 1, 2},
	{CHECKSIG, "checksig", 0, nil, 1, 2},
	{ERRHALT, "errhalt", 0, nil, 0, 1},
	{HALT, "halt", 0, nil, 0, 1},
}
