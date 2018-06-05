package vm

import (
	"bytes"
	"encoding/binary"
	"math/big"
	"testing"

	"github.com/bazo-blockchain/bazo-miner/protocol"
)

func TestVM_NewTestVM(t *testing.T) {
	vm := NewTestVM([]byte{})

	if len(vm.code) > 0 {
		t.Errorf("Actual code length is %v, should be 0 after initialization", len(vm.code))
	}

	if vm.pc != 0 {
		t.Errorf("Actual pc counter is %v, should be 0 after initialization", vm.pc)
	}
}

func TestVM_Exec_GasConsumption(t *testing.T) {
	code := []byte{
		PUSH, 0, 8,
		PUSH, 0, 8,
		ADD,
		HALT,
	}

	vm := NewTestVM([]byte{})
	mc := NewMockContext(code)
	mc.Fee = 3
	vm.context = mc

	vm.Exec(false)
	ba, _ := vm.evaluationStack.Pop()
	val := ba

	if val.Int64() != 16 {
		t.Errorf("Expected first value to be 16 but was %v", val)
	}
}

func TestVM_Exec_PushOutOfBounds(t *testing.T) {
	code := []byte{
		PUSH, 0, 125,
		PUSH, 126, 12,
		HALT,
	}

	vm := NewTestVM([]byte{})
	mc := NewMockContext(code)
	vm.context = mc
	vm.Exec(false)

	tos, err := vm.evaluationStack.PopBytes()

	if err != nil {
		t.Errorf("%v", err)
	}

	actual := string(tos)
	expected := "push: instructionSet out of bounds"
	if actual != expected {
		t.Errorf("Expected '%v' to be returned but got '%v'", expected, actual)
	}
}

func TestVM_Exec_Addition(t *testing.T) {
	code := []byte{
		PUSH, 0, 125,
		PUSH, 1, 168, 22,
		ADD,
		HALT,
	}

	vm := NewTestVM([]byte{})
	mc := NewMockContext(code)
	vm.context = mc
	vm.Exec(false)

	tos, err := vm.evaluationStack.Peek()

	if err != nil {
		t.Errorf("%v", err)
	}

	if tos.Int64() != int64(43155) {
		t.Errorf("Actual value is %v, should be 53 after adding up 50 and 3", tos.Int64())
	}
}

func TestVM_Exec_Subtraction(t *testing.T) {
	code := []byte{
		PUSH, 0, 6,
		PUSH, 0, 3,
		SUB,
		HALT,
	}

	vm := NewTestVM([]byte{})
	mc := NewMockContext(code)
	vm.context = mc
	vm.Exec(false)

	tos, err := vm.evaluationStack.Peek()

	if err != nil {
		t.Errorf("%v", err)
	}

	if tos.Int64() != 3 {
		t.Errorf("Actual value is %v, should be 3 after subtracting 2 from 5", tos)
	}
}

func TestVM_Exec_SubtractionWithNegativeResults(t *testing.T) {
	code := []byte{
		PUSH, 0, 3,
		PUSH, 0, 6,
		SUB,
		HALT,
	}

	vm := NewTestVM([]byte{})
	mc := NewMockContext(code)
	vm.context = mc
	vm.Exec(false)

	tos, err := vm.evaluationStack.Peek()

	if err != nil {
		t.Errorf("%v", err)
	}

	if tos.Int64() != -3 {
		t.Errorf("Actual value is %v, should be -3 after subtracting 6 from 3", tos)
	}
}

func TestVM_Exec_Multiplication(t *testing.T) {
	code := []byte{
		PUSH, 0, 5,
		PUSH, 0, 2,
		MULT,
		HALT,
	}

	vm := NewTestVM([]byte{})
	mc := NewMockContext(code)
	vm.context = mc
	vm.Exec(false)

	tos, err := vm.evaluationStack.Peek()

	if err != nil {
		t.Errorf("%v", err)
	}

	if tos.Int64() != 10 {
		t.Errorf("Actual value is %v, should be 10 after multiplying 2 with 5", tos)
	}
}

func TestVM_Exec_Modulo(t *testing.T) {
	code := []byte{
		PUSH, 0, 5,
		PUSH, 0, 2,
		MOD,
		HALT,
	}

	vm := NewTestVM([]byte{})
	mc := NewMockContext(code)
	vm.context = mc
	vm.Exec(false)

	tos, err := vm.evaluationStack.Peek()

	if err != nil {
		t.Errorf("%v", err)
	}

	if tos.Int64() != 1 {
		t.Errorf("Actual value is %v, should be 1 after 5 mod 2", tos)
	}
}

func TestVM_Exec_Negate(t *testing.T) {
	code := []byte{
		PUSH, 0, 5,
		NEG,
		HALT,
	}

	vm := NewTestVM([]byte{})
	mc := NewMockContext(code)
	vm.context = mc
	vm.Exec(false)

	tos, err := vm.evaluationStack.Peek()

	if err != nil {
		t.Errorf("%v", err)
	}

	if tos.Int64() != -5 {
		t.Errorf("Actual value is %v, should be -5 after negating 5", tos)
	}
}

func TestVM_Exec_Division(t *testing.T) {
	code := []byte{
		PUSH, 0, 6,
		PUSH, 0, 2,
		DIV,
		HALT,
	}

	vm := NewTestVM([]byte{})
	mc := NewMockContext(code)
	vm.context = mc
	vm.Exec(false)

	tos, err := vm.evaluationStack.Peek()

	if err != nil {
		t.Errorf("%v", err)
	}

	if tos.Int64() != 3 {
		t.Errorf("Actual value is %v, should be 10 after dividing 6 by 2", tos)
	}
}

