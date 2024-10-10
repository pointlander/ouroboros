// Copyright 2024 The Ouroboros Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"fmt"
	"math"
	"math/rand"
)

func main() {
	rng := rand.New(rand.NewSource(1))
	inputs := NewMatrix(Samples, Samples)
	for i := 0; i < Samples*Samples; i++ {
		inputs.Data = append(inputs.Data, complex(rng.Float64(), 0))
	}
	for i := 0; i < 33; i++ {
		outputs := Process(rng, inputs)
		min, max := math.MaxFloat64, -math.MaxFloat64
		for _, value := range outputs.Data {
			v := real(value)
			if v < min {
				min = v
			}
			if v > max {
				max = v
			}
		}
		entropy := 0.0
		for j := 0; j < outputs.Rows; j++ {
			for k := 0; k < outputs.Cols; k++ {
				value := real(outputs.Data[j*outputs.Cols+k])
				if value == 0 {
					continue
				}
				entropy += value * math.Log2(value)
			}
		}
		entropy = -entropy
		entropy /= float64(outputs.Rows)
		fmt.Println(i, entropy, min, max)
		for j := 0; j < Samples*Samples/2; j++ {
			outputs.Data[rng.Intn(Samples*Samples)] = complex(0, 0)
		}
		inputs = outputs
	}
}
