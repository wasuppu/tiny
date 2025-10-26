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

type DepthTestMode int

const (
	Never DepthTestMode = iota
	Always
	Less
	LessEqual
	Greater
	GreaterEqual
	Equal
	NotEqual
)

type DepthSettings struct {
	write bool
	mode  DepthTestMode
}

func NewDepthSettings() DepthSettings {
	return DepthSettings{true, Always}
}

func DepthTestPassed(mode DepthTestMode, value, reference uint32) bool {
	switch mode {
	case Always:
		return true
	case Never:
		return false
	case Less:
		return value < reference
	case LessEqual:
		return value <= reference
	case Greater:
		return value > reference
	case GreaterEqual:
		return value >= reference
	case Equal:
		return value == reference
	case NotEqual:
		return value != reference
	default: // Unreachable
		return true
	}
}

type CullMode int

const (
	None CullMode = iota
	CW
	CCW
)

type DrawCommand struct {
	mesh      Mesh
	cullMode  CullMode
	depth     DepthSettings
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

type Pixel interface {
	color.RGBA | uint32
}

type ImageView[T Pixel] struct {
	pixels []T
	width  int32
	height int32
}

func (iv *ImageView[T]) At(x, y int32) T {
	return iv.pixels[x+y*iv.width]
}

func (iv *ImageView[T]) Set(x, y int32, v T) bool {
	if x >= 0 && x < iv.width && y >= 0 && y < iv.height {
		iv.pixels[x+y*iv.width] = v
		return true
	}
	return false
}

type Colorbuffer = ImageView[color.RGBA]
type Depthbuffer = ImageView[uint32]

func ClearColorbuffer(iv *Colorbuffer, c Vec4f) {
	v := c.ToColor()
	for i := range iv.pixels {
		iv.pixels[i] = v
	}
}

func ClearDepthbuffer(iv *Depthbuffer, value uint32) {
	for i := range iv.pixels {
		iv.pixels[i] = value
	}
}

func Allocate[T Pixel](width, height int32) ImageView[T] {
	return ImageView[T]{make([]T, width*height), width, height}
}

type Framebuffer struct {
	color *Colorbuffer
	depth *Depthbuffer
}

func (f Framebuffer) Width() int32 {
	if f.color != nil {
		return f.color.width
	}
	return f.depth.width
}

func (f Framebuffer) Height() int32 {
	if f.color != nil {
		return f.color.height
	}
	return f.depth.height
}

func (framebuffer *Framebuffer) Draw(viewport Viewport, command DrawCommand) {
	for vertexIndex := uint32(0); vertexIndex+2 < command.mesh.count; vertexIndex += 3 {
		i0 := vertexIndex + 0
		i1 := vertexIndex + 1
		i2 := vertexIndex + 2

		if len(command.mesh.indices) > 0 {
			i0 = command.mesh.indices[i0]
			i1 = command.mesh.indices[i1]
			i2 = command.mesh.indices[i2]
		}

		clippedVertices := [12]Vertex{}

		clippedVertices[0].position = command.transform.Mulv(command.mesh.positions[i0].AsPoint())
		clippedVertices[1].position = command.transform.Mulv(command.mesh.positions[i1].AsPoint())
		clippedVertices[2].position = command.transform.Mulv(command.mesh.positions[i2].AsPoint())

		clippedVertices[0].color = command.mesh.colors[i0]
		clippedVertices[1].color = command.mesh.colors[i1]
		clippedVertices[2].color = command.mesh.colors[i2]

		result := ClipTriangle(clippedVertices[:3])
		for i := 0; i < len(result); i += 3 {
			v0 := result[i].position
			v1 := result[i+1].position
			v2 := result[i+2].position

			c0 := result[i].color
			c1 := result[i+1].color
			c2 := result[i+2].color

			v0 = v0.PerspectiveDivide()
			v1 = v1.PerspectiveDivide()
			v2 = v2.PerspectiveDivide()

			v0 = viewport.Apply(v0)
			v1 = viewport.Apply(v1)
			v2 = viewport.Apply(v2)

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
			xmax := int32(min(viewport.xmax, framebuffer.Width()) - 1)
			ymin := int32(max(viewport.ymin, 0))
			ymax := int32(min(viewport.ymax, framebuffer.Height()) - 1)

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
						l0 := det12p / det012 * v0.W()
						l1 := det20p / det012 * v1.W()
						l2 := det01p / det012 * v2.W()

						lsum := l0 + l1 + l2

						l0 /= lsum
						l1 /= lsum
						l2 /= lsum

						if len(framebuffer.depth.pixels) != 0 {
							z := v0.Z()*l0 + v1.Z()*l1 + v2.Z()*l2

							// Convert from [-1, 1] to [0, UINT32_MAX]
							depth := uint32((0.5 + 0.5*z) * math.MaxUint32)

							if !DepthTestPassed(command.depth.mode, uint32(depth), framebuffer.depth.At(x, y)) {
								continue
							}

							if command.depth.write {
								framebuffer.depth.Set(x, y, depth)
							}
						}

						c := c0.Muln(l0).Add(c1.Muln(l1)).Add(c2.Muln(l2))
						framebuffer.color.Set(x, y, c.ToColor())
					}
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

	depthbuffer := Depthbuffer{}

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
					depthbuffer = Depthbuffer{}
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

		if len(depthbuffer.pixels) == 0 {
			depthbuffer = Allocate[uint32](width, height)
		}

		now := time.Now()
		dt := now.Sub(lastFrameStart).Seconds()
		lastFrameStart = now
		tm += dt
		// fmt.Println(dt)

		drawSurface.Lock()
		ps := drawSurface.Pixels()
		pixels := unsafe.Slice((*color.RGBA)(unsafe.Pointer(&ps[0])), width*height)
		framebuffer := Framebuffer{
			color: &Colorbuffer{pixels: pixels, width: width, height: height},
			depth: &depthbuffer,
		}
		ClearColorbuffer(framebuffer.color, Vec4f{0.9, 0.9, 0.9, 1.0})
		ClearDepthbuffer(framebuffer.depth, math.MaxUint32)
		drawSurface.Unlock()

		viewport := Viewport{0, 0, width, height}

		cubeScale := 1.0

		model := ScaleF(cubeScale).Mul(RotateZX(tm)).Mul(RotateXY(tm * 1.61))
		view := Translate(Vec3f{0, 0, -5})
		projection := Perspective(0.01, 10, math.Pi/3, float64(width)/float64(height))

		command := DrawCommand{
			mesh:      Cube,
			depth:     DepthSettings{true, Less},
			transform: projection.Mul(view).Mul(model),
		}
		framebuffer.Draw(viewport, command)

		// for i := -2; i <= 2; i++ {
		// 	command := DrawCommand{
		// 		mesh:      Cube,
		// 		cullMode:  None,
		// 		depth:     DepthSettings{true, Less},
		// 		transform: projection.Mul(view).Mul(Translate(Vec3f{float64(i), 0, 0})).Mul(model),
		// 	}
		// 	framebuffer.Draw(viewport, command)
		// }

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
