package main

import (
	"log"
	"os"
)

func main() {
	oldLocation := "~/Desktop/test.png"
	newLocation := "~/Desktop/USMC/test.png"
	err := os.Rename(oldLocation, newLocation)
	if err != nil {
		log.Fatal(err)
	}
}
