package vm

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"encoding/binary"
	"errors"
	"fmt"
	"math/big"

	"github.com/bazo-blockchain/bazo-miner/protocol"

	"golang.org/x/crypto/sha3"
)

type Context interface {
	GetContract() []byte
	GetContractVariable(index int) ([]byte, error)
	SetContractVariable(index int, value []byte) error
	GetAddress() [64]byte
	GetIssuer() [32]byte
	GetBalance() uint64
	GetSender() [32]byte
	GetAmount() uint64
	GetTransactionData() []byte
	GetFee() uint64
	GetSig1() [64]byte
}

type VM struct {
	code            []byte
	pc              int // Program counter
	fee             uint64
	evaluationStack *Stack
	callStack       *CallStack
	context         Context
}

func NewVM(context Context) VM {
	return VM{
		code:            []byte{},
		pc:              0,
		fee:             0,
		evaluationStack: NewStack(),
		callStack:       NewCallStack(),
		context:         context,
	}
}

func NewTestVM(byteCode []byte) VM {
	return VM{
		code:            []byte{},
		pc:              0,
		fee:             0,
		evaluationStack: NewStack(),
		callStack:       NewCallStack(),
		context:         NewMockContext(byteCode),
	}
}

// Private function, that can be activated by Exec call, useful for debugging
func (vm *VM) trace() {
	stack := vm.evaluationStack
	addr := vm.pc

	byteCode := int(vm.code[vm.pc])
	if len(OpCodes) <= byteCode {
		stack.Push([]byte("Trace: invalid opcode "))
		return
	}
	opCode := OpCodes[byteCode]

	var args []byte
	var formattedArgs string
	var counter int

	for _, argType := range opCode.ArgTypes {
		switch argType {
		case BYTES:
			if len(vm.code)-vm.pc > 0 {
				nargs := int(vm.code[vm.pc+1])

				if nargs < (len(vm.code)-vm.pc)-1 {
					args = vm.code[vm.pc+2+counter : vm.pc+nargs+3+counter]
					counter += nargs
					formattedArgs = formattedArgs + fmt.Sprintf("%v (bytes) ", args[:])
				}
			}

		case BYTE:
			if len(vm.code)-vm.pc > 0 {
				args = []byte{vm.code[vm.pc+1+counter]}
				counter += 1
				formattedArgs += fmt.Sprintf("%v (byte) ", args[:])
			}

		case ADDR:
			if len(vm.code)-vm.pc > 31 {
				args = vm.code[vm.pc+1+counter : vm.pc+33+counter]
				counter += 32
				formattedArgs += fmt.Sprintf("%v (bazo address) ", args[:])
			}

		case LABEL:
			if len(vm.code)-vm.pc > 1 {
				args = vm.code[vm.pc+1+counter : vm.pc+3+counter]
				counter += 2
				formattedArgs += fmt.Sprintf("%v (address) ", ByteArrayToInt(args[:]))
			}
		}
	}

	reversedStack := make([]protocol.ByteArray, stack.GetLength())
	maxIndex := len(stack.Stack) - 1
	for i := maxIndex; i >= 0; i-- {
		reversedStack[maxIndex-i] = stack.Stack[i]
	}

	fmt.Printf("\t  Stack: %v \n", reversedStack)
	fmt.Printf("\t  %v of max. %v Bytes in use \n", stack.memoryUsage, stack.memoryMax)
	fmt.Printf("⋅⋅⋅⋅⋅⋅⋅⋅⋅⋅⋅⋅⋅⋅⋅⋅⋅⋅⋅⋅⋅⋅⋅⋅⋅⋅⋅⋅⋅⋅⋅⋅⋅⋅⋅⋅⋅⋅⋅⋅⋅⋅⋅⋅⋅⋅⋅⋅⋅⋅⋅⋅⋅⋅⋅⋅⋅⋅⋅⋅⋅⋅⋅⋅⋅⋅⋅⋅⋅⋅⋅⋅⋅⋅⋅⋅⋅⋅⋅⋅⋅⋅⋅⋅⋅⋅⋅⋅⋅⋅⋅⋅⋅⋅⋅\n")
	fmt.Printf("%04d: %-6s %v \n", addr, opCode.Name, formattedArgs)
}

