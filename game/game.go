// ============================================================================
// File: game.go
// Author: Ding Lionel, Darx Christian
// Email: lionel.ding@hes-so.ch, christia.darx@hes-so.ch
// Date: 28.11.2025
// Description: Go file managing the game.
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
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

// Initialize the game
func (g *Game) Init() {
	g.State = GameStateIntro
	g.Board = NewBoard(BoardSize)
}

// To pass to the next state when the user clicks or presses space
func (g *Game) Update() error {
	if g.State == GameStateIntro {
		if inpututil.IsKeyJustPressed(ebiten.KeySpace) || inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
			g.State = GameStateGame
		}
	} else if g.State == GameStateGame {
		if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
			x, y := ebiten.CursorPosition()
			if g.Renderer != nil {
				row, col, onBoard := g.Renderer.GetGridPosition(x, y)
				if onBoard {
					g.Board.PlaceStone(row, col)
				}
			}
		}
	}
	return nil
}

// Draw the game
func (g *Game) Draw(screen *ebiten.Image) {
	if g.Renderer != nil {
		g.Renderer.Draw(screen, g)
	}
}

// Layout the game
func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	if g.Renderer != nil {
		return g.Renderer.Layout(outsideWidth, outsideHeight)
	}
	return 600, 600
}
