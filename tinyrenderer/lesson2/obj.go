package main

import (
	"bufio"
	"os"

	"strconv"
	"strings"
)

type Obj struct {
	verts []Vec3f
	faces [][]int
}

func LoadObj(filename string) *Obj {
	obj := Obj{}
	obj.read(filename)
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
			obj.verts = append(obj.verts, Vec3f{f[0], f[1], f[2]})
		case "f": // face
			fs := make([]int, len(args))
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

func (obj Obj) Face(i int) []int {
	return obj.faces[i]
}

func (obj Obj) Vert(i int) Vec3f {
	return obj.verts[i]
}

func parseFloats(items []string) []float64 {
	result := make([]float64, len(items))
	for i, item := range items {
		f, _ := strconv.ParseFloat(item, 64)
		result[i] = f
	}
	return result
}
