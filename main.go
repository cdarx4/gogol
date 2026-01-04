// ============================================================================
// File: main.go
// Author: Ding Lionel, Darx Christian
// Email: lionel.ding@hes-so.ch, christia.darx@hes-so.ch
// Date: 28.11.2025
// Description: Main file for this GoGol game.
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

package main

import (
	"log"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/joho/godotenv"

	"heia2526/gogol/ai"
	"heia2526/gogol/game"
	"heia2526/gogol/ui"
)

// Define the window size and title
const (
	WindowWidth  = 600
	WindowHeight = 600
	WindowTitle  = "GoGol - Go game"
)

// Main entry point of the program
func main() {
	// Try loading from .env file for local development (optional)
	// This will be ignored if the file doesn't exist (e.g., in WASM or CI)
	// Environment variables can be set directly or via system env vars
	_ = godotenv.Load() // Silently ignore if .env file doesn't exist

	g := &game.Game{}
	g.Init()

	// Try to load AI model for PvE mode (non-blocking - game will handle failure gracefully)
	model, err := ai.LoadModel("models/champion.json")
	if err != nil {
		log.Printf("Note: AI model not found at models/champion.json. PvE mode will not be available. Error: %v", err)
	} else {
		g.SetBot(ai.NewMLPPlayer(model))
	}

	renderer := ui.NewRenderer()
	g.Renderer = renderer

	// Set the window size and title
	ebiten.SetWindowSize(WindowWidth, WindowHeight)
	ebiten.SetWindowTitle(WindowTitle)

	// Runs the game
	if err := ebiten.RunGame(g); err != nil {
		log.Fatal(err)
	}
}
