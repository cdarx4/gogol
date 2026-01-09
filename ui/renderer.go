// ============================================================================
// File: renderer.go
// Author: Ding Lionel, Darx Christian
// Email: lionel.ding@hes-so.ch, christia.darx@hes-so.ch
// Date: 28.11.2025
// Description: Renderer for the GoGol game.
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

package ui

import (
	"bytes"
	"fmt"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/examples/resources/fonts"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"

	"heia2526/gogol/assets"
	"heia2526/gogol/game"
)

const (
	// Board sizing/layout constants
	cellSize    = 60
	boardMargin = 50
	lineWidth   = 2

	// UI text and font sizes
	introTitle       = "GoGol"
	PVPText          = "Press P or Space to start a PVP game"
	PVEText          = "Press B to start a PVE game"
	ThinkingText     = "Thinking..."
	PassText         = "Press S to pass"
	GameOverText     = "\nPress Space to Restart"
	TitleFontSize    = 24
	SubTitleFontSize = 12

	// Stone rendering constants
	stoneScaleFactor = 0.9

	// Star point (hoshi) rendering constants
	starPointRadius = 4

	// Game over overlay constants
	gameOverRectWidth  = 400
	gameOverRectHeight = 100
	gameOverAlpha      = 180
)

// Colors and font sources for the UI
var (
	boardColor      = color.RGBA{220, 179, 92, 255}
	lineColor       = color.RGBA{0, 0, 0, 255}
	mplusFaceSource *text.GoTextFaceSource
)

// Renderer handles all rendering operations for the game.
type Renderer struct {
	// BlackStone is the image used for black stones.
	BlackStone *ebiten.Image
	// WhiteStone is the image used for white stones.
	WhiteStone *ebiten.Image
}

// NewRenderer creates and initializes a new renderer with loaded stone images and fonts.
func NewRenderer() *Renderer {
	// Load black stone image from embedded assets
	blackStoneImage, _, err := ebitenutil.NewImageFromReader(bytes.NewReader(assets.BlackStonePNG))
	if err != nil {
		panic(fmt.Errorf("failed to load embedded black stone: %w", err))
	}

	// Load white stone image from embedded assets
	whiteStoneImage, _, err := ebitenutil.NewImageFromReader(bytes.NewReader(assets.WhiteStonePNG))
	if err != nil {
		panic(fmt.Errorf("failed to load embedded white stone: %w", err))
	}

	// Load font face source for text rendering
	fontFaceSource, err := text.NewGoTextFaceSource(bytes.NewReader(fonts.PressStart2P_ttf))
	if err != nil {
		panic(err)
	}
	mplusFaceSource = fontFaceSource

	return &Renderer{
		BlackStone: blackStoneImage,
		WhiteStone: whiteStoneImage,
	}
}

// Draw draws the game based on its current state.
func (r *Renderer) Draw(screen *ebiten.Image, currentGame *game.Game) {
	switch currentGame.State {
	case game.GameStateIntro:
		r.drawIntro(screen)
	case game.GameStateGame:
		r.drawBoard(screen, currentGame)
	case game.GameStateEnd:
		r.drawBoard(screen, currentGame)
		r.drawGameOver(screen, currentGame)
	}
}

// Layout returns the logical screen size for the game.
func (r *Renderer) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	boardWidth := (game.BoardSize-1)*cellSize + boardMargin*2
	boardHeight := (game.BoardSize-1)*cellSize + boardMargin*2
	return boardWidth, boardHeight
}

