package vm

import (
	"bytes"
	"encoding/binary"
	"math/big"
	"testing"

	"fmt"

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
		PUSH, 1, 0, 8,
		PUSH, 1, 0, 8,
		ADD,
		HALT,
	}

	vm := NewTestVM([]byte{})
	mc := NewMockContext(code)
	mc.Fee = 30
	vm.context = mc

	vm.Exec(false)

	ba, _ := vm.evaluationStack.Pop()
	expected := 16
	actual := ByteArrayToInt(ba)

	if expected != actual {
		t.Errorf("Expected first value to be '%v' but was '%v'", expected, actual)
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

	tos, err := vm.evaluationStack.Pop()
	if err != nil {
		t.Errorf("%v", err)
	}

	actual := string(tos)
	expected := "push: Instruction set out of bounds"

	if actual != expected {
		t.Errorf("Expected '%v' to be returned but got '%v'", expected, actual)
	}
}

func TestVM_Exec_Addition(t *testing.T) {
	code := []byte{
		PUSH, 1, 0, 125,
		PUSH, 2, 0, 168, 22,
		ADD,
		HALT,
	}

	vm := NewTestVM([]byte{})
	mc := NewMockContext(code)
	vm.context = mc
	vm.Exec(false)

	tos, _ := vm.evaluationStack.Pop()

	expected := 43155
	actual := ByteArrayToInt(tos)

	if expected != actual {
		t.Errorf("Expected result to be '%v' but was '%v'", expected, actual)
	}
}

func TestVM_Exec_Subtraction(t *testing.T) {
	code := []byte{
		PUSH, 1, 0, 6,
		PUSH, 1, 0, 3,
		SUB,
		HALT,
	}

	vm := NewTestVM([]byte{})
	mc := NewMockContext(code)
	vm.context = mc
	vm.Exec(false)

	tos, _ := vm.evaluationStack.Pop()

	expected := 3
	actual := ByteArrayToInt(tos)

	if expected != actual {
		t.Errorf("Expected result to be '%v' but was '%v'", expected, actual)
	}
}

func TestVM_Exec_SubtractionWithNegativeResults(t *testing.T) {
	code := []byte{
		PUSH, 1, 0, 3,
		PUSH, 1, 0, 6,
		SUB,
		HALT,
	}

	vm := NewTestVM([]byte{})
	mc := NewMockContext(code)
	vm.context = mc
	vm.Exec(false)

	tos, _ := vm.evaluationStack.Pop()

	expected := -3
	actual := ByteArrayToInt(tos[1:])

	if tos[0] == 0x01 {
		actual = actual * -1
	}

	if expected != actual {
		t.Errorf("Expected result to be '%v' but was '%v'", expected, actual)
	}
}

func TestVM_Exec_Multiplication(t *testing.T) {
	code := []byte{
		PUSH, 1, 0, 5,
		PUSH, 1, 0, 2,
		MULT,
		HALT,
	}

	vm := NewTestVM([]byte{})
	mc := NewMockContext(code)
	vm.context = mc
	vm.Exec(false)

	tos, _ := vm.evaluationStack.Pop()

	expected := 10
	actual := ByteArrayToInt(tos)

	if expected != actual {
		t.Errorf("Expected result to be '%v' but was '%v'", expected, actual)
	}
}

func TestVM_Exec_Modulo(t *testing.T) {
	code := []byte{
		PUSH, 1, 0, 5,
		PUSH, 1, 0, 2,
		MOD,
		HALT,
	}

	vm := NewTestVM([]byte{})
	mc := NewMockContext(code)
	vm.context = mc
	vm.Exec(false)

	tos, _ := vm.evaluationStack.Pop()

	expected := 1
	actual := ByteArrayToInt(tos)

	if expected != actual {
		t.Errorf("Expected result to be '%v' but was '%v'", expected, actual)
	}
}

func TestVM_Exec_Negate(t *testing.T) {
	code := []byte{
		PUSH, 1, 0, 5,
		NEG,
		HALT,
	}

	vm := NewTestVM([]byte{})
	mc := NewMockContext(code)
	vm.context = mc
	vm.Exec(false)

	tos, _ := vm.evaluationStack.Pop()

	expected := big.NewInt(-5)
	actual, _ := SignedBigIntConversion(tos, nil)

	if !(expected.Cmp(&actual) == 0) {
		t.Errorf("Expected result to be '%v' but was '%v'", expected, actual)
	}
}

func TestVM_Exec_Division(t *testing.T) {
	code := []byte{
		PUSH, 1, 0, 6,
		PUSH, 1, 0, 2,
		DIV,
		HALT,
	}

	vm := NewTestVM([]byte{})
	mc := NewMockContext(code)
	vm.context = mc
	vm.Exec(false)

	tos, _ := vm.evaluationStack.Pop()

	expected := 3
	actual := ByteArrayToInt(tos)

	if expected != actual {
		t.Errorf("Expected result to be '%v' but was '%v'", expected, actual)
	}
}

