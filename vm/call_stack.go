package vm

import (
	"errors"
	"math/big"
)

type Frame struct {
	variables     map[int]big.Int
	returnAddress int
}

type CallStack struct {
	values []*Frame
}

func NewCallStack() *CallStack {
	return &CallStack{}
}

func (cs CallStack) GetLength() int {
	return len(cs.values)
}

func (cs *CallStack) Push(element *Frame) {
	cs.values = append(cs.values[:cs.GetLength()], element)
}

func (cs *CallStack) Pop() (frame *Frame, err error) {
	if (*cs).GetLength() > 0 {
		element := (*cs).values[cs.GetLength()-1]
		cs.values = cs.values[:cs.GetLength()-1]
		return element, nil
	} else {
		return nil, errors.New("pop() on empty callStack")
	}
}

func (cs *CallStack) Peek() (frame *Frame, err error) {
	if (*cs).GetLength() > 0 {
		return (*cs).values[cs.GetLength()-1], nil
	} else {
		return nil, errors.New("peek() on empty callStack")
	}
}
