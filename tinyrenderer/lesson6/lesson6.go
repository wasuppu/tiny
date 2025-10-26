package main

import (
	"fmt"
	"image/color"
	"math"
	"path/filepath"
	"runtime"
)

var (
	width    = 800
	height   = 800
	obj      *Obj
	lightDir = Vec3f{1, 1, 1}.Normalize()
	// eye        = Vec3f{0, -1, 3}
	eye        = Vec3f{1, 1, 3}
	center     = Vec3f{0, 0, 0}
	up         = Vec3f{0, 1, 0}
	modelView  = LookAt(eye, center, up)
	viewport   = Viewport(width/8, height/8, width*3/4, height*3/4)
	project    = Projection(-1 / eye.Sub(center).Length())
	basepath   string
	parentpath string
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

func Triangle10(pts [3]Vec4f, shader IShader, img *TGAImage, zbuffer *TGAImage) {
	bboxmin := Vec2f{math.MaxFloat64, math.MaxFloat64}
	bboxmax := Vec2f{-math.MaxFloat64, -math.MaxFloat64}
	for i := range 3 {
		for j := range 2 {
			bboxmin[j] = math.Min(bboxmin[j], pts[i][j]/pts[i][3])
			bboxmax[j] = math.Max(bboxmax[j], pts[i][j]/pts[i][3])
		}
	}

	p := Vec2i{}
	for p[0] = int(bboxmin.X()); p.X() <= int(bboxmax.X()); p[0]++ {
		for p[1] = int(bboxmin.Y()); p.Y() <= int(bboxmax.Y()); p[1]++ {
			b := Barycentric3(pts[0].Divn(pts[0][3]).V2(), pts[1].Divn(pts[1][3]).V2(), pts[2].Divn(pts[2][3]).V2(), p.F())
			z := pts[0][2]*b.X() + pts[1][2]*b.Y() + pts[2][2]*b.Z()
			w := pts[0][3]*b.X() + pts[1][3]*b.Y() + pts[2][3]*b.Z()
			fragDepth := math.Max(0, math.Min(255, float64(int(z/w+0.5))))

			if b.X() < 0 || b.Y() < 0 || b.Z() < 0 || float64(zbuffer.At(p.X(), p.Y()).R) > fragDepth {
				continue
			}
			c, discard := shader.Fragment(b)
			if !discard {
				zbuffer.Set(p.X(), p.Y(), color.Gray{uint8(fragDepth)})
				img.Set(p.X(), p.Y(), c)
			}
		}
	}
}

type IShader interface {
	Vertex(iface, nthvert int) Vec4f
	Fragment(Vec3f) (color.Color, bool)
}

type GouraudShader1 struct {
	varyingIntensity Vec3f // written by vertex shader, read by fragment shader
}

func (s *GouraudShader1) Vertex(iface, nthvert int) Vec4f {
	glVertex := obj.Vertf(iface, nthvert).V4(1)                                       // read the vertex from .obj file
	glVertex = viewport.Mul(project).Mul(modelView).Mulv(glVertex)                    // transform it to screen coordinates
	s.varyingIntensity[nthvert] = math.Max(0, obj.Norm(iface, nthvert).Dot(lightDir)) // get diffuse lighting intensity
	return glVertex
}

func (s GouraudShader1) Fragment(bar Vec3f) (color.Color, bool) {
	intensity := s.varyingIntensity.Dot(bar)                      // interpolate intensity for the current pixel
	c := ScaleColorRGB(color.RGBA{255, 255, 255, 255}, intensity) // well duh
	return c, false                                               // no, we do not discard this pixel
}

type GouraudShader2 struct {
	varyingIntensity Vec3f // written by vertex shader, read by fragment shader
}

func (s *GouraudShader2) Vertex(iface, nthvert int) Vec4f {
	glVertex := obj.Vertf(iface, nthvert).V4(1)                                       // read the vertex from .obj file
	glVertex = viewport.Mul(project).Mul(modelView).Mulv(glVertex)                    // transform it to screen coordinates
	s.varyingIntensity[nthvert] = math.Max(0, obj.Norm(iface, nthvert).Dot(lightDir)) // get diffuse lighting intensity
	return glVertex
}

func (s GouraudShader2) Fragment(bar Vec3f) (color.Color, bool) {
	intensity := s.varyingIntensity.Dot(bar) // interpolate intensity for the current pixel
	if intensity > 0.85 {
		intensity = 1
	} else if intensity > 0.60 {
		intensity = 0.80
	} else if intensity > 0.45 {
		intensity = 0.60
	} else if intensity > 0.30 {
		intensity = 0.45
	} else if intensity > 0.15 {
		intensity = 0.30
	} else {
		intensity = 0
	}
	c := ScaleColorRGB(color.RGBA{255, 155, 0, 255}, intensity) // well duh
	return c, false                                             // no, we do not discard this pixel
}

type Shader1 struct {
	varyingIntensity Vec3f
	veryingUV        Mat2x3
}

func (s *Shader1) Vertex(iface, nthvert int) Vec4f {
	s.veryingUV.SetCol(nthvert, obj.UVf(iface, nthvert))
	s.varyingIntensity[nthvert] = math.Max(0, obj.Norm(iface, nthvert).Dot(lightDir))
	glVert := obj.Vertf(iface, nthvert).V4(1)
	return viewport.Mul(project).Mul(modelView).Mulv(glVert)
}

func (s Shader1) Fragment(bar Vec3f) (color.Color, bool) {
	intensity := s.varyingIntensity.Dot(bar)
	uv := s.veryingUV.M().Mulv(bar.V())
	c := ScaleColorRGB(obj.Diffusef(Vec2f(uv)), intensity)
	return c, false
}

type Shader2 struct {
	veryingUV  Mat2x3
	uniformM   Mat4
	uniformMIT Mat4
}

func (s *Shader2) Vertex(iface, nthvert int) Vec4f {
	s.veryingUV.SetCol(nthvert, obj.UVf(iface, nthvert))
	glVertex := obj.Vertf(iface, nthvert).V4(1)                // read the vertex from .obj file
	return viewport.Mul(project).Mul(modelView).Mulv(glVertex) // transform it to screen coordinates
}

func (s Shader2) Fragment(bar Vec3f) (color.Color, bool) {
	uv := Vec2f(s.veryingUV.M().Mulv(bar[:])) // interpolate uv for the current pixel
	n := s.uniformMIT.Mulv(obj.NormUV(uv).V4(1)).V3().Normalize()
	l := s.uniformM.Mulv(lightDir.V4(1)).V3().Normalize()
	intensity := math.Max(0, n.Dot(l))
	c := ScaleColorRGB(obj.Diffusef(uv), intensity) // well duh
	return c, false                                 // no, we do not discard this pixel
}

type Shader3 struct {
	veryingUV  Mat2x3
	uniformM   Mat4
	uniformMIT Mat4
}

func (s *Shader3) Vertex(iface, nthvert int) Vec4f {
	s.veryingUV.SetCol(nthvert, obj.UVf(iface, nthvert))
	glVertex := obj.Vertf(iface, nthvert).V4(1)                // read the vertex from .obj file
	return viewport.Mul(project).Mul(modelView).Mulv(glVertex) // transform it to screen coordinates
}

func (s Shader3) Fragment(bar Vec3f) (color.Color, bool) {
	uv := Vec2f(s.veryingUV.M().Mulv(bar[:])) // interpolate uv for the current pixel
	n := s.uniformMIT.Mulv(obj.NormUV(uv).V4(1)).V3().Normalize()
	l := s.uniformM.Mulv(lightDir.V4(1)).V3().Normalize()
	r := n.Muln(n.Dot(l) * 2).Sub(l).Normalize() // reflected light
	spec := math.Pow(math.Max(r.Z(), 0), obj.Specular(uv))
	diff := math.Max(0, n.Dot(l))

	c1 := obj.Diffusef(uv)
	c := c1
	c.B = uint8(math.Min(5+float64(c1.B)*(diff+0.6*spec), 255))
	c.G = uint8(math.Min(5+float64(c1.G)*(diff+0.6*spec), 255))
	c.R = uint8(math.Min(5+float64(c1.R)*(diff+0.6*spec), 255))
	return c, false // no, we do not discard this pixel
}

func main() {
	render := func(shader IShader, filename string) {
		obj = LoadObj(filepath.Join(parentpath, "objs", "african_head.obj"))

		tga := NewTgaImg(width, height)
		zbuffer := NewGreyTgaImage(width, height)

		for i := range obj.NFaces() {
			screenCoords := [3]Vec4f{}
			for j := range 3 {
				screenCoords[j] = shader.Vertex(i, j)
			}
			Triangle10(screenCoords, shader, tga, zbuffer)
		}

		tga.FlipVertically()
		zbuffer.FlipVertically()
		tga.UseRLE = true
		zbuffer.UseRLE = true
		tga.Write(filepath.Join(parentpath, filename))
		zbuffer.Write(filepath.Join(parentpath, "lesson6-zubffer.tga"))
	}

	shaders := []IShader{&GouraudShader1{}, &GouraudShader2{}, &Shader1{}}
	for i, shader := range shaders {
		render(shader, fmt.Sprintf("lesson6-%d.tga", i+1))
	}

	shader2 := Shader2{}
	shader2.uniformM = project.Mul(modelView)
	shader2.uniformMIT = project.Mul(modelView).Inverse().Transpose()
	render(&shader2, fmt.Sprintf("lesson6-%d.tga", 4))

	shader3 := Shader3{}
	shader3.uniformM = project.Mul(modelView)
	shader3.uniformMIT = project.Mul(modelView).Inverse().Transpose()
	render(&shader3, fmt.Sprintf("lesson6-%d.tga", 5))
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