func TestVM_Exec_DivisionByZero(t *testing.T) {
	code := []byte{
		PUSH, 1, 0, 6,
		PUSH, 1, 0, 0,
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
	actual := string(result)
	if actual != expected {
		t.Errorf("Expected tos to be '%v' error message but was '%v'", expected, actual)
	}
}

func TestVM_Exec_Eq(t *testing.T) {
	code := []byte{
		PUSH, 1, 0, 6,
		PUSH, 1, 0, 6,
		EQ,
		HALT,
	}

	vm := NewTestVM([]byte{})
	mc := NewMockContext(code)
	vm.context = mc
	vm.Exec(false)

	tos, err := vm.evaluationStack.Pop()
	if err != nil {
		t.Errorf("%v", err)
	}

	if !ByteArrayToBool(tos) {
		t.Errorf("Actual value is %v, should be 1 after comparing 6 with 6", tos[0])
	}
}

func TestVM_Exec_Neq(t *testing.T) {
	code := []byte{
		PUSH, 1, 0, 6,
		PUSH, 1, 0, 5,
		NEQ,
		HALT,
	}

	vm := NewTestVM([]byte{})
	mc := NewMockContext(code)
	vm.context = mc
	vm.Exec(false)

	tos, err := vm.evaluationStack.Pop()
	if err != nil {
		t.Errorf("%v", err)
	}

	if !ByteArrayToBool(tos) {
		t.Errorf("Actual value is %v, should be 1 after comparing 6 with 5 to not be equal", tos[0])
	}
}

func TestVM_Exec_Lt(t *testing.T) {
	code := []byte{
		PUSH, 1, 0, 4,
		PUSH, 1, 0, 6,
		LT,
		HALT,
	}

	vm := NewTestVM([]byte{})
	mc := NewMockContext(code)
	vm.context = mc
	vm.Exec(false)

	tos, err := vm.evaluationStack.Pop()

	if err != nil {
		t.Errorf("%v", err)
	}

	if !ByteArrayToBool(tos) {
		t.Errorf("Actual value is %v, should be 1 after evaluating 4 < 6", tos[0])
	}
}

func TestVM_Exec_Gt(t *testing.T) {
	code := []byte{
		PUSH, 1, 0, 6,
		PUSH, 1, 0, 4,
		GT,
		HALT,
	}

	vm := NewTestVM([]byte{})
	mc := NewMockContext(code)
	vm.context = mc
	vm.Exec(false)

	tos, err := vm.evaluationStack.Pop()
	if err != nil {
		t.Errorf("%v", err)
	}

	if !ByteArrayToBool(tos) {
		t.Errorf("Actual value is %v, should be 1 after evaluating 6 > 4", tos[0])
	}
}

func TestVM_Exec_Lte_islower(t *testing.T) {
	code := []byte{
		PUSH, 1, 0, 4,
		PUSH, 1, 0, 6,
		LTE,
		HALT,
	}

	vm := NewTestVM([]byte{})
	mc := NewMockContext(code)
	vm.context = mc
	vm.Exec(false)

	tos, err := vm.evaluationStack.Pop()
	if err != nil {
		t.Errorf("%v", err)
	}

	if !ByteArrayToBool(tos) {
		t.Errorf("Actual value is %v, should be 1 after evaluating 4 <= 6", tos[0])
	}

}

func TestVM_Exec_Lte_isequals(t *testing.T) {
	code := []byte{
		PUSH, 1, 0, 6,
		PUSH, 1, 0, 6,
		LTE,
		HALT,
	}

	vm := NewTestVM([]byte{})
	mc := NewMockContext(code)
	vm.context = mc
	vm.Exec(false)

	tos, err := vm.evaluationStack.Pop()
	if err != nil {
		t.Errorf("%v", err)
	}

	if !ByteArrayToBool(tos) {
		t.Errorf("Actual value is %v, should be 1 after evaluating 6 <= 6", tos[0])
	}
}

func TestVM_Exec_Gte_isGreater(t *testing.T) {
	code := []byte{
		PUSH, 1, 0, 6,
		PUSH, 1, 0, 4,
		GTE,
		HALT,
	}

	vm := NewTestVM([]byte{})
	mc := NewMockContext(code)
	vm.context = mc
	vm.Exec(false)

	tos, err := vm.evaluationStack.Pop()
	if err != nil {
		t.Errorf("%v", err)
	}

	if !ByteArrayToBool(tos) {
		t.Errorf("Actual value is %v, should be 1 after evaluating 6 >= 4", tos[0])
	}
}

func TestVM_Exec_Gte_isEqual(t *testing.T) {
	code := []byte{
		PUSH, 1, 0, 6,
		PUSH, 1, 0, 6,
		GTE,
		HALT,
	}

	vm := NewTestVM([]byte{})
	mc := NewMockContext(code)
	vm.context = mc
	vm.Exec(false)

	tos, err := vm.evaluationStack.Pop()
	if err != nil {
		t.Errorf("%v", err)
	}

	if !ByteArrayToBool(tos) {
		t.Errorf("Actual value is %v, should be 1 after evaluating 6 >= 6", tos[0])
	}
}

func TestVM_Exec_Shiftl(t *testing.T) {
	code := []byte{
		PUSH, 1, 0, 1,
		SHIFTL, 3,
		HALT,
	}

	vm := NewTestVM([]byte{})
	mc := NewMockContext(code)
	vm.context = mc
	vm.Exec(false)

	tos, _ := vm.evaluationStack.Pop()

	expected := 8
	actual := ByteArrayToInt(tos)

	if expected != actual {
		t.Errorf("Expected result to be '%v' but was '%v'", expected, actual)
	}
}

func TestVM_Exec_Shiftr(t *testing.T) {
	code := []byte{
		PUSH, 1, 0, 8,
		SHIFTR, 3,
		HALT,
	}

	vm := NewTestVM([]byte{})
	mc := NewMockContext(code)
	vm.context = mc
	vm.Exec(false)

	tos, _ := vm.evaluationStack.Pop()

	expected := 1
	actual := ByteArrayToInt(tos)

	if expected != actual {
		t.Errorf("Expected result to be '%v' but was '%v'", expected, actual)
	}
}

func TestVM_Exec_Jmpif(t *testing.T) {
	code := []byte{
		PUSH, 1, 0, 3,
		PUSH, 1, 0, 4,
		ADD,
		PUSH, 1, 0, 20,
		LT,
		JMPIF, 0, 21,
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

	tos, _ := vm.evaluationStack.Pop()

	expected := 3
	actual := ByteArrayToInt(tos)

	if expected != actual {
		t.Errorf("Expected result to be '%v' but was '%v'", expected, actual)
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

	tos, _ := vm.evaluationStack.Pop()

	expected := 2
	actual := ByteArrayToInt(tos)

	if expected != actual {
		t.Errorf("Expected result to be '%v' but was '%v'", expected, actual)
	}

	expected = 0
	actual = vm.callStack.GetLength()
	if expected != actual {
		t.Errorf("After calling and returning, callStack lenght should be %v, but was %v", expected, actual)
	}
}

func TestVM_Exec_Callif_true(t *testing.T) {
	code := []byte{
		PUSH, 1, 0, 10,
		PUSH, 1, 0, 8,
		PUSH, 1, 0, 10,
		PUSH, 1, 0, 10,
		EQ,
		CALLIF, 0, 24, 2,
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

	tos, _ := vm.evaluationStack.Pop()

	expected := 2
	actual := ByteArrayToInt(tos)

	if expected != actual {
		t.Errorf("Expected result to be '%v' but was '%v'", expected, actual)
	}

	expected = 0
	actual = vm.callStack.GetLength()
	if expected != actual {
		t.Errorf("After calling and returning, callStack lenght should be %v, but was %v", expected, actual)
	}
}

func TestVM_Exec_Callif_false(t *testing.T) {
	code := []byte{
		PUSH, 1, 0, 10,
		PUSH, 1, 0, 8,
		PUSH, 1, 0, 10,
		PUSH, 1, 0, 2,
		EQ,
		CALLIF, 0, 25, 2,
		HALT,
		NOP,
		NOP,
		LOAD, 0, // Begin of called function at address 21
		LOAD, 1,
		SUB,
		RET,
	}

	vm := NewTestVM([]byte{})
	mc := NewMockContext(code)
	vm.context = mc
	vm.Exec(false)

	tos, _ := vm.evaluationStack.Pop()

	expected := 8
	actual := ByteArrayToInt(tos)

	if expected != actual {
		t.Errorf("Expected result to be '%v' but was '%v'", expected, actual)
	}

	expected = 0
	actual = vm.callStack.GetLength()
	if expected != actual {
		t.Errorf("After skipping callif, callStack lenght should be '%v', but was '%v'", expected, actual)
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
		t.Errorf("%v", err)
	}

	expected := 3
	actual := ByteArrayToInt(tos)

	if expected != actual {
		t.Errorf("Expected element size to be '%v' but was '%v'", expected, actual)
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

	expected := []byte{0}
	actual, _ := vm.evaluationStack.Pop()

	if !bytes.Equal(expected, actual) {
		t.Errorf("Expected result to be '%v' but was '%v'", expected, actual)
	}

	result, err := vm.evaluationStack.Pop()
	if err != nil {
		t.Errorf("%v", err)
	}

	expectedString := "Hi There!!"
	actualString := string(result)
	if expectedString != actualString {
		t.Errorf("The String on the Stack should be '%v' but was %v", expectedString, actualString)
	}

	expected = []byte{26}
	actual, _ = vm.evaluationStack.Pop()

	if !bytes.Equal(expected, actual) {
		t.Errorf("Expected result to be '%v' but was '%v'", expected, actual)
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
	mc.Fee = 100000
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
	tos, _ := vm.evaluationStack.Pop()

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
	tos, _ := vm.evaluationStack.Pop()

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
	tos, _ := vm.evaluationStack.Pop()

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
	tos, _ := vm.evaluationStack.Pop()

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
	mc.Data = td

	vm.context = mc
	vm.Exec(false)

	functionHash, _ := vm.evaluationStack.Pop()

	if !bytes.Equal(functionHash, td[5:]) {
		t.Errorf("expected '%# x' but got '%# x'", td[5:], functionHash)
	}

	arg1, _ := vm.evaluationStack.Pop()
	if !bytes.Equal(arg1, td[3:4]) {
		t.Errorf("expected '%# x' but got '%# x'", td[3:4], arg1)
	}

	arg2, _ := vm.evaluationStack.Pop()
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

	actual, _ := vm.evaluationStack.Pop()
	expected := []byte{227, 237, 86, 189, 8, 109, 137, 88, 72, 58, 18, 115, 79, 160, 174, 127, 92, 139, 177, 96, 239, 144, 146, 198, 126, 130, 237, 155, 25, 228, 199, 178}
	if !bytes.Equal(actual, expected) {
		t.Errorf("Expected value to be \n '%v', \n but was \n '%v' \n after jumping to halt", expected, actual)
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

	expected := 4
	actual := ByteArrayToInt(tos)
	if actual != expected {
		t.Errorf("Expected result to be '%v' but was '%v'", expected, actual)
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

	actual, err := vm.evaluationStack.Pop()
	if err != nil {
		t.Errorf("%v", err)
	}

	expected := []byte{0x01, 0x00, 0x00}

	if !bytes.Equal(expected, actual) {
		t.Errorf("expected the Value of the new Map to be '[%v]' but was '[%v]'", expected, actual)
	}
}

func TestVM_Exec_MapHasKey_true(t *testing.T) {
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

		MAPHASKEY,

		HALT,
	}

	vm := NewTestVM([]byte{})
	mc := NewMockContext(code)
	vm.context = mc

	exec := vm.Exec(false)
	if !exec {
		errorMessage, _ := vm.evaluationStack.Pop()
		t.Errorf("VM.Exec terminated with Error: %v", string(errorMessage))
	}

	tos, err := vm.evaluationStack.Pop()
	if err != nil {
		t.Errorf("%v", err)
	}

	expected := true //Just for readability
	actual := ByteArrayToBool(tos)
	if expected != actual {
		t.Errorf("invalid value, Expected '%v' but was '%v'", expected, actual)
	}
}

func TestVM_Exec_MapHasKey_false(t *testing.T) {
	code := []byte{
		PUSH, 0x00, 0x06, //The key for MAPGETVAL

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

		MAPHASKEY,

		HALT,
	}

	vm := NewTestVM([]byte{})
	mc := NewMockContext(code)
	vm.context = mc

	exec := vm.Exec(false)
	if !exec {
		errorMessage, _ := vm.evaluationStack.Pop()
		t.Errorf("VM.Exec terminated with Error: %v", string(errorMessage))
	}

	tos, err := vm.evaluationStack.Pop()
	if err != nil {
		t.Errorf("%v", err)
	}

	expected := false //Just for readability
	actual := ByteArrayToBool(tos)
	if expected != actual {
		t.Errorf("invalid value, Expected '%v' but was '%v'", expected, actual)
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
		errorMessage, _ := vm.evaluationStack.Pop()
		t.Errorf("VM.Exec terminated with Error: %v", string(errorMessage))
	}

	m, err := vm.evaluationStack.Pop()
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
	mc.Fee = 300
	vm.context = mc

	exec := vm.Exec(false)
	if !exec {
		errorMessage, _ := vm.evaluationStack.Pop()
		t.Errorf("VM.Exec terminated with Error: %v", string(errorMessage))
	}

	actual, err := vm.evaluationStack.Pop()
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
		PUSH, 0x01, 0x55, 0x55, //Value to be reset by MAPSETVAL
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
	mc.Fee = 300
	vm.context = mc
	exec := vm.Exec(false)

	if !exec {
		errorMessage, _ := vm.evaluationStack.Pop()
		t.Errorf("VM.Exec terminated with Error: %v", string(errorMessage))
	}

	mbi, err := vm.evaluationStack.Pop()
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
	mc.Fee = 300
	vm.context = mc

	exec := vm.Exec(false)
	if !exec {
		errorMessage, _ := vm.evaluationStack.Pop()
		t.Errorf("VM.Exec terminated with Error: %v", string(errorMessage))
	}

	mapAsByteArray, err := vm.evaluationStack.Pop()
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
		t.Errorf("VM.Exec terminated with Error: %v", string(errorMessage))
	}

	arr, err := vm.evaluationStack.Pop()
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
		PUSH, 0x01, 0xFF, 0x00,
		NEWARR,
		ARRAPPEND,
		HALT,
	}

	vm := NewTestVM([]byte{})
	mc := NewMockContext(code)
	vm.context = mc
	exec := vm.Exec(false)
	if !exec {
		errorMessage, _ := vm.evaluationStack.Pop()
		t.Errorf("VM.Exec terminated with Error: %v", string(errorMessage))
	}

	arr, err := vm.evaluationStack.Pop()
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
		PUSH, 0x01, 0x00, 0x00,
		PUSH, 0x01, 0x00, 0x00,

		PUSH, 0x01, 0xFF, 0x00,

		PUSH, 0x01, 0xFF, 0x00,
		NEWARR,
		ARRAPPEND,
		ARRAPPEND,
		ARRINSERT,
		HALT,
	}

	vm := NewTestVM([]byte{})
	mc := NewMockContext(code)
	mc.Fee = 300
	vm.context = mc
	exec := vm.Exec(false)
	if !exec {
		errorMessage, _ := vm.evaluationStack.Pop()
		t.Errorf("VM.Exec terminated with Error: %v", string(errorMessage))
	}

	actual, err := vm.evaluationStack.Pop()
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
		PUSH, 0x01, 0x01, 0x00, //Index of element to remove
		PUSH, 0x01, 0xBB, 0x00,
		PUSH, 0x01, 0xAA, 0x00,
		PUSH, 0x01, 0xFF, 0x00,

		NEWARR,

		ARRAPPEND,
		ARRAPPEND,
		ARRAPPEND,
		ARRREMOVE,
		HALT,
	}

	vm := NewTestVM([]byte{})
	mc := NewMockContext(code)
	mc.Fee = 300
	vm.context = mc
	exec := vm.Exec(false)

	if !exec {
		errorMessage, _ := vm.evaluationStack.Pop()
		t.Errorf("VM.Exec terminated with Error: %v", string(errorMessage))
	}

	a, err := vm.evaluationStack.Pop()
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
		PUSH, 0x01, 0x02, 0x00, // index for ARRAT
		PUSH, 0x01, 0xBB, 0x00,
		PUSH, 0x01, 0xAA, 0x00,
		PUSH, 0x01, 0xFF, 0x00,

		NEWARR,

		ARRAPPEND,
		ARRAPPEND,
		ARRAPPEND,

		ARRAT,
		HALT,
	}

	vm := NewTestVM([]byte{})
	mc := NewMockContext(code)
	mc.Fee = 200
	vm.context = mc
	exec := vm.Exec(false)

	if !exec {
		errorMessage, _ := vm.evaluationStack.Pop()
		t.Errorf("VM.Exec terminated with Error: %v", string(errorMessage))
	}

	actual, err1 := vm.evaluationStack.Pop()

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
	actual := string(tos)
	if actual != expected {
		t.Errorf("Expected error message to be '%v' but was '%v'", expected, actual)
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

	expected := "push: Instruction set out of bounds"
	actual := string(tos)
	if actual != expected {
		t.Errorf("Expected error message to be '%v' but was '%v'", expected, actual)
	}
}

func TestVM_Exec_PopOnEmptyStack(t *testing.T) {
	code := []byte{
		PUSH, 0x00, 0x01,
		SHA3,
		SUB, 0x02, 0x03,
	}

	vm := NewTestVM([]byte{})
	mc := NewMockContext(code)
	vm.context = mc
	mc.Fee = 100
	vm.Exec(false)

	tos, _ := vm.evaluationStack.Pop()

	expected := "sub: Invalid signing bit"
	actual := string(tos)
	if actual != expected {
		t.Errorf("Expected error message to be '%v' but was '%v'", expected, actual)
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

	expected := "vm.exec(): Instruction set out of bounds"
	actual := string(tos)
	if actual != expected {
		t.Errorf("Expected error message to be '%v' but was '%v'", expected, actual)
	}
}

func TestVM_Exec_FuzzReproduction_InstructionSetOutOfBounds2(t *testing.T) {
	code := []byte{
		CALLEXT, 231,
	}

	vm := NewTestVM([]byte{})
	mc := NewMockContext(code)
	vm.context = mc
	mc.Fee = 100000
	vm.Exec(false)

	tos, _ := vm.evaluationStack.Pop()

	expected := "callext: Instruction set out of bounds"
	actual := string(tos)
	if actual != expected {
		t.Errorf("Expected error message to be '%v' but was '%v'", expected, actual)
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
	actual := string(tos)
	if actual != expected {
		t.Errorf("Expected error message to be '%v' but was '%v'", expected, actual)
	}
}

func TestVM_Exec_FuzzReproduction_IndexOutOfBounds2(t *testing.T) {
	code := []byte{
		PUSH, 4, 46, 110, 66, 50, 255, SSTORE, 123, 119,
	}

	vm := NewTestVM([]byte{})
	mc := NewMockContext(code)
	vm.context = mc
	mc.Fee = 100000
	vm.Exec(false)

	tos, _ := vm.evaluationStack.Pop()

	expected := "sstore: Index out of bounds"
	actual := string(tos)
	if actual != expected {
		t.Errorf("Expected error message to be '%v' but was '%v'", expected, actual)
	}
}

func TestVM_Exec_FunctionCallSub(t *testing.T) {
	code := []byte{
		// start ABI
		CALLDATA,
		DUP,
		PUSH, 1, 0, 1,
		EQ,
		JMPIF, 0, 20,
		DUP,
		PUSH, 1, 0, 2,
		EQ,
		JMPIF, 0, 23,
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

	mc.Data = []byte{
		1, 0, 5,
		1, 0, 2,
		1, 0, 1, // Function hash
	}

	vm.context = mc
	vm.Exec(false)

	tos, _ := vm.evaluationStack.Pop()

	expected := 3
	actual := ByteArrayToInt(tos)
	if actual != expected {
		t.Errorf("Expected result to be '%v' but was '%v'", expected, actual)
	}
}

func TestVM_Exec_FunctionCall(t *testing.T) {
	code := []byte{
		// start ABI
		CALLDATA,
		DUP,
		PUSH, 1, 0, 1,
		EQ,
		JMPIF, 0, 20,
		DUP,
		PUSH, 1, 0, 2,
		EQ,
		JMPIF, 0, 23,
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

	mc.Data = []byte{
		1, 0, 2,
		1, 0, 5,
		1, 0, 2, // Function hash
	}

	vm.context = mc
	vm.Exec(false)

	tos, _ := vm.evaluationStack.Pop()

	expected := 7
	actual := ByteArrayToInt(tos)
	if actual != expected {
		t.Errorf("Expected result to be '%v' but was '%v'", expected, actual)
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

	expected := "arrat: pop() on empty stack"
	actual := string(tos)
	if actual != expected {
		t.Errorf("Expected error message to be '%v' but was '%v'", expected, actual)
	}
}

func TestVM_Exec_FuzzReproduction_ContextOpCode1(t *testing.T) {
	code := []byte{
		CALLER, CALLER, ARRAPPEND,
	}

	vm := NewTestVM([]byte{})
	mc := NewMockContext(code)
	mc.Fee = 200
	vm.context = mc
	vm.Exec(false)

	tos, _ := vm.evaluationStack.Pop()

	expected := "arrappend: not a valid array"
	actual := string(tos)
	if actual != expected {
		t.Errorf("Expected error message to be '%v' but was '%v'", expected, actual)
	}
}

func TestVM_Exec_FuzzReproduction_ContextOpCode2(t *testing.T) {
	code := []byte{
		ADDRESS, CALLER, ARRAPPEND,
	}

	vm := NewTestVM([]byte{})
	mc := NewMockContext(code)
	mc.Fee = 200
	vm.context = mc
	vm.Exec(false)

	tos, _ := vm.evaluationStack.Pop()

	expected := "arrappend: not a valid array"
	actual := string(tos)
	if actual != expected {
		t.Errorf("Expected error message to be '%v' but was '%v'", expected, actual)
	}
}

func TestVM_Exec_FuzzReproduction_EdgecaseLastOpcodePlusOne(t *testing.T) {
	code := []byte{
		HALT + 1,
	}

	vm := NewTestVM([]byte{})
	mc := NewMockContext(code)
	vm.context = mc
	vm.Exec(false)

	tos, _ := vm.evaluationStack.Pop()

	expected := "vm.exec(): Not a valid opCode"
	actual := string(tos)
	if actual != expected {
		t.Errorf("Expected error message to be '%v' but was '%v'", expected, actual)
	}
}

func TestVM_PopBytes(t *testing.T) {
	code := []byte{
		PUSH, 1, 0, 8,
		PUSH, 1, 0, 8,
		ADD,
		HALT,
	}

	vm := NewTestVM([]byte{})
	mc := NewMockContext(code)
	mc.Fee = 11
	vm.context = mc

	vm.Exec(false)

	tos, _ := vm.evaluationStack.Pop()

	expected := 16
	actual := ByteArrayToInt(tos)
	if actual != expected {
		t.Errorf("Expected ToS to be '%v' but was '%v'", expected, actual)
	}

	expectedFee := 4
	actualFee := vm.fee

	if int(actualFee) != expectedFee {
		t.Errorf("Expected actual fee to be '%v' but was '%v'", expected, actual)
	}
}

func TestVM_FuzzTest_Reproduction(t *testing.T) {
	code := []byte{
		42, 0, 11, 1, 155, 6, 4, 13, 80, 89, 144, 14, 178, 188, 176, 41, 215, 171, 74, 28, 97, 232, 200, 151, 211, 147, 185, 143, 13, 220, 87, 77, 33, 223, 218, 249, 39, 126, 162, 59, 136, 178, 192, 120, 189, 37, 32, 37, 99, 130, 12, 145, 66, 131, 252, 30, 213, 1, 193, 101, 2, 15, 216, 19, 252, 78, 121, 20, 24, 216,
	}

	vm := NewTestVM([]byte{})
	mc := NewMockContext(code)
	mc.Fee = 11
	vm.context = mc

	vm.Exec(false)
}

func TestVM_FuzzTest_Reproduction_IndexOutOfRange(t *testing.T) {
	code := []byte{
		36, 16, 19, 33, 46, 55, 188,
	}

	vm := NewTestVM([]byte{})
	mc := NewMockContext(code)
	mc.Fee = 100
	vm.context = mc

	vm.Exec(false)
}

func TestVM_GasCalculation(t *testing.T) {
	code := []byte{
		PUSH, 64, 0, 8, 179, 91, 9, 9, 6, 136, 231, 56, 7, 146, 99, 170, 98, 183, 40, 118, 185, 95,
		106, 14, 143, 25, 99, 79, 76, 222, 197, 5, 218, 90, 216, 47, 218, 74, 53, 139, 62, 28, 104,
		180, 139, 65, 103, 193, 244, 169, 85, 39, 160, 218, 158, 207, 118, 37, 78, 42, 186, 64, 4, 70, 70, 190, 177,
		PUSH, 1, 0, 8,
		ADD,
		HALT,
	}

	vm := NewTestVM([]byte{})
	mc := NewMockContext(code)
	mc.Fee = 11
	vm.context = mc

	vm.Exec(false)

	expectedFee := 2
	actualFee := vm.fee

	if int(actualFee) != expectedFee {
		t.Errorf("Expected actual fee to be '%v' but was '%v'", expectedFee, actualFee)
	}
}

func TestVM_PopBytesOutOfGas(t *testing.T) {
	code := []byte{
		PUSH, 1, 0, 8,
		PUSH, 1, 0, 8,
		ADD,
		HALT,
	}

	vm := NewTestVM([]byte{})
	mc := NewMockContext(code)
	mc.Fee = 3
	vm.context = mc

	vm.Exec(false)

	tos, _ := vm.evaluationStack.Pop()

	expected := "add: Out of gas"
	actual := string(tos)
	if actual != expected {
		t.Errorf("Expected ToS to be '%v' but was '%v'", expected, actual)
	}

	expectedFee := 0
	actualFee := vm.fee

	if int(actualFee) != expectedFee {
		t.Errorf("Expected actual fee to be '%v' but was '%v'", expected, actual)
	}
}

func BenchmarkVM_Exec_ModularExponentiation_GoImplementation(b *testing.B) {
	benchmarks := []struct {
		name string
		bLen int
	}{
		{"bIs32B", 32},
		{"bIs128B", 128},
		{"bIs255B", 255},
	}

	var base big.Int
	var exponent big.Int
	var modulus big.Int

	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			for n := 0; n < b.N; n++ {

				base.SetBytes(protocol.RandomBytesWithLength(bm.bLen))
				exponent.SetBytes(protocol.RandomBytesWithLength(1))
				modulus.SetBytes(protocol.RandomBytesWithLength(2))

				modularExpGo(base, exponent, modulus)
			}

			b.ReportAllocs()
		})
	}
}

func BenchmarkVM_Exec_ModularExponentiation_ContractImplementation(b *testing.B) {
	benchmarks := []struct {
		name string
		bLen int
	}{
		{"bIs32B", 32},
		{"bIs128B", 128},
		{"bIs255B", 255},
	}

	var base big.Int
	var exponent big.Int
	var modulus big.Int

	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			for n := 0; n < b.N; n++ {
				base.SetBytes(protocol.RandomBytesWithLength(bm.bLen))
				exponent.SetBytes(protocol.RandomBytesWithLength(1))
				modulus.SetBytes(protocol.RandomBytesWithLength(2))

				contract := modularExpContract(base, exponent, modulus)

				vm := NewTestVM([]byte{})
				mc := NewMockContext(contract)
				mc.Fee = 1000000000000
				vm.context = mc

				if vm.Exec(false) != true {
					tos, err := vm.evaluationStack.Pop()
					fmt.Println(string(tos), err)
					b.Fail()
				}
				vm.pc = 0
				mc.Fee = 10000000000000
			}

			b.ReportAllocs()
			fmt.Println(b.Name())
		})
	}
}

func modularExpGo(base big.Int, exponent big.Int, modulus big.Int) *big.Int {
	if modulus.Cmp(big.NewInt(0)) == 0 {
		return big.NewInt(0)
	}
	start := big.NewInt(1)
	c := big.NewInt(1)
	for i := new(big.Int).Set(start); i.Cmp(&exponent) < 0; i.Add(i, big.NewInt(1)) {
		c = c.Mul(c, &base)
		c = c.Mod(c, &modulus)
	}
	return c
}

func modularExpContract(base big.Int, exponent big.Int, modulus big.Int) []byte {
	baseVal := BigIntToPushableBytes(base)
	exponentVal := BigIntToPushableBytes(exponent)
	modulusVal := BigIntToPushableBytes(modulus)

	addressBeforeExp := UInt16ToByteArray(uint16(39) + uint16(len(baseVal)) + uint16(len(modulusVal)))
	addressAfterExp := UInt16ToByteArray(uint16(66) + uint16(len(baseVal)) + uint16(len(modulusVal)) + uint16(len(exponentVal)))
	addressForLoop := UInt16ToByteArray(uint16(20) + uint16(len(baseVal)) + uint16(len(modulusVal)) + uint16(len(exponentVal)))

	contract := []byte{
		PUSH,
	}
	contract = append(contract, baseVal...)
	contract = append(contract, PUSH)
	contract = append(contract, modulusVal...)
	contract = append(contract, []byte{
		DUP,
		PUSH, 1, 0, 0,
		EQ,
		JMPIF,
	}...)
	contract = append(contract, addressBeforeExp[1])
	contract = append(contract, addressBeforeExp[0])
	contract = append(contract, []byte{
		PUSH, 1, 0, 1, // Counter (c)
		PUSH, 1, 0, 0, //i
		PUSH,
	}...)
	contract = append(contract, exponentVal...)
	contract = append(contract, []byte{
		//LOOP start
		//Duplicate arguments
		ROLL, 2,
		DUP, //Stack: [[0 11 75] [0 11 75] [0 13] [0 0] [0 1] [0 4]]
		ROLL, 4,
		DUP, // STACK Stack: [[04] [0 4] [0 11 75] [0 11 75] [0 13] [0 0] [0 1]]
		// PUT in order
		ROLL, 1, //Stack: [[0 11 75] [0 4] [0 4] [0 11 75] [0 13] [0 0] [0 1]]
		ROLL, 4, //Stack: [[0 0] [0 11 75] [0 4] [0 4] [0 11 75] [0 13] [0 1]]
		ROLL, 4, //Stack: [[0 13] [0 0] [0 11 75] [0 4] [0 4] [0 11 75] [0 1]]
		ROLL, 3, //Stack: [[0 4] [0 13] [0 0] [0 11 75] [0 4] [0 11 75] [0 1]]
		ROLL, 4, //Stack: [[0 11 75] [0 4] [0 13] [0 0] [0 11 75] [0 4] [0 1]]
		ROLL, 5, //Stack: [[0 1] [0 11 75] [0 4] [0 13] [0 0] [0 11 75] [0 4]]
		// Order: counter, modulus, base, exp, i, modulus, base
		CALL,
	}...)
	contract = append(contract, byte(addressAfterExp[1]))
	contract = append(contract, byte(addressAfterExp[0]))
	contract = append(contract, []byte{
		3,
		// PUT in order
		ROLL, 1,
		ROLL, 1,

		// Order: exp, i - counter, modulus, base,
		DUP,
		ROLL, 1,
		PUSH, 1, 0, 1,
		ADD,
		DUP,
		ROLL, 1,
		ROLL, 1,
		ROLL, 2,
		LT,
		JMPIF,
	}...)
	contract = append(contract, addressForLoop[1])
	contract = append(contract, addressForLoop[0])
	contract = append(contract, []byte{
		// LOOP END
		HALT,

		// FUNCTION Order: c, modulus, base,
		LOAD, 2,
		LOAD, 0,
		MULT,
		LOAD, 1,
		MOD,
		RET,
	}...)

	return contract
}

func TestVm_Exec_Loop(t *testing.T) {
	code := []byte{
		PUSH, 1, 0, 0, //i
		PUSH, 1, 0, 13, // Exp

		// Order: exp, i
		DUP,
		ROLL, 1,
		PUSH, 1, 0, 1,
		ADD,
		DUP,
		ROLL, 1,
		ROLL, 1,
		ROLL, 2,
		LT,
		JMPIF, 0, 8, // Adjust address
		// LOOP END
		HALT,
	}

	vm := NewTestVM([]byte{})
	mc := NewMockContext(code)
	mc.Fee = 1000
	vm.context = mc
	vm.Exec(false)

	expected := 13
	actual, _ := vm.evaluationStack.Pop()

	if ByteArrayToInt(actual[1:]) != expected {
		t.Errorf("Expected actual result to be '%v' but was '%v'", expected, actual)
	}
}

func TestVm_Exec_ModularExponentiation_ContractImplementation(t *testing.T) {
	code := []byte{
		PUSH, 1, 0, 4, // Base
		PUSH, 2, 0, 1, 241, // Modulus
		// IF modulus equals 0
		DUP,
		PUSH, 1, 0, 0,
		EQ,
		JMPIF, 0, 46, // Adjust address

		PUSH, 1, 0, 1, // Counter (c)
		PUSH, 1, 0, 0, //i
		PUSH, 1, 0, 13, // Exp

		//LOOP start
		//Duplicate arguments
		ROLL, 2,
		DUP, //Stack: [[0 11 75] [0 11 75] [0 13] [0 0] [0 1] [0 4]]
		ROLL, 4,
		DUP, // STACK Stack: [[04] [0 4] [0 11 75] [0 11 75] [0 13] [0 0] [0 1]]
		// PUT in order
		ROLL, 1, //Stack: [[0 11 75] [0 4] [0 4] [0 11 75] [0 13] [0 0] [0 1]]
		ROLL, 4, //Stack: [[0 0] [0 11 75] [0 4] [0 4] [0 11 75] [0 13] [0 1]]
		ROLL, 4, //Stack: [[0 13] [0 0] [0 11 75] [0 4] [0 4] [0 11 75] [0 1]]
		ROLL, 3, //Stack: [[0 4] [0 13] [0 0] [0 11 75] [0 4] [0 11 75] [0 1]]
		ROLL, 4, //Stack: [[0 11 75] [0 4] [0 13] [0 0] [0 11 75] [0 4] [0 1]]
		ROLL, 5, //Stack: [[0 1] [0 11 75] [0 4] [0 13] [0 0] [0 11 75] [0 4]]
		// Order: counter, modulus, base, exp, i, modulus, base
		CALL, 0, 76, 3,
		// PUT in order
		ROLL, 1,
		ROLL, 1,

		// Order: exp, i - counter, modulus, base,
		DUP,
		ROLL, 1,
		PUSH, 1, 0, 1,
		ADD,
		DUP,
		ROLL, 1,
		ROLL, 1,
		ROLL, 2,
		LT,
		JMPIF, 0, 30, // Adjust address
		// LOOP END
		HALT,

		// FUNCTION Order: c, modulus, base,
		LOAD, 2,
		LOAD, 0,
		MULT,
		LOAD, 1,
		MOD,
		RET,
	}

	vm := NewTestVM([]byte{})
	mc := NewMockContext(code)
	mc.Fee = 1000
	vm.context = mc
	vm.Exec(false)

	expected := 445
	vm.evaluationStack.Pop()
	vm.evaluationStack.Pop()
	actual, _ := vm.evaluationStack.Pop()

	if ByteArrayToInt(actual[1:]) != expected {
		t.Errorf("Expected actual result to be '%v' but was '%v'", expected, actual)
	}
}
