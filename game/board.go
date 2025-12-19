// ============================================================================
// File: board.go
// Author: Ding Lionel, Darx Christian
// Email: lionel.ding@hes-so.ch, christia.darx@hes-so.ch
// Date: 28.11.2025
// Description: Go file managing the board.
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

package game

// ---------- Board creation ----------

// Creates and returns a new empty board of given size.
// Panics if size is not positive.
func NewBoard(size int) *Board {
	if size <= 0 {
		panic("board size must be positive")
	}

	grid := make([][]*Stone, size)
	for i := range grid {
		grid[i] = make([]*Stone, size)
	}

	return &Board{
		Size:          size,
		Grid:          grid,
		CurrentPlayer: StartingPlayer,
	}
}

// ---------- Helpers ----------

// Returns true if (x, y) is within the board boundaries.
func (b *Board) isInBounds(x, y int) bool {
	return x >= 0 && x < b.Size && y >= 0 && y < b.Size
}

// Returns true if (x, y) is occupied by a stone.
func (b *Board) isOccupied(x, y int) bool {
	return b.Grid[x][y] != nil
}

// ---------- Group & liberty recomputation ----------

// Recomputes the liberties of a group by checking all empty adjacent points for every stone in the group.
func (b *Board) recomputeGroupLiberties(group *Group) {
	group.Liberties = make(map[Position]struct{})

	for _, s := range group.Stones {
		for _, d := range AdjacentDirections {
			nx, ny := s.X+d.X, s.Y+d.Y
			if b.isInBounds(nx, ny) && b.Grid[nx][ny] == nil {
				group.Liberties[Position{X: nx, Y: ny}] = struct{}{}
			}
		}
	}
}

// Recomputes liberties for all groups on the board.
func (b *Board) recomputeAllGroupLiberties() {
	for _, g := range b.Groups {
		b.recomputeGroupLiberties(g)
	}
}

// ---------- Group creation / merging ----------

// Creates a new group for the given stone and adds it to the board's group list.
func (b *Board) createGroup(stone *Stone) *Group {
	group := &Group{
		ID:     b.nextGroupId,
		Player: stone.Player,
		Stones: []*Stone{stone},
	}
	b.nextGroupId++
	b.Groups = append(b.Groups, group)
	stone.Group = group
	return group
}

// Merges other groups into the target group (used when placing a stone connects separate friendly groups).
func (b *Board) mergeGroups(target *Group, others []*Group) {
	for _, g := range others {
		for _, s := range g.Stones {
			s.Group = target
			target.Stones = append(target.Stones, s)
		}
	}
	b.removeGroups(others)
}

// Removes the listed groups from the board's group list.
func (b *Board) removeGroups(groups []*Group) {
	filtered := []*Group{}
	for _, g := range b.Groups {
		keep := true
		for _, r := range groups {
			if g.ID == r.ID {
				keep = false
				break
			}
		}
		if keep {
			filtered = append(filtered, g)
		}
	}
	b.Groups = filtered
}

// ---------- Capture logic ----------

// Removes all stones in the group from the board and updates group list.
func (b *Board) captureGroup(group *Group) {
	for _, s := range group.Stones {
		b.Grid[s.X][s.Y] = nil
	}
	b.removeGroups([]*Group{group})
}

// ---------- Suicide check (simulation) ----------

