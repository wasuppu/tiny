package main

import (
	"fmt"
	"image"
	"image/color"
	"image/png"
	"log"
	"math"
	"math/rand"
	"os"
	"path/filepath"
	"runtime"
)

type vec3 = [3]float64

var (
	basepath string
	rootpath string
)

func init() {
	_, exepath, _, _ := runtime.Caller(0)
	basepath = filepath.Dir(exepath)
	rootpath = filepath.Dir(filepath.Dir(basepath))
}

const (
	WINDOW_WIDTH  = 1024
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

func drawTrangle(img []color.Color, imgW, imgH, x, y, w, h int, c color.Color) {
	for i := range w {
		for j := range h {
			cx := x + i
			cy := y + j
			if cx >= imgW || cy >= imgH { // no need to check negative values (unsigned variables)
				continue
			}
			img[cx+cy*imgW] = c
		}
	}
}

func main() {
	playerX := 3.456   // player x position
	playerY := 2.345   // player y position
	playerA := 1.523   // player view direction
	fov := math.Pi / 3 // field of view

	framebuffer := make([]color.Color, WINDOW_WIDTH*WINDOW_HEIGHT)

	ncolors := 10
	colors := make([]color.Color, ncolors)
	for i := range ncolors {
		colors[i] = toColor(vec3{float64(rand.Int() % 255), float64(rand.Int() % 255), float64(rand.Int() % 255)})
	}

	rectW := WINDOW_WIDTH / (MAP_WIDTH * 2)
	rectH := WINDOW_HEIGHT / MAP_HEIGHT

	outdir := filepath.Join(rootpath, "out")
	if err := os.MkdirAll(outdir, 0755); err != nil {
		log.Fatalln("Failed to create directories:", err)
	}
	for frame := range 360 {
		outname := fmt.Sprintf("%0.5d", frame)
		playerA += 2 * math.Pi / 360

		for i := range framebuffer { // // the image itself, initialized to white
			framebuffer[i] = color.RGBA{255, 255, 255, 255}
		}

		for j := range MAP_HEIGHT { // draw the map
			for i := range MAP_WIDTH {
				if worldMap[i+j*MAP_WIDTH] == ' ' { // skip empty spaces
					continue
				}
				rectX := i * rectW
				rectY := j * rectH
				icolor := int(worldMap[i+j*MAP_WIDTH] - '0')
				assert(icolor < ncolors, "icolor < ncolors")
				drawTrangle(framebuffer, WINDOW_WIDTH, WINDOW_HEIGHT, rectX, rectY, rectW, rectH, colors[icolor])
			}
		}

		for i := range WINDOW_WIDTH / 2 { // draw the visibility cone AND the "3D" view
			angle := playerA - fov/2 + fov*float64(i)/float64(WINDOW_WIDTH/2)

			for t := 0.0; t < 20; t += 0.01 {
				cx := playerX + t*math.Cos(angle)
				cy := playerY + t*math.Sin(angle)

				pixX := int(cx * float64(rectW))
				pixY := int(cy * float64(rectH))
				framebuffer[pixX+pixY*WINDOW_WIDTH] = color.RGBA{160, 160, 160, 255} // this draws the visibility cone

				if worldMap[int(cx)+int(cy)*MAP_WIDTH] != ' ' { // our ray touches a wall, so draw the vertical column to create an illusion of 3D
					icolor := int(worldMap[int(cx)+int(cy)*MAP_WIDTH] - '0')
					assert(icolor < ncolors, "icolor < ncolors")
					columnHeight := int(WINDOW_HEIGHT / (t * math.Cos(angle-playerA)))
					drawTrangle(framebuffer, WINDOW_WIDTH, WINDOW_HEIGHT, WINDOW_WIDTH/2+i, WINDOW_HEIGHT/2-columnHeight/2, 1, columnHeight, colors[icolor])
					break
				}
			}
		}

		writePng(filepath.Join(outdir, outname), framebuffer, WINDOW_WIDTH, WINDOW_HEIGHT)
	}
}

func assert(cond bool, msg string) {
	if !cond {
		panic(msg)
	}
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
