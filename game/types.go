package game

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

// Represents the different states of the game
type GameStates string

// Represents the different players
type Player int

// For the stone groups
// TODO add the liberties
type Group struct {
	ID     int
	Player Player
	Stones []*Stone
}

// Renderer interface for the game
type Renderer interface {
	Draw(screen *ebiten.Image, game *Game)
	Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int)
}

// Represents the struct for the game itself
type Game struct {
	State    GameStates
	Renderer Renderer
}

// Different states of the game
// TODO add a configuration state later
const (
	GameStateIntro GameStates = "intro"
	GameStateGame  GameStates = "game"
	GameStateEnd   GameStates = "end"
)

// Different players White/Black
const (
	PlayerBlack Player = iota
	PlayerWhite
)

// Size of the board 9x9
const BoardSize = 9

// Initialize the game
func (g *Game) Init() {
	g.State = GameStateIntro
}

// To pass to the next state when the user clicks or presses space
func (g *Game) Update() error {
	if g.State == GameStateIntro {
		if inpututil.IsKeyJustPressed(ebiten.KeySpace) || inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
			g.State = GameStateGame
		}
	}
	return nil
}

// Draw the game
func (g *Game) Draw(screen *ebiten.Image) {
	if g.Renderer != nil {
		g.Renderer.Draw(screen, g)
	}
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	if g.Renderer != nil {
		return g.Renderer.Layout(outsideWidth, outsideHeight)
	}
	return 600, 600
}
