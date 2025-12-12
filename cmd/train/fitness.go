// ============================================================================
// File: fitness.go
// Author: Ding Lionel, Darx Christian
// Email: lionel.ding@hes-so.ch, christia.darx@hes-so.ch
// Date: 28.11.2025
// Description: Fitness evaluation for genetic algorithm.
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
	"sync"

	"heia2526/gogol/ai"
	"heia2526/gogol/game"
)

// EvaluateFitness evaluates the fitness of a model by playing against opponents
// Fitness: +1.0 per win, -1.0 per loss, +0.01 * score_difference
// This is the sequential version (kept for compatibility)
func EvaluateFitness(model *ai.MLP, opponents []*ai.MLP, gamesPerOpponent int) float64 {
	return EvaluateFitnessParallel(model, opponents, gamesPerOpponent)
}

// EvaluateFitnessParallel evaluates fitness with parallelized game simulations
// Fitness: +1.0 per win, -1.0 per loss, +0.01 * score_difference
func EvaluateFitnessParallel(model *ai.MLP, opponents []*ai.MLP, gamesPerOpponent int) float64 {
	if len(opponents) == 0 {
		return 0.0
	}

	// Collect all game tasks
	type gameTask struct {
		opponent    *ai.MLP
		playerColor game.Player
	}

	tasks := make([]gameTask, 0, len(opponents)*gamesPerOpponent*2)
	for _, opponent := range opponents {
		for i := 0; i < gamesPerOpponent; i++ {
			// Play two games: once as Black, once as White
			tasks = append(tasks, gameTask{opponent: opponent, playerColor: game.PlayerBlack})
			tasks = append(tasks, gameTask{opponent: opponent, playerColor: game.PlayerWhite})
		}
	}

	// Parallelize game simulations
	var wg sync.WaitGroup
	fitnessChan := make(chan float64, len(tasks))
	
	// Use a worker pool to limit concurrent games (avoid overwhelming CPU)
	numWorkers := len(tasks)
	if numWorkers > 20 {
		numWorkers = 20 // Cap at 20 concurrent games per individual
	}

	taskChan := make(chan gameTask, len(tasks))
	
	// Start workers
	for w := 0; w < numWorkers; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for task := range taskChan {
				var result GameResult
				if task.playerColor == game.PlayerBlack {
					result = SimulateGame(model, task.opponent)
				} else {
					result = SimulateGame(task.opponent, model)
				}

				// Calculate fitness contribution
				fitness := 0.0
				if result.IsDraw {
					fitness = 0.0
				} else if result.Winner != nil {
					if *result.Winner == task.playerColor {
						fitness = 1.0 // Win
					} else {
						fitness = -1.0 // Loss
					}
				}

				// Add score difference bonus (always from model's perspective)
				if task.playerColor == game.PlayerBlack {
					fitness += 0.01 * float64(result.ScoreDiff)
				} else {
					fitness += 0.01 * float64(-result.ScoreDiff) // Flip for White
				}

				fitnessChan <- fitness
			}
		}()
	}

	// Send tasks
	for _, task := range tasks {
		taskChan <- task
	}
	close(taskChan)

	// Wait for completion and collect results
	go func() {
		wg.Wait()
		close(fitnessChan)
	}()

	totalFitness := 0.0
	totalGames := 0
	for fitness := range fitnessChan {
		totalFitness += fitness
		totalGames++
	}

	if totalGames == 0 {
		return 0.0
	}

	return totalFitness / float64(totalGames)
}

