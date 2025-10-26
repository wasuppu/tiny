package main

import (
	"image"
	"image/color"
	"image/png"
	"log"
	"math"
	"os"
	"path/filepath"
	"runtime"
)

const (
	WINDOW_WIDTH  = 1024
	WINDOW_HEIGHT = 512
	MAP_WIDTH     = 16
	MAP_HEIGHT    = 16
)

var (
	basepath string
	rootpath string
)

func init() {
	_, exepath, _, _ := runtime.Caller(0)
	basepath = filepath.Dir(exepath)
	rootpath = filepath.Dir(filepath.Dir(basepath))
}

var worldMap = []byte{
	'0', '0', '0', '0', '2', '2', '2', '2', '2', '2', '2', '2', '0', '0', '0', '0',
	'1', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', '0',
	'1', ' ', ' ', ' ', ' ', ' ', ' ', '1', '1', '1', '1', '1', ' ', ' ', ' ', '0',
	'1', ' ', ' ', ' ', ' ', ' ', '0', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', '0',
	'0', ' ', ' ', ' ', ' ', ' ', '0', ' ', ' ', '1', '1', '1', '0', '0', '0', '0',
	'0', ' ', ' ', ' ', ' ', ' ', '3', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', '0',
	'0', ' ', ' ', ' ', '1', '0', '0', '0', '0', ' ', ' ', ' ', ' ', ' ', ' ', '0',
	'0', ' ', ' ', ' ', '3', ' ', ' ', ' ', '1', '1', '1', '0', '0', ' ', ' ', '0',
	'5', ' ', ' ', ' ', '4', ' ', ' ', ' ', '0', ' ', ' ', ' ', ' ', ' ', ' ', '0',
	'5', ' ', ' ', ' ', '4', ' ', ' ', ' ', '1', ' ', ' ', '0', '0', '0', '0', '0',
	'0', ' ', ' ', ' ', ' ', ' ', ' ', ' ', '1', ' ', ' ', ' ', ' ', ' ', ' ', '0',
	'2', ' ', ' ', ' ', ' ', ' ', ' ', ' ', '1', ' ', ' ', ' ', ' ', ' ', ' ', '0',
	'0', ' ', ' ', ' ', ' ', ' ', ' ', ' ', '0', ' ', ' ', ' ', ' ', ' ', ' ', '0',
	'0', ' ', '0', '0', '0', '0', '0', '0', '0', ' ', ' ', ' ', ' ', ' ', ' ', '0',
	'0', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', '0',
	'0', '0', '0', '2', '2', '2', '2', '2', '2', '2', '2', '0', '0', '0', '0', '0',
}

func loadTexture(filename string) ([]color.Color, int, int, error) {
	texturepath := filepath.Join(rootpath, "textures", filename)
	img, err := openImg(texturepath)
	if err != nil {
		return nil, -1, -1, err
	}
	width, height := img.Bounds().Dx(), img.Bounds().Dy()
	return readColorFromImg(img), width, height, nil
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

	walltext, walltextWidth, walltextHeight, err := loadTexture("walltext.png")
	if err != nil {
		log.Fatalf("%+v", err)
	}
	walltextCnt := walltextWidth / walltextHeight
	walltextSize := walltextWidth / walltextCnt

	rectW := WINDOW_WIDTH / (MAP_WIDTH * 2)
	rectH := WINDOW_HEIGHT / MAP_HEIGHT

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
			texid := int(worldMap[i+j*MAP_WIDTH] - '0')
			assert(texid < walltextCnt, "texid < walltextCnt")
			drawTrangle(framebuffer, WINDOW_WIDTH, WINDOW_HEIGHT, rectX, rectY, rectW, rectH, walltext[texid*walltextSize]) // the color is taken from the upper left pixel of the texture #texid
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
				texid := int(worldMap[int(cx)+int(cy)*MAP_WIDTH] - '0')
				assert(texid < walltextCnt, "texid < walltextCnt")
				columnHeight := int(WINDOW_HEIGHT / (t * math.Cos(angle-playerA)))
				drawTrangle(framebuffer, WINDOW_WIDTH, WINDOW_HEIGHT, WINDOW_WIDTH/2+i, WINDOW_HEIGHT/2-columnHeight/2, 1, columnHeight, walltext[texid*walltextSize])
				break
			}
		}
	}

	writePng("out", framebuffer, WINDOW_WIDTH, WINDOW_HEIGHT)
}

func assert(cond bool, msg string) {
	if !cond {
		panic(msg)
	}
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

func openImg(filename string) (image.Image, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	img, _, err := image.Decode(file)
	if err != nil {
		return nil, err
	}

	return img, nil
}

func readColorFromImg(img image.Image) []color.Color {
	var pixels []color.Color

	bounds := img.Bounds()
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			pixels = append(pixels, img.At(x, y))
		}
	}

	return pixels
}