func (vm *VM) Exec(trace bool) bool {

	vm.code = vm.context.GetContract()
	vm.fee = vm.context.GetFee()

	if len(vm.code) > 100000 {
		vm.evaluationStack.Push([]byte("vm.exec(): Instruction set to big"))
		return false
	}

	// Infinite Loop until return called
	for {
		if trace {
			vm.trace()
		}

		// Fetch
		byteCode, err := vm.fetch("vm.exec()")
		if err != nil {
			vm.evaluationStack.Push([]byte("vm.exec(): " + err.Error()))
			return false
		}

		// Return false if instruction is not an opCode
		if len(OpCodes) <= int(byteCode) {
			vm.evaluationStack.Push([]byte("vm.exec(): Not a valid opCode"))
			return false
		}

		opCode := OpCodes[byteCode]
		// Subtract gas used for operation
		if vm.fee < opCode.gasPrice {
			vm.evaluationStack.Push([]byte("vm.exec(): out of gas"))
			return false
		} else {
			vm.fee -= opCode.gasPrice
		}

		// Decode
		switch opCode.code {

		case PUSH:
			arg, errArg1 := vm.fetch(opCode.Name)
			byteCount := int(arg) + 1 // Amount of bytes pushed, maximum amount of bytes that can be pushed is 256
			bytes, errArg2 := vm.fetchMany(opCode.Name, byteCount)

			if !vm.checkErrors(opCode.Name, errArg1, errArg2) {
				return false
			}

			err = vm.evaluationStack.Push(bytes)

			if err != nil {
				vm.evaluationStack.Push([]byte(opCode.Name + ": " + err.Error()))
				return false
			}

		case DUP:
			tos, err := vm.PopBytes(opCode)

			if !vm.checkErrors(opCode.Name, err) {
				return false
			}

			err = vm.evaluationStack.Push(tos)

			if err != nil {
				vm.evaluationStack.Push([]byte(opCode.Name + ": " + err.Error()))
				return false
			}

			err = vm.evaluationStack.Push(tos)

			if err != nil {
				vm.evaluationStack.Push([]byte(opCode.Name + ": " + err.Error()))
				return false
			}

		case ROLL:
			arg, err := vm.fetch(opCode.Name) // arg shows how many have to be rolled
			index := vm.evaluationStack.GetLength() - (int(arg) + 2)

			if !vm.checkErrors(opCode.Name, err) {
				return false
			}

			if index != -1 {
				if int(arg) >= vm.evaluationStack.GetLength() {
					vm.evaluationStack.Push([]byte(opCode.Name + ": index out of bounds"))
					return false
				}

				newTos, err := vm.evaluationStack.PopIndexAt(index)

				if err != nil {
					vm.evaluationStack.Push([]byte(opCode.Name + ": " + err.Error()))
					return false
				}

				err = vm.evaluationStack.Push(newTos)

				if err != nil {
					vm.evaluationStack.Push([]byte(opCode.Name + ": " + err.Error()))
					return false
				}
			}

		case POP:
			_, rerr := vm.PopBytes(opCode)
			if !vm.checkErrors(opCode.Name, rerr) {
				return false
			}

		case ADD:
			right, rerr := vm.PopSignedBigInt(opCode)
			left, lerr := vm.PopSignedBigInt(opCode)

			if !vm.checkErrors(opCode.Name, rerr, lerr) {
				return false
			}

			left.Add(&left, &right)
			err := vm.evaluationStack.Push(SignedByteArrayConversion(left))

			if err != nil {
				vm.evaluationStack.Push([]byte(opCode.Name + ": " + err.Error()))
				return false
			}

		case SUB:
			right, rerr := vm.PopSignedBigInt(opCode)
			left, lerr := vm.PopSignedBigInt(opCode)

			if !vm.checkErrors(opCode.Name, rerr, lerr) {
				return false
			}

			left.Sub(&left, &right)
			err := vm.evaluationStack.Push(SignedByteArrayConversion(left))

			if err != nil {
				vm.evaluationStack.Push([]byte(opCode.Name + ": " + err.Error()))
				return false
			}

		case MULT:
			right, rerr := vm.PopSignedBigInt(opCode)
			left, lerr := vm.PopSignedBigInt(opCode)

			if !vm.checkErrors(opCode.Name, rerr, lerr) {
				return false
			}

			left.Mul(&left, &right)
			err := vm.evaluationStack.Push(SignedByteArrayConversion(left))

			if err != nil {
				vm.evaluationStack.Push([]byte(opCode.Name + ": " + err.Error()))
				return false
			}

		case DIV:
			right, rerr := vm.PopSignedBigInt(opCode)
			left, lerr := vm.PopSignedBigInt(opCode)

			if !vm.checkErrors(opCode.Name, rerr, lerr) {
				return false
			}

			if right.Cmp(big.NewInt(0)) == 0 {
				vm.evaluationStack.Push([]byte(opCode.Name + ": Division by Zero"))
				return false
			}

			left.Div(&left, &right)
			err := vm.evaluationStack.Push(SignedByteArrayConversion(left))

			if err != nil {
				vm.evaluationStack.Push([]byte(opCode.Name + ": " + err.Error()))
				return false
			}

		case MOD:
			right, rerr := vm.PopSignedBigInt(opCode)
			left, lerr := vm.PopSignedBigInt(opCode)

			if !vm.checkErrors(opCode.Name, rerr, lerr) {
				return false
			}

			if right.Cmp(big.NewInt(0)) == 0 {
				vm.evaluationStack.Push([]byte(opCode.Name + ": Division by Zero"))
				return false
			}

			left.Mod(&left, &right)
			err := vm.evaluationStack.Push(SignedByteArrayConversion(left))

			if err != nil {
				vm.evaluationStack.Push([]byte(opCode.Name + ": " + err.Error()))
				return false
			}

		case NEG:
			tos, err := vm.PopSignedBigInt(opCode)

			if err != nil {
				vm.evaluationStack.Push([]byte(opCode.Name + ": " + err.Error()))
				return false
			}

			tos.Neg(&tos)

			vm.evaluationStack.Push(SignedByteArrayConversion(tos))

		case EQ:
			right, rerr := vm.PopUnsignedBigInt(opCode)
			left, lerr := vm.PopUnsignedBigInt(opCode)
			if !vm.checkErrors(opCode.Name, rerr, lerr) {
				return false
			}

			result := left.Cmp(&right) == 0
			vm.evaluationStack.Push(BoolToByteArray(result))

		case NEQ:
			right, rerr := vm.PopUnsignedBigInt(opCode)
			left, lerr := vm.PopUnsignedBigInt(opCode)
			if !vm.checkErrors(opCode.Name, rerr, lerr) {
				return false
			}

			result := left.Cmp(&right) != 0
			vm.evaluationStack.Push(BoolToByteArray(result))

		case LT:
			right, rerr := vm.PopSignedBigInt(opCode)
			left, lerr := vm.PopSignedBigInt(opCode)
			if !vm.checkErrors(opCode.Name, rerr, lerr) {
				return false
			}

			result := left.Cmp(&right) == -1
			vm.evaluationStack.Push(BoolToByteArray(result))

		case GT:
			right, rerr := vm.PopSignedBigInt(opCode)
			left, lerr := vm.PopSignedBigInt(opCode)
			if !vm.checkErrors(opCode.Name, rerr, lerr) {
				return false
			}

			result := left.Cmp(&right) == 1
			vm.evaluationStack.Push(BoolToByteArray(result))

		case LTE:
			right, rerr := vm.PopSignedBigInt(opCode)
			left, lerr := vm.PopSignedBigInt(opCode)
			if !vm.checkErrors(opCode.Name, rerr, lerr) {
				return false
			}

			result := left.Cmp(&right) == -1 || left.Cmp(&right) == 0
			vm.evaluationStack.Push(BoolToByteArray(result))

		case GTE:
			right, rerr := vm.PopSignedBigInt(opCode)
			left, lerr := vm.PopSignedBigInt(opCode)
			if !vm.checkErrors(opCode.Name, rerr, lerr) {
				return false
			}

			result := left.Cmp(&right) == 1 || left.Cmp(&right) == 0
			vm.evaluationStack.Push(BoolToByteArray(result))

		case SHIFTL:
			nrOfShifts, errArg := vm.fetch(opCode.Name)
			tos, errStack := vm.PopSignedBigInt(opCode)

			if !vm.checkErrors(opCode.Name, errArg, errStack) {
				return false
			}

			tos.Lsh(&tos, uint(nrOfShifts))
			err = vm.evaluationStack.Push(SignedByteArrayConversion(tos))

			if err != nil {
				vm.evaluationStack.Push([]byte(opCode.Name + ": " + err.Error()))
				return false
			}

		case SHIFTR:
			nrOfShifts, errArg := vm.fetch(opCode.Name)
			tos, errStack := vm.PopSignedBigInt(opCode)

			if !vm.checkErrors(opCode.Name, errArg, errStack) {
				return false
			}

			tos.Rsh(&tos, uint(nrOfShifts))
			err = vm.evaluationStack.Push(SignedByteArrayConversion(tos))

			if err != nil {
				vm.evaluationStack.Push([]byte(opCode.Name + ": " + err.Error()))
				return false
			}

		case NOP:
			_, err := vm.fetch(opCode.Name)

			if err != nil {
				vm.evaluationStack.Push([]byte(opCode.Name + ": " + err.Error()))
				return false
			}

		case JMP:
			nextInstruction, err := vm.fetchMany(opCode.Name, 2)

			if !vm.checkErrors(opCode.Name, err) {
				return false
			}

			var jumpTo big.Int
			jumpTo.SetBytes(nextInstruction)

			vm.pc = int(jumpTo.Int64())

		case JMPIF:
			nextInstruction, errArg := vm.fetchMany(opCode.Name, 2)
			right, errStack := vm.PopBytes(opCode)
			if !vm.checkErrors(opCode.Name, errArg, errStack) {
				return false
			}

			if ByteArrayToBool(right) {
				vm.pc = ByteArrayToInt(nextInstruction)
			}

		case CALL:
			returnAddressBytes, errArg1 := vm.fetchMany(opCode.Name, 2) // Shows where to jump after executing
			argsToLoad, errArg2 := vm.fetch(opCode.Name)                // Shows how many elements have to be popped from evaluationStack

			if !vm.checkErrors(opCode.Name, errArg1, errArg2) {
				return false
			}

			var returnAddress big.Int
			returnAddress.SetBytes(returnAddressBytes)

			if int(returnAddress.Int64()) == 0 || int(returnAddress.Int64()) > len(vm.code) {
				vm.evaluationStack.Push([]byte(opCode.Name + ": ReturnAddress out of bounds"))
				return false
			}

			frame := &Frame{returnAddress: vm.pc, variables: make(map[int]big.Int)}

			for i := int(argsToLoad) - 1; i >= 0; i-- {
				frame.variables[i], err = vm.PopUnsignedBigInt(opCode)
				if err != nil {
					vm.evaluationStack.Push([]byte(opCode.Name + ": " + err.Error()))
					return false
				}
			}

			vm.callStack.Push(frame)
			vm.pc = int(returnAddress.Int64())

		case CALLIF:
			returnAddressBytes, errArg1 := vm.fetchMany(opCode.Name, 2) // Shows where to jump after executing
			argsToLoad, errArg2 := vm.fetch(opCode.Name)                // Shows how many elements have to be popped from evaluationStack
			right, errStack := vm.PopBytes(opCode)

			if !vm.checkErrors(opCode.Name, errArg1, errArg2, errStack) {
				return false
			}

			if ByteArrayToBool(right) {
				var returnAddress big.Int
				returnAddress.SetBytes(returnAddressBytes)

				if int(returnAddress.Int64()) == 0 || int(returnAddress.Int64()) > len(vm.code) {
					vm.evaluationStack.Push([]byte(opCode.Name + ": ReturnAddress out of bounds"))
					return false
				}

				frame := &Frame{returnAddress: vm.pc, variables: make(map[int]big.Int)}

				for i := int(argsToLoad) - 1; i >= 0; i-- {
					frame.variables[i], err = vm.PopUnsignedBigInt(opCode)
					if err != nil {
						vm.evaluationStack.Push([]byte(opCode.Name + ": " + err.Error()))
						return false
					}
				}
				vm.callStack.Push(frame)
				vm.pc = int(returnAddress.Int64())
			}

		case CALLEXT:
			transactionAddress, errArg1 := vm.fetchMany(opCode.Name, 32) // Addresses are 32 bytes (var name: transactionAddress)
			functionHash, errArg2 := vm.fetchMany(opCode.Name, 4)        // Function hash identifies function in external smart contract, first 4 byte of SHA3 hash (var name: functionHash)
			argsToLoad, errArg3 := vm.fetch(opCode.Name)                 // Shows how many arguments to pop from stack and pass to external function (var name: argsToLoad)

			if !vm.checkErrors(opCode.Name, errArg1, errArg2, errArg3) {
				return false
			}

			fmt.Sprint("CALLEXT", transactionAddress, functionHash, argsToLoad)
			//TODO: Invoke new transaction with function hash and arguments, waiting for integration in bazo blockchain to finish

		case RET:
			callstackTos, err := vm.callStack.Peek()

			if !vm.checkErrors(opCode.Name, err) {
				return false
			}

			vm.callStack.Pop()
			vm.pc = callstackTos.returnAddress

		case SIZE:
			element, err := vm.PopBytes(opCode)
			if err != nil {
				vm.evaluationStack.Push([]byte(opCode.Name + ": " + err.Error()))
				return false
			}

			size := UInt64ToByteArray(uint64(len(element)))

			err = vm.evaluationStack.Push(size)
			if err != nil {
				vm.evaluationStack.Push([]byte(opCode.Name + ": " + err.Error()))
				return false
			}

		case SSTORE:
			index, errArgs := vm.fetch(opCode.Name)
			value, errStack := vm.PopBytes(opCode)
			if !vm.checkErrors(opCode.Name, errArgs, errStack) {
				return false
			}

			err = vm.context.SetContractVariable(int(index), value)
			if err != nil {
				vm.evaluationStack.Push([]byte(opCode.Name + ": " + err.Error()))
				return false
			}

		case STORE:
			address, errArgs := vm.fetch(opCode.Name)
			right, errStack := vm.PopSignedBigInt(opCode)

			if !vm.checkErrors(opCode.Name, errArgs, errStack) {
				return false
			}

			callstackTos, err := vm.callStack.Peek()

			if err != nil {
				vm.evaluationStack.Push([]byte(opCode.Name + ": " + err.Error()))
				return false
			}

			callstackTos.variables[int(address)] = right

		case SLOAD:
			index, err := vm.fetch(opCode.Name)
			if !vm.checkErrors(opCode.Name, err) {
				return false
			}

			value, err := vm.context.GetContractVariable(int(index))
			if err != nil {
				vm.evaluationStack.Push([]byte(opCode.Name + ": " + err.Error()))
				return false
			}

			err = vm.evaluationStack.Push(value)
			if err != nil {
				vm.evaluationStack.Push([]byte(opCode.Name + ": " + err.Error()))
				return false
			}

		case LOAD:
			address, errArg := vm.fetch(opCode.Name)
			callstackTos, errCallStack := vm.callStack.Peek()

			if !vm.checkErrors(opCode.Name, errArg, errCallStack) {
				return false
			}

			val := callstackTos.variables[int(address)]

			err := vm.evaluationStack.Push(SignedByteArrayConversion(val))

			if err != nil {
				vm.evaluationStack.Push([]byte(opCode.Name + ": " + err.Error()))
				return false
			}

		case ADDRESS:
			address := vm.context.GetAddress()
			err := vm.evaluationStack.Push(address[:])

			if err != nil {
				vm.evaluationStack.Push([]byte(opCode.Name + ": " + err.Error()))
				return false
			}

		case ISSUER:
			issuer := vm.context.GetIssuer()
			err := vm.evaluationStack.Push(issuer[:])

			if err != nil {
				vm.evaluationStack.Push([]byte(opCode.Name + ": " + err.Error()))
				return false
			}

		case BALANCE:
			balance := make([]byte, 8)
			binary.LittleEndian.PutUint64(balance, vm.context.GetBalance())

			err := vm.evaluationStack.Push(balance)

			if err != nil {
				vm.evaluationStack.Push([]byte(opCode.Name + ": " + err.Error()))
				return false
			}

		case CALLER:
			caller := vm.context.GetSender()
			err := vm.evaluationStack.Push(caller[:])

			if err != nil {
				vm.evaluationStack.Push([]byte(opCode.Name + ": " + err.Error()))
				return false
			}

		case CALLVAL:
			value := make([]byte, 8)
			binary.LittleEndian.PutUint64(value, vm.context.GetAmount())

			err := vm.evaluationStack.Push(value[:])

			if err != nil {
				vm.evaluationStack.Push([]byte(opCode.Name + ": " + err.Error()))
				return false
			}

		case CALLDATA:
			td := vm.context.GetTransactionData()
			for i := 0; i < len(td); i++ {
				length := int(td[i]) // Length of parameters

				if len(td)-i-1 <= length {
					vm.evaluationStack.Push([]byte(opCode.Name + ": Index out of bounds"))
					return false
				}

				err := vm.evaluationStack.Push(td[i+1 : i+length+2])

				if err != nil {
					vm.evaluationStack.Push([]byte(opCode.Name + ": " + err.Error()))
					return false
				}

				i += int(td[i]) + 1 // Increase to next parameter length
			}

		case NEWMAP:
			m := NewMap()

			err = vm.evaluationStack.Push(m)
			if err != nil {
				vm.evaluationStack.Push([]byte(opCode.Name + ": " + err.Error()))
				return false
			}

		case MAPHASKEY:
			mba, err := vm.PopBytes(opCode)
			if err != nil {
				vm.evaluationStack.Push([]byte(opCode.Name + ": " + err.Error()))
				return false
			}

			m, err := MapFromByteArray(mba)
			if err != nil {
				vm.evaluationStack.Push([]byte(opCode.Name + ": " + err.Error()))
				return false
			}

			k, err := vm.PopBytes(opCode)
			if err != nil {
				vm.evaluationStack.Push([]byte(opCode.Name + ": " + err.Error()))
				return false
			}

			result, err := m.MapContainsKey(k)
			if err != nil {
				vm.evaluationStack.Push([]byte(opCode.Name + ": " + err.Error()))
				return false
			}

			vm.evaluationStack.Push(BoolToByteArray(result))

		case MAPPUSH:
			mba, err := vm.PopBytes(opCode)
			if err != nil {
				vm.evaluationStack.Push([]byte(opCode.Name + ": " + err.Error()))
				return false
			}

			k, err := vm.PopBytes(opCode)
			if err != nil {
				vm.evaluationStack.Push([]byte(opCode.Name + ": " + err.Error()))
				return false
			}

			v, err := vm.PopBytes(opCode)
			if err != nil {
				vm.evaluationStack.Push([]byte(opCode.Name + ": " + err.Error()))
				return false
			}

			m, err := MapFromByteArray(mba)
			if err != nil {
				vm.evaluationStack.Push([]byte(opCode.Name + ": " + err.Error()))
				return false
			}

			err = m.Append(k, v)
			if err != nil {
				vm.evaluationStack.Push([]byte(opCode.Name + ": " + err.Error()))
				return false
			}

			err = vm.evaluationStack.Push(m)
			if err != nil {
				vm.evaluationStack.Push([]byte(opCode.Name + ": " + err.Error()))
				return false
			}

		case MAPGETVAL:
			mapAsByteArray, err := vm.PopBytes(opCode)
			if err != nil {
				vm.evaluationStack.Push([]byte(opCode.Name + ": " + err.Error()))
				return false
			}

			k, err := vm.PopBytes(opCode)
			if err != nil {
				vm.evaluationStack.Push([]byte(opCode.Name + ": " + err.Error()))
				return false
			}

			m, err := MapFromByteArray(mapAsByteArray)
			if err != nil {
				vm.evaluationStack.Push([]byte(opCode.Name + ": " + err.Error()))
				return false
			}

			v, err := m.GetVal(k)
			if err != nil {
				vm.evaluationStack.Push([]byte(opCode.Name + ": " + err.Error()))
				return false
			}

			err = vm.evaluationStack.Push(v)
			if err != nil {
				vm.evaluationStack.Push([]byte(opCode.Name + ": " + err.Error()))
				return false
			}

		case MAPSETVAL:
			mapAsByteArray, err := vm.PopBytes(opCode)
			if err != nil {
				vm.evaluationStack.Push([]byte(opCode.Name + ": " + err.Error()))
				return false
			}

			m, err := MapFromByteArray(mapAsByteArray)
			if err != nil {
				vm.evaluationStack.Push([]byte(opCode.Name + ": " + err.Error()))
				return false
			}

			k, err := vm.PopBytes(opCode)
			if err != nil {
				vm.evaluationStack.Push([]byte(opCode.Name + ": " + err.Error()))
				return false
			}

			v, err := vm.PopBytes(opCode)
			if err != nil {
				vm.evaluationStack.Push([]byte(opCode.Name + ": " + err.Error()))
				return false
			}

			err = m.SetVal(k, v)
			if err != nil {
				vm.evaluationStack.Push([]byte(opCode.Name + ": " + err.Error()))
				return false
			}

			err = vm.evaluationStack.Push(m)
			if err != nil {
				vm.evaluationStack.Push([]byte(opCode.Name + ": " + err.Error()))
				return false
			}

		case MAPREMOVE:
			mapAsByteArray, err := vm.PopBytes(opCode)
			if err != nil {
				vm.evaluationStack.Push([]byte(opCode.Name + ": " + err.Error()))
				return false
			}

			k, err := vm.PopBytes(opCode)
			if err != nil {
				vm.evaluationStack.Push([]byte(opCode.Name + ": " + err.Error()))
				return false
			}

			m, err := MapFromByteArray(mapAsByteArray)
			if err != nil {
				vm.evaluationStack.Push([]byte(opCode.Name + ": " + err.Error()))
				return false
			}

			err = m.Remove(k)
			if err != nil {
				vm.evaluationStack.Push([]byte(opCode.Name + ": " + err.Error()))
				return false
			}

			err = vm.evaluationStack.Push(m)
			if err != nil {
				vm.evaluationStack.Push([]byte(opCode.Name + ": " + err.Error()))
				return false
			}

		case NEWARR:
			a := NewArray()
			vm.evaluationStack.Push(a)

		case ARRAPPEND:
			a, aerr := vm.PopBytes(opCode)
			v, verr := vm.PopBytes(opCode)
			if !vm.checkErrors(opCode.Name, verr, aerr) {
				return false
			}

			arr, err := ArrayFromByteArray(a)
			if err != nil {
				vm.evaluationStack.Push([]byte(opCode.Name + ": " + err.Error()))
				return false
			}

			err = arr.Append(v)
			if err != nil {
				vm.evaluationStack.Push([]byte(opCode.Name + ": Invalid argument size of ARRAPPEND"))
				return false
			}

			err = vm.evaluationStack.Push(arr)
			if err != nil {
				vm.evaluationStack.Push([]byte(opCode.Name + ": " + err.Error()))
				return false
			}

		case ARRINSERT:
			a, err := vm.PopBytes(opCode)
			if err != nil {
				vm.evaluationStack.Push([]byte(opCode.Name + ": " + err.Error()))
				return false
			}

			i, err := vm.PopBytes(opCode)
			if err != nil {
				vm.evaluationStack.Push([]byte(opCode.Name + ": " + err.Error()))
				return false
			}

			if len(i) > 2 {
				vm.evaluationStack.Push([]byte(opCode.Name + ": Wrong index size"))
				return false
			}

			element, err := vm.PopBytes(opCode)
			if err != nil {
				vm.evaluationStack.Push([]byte(opCode.Name + ": " + err.Error()))
				return false
			}

			arr, err := ArrayFromByteArray(a)
			if err != nil {
				vm.evaluationStack.Push([]byte(opCode.Name + ": " + err.Error()))
				return false
			}

			index, err := ByteArrayToUI16(i)
			if err != nil {
				vm.evaluationStack.Push([]byte(opCode.Name + ": " + err.Error()))
				return false
			}

			size, err := arr.getSize()
			if err != nil {
				vm.evaluationStack.Push([]byte(opCode.Name + ": " + err.Error()))
				return false
			}

			if index >= size {
				vm.evaluationStack.Push([]byte(opCode.Name + ": Index out of bounds"))
				return false
			}

			err = arr.Insert(index, element)
			if err != nil {
				vm.evaluationStack.Push([]byte(opCode.Name + ": " + err.Error()))
				return false
			}

			err = vm.evaluationStack.Push(arr)
			if err != nil {
				vm.evaluationStack.Push([]byte(opCode.Name + ": " + err.Error()))
				return false
			}

		case ARRREMOVE:
			a, err := vm.PopBytes(opCode)
			if err != nil {
				vm.evaluationStack.Push([]byte(opCode.Name + ": " + err.Error()))
				return false
			}

			i, err := vm.PopBytes(opCode)
			if err != nil {
				vm.evaluationStack.Push([]byte(opCode.Name + ": " + err.Error()))
				return false
			}

			if len(i) > 2 {
				vm.evaluationStack.Push([]byte(opCode.Name + ": Wrong index size"))
				return false
			}

			index, err := ByteArrayToUI16(i)
			if err != nil {
				vm.evaluationStack.Push([]byte(opCode.Name + ": " + err.Error()))
				return false
			}

			arr, err := ArrayFromByteArray(a)
			if err != nil {
				vm.evaluationStack.Push([]byte(opCode.Name + ": " + err.Error()))
				return false
			}

			err = arr.Remove(index)
			if err != nil {
				vm.evaluationStack.Push([]byte(opCode.Name + ": " + err.Error()))
				return false
			}

			err = vm.evaluationStack.Push(arr)
			if err != nil {
				vm.evaluationStack.Push([]byte(opCode.Name + ": " + err.Error()))
				return false
			}

		case ARRAT:
			a, err := vm.PopBytes(opCode)
			if err != nil {
				vm.evaluationStack.Push([]byte(opCode.Name + ": " + err.Error()))
				return false
			}

			i, err := vm.PopBytes(opCode)
			if err != nil {
				vm.evaluationStack.Push([]byte(opCode.Name + ": " + err.Error()))
				return false
			}

			index, err := ByteArrayToUI16(i)
			if err != nil {
				vm.evaluationStack.Push([]byte(opCode.Name + ": " + err.Error()))
				return false
			}

			arr, err := ArrayFromByteArray(a)
			if err != nil {
				vm.evaluationStack.Push([]byte(opCode.Name + ": " + err.Error()))
				return false
			}

			element, err := arr.At(index)
			if err != nil {
				vm.evaluationStack.Push([]byte(opCode.Name + ": " + err.Error()))
				return false
			}

			err = vm.evaluationStack.Push(element)
			if err != nil {
				vm.evaluationStack.Push([]byte(opCode.Name + ": " + err.Error()))
				return false
			}

		case SHA3:
			right, err := vm.PopBytes(opCode)
			if err != nil {
				vm.evaluationStack.Push([]byte(opCode.Name + ": " + err.Error()))
				return false
			}

			hasher := sha3.New256()
			hasher.Write(right)
			hash := hasher.Sum(nil)

			err = vm.evaluationStack.Push(hash)
			if err != nil {
				vm.evaluationStack.Push([]byte(opCode.Name + ": " + err.Error()))
				return false
			}

		case CHECKSIG:
			publicKeySig, errArg1 := vm.PopBytes(opCode)
			hash, errArg2 := vm.PopBytes(opCode)

			if !vm.checkErrors(opCode.Name, errArg1, errArg2) {
				return false
			}

			if len(publicKeySig) != 64 {
				vm.evaluationStack.Push([]byte(opCode.Name + ": Not a valid address"))
				return false
			}

			if len(hash) != 32 {
				vm.evaluationStack.Push([]byte(opCode.Name + ": Not a valid hash"))
				return false
			}

			pubKey1Sig1, pubKey2Sig1 := new(big.Int), new(big.Int)
			r, s := new(big.Int), new(big.Int)

			pubKey1Sig1.SetBytes(publicKeySig[:32])
			pubKey2Sig1.SetBytes(publicKeySig[32:])

			sig1 := vm.context.GetSig1()
			r.SetBytes(sig1[:32])
			s.SetBytes(sig1[32:])

			pubKey := ecdsa.PublicKey{elliptic.P256(), pubKey1Sig1, pubKey2Sig1}

			result := ecdsa.Verify(&pubKey, hash, r, s)
			vm.evaluationStack.Push(BoolToByteArray(result))

		case ERRHALT:
			return false

		case HALT:
			return true
		}
	}
}

