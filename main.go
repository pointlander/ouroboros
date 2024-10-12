// Copyright 2024 The Ouroboros Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"flag"
	"fmt"
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

var (
	// FlagPlot plot the results as a histogram
	FlagPlot = flag.Bool("plot", false, "plot the results as a histogram")
)

func main() {
	flag.Parse()

	speech := htgotts.Speech{Folder: "audio", Language: voices.English, Handler: &handlers.Native{}}
	speech.Speak("Starting...")

	rng := rand.New(rand.NewSource(1))
	inputs := NewMatrix(Samples, Samples)
	for i := 0; i < Samples*Samples; i++ {
		inputs.Data = append(inputs.Data, rng.Float64())
	}
	for i := 0; i < 33; i++ {
		outputs := Process(rng, inputs)
		min, max := math.MaxFloat64, -math.MaxFloat64
		for _, value := range outputs.Data {
			v := value
			if v < min {
				min = v
			}
			if v > max {
				max = v
			}
		}
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

		{
			sort.Slice(samples, func(i, j int) bool {
				return samples[i] < samples[j]
			})
			min, max = samples[0], samples[len(samples)-1]
			window := (max - min) / 100.0
		outer:
			for i, start := range samples {
				for j := i + 8; j < len(samples); j++ {
					end := samples[j]
					if (end - start) < window {
						fmt.Println("fire", start, end)
						break outer
					}
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

			err = p.Save(8*vg.Inch, 8*vg.Inch, fmt.Sprintf("output/%d_entropy.png", i))
			if err != nil {
				panic(err)
			}
		}
		entropy = -entropy
		entropy /= float64(outputs.Rows)
		fmt.Println(i, entropy, min, max)
		if i < 10 {
			value := 0.0
			if i%2 == 1 {
				value = 1
			}
			for j := 0; j < Samples*Samples/2; j++ {
				outputs.Data[rng.Intn(Samples*Samples)] *= value
			}
		}
		inputs = outputs
	}
}
