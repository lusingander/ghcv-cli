package main

import (
	"log"
	"os"
)

func run(args []string) error {
	return nil
}

func main() {
	if err := run(os.Args); err != nil {
		log.Fatal(err)
	}
}
