package main

import (
	"image/color"
	"log"
	"math"
	"path/filepath"
	"runtime"
	"sort"
)

var (
	basepath string
	rootpath string
)

func init() {
	_, exepath, _, _ := runtime.Caller(0)
	basepath = filepath.Dir(exepath)
	rootpath = filepath.Dir(filepath.Dir(basepath))
}

func drawSprite(sprite *Sprite, depthBuffer []float64, fb *Framebuffer, player *Player, texsprites *Texture) {
	// absolute direction from the player to the sprite (in radians)
	spriteDir := math.Atan2(sprite.y-player.y, sprite.x-player.x)
	for spriteDir-player.a > math.Pi {
		spriteDir -= 2 * math.Pi // remove unncesessary periods from the relative direction
	}

	for spriteDir-player.a < -math.Pi {
		spriteDir += 2 * math.Pi
	}

	spriteScreenSize := int(math.Min(1000, float64(int(float64(fb.h)/sprite.playerDist))))                               // screen sprite size
	hoffset := int((spriteDir-player.a)/player.fov*(float64(fb.w)/2) + (float64(fb.w)/2)/2 - float64(texsprites.size)/2) // do not forget the 3D view takes only a half of the framebuffer
	voffset := fb.h/2 - spriteScreenSize/2

	for i := range spriteScreenSize {
		if hoffset+i < 0 || hoffset+i >= fb.w/2 {
			continue
		}
		if depthBuffer[hoffset+i] < sprite.playerDist {
			continue // this sprite column is occluded
		}
		for j := range spriteScreenSize {
			if voffset+j < 0 || voffset+j >= fb.h {
				continue
			}
			c := texsprites.Get(i*texsprites.size/spriteScreenSize, j*texsprites.size/spriteScreenSize, sprite.texid)
			_, _, _, a := c.RGBA()
			if a > 128 {
				fb.SetPixel(fb.w/2+hoffset+i, voffset+j, c)
			}
		}
	}
}

func mapShowSprite(sprite *Sprite, fb *Framebuffer, m *Map) {
	rectW := fb.w / (m.w * 2) // size of one map cell on the screen
	rectH := fb.h / m.h
	fb.DrawRectangle(int(sprite.x*float64(rectW)-3), int(sprite.y*float64(rectH)-3), 6, 6, color.RGBA{255, 0, 0, 255})
}

func wallTexcoordX(cx, cy float64, texwalls *Texture) int {
	hitx := cx - math.Floor(cx+0.5) // hitx and hity contain (signed) fractional parts of cx and cy,
	hity := cy - math.Floor(cy+0.5) // they vary between -0.5 and +0.5, and one of them is supposed to be very close to 0
	texcoordX := int(hitx * float64(texwalls.size))
	if math.Abs(hity) > math.Abs(hitx) { // we need to determine whether we hit a "vertical" or a "horizontal" wall (w.r.t the map)
		texcoordX = int(hity * float64(texwalls.size))
	}
	if texcoordX < 0 {
		texcoordX += texwalls.size // do not forget x_texcoord can be negative, fix that
	}
	return texcoordX
}

func render(fb *Framebuffer, m *Map, player *Player, sprites []Sprite, texwalls *Texture, texmonst *Texture) {
	fb.Clear(color.RGBA{255, 255, 255, 255}) // clear the screen

	rectW := fb.w / (m.w * 2) // size of one map cell on the screen
	rectH := fb.h / m.h
	for j := range m.h { // draw the map
		for i := range m.w {
			if m.IsEmpty(i, j) { // skip empty spaces
				continue
			}
			rectX := i * rectW
			rectY := j * rectH
			texid := m.Get(i, j)
			fb.DrawRectangle(rectX, rectY, rectW, rectH, texwalls.Get(0, 0, texid)) //  the color is taken from the upper left pixel of the texture #texid
		}
	}

	depthBuffer := make([]float64, fb.w/2)
	for i := range depthBuffer {
		depthBuffer[i] = 1e3
	}

	for i := range fb.w / 2 { // draw the visibility cone AND the "3D" view
		angle := player.a - player.fov/2 + player.fov*float64(i)/(float64(fb.w)/2)
		for t := 0.0; t < 20; t += .01 { // ray marching loop
			x := player.x + t*math.Cos(angle)
			y := player.y + t*math.Sin(angle)
			fb.SetPixel(int(x*float64(rectW)), int(y*float64(rectH)), color.RGBA{160, 160, 160, 255}) // this draws the visibility cone

			if m.IsEmpty(int(x), int(y)) {
				continue
			}

			texid := m.Get(int(x), int(y)) // our ray touches a wall, so draw the vertical column to create an illusion of 3D
			dist := t * math.Cos(angle-player.a)
			depthBuffer[i] = dist
			columnHeight := int(float64(fb.h) / dist)
			texcoordX := wallTexcoordX(x, y, texwalls)
			column := texwalls.TextureColumn(texid, texcoordX, columnHeight)
			pixX := fb.w/2 + i            // we are drawing at the right half of the screen, thus +fb.w/2
			for j := range columnHeight { // copy the texture column to the framebuffer
				pixY := j + fb.h/2 - columnHeight/2
				if pixY < 0 || pixY >= fb.h {
					continue
				}
				fb.SetPixel(pixX, pixY, column[j])
			}
			break
		} // ray marching loop
	} // field of view ray sweeping

	for i := range len(sprites) { // update the distances from the player to each sprite
		sprites[i].playerDist = math.Sqrt(math.Pow(player.x-sprites[i].x, 2) + math.Pow(player.y-sprites[i].y, 2))
	}
	sort.Slice(sprites, func(i, j int) bool {
		return sprites[i].playerDist > sprites[j].playerDist
	})

	for i := range len(sprites) { // draw the sprites
		mapShowSprite(&sprites[i], fb, m)
		drawSprite(&sprites[i], depthBuffer, fb, player, texmonst)
	}
}

func main() {
	fb := NewFramebuffer(1024, 512, color.RGBA{255, 255, 255, 255})
	player := Player{3.456, 2.345, 1.523, math.Pi / 3}
	m := NewMap()
	texwalls := NewTexture("walltext.png")
	texmonst := NewTexture("monsters.png")
	if texwalls.count == 0 || texmonst.count == 0 {
		log.Fatalln("Failed to load textures")
	}
	sprites := []Sprite{{3.523, 3.812, 2, 0}, {1.834, 8.765, 0, 0}, {5.323, 5.365, 1, 0}, {4.123, 10.265, 1, 0}}

	render(fb, m, &player, sprites, texwalls, texmonst)
	writePng("out", fb.ps, fb.w, fb.h)
}