func TestVM_Exec_DivisionByZero(t *testing.T) {
	code := []byte{
		PUSH, 0, 6,
		PUSH, 0, 0,
		DIV,
		HALT,
	}

	vm := NewTestVM([]byte{})
	mc := NewMockContext(code)
	vm.context = mc
	vm.Exec(false)

	result, err := vm.evaluationStack.Pop()

	if err != nil {
		t.Errorf("%v", err)
	}

	expected := "div: Division by Zero"
	actual := BigIntToString(result)
	if actual != expected {
		t.Errorf("Expected tos to be '%v' error message but was '%v'", expected, actual)
	}
}

func TestVM_Exec_Eq(t *testing.T) {
	code := []byte{
		PUSH, 0, 6,
		PUSH, 0, 6,
		EQ,
		HALT,
	}

	vm := NewTestVM([]byte{})
	mc := NewMockContext(code)
	vm.context = mc
	vm.Exec(false)

	tos, err := vm.evaluationStack.PopBytes()
	if err != nil {
		t.Errorf("%v", err)
	}

	if !ByteArrayToBool(tos) {
		t.Errorf("Actual value is %v, should be 1 after comparing 6 with 6", tos[0])
	}
}

func TestVM_Exec_Neq(t *testing.T) {
	code := []byte{
		PUSH, 0, 6,
		PUSH, 0, 5,
		NEQ,
		HALT,
	}

	vm := NewTestVM([]byte{})
	mc := NewMockContext(code)
	vm.context = mc
	vm.Exec(false)

	tos, err := vm.evaluationStack.PopBytes()
	if err != nil {
		t.Errorf("%v", err)
	}

	if !ByteArrayToBool(tos) {
		t.Errorf("Actual value is %v, should be 1 after comparing 6 with 5 to not be equal", tos[0])
	}
}

func TestVM_Exec_Lt(t *testing.T) {
	code := []byte{
		PUSH, 0, 4,
		PUSH, 0, 6,
		LT,
		HALT,
	}

	vm := NewTestVM([]byte{})
	mc := NewMockContext(code)
	vm.context = mc
	vm.Exec(false)

	tos, err := vm.evaluationStack.PopBytes()

	if err != nil {
		t.Errorf("%v", err)
	}

	if !ByteArrayToBool(tos) {
		t.Errorf("Actual value is %v, should be 1 after evaluating 4 < 6", tos[0])
	}
}

func TestVM_Exec_Gt(t *testing.T) {
	code := []byte{
		PUSH, 0, 6,
		PUSH, 1, 0, 4,
		GT,
		HALT,
	}

	vm := NewTestVM([]byte{})
	mc := NewMockContext(code)
	vm.context = mc
	vm.Exec(false)

	tos, err := vm.evaluationStack.PopBytes()
	if err != nil {
		t.Errorf("%v", err)
	}

	if !ByteArrayToBool(tos) {
		t.Errorf("Actual value is %v, should be 1 after evaluating 6 > 4", tos[0])
	}
}

func TestVM_Exec_Lte_islower(t *testing.T) {
	code := []byte{
		PUSH, 0, 4,
		PUSH, 0, 6,
		LTE,
		HALT,
	}

	vm := NewTestVM([]byte{})
	mc := NewMockContext(code)
	vm.context = mc
	vm.Exec(false)

	tos, err := vm.evaluationStack.PopBytes()
	if err != nil {
		t.Errorf("%v", err)
	}

	if !ByteArrayToBool(tos) {
		t.Errorf("Actual value is %v, should be 1 after evaluating 4 <= 6", tos[0])
	}


}

func TestVM_Exec_Lte_isequals(t *testing.T) {
	code := []byte{
		PUSH, 0, 6,
		PUSH, 0, 6,
		LTE,
		HALT,
	}

	vm := NewTestVM([]byte{})
	mc := NewMockContext(code)
	vm.context = mc
	vm.Exec(false)

	tos, err := vm.evaluationStack.PopBytes()
	if err != nil {
		t.Errorf("%v", err)
	}

	if !ByteArrayToBool(tos) {
		t.Errorf("Actual value is %v, should be 1 after evaluating 6 <= 6", tos[0])
	}
}

func TestVM_Exec_Gte_isGreater(t *testing.T) {
	code := []byte{
		PUSH, 0, 6,
		PUSH, 0, 4,
		GTE,
		HALT,
	}

	vm := NewTestVM([]byte{})
	mc := NewMockContext(code)
	vm.context = mc
	vm.Exec(false)

	tos, err := vm.evaluationStack.PopBytes()
	if err != nil {
		t.Errorf("%v", err)
	}

	if !ByteArrayToBool(tos) {
		t.Errorf("Actual value is %v, should be 1 after evaluating 6 >= 4", tos[0])
	}
}

func TestVM_Exec_Gte_isEqual(t *testing.T) {
	code := []byte{
		PUSH, 0, 6,
		PUSH, 0, 6,
		GTE,
		HALT,
	}

	vm := NewTestVM([]byte{})
	mc := NewMockContext(code)
	vm.context = mc
	vm.Exec(false)

	tos, err := vm.evaluationStack.PopBytes()
	if err != nil {
		t.Errorf("%v", err)
	}

	if !ByteArrayToBool(tos) {
		t.Errorf("Actual value is %v, should be 1 after evaluating 6 >= 6", tos[0])
	}
}

func TestVM_Exec_Shiftl(t *testing.T) {
	code := []byte{
		PUSH, 0, 1,
		SHIFTL, 3,
		HALT,
	}

	vm := NewTestVM([]byte{})
	mc := NewMockContext(code)
	vm.context = mc
	vm.Exec(false)

	tos, _ := vm.evaluationStack.Pop()

	if tos.Int64() != 8 {
		t.Errorf("Expected result to be 8 but was %v", tos)
	}
}

