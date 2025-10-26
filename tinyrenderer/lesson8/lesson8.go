package main

import (
	"fmt"
	"image/color"
	"math"
	"math/rand"
	"os"
	"path/filepath"
	"runtime"
)

const (
	width  = 800
	height = 800
	isize  = 1024
)

var (
	obj          *Obj
	eye          = Vec3f{1.2, -0.8, 3}
	center       = Vec3f{0, 0, 0}
	up           = Vec3f{0, 1, 0}
	total        *TGAImage
	occl         *TGAImage
	modelView    Mat4
	viewport     Mat4
	project      Mat4
	shadowbuffer [width * height]float64
	basepath     string
	parentpath   string
)

func init() {
	_, exepath, _, _ := runtime.Caller(0)
	basepath = filepath.Dir(exepath)
	parentpath = filepath.Dir(basepath)
}

func Barycentric3(a, b, c, p Vec2f) Vec3f {
	s := [2]Vec3f{}
	for i := range 2 {
		s[i][0] = c[i] - a[i]
		s[i][1] = b[i] - a[i]
		s[i][2] = a[i] - p[i]
	}
	u := s[0].Cross(s[1])
	if math.Abs(u[2]) > 1e-2 { // dont forget that u[2] is integer. If it is zero then triangle ABC is degenerate
		return Vec3f{1 - (u.X()+u.Y())/u.Z(), u.Y() / u.Z(), u.X() / u.Z()}
	}
	return Vec3f{-1, 1, 1} // in this case generate negative coordinates, it will be thrown away by the rasterizator
}

func Triangle11(clipc Mat4x3, shader IShader2, img *TGAImage, zbuffer []float64) {
	pts := viewport.Mul4x3(clipc).Transpose()
	pts2 := Mat3x2{}
	for i := range 3 {
		pts2[i] = pts[i].Divn(pts[i][3]).V2()
	}

	bboxmin := Vec2f{math.MaxFloat64, math.MaxFloat64}
	bboxmax := Vec2f{-math.MaxFloat64, -math.MaxFloat64}
	clamp := Vec2f{float64(img.Width - 1), float64(img.Height - 1)}
	for i := range 3 {
		for j := range 2 {
			bboxmin[j] = math.Max(0, math.Min(bboxmin[j], pts2[i][j]))
			bboxmax[j] = math.Min(clamp[j], math.Max(bboxmax[j], pts2[i][j]))
		}
	}

	p := Vec2i{}
	for p[0] = int(bboxmin.X()); p.X() <= int(bboxmax.X()); p[0]++ {
		for p[1] = int(bboxmin.Y()); p.Y() <= int(bboxmax.Y()); p[1]++ {
			bcScreen := Barycentric3(pts2[0], pts2[1], pts2[2], p.F())
			bcClip := Vec3f{bcScreen.X() / pts[0][3], bcScreen.Y() / pts[1][3], bcScreen.Z() / pts[2][3]}
			bcClip = bcClip.Divn(bcClip.X() + bcClip.Y() + bcClip.Z())
			fragDepth := clipc[2].Dot(bcClip)

			if bcScreen.X() < 0 || bcScreen.Y() < 0 || bcScreen.Z() < 0 || zbuffer[p.X()+p.Y()*img.Width] > fragDepth {
				continue
			}

			c, discard := shader.Fragment(Vec3f{float64(p.X()), float64(p.Y()), fragDepth}, bcClip)
			if !discard {
				zbuffer[p.X()+p.Y()*img.Width] = fragDepth
				img.Set(p.X(), p.Y(), c)
			}
		}
	}
}

type IShader2 interface {
	Vertex(iface, nthvert int) Vec4f
	Fragment(glFragCoord Vec3f, bar Vec3f) (color.Color, bool)
}

type ZShader struct {
	varyingTri Mat4x3
}

func (s *ZShader) Vertex(iface, nthvert int) Vec4f {
	glVertex := project.Mul(modelView).Mulv(obj.Vertf(iface, nthvert).V4(1))
	s.varyingTri.SetCol(nthvert, glVertex)
	return glVertex
}

func (s ZShader) Fragment(glFragCoord Vec3f, bar Vec3f) (color.Color, bool) {
	return ScaleColorRGB(color.RGBA{255, 255, 255, 255}, (glFragCoord.Z()+1)/2), false
}

