package main

import (
	"fmt"
	"log"
	"os"

	"github.com/standardrhyme/stegsecure/pkg/interceptionfs"
	"github.com/standardrhyme/stegsecure/pkg/steganalysis"
)

var (
	DEBUG = false
)

func testInterception() {
	// Initialize the filesystem
	fs, err := interceptionfs.Init(steganalysis.AnalyzeGo)
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
