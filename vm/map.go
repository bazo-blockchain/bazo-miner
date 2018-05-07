package vm

import (
	"bytes"
	"errors"
	"log"
	"math/big"
)

type Map []byte

func NewMap() Map {
	return []byte{0x01, 0x00, 0x00}
}

func (m *Map) ToBigInt() big.Int {
	mp := big.Int{}
	mp.SetBytes(*m)
	return mp
}

func MapFromBigInt(m big.Int) (Map, error) {
	ba := m.Bytes()
	if len(ba) <= 0 {
		return Map{}, errors.New("empty map")
	}
	if ba[0] != 0x01 {
		return Map{}, errors.New("invalid datatype supplied")
	}
	return Map(ba), nil
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

func (m *Map) GetVal(key []byte) ([]byte, error) {
	offset := 3
	l := len(*m)

	//bai stands for byteArrayIndex and is the index on the
	//byte array which the map is built upon
	for bai := offset; bai < l; {
		if l == 3 {
			return []byte{}, errors.New("no elements in map")
		}

		sizeOfKey, err := ByteArrayToUI16((*m)[bai : bai+2])

		if err != nil {
			return []byte{}, err
		}

		valueSizeStartIndex := bai + 2 + int(sizeOfKey)

		k := (*m)[bai+2 : valueSizeStartIndex]

		sizeOfValue, err := ByteArrayToUI16((*m)[valueSizeStartIndex : valueSizeStartIndex+2])
		if err != nil {
			return []byte{}, err
		}

		valueEndIndex := valueSizeStartIndex + 2 + int(sizeOfValue)
		v := (*m)[valueSizeStartIndex+2 : valueEndIndex]
		if bytes.Equal(key, k) {
			return v, nil
		}

		if bai == valueEndIndex {
			return []byte{}, errors.New("element sizes are 0")
		}
		bai = valueEndIndex
	}

	return []byte{}, errors.New("key not found")
}

func (m *Map) Remove(key []byte) error {
	offset := 3
	l := len(*m)

	//bai stands for byteArrayIndex and is the index on the
	//byte array which the map is built upon
	for bai := offset; bai < l; {
		if l == 3 {
			return errors.New("no elements in map")
		}

		sizeOfKey, err := ByteArrayToUI16((*m)[bai : bai+2])
		if err != nil {
			return err
		}

		valueSizeStartIndex := bai + 2 + int(sizeOfKey)

		k := (*m)[bai+2 : valueSizeStartIndex]
		sizeOfValue, err := ByteArrayToUI16((*m)[valueSizeStartIndex : valueSizeStartIndex+2])
		if err != nil {
			return err
		}

		valueEndIndex := valueSizeStartIndex + 2 + int(sizeOfValue)
		if bytes.Equal(key, k) {
			tmp := append([]byte{}, (*m)[:bai]...)
			*m = append(tmp, (*m)[valueEndIndex:]...)
			m.DecrementSize()
			return nil
		}

		if bai == valueEndIndex {
			return errors.New("element sizes are 0")
		}
		bai = valueEndIndex
	}

	return errors.New("key not found")
}
