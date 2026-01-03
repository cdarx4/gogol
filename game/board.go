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
func NewBoard(size int) *Board {
	grid := make([][]*Stone, size)
	for i := range grid {
		grid[i] = make([]*Stone, size)
	}

	b := &Board{
		Size:          size,
		Grid:          grid,
		currentPlayer: StartingPlayer,
		History:       []string{},
	}
	b.History = append(b.History, b.String())
	return b
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
	group.Liberties = make(map[[2]int]struct{})

	for _, s := range group.Stones {
		for _, d := range AdjacentDirections {
			nx, ny := s.X+d.dx, s.Y+d.dy
			if b.isInBounds(nx, ny) && b.Grid[nx][ny] == nil {
				group.Liberties[[2]int{nx, ny}] = struct{}{}
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
	for _, g := range opponent {
		b.recomputeGroupLiberties(g)
		if len(g.Liberties) == 1 {
			if _, ok := g.Liberties[[2]int{x, y}]; ok {
				return false
			}
		}
	}

	// Simulate merged group liberties after this move.
	liberties := make(map[[2]int]struct{})
	for _, d := range AdjacentDirections {
		nx, ny := x+d.dx, y+d.dy
		if b.isInBounds(nx, ny) && b.Grid[nx][ny] == nil {
			liberties[[2]int{nx, ny}] = struct{}{}
		}
	}

	// Add liberties from adjacent friendly groups (excluding the spot itself)
	for _, g := range friendly {
		b.recomputeGroupLiberties(g)
		for l := range g.Liberties {
			if l != ([2]int{x, y}) {
				liberties[l] = struct{}{}
			}
		}
	}

	// If no liberties remain, it's suicide.
	return len(liberties) == 0
}

// ---------- Ko Check ----------

// Checks if the move would violate the Ko rule (Positional Superko).
func (b *Board) isKo(x, y int, player Player) bool {
	nextState := b.simulateNextState(x, y, player)
	for _, state := range b.History {
		if state == nextState {
			return true
		}
	}
	return false
}

// Simulates the next board state string after a move.
func (b *Board) simulateNextState(x, y int, player Player) string {
	// 1. Create a simplified copy of the grid (0=empty, 1=black, 2=white)
	sim := make([][]int, b.Size)
	for i := 0; i < b.Size; i++ {
		sim[i] = make([]int, b.Size)
		for j := 0; j < b.Size; j++ {
			if b.Grid[i][j] != nil {
				if b.Grid[i][j].Player == PlayerBlack {
					sim[i][j] = 1
				} else {
					sim[i][j] = 2
				}
			}
		}
	}

	// 2. Place the stone
	myVal := 1
	if player == PlayerWhite {
		myVal = 2
	}
	sim[x][y] = myVal

	// 3. Check for captures of opponents
	oppVal := 1
	if myVal == 1 {
		oppVal = 2
	}

	dirs := [][2]int{{-1, 0}, {1, 0}, {0, -1}, {0, 1}}

	// Helper: recursive floodfill to check liberties and remove if captured
	// Returns true if group was captured (removed)
	var checkAndRemove func(sx, sy int)
	checkAndRemove = func(sx, sy int) {
		if sx < 0 || sx >= b.Size || sy < 0 || sy >= b.Size {
			return
		}
		if sim[sx][sy] != oppVal {
			return
		}

		// Find group and check liberties
		group := [][2]int{}
		q := [][2]int{{sx, sy}}
		visited := map[[2]int]bool{{sx, sy}: true}
		hasLiberties := false

		head := 0
		for head < len(q) {
			cur := q[head]
			head++
			group = append(group, cur)
			cx, cy := cur[0], cur[1]

			for _, d := range dirs {
				nx, ny := cx+d[0], cy+d[1]
				if nx < 0 || nx >= b.Size || ny < 0 || ny >= b.Size {
					continue
				}
				val := sim[nx][ny]
				if val == 0 {
					hasLiberties = true
				} else if val == oppVal {
					if !visited[[2]int{nx, ny}] {
						visited[[2]int{nx, ny}] = true
						q = append(q, [2]int{nx, ny})
					}
				}
			}
		}

		if !hasLiberties {
			// Remove group
			for _, s := range group {
				sim[s[0]][s[1]] = 0
			}
		}
	}

	// Check all neighbors
	for _, d := range dirs {
		checkAndRemove(x+d[0], y+d[1])
	}

	// 4. Generate string representation
	out := ""
	for y := 0; y < b.Size; y++ {
		for x := 0; x < b.Size; x++ {
			val := sim[x][y]
			if val == 0 {
				out += ". "
			} else if val == 1 {
				out += "B "
			} else {
				out += "W "
			}
		}
		out += "\n"
	}
	return out
}

// ---------- Adjacent groups ----------

// Returns a slice of all unique groups directly adjacent to (x, y).
func (b *Board) adjacentGroups(x, y int) []*Group {
	seen := map[int]struct{}{}
	var result []*Group

	for _, d := range AdjacentDirections {
		nx, ny := x+d.dx, y+d.dy
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
	if b.isSuicide(x, y, b.currentPlayer) {
		return false
	}

	// 3. Illegal move if it violates Ko rule (reps state).
	if b.isKo(x, y, b.currentPlayer) {
		return false
	}

	// 3. Place the new stone and create its group.
	stone := &Stone{X: x, Y: y, Player: b.currentPlayer}
	b.Grid[x][y] = stone
	newGroup := b.createGroup(stone)

	// 4. Merge with any adjacent friendly groups created by this placement.
	adj := b.adjacentGroups(x, y)
	toMerge := []*Group{}

	for _, g := range adj {
		if g.Player == b.currentPlayer {
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
		if g.Player != b.currentPlayer && len(g.Liberties) == 0 {
			b.captureGroup(g)
		}
	}

	// 8. Recompute liberties again after any captures.
	b.recomputeAllGroupLiberties()

	// 9. Register new board state
	b.History = append(b.History, b.String())

	// 10. Switch to the next player's turn.
	b.switchTurn()
	return true
}

// ---------- Turn ----------

// Switches current player to their opponent.
func (b *Board) switchTurn() {
	b.currentPlayer = b.currentPlayer.Opponent()
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

// IsFull checks if the board is completely full (no empty spots)
func (b *Board) IsFull() bool {
	for x := 0; x < b.Size; x++ {
		for y := 0; y < b.Size; y++ {
			if b.Grid[x][y] == nil {
				return false
			}
		}
	}
	return true
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
