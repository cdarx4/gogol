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

// Init initializes the game to its starting state.
// Preserves the bot if it was already set (for game restarts).
func (g *Game) Init() {
	g.State = GameStateIntro
	g.Board = NewBoard(BoardSize)
	g.Mode = GameModePvP
	g.BotMoveChan = make(chan BotMoveResult)
	// Note: g.Bot is intentionally not reset here to preserve it across game restarts
}

// Update updates the game state. It implements ebiten.Game.
func (g *Game) Update() error {
	if g.State == GameStateIntro {
		return g.handleIntroState()
	}

	if g.State == GameStateGame {
		// Check whose turn it is and handle accordingly
		if g.isBotTurn() {
			// Bot's turn: either waiting for result or needs to start thinking
			if g.IsBotThinking {
				g.handleBotResult()
			} else {
				g.handleBotTurn()
			}
		} else {
			// Player's turn
			g.handlePlayerInput()
		}

		if g.Board.PassCount == 2 {
			g.State = GameStateEnd
		}
		return nil
	}

	if g.State == GameStateEnd {
		if inpututil.IsKeyJustPressed(ebiten.KeySpace) {
			g.State = GameStateIntro
			g.Init()
		}
	}
	return nil
}

// Draw draws the game. It implements ebiten.Game.
func (g *Game) Draw(screen *ebiten.Image) {
	if g.Renderer != nil {
		g.Renderer.Draw(screen, g)
	}
}

// Layout returns the game's logical screen size. It implements ebiten.Game.
func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	if g.Renderer != nil {
		return g.Renderer.Layout(outsideWidth, outsideHeight)
	}
	return 600, 600
}

// SetBot sets the AI bot for the game.
func (g *Game) SetBot(bot AIPlayer) {
	g.Bot = bot
}

// handleIntroState processes input while the game is in the intro state.
func (g *Game) handleIntroState() error {
	if inpututil.IsKeyJustPressed(ebiten.KeyP) {
		g.Mode = GameModePvP
		g.State = GameStateGame
		return nil
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyB) {
		g.Mode = GameModePvE
		g.State = GameStateGame
		return nil
	}

	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		g.Mode = GameModePvP
		g.State = GameStateGame
		return nil
	}

	return nil
}

// handleBotResult checks for and processes bot move results when the bot is thinking.
func (g *Game) handleBotResult() {
	select {
	case botMoveResult := <-g.BotMoveChan:
		g.IsBotThinking = false
		g.processBotMove(botMoveResult)
	default:
		// No result yet, keep waiting
	}
}

// processBotMove processes a completed bot move result.
func (g *Game) processBotMove(botMoveResult BotMoveResult) {
	if botMoveResult.Err != nil {
		fmt.Println("Bot error:", botMoveResult.Err, "- passing instead")
		g.handlePass()
		return
	}

	if botMoveResult.Pass {
		g.handlePass()
		return
	}

	if g.Board.PlaceStone(botMoveResult.X, botMoveResult.Y) {
		g.checkGameEnd()
		return
	}

	// Move failed (shouldn't happen, but handle gracefully)
	fmt.Println("Bot move failed, passing instead")
	g.handlePass()
}

// handlePlayerInput processes player input during gameplay.
func (g *Game) handlePlayerInput() {
	if inpututil.IsKeyJustPressed(ebiten.KeyS) {
		g.handlePass()
		return
	}

	if !inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		return
	}

	cursorX, cursorY := ebiten.CursorPosition()
	gridX, gridY, onBoard := g.Renderer.GetGridPosition(cursorX, cursorY)
	if !onBoard {
		return
	}

	if g.Board.PlaceStone(gridX, gridY) {
		g.checkGameEnd()
	}
}

// isBotTurn reports whether it is currently the bot's turn.
func (g *Game) isBotTurn() bool {
	return g.Mode == GameModePvE && g.Board.CurrentPlayer == PlayerWhite
}

// handleBotTurn starts the bot's move calculation in a goroutine.
func (g *Game) handleBotTurn() {
	if g.Bot == nil {
		g.handlePass()
		return
	}

	g.IsBotThinking = true
	go func() {
		moveX, moveY, err := g.Bot.NextMove(g.Board, PlayerWhite)
		isPassMove := (moveX == -1 && moveY == -1)
		g.BotMoveChan <- BotMoveResult{X: moveX, Y: moveY, Pass: isPassMove, Err: err}
	}()
}

// handlePass executes a pass move.
func (g *Game) handlePass() {
	g.Board.Pass()
}

// checkGameEnd checks if the board is full and ends the game if so.
func (g *Game) checkGameEnd() {
	if g.Board.IsFull() {
		g.State = GameStateEnd
	}
}
