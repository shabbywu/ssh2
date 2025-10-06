package main

import (
	"log"
	"os"
	"ssh2/cmd"
)

func main() {
	err := cmd.App.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
