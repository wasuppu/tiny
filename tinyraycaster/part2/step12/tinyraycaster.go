package main

import (
	"fmt"
	"image/color"
	"log"
	"math"
	"os"
	"path/filepath"
	"runtime"
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

func wallTexcoordX(cx, cy float64, texwalls *Texture) int {
	hitx := cx - math.Floor(cx+0.5) // hitx and hity contain (signed) fractional parts of cx and cy,
	hity := cy - math.Floor(cy+0.5) // they vary between -0.5 and +0.5, and one of them is supposed to be very close to 0
	texcoordX := int(hitx * float64(texwalls.size))
	if math.Abs(hity) > math.Abs(hitx) { // we need to determine whether we hit a "vertical" or a "horizontal" wall (w.r.t the map)
		texcoordX = int(hity * float64(texwalls.size))
	}
	if texcoordX < 0 {
		texcoordX += texwalls.size // do not forget x_texcoord can be negative, fix that
	}
	return texcoordX
}

func render(fb *Framebuffer, m *Map, player *Player, texwalls *Texture) {
	fb.Clear(color.RGBA{255, 255, 255, 255}) // clear the screen

	rectW := fb.w / (m.w * 2) // size of one map cell on the screen
	rectH := fb.h / m.h
	for j := range m.h { // draw the map
		for i := range m.w {
			if m.IsEmpty(i, j) { // skip empty spaces
				continue
			}
			rectX := i * rectW
			rectY := j * rectH
			texid := m.Get(i, j)
			fb.DrawRectangle(rectX, rectY, rectW, rectH, texwalls.Get(0, 0, texid)) //  the color is taken from the upper left pixel of the texture #texid
		}
	}

	for i := range fb.w / 2 { // draw the visibility cone AND the "3D" view
		angle := player.a - player.fov/2 + player.fov*float64(i)/(float64(fb.w)/2)
		for t := 0.0; t < 20; t += .01 { // ray marching loop
			x := player.x + t*math.Cos(angle)
			y := player.y + t*math.Sin(angle)
			fb.SetPixel(int(x*float64(rectW)), int(y*float64(rectH)), color.RGBA{160, 160, 160, 255}) // this draws the visibility cone

			if m.IsEmpty(int(x), int(y)) {
				continue
			}

			texid := m.Get(int(x), int(y)) // our ray touches a wall, so draw the vertical column to create an illusion of 3D
			columnHeight := int(float64(fb.h) / (t * math.Cos(angle-player.a)))
			texcoordX := wallTexcoordX(x, y, texwalls)
			column := texwalls.TextureColumn(texid, texcoordX, columnHeight)
			pixX := fb.w/2 + i            // we are drawing at the right half of the screen, thus +fb.w/2
			for j := range columnHeight { // copy the texture column to the framebuffer
				pixY := j + fb.h/2 - columnHeight/2
				if pixY < 0 || pixY >= fb.h {
					continue
				}
				fb.SetPixel(pixX, pixY, column[j])
			}
			break
		} // ray marching loop
	} // field of view ray sweeping
}

func main() {
	fb := NewFramebuffer(1024, 512, color.RGBA{255, 255, 255, 255})
	player := Player{3.456, 2.345, 1.523, math.Pi / 3}
	m := NewMap()
	texwalls := NewTexture("walltext.png")
	if texwalls.count == 0 {
		log.Fatalln("Failed to load wall textures")
	}

	outdir := filepath.Join(rootpath, "out")
	if err := os.MkdirAll(outdir, 0755); err != nil {
		log.Fatalln("Failed to create directories:", err)
	}
	for frame := range 360 {
		outname := fmt.Sprintf("%0.5d", frame)
		player.a += 2 * math.Pi / 360
		render(fb, m, &player, texwalls)

		writePng(filepath.Join(outdir, outname), fb.ps, fb.w, fb.h)
	}
}
