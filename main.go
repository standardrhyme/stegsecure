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

func testInterception(path string) {
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
	err = fs.Mount(path)
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

	err = <-errchan
	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	if os.Geteuid() != 0 {
		log.Fatalln("Must be run as root!")
	}

	if (len(os.Args[1:]) < 1) {
		fmt.Println("Mounting to testdir/Downloads. If you want to set the folder to mount to, use: sudo go run main.go MOUNTPATH")
		testInterception("testdir/Downloads")
	} else {
		testInterception(os.Args[1])
	}
}
