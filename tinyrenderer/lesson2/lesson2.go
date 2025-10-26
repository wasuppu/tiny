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

func Triangle1(t0, t1, t2 Vec2i, tga *TGAImage, c color.Color) {
	Line6(t0, t1, tga, c)
	Line6(t1, t2, tga, c)
	Line6(t2, t0, tga, c)
}

func Section1() {
	width := 200
	height := 200

	tga := NewTgaImg(width, height)
	t0 := [3]Vec2i{{10, 70}, {50, 160}, {70, 80}}
	t1 := [3]Vec2i{{180, 50}, {150, 1}, {70, 180}}
	t2 := [3]Vec2i{{180, 150}, {120, 160}, {130, 180}}
	Triangle1(t0[0], t0[1], t0[2], tga, red)
	Triangle1(t1[0], t1[1], t1[2], tga, white)
	Triangle1(t2[0], t2[1], t2[2], tga, green)

	tga.FlipVertically()
	tga.UseRLE = true
	tga.Write(filepath.Join(parentpath, "lesson2-1.tga"))
}

func Triangle2(t0, t1, t2 Vec2i, tga *TGAImage, c color.Color) {
	// sort the vertices, t0, t1, t2 lower−to−upper (bubblesort yay!)
	if t0.Y() > t1.Y() {
		swap(&t0, &t1)
	}
	if t0.Y() > t2.Y() {
		swap(&t0, &t2)
	}
	if t1.Y() > t2.Y() {
		swap(&t1, &t2)
	}

	Line6(t0, t1, tga, green)
	Line6(t1, t2, tga, green)
	Line6(t2, t0, tga, red)
}

func Section2() {
	width := 200
	height := 200

	tga := NewTgaImg(width, height)
	t0 := [3]Vec2i{{10, 70}, {50, 160}, {70, 80}}
	t1 := [3]Vec2i{{180, 50}, {150, 1}, {70, 180}}
	t2 := [3]Vec2i{{180, 150}, {120, 160}, {130, 180}}
	Triangle2(t0[0], t0[1], t0[2], tga, red)
	Triangle2(t1[0], t1[1], t1[2], tga, white)
	Triangle2(t2[0], t2[1], t2[2], tga, green)

	tga.FlipVertically()
	tga.UseRLE = true
	tga.Write(filepath.Join(parentpath, "lesson2-2.tga"))
}

func Triangle3(t0, t1, t2 Vec2i, tga *TGAImage, c color.Color) {
	// sort the vertices, t0, t1, t2 lower−to−upper (bubblesort yay!)
	if t0.Y() > t1.Y() {
		swap(&t0, &t1)
	}
	if t0.Y() > t2.Y() {
		swap(&t0, &t2)
	}
	if t1.Y() > t2.Y() {
		swap(&t1, &t2)
	}
	totalHeight := t2.Y() - t0.Y()

	for y := t0.Y(); y <= t1.Y(); y++ {
		segmentHeight := t1.Y() - t0.Y() + 1
		alpha := float64(y-t0.Y()) / float64(totalHeight)
		beta := float64(y-t0.Y()) / float64(segmentHeight) // be careful with divisions by zero
		a := t0.Add(t2.Sub(t0).Muln(alpha))
		b := t0.Add(t1.Sub(t0).Muln(beta))
		tga.Set(a.X(), y, red)
		tga.Set(b.X(), y, green)
	}
}

func Section3() {
	width := 200
	height := 200

	tga := NewTgaImg(width, height)
	t0 := [3]Vec2i{{10, 70}, {50, 160}, {70, 80}}
	t1 := [3]Vec2i{{180, 50}, {150, 1}, {70, 180}}
	t2 := [3]Vec2i{{180, 150}, {120, 160}, {130, 180}}
	Triangle3(t0[0], t0[1], t0[2], tga, red)
	Triangle3(t1[0], t1[1], t1[2], tga, white)
	Triangle3(t2[0], t2[1], t2[2], tga, green)

	tga.FlipVertically()
	tga.UseRLE = true
	tga.Write(filepath.Join(parentpath, "lesson2-3.tga"))
}

