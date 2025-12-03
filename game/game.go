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
	"fmt"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

// Initialize the game
func (g *Game) Init() {
	g.State = GameStateIntro
	g.Board = NewBoard(BoardSize)
	g.Mode = GameModePvP
	g.BotMoveChan = make(chan BotMoveResult)
}

// To pass to the next state when the user clicks or presses space
func (g *Game) Update() error {
	if g.State == GameStateIntro {
		if inpututil.IsKeyJustPressed(ebiten.KeyP) {
			g.Mode = GameModePvP
			g.State = GameStateGame
		} else if inpututil.IsKeyJustPressed(ebiten.KeyB) {
			g.Mode = GameModePvE
			g.State = GameStateGame
		} else if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
			g.Mode = GameModePvP
			g.State = GameStateGame
		}
	} else if g.State == GameStateGame {
		// Check for bot result
		select {
		case result := <-g.BotMoveChan:
			g.IsBotThinking = false
			if result.Err == nil {
				if g.Board.PlaceStone(result.X, result.Y) {
					g.PrintGame()
				}
			} else {
				fmt.Println("Bot error:", result.Err)
			}
		default:
			// No result yet
		}

		if g.IsBotThinking {
			return nil // Don't allow other inputs while thinking
		}

		if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
			x, y := ebiten.CursorPosition()
			if g.Renderer != nil {
				row, col, onBoard := g.Renderer.GetGridPosition(x, y)
				if onBoard {
					// In PvE, only allow player (Black) to move manually
					// In PvP, both can move manually (turn logic handled by Board)
					if g.Mode == GameModePvE && g.Board.currentPlayer != PlayerBlack {
						// It's bot's turn, ignore click
					} else {
						if g.Board.PlaceStone(row, col) {
							g.PrintGame()
						}
					}
				}

			}
		}

		// Bot turn (White) - Only in PvE mode
		if g.Mode == GameModePvE && g.Board.currentPlayer == PlayerWhite && !g.IsBotThinking {
			g.IsBotThinking = true
			go func() {
				x, y, err := GetNextMove(g.Board, PlayerWhite)
				g.BotMoveChan <- BotMoveResult{X: x, Y: y, Err: err}
			}()
		}
	}
	return nil
}

// Print the current game state
func (g *Game) PrintGame() {
	fmt.Println("Current Game State:")
	fmt.Println(g.Board.String())
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
