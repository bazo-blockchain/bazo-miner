package vm

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"errors"
	"fmt"
	"math/big"

	"encoding/binary"

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
	evaluationStack *Stack
	callStack       *CallStack
	context         Context
}

func NewVM(context Context) VM {
	return VM{
		code:            []byte{},
		pc:              0,
		evaluationStack: NewStack(),
		callStack:       NewCallStack(),
		context:         context,
	}
}

func NewTestVM(byteCode []byte) VM {
	return VM{
		code:            []byte{},
		pc:              0,
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
		stack.PushBytes([]byte("Trace: invalid opcode "))
		return
	}
	opCode := OpCodes[byteCode]

	var args []byte

	switch opCode.Name {
	case "push":
		nargs := int(vm.code[vm.pc+1])

		if vm.pc+nargs < (len(vm.code) - vm.pc) {
			args = vm.code[vm.pc+2 : vm.pc+nargs+3]
			fmt.Printf("%04d: %-6s %-10v %v\n", addr, opCode.Name, ByteArrayToInt(args), stack)
		}

		//TODO - Fix CALLEXT case, leads to index out of bounds exception
	/*case "callext":
	address := vm.code[vm.pc+1 : vm.pc+33]
	functionHash := vm.code[vm.pc+33 : vm.pc+37]
	nargs := int(vm.code[vm.pc+37])

	fmt.Printf("%04d: %-6s %x %x %v %v\n", addr, opCode.Name, address, functionHash, nargs, stack)
	*/

	case "newarr":
	case "arrappend":
	case "arrinsert":
	case "arrremove":
	case "arrat":
		args = vm.code[vm.pc+1 : vm.pc+opCode.Nargs+1]
		fmt.Printf("%04d: %-6s %v ", addr, opCode.Name, args)

		for _, e := range stack.Stack {
			fmt.Printf("[%# x]", e.Bytes())
			fmt.Printf(" ")
		}

		fmt.Printf("\n")

	default:
		args = vm.code[vm.pc+1 : vm.pc+opCode.Nargs+1]
		fmt.Printf("%04d: %-6s %v %v\n", addr, opCode.Name, args, stack)
	}
}

