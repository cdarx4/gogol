// ============================================================================
// File: encode_test.go
// Author: Ding Lionel, Darx Christian
// Email: lionel.ding@hes-so.ch, christia.darx@hes-so.ch
// Date: 04.01.2026
// Description: Tests for board encoding functionality.
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

	"heia2526/gogol/game"
)

func TestEncodeBoard_EmptyBoard(t *testing.T) {
	b := game.NewBoard(9)
	encoded := EncodeBoard(b, game.PlayerBlack)

	if len(encoded) != 81 {
		t.Fatalf("expected 81 elements, got %d", len(encoded))
	}

	for i, val := range encoded {
		if val != 0.0 {
			t.Errorf("expected all zeros for empty board, got %f at index %d", val, i)
		}
	}
}

func TestEncodeBoard_WithStones(t *testing.T) {
	b := game.NewBoard(9)

	// Place black stone at (4, 4)
	b.PlaceStone(4, 4)

	// Place white stone at (5, 5)
	b.PlaceStone(5, 5)

	// Encode from Black's perspective
	encoded := EncodeBoard(b, game.PlayerBlack)

	if len(encoded) != 81 {
		t.Fatalf("expected 81 elements, got %d", len(encoded))
	}

	// Check black stone at (4, 4) - index = 4*9 + 4 = 40
	idx := 4*9 + 4
	if encoded[idx] != 1.0 {
		t.Errorf("expected 1.0 at index %d (black stone), got %f", idx, encoded[idx])
	}

	// Check white stone at (5, 5) - index = 5*9 + 5 = 50
	idx = 5*9 + 5
	if encoded[idx] != -1.0 {
		t.Errorf("expected -1.0 at index %d (white stone), got %f", idx, encoded[idx])
	}
}

func TestEncodeBoard_PerspectiveSwitching(t *testing.T) {
	b := game.NewBoard(9)
	b.PlaceStone(4, 4) // Black stone

	// Encode from Black's perspective
	encodedBlack := EncodeBoard(b, game.PlayerBlack)

	// Encode from White's perspective
	encodedWhite := EncodeBoard(b, game.PlayerWhite)

	idx := 4*9 + 4
	if encodedBlack[idx] != 1.0 {
		t.Errorf("from Black perspective: expected 1.0, got %f", encodedBlack[idx])
	}

	if encodedWhite[idx] != -1.0 {
		t.Errorf("from White perspective: expected -1.0, got %f", encodedWhite[idx])
	}
}

func TestEncodeBoard_MultipleStones(t *testing.T) {
	b := game.NewBoard(9)

	// Place several stones
	b.PlaceStone(0, 0) // Black
	b.PlaceStone(1, 0) // White
	b.PlaceStone(2, 0) // Black
	b.PlaceStone(3, 0) // White

	encoded := EncodeBoard(b, game.PlayerBlack)

	// Check positions - encoding uses y*size + x
	// (0,0) -> 0*9+0 = 0, (1,0) -> 0*9+1 = 1, (2,0) -> 0*9+2 = 2, (3,0) -> 0*9+3 = 3
	expectations := []struct {
		x, y int
		want float64
	}{
		{0, 0, 1.0},  // Black at (0,0)
		{1, 0, -1.0}, // White at (1,0)
		{2, 0, 1.0},  // Black at (2,0)
		{3, 0, -1.0}, // White at (3,0)
	}

	for _, exp := range expectations {
		idx := exp.y*9 + exp.x
		if encoded[idx] != exp.want {
			t.Errorf("at (%d,%d) index %d: expected %f, got %f", exp.x, exp.y, idx, exp.want, encoded[idx])
		}
	}
}
