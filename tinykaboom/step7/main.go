package main

import (
	"image"
	"image/color"
	"image/png"
	"math"
	"os"
)

const (
	SPHERE_RADIUS   = 1.5
	NOISE_AMPLITUDE = 1.0
)

var (
	yellow   = Vec3f{1.7, 1.3, 1.0} // note that the color is "hot", i.e. has components >1
	orangle  = Vec3f{1.0, 0.6, 0.0}
	red      = Vec3f{1.0, 0.0, 0.0}
	darkgray = Vec3f{0.2, 0.2, 0.2}
	gray     = Vec3f{0.4, 0.4, 0.4}
)

func lerp[T float64 | int](v0, v1 T, t float64) T {
	return T(float64(v0) + float64(v1-v0)*math.Max(0, math.Min(1, t)))
}

func lerpV(v0, v1 Vec3f, t float64) Vec3f {
	return v0.Add(v1.Sub(v0).Muln(math.Max(0, math.Min(1, t))))
}

func hash(n float64) float64 {
	x := math.Sin(n) * 43758.5453
	return x - math.Floor(x)
}

func noise(x Vec3f) float64 {
	p := Vec3f{math.Floor(x.X()), math.Floor(x.Y()), math.Floor(x.Z())}
	f := Vec3f{x.X() - p.X(), x.Y() - p.Y(), x.Z() - p.Z()}
	f = f.Muln(f.Dot(Vec3f{3, 3, 3}.Sub(f.Muln(2))))
	n := p.Dot(Vec3f{1, 57, 113})
	return lerp(
		lerp(
			lerp(hash(n+0.), hash(n+1), f.X()),
			lerp(hash(n+57), hash(n+58), f.X()),
			f.Y()),
		lerp(
			lerp(hash(n+113), hash(n+114), f.X()),
			lerp(hash(n+170), hash(n+171), f.X()),
			f.Y()),
		f.Z())
}

func rotate(v Vec3f) Vec3f {
	return Vec3f{Vec3f{0.00, 0.80, 0.60}.Dot(v),
		Vec3f{-0.80, 0.36, -0.48}.Dot(v),
		Vec3f{-0.60, -0.48, 0.64}.Dot(v),
	}
}

func fractalBrownianMotion(x Vec3f) float64 {
	p := rotate(x)
	f := 0.0
	f += 0.5000 * noise(p)
	p = p.Muln(2.32)
	f += 0.2500 * noise(p)
	p = p.Muln(3.03)
	f += 0.1250 * noise(p)
	p = p.Muln(2.61)
	f += 0.0625 * noise(p)
	return f / 0.9375
}

func peletteFire(d float64) Vec3f {
	x := math.Max(0, math.Min(1, d))
	if x < 0.25 {
		return lerpV(gray, darkgray, x*4)
	} else if x < 0.5 {
		return lerpV(darkgray, red, x*4-1)
	} else if x < 0.75 {
		return lerpV(red, orangle, x*4-2)
	} else {
		return lerpV(orangle, yellow, x*4-3)
	}
}

func signedDistance(p Vec3f) float64 {
	displacement := -fractalBrownianMotion(p.Muln(3.4)) * NOISE_AMPLITUDE
	return p.Length() - (SPHERE_RADIUS + displacement)
}

func sphereTrace(orig, dir Vec3f, pos *Vec3f) bool {
	if orig.Dot(orig)-math.Pow(orig.Dot(dir), 2) > math.Pow(SPHERE_RADIUS, 2) {
		return false // early discard
	}

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
				noiseLevel := (SPHERE_RADIUS - hit.Length()) / NOISE_AMPLITUDE
				lightDir := Vec3f{10, 10, 10}.Sub(hit).Normalize() // one light is placed to (10,10,10)
				lightIntensity := math.Max(0.4, lightDir.Dot(distanceFieldNormal(hit)))
				framebuffer[i+j*width] = toColor(peletteFire((-0.2 + noiseLevel) * 2).Muln(lightIntensity))
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

func (v Vec3f) Divn(t float64) Vec3f {
	return Vec3f{v[0] / t, v[1] / t, v[2] / t}
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
