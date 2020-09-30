package main

import (
	"bytes"
	"crypto/sha256"
	"fmt"
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
	blocks []*Block
}

func (c *BlockChain) AddBlock(data string) {
	// pb: prevBlock
	pb := c.blocks[len(c.blocks)-1]
	nb := NewBlock(pb.Hash, data)
	c.blocks = append(c.blocks, nb)
}

func NewBlockChain() *BlockChain {
	return &BlockChain{[]*Block{Genesis()}}
}

func main() {
	chain := NewBlockChain()

	chain.AddBlock("First Block after Genesis")
	chain.AddBlock("Second Block after Genesis")
	chain.AddBlock("Third Block after Genesis")

	for _, block := range chain.blocks {
		fmt.Printf("Previous Hash: %x\n", block.PrevHash)
		fmt.Printf("Data in Block: %s\n", block.Data)
		fmt.Printf("Hash: %x\n\n", block.Hash)
	}
}
