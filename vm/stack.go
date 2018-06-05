package vm

import (
	"errors"
	"math/big"
)

type Stack struct {
	Stack       []big.Int
	memoryUsage uint32 // In bytes
	memoryMax   uint32
}

func NewStack() *Stack {
	return &Stack{
		Stack:       nil,
		memoryUsage: 0,
		memoryMax:   1000000, // Max 1000000 Bytes = 1MB
	}
}

func (s Stack) GetLength() int {
	return len(s.Stack)
}

func (s *Stack) Push(element big.Int) error {
	if (*s).hasEnoughMemory(element.BitLen()) {
		s.memoryUsage += getElementMemoryUsage(element.BitLen())
		s.Stack = append(s.Stack, element)
		return nil
	} else {
		return errors.New("Stack out of memory")
	}
}

func (s *Stack) PushBytes(element []byte) error {
	var bigInt big.Int
	bigInt.SetBytes(element)

	return s.Push(bigInt)
}

func (s *Stack) PopIndexAt(index int) ([]byte, error) {
	if (*s).GetLength() >= index {
		element := (*s).Stack[index]
		s.memoryUsage -= getElementMemoryUsage(element.BitLen())
		s.Stack = append((*s).Stack[:index], (*s).Stack[index+1:]...)
		return element.Bytes(), nil
	} else {
		return []byte{0}, errors.New("index out of bounds")
	}
}

func (s *Stack) Pop() (element big.Int, err error) {
	if (*s).GetLength() > 0 {
		element = (*s).Stack[s.GetLength()-1]
		s.memoryUsage -= getElementMemoryUsage(element.BitLen())
		s.Stack = s.Stack[:s.GetLength()-1]
		return element, nil
	} else {
		return *new(big.Int).SetInt64(0), errors.New("pop() on empty stack")
	}
}

func (s *Stack) PopBytes() (element []byte, err error) {
	bigInt, err := s.Pop()
	return bigInt.Bytes(), err
}

func (s *Stack) Peek() (element big.Int, err error) {
	if (*s).GetLength() > 0 {
		element = (*s).Stack[s.GetLength()-1]
		return element, nil
	} else {
		return *new(big.Int).SetInt64(0), errors.New("peek() on empty Stack")
	}
}

func (s *Stack) PeekBytes() (element []byte, err error) {
	bigInt, err := s.Peek()
	return bigInt.Bytes(), err
}

// Function turns bit into bytes and rounds up
func getElementMemoryUsage(element int) uint32 {
	return uint32(((element + 7) / 8) + 1)
}

// Function checks, if enough memory is available to push the element
func (s *Stack) hasEnoughMemory(elementSize int) bool {
	return s.memoryMax >= getElementMemoryUsage(elementSize)+s.memoryUsage
}
