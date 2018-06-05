package vm

import (
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
	val, err := s.PeekBytes()

	if err == nil {
		t.Errorf("Expected stack to be empty but contained the element: %v", val)
	}
}

func TestStack_PopIfRemoves(t *testing.T) {
	s := NewStack()

	s.PushBytes(UInt64ToByteArray(454))
	s.PushBytes(UInt64ToByteArray(46542))
	s.PushBytes(UInt64ToByteArray(841324768))

	tos, _ := s.PopBytes()

	expected := 841324768
	actual := ByteArrayToInt(tos)

	if expected != actual {
		t.Errorf("Expected result to be '%v' but was '%v'", expected, actual)
	}

	s.PopBytes()
	s.PopBytes()

	expected = 0
	actual = s.GetLength()
	if expected != actual {
		t.Errorf("Expected stack size to be '%v' but was '%v'", expected, actual)
	}
}

func TestStack_Peek(t *testing.T) {
	s := NewStack()

	s.PushBytes([]byte{0x01, 0x00})
	s.PeekBytes()

	expected := 1
	actual := s.GetLength()

	if expected != actual {
		t.Errorf("Expected Stack with size '%v' but got %v",expected, s.GetLength())
	}
}

func TestStack_PopIndexAt(t *testing.T) {
	s := NewStack()

	s.PushBytes([]byte{0x01, 0x00})
	s.PushBytes([]byte{0x02, 0x00})
	s.PushBytes([]byte{0x05})
	s.PushBytes([]byte{0x04, 0x00})
	element, _ := s.PopIndexAt(2)

	expected := 3
	actual := s.GetLength()
	if expected != actual {
		t.Errorf("Expected Stack with size '%v' but got %v",expected, s.GetLength())
	}

	expected = 5
	actual = ByteArrayToInt(element)
	if expected != actual {
		t.Errorf("Expected element to be '%v' but got '%v'", expected, actual)
	}
}

func TestStack_PushAndPopElement(t *testing.T) {
	s := NewStack()

	expected := 0
	actual := s.GetLength()
	if expected != actual {
		t.Errorf("Expected size before push to be '%v', but was '%v'", expected, actual)
	}

	s.PushBytes([]byte{0x02})

	expected = 1
	actual = s.GetLength()
	if expected != actual {
		t.Errorf("Expected size to be '%v' but was '%v'", expected, actual)
	}

	tos, _ := s.PopBytes()

	expected = 2
	actual = ByteArrayToInt(tos)
	if actual != expected {
		t.Errorf("Expected val of element to be '%v', but was '%v'", expected, actual)
	}

	s.PushBytes([]byte{0x05})

	expected = 1
	actual = s.GetLength()
	if actual != expected {
		t.Errorf("Expected size to be '%v' but was '%v'", expected, actual)
	}
}

func TestStack_MemoryUsage(t *testing.T) {
	s := NewStack()

	byteArray1 := []byte{123, 48, 56, 126}           // 4 + 1
	byteArray2 := []byte{175, 135, 44, 132, 48, 134} // 6 + 1
	byteArray3 := []byte{123, 132}
	byteArray4 := []byte{123, 48, 56, 126, 123, 48, 56, 126, 123, 48, 56, 126, 123, 48, 56, 126}

	if s.memoryUsage != uint32(0) {
		t.Errorf("Expected memory usage to be 0 before pushing anything but was %v", s.memoryUsage)
	}

	s.PushBytes(byteArray1)

	if s.memoryUsage != uint32(5) {
		t.Errorf("Expected memory usage to be 5 after pushing big.Int made from 4 bytes but was %v", s.memoryUsage)
	}

	s.PushBytes(byteArray2)

	if s.memoryUsage != uint32(12) {
		t.Errorf("Expected memory usage to be 12 after pushing big.Int made from 6 bytes but was %v", s.memoryUsage)
	}

	s.PushBytes(byteArray3)

	if s.memoryUsage != uint32(15) {
		t.Errorf("Expected memory usage to be 15 after pushing big.Int made from 6 bytes but was %v", s.memoryUsage)
	}

	s.PushBytes(byteArray4)

	if s.memoryUsage != uint32(32) {
		t.Errorf("Expected memory usage to be 32 after pushing big.Int made from 16 bytes but was %v", s.memoryUsage)
	}

	s.PopBytes()

	if s.memoryUsage != uint32(15) {
		t.Errorf("Expected memory usage to be 15 after pushing big.Int made from 6 bytes but was %v", s.memoryUsage)
	}

	s.PopBytes()

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
