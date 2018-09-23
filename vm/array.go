package vm

import (
	"errors"
)

type action func(array *Array, index uint16, elementSize uint16) ([]byte, error)
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
		return Array{}, errors.New("not a valid array")
	}
	return Array(arr), nil
}

func (a *Array) getSize() (uint16, error) {
	if len(*a) < 3 {
		return 0, errors.New("not a valid array")
	}
	value, err := ByteArrayToUI16((*a)[1:3])
	if err != nil {
		return 0, errors.New("cannot get size of array")
	}
	return value, nil
}

func (a *Array) setSize(ba []byte) {
	//No checks because it is always used with getSize,
	//If the array is incorrect, getSize() should already have noticed
	(*a)[1] = ba[0]
	(*a)[2] = ba[1]
}

func (a *Array) IncrementSize() error {
	s, err := a.getSize()
	if err != nil {
		return errors.New("could not increase size")
	}
	s++
	a.setSize(UInt16ToByteArray(s))
	return nil
}

func (a *Array) DecrementSize() error {
	s, err := a.getSize()
	if err != nil {
		return err
	}

	if s <= 0 {
		return errors.New("Array size is already 0")
	}
	s--
	a.setSize(UInt16ToByteArray(s))
	return nil
}

func (a *Array) At(index uint16) ([]byte, error) {
	var f action = func(array *Array, i uint16, s uint16) ([]byte, error) {
		return (*array)[i+2 : i+2+s], nil
	}
	result, err := a.goToIndex(index, f)
	return result, err
}

func (a *Array) Insert(index uint16, element []byte) error {
	err := a.Remove(index)
	if err != nil {
		return err
	}

	var f action = func(array *Array, i uint16, s uint16) ([]byte, error) {
		tmp := Array{}
		tmp = append(tmp, (*a)[:i]...)
		tmp.Append(element)
		*a = append(tmp, (*a)[i:]...)
		return []byte{}, nil
	}
	_, err = a.goToIndex(index, f)
	return err
}

func (a *Array) Append(ba []byte) error {
	length := len(ba)

	if length > int(UINT16_MAX) {
		return errors.New("Element Size overflow")
	}

	sb := UInt16ToByteArray(uint16(len(ba)))
	*a = append(*a, append(sb, ba...)...)
	err := a.IncrementSize()
	return err
}

func (a *Array) Remove(index uint16) error {
	var f action = func(array *Array, k uint16, s uint16) ([]byte, error) {
		tmp := Array{}
		tmp = append(tmp, (*a)[:k]...)
		*a = append(tmp, (*a)[k+2+s:]...)
		return []byte{}, nil
	}
	_, err := a.goToIndex(index, f)
	if err != nil {
		return err
	}

	err = a.DecrementSize()
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
