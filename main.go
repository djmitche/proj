package main

import (
	"github.com/djmitche/proj/proj"
	"log"
)

func main() {
	err := proj.Main()
	if err != nil {
		fmt.Println(err)
	}
}