func TestVM_Exec_Shiftr(t *testing.T) {
	code := []byte{
		PUSH, 0, 8,
		SHIFTR, 3,
		HALT,
	}

	vm := NewTestVM([]byte{})
	mc := NewMockContext(code)
	vm.context = mc
	vm.Exec(false)

	tos, _ := vm.evaluationStack.Pop()

	if tos.Int64() != 1 {
		t.Errorf("Expected result to be 1 but was %v", tos)
	}
}

func TestVM_Exec_Jmpif(t *testing.T) {
	code := []byte{
		PUSH, 0, 3,
		PUSH, 0, 4,
		ADD,
		PUSH, 0, 20,
		LT,
		JMPIF, 0, 18,
		PUSH, 0, 3,
		NOP,
		NOP,
		NOP,
		HALT,
	}

	vm := NewTestVM([]byte{})
	mc := NewMockContext(code)
	vm.context = mc
	vm.Exec(false)

	if vm.evaluationStack.GetLength() != 0 {
		t.Errorf("After calling and returning, callStack lenght should be 0, but is %v", vm.evaluationStack.GetLength())
	}
}

func TestVM_Exec_Jmp(t *testing.T) {
	code := []byte{
		PUSH, 0, 3,
		JMP, 0, 14,
		PUSH, 0, 4,
		ADD,
		PUSH, 0, 15,
		ADD,
		HALT,
	}

	vm := NewTestVM([]byte{})
	mc := NewMockContext(code)
	vm.context = mc
	vm.Exec(false)

	tos, err := vm.evaluationStack.Peek()

	if err != nil {
		t.Errorf("%v", err)
	}

	if tos.Int64() != 3 {
		t.Errorf("Actual value is %v, should be 3 after jumping to halt", tos)
	}
}

func TestVM_Exec_Call(t *testing.T) {
	code := []byte{
		PUSH, 0, 10,
		PUSH, 0, 8,
		CALL, 0, 13, 2,
		HALT,
		NOP,
		NOP,
		LOAD, 0, // Begin of called function at address 14
		LOAD, 1,
		SUB,
		RET,
	}

	vm := NewTestVM([]byte{})
	mc := NewMockContext(code)
	vm.context = mc
	vm.Exec(false)

	tos, err := vm.evaluationStack.Peek()

	if err != nil {
		t.Errorf("Expected empty Stack to throw an error when using peek() but it didn't")
	}

	if tos.Int64() != 2 {
		t.Errorf("Actual value is %v, sould be 3 after jumping to halt", tos)
	}

	callStackLenght := vm.callStack.GetLength()

	if callStackLenght != 0 {
		t.Errorf("After calling and returning, callStack lenght should be 0, but is %v", callStackLenght)
	}
}

func TestVM_Exec_Callif(t *testing.T) {
	code := []byte{
		PUSH, 0, 10,
		PUSH, 0, 8,
		PUSH, 0, 10,
		PUSH, 0, 10,
		EQ,
		CALLIF, 0, 20, 2,
		HALT,
		NOP,
		NOP,
		LOAD, 0, // Begin of called function at address 20
		LOAD, 1,
		SUB,
		RET,
	}

	vm := NewTestVM([]byte{})
	mc := NewMockContext(code)
	vm.context = mc
	vm.Exec(false)

	tos, err := vm.evaluationStack.Peek()

	if err != nil {
		t.Errorf("Expected empty Stack to throw an error when using peek() but it didn't")
	}

	if tos.Int64() != 2 {
		t.Errorf("Actual value is %v, sould be 3 after jumping to halt", tos)
	}

	callStackLenght := vm.callStack.GetLength()

	if callStackLenght != 0 {
		t.Errorf("After calling and returning, callStack lenght should be 0, but is %v", callStackLenght)
	}

	code1 := []byte{
		PUSH, 0, 10,
		PUSH, 0, 8,
		PUSH, 0, 10,
		PUSH, 0, 2,
		EQ,
		CALLIF, 0, 21, 2,
		HALT,
		NOP,
		NOP,
		LOAD, 0, // Begin of called function at address 21
		LOAD, 1,
		SUB,
		RET,
	}

	vm1 := NewTestVM([]byte{})
	mc1 := NewMockContext(code1)
	vm1.context = mc1
	vm1.Exec(false)

	tos1, err := vm1.evaluationStack.Peek()

	if err != nil {
		t.Errorf("Expected empty Stack to throw an error when using peek() but it didn't")
	}

	if tos1.Int64() != 8 {
		t.Errorf("Actual value is %v, sould be 8 after ignoring callif", tos)
	}

	callStackLenght1 := vm1.callStack.GetLength()

	if callStackLenght1 != 0 {
		t.Errorf("After skipping callif, callStack lenght should be 0, but is %v", callStackLenght)
	}
}

func TestVM_Exec_TosSize(t *testing.T) {
	code := []byte{
		PUSH, 2, 10, 4, 5,
		SIZE,
		HALT,
	}

	vm := NewTestVM([]byte{})
	mc := NewMockContext(code)
	vm.context = mc
	vm.Exec(false)

	tos, err := vm.evaluationStack.Pop()

	if err != nil {
		t.Errorf("Expected empty stack to throw an error when using peek() but it didn't")
	}

	if tos.Int64() != 4 {
		t.Errorf("Expected TOS size to be 4, but got %v", tos)
	}
}

