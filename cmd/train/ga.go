// ============================================================================
// File: ga.go
// Author: Ding Lionel, Darx Christian
// Email: lionel.ding@hes-so.ch, christia.darx@hes-so.ch
// Date: 28.11.2025
// Description: Genetic algorithm core for neuro-evolution.
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
	"math/rand"
	"sort"

	"heia2526/gogol/ai"
)

// Individual represents a single individual in the population
type Individual struct {
	Model   *ai.MLP
	Fitness float64
}

// Population is a collection of individuals
type Population []Individual

// Evolve performs one generation of evolution
// Returns a new population
func Evolve(pop Population, eliteSize int, mutationSigma float64) Population {
	// Sort by fitness (descending)
	sort.Slice(pop, func(i, j int) bool {
		return pop[i].Fitness > pop[j].Fitness
	})

	newPop := make(Population, len(pop))

	// Elitism: top K individuals survive unchanged
	for i := 0; i < eliteSize && i < len(pop); i++ {
		newPop[i] = Individual{
			Model:   pop[i].Model.Clone(),
			Fitness: 0.0, // Will be re-evaluated
		}
	}

	// Fill rest of population by cloning and mutating elite
	for i := eliteSize; i < len(pop); i++ {
		// Select random elite individual
		parentIdx := rand.Intn(eliteSize)
		child := Individual{
			Model:   pop[parentIdx].Model.Clone(),
			Fitness: 0.0,
		}
		// Mutate
		child.Model.Mutate(mutationSigma)
		newPop[i] = child
	}

	return newPop
}

