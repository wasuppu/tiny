package main

import (
	"image"
	"image/color"
	"image/png"
	"math"
	"os"
)

type Material struct {
	diffuseColor Vec3f
}

type Sphere struct {
	center   Vec3f
	radius   float64
	material Material
}

func (s Sphere) rayIntersect(orig, dir Vec3f, t0 *float64) bool {
	l := s.center.Sub(orig)
	tca := l.Dot(dir)
	d2 := l.Dot(l) - tca*tca

	if d2 > s.radius*s.radius {
		return false
	}
	thc := math.Sqrt(s.radius*s.radius - d2)
	*t0 = tca - thc
	t1 := tca + thc

	if *t0 < 0 {
		t0 = &t1
	}
	if *t0 < 0 {
		return false
	}
	return true
}

func sceneIntersect(orig, dir Vec3f, spheres []Sphere, material *Material) bool {
	sphereDist := math.MaxFloat64
	for i := range spheres {
		var disti float64
		if spheres[i].rayIntersect(orig, dir, &disti) && disti < sphereDist {
			sphereDist = disti
			*material = spheres[i].material
		}
	}
	return sphereDist < 1000
}

func castRay(orig, dir Vec3f, spheres []Sphere) Vec3f {
	material := Material{}
	if !sceneIntersect(orig, dir, spheres, &material) {
		return Vec3f{0.2, 0.7, 0.8} // background color
	}
	return material.diffuseColor
}

func render(spheres []Sphere) {
	width := 1024
	height := 768
	fov := int(math.Round(math.Pi)) / 2

	framebuffer := make([]color.Color, width*height)

	for j := range height {
		for i := range width {
			x := (2*(float64(i)+0.5)/float64(width) - 1) * math.Tan(float64(fov)/2) * float64(width) / float64(height)
			y := -(2*(float64(j)+0.5)/float64(height) - 1) * math.Tan(float64(fov)/2)
			dir := Vec3f{x, y, -1}.Normalize()
			framebuffer[i+j*width] = toColor(castRay(Vec3f{0, 0, 0}, dir, spheres))
		}
	}

	writePng("out", framebuffer, width, height)
}

func main() {
	ivory := Material{Vec3f{0.4, 0.4, 0.3}}
	redRubber := Material{Vec3f{0.3, 0.1, 0.1}}

	spheres := []Sphere{}
	spheres = append(spheres, Sphere{Vec3f{-3, 0, -16}, 2, ivory})
	spheres = append(spheres, Sphere{Vec3f{-1.0, -1.5, -12}, 2, redRubber})
	spheres = append(spheres, Sphere{Vec3f{1.5, -0.5, -18}, 3, redRubber})
	spheres = append(spheres, Sphere{Vec3f{7, 5, -18}, 4, ivory})

	render(spheres)
}

func toColor(v Vec3f) color.RGBA {
	return color.RGBA{uint8(255 * v[0]), uint8(255 * v[1]), uint8(255 * v[2]), 0xff}
}

func writePng(name string, pixels []color.Color, width, height int) {
	f, _ := os.Create(name + ".png")
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	for j := range height {
		for i := range width {
			img.Set(i, j, pixels[i+j*width])
		}
	}
	png.Encode(f, img)
}

type Vec3f [3]float64

func (v Vec3f) X() float64 {
	return v[0]
}

func (v Vec3f) Y() float64 {
	return v[1]
}

func (v Vec3f) Z() float64 {
	return v[2]
}

func (v Vec3f) Add(o Vec3f) Vec3f {
	return Vec3f{v[0] + o[0], v[1] + o[1], v[2] + o[2]}
}

func (v Vec3f) Sub(o Vec3f) Vec3f {
	return Vec3f{v[0] - o[0], v[1] - o[1], v[2] - o[2]}
}

func (v Vec3f) Muln(t float64) Vec3f {
	return Vec3f{v[0] * t, v[1] * t, v[2] * t}
}

func (v Vec3f) Dot(o Vec3f) float64 {
	return v[0]*o[0] + v[1]*o[1] + v[2]*o[2]
}

func (v Vec3f) Cross(o Vec3f) Vec3f {
	x := v[1]*o[2] - v[2]*o[1]
	y := v[2]*o[0] - v[0]*o[2]
	z := v[0]*o[1] - v[1]*o[0]

	return Vec3f{x, y, z}
}

func (v Vec3f) Length() float64 {
	return math.Sqrt(float64(v.Dot(v)))
}

func (v Vec3f) Normalize() Vec3f {
	return v.Muln(1 / v.Length())
}
