package main

import (
	"image"
	"image/color"
	"image/png"
	"os"
)

type vec3 = [3]float64

const (
	WINDOW_WIDTH  = 512
	WINDOW_HEIGHT = 512
	MAP_WIDTH     = 16
	MAP_HEIGHT    = 16
)

func main() {
	framebuffer := make([]color.Color, WINDOW_WIDTH*WINDOW_HEIGHT)
	for j := range WINDOW_HEIGHT { // fill the screen with color gradients
		for i := range WINDOW_WIDTH {
			framebuffer[i+j*WINDOW_WIDTH] = toColor(vec3{float64(j) / float64(WINDOW_HEIGHT), float64(i) / float64(WINDOW_WIDTH), 0})
		}
	}
	writePng("out", framebuffer, WINDOW_WIDTH, WINDOW_HEIGHT)
}

func toColor(v vec3) color.RGBA {
	return color.RGBA{uint8(255 * v[0]), uint8(255 * v[1]), uint8(255 * v[2]), 0xff}
}

func writePng(name string, pixels []color.Color, WINDOW_WIDTH, WINDOW_HEIGHT int) {
	f, _ := os.Create(name + ".png")
	img := image.NewRGBA(image.Rect(0, 0, WINDOW_WIDTH, WINDOW_HEIGHT))
	for j := range WINDOW_HEIGHT {
		for i := range WINDOW_WIDTH {
			img.Set(i, j, pixels[i+j*WINDOW_WIDTH])
		}
	}
	png.Encode(f, img)
}
