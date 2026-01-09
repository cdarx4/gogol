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

// Constants for simulated grid cell values used in Ko checking.
const (
	// EmptyCellValue represents an empty cell in the simulated grid.
	EmptyCellValue = 0
	// BlackCellValue represents a black stone in the simulated grid.
	BlackCellValue = 1
	// WhiteCellValue represents a white stone in the simulated grid.
	WhiteCellValue = 2

	// PassMoveCoordinate represents the coordinate used for a pass move.
	PassMoveCoordinate = -1
)

// SimulatedGrid represents a simplified board state for Ko checking
type SimulatedGrid [][]int

// ---------- Board creation ----------

// NewBoard creates and returns a new empty board of the given size.
func NewBoard(size int) *Board {
	grid := make([][]*Stone, size)
	for i := range grid {
		grid[i] = make([]*Stone, size)
	}

	b := &Board{
		Size:          size,
		Grid:          grid,
		CurrentPlayer: StartingPlayer,
		History:       []string{},
	}
	b.History = append(b.History, b.String())
	return b
}

// ---------- Helpers ----------

// isInBounds reports whether the position (x, y) is within the board boundaries.
func (b *Board) isInBounds(x, y int) bool {
	return x >= 0 && x < b.Size && y >= 0 && y < b.Size
}

// isOccupied reports whether the position (x, y) is occupied by a stone.
func (b *Board) isOccupied(x, y int) bool {
	return b.Grid[x][y] != nil
}

// ---------- Group & liberty recomputation ----------

// recomputeGroupLiberties recomputes the liberties of a group by checking all empty adjacent points for every stone in the group.
func (b *Board) recomputeGroupLiberties(group *Group) {
	group.Liberties = make(map[[2]int]struct{})

	for _, stone := range group.Stones {
		for _, direction := range AdjacentDirections {
			neighborX, neighborY := stone.X+direction.DeltaX, stone.Y+direction.DeltaY
			if b.isInBounds(neighborX, neighborY) && b.Grid[neighborX][neighborY] == nil {
				group.Liberties[[2]int{neighborX, neighborY}] = struct{}{}
			}
		}
	}
}

// recomputeAllGroupLiberties recomputes liberties for all groups on the board.
func (b *Board) recomputeAllGroupLiberties() {
	for _, group := range b.Groups {
		b.recomputeGroupLiberties(group)
	}
}

// ---------- Group creation / merging ----------

// createGroup creates a new group for the given stone and adds it to the board's group list.
func (b *Board) createGroup(stone *Stone) *Group {
	group := &Group{
		ID:     b.nextGroupID,
		Player: stone.Player,
		Stones: []*Stone{stone},
	}
	b.nextGroupID++
	b.Groups = append(b.Groups, group)
	stone.Group = group
	return group
}

// mergeGroups merges other groups into the target group. It is used when placing a stone connects separate friendly groups.
func (b *Board) mergeGroups(target *Group, others []*Group) {
	for _, groupToMerge := range others {
		for _, stone := range groupToMerge.Stones {
			stone.Group = target
			target.Stones = append(target.Stones, stone)
		}
	}
	b.removeGroups(others)
}

// removeGroups removes the listed groups from the board's group list.
func (b *Board) removeGroups(groups []*Group) {
	filtered := []*Group{}
	for _, group := range b.Groups {
		shouldKeep := true
		for _, groupToRemove := range groups {
			if group.ID == groupToRemove.ID {
				shouldKeep = false
				break
			}
		}
		if shouldKeep {
			filtered = append(filtered, group)
		}
	}
	b.Groups = filtered
}

// ---------- Capture logic ----------

// captureGroup removes all stones in the group from the board and updates the group list.
func (b *Board) captureGroup(group *Group) {
	for _, stone := range group.Stones {
		b.Grid[stone.X][stone.Y] = nil
	}
	b.removeGroups([]*Group{group})
}

// ---------- Suicide check (simulation) ----------