type Shader7 struct {
	veryingUV  Mat2x3
	varyingTri Mat4x3
}

func (s *Shader7) Vertex(iface, nthvert int) Vec4f {
	s.veryingUV.SetCol(nthvert, obj.UVf(iface, nthvert))
	glVertex := project.Mul(modelView).Mulv(obj.Vertf(iface, nthvert).V4(1))
	s.varyingTri.SetCol(nthvert, glVertex)
	return glVertex
}

func (s Shader7) Fragment(glFragCoord Vec3f, bar Vec3f) (color.Color, bool) {
	uv := s.veryingUV.Mulv(bar)
	if math.Abs(shadowbuffer[int(glFragCoord.X()+glFragCoord.Y()*width)])-glFragCoord.Z() < 1e-2 {
		occl.Set(int(uv.X()*isize), int(uv.Y()*isize), color.Gray{255})
	}
	c := color.RGBA{255, 0, 0, 255}
	return c, false
}

type AOShader struct {
	veryingUV  Mat2x3
	varyingTri Mat4x3
	aoimage    TGAImage
}

func (s *AOShader) Vertex(iface, nthvert int) Vec4f {
	s.veryingUV.SetCol(nthvert, obj.UVf(iface, nthvert))
	glVertex := project.Mul(modelView).Mulv(obj.Vertf(iface, nthvert).V4(1))
	s.varyingTri.SetCol(nthvert, glVertex)
	return glVertex
}

func (s AOShader) Fragment(glFragCoord Vec3f, bar Vec3f) (color.Color, bool) {
	uv := s.veryingUV.Mulv(bar)
	t := s.aoimage.At(int(uv.X()*isize), int(uv.Y()*isize)).R
	c := color.RGBA{t, t, t, 255}
	return c, false
}

type ZShader2 struct {
	varyingTri Mat4x3
}

func (s *ZShader2) Vertex(iface, nthvert int) Vec4f {
	glVertex := project.Mul(modelView).Mulv(obj.Vertf(iface, nthvert).V4(1))
	s.varyingTri.SetCol(nthvert, glVertex)
	return glVertex
}

func (s ZShader2) Fragment(glFragCoord Vec3f, bar Vec3f) (color.Color, bool) {
	return color.RGBA{0, 0, 0, 255}, false
}

func randPointOnUnitShpere() Vec3f {
	u := rand.Float64()
	v := rand.Float64()
	theta := 2 * math.Pi * u
	phi := math.Acos(2*v - 1)
	return Vec3f{math.Sin(phi) * math.Cos(theta), math.Sin(phi) * math.Sin(theta), math.Cos(phi)}
}

