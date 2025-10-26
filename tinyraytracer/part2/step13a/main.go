package main

import (
	"image"
	"image/color"
	_ "image/jpeg"
	"image/png"
	"math"
	"os"
)

type Sphere struct {
	center Vec3f
	radius float64
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
		*t0 = t1
	}
	if *t0 < 0 {
		return false
	}
	return true
}

func sceneIntersect(orig, dir Vec3f, spheres []Sphere) float64 {
	sphereDist := math.MaxFloat64
	for i := range spheres {
		var disti float64
		if spheres[i].rayIntersect(orig, dir, &disti) && disti < sphereDist {
			sphereDist = disti
		}
	}

	checkerboardDist := math.MaxFloat64
	// intersect the ray with the checkerboard, avoid division by zero
	if math.Abs(dir.Y()) > 1e-3 {
		// the checkerboard plane has equation y = -4
		d := -(orig.Y() + 4) / dir.Y()
		p := orig.Add(dir.Muln(d))
		if d > 0 && math.Abs(p.X()) < 10 && p.Z() < -10 && p.Z() > -30 && d < sphereDist {
			checkerboardDist = d
		}
	}

	return math.Min(sphereDist, checkerboardDist)
}

func computeDepthmap(width, height int, fov, far float64, spheres []Sphere, zbuffer []float64) {
	for j := range height {
		for i := range width {
			x := (float64(i) + 0.5) - float64(width)/2
			y := -(float64(j) + 0.5) + float64(height)/2
			z := -float64(height) / (2 * math.Tan(fov/2.))
			dir := Vec3f{x, y, z}.Normalize()
			zbuffer[i+j*width] = sceneIntersect(Vec3f{0, 0, 0}, dir, spheres)
		}
	}

	min := math.MaxFloat64
	max := -math.MaxFloat64
	for i := range height * width {
		min = math.Min(min, zbuffer[i])
		max = math.Max(max, math.Min(zbuffer[i], far))
	}

	for i := range height * width {
		zbuffer[i] = 1 - (math.Min(zbuffer[i], far)-min)/(max-min)
	}
}

func render(spheres []Sphere) {
	width := 1024
	height := 768
	fov := math.Round(math.Pi) / 3

	zbuffer := make([]float64, width*height)
	computeDepthmap(width, height, fov, 23, spheres, zbuffer)

	framebuffer := make([]color.Color, width*height)
	for j := range height {
		for i := range width {
			framebuffer[i+j*width] = color.Gray{uint8(255 * zbuffer[i+j*width])}
		}
	}

	writePng("out", framebuffer, width, height)
}

func main() {
	spheres := []Sphere{}
	spheres = append(spheres, Sphere{Vec3f{-3, 0, -16}, 2})
	spheres = append(spheres, Sphere{Vec3f{-1.0, -1.5, -12}, 2})
	spheres = append(spheres, Sphere{Vec3f{1.5, -0.5, -18}, 3})
	spheres = append(spheres, Sphere{Vec3f{7, 5, -18}, 4})

	render(spheres)
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
