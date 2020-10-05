package blockchain

import (
	"bytes"
	"crypto/sha256"
	"encoding/gob"
)

type Block struct {
	PrevHash     []byte
	Transactions []*Transaction
	Hash         []byte
	Nonce        int
}

func NewBlock(prevHash []byte, txs []*Transaction) *Block {
	b := &Block{prevHash, txs, nil, 0}
	pow := NewProof(b)
	nonce, checksum := pow.Run()

	b.Nonce = nonce
	b.Hash = checksum

	return b
}

func Genesis(coinbase *Transaction) *Block {
	return NewBlock(nil, []*Transaction{coinbase})
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

func (b *Block) TransactionsChecksum() []byte {
	var txHashes [][]byte
	for _, tx := range b.Transactions {
		txHashes = append(txHashes, tx.ID)
	}

	checksum := sha256.Sum256(bytes.Join(txHashes, []byte{}))

	return checksum[:]
}
