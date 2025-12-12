// ============================================================================
// File: types.go
// Author: Ding Lionel, Darx Christian
// Email: lionel.ding@hes-so.ch, christia.darx@hes-so.ch
// Date: 28.11.2025
// Description: Go file managing the types.
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
	"github.com/hajimehoshi/ebiten/v2"
)

// Represents the different states of the game
type GameStates string

// GameMode represents the mode of the game (PvP or PvE)
type GameMode int

const (
	GameModePvP GameMode = iota
	GameModePvE
)

// Represents the different players
type Player int

const (
	PlayerBlack Player = iota
	PlayerWhite
)

const StartingPlayer = PlayerBlack
const BoardSize = 9

// Represents a stone on the board
type Stone struct {
	X, Y   int
	Player Player
	Group  *Group
}

// Represents a connected group of stones
type Group struct {
	ID        int
	Player    Player
	Stones    []*Stone
	Liberties map[[2]int]struct{}
}

// Represents the board
type Board struct {
	Size          int
	Grid          [][]*Stone
	Groups        []*Group
	nextGroupId   int
	currentPlayer Player
}

// Renderer interface for the game
type Renderer interface {
	Draw(screen *ebiten.Image, game *Game)
	Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int)
	GetGridPosition(x, y int) (row, col int, onBoard bool)
}

// Represents the struct for the game itself
type Game struct {
	State         GameStates
	Renderer      Renderer
	Board         *Board
	Mode          GameMode
	IsBotThinking bool
	BotMoveChan   chan BotMoveResult
}

type BotMoveResult struct {
	X, Y int
	Err  error
}

const (
	GameStateIntro GameStates = "intro"
	GameStateGame  GameStates = "game"
	GameStateEnd   GameStates = "end"
)

var AdjacentDirections = []struct{ dx, dy int }{
	{-1, 0}, {1, 0}, {0, -1}, {0, 1},
}

func (p Player) Opponent() Player {
	if p == PlayerBlack {
		return PlayerWhite
	}
	return PlayerBlack
}
