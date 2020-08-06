package main

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/gob"
	"fmt"
	"log"
	"math/big"
	"strings"
)

const reward = 50

//1 定义交易结构

type Transaction struct {
	TXID      []byte
	TXInputs  []TXInput
	TXOutputs []TXOutput
}

//交易输入
type TXInput struct {
	//引用的交易ID
	TXid []byte
	//引用的outPut index
	Index int64
	//解锁脚本
	//Sig string

	//真正的数字签名，由r，s拼成的[]byte
	Signature []byte

	//约定，这里的PubKey不存储原始的公钥，而是存储X和Y拼接的字符串，在校验端重新拆分（参考r,s传递）
	//注意，是公钥，不是哈希，也不是地址
	PubKey []byte
}

type TXOutput struct {
	//转账的金额
	Value float64
	//锁定脚本
	//PubKeyHash string

	//收款方的公钥的哈希，注意，是哈希而不是公钥，也不是地址
	PubKeyHash []byte
}

//根据地址获取改地址的公钥hash
func (output *TXOutput) Lock(address string) {
	output.PubKeyHash = GetPubKeyFromAddress(address)
}

//给TXOutput提供一个创建的方法，否则无法调用Lock
func NewTXOutput(value float64, address string) *TXOutput {
	output := TXOutput{
		Value: value,
	}

	output.Lock(address)
	return &output
}

//2. 提供创建交易方法(挖矿交易)
func NewCoinbaseTX(address string, data string) *Transaction {
	//挖矿交易的特点：
	//1. 只有一个input
	//2. 无需引用交易id
	//3. 无需引用index
	//矿工由于挖矿时无需指定签名，所以这个sig字段可以由矿工自由填写数据，一般是填写矿池的名字
	//签名先填写为空，后面创建完整交易后，最后做一次签名即可
	input := TXInput{[]byte{}, -1, nil, []byte(data)}

	//新的创建方法
	output := NewTXOutput(reward, address)

	//对于挖矿交易来说，只有一个input和一output
	tx := Transaction{[]byte{}, []TXInput{input}, []TXOutput{*output}}
	tx.SetHash()

	return &tx
}

//实现一个函数，判断当前的交易是否为挖矿交易
func (tx *Transaction) IsCoinbase() bool {
	//1. 交易input只有一个
	//if len(tx.TXInputs) == 1  {
	//	input := tx.TXInputs[0]
	//	//2. 交易id为空
	//	//3. 交易的index 为 -1
	//	if !bytes.Equal(input.TXid, []byte{}) || input.Index != -1 {
	//		return false
	//	}
	//}
	//return true
	if len(tx.TXInputs) == 1 && len(tx.TXInputs[0].TXid) == 0 && tx.TXInputs[0].Index == -1 {
		return true
	}

	return false
}

//设置交易ID
func (tx *Transaction) SetHash() {
	var buffer bytes.Buffer

	encoder := gob.NewEncoder(&buffer)

	err := encoder.Encode(tx)
	if err != nil {
		log.Panic(err)
	}

	data := buffer.Bytes()
	hash := sha256.Sum256(data)
	tx.TXID = hash[:]
}

func NewTransaction(from, to string, amount float64, bc *BlockChain) *Transaction {
	//1. 创建交易之后要进行数字签名->所以需要私钥->打开钱包"NewWallets()"
	ws := NewWallets()

	//2. 找到自己的钱包，根据地址返回自己的wallet
	wallet := ws.WalletsMap[from]
	if wallet == nil {
		fmt.Printf("没有找到该地址的 %s 钱包，交易创建失败!\n", from)
		return nil
	}

	//3. 得到对应的公钥，私钥
	pubKey := wallet.PubKey
	privateKey := wallet.Private

	pubKeyHash := HashPubKey(pubKey)

	//1. 找到最合理UTXO集合 map[string][]uint64
	utxos, resValue := bc.FindNeedUTXOs(pubKeyHash, amount)

	if resValue < amount {
		fmt.Printf("余额不足，交易失败!")
		return nil
	}

	var inputs []TXInput
	var outputs []TXOutput

	//2. 创建交易输入, 将这些UTXO逐一转成inputs
	//map[2222] = []int64{0}
	//map[3333] = []int64{0, 1}
	for id, indexArray := range utxos {
		for _, i := range indexArray {
//			input := TXInput{[]byte(id), int64(i), from}
			input := TXInput{[]byte(id), int64(i), nil, pubKey}
			inputs = append(inputs, input)
		}
	}

	//创建交易输出
	//output := TXOutput{amount, to}
	output := NewTXOutput(amount, to)
	outputs = append(outputs, *output)

	//找零
	if resValue > amount {
		output = NewTXOutput(resValue-amount, from)
		outputs = append(outputs, *output)
	}

	tx := Transaction{[]byte{}, inputs, outputs}
	tx.SetHash()

	bc.SignTransaction(&tx, privateKey)
	return &tx
}

