package main

import (
	"math"
)

type Number interface {
	int | float64 | float32
}

type Vec[T Number] []T

type Vecf = Vec[float64]
type Veci = Vec[int]

func (v Vec[T]) Clone() Vec[T] {
	u := make(Vec[T], len(v))
	copy(u, v)
	return u
}

func (v Vec[T]) ToInt() Veci {
	u := make(Vec[int], len(v))
	for i := range u {
		u[i] = int(math.Round(float64(v[i])))
	}
	return u
}

func (v Vec[T]) ToFloat() Vecf {
	u := make(Vec[float64], len(v))
	for i := range u {
		u[i] = float64(v[i])
	}
	return u
}

func (v Vec[T]) Proj(dim int, fill T) Vec[T] {
	u := make(Vec[T], dim)
	for i := range dim {
		if i >= len(v) {
			u[i] = fill
		} else {
			u[i] = v[i]
		}
	}
	return u
}

func (v Vec[T]) Add(o Vec[T]) Vec[T] {
	assert(len(v) == len(o), "length of vector should be same")
	u := v.Clone()

	for i := range u {
		u[i] += o[i]
	}

	return u
}

func (v Vec[T]) Sub(o Vec[T]) Vec[T] {
	assert(len(v) == len(o), "length of vector should be same")
	u := v.Clone()
	for i := range u {
		u[i] -= o[i]
	}
	return u
}

func (v Vec[T]) Muln(t float64) Vec[T] {
	u := v.Clone()
	for i := range u {
		u[i] = T(float64(u[i]) * t)
	}
	return u
}

func (v Vec[T]) Mulm(m Matrix) Vecf {
	if len(v) != m.Nrows() {
		panic("column of matrix must be same as length of vector")
	}

	r, c := m.Nrows(), m.Ncols()
	u := make(Vecf, r)
	for j := range c {
		for i := range r {
			u[j] += float64(v[i]) * m[i][j]
		}
	}
	return u
}

func (v Vec[T]) Divn(t float64) Vec[T] {
	u := v.Clone()
	for i := range u {
		u[i] = T(float64(u[i]) / t)
	}
	return u
}

func (v Vec[T]) Dot(o Vec[T]) T {
	assert(len(v) == len(o), "length of vector should be same")
	s := 0.0
	for i := range v {
		s += float64(v[i] * o[i])
	}
	return T(s)
}

func (v Vec[T]) Length() float64 {
	return math.Sqrt(float64(v.Dot(v)))
}

func (v Vec[T]) Normalize() Vec[T] {
	return v.Divn(v.Length())
}

func assert(cond bool, msg string) {
	if !cond {
		panic(msg)
	}
}

type Vec2[T Number] [2]T
type Vec2f = Vec2[float64]
type Vec2i = Vec2[int]

type Point2 = Vec3f

func (v Vec2[T]) V() Vec[T] {
	return v[:]
}

func (v Vec2[T]) V3(fill T) Vec3[T] {
	return Vec3[T]{v[0], v[1], fill}
}

func (v Vec2[T]) V4(fill T) Vec4[T] {
	return Vec4[T]{v[0], v[1], fill, fill}
}

func (v Vec2[T]) F() Vec2f {
	u := Vec2f{}
	for i := range u {
		u[i] = float64(v[i])
	}
	return u
}

func (v Vec2[T]) I() Vec2i {
	u := Vec2i{}
	for i := range u {
		u[i] = int(float64(v[i]) + .5)
	}
	return u
}

func (v Vec2[T]) X() T {
	return v[0]
}

func (v Vec2[T]) S() T {
	return v[0]
}

func (v Vec2[T]) Y() T {
	return v[1]
}

func (v Vec2[T]) T() T {
	return v[1]
}

func (v Vec2[T]) Add(o Vec2[T]) Vec2[T] {
	return Vec2[T]{v[0] + o[0], v[1] + o[1]}
}

func (v Vec2[T]) Sub(o Vec2[T]) Vec2[T] {
	return Vec2[T]{v[0] - o[0], v[1] - o[1]}
}

func (v Vec2[T]) Muln(t float64) Vec2[T] {
	return Vec2[T]{T(float64(v[0]) * t), T(float64(v[1]) * t)}
}

func (v Vec2[T]) Divn(t float64) Vec2[T] {
	return Vec2[T]{T(float64(v[0]) / t), T(float64(v[1]) / t)}
}

func (v Vec2[T]) Dot(o Vec2[T]) T {
	return v[0]*o[0] + v[1]*o[1]
}

