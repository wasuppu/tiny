package main

import (
	"fmt"
	"image/color"
	"math"
)

var (
	cubePositions = []Vec3f{
		// -X face
		{-1, -1, -1},
		{-1, 1, -1},
		{-1, -1, 1},
		{-1, 1, 1},

		// +X face
		{1, -1, -1},
		{1, 1, -1},
		{1, -1, 1},
		{1, 1, 1},

		// -Y face
		{-1, -1, -1},
		{1, -1, -1},
		{-1, -1, 1},
		{1, -1, 1},

		// +Y face
		{-1, 1, -1},
		{1, 1, -1},
		{-1, 1, 1},
		{1, 1, 1},

		// -Z face
		{-1, -1, -1},
		{1, -1, -1},
		{-1, 1, -1},
		{1, 1, -1},

		// +Z face
		{-1, -1, 1},
		{1, -1, 1},
		{-1, 1, 1},
		{1, 1, 1},
	}
	cubeColors = []Vec4f{
		// -X face
		{0, 1, 1, 1},
		{0, 1, 1, 1},
		{0, 1, 1, 1},
		{0, 1, 1, 1},

		// +X face
		{1, 0, 0, 1},
		{1, 0, 0, 1},
		{1, 0, 0, 1},
		{1, 0, 0, 1},

		// -Y face
		{1, 0, 1, 1},
		{1, 0, 1, 1},
		{1, 0, 1, 1},
		{1, 0, 1, 1},

		// +Y face
		{0, 1, 0, 1},
		{0, 1, 0, 1},
		{0, 1, 0, 1},
		{0, 1, 0, 1},

		// -Z face
		{1, 1, 0, 1},
		{1, 1, 0, 1},
		{1, 1, 0, 1},
		{1, 1, 0, 1},

		// +Z face
		{0, 0, 1, 1},
		{0, 0, 1, 1},
		{0, 0, 1, 1},
		{0, 0, 1, 1},
	}
	cubeIndices = []uint32{
		// -X face
		0, 2, 1,
		1, 2, 3,

		// +X face
		4, 5, 6,
		6, 5, 7,

		// -Y face
		8, 9, 10,
		10, 9, 11,

		// +Y face
		12, 14, 13,
		14, 15, 13,

		// -Z face
		16, 18, 17,
		17, 18, 19,

		// +Z face
		20, 21, 22,
		21, 23, 22,
	}
	Cube = Mesh{cubePositions, cubeColors, cubeIndices, 36}
)

type Mesh struct {
	positions []Vec3f
	colors    []Vec4f
	indices   []uint32
	count     uint32
}

type Vertex struct {
	position Vec4f
	color    Vec4f
}

// f(t) = at+b
// f(0) = v0 = b
// f(1) = v1 = a+v0 => a = v1 - v0
// f(t) = v0 + (v1 - v0) * t
// f(t) = 0 => t = -v0 / (v1 - v0) = v0 / (v0 - v1)
func clipIntersectEdge(v0, v1 Vertex, value0, value1 float64) Vertex {
	t := value0 / (value0 - value1)

	position := v0.position.Muln(1 - t).Add(v1.position.Muln(t))
	color := v0.color.Muln(1 - t).Add(v1.color.Muln(t))
	return Vertex{position, color}
}

func clipTriangle(triangle [3]Vertex, equation Vec4f, result []Vertex) []Vertex {
	values := [3]float64{
		triangle[0].position.Dot(equation),
		triangle[1].position.Dot(equation),
		triangle[2].position.Dot(equation),
	}

	mask := 0
	if values[0] < 0 {
		mask |= 1
	}
	if values[1] < 0 {
		mask |= 2
	}
	if values[2] < 0 {
		mask |= 4
	}

	v01 := clipIntersectEdge(triangle[0], triangle[1], values[0], values[1])
	v02 := clipIntersectEdge(triangle[0], triangle[2], values[0], values[2])
	v10 := clipIntersectEdge(triangle[1], triangle[0], values[1], values[0])
	v12 := clipIntersectEdge(triangle[1], triangle[2], values[1], values[2])
	v20 := clipIntersectEdge(triangle[2], triangle[0], values[2], values[0])
	v21 := clipIntersectEdge(triangle[2], triangle[1], values[2], values[1])

	switch mask {
	case 0b000:
		// All vertices are inside allowed half-space
		// No clipping required, copy the triangle to output
		result = append(result, triangle[0], triangle[1], triangle[2])
	case 0b001:
		// Vertex 0 is outside allowed half-space
		// Replace it with points on edges 01 and 02
		// And re-triangulate
		result = append(result, v01, triangle[1], triangle[2], v01, triangle[2], v02)
	case 0b010:
		// Vertex 1 is outside allowed half-space
		// Replace it with points on edges 10 and 12
		// And re-triangulate
		result = append(result, triangle[0], v10, triangle[2], triangle[2], v10, v12)
	case 0b011:
		// Vertices 0 and 1 are outside allowed half-space
		// Replace them with points on edges 02 and 12
		result = append(result, v02, v12, triangle[2])
	case 0b100:
		// Vertex 2 is outside allowed half-space
		// Replace it with points on edges 20 and 21
		// And re-triangulate
		result = append(result, triangle[0], triangle[1], v20, v20, triangle[1], v21)
	case 0b101:
		// Vertices 0 and 2 are outside allowed half-space
		// Replace them with points on edges 01 and 21
		result = append(result, v01, triangle[1], v21)
	case 0b110:
		// Vertices 1 and 2 are outside allowed half-space
		// Replace them with points on edges 10 and 20
		result = append(result, triangle[0], v10, v20)
	case 0b111:
		// All vertices are outside allowed half-space
		// Clip the whole triangle, result is empty
	}

	return result
}

