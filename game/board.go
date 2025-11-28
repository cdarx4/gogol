package game

type Board struct {
	Size          int
	Grid          [][]*Stone
	Groups        []*Group
	nextGroupId   int
	currentPlayer Player
}
