package main

import (
	"bytes"
	"encoding/gob"
	"fmt"
)

type Person struct {
	Name string
	Age  int
}

func main() {
	var buffer bytes.Buffer

	encoder := gob.NewEncoder(&buffer)

	lily := Person{"lily", 28}

	err := encoder.Encode(lily)
	if err != nil {
		panic(err)
	}
	fmt.Println("after serialize :", buffer.Bytes())

	var LILY Person

	decoder := gob.NewDecoder(&buffer)
	err = decoder.Decode(&LILY)
	if err != nil {
		fmt.Println("decode failed!", err)
	}

	fmt.Println(LILY)

}
