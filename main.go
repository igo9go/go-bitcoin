package main

import (
	"fmt"
	"unsafe"
)


func main() {
	bc := NewBlockChain()
	bc.AddBlock("hello")
	bc.AddBlock("world")

	for i, block := range bc.blocks {

		fmt.Printf("======== 当前区块高度： %d ========\n", i)
		fmt.Printf("前区块哈希值： %x\n", block.PrevHash)
		fmt.Printf("当前区块哈希值： %x\n", block.Hash)
		fmt.Printf("区块数据 :%s\n", block.Data)
		fmt.Printf("随机数 :%x\n", block.Nonce)

		//fmt.Println("hello")
	}

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
