package main

import (
	"image/color"
	"log"
	"time"
	"unsafe"

	"github.com/veandco/go-sdl2/sdl"
)

var (
	width  int32 = 800
	height int32 = 600
)

type Vec4f [4]float64

func (v Vec4f) toColor() color.RGBA {
	r := uint8(max(0, min(255, v[0]*255)))
	g := uint8(max(0, min(255, v[1]*255)))
	b := uint8(max(0, min(255, v[2]*255)))
	a := uint8(max(0, min(255, v[3]*255)))
	return color.RGBA{r, g, b, a}
}

type ImageView struct {
	pixels []color.RGBA
	width  int32
	height int32
}

// actually it's ABGR8888
func (iv *ImageView) Clear(c Vec4f) {
	v := c.toColor()
	for i := range iv.pixels {
		iv.pixels[i] = v
	}
}

func main() {
	sdl.Init(sdl.INIT_VIDEO)
	window, err := sdl.CreateWindow("Tiny rasterizer",
		sdl.WINDOWPOS_UNDEFINED,
		sdl.WINDOWPOS_UNDEFINED,
		width, height,
		sdl.WINDOW_RESIZABLE|sdl.WINDOW_SHOWN)
	if err != nil {
		log.Fatalf("%+v\n", err)
	}
	defer window.Destroy()

	var drawSurface *sdl.Surface
	lastFrameStart := time.Now()
	running := true

	mouseX := 0
	mouseY := 0
	for running {
		for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
			switch t := event.(type) {
			case *sdl.WindowEvent:
				if t.Event == sdl.WINDOWEVENT_SIZE_CHANGED {
					if drawSurface != nil {
						drawSurface.Free()
					}
					drawSurface = nil
					width = t.Data1
					height = t.Data2
				}
			case *sdl.MouseMotionEvent:
				mouseX = int(t.X)
				mouseY = int(t.Y)
				_ = mouseX
				_ = mouseY
			case *sdl.QuitEvent:
				running = false
			}
		}

		if !running {
			break
		}

		if drawSurface == nil {
			drawSurface, err = sdl.CreateRGBSurfaceWithFormat(0, width, height, 32, uint32(sdl.PIXELFORMAT_RGBA32))
			if err != nil {
				log.Fatalf("%+v\n", err)
			}
			drawSurface.SetBlendMode(sdl.BLENDMODE_NONE)
		}

		now := time.Now()
		dt := now.Sub(lastFrameStart).Seconds()
		lastFrameStart = now
		_ = dt

		drawSurface.Lock()
		ps := drawSurface.Pixels()
		pixels := unsafe.Slice((*color.RGBA)(unsafe.Pointer(&ps[0])), width*height)
		colorBuffer := ImageView{pixels: pixels, width: width, height: height}
		colorBuffer.Clear(Vec4f{0.8, 0.9, 1.0, 1.0})
		drawSurface.Unlock()

		rect := sdl.Rect{X: 0, Y: 0, W: width, H: height}

		if windowSurface, err := window.GetSurface(); err != nil {
			log.Fatalf("%+v\n", err)
		} else {
			windowSurface.FillRect(nil, 0)
			drawSurface.Blit(&rect, windowSurface, &rect)
		}

		if err = window.UpdateSurface(); err != nil {
			log.Fatalf("%+v\n", err)
		}
	}
}