func TestVM_Exec_CallExt(t *testing.T) {
	code := []byte{
		PUSH, 0, 10,
		PUSH, 0, 8,
		CALLEXT, 227, 237, 86, 189, 8, 109, 137, 88, 72, 58, 18, 115, 79, 160, 174, 127, 92, 139, 177, 96, 239, 144, 146, 198, 126, 130, 237, 155, 25, 228, 199, 178, 41, 24, 45, 14, 2,
		HALT,
	}

	vm := NewTestVM([]byte{})
	mc := NewMockContext(code)
	vm.context = mc
	vm.Exec(false)

}

func TestVM_Exec_Sload(t *testing.T) {
	code := []byte{
		SLOAD, 1,
		SLOAD, 0,
		SLOAD, 2,
		HALT,
	}

	vm := NewTestVM([]byte{})
	mc := NewMockContext(code)
	mc.ContractVariables = []protocol.ByteArray{[]byte("Hi There!!"), []byte{26}, []byte{0}}
	vm.context = mc

	vm.Exec(false)

	expected := big.NewInt(0)
	actual, _ := vm.evaluationStack.Pop()

	if actual.Cmp(expected) != 0 {
		t.Errorf("Unexpected value retrieved")
	}

	result, err := vm.evaluationStack.Pop()

	if err != nil {
		t.Errorf("%v", err)
	}

	resultString := BigIntToString(result)
	if resultString != "Hi There!!" {
		t.Errorf("The String on the Stack should be 'Hi There!!' but was %v", resultString)
	}

	expected = big.NewInt(26)
	actual, _ = vm.evaluationStack.Pop()

	if actual.Cmp(expected) != 0 {
		t.Errorf("Unexpected value retrieved")
	}
}

func TestVM_Exec_Sstore(t *testing.T) {
	code := []byte{
		PUSH, 9, 72, 105, 32, 84, 104, 101, 114, 101, 33, 33,
		SSTORE, 0,
		HALT,
	}

	vm := NewTestVM([]byte{})
	mc := NewMockContext(code)
	mc.ContractVariables = []protocol.ByteArray{[]byte("Something")}
	vm.context = mc
	vm.Exec(false)
	mc.PersistChanges()

	v, _ := vm.context.GetContractVariable(0)
	result := string(v)
	if result != "Hi There!!" {
		t.Errorf("The String on the Stack should be 'Hi There!!' but was '%v'", result)
	}
}

func TestVM_Exec_Address(t *testing.T) {
	code := []byte{
		ADDRESS,
		HALT,
	}

	vm := NewTestVM([]byte{})
	mc := NewMockContext(code)
	ba := [64]byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}
	mc.Address = ba
	vm.context = mc

	vm.Exec(false)
	tos, _ := vm.evaluationStack.PopBytes()

	if len(tos) != 64 {
		t.Errorf("Expected TOS size to be 64, but got %v", len(tos))
	}

	//This just tests 1/8 of the address as Uint64 are 64 bits and the address is 64 bytes
	actual := binary.LittleEndian.Uint64(tos)
	var expected uint64 = 18446744073709551615

	if expected != actual {
		t.Errorf("Expected TOS size to be '%v', but got '%v'", expected, actual)
	}
}

func TestVM_Exec_Balance(t *testing.T) {
	code := []byte{
		BALANCE,
		HALT,
	}

	vm := NewTestVM([]byte{})
	mc := NewMockContext(code)
	mc.Balance = uint64(100)
	vm.context = mc

	vm.Exec(false)
	tos, _ := vm.evaluationStack.PopBytes()

	if len(tos) != 8 {
		t.Errorf("Expected TOS size to be 64, but got %v", len(tos))
	}

	actual := binary.LittleEndian.Uint64(tos)
	var expected uint64 = 100

	if actual != expected {
		t.Errorf("Expected TOS to be '%v', but got '%v'", expected, actual)
	}
}

func TestVM_Exec_Caller(t *testing.T) {
	code := []byte{
		CALLER,
		HALT,
	}

	vm := NewTestVM([]byte{})
	mc := NewMockContext(code)
	from := [32]byte{
		0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
		0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
		0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
		0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
	}
	mc.From = from
	vm.context = mc

	vm.Exec(false)
	tos, _ := vm.evaluationStack.PopBytes()

	if len(tos) != 32 {
		t.Errorf("Expected TOS size to be 32, but got %v", len(tos))
	}

	if !bytes.Equal(tos, from[:]) {
		t.Errorf("Retrieved unexpected value")
	}
}

func TestVM_Exec_Callval(t *testing.T) {
	code := []byte{
		CALLVAL,
		HALT,
	}

	vm := NewTestVM([]byte{})
	mc := NewMockContext(code)
	mc.Amount = uint64(100)
	vm.context = mc

	vm.Exec(false)
	tos, _ := vm.evaluationStack.PopBytes()

	if len(tos) != 8 {
		t.Errorf("Expected TOS size to be 8, but got %v", len(tos))
	}

	result := binary.LittleEndian.Uint64(tos)

	if result != 100 {
		t.Errorf("Expected value to be 100, but got %v", result)
	}
}

func TestVM_Exec_Calldata(t *testing.T) {
	code := []byte{
		CALLDATA,
		HALT,
	}

	vm := NewTestVM([]byte{})
	mc := NewMockContext(code)
	mc.Fee = 50

	td := []byte{
		0, 0x02,
		0, 0x05,
		3, 0x10, 0x12, 0x4, 0x12, // Function hash
	}
	mc.transactionData = td

	vm.context = mc
	vm.Exec(false)

	functionHash, _ := vm.evaluationStack.PopBytes()

	if !bytes.Equal(functionHash, td[5:]) {
		t.Errorf("expected '%# x' but got '%# x'", td[5:], functionHash)
	}

	arg1, _ := vm.evaluationStack.PopBytes()
	if !bytes.Equal(arg1, td[3:4]) {
		t.Errorf("expected '%# x' but got '%# x'", td[3:4], arg1)
	}

	arg2, _ := vm.evaluationStack.PopBytes()
	if !bytes.Equal(arg2, td[1:2]) {
		t.Errorf("expected '%# x' but got '%# x'", td[1:2], arg2)
	}
}

