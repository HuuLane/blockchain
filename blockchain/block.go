package blockchain

import (
	"bytes"
	"encoding/gob"
)

type Block struct {
	PrevHash []byte
	Data     []byte
	Hash     []byte
	Nonce    int
}

func NewBlock(prevHash []byte, data string) *Block {
	b := &Block{prevHash, []byte(data), nil, 0}
	pow := NewProof(b)
	nonce, checksum := pow.Run()

	b.Nonce = nonce
	b.Hash = checksum

	return b
}

func Genesis() *Block {
	return NewBlock(nil, "Genesis")
}

func (b *Block) Serialize() []byte {
	var res bytes.Buffer

	encoder := gob.NewEncoder(&res)
	err := encoder.Encode(b)
	Handle(err)

	return res.Bytes()
}

func Deserialize(data []byte) *Block {
	var block Block

	decoder := gob.NewDecoder(bytes.NewReader(data))
	err := decoder.Decode(&block)
	Handle(err)

	return &block
}
