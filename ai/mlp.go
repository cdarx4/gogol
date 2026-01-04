// ============================================================================
// File: mlp.go
// Author: Ding Lionel, Darx Christian
// Email: lionel.ding@hes-so.ch, christia.darx@hes-so.ch
// Date: 28.11.2025
// Description: Multi-layer perceptron implementation for neuro-evolution.
// Version: 1.0
//
// License: MIT
// Copyright 2025, School of Engineering and Architecture of Fribourg
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
// ============================================================================

package ai

// Layer represents a single layer in the MLP
type Layer struct {
	Weights [][]float64
	Biases  []float64
}

// MLP represents a multi-layer perceptron
type MLP struct {
	Layers []Layer
}

// Forward performs a forward pass through the network
// Returns the output value (single float64 for value network)
func (m *MLP) Forward(input []float64) float64 {
	if len(input) != len(m.Layers[0].Weights) {
		panic("input size mismatch")
	}

	// Start with input
	current := make([]float64, len(input))
	copy(current, input)

	// Pass through each layer
	for i, layer := range m.Layers {
		outputSize := len(layer.Biases)
		next := make([]float64, outputSize)

		// Compute weighted sum + bias
		for j := 0; j < outputSize; j++ {
			sum := layer.Biases[j]
			for k := 0; k < len(current); k++ {
				sum += current[k] * layer.Weights[k][j]
			}
			next[j] = sum
		}

		// Apply activation function
		// ReLU for hidden layers, linear for output
		if i < len(m.Layers)-1 {
			// Hidden layer: ReLU
			for j := 0; j < len(next); j++ {
				if next[j] < 0 {
					next[j] = 0
				}
			}
		}
		// Output layer: linear (no activation)

		current = next
	}

	// Return single output value
	return current[0]
}
