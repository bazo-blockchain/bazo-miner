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

	s.Push(UInt64ToByteArray(454))
	s.Push(UInt64ToByteArray(46542))
	s.Push(UInt64ToByteArray(841324768))

	tos, _ := s.Pop()

	expected := 841324768
	actual := ByteArrayToInt(tos)

	if expected != actual {
		t.Errorf("Expected result to be '%v' but was '%v'", expected, actual)
	}

	s.Pop()
	s.Pop()

	expected = 0
	actual = s.GetLength()
	if expected != actual {
		t.Errorf("Expected stack size to be '%v' but was '%v'", expected, actual)
	}
}

func TestStack_Peek(t *testing.T) {
	s := NewStack()

	s.Push([]byte{0x01, 0x00})
	s.PeekBytes()

	expected := 1
	actual := s.GetLength()

	if expected != actual {
		t.Errorf("Expected Stack with size '%v' but got %v", expected, s.GetLength())
	}
}

func TestStack_PopIndexAt(t *testing.T) {
	s := NewStack()

	s.Push([]byte{0x01, 0x00})
	s.Push([]byte{0x02, 0x00})
	s.Push([]byte{0x05})
	s.Push([]byte{0x04, 0x00})
	element, _ := s.PopIndexAt(2)

	expected := 3
	actual := s.GetLength()
	if expected != actual {
		t.Errorf("Expected Stack with size '%v' but got %v", expected, s.GetLength())
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

	s.Push([]byte{0x02})

	expected = 1
	actual = s.GetLength()
	if expected != actual {
		t.Errorf("Expected size to be '%v' but was '%v'", expected, actual)
	}

	tos, _ := s.Pop()

	expected = 2
	actual = ByteArrayToInt(tos)
	if actual != expected {
		t.Errorf("Expected val of element to be '%v', but was '%v'", expected, actual)
	}

	s.Push([]byte{0x05})

	expected = 1
	actual = s.GetLength()
	if actual != expected {
		t.Errorf("Expected size to be '%v' but was '%v'", expected, actual)
	}
}

func TestStack_MemoryUsage(t *testing.T) {
	s := NewStack()

	byteArray1 := []byte{123, 48, 56, 126}
	byteArray2 := []byte{175, 135, 44, 132, 48, 134}
	byteArray3 := []byte{123, 132}
	byteArray4 := []byte{123, 48, 56, 126, 123, 48, 56, 126, 123, 48, 56, 126, 123, 48, 56, 126}

	if s.memoryUsage != uint32(0) {
		t.Errorf("Expected memory usage to be 0 before pushing anything but was %v", s.memoryUsage)
	}

	s.Push(byteArray1)

	expected := uint32(4)
	actual := s.memoryUsage
	if expected != actual {
		t.Errorf("Expected memory usage to be '%v' but was '%v'", expected, actual)
	}

	s.Push(byteArray2)

	expected = uint32(10)
	actual = s.memoryUsage
	if expected != actual {
		t.Errorf("Expected memory usage to be '%v' but was '%v'", expected, actual)
	}

	s.Push(byteArray3)

	expected = uint32(12)
	actual = s.memoryUsage
	if expected != actual {
		t.Errorf("Expected memory usage to be '%v' but was '%v'", expected, actual)
	}

	s.Push(byteArray4)

	expected = uint32(28)
	actual = s.memoryUsage
	if expected != actual {
		t.Errorf("Expected memory usage to be '%v' but was '%v'", expected, actual)
	}

	s.Pop()

	expected = uint32(12)
	actual = s.memoryUsage
	if expected != actual {
		t.Errorf("Expected memory usage to be '%v' but was '%v'", expected, actual)
	}

	s.Pop()

	expected = uint32(10)
	actual = s.memoryUsage
	if expected != actual {
		t.Errorf("Expected memory usage to be '%v' but was '%v'", expected, actual)
	}
}
