package main

import (
	"fmt"
	"log"
	"os"

	"github.com/standardrhyme/stegsecure/pkg/interceptionfs"
)

var (
	DEBUG = true
)

func testInterception() {
	// Initialize the filesystem
	fs, err := interceptionfs.Init(func(n interceptionfs.Node) {
		node, err := n.GetNode()
		if err != nil {
			return
		}

		fmt.Println("file was modified:", n.Name(), node)

		fh, ok := n.(*interceptionfs.FileHandle)
		if !ok {
			// Not a file handle
			return
		}

		fmt.Println(fh.File)
	})
	if err != nil {
		log.Fatal(err)
	}

	if DEBUG {
		fs.Debug = func(msg interface{}) {
			fmt.Println("[DEBUG]", msg)
		}
	}

	// Mount the filesystem to the Downloads folder
	err = fs.Mount("testdir/Downloads")
	if err != nil {
		log.Fatal(err)
	}

	// Serve the filesystem (attach the mount to the FS struct) and watch for errors.
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
	if os.Geteuid() != 0 {
		log.Fatalln("Must be run as root!")
	}

	testInterception()
}