func ClipTriangle(vertices []Vertex) []Vertex {
	equations := [2]Vec4f{
		{0, 0, 1, 1},  // Z > -W  <=>   Z + W > 0
		{0, 0, -1, 1}, // Z <  W  <=> - Z + W > 0
	}

	result := []Vertex{}
	temp := []Vertex{vertices[0], vertices[1], vertices[2]}
	for _, equation := range equations {
		resultend := []Vertex{}
		for i := 0; i < len(temp); i += 3 {
			triangle := [3]Vertex{temp[i], temp[i+1], temp[i+2]}
			resultend = clipTriangle(triangle, equation, resultend)
		}
		result = append(result, resultend...)
		temp = resultend
	}
	return result
}

type Vec3f [3]float64
type Vec4f [4]float64
type Mat4 [4]Vec4f

func (v Vec4f) String() string {
	if _, ok := any(v[0]).(int); ok {
		return fmt.Sprintf("%d, %d, %d, %d", int(v[0]), int(v[1]), int(v[2]), int(v[3]))
	} else if _, ok := any(v[0]).(float64); ok {
		return fmt.Sprintf("%.6g %.6g %.6g %.7g", float64(v[0]), float64(v[1]), float64(v[2]), float64(v[3]))
	} else {
		return fmt.Sprintf("%d, %d, %d, %d", int(v[0]), int(v[1]), int(v[2]), int(v[3]))
	}
}

func (v Vec3f) X() float64 {
	return v[0]
}

func (v Vec3f) Y() float64 {
	return v[1]
}

func (v Vec3f) Z() float64 {
	return v[2]
}

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

func (v Vec4f) Add(o Vec4f) Vec4f {
	return Vec4f{v[0] + o[0], v[1] + o[1], v[2] + o[2], v[3] + o[3]}
}

func (v Vec4f) Sub(o Vec4f) Vec4f {
	return Vec4f{v[0] - o[0], v[1] - o[1], v[2] - o[2], v[3] - o[3]}
}

func (v Vec4f) Muln(t float64) Vec4f {
	return Vec4f{v[0] * t, v[1] * t, v[2] * t, v[3] * t}
}

func (v Vec4f) Dot(o Vec4f) float64 {
	return v[0]*o[0] + v[1]*o[1] + v[2]*o[2] + v[3]*o[3]
}

func (v Vec4f) Det2D(o Vec4f) float64 {
	return v[0]*o[1] - v[1]*o[0]
}

func (v Vec4f) PerspectiveDivide() Vec4f {
	w := 1 / v[3]
	return Vec4f{v[0] * w, v[1] * w, v[2] * w, w}
}

func (m Mat4) Mulv(v Vec4f) Vec4f {
	u := Vec4f{}
	for i := range 4 {
		u[i] = m[i].Dot(v)
	}
	return u
}

func (m Mat4) Mul(n Mat4) Mat4 {
	a := Mat4{}
	for i := range 4 {
		for j := range 4 {
			for k := range 4 {
				a[i][j] += m[i][k] * n[k][j]
			}
		}
	}
	return a
}

func ID4() Mat4 {
	m := Mat4{}
	for i := range 4 {
		for j := range 4 {
			if i == j {
				m[i][j] = 1
			} else {
				m[i][j] = 0
			}
		}
	}
	return m
}

func ScaleF(s float64) Mat4 {
	return Scale(Vec3f{s, s, s})
}

func Scale(v Vec3f) Mat4 {
	return Mat4{
		{v.X(), 0, 0, 0},
		{0, v.Y(), 0, 0},
		{0, 0, v.Z(), 0},
		{0, 0, 0, 1},
	}
}

func Translate(v Vec3f) Mat4 {
	return Mat4{
		{1, 0, 0, v.X()},
		{0, 1, 0, v.Y()},
		{0, 0, 1, v.Z()},
		{0, 0, 0, 1},
	}
}

func RotateXY(angle float64) Mat4 {
	sin, cos := math.Sincos(angle)
	return Mat4{
		{cos, -sin, 0, 0},
		{sin, cos, 0, 0},
		{0, 0, 1, 0},
		{0, 0, 0, 1},
	}
}

func RotateYZ(angle float64) Mat4 {
	sin, cos := math.Sincos(angle)
	return Mat4{
		{1, 0, 0, 0},
		{0, cos, -sin, 0},
		{0, sin, cos, 0},
		{0, 0, 0, 1},
	}
}

func RotateZX(angle float64) Mat4 {
	sin, cos := math.Sincos(angle)
	return Mat4{
		{cos, 0, sin, 0},
		{0, 1, 0, 0},
		{-sin, 0, cos, 0},
		{0, 0, 0, 1},
	}
}

func Perspective(near, far, fov, aspect float64) Mat4 {
	top := near * math.Tan(fov/2)
	right := top * aspect

	return Mat4{
		{near / right, 0, 0, 0},
		{0, near / top, 0, 0},
		{0, 0, -(far + near) / (far - near), -2 * far * near / (far - near)},
		{0, 0, -1, 0},
	}
}
