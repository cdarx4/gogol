// ============================================================================
// File: main.go
// Author: Ding Lionel, Darx Christian
// Email: lionel.ding@hes-so.ch, christia.darx@hes-so.ch
// Date: 28.11.2025
// Description: Training entry point for neuro-evolution AI.
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
	"flag"
	"fmt"
	"log"
	"math/rand"
	"os"
	"runtime"
	"sync"
	"sync/atomic"
	"time"

	"heia2526/gogol/ai"
)

func main() {
	// CLI flags
	population := flag.Int("population", 50, "Population size")
	generations := flag.Int("generations", 100, "Number of generations")
	gamesPerModel := flag.Int("games-per-model", 10, "Games per model for fitness evaluation")
	mutationSigma := flag.Float64("mutation-sigma", 0.1, "Mutation standard deviation")
	eliteSize := flag.Int("elite-size", 5, "Number of elite individuals to preserve")
	output := flag.String("output", "models/champion.json", "Output path for champion model")
	metricsPath := flag.String("metrics", "models/training_history.json", "Output path for training metrics")
	csvPath := flag.String("csv", "models/training_history.csv", "CSV export path for metrics (Excel-compatible)")
	workers := flag.Int("workers", runtime.NumCPU(), "Number of parallel workers (default: CPU count)")
	flag.Parse()

	// Note: As of Go 1.20+, rand.Seed is deprecated.
	// The global random generator is automatically seeded, so no explicit seeding needed.

	log.Printf("Starting training with population=%d, generations=%d", *population, *generations)
	log.Printf("Games per model=%d, mutation sigma=%.3f, elite size=%d", *gamesPerModel, *mutationSigma, *eliteSize)
	log.Printf("Using %d parallel workers", *workers)

	// Initialize population
	pop := make(Population, *population)
	for i := range pop {
		// Architecture: 81 (input) -> 64 (hidden) -> 1 (output)
		pop[i] = Individual{
			Model:   ai.NewMLP([]int{81, 64, 1}),
			Fitness: 0.0,
		}
	}

	// Initialize training history
	history := TrainingHistory{
		Metrics: make([]GenerationMetrics, 0, *generations),
	}

	// Track overall training time
	trainingStart := time.Now()

	// Training loop
	for gen := 0; gen < *generations; gen++ {
		genStart := time.Now()
		log.Printf("Generation %d/%d", gen+1, *generations)

		// Evaluate fitness for all individuals in parallel with progress bar
		log.Printf("  Evaluating fitness...")
		evalStart := time.Now()
		evaluateFitnessParallel(pop, *gamesPerModel, *workers, evalStart)
		fmt.Fprintf(os.Stderr, "\n") // New line after progress bar
		evalDuration := time.Since(evalStart)
		log.Printf("  Fitness evaluation completed in %v", evalDuration)

		// Record metrics
		metrics := RecordGeneration(pop, gen)
		history.Metrics = append(history.Metrics, metrics)
		genDuration := time.Since(genStart)
		totalElapsed := time.Since(trainingStart)
		avgGenTime := totalElapsed / time.Duration(gen+1)
		estimatedRemaining := avgGenTime * time.Duration(*generations-gen-1)

		log.Printf("  Best fitness: %.3f, Average: %.3f, Worst: %.3f",
			metrics.BestFitness, metrics.AverageFitness, metrics.WorstFitness)
		log.Printf("  Generation time: %v | Total: %v | Est. remaining: %v",
			genDuration.Round(time.Second), totalElapsed.Round(time.Second), estimatedRemaining.Round(time.Second))

		// Save metrics and CSV after each generation
		if err := SaveHistory(history, *metricsPath); err != nil {
			log.Printf("  Warning: Failed to save metrics: %v", err)
		}
		// Always export CSV for Excel compatibility
		if err := ExportCSV(history, *csvPath); err != nil {
			log.Printf("  Warning: Failed to export CSV: %v", err)
		} else {
			log.Printf("  Metrics exported to %s", *csvPath)
		}

		// Save best model after each generation
		best := getBestIndividual(pop)
		bestPath := fmt.Sprintf("models/best_gen_%d.json", gen+1)
		if err := ai.SaveModel(best.Model, bestPath); err != nil {
			log.Printf("  Warning: Failed to save best model: %v", err)
		} else {
			log.Printf("  Best model saved to %s (fitness: %.3f)", bestPath, best.Fitness)
		}

		// Evolution: select, reproduce, mutate (except last generation)
		if gen < *generations-1 {
			log.Printf("  Evolving population...")
			pop = Evolve(pop, *eliteSize, *mutationSigma)
		}
	}

	// Save final champion
	best := getBestIndividual(pop)
	totalDuration := time.Since(trainingStart)
	log.Printf("Training complete! Best fitness: %.3f", best.Fitness)
	log.Printf("Total training time: %v", totalDuration.Round(time.Second))
	if err := ai.SaveModel(best.Model, *output); err != nil {
		log.Fatalf("Failed to save champion model: %v", err)
	}
	log.Printf("Champion saved to %s", *output)

	// Final metrics save
	if err := SaveHistory(history, *metricsPath); err != nil {
		log.Printf("Warning: Failed to save final metrics: %v", err)
	}
	// Always export final CSV
	if err := ExportCSV(history, *csvPath); err != nil {
		log.Printf("Warning: Failed to export final CSV: %v", err)
	} else {
		log.Printf("Training metrics exported to %s (Excel-compatible)", *csvPath)
	}
}