func TestVM_Exec_Sha3(t *testing.T) {
	code := []byte{
		PUSH, 0, 3,
		SHA3,
		HALT,
	}

	vm := NewTestVM([]byte{})
	mc := NewMockContext(code)
	vm.context = mc
	vm.Exec(false)

	val, _ := vm.evaluationStack.PopBytes()

	if !bytes.Equal(val, []byte{227, 237, 86, 189, 8, 109, 137, 88, 72, 58, 18, 115, 79, 160, 174, 127, 92, 139, 177, 96, 239, 144, 146, 198, 126, 130, 237, 155, 25, 228, 199, 178}) {
		t.Errorf("Actual value is %v, should be {227, 237, 86, 189...} after jumping to halt", val)
	}
}

func TestVM_Exec_Roll(t *testing.T) {
	code := []byte{
		PUSH, 0, 3,
		PUSH, 0, 4,
		PUSH, 0, 5,
		PUSH, 0, 6,
		PUSH, 0, 7,
		ROLL, 2,
		HALT,
	}

	vm := NewTestVM([]byte{})
	mc := NewMockContext(code)
	vm.context = mc
	vm.Exec(false)

	tos, _ := vm.evaluationStack.Pop()

	if tos.Int64() != 4 {
		t.Errorf("Actual value is %v, should be 4 after rolling with two as arg", tos)
	}
}

func TestVM_Exec_NewMap(t *testing.T) {
	code := []byte{
		NEWMAP,
		HALT,
	}

	vm := NewTestVM([]byte{})
	mc := NewMockContext(code)
	vm.context = mc
	vm.Exec(false)

	actual, err := vm.evaluationStack.PopBytes()
	if err != nil {
		t.Errorf("%v", err)
	}

	expected := []byte{0x01, 0x00, 0x00}

	if !bytes.Equal(expected, actual) {
		t.Errorf("expected the Value of the new Map to be '[%v]' but was '[%v]'", expected, actual)
	}
}

func TestVM_Exec_MapPush(t *testing.T) {
	code := []byte{
		PUSH, 1, 72, 105,
		PUSH, 0, 0x03,
		NEWMAP,
		MAPPUSH,
		HALT,
	}

	vm := NewTestVM([]byte{})
	mc := NewMockContext(code)
	vm.context = mc
	exec := vm.Exec(false)

	if !exec {
		errorMessage, _ := vm.evaluationStack.PopBytes()
		t.Errorf("VM.Exec terminated with Error: %v", string(errorMessage))
	}

	m, err := vm.evaluationStack.PopBytes()
	if err != nil {
		t.Errorf("%v", err)
	}

	mp, err2 := MapFromByteArray(m)
	if err2 != nil {
		t.Errorf("%v", err)
	}

	datastructure := mp[0]
	size, err := mp.getSize()

	if err != nil {
		t.Error(err)
	}

	if datastructure != 0x01 {
		t.Errorf("Invalid Datastructure ID, Expected 0x01 but was %v", datastructure)
	}

	if size != 1 {
		t.Errorf("invalid size, Expected 1 but was %v", size)
	}

}

func TestVM_Exec_MapGetVAL(t *testing.T) {
	code := []byte{
		PUSH, 0x00, 0x01, //The key for MAPGETVAL

		PUSH, 0x01, 0x48, 0x48,
		PUSH, 0x00, 0x01,

		PUSH, 0x01, 0x69, 0x69,
		PUSH, 0x00, 0x02,

		PUSH, 0x01, 0x48, 0x69,
		PUSH, 0x00, 0x03,

		NEWMAP,

		MAPPUSH,
		MAPPUSH,
		MAPPUSH,

		MAPGETVAL,

		HALT,
	}

	vm := NewTestVM([]byte{})
	mc := NewMockContext(code)
	vm.context = mc

	exec := vm.Exec(false)
	if !exec {
		errorMessage, _ := vm.evaluationStack.Pop()
		t.Errorf("VM.Exec terminated with Error: %v", BigIntToString(errorMessage))
	}

	actual, err := vm.evaluationStack.PopBytes()
	if err != nil {
		t.Errorf("%v", err)
	}

	expected := []byte{72, 72}
	if !bytes.Equal(actual, expected) {
		t.Errorf("invalid value, Expected '%v' but was '%v'", expected, actual)
	}
}

func TestVM_Exec_MapSetVal(t *testing.T) {
	code := []byte{
		PUSH, 0x01, 0x55, 0x55,  //Value to be reset by MAPSETVAL
		PUSH, 0x00, 0x03,

		PUSH, 0x01, 0x48, 0x69,
		PUSH, 0x00, 0x03,

		PUSH, 0x01, 0x69, 0x69,
		PUSH, 0x00, 0x02,

		NEWMAP,

		MAPPUSH,
		MAPPUSH,

		MAPSETVAL,

		HALT,
	}

	vm := NewTestVM([]byte{})
	mc := NewMockContext(code)
	vm.context = mc
	exec := vm.Exec(false)

	if !exec {
		errorMessage, _ := vm.evaluationStack.PopBytes()
		t.Errorf("VM.Exec terminated with Error: %v", string(errorMessage))
	}

	mbi, err := vm.evaluationStack.PopBytes()
	if err != nil {
		t.Errorf("%v", err)
	}
	actual, err := MapFromByteArray(mbi)
	if err != nil {
		t.Errorf("%v", err)
	}

	expected := []byte{0x01,
		0x02, 0x00,
		0x01, 0x00, 0x02,
		0x02, 0x00, 0x69, 0x69,
		0x01, 0x00, 0x03,
		0x02, 0x00, 0x55, 0x55,
	}

	if !bytes.Equal(actual, expected) {
		t.Errorf("invalid datastructure, Expected '[%# x]' but was '[%# x]'", expected, actual)
	}
}

