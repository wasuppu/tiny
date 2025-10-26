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
	blue       = color.RGBA{0, 0, 255, 255}
	yellow     = color.RGBA{255, 255, 0, 255}
	basepath   string
	parentpath string
)

func init() {
	_, exepath, _, _ := runtime.Caller(0)
	basepath = filepath.Dir(exepath)
	parentpath = filepath.Dir(basepath)
}

func Triangle9(t0, t1, t2 Vec3i, ity0, ity1, ity2 float64, tga *TGAImage, zbuffer []int) {
	if t0.Y() == t1.Y() && t0.Y() == t2.Y() { // I dont care about degenerate triangles
		return
	}
	if t0.Y() > t1.Y() {
		swap(&t0, &t1)
		swap(&ity0, &ity1)
	}
	if t0.Y() > t2.Y() {
		swap(&t0, &t2)
		swap(&ity0, &ity2)
	}
	if t1.Y() > t2.Y() {
		swap(&t1, &t2)
		swap(&ity1, &ity2)
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

		itya := ity0 + (ity2-ity0)*alpha
		ityb := ity0 + (ity1-ity0)*beta
		if secondHalf {
			ityb = ity1 + (ity2-ity1)*beta
		}

		if a.X() > b.X() {
			swap(&a, &b)
			swap(&itya, &ityb)
		}

		for j := a.X(); j <= b.X(); j++ {
			phi := float64(j-a.X()) / float64(b.X()-a.X())
			if b.X() == a.X() {
				phi = 1
			}
			p := a.F().Add(b.Sub(a).F().Muln(phi)).I()
			ityp := itya + (ityb-itya)*phi
			idx := p.X() + p.Y()*tga.Width

			if p.X() >= tga.Width || p.Y() >= tga.Height || p.X() < 0 || p.Y() < 0 {
				continue
			}

			if zbuffer[idx] < p.Z() {
				zbuffer[idx] = p.Z()
				tga.Set(p.X(), p.Y(), ScaleColorRGB(color.RGBA{255, 255, 255, 255}, ityp))
			}
		}
	}
}

func main() {
	width := 800
	height := 800

	lightDir := Vec3f{1, -1, 1}.Normalize()
	eye := Vec3f{1, 1, 3}
	center := Vec3f{0, 0, 0}

	obj := LoadObj(filepath.Join(parentpath, "objs", "african_head.obj"))

	zbuffer := make([]int, int(width*height))
	for i := range zbuffer {
		zbuffer[i] = math.MinInt
	}

	{ // draw the model
		modelView := LookAt(eye, center, Vec3f{0, 1, 0})
		projection := ID4()
		projection[3][2] = -1 / eye.Sub(center).Length()
		viewport := Viewport(width/8, height/8, width*3/4, height*3/4)

		tga := NewTgaImg(width, height)
		for i := range obj.NFaces() {
			face := obj.Face(i)
			screenCoords := [3]Vec3i{}
			intensity := [3]float64{}
			for j := range 3 {
				v := obj.Vert(face[j])
				screenCoords[j] = viewport.Mul(projection).Mul(modelView).M().Mul(v.M()).V().I()
				intensity[j] = obj.Norm(i, j).Dot(lightDir)
			}

			Triangle9(screenCoords[0], screenCoords[1], screenCoords[2], intensity[0], intensity[1], intensity[2], tga, zbuffer)
		}

		tga.FlipVertically() // i want to have the origin at the left bottom corner of the image
		tga.UseRLE = true
		tga.Write(filepath.Join(parentpath, "lesson5.tga"))
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
		tga.Write(filepath.Join(parentpath, "lesson5-zubffer.tga"))
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

func swap[T any](v1 *T, v2 *T) {
	*v1, *v2 = *v2, *v1
}