// selectOpponents selects random opponents from population (excluding self)
func selectOpponents(pop Population, excludeIdx int, count int) []*ai.MLP {
	opponents := make([]*ai.MLP, 0, count)
	indices := make(map[int]bool)
	indices[excludeIdx] = true

	for len(opponents) < count && len(opponents) < len(pop)-1 {
		idx := rand.Intn(len(pop))
		if !indices[idx] {
			indices[idx] = true
			opponents = append(opponents, pop[idx].Model)
		}
	}

	return opponents
}

// getBestIndividual returns the individual with highest fitness
func getBestIndividual(pop Population) Individual {
	best := pop[0]
	for i := 1; i < len(pop); i++ {
		if pop[i].Fitness > best.Fitness {
			best = pop[i]
		}
	}
	return best
}

// evaluateFitnessParallel evaluates fitness for all individuals in parallel using worker pool pattern
func evaluateFitnessParallel(pop Population, gamesPerModel, numWorkers int, startTime time.Time) {
	// Create job channel
	jobs := make(chan int, len(pop))
	results := make(chan struct {
		idx     int
		fitness float64
	}, len(pop))

	// Track completed count atomically for progress updates
	var completed int64

	// Start worker goroutines
	var wg sync.WaitGroup
	for w := 0; w < numWorkers; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for idx := range jobs {
				// Select random opponents from population
				opponents := selectOpponents(pop, idx, 5) // Play against 5 random opponents
				fitness := EvaluateFitnessParallel(pop[idx].Model, opponents, gamesPerModel)

				results <- struct {
					idx     int
					fitness float64
				}{idx: idx, fitness: fitness}

				// Update progress atomically
				completed := atomic.AddInt64(&completed, 1)
				progress := float64(completed) / float64(len(pop))
				elapsed := time.Since(startTime)
				estimatedTotal := time.Duration(float64(elapsed) / progress)
				remaining := estimatedTotal - elapsed

				// Thread-safe progress bar update
				printProgressBar(int(completed), len(pop), progress, elapsed, remaining)
			}
		}()
	}

	// Send jobs
	for i := range pop {
		jobs <- i
	}
	close(jobs)

	// Wait for all workers to finish
	go func() {
		wg.Wait()
		close(results)
	}()

	// Collect results
	for result := range results {
		pop[result.idx].Fitness = result.fitness
	}
}

// printProgressBar prints a progress bar to stderr (so it doesn't interfere with log output)
// Thread-safe when called from multiple goroutines (uses atomic operations)
func printProgressBar(current, total int, progress float64, elapsed, remaining time.Duration) {
	const barWidth = 40
	filled := int(progress * barWidth)
	bar := make([]byte, barWidth+2)
	bar[0] = '['
	for i := 1; i <= barWidth; i++ {
		if i <= filled {
			bar[i] = '='
		} else {
			bar[i] = ' '
		}
	}
	bar[barWidth+1] = ']'

	percent := progress * 100
	fmt.Fprintf(os.Stderr, "\r  %s %.1f%% [%d/%d] | Elapsed: %v | Remaining: ~%v",
		string(bar), percent, current, total, elapsed.Round(time.Second), remaining.Round(time.Second))
}
