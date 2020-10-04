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

var (
	singleton *BlockChain
)

func NewBlockChain() *BlockChain {
	if singleton == nil {
		singleton = &BlockChain{[]*Block{Genesis()}}
	}
	return singleton
}
