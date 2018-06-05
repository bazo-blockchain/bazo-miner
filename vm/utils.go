package vm

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
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

func BoolToByteArray(value bool) []byte{
	var result byte
	if value {
		result = 1
	}
	return []byte{result}
}

func ByteArrayToBool(ba []byte) bool {
	return ba[0] == 1
}

func ConvertToBigInt(ba []byte, err error) (big.Int, error) {
	result := big.Int{}
	result.SetBytes(ba)
	return result, err
}

func ConvertToByteArray(bi big.Int) ([]byte) {
	result := bi.Bytes()
	return result
}