// Copyright 2024 The Ouroboros Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"flag"
	"fmt"
	"image"
	"math"
	"math/rand"
	"sort"

	htgotts "github.com/hegedustibor/htgo-tts"
	handlers "github.com/hegedustibor/htgo-tts/handlers"
	voices "github.com/hegedustibor/htgo-tts/voices"
	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/vg"
)

// Neuron is a neuron
type Neuron struct {
	Input chan []float64
	Name  string
}

var (
	// FlagPlot plot the results as a histogram
	FlagPlot = flag.Bool("plot", false, "plot the results as a histogram")
	// FlagMaze maze mode
	FlagMaze = flag.Bool("maze", false, "maze mode")
)

// Light lights the neuron
func (n Neuron) Light(seed int64, say chan string) {
	rng := rand.New(rand.NewSource(seed))
	inputs := NewMatrix(Samples, Samples)
	for i := 0; i < Samples*Samples; i++ {
		inputs.Data = append(inputs.Data, rng.Float64())
	}
	i := 0
	for {
		outputs := Process(rng, inputs)
		samples := make(plotter.Values, 0, 8)
		entropy := 0.0
		for j := 0; j < outputs.Rows; j++ {
			rowEntropy := 0.0
			for k := 0; k < outputs.Cols; k++ {
				value := outputs.Data[j*outputs.Cols+k]
				if value == 0 {
					continue
				}
				entropy += value * math.Log2(value)
				rowEntropy += value * math.Log2(value)
			}
			samples = append(samples, -rowEntropy)
		}

		select {
		case sense := <-n.Input:
			for j := range outputs.Data {
				if rng.Intn(2) == 0 {
					outputs.Data[j] = sense[rng.Intn(len(sense))]
				}
			}

		default:
		}

		sort.Slice(samples, func(i, j int) bool {
			return samples[i] < samples[j]
		})
		min, max := samples[0], samples[len(samples)-1]
		window := (max - min) / 100.0
	outer:
		for i, start := range samples {
			for j := i + 8; j < len(samples); j++ {
				end := samples[j]
				if (end - start) < window {
					say <- n.Name
					break outer
				}
			}
		}

		if *FlagPlot {
			p := plot.New()
			p.Title.Text = "entropy"

			histogram, err := plotter.NewHist(samples, 100)
			if err != nil {
				panic(err)
			}
			p.Add(histogram)

			err = p.Save(8*vg.Inch, 8*vg.Inch, fmt.Sprintf("output/%s_%d_entropy.png", n.Name, i))
			if err != nil {
				panic(err)
			}
		}
		entropy = -entropy
		entropy /= float64(outputs.Rows)
		inputs = outputs
		i++
	}
}

// Frame is a video frame
type Frame struct {
	Frame *image.YCbCr
	Thumb image.Image
	Gray  *image.Gray
}

// Maze is maze mode
func Maze() {

}

func main() {
	flag.Parse()

	if *FlagMaze {
		Maze()
		return
	}

	speech := htgotts.Speech{Folder: "audio", Language: voices.English, Handler: &handlers.MPlayer{}}
	speech.Speak("Starting...")

	left := Neuron{
		Input: make(chan []float64, 8),
		Name:  "left",
	}
	right := Neuron{
		Input: make(chan []float64, 8),
		Name:  "right",
	}
	forward := Neuron{
		Input: make(chan []float64, 8),
		Name:  "forward",
	}
	camera := NewV4LCamera()
	say := make(chan string, 8)
	go func() {
		for s := range say {
			speech.Speak(s)
		}
	}()

	go left.Light(1, say)
	go right.Light(2, say)
	go forward.Light(3, say)
	go camera.Start("/dev/video0")

	for img := range camera.Images {
		width := img.Gray.Bounds().Max.X
		height := img.Gray.Bounds().Max.Y
		dat := make([]float64, width*height)
		for y := 0; y < height; y++ {
			for x := 0; x < width; x++ {
				dat[y*width+x] = float64(img.Gray.GrayAt(x, y).Y) / 255
			}
		}
		select {
		case left.Input <- dat:
		default:
		}
		select {
		case right.Input <- dat:
		default:
		}
		select {
		case forward.Input <- dat:
		default:
		}
	}
}
