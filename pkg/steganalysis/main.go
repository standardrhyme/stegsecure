package steganalysis

import (
	"bytes"
	"fmt"
	"image"
	"image/png"
	"os"
	"os/exec"
	"strings"

	"github.com/standardrhyme/stegsecure/pkg/interceptionfs"
)

// Returns whether the file has steganographic content.
func runPython(path string) bool {
	fmt.Println("")
	out, err := exec.Command("python3", "./python-scripts/samplepairs.py", path).Output()
	if err != nil {
		return false
	}

	parts := strings.Split(strings.TrimSpace(string(out)), "\n")
	output := parts[len(parts) - 1]
	fmt.Print(string(out))
	return output == "TRUE"
}

// Returns whether the file has steganographic content.
func AnalyzeCreate(b []byte) bool {
	// convert []byte to image for saving to file
	img, _, _ := image.Decode(bytes.NewReader(b))

	//save the imgByte to file
	out, err := os.CreateTemp("", "*")
	defer os.Remove(out.Name())
	defer out.Close()

	if err != nil {
		fmt.Println(err)
		return false
	}

	err = png.Encode(out, img)

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	fmt.Println(out.Name())

	//Run python script on file
	return runPython(out.Name())
}

func Analyze(n interceptionfs.Node) {
	fh, ok := n.(*interceptionfs.FileHandle)
	if !ok {
		// Not a file handle
		return
	}

	fmt.Println()
	fmt.Println("===========")
	fmt.Println("FILE NAME: ", fh.Name())

	if !strings.HasSuffix(fh.Name(), ".png") {
		fh.File.Release()
		return
	}

	data, err := fh.InternalReadAll()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}

	if AnalyzeCreate(data) {
		fmt.Println("SANITIZE")
	}

	fh.File.Release()
}
