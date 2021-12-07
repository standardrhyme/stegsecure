package steganalysis

import (
	// "bufio"
	"bytes"
	"fmt"
	"image"
	"image/png"
	"io/ioutil"
	"log"
	"os"
	"os/exec"

	"github.com/standardrhyme/stegsecure/pkg/interceptionfs"
)

func runPython() {
	fmt.Println("")
	cmd := exec.Command("python3", "./python-scripts/samplepairs.py", "testpng.png")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	log.Println(cmd.Run())
}

func AnalyzeCreate(b []byte) {
	err := ioutil.WriteFile("test", b, 0755)
	if err != nil {
		fmt.Printf("Unable to write file: %v", err)
	}

	// convert []byte to image for saving to file
	img, _, _ := image.Decode(bytes.NewReader(b))

	//save the imgByte to file
	out, err := os.Create("./testpng.png")

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	err = png.Encode(out, img)

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	//Run python script on file
	runPython()

}

func Analyze(n interceptionfs.Node) {
	fh, ok := n.(*interceptionfs.FileHandle)
	if !ok {
		// Not a file handle
		return
	}

	data, err := fh.InternalReadAll()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}

	AnalyzeCreate(data)
}
