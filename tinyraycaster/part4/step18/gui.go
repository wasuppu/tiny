package main

import (
	"fmt"
	"image/color"
	"log"
	"math"
	"unsafe"

	"github.com/veandco/go-sdl2/sdl"
)

func main() {
	fb := NewFramebuffer(1024, 512, color.RGBA{255, 255, 255, 255})
	player := Player{3.456, 2.345, 1.523, math.Pi / 3}
	m := NewMap()
	texwalls := NewTexture("walltext.png")
	texmonst := NewTexture("monsters.png")
	if texwalls.count == 0 || texmonst.count == 0 {
		log.Fatalln("Failed to load textures")
	}
	sprites := []Sprite{{3.523, 3.812, 2, 0}, {1.834, 8.765, 0, 0}, {5.323, 5.365, 1, 0}, {4.123, 10.265, 1, 0}}

	render(fb, m, &player, sprites, texwalls, texmonst)

	if err := sdl.Init(sdl.INIT_VIDEO); err != nil {
		log.Fatalf("%+v\n", fmt.Errorf("couldn't initialize SDL: %s", err))
	}

	window, renderer, err := sdl.CreateWindowAndRenderer(int32(fb.w), int32(fb.h), sdl.WINDOW_SHOWN|sdl.WINDOW_INPUT_FOCUS)
	if err != nil {
		log.Fatalf("%+v\n", fmt.Errorf("couldn't create window and renderer: %s", err))
	}

	texture, err := renderer.CreateTexture(sdl.PIXELFORMAT_ABGR8888, sdl.TEXTUREACCESS_STREAMING, int32(fb.w), int32(fb.h))
	if err != nil {
		log.Fatalf("%+v\n", fmt.Errorf("couldn't create texture: %s", err))
	}

	data := colorsToPixels(fb.ps)
	texture.Update(nil, unsafe.Pointer(&data[0]), fb.w*4)
	running := true
	for running {
		for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
			switch event.(type) {
			case *sdl.QuitEvent:
				running = false
			}
		}

		renderer.Clear()
		renderer.Copy(texture, nil, nil)
		renderer.Present()
	}

	texture.Destroy()
	renderer.Destroy()
	window.Destroy()
	sdl.Quit()
}

func colorsToPixels(colors []color.Color) []byte {
	pixels := make([]byte, len(colors)*4)
	for i, c := range colors {
		r, g, b, a := c.RGBA()
		pixels[i*4] = byte(r >> 8)
		pixels[i*4+1] = byte(g >> 8)
		pixels[i*4+2] = byte(b >> 8)
		pixels[i*4+3] = byte(a >> 8)
	}
	return pixels
}