func (v Vec2[T]) Cross(o Vec2[T]) T {
	return v[0]*o[1] - v[1]*o[0]
}

func (v Vec2[T]) Length() float64 {
	return math.Sqrt(float64(v.Dot(v)))
}

func (v Vec2[T]) Normalize() Vec2[T] {
	return v.Divn(v.Length())
}

type Vec3[T Number] [3]T
type Vec3f = Vec3[float64]
type Vec3i = Vec3[int]

type Point3 = Vec3f

func (v Vec3[T]) V() Vec[T] {
	return v[:]
}

func (v Vec3[T]) V4(fill T) Vec4[T] {
	return Vec4[T]{v[0], v[1], v[2], fill}
}

func (v Vec3[T]) M() Matrix {
	m := NewMatrix(4, 1)
	m[0][0] = float64(v.X())
	m[1][0] = float64(v.Y())
	m[2][0] = float64(v.Z())
	m[3][0] = 1
	return m
}

func (v Vec3[T]) I() Vec3i {
	u := Vec3i{}
	for i := range u {
		u[i] = int(float64(v[i]) + .5)
	}
	return u
}

func (v Vec3[T]) F() Vec3f {
	u := Vec3f{}
	for i := range u {
		u[i] = float64(v[i])
	}
	return u
}

func (v Vec3[T]) X() T {
	return v[0]
}

func (v Vec3[T]) Y() T {
	return v[1]
}

func (v Vec3[T]) Z() T {
	return v[2]
}

func (v Vec3[T]) Add(o Vec3[T]) Vec3[T] {
	return Vec3[T]{v[0] + o[0], v[1] + o[1], v[2] + o[2]}
}

func (v Vec3[T]) Sub(o Vec3[T]) Vec3[T] {
	return Vec3[T]{v[0] - o[0], v[1] - o[1], v[2] - o[2]}
}

func (v Vec3[T]) Muln(t float64) Vec3[T] {
	return Vec3[T]{T(float64(v[0]) * t), T(float64(v[1]) * t), T(float64(v[2]) * t)}
}

func (v Vec3[T]) Mulm(m Mat3) Vec3f {
	u := Vec3f{}
	for j := range 3 {
		for i := range 3 {
			u[j] += float64(v[i]) * m[i][j]
		}
	}
	return u
}

func (v Vec3[T]) Divn(t float64) Vec3[T] {
	return Vec3[T]{T(float64(v[0]) / t), T(float64(v[1]) / t), T(float64(v[2]) / t)}
}

func (v Vec3[T]) Dot(o Vec3[T]) T {
	return v[0]*o[0] + v[1]*o[1] + v[2]*o[2]
}

func (v Vec3[T]) Cross(o Vec3[T]) Vec3[T] {
	x := v[1]*o[2] - v[2]*o[1]
	y := v[2]*o[0] - v[0]*o[2]
	z := v[0]*o[1] - v[1]*o[0]

	return Vec3[T]{x, y, z}
}

func (v Vec3[T]) Length() float64 {
	return math.Sqrt(float64(v.Dot(v)))
}

func (v Vec3[T]) Normalize() Vec3[T] {
	return v.Divn(v.Length())
}

type Vec4[T Number] [4]T
type Vec4f = Vec4[float64]
type Vec4i = Vec4[int]

type Point4 = Vec4f

func (v Vec4[T]) V2() Vec2[T] {
	return Vec2[T]{v[0], v[1]}
}

func (v Vec4[T]) V3() Vec3[T] {
	return Vec3[T]{v[0], v[1], v[2]}
}

func (v Vec4[T]) X() T {
	return v[0]
}

func (v Vec4[T]) Y() T {
	return v[1]
}

func (v Vec4[T]) Z() T {
	return v[2]
}

func (v Vec4[T]) W() T {
	return v[3]
}

func (v Vec4[T]) Add(o Vec4[T]) Vec4[T] {
	return Vec4[T]{v[0] + o[0], v[1] + o[1], v[2] + o[2], v[3] + o[3]}
}

func (v Vec4[T]) Sub(o Vec4[T]) Vec4[T] {
	return Vec4[T]{v[0] - o[0], v[1] - o[1], v[2] - o[2], v[3] - o[3]}
}

func (v Vec4[T]) Muln(t float64) Vec4[T] {
	return Vec4[T]{T(float64(v[0]) * t), T(float64(v[1]) * t), T(float64(v[2]) * t), T(float64(v[3]) * t)}
}