// Checks if placing a stone of `player` at (x, y) would be a suicide (illegal self-capture).
// Returns true if the move is suicide and thus illegal.
func (b *Board) isSuicide(x, y int, player Player) bool {
	// Note: tempStone is not used but would represent the hypothetical stone.
	// This function simulates the consequences of the move without modifying board state.

	adjGroups := b.adjacentGroups(x, y)
	friendly := []*Group{}
	opponent := []*Group{}

	// Separate adjacent friendly and opponent groups.
	for _, g := range adjGroups {
		if g.Player == player {
			friendly = append(friendly, g)
		} else {
			opponent = append(opponent, g)
		}
	}

	// If any opponent group would be captured by this move, it's not suicide.
	movePos := Position{X: x, Y: y}
	for _, g := range opponent {
		b.recomputeGroupLiberties(g)
		if len(g.Liberties) == 1 {
			if _, ok := g.Liberties[movePos]; ok {
				return false
			}
		}
	}

	// Simulate merged group liberties after this move.
	liberties := make(map[Position]struct{})
	for _, d := range AdjacentDirections {
		nx, ny := x+d.X, y+d.Y
		if b.isInBounds(nx, ny) && b.Grid[nx][ny] == nil {
			liberties[Position{X: nx, Y: ny}] = struct{}{}
		}
	}

	// Add liberties from adjacent friendly groups (excluding the spot itself)
	for _, g := range friendly {
		b.recomputeGroupLiberties(g)
		for l := range g.Liberties {
			if l != movePos {
				liberties[l] = struct{}{}
			}
		}
	}

	// If no liberties remain, it's suicide.
	return len(liberties) == 0
}

// ---------- Adjacent groups ----------

// Returns a slice of all unique groups directly adjacent to (x, y).
func (b *Board) adjacentGroups(x, y int) []*Group {
	seen := map[int]struct{}{}
	var result []*Group

	for _, d := range AdjacentDirections {
		nx, ny := x+d.X, y+d.Y
		if !b.isInBounds(nx, ny) {
			continue
		}
		s := b.Grid[nx][ny]
		if s != nil && s.Group != nil {
			if _, ok := seen[s.Group.ID]; !ok {
				seen[s.Group.ID] = struct{}{}
				result = append(result, s.Group)
			}
		}
	}
	return result
}

// ---------- Place stone (main rule engine) ----------

// Places a stone for the current player at (x, y) if legal, enacts all necessary group/capture logic, and advances the turn.
// Returns true if the move was successful.
func (b *Board) PlaceStone(x, y int) bool {
	// Reset pass count
	b.PassCount = 0

	// 1. Basic legality checks: must be on board, location must be empty.
	if !b.isInBounds(x, y) || b.isOccupied(x, y) {
		return false
	}

	// 2. Illegal move if it would be suicide (self-capture).
	if b.isSuicide(x, y, b.CurrentPlayer) {
		return false
	}

	// 3. Place the new stone and create its group.
	stone := &Stone{X: x, Y: y, Player: b.CurrentPlayer}
	b.Grid[x][y] = stone
	newGroup := b.createGroup(stone)

	// 4. Merge with any adjacent friendly groups created by this placement.
	adj := b.adjacentGroups(x, y)
	toMerge := []*Group{}

	for _, g := range adj {
		if g.Player == b.CurrentPlayer {
			toMerge = append(toMerge, g)
		}
	}

	if len(toMerge) > 0 {
		b.mergeGroups(newGroup, toMerge)
	}

	// 5. Recompute liberties after merge and before captures.
	b.recomputeAllGroupLiberties()

	// 6. Capture any adjacent opponent groups that have lost all liberties.
	for _, g := range b.Groups {
		if g.Player != b.CurrentPlayer && len(g.Liberties) == 0 {
			b.captureGroup(g)
		}
	}

	// 7. Recompute liberties again after any captures.
	b.recomputeAllGroupLiberties()

	// 8. Switch to the next player's turn.
	b.switchTurn()
	return true
}

// ---------- Turn ----------

// Switches current player to their opponent.
func (b *Board) switchTurn() {
	b.CurrentPlayer = b.CurrentPlayer.Opponent()
}

// Pass allows the current player to skip their turn.
func (b *Board) Pass() {
	b.switchTurn()
	b.PassCount++
}

// GetStoneCount returns the number of stones for each player
func (b *Board) GetStoneCount() (blackCount, whiteCount int) {
	for x := 0; x < b.Size; x++ {
		for y := 0; y < b.Size; y++ {
			if b.Grid[x][y] != nil {
				if b.Grid[x][y].Player == PlayerBlack {
					blackCount++
				} else {
					whiteCount++
				}
			}
		}
	}
	return
}

// GetWinner returns the winner based on stone count
// Returns nil if draw
func (b *Board) GetWinner() (winner *Player, isDraw bool) {
	black, white := b.GetStoneCount()
	if black > white {
		p := PlayerBlack
		return &p, false
	} else if white > black {
		p := PlayerWhite
		return &p, false
	}
	return nil, true
}