func maxElevationAngle(zbuffer []float64, p, dir Vec2f) float64 {
	maxangle := 0.0
	for t := 0.0; t < 1000; t += 1.0 {
		cur := p.Add(dir.Muln(t))
		if cur.X() >= width || cur.Y() >= height || cur.X() < 0 || cur.Y() < 0 {
			return maxangle
		}

		distance := p.Sub(cur).Length()
		if distance < 1 {
			continue
		}
		elevation := zbuffer[int(cur.X())+int(cur.Y())*width] - zbuffer[int(p.X())+int(p.Y())*width]
		maxangle = math.Max(maxangle, math.Atan(elevation/distance))
	}
	return maxangle
}

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "Usage: go run ./%s obj/model.obj\n", filepath.Base(os.Args[0]))
		os.Exit(1)
	}

	obj = LoadObj(os.Args[1])

	// There are some errors in occl, which resulted in error of occlusion and in turn led to incorrect aoimage
	// it seems to be related to the implementation of TGA images, i don't know, anyway, i don't have enough patience to solve it. sigh
	{
		total = NewGreyTgaImage(isize, isize)
		occl = NewGreyTgaImage(isize, isize)
		zbuffer := make([]float64, int(width*height))

		nrenders := 1
		for iter := 1; iter <= nrenders; iter++ {
			fmt.Printf("%d from %d\n", iter, nrenders)
			for i := range 3 {
				up[i] = rand.Float64()
			}
			eye = randPointOnUnitShpere()
			eye[1] = math.Abs(eye.Y())
			fmt.Println("v", eye)

			for i := range zbuffer {
				zbuffer[i] = -math.MaxFloat64
				shadowbuffer[i] = -math.MaxFloat64
			}

			frame := NewTgaImg(width, height)
			modelView = LookAt(eye, center, up)
			viewport = Viewport(width/8, height/8, width*3/4, height*3/4)
			project = Projection(0)

			zshader := ZShader{}
			for i := range obj.NFaces() {
				for j := range 3 {
					zshader.Vertex(i, j)
				}
				Triangle11(zshader.varyingTri, &zshader, frame, shadowbuffer[:])
			}
			frame.FlipVertically()
			frame.UseRLE = true
			frame.Write(filepath.Join(parentpath, "lesson8.tga"))

			shader := Shader7{}
			occl.Clear()
			for i := range obj.NFaces() {
				for j := range 3 {
					shader.Vertex(i, j)
				}
				Triangle11(shader.varyingTri, &shader, frame, zbuffer)
			}

			occl.GaussianBlur(5)
			for i := range isize {
				for j := range isize {
					t := float64(total.At(i, j).R)
					o := float64(occl.At(i, j).R)
					total.Set(i, j, color.Gray{uint8(float64(t*float64(iter-1)+o)/float64(iter) + 0.5)})
				}
			}
		}
		total.FlipVertically()
		total.UseRLE = true
		total.Write(filepath.Join(parentpath, "lesson8-occlusion.tga"))
		occl.FlipVertically()
		occl.UseRLE = true
		occl.Write(filepath.Join(parentpath, "lesson8-occl.tga"))
	}

	{
		frame := NewTgaImg(width, height)
		zbuffer := make([]float64, int(width*height))
		for i := range zbuffer {
			zbuffer[i] = -math.MaxFloat64
		}

		modelView = LookAt(eye, center, up)
		viewport = Viewport(width/8, height/8, width*3/4, height*3/4)
		project = Projection(-1 / eye.Sub(center).Length())

		aoshader := AOShader{}
		aoshader.aoimage.read(filepath.Join(parentpath, "occlusion.tga"))
		aoshader.aoimage.FlipVertically()

		for i := range obj.NFaces() {
			for j := range 3 {
				aoshader.Vertex(i, j)
			}
			Triangle11(aoshader.varyingTri, &aoshader, frame, zbuffer)
		}

		frame.FlipVertically()
		frame.UseRLE = true
		frame.Write(filepath.Join(parentpath, "lesson8-ao.tga"))
	}

	{
		frame := NewTgaImg(width, height)
		zbuffer := make([]float64, int(width*height))
		for i := range zbuffer {
			zbuffer[i] = -math.MaxFloat64
		}

		modelView = LookAt(eye, center, up)
		viewport = Viewport(width/8, height/8, width*3/4, height*3/4)
		project = Projection(-1 / eye.Sub(center).Length())

		zshader := ZShader2{}
		for i := range obj.NFaces() {
			for j := range 3 {
				zshader.Vertex(i, j)
			}
			Triangle11(zshader.varyingTri, &zshader, frame, zbuffer)
		}

		for x := range width {
			for y := range height {
				if zbuffer[x+y*width] < -1e5 {
					continue
				}
				total := 0.0
				for a := 0.0; a < math.Pi*2-1e-4; a += math.Pi / 4 {
					s, c := math.Sincos(a)
					total += math.Pi/2 - maxElevationAngle(zbuffer, Vec2f{float64(x), float64(y)}, Vec2f{c, s})
				}
				total /= math.Pi / 2 * 8
				total = math.Pow(total, 100)
				frame.Set(x, y, color.RGBA{uint8(total * 255), uint8(total * 255), uint8(total * 255), 255})
			}
		}

		frame.FlipVertically()
		frame.UseRLE = true
		frame.Write(filepath.Join(parentpath, "lesson8-2.tga"))
	}
}

func ScaleColorRGB(c color.RGBA, intensity float64) color.RGBA {
	if intensity > 1 {
		intensity = 1
	} else if intensity < 0 {
		intensity = 0
	}

	r := uint8(float64(c.R) * intensity)
	g := uint8(float64(c.G) * intensity)
	b := uint8(float64(c.B) * intensity)
	a := c.A

	return color.RGBA{R: r, G: g, B: b, A: a}
}
