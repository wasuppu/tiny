package main

import (
	"fmt"
	"image/color"
	"log"
	"math"
	"sort"
	"time"
	"unsafe"

	"github.com/veandco/go-sdl2/sdl"
)

func main() {
	fb := NewFramebuffer(1024, 512, color.RGBA{255, 255, 255, 255})
	gs := GameState{
		NewMap(),
		&Player{3.456, 2.345, 1.523, math.Pi / 3, 0, 0}, // player
		[]Sprite{{3.523, 3.812, 2, 0}, // monsters lists
			{1.834, 8.765, 0, 0},
			{5.323, 5.365, 1, 0},
			{14.32, 13.36, 3, 0},
			{4.123, 10.76, 1, 0}},
		NewTexture("walltext.png"),
		NewTexture("monsters.png"),
	}

	if gs.texwalls.count == 0 || gs.texmonst.count == 0 {
		log.Fatalln("Failed to load textures")
	}

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

	t1 := time.Now()
	running := true
	for running {
		{
			t2 := time.Now()
			dura := t2.Sub(t1)
			if dura.Milliseconds() < 20 {
				time.Sleep(3 * time.Nanosecond)
				continue
			}
			t1 = t2
		}

		// poll events and update player's state (walk/turn flags);
		{
			for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
				switch t := event.(type) {
				case *sdl.QuitEvent:
					running = false
				case *sdl.KeyboardEvent:
					switch t.Type {
					case sdl.KEYDOWN:
						switch t.Keysym.Sym {
						case sdl.K_ESCAPE:
							running = false
						case sdl.K_a:
							gs.player.turn = -1
						case sdl.K_d:
							gs.player.turn = 1
						case sdl.K_w:
							gs.player.walk = 1
						case sdl.K_s:
							gs.player.walk = -1
						}

					case sdl.KEYUP:
						switch t.Keysym.Sym {
						case sdl.K_a, sdl.K_d:
							gs.player.turn = 0
						case sdl.K_w, sdl.K_s:
							gs.player.walk = 0
						}
					}
				}
			}
		}

		// update player's position
		{
			gs.player.a += float64(gs.player.turn) * 0.05
			nx := gs.player.x + float64(gs.player.walk)*math.Cos(gs.player.a)*0.05
			ny := gs.player.y + float64(gs.player.walk)*math.Sin(gs.player.a)*0.05

			if int(nx) >= 0 && int(nx) < gs.m.w && int(ny) >= 0 && int(ny) < gs.m.h {
				if gs.m.IsEmpty(int(nx), int(gs.player.y)) {
					gs.player.x = nx
				}

				if gs.m.IsEmpty(int(gs.player.x), int(ny)) {
					gs.player.y = ny
				}
			}

			// update the distances from the player to each sprite
			for i := range gs.monsters {
				gs.monsters[i].playerDist = math.Sqrt(math.Pow(gs.player.x-gs.monsters[i].x, 2) + math.Pow(gs.player.y-gs.monsters[i].y, 2))
			}
			sort.Slice(gs.monsters, func(i, j int) bool {
				return gs.monsters[i].playerDist > gs.monsters[j].playerDist // sort it from farthest to closest
			})
		}

		render(fb, &gs) // update the distances from the player to each sprite

		// copy the framebuffer contents to the screen
		{
			data := colorsToPixels(fb.ps)
			texture.Update(nil, unsafe.Pointer(&data[0]), fb.w*4)
			renderer.Clear()
			renderer.Copy(texture, nil, nil)
			renderer.Present()
		}
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
