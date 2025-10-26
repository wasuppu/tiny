package main

import (
	"bufio"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

var (
	basename string
	dirname  string
)

type Obj struct {
	verts    []Vec3f
	faces    []Vec3i
	material Material
}

func LoadObj(filename string) *Obj {
	dirname = filepath.Dir(filename)
	basename = strings.TrimSuffix(filepath.Base(filename), filepath.Ext(filename))
	obj := Obj{}
	err := obj.read(filename)
	if err != nil {
		panic(err)
	}
	return &obj
}

func (obj *Obj) read(filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		fields := strings.Fields(line)
		if len(fields) == 0 {
			continue
		}

		key := fields[0]
		args := fields[1:]
		switch key {
		case "v": // vertex
			f := parseFloats(args)
			v := Vec3f{}
			for i := range 3 {
				v[i] = f[i]
			}
			obj.verts = append(obj.verts, v)
		case "f": // face
			fs := Vec3i{}
			for i, arg := range args {
				vertex := strings.Split(arg, "/")
				fs[i], _ = strconv.Atoi(vertex[0])
				fs[i]-- // in wavefront obj all indices start at 1, not zero
			}
			obj.faces = append(obj.faces, fs)
		}
	}

	return nil
}

func (obj Obj) NVerts() int {
	return len(obj.verts)
}

func (obj Obj) NFaces() int {
	return len(obj.faces)
}

func (obj Obj) Point(i int) Vec3f {
	return obj.verts[i]
}

func (obj Obj) Vert(fi, li int) int {
	return obj.faces[fi][li]
}

func (obj Obj) rayIntersect(orig, dir Vec3f) (bool, float64, Material, Vec3f, Vec3f) {
	dist := math.MaxFloat64
	var t0 float64
	if obj.rayAABBIntersect(orig, dir) {
		for i := range obj.faces {
			if obj.rayTriangleIntersect(i, orig, dir, &t0) && t0 < dist {
				face := obj.faces[i]
				hit := orig.Add(dir.Muln(t0))
				n := obj.Point(face[1]).Sub(obj.Point(face[0])).Cross(obj.Point(face[2]).Sub(obj.Point(face[0]))).Normalize()
				return true, t0, obj.material, hit, n
			}
		}
	}
	return false, math.MaxFloat64, Material{}, Vec3f{}, Vec3f{}
}

func (obj Obj) rayTriangleIntersect(fi int, orig, dir Vec3f, tnear *float64) bool {
	edge1 := obj.Point(obj.Vert(fi, 1)).Sub(obj.Point(obj.Vert(fi, 0)))
	edge2 := obj.Point(obj.Vert(fi, 2)).Sub(obj.Point(obj.Vert(fi, 0)))
	pvec := dir.Cross(edge2)
	det := edge1.Dot(pvec)
	if det < 1e-5 {
		return false
	}

	tvec := orig.Sub(obj.Point(obj.Vert(fi, 0)))
	u := tvec.Dot(pvec)
	if u < 0 || u > det {
		return false
	}

	qvec := tvec.Cross(edge1)
	v := dir.Dot(qvec)
	if v < 0 || u+v > det {
		return false
	}

	*tnear = edge2.Dot(qvec) * (1 / det)
	return *tnear > 1e-5
}

func (obj Obj) rayAABBIntersect(orig, dir Vec3f) bool {
	min, max := obj.getBBox()

	min = min.Sub(orig).Divn(dir.X())
	max = max.Sub(orig).Divn(dir.X())

	if dir[0] < 0 {
		min[0], max[0] = max[0], min[0]
	}

	if dir[1] < 0 {
		min[1], max[1] = max[1], min[1]
	}

	if dir[2] < 0 {
		min[2], max[2] = max[2], min[2]
	}

	t0 := math.Max(min[0], math.Max(min[1], min[2]))
	t1 := math.Max(max[0], math.Max(max[1], max[2]))

	return t1 > t0 && t1 >= 0
}

func (obj Obj) getBBox() (Vec3f, Vec3f) {
	min := obj.verts[0]
	max := obj.verts[0]
	for i := 1; i < len(obj.verts); i++ {
		for j := range 3 {
			min[j] = math.Min(min[j], obj.verts[i][j])
			max[j] = math.Max(max[j], obj.verts[i][j])
		}
	}

	return min, max
}

func parseFloats(items []string) []float64 {
	result := make([]float64, len(items))
	for i, item := range items {
		f, _ := strconv.ParseFloat(item, 64)
		result[i] = f
	}
	return result
}
