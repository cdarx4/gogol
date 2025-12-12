// ============================================================================
// File: metrics.go
// Author: Ding Lionel, Darx Christian
// Email: lionel.ding@hes-so.ch, christia.darx@hes-so.ch
// Date: 28.11.2025
// Description: Training metrics collection and export.
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

package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"time"
)

// GenerationMetrics represents metrics for a single generation
type GenerationMetrics struct {
	Generation    int       `json:"generation"`
	Timestamp     time.Time `json:"timestamp"`
	BestFitness   float64   `json:"best_fitness"`
	AverageFitness float64  `json:"average_fitness"`
	WorstFitness  float64   `json:"worst_fitness"`
	TopNFitness   []float64 `json:"top_n_fitness"` // Top 5 individuals
}

// TrainingHistory contains all generation metrics
type TrainingHistory struct {
	Metrics []GenerationMetrics `json:"metrics"`
}

// RecordGeneration records metrics for the current generation
func RecordGeneration(pop Population, gen int) GenerationMetrics {
	if len(pop) == 0 {
		return GenerationMetrics{
			Generation:    gen,
			Timestamp:     time.Now(),
			BestFitness:   0.0,
			AverageFitness: 0.0,
			WorstFitness:  0.0,
			TopNFitness:   []float64{},
		}
	}

	// Calculate statistics
	fitnesses := make([]float64, len(pop))
	for i := range pop {
		fitnesses[i] = pop[i].Fitness
	}

	// Sort for statistics
	sorted := make([]float64, len(fitnesses))
	copy(sorted, fitnesses)
	sort.Float64s(sorted)

	best := sorted[len(sorted)-1]
	worst := sorted[0]

	// Calculate average
	sum := 0.0
	for _, f := range fitnesses {
		sum += f
	}
	average := sum / float64(len(fitnesses))

	// Get top 5
	topN := 5
	if len(sorted) < topN {
		topN = len(sorted)
	}
	topNFitness := make([]float64, topN)
	for i := 0; i < topN; i++ {
		topNFitness[i] = sorted[len(sorted)-1-i] // Top N (highest)
	}

	return GenerationMetrics{
		Generation:     gen,
		Timestamp:      time.Now(),
		BestFitness:    best,
		AverageFitness: average,
		WorstFitness:   worst,
		TopNFitness:    topNFitness,
	}
}

// SaveHistory saves training history to JSON file
func SaveHistory(history TrainingHistory, path string) error {
	jsonData, err := json.MarshalIndent(history, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal history: %w", err)
	}

	if err := os.WriteFile(path, jsonData, 0644); err != nil {
		return fmt.Errorf("failed to write history file: %w", err)
	}

	return nil
}

// ExportCSV exports training history to CSV file (Excel-compatible)
func ExportCSV(history TrainingHistory, path string) error {
	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("failed to create CSV file: %w", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write header with Excel-friendly column names
	header := []string{
		"Generation",
		"Timestamp",
		"Best Fitness",
		"Average Fitness",
		"Worst Fitness",
		"Top 1 Fitness",
		"Top 2 Fitness",
		"Top 3 Fitness",
		"Top 4 Fitness",
		"Top 5 Fitness",
		"Fitness Range", // Best - Worst
		"Improvement",   // Best fitness improvement from previous generation
	}
	if err := writer.Write(header); err != nil {
		return fmt.Errorf("failed to write CSV header: %w", err)
	}

	// Track previous best for improvement calculation
	previousBest := 0.0
	firstGen := true

	// Write data rows
	for _, m := range history.Metrics {
		// Calculate fitness range
		fitnessRange := m.BestFitness - m.WorstFitness

		// Calculate improvement from previous generation
		improvement := 0.0
		if !firstGen {
			improvement = m.BestFitness - previousBest
		}
		firstGen = false
		previousBest = m.BestFitness

		row := []string{
			fmt.Sprintf("%d", m.Generation),
			m.Timestamp.Format("2006-01-02 15:04:05"), // Excel-friendly date format
			fmt.Sprintf("%.6f", m.BestFitness),
			fmt.Sprintf("%.6f", m.AverageFitness),
			fmt.Sprintf("%.6f", m.WorstFitness),
		}

		// Add top N fitness values
		for i := 0; i < 5; i++ {
			if i < len(m.TopNFitness) {
				row = append(row, fmt.Sprintf("%.6f", m.TopNFitness[i]))
			} else {
				row = append(row, "")
			}
		}

		// Add calculated columns
		row = append(row, fmt.Sprintf("%.6f", fitnessRange))
		row = append(row, fmt.Sprintf("%.6f", improvement))

		if err := writer.Write(row); err != nil {
			return fmt.Errorf("failed to write CSV row: %w", err)
		}
	}

	return nil
}

