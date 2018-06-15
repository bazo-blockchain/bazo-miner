package vm

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"math/big"
)

const UINT16_MAX uint16 = 65535

func UInt64ToByteArray(element uint64) []byte {
	ba := make([]byte, 8)
	binary.BigEndian.PutUint64(ba, uint64(element))
	return ba
}

func UInt16ToByteArray(element uint16) []byte {
	ba := make([]byte, 2)
	binary.LittleEndian.PutUint16(ba, uint16(element))
	return ba
}

func ByteArrayToUI16(element []byte) (uint16, error) {
	if bytes.Equal([]byte{}, element) {
		return 0, nil
	}
	if len(element) != 2 {
		return 0, errors.New("byte array to uint16 invalid parameters provided")
	}

	return binary.LittleEndian.Uint16(element), nil
}

func StrToBigInt(element string) big.Int {
	var result big.Int
	hexEncoded := hex.EncodeToString([]byte(element))
	result.SetString(hexEncoded, 16)
	return result
}

func ByteArrayToInt(element []byte) int {
	ba := make([]byte, 8-len(element))
	ba = append(ba, element...)
	return int(binary.BigEndian.Uint64(ba))
}

func BigIntToString(element big.Int) string {
	ba := element.Bytes()
	return string(ba[:])
}

func BoolToByteArray(value bool) []byte {
	var result byte
	if value {
		result = 1
	}
	return []byte{result}
}

func ByteArrayToBool(ba []byte) bool {
	return ba[0] == 1
}

func SignedBigIntConversion(ba []byte, err error) (big.Int, error) {
	if err != nil {
		return big.Int{}, err
	} else {
		result := big.Int{}

		if ba[0] != 0x01 && ba[0] != 0x00 {
			return big.Int{}, errors.New("Invalid signing bit")
		}

		result.SetBytes(ba[1:])

		if ba[0] == 0x01 {
			result.Neg(&result)
		}

		return result, err
	}
}

func UnsignedBigIntConversion(ba []byte, err error) (big.Int, error) {
	if err != nil {
		return big.Int{}, err
	} else {
		result := big.Int{}
		result.SetBytes(ba)
		return result, err
	}
}

func SignedByteArrayConversion(bi big.Int) []byte {
	var result []byte
	if bi.Sign() == 0 || bi.Sign() == 1 {
		result = []byte{0x00}
	} else {
		result = []byte{0x01}
	}
	result = append(result, bi.Bytes()...)

	return result
}

func BigIntToPushableBytes(element big.Int) []byte {
	baseLength := byte(len(element.Bytes()))

	var baseVal []byte
	baseVal = append(baseVal, baseLength)

	if element.IsUint64() == true {
		baseVal = append(baseVal, 0) // signing byte
	} else {
		baseVal = append(baseVal, 1) // signing byte
	}

	baseVal = append(baseVal, element.Bytes()...) // value
	return baseVal
}
