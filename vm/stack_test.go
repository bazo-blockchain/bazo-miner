package vm

import (
	"math/big"
	"testing"
)

func TestStack_NewStack(t *testing.T) {
	s := NewStack()
	if s.GetLength() != 0 {
		t.Errorf("Expected Stack with size 0 but got %v", s.GetLength())
	}
}

func TestStack_PopWhenEmpty(t *testing.T) {
	s := NewStack()
	val, err := s.Peek()

	if err == nil {
		t.Errorf("Throw error because val was %v", val)
	}
}

func TestStack_PopIfRemoves(t *testing.T) {
	s := NewStack()

	var val1 big.Int
	val1.SetInt64(454)

	var val2 big.Int
	val2.SetInt64(46542)

	var val3 big.Int
	val3.SetInt64(-841324768)

	s.Push(val1)
	s.Push(val2)
	s.Push(val3)

	tos, _ := s.Pop()

	if tos.Int64() != int64(-841324768) {
		t.Errorf("Expected 123 got something else")
	}

	s.Pop()
	s.Pop()

	if s.GetLength() != 0 {
		t.Errorf("Expected empty Stack to throw an error when using pop() but it didn't")
	}
}

func TestStack_Peek(t *testing.T) {
	s := NewStack()

	var val big.Int
	val.SetInt64(-841324768)

	s.Push(val)
	s.Peek()

	if s.GetLength() != 1 {
		t.Errorf("Expected Stack with size 1 but got %v", s.GetLength())
	}
}

func TestStack_PopIndexAt(t *testing.T) {
	s := NewStack()

	s.Push(*big.NewInt(int64(3)))
	s.Push(*big.NewInt(int64(4)))
	s.Push(*big.NewInt(int64(5)))
	s.Push(*big.NewInt(int64(6)))
	element, _ := s.PopIndexAt(2)

	if s.GetLength() != 3 {
		t.Errorf("Expected Stack with size 3 but got %v", s.GetLength())
	}

	if element.Int64() != 5 {
		t.Errorf("Expected element to be 5 but got %v", element)
	}
}

func TestStack_PushAndPopElement(t *testing.T) {
	s := NewStack()

	if s.GetLength() != 0 {
		t.Errorf("Expected size before push to be 0, but was %v", s.GetLength())
	}

	s.Push(*big.NewInt(int64(2)))

	if s.GetLength() != 1 {
		t.Errorf("Expected size to be 1 but was %v", s.GetLength())
	}

	val, _ := s.Pop()
	if val.Int64() != 2 {
		t.Errorf("Expected val of element to be 2, but was %v", val)
	}

	s.Push(*big.NewInt(int64(5)))

	if s.GetLength() != 1 {
		t.Errorf("Expected size to be 1 but was %v", s.GetLength())
	}
}

func TestStack_MemoryUsage(t *testing.T) {
	s := NewStack()

	byteArray1 := []byte{123, 48, 56, 126}           // 4 + 1
	byteArray2 := []byte{175, 135, 44, 132, 48, 134} // 6 + 1
	byteArray3 := []byte{123, 132}
	byteArray4 := []byte{123, 48, 56, 126, 123, 48, 56, 126, 123, 48, 56, 126, 123, 48, 56, 126}
	bigInt1 := new(big.Int).SetBytes(byteArray1)
	bigInt2 := new(big.Int).SetBytes(byteArray2)
	bigInt3 := new(big.Int).SetBytes(byteArray3)
	bigInt4 := new(big.Int).SetBytes(byteArray4)

	if s.memoryUsage != uint32(0) {
		t.Errorf("Expected memory usage to be 0 before pushing anything but was %v", s.memoryUsage)
	}

	s.Push(*bigInt1)

	if s.memoryUsage != uint32(5) {
		t.Errorf("Expected memory usage to be 5 after pushing big.Int made from 4 bytes but was %v", s.memoryUsage)
	}

	s.Push(*bigInt2)

	if s.memoryUsage != uint32(12) {
		t.Errorf("Expected memory usage to be 12 after pushing big.Int made from 6 bytes but was %v", s.memoryUsage)
	}

	s.Push(*bigInt3)

	if s.memoryUsage != uint32(15) {
		t.Errorf("Expected memory usage to be 15 after pushing big.Int made from 6 bytes but was %v", s.memoryUsage)
	}

	s.Push(*bigInt4)

	if s.memoryUsage != uint32(32) {
		t.Errorf("Expected memory usage to be 32 after pushing big.Int made from 16 bytes but was %v", s.memoryUsage)
	}

	s.Pop()

	if s.memoryUsage != uint32(15) {
		t.Errorf("Expected memory usage to be 15 after pushing big.Int made from 6 bytes but was %v", s.memoryUsage)
	}

	s.Pop()

	if s.memoryUsage != uint32(12) {
		t.Errorf("Expected memory usage to be 12 after pushing big.Int made from 6 bytes but was %v", s.memoryUsage)
	}
}

func TestStack_RoundingFunction(t *testing.T) {
	if getElementMemoryUsage(456) != 58 {
		t.Errorf("Expected memory usage to be 58 bytes, after adding 8 to 456, dividing by 8 and rounding up, but got %v", getElementMemoryUsage(456))
	}

	if getElementMemoryUsage(457) != 59 {
		t.Errorf("Expected memory usage to be 59 bytes, after adding 8 to 457, dividing by 8 and rounding up, but got %v", getElementMemoryUsage(457))
	}

	if getElementMemoryUsage(8) != 2 {
		t.Errorf("Expected memory usage to be 2 bytes, after adding 8 to 8, dividing by 8 and rounding up, but got %v", getElementMemoryUsage(8))
	}

	if getElementMemoryUsage(32) != 5 {
		t.Errorf("Expected memory usage to be 5 bytes, after adding 8 to 456, dividing by 8 and rounding up, but got %v", getElementMemoryUsage(32))
	}

}
