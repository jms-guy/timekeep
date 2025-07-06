package main

import (
	"flag"
	"fmt"
	"log"
	"os"
)

func main() {
	var serviceFlag = flag.String("service", "", "Control the system service")
	flag.Parse()

	f, err := os.OpenFile("debug.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalln(fmt.Errorf("error opening file: %v", err))
	}
	defer f.Close()

	log.SetOutput(f)
	runService("ptracker", false)
}