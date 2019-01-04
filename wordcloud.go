package wordcloud

import (
	"errors"
	"image"
	"image/color"
	"io/ioutil"
	"math/rand"

	"github.com/disintegration/imaging"

	"github.com/golang/freetype"

	"github.com/golang/freetype/truetype"
	"golang.org/x/image/font"
)

type Word struct {
	Text  string
	Count float64
	Size  float64
}

type WordCloud struct {
	w, h         int
	font         *truetype.Font
	drawer       *font.Drawer // cache until font size no longer matches
	prevFontSize float64      // for caching the font by size
	spiraller    *Spiraller
	color        color.RGBA
	img          *image.NRGBA
}

type GeneratorOptions struct {
	Rand          *rand.Rand
	Angles        []float64
	FontPath      string
	Width, Height int
	MinSize       float64
	Words         []*Word
}

type rect struct {
	StartX int
	StartY int
	EndX   int
	EndY   int
}

func (wc *WordCloud) rect(x0, y0, x1, y1 int) *rect {
	return &rect{
		StartX: x0,
		EndX:   x1,
		StartY: y0,
		EndY:   y1,
	}
}

func (wc *WordCloud) collides(r *rect) bool {
	for x := r.StartX; x < r.EndX; x++ {
		for y := r.StartY; y < r.EndY; y++ {
			if (wc.img.Pix[wc.img.PixOffset(x, y)+3]) != 0 {
				return true
			}
		}
	}
	return false
}

func Crop(img *image.NRGBA) *image.NRGBA {
	minX, minY, maxX, maxY := -1, -1, -1, -1
	for x := 0; x < img.Bounds().Max.X; x++ {
		for y := 0; y < img.Bounds().Max.Y; y++ {
			if (img.Pix[img.PixOffset(x, y)+3]) != 0 {
				if minX == -1 || x < minX {
					minX = x
				}
				if maxX == -1 || x > maxX {
					maxX = x
				}
				if minY == -1 || y < minY {
					minY = y
				}
				if maxY == -1 || y > maxY {
					maxY = y
				}
			}
		}
	}

	return imaging.Crop(img, image.Rect(minX, minY, maxX+1, maxY+1))
}

func (wc *WordCloud) getNextPosition(w, h int) *image.Point {
	var (
		centerX = w / 2
		centerY = h / 2
		rect    *rect
	)

	for {
		pos := wc.spiraller.Next()
		if pos == nil {
			return nil
		}

		if pos.X-centerX < 0 ||
			pos.X+centerX >= wc.w ||
			pos.Y-centerY < 0 ||
			pos.Y+centerY >= wc.h {
			continue
		}

		rect = wc.rect(pos.X-centerX, pos.Y-centerY, pos.X+centerX, pos.Y+centerY)

		if wc.collides(rect) {
			continue
		}

		pos.X -= centerX
		pos.Y -= centerY
		return pos
	}
}

// Generate generates an image given the options
func (wc *WordCloud) Generate(options *GeneratorOptions) (*image.NRGBA, error) {
	if options.Rand == nil {
		return nil, errors.New("options missing rand")
	}

	wc.w, wc.h = options.Width, options.Height
	wc.spiraller = NewSpiraller(2, 5, wc.w, wc.h)
	wc.color = color.RGBA{A: 255}
	wc.img = image.NewNRGBA(image.Rect(0, 0, wc.w, wc.h))

	fontfile, err := ioutil.ReadFile(options.FontPath)
	if err != nil {
		return nil, err
	}

	f, err := freetype.ParseFont(fontfile)
	if err != nil {
		return nil, err
	}

	wc.prevFontSize = -1
	wc.font = f
	numAngles := len(options.Angles)
	i := 0

	for {
		if i >= len(options.Words) {
			break
		}
		word := options.Words[i]

		if wc.prevFontSize == -1 || wc.prevFontSize != word.Size {
			wc.prevFontSize = word.Size
			wc.drawer = &font.Drawer{
				Face: truetype.NewFace(wc.font, &truetype.Options{
					Size:    word.Size,
					DPI:     72,
					Hinting: font.HintingFull,
				}),
			}
		}

		rand := options.Rand.Int()
		wc.color.R = uint8(rand & 0xff)
		wc.color.G = uint8(rand >> 3 & 0xff)
		wc.color.B = uint8(rand >> 6 & 0xff)

		d := wc.drawer.MeasureString(word.Text)
		w, h := d.Ceil(), wc.drawer.Face.Metrics().Height.Ceil()*2

		tmp := image.NewNRGBA(image.Rect(0, 0, w, h))
		wc.drawer.Src = image.NewUniform(wc.color)
		wc.drawer.Dst = tmp
		wc.drawer.Dot = freetype.Pt(0, h/2)
		wc.drawer.DrawString(word.Text)

		if i != 0 {
			angle := options.Angles[options.Rand.Intn(numAngles)]
			if angle != 0 {
				tmp = imaging.Rotate(tmp, angle, color.Transparent)
			}
		}

		tmp = Crop(tmp)
		w = tmp.Bounds().Max.X
		h = tmp.Bounds().Max.Y

		// TODO: Generating bounding boxes or something

		wc.spiraller.Reset()
		pt := wc.getNextPosition(w, h)
		if pt == nil {
			if options.MinSize != 0 && word.Size >= options.MinSize {
				// shrink the font size and retry
				word.Size *= 0.75
				continue
			}
			// no other places for the current word, skip it
			i++
			continue
		}

		wc.img = imaging.Overlay(wc.img, tmp, *pt, 1.0)
		i++
	}

	return wc.img, nil
}
