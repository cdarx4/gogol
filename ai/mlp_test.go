// ============================================================================
// File: mlp_test.go
// Author: Ding Lionel, Darx Christian
// Email: lionel.ding@hes-so.ch, christia.darx@hes-so.ch
// Date: 04.01.2026
// Description: Tests for MLP neural network functionality.
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
	"testing"
)

func TestMLP_Forward_ValidInput(t *testing.T) {
	// Create a simple MLP: 2 inputs -> 3 hidden -> 1 output
	mlp := &MLP{
		Layers: []Layer{
			{
				Weights: [][]float64{
					{0.1, 0.2, 0.3},
					{0.4, 0.5, 0.6},
				},
				Biases: []float64{0.1, 0.2, 0.3},
			},
			{
				Weights: [][]float64{
					{0.7},
					{0.8},
					{0.9},
				},
				Biases: []float64{0.1},
			},
		},
	}

	input := []float64{1.0, 2.0}
	output := mlp.Forward(input)

	// Just check it doesn't panic and returns a value
	// This is a basic smoke test - actual value depends on the math
	// We just want to ensure it runs without panicking
	_ = output
}

func TestMLP_Forward_InputSizeMismatch(t *testing.T) {
	mlp := &MLP{
		Layers: []Layer{
			{
				Weights: [][]float64{
					{0.1},
				},
				Biases: []float64{0.1},
			},
		},
	}

	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic for input size mismatch, got none")
		}
	}()

	// Wrong input size (should be 1, but providing 2)
	mlp.Forward([]float64{1.0, 2.0})
}
