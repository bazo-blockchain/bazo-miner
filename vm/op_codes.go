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
	CALLEXT
	RET
	SIZE
	STORE
	SSTORE
	LOAD
	SLOAD
	ADDRESS // Address of account
	BALANCE // Balance of account
	CALLER
	CALLVAL  // Amount of bazo coins transacted in transaction
	CALLDATA // Parameters and function signature hash
	NEWMAP
	MAPPUSH
	MAPGETVAL
	MAPREMOVE
	NEWARR
	ARRAPPEND
	//ARRINSERT
	ARRREMOVE
	ARRAT
	SHA3
	CHECKSIG
	ERRHALT
	HALT
)

const (
	INT = iota + 1
	BYTE
	BYTES
	LABEL
)

type OpCode struct {
	Name     string
	Nargs    int
	ArgTypes []int
	gasPrice uint64
}

var OpCodes = map[int]OpCode{
	PUSH:      {"push", 1, []int{INT}, 1},
	DUP:       {"dup", 0, []int{}, 1},
	ROLL:      {"roll", 1, []int{INT}, 1},
	POP:       {"pop", 0, []int{}, 1},
	ADD:       {"add", 0, []int{}, 1},
	SUB:       {"sub", 0, []int{}, 1},
	MULT:      {"mult", 0, []int{}, 1},
	DIV:       {"div", 0, []int{}, 1},
	MOD:       {"mod", 0, []int{}, 1},
	NEG:       {"neg", 0, []int{}, 1},
	EQ:        {"eq", 0, []int{}, 1},
	NEQ:       {"neq", 0, []int{}, 1},
	LT:        {"lt", 0, []int{}, 1},
	GT:        {"gt", 0, []int{}, 1},
	LTE:       {"lte", 0, []int{}, 1},
	GTE:       {"gte", 0, []int{}, 1},
	SHIFTL:    {"shiftl", 1, []int{BYTE}, 1},
	SHIFTR:    {"shiftl", 1, []int{BYTE}, 1},
	NOP:       {"nop", 0, []int{}, 1},
	JMP:       {"jmp", 1, []int{LABEL}, 1},
	JMPIF:     {"jmpif", 1, []int{LABEL}, 1},
	CALL:      {"call", 2, []int{LABEL, BYTE}, 1},
	CALLEXT:   {"callext", 3, []int{BYTES, BYTES, BYTE}, 1},
	RET:       {"ret", 0, []int{}, 1},
	SIZE:      {"size", 0, []int{}, 1},
	ADDRESS:   {"address", 0, []int{}, 1},
	BALANCE:   {"balance", 0, []int{}, 1},
	CALLER:    {"caller", 0, []int{}, 1},
	CALLVAL:   {"callval", 0, []int{}, 1},
	CALLDATA:  {"calldata", 0, []int{}, 1},
	STORE:     {"store", 0, []int{}, 1},
	SSTORE:    {"sstore", 1, []int{INT}, 1},
	LOAD:      {"load", 1, []int{BYTE}, 1},
	SLOAD:     {"sload", 1, []int{INT}, 1},
	NEWMAP:    {"newmap", 0, []int{}, 1},
	MAPPUSH:   {"mappush", 0, []int{}, 1},
	MAPGETVAL: {"mapgetval", 0, []int{}, 1},
	MAPREMOVE: {"mapremove", 0, []int{}, 1},
	NEWARR:    {"newarr", 0, []int{}, 1},
	ARRAPPEND: {"arrappend", 0, []int{}, 1},
	//ARRINSERT:    {"arrinsert", 0, []int{}, 1},
	ARRREMOVE: {"arrremove", 0, []int{}, 1},
	ARRAT:     {"arrat", 0, []int{}, 1},
	SHA3:      {"sha3", 0, []int{}, 1},
	CHECKSIG:  {"checksig", 0, []int{}, 1},
	HALT:      {"halt", 0, []int{}, 0},
	ERRHALT:   {"errhalt", 0, []int{}, 0},
}
