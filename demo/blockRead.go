package main

import (
	"bytes"
	"fmt"
	"github.com/boltdb/bolt"
	"log"
)

func main()  {
	db, err := bolt.Open("blockChain.db", 0600, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()


	db.View(func(tx *bolt.Tx) error {
		// Assume bucket exists and has keys
		b := tx.Bucket([]byte("blockBucket"))
		//从第一个key-> value 进行遍历，到最后一个固定的key时直接返回
		b.ForEach(func(k, v []byte) error {
			if bytes.Equal(k, []byte("LastHashKey")) {
				return nil
			}

			fmt.Printf("key=%x, value=%s\n", k, v)

			//block := Deserialize(v)
			//fmt.Printf("===========================\n\n")
			//fmt.Printf("版本号: %d\n", block.Version)
			//fmt.Printf("前区块哈希值: %x\n", block.PrevHash)
			//fmt.Printf("梅克尔根: %x\n", block.MerKleRoot)
			//fmt.Printf("时间戳: %d\n", block.TimeStamp)
			//fmt.Printf("难度值(随便写的）: %d\n", block.Difficulty)
			//fmt.Printf("随机数 : %d\n", block.Nonce)
			//fmt.Printf("当前区块哈希值: %x\n", block.Hash)
			//fmt.Printf("区块数据 :%s\n", block.Data)
			return nil
		})
		return nil
	})
}
