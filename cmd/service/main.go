package main

import (
	"log"

	_ "modernc.org/sqlite"
)

func main() {
	err := RunService("Timekeep", false)
	if err != nil {
		log.Fatalln(err)
	}
}