func (v Vec4[T]) Divn(t float64) Vec4[T] {
	return Vec4[T]{T(float64(v[0]) / t), T(float64(v[1]) / t), T(float64(v[2]) / t), T(float64(v[3]) / t)}
}

func (v Vec4[T]) Mulm(m Mat4) Vec4f {
	u := Vec4f{}
	for j := range 4 {
		for i := range 4 {
			u[j] += float64(v[i]) * m[i][j]
		}
	}
	return u
}

func (v Vec4[T]) Dot(o Vec4[T]) T {
	return v[0]*o[0] + v[1]*o[1] + v[2]*o[2] + v[3]*o[3]
}

func (v Vec4[T]) Length() float64 {
	return math.Sqrt(float64(v.Dot(v)))
}

func (v Vec4[T]) Normalize() Vec4[T] {
	return v.Divn(v.Length())
}

type Mat3 [3]Vec3f

func (m Mat3) M() Matrix {
	mat := make([]Vecf, 3)
	for i := range m {
		mat[i] = m[i][:]
	}
	return mat
}

func (m *Mat3) Mm() Matrix {
	mat := make([]Vecf, 2)
	for i := range m {
		mat[i] = m[i][:]
	}
	return mat
}

func (m Mat3) Col(j int) Vec3f {
	v := Vec3f{}
	for i := range m {
		v[i] = m[i][j]
	}
	return v
}

func (m *Mat3) SetCol(idx int, v Vec3f) {
	r, c := 3, 3
	if idx >= c {
		panic("column index greater than the number of columns")
	}
	for i := range r {
		m[i][idx] = v[i]
	}
}

func ID3() Mat3 {
	m := Mat3{}
	for i := range 3 {
		for j := range 3 {
			if i == j {
				m[i][j] = 1
			} else {
				m[i][j] = 0
			}
		}
	}
	return m
}

func (m Mat3) Add(n Mat3) Mat3 {
	a := Mat3{}
	for i := range 3 {
		a[i] = m[i].Add(n[i])
	}
	return a
}

func (m Mat3) Sub(n Mat3) Mat3 {
	a := Mat3{}
	for i := range 3 {
		a[i] = m[i].Sub(n[i])
	}
	return a
}

func (m Mat3) Mul(n Mat3) Mat3 {
	a := Mat3{}
	for i := range 3 {
		for j := range 3 {
			for k := range 3 {
				a[i][j] += m[i][k] * n[k][j]
			}
		}
	}
	return a
}

func (m Mat3) Muln(t float64) Mat3 {
	a := Mat3{}
	for i := range 3 {
		a[i] = m[i].Muln(t)
	}
	return a
}

func (m Mat3) Mulv(v Vec3f) Vec3f {
	u := Vec3f{}
	for i := range 3 {
		u[i] = m[i].Dot(v)
	}
	return u
}

func (m Mat3) Divn(t float64) Mat3 {
	a := Mat3{}
	for i := range 3 {
		a[i] = m[i].Divn(t)
	}
	return a
}

func (m Mat3) Transpose() Mat3 {
	n := Mat3{}
	for i := range 3 {
		for j := range 3 {
			n[j][i] = m[i][j]
		}
	}
	return n
}

func (m Mat3) Determinant() float64 {
	return m[0][0]*(m[1][1]*m[2][2]-m[2][1]*m[1][2]) -
		m[1][0]*(m[0][1]*m[2][2]-m[2][1]*m[0][2]) +
		m[2][0]*(m[0][1]*m[1][2]-m[1][1]*m[0][2])
}

func (m Mat3) Inverse() Mat3 {
	n := Mat3{}
	d := m.Determinant()

	n[0][0] = (m[1][1]*m[2][2] - m[2][1]*m[1][2]) / d
	n[0][1] = -(m[0][1]*m[2][2] - m[2][1]*m[0][2]) / d
	n[0][2] = (m[0][1]*m[1][2] - m[1][1]*m[0][2]) / d
	n[1][0] = -(m[1][0]*m[2][2] - m[2][0]*m[1][2]) / d
	n[1][1] = (m[0][0]*m[2][2] - m[2][0]*m[0][2]) / d
	n[1][2] = -(m[0][0]*m[1][2] - m[1][0]*m[0][2]) / d
	n[2][0] = (m[1][0]*m[2][1] - m[2][0]*m[1][1]) / d
	n[2][1] = -(m[0][0]*m[2][1] - m[2][0]*m[0][1]) / d
	n[2][2] = (m[0][0]*m[1][1] - m[1][0]*m[0][1]) / d

	return n
}

