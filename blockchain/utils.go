package blockchain

import (
	"bytes"
	"encoding/binary"
	"log"
	"math/big"
)

func Handle(err error) {
	if err != nil {
		log.Panic(err)
	}
}

func New32BytesArrWithRsh(n int) *[32]byte {
	if n < 1 || n > 32 {
		panic("shift out of bounds")
	}
	var res [32]byte

	i := big.NewInt(1)
	i.Lsh(i, uint(256-n))
	i.FillBytes(res[:])

	return &res
}

func ToBytes(num int64) []byte {
	buff := new(bytes.Buffer)
	err := binary.Write(buff, binary.BigEndian, num)
	if err != nil {
		log.Panic(err)
	}

	return buff.Bytes()
}

func Cmp32BytesArr(a, b *[32]byte) int {
	// for BigEndian
	for i := 0; i < 32; i++ {
		if a[i] < b[i] {
			return -1
		} else if a[i] > b[i] {
			return 1
		}
	}
	return 0
}
