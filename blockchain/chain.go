package blockchain

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