func Triangle4(t0, t1, t2 Vec2i, tga *TGAImage, c color.Color) {
	// sort the vertices, t0, t1, t2 lower−to−upper (bubblesort yay!)
	if t0.Y() > t1.Y() {
		swap(&t0, &t1)
	}
	if t0.Y() > t2.Y() {
		swap(&t0, &t2)
	}
	if t1.Y() > t2.Y() {
		swap(&t1, &t2)
	}
	totalHeight := t2.Y() - t0.Y()
	for y := t0.Y(); y <= t1.Y(); y++ {
		segmentHeight := t1.Y() - t0.Y() + 1
		alpha := float64(y-t0.Y()) / float64(totalHeight)
		beta := float64(y-t0.Y()) / float64(segmentHeight) // be careful with divisions by zero
		a := t0.Add(t2.Sub(t0).Muln(alpha))
		b := t0.Add(t1.Sub(t0).Muln(beta))
		if a.X() > b.X() {
			swap(&a, &b)
		}
		for j := a.X(); j <= b.X(); j++ {
			tga.Set(j, y, c) // attention, due to int casts t0.y+i != A.y
		}
	}
	for y := t1.Y(); y <= t2.Y(); y++ {
		segmentHeight := t2.Y() - t1.Y() + 1
		alpha := float64(y-t0.Y()) / float64(totalHeight)
		beta := float64(y-t1.Y()) / float64(segmentHeight) // be careful with divisions by zero
		a := t0.Add(t2.Sub(t0).Muln(alpha))
		b := t1.Add(t2.Sub(t1).Muln(beta))
		if a.X() > b.X() {
			swap(&a, &b)
		}
		for j := a.X(); j <= b.X(); j++ {
			tga.Set(j, y, c) // attention, due to int casts t0.y+i != A.y
		}
	}
}

func Section4() {
	width := 200
	height := 200

	tga := NewTgaImg(width, height)
	t0 := [3]Vec2i{{10, 70}, {50, 160}, {70, 80}}
	t1 := [3]Vec2i{{180, 50}, {150, 1}, {70, 180}}
	t2 := [3]Vec2i{{180, 150}, {120, 160}, {130, 180}}
	Triangle4(t0[0], t0[1], t0[2], tga, red)
	Triangle4(t1[0], t1[1], t1[2], tga, white)
	Triangle4(t2[0], t2[1], t2[2], tga, green)

	tga.FlipVertically()
	tga.UseRLE = true
	tga.Write(filepath.Join(parentpath, "lesson2-4.tga"))
}

func Triangle5(t0, t1, t2 Vec2i, tga *TGAImage, c color.Color) {
	if t0.Y() == t1.Y() && t0.Y() == t2.Y() { // I dont care about degenerate triangles
		return
	}
	if t0.Y() > t1.Y() {
		swap(&t0, &t1)
	}
	if t0.Y() > t2.Y() {
		swap(&t0, &t2)
	}
	if t1.Y() > t2.Y() {
		swap(&t1, &t2)
	}
	totalHeight := t2.Y() - t0.Y()
	for i := range totalHeight {
		secondHalf := i > t1.Y()-t0.Y() || t1.Y() == t0.Y()
		var segmentHeight int
		if secondHalf {
			segmentHeight = t2.Y() - t1.Y()
		} else {
			segmentHeight = t1.Y() - t0.Y()
		}
		alpha := float64(i) / float64(totalHeight)
		var beta float64 // be careful: with above conditions no division by zero here
		if secondHalf {
			beta = float64(i-(t1.Y()-t0.Y())) / float64(segmentHeight)
		} else {
			beta = float64(i) / float64(segmentHeight)
		}

		a := t0.Add(t2.Sub(t0).Muln(alpha))
		var b Vec2i
		if secondHalf {
			b = t1.Add(t2.Sub(t1).Muln(beta))
		} else {
			b = t0.Add(t1.Sub(t0).Muln(beta))
		}
		if a.X() > b.X() {
			swap(&a, &b)
		}
		for j := a.X(); j <= b.X(); j++ {
			tga.Set(j, t0.Y()+i, c) // attention, due to int casts t0.y+i != A.y
		}
	}
}

func Section5() {
	width := 200
	height := 200

	tga := NewTgaImg(width, height)
	t0 := [3]Vec2i{{10, 70}, {50, 160}, {70, 80}}
	t1 := [3]Vec2i{{180, 50}, {150, 1}, {70, 180}}
	t2 := [3]Vec2i{{180, 150}, {120, 160}, {130, 180}}
	Triangle5(t0[0], t0[1], t0[2], tga, red)
	Triangle5(t1[0], t1[1], t1[2], tga, white)
	Triangle5(t2[0], t2[1], t2[2], tga, green)

	tga.FlipVertically()
	tga.UseRLE = true
	tga.Write(filepath.Join(parentpath, "lesson2-5.tga"))
}

