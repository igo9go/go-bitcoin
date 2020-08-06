package main


import (
	"crypto/sha256"
	"fmt"
)

func main() {
	var data = "helloworld"
	//10s中左右
	for i := 0; i < 1000000; i++ {
		hash := sha256.Sum256([]byte(data + string(i)))
		fmt.Printf("hash : %x, %d\n", string(hash[:]), i)
	}
}