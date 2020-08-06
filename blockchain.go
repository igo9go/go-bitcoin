package main

import (
	"bytes"
	"crypto/ecdsa"
	"fmt"
	"github.com/boltdb/bolt"
	"log"
	"errors"
)

type BlockChain struct {
	//blocks []*Block

	db   *bolt.DB //key hash  value block.tobyte
	tail []byte   //存储最后一个区块的哈希
}

const blockChainDb = "blockChain.db"
const blockBucket = "blockBucket"

func NewBlockChain(address string) *BlockChain {
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
			genesisBlock := GenesisBlock(address)

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

func GenesisBlock(address string) *Block {
	coinBase := NewCoinbaseTX(address, "first block")
	return NewBlock([]*Transaction{coinBase}, []byte{})
}

func (bc *BlockChain) AddBlock(txs []*Transaction) {

	for _, tx := range txs {
		if !bc.VerifyTransaction(tx) {
			fmt.Printf("矿工发现无效交易!")
			return
		}
	}

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
			newBlock := NewBlock(txs, lastBlockHash)
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
			fmt.Printf("区块数据 :%s\n", block.Transactions[0].TXInputs[0].PubKey)
			return nil
		})
		return nil
	})
}

//找到指定地址的所有的utxo
func (bc *BlockChain) FindUTXOs(pubKeyHash []byte) []TXOutput {
	var UTXO []TXOutput

	txs := bc.FindUTXOTransactions(pubKeyHash)

	for _, tx := range txs {
		for _, output := range tx.TXOutputs {
			if bytes.Equal(pubKeyHash, output.PubKeyHash) {
				UTXO = append(UTXO, output)
			}
		}
	}

	return UTXO
}

func (bc *BlockChain) FindNeedUTXOs(senderPubKeyHash []byte, amount float64) (map[string][]uint64, float64) {

	//找到的合理的utxos集合
	utxos := make(map[string][]uint64)
	var calc float64

	txs := bc.FindUTXOTransactions(senderPubKeyHash)

	for _, tx := range txs {
		for i, output := range tx.TXOutputs {
			if bytes.Equal(senderPubKeyHash, output.PubKeyHash) {
				//fmt.Printf("222222")
				//UTXO = append(UTXO, output)
				//fmt.Printf("333333 : %f\n", UTXO[0].Value)
				//我们要实现的逻辑就在这里，找到自己需要的最少的utxo
				//3. 比较一下是否满足转账需求
				//   a. 满足的话，直接返回 utxos, calc
				//   b. 不满足继续统计

				if calc < amount {
					//1. 把utxo加进来，
					//utxos := make(map[string][]uint64)
					//array := utxos[string(tx.TXID)] //确认一下是否可行！！
					//array = append(array, uint64(i))
					utxos[string(tx.TXID)] = append(utxos[string(tx.TXID)], uint64(i))
					//2. 统计一下当前utxo的总额
					//第一次进来: calc =3,  map[3333] = []uint64{0}
					//第二次进来: calc =3 + 2,  map[3333] = []uint64{0, 1}
					//第三次进来：calc = 3 + 2 + 10， map[222] = []uint64{0}
					calc += output.Value

					//加完之后满足条件了，
					if calc >= amount {
						//break
						fmt.Printf("找到了满足的金额：%f\n", calc)
						return utxos, calc
					}
				} else {
					fmt.Printf("不满足转账金额,当前总额：%f， 目标金额: %f\n", calc, amount)
				}
			}
		}
	}

	return utxos, calc
}