func Barycentric1(pts [3]Vec2i, p Vec2i) Vec3f {
	u := Vec3f{float64(pts[2][0] - pts[0][0]), float64(pts[1][0] - pts[0][0]), float64(pts[0][0] - p[0])}.Cross(Vec3f{float64(pts[2][1] - pts[0][1]), float64(pts[1][1] - pts[0][1]), float64(pts[0][1] - p[1])})
	/* `pts` and `P` has integer value as coordinates
	   so `abs(u[2])` < 1 means `u[2]` is 0, that means
	   triangle is degenerate, in this case return something with negative coordinates */
	if math.Abs(u.Z()) < 1 {
		return Vec3f{-1, 1, 1}
	}

	return Vec3f{1 - (u.X()+u.Y())/u.Z(), u.Y() / u.Z(), u.X() / u.Z()}
}

func Triangle6(pts [3]Vec2i, tga *TGAImage, c color.Color) {
	bboxmin := Vec2i{tga.Width - 1, tga.Height - 1}
	bboxmax := Vec2i{0, 0}
	clamp := Vec2i{tga.Width - 1, tga.Height - 1}

	for i := range 3 {
		bboxmin[0] = int(math.Max(0, math.Min(float64(bboxmin.X()), float64(pts[i].X()))))
		bboxmin[1] = int(math.Max(0, math.Min(float64(bboxmin.Y()), float64(pts[i].Y()))))

		bboxmax[0] = int(math.Min(float64(clamp.X()), math.Max(float64(bboxmax.X()), float64(pts[i].X()))))
		bboxmax[1] = int(math.Min(float64(clamp.Y()), math.Max(float64(bboxmax.Y()), float64(pts[i].Y()))))
	}

	p := Vec2i{}
	for p[0] = bboxmin.X(); p.X() <= bboxmax.X(); p[0]++ {
		for p[1] = bboxmin.Y(); p.Y() <= bboxmax.Y(); p[1]++ {
			bcScreen := Barycentric1(pts, p)
			if bcScreen.X() < 0 || bcScreen.Y() < 0 || bcScreen.Z() < 0 {
				continue
			}
			tga.Set(p.X(), p.Y(), c)
		}
	}
}

func Section6() {
	width := 200
	height := 200

	tga := NewTgaImg(width, height)
	pts := [3]Vec2i{{10, 10}, {100, 30}, {190, 160}}
	Triangle6(pts, tga, red)

	tga.FlipVertically()
	tga.UseRLE = true
	tga.Write(filepath.Join(parentpath, "lesson2-6.tga"))
}

func Section7() {
	width := 800
	height := 800

	tga := NewTgaImg(width, height)
	obj := LoadObj(filepath.Join(parentpath, "objs", "african_head.obj"))

	for i := range obj.NFaces() {
		face := obj.Face(i)
		screenCoords := [3]Vec2i{}
		for j := range 3 {
			worldCoords := obj.Vert(face[j])
			screenCoords[j] = Vec2i{int((worldCoords.X() + 1) * float64(width) / 2), int((worldCoords.Y() + 1) * float64(height) / 2)}
		}
		Triangle5(screenCoords[0], screenCoords[1], screenCoords[2], tga, color.RGBA{uint8(rand.Intn(255)), uint8(rand.Intn(255)), uint8(rand.Intn(255)), 255})
	}

	tga.FlipVertically()
	tga.UseRLE = true
	tga.Write(filepath.Join(parentpath, "lesson2-7.tga"))
}

func Section8() {
	width := 800
	height := 800

	tga := NewTgaImg(width, height)
	obj := LoadObj(filepath.Join(parentpath, "objs", "african_head.obj"))

	lightDir := Vec3f{0, 0, -1}
	for i := range obj.NFaces() {
		face := obj.Face(i)
		screenCoords := [3]Vec2i{}
		worldCoords := []Vec3f{}
		for j := range 3 {
			v := obj.Vert(face[j])
			screenCoords[j] = Vec2i{int((v.X() + 1) * float64(width) / 2), int((v.Y() + 1) * float64(height) / 2)}
			worldCoords = append(worldCoords, v)
		}
		n := worldCoords[2].Sub(worldCoords[0]).Cross(worldCoords[1].Sub(worldCoords[0]))
		n = n.Normalize()
		intensity := n.Dot(lightDir)
		if intensity > 0 {
			Triangle5(screenCoords[0], screenCoords[1], screenCoords[2], tga, color.RGBA{uint8(intensity * 255), uint8(intensity * 255), uint8(intensity * 255), 255})
		}
	}

	tga.FlipVertically()
	tga.UseRLE = true
	tga.Write(filepath.Join(parentpath, "lesson2-8.tga"))
}

func main() {
	Section1()
	Section2()
	Section3()
	Section4()
	Section5()
	Section6()
	Section7()
	Section8()
}

func swap[T any](v1 *T, v2 *T) {
	*v1, *v2 = *v2, *v1
}
