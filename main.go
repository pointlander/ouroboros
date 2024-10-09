// Copyright 2024 The Ouroboros Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"fmt"
	"math"
	"math/rand"
)

// Values are the values
type Values struct {
	Values [Samples]float64
}

func main() {
	rng := rand.New(rand.NewSource(1))
	inputs := NewMatrix(Samples, Samples)
	for i := 0; i < Samples*Samples; i++ {
		inputs.Data = append(inputs.Data, complex(rng.NormFloat64(), 0))
	}
	for i := 0; i < 33; i++ {
		outputs := Process(rng, inputs)
		sum, min, max := 0.0, math.MaxFloat64, -math.MaxFloat64
		for _, value := range outputs.Data {
			v := real(value)
			sum += v
			if v < min {
				min = v
			}
			if v > max {
				max = v
			}
		}
		sum /= float64(len(outputs.Data))
		fmt.Println(i, sum, min, max)
		for j := 0; j < Samples*Samples/2; j++ {
			outputs.Data[rng.Intn(Samples*Samples)] = complex(0, 0)
		}
		inputs = outputs
	}
}
