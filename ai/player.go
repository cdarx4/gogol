// ============================================================================
// File: player.go
// Author: Ding Lionel, Darx Christian
// Email: lionel.ding@hes-so.ch, christia.darx@hes-so.ch
// Date: 28.11.2025
// Description: AI player interface and MLP-based player implementation.
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
	"fmt"
	"math"
	"math/rand"
	"sort"

	"heia2526/gogol/game"
)

// MLPPlayer is an AI player that uses an MLP to evaluate positions
// It implements the game.AIPlayer interface
type MLPPlayer struct {
	Model         *MLP
	Deterministic bool // If true, always plays best move (no randomness)
}

// NewMLPPlayer creates a new MLPPlayer with the given model.
// The player uses stochastic move selection (80% best move, 20% weighted random from top 5).
func NewMLPPlayer(model *MLP) *MLPPlayer {
	return &MLPPlayer{
		Model:         model,
		Deterministic: false,
	}
}

// NewDeterministicMLPPlayer creates a deterministic MLPPlayer that always plays the best move.
func NewDeterministicMLPPlayer(model *MLP) *MLPPlayer {
	return &MLPPlayer{
		Model:         model,
		Deterministic: true,
	}
}

// MoveScore represents a move with its evaluation score
type MoveScore struct {
	Move  [2]int
	Score float64
}

// NextMove selects the next move using the MLP to evaluate positions.
// It returns the coordinates (x, y) of the selected move, or (-1, -1) to indicate a pass.
func (p *MLPPlayer) NextMove(b *game.Board, player game.Player) (x, y int, err error) {
	if p.Model == nil {
		return -1, -1, fmt.Errorf("MLP model is nil")
	}

	return p.nextMoveSinglePly(b, player)
}

// nextMoveSinglePly evaluates all legal moves using a 1-ply lookahead.
// In deterministic mode, it always returns the best move.
// Otherwise, it uses stochastic selection: 80% best move, 20% weighted random from top 5.
func (p *MLPPlayer) nextMoveSinglePly(b *game.Board, player game.Player) (x, y int, err error) {
	// Get all legal moves
	legalMoves := b.GetLegalMoves(player)
	if len(legalMoves) == 0 {
		return -1, -1, fmt.Errorf("no legal moves available")
	}

	// Evaluate each move
	moveScores := make([]MoveScore, 0, len(legalMoves))
	for _, move := range legalMoves {
		x, y := move[0], move[1]

		// For pass move, use a default low score
		if x == -1 && y == -1 {
			moveScores = append(moveScores, MoveScore{
				Move:  move,
				Score: -1000.0, // Very low score for pass
			})
			continue
		}

		// Clone board and apply move
		cloned, success := b.CloneAndPlay(x, y, player)
		if !success {
			continue // Skip illegal moves (shouldn't happen, but be safe)
		}

		encoded := EncodeBoard(cloned, cloned.CurrentPlayer.Opponent())

		// Evaluate with MLP (value from the perspective of the player who just moved)
		score := p.Model.Forward(encoded)
		moveScores = append(moveScores, MoveScore{
			Move:  move,
			Score: score,
		})
	}

	if len(moveScores) == 0 {
		return -1, -1, fmt.Errorf("no valid moves after evaluation")
	}

	// Sort by score (descending)
	sort.Slice(moveScores, func(i, j int) bool {
		return moveScores[i].Score > moveScores[j].Score
	})

	// Selection strategy: deterministic or stochastic
	if p.Deterministic {
		// Deterministic mode: always return best move
		return moveScores[0].Move[0], moveScores[0].Move[1], nil
	}

	// Stochastic mode: 80% best move, 20% weighted random from top 5
	if rand.Float64() < 0.8 {
		// Return best move
		return moveScores[0].Move[0], moveScores[0].Move[1], nil
	}

	// Weighted random from top 5 (or fewer if less than 5 moves available)
	topN := 5
	if len(moveScores) < topN {
		topN = len(moveScores)
	}

	// Calculate weights (exponential decay: best move has highest weight)
	weights := make([]float64, topN)
	totalWeight := 0.0
	for i := 0; i < topN; i++ {
		// Convert score to weight (higher score = higher weight)
		// Use exponential: weight = exp(score / temperature)
		temperature := 1.0
		weight := math.Exp(moveScores[i].Score / temperature)
		weights[i] = weight
		totalWeight += weight
	}

	// Select based on weights
	r := rand.Float64() * totalWeight
	cumulative := 0.0
	for i := 0; i < topN; i++ {
		cumulative += weights[i]
		if r <= cumulative {
			return moveScores[i].Move[0], moveScores[i].Move[1], nil
		}
	}

	// Fallback to best move
	return moveScores[0].Move[0], moveScores[0].Move[1], nil
}
