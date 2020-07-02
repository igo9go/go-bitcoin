package main

import (
	"fmt"
	"unsafe"
)


func main() {
	bc := NewBlockChain()
	bc.AddBlock("121")
	bc.AddBlock("222")

	//bc.Printchain()
	defer bc.db.Close()

	//定义迭代器
	it := NewBlockChainIterator(bc)

	for {
		//调用迭代器访问函数，返回当前block，并且向左移动
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

		//终止条件
		if len(block.PrevHash)  == 0 {
			fmt.Println("print over!")
			break
		}
	}
	//for i, block := range bc.blocks {
	//
	//	fmt.Printf("======== 当前区块高度： %d ========\n", i)
	//	fmt.Printf("前区块哈希值： %x\n", block.PrevHash)
	//	fmt.Printf("当前区块哈希值： %x\n", block.Hash)
	//	fmt.Printf("区块数据 :%s\n", block.Data)
	//	fmt.Printf("随机数 :%x\n", block.Nonce)
	//	//fmt.Println("hello")
	//}

	s := int16(0x1234)
	b := int8(s)
	//0x1234
	// 低 --------》 高
	// 12 34  -> 大端 -> 高尾端
	// 34 12  -> 小端 -> 低尾端

	fmt.Println("int16字节大小为", unsafe.Sizeof(s)) //结果为2
	if 0x34 == b {
		fmt.Println("little endian")
	} else {
		fmt.Println("big endian")
	}
}
