package main

import (
	"image"
	"image/color"
	"image/png"
	"math"
	"os"
)

type vec3 = [3]float64

const (
	WINDOW_WIDTH  = 512
	WINDOW_HEIGHT = 512
	MAP_WIDTH     = 16
	MAP_HEIGHT    = 16
)

var worldMap = []byte{
	'0', '0', '0', '0', '2', '2', '2', '2', '2', '2', '2', '2', '0', '0', '0', '0',
	'1', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', '0',
	'1', ' ', ' ', ' ', ' ', ' ', ' ', '1', '1', '1', '1', '1', ' ', ' ', ' ', '0',
	'1', ' ', ' ', ' ', ' ', ' ', '0', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', '0',
	'0', ' ', ' ', ' ', ' ', ' ', '0', ' ', ' ', '1', '1', '1', '0', '0', '0', '0',
	'0', ' ', ' ', ' ', ' ', ' ', '3', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', '0',
	'0', ' ', ' ', ' ', '1', '0', '0', '0', '0', ' ', ' ', ' ', ' ', ' ', ' ', '0',
	'0', ' ', ' ', ' ', '0', ' ', ' ', ' ', '1', '1', '1', '0', '0', ' ', ' ', '0',
	'0', ' ', ' ', ' ', '0', ' ', ' ', ' ', '0', ' ', ' ', ' ', ' ', ' ', ' ', '0',
	'0', ' ', ' ', ' ', '0', ' ', ' ', ' ', '1', ' ', ' ', '0', '0', '0', '0', '0',
	'0', ' ', ' ', ' ', ' ', ' ', ' ', ' ', '1', ' ', ' ', ' ', ' ', ' ', ' ', '0',
	'2', ' ', ' ', ' ', ' ', ' ', ' ', ' ', '1', ' ', ' ', ' ', ' ', ' ', ' ', '0',
	'0', ' ', ' ', ' ', ' ', ' ', ' ', ' ', '0', ' ', ' ', ' ', ' ', ' ', ' ', '0',
	'0', ' ', '0', '0', '0', '0', '0', '0', '0', ' ', ' ', ' ', ' ', ' ', ' ', '0',
	'0', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', '0',
	'0', '0', '0', '2', '2', '2', '2', '2', '2', '2', '2', '0', '0', '0', '0', '0',
}

func drawTrangle(img []color.Color, imgW, x, y, w, h int, c color.Color) {
	for i := range w {
		for j := range h {
			cx := x + i
			cy := y + j
			img[cx+cy*imgW] = c
		}
	}
}

func main() {
	playerX := 3.456 // player x position
	playerY := 2.345 // player y position
	playerA := 1.523 // player view direction

	framebuffer := make([]color.Color, WINDOW_WIDTH*WINDOW_HEIGHT)
	for j := range WINDOW_HEIGHT { // fill the screen with color gradients
		for i := range WINDOW_WIDTH {
			framebuffer[i+j*WINDOW_WIDTH] = toColor(vec3{float64(j) / float64(WINDOW_HEIGHT), float64(i) / float64(WINDOW_WIDTH), 0})
		}
	}

	rectW := WINDOW_WIDTH / MAP_WIDTH
	rectH := WINDOW_HEIGHT / MAP_HEIGHT
	for j := range MAP_HEIGHT { // draw the map
		for i := range MAP_WIDTH {
			if worldMap[i+j*MAP_WIDTH] == ' ' { // skip empty spaces
				continue
			}
			rectX := i * rectW
			rectY := j * rectH
			drawTrangle(framebuffer, WINDOW_WIDTH, rectX, rectY, rectW, rectH, color.RGBA{0, 255, 255, 255})
		}
	}

	// draw the player on the map
	drawTrangle(framebuffer, WINDOW_WIDTH, int(playerX*float64(rectW)), int(playerY*float64(rectH)), 5, 5, color.RGBA{255, 255, 255, 255})

	for t := 0.0; t < 20; t += 0.05 {
		cx := playerX + t*math.Cos(playerA)
		cy := playerY + t*math.Sin(playerA)
		if worldMap[int(cx)+int(cy)*MAP_WIDTH] != ' ' {
			break
		}

		pixX := int(cx * float64(rectW))
		pixY := int(cy * float64(rectH))
		framebuffer[pixX+pixY*WINDOW_WIDTH] = color.RGBA{255, 255, 255, 255}
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
