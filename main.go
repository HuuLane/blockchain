package main

import (
	"os"

	"github.com/HuuLane/stupidcoin/cli"
)

func main() {
	defer os.Exit(0)
	cli := cli.CommandLine{}
	cli.Run()
}
