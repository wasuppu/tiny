package main

import (
	"image"
	"image/color"
	"image/png"
	"math"
	"os"
)

var (
	ivory     = Material{1.0, []float64{0.9, 0.5, 0.1, 0.0}, vec3{0.4, 0.4, 0.3}, 50}
	glass     = Material{1.5, []float64{0.0, 0.9, 0.1, 0.8}, vec3{0.6, 0.7, 0.8}, 125}
	redRubber = Material{1.0, []float64{1.4, 0.3, 0.0, 0.0}, vec3{0.3, 0.1, 0.1}, 10}
	mirror    = Material{1.0, []float64{0.0, 16.0, 0.8, 0.0}, vec3{1.0, 1.0, 1.0}, 1425}
	spheres   = []Sphere{
		{vec3{-3, 0, -16}, 2, ivory},
		{vec3{-1.0, -1.5, -12}, 2, glass},
		{vec3{1.5, -0.5, -18}, 3, redRubber},
		{vec3{7, 5, -18}, 4, mirror}}
	lights = []vec3{
		{-20, 20, 20},
		{30, 50, -25},
		{30, 20, 30}}
)

type Material struct {
	refractiveIndex  float64
	albedo           []float64
	diffuseColor     vec3
	specularExponent float64
}

func NewMaterial() Material {
	return Material{refractiveIndex: 1, albedo: []float64{2, 0, 0, 0}}
}

type Sphere struct {
	center   vec3
	radius   float64
	material Material
}

func (s Sphere) rayIntersect(orig, dir vec3) (bool, float64) {
	L := s.center.Sub(orig)
	tca := L.Dot(dir)
	d2 := L.Dot(L) - tca*tca
	if d2 > s.radius*s.radius {
		return false, 0
	}
	thc := math.Sqrt(s.radius*s.radius - d2)
	t0 := tca - thc
	t1 := tca + thc
	// offset the original point by .001 to avoid occlusion by the object itself
	if t0 > 0.001 {
		return true, t0
	}
	if t1 > 0.001 {
		return true, t1
	}
	return false, 0
}

func reflect(I, N vec3) vec3 {
	return I.Sub(N.Muln(2).Muln(I.Dot(N)))
}

func refract(I, N vec3, etat, etai float64) vec3 {
	cosi := -math.Max(-1, math.Min(1, I.Dot(N)))
	// if the ray is inside the object, swap the indices and invert the normal to get the correct result
	if cosi < 0 {
		return refract(I, N.Muln(-1), etai, etat)
	}
	eta := etai / etat
	k := 1 - eta*eta*(1-cosi*cosi)

	if k < 0 {
		return vec3{1, 0, 0}
	} else {
		// k<0 = total reflection, no ray to refract. I refract it anyways, this has no physical meaning
		return I.Muln(eta).Add(N.Muln(eta*cosi - math.Sqrt(k)))
	}
}

func sceneIntersect(orig, dir vec3) (bool, vec3, vec3, Material) {
	var pt, N vec3
	material := NewMaterial()
	nearestDist := 1e10
	// intersect the ray with the checkerboard, avoid division by zero
	if math.Abs(dir.Y()) > 0.001 {
		// the checkerboard plane has equation y = -4
		d := -(orig.Y() + 4) / dir.Y()
		p := orig.Add(dir.Muln(d))
		if d > 0.001 && d < nearestDist && math.Abs(p.X()) < 10 && p.Z() < -10 && p.Z() > -30 {
			nearestDist = d
			pt = p
			N = vec3{0, 1, 0}
			material.diffuseColor = vec3{.3, .2, .1}
			if (int(0.5*pt.X()+1000)+int(0.5*pt.Z()))&1 != 0 {
				material.diffuseColor = vec3{.3, .3, .3}
			}
		}
	}

	for _, s := range spheres {
		intersection, d := s.rayIntersect(orig, dir)
		if !intersection || d > nearestDist {
			continue
		}
		nearestDist = d
		pt = orig.Add(dir.Muln(nearestDist))
		N = pt.Sub(s.center).Normalize()
		material = s.material
	}
	return nearestDist < 1000, pt, N, material
}

func castRay(orig, dir vec3, depth int) vec3 {
	hit, point, N, material := sceneIntersect(orig, dir)

	if depth > 4 || !hit {
		return vec3{0.2, 0.7, 0.8} // background color
	}

	reflectDir := reflect(dir, N).Normalize()
	refractDir := refract(dir, N, material.refractiveIndex, 1).Normalize()
	reflectColor := castRay(point, reflectDir, depth+1)
	refractColor := castRay(point, refractDir, depth+1)

	diffuseLightIntensity := 0.0
	specularLightIntensity := 0.0
	for _, light := range lights {
		lightDir := light.Sub(point).Normalize()
		hit, shadowPt, trashnrm, trashmat := sceneIntersect(point, lightDir)
		_, _ = trashnrm, trashmat
		if hit && shadowPt.Sub(point).Length() < light.Sub(point).Length() {
			continue
		}
		diffuseLightIntensity += math.Max(0, lightDir.Dot(N))
		specularLightIntensity += math.Pow(math.Max(0, -reflect(lightDir.Muln(-1), N).Dot(dir)), material.specularExponent)
	}

	return material.diffuseColor.Muln(diffuseLightIntensity).Muln(material.albedo[0]).Add(vec3{1, 1, 1}.Muln(specularLightIntensity).Muln(material.albedo[1])).Add(reflectColor.Muln(material.albedo[2])).Add(refractColor.Muln(material.albedo[3]))
}

func render() {
	width := 1024
	height := 768
	fov := 1.05 // 60 degrees field of view in radians

	framebuffer := make([]color.Color, width*height)

	for pix := range width * height {
		dirx := (float64(pix%width) + 0.5) - float64(width/2)
		diry := -(float64(pix/width) + 0.5) + float64(height/2) // this flips the image at the same time
		dirz := -float64(height) / (2 * math.Tan(fov/2))
		dir := vec3{dirx, diry, dirz}.Normalize()
		framebuffer[pix] = toColor(castRay(vec3{0, 0, 0}, dir, 0))
	}

	writePng("out", framebuffer, width, height)
}

func main() {
	render()
}

func toColor(v vec3) color.RGBA {
	max := math.Max(1, math.Max(v[0], math.Max(v[1], v[2])))
	return color.RGBA{uint8(255 * v[0] / max), uint8(255 * v[1] / max), uint8(255 * v[2] / max), 0xff}
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

type vec3 [3]float64

func (v vec3) X() float64 {
	return v[0]
}

func (v vec3) Y() float64 {
	return v[1]
}

func (v vec3) Z() float64 {
	return v[2]
}

func (v vec3) Add(o vec3) vec3 {
	return vec3{v[0] + o[0], v[1] + o[1], v[2] + o[2]}
}

func (v vec3) Sub(o vec3) vec3 {
	return vec3{v[0] - o[0], v[1] - o[1], v[2] - o[2]}
}

func (v vec3) Muln(t float64) vec3 {
	return vec3{v[0] * t, v[1] * t, v[2] * t}
}

func (v vec3) Dot(o vec3) float64 {
	return v[0]*o[0] + v[1]*o[1] + v[2]*o[2]
}

func (v vec3) Length() float64 {
	return math.Sqrt(float64(v.Dot(v)))
}

func (v vec3) Normalize() vec3 {
	return v.Muln(1 / v.Length())
}
