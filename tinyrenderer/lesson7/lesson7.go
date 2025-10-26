package main

import (
	"fmt"
	"image/color"
	"math"
	"os"
	"path/filepath"
	"runtime"
)

const (
	width  = 800
	height = 800
	depth  = 2000.0
)

var (
	obj          *Obj
	lightDir     = Vec3f{1, 1, 0}
	eye          = Vec3f{1, 1, 4}
	center       = Vec3f{0, 0, 0}
	up           = Vec3f{0, 1, 0}
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

func Triangle12(pts [3]Vec4f, shader IShader, img *TGAImage, zbuffer []float64) {
	bboxmin := Vec2f{math.MaxFloat64, math.MaxFloat64}
	bboxmax := Vec2f{-math.MaxFloat64, -math.MaxFloat64}
	for i := range 3 {
		for j := range 2 {
			bboxmin[j] = math.Min(bboxmin[j], pts[i][j]/pts[i][3])
			bboxmax[j] = math.Max(bboxmax[j], pts[i][j]/pts[i][3])
		}
	}

	p := Vec2i{}
	for p[0] = int(bboxmin.X()); float64(p.X()) <= bboxmax.X(); p[0]++ {
		for p[1] = int(bboxmin.Y()); float64(p.Y()) <= bboxmax.Y(); p[1]++ {
			b := Barycentric3(pts[0].Divn(pts[0][3]).V2(), pts[1].Divn(pts[1][3]).V2(), pts[2].Divn(pts[2][3]).V2(), p.F())
			z := pts[0][2]*b.X() + pts[1][2]*b.Y() + pts[2][2]*b.Z()
			w := pts[0][3]*b.X() + pts[1][3]*b.Y() + pts[2][3]*b.Z()
			fragDepth := int(z / w)
			if b.X() < 0 || b.Y() < 0 || b.Z() < 0 || zbuffer[p.X()+p.Y()*img.Width] > float64(fragDepth) {
				continue
			}
			c, discard := shader.Fragment(b)
			if !discard {
				zbuffer[p.X()+p.Y()*img.Width] = float64(fragDepth)
				img.Set(p.X(), p.Y(), c)
			}
		}
	}
}

type IShader interface {
	Vertex(iface, nthvert int) Vec4f
	Fragment(Vec3f) (color.Color, bool)
}

type DepthShader struct {
	varyingTri Mat3
}

func (s *DepthShader) Vertex(iface, nthvert int) Vec4f {
	glVertex := obj.Vertf(iface, nthvert).V4(1) // read the vertex from .obj file
	glVertex = viewport.Mul(project).Mul(modelView).Mulv(glVertex)
	s.varyingTri.SetCol(nthvert, glVertex.Divn(glVertex[3]).V3())
	return glVertex
}

func (s DepthShader) Fragment(bar Vec3f) (color.Color, bool) {
	p := s.varyingTri.Mulv(bar)
	c := ScaleColorRGB(color.RGBA{255, 255, 255, 255}, p.Z()/depth)
	return c, false
}

type Shader6 struct {
	uniformM       Mat4   //  Projection*ModelView
	uniformMIT     Mat4   // (Projection*ModelView).invert_transpose()
	uniformMshadow Mat4   // transform framebuffer screen coordinates to shadowbuffer screen coordinates
	veryingUV      Mat2x3 // triangle uv coordinates, written by the vertex shader, read by the fragment shader
	varyingTri     Mat3   // triangle coordinates before Viewport transform, written by VS, read by FS
}

func (s *Shader6) Vertex(iface, nthvert int) Vec4f {
	s.veryingUV.SetCol(nthvert, obj.UVf(iface, nthvert))
	glVertex := viewport.Mul(project).Mul(modelView).Mulv(obj.Vertf(iface, nthvert).V4(1)) // read the vertex from .obj file
	s.varyingTri.SetCol(nthvert, glVertex.Divn(glVertex[3]).V3())
	return glVertex
}

func (s Shader6) Fragment(bar Vec3f) (color.Color, bool) {
	sbp := s.uniformMshadow.Mulv(s.varyingTri.Mulv(bar).V4(1)) // corresponding point in the shadow buffer
	sbp = sbp.Divn(sbp[3])
	idx := int(sbp[0]) + int(sbp[1])*width // index in the shadowbuffer array
	shadow := 0.3
	if shadowbuffer[idx] < sbp[2]+43.34 {
		shadow = shadow + 0.7 // magic coeff to avoid z-fighting
	}
	uv := s.veryingUV.Mulv(bar)                                   // interpolate uv for the current pixel
	n := s.uniformMIT.Mulv(obj.NormUV(uv).V4(1)).V3().Normalize() // normal
	l := s.uniformM.Mulv(lightDir.V4(1)).V3().Normalize()         // light vector
	r := n.Muln(n.Dot(l) * 2).Sub(l).Normalize()                  // reflected light
	spec := math.Pow(math.Max(r.Z(), 0), obj.Specular(uv))
	diff := math.Max(0, n.Dot(l))

	c1 := obj.Diffusef(uv)
	c := c1
	c.B = uint8(math.Min(20+float64(c1.B)*shadow*(1.2*diff+0.6*spec), 255))
	c.G = uint8(math.Min(20+float64(c1.G)*shadow*(1.2*diff+0.6*spec), 255))
	c.R = uint8(math.Min(20+float64(c1.R)*shadow*(1.2*diff+0.6*spec), 255))
	return c, false
}

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "Usage: go run ./%s obj/model.obj\n", filepath.Base(os.Args[0]))
		os.Exit(1)
	}

	zbuffer := make([]float64, int(width*height))
	for i := range zbuffer {
		zbuffer[i] = -math.MaxFloat64
		shadowbuffer[i] = -math.MaxFloat64
	}

	obj = LoadObj(os.Args[1])
	lightDir = lightDir.Normalize()

	{ // rendering the shadow buffer
		tga := NewTgaImg(width, height)

		modelView = LookAt(lightDir, center, up)
		viewport = Viewport(width/8, height/8, width*3/4, height*3/4)
		project = Projection(0)

		depthshader := DepthShader{}
		screenCoords := [3]Vec4f{}
		for i := range obj.NFaces() {
			for j := range 3 {
				screenCoords[j] = depthshader.Vertex(i, j)
			}
			Triangle12(screenCoords, &depthshader, tga, shadowbuffer[:])
		}

		tga.FlipVertically()
		tga.UseRLE = true
		tga.Write(filepath.Join(parentpath, "lesson7-depth.tga"))
	}

	m := viewport.Mul(project).Mul(modelView)

	{ // rendering the frame buffer
		tga := NewTgaImg(width, height)

		modelView = LookAt(eye, center, up)
		viewport = Viewport(width/8, height/8, width*3/4, height*3/4)
		project = Projection(-1.0 / (eye.Sub(center).Length()))

		shader := Shader6{uniformM: modelView, uniformMIT: project.Mul(modelView).Inverse().Transpose(), uniformMshadow: m.Mul(viewport.Mul(project).Mul(modelView).Inverse())}
		screenCoords := [3]Vec4f{}
		for i := range obj.NFaces() {
			for j := range 3 {
				screenCoords[j] = shader.Vertex(i, j)
			}
			Triangle12(screenCoords, &shader, tga, zbuffer)
		}

		tga.FlipVertically()
		tga.UseRLE = true
		tga.Write(filepath.Join(parentpath, "lesson7.tga"))
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
