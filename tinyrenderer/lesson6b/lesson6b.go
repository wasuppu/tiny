package main

import (
	"fmt"
	"image/color"
	"math"
	"os"
	"path/filepath"
	"runtime"
)

var (
	width      = 800
	height     = 800
	obj        *Obj
	lightDir   = Vec3f{1, 1, 1}
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

func Triangle11(clipc Mat4x3, shader IShader, img *TGAImage, zbuffer []float64) {
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

			c, discard := shader.Fragment(bcClip)
			if !discard {
				zbuffer[p.X()+p.Y()*img.Width] = fragDepth
				img.Set(p.X(), p.Y(), c)
			}
		}
	}
}

type IShader interface {
	Vertex(iface, nthvert int) Vec4f
	Fragment(Vec3f) (color.Color, bool)
	GetVaryingTri() Mat4x3
}

type Shader4 struct {
	veryingUV  Mat2x3 // triangle uv coordinates, written by the vertex shader, read by the fragment shader
	varyingTri Mat4x3 // triangle coordinates (clip coordinates), written by VS, read by FS
	veryingNrm Mat3   // normal per vertex to be interpolated by FS
}

func (s *Shader4) Vertex(iface, nthvert int) Vec4f {
	s.veryingUV.SetCol(nthvert, obj.UVf(iface, nthvert))
	s.veryingNrm.SetCol(nthvert, project.Mul(modelView).Inverse().Transpose().Mulv(obj.Norm(iface, nthvert).V4(0)).V3())
	glVertex := project.Mul(modelView).Mulv(obj.Vertf(iface, nthvert).V4(1)) // read the vertex from .obj file
	s.varyingTri.SetCol(nthvert, glVertex)
	return glVertex
}

func (s Shader4) Fragment(bar Vec3f) (color.Color, bool) {
	bn := s.veryingNrm.Mulv(bar).Normalize()
	uv := s.veryingUV.Mulv(bar)
	diff := math.Max(0, bn.Dot(lightDir))
	c := ScaleColorRGB(obj.Diffusef(uv), diff) // well duh
	return c, false                            // no, we do not discard this pixel
}

func (s Shader4) GetVaryingTri() Mat4x3 {
	return s.varyingTri
}

type Shader5 struct {
	veryingUV  Mat2x3 // triangle uv coordinates, written by the vertex shader, read by the fragment shader
	varyingTri Mat4x3 // triangle coordinates (clip coordinates), written by VS, read by FS
	veryingNrm Mat3   // normal per vertex to be interpolated by FS
	ndcTri     Mat3   // triangle in normalized device coordinates
}

func (s *Shader5) Vertex(iface, nthvert int) Vec4f {
	s.veryingUV.SetCol(nthvert, obj.UVf(iface, nthvert))
	s.veryingNrm.SetCol(nthvert, project.Mul(modelView).Inverse().Transpose().Mulv(obj.Norm(iface, nthvert).V4(0)).V3())
	glVertex := project.Mul(modelView).Mulv(obj.Vertf(iface, nthvert).V4(1)) // read the vertex from .obj file
	s.varyingTri.SetCol(nthvert, glVertex)
	s.ndcTri.SetCol(nthvert, glVertex.Divn(glVertex[3]).V3())
	return glVertex
}

func (s Shader5) Fragment(bar Vec3f) (color.Color, bool) {
	bn := s.veryingNrm.Mulv(bar).Normalize()
	uv := s.veryingUV.Mulv(bar)

	a := Mat3{}
	a[0] = s.ndcTri.Col(1).Sub(s.ndcTri.Col(0))
	a[1] = s.ndcTri.Col(2).Sub(s.ndcTri.Col(0))
	a[2] = bn

	ai := a.Inverse()

	i := ai.Mulv(Vec3f{s.veryingUV[0][1] - s.veryingUV[0][0], s.veryingUV[0][2] - s.veryingUV[0][0], 0})
	j := ai.Mulv(Vec3f{s.veryingUV[1][1] - s.veryingUV[1][0], s.veryingUV[1][2] - s.veryingUV[1][0], 0})

	b := Mat3{}
	b.SetCol(0, i.Normalize())
	b.SetCol(1, j.Normalize())
	b.SetCol(2, bn)

	n := b.Mulv(obj.NormUV(uv)).Normalize()

	diff := math.Max(0, n.Dot(lightDir))

	c := ScaleColorRGB(obj.Diffusef(uv), diff) // well duh
	return c, false                            // no, we do not discard this pixel
}

func (s Shader5) GetVaryingTri() Mat4x3 {
	return s.varyingTri
}

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "Usage: go run ./%s obj/model.obj\n", filepath.Base(os.Args[0]))
		os.Exit(1)
	}

	render := func(shader IShader, filename string) {
		zbuffer := make([]float64, int(width*height))
		for i := range zbuffer {
			zbuffer[i] = -math.MaxFloat64
		}

		tga := NewTgaImg(width, height)
		lightDir = project.Mul(modelView).Mulv(lightDir.V4(0)).V3().Normalize()

		for i := 1; i < len(os.Args); i++ {
			obj = LoadObj(os.Args[i])

			for i := range obj.NFaces() {
				for j := range 3 {
					shader.Vertex(i, j)
				}
				Triangle11(shader.GetVaryingTri(), shader, tga, zbuffer)
			}
		}

		tga.FlipVertically()
		tga.UseRLE = true
		tga.Write(filepath.Join(parentpath, filename))
	}

	shaders := []IShader{&Shader4{}, &Shader5{}}
	for i, shader := range shaders {
		render(shader, fmt.Sprintf("lesson6b-%d.tga", i+1))
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
