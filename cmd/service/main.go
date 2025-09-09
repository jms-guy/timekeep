package main

import (
	"log"
)

func main() {
	err := RunService("Timekeep", false)
	if err != nil {
		log.Fatalln(err)
	}
}
