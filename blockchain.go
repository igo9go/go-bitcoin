package main

type BlockChain struct {
	blocks []*Block
}

func NewBlockChain() *BlockChain {
	genesisBlock := GenesisBlock()
	return &BlockChain{
		blocks: []*Block{genesisBlock},
	}
}

func GenesisBlock() *Block {
	return NewBlock("first block", []byte{})
}

func (bc *BlockChain) AddBlock(data string) {
	prevHash := bc.blocks[len(bc.blocks)-1].Hash

	block := NewBlock(data, prevHash)
	bc.blocks = append(bc.blocks, block)
}
