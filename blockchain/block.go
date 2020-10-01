package blockchain

import (
	"bytes"
	"crypto/sha256"
)

type Block struct {
	PrevHash []byte
	Data     []byte
	Hash     []byte
}

func (b *Block) DeriveHash() {
	info := bytes.Join([][]byte{b.PrevHash, b.Data}, nil)
	hash := sha256.Sum256(info)
	// array to slice
	b.Hash = hash[:]
}

func NewBlock(prevHash []byte, data string) *Block {
	b := &Block{prevHash, []byte(data), nil}
	b.DeriveHash()
	return b
}

func Genesis() *Block {
	return NewBlock(nil, "Genesis")
}

type BlockChain struct {
	Blocks []*Block
}

func (c *BlockChain) AddBlock(data string) {
	// pb: prevBlock
	pb := c.Blocks[len(c.Blocks)-1]
	nb := NewBlock(pb.Hash, data)
	c.Blocks = append(c.Blocks, nb)
}

func NewBlockChain() *BlockChain {
	return &BlockChain{[]*Block{Genesis()}}
}
