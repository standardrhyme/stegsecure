package main

import (
	"fmt"
	"log"

	"github.com/standardrhyme/stegsecure/pkg/interceptionfs"
)

func testInterception() {
	fs, err := interceptionfs.Init(func(n *interceptionfs.Node) {
		fmt.Println("file was modified:", n)
	})
	if err != nil {
		log.Fatal(err)
	}

	fs.Debug = func(msg interface{}) {
		fmt.Println("[DEBUG]", msg)
	}

	err = fs.Mount("mnt")
	if err != nil {
		log.Fatal(err)
	}

	errchan := make(chan error)
	err = fs.Serve(errchan)
	if err != nil {
		log.Fatal(err)
	}

	defer fs.Close()

	fmt.Println(fs)
	err = <-errchan
	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	testInterception()
}
