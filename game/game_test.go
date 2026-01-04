// ============================================================================
// File: game_test.go
// Author: Ding Lionel, Darx Christian
// Email: lionel.ding@hes-so.ch, christia.darx@hes-so.ch
// Date: 04.01.2026
// Description: Tests for game logic functions.
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

package game

import (
	"testing"
)

func TestGame_isBotTurn_PvEWhite(t *testing.T) {
	game := &Game{}
	game.Init()
	game.Mode = GameModePvE
	game.Board.CurrentPlayer = PlayerWhite

	if !game.isBotTurn() {
		t.Error("isBotTurn should return true for PvE mode with White's turn")
	}
}

func TestGame_isBotTurn_PvEBlack(t *testing.T) {
	game := &Game{}
	game.Init()
	game.Mode = GameModePvE
	game.Board.CurrentPlayer = PlayerBlack

	if game.isBotTurn() {
		t.Error("isBotTurn should return false for PvE mode with Black's turn")
	}
}

func TestGame_isBotTurn_PvP(t *testing.T) {
	game := &Game{}
	game.Init()
	game.Mode = GameModePvP
	game.Board.CurrentPlayer = PlayerWhite

	if game.isBotTurn() {
		t.Error("isBotTurn should return false for PvP mode")
	}
}

func TestGame_isBotTurn_PvPBlack(t *testing.T) {
	game := &Game{}
	game.Init()
	game.Mode = GameModePvP
	game.Board.CurrentPlayer = PlayerBlack

	if game.isBotTurn() {
		t.Error("isBotTurn should return false for PvP mode with Black's turn")
	}
}