// drawIntro draws the introduction screen.
func (r *Renderer) drawIntro(screen *ebiten.Image) {
	// Fill the background with the board color
	screen.Fill(boardColor)

	screenSize := screen.Bounds().Size()
	screenWidth, screenHeight := float64(screenSize.X), float64(screenSize.Y)

	// Set up the text faces for title and subtitles
	titleFace := &text.GoTextFace{
		Source: mplusFaceSource,
		Size:   TitleFontSize,
	}
	subtitleFace := &text.GoTextFace{
		Source: mplusFaceSource,
		Size:   SubTitleFontSize,
	}

	// Measure how big each text will be
	titleWidth, titleHeight := text.Measure(introTitle, titleFace, TitleFontSize)
	pvpWidth, pvpHeight := text.Measure(PVPText, subtitleFace, SubTitleFontSize)
	pveWidth, pveHeight := text.Measure(PVEText, subtitleFace, SubTitleFontSize)

	// Calculate spacing and total height to center everything vertically
	lineSpacing := float64(SubTitleFontSize)
	totalTextHeight := titleHeight + pvpHeight + pveHeight + 2*lineSpacing
	currentY := (screenHeight - totalTextHeight) / 2

	// Draw the title centered
	textOptions := &text.DrawOptions{}
	textOptions.GeoM.Translate((screenWidth-titleWidth)/2, currentY)
	textOptions.ColorScale.ScaleWithColor(color.White)
	text.Draw(screen, introTitle, titleFace, textOptions)

	currentY += titleHeight + lineSpacing

	// Draw the PVP option text
	textOptions.GeoM.Reset()
	textOptions.GeoM.Translate((screenWidth-pvpWidth)/2, currentY)
	textOptions.ColorScale.ScaleWithColor(color.White)
	text.Draw(screen, PVPText, subtitleFace, textOptions)

	currentY += pvpHeight + lineSpacing

	// Draw the PVE option text
	textOptions.GeoM.Reset()
	textOptions.GeoM.Translate((screenWidth-pveWidth)/2, currentY)
	textOptions.ColorScale.ScaleWithColor(color.White)
	text.Draw(screen, PVEText, subtitleFace, textOptions)
}

// drawBoard draws the game board with grid lines, star points, stones, and UI text.
func (r *Renderer) drawBoard(screen *ebiten.Image, currentGame *game.Game) {
	// Fill the background with the board color
	screen.Fill(boardColor)

	// Calculate where the board starts and its dimensions
	boardWidth := (game.BoardSize - 1) * cellSize
	boardHeight := (game.BoardSize - 1) * cellSize
	startX := float32(boardMargin)
	startY := float32(boardMargin)

	// Draw all the vertical grid lines
	for columnIndex := 0; columnIndex < game.BoardSize; columnIndex++ {
		lineX := startX + float32(columnIndex*cellSize)
		vector.StrokeLine(screen, lineX, startY, lineX, startY+float32(boardHeight), lineWidth, lineColor, false)
	}

	// Draw all the horizontal grid lines
	for rowIndex := 0; rowIndex < game.BoardSize; rowIndex++ {
		lineY := startY + float32(rowIndex*cellSize)
		vector.StrokeLine(screen, startX, lineY, startX+float32(boardWidth), lineY, lineWidth, lineColor, false)
	}

	// Draw the star points (hoshi) at the standard positions
	starPoints := [][]int{
		{2, 2}, {2, 6},
		{6, 2}, {6, 6},
		{4, 4},
	}
	for _, point := range starPoints {
		starX := startX + float32(point[0]*cellSize)
		starY := startY + float32(point[1]*cellSize)
		vector.FillCircle(screen, starX, starY, starPointRadius, lineColor, false)
	}

	// Draw all the stones on the board
	if currentGame.Board != nil {
		for x := 0; x < game.BoardSize; x++ {
			for y := 0; y < game.BoardSize; y++ {
				stone := currentGame.Board.Grid[x][y]
				if stone != nil {
					stoneX := float64(startX) + float64(x*cellSize)
					stoneY := float64(startY) + float64(y*cellSize)

					// Choose the right stone image based on player
					var stoneImage *ebiten.Image
					if stone.Player == game.PlayerBlack {
						stoneImage = r.BlackStone
					} else {
						stoneImage = r.WhiteStone
					}

					// Draw the stone image if we have it
					if stoneImage != nil {
						drawOptions := &ebiten.DrawImageOptions{}
						imageSize := stoneImage.Bounds().Size()
						imageWidth, imageHeight := imageSize.X, imageSize.Y

						// Scale the image to fit nicely in the cell
						scale := float64(cellSize) * stoneScaleFactor / float64(imageWidth)
						drawOptions.GeoM.Scale(scale, scale)

						// Center the image on the intersection
						drawOptions.GeoM.Translate(stoneX-float64(imageWidth)*scale/2, stoneY-float64(imageHeight)*scale/2)

						screen.DrawImage(stoneImage, drawOptions)
					} else {
						// Fallback: draw a circle if image loading failed
						radius := float32(cellSize) / 2 * stoneScaleFactor
						var stoneColor color.Color
						if stone.Player == game.PlayerBlack {
							stoneColor = color.Black
						} else {
							stoneColor = color.White
						}
						vector.FillCircle(screen, float32(stoneX), float32(stoneY), radius, stoneColor, true)
					}
				}
			}
		}
	}

	// Show UI text at the bottom of the screen
	screenSize := screen.Bounds().Size()
	screenWidth, screenHeight := screenSize.X, screenSize.Y

	if currentGame.IsBotThinking {
		ebitenutil.DebugPrintAt(screen, ThinkingText, screenWidth/2-30, screenHeight-20)
	} else {
		ebitenutil.DebugPrintAt(screen, PassText, screenWidth/2-50, screenHeight-20)
	}
}

