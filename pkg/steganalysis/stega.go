package steganalysis

import (
	"bytes"
	"fmt"
	"image"
	imcolor "image/color"
	"log"
	"math"
	"os"
	"strings"

	"github.com/standardrhyme/stegsecure/pkg/interceptionfs"
	"github.com/standardrhyme/stegsecure/pkg/sanitize"
)

// onlyLsb returns the LSB of x
func onlyLsb(x uint32) uint32 {
	return x & 1
}

// exceptLsb returns all the bits other than the last one of x
func exceptLsb(x uint32) uint32 {
	return x << 1
}

// updateParams updates the provided parameters, depending on the LSB and MSBs
func updateParams(u uint32, v uint32, params *[4]float64) {
	uMsb := exceptLsb(u)
	uLsb := onlyLsb(u)
	vMsb := exceptLsb(v)
	vLsb := onlyLsb(v)

	// if only the LSB are different
	if (uMsb == vMsb) && (uLsb != vMsb) {
		params[0]++
	}

	// if they are the same
	if u == v {
		params[3]++
	}

	if ((vLsb == 0) && (u < v)) || ((vLsb == 1) && (u > v)) {
		params[1]++
	}

	if ((vLsb == 0) && (u > v)) || ((vLsb == 1) && (u < v)) {
		params[2]++
	}
}

func getColorComponent(color imcolor.Color, index int) uint32 {
	r, g, b, a := color.RGBA()
	switch index {
	case 0:
		return r
	case 1:
		return g
	case 2:
		return b
	default:
		return a
	}
}

func analyzeSamplePairs(im image.Image, bounds image.Rectangle) (bool, float64) {
	// Based off of https://github.com/b3dk7/StegExpose/blob/master/SamplePairs.java
	avg := float64(0)
	for color := 0; color < 3; color++ {
		//                   W X Y Z
		params := &[4]float64{0,0,0,0}
		P := float64(0)

		// Compute horizontal pairs
		for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
			for x := bounds.Min.X; x < bounds.Max.X - 1; x += 2 {
				// Extract the color value for the pair pixels
				u := getColorComponent(im.At(x, y), color)
				v := getColorComponent(im.At(x+1, y), color)

				updateParams(u, v, params)

				P++
			}
		}

		// Compute vertical pairs
		for y := bounds.Min.Y; y < bounds.Max.Y - 1; y += 2 {
			for x := bounds.Min.X; x < bounds.Max.X; x++ {
				// Extract the color value for the pair pixels
				u := getColorComponent(im.At(x, y), color)
				v := getColorComponent(im.At(x, y+1), color)

				updateParams(u, v, params)

				P++
			}
		}

		// W, X, Y, Z := params
		W := params[0]
		X := params[1]
		Y := params[2]
		Z := params[3]

		a := (W+Z) / 2
		b := (2*X) - P
		c := Y-X

		var x float64

		if a == 0 {
			x = c / b
		}

		// Solve for the largest root
		discriminant := math.Pow(b, 2) - (4*a*c)
		if discriminant >= 0 {
			posRoot := ((-1*b) + math.Sqrt(discriminant)) / (2.0*a)
			negRoot := ((-1*b) - math.Sqrt(discriminant)) / (2.0*a)

			if math.Abs(posRoot) <= math.Abs(negRoot) {
				x = posRoot
			} else {
				x = negRoot
			}
		} else {
			x = c / b
		}

		avg += x
	}

	average := avg / 3
	probability := math.Min(math.Abs(float64(average)), 1)

	return probability > 0.5, probability
}

func analyzeBytes(b []byte) bool {
	r := bytes.NewReader(b)
	im, _, err := image.Decode(r)
	if err != nil {
		fmt.Println(err)
		return false
	}
	bounds := im.Bounds()

	result, _ := analyzeSamplePairs(im, bounds)
	return result
}

func AnalyzeGo(n interceptionfs.Node) {
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

	if analyzeBytes(data) {
		fmt.Println("SANITIZE")
		cleaned := sanitize.SanitizeBytes(data, "png")
		fh.InternalOverwrite(cleaned)
	}

	fh.File.Release()
}

func main() {
	// Ask the user what image they would like to analyze
	fmt.Println("Enter the name of the image you would like to analyze: ")
	imageName := ""
	_, err := fmt.Scanf("%s", imageName)
	if err != nil {
		fmt.Println("Please input a valid directory / file name")
		log.Fatal(err)
	}

	// Open the image
	reader, err := os.Open(imageName)
	if err != nil {
		log.Fatal(err)
	}
	defer reader.Close()

	im, _, err := image.Decode(reader)
	if err != nil {
		log.Fatal(err)
	}
	bounds := im.Bounds()

	result, probability := analyzeSamplePairs(im, bounds)

	println("Probability of being a stego image:", probability)
	if result {
		fmt.Println("This is probably a stego image.")
	} else {
		fmt.Println("This is probably not a stego image.")
	}
}

