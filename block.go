package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"time"
)

//0. 定义结构
type Block struct {
	//版本号
	Version uint64
	//前区块哈希
	PrevHash []byte

	//梅克尔根(就是一个哈希值，v4版本介绍）
	MerKleRoot []byte

	//时间戳
	TimeStamp uint64

	//难度值(调整比特币挖矿的难度)
	Difficulty uint64

	//随机数，这就是挖矿时所要寻找的数
	Nonce uint64

	// 当前区块哈希
	Hash []byte
	// 数据
	Data []byte
}

//2. 创建区块
func NewBlock(data string, prevBlockHash []byte) *Block {
	block := Block{
		Version:    00,
		PrevHash:   prevBlockHash,
		MerKleRoot: []byte{}, //先填为空，v4版本再详解
		TimeStamp:  uint64(time.Now().Unix()),
		Difficulty: 100,
		Nonce:      100,
		Data:       []byte(data),
		//Hash:    []byte{}, //先填为空，后面再进行计算
	}

	//block.SetHash()

	//挖矿
	pow := NewProofOfWork(&block)
	hash, nonce := pow.Run()
	//根据挖矿结果对区块数据进行更新（补充）
	block.Hash = hash
	block.Nonce = nonce
	return &block
}

//3. 生成哈希
func (block *Block) SetHash() {
	//1. 拼装数据
	var blockByteInfo []byte //存储拼接好的数据，最后作为sha256函数的参数
	//1. 拼接当前区块的数据
	//blockByteInfo = append(blockByteInfo, block.PrevHash...)
	//blockByteInfo = append(blockByteInfo, block.Data...)
	//blockByteInfo = append(blockByteInfo, block.MerKleRoot...)
	//blockByteInfo = append(blockByteInfo, Uint64ToByte(block.Version)...)
	//blockByteInfo = append(blockByteInfo, Uint64ToByte(block.TimeStamp)...)
	//blockByteInfo = append(blockByteInfo, Uint64ToByte(block.Difficulty)...)
	//blockByteInfo = append(blockByteInfo, Uint64ToByte(block.Nonce)...)
	//2. sha256

	tmp := [][]byte{
		block.PrevHash,
		block.Data,
		block.MerKleRoot,
		Uint64ToByte(block.Version),
		Uint64ToByte(block.TimeStamp),
		Uint64ToByte(block.Difficulty),
		Uint64ToByte(block.Nonce)}

	blockByteInfo = bytes.Join(tmp, []byte(""))

	hash := sha256.Sum256(blockByteInfo)
	block.Hash = hash[:]
}

func Uint64ToByte(num uint64) []byte {
	var buffer bytes.Buffer

	err := binary.Write(&buffer, binary.BigEndian, num)

	if err != nil {
		panic(err)
	}

	return buffer.Bytes()
}