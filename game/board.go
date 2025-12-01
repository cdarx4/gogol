package game

type Board struct {
	Size          int
	Grid          [][]*Stone
	Groups        []*Group
	nextGroupId   int
	currentPlayer Player
}

func NewBoard(size int) *Board {
	grid := make([][]*Stone, size)
	for i := range grid {
		grid[i] = make([]*Stone, size)
	}
	return &Board{
		Size:          size,
		Grid:          grid,
		currentPlayer: PlayerBlack,
	}
}

func (b *Board) PlaceStone(x, y int) bool {
	if x < 0 || x >= b.Size || y < 0 || y >= b.Size {
		return false
	}

	// Check if the spot is already occupied
	if b.Grid[x][y] != nil {
		return false
	}

	stone := &Stone{
		X:      x,
		Y:      y,
		Player: b.currentPlayer,
	}
	b.Grid[x][y] = stone

	// Switch turns
	if b.currentPlayer == PlayerBlack {
		b.currentPlayer = PlayerWhite
	} else {
		b.currentPlayer = PlayerBlack
	}

	return true
}
