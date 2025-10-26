package main

import (
	"bufio"
	"image/color"
	"os"
	"path/filepath"

	"strconv"
	"strings"
)

type Obj struct {
	verts      []Vec3f
	faces      [][]Vec3i
	norms      []Vec3f
	uvs        []Vec2f
	diffusemap *TGAImage
}

func LoadObj(filename string) *Obj {
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
			fs := make([]Vec3i, len(args))
			for i, arg := range args {
				vertex := strings.Split(arg, "/")
				for j := range 3 {
					fs[i][j], _ = strconv.Atoi(vertex[j])
					fs[i][j]-- // in wavefront obj all indices start at 1, not zero
				}
			}
			obj.faces = append(obj.faces, fs)
		case "vn": // normal
			f := parseFloats(args)
			n := Vec3f{}
			for i := range 3 {
				n[i] = f[i]
			}
			obj.norms = append(obj.norms, n)
		case "vt": // texture
			f := parseFloats(args)
			uv := Vec2f{}
			for i := range 2 {
				uv[i] = f[i]
			}
			obj.uvs = append(obj.uvs, uv)
		}
	}

	dir := filepath.Dir(filename)
	baseName := strings.TrimSuffix(filepath.Base(filename), filepath.Ext(filename))
	obj.loadTexture(filepath.Join(dir, baseName+"_diffuse.tga"))
	return nil
}

func (obj *Obj) loadTexture(filename string) (err error) {
	obj.diffusemap, err = LoadTGAImage(filename)
	if err != nil {
		return
	}
	obj.diffusemap.FlipVertically()
	return nil
}

func (obj Obj) NVerts() int {
	return len(obj.verts)
}

func (obj Obj) NFaces() int {
	return len(obj.faces)
}

func (obj Obj) Face(idx int) []int {
	var face []int
	for i := range len(obj.faces[idx]) {
		face = append(face, obj.faces[idx][i][0])
	}
	return face
}

func (obj Obj) Vert(i int) Vec3f {
	return obj.verts[i]
}

func (obj Obj) UV(iface, nvert int) Vec2i {
	if obj.diffusemap == nil {
		return Vec2i{}
	}

	idx := obj.faces[iface][nvert][1]
	return Vec2i{int(obj.uvs[idx].X() * float64(obj.diffusemap.Width)), int(obj.uvs[idx].Y() * float64(obj.diffusemap.Height))}
}

func (obj Obj) Diffuse(uv Vec2i) color.RGBA {
	return obj.diffusemap.At(uv.X(), uv.Y())
}

func parseFloats(items []string) []float64 {
	result := make([]float64, len(items))
	for i, item := range items {
		f, _ := strconv.ParseFloat(item, 64)
		result[i] = f
	}
	return result
}
