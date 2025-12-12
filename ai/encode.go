// ============================================================================
// File: encode.go
// Author: Ding Lionel, Darx Christian
// Email: lionel.ding@hes-so.ch, christia.darx@hes-so.ch
// Date: 28.11.2025
// Description: Board encoding for MLP input.
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

import "heia2526/gogol/game"

// EncodeBoard encodes the board state as a vector from the player's perspective
// Returns an 81-element vector (9x9 board) where:
//   +1.0 = current player's stone
//   -1.0 = opponent's stone
//   0.0  = empty
func EncodeBoard(b *game.Board, p game.Player) []float64 {
	size := b.Size
	encoded := make([]float64, size*size)

	opponent := p.Opponent()

	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			idx := y*size + x
			stone := b.Grid[x][y]

			if stone == nil {
				encoded[idx] = 0.0
			} else if stone.Player == p {
				encoded[idx] = 1.0
			} else if stone.Player == opponent {
				encoded[idx] = -1.0
			} else {
				encoded[idx] = 0.0
			}
		}
	}

	return encoded
}
