package blockchain

type Block struct {
	PrevHash []byte
	Data     []byte
	Hash     []byte
	Nonce    int
}

func NewBlock(prevHash []byte, data string) *Block {
	b := &Block{prevHash, []byte(data), nil, 0}
	pow := NewProofOfWork(b)
	nonce, checksum := pow.Run()

	b.Nonce = nonce
	b.Hash = checksum

	return b
}

func Genesis() *Block {
	return NewBlock(nil, "Genesis")
}
