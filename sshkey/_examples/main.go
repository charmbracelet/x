package main

import (
	"fmt"

	"github.com/charmbracelet/x/sshkey"
)

func main() {
	// password is "asd".
	signer, err := sshkey.Open("./key")
	if err != nil {
		panic(err)
	}

	if signer != nil {
		fmt.Println("Key opened!")
	}
}
