// ============================================================================
// File: model_io_test.go
// Author: Ding Lionel, Darx Christian
// Email: lionel.ding@hes-so.ch, christia.darx@hes-so.ch
// Date: 04.01.2026
// Description: Tests for model loading functionality.
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
	"os"
	"path/filepath"
	"testing"
)

func TestLoadModel_ValidModel(t *testing.T) {
	// Create a temporary valid model file
	tmpDir := t.TempDir()
	modelPath := filepath.Join(tmpDir, "test_model.json")

	// Create a minimal valid model
	modelData := ModelData{
		Version:      ModelVersion,
		Architecture: []int{81, 64, 32, 1},
		Layers: []LayerData{
			{
				Weights: make([][]float64, 81),
				Biases:  make([]float64, 64),
			},
			{
				Weights: make([][]float64, 64),
				Biases:  make([]float64, 32),
			},
			{
				Weights: make([][]float64, 32),
				Biases:  make([]float64, 1),
			},
		},
	}

	// Initialize weights
	for i := range modelData.Layers[0].Weights {
		modelData.Layers[0].Weights[i] = make([]float64, 64)
	}
	for i := range modelData.Layers[1].Weights {
		modelData.Layers[1].Weights[i] = make([]float64, 32)
	}
	for i := range modelData.Layers[2].Weights {
		modelData.Layers[2].Weights[i] = make([]float64, 1)
	}

	jsonData, err := json.MarshalIndent(modelData, "", "  ")
	if err != nil {
		t.Fatalf("failed to marshal test model: %v", err)
	}

	if err := os.WriteFile(modelPath, jsonData, 0644); err != nil {
		t.Fatalf("failed to write test model: %v", err)
	}

	// Test loading using os.DirFS
	fsys := os.DirFS(tmpDir)
	model, err := LoadModel(fsys, "test_model.json")
	if err != nil {
		t.Fatalf("LoadModel failed: %v", err)
	}

	if model == nil {
		t.Fatal("LoadModel returned nil model")
	}

	if len(model.Layers) != 3 {
		t.Errorf("expected 3 layers, got %d", len(model.Layers))
	}
}

func TestLoadModel_InvalidPath(t *testing.T) {
	fsys := os.DirFS(".")
	_, err := LoadModel(fsys, "nonexistent/path/model.json")
	if err == nil {
		t.Fatal("expected error for nonexistent file, got nil")
	}
}

func TestLoadModel_InvalidJSON(t *testing.T) {
	tmpDir := t.TempDir()
	modelPath := filepath.Join(tmpDir, "invalid.json")

	if err := os.WriteFile(modelPath, []byte("not json"), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	fsys := os.DirFS(tmpDir)
	_, err := LoadModel(fsys, "invalid.json")
	if err == nil {
		t.Fatal("expected error for invalid JSON, got nil")
	}
}

func TestLoadModel_InvalidArchitecture(t *testing.T) {
	tmpDir := t.TempDir()
	modelPath := filepath.Join(tmpDir, "invalid_arch.json")

	// Model with only 1 layer (invalid - need at least 2)
	modelData := ModelData{
		Version:      ModelVersion,
		Architecture: []int{81},
		Layers:       []LayerData{},
	}

	jsonData, err := json.Marshal(modelData)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	if err := os.WriteFile(modelPath, jsonData, 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	fsys := os.DirFS(tmpDir)
	_, err = LoadModel(fsys, "invalid_arch.json")
	if err == nil {
		t.Fatal("expected error for invalid architecture, got nil")
	}
}

func TestLoadModel_LayerCountMismatch(t *testing.T) {
	tmpDir := t.TempDir()
	modelPath := filepath.Join(tmpDir, "mismatch.json")

	// Architecture says 3 layers but only 2 provided
	modelData := ModelData{
		Version:      ModelVersion,
		Architecture: []int{81, 64, 32, 1},
		Layers: []LayerData{
			{Weights: make([][]float64, 81), Biases: make([]float64, 64)},
			{Weights: make([][]float64, 64), Biases: make([]float64, 32)},
			// Missing third layer
		},
	}

	jsonData, err := json.Marshal(modelData)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	if err := os.WriteFile(modelPath, jsonData, 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	fsys := os.DirFS(tmpDir)
	_, err = LoadModel(fsys, "mismatch.json")
	if err == nil {
		t.Fatal("expected error for layer count mismatch, got nil")
	}
}
