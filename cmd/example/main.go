package main

import (
	"fmt"
	"image/png"
	"math/rand"
	"os"
	"sort"
	"time"
	"wordcloud"
)

func main() {
	source := rand.NewSource(time.Now().UnixNano())
	rand := rand.New(source)

	minSize := 15
	baseFontSize := 100
	var words []*wordcloud.Word
	for i := 0; i < 100; i++ {
		words = append(words, &wordcloud.Word{
			Text: fmt.Sprintf("Test %d", i),
			Size: float64(rand.Intn(baseFontSize-minSize) + minSize),
		})
	}

	sort.Slice(words, func(i, j int) bool {
		return words[i].Size > words[j].Size
	})

	f, err := os.Create("test.png")
	if err != nil {
		panic(err)
	}
	defer f.Close()

	generator := wordcloud.WordCloud{}
	img, err := generator.Generate(&wordcloud.GeneratorOptions{
		Width:    400,
		Height:   400,
		MinSize:  15,
		FontPath: "/usr/share/fonts/TTF/UbuntuMono-R.ttf",
		Words:    words,
		Rand:     rand,
		Angles: []float64{
			0,
			45,
			-45,
			90,
		},
	})
	if err != nil {
		panic(err)
	}

	// Trim off the blank edges.
	img = wordcloud.Crop(img)
	png.Encode(f, img)
}