func (vm *VM) fetch(errorLocation string) (element byte, err error) {
	tempPc := vm.pc
	if len(vm.code) > tempPc {
		vm.pc++
		return vm.code[tempPc], nil
	} else {
		return 0, errors.New("Instruction set out of bounds")
	}
}

func (vm *VM) fetchMany(errorLocation string, argument int) (elements []byte, err error) {
	tempPc := vm.pc
	if len(vm.code)-tempPc > argument {
		vm.pc += argument
		return vm.code[tempPc : tempPc+argument], nil
	} else {
		return []byte{}, errors.New("Instruction set out of bounds")
	}
}

func (vm *VM) checkErrors(errorLocation string, errors ...error) bool {
	for i, err := range errors {
		if err != nil {
			vm.evaluationStack.Push([]byte(errorLocation + ": " + errors[i].Error()))
			return false
		}
	}
	return true
}

func (vm *VM) PopBytes(opCode OpCode) (elements []byte, err error) {
	bytes, err := vm.evaluationStack.Pop()
	if err != nil {
		return nil, err
	}

	elementSize := (len(bytes) + 64 - 1) / 64

	gasCost := opCode.gasFactor * uint64(elementSize)
	if int64(vm.fee-gasCost) < 0 {
		return nil, errors.New("Out of gas")
	}

	vm.fee -= gasCost

	return bytes, nil
}

