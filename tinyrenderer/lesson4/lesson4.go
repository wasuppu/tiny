package main

import (
	"image/color"
	"math"
	"path/filepath"
	"runtime"
)

var (
	white      = color.RGBA{255, 255, 255, 255}
	red        = color.RGBA{255, 0, 0, 255}
	green      = color.RGBA{0, 255, 0, 255}
	yellow     = color.RGBA{255, 255, 0, 255}
	basepath   string
	parentpath string
)

func init() {
	_, exepath, _, _ := runtime.Caller(0)
	basepath = filepath.Dir(exepath)
	parentpath = filepath.Dir(basepath)
}

func m2v(m Matrix) Vec3f {
	return Vec3f{m[0][0] / m[3][0], m[1][0] / m[3][0], m[2][0] / m[3][0]}
}

func v2m(v Vec3f) Matrix {
	m := NewMatrix(4, 1)
	m[0][0] = v.X()
	m[1][0] = v.Y()
	m[2][0] = v.Z()
	m[3][0] = 1
	return m
}

func Line7(p0, p1 Vec3i, tga *TGAImage, c color.Color) {
	steep := false
	if math.Abs(float64(p0.X()-p1.X())) < math.Abs(float64(p0.Y()-p1.Y())) {
		swap(&p0[0], &p0[1])
		swap(&p1[0], &p1[1])
		steep = true
	}
	if p0.X() > p1.X() {
		swap(&p0, &p1)
	}

	for x := p0.X(); x <= p1.X(); x++ {
		t := float64(x-p0.X()) / float64(p1.X()-p0.X())
		y := int(float64(p0.Y())*(1-t) + float64(p1.Y())*t + 0.5)
		if steep {
			tga.Set(y, x, c)
		} else {
			tga.Set(x, y, c)
		}
	}
}

func Section1() {
	width := 100
	height := 100

	obj := LoadObj(filepath.Join(parentpath, "objs", "cube.obj"))
	tga := NewTgaImg(width, height)

	vp := Viewport(width/4, width/4, width/2, height/2).M()

	{ // draw the axes
		x := Vec3f{1, 0, 0}
		y := Vec3f{0, 1, 0}
		o := Vec3f{0, 0, 0}

		o = m2v(vp.Mul(v2m(o)))
		x = m2v(vp.Mul(v2m(x)))
		y = m2v(vp.Mul(v2m(y)))

		Line7(o.I(), x.I(), tga, red)
		Line7(o.I(), y.I(), tga, green)
	}

	for i := range obj.NFaces() {
		face := obj.Face(i)
		for j := range len(face) {
			wp0 := obj.Vert(face[j])
			wp1 := obj.Vert(face[(j+1)%len(face)])
			{ // draw the original model
				sp0 := m2v(vp.Mul(v2m(wp0)))
				sp1 := m2v(vp.Mul(v2m(wp1)))
				Line7(sp0.I(), sp1.I(), tga, white)
			}
			{ // draw the deformed model
				t := Zoom(1.5).M()
				// t := Identity(4)
				// t[0][1] = 0.333
				// t := Translate(Vec3f{.33, .5, 0}).Transpose().Mul(RotationZ(Radians(10))).M()
				sp0 := m2v(vp.Mul(t).Mul(v2m(wp0)))
				sp1 := m2v(vp.Mul(t).Mul(v2m(wp1)))
				Line7(sp0.I(), sp1.I(), tga, yellow)
			}
		}
		break
	}

	tga.FlipVertically()
	tga.UseRLE = true
	tga.Write(filepath.Join(parentpath, "lesson4-1.tga"))
}