func TestVM_Exec_MapRemove(t *testing.T) {
	code := []byte{
		PUSH, 0x00, 0x03, // The Key to be removed with MAPREMOVE

		PUSH, 0x01, 0x48, 0x69,
		PUSH, 0x00, 0x03,

		PUSH, 0x01, 0x48, 0x48,
		PUSH, 0x00, 0x01,

		PUSH, 0x01, 0x69, 0x69,
		PUSH, 0x00, 0x02,

		NEWMAP,

		MAPPUSH,
		MAPPUSH,
		MAPPUSH,

		MAPREMOVE,
		HALT,
	}

	vm := NewTestVM([]byte{})
	mc := NewMockContext(code)
	vm.context = mc

	exec := vm.Exec(false)
	if !exec {
		errorMessage, _ := vm.evaluationStack.Pop()
		t.Errorf("VM.Exec terminated with Error: %v", BigIntToString(errorMessage))
	}

	mapAsByteArray, err := vm.evaluationStack.PopBytes()
	if err != nil {
		t.Errorf("%v", err)
	}

	actual, err := MapFromByteArray(mapAsByteArray)
	if err != nil {
		t.Errorf("%v", err)
	}

	expected := []byte{0x01,
		0x02, 0x00,
		0x01, 0x00, 0x02,
		0x02, 0x00, 0x69, 0x69,
		0x01, 0x00, 0x01,
		0x02, 0x00, 0x48, 0x48,
	}

	if !bytes.Equal(actual, expected) {
		t.Errorf("invalid datastructure, Expected '[%# x]' but was '[%# x]'", expected, actual)
	}
}

func TestVM_Exec_NewArr(t *testing.T) {
	code := []byte{
		NEWARR,
		HALT,
	}

	vm := NewTestVM([]byte{})
	mc := NewMockContext(code)
	vm.context = mc
	exec := vm.Exec(false)

	if !exec {
		errorMessage, _ := vm.evaluationStack.Pop()
		t.Errorf("VM.Exec terminated with Error: %v", BigIntToString(errorMessage))
	}

	arr, err := vm.evaluationStack.PopBytes()
	if err != nil {
		t.Errorf("%v", err)
	}
	expectedSize := []byte{0x00, 0x00}
	actualSize := arr[1:3]
	if !bytes.Equal(expectedSize, actualSize) {
		t.Errorf("invalid size, Expected %v but was '%v'", expectedSize, actualSize)
	}
}

func TestVM_Exec_ArrAppend(t *testing.T) {
	code := []byte{
		NEWARR,
		PUSH, 0x01, 0xFF, 0x00,
		ARRAPPEND,
		HALT,
	}

	vm := NewTestVM([]byte{})
	mc := NewMockContext(code)
	vm.context = mc
	exec := vm.Exec(false)
	if !exec {
		errorMessage, _ := vm.evaluationStack.PopBytes()
		t.Errorf("VM.Exec terminated with Error: %v", string(errorMessage))
	}

	arr, err := vm.evaluationStack.PopBytes()
	if err != nil {
		t.Errorf("%v", err)
	}

	actual := arr[5:7]
	expected := []byte{0xFF, 0x00}
	if !bytes.Equal(expected, actual) {
		t.Errorf("invalid element appended, Expected '%v' but was '%v'", expected, actual)
	}
}

func TestVM_Exec_ArrInsert(t *testing.T) {
	code := []byte{
		NEWARR,
		PUSH, 0x01, 0xFF, 0x00,
		ARRAPPEND,
		PUSH, 0x01, 0xFF, 0x00,
		ARRAPPEND,
		PUSH, 0x01, 0x00, 0x00,
		PUSH, 0x01, 0x00, 0x00,
		ARRINSERT,
		HALT,
	}

	vm := NewTestVM([]byte{})
	mc := NewMockContext(code)
	vm.context = mc
	exec := vm.Exec(false)
	if !exec {
		errorMessage, _ := vm.evaluationStack.PopBytes()
		t.Errorf("VM.Exec terminated with Error: %v", string(errorMessage))
	}

	actual, err := vm.evaluationStack.PopBytes()
	if err != nil {
		t.Errorf("%v", err)
	}

	expectedSize := []byte{0x02}
	if !bytes.Equal(expectedSize, actual[1:2]) {
		t.Errorf("invalid element appended, Expected '[%# x]' but was '[%# x]'", expectedSize, actual[1:2])
	}

	expectedValue := []byte{0x00, 0x00}
	if !bytes.Equal(expectedValue, actual[5:7]) {
		t.Errorf("invalid element appended, Expected '[%# x' but was '[%# x]'", expectedValue, actual[5:7])
	}
}

