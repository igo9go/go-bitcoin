package main

import (
	"bytes"
	"fmt"
	"github.com/boltdb/bolt"
	"log"
)

type BlockChain struct {
	//blocks []*Block

	db   *bolt.DB //key hash  value block.tobyte
	tail []byte   //存储最后一个区块的哈希
}

const blockChainDb = "blockChain.db"
const blockBucket = "blockBucket"

func NewBlockChain() *BlockChain {
	//最后一个区块的哈希， 从数据库中读出来的
	var lastHash []byte

	//1. 打开数据库
	db, err := bolt.Open(blockChainDb, 0600, nil)
//	defer db.Close()

	if err != nil {
		log.Panic("打开数据库失败！")
	}

	//将要操作数据库（改写）
	db.Update(func(tx *bolt.Tx) error {
		//2. 找到抽屉bucket(如果没有，就创建）
		bucket := tx.Bucket([]byte(blockBucket))
		if bucket == nil {
			//没有抽屉，我们需要创建
			bucket, err = tx.CreateBucket([]byte(blockBucket))
			if err != nil {
				log.Panic("创建bucket(b1)失败")
			}

			//创建一个创世块，并作为第一个区块添加到区块链中
			genesisBlock := GenesisBlock()

			//3. 写数据
			//hash作为key， block的字节流作为value，尚未实现
			bucket.Put(genesisBlock.Hash, genesisBlock.Serialize())
			bucket.Put([]byte("LastHashKey"), genesisBlock.Hash)
			lastHash = genesisBlock.Hash

			fmt.Println("写入创世块")
		} else {
			lastHash = bucket.Get([]byte("LastHashKey"))
		}

		return nil
	})

	return &BlockChain{db, lastHash}
}

func GenesisBlock() *Block {
	return NewBlock("first block", []byte{})
}

func (bc *BlockChain) AddBlock(data string) {
	/*
	   //获取最后一个区块
	   lastBlock := bc.blocks[len(bc.blocks) -1]
	   //获取最后一个区块的哈希,作为最新（当前）区块的前哈希
	   prevHash := lastBlock.Hash

	   block := NewBlock(data, prevHash)
	   bc.blocks = append(bc.blocks, &block)
	*/

	//获取最后区块的哈希值
	lastBlockHash := bc.tail
	fmt.Println("添加", lastBlockHash)

	bc.db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(blockBucket))
		fmt.Println("开始添加")
		//如果是空的，表明这个bucket没有创建，我们就要去创建它，然后再写数据。
		if bucket == nil {
			panic("err")
		} else {
			//创建新区块
			newBlock := NewBlock(data, lastBlockHash)
			//添加区块
			bucket.Put(newBlock.Hash, newBlock.Serialize())
			//更新最后区块的哈希值
			bucket.Put([]byte("LastHashKey"), newBlock.Hash)

			//这个别忘了，我们需要返回它
			bc.tail = newBlock.Hash

			fmt.Println("最后hash", newBlock.Hash)
			return nil
		}
		return nil
	})
	fmt.Println("成功添加一个区块")
}

func (bc *BlockChain) Printchain() {

	bc.db.View(func(tx *bolt.Tx) error {
		// Assume bucket exists and has keys
		b := tx.Bucket([]byte("blockBucket"))
		//从第一个key-> value 进行遍历，到最后一个固定的key时直接返回
		b.ForEach(func(k, v []byte) error {
			if bytes.Equal(k, []byte("LastHashKey")) {
				return nil
			}

			block := Deserialize(v)
			//fmt.Printf("key=%x, value=%s\n", k, v)
			fmt.Printf("===========================\n\n")
			fmt.Printf("版本号: %d\n", block.Version)
			fmt.Printf("前区块哈希值: %x\n", block.PrevHash)
			fmt.Printf("梅克尔根: %x\n", block.MerKleRoot)
			fmt.Printf("时间戳: %d\n", block.TimeStamp)
			fmt.Printf("难度值(随便写的）: %d\n", block.Difficulty)
			fmt.Printf("随机数 : %d\n", block.Nonce)
			fmt.Printf("当前区块哈希值: %x\n", block.Hash)
			fmt.Printf("区块数据 :%s\n", block.Data)
			return nil
		})
		return nil
	})
}
