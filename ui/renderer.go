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

type Renderer struct{}

func (r *Renderer) Draw(screen *ebiten.Image, g *game.Game) {
	switch g.State {
	case game.GameStateIntro:
		r.drawIntro(screen)
	case game.GameStateGame:
		r.drawBoard(screen)
	case game.GameStateEnd:
		r.drawBoard(screen)
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

func (r *Renderer) drawBoard(screen *ebiten.Image) {
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
}