func (vm *VM) PopSignedBigInt(opCode OpCode) (bigInt big.Int, err error) {
	bytes, err := vm.evaluationStack.Pop()
	if err != nil {
		return *big.NewInt(0), err
	}

	elementSize := (len(bytes) + 64 - 1) / 64

	gasCost := opCode.gasFactor * uint64(elementSize)
	if int64(vm.fee-gasCost) < 0 {
		return *big.NewInt(0), errors.New("Out of gas")
	}

	vm.fee -= gasCost

	result, err := SignedBigIntConversion(bytes, err)
	return result, err
}

func (vm *VM) PopUnsignedBigInt(opCode OpCode) (bigInt big.Int, err error) {
	bytes, err := vm.evaluationStack.Pop()
	if err != nil {
		return *big.NewInt(0), err
	}

	elementSize := (len(bytes) + 64 - 1) / 64

	gasCost := opCode.gasFactor * uint64(elementSize)
	if int64(vm.fee-gasCost) < 0 {
		return *big.NewInt(0), errors.New("Out of gas")
	}

	vm.fee -= gasCost

	result, err := UnsignedBigIntConversion(bytes, err)
	return result, err
}

func (vm *VM) GetErrorMsg() string {
	tos, err := vm.evaluationStack.PeekBytes()
	if err != nil {
		return "Peek on empty Stack"
	}
	return string(tos)
}
