package main

import (
	"bufio"
	"fmt"
	"image/color"
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
	verts       []Vec3f
	faces       [][]Vec3i
	norms       []Vec3f
	uvs         []Vec2f
	diffusemap  *TGAImage
	normalmap   *TGAImage
	specularmap *TGAImage
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

	obj.loadTexture("_diffuse.tga", &obj.diffusemap)
	obj.loadTexture("_nm.tga", &obj.normalmap)
	obj.loadTexture("_spec.tga", &obj.specularmap)

	return nil
}

func (obj *Obj) loadTexture(typename string, img **TGAImage) (err error) {
	completeName := basename + typename
	*img, err = LoadTGAImage(filepath.Join(dirname, completeName))
	if err != nil {
		fmt.Fprintf(os.Stderr, "texture file %s loading failed\n", completeName)
		return
	}

	(*img).FlipVertically()
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

func (obj Obj) Vertf(iface, nvert int) Vec3f {
	return obj.verts[obj.faces[iface][nvert][0]]
}

func (obj Obj) UV(iface, nvert int) Vec2i {
	if obj.diffusemap == nil {
		return Vec2i{}
	}

	idx := obj.faces[iface][nvert][1]
	return Vec2i{int(obj.uvs[idx].X() * float64(obj.diffusemap.Width)), int(obj.uvs[idx].Y() * float64(obj.diffusemap.Height))}
}

func (obj Obj) UVf(iface, nthvert int) Vec2f {
	return obj.uvs[obj.faces[iface][nthvert][1]]
}

func (obj Obj) Diffuse(uv Vec2i) color.RGBA {
	return obj.diffusemap.At(uv.X(), uv.Y())
}

func (obj Obj) Diffusef(uvf Vec2f) color.RGBA {
	uv := Vec2i{int(uvf[0] * float64(obj.diffusemap.Width)), int(uvf[1] * float64(obj.diffusemap.Height))}
	return obj.diffusemap.At(uv.X(), uv.Y())
}

func (obj Obj) Norm(iface, nvert int) Vec3f {
	idx := obj.faces[iface][nvert][2]
	return obj.norms[idx].Normalize()
}

func (obj Obj) NormUV(uvf Vec2f) Vec3f {
	uv := Vec2i{int(uvf[0] * float64(obj.normalmap.Width)), int(uvf[1] * float64(obj.normalmap.Height))}
	c := obj.normalmap.At(uv.X(), uv.Y())

	res := Vec3f{}
	res[2] = float64(c.B)/255.*2. - 1.
	res[1] = float64(c.G)/255.*2. - 1.
	res[0] = float64(c.R)/255.*2. - 1.
	return res
}

func (obj Obj) Specular(uvf Vec2f) float64 {
	uv := Vec2i{int(uvf[0] * float64(obj.specularmap.Width)), int(uvf[1] * float64(obj.specularmap.Height))}
	return float64(obj.specularmap.At(uv.X(), uv.Y()).R) / 1.0
}

func parseFloats(items []string) []float64 {
	result := make([]float64, len(items))
	for i, item := range items {
		f, _ := strconv.ParseFloat(item, 64)
		result[i] = f
	}
	return result
}
