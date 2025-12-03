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
	_ "embed"
	"log"
	"os"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/joho/godotenv"

	"heia2526/gogol/game"
	"heia2526/gogol/ui"
)

//go:embed .env
var envFile string

// Define the window size and title
const (
	WindowWidth  = 600
	WindowHeight = 600
	WindowTitle  = "GoGol - 9x9 Board"
	EnvFileError = "No .env file found"
)

// Main entry point of the program
func main() {
	// Parse the embedded .env file
	envMap, err := godotenv.Unmarshal(envFile)
	if err != nil {
		log.Println(EnvFileError)
	}

	// Set environment variables
	for key, value := range envMap {
		os.Setenv(key, value)
	}

	// Also try loading from file for local development if embedded is empty (optional, but good for dev)
	// However, since we embed it, it should be there.
	// If we want to allow overriding with a local file, we could call Load() afterwards.
	// But for WASM, Load() fails. So let's stick to the embedded one or try Load and ignore error.
	if err := godotenv.Load(); err != nil {
		// Ignore error if file not found (likely in WASM or if relying on embed)
		if !strings.Contains(err.Error(), "no such file") {
			// log.Println("Error loading .env file:", err)
		}
	}

	g := &game.Game{}
	g.Init()

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