// isSuicide reports whether placing a stone of the given player at (x, y) would be a suicide (illegal self-capture).
// It returns true if the move is suicide and thus illegal.
func (b *Board) isSuicide(x, y int, player Player) bool {
	// Find all groups adjacent to where we want to place the stone
	adjacentGroups := b.adjacentGroups(x, y)
	friendlyGroups := []*Group{}
	opponentGroups := []*Group{}

	// Split them into friendly and opponent groups
	for _, group := range adjacentGroups {
		if group.Player == player {
			friendlyGroups = append(friendlyGroups, group)
		} else {
			opponentGroups = append(opponentGroups, group)
		}
	}

	// If placing here would capture an opponent group, it's definitely not suicide
	for _, opponentGroup := range opponentGroups {
		b.recomputeGroupLiberties(opponentGroup)
		// If the opponent group only has one liberty and it's exactly where we're placing, we capture it
		if len(opponentGroup.Liberties) == 1 {
			if _, ok := opponentGroup.Liberties[[2]int{x, y}]; ok {
				return false
			}
		}
	}

	// Calculate what liberties the merged group would have after placing the stone
	mergedGroupLiberties := make(map[[2]int]struct{})

	// First, add the direct liberties around the placement position
	for _, direction := range AdjacentDirections {
		neighborX, neighborY := x+direction.DeltaX, y+direction.DeltaY
		if b.isInBounds(neighborX, neighborY) && b.Grid[neighborX][neighborY] == nil {
			mergedGroupLiberties[[2]int{neighborX, neighborY}] = struct{}{}
		}
	}

	// Then add liberties from any adjacent friendly groups (but not the placement position itself)
	for _, friendlyGroup := range friendlyGroups {
		b.recomputeGroupLiberties(friendlyGroup)
		for libertyPosition := range friendlyGroup.Liberties {
			if libertyPosition != ([2]int{x, y}) {
				mergedGroupLiberties[libertyPosition] = struct{}{}
			}
		}
	}

	// If the merged group has no liberties, this move is suicide
	return len(mergedGroupLiberties) == 0
}

// ---------- Ko Check ----------

// isKo reports whether the move would violate the Ko rule (Positional Superko).
func (b *Board) isKo(x, y int, player Player) bool {
	nextState := b.simulateNextState(x, y, player)
	for _, previousState := range b.History {
		if previousState == nextState {
			return true
		}
	}
	return false
}

// simulateNextState simulates the next board state string after a move.
func (b *Board) simulateNextState(x, y int, player Player) string {
	simulatedGrid := b.createSimulatedGrid()
	b.placeStoneInSimulation(simulatedGrid, x, y, player)
	b.captureOpponentGroupsInSimulation(simulatedGrid, x, y, player)
	return b.simulatedGridToString(simulatedGrid)
}

// createSimulatedGrid creates a simplified copy of the grid for simulation
func (b *Board) createSimulatedGrid() SimulatedGrid {
	// Create a grid of the same size
	simulatedGrid := make([][]int, b.Size)
	for y := 0; y < b.Size; y++ {
		simulatedGrid[y] = make([]int, b.Size)

		// Convert each stone to a simple integer value
		for x := 0; x < b.Size; x++ {
			if b.Grid[x][y] != nil {
				if b.Grid[x][y].Player == PlayerBlack {
					simulatedGrid[y][x] = BlackCellValue
				} else {
					simulatedGrid[y][x] = WhiteCellValue
				}
			}
		}
	}
	return simulatedGrid
}

// placeStoneInSimulation places a stone in the simulated grid
func (b *Board) placeStoneInSimulation(simulatedGrid SimulatedGrid, x, y int, player Player) {
	playerValue := BlackCellValue
	if player == PlayerWhite {
		playerValue = WhiteCellValue
	}
	simulatedGrid[y][x] = playerValue
}

// captureOpponentGroupsInSimulation checks and removes captured opponent groups
func (b *Board) captureOpponentGroupsInSimulation(simulatedGrid SimulatedGrid, x, y int, player Player) {
	// Figure out what value represents the opponent in the simulated grid
	opponentValue := BlackCellValue
	if player == PlayerBlack {
		opponentValue = WhiteCellValue
	}

	// Check all four neighbors to see if any opponent groups got captured
	for _, direction := range AdjacentDirections {
		xDelta := direction.DeltaX
		yDelta := direction.DeltaY
		neighborX := x + xDelta
		neighborY := y + yDelta
		b.checkAndRemoveCapturedGroup(simulatedGrid, neighborX, neighborY, opponentValue)
	}
}

