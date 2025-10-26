package main

import (
	"image/color"
	"math"
	"math/rand"
	"path/filepath"
	"runtime"
)

var (
	white      = color.RGBA{255, 255, 255, 255}
	red        = color.RGBA{255, 0, 0, 255}
	green      = color.RGBA{0, 255, 0, 255}
	blue       = color.RGBA{0, 0, 255, 255}
	basepath   string
	parentpath string
)

func init() {
	_, exepath, _, _ := runtime.Caller(0)
	basepath = filepath.Dir(exepath)
	parentpath = filepath.Dir(basepath)
}

func Line6(p0, p1 Vec2i, tga *TGAImage, c color.Color) {
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

func Rasterize(p0, p1 Vec2i, tga *TGAImage, c color.Color, ybuffer []int) {
	if p0.X() > p1.X() {
		swap(&p0, &p1)
	}

	for x := p0.X(); x <= p1.X(); x++ {
		t := float64(x-p0.X()) / float64(p1.X()-p0.X())
		y := int(float64(p0.Y())*(1-t) + float64(p1.Y())*t)
		if ybuffer[x] < y {
			ybuffer[x] = y
			tga.Set(x, 0, c)
		}
	}
}

func Section1() {
	{
		width := 800
		height := 500
		// just dumping the 2d scene (yay we have enough dimensions!)
		tga := NewTgaImg(width, height)

		// scene "2d mesh"
		Line6(Vec2i{20, 34}, Vec2i{744, 400}, tga, red)
		Line6(Vec2i{120, 434}, Vec2i{444, 400}, tga, green)
		Line6(Vec2i{330, 463}, Vec2i{594, 200}, tga, blue)

		// screen line
		Line6(Vec2i{10, 10}, Vec2i{790, 10}, tga, white)

		tga.FlipVertically()
		tga.UseRLE = true
		tga.Write(filepath.Join(parentpath, "lesson3-1-scene.tga"))
	}

	{
		width := 800
		tga := NewTgaImg(width, 16)

		ybuffer := [800]int{}
		for i := range width {
			ybuffer[i] = math.MinInt
		}

		Rasterize(Vec2i{20, 34}, Vec2i{744, 400}, tga, red, ybuffer[:])
		Rasterize(Vec2i{120, 434}, Vec2i{444, 400}, tga, green, ybuffer[:])
		Rasterize(Vec2i{330, 463}, Vec2i{594, 200}, tga, blue, ybuffer[:])

		// 1-pixel wide image is bad for eyes, lets widen it
		for i := range width {
			for j := 1; j < 16; j++ {
				tga.Set(i, j, tga.At(i, 0))
			}
		}

		tga.FlipVertically() // i want to have the origin at the left bottom corner of the image
		tga.UseRLE = true
		tga.Write(filepath.Join(parentpath, "lesson3-1-render.tga"))
	}
}

func Barycentric2(a, b, c, p Vec3f) Vec3f {
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

func Triangle7(pts [3]Vec3f, zbuffer []float64, tga *TGAImage, c color.Color) {
	bboxmin := Vec2f{math.MaxFloat64, math.MaxFloat64}
	bboxmax := Vec2f{-math.MaxFloat64, -math.MaxFloat64}
	clamp := Vec2f{float64(tga.Width - 1), float64(tga.Height - 1)}
	for i := range 3 {
		for j := range 2 {
			bboxmin[j] = math.Max(0, math.Min(bboxmin[j], pts[i][j]))
			bboxmax[j] = math.Min(clamp[j], math.Max(bboxmax[j], pts[i][j]))
		}
	}

	p := Vec3f{}
	for p[0] = bboxmin.X(); p.X() <= bboxmax.X(); p[0]++ {
		for p[1] = bboxmin.Y(); p.Y() <= bboxmax.Y(); p[1]++ {
			bcScreen := Barycentric2(pts[0], pts[1], pts[2], p)
			if bcScreen.X() < 0 || bcScreen.Y() < 0 || bcScreen.Z() < 0 {
				continue
			}
			p[2] = 0
			for i := range 3 {
				p[2] += pts[i][2] * bcScreen[i]
			}
			if zbuffer[int(p.X()+p.Y()*float64(tga.Width))] < p.Z() {
				zbuffer[int(p.X()+p.Y()*float64(tga.Width))] = p.Z()
				tga.Set(int(p.X()), int(p.Y()), c)
			}
		}
	}
}

func world2screen(v Vec3f, width, height int) Vec3f {
	return Vec3f{float64(int((v.X()+1.0)*float64(width)/2 + 0.5)), float64(int((v.Y()+1.0)*float64(height)/2 + 0.5)), v.Z()}
}

func Section2() {
	width := 800
	height := 800

	tga := NewTgaImg(width, height)
	obj := LoadObj(filepath.Join(parentpath, "objs", "african_head.obj"))

	zbuffer := make([]float64, int(width*height))
	for i := range zbuffer {
		zbuffer[i] = -math.MaxFloat64
	}

	for i := range obj.NFaces() {
		face := obj.Face(i)
		pts := [3]Vec3f{}
		for j := range 3 {
			pts[j] = world2screen(obj.Vert(face[j]), width, height)
		}

		Triangle7(pts, zbuffer, tga, color.RGBA{uint8(rand.Intn(255)), uint8(rand.Intn(255)), uint8(rand.Intn(255)), 255})
	}

	tga.FlipVertically()
	tga.UseRLE = true
	tga.Write(filepath.Join(parentpath, "lesson3-2.tga"))
}

func Section3() {
	width := 800
	height := 800

	tga := NewTgaImg(width, height)
	obj := LoadObj(filepath.Join(parentpath, "objs", "african_head.obj"))

	zbuffer := make([]float64, int(width*height))
	for i := range zbuffer {
		zbuffer[i] = -math.MaxFloat64
	}

	lightDir := Vec3f{0, 0, -1}
	for i := range obj.NFaces() {
		face := obj.Face(i)
		pts := [3]Vec3f{}
		worldCoords := []Vec3f{}
		for j := range 3 {
			v := obj.Vert(face[j])
			pts[j] = world2screen(v, width, height)
			worldCoords = append(worldCoords, v)
		}
		n := worldCoords[2].Sub(worldCoords[0]).Cross(worldCoords[1].Sub(worldCoords[0]))
		n = n.Normalize()
		intensity := n.Dot(lightDir)
		if intensity > 0 {
			Triangle7(pts, zbuffer, tga, color.RGBA{uint8(intensity * 255), uint8(intensity * 255), uint8(intensity * 255), 255})
		}
	}

	tga.FlipVertically()
	tga.UseRLE = true
	tga.Write(filepath.Join(parentpath, "lesson3-3.tga"))
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

func Section4() {
	width := 800
	height := 800
	depth := 255

	obj := LoadObj(filepath.Join(parentpath, "objs", "african_head.obj"))

	zbuffer := make([]int, int(width*height))
	for i := range zbuffer {
		zbuffer[i] = math.MinInt
	}

	{ // draw the model
		tga := NewTgaImg(width, height)
		lightDir := Vec3f{0, 0, -1}
		for i := range obj.NFaces() {
			face := obj.Face(i)
			screenCoords := [3]Vec3i{}
			worldCoords := [3]Vec3f{}
			for j := range 3 {
				v := obj.Vert(face[j])
				screenCoords[j] = Vec3i{int((v.X() + 1.) * float64(width) / 2), int((v.Y() + 1.0) * float64(height) / 2), int((v.Z() + 1.0) * float64(depth) / 2)}
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
		tga.Write(filepath.Join(parentpath, "lesson3-4.tga"))
	}

	{ // dump z-buffer (debugging purposes only)
		tga := NewGreyTgaImage(width, height)
		for i := range width {
			for j := range height {
				tga.Set(i, j, color.Gray{uint8(zbuffer[i+j*width])})
			}
		}
		tga.FlipVertically()
		tga.UseRLE = true
		tga.Write(filepath.Join(parentpath, "lesson3-4-zbuffer.tga"))
	}
}

func main() {
	Section1()
	Section2()
	Section3()
	Section4()
}

func swap[T any](v1 *T, v2 *T) {
	*v1, *v2 = *v2, *v1
}
