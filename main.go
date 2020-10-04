package main

import (
	"fmt"
	"strconv"

	"github.com/HuuLane/blockchain/blockchain"
)

func main() {
	chain := blockchain.New()

	chain.AddBlock("First Block after Genesis")
	chain.AddBlock("Second Block after Genesis")
	chain.AddBlock("Third Block after Genesis")

	iter := chain.Iterator()
	for b := iter.Next(); b.PrevHash != nil; b = iter.Next() {

		fmt.Printf("Previous Hash: %x\n", b.PrevHash)
		fmt.Printf("Data in Block: %s\n", b.Data)
		fmt.Printf("Hash: %x\n", b.Hash)

		pow := blockchain.NewProofOfWork(b)
		fmt.Printf("PoW: %s\n", strconv.FormatBool(pow.Validate()))
		fmt.Println()

	}
}
