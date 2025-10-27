package main

import (
	"flag"
	"log"

	_ "modernc.org/sqlite"
)

// Service entry point
func main() {
	debug := flag.Bool("debug", false, "Set debug mode")

	flag.Parse()

	// OS specific RunService function
	err := RunService("Timekeep", debug)
	if err != nil {
		log.Fatalln(err)
	}
}
