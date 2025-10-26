package main

import (
	"image/color"
	"log"
	"math"
	"time"
	"unsafe"

	"github.com/veandco/go-sdl2/sdl"
)

var (
	width  int32 = 800
	height int32 = 600
)

type CullMode int

const (
	None CullMode = iota
	CW
	CCW
)

type DrawCommand struct {
	mesh      Mesh
	cullMode  CullMode
	transform Mat4
}

func NewDrawCommand() DrawCommand {
	return DrawCommand{cullMode: None, transform: ID4()}
}

type ImageView struct {
	pixels []color.RGBA
	width  int32
	height int32
}

// actually it's ABGR8888
func (iv *ImageView) Clear(c Vec4f) {
	v := c.ToColor()
	for i := range iv.pixels {
		iv.pixels[i] = v
	}
}

func (iv *ImageView) Set(x, y int32, c color.RGBA) bool {
	if x >= 0 && x < iv.width && y >= 0 && y < iv.height {
		iv.pixels[x+y*iv.width] = c
		return true
	}
	return false
}

func (iv *ImageView) Draw(command DrawCommand) {
	for vertexIndex := uint32(0); vertexIndex+2 < command.mesh.count; vertexIndex += 3 {
		v0 := command.transform.Mulv(command.mesh.positions[vertexIndex+0].AsPoint())
		v1 := command.transform.Mulv(command.mesh.positions[vertexIndex+1].AsPoint())
		v2 := command.transform.Mulv(command.mesh.positions[vertexIndex+2].AsPoint())

		c0 := command.mesh.colors[vertexIndex+0]
		c1 := command.mesh.colors[vertexIndex+1]
		c2 := command.mesh.colors[vertexIndex+2]

		det012 := v1.Sub(v0).Det2D(v2.Sub(v0))

		// Is it counterclockwise on screen?
		ccw := det012 < 0.0

		switch command.cullMode {
		case None:
		case CW:
			if !ccw {
				continue // move to the next triangle
			}
		case CCW:
			if ccw {
				continue // move to the next triangle
			}
		}

		if ccw {
			v1, v2 = v2, v1
			c1, c2 = c2, c1
			// det012 = -det012
		}

		xmin := int32(min(math.Floor(v0.X()), math.Floor(v1.X()), math.Floor(v2.X())))
		xmax := int32(max(math.Floor(v0.X()), math.Floor(v1.X()), math.Floor(v2.X())))
		ymin := int32(min(math.Floor(v0.Y()), math.Floor(v1.Y()), math.Floor(v2.Y())))
		ymax := int32(max(math.Floor(v0.Y()), math.Floor(v1.Y()), math.Floor(v2.Y())))

		xmin = max(0, xmin)
		xmax = min(iv.width-1, xmax)
		ymin = max(0, ymin)
		ymax = min(iv.height-1, ymax)

		for y := ymin; y <= ymax; y++ {
			for x := xmin; x <= xmax; x++ {
				p := Vec4f{float64(x) + 0.5, float64(y) + 0.5, 0, 0}

				det01p := v1.Sub(v0).Det2D(p.Sub(v0))
				det12p := v2.Sub(v1).Det2D(p.Sub(v1))
				det20p := v0.Sub(v2).Det2D(p.Sub(v2))

				if det01p >= 0 && det12p >= 0 && det20p >= 0 {
					l0 := det12p / det012
					l1 := det20p / det012
					l2 := det01p / det012

					iv.Set(x, y, c0.Muln(l0).Add(c1.Muln(l1)).Add(c2.Muln(l2)).ToColor())
				}
			}
		}
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
		// fmt.Println(dt)

		drawSurface.Lock()
		ps := drawSurface.Pixels()
		pixels := unsafe.Slice((*color.RGBA)(unsafe.Pointer(&ps[0])), width*height)
		colorBuffer := ImageView{pixels: pixels, width: width, height: height}
		colorBuffer.Clear(Vec4f{0.8, 0.9, 1.0, 1.0})
		drawSurface.Unlock()

		positions := []Vec3f{
			{0, 0, 0},
			{100, 0, 0},
			{0, 100, 0},
		}

		colors := []Vec4f{
			{1, 0, 0, 1},
			{0, 1, 0, 1},
			{0, 0, 1, 1},
		}

		for i := range 100 {
			command := DrawCommand{
				mesh:     Mesh{positions: positions, colors: colors, count: 3},
				cullMode: None,
				transform: Mat4{
					{1, 0, 0, float64(mouseX + 100*(i%10))},
					{0, 1, 0, float64(mouseY + 100*(i/10))},
					{0, 0, 1, 0},
					{0, 0, 0, 1},
				},
			}

			colorBuffer.Draw(command)
		}

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
