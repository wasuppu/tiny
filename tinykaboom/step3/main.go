package main

import (
	"image"
	"image/color"
	"image/png"
	"math"
	"os"
)

const (
	SPHERE_RADIUS = 1.5
)

func signedDistance(p Vec3f) float64 {
	return p.Length() - SPHERE_RADIUS
}

func sphereTrace(orig, dir Vec3f, pos *Vec3f) bool {
	*pos = orig
	for range 128 {
		d := signedDistance(*pos)
		if d < 0 {
			return true
		}
		*pos = pos.Add(dir.Muln(math.Max(d*0.1, 0.01)))
	}
	return false
}

func distanceFieldNormal(pos Vec3f) Vec3f {
	eps := 0.1
	d := signedDistance(pos)
	nx := signedDistance(pos.Add(Vec3f{eps, 0, 0})) - d
	ny := signedDistance(pos.Add(Vec3f{0, eps, 0})) - d
	nz := signedDistance(pos.Add(Vec3f{0, 0, eps})) - d
	return Vec3f{nx, ny, nz}.Normalize()
}

func main() {
	width := 640
	height := 480
	fov := math.Pi / 3

	framebuffer := make([]color.Color, width*height)

	// actual rendering loop
	for j := range height {
		for i := range width {
			x := (float64(i) + 0.5) - float64(width)/2
			y := -(float64(j) + 0.5) + float64(height)/2 // this flips the image at the same time
			z := -float64(height) / (2 * math.Tan(fov/2.))
			dir := Vec3f{x, y, z}.Normalize()
			var hit Vec3f
			// the camera is placed to (0,0,3) and it looks along the -z axis
			if sphereTrace(Vec3f{0, 0, 3}, dir, &hit) {
				lightDir := Vec3f{10, 10, 10}.Sub(hit).Normalize() // one light is placed to (10,10,10)
				lightIntensity := math.Max(0.4, lightDir.Dot(distanceFieldNormal(hit)))
				displacement := (math.Sin(16*hit.X())*math.Sin(16*hit.Y())*math.Sin(16*hit.Z()) + 1) / 2
				framebuffer[i+j*width] = toColor(Vec3f{1, 1, 1}.Muln(displacement * lightIntensity))
			} else {
				framebuffer[i+j*width] = toColor(Vec3f{0.2, 0.7, 0.8}) // background color
			}
		}
	}

	writePng("out", framebuffer, width, height) // save the framebuffer to file
}

func toColor(v Vec3f) color.RGBA {
	max := math.Max(v[0], math.Max(v[1], v[2]))
	if max > 1 {
		v = v.Muln(1 / max)
	}
	return color.RGBA{
		uint8(255 * clamp(v[0], 0, 1)),
		uint8(255 * clamp(v[1], 0, 1)),
		uint8(255 * clamp(v[2], 0, 1)),
		0xff,
	}
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

func clamp[T float64 | int](val T, min T, max T) T {
	if val < min {
		return min
	} else if val > max {
		return max
	}
	return val
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