//签名的具体实现,
// 参数为：私钥，inputs里面所有引用的交易的结构map[string]Transaction
//map[2222]Transaction222
//map[3333]Transaction333
func (tx *Transaction) Sign(privateKey *ecdsa.PrivateKey, prevTXs map[string]Transaction) {
	fmt.Println("begin Sign ...")
	//挖矿交易不需要签名
	if tx.IsCoinbase() {
		return
	}

	fmt.Printf("====== Sign tx : %x ======\n", tx.TXID)

	//确保引用的交易都是有效的，tx中的每个input都应该对应一条交易
	for _, input := range tx.TXInputs {
		if prevTXs[string(input.TXid)].TXID == nil {
			//if prevTXs[hex.EncodeToString(input.TXID)].TXID == nil {
			log.Panic("Previous txs are not valid!")
		}
	}

	//获取要签名的信息
	//我们对把当前的交易签名，然后存放到Signature中。
	//将当前的交易复制一份，然后做签名处理
	txCopy := tx.TrimmedCopy()

	//再次强调一下我们要签名的三部分：
	//- 欲使用的utxo中的pubKeyHash（这描述了付款人）
	//- 新生成的utxo中的pubKeyHash（这描述了收款人）
	//- 转账金额

	//注意，遍历的是txCopy，而不是tx本身
	for index, input := range txCopy.TXInputs {
		prevTX := prevTXs[string(input.TXid)]

		//使用input的PublicKey字段暂时存储一下这个想要解锁的utxo的公钥哈希
		//这里有个坑！！！
		//我一直使用input.PublicKey来赋值，但是其实range会复制一个新的变量，而不会修改你遍历的的txCopy， 这里一定要小心！！！，所以下面的这就要修改如下：
		//input.PublicKey = prevTX.TXOutputs[input.VoutIndex].PublicKeyHash
		txCopy.TXInputs[index].PubKey = prevTX.TXOutputs[input.Index].PubKeyHash
		//至此三个想要签名的数据都有了, 如何签名？签名一般对数据的哈希值进行签名
		//要签名的哈希存放在txCopy.TXID中
		txCopy.SetHash()
		//一定要赋值为nil，下面交易签名还要用的，虽然不用这个input了，但是它会污染其他交易
		//在校验过程时，也会生成txCopy，设置nil双方才能与生成一致的数据
		txCopy.TXInputs[index].PubKey = nil
		fmt.Printf("data to sign (TXID) : %x\n", txCopy.TXID)

		r, s, err := ecdsa.Sign(rand.Reader, privateKey, txCopy.TXID)
		if err != nil {
			fmt.Println("ecdsa.Sign failed!")
			log.Panic(err)
		}

		//将得到的r,s拼接成[]byte，放在signature中
		signature := append(r.Bytes(), s.Bytes()...)
		//tx的索引和txCopy的是一致的
		tx.TXInputs[index].Signature = signature
		fmt.Printf("signature : %x\n", signature)
	}
}

func (tx *Transaction) TrimmedCopy() Transaction {
	var inputs []TXInput
	var outputs []TXOutput

	for _, input := range tx.TXInputs {
		inputs = append(inputs, TXInput{input.TXid, input.Index, nil, nil})
	}

	for _, output := range tx.TXOutputs {
		outputs = append(outputs, TXOutput{output.Value, output.PubKeyHash})
	}

	txCopy := Transaction{tx.TXID, inputs, outputs}
	return txCopy
}


//分析校验：
//所需要的数据：公钥，数据(txCopy，生成哈希), 签名
//我们要对每一个签名过得input进行校验

func (tx *Transaction) Verify(prevTXs map[string]Transaction) bool {
	if tx.IsCoinbase() {
		return true
	}

	//1. 得到签名的数据
	txCopy := tx.TrimmedCopy()

	for i, input := range tx.TXInputs {
		prevTX := prevTXs[string(input.TXid)]
		if len(prevTX.TXID) == 0 {
			log.Panic("引用的交易无效")
		}

		txCopy.TXInputs[i].PubKey = prevTX.TXOutputs[input.Index].PubKeyHash
		txCopy.SetHash()
		dataHash := txCopy.TXID
		//2. 得到Signature, 反推会r,s
		signature := input.Signature //拆，r,s
		//3. 拆解PubKey, X, Y 得到原生公钥
		pubKey := input.PubKey //拆，X, Y


		//1. 定义两个辅助的big.int
		r := big.Int{}
		s := big.Int{}

		//2. 拆分我们signature，平均分，前半部分给r, 后半部分给s
		r.SetBytes(signature[:len(signature)/2 ])
		s.SetBytes(signature[len(signature)/2:])


		//a. 定义两个辅助的big.int
		X := big.Int{}
		Y := big.Int{}

		//b. pubKey，平均分，前半部分给X, 后半部分给Y
		X.SetBytes(pubKey[:len(pubKey)/2 ])
		Y.SetBytes(pubKey[len(pubKey)/2:])

		pubKeyOrigin := ecdsa.PublicKey{elliptic.P256(), &X, &Y}

		//4. Verify
		if !ecdsa.Verify(&pubKeyOrigin, dataHash, &r, &s) {
			return false
		}
	}

	return true
}


func (tx Transaction) String() string {
	var lines []string

	lines = append(lines, fmt.Sprintf("--- Transaction %x:", tx.TXID))

	for i, input := range tx.TXInputs {

		lines = append(lines, fmt.Sprintf("     Input %d:", i))
		lines = append(lines, fmt.Sprintf("       TXID:      %x", input.TXid))
		lines = append(lines, fmt.Sprintf("       Out:       %d", input.Index))
		lines = append(lines, fmt.Sprintf("       Signature: %x", input.Signature))
		lines = append(lines, fmt.Sprintf("       PubKey:    %x", input.PubKey))
	}

	for i, output := range tx.TXOutputs{
		lines = append(lines, fmt.Sprintf("     Output %d:", i))
		lines = append(lines, fmt.Sprintf("       Value:  %f", output.Value))
		lines = append(lines, fmt.Sprintf("       Script: %x", output.PubKeyHash))
	}

	return strings.Join(lines, "\n")
}

