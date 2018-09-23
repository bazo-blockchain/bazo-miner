package vm

import (
	"bytes"
	"testing"
)

func Test_NewMap(t *testing.T) {
	m := NewMap()

	if len(m) != 3 {
		t.Errorf("Expected a Byte Array with size 3 but got %v", len(m))
	}
}

func TestMap_IncerementSize(t *testing.T) {
	m := NewMap()

	s, err := ByteArrayToUI16(m[1:3])
	if s != 0 || err != nil {
		t.Errorf("Invalid Array Size, Expected 0 but got %v", s)
	}

	m.IncrementSize()
	si, err := ByteArrayToUI16(m[1:3])
	if si != 1 || err != nil {
		t.Errorf("Invalid Map Size, Expected 1 after increment but got %v", si)
	}
}

func TestMap_DecrementSize(t *testing.T) {
	a := Array([]byte{0x02, 0x02, 0x00})

	s, err := ByteArrayToUI16(a[1:3])
	if s != 2 || err != nil {
		t.Errorf("Invalid Array Size, Expected 2 but got %v", s)
	}

	a.DecrementSize()
	sd, err := ByteArrayToUI16(a[1:3])
	if sd != 1 || err != nil {
		t.Errorf("Invalid Array Size, Expected 1 after decrement but got %v", sd)
	}
}

func TestMap_Append(t *testing.T) {
	m := NewMap()
	k := []byte{0x01}
	v := []byte{0x64, 0x00}
	err := m.Append(k, v)

	if err != nil {
		t.Errorf("%v", err)
	}

	ba := []byte(m)
	if len(ba) != 10 {
		t.Errorf("Expected a Byte Array with size 10 but got %v", len(ba))
	}

	if ba[5] != 0x01 {
		t.Errorf("Unexpected key or unexpected internal data structure")
	}

	if ba[8] != 0x64 {
		t.Errorf("Unexpected key or unexpected internal data structure")
	}
}

func TestMap_ContainsKey_EmptyMap(t *testing.T) {
	m := NewMap()
	actual, _ := m.MapContainsKey([]byte{0x00})
	if actual {
		t.Errorf("Didn't expect map to contain key")
	}
}

func TestMap_ContainsKey_false(t *testing.T) {
	m := NewMap()
	m.Append([]byte{0x02}, []byte{0x01, 0x02})
	m.Append([]byte{0x05}, []byte{0x01, 0x02})
	m.Append([]byte{0x03}, []byte{0x01, 0x02})

	actual, _ := m.MapContainsKey([]byte{0x00})
	if actual {
		t.Errorf("Didn't expect map to contain key")
	}
}

func TestMap_ContainsKey_true(t *testing.T) {
	m := NewMap()
	m.Append([]byte{0x02}, []byte{0x01, 0x02})
	m.Append([]byte{0x05}, []byte{0x01, 0x02})
	m.Append([]byte{0x00}, []byte{0x01, 0x02})
	m.Append([]byte{0x03}, []byte{0x01, 0x02})
	actual, _ := m.MapContainsKey([]byte{0x00})
	if !actual {
		t.Errorf("Expected map to contain key")
	}
}

func TestMap_GetVal(t *testing.T) {
	m := NewMap()
	m.Append([]byte{0x00}, []byte{0x00})
	m.Append([]byte{0x01}, []byte{0x01, 0x01})
	m.Append([]byte{0x02, 0x00}, []byte{0x02, 0x02, 0x02})
	m.Append([]byte{0x03, 0x00, 0x00}, []byte{0x03, 0x03, 0x03, 0x03, 0x03})

	expected4 := []byte{0x03, 0x03, 0x03, 0x03, 0x03}
	actual4, err4 := m.GetVal([]byte{0x03, 0x00, 0x00})
	if err4 != nil {
		t.Errorf("%v", err4)
	}
	if !bytes.Equal(expected4, actual4) {
		t.Errorf("Unexpected value, Expected '%# x' but was '%# x'", expected4, actual4)
	}

	expected3 := []byte{0x02, 0x02, 0x02}
	actual3, err3 := m.GetVal([]byte{0x02, 0x00})
	if err3 != nil {
		t.Errorf("%v", err3)
	}
	if !bytes.Equal(expected3, actual3) {
		t.Errorf("Unexpected value, Expected '%# x' but was '%# x'", expected3, actual3)
	}
}

func TestMap_SetVal(t *testing.T) {
	actual := NewMap()
	actual.Append([]byte{0x00}, []byte{0x00})
	actual.Append([]byte{0x02, 0x00}, []byte{0x02, 0x02, 0x02})

	size, err := actual.getSize()
	if size != 2 || err != nil {
		t.Errorf("Expected map size to be '4' but was '%v'", size)
	}

	expected := NewMap()
	expected.Append([]byte{0x00}, []byte{0x00})
	expected.Append([]byte{0x02, 0x00}, []byte{0x04, 0x04, 0x04, 0x04, 0x04, 0x04, 0x04, 0x04, 0x04})

	err = actual.SetVal([]byte{0x02, 0x00}, []byte{0x04, 0x04, 0x04, 0x04, 0x04, 0x04, 0x04, 0x04, 0x04})

	if err != nil {
		t.Errorf("%v", err)
	}

	size, err = actual.getSize()
	if size != 2 || err != nil {
		t.Errorf("Expected map size to be '4' but was '%v'", size)
	}

	if !bytes.Equal(expected, actual) {
		t.Errorf("Expected map to be '[%# x]' but was '[%# x]' after setVal", expected, actual)
	}

}

func TestMap_Remove(t *testing.T) {
	actual := NewMap()
	actual.Append([]byte{0x00}, []byte{0x00})
	actual.Append([]byte{0x01}, []byte{0x01, 0x01})
	actual.Append([]byte{0x02, 0x00}, []byte{0x02, 0x02, 0x02})
	actual.Append([]byte{0x03, 0x00, 0x00}, []byte{0x03, 0x03, 0x03, 0x03, 0x03})

	size, err := actual.getSize()
	if size != 4 || err != nil {
		t.Errorf("Expected map size to be '4' but was '%v'", size)
	}

	expected := NewMap()
	expected.Append([]byte{0x00}, []byte{0x00})
	expected.Append([]byte{0x01}, []byte{0x01, 0x01})
	expected.Append([]byte{0x03, 0x00, 0x00}, []byte{0x03, 0x03, 0x03, 0x03, 0x03})

	err = actual.Remove([]byte{0x02, 0x00})
	if err != nil {
		t.Errorf("%v", err)
	}

	size, err = actual.getSize()
	if size != 3 || err != nil {
		t.Errorf("Expected map size to be '3' but was '%v'", size)
	}

	if !bytes.Equal(expected, actual) {
		t.Errorf("Expected map to be '[%# x]' but was '[%# x]' after element removal", expected, actual)
	}
}