func (bc *BlockChain) FindUTXOTransactions(senderPubKeyHash []byte) []*Transaction {
	var txs []*Transaction //存储所有包含utxo交易集合
	//我们定义一个map来保存消费过的output，key是这个output的交易id，value是这个交易中索引的数组
	//map[交易id][]int64
	spentOutputs := make(map[string][]int64)

	//创建迭代器
	it := bc.NewIterator()

	for {
		//1.遍历区块
		block := it.Next()

		//2. 遍历交易
		for _, tx := range block.Transactions {
			//fmt.Printf("current txid : %x\n", tx.TXID)

		OUTPUT:
			//3. 遍历output，找到和自己相关的utxo(在添加output之前检查一下是否已经消耗过)
			//	i : 0, 1, 2, 3
			for i, output := range tx.TXOutputs {
				//fmt.Printf("current index : %d\n", i)
				//在这里做一个过滤，将所有消耗过的outputs和当前的所即将添加output对比一下
				//如果相同，则跳过，否则添加
				//如果当前的交易id存在于我们已经表示的map，那么说明这个交易里面有消耗过的output

				//map[2222] = []int64{0}
				//map[3333] = []int64{0, 1}
				//这个交易里面有我们消耗过得output，我们要定位它，然后过滤掉
				if spentOutputs[string(tx.TXID)] != nil {
					for _, j := range spentOutputs[string(tx.TXID)] {
						//[]int64{0, 1} , j : 0, 1
						if int64(i) == j {
							//fmt.Printf("111111")
							//当前准备添加output已经消耗过了，不要再加了
							continue OUTPUT
						}
					}
				}

				//这个output和我们目标的地址相同，满足条件，加到返回UTXO数组中
				if bytes.Equal(output.PubKeyHash, senderPubKeyHash) {
					//fmt.Printf("222222")
					//UTXO = append(UTXO, output)

					//!!!!!重点
					//返回所有包含我的outx的交易的集合
					txs = append(txs, tx)

					//fmt.Printf("333333 : %f\n", UTXO[0].Value)
				} else {
					//fmt.Printf("333333")
				}
			}

			//如果当前交易是挖矿交易的话，那么不做遍历，直接跳过

			if !tx.IsCoinbase() {
				//4. 遍历input，找到自己花费过的utxo的集合(把自己消耗过的标示出来)
				for _, input := range tx.TXInputs {
					//判断一下当前这个input和目标（李四）是否一致，如果相同，说明这个是李四消耗过的output,就加进来
					//if input.Sig == address {
					//if input.PubKey == senderPubKeyHash  //这是肯定不对的，要做哈希处理
					pubKeyHash := HashPubKey(input.PubKey)
					if bytes.Equal(pubKeyHash, senderPubKeyHash) {
						//spentOutputs := make(map[string][]int64)
						//indexArray := spentOutputs[string(input.TXid)]
						//indexArray = append(indexArray, input.Index)
						spentOutputs[string(input.TXid)] = append(spentOutputs[string(input.TXid)], input.Index)
						//map[2222] = []int64{0}
						//map[3333] = []int64{0, 1}
					}
				}
			} else {
				//fmt.Printf("这是coinbase，不做input遍历！")
			}
		}

		if len(block.PrevHash) == 0 {
			break
			fmt.Printf("区块遍历完成退出!")
		}
	}

	return txs
}


//根据id查找交易本身，需要遍历整个区块链
func (bc *BlockChain) FindTransactionByTXid(id []byte) (Transaction, error) {

	//4. 如果没找到，返回空Transaction，同时返回错误状态

	it := bc.NewIterator()

	//1. 遍历区块链
	for {
		block := it.Next()
		//2. 遍历交易
		for _, tx := range block.Transactions {
			//3. 比较交易，找到了直接退出
			if bytes.Equal(tx.TXID, id) {
				return *tx, nil
			}
		}

		if len(block.PrevHash) == 0 {
			fmt.Printf("区块链遍历结束!\n")
			break
		}
	}

	return Transaction{}, errors.New("无效的交易id，请检查!")
}

func (bc *BlockChain) SignTransaction(tx *Transaction, privateKey *ecdsa.PrivateKey) {
	//签名，交易创建的最后进行签名
	prevTXs := make(map[string]Transaction)

	//找到所有引用的交易
	//1. 根据inputs来找，有多少input, 就遍历多少次
	//2. 找到目标交易，（根据TXid来找）
	//3. 添加到prevTXs里面
	for _, input := range tx.TXInputs {
		//根据id查找交易本身，需要遍历整个区块链
		tx, err := bc.FindTransactionByTXid(input.TXid)

		if err != nil {
			log.Panic(err)
		}

		prevTXs[string(input.TXid)] = tx
		//第一个input查找之后：prevTXs：
		// map[2222]Transaction222

		//第二个input查找之后：prevTXs：
		// map[2222]Transaction222
		// map[3333]Transaction333

		//第三个input查找之后：prevTXs：
		// map[2222]Transaction222
		// map[3333]Transaction333(只不过是重新写了一次)
	}

	tx.Sign(privateKey, prevTXs)
}

func (bc *BlockChain) VerifyTransaction(tx *Transaction) bool {

	if tx.IsCoinbase() {
		return true
	}
	//签名，交易创建的最后进行签名
	prevTXs := make(map[string]Transaction)

	//找到所有引用的交易
	//1. 根据inputs来找，有多少input, 就遍历多少次
	//2. 找到目标交易，（根据TXid来找）
	//3. 添加到prevTXs里面
	for _, input := range tx.TXInputs {
		//根据id查找交易本身，需要遍历整个区块链
		tx, err := bc.FindTransactionByTXid(input.TXid)

		if err != nil {
			log.Panic(err)
		}

		prevTXs[string(input.TXid)] = tx

	}

	return tx.Verify(prevTXs)
}
