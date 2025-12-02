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
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/vector"

	"heia2526/gogol/game"
)

const (
	cellSize    = 60
	boardMargin = 50
	lineWidth   = 2
	introTitle  = "GoGol"
	introSub    = "Press SPACE or Click to Start"
)

var (
	boardColor = color.RGBA{220, 179, 92, 255}
	lineColor  = color.RGBA{0, 0, 0, 255}
)

type Renderer struct {
	BlackStone *ebiten.Image
	WhiteStone *ebiten.Image
}

func NewRenderer() *Renderer {
	black, _, err := ebitenutil.NewImageFromFile("images/black-stone.png")
	if err != nil {
		panic(err)
	}
	white, _, err := ebitenutil.NewImageFromFile("images/white-stone.png")
	if err != nil {
		panic(err)
	}
	return &Renderer{
		BlackStone: black,
		WhiteStone: white,
	}
}

func (r *Renderer) Draw(screen *ebiten.Image, g *game.Game) {
	switch g.State {
	case game.GameStateIntro:
		r.drawIntro(screen)
	case game.GameStateGame:
		r.drawBoard(screen, g)
	case game.GameStateEnd:
		r.drawBoard(screen, g)
	}
}

func (r *Renderer) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	boardWidth := (game.BoardSize-1)*cellSize + boardMargin*2
	boardHeight := (game.BoardSize-1)*cellSize + boardMargin*2
	return boardWidth, boardHeight
}

func (r *Renderer) drawIntro(screen *ebiten.Image) {
	// Fill background
	screen.Fill(boardColor)

	// Draw title "GoGol"
	size := screen.Bounds().Size()
	width, height := size.X, size.Y
	ebitenutil.DebugPrintAt(screen, introTitle, width/2-30, height/2-20)
	ebitenutil.DebugPrintAt(screen, introSub, width/2-100, height/2+20)
}

func (r *Renderer) drawBoard(screen *ebiten.Image, g *game.Game) {
	// Fill background with board color (light beige/wood)
	screen.Fill(boardColor)
	// Calculate board dimensions
	boardWidth := (game.BoardSize - 1) * cellSize
	boardHeight := (game.BoardSize - 1) * cellSize
	startX := float32(boardMargin)
	startY := float32(boardMargin)

	// Draw vertical lines
	for i := 0; i < game.BoardSize; i++ {
		x := startX + float32(i*cellSize)
		vector.StrokeLine(screen, x, startY, x, startY+float32(boardHeight), lineWidth, lineColor, false)
	}

	// Draw horizontal lines
	for i := 0; i < game.BoardSize; i++ {
		y := startY + float32(i*cellSize)
		vector.StrokeLine(screen, startX, y, startX+float32(boardWidth), y, lineWidth, lineColor, false)
	}

	// Draw star points (hoshi)
	starPoints := [][]int{
		{2, 2}, {2, 6},
		{6, 2}, {6, 6},
		{4, 4},
	}
	starRadius := float32(4)
	for _, point := range starPoints {
		x := startX + float32(point[0]*cellSize)
		y := startY + float32(point[1]*cellSize)
		vector.FillCircle(screen, x, y, starRadius, lineColor, false)
	}

	// Draw stones
	if g.Board != nil {
		for i := 0; i < game.BoardSize; i++ {
			for j := 0; j < game.BoardSize; j++ {
				stone := g.Board.Grid[i][j]
				if stone != nil {
					x := float64(startX) + float64(i*cellSize)
					y := float64(startY) + float64(j*cellSize)

					var img *ebiten.Image
					if stone.Player == game.PlayerBlack {
						img = r.BlackStone
					} else {
						img = r.WhiteStone
					}

					if img != nil {
						op := &ebiten.DrawImageOptions{}
						// Center the image
						size := img.Bounds().Size()
						width, height := size.X, size.Y
						// Scale to fit cellSize (slightly smaller)
						scale := float64(cellSize) * 0.9 / float64(width)
						op.GeoM.Scale(scale, scale)
						// Center the image on the corss section
						op.GeoM.Translate(x-float64(width)*scale/2, y-float64(height)*scale/2)

						screen.DrawImage(img, op)
					} else {
						// Fallback if image load failed (shouldn't happen with NewRenderer panic)
						radius := float32(cellSize) / 2 * 0.9
						var c color.Color
						if stone.Player == game.PlayerBlack {
							c = color.Black
						} else {
							c = color.White
						}
						vector.FillCircle(screen, float32(x), float32(y), radius, c, true)
					}
				}
			}
		}
	}
}

func (r *Renderer) GetGridPosition(x, y int) (row, col int, onBoard bool) {
	// Calculate board dimensions
	startX := boardMargin
	startY := boardMargin

	// Check if click is within reasonable bounds of the board
	// We allow some margin around the board for clicking
	boardWidth := (game.BoardSize - 1) * cellSize
	boardHeight := (game.BoardSize - 1) * cellSize

	if x < startX-cellSize/2 || x > startX+boardWidth+cellSize/2 ||
		y < startY-cellSize/2 || y > startY+boardHeight+cellSize/2 {
		return 0, 0, false
	}

	// Calculate nearest intersection
	// (x - startX) / cellSize
	// We want to round to the nearest integer

	fx := float64(x - startX)
	fy := float64(y - startY)

	ix := int((fx + float64(cellSize)/2) / float64(cellSize))
	iy := int((fy + float64(cellSize)/2) / float64(cellSize))

	if ix >= 0 && ix < game.BoardSize && iy >= 0 && iy < game.BoardSize {
		return ix, iy, true
	}

	return 0, 0, false
}
