package main

import (
	"os"
	"fmt"
)

type CLI struct {
	bc *BlockChain
}

const Usage = `
    ./blockchain addBlock --data DATA "add a block"
    ./blockchain printChain "print block Chain"
`

func (cli *CLI) Run() {
	if len(os.Args) < 2 {
		fmt.Println(Usage)
		os.Exit(1)
	}

	cmd := os.Args[1]

	switch cmd {
	case "addBlock":
		if len(os.Args) > 3 && os.Args[2] == "--data" {
			data := os.Args[3]
			if data == "" {
				fmt.Println("data should not be empty!")
				os.Exit(1)
			}
			cli.addBlock(data)
		} else {
			fmt.Println(Usage)
		}
	case "printChain":
		cli.printChain()
	default:
		fmt.Println(Usage)
	}
}

func (cli *CLI)addBlock(data string)  {
	cli.bc.AddBlock(data)
}

func (cli *CLI)printChain()  {

	//定义迭代器
	it := NewBlockChainIterator(cli.bc)
	for {

		block := it.GetBlockAndMoveLeft()

		fmt.Println(" ============== =============")
		fmt.Printf("Version : %d\n", block.Version)
		fmt.Printf("PrevBlockHash : %x\n", block.PrevHash)
		fmt.Printf("Hash : %x\n", block.Hash)
		fmt.Printf("MerkleRoot : %x\n", block.MerKleRoot)
		fmt.Printf("TimeStamp : %d\n", block.TimeStamp)
		fmt.Printf("Difficuty : %d\n", block.Difficulty)
		fmt.Printf("Nonce : %d\n", block.Nonce)
		fmt.Printf("Data : %s\n", block.Data)
		pow := NewProofOfWork(&block)
		fmt.Printf("IsValid : %v\n", pow.IsValid())


		if len(block.PrevHash)  == 0 {
			fmt.Println("print over!")
			break
		}
	}
}