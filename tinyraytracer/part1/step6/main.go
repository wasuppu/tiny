package main

import (
	"image"
	"image/color"
	"image/png"
	"math"
	"os"
)

type Light struct {
	position  Vec3f
	intensity float64
}

type Material struct {
	albedo           Vec2f
	diffuseColor     Vec3f
	specularExponent float64
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

func reflect(i, n Vec3f) Vec3f {
	return i.Sub(n.Muln(2).Muln(i.Dot(n)))
}

func sceneIntersect(orig, dir Vec3f, spheres []Sphere, material *Material, hit *Vec3f, n *Vec3f) bool {
	sphereDist := math.MaxFloat64
	for i := range spheres {
		var disti float64
		if spheres[i].rayIntersect(orig, dir, &disti) && disti < sphereDist {
			sphereDist = disti
			*material = spheres[i].material
			*hit = orig.Add(dir.Muln(disti))
			*n = (*hit).Sub(spheres[i].center).Normalize()
		}
	}
	return sphereDist < 1000
}

func castRay(orig, dir Vec3f, spheres []Sphere, lights []Light) Vec3f {
	var point, n Vec3f
	material := Material{}
	if !sceneIntersect(orig, dir, spheres, &material, &point, &n) {
		return Vec3f{0.2, 0.7, 0.8} // background color
	}

	diffuseLightIntensity := 0.0
	specularLightIntensity := 0.0
	for i := range len(lights) {
		lightDir := lights[i].position.Sub(point).Normalize()
		lightDistance := lights[i].position.Sub(point).Length()

		shadowOrig := point.Add(n.Muln(1e-3)) // checking if the point lies in the shadow of the lights[i]
		if lightDir.Dot(n) < 0 {
			shadowOrig = point.Sub(n.Muln(1e-3))
		}
		var shadowPt, shadowN Vec3f
		tmpMaterial := Material{albedo: Vec2f{1, 0}}
		if sceneIntersect(shadowOrig, lightDir, spheres, &tmpMaterial, &shadowPt, &shadowN) && shadowPt.Sub(shadowOrig).Length() < lightDistance {
			continue
		}

		diffuseLightIntensity += lights[i].intensity * math.Max(0, lightDir.Dot(n))

		specularLightIntensity += math.Pow(math.Max(0, -reflect(lightDir.Muln(-1), n).Dot(dir)), material.specularExponent) * lights[i].intensity
	}

	return material.diffuseColor.Muln(diffuseLightIntensity).Muln(material.albedo[0]).Add(Vec3f{1, 1, 1}.Muln(specularLightIntensity).Muln(material.albedo[1]))
}

func render(spheres []Sphere, lights []Light) {
	width := 1024
	height := 768
	fov := int(math.Round(math.Pi)) / 2

	framebuffer := make([]color.Color, width*height)

	for j := range height {
		for i := range width {
			x := (2*(float64(i)+0.5)/float64(width) - 1) * math.Tan(float64(fov)/2) * float64(width) / float64(height)
			y := -(2*(float64(j)+0.5)/float64(height) - 1) * math.Tan(float64(fov)/2)
			dir := Vec3f{x, y, -1}.Normalize()
			framebuffer[i+j*width] = toColor(castRay(Vec3f{0, 0, 0}, dir, spheres, lights))
		}
	}

	writePng("out", framebuffer, width, height)
}

func main() {
	ivory := Material{Vec2f{0.6, 0.3}, Vec3f{0.4, 0.4, 0.3}, 50}
	redRubber := Material{Vec2f{0.9, 0.1}, Vec3f{0.3, 0.1, 0.1}, 10}

	spheres := []Sphere{}
	spheres = append(spheres, Sphere{Vec3f{-3, 0, -16}, 2, ivory})
	spheres = append(spheres, Sphere{Vec3f{-1.0, -1.5, -12}, 2, redRubber})
	spheres = append(spheres, Sphere{Vec3f{1.5, -0.5, -18}, 3, redRubber})
	spheres = append(spheres, Sphere{Vec3f{7, 5, -18}, 4, ivory})

	lights := []Light{}
	lights = append(lights, Light{Vec3f{-20, 20, 20}, 1.5})
	lights = append(lights, Light{Vec3f{30, 50, -25}, 1.8})
	lights = append(lights, Light{Vec3f{30, 20, 30}, 1.7})

	render(spheres, lights)
}

func toColor(v Vec3f) color.RGBA {
	max := math.Max(v[0], math.Max(v[1], v[2]))
	if max > 1 {
		v = v.Muln(1 / max)
	}
	return color.RGBA{uint8(255 * clamp(v[0], 0, 1)), uint8(255 * clamp(v[1], 0, 1)), uint8(255 * clamp(v[2], 0, 1)), 0xff}
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

type Vec2f [2]float64
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

func clamp[T float64 | int](val T, min T, max T) T {
	if val < min {
		return min
	} else if val > max {
		return max
	}
	return val
}
