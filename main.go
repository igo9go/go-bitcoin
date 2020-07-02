package main


func main() {
	bc := NewBlockChain()
	cli := CLI{bc}
	cli.Run()
}
