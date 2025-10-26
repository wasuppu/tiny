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

type Viewport struct {
	xmin, ymin, xmax, ymax int32
}

func (vp Viewport) Apply(v Vec4f) Vec4f {
	v[0] = float64(vp.xmin) + float64(vp.xmax-vp.xmin)*(0.5+0.5*v.X())
	v[1] = float64(vp.ymin) + float64(vp.ymax-vp.ymin)*(0.5-0.5*v.Y())
	return v
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

func (iv *ImageView) Draw(viewport Viewport, command DrawCommand) {
	for vertexIndex := uint32(0); vertexIndex+2 < command.mesh.count; vertexIndex += 3 {
		i0 := vertexIndex + 0
		i1 := vertexIndex + 1
		i2 := vertexIndex + 2

		if len(command.mesh.indices) > 0 {
			i0 = command.mesh.indices[i0]
			i1 = command.mesh.indices[i1]
			i2 = command.mesh.indices[i2]
		}

		v0 := command.transform.Mulv(command.mesh.positions[i0].AsPoint())
		v1 := command.transform.Mulv(command.mesh.positions[i1].AsPoint())
		v2 := command.transform.Mulv(command.mesh.positions[i2].AsPoint())

		v0 = viewport.Apply(v0)
		v1 = viewport.Apply(v1)
		v2 = viewport.Apply(v2)

		c0 := command.mesh.colors[i0]
		c1 := command.mesh.colors[i1]
		c2 := command.mesh.colors[i2]

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
			det012 = -det012
		}

		xmin := int32(max(viewport.xmin, 0))
		xmax := int32(min(viewport.xmax, iv.width) - 1)
		ymin := int32(max(viewport.ymin, 0))
		ymax := int32(min(viewport.ymax, iv.height) - 1)

		xmin = int32(max(float64(xmin), min(math.Floor(v0.X()), math.Floor(v1.X()), math.Floor(v2.X()))))
		xmax = int32(min(float64(xmax), max(math.Floor(v0.X()), math.Floor(v1.X()), math.Floor(v2.X()))))
		ymin = int32(max(float64(ymin), min(math.Floor(v0.Y()), math.Floor(v1.Y()), math.Floor(v2.Y()))))
		ymax = int32(min(float64(ymax), max(math.Floor(v0.Y()), math.Floor(v1.Y()), math.Floor(v2.Y()))))

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
	tm := 0.0
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
		tm += dt
		// fmt.Println(dt)

		drawSurface.Lock()
		ps := drawSurface.Pixels()
		pixels := unsafe.Slice((*color.RGBA)(unsafe.Pointer(&ps[0])), width*height)
		colorBuffer := ImageView{pixels: pixels, width: width, height: height}
		colorBuffer.Clear(Vec4f{0.9, 0.9, 0.9, 1.0})
		drawSurface.Unlock()

		viewport := Viewport{0, 0, width, height}

		positions := []Vec3f{
			{-0.5, -0.5, 0},
			{-0.5, 0.5, 0},
			{0.5, -0.5, 0},
			{0.5, 0.5, 0},
		}

		colors := []Vec4f{
			{1, 0, 0, 1},
			{0, 1, 0, 1},
			{0, 0, 1, 1},
			{1, 1, 1, 1},
		}

		indices := []uint32{
			0, 1, 2,
			2, 1, 3,
		}

		transform := RotateZX(tm)

		command := DrawCommand{
			mesh:      Mesh{positions: positions, colors: colors, indices: indices, count: 6},
			cullMode:  None,
			transform: transform,
		}
		colorBuffer.Draw(viewport, command)

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
