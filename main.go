package main

import (
	"fmt"
	"github.com/djmitche/proj/internal"
	"log"
	"os"
)

func main() {
	log.SetFlags(0)
	log.SetPrefix("proj: ")

	err := proj.Main()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %s\n", err)
		os.Exit(1)
	}
}
