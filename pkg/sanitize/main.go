package main

import (
	"fmt"
	"image"
	"image/color"
	"image/jpeg"
	_ "image/jpeg"
	"image/png"
	"os"
	"reflect"
)

var sampleRGBA image.RGBA
var sampleYCbCr image.YCbCr

type CleanImg struct {
	image.Image
	custom map[image.Point]color.Color
}

func NewCleanImg(img image.Image) *CleanImg {
	return &CleanImg{img, map[image.Point]color.Color{}}
}

func (m *CleanImg) Set(x, y int, c color.Color) {
	m.custom[image.Point{X: x, Y: y}] = c
}

func (m *CleanImg) At(x, y int) color.Color {
	// Explicitly changed part: custom colors of the changed pixels:
	if c := m.custom[image.Point{X: x, Y: y}]; c != nil {
		return c
	}
	// Unchanged part: colors of the original image:
	return m.Image.At(x, y)
}

func Sanitize(path string) error {
	pic, err := os.Open(path)
	if err != nil {
		return err
	}
	old, format, err := image.Decode(pic)
	clean := NewCleanImg(old)
	bound := old.Bounds()
	var pixelFormat int
	switch reflect.TypeOf(old) {
	case reflect.TypeOf(&sampleRGBA):
		pixelFormat = 1
	case reflect.TypeOf(&sampleYCbCr):
		pixelFormat = 2
	default:
		fmt.Println("Unsupported Format")
		return nil
	}
	for y := bound.Min.Y; y < bound.Max.Y; y++ {
		go SanitizeRow(bound.Min.X, bound.Max.X, y, old, *clean, pixelFormat)
	}
	out, err := os.Create(path)
	if err != nil {
		return err
	}
	if format == "png" {
		err = png.Encode(out, clean)
	} else if format == "jpeg" {
		err = jpeg.Encode(out, clean, nil)
	} else {
		println("Unsupported format")
		return nil
	}
	if err != nil {
		return err
	}
	err = pic.Close()
	return err
}

func SanitizeRow(min int, max int, y int, old image.Image, clean CleanImg, pixelFormat int) {
	if pixelFormat == 1 {
		for x := min; x < max; x++ {
			SanitizePixelRGBA(x, y, old, clean)
		}
	}
	if pixelFormat == 2 {
		for x := min; x < max; x++ {
			SanitizePixelYCbCr(x, y, old, clean)
		}
	}
}

func SanitizePixelRGBA(x int, y int, old image.Image, clean CleanImg) {
	if reflect.TypeOf(old) == reflect.TypeOf(&sampleRGBA) {
		pixel := old.At(x, y)
		r, g, b, a := pixel.RGBA()
		clean.Set(x, y, color.RGBA{R: uint8(2 * (r / 2)), G: uint8(2 * (g / 2)), B: uint8(2 * (b / 2)), A: uint8(a)})
	} else if reflect.TypeOf(old) == reflect.TypeOf(&sampleYCbCr) {

	}
}

func SanitizePixelYCbCr(x int, y int, old image.Image, clean CleanImg) {
	pixel := old.At(x, y)
	r, g, b, _ := pixel.RGBA()
	r /= 256
	g /= 256
	b /= 256
	Y, Cb, Cr := color.RGBToYCbCr(uint8(2*(r/2)), uint8(2*(g/2)), uint8(2*(b/2)))
	clean.Set(x, y, color.YCbCr{Y: Y, Cb: Cb, Cr: Cr})
}

//func main() {
//	imageName := "C:\\Users\\Zhiyuan Huang\\Desktop\\final pro\\stegsecure\\pkg\\sanitize\\sample.jpg"
//	reader, err := os.Open(imageName)
//	if err != nil {
//		log.Fatal(err)
//	}
//	im, _, err := image.Decode(reader)
//	fmt.Println(im.At(9, 13))
//	defer reader.Close()
//	Sanitize(imageName)
//	reader2, err := os.Open(imageName)
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	im2, _, err := image.Decode(reader2)
//	fmt.Println(im2.At(9, 13))
//	reader2.Close()
//}
