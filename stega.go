package main

import (
	"fmt"
	"image"
	"log"
	"math"
	"os"
)

// onlyLsb returns the LSB of x
func onlyLsb(x int) int {
	return x & 1
}

// exceptLsb returns all the bits other than the last one of x
func exceptLsb(x int) int {
	return x << 1
}

// updateParams updates the provided parameters, depending on the LSB and MSBs
func updateParams(u int, v int, params int[]) {
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

func analyzeSamplePairs(im image.Image, bounds image.Rectangle) (bool, float64) {
	// Based off of https://github.com/b3dk7/StegExpose/blob/master/SamplePairs.java
	avg := 0
	for color := 0; color < 3; color++ {

		//         W  X  Y  Z
		params := [4]int{0,0,0,0} // *** Go tuple issue ***
		P := 0 // *** Go tuple issue ***
		// pair := make([][]uint8{} // *** Go tuple issue ***
		pair := [][]int[] // *** Go tuple issue ***

		// Compute horizontal pairs
		for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
			for x := bounds.Min.X; x < bounds.Max.X - 1; x += 2 {
				pair := {x, y}, {x + 1, y} // *** Go tuple issue ***

				// Extract the color value for the pair pixels
				u := pair[0][color]
				v := pair[1][color]

				updateParams(u, v, params)

				P++
			}
		}

		// Compute vertical pairs
		for y := bounds.Min.Y; y < bounds.Max.Y - 1; y += 2 {
			for x := bounds.Min.X; x < bounds.Max.X; x++ {
				pair := {{x, y}, {x, y + 1}} // *** Go tuple issue ***

				// Extract the color value for the pair pixels
				u := pair[0][color]
				v := pair[1][color]

				updateParams(u, v, params)

				P++
			}
		}

		// W, X, Y, Z := params
		W := params[0]
		X := params[1]
		Y := params[2]
		Z := params[3]

		a := float64((W + Z) / 2.0)
		b := float64((2.0 * X) - P)
		c := float64(Y - X)

		if a == 0 {
			x := c / b
		}

		// Solve for the largest root
		discriminant := math.Pow(b,2.0) - (4.0*a*c)
		if discriminant >= 0 {
			posRoot := ((-1.0*b) + math.Pow(discriminant, 0.5)) / (2.0*a)
			negRoot := ((-1.0*b) - math.Pow(discriminant, 0.5)) / (2.0*a)

			if math.Abs(posRoot) <= math.Abs(negRoot) {
				x := posRoot
				else {
				x := negRoot
				}
			}
			else {
				x := c / b
			}
		}
		avg += x

	average := avg / 3
	probability := math.Min(math.Abs(float64(average)), 1)
	}

return probability > 0.5, probability
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
	}
	else {
		fmt.Println("This is probably not a stego image.")
	}
}

