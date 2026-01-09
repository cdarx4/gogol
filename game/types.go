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

// GameStates represents the different states of the game.
type GameStates string

// GameMode represents the mode of the game (PvP or PvE)
type GameMode int

const (
	// GameModePvP is player versus player mode.
	GameModePvP GameMode = iota
	// GameModePvE is player versus environment (AI) mode.
	GameModePvE
)

// Player represents the different players in the game.
type Player int

const (
	// PlayerBlack represents the black player.
	PlayerBlack Player = iota
	// PlayerWhite represents the white player.
	PlayerWhite
)

// StartingPlayer is the player who moves first.
const StartingPlayer = PlayerBlack

// BoardSize is the default size of the game board.
const BoardSize = 9

// Stone represents a stone on the board.
type Stone struct {
	X, Y   int
	Player Player
	Group  *Group
}

// Group represents a connected group of stones.
type Group struct {
	ID        int
	Player    Player
	Stones    []*Stone
	Liberties map[[2]int]struct{}
}

// Board represents the game board.
type Board struct {
	Size          int
	Grid          [][]*Stone
	Groups        []*Group
	nextGroupID   int
	CurrentPlayer Player
	PassCount     int
	History       []string
}

// Renderer is the interface for rendering the game.
type Renderer interface {
	Draw(screen *ebiten.Image, game *Game)
	Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int)
	GetGridPosition(x, y int) (gridX, gridY int, onBoard bool)
}

// AIPlayer is the interface for AI players.
type AIPlayer interface {
	NextMove(b *Board, p Player) (x, y int, err error)
}

// Game represents the game state and logic.
type Game struct {
	State         GameStates
	Renderer      Renderer
	Board         *Board
	Mode          GameMode
	IsBotThinking bool
	BotMoveChan   chan BotMoveResult
	Bot           AIPlayer
}

// BotMoveResult represents the result of a bot's move calculation.
type BotMoveResult struct {
	X, Y int
	Pass bool
	Err  error
}

const (
	// GameStateIntro is the initial state of the game.
	GameStateIntro GameStates = "intro"
	// GameStateGame is the state when the game is in progress.
	GameStateGame GameStates = "game"
	// GameStateEnd is the state when the game has ended.
	GameStateEnd GameStates = "end"
)

// Direction represents a direction vector for adjacent positions
type Direction struct {
	DeltaX, DeltaY int
}

// Defining all the directions possible to form a group
var AdjacentDirections = []Direction{
	{-1, 0}, {1, 0}, {0, -1}, {0, 1},
}

// Opponent returns the opponent of the current player.
func (p Player) Opponent() Player {
	if p == PlayerBlack {
		return PlayerWhite
	}
	return PlayerBlack
}