func Triangle8(t0, t1, t2 Vec3i, uv0, uv1, uv2 Vec2i, tga *TGAImage, intensity float64, zbuffer []int, obj *Obj) {
	if t0.Y() == t1.Y() && t0.Y() == t2.Y() { // I dont care about degenerate triangles
		return
	}
	if t0.Y() > t1.Y() {
		swap(&t0, &t1)
		swap(&uv0, &uv1)
	}
	if t0.Y() > t2.Y() {
		swap(&t0, &t2)
		swap(&uv0, &uv2)
	}
	if t1.Y() > t2.Y() {
		swap(&t1, &t2)
		swap(&uv1, &uv2)
	}

	totalHeight := t2.Y() - t0.Y()
	for i := range totalHeight {
		secondHalf := i > t1.Y()-t0.Y() || t1.Y() == t0.Y()
		segmentHeight := t1.Y() - t0.Y()
		if secondHalf {
			segmentHeight = t2.Y() - t1.Y()
		}

		alpha := float64(i) / float64(totalHeight)
		beta := float64(i) / float64(segmentHeight) // be careful: with above conditions no division by zero here
		if secondHalf {
			beta = float64(i-(t1.Y()-t0.Y())) / float64(segmentHeight)
		}

		a := t0.Add(t2.Sub(t0).F().Muln(alpha).I())
		b := t0.Add(t1.Sub(t0).F().Muln(beta).I())
		if secondHalf {
			b = t1.Add(t2.Sub(t1).F().Muln(beta).I())
		}

		uva := uv0.Add(uv2.Sub(uv0).Muln(alpha))
		uvb := uv0.Add(uv1.Sub(uv0).Muln(beta))
		if secondHalf {
			uvb = uv1.Add(uv2.Sub(uv1).Muln(beta))
		}

		if a.X() > b.X() {
			swap(&a, &b)
			swap(&uva, &uvb)
		}

		for j := a.X(); j <= b.X(); j++ {
			phi := float64(j-a.X()) / float64(b.X()-a.X())
			if b.X() == a.X() {
				phi = 1
			}
			p := a.F().Add(b.Sub(a).F().Muln(phi)).I()
			uvp := uva.Add(uvb.Sub(uva).Muln(phi))
			idx := p.X() + p.Y()*tga.Width

			if zbuffer[idx] < p.Z() {
				zbuffer[idx] = p.Z()
				c := obj.Diffuse(uvp)
				tga.Set(p.X(), p.Y(), color.RGBA{uint8(float64(c.R) * intensity), uint8(float64(c.G) * intensity), uint8(float64(c.B) * intensity), 255})
			}
		}
	}
}

func Section2() {
	width := 800
	height := 800

	lightDir := Vec3f{0, 0, -1}
	camera := Vec3f{0, 0, 3}

	obj := LoadObj(filepath.Join(parentpath, "objs", "african_head.obj"))

	zbuffer := make([]int, int(width*height))
	for i := range zbuffer {
		zbuffer[i] = math.MinInt
	}

	{
		tga := NewTgaImg(width, height)

		projection := ID4()
		projection[3][2] = -1 / camera.Z()
		viewport := Viewport(width/8, height/8, width*3/4, height*3/4)

		for i := range obj.NFaces() {
			face := obj.Face(i)
			screenCoords := [3]Vec3i{}
			worldCoords := [3]Vec3f{}
			for j := range 3 {
				v := obj.Vert(face[j])
				screenCoords[j] = m2v(viewport.Mul(projection).M().Mul(v2m(v))).I()
				worldCoords[j] = v
			}
			n := worldCoords[2].Sub(worldCoords[0]).Cross(worldCoords[1].Sub(worldCoords[0]))
			n = n.Normalize()
			intensity := n.Dot(lightDir)
			if intensity > 0 {
				uv := [3]Vec2i{}
				for k := range 3 {
					uv[k] = obj.UV(i, k)
				}
				Triangle8(screenCoords[0], screenCoords[1], screenCoords[2], uv[0], uv[1], uv[2], tga, intensity, zbuffer, obj)
			}
		}
		tga.FlipVertically()
		tga.UseRLE = true
		tga.Write(filepath.Join(parentpath, "lesson4-2.tga"))
	}

	{
		tga := NewGreyTgaImage(width, height)
		for i := range width {
			for j := range height {
				tga.Set(i, j, color.Gray{uint8(zbuffer[i+j*width])})
			}
		}
		tga.FlipVertically()
		tga.UseRLE = true
		tga.Write(filepath.Join(parentpath, "lesson4-2-zbuffer.tga"))
	}

}

func main() {
	Section1()
	Section2()
}

func swap[T any](v1 *T, v2 *T) {
	*v1, *v2 = *v2, *v1
}
