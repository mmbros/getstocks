package main

import (
	"os"

	"github.com/mmbros/getstocks/cli"
)

func main() {
	os.Exit(cli.Run())
}