type Mat4 [4]Vec4f

func (m Mat4) V() Vec3f {
	return Vec3f{m[0][0] / m[3][0], m[1][0] / m[3][0], m[2][0] / m[3][0]}
}

func (m Mat4) M() Matrix {
	mat := make([]Vecf, 4)
	for i := range m {
		mat[i] = m[i][:]
	}
	return mat
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

func (m Mat4) Add(n Mat4) Mat4 {
	a := Mat4{}
	for i := range 4 {
		a[i] = m[i].Add(n[i])
	}
	return a
}

func (m Mat4) Sub(n Mat4) Mat4 {
	a := Mat4{}
	for i := range 4 {
		a[i] = m[i].Sub(n[i])
	}
	return a
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

func (m Mat4) Mul4x3(n Mat4x3) Mat4x3 {
	a := Mat4x3{}
	for i := range 4 {
		for j := range 3 {
			for k := range 4 {
				a[i][j] += m[i][k] * n[k][j]
			}
		}
	}
	return a
}

func (m Mat4) Muln(t float64) Mat4 {
	a := Mat4{}
	for i := range 4 {
		a[i] = m[i].Muln(t)
	}
	return a
}

func (m Mat4) Mulv(v Vec4f) Vec4f {
	u := Vec4f{}
	for i := range 4 {
		u[i] = m[i].Dot(v)
	}
	return u
}

func (m Mat4) Divn(t float64) Mat4 {
	a := Mat4{}
	for i := range 4 {
		a[i] = m[i].Divn(t)
	}
	return a
}

func (m Mat4) Transpose() Mat4 {
	n := Mat4{}
	for i := range 4 {
		for j := range 4 {
			n[j][i] = m[i][j]
		}
	}
	return n
}

func (m Mat4) Determinant() float64 {
	return (m[0][0]*m[1][1]*m[2][2]*m[3][3] - m[0][0]*m[1][1]*m[2][3]*m[3][2] +
		m[0][0]*m[1][2]*m[2][3]*m[3][1] - m[0][0]*m[1][2]*m[2][1]*m[3][3] +
		m[0][0]*m[1][3]*m[2][1]*m[3][2] - m[0][0]*m[1][3]*m[2][2]*m[3][1] -
		m[0][1]*m[1][2]*m[2][3]*m[3][0] + m[0][1]*m[1][2]*m[2][0]*m[3][3] -
		m[0][1]*m[1][3]*m[2][0]*m[3][2] + m[0][1]*m[1][3]*m[2][2]*m[3][0] -
		m[0][1]*m[1][0]*m[2][2]*m[3][3] + m[0][1]*m[1][0]*m[2][3]*m[3][2] +
		m[0][2]*m[1][3]*m[2][0]*m[3][1] - m[0][2]*m[1][3]*m[2][1]*m[3][0] +
		m[0][2]*m[1][0]*m[2][1]*m[3][3] - m[0][2]*m[1][0]*m[2][3]*m[3][1] +
		m[0][2]*m[1][1]*m[2][3]*m[3][0] - m[0][2]*m[1][1]*m[2][0]*m[3][3] -
		m[0][3]*m[1][0]*m[2][1]*m[3][2] + m[0][3]*m[1][0]*m[2][2]*m[3][1] -
		m[0][3]*m[1][1]*m[2][2]*m[3][0] + m[0][3]*m[1][1]*m[2][0]*m[3][2] -
		m[0][3]*m[1][2]*m[2][0]*m[3][1] + m[0][3]*m[1][2]*m[2][1]*m[3][0])
}

func (m Mat4) Inverse() Mat4 {
	n := Mat4{}
	d := m.Determinant()

	n[0][0] = (m[1][2]*m[2][3]*m[3][1] - m[1][3]*m[2][2]*m[3][1] + m[1][3]*m[2][1]*m[3][2] - m[1][1]*m[2][3]*m[3][2] - m[1][2]*m[2][1]*m[3][3] + m[1][1]*m[2][2]*m[3][3]) / d
	n[0][1] = (m[0][3]*m[2][2]*m[3][1] - m[0][2]*m[2][3]*m[3][1] - m[0][3]*m[2][1]*m[3][2] + m[0][1]*m[2][3]*m[3][2] + m[0][2]*m[2][1]*m[3][3] - m[0][1]*m[2][2]*m[3][3]) / d
	n[0][2] = (m[0][2]*m[1][3]*m[3][1] - m[0][3]*m[1][2]*m[3][1] + m[0][3]*m[1][1]*m[3][2] - m[0][1]*m[1][3]*m[3][2] - m[0][2]*m[1][1]*m[3][3] + m[0][1]*m[1][2]*m[3][3]) / d
	n[0][3] = (m[0][3]*m[1][2]*m[2][1] - m[0][2]*m[1][3]*m[2][1] - m[0][3]*m[1][1]*m[2][2] + m[0][1]*m[1][3]*m[2][2] + m[0][2]*m[1][1]*m[2][3] - m[0][1]*m[1][2]*m[2][3]) / d
	n[1][0] = (m[1][3]*m[2][2]*m[3][0] - m[1][2]*m[2][3]*m[3][0] - m[1][3]*m[2][0]*m[3][2] + m[1][0]*m[2][3]*m[3][2] + m[1][2]*m[2][0]*m[3][3] - m[1][0]*m[2][2]*m[3][3]) / d
	n[1][1] = (m[0][2]*m[2][3]*m[3][0] - m[0][3]*m[2][2]*m[3][0] + m[0][3]*m[2][0]*m[3][2] - m[0][0]*m[2][3]*m[3][2] - m[0][2]*m[2][0]*m[3][3] + m[0][0]*m[2][2]*m[3][3]) / d
	n[1][2] = (m[0][3]*m[1][2]*m[3][0] - m[0][2]*m[1][3]*m[3][0] - m[0][3]*m[1][0]*m[3][2] + m[0][0]*m[1][3]*m[3][2] + m[0][2]*m[1][0]*m[3][3] - m[0][0]*m[1][2]*m[3][3]) / d
	n[1][3] = (m[0][2]*m[1][3]*m[2][0] - m[0][3]*m[1][2]*m[2][0] + m[0][3]*m[1][0]*m[2][2] - m[0][0]*m[1][3]*m[2][2] - m[0][2]*m[1][0]*m[2][3] + m[0][0]*m[1][2]*m[2][3]) / d
	n[2][0] = (m[1][1]*m[2][3]*m[3][0] - m[1][3]*m[2][1]*m[3][0] + m[1][3]*m[2][0]*m[3][1] - m[1][0]*m[2][3]*m[3][1] - m[1][1]*m[2][0]*m[3][3] + m[1][0]*m[2][1]*m[3][3]) / d
	n[2][1] = (m[0][3]*m[2][1]*m[3][0] - m[0][1]*m[2][3]*m[3][0] - m[0][3]*m[2][0]*m[3][1] + m[0][0]*m[2][3]*m[3][1] + m[0][1]*m[2][0]*m[3][3] - m[0][0]*m[2][1]*m[3][3]) / d
	n[2][2] = (m[0][1]*m[1][3]*m[3][0] - m[0][3]*m[1][1]*m[3][0] + m[0][3]*m[1][0]*m[3][1] - m[0][0]*m[1][3]*m[3][1] - m[0][1]*m[1][0]*m[3][3] + m[0][0]*m[1][1]*m[3][3]) / d
	n[2][3] = (m[0][3]*m[1][1]*m[2][0] - m[0][1]*m[1][3]*m[2][0] - m[0][3]*m[1][0]*m[2][1] + m[0][0]*m[1][3]*m[2][1] + m[0][1]*m[1][0]*m[2][3] - m[0][0]*m[1][1]*m[2][3]) / d
	n[3][0] = (m[1][2]*m[2][1]*m[3][0] - m[1][1]*m[2][2]*m[3][0] - m[1][2]*m[2][0]*m[3][1] + m[1][0]*m[2][2]*m[3][1] + m[1][1]*m[2][0]*m[3][2] - m[1][0]*m[2][1]*m[3][2]) / d
	n[3][1] = (m[0][1]*m[2][2]*m[3][0] - m[0][2]*m[2][1]*m[3][0] + m[0][2]*m[2][0]*m[3][1] - m[0][0]*m[2][2]*m[3][1] - m[0][1]*m[2][0]*m[3][2] + m[0][0]*m[2][1]*m[3][2]) / d
	n[3][2] = (m[0][2]*m[1][1]*m[3][0] - m[0][1]*m[1][2]*m[3][0] - m[0][2]*m[1][0]*m[3][1] + m[0][0]*m[1][2]*m[3][1] + m[0][1]*m[1][0]*m[3][2] - m[0][0]*m[1][1]*m[3][2]) / d
	n[3][3] = (m[0][1]*m[1][2]*m[2][0] - m[0][2]*m[1][1]*m[2][0] + m[0][2]*m[1][0]*m[2][1] - m[0][0]*m[1][2]*m[2][1] - m[0][1]*m[1][0]*m[2][2] + m[0][0]*m[1][1]*m[2][2]) / d

	return n
}

type Mat2x3 [2]Vec3f

func (m Mat2x3) M() Matrix {
	mat := make([]Vecf, 2)
	for i := range m {
		mat[i] = m[i][:]
	}
	return mat
}

// When returned matrix changes, the original matrix will also change accordingly
func (m *Mat2x3) Mm() Matrix {
	mat := make([]Vecf, 2)
	for i := range m {
		mat[i] = m[i][:]
	}
	return mat
}

func (m *Mat2x3) SetCol(idx int, v Vec2f) {
	r, c := 2, 3
	if idx >= c {
		panic("column index greater than the number of columns")
	}
	for i := range r {
		m[i][idx] = v[i]
	}
}

func (m Mat2x3) Mulv(v Vec3f) Vec2f {
	u := Vec2f{}
	for i := range 2 {
		u[i] = m[i].Dot(v)
	}
	return u
}

type Mat3x2 [3]Vec2f

type Mat4x3 [4]Vec3f

func (m *Mat4x3) SetCol(idx int, v Vec4f) {
	r, c := 4, 3
	if idx >= c {
		panic("column index greater than the number of columns")
	}
	for i := range r {
		m[i][idx] = v[i]
	}
}

func (m Mat4x3) Transpose() Mat3x4 {
	r, c := 4, 3
	n := Mat3x4{}
	for i := range r {
		for j := range c {
			n[j][i] = m[i][j]
		}
	}
	return n
}

type Mat3x4 [3]Vec4f

type Matrix []Vecf

func NewMatrix(rows, cols int) Matrix {
	m := make(Matrix, rows)
	for i := range m {
		m[i] = make(Vecf, cols)
	}
	return m
}

func NewSquareMatrix(n int) Matrix {
	return NewMatrix(n, n)
}

func Identity(dimensions int) Matrix {
	m := NewMatrix(dimensions, dimensions)
	for i := range dimensions {
		for j := range dimensions {
			if i == j {
				m[i][j] = 1
			}
		}
	}
	return m
}

func (m Matrix) V() Vec3f {
	return Vec3f{m[0][0] / m[3][0], m[1][0] / m[3][0], m[2][0] / m[3][0]}
}

func (m Matrix) Nrows() int {
	return len(m)
}

func (m Matrix) Ncols() int {
	if len(m) == 0 {
		return 0
	} else {
		return len(m[0])
	}
}

func (m Matrix) Row(i int) Vecf {
	return m[i]
}

func (m Matrix) Col(j int) Vecf {
	r := m.Nrows()
	v := make(Vecf, r)
	for i := range m {
		v[i] = m[i][j]
	}
	return v
}

func (m Matrix) SetCol(idx int, v Vecf) {
	r, c := m.Nrows(), m.Ncols()
	if idx >= c {
		panic("column index greater than the number of columns")
	}
	for i := range r {
		m[i][idx] = v[i]
	}
}

func (m Matrix) IsMatrix() bool {
	if m.Nrows() == 0 {
		return false
	}
	for _, row := range m {
		if len(m[0]) != len(row) {
			return false
		}
	}
	return true
}

func (m Matrix) IsSquare() bool {
	return m.Ncols() == m.Nrows()
}

func (m Matrix) Clone() Matrix {
	n := make(Matrix, len(m))
	for i := range m {
		n[i] = make([]float64, len(m[i]))
		copy(n[i], m[i])
	}
	return n
}

func (m Matrix) Add(n Matrix) Matrix {
	r, c := m.Nrows(), m.Ncols()
	if r != n.Nrows() || c != n.Ncols() {
		panic("not homogeneous matrix")
	}
	a := m.Clone()
	for i := range r {
		for j := range c {
			a[i][j] += n[i][j]
		}
	}
	return a
}

func (m Matrix) Sub(n Matrix) Matrix {
	r, c := m.Nrows(), m.Ncols()
	if r != n.Nrows() || c != n.Ncols() {
		panic("not homogeneous matrix")
	}
	a := m.Clone()
	for i := range r {
		for j := range c {
			a[i][j] -= n[i][j]
		}
	}
	return a
}

func (m Matrix) Mul(n Matrix) Matrix {
	mr, mc := m.Nrows(), m.Ncols()
	nr, nc := n.Nrows(), n.Ncols()

	if mc != nr {
		panic("two matrices cannot be multiplied")
	}
	a := NewMatrix(mr, nc)

	for i := range mr {
		for j := range nc {
			for k := range mc {
				a[i][j] += m[i][k] * n[k][j]
			}
		}
	}
	return a
}

func (m Matrix) Muln(t float64) Matrix {
	r, c := m.Nrows(), m.Ncols()
	a := m.Clone()
	for i := range r {
		for j := range c {
			a[i][j] *= t
		}
	}
	return a
}

func (m Matrix) Mulv(v Vecf) Vecf {
	if m.Ncols() != len(v) {
		panic("column of matrix must be same as length of vector")
	}

	r, c := m.Nrows(), m.Ncols()
	u := make(Vecf, r)
	for i := range r {
		for j := range c {
			u[i] += v[j] * m[i][j]
		}
	}
	return u
}

func (m Matrix) Divn(n float64) Matrix {
	r, c := m.Nrows(), m.Ncols()
	a := m.Clone()
	for i := range r {
		for j := range c {
			a[i][j] /= n
		}
	}
	return a
}

func (m Matrix) Transpose() Matrix {
	r, c := m.Nrows(), m.Ncols()
	n := NewMatrix(c, r)
	for i := range r {
		for j := range c {
			n[j][i] = m[i][j]
		}
	}
	return n
}

func (m Matrix) Determinant() float64 {
	r, c := m.Nrows(), m.Ncols()
	if r != c {
		panic("not square matrix")
	}

	if len(m) == 1 {
		return m[0][0]
	}

	if len(m) == 2 {
		return m[0][0]*m[1][1] - m[0][1]*m[1][0]
	}

	ret := 0.0
	for j := range c {
		ret += m[0][j] * m.Cofactor(0, j)
	}

	return ret
}

func (m Matrix) SubMatrix(x, y int) Matrix {
	r, c := m.Nrows(), m.Ncols()
	n := NewMatrix(r-1, c-1)
	for i := range r {
		for j := range c {
			if i < x && j < y {
				n[i][j] = m[i][j]
			} else if i < x && j > y {
				n[i][j-1] = m[i][j]
			} else if i > x && j < y {
				n[i-1][j] = m[i][j]
			} else if i > x && j > y {
				n[i-1][j-1] = m[i][j]
			}
		}
	}
	return n
}

func (m Matrix) Minor(x, y int) float64 {
	return m.SubMatrix(x, y).Determinant()
}

func (m Matrix) Cofactor(x, y int) float64 {
	return math.Pow(-1, float64(x+y)) * m.Minor(x, y)
}

func (m Matrix) Adjugate() Matrix {
	r, c := m.Nrows(), m.Ncols()
	a := NewMatrix(r, c)
	for i := range r {
		for j := range c {
			a[i][j] = m.Cofactor(j, i)
		}
	}
	return a
}

func (m Matrix) Inverse() Matrix {
	r, c := m.Nrows(), m.Ncols()
	if r != c {
		panic("not square matrix")
	}

	d := m.Determinant()
	if d == 0 {
		panic("determinant is 0, can't be inversed")
	}
	n := m.Adjugate().Divn(d)
	return n
}

func (m Matrix) Inverse2() Matrix {
	r, c := m.Nrows(), m.Ncols()
	if r != c {
		panic("not square matrix")
	}

	// augmenting the square matrix with the identity matrix of the same dimensions a => [ai]
	n := NewMatrix(r, c*2)
	for i := range r {
		for j := range c {
			n[i][j] = m[i][j]
		}
	}
	for i := range r {
		n[i][i+c] = 1
	}
	// first pass
	for i := range r - 1 {
		// normalize the first row
		for j := 2*c - 1; j >= 0; j-- {
			n[i][j] /= n[i][i]
		}
		for k := i + 1; k < r; k++ {
			coeff := n[k][i]
			for h := 0; h < 2*c; h++ {
				n[k][h] -= n[i][h] * coeff
			}
		}
	}
	// normalize the last row
	for j := 2*c - 1; j >= r-1; j-- {
		n[r-1][j] /= n[r-1][r-1]
	}
	// second pass
	for i := r - 1; i > 0; i-- {
		for k := i - 1; k >= 0; k-- {
			coeff := n[k][i]
			for j := 0; j < 2*c; j++ {
				n[k][j] -= n[i][j] * coeff
			}
		}
	}
	// cut the identity matrix back
	t := NewMatrix(r, c)
	for i := range r {
		for j := range c {
			t[i][j] = n[i][j+c]
		}
	}

	return t
}

func LookAt(eye, center, up Vec3f) Mat4 {
	z := eye.Sub(center).Normalize()
	x := up.Cross(z).Normalize()
	y := z.Cross(x).Normalize()
	m := ID4()
	tr := ID4()

	for i := range 3 {
		m[0][i] = x[i]
		m[1][i] = y[i]
		m[2][i] = z[i]
		tr[i][3] = -center[i]
	}
	return m.Mul(tr)
}

func Projection(coeff float64) Mat4 {
	m := ID4()
	m[3][2] = coeff
	return m
}

func Viewport(x, y, w, h int) Mat4 {
	m := ID4()
	m[0][3] = float64(x) + float64(w)/2
	m[1][3] = float64(y) + float64(h)/2
	m[2][3] = 1
	m[0][0] = float64(w) / 2
	m[1][1] = float64(h) / 2
	m[2][2] = 0
	return m
}

func RotationX(angle float64) Mat4 {
	sinangle, cosangle := math.Sincos(angle)

	m := ID4()
	m[1][1] = cosangle
	m[2][2] = cosangle
	m[1][2] = -sinangle
	m[2][1] = sinangle
	return m
}

func RotationY(angle float64) Mat4 {
	sinangle, cosangle := math.Sincos(angle)
	m := ID4()
	m[0][0] = cosangle
	m[2][2] = cosangle
	m[0][2] = sinangle
	m[2][0] = -sinangle
	return m
}

func RotationZ(angle float64) Mat4 {
	sinangle, cosangle := math.Sincos(angle)
	m := ID4()
	m[0][0] = cosangle
	m[1][1] = cosangle
	m[0][1] = -sinangle
	m[1][0] = sinangle
	return m
}

func Zoom(factor float64) Mat4 {
	m := ID4()
	m[0][0] = factor
	m[1][1] = factor
	m[2][2] = factor
	return m
}

func LookAtMatrix(eye, center, up Vec3f) Mat4 {
	up = up.Normalize()
	f := center.Sub(eye).Normalize()
	s := f.Cross(up).Normalize()
	u := s.Cross(f)

	m := Mat4{
		{s.X(), u.X(), -f.X(), 0},
		{s.Y(), u.Y(), -f.Y(), 0},
		{s.Z(), u.Z(), -f.Z(), 0},
		{-s.Dot(eye), -u.Dot(eye), f.Dot(eye), 1},
	}

	return m
}

func Perspective(fovy, aspect, znear, zfar float64) Mat4 {
	tanHalfFovy := math.Tan(fovy / 2)
	m := Mat4{}
	m[0][0] = 1 / (aspect * tanHalfFovy)
	m[1][1] = 1 / tanHalfFovy
	m[2][2] = -(zfar + znear) / (zfar - znear)
	m[2][3] = -1
	m[3][2] = -(zfar * znear) / (zfar - znear)
	return m
}

func Translate(v Vec3f) Mat4 {
	m := ID4()
	m[3][0] = v.X()
	m[3][1] = v.Y()
	m[3][2] = v.Z()
	return m
}

func (m Mat4) Translate(v Vec3f) Mat4 {
	return Translate(v).Mul(m)
}

func Rotate(v Vec3f, a float64) Mat4 {
	v = v.Normalize()
	s, c := math.Sincos(a)
	m := 1 - c

	return Mat4{
		{m*v.X()*v.X() + c, m*v.X()*v.Y() + v.Z()*s, m*v.Z()*v.X() - v.Y()*s, 0},
		{m*v.X()*v.Y() - v.Z()*s, m*v.Y()*v.Y() + c, m*v.Y()*v.Z() + v.X()*s, 0},
		{m*v.Z()*v.X() + v.Y()*s, m*v.Y()*v.Z() - v.X()*s, m*v.Z()*v.Z() + c, 0},
		{0, 0, 0, 0}}
}

func (m Mat4) Rotate(a float64, v Vec3f) Mat4 {
	r := Rotate(v, a).Mul(m)
	r[3] = m[3]
	return r
}

func Clamp(val float64, min float64, max float64) float64 {
	if val < min {
		return min
	} else if val > max {
		return max
	}
	return val
}

func Radians(angle float64) float64 {
	return angle * math.Pi / 180
}