// checkAndRemoveCapturedGroup checks if a group at the given position is captured and removes it
func (b *Board) checkAndRemoveCapturedGroup(simulatedGrid SimulatedGrid, startX, startY int, opponentValue int) {
	// Make sure the position is within bounds
	if startX < 0 || startX >= b.Size || startY < 0 || startY >= b.Size {
		return
	}

	// Make sure there's actually an opponent stone at this position
	if simulatedGrid[startY][startX] != opponentValue {
		return
	}

	// Use flood fill to find all connected opponent stones in this group
	groupPositions := [][2]int{}
	queue := [][2]int{{startX, startY}}
	visited := map[[2]int]bool{{startX, startY}: true}
	hasLiberties := false
	queueHead := 0

	// Explore all connected stones using breadth-first search
	for queueHead < len(queue) {
		currentPosition := queue[queueHead]
		queueHead++
		groupPositions = append(groupPositions, currentPosition)
		currentX, currentY := currentPosition[0], currentPosition[1]

		// Check all four adjacent neighbors
		for _, direction := range AdjacentDirections {
			xDelta := direction.DeltaX
			yDelta := direction.DeltaY
			neighborX := currentX + xDelta
			neighborY := currentY + yDelta

			// Skip if neighbor is outside the board
			if neighborX < 0 || neighborX >= b.Size || neighborY < 0 || neighborY >= b.Size {
				continue
			}

			cellValue := simulatedGrid[neighborY][neighborX]

			// If we find an empty cell, the group has at least one liberty
			if cellValue == EmptyCellValue {
				hasLiberties = true
			} else if cellValue == opponentValue {
				// Found another opponent stone, add it to the queue if we haven't seen it yet
				if !visited[[2]int{neighborX, neighborY}] {
					visited[[2]int{neighborX, neighborY}] = true
					queue = append(queue, [2]int{neighborX, neighborY})
				}
			}
		}
	}

	// If the group has no liberties, remove all its stones
	if !hasLiberties {
		for _, position := range groupPositions {
			simulatedGrid[position[1]][position[0]] = EmptyCellValue
		}
	}
}

// simulatedGridToString converts the simulated grid to a string representation
func (b *Board) simulatedGridToString(simulatedGrid SimulatedGrid) string {
	boardString := ""

	// Convert each cell to a character
	for y := 0; y < b.Size; y++ {
		for x := 0; x < b.Size; x++ {
			cellValue := simulatedGrid[y][x]
			if cellValue == EmptyCellValue {
				boardString += ". "
			} else if cellValue == BlackCellValue {
				boardString += "B "
			} else {
				boardString += "W "
			}
		}
		boardString += "\n"
	}
	return boardString
}

// ---------- Adjacent groups ----------

// adjacentGroups returns a slice of all unique groups directly adjacent to (x, y).
func (b *Board) adjacentGroups(x, y int) []*Group {
	seenGroupIDs := map[int]struct{}{}
	var result []*Group

	for _, direction := range AdjacentDirections {
		neighborX, neighborY := x+direction.DeltaX, y+direction.DeltaY
		if !b.isInBounds(neighborX, neighborY) {
			continue
		}
		stone := b.Grid[neighborX][neighborY]
		if stone != nil && stone.Group != nil {
			if _, ok := seenGroupIDs[stone.Group.ID]; !ok {
				seenGroupIDs[stone.Group.ID] = struct{}{}
				result = append(result, stone.Group)
			}
		}
	}
	return result
}

// ---------- Place stone (main rule engine) ----------

