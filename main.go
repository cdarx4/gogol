package main

import (
	"image/color"
	"log"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

const (
	boardSize   = 9
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

type GameStates string

const (
	GameStateIntro GameStates = "intro"
	GameStateGame  GameStates = "game"
	GameStateEnd   GameStates = "end"
)

type Game struct {
	state GameStates
}

func (g *Game) Update() error {
	// Transition from intro to game on any key press or mouse click
	if g.state == GameStateIntro {
		if inpututil.IsKeyJustPressed(ebiten.KeySpace) || inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
			g.state = GameStateGame
		}
	}
	return nil
}

func (g *Game) drawIntro(screen *ebiten.Image) {
	// Fill background
	screen.Fill(boardColor)

	// Draw title "GoGol"
	size := screen.Bounds().Size()
	width, height := size.X, size.Y
	ebitenutil.DebugPrintAt(screen, introTitle, width/2-30, height/2-20)
	ebitenutil.DebugPrintAt(screen, introSub, width/2-100, height/2+20)
}

func (g *Game) drawBoard(screen *ebiten.Image) {
	// Fill background with board color (light beige/wood)
	screen.Fill(boardColor)
	// Calculate board dimensions
	boardWidth := (boardSize - 1) * cellSize
	boardHeight := (boardSize - 1) * cellSize
	startX := float32(boardMargin)
	startY := float32(boardMargin)

	// Draw vertical lines
	for i := 0; i < boardSize; i++ {
		x := startX + float32(i*cellSize)
		vector.StrokeLine(screen, x, startY, x, startY+float32(boardHeight), lineWidth, lineColor, false)
	}

	// Draw horizontal lines
	for i := 0; i < boardSize; i++ {
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

func (g *Game) Draw(screen *ebiten.Image) {
	switch g.state {
	case GameStateIntro:
		g.drawIntro(screen)
	case GameStateGame:
		g.drawBoard(screen)
	case GameStateEnd:
		g.drawBoard(screen)
	}
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	boardWidth := (boardSize-1)*cellSize + boardMargin*2
	boardHeight := (boardSize-1)*cellSize + boardMargin*2
	return boardWidth, boardHeight
}

// Initialize the game
func (g *Game) Init() {
	g.state = GameStateIntro
}

func main() {
	game := &Game{}
	game.Init()
	ebiten.SetWindowSize((boardSize-1)*cellSize+boardMargin*2, (boardSize-1)*cellSize+boardMargin*2)
	ebiten.SetWindowTitle("GoGol - 9x9 Board")
	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}