// ---------- Board cloning ----------

// Clone creates a deep copy of the board
func (b *Board) Clone() *Board {
	cloned := &Board{
		Size:          b.Size,
		Grid:          make([][]*Stone, b.Size),
		Groups:        make([]*Group, 0, len(b.Groups)),
		nextGroupId:   b.nextGroupId,
		CurrentPlayer: b.CurrentPlayer,
		PassCount:     b.PassCount,
	}

	// Create new grid
	for i := range cloned.Grid {
		cloned.Grid[i] = make([]*Stone, b.Size)
	}

	// Map old groups to new groups
	groupMap := make(map[int]*Group)

	// Clone stones and create groups
	for x := 0; x < b.Size; x++ {
		for y := 0; y < b.Size; y++ {
			if b.Grid[x][y] != nil {
				oldStone := b.Grid[x][y]
				newStone := &Stone{
					X:      oldStone.X,
					Y:      oldStone.Y,
					Player: oldStone.Player,
				}
				cloned.Grid[x][y] = newStone

				// Create or reuse group
				if oldStone.Group != nil {
					if newGroup, exists := groupMap[oldStone.Group.ID]; !exists {
						newGroup = &Group{
							ID:        oldStone.Group.ID,
							Player:    oldStone.Group.Player,
							Stones:    []*Stone{newStone},
							Liberties: make(map[Position]struct{}),
						}
						groupMap[oldStone.Group.ID] = newGroup
						cloned.Groups = append(cloned.Groups, newGroup)
						newStone.Group = newGroup
					} else {
						newGroup.Stones = append(newGroup.Stones, newStone)
						newStone.Group = newGroup
					}
				}
			}
		}
	}

	// Recompute liberties for all groups
	cloned.recomputeAllGroupLiberties()

	return cloned
}

// ---------- Legal moves ----------

// GetLegalMoves returns all legal moves for the given player
// Includes pass move represented as Position{X: -1, Y: -1}
func (b *Board) GetLegalMoves(p Player) []Position {
	moves := make([]Position, 0)

	// Add pass move
	moves = append(moves, Position{X: -1, Y: -1})

	// Check all board positions
	for x := 0; x < b.Size; x++ {
		for y := 0; y < b.Size; y++ {
			// Check if position is empty
			if !b.isOccupied(x, y) {
				// Check if move is legal (not suicide)
				if !b.isSuicide(x, y, p) {
					moves = append(moves, Position{X: x, Y: y})
				}
			}
		}
	}

	return moves
}

// CloneAndPlay clones the board, applies the move, and returns the new board
// Returns (cloned board, success)
// Does not modify the original board
func (b *Board) CloneAndPlay(x, y int, p Player) (*Board, bool) {
	cloned := b.Clone()

	// Handle pass move
	pos := Position{X: x, Y: y}
	if pos.IsPass() {
		cloned.Pass()
		return cloned, true
	}

	// Check if move is legal on cloned board
	if !cloned.isInBounds(x, y) || cloned.isOccupied(x, y) {
		return cloned, false
	}

	if cloned.isSuicide(x, y, p) {
		return cloned, false
	}

	// Temporarily set current player to apply move
	originalPlayer := cloned.CurrentPlayer
	cloned.CurrentPlayer = p

	// Apply the move
	success := cloned.PlaceStone(x, y)

	// Restore original player state (PlaceStone switches turn, so we need to handle this)
	if !success {
		cloned.CurrentPlayer = originalPlayer
	}

	return cloned, success
}

// ---------- Debug ----------

// Returns a string representing the board as ASCII art (B/W/. for Black/White/empty).
func (b *Board) String() string {
	out := ""
	for y := 0; y < b.Size; y++ {
		for x := 0; x < b.Size; x++ {
			s := b.Grid[x][y]
			if s == nil {
				out += ". "
			} else if s.Player == PlayerBlack {
				out += "B "
			} else {
				out += "W "
			}
		}
		out += "\n"
	}
	return out
}
