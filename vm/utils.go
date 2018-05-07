package vm

import (
	rand1 "crypto/rand"
	"encoding/binary"
	"encoding/hex"
	"math/big"
	rand2 "math/rand"
	"time"

	"github.com/pkg/errors"
)

const UINT16_MAX uint16 = 65535

func UInt64ToByteArray(element uint64) []byte {
	ba := make([]byte, 8)
	binary.LittleEndian.PutUint64(ba, uint64(element))
	return ba
}

func UInt16ToByteArray(element uint16) []byte {
	ba := make([]byte, 2)
	binary.LittleEndian.PutUint16(ba, uint16(element))
	return ba
}

func ByteArrayToUI16(element []byte) (uint16, error) {
	if len(element) != 2 {
		return 0, errors.New("not a valid uint16 representation")
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
	ba := make([]byte, 64-len(element))
	ba = append(element, ba...)
	return int(binary.LittleEndian.Uint64(ba))
}

func BigIntToString(element big.Int) string {
	ba := element.Bytes()
	return string(ba[:])
}

func StrToByteArray(element string) []byte {
	return []byte(element)
}

func ByteArrayToString(element []byte) string {
	return string(element[:])
}

func RandomBytes() []byte {
	byteArray := make([]byte, RandomInt())
	rand1.Read(byteArray)
	return byteArray
}

func RandomInt() int {
	rand2.Seed(time.Now().Unix())
	return rand2.Intn(1000)
}
