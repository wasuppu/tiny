package main

import (
	"encoding/binary"
	"fmt"
	"image/color"
	"io"
	"math"
	"os"
)

const (
	GRAYSCALE = 1
	RGB       = 3
	RGBA      = 4
)

type TGAHeader struct {
	IdLength        byte
	ColorMapType    byte
	DataTypeCode    byte
	ColorMapOrigin  int16 // Color map specification begin
	ColorMapLength  int16
	ColorMapDepth   byte
	XOrigin         int16 // Image specification begin
	YOrigin         int16
	Width           int16
	Height          int16
	BitsPerPixel    byte
	ImageDescriptor byte
}

func LoadTGAImage(filename string) (*TGAImage, error) {
	tga := TGAImage{}
	err := tga.read(filename)
	if err != nil {
		return nil, err
	}
	return &tga, nil
}

func NewTgaImg(width, height int) *TGAImage {
	tga := TGAImage{Width: width, Height: height, bytespp: RGB, data: make([]byte, width*height*RGB)}
	tga.setColorMethod(RGB << 3)
	return &tga
}

func NewGreyTgaImage(width, height int) *TGAImage {
	tga := TGAImage{Width: width, Height: height, bytespp: GRAYSCALE, data: make([]byte, width*height*GRAYSCALE)}
	tga.setColorMethod(GRAYSCALE << 3)
	return &tga
}

type TGAImage struct {
	Width   int
	Height  int
	bytespp int
	UseRLE  bool
	data    []byte
	dec     func([]byte) color.RGBA
	enc     func([]byte, color.Color)
	Header  *TGAHeader
}

func (tga *TGAImage) Write(filename string) error {
	developerArea := [4]byte{0, 0, 0, 0}
	extensionArea := [4]byte{0, 0, 0, 0}
	footer := []byte("TRUEVISION-XFILE.\x00")

	file, err := os.Create(filename)
	if err != nil {
		return err
	}

	header := new(TGAHeader)
	header.BitsPerPixel = byte(tga.bytespp << 3)
	header.Width = int16(tga.Width)
	header.Height = int16(tga.Height)

	if tga.bytespp == GRAYSCALE && tga.UseRLE {
		header.DataTypeCode = 11
	} else if tga.bytespp == GRAYSCALE && !tga.UseRLE {
		header.DataTypeCode = 3
	} else if tga.bytespp != GRAYSCALE && tga.UseRLE {
		header.DataTypeCode = 10
	} else if tga.bytespp != GRAYSCALE && !tga.UseRLE {
		header.DataTypeCode = 2
	}

	header.ImageDescriptor = 0x20
	err = binary.Write(file, binary.LittleEndian, header)
	if err != nil {
		return fmt.Errorf("can't dump the tga file\n%v", err)
	}

	if !tga.UseRLE {
		err = binary.Write(file, binary.LittleEndian, tga.data)
		if err != nil {
			return fmt.Errorf("can't unload raw data\n%v", err)
		}
	} else {
		err = tga.unloadRleData(file)
		if err != nil {
			return fmt.Errorf("can't unload raw data\n%v", err)
		}
	}

	err = binary.Write(file, binary.LittleEndian, developerArea)
	if err != nil {
		return fmt.Errorf("can't dump the tga file\n%v", err)
	}
	err = binary.Write(file, binary.LittleEndian, extensionArea)
	if err != nil {
		return fmt.Errorf("can't dump the tga file\n%v", err)
	}
	err = binary.Write(file, binary.LittleEndian, footer)
	if err != nil {
		return fmt.Errorf("can't dump the tga file\n%v", err)
	}
	return nil
}

func (tga *TGAImage) read(filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	header := new(TGAHeader)
	err = binary.Read(file, binary.LittleEndian, header)
	if err != nil {
		return err
	}

	tga.Width = int(header.Width)
	tga.Height = int(header.Height)
	tga.bytespp = int(header.BitsPerPixel) >> 3
	if tga.Width <= 0 || tga.Height <= 0 || (tga.bytespp != GRAYSCALE && tga.bytespp != RGB && tga.bytespp != RGBA) {
		return fmt.Errorf("bad bpp (or width/height) value")
	}

	err = tga.setColorMethod(int(header.BitsPerPixel))
	if err != nil {
		return err
	}

	tga.Header = header

	nbytes := int(tga.bytespp) * int(tga.Width) * int(tga.Height)
	if header.DataTypeCode == 3 || header.DataTypeCode == 2 {
		tga.data = make([]byte, nbytes)
		_, err = io.ReadFull(file, tga.data[:])
		if err != nil {
			return fmt.Errorf("an error occured while reading the data")
		}
	} else if header.DataTypeCode == 10 || header.DataTypeCode == 11 {
		err = tga.loadRleData(file)
		if err != nil {
			return fmt.Errorf("an error occured while reading the data")
		}
	} else {
		return fmt.Errorf("unknown file format %d", header.DataTypeCode)
	}

	if header.ImageDescriptor&0x20 == 0 {
		tga.FlipVertically()
	}

	if header.ImageDescriptor&0x10 != 0 {
		tga.FlipHorizontally()
	}

	return nil
}

