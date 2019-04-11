package vm

import (
	"errors"
)

type Stack struct {
	Stack       [][]byte
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

func (s *Stack) Push(element []byte) error {
	if (*s).hasEnoughMemory(len(element)) {
		s.memoryUsage += uint32(len(element))
		s.Stack = append(s.Stack, element)
		return nil
	} else {
		return errors.New("Stack out of memory")
	}
}

func (s *Stack) PopIndexAt(index int) ([]byte, error) {
	if (*s).GetLength() >= index {
		element := (*s).Stack[index]
		s.memoryUsage -= uint32(len(element))
		s.Stack = append((*s).Stack[:index], (*s).Stack[index+1:]...)
		return element, nil
	} else {
		return []byte{}, errors.New("index out of bounds")
	}
}

func (s *Stack) Pop() (element []byte, err error) {
	if (*s).GetLength() > 0 {
		element = (*s).Stack[s.GetLength()-1]
		s.memoryUsage -= uint32(len(element))
		s.Stack = s.Stack[:s.GetLength()-1]
		return element, nil
	} else {
		return []byte{}, errors.New("pop() on empty stack")
	}
}

func (s *Stack) PeekBytes() (element []byte, err error) {
	if (*s).GetLength() > 0 {
		element = (*s).Stack[s.GetLength()-1]
		return element, nil
	} else {
		return []byte{}, errors.New("peek() on empty Stack")
	}
}

// Function checks, if enough memory is available to push the element
func (s *Stack) hasEnoughMemory(elementSize int) bool {
	return s.memoryMax >= uint32(elementSize)+s.memoryUsage
}
