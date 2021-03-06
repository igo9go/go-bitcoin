package main

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"math/big"
)

type ProofOfWork struct {
	block *Block

	//目标值
	target *big.Int
}

func NewProofOfWork(block *Block) *ProofOfWork {
	pow := &ProofOfWork{block: block}

	//我们指定的难度值，现在是一个string类型，需要进行转换
	targetStr := "0000100000000000000000000000000000000000000000000000000000000000"

	//引入的辅助变量，目的是将上面的难度值转成big.int
	tmpInt := big.Int{}
	//将难度值赋值给big.int，指定16进制的格式
	tmpInt.SetString(targetStr, 16)

	pow.target = &tmpInt

	//targetLocal := big.NewInt(1)
	////targetLocal.Lsh(targetLocal, 256)
	////targetLocal.Rsh(targetLocal, difficulty)
	//targetLocal.Lsh(targetLocal, 256-difficulty)
	//pow.target = targetLocal
	return pow
}

//3. 提供计算不断计算hash的哈数
//
//- Run()

func (pow *ProofOfWork) Run() ([]byte, uint64) {
	//1. 拼装数据（区块的数据，还有不断变化的随机数）
	//2. 做哈希运算
	//3. 与pow中的target进行比较
	//a. 找到了，退出返回
	//b. 没找到，继续找，随机数加1

	var nonce uint64
	var hash [32]byte

	for {
		//1. 拼装数据（区块的数据，还有不断变化的随机数）
		hash = sha256.Sum256(pow.prepareData(nonce))

		tmpInt := big.Int{}
		tmpInt.SetBytes(hash[:])

		if tmpInt.Cmp(pow.target) == -1 {
			fmt.Printf("挖矿成功！hash : %x, nonce : %d\n", hash, nonce)
			return hash[:], nonce
		} else {
			nonce++
		}
	}
}

func (pow *ProofOfWork) prepareData(nonce uint64) []byte {
	block := pow.block
	tmp := [][]byte{
		Uint64ToByte(block.Version),
		block.PrevHash,
		block.MerKleRoot,
		Uint64ToByte(block.TimeStamp),
		Uint64ToByte(block.Difficulty),
		Uint64ToByte(nonce),
		//block.Data,

	}

	data := bytes.Join(tmp, []byte{})
	return data
}
//
//4. 提供一个校验函数
//
//- IsValid()

func (pow *ProofOfWork)IsValid()  bool{
	hash := sha256.Sum256(pow.prepareData(pow.block.Nonce))
	//fmt.Printf("is valid hash : %x, %d\n", hash[:], pow.block.Nonce)

	tTmp := big.Int{}
	tTmp.SetBytes(hash[:])
	if tTmp.Cmp(pow.target)  == -1 {
		return true
	}

	return false

	//return tTmp.Cmp(&pow.target)  == -1
}