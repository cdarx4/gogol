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

import (
	"math/rand"
)

// Layer represents a single layer in the MLP
type Layer struct {
	Weights [][]float64
	Biases  []float64
}

// MLP represents a multi-layer perceptron
type MLP struct {
	Layers []Layer
}

// NewMLP creates a new MLP with the given architecture
// sizes: [input_size, hidden1_size, hidden2_size, ..., output_size]
func NewMLP(sizes []int) *MLP {
	if len(sizes) < 2 {
		panic("MLP must have at least input and output layers")
	}

	mlp := &MLP{
		Layers: make([]Layer, len(sizes)-1),
	}

	// Initialize each layer
	for i := 0; i < len(sizes)-1; i++ {
		inputSize := sizes[i]
		outputSize := sizes[i+1]

		layer := Layer{
			Weights: make([][]float64, inputSize),
			Biases:  make([]float64, outputSize),
		}

		// Initialize weights with Normal(0, 0.1)
		for j := 0; j < inputSize; j++ {
			layer.Weights[j] = make([]float64, outputSize)
			for k := 0; k < outputSize; k++ {
				layer.Weights[j][k] = rand.NormFloat64() * 0.1
			}
		}

		// Initialize biases with Normal(0, 0.1)
		for j := 0; j < outputSize; j++ {
			layer.Biases[j] = rand.NormFloat64() * 0.1
		}

		mlp.Layers[i] = layer
	}

	return mlp
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

// Mutate applies random mutations to the network weights and biases
// sigma: standard deviation for mutation noise
func (m *MLP) Mutate(sigma float64) {
	for i := range m.Layers {
		// Mutate weights
		for j := range m.Layers[i].Weights {
			for k := range m.Layers[i].Weights[j] {
				m.Layers[i].Weights[j][k] += rand.NormFloat64() * sigma
			}
		}

		// Mutate biases
		for j := range m.Layers[i].Biases {
			m.Layers[i].Biases[j] += rand.NormFloat64() * sigma
		}
	}
}

// Clone creates a deep copy of the MLP
func (m *MLP) Clone() *MLP {
	cloned := &MLP{
		Layers: make([]Layer, len(m.Layers)),
	}

	for i, layer := range m.Layers {
		cloned.Layers[i] = Layer{
			Weights: make([][]float64, len(layer.Weights)),
			Biases:  make([]float64, len(layer.Biases)),
		}

		// Copy weights
		for j := range layer.Weights {
			cloned.Layers[i].Weights[j] = make([]float64, len(layer.Weights[j]))
			copy(cloned.Layers[i].Weights[j], layer.Weights[j])
		}

		// Copy biases
		copy(cloned.Layers[i].Biases, layer.Biases)
	}

	return cloned
}

// GetArchitecture returns the architecture of the MLP
func (m *MLP) GetArchitecture() []int {
	if len(m.Layers) == 0 {
		return []int{}
	}

	sizes := make([]int, len(m.Layers)+1)
	sizes[0] = len(m.Layers[0].Weights) // Input size
	for i := 0; i < len(m.Layers); i++ {
		sizes[i+1] = len(m.Layers[i].Biases) // Output size of layer i
	}
	return sizes
}

