// ============================================================================
// File: selfplay.go
// Author: Ding Lionel, Darx Christian
// Email: lionel.ding@hes-so.ch, christia.darx@hes-so.ch
// Date: 28.11.2025
// Description: Self-play simulation for fitness evaluation.
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
	"heia2526/gogol/ai"
	"heia2526/gogol/game"
)

// GameResult represents the result of a game
type GameResult struct {
	Winner     *game.Player
	IsDraw     bool
	ScoreDiff  int // Black score - White score
	BlackScore int
	WhiteScore int
}

// SimulateGame simulates a game between two MLP players
// Returns the game result
func SimulateGame(p1, p2 *ai.MLP) GameResult {
	board := game.NewBoard(game.BoardSize)
	player1 := ai.NewMLPPlayer(p1)
	player2 := ai.NewMLPPlayer(p2)

	maxMoves := 2 * game.BoardSize * game.BoardSize
	moveCount := 0
	consecutivePasses := 0

	for moveCount < maxMoves {
		var x, y int
		var err error

		if board.CurrentPlayer == game.PlayerBlack {
			x, y, err = player1.NextMove(board, game.PlayerBlack)
		} else {
			x, y, err = player2.NextMove(board, game.PlayerWhite)
		}

		if err != nil {
			// Error getting move, pass
			board.Pass()
			consecutivePasses++
		} else if x == -1 && y == -1 {
			// Pass move
			board.Pass()
			consecutivePasses++
		} else {
			// Regular move
			if board.PlaceStone(x, y) {
				consecutivePasses = 0
			} else {
				// Illegal move, pass instead
				board.Pass()
				consecutivePasses++
			}
		}

		moveCount++

		// End game on two consecutive passes
		if consecutivePasses >= 2 {
			break
		}
	}

	// Calculate scores (Chinese rules: stone count)
	blackScore, whiteScore := board.GetStoneCount()
	scoreDiff := blackScore - whiteScore

	winner, isDraw := board.GetWinner()

	return GameResult{
		Winner:     winner,
		IsDraw:     isDraw,
		ScoreDiff:  scoreDiff,
		BlackScore: blackScore,
		WhiteScore: whiteScore,
	}
}
