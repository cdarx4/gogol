// ============================================================================
// File: model_io.go
// Author: Ding Lionel, Darx Christian
// Email: lionel.ding@hes-so.ch, christia.darx@hes-so.ch
// Date: 28.11.2025
// Description: Model persistence (save/load) for MLP.
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
	"encoding/json"
	"fmt"
	"os"
)

const ModelVersion = "1.0"

// ModelData represents the serialized form of an MLP
type ModelData struct {
	Version    string      `json:"version"`
	Architecture []int     `json:"architecture"`
	Layers     []LayerData `json:"layers"`
}

// LayerData represents a serialized layer
type LayerData struct {
	Weights [][]float64 `json:"weights"`
	Biases  []float64   `json:"biases"`
}

// SaveModel saves an MLP to a JSON file
func SaveModel(m *MLP, path string) error {
	arch := m.GetArchitecture()
	layers := make([]LayerData, len(m.Layers))

	for i, layer := range m.Layers {
		layers[i] = LayerData(layer)
	}

	data := ModelData{
		Version:     ModelVersion,
		Architecture: arch,
		Layers:      layers,
	}

	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal model: %w", err)
	}

	if err := os.WriteFile(path, jsonData, 0644); err != nil {
		return fmt.Errorf("failed to write model file: %w", err)
	}

	return nil
}

// LoadModel loads an MLP from a JSON file
func LoadModel(path string) (*MLP, error) {
	jsonData, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read model file: %w", err)
	}

	var data ModelData
	if err := json.Unmarshal(jsonData, &data); err != nil {
		return nil, fmt.Errorf("failed to unmarshal model: %w", err)
	}

	if len(data.Architecture) < 2 {
		return nil, fmt.Errorf("invalid architecture: must have at least 2 layers")
	}

	if len(data.Layers) != len(data.Architecture)-1 {
		return nil, fmt.Errorf("layer count mismatch: expected %d, got %d", len(data.Architecture)-1, len(data.Layers))
	}

	mlp := &MLP{
		Layers: make([]Layer, len(data.Layers)),
	}

	for i, layerData := range data.Layers {
		mlp.Layers[i] = Layer(layerData)
	}

	return mlp, nil
}

