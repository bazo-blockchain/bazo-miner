package vm

import (
	"bytes"
	"errors"
	"log"
)

type Map []byte

func NewMap() Map {
	return []byte{0x01, 0x00, 0x00}
}

func MapFromByteArray(m []byte) (Map, error) {
	if len(m) <= 0 {
		return Map{}, errors.New("empty map")
	}
	if m[0] != 0x01 {
		return Map{}, errors.New("invalid datatype supplied")
	}
	return Map(m), nil
}

func (m *Map) getSize() (uint16, error) {
	value, err := ByteArrayToUI16((*m)[1:3])
	if err != nil {
		return 0, errors.New("cannot get size of map")
	}
	return value, nil
}

func (m *Map) setSize(ba []byte) {
	(*m)[1] = ba[0]
	(*m)[2] = ba[1]
}

func (m *Map) IncrementSize() {
	s, err := m.getSize()
	if err != nil {
		log.Fatal("could not increment size")
	}
	s++
	m.setSize(UInt16ToByteArray(s))
}

func (m *Map) DecrementSize() error {
	s, err := m.getSize()
	if err != nil {
		log.Fatal("could not decrement size")
	}

	if s <= 0 {
		return errors.New("Map size already 0")
	}
	s--
	m.setSize(UInt16ToByteArray(s))
	return nil
}

func (m *Map) MapContainsKey(key []byte) (bool, error) {
	offset := 3
	l := len(*m)

	for index := offset; index < l; {
		if l == 3 {
			return false, nil
		}

		k, keyEndsBefore, err := getElement(m, index)

		sizeOfValue, err := getElementSize(m, keyEndsBefore)
		if err != nil {
			return false, err
		}

		valueStartsAt := keyEndsBefore //Just for better readability
		if bytes.Equal(key, k) {
			return true, err
		}
		valueEndsBefore := nextElementStartsAt(valueStartsAt, sizeOfValue)

		if index == valueEndsBefore {
			return false, errors.New("element sizes are 0")
		}
		index = valueEndsBefore
	}
	return false, nil
}

func (m *Map) Append(key []byte, value []byte) error {
	sk := len(key)
	sv := len(value)
	if sk > int(UINT16_MAX) || sv > int(UINT16_MAX) {
		return errors.New("key or value size overflows uint16")
	}

	tmp := append(*m, UInt16ToByteArray(uint16(sk))...)
	tmp = append(tmp, key...)
	tmp = append(tmp, UInt16ToByteArray(uint16(sv))...)
	tmp = append(tmp, value...)
	*m = tmp
	m.IncrementSize()
	return nil
}

func (m *Map) SetVal(key []byte, value []byte) error {
	err := m.Remove(key)
	m.Append(key, value)
	if err != nil {
		return err
	}
	return nil
}

func (m *Map) GetVal(key []byte) ([]byte, error) {
	offset := 3
	l := len(*m)

	for index := offset; index < l; {
		if l == 3 {
			return []byte{}, errors.New("no elements in map")
		}

		k, valueStartsAt, err := getElement(m, index)
		if err != nil {
			return []byte{}, err
		}

		v, nextElementStartsAt, err := getElement(m, valueStartsAt)
		if err != nil {
			return []byte{}, err
		}

		if bytes.Equal(key, k) {
			return v, nil
		}

		if index == nextElementStartsAt {
			return []byte{}, errors.New("element sizes are 0")
		}
		index = nextElementStartsAt
	}

	return []byte{}, errors.New("key not found")
}

func (m *Map) Remove(key []byte) error {
	offset := 3
	l := len(*m)

	for index := offset; index < l; {
		if l == 3 {
			return errors.New("no elements in map")
		}

		k, keyEndsBefore, err := getElement(m, index)
		if err != nil {
			return err
		}

		sizeOfValue, err := getElementSize(m, keyEndsBefore)
		if err != nil {
			return err
		}

		valueStartsAt := keyEndsBefore //Just for better readability
		valueEndsBefore := nextElementStartsAt(valueStartsAt, sizeOfValue)
		if bytes.Equal(key, k) {
			tmp := append([]byte{}, (*m)[:index]...)
			*m = append(tmp, (*m)[valueEndsBefore:]...)
			m.DecrementSize()
			return nil
		}

		if index == valueEndsBefore {
			return errors.New("element sizes are 0")
		}
		index = valueEndsBefore
	}
	return errors.New("key not found")
}

func getElement(m *Map, startsAt int) (element []byte, endsBefore int, err error) {
	size, err := getElementSize(m, startsAt)
	if err != nil {
		return []byte{}, 0, err
	}
	endsBefore = nextElementStartsAt(startsAt, size)
	element, err = getBytesOfElement(m, startsAt, endsBefore)
	if err != nil {
		return []byte{}, 0, err
	}
	return element, endsBefore, err
}

func getBytesOfElement(m *Map, startsAt int, endsBefore int) ([]byte, error) {
	if startsAt >= endsBefore {
		return []byte{}, errors.New("can't retrieve element")
	}
	length := len(*m)

	if length < startsAt+2 || length < endsBefore {
		return []byte{}, errors.New("map internals error")
	}

	return (*m)[startsAt+2 : endsBefore], nil
}
func nextElementStartsAt(index int, elementSize uint16) int {
	return index + 2 + int(elementSize)
}

func getElementSize(m *Map, index int) (uint16, error) {
	elementSize, err := ByteArrayToUI16((*m)[index : index+2])
	return elementSize, err
}
