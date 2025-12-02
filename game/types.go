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

// Represents the different players
type Player int

// Represents a stone
type Stone struct {
	X, Y    int
	Player  Player
	GroupId int
}

// For the stone groups
// TODO add the liberties
type Group struct {
	ID        int
	Player    Player
	Stones    []*Stone
	Liberties int
}

// Renderer interface for the game
type Renderer interface {
	Draw(screen *ebiten.Image, game *Game)
	Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int)
	GetGridPosition(x, y int) (row, col int, onBoard bool)
}

// Represents the struct for the game itself
type Game struct {
	State    GameStates
	Renderer Renderer
	Board    *Board
}

// Different states of the game
// TODO add a configuration state later
const (
	GameStateIntro GameStates = "intro"
	GameStateGame  GameStates = "game"
	GameStateEnd   GameStates = "end"
)

// Different players White/Black
const (
	PlayerBlack Player = iota
	PlayerWhite
)

// Size of the board 9x9
const BoardSize = 9

func (p Player) String() string {
	if p == PlayerBlack {
		return "Black"
	} else if p == PlayerWhite {
		return "White"
	}
	return "Unknown"
}
