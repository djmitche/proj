package main

import (
	"github.com/djmitche/proj/proj"
	"log"
)

func main() {
	log.SetFlags(0)
	log.SetPrefix("proj: ")

	err := proj.Main()
	if err != nil {
		log.Println(err)
	}
}