func TestVM_Exec_ArrRemove(t *testing.T) {
	code := []byte{
		NEWARR,
		PUSH, 0x01, 0xFF, 0x00,
		ARRAPPEND,
		PUSH, 0x01, 0xAA, 0x00,
		ARRAPPEND,
		PUSH, 0x01, 0xBB, 0x00,
		ARRAPPEND,
		ARRREMOVE, 0x01, 0x00,
		HALT,
	}

	vm := NewTestVM([]byte{})
	mc := NewMockContext(code)
	vm.context = mc
	exec := vm.Exec(false)

	if !exec {
		errorMessage, _ := vm.evaluationStack.PopBytes()
		t.Errorf("VM.Exec terminated with Error: %v", string(errorMessage))
	}

	a, err := vm.evaluationStack.PopBytes()
	if err != nil {
		t.Errorf("%v", err)
	}

	arr, bierr := ArrayFromByteArray(a)
	if bierr != nil {
		t.Errorf("%v", err)
	}

	size, err := arr.getSize()
	if err != nil {
		t.Error(err)
	}

	if size != uint16(2) {
		t.Errorf("invalid array size, Expected 2 but was '%v'", size)
	}

	expectedSecondElement := []byte{0xBB, 0x00}
	actualSecondElement, err2 := arr.At(uint16(1))
	if err2 != nil {
		t.Errorf("%v", err)
	}

	if !bytes.Equal(expectedSecondElement, actualSecondElement) {
		t.Errorf("invalid element on second index, Expected '[%# x]' but was '[%# x]'", expectedSecondElement, actualSecondElement)
	}
}

func TestVM_Exec_ArrAt(t *testing.T) {
	code := []byte{
		NEWARR,
		PUSH, 0x01, 0xFF, 0x00,
		ARRAPPEND,
		PUSH, 0x01, 0xAA, 0x00,
		ARRAPPEND,
		PUSH, 0x01, 0xBB, 0x00,
		ARRAPPEND,
		ARRAT, 0x02, 0x00,
		HALT,
	}

	vm := NewTestVM([]byte{})
	mc := NewMockContext(code)
	vm.context = mc
	exec := vm.Exec(false)

	if !exec {
		errorMessage, _ := vm.evaluationStack.PopBytes()
		t.Errorf("VM.Exec terminated with Error: %v", string(errorMessage))
	}

	actual, err1 := vm.evaluationStack.PopBytes()

	if err1 != nil {
		t.Errorf("%v", err1)
	}

	expected := []byte{0xBB, 0x00}
	if !bytes.Equal(expected, actual) {
		t.Errorf("invalid element on first index, Expected '[%# x]' but was '[%# x]'", expected, actual)
	}

}

func TestVM_Exec_NonValidOpCode(t *testing.T) {
	code := []byte{
		89,
	}

	vm := NewTestVM([]byte{})
	mc := NewMockContext(code)
	vm.context = mc
	vm.Exec(false)

	tos, _ := vm.evaluationStack.Pop()

	expected := "vm.exec(): Not a valid opCode"
	actual := BigIntToString(tos)
	if actual != expected {
		t.Errorf("Expected tos to be '%v' error message but was '%v'", expected, actual)
	}
}

func TestVM_Exec_ArgumentsExceedInstructionSet(t *testing.T) {
	code := []byte{
		PUSH, 0x00, 0x00, PUSH, 0x0b, 0x01, 0x00, 0x03, 0x12, 0x05,
	}

	vm := NewTestVM([]byte{})
	mc := NewMockContext(code)
	vm.context = mc
	vm.Exec(false)

	tos, _ := vm.evaluationStack.Pop()

	expected := "push: instructionSet out of bounds"
	actual := BigIntToString(tos)
	if actual != expected {
		t.Errorf("Expected tos to be '%v' error message but was '%v'", expected, actual)
	}
}

func TestVM_Exec_PopOnEmptyStack(t *testing.T) {
	code := []byte{
		PUSH, 0x00, 0x01, SHA3, 0x05, 0x02, 0x03,
	}

	vm := NewTestVM([]byte{})
	mc := NewMockContext(code)
	vm.context = mc
	vm.Exec(false)

	tos, _ := vm.evaluationStack.Pop()

	expected := "sub: pop() on empty stack"
	actual := BigIntToString(tos)
	if actual != expected {
		t.Errorf("Expected tos to be '%v' error message but was '%v'", expected, actual)
	}
}

func TestVM_Exec_FuzzReproduction_InstructionSetOutOfBounds(t *testing.T) {
	code := []byte{
		PUSH, 0, 20,
		ROLL, 0,
	}

	vm := NewTestVM([]byte{})
	mc := NewMockContext(code)
	vm.context = mc
	vm.Exec(false)

	tos, _ := vm.evaluationStack.Pop()

	expected := "vm.exec(): instructionSet out of bounds"
	actual := BigIntToString(tos)
	if actual != expected {
		t.Errorf("Expected tos to be '%v' error message but was '%v'", expected, actual)
	}
}

func TestVM_Exec_FuzzReproduction_InstructionSetOutOfBounds2(t *testing.T) {
	code := []byte{
		CALLEXT, 231,
	}

	vm := NewTestVM([]byte{})
	mc := NewMockContext(code)
	vm.context = mc
	vm.Exec(false)

	tos, _ := vm.evaluationStack.Pop()

	expected := "callext: instructionSet out of bounds"
	actual := BigIntToString(tos)
	if actual != expected {
		t.Errorf("Expected tos to be '%v' error message but was '%v'", expected, actual)
	}
}

func TestVM_Exec_FuzzReproduction_IndexOutOfBounds1(t *testing.T) {
	code := []byte{
		SLOAD, 0, 0, 33,
	}

	vm := NewTestVM([]byte{})
	mc := NewMockContext(code)
	vm.context = mc
	vm.Exec(false)

	tos, _ := vm.evaluationStack.Pop()

	expected := "sload: Index out of bounds"
	actual := BigIntToString(tos)
	if actual != expected {
		t.Errorf("Expected tos to be '%v' error message but was '%v'", expected, actual)
	}
}