func (tga *TGAImage) loadRleData(r io.Reader) error {
	tga.UseRLE = true
	pixelCount := tga.Width * tga.Height
	currentpixel := 0
	colorbuf := make([]byte, tga.bytespp)
	for {
		buf := make([]byte, 1)
		_, err := r.Read(buf)
		chunkHeader := buf[0]
		if err != nil {
			return fmt.Errorf("an error occured while reading the data\n%v", err)
		}
		if chunkHeader < 128 {
			chunkHeader++
			for range int(chunkHeader) {
				_, err := r.Read(colorbuf)
				if err != nil {
					return fmt.Errorf("an error occured while reading the data\n%v", err)
				}

				for t := range tga.bytespp {
					tga.data = append(tga.data, colorbuf[t])
				}
				currentpixel++
				if currentpixel > pixelCount {
					return fmt.Errorf("too many pixels read")
				}
			}
		} else {
			chunkHeader -= 127
			_, err := r.Read(colorbuf)
			if err != nil {
				return fmt.Errorf("an error occured while reading the header\n%v", err)
			}
			for range int(chunkHeader) {
				for t := range tga.bytespp {
					tga.data = append(tga.data, colorbuf[t])
				}
				currentpixel++
				if currentpixel > pixelCount {
					return fmt.Errorf("too many pixels read")
				}
			}
		}

		if currentpixel >= pixelCount {
			break
		}
	}

	return nil
}

func (tga *TGAImage) unloadRleData(w io.Writer) (err error) {
	maxChunkLength := 128
	npixels := tga.Width * tga.Height
	curpix := 0
	j := 0
	for curpix < npixels {
		j++
		chunkStart := curpix * tga.bytespp
		curbyte := curpix * tga.bytespp
		runLength := 1
		raw := true

		for curpix+runLength < npixels && runLength < maxChunkLength {
			succEq := true

			for t := 0; succEq && t < tga.bytespp; t++ {
				succEq = (tga.data[curbyte+t] == tga.data[curbyte+t+tga.bytespp])
			}

			curbyte += tga.bytespp
			if runLength == 1 {
				raw = !succEq
			}

			if raw && succEq {
				runLength--
				break
			}

			if !raw && !succEq {
				break
			}

			runLength++
		}

		curpix += runLength
		if raw {
			_, err = w.Write([]byte{byte(runLength - 1)})
		} else {
			_, err = w.Write([]byte{byte(runLength + 127)})
		}
		if err != nil {
			return fmt.Errorf("can't dump the tga file\n%v", err)
		}

		if raw {
			_, err = w.Write(tga.data[chunkStart : chunkStart+runLength*tga.bytespp])
		} else {
			_, err = w.Write(tga.data[chunkStart : chunkStart+tga.bytespp])
		}

		if err != nil {
			return fmt.Errorf("can't dump the tga file\n%v", err)
		}
	}

	return nil
}

func (tga *TGAImage) FlipHorizontally() error {
	if len(tga.data) == 0 {
		return fmt.Errorf("tga data is empty")
	}
	bytesPerLine := tga.Width * tga.bytespp
	half := tga.Width >> 1
	for j := range tga.Height {
		for i := range half {
			k := (tga.Width - 1 - i)
			for b := range tga.bytespp {
				p1 := (i*tga.bytespp + b) + (j * bytesPerLine)
				p2 := (k*tga.bytespp + b) + (j * bytesPerLine)
				tga.data[p1], tga.data[p2] = tga.data[p2], tga.data[p1]
			}
		}
	}
	return nil
}

func (tga *TGAImage) FlipVertically() error {
	if len(tga.data) == 0 {
		return fmt.Errorf("tga data is empty")
	}
	bytesPerLine := tga.Width * tga.bytespp
	half := tga.Height >> 1
	for j := range half {
		for i := range tga.Width {
			k := (tga.Height - 1 - j)
			for b := range tga.bytespp {
				p1 := (i*tga.bytespp + b) + (j * bytesPerLine)
				p2 := (i*tga.bytespp + b) + (k * bytesPerLine)
				tga.data[p1], tga.data[p2] = tga.data[p2], tga.data[p1]
			}
		}
	}
	return nil
}

func (tga *TGAImage) Bpp() int {
	return tga.bytespp
}

func (tga *TGAImage) Data() []byte {
	return tga.data
}

func (tga *TGAImage) At(x, y int) color.RGBA {
	if x >= tga.Width || y >= tga.Height || x < 0 || y < 0 {
		return color.RGBA{}
	}
	pos := (x + y*tga.Width) * tga.bytespp
	return tga.dec(tga.data[pos:])
}

