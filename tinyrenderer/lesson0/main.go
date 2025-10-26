package main

import (
	"image/color"
	"path/filepath"
	"runtime"
)

var (
	white      = color.RGBA{255, 255, 255, 255}
	red        = color.RGBA{255, 0, 0, 255}
	basepath   string
	parentpath string
)

func init() {
	_, exepath, _, _ := runtime.Caller(0)
	basepath = filepath.Dir(exepath)
	parentpath = filepath.Dir(basepath)
}

func main() {
	tga := NewTgaImg(100, 100)
	tga.Set(52, 41, red)
	tga.FlipVertically()
	tga.Write(filepath.Join(parentpath, "lesson0.tga"))
}
