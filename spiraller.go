package wordcloud

import (
	"image"
	"math"
)

type Spiraller struct {
	Width, Height          float64
	centerX, centerY       int
	degressIncr            float64
	n                      float64
	a                      float64
	b                      float64
	maxN, maxE, maxS, maxW bool
	init                   bool
}

func NewSpiraller(a, dist float64, width, height int) *Spiraller {
	return &Spiraller{
		Width:       float64(width),
		Height:      float64(height),
		centerX:     width / 2,
		centerY:     height / 2,
		degressIncr: 0.1 * math.Pi / 180,
		b:           dist,
		a:           a,
	}
}

type Position struct {
	X int
	Y int
}

func (s *Spiraller) Next() *image.Point {
	if !s.init {
		s.init = true
		return &image.Point{X: s.centerX, Y: s.centerY}
	}

	if s.maxN && s.maxE && s.maxS && s.maxW {
		return nil
	}
	r := s.a + s.b*s.n
	x := r * math.Cos(s.n)
	y := r * math.Sin(s.n)
	if x >= s.Width {
		s.maxE = true
	}
	if x < 0 {
		s.maxW = true
	}
	if y >= s.Height {
		s.maxS = true
	}
	if y < 0 {
		s.maxN = true
	}
	s.n += s.degressIncr
	return &image.Point{
		X: int(x) + s.centerX,
		Y: int(y) + s.centerY,
	}
}

func (s *Spiraller) Reset() {
	s.maxN, s.maxE, s.maxS, s.maxW = false, false, false, false
	s.init = false
	s.n = 0.0
}
