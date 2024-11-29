package utils

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"github.com/pactus-project/pactus/util/bech32m"
	"math/big"
)

func LoadUint8(reader *bytes.Reader) (uint8, error) {
	b := make([]byte, 1)
	_, err := reader.Read(b)
	return b[0], err
}
func Uint8ToBytes(n uint8) []byte {
	return []byte{n}
}

func BoolToByte(ok bool) uint8 {
	if ok {
		return 1
	}
	return 0
}

func LoadBool(reader *bytes.Reader) (bool, error) {
	b := make([]byte, 1)
	_, err := reader.Read(b)
	if err != nil {
		return false, err
	}
	if b[0] == 0 {
		return false, nil
	}
	if b[0] == 1 {
		return true, nil
	}
	return false, fmt.Errorf("invalid value for bool | v=%v", b[0])
}
func LoadUint16(reader *bytes.Reader) (uint16, error) {
	b := make([]byte, 2)
	_, err := reader.Read(b)
	return binary.LittleEndian.Uint16(b), err
}
func Uint16ToLittleBytes(n uint16) []byte {
	b := make([]byte, 2)
	binary.LittleEndian.PutUint16(b, n)
	return b
}

func LoadUint32(reader *bytes.Reader) (uint32, error) {
	b := make([]byte, 4)
	_, err := reader.Read(b)
	return binary.LittleEndian.Uint32(b), err
}
func Uint32ToLittleBytes(n uint32) []byte {
	b := make([]byte, 4)
	binary.LittleEndian.PutUint32(b, n)
	return b
}

func LoadUint64(reader *bytes.Reader) (uint64, error) {
	b := make([]byte, 8)
	_, err := reader.Read(b)
	return binary.LittleEndian.Uint64(b), err
}

func Uint64ToLittleBytes(n uint64) []byte {
	b := make([]byte, 8)
	binary.LittleEndian.PutUint64(b, n)
	return b
}

func LoadInt64(reader *bytes.Reader) (int64, error) {
	b := make([]byte, 8)
	_, err := reader.Read(b)
	return int64(binary.LittleEndian.Uint64(b)), err
}

func Int64ToLittleBytes(n int64) []byte {
	b := make([]byte, 8)
	binary.LittleEndian.PutUint64(b, uint64(n))
	return b
}

func LoadUint128(reader *bytes.Reader) (*big.Int, error) {
	b := make([]byte, 16)
	_, err := reader.Read(b)
	if err != nil {
		return nil, err
	}
	return LoadBigIntFromLittleBytes(b), nil
}

func Uint128ToLittleBytes(n *big.Int) []byte {
	return BigIntToLittleBytes(n.String(), 16)
}

func LoadUint256(reader *bytes.Reader) (*big.Int, error) {
	b := make([]byte, 32)
	_, err := reader.Read(b)
	if err != nil {
		return nil, err
	}
	return LoadBigIntFromLittleBytes(b), nil
}

func Uint256ToLittleBytes(n *big.Int) []byte {
	return BigIntToLittleBytes(n.String(), 32)
}

func ReverseBytes(b []byte) []byte {
	var reverse []byte
	for i, _ := range b {
		reverse = append(reverse, b[len(b)-i-1])
	}
	return reverse
}
func LoadBigIntFromLittleBytes(b []byte) *big.Int {
	return new(big.Int).SetBytes(ReverseBytes(b))
}

func BigIntToLittleBytes(bigIntString string, length int) []byte {
	n := new(big.Int)
	n.SetString(bigIntString, 10)
	if n.BitLen() > length*8 {
		// max 2**256-1
		return []byte{}
	}
	p := make([]byte, length)
	for i, _ := range p {
		// (n >> i * 8) & 0xff
		z := new(big.Int).And(new(big.Int).Rsh(n, uint(i)*8), new(big.Int).SetInt64(0xff))
		p[i] = uint8(z.Int64())
	}
	return p
}

func UintToBytesWithLittle(n uint64, length ...int) []byte {
	// max 2**64-1
	l := 1
	if len(length) > 0 {
		l = length[0]
	}
	var b []byte
	for i := 0; i < l; i++ {
		b = append(b, byte((n>>(i*8))&0xff))
	}
	return b
}

func Bech32Decode(s string) (string, []byte, error) {
	hrp, b, err := bech32m.Decode(s)
	if err != nil {
		return "", nil, err
	}
	b, err = bech32m.ConvertBits(b, 5, 8, false)
	return hrp, b, err
}
