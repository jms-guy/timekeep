package main

import (
	"fmt"
	"log"
	"os"
)

func main() {
	f, err := os.OpenFile("debug.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalln(fmt.Errorf("error opening file: %v", err))
	}
	defer f.Close()

	log.SetOutput(f)
	err = RunService("Timekeep", false)
	if err != nil {
		log.Fatalln(err)
	}
}
