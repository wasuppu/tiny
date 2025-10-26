package main

import (
	"image/color"
	"log"
	"path/filepath"
)

type Texture struct {
	w, h        int           // overall image dimensions
	count, size int           // number of textures and size in pixels
	ps          []color.Color // textures storage container
}

func NewTexture(filename string) *Texture {
	t := Texture{}
	err := t.LoadTexture(filename)
	if err != nil {
		log.Fatalf("%+v", err)
	}
	return &t
}

func (t *Texture) LoadTexture(filename string) error {
	texturepath := filepath.Join(rootpath, "textures", filename)
	img, err := openImg(texturepath)
	if err != nil {
		return err
	}
	t.w, t.h = img.Bounds().Dx(), img.Bounds().Dy()
	t.count = t.w / t.h
	t.size = t.w / t.count
	t.ps = readColorFromImg(img)
	return nil
}

func (t *Texture) TextureColumn(texid, texcoord, columnHeight int) []color.Color {
	column := make([]color.Color, columnHeight)
	for y := range columnHeight {
		pixX := texid*t.size + texcoord
		pixY := (y * t.size) / columnHeight
		column[y] = t.ps[pixX+pixY*t.w]
	}
	return column
}

// get the pixel (i,j) from the texture idx
func (t *Texture) Get(i, j, idx int) color.Color {
	return t.ps[i+idx*t.size+j*t.w]
}
