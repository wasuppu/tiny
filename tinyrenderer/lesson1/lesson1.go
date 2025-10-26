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
	basepath   string
	parentpath string
)

func init() {
	_, exepath, _, _ := runtime.Caller(0)
	basepath = filepath.Dir(exepath)
	parentpath = filepath.Dir(basepath)
}

func Line1(x0, y0, x1, y1 int, tga *TGAImage, c color.Color) {
	for t := .0; t < 1.; t += .01 {
		x := int(float64(x0) + float64(x1-x0)*t)
		y := int(float64(y0) + float64(y1-y0)*t)
		tga.Set(x, y, c)
	}
}

func Section1() {
	tga := NewTgaImg(100, 100)
	Line1(13, 20, 80, 40, tga, white)

	tga.FlipVertically()
	tga.UseRLE = true
	tga.Write(filepath.Join(parentpath, "lesson1-1.tga"))
}

func Line2(x0, y0, x1, y1 int, tga *TGAImage, c color.Color) {
	for x := x0; x <= x1; x++ {
		t := float64(x-x0) / float64(x1-x0)
		y := int(float64(y0)*(1.-t) + float64(y1)*t)
		tga.Set(x, y, c)
	}
}

func Section2() {
	tga := NewTgaImg(100, 100)
	Line2(13, 20, 80, 40, tga, white)
	Line2(20, 13, 40, 80, tga, red)
	Line2(80, 40, 13, 20, tga, red)

	tga.FlipVertically()
	tga.UseRLE = true
	tga.Write(filepath.Join(parentpath, "lesson1-2.tga"))
}

func Line3(x0, y0, x1, y1 int, tga *TGAImage, c color.Color) {
	steep := false
	if math.Abs(float64(x0-x1)) < math.Abs(float64(y0-y1)) {
		x0, y0 = y0, x0
		x1, y1 = y1, x1
		steep = true
	}
	if x0 > x1 {
		x0, x1 = x1, x0
		y0, y1 = y1, y0
	}

	for x := x0; x <= x1; x++ {
		t := float64(x-x0) / float64(x1-x0)
		y := int(float64(y0)*(1-t) + float64(y1)*t)
		if steep {
			tga.Set(y, x, c)
		} else {
			tga.Set(x, y, c)
		}
	}
}

func Section3() {
	tga := NewTgaImg(100, 100)

	Line3(13, 20, 80, 40, tga, white)
	Line3(20, 13, 40, 80, tga, red)
	Line3(80, 40, 13, 20, tga, red)

	tga.FlipVertically()
	tga.UseRLE = true
	tga.Write(filepath.Join(parentpath, "lesson1-3.tga"))
}

func Line4(x0, y0, x1, y1 int, tga *TGAImage, c color.Color) {
	steep := false
	if math.Abs(float64(x0-x1)) < math.Abs(float64(y0-y1)) {
		x0, y0 = y0, x0
		x1, y1 = y1, x1
		steep = true
	}
	if x0 > x1 {
		x0, x1 = x1, x0
		y0, y1 = y1, y0
	}

	dx := x1 - x0
	dy := y1 - y0
	derr := math.Abs(float64(dy) / float64(dx))
	err := 0.0
	y := y0
	for x := x0; x <= x1; x++ {
		if steep {
			tga.Set(y, x, c)
		} else {
			tga.Set(x, y, c)
		}

		err += derr
		if err > .5 {
			if y1 > y0 {
				y += 1
			} else {
				y -= 1
			}
			err -= 1.
		}
	}
}

func Section4() {
	tga := NewTgaImg(100, 100)

	Line4(13, 20, 80, 40, tga, white)
	Line4(20, 13, 40, 80, tga, red)
	Line4(80, 40, 13, 20, tga, red)

	tga.FlipVertically()
	tga.UseRLE = true
	tga.Write(filepath.Join(parentpath, "lesson1-4.tga"))
}

func Line5(x0, y0, x1, y1 int, tga *TGAImage, c color.Color) {
	steep := false
	if math.Abs(float64(x0-x1)) < math.Abs(float64(y0-y1)) {
		x0, y0 = y0, x0
		x1, y1 = y1, x1
		steep = true
	}
	if x0 > x1 {
		x0, x1 = x1, x0
		y0, y1 = y1, y0
	}

	dx := x1 - x0
	dy := y1 - y0
	derr2 := math.Abs(float64(dy)) * 2
	err2 := 0.0
	y := y0
	for x := x0; x <= x1; x++ {
		if steep {
			tga.Set(y, x, c)
		} else {
			tga.Set(x, y, c)
		}

		err2 += derr2
		if err2 > float64(dx) {
			if y1 > y0 {
				y += 1
			} else {
				y -= 1
			}
			err2 -= float64(dx) * 2
		}
	}
}

func Section5() {
	tga := NewTgaImg(100, 100)

	Line5(13, 20, 80, 40, tga, white)
	Line5(20, 13, 40, 80, tga, red)
	Line5(80, 40, 13, 20, tga, red)

	tga.FlipVertically()
	tga.UseRLE = true
	tga.Write(filepath.Join(parentpath, "lesson1-5.tga"))
}

func Section6() {
	width := 800
	height := 800

	tga := NewTgaImg(width, height)
	obj := LoadObj(filepath.Join(parentpath, "objs", "african_head.obj"))

	for i := range obj.NFaces() {
		face := obj.Face(i)
		for j := range 3 {
			v0 := obj.Vert(face[j])
			v1 := obj.Vert(face[(j+1)%3])
			x0 := int((v0.X() + 1.) * float64(width) / 2)
			y0 := int((v0.Y() + 1.) * float64(height) / 2)
			x1 := int((v1.X() + 1.) * float64(width) / 2)
			y1 := int((v1.Y() + 1.) * float64(height) / 2)
			Line3(x0, y0, x1, y1, tga, white)
		}
	}

	tga.FlipVertically()
	tga.UseRLE = true
	tga.Write(filepath.Join(parentpath, "lesson1-6.tga"))
}

func main() {
	Section1()
	Section2()
	Section3()
	Section4()
	Section5()
	Section6()
}
