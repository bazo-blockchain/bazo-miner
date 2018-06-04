package vm

import (
	"bytes"
	"errors"
	"log"
)

type action func(array *Array, k uint16, s uint16) ([]byte, error)
type Array []byte

func NewArray() Array {
	ba := []byte{0x02}
	size := []byte{0x00, 0x00}
	return append(ba, size...)
}

func ArrayFromByteArray(arr []byte) (Array, error) {
	if len(arr) == 0 {
		return Array{}, errors.New("not a valid array")
	}

	if arr[0] != 0x02 {
		return Array{}, errors.New("invalid data type supplied")
	}
	return Array(arr), nil
}

func (a *Array) getSize() (uint16, error) {
	value, err := ByteArrayToUI16((*a)[1:3])
	if err != nil {
		return 0, errors.New("cannot get size of array")
	}
	return value, nil
}

func (a *Array) setSize(ba []byte) {
	(*a)[1] = ba[0]
	(*a)[2] = ba[1]
}

func (a *Array) IncrementSize() {
	s, err := a.getSize()
	if err != nil {
		log.Fatal("could not increase size")
	}
	s++
	a.setSize(UInt16ToByteArray(s))
}

func (a *Array) DecrementSize() error {
	s, err := a.getSize()
	if err != nil {
		log.Fatal("could not decrement size")
	}

	if s <= 0 {
		return errors.New("Array size already 0")
	}
	s--
	a.setSize(UInt16ToByteArray(s))
	return nil
}

func (a *Array) At(index uint16) ([]byte, error) {
	var f action = func(array *Array, k uint16, s uint16) ([]byte, error) {
		return (*array)[k+2 : k+2+s], nil
	}
	result, err := a.goToIndex(index, f)
	return result, err
}

func (a *Array) Insert(index uint16, element []byte) error {
	a.Remove(index)
	var f action = func(array *Array, k uint16, s uint16) ([]byte, error) {
		tmp := Array{}
		tmp = append(tmp, (*a)[:k]...)
		tmp.Append(element)
		*a = append(tmp, (*a)[k:]...)
		return []byte{}, nil
	}
	_, err := a.goToIndex(index, f)
	return err
}

func (a *Array) Append(ba []byte) error {
	if bytes.Equal(ba, []byte{}) {
		ba = []byte{0x00, 0x00}
	}
	s := len(ba)

	if s > int(UINT16_MAX) {
		return errors.New("Element Size overflow")
	}

	sb := UInt16ToByteArray(uint16(len(ba)))
	*a = append(*a, append(sb, ba...)...)
	a.IncrementSize()
	return nil
}

func (a *Array) Remove(index uint16) error {
	var f action = func(array *Array, k uint16, s uint16) ([]byte, error) {
		tmp := Array{}
		tmp = append(tmp, (*a)[:k]...)
		*a = append(tmp, (*a)[k+2+s:]...)
		return []byte{}, nil
	}
	_, err := a.goToIndex(index, f)
	a.DecrementSize()
	return err
}

func (a *Array) goToIndex(index uint16, f action) ([]byte, error) {
	var offset uint16 = 3

	size, err := a.getSize()
	if err != nil {
		return []byte{}, err
	}

	if size < index {
		return []byte{}, errors.New("array index out of bounds")
	}

	var currentElement uint16 = 0
	//Since the Elements can be of variable size,
	//each Element has to be visited to know how many bytes it occupies

	var indexOnByteArray = offset
	for ; indexOnByteArray < uint16(len(*a)) && currentElement <= index; currentElement++ {
		elementSize, err := ByteArrayToUI16((*a)[indexOnByteArray : indexOnByteArray+2])

		if err != nil {
			return []byte{}, err
		}

		if currentElement == index {
			result, err := f(a, indexOnByteArray, elementSize)
			return result, err
		}
		indexOnByteArray += 2 + elementSize
	}

	return []byte{}, errors.New("array internals error")
}
