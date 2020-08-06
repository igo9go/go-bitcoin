package main


func main() {
	bc := NewBlockChain("1PotVgZsgUM3kPZ4iABBbN2253AeKm3Uc9")
	cli := CLI{bc}
	cli.Run()
}
