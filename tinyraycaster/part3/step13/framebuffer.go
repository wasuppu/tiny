package main

import "image/color"

type Framebuffer struct {
	w, h int           // image dimensions
	ps   []color.Color // storage container
}

func NewFramebuffer(winW, winH int, c color.Color) *Framebuffer {
	fb := Framebuffer{w: winW, h: winH}
	fb.ps = make([]color.Color, fb.w*fb.h)
	fb.Clear(c)
	return &fb
}

func (fb *Framebuffer) SetPixel(x, y int, c color.Color) {
	fb.ps[x+y*fb.w] = c
}

func (fb *Framebuffer) DrawRectangle(rectX, rectY, rectW, rectH int, c color.Color) {
	for i := range rectW {
		for j := range rectH {
			cx := rectX + i
			cy := rectY + j
			if cx < fb.w && cy < fb.h { // no need to check negative values (unsigned variables)
				fb.SetPixel(cx, cy, c)
			}
		}
	}
}

func (fb *Framebuffer) Clear(c color.Color) {
	for i := range fb.ps {
		fb.ps[i] = c
	}
}