func (vm *VM) Exec(trace bool) bool {

	vm.code = vm.context.GetContract()

	if len(vm.code) > 100000 {
		vm.evaluationStack.PushBytes([]byte("vm.exec(): Instruction set to big"))
		return false
	}

	fee := vm.context.GetFee()

	// Infinite Loop until return called
	for {
		if trace {
			vm.trace()
		}

		// Fetch
		byteCode, err := vm.fetch("vm.exec()")
		if err != nil {
			vm.evaluationStack.PushBytes([]byte("vm.exec(): " + err.Error()))
			return false
		}
		// Return false if instruction is not an opCode
		if len(OpCodes) <= int(byteCode) {
			vm.evaluationStack.PushBytes([]byte("vm.exec(): Not a valid opCode"))
			return false
		}

		opCode := OpCodes[byteCode]
		// Subtract gas used for operation
		if fee < opCode.gasPrice {
			vm.evaluationStack.PushBytes([]byte("vm.exec(): out of gas"))
			return false
		} else {
			fee -= opCode.gasPrice
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

			err = vm.evaluationStack.PushBytes(bytes)

			if err != nil {
				vm.evaluationStack.PushBytes([]byte(opCode.Name + ": " + err.Error()))
				return false
			}

		case DUP:
			tos, err := vm.evaluationStack.PeekBytes()
			if err != nil {
				vm.evaluationStack.PushBytes([]byte(opCode.Name + ": " + err.Error()))
				return false
			}

			err = vm.evaluationStack.PushBytes(tos)
			if err != nil {
				vm.evaluationStack.PushBytes([]byte(opCode.Name + ": " + err.Error()))
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
					vm.evaluationStack.PushBytes([]byte(opCode.Name + ": index out of bounds"))
					return false
				}

				newTos, err := vm.evaluationStack.PopIndexAt(index)

				if err != nil {
					vm.evaluationStack.PushBytes([]byte(opCode.Name + ": " + err.Error()))
					return false
				}

				err = vm.evaluationStack.PushBytes(ConvertToByteArray(newTos))

				if err != nil {
					vm.evaluationStack.PushBytes([]byte(opCode.Name + ": " + err.Error()))
					return false
				}
			}

		case POP:
			_, rerr := vm.evaluationStack.PopBytes()
			if !vm.checkErrors(opCode.Name, rerr) {
				return false
			}

		case ADD:
			right, rerr := ConvertToBigInt(vm.evaluationStack.PopBytes())
			left, lerr := ConvertToBigInt(vm.evaluationStack.PopBytes())

			if !vm.checkErrors(opCode.Name, rerr, lerr) {
				return false
			}

			left.Add(&left, &right)
			err := vm.evaluationStack.PushBytes(ConvertToByteArray(left))

			if err != nil {
				vm.evaluationStack.PushBytes([]byte(opCode.Name + ": " + err.Error()))
				return false
			}

		case SUB:
			right, rerr := ConvertToBigInt(vm.evaluationStack.PopBytes())
			left, lerr := ConvertToBigInt(vm.evaluationStack.PopBytes())

			if !vm.checkErrors(opCode.Name, rerr, lerr) {
				return false
			}

			left.Sub(&left, &right)
			err := vm.evaluationStack.PushBytes(ConvertToByteArray(left))

			if err != nil {
				vm.evaluationStack.PushBytes([]byte(opCode.Name + ": " + err.Error()))
				return false
			}

		case MULT:
			right, rerr := ConvertToBigInt(vm.evaluationStack.PopBytes())
			left, lerr := ConvertToBigInt(vm.evaluationStack.PopBytes())

			if !vm.checkErrors(opCode.Name, rerr, lerr) {
				return false
			}

			left.Mul(&left, &right)
			err := vm.evaluationStack.PushBytes(ConvertToByteArray(left))

			if err != nil {
				vm.evaluationStack.PushBytes([]byte(opCode.Name + ": " + err.Error()))
				return false
			}

		case DIV:
			right, rerr := ConvertToBigInt(vm.evaluationStack.PopBytes())
			left, lerr := ConvertToBigInt(vm.evaluationStack.PopBytes())

			if !vm.checkErrors(opCode.Name, rerr, lerr) {
				return false
			}

			if right.Cmp(big.NewInt(0)) == 0 {
				vm.evaluationStack.PushBytes([]byte(opCode.Name + ": Division by Zero"))
				return false
			}

			left.Div(&left, &right)
			err := vm.evaluationStack.PushBytes(ConvertToByteArray(left))

			if err != nil {
				vm.evaluationStack.PushBytes([]byte(opCode.Name + ": " + err.Error()))
				return false
			}

		case MOD:
			right, rerr := ConvertToBigInt(vm.evaluationStack.PopBytes())
			left, lerr := ConvertToBigInt(vm.evaluationStack.PopBytes())

			if !vm.checkErrors(opCode.Name, rerr, lerr) {
				return false
			}

			if right.Cmp(big.NewInt(0)) == 0 {
				vm.evaluationStack.PushBytes([]byte(opCode.Name + ": Division by Zero"))
				return false
			}

			left.Mod(&left, &right)
			err := vm.evaluationStack.PushBytes(ConvertToByteArray(left))

			if err != nil {
				vm.evaluationStack.PushBytes([]byte(opCode.Name + ": " + err.Error()))
				return false
			}

		case NEG:
			tos, err := ConvertToBigInt(vm.evaluationStack.PopBytes())

			if err != nil {
				vm.evaluationStack.PushBytes([]byte(opCode.Name + ": " + err.Error()))
				return false
			}

			tos.Neg(&tos)

			vm.evaluationStack.PushBytes(ConvertToByteArray(tos))

		case EQ:
			right, rerr := ConvertToBigInt(vm.evaluationStack.PopBytes())
			left, lerr := ConvertToBigInt(vm.evaluationStack.PopBytes())
			if !vm.checkErrors(opCode.Name, rerr, lerr) {
				return false
			}

			result := left.Cmp(&right) == 0
			vm.evaluationStack.PushBytes(BoolToByteArray(result))

		case NEQ:
			right, rerr := ConvertToBigInt(vm.evaluationStack.PopBytes())
			left, lerr := ConvertToBigInt(vm.evaluationStack.PopBytes())
			if !vm.checkErrors(opCode.Name, rerr, lerr) {
				return false
			}

			result := left.Cmp(&right) != 0
			vm.evaluationStack.PushBytes(BoolToByteArray(result))

		case LT:
			right, rerr := ConvertToBigInt(vm.evaluationStack.PopBytes())
			left, lerr := ConvertToBigInt(vm.evaluationStack.PopBytes())
			if !vm.checkErrors(opCode.Name, rerr, lerr) {
				return false
			}

			result := left.Cmp(&right) == -1
			vm.evaluationStack.PushBytes(BoolToByteArray(result))

		case GT:
			right, rerr := ConvertToBigInt(vm.evaluationStack.PopBytes())
			left, lerr := ConvertToBigInt(vm.evaluationStack.PopBytes())
			if !vm.checkErrors(opCode.Name, rerr, lerr) {
				return false
			}

			result := left.Cmp(&right) == 1
			vm.evaluationStack.PushBytes(BoolToByteArray(result))

		case LTE:
			right, rerr := ConvertToBigInt(vm.evaluationStack.PopBytes())
			left, lerr := ConvertToBigInt(vm.evaluationStack.PopBytes())
			if !vm.checkErrors(opCode.Name, rerr, lerr) {
				return false
			}

			result := left.Cmp(&right) == -1 || left.Cmp(&right) == 0
			vm.evaluationStack.PushBytes(BoolToByteArray(result))

		case GTE:
			right, rerr := ConvertToBigInt(vm.evaluationStack.PopBytes())
			left, lerr := ConvertToBigInt(vm.evaluationStack.PopBytes())
			if !vm.checkErrors(opCode.Name, rerr, lerr) {
				return false
			}

			result := left.Cmp(&right) == 1 || left.Cmp(&right) == 0
			vm.evaluationStack.PushBytes(BoolToByteArray(result))

		case SHIFTL:
			nrOfShifts, errArg := vm.fetch(opCode.Name)
			tos, errStack := ConvertToBigInt(vm.evaluationStack.PopBytes())

			if !vm.checkErrors(opCode.Name, errArg, errStack) {
				return false
			}

			tos.Lsh(&tos, uint(nrOfShifts))
			err = vm.evaluationStack.PushBytes(ConvertToByteArray(tos))

			if err != nil {
				vm.evaluationStack.PushBytes([]byte(opCode.Name + ": " + err.Error()))
				return false
			}

		case SHIFTR:
			nrOfShifts, errArg := vm.fetch(opCode.Name)
			tos, errStack := ConvertToBigInt(vm.evaluationStack.PopBytes())

			if !vm.checkErrors(opCode.Name, errArg, errStack) {
				return false
			}

			tos.Rsh(&tos, uint(nrOfShifts))
			err = vm.evaluationStack.PushBytes(ConvertToByteArray(tos))

			if err != nil {
				vm.evaluationStack.PushBytes([]byte(opCode.Name + ": " + err.Error()))
				return false
			}

		case NOP:
			_, err := vm.fetch(opCode.Name)

			if err != nil {
				vm.evaluationStack.PushBytes([]byte(opCode.Name + ": " + err.Error()))
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
			right, errStack := ConvertToBigInt(vm.evaluationStack.PopBytes())

			if !vm.checkErrors(opCode.Name, errArg, errStack) {
				return false
			}

			var jumpTo big.Int
			jumpTo.SetBytes(nextInstruction)

			if right.Int64() == 1 {
				vm.pc = int(jumpTo.Int64())
			}

		case CALL:
			returnAddressBytes, errArg1 := vm.fetchMany(opCode.Name, 2) // Shows where to jump after executing
			argsToLoad, errArg2 := vm.fetch(opCode.Name)              // Shows how many elements have to be popped from evaluationStack

			if !vm.checkErrors(opCode.Name, errArg1, errArg2) {
				return false
			}

			var returnAddress big.Int
			returnAddress.SetBytes(returnAddressBytes)

			if int(returnAddress.Int64()) == 0 || int(returnAddress.Int64()) > len(vm.code) {
				vm.evaluationStack.PushBytes([]byte(opCode.Name + ": ReturnAddress out of bounds"))
				return false
			}

			frame := &Frame{returnAddress: vm.pc, variables: make(map[int]big.Int)}

			for i := int(argsToLoad) - 1; i >= 0; i-- {
				frame.variables[i], err = ConvertToBigInt(vm.evaluationStack.PopBytes())
				if err != nil {
					vm.evaluationStack.PushBytes([]byte(opCode.Name + ": " + err.Error()))
					return false
				}
			}

			vm.callStack.Push(frame)
			vm.pc = int(returnAddress.Int64())

		case CALLIF:
			returnAddressBytes, errArg1 := vm.fetchMany(opCode.Name, 2) // Shows where to jump after executing
			argsToLoad, errArg2 := vm.fetch(opCode.Name)              // Shows how many elements have to be popped from evaluationStack
			right, errStack := ConvertToBigInt(vm.evaluationStack.PopBytes())

			if !vm.checkErrors(opCode.Name, errArg1, errArg2, errStack) {
				return false
			}

			if right.Int64() == 1 {
				var returnAddress big.Int
				returnAddress.SetBytes(returnAddressBytes)

				if int(returnAddress.Int64()) == 0 || int(returnAddress.Int64()) > len(vm.code) {
					vm.evaluationStack.PushBytes([]byte(opCode.Name + ": ReturnAddress out of bounds"))
					return false
				}

				frame := &Frame{returnAddress: vm.pc, variables: make(map[int]big.Int)}

				for i := int(argsToLoad) - 1; i >= 0; i-- {
					frame.variables[i], err = ConvertToBigInt(vm.evaluationStack.PopBytes())
					if err != nil {
						vm.evaluationStack.PushBytes([]byte(opCode.Name + ": " + err.Error()))
						return false
					}
				}
				vm.callStack.Push(frame)
				vm.pc = int(returnAddress.Int64())
			}

		case CALLEXT:
			transactionAddress, errArg1 := vm.fetchMany(opCode.Name, 32) // Addresses are 32 bytes (var name: transactionAddress)
			functionHash, errArg2 := vm.fetchMany(opCode.Name, 4)        // Function hash identifies function in external smart contract, first 4 byte of SHA3 hash (var name: functionHash)
			argsToLoad, errArg3 := vm.fetch(opCode.Name)               // Shows how many arguments to pop from stack and pass to external function (var name: argsToLoad)

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
			right, err := ConvertToBigInt(vm.evaluationStack.PopBytes())

			if !vm.checkErrors(opCode.Name, err) {
				return false
			}

			err = vm.evaluationStack.PushBytes(ConvertToByteArray(*big.NewInt(int64(getElementMemoryUsage(right.BitLen())))))

			if err != nil {
				vm.evaluationStack.PushBytes([]byte(opCode.Name + ": " + err.Error()))
				return false
			}

		case SSTORE:
			index, errArgs := vm.fetch(opCode.Name)
			value, errStack := vm.evaluationStack.PopBytes()

			if !vm.checkErrors(opCode.Name, errArgs, errStack) {
				return false
			}

			err = vm.context.SetContractVariable(int(index), value)
			if err != nil {
				vm.evaluationStack.PushBytes([]byte(opCode.Name + ": " + err.Error()))
				return false
			}

		case STORE:
			address, errArgs := vm.fetch(opCode.Name)
			right, errStack := ConvertToBigInt(vm.evaluationStack.PopBytes())

			if !vm.checkErrors(opCode.Name, errArgs, errStack) {
				return false
			}

			callstackTos, err := vm.callStack.Peek()

			if err != nil {
				vm.evaluationStack.PushBytes([]byte(opCode.Name + ": " + err.Error()))
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
				vm.evaluationStack.PushBytes([]byte(opCode.Name + ": " + err.Error()))
				return false
			}

			err = vm.evaluationStack.PushBytes(value)
			if err != nil {
				vm.evaluationStack.PushBytes([]byte(opCode.Name + ": " + err.Error()))
				return false
			}

		case LOAD:
			address, errArg := vm.fetch(opCode.Name)
			callstackTos, errCallStack := vm.callStack.Peek()

			if !vm.checkErrors(opCode.Name, errArg, errCallStack) {
				return false
			}

			val := callstackTos.variables[int(address)]

			err := vm.evaluationStack.PushBytes(ConvertToByteArray(val))

			if err != nil {
				vm.evaluationStack.PushBytes([]byte(opCode.Name + ": " + err.Error()))
				return false
			}

		case ADDRESS:
			address := vm.context.GetAddress()
			err := vm.evaluationStack.PushBytes(address[:])

			if err != nil {
				vm.evaluationStack.PushBytes([]byte(opCode.Name + ": " + err.Error()))
				return false
			}

		case ISSUER:
			issuer := vm.context.GetIssuer()
			err := vm.evaluationStack.PushBytes(issuer[:])

			if err != nil {
				vm.evaluationStack.PushBytes([]byte(opCode.Name + ": " + err.Error()))
				return false
			}

		case BALANCE:
			balance := make([]byte, 8)
			binary.LittleEndian.PutUint64(balance, vm.context.GetBalance())

			err := vm.evaluationStack.PushBytes(balance)

			if err != nil {
				vm.evaluationStack.PushBytes([]byte(opCode.Name + ": " + err.Error()))
				return false
			}

		case CALLER:
			caller := vm.context.GetSender()
			err := vm.evaluationStack.PushBytes(caller[:])

			if err != nil {
				vm.evaluationStack.PushBytes([]byte(opCode.Name + ": " + err.Error()))
				return false
			}

		case CALLVAL:
			value := make([]byte, 8)
			binary.LittleEndian.PutUint64(value , vm.context.GetAmount())

			err := vm.evaluationStack.PushBytes(value[:])

			if err != nil {
				vm.evaluationStack.PushBytes([]byte(opCode.Name + ": " + err.Error()))
				return false
			}

		case CALLDATA:
			td := vm.context.GetTransactionData()
			for i := 0; i < len(td); i++ {
				length := int(td[i]) // Length of parameters

				if len(td)-i-1 <= length {
					vm.evaluationStack.PushBytes([]byte(opCode.Name + ": Index out of bounds"))
					return false
				}

				err := vm.evaluationStack.PushBytes(td[i+1 : i+length+2])

				if err != nil {
					vm.evaluationStack.PushBytes([]byte(opCode.Name + ": " + err.Error()))
					return false
				}

				i += int(td[i]) + 1 // Increase to next parameter length
			}

		case NEWMAP:
			m := NewMap()

			err = vm.evaluationStack.PushBytes(m)
			if err != nil {
				vm.evaluationStack.PushBytes([]byte(opCode.Name + ": " + err.Error()))
				return false
			}

		case MAPPUSH:
			mba, err := vm.evaluationStack.PopBytes()
			if err != nil {
				vm.evaluationStack.PushBytes([]byte(opCode.Name + ": " + err.Error()))
				return false
			}

			k, err := vm.evaluationStack.PopBytes()
			if err != nil {
				vm.evaluationStack.PushBytes([]byte(opCode.Name + ": " + err.Error()))
				return false
			}

			v, err := vm.evaluationStack.PopBytes()
			if err != nil {
				vm.evaluationStack.PushBytes([]byte(opCode.Name + ": " + err.Error()))
				return false
			}

			m, err := MapFromByteArray(mba)
			if err != nil {
				vm.evaluationStack.PushBytes([]byte(opCode.Name + ": " + err.Error()))
				return false
			}

			err = m.Append(k, v)
			if err != nil {
				vm.evaluationStack.PushBytes([]byte(opCode.Name + ": " + err.Error()))
				return false
			}

			err = vm.evaluationStack.PushBytes(m)
			if err != nil {
				vm.evaluationStack.PushBytes([]byte(opCode.Name + ": " + err.Error()))
				return false
			}

		case MAPGETVAL:
			mapAsByteArray, err := vm.evaluationStack.PopBytes()
			if err != nil {
				vm.evaluationStack.PushBytes([]byte(opCode.Name + ": " + err.Error()))
				return false
			}

			k, err := vm.evaluationStack.PopBytes()
			if err != nil {
				vm.evaluationStack.PushBytes([]byte(opCode.Name + ": " + err.Error()))
				return false
			}


			m, err := MapFromByteArray(mapAsByteArray)
			if err != nil {
				vm.evaluationStack.PushBytes([]byte(opCode.Name + ": " + err.Error()))
				return false
			}

			v, err := m.GetVal(k)
			if err != nil {
				vm.evaluationStack.PushBytes([]byte(opCode.Name + ": " + err.Error()))
				return false
			}

			err = vm.evaluationStack.PushBytes(v)
			if err != nil {
				vm.evaluationStack.PushBytes([]byte(opCode.Name + ": " + err.Error()))
				return false
			}

		case MAPSETVAL:
			mapAsByteArray, err := vm.evaluationStack.PopBytes()
			if err != nil {
				vm.evaluationStack.PushBytes([]byte(opCode.Name + ": " + err.Error()))
				return false
			}

			m, err := MapFromByteArray(mapAsByteArray)
			if err != nil {
				vm.evaluationStack.PushBytes([]byte(opCode.Name + ": " + err.Error()))
				return false
			}

			k, err := vm.evaluationStack.PopBytes()
			if err != nil {
				vm.evaluationStack.PushBytes([]byte(opCode.Name + ": " + err.Error()))
				return false
			}

			v, err := vm.evaluationStack.PopBytes()
			if err != nil {
				vm.evaluationStack.PushBytes([]byte(opCode.Name + ": " + err.Error()))
				return false
			}

			err = m.SetVal(k, v)
			if err != nil {
				vm.evaluationStack.PushBytes([]byte(opCode.Name + ": " + err.Error()))
				return false
			}

			err = vm.evaluationStack.PushBytes(m)
			if err != nil {
				vm.evaluationStack.PushBytes([]byte(opCode.Name + ": " + err.Error()))
				return false
			}

		case MAPREMOVE:
			mapAsByteArray, err := vm.evaluationStack.PopBytes()
			if err != nil {
				vm.evaluationStack.PushBytes([]byte(opCode.Name + ": " + err.Error()))
				return false
			}

			k, err := vm.evaluationStack.PopBytes()
			if err != nil {
				vm.evaluationStack.PushBytes([]byte(opCode.Name + ": " + err.Error()))
				return false
			}

			m, err := MapFromByteArray(mapAsByteArray)
			if err != nil {
				vm.evaluationStack.PushBytes([]byte(opCode.Name + ": " + err.Error()))
				return false
			}

			err = m.Remove(k)
			if err != nil {
				vm.evaluationStack.PushBytes([]byte(opCode.Name + ": " + err.Error()))
				return false
			}

			err = vm.evaluationStack.PushBytes(m)
			if err != nil {
				vm.evaluationStack.PushBytes([]byte(opCode.Name + ": " + err.Error()))
				return false
			}

		case NEWARR:
			a := NewArray()
			vm.evaluationStack.PushBytes(a)

		case ARRAPPEND:
			v, verr := vm.evaluationStack.PopBytes()
			a, aerr := vm.evaluationStack.PopBytes()
			if !vm.checkErrors(opCode.Name, verr, aerr) {
				return false
			}

			arr, err := ArrayFromByteArray(a)
			if err != nil {
				vm.evaluationStack.PushBytes([]byte(opCode.Name + ": " + err.Error()))
				return false
			}

			err = arr.Append(v)
			if err != nil {
				vm.evaluationStack.PushBytes([]byte(opCode.Name + ": Invalid argument size of ARRAPPEND"))
				return false
			}

			err = vm.evaluationStack.PushBytes(arr)
			if err != nil {
				vm.evaluationStack.PushBytes([]byte(opCode.Name + ": " + err.Error()))
				return false
			}

		case ARRINSERT:
			i, err := vm.evaluationStack.PopBytes()
			if err != nil {
				vm.evaluationStack.PushBytes([]byte(opCode.Name + ": " + err.Error()))
				return false
			}

			if len(i) > 2 {
				vm.evaluationStack.PushBytes([]byte(opCode.Name + ": Wrong index size"))
				return false
			}

			element, err := vm.evaluationStack.PopBytes()
			if err != nil {
				vm.evaluationStack.PushBytes([]byte(opCode.Name + ": " + err.Error()))
				return false
			}

			a, err := vm.evaluationStack.PopBytes()
			if err != nil {
				vm.evaluationStack.PushBytes([]byte(opCode.Name + ": " + err.Error()))
				return false
			}

			arr, err := ArrayFromByteArray(a)
			if err != nil {
				vm.evaluationStack.PushBytes([]byte(opCode.Name + ": " + err.Error()))
				return false
			}

			index, err := ByteArrayToUI16(i)
			if err != nil {
				vm.evaluationStack.PushBytes([]byte(opCode.Name + ": " + err.Error()))
				return false
			}

			size, err := arr.getSize()
			if err != nil {
				vm.evaluationStack.PushBytes([]byte(opCode.Name + ": " + err.Error()))
				return false
			}

			if index >= size {
				vm.evaluationStack.PushBytes([]byte(opCode.Name + ": Index out of bounds"))
				return false
			}

			err = arr.Insert(index, element)
			if err != nil {
				vm.evaluationStack.PushBytes([]byte(opCode.Name + ": " + err.Error()))
				return false
			}

			err = vm.evaluationStack.PushBytes(arr)
			if err != nil {
				vm.evaluationStack.PushBytes([]byte(opCode.Name + ": " + err.Error()))
				return false
			}

		case ARRREMOVE:
			a, aerr := vm.evaluationStack.PopBytes()
			ba, errArgs := vm.fetchMany(opCode.Name, 2)
			index, err := ByteArrayToUI16(ba)
			if !vm.checkErrors(opCode.Name, aerr, errArgs, err) {
				return false
			}

			arr, err := ArrayFromByteArray(a)
			if err != nil {
				vm.evaluationStack.PushBytes([]byte(opCode.Name + ": " + err.Error()))
				return false
			}

			err = arr.Remove(index)
			if err != nil {
				vm.evaluationStack.PushBytes([]byte(opCode.Name + ": " + err.Error()))
				return false
			}

			err = vm.evaluationStack.PushBytes(arr)
			if err != nil {
				vm.evaluationStack.PushBytes([]byte(opCode.Name + ": " + err.Error()))
				return false
			}

		case ARRAT:
			a, err := vm.evaluationStack.PeekBytes()
			if err != nil {
				vm.evaluationStack.PushBytes([]byte(opCode.Name + ": " + err.Error()))
				return false
			}

			ba, err := vm.fetchMany(opCode.Name, 2)
			if err != nil {
				vm.evaluationStack.PushBytes([]byte(opCode.Name + ": " + err.Error()))
				return false
			}

			index, err := ByteArrayToUI16(ba)
			if err != nil {
				vm.evaluationStack.PushBytes([]byte(opCode.Name + ": " + err.Error()))
				return false
			}

			arr, err := ArrayFromByteArray(a)
			if err != nil {
				vm.evaluationStack.PushBytes([]byte(opCode.Name + ": " + err.Error()))
				return false
			}

			element, err := arr.At(index)
			if err != nil {
				vm.evaluationStack.PushBytes([]byte(opCode.Name + ": " + err.Error()))
				return false
			}

			err = vm.evaluationStack.PushBytes(element)
			if err != nil {
				vm.evaluationStack.PushBytes([]byte(opCode.Name + ": " + err.Error()))
				return false
			}

		case SHA3:
			right, err := ConvertToBigInt(vm.evaluationStack.PopBytes())

			if err != nil {
				vm.evaluationStack.PushBytes([]byte(opCode.Name + ": " + err.Error()))
				return false
			}

			hasher := sha3.New256()
			hasher.Write(right.Bytes())
			hash := hasher.Sum(nil)

			err = vm.evaluationStack.PushBytes(hash)

			if err != nil {
				vm.evaluationStack.PushBytes([]byte(opCode.Name + ": " + err.Error()))
				return false
			}

		case CHECKSIG:
			publicKeySig, errArg1 := ConvertToBigInt(vm.evaluationStack.PopBytes()) // PubKeySig
			hash, errArg2 := ConvertToBigInt(vm.evaluationStack.PopBytes())         // Hash

			if !vm.checkErrors(opCode.Name, errArg1, errArg2) {
				return false
			}

			if len(publicKeySig.Bytes()) != 64 {
				vm.evaluationStack.PushBytes([]byte(opCode.Name + ": Not a valid address"))
				return false
			}

			if len(hash.Bytes()) != 32 {
				vm.evaluationStack.PushBytes([]byte(opCode.Name + ": Not a valid hash"))
				return false
			}

			pubKey1Sig1, pubKey2Sig1 := new(big.Int), new(big.Int)
			r, s := new(big.Int), new(big.Int)

			pubKey1Sig1.SetBytes(publicKeySig.Bytes()[:32])
			pubKey2Sig1.SetBytes(publicKeySig.Bytes()[32:])

			sig1 := vm.context.GetSig1()
			r.SetBytes(sig1[:32])
			s.SetBytes(sig1[32:])

			pubKey := ecdsa.PublicKey{elliptic.P256(), pubKey1Sig1, pubKey2Sig1}

			result := ecdsa.Verify(&pubKey, hash.Bytes(), r, s)
			vm.evaluationStack.PushBytes(BoolToByteArray(result))

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
		return 0, errors.New("instructionSet out of bounds")
	}
}

func (vm *VM) fetchMany(errorLocation string, argument int) (elements []byte, err error) {
	tempPc := vm.pc
	if len(vm.code)-tempPc > argument {
		vm.pc += argument
		return vm.code[tempPc : tempPc+argument], nil
	} else {
		return []byte{}, errors.New("instructionSet out of bounds")
	}
}

func (vm *VM) checkErrors(errorLocation string, errors ...error) bool {
	for i, err := range errors {
		if err != nil {
			vm.evaluationStack.PushBytes([]byte(errorLocation + ": " + errors[i].Error()))
			return false
		}
	}
	return true
}

func (vm *VM) GetErrorMsg() string {
	tos, err := vm.evaluationStack.PeekBytes()
	if err != nil {
		return "Peek on empty Stack"
	}
	return string(tos)
}