func TestVM_Exec_FuzzReproduction_IndexOutOfBounds2(t *testing.T) {
	code := []byte{
		PUSH, 4, 46, 110, 66, 50, 255, SSTORE, 123, 119,
	}

	vm := NewTestVM([]byte{})
	mc := NewMockContext(code)
	vm.context = mc
	vm.Exec(false)

	tos, _ := vm.evaluationStack.Pop()

	expected := "sstore: Index out of bounds"
	actual := BigIntToString(tos)
	if actual != expected {
		t.Errorf("Expected tos to be '%v' error message but was '%v'", expected, actual)
	}
}

func TestVM_Exec_FunctionCallSub(t *testing.T) {
	code := []byte{
		// start ABI
		CALLDATA,
		DUP,
		PUSH, 0, 1,
		EQ,
		JMPIF, 0, 18,
		DUP,
		PUSH, 0, 2,
		EQ,
		JMPIF, 0, 21,
		HALT,
		// end ABI
		POP,
		SUB,
		HALT,
		POP,
		ADD,
		HALT,
	}

	vm := NewTestVM([]byte{})
	mc := NewMockContext(code)

	mc.transactionData = []byte{
		0, 2,
		0, 5,
		0, 1, // Function hash
	}

	vm.context = mc
	vm.Exec(false)

	tos, _ := vm.evaluationStack.Pop()

	if tos.Uint64() != 3 {
		t.Errorf("Expected tos to be '3' error message but was %v", tos.Uint64())
	}
}

func TestVM_Exec_FunctionCall(t *testing.T) {
	code := []byte{
		// start ABI
		CALLDATA,
		DUP,
		PUSH, 0, 1,
		EQ,
		JMPIF, 0, 18,
		DUP,
		PUSH, 0, 2,
		EQ,
		JMPIF, 0, 21,
		HALT,
		// end ABI
		POP,
		SUB,
		HALT,
		POP,
		ADD,
		HALT,
	}

	vm := NewTestVM([]byte{})
	mc := NewMockContext(code)

	mc.transactionData = []byte{
		0, 2,
		0, 5,
		0, 2, // Function hash
	}

	vm.context = mc
	vm.Exec(false)

	tos, _ := vm.evaluationStack.Pop()

	if tos.Uint64() != 7 {
		t.Errorf("Expected tos to be '7' error message but was %v", tos.Uint64())
	}
}

func TestVM_Exec_GithubIssue13(t *testing.T) {
	code := []byte{
		ADDRESS, ARRAT,
	}

	vm := NewTestVM([]byte{})
	mc := NewMockContext(code)
	vm.context = mc
	vm.Exec(false)

	tos, _ := vm.evaluationStack.Pop()

	expected := "arrat: instructionSet out of bounds"
	actual := BigIntToString(tos)
	if actual != expected {
		t.Errorf("Expected tos to be '%v' error message but was '%v'", expected, actual)
	}
}

func TestVm_Exec_FuzzReproduction_ContextOpCode1(t *testing.T) {
	code := []byte{
		CALLER, CALLER, ARRAPPEND,
	}

	vm := NewTestVM([]byte{})
	mc := NewMockContext(code)
	vm.context = mc
	vm.Exec(false)

	tos, _ := vm.evaluationStack.Pop()

	expected := "arrappend: not a valid array"
	actual := BigIntToString(tos)
	if actual != expected {
		t.Errorf("Expected tos to be '%v' error message but was '%v'", expected, actual)
	}
}

func TestVm_Exec_FuzzReproduction_ContextOpCode2(t *testing.T) {
	code := []byte{
		ADDRESS, CALLER, ARRAPPEND,
	}

	vm := NewTestVM([]byte{})
	mc := NewMockContext(code)
	vm.context = mc
	vm.Exec(false)

	tos, _ := vm.evaluationStack.Pop()

	expected := "arrappend: not a valid array"
	actual := BigIntToString(tos)
	if actual != expected {
		t.Errorf("Expected tos to be '%v' error message but was '%v'", expected, actual)
	}
}

func TestVm_Exec_FuzzReproduction_indexOutOfRange(t *testing.T) {
	code := []byte{
		51,
	}

	vm := NewTestVM([]byte{})
	mc := NewMockContext(code)
	vm.context = mc
	vm.Exec(false)

	tos, _ := vm.evaluationStack.Pop()

	expected := "vm.exec(): Not a valid opCode"
	if BigIntToString(tos) != expected {
		t.Errorf("expected: '%v' but was '%v'", expected, BigIntToString(tos))
	}
}

func BenchmarkVM_Exec_ModularExponentiation_GoImplementation(b *testing.B) {
	var base big.Int
	var exponent big.Int
	var modulus big.Int

	//base.SetInt64(4)
	for n := 0; n < b.N; n++ {
		base.SetBytes(protocol.RandomBytesWithLength(10000))
		exponent.SetBytes(protocol.RandomBytesWithLength(1))
		modulus.SetBytes(protocol.RandomBytesWithLength(2))

		modularExp(base, exponent, modulus)
	}
}

func modularExp(base big.Int, exponent big.Int, modulus big.Int) *big.Int {
	if modulus.CmpAbs(big.NewInt(int64(1))) == 0 {
		return big.NewInt(0)
	}
	start := big.NewInt(1)
	c := big.NewInt(1)
	for i := new(big.Int).Set(start); i.Cmp(&exponent) < 0; i.Add(i, big.NewInt(1)) {
		c = c.Mul(c, &base)
		c = c.Mod(c, &modulus)
	}
	return start
}

func TestVm_Exec_ModularExponentiation_ContractImplementation(t *testing.T) {
	code := []byte{
		0, 0, 13, 1, 0, 0, 0, 2, 0, 0, 1, 2, 0, 1, 0, 0, 1, 4, 12, 20, 0, 7, 49,
	}

	vm := NewTestVM([]byte{})
	mc := NewMockContext(code)
	vm.context = mc
	vm.Exec(false)
}
