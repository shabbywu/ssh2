package main

import (
	"log"
	"ssh2/cmd"
	"os"
)


func main () {
	err := cmd.App.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}