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

	"heia2526/gogol/ai"
	"heia2526/gogol/assets"
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

	g := &game.Game{}
	g.Init()

	// Load AI model from embedded assets for PvE mode (WASM compatible)
	// Non-blocking - game will handle failure gracefully
	model, err := ai.LoadModelFromBytes(assets.ChampionModelJSON)
	if err != nil {
		log.Printf("Note: Failed to load embedded AI model. PvE mode will not be available. Error: %v", err)
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