// PlaceStone places a stone for the current player at (x, y) if legal, enacts all necessary group/capture logic, and advances the turn.
// It returns true if the move was successful.
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

	// 3. Illegal move if it violates Ko rule (reps state).
	if b.isKo(x, y, b.CurrentPlayer) {
		return false
	}

	// 4. Place the new stone and create its group.
	stone := &Stone{X: x, Y: y, Player: b.CurrentPlayer}
	b.Grid[x][y] = stone
	newGroup := b.createGroup(stone)

	// 5. Merge with any adjacent friendly groups created by this placement.
	adjacentGroups := b.adjacentGroups(x, y)
	groupsToMerge := []*Group{}

	// Find all adjacent groups that belong to the current player
	for _, group := range adjacentGroups {
		if group.Player == b.CurrentPlayer {
			groupsToMerge = append(groupsToMerge, group)
		}
	}

	// Merge them all together if there are any
	if len(groupsToMerge) > 0 {
		b.mergeGroups(newGroup, groupsToMerge)
	}

	// 6. Recompute liberties after merge and before captures.
	b.recomputeAllGroupLiberties()

	// 7. Capture any adjacent opponent groups that have lost all liberties.
	for _, opponentGroup := range b.Groups {
		if opponentGroup.Player != b.CurrentPlayer && len(opponentGroup.Liberties) == 0 {
			b.captureGroup(opponentGroup)
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

// switchTurn switches the current player to their opponent.
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
	// Go through the whole board and count stones
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

// GetWinner returns the winner based on stone count.
// It returns nil if the game is a draw.
func (b *Board) GetWinner() (winner *Player, isDraw bool) {
	blackCount, whiteCount := b.GetStoneCount()

	// Whoever has more stones wins
	if blackCount > whiteCount {
		winnerPlayer := PlayerBlack
		return &winnerPlayer, false
	} else if whiteCount > blackCount {
		winnerPlayer := PlayerWhite
		return &winnerPlayer, false
	}

	// If counts are equal, it's a draw
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

// ---------- Clone ----------

// Clone creates a deep copy of the board
func (b *Board) Clone() *Board {
	// Copy all the basic fields
	cloned := &Board{
		Size:          b.Size,
		Grid:          make([][]*Stone, b.Size),
		Groups:        make([]*Group, 0, len(b.Groups)),
		nextGroupID:   b.nextGroupID,
		CurrentPlayer: b.CurrentPlayer,
		PassCount:     b.PassCount,
		History:       make([]string, len(b.History)),
	}
	copy(cloned.History, b.History)

	// Allocate the grid rows
	for i := range cloned.Grid {
		cloned.Grid[i] = make([]*Stone, b.Size)
	}

	// Keep track of which old groups map to which new groups
	groupMap := make(map[int]*Group)

	// Go through every position and clone any stones we find
	for x := 0; x < b.Size; x++ {
		for y := 0; y < b.Size; y++ {
			if b.Grid[x][y] != nil {
				oldStone := b.Grid[x][y]

				// Create a new stone with the same properties
				newStone := &Stone{
					X:      oldStone.X,
					Y:      oldStone.Y,
					Player: oldStone.Player,
				}
				cloned.Grid[x][y] = newStone

				// If the stone belongs to a group, we need to recreate that group structure
				if oldStone.Group != nil {
					// Check if we've already created this group
					if newGroup, exists := groupMap[oldStone.Group.ID]; !exists {
						// Create a new group and remember it
						newGroup = &Group{
							ID:        oldStone.Group.ID,
							Player:    oldStone.Group.Player,
							Stones:    []*Stone{newStone},
							Liberties: make(map[[2]int]struct{}),
						}
						groupMap[oldStone.Group.ID] = newGroup
						cloned.Groups = append(cloned.Groups, newGroup)
						newStone.Group = newGroup
					} else {
						// This group already exists, just add the stone to it
						newGroup.Stones = append(newGroup.Stones, newStone)
						newStone.Group = newGroup
					}
				}
			}
		}
	}

	// Recompute all the liberties since we've rebuilt the groups
	cloned.recomputeAllGroupLiberties()

	return cloned
}

// ---------- Legal moves ----------

// GetLegalMoves returns all legal moves for the given player
// Includes pass move represented as (-1, -1)
func (b *Board) GetLegalMoves(player Player) [][2]int {
	moves := make([][2]int, 0)

	// Passing is always a legal move
	moves = append(moves, [2]int{PassMoveCoordinate, PassMoveCoordinate})

	// Check every position on the board
	for x := 0; x < b.Size; x++ {
		for y := 0; y < b.Size; y++ {
			// Position needs to be empty
			if !b.isOccupied(x, y) {
				// And it can't be suicide
				if !b.isSuicide(x, y, player) {
					moves = append(moves, [2]int{x, y})
				}
			}
		}
	}

	return moves
}

// CloneAndPlay clones the board, applies the move, and returns the new board
// Returns (cloned board, success)
// Does not modify the original board
func (b *Board) CloneAndPlay(x, y int, player Player) (*Board, bool) {
	cloned := b.Clone()

	// Handle pass moves
	if x == PassMoveCoordinate && y == PassMoveCoordinate {
		cloned.Pass()
		return cloned, true
	}

	// Make sure the move is valid
	if !cloned.isInBounds(x, y) || cloned.isOccupied(x, y) {
		return cloned, false
	}

	// Can't be suicide
	if cloned.isSuicide(x, y, player) {
		return cloned, false
	}

	// Set the current player so we can apply the move
	originalPlayer := cloned.CurrentPlayer
	cloned.CurrentPlayer = player

	// Try to place the stone
	success := cloned.PlaceStone(x, y)

	// If it failed, restore the original player
	if !success {
		cloned.CurrentPlayer = originalPlayer
	}

	return cloned, success
}

// ---------- Debug ----------

// String returns a string representing the board as ASCII art (B/W/. for Black/White/empty).
func (b *Board) String() string {
	boardString := ""
	for y := 0; y < b.Size; y++ {
		for x := 0; x < b.Size; x++ {
			stone := b.Grid[x][y]
			if stone == nil {
				boardString += ". "
			} else if stone.Player == PlayerBlack {
				boardString += "B "
			} else {
				boardString += "W "
			}
		}
		boardString += "\n"
	}
	return boardString
}
