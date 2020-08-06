package main

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"github.com/btcsuite/btcutil/base58"
	"golang.org/x/crypto/ripemd160"
)

type Wallet struct {
	//约定，这里的PubKey不存储原始的公钥，而是存储X和Y拼接的字符串，在校验端重新拆分（参考r,s传递）
	PubKey  []byte
	Private *ecdsa.PrivateKey
}

func NewWallet() *Wallet {
	//创建曲线
	curve := elliptic.P256()
	//生成私钥
	privateKey, err := ecdsa.GenerateKey(curve, rand.Reader)
	if err != nil {
		panic(err)
	}

	pubKeyOrig := privateKey.PublicKey

	//拼接X, Y
	pubKey := append(pubKeyOrig.X.Bytes(), pubKeyOrig.Y.Bytes()...)

	return &Wallet{Private: privateKey, PubKey: pubKey}
}

//生成地址
func (w *Wallet) NewAddress() string {

	version := byte(00)

	pubKey := w.PubKey

	rip160HashValue := HashPubKey(pubKey)

	//拼接version
	payload := append([]byte{version}, rip160HashValue...)

	//checksum
	checkCode := CheckSum(payload) //取前4字节

	//25字节数据
	payload = append(payload, checkCode...) //21 + 4

	//go语言有一个库，叫做btcd,这个是go语言实现的比特币全节点源码
	address := base58.Encode(payload)

	return address
}

func HashPubKey(data []byte) []byte {
	hash := sha256.Sum256(data)
	//理解为编码器
	rip160hasher := ripemd160.New()
	_, err := rip160hasher.Write(hash[:])

	if err != nil {
		panic(err)
	}

	//返回rip160的哈希结果
	rip160HashValue := rip160hasher.Sum(nil)
	return rip160HashValue
}

func CheckSum(data []byte) []byte {
	//两次sha256
	hash1 := sha256.Sum256(data)
	hash2 := sha256.Sum256(hash1[:])

	//前4字节校验码
	checkCode := hash2[:4]
	return checkCode
}

func IsValidAddress(address string) bool {
	addressByte := base58.Decode(address)

	//防止输入的过短引起下面访问越界
	if len(addressByte) < 4 {
		return false
	}

	payload := addressByte[:len(addressByte)-4]
	checksum1 := addressByte[len(addressByte)-4:] //后4位 校验位
	checksum2 := CheckSum(payload)
	return bytes.Equal(checksum1, checksum2)
}
