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
	g.Bot = nil
}

// Update processes input and updates game state
func (g *Game) Update() error {
	switch g.State {
	case GameStateIntro:
		return g.handleIntroState()
	case GameStateGame:
		return g.handleGameState()
	case GameStateEnd:
		return g.handleEndState()
	default:
		return nil
	}
}

// handleIntroState processes input for the intro screen
func (g *Game) handleIntroState() error {
	if inpututil.IsKeyJustPressed(ebiten.KeyP) || inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		g.Mode = GameModePvP
		g.State = GameStateGame
	} else if inpututil.IsKeyJustPressed(ebiten.KeyB) {
		g.Mode = GameModePvE
		g.State = GameStateGame
		// Bot will be loaded lazily when needed
	}
	return nil
}

// handleGameState processes input and game logic during gameplay
func (g *Game) handleGameState() error {
	// Check for bot result
	if err := g.processBotMove(); err != nil {
		return err
	}

	if g.IsBotThinking {
		return nil // Don't allow other inputs while thinking
	}

	// Handle all player inputs (mouse clicks and keyboard)
	if err := g.handlePlayerInput(); err != nil {
		return err
	}

	// Trigger bot move if needed
	if err := g.triggerBotMove(); err != nil {
		return err
	}

	// Check for game end condition
	if g.Board.PassCount == 2 {
		g.State = GameStateEnd
	}

	return nil
}

// processBotMove checks for and processes bot move results
func (g *Game) processBotMove() error {
	select {
	case result := <-g.BotMoveChan:
		g.IsBotThinking = false
		if result.Err == nil {
			if !g.Board.PlaceStone(result.X, result.Y) {
				return fmt.Errorf("invalid move")
			}
		} else {
			return fmt.Errorf("bot error: %w", result.Err)
		}
	default:
		// No result yet
	}
	return nil
}

// handlePlayerInput processes all player inputs (mouse clicks for placing stones and keyboard for passing)
func (g *Game) handlePlayerInput() error {
	// Handle pass key
	if inpututil.IsKeyJustPressed(ebiten.KeyS) {
		g.Board.Pass()
		return nil
	}

	// Handle mouse click for placing stones
	if !inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		return nil
	}

	x, y := ebiten.CursorPosition()
	if g.Renderer == nil {
		return fmt.Errorf("renderer is nil, cannot process input")
	}

	row, col, onBoard := g.Renderer.GetGridPosition(x, y)
	if !onBoard {
		return fmt.Errorf("invalid position")
	}

	// In PvE, only allow player (Black) to move manually
	// In PvP, both can move manually (turn logic handled by Board)
	if g.Mode == GameModePvE && g.Board.CurrentPlayer != PlayerBlack {
		// It's bot's turn, ignore click
		return nil
	}

	if !g.Board.PlaceStone(row, col) {
		return fmt.Errorf("invalid move")
	}

	return nil
}

// triggerBotMove starts bot thinking if it's the bot's turn
func (g *Game) triggerBotMove() error {
	// Bot turn (White) - Only in PvE mode
	if g.Mode != GameModePvE || g.Board.CurrentPlayer != PlayerWhite || g.IsBotThinking {
		return nil
	}

	if g.Bot == nil {
		// Try to load bot lazily
		g.Bot = g.loadBot()
		if g.Bot == nil {
			// No bot available, pass
			g.Board.Pass()
			return nil
		}
	}

	g.IsBotThinking = true
	go func() {
		x, y, err := g.Bot.NextMove(g.Board, PlayerWhite)
		g.BotMoveChan <- BotMoveResult{X: x, Y: y, Err: err}
	}()
	return nil
}

// handleEndState processes input for the end game screen
func (g *Game) handleEndState() error {
	if inpututil.IsKeyJustPressed(ebiten.KeySpace) {
		g.State = GameStateIntro
		g.Init()
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

// loadBot attempts to load the AI bot (called lazily to avoid circular import)
// Returns nil if loading fails - actual implementation should be set via SetBot
func (g *Game) loadBot() AIPlayer {
	// This is a placeholder - actual loading happens via SetBot from main.go
	// to avoid circular import between game and ai packages
	return g.Bot
}

// SetBot sets the AI bot for the game (called from main.go to break circular import)
func (g *Game) SetBot(bot AIPlayer) {
	g.Bot = bot
}
