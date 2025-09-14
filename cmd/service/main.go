package main

import (
	"flag"
	"log"

	_ "modernc.org/sqlite"
)

func main() {
	debug := flag.Bool("debug", false, "Set debug mode")

	flag.Parse()

	err := RunService("Timekeep", debug)
	if err != nil {
		log.Fatalln(err)
	}
}