// GetGridPosition converts screen coordinates to board grid coordinates.
func (r *Renderer) GetGridPosition(screenX, screenY int) (gridX, gridY int, onBoard bool) {
	// Figure out where the board starts
	startX := boardMargin
	startY := boardMargin

	// Calculate the board's actual size
	boardWidth := (game.BoardSize - 1) * cellSize
	boardHeight := (game.BoardSize - 1) * cellSize

	// Check if the click is anywhere near the board (with some margin for easier clicking)
	if screenX < startX-cellSize/2 || screenX > startX+boardWidth+cellSize/2 ||
		screenY < startY-cellSize/2 || screenY > startY+boardHeight+cellSize/2 {
		return 0, 0, false
	}

	// Convert screen coordinates to grid coordinates by rounding to nearest intersection
	floatX := float64(screenX - startX)
	floatY := float64(screenY - startY)

	gridX = int((floatX + float64(cellSize)/2) / float64(cellSize))
	gridY = int((floatY + float64(cellSize)/2) / float64(cellSize))

	// Make sure the coordinates are actually on the board
	if gridX >= 0 && gridX < game.BoardSize && gridY >= 0 && gridY < game.BoardSize {
		return gridX, gridY, true
	}

	return 0, 0, false
}

// drawGameOver draws the game over overlay with winner information.
func (r *Renderer) drawGameOver(screen *ebiten.Image, currentGame *game.Game) {
	winner, isDraw := currentGame.Board.GetWinner()
	blackCount, whiteCount := currentGame.Board.GetStoneCount()

	// Build the message based on who won
	var messageText string
	if isDraw {
		messageText = fmt.Sprintf("Draw! Black: %d, White: %d", blackCount, whiteCount)
	} else if *winner == game.PlayerBlack {
		messageText = fmt.Sprintf("Black Wins! Black: %d, White: %d", blackCount, whiteCount)
	} else {
		messageText = fmt.Sprintf("White Wins! Black: %d, White: %d", blackCount, whiteCount)
	}
	messageText += GameOverText

	// Get screen dimensions to center the overlay
	screenSize := screen.Bounds().Size()
	screenWidth, screenHeight := screenSize.X, screenSize.Y

	// Draw a semi-transparent black background for the overlay
	overlayImage := ebiten.NewImage(gameOverRectWidth, gameOverRectHeight)
	overlayImage.Fill(color.RGBA{0, 0, 0, gameOverAlpha})
	drawOptions := &ebiten.DrawImageOptions{}
	drawOptions.GeoM.Translate(float64(screenWidth/2-gameOverRectWidth/2), float64(screenHeight/2-gameOverRectHeight/2))
	screen.DrawImage(overlayImage, drawOptions)

	// Draw the message text centered on the overlay
	ebitenutil.DebugPrintAt(screen, messageText, screenWidth/2-100, screenHeight/2-20)
}