func (tga *TGAImage) setColorMethod(bpp int) error {
	switch bpp {
	case 8:
		tga.dec = decodeGray
		tga.enc = encodeGray
	case 16:
		tga.dec = decode16
		tga.enc = encode16
	case 24:
		tga.dec = decode24
		tga.enc = encode24
	case 32:
		tga.dec = decode32
		tga.enc = encode32
	default:
		return fmt.Errorf("unsupported bpp size: %d", bpp)
	}
	return nil
}

func decodeGray(p []byte) color.RGBA {
	return color.RGBA{p[0], p[0], p[0], 255}
}

func decode16(p []byte) color.RGBA {
	r := (p[1] & 0x7C) << 1
	g := ((p[1] & 0x3) << 6) | ((p[0] & 0xE0) >> 2)
	b := (p[0] & 0x1F) << 3
	a := uint8((uint16(p[1]&0x80) << 1) - 1)
	return color.RGBA{r, g, b, a}
}

func decode24(p []byte) color.RGBA {
	return color.RGBA{p[2], p[1], p[0], 255}
}

func decode32(p []byte) color.RGBA {
	return color.RGBA{p[2], p[1], p[0], p[3]}
}

func (tga *TGAImage) Set(x, y int, c color.Color) error {
	if x >= tga.Width || y >= tga.Height || x < 0 || y < 0 {
		return fmt.Errorf("exceeding the length or width range")
	}
	pos := (x + y*tga.Width) * tga.bytespp
	rgba, _ := color.RGBAModel.Convert(c).(color.RGBA)
	tga.enc(tga.data[pos:], rgba)
	return nil
}

func encodeGray(p []byte, pixel color.Color) {
	c, ok := color.RGBAModel.Convert(pixel).(color.Gray)
	if ok {
		p[0] = c.Y
	} else {
		c := color.RGBAModel.Convert(pixel).(color.RGBA)
		g, _, _, _ := c.RGBA()
		p[0] = uint8(g)
	}
}

func encode16(p []byte, pixel color.Color) {
	c := color.RGBAModel.Convert(pixel).(color.RGBA)
	r, g, b, a := c.RGBA()
	v := (uint16(r)>>3)<<1 |
		(uint16(g)>>3)<<6 |
		(uint16(b)>>3)<<11 |
		(uint16(a) >> 7)
	binary.LittleEndian.PutUint16(p[:2], v)
}

func encode24(p []byte, pixel color.Color) {
	c := color.RGBAModel.Convert(pixel).(color.RGBA)
	r, g, b, _ := c.RGBA()
	p[0], p[1], p[2] = byte(b), byte(g), byte(r)
}

func encode32(p []byte, pixel color.Color) {
	c := color.RGBAModel.Convert(pixel).(color.RGBA)
	r, g, b, a := c.RGBA()
	p[0], p[1], p[2], p[3] = byte(b), byte(g), byte(r), byte(a)
}

func (tga *TGAImage) Clear() {
	for i := range tga.data {
		tga.data[i] = 0
	}
}

func gaussianKernel(radius int) []float64 {
	size := radius*2 + 1
	norm := 1.0 / (math.Sqrt(2*math.Pi) * float64(radius))
	gaussianKernel := make([]float64, size)
	coeff := -1.0 / (2.0 * float64(radius) * float64(radius))

	sum := 0.0
	for i := range size {
		gaussianKernel[i] = norm * math.Exp(math.Pow(float64(i-radius), 2)*coeff)
		sum += gaussianKernel[i]
	}

	for i := range size {
		gaussianKernel[i] /= sum
	}

	return gaussianKernel
}

func (tga *TGAImage) GaussianBlur(radius int) {
	kernel := gaussianKernel(radius)
	tmp := tga
	size := radius*2 + 1

	for j := size; j < tga.Height; j++ {
		for i := range tga.Width {
			rgba := []float64{0, 0, 0, 0}
			for k := range size {
				c := tga.At(i, j-size+k)
				rgba[0] += float64(c.R) * kernel[k]
				rgba[1] += float64(c.G) * kernel[k]
				rgba[2] += float64(c.B) * kernel[k]
				rgba[3] += float64(c.A) * kernel[k]
			}
			ic := color.Gray{uint8(rgba[0])}
			tmp.Set(i, j, ic)
		}
	}

	for j := size; j < tga.Height; j++ {
		for i := range tga.Width {
			rgba := []float64{0, 0, 0, 0}
			for k := range size {
				c := tga.At(i-size+k, j)
				rgba[0] += float64(c.R) * kernel[k]
				rgba[1] += float64(c.G) * kernel[k]
				rgba[2] += float64(c.B) * kernel[k]
				rgba[3] += float64(c.A) * kernel[k]
			}
			ic := color.Gray{uint8(rgba[0])}
			tmp.Set(i, j, ic)
		}
	}
}
