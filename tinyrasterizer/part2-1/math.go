package main

import "image/color"

type Mesh struct {
	positions []Vec3f
	color     Vec4f
	count     uint32
}

type Vec3f [3]float64
type Vec4f [4]float64
type Mat4 [4]Vec4f

func (v Vec3f) AsVector() Vec4f {
	return Vec4f{v[0], v[1], v[2], 0}
}

func (v Vec3f) AsPoint() Vec4f {
	return Vec4f{v[0], v[1], v[2], 1}
}

func (v Vec4f) X() float64 {
	return v[0]
}

func (v Vec4f) Y() float64 {
	return v[1]
}

func (v Vec4f) Z() float64 {
	return v[2]
}

func (v Vec4f) W() float64 {
	return v[3]
}

func (v Vec4f) ToColor() color.RGBA {
	r := uint8(max(0, min(255, v[0]*255)))
	g := uint8(max(0, min(255, v[1]*255)))
	b := uint8(max(0, min(255, v[2]*255)))
	a := uint8(max(0, min(255, v[3]*255)))
	return color.RGBA{r, g, b, a}
}

func (v Vec4f) Sub(o Vec4f) Vec4f {
	return Vec4f{v[0] - o[0], v[1] - o[1], v[2] - o[2], v[3] - o[3]}
}

func (v Vec4f) Dot(o Vec4f) float64 {
	return v[0]*o[0] + v[1]*o[1] + v[2]*o[2] + v[3]*o[3]
}

func (v Vec4f) Det2D(o Vec4f) float64 {
	return v[0]*o[1] - v[1]*o[0]
}
