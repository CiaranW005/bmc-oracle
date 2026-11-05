package main

import (
	"fmt"
	"log"
	"os"

	"golang.org/x/crypto/bcrypt"
	"golang.org/x/term"
)

func main() {
	fmt.Println("Enter BMC password: ")

	password, err := term.ReadPassword(int(os.Stdin.Fd()))
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println()

	// Generate the bcrypt hash to store in BMC_PASSWORD_HASH.
	// The server checks login passwords against this hash.
	hash, err := bcrypt.GenerateFromPassword(
		password,
		bcrypt.DefaultCost,
	)
	if err != nil {
		panic(err)
	}

	fmt.Println(string(hash))
}
