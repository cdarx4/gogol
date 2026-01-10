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

import "slices"

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

// isLegalPlacement checks all placement legality rules for a non-pass move.
func (b *Board) isLegalPlacement(x, y int, player Player) bool {
	if !b.isInBounds(x, y) || b.isOccupied(x, y) {
		return false
	}
	if b.isSuicide(x, y, player) {
		return false
	}
	if b.isKo(x, y, player) {
		return false
	}
	return true
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
	friendlyGroups, opponentGroups := b.splitAdjacentGroupsByPlayer(x, y, player)

	if b.moveCapturesOpponentAt(x, y, opponentGroups) {
		return false
	}

	mergedLiberties := b.computeMergedLibertiesAfterPlacement(x, y, friendlyGroups)
	return len(mergedLiberties) == 0
}

// splitAdjacentGroupsByPlayer returns adjacent groups split into friendly and opponent groups for the given player.
func (b *Board) splitAdjacentGroupsByPlayer(x, y int, player Player) (friendly, opponent []*Group) {
	for _, group := range b.adjacentGroups(x, y) {
		if group.Player == player {
			friendly = append(friendly, group)
			continue
		}
		opponent = append(opponent, group)
	}
	return friendly, opponent
}

// moveCapturesOpponentAt reports whether placing at (x, y) would capture at least one adjacent opponent group.
func (b *Board) moveCapturesOpponentAt(x, y int, opponentGroups []*Group) bool {
	placement := [2]int{x, y}

	for _, group := range opponentGroups {
		b.recomputeGroupLiberties(group)
		if len(group.Liberties) != 1 {
			continue
		}
		if _, ok := group.Liberties[placement]; ok {
			return true
		}
	}
	return false
}

// computeMergedLibertiesAfterPlacement computes the liberties the (possibly merged) friendly group would have after placement.
func (b *Board) computeMergedLibertiesAfterPlacement(x, y int, friendlyGroups []*Group) map[[2]int]struct{} {
	merged := b.directLibertiesAround(x, y)

	placement := [2]int{x, y}
	for _, group := range friendlyGroups {
		b.recomputeGroupLiberties(group)
		for liberty := range group.Liberties {
			if liberty == placement {
				continue
			}
			merged[liberty] = struct{}{}
		}
	}

	return merged
}

// directLibertiesAround returns empty adjacent points around (x, y).
func (b *Board) directLibertiesAround(x, y int) map[[2]int]struct{} {
	liberties := make(map[[2]int]struct{})

	for _, direction := range AdjacentDirections {
		nx, ny := x+direction.DeltaX, y+direction.DeltaY
		if b.isInBounds(nx, ny) && b.Grid[nx][ny] == nil {
			liberties[[2]int{nx, ny}] = struct{}{}
		}
	}

	return liberties
}

// ---------- Ko Check ----------

// isKo reports whether the move would violate the Ko rule (Positional Superko).
func (b *Board) isKo(x, y int, player Player) bool {
	nextState := b.simulateNextState(x, y, player)
	return slices.Contains(b.History, nextState)
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

// checkAndRemoveCapturedGroup checks if a group at the given position is captured and removes it.
func (b *Board) checkAndRemoveCapturedGroup(simulatedGrid SimulatedGrid, startX, startY int, opponentValue int) {
	if !b.isInBounds(startX, startY) {
		return
	}
	if simulatedGrid[startY][startX] != opponentValue {
		return
	}

	groupPositions, hasLiberties := b.simulateGroupFloodFill(simulatedGrid, startX, startY, opponentValue)
	if hasLiberties {
		return
	}

	b.clearSimulatedPositions(simulatedGrid, groupPositions)
}

// simulateGroupFloodFill returns all connected opponent positions starting from (startX, startY) and whether the group has liberties.
func (b *Board) simulateGroupFloodFill(simulatedGrid SimulatedGrid, startX, startY int, opponentValue int) (positions [][2]int, hasLiberties bool) {
	queue := [][2]int{{startX, startY}}
	visited := map[[2]int]bool{{startX, startY}: true}

	for head := 0; head < len(queue); head++ {
		pos := queue[head]
		positions = append(positions, pos)

		x, y := pos[0], pos[1]
		for _, direction := range AdjacentDirections {
			nx, ny := x+direction.DeltaX, y+direction.DeltaY
			if !b.isInBounds(nx, ny) {
				continue
			}

			cell := simulatedGrid[ny][nx]
			switch cell {
			case EmptyCellValue:
				hasLiberties = true
			case opponentValue:
				next := [2]int{nx, ny}
				if visited[next] {
					continue
				}
				visited[next] = true
				queue = append(queue, next)
			}
		}
	}

	return positions, hasLiberties
}

// simulatedGridToString converts the simulated grid to a string representation
func (b *Board) simulatedGridToString(simulatedGrid SimulatedGrid) string {
	boardString := ""

	// Convert each cell to a character
	for y := 0; y < b.Size; y++ {
		for x := 0; x < b.Size; x++ {
			cellValue := simulatedGrid[y][x]
			switch cellValue {
			case EmptyCellValue:
				boardString += ". "
			case BlackCellValue:
				boardString += "B "
			default:
				boardString += "W "
			}
		}
		boardString += "\n"
	}
	return boardString
}

// clearSimulatedPositions clears all given positions in the simulated grid.
func (b *Board) clearSimulatedPositions(simulatedGrid SimulatedGrid, positions [][2]int) {
	for _, pos := range positions {
		simulatedGrid[pos[1]][pos[0]] = EmptyCellValue
	}
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
	b.PassCount = 0

	if !b.isLegalPlacement(x, y, b.CurrentPlayer) {
		return false
	}

	newGroup := b.placeStoneAndCreateGroup(x, y, b.CurrentPlayer)
	b.mergeNewGroupWithAdjacentFriendlies(x, y, newGroup, b.CurrentPlayer)

	b.recomputeAllGroupLiberties()
	b.captureDeadOpponentGroups(b.CurrentPlayer)
	b.recomputeAllGroupLiberties()

	b.History = append(b.History, b.String())
	b.switchTurn()
	return true
}

// placeStoneAndCreateGroup places the stone on the board and creates its initial group.
func (b *Board) placeStoneAndCreateGroup(x, y int, player Player) *Group {
	stone := &Stone{X: x, Y: y, Player: player}
	b.Grid[x][y] = stone
	return b.createGroup(stone)
}

// mergeNewGroupWithAdjacentFriendlies merges the new group with any adjacent friendly groups.
func (b *Board) mergeNewGroupWithAdjacentFriendlies(x, y int, newGroup *Group, player Player) {
	var toMerge []*Group
	for _, group := range b.adjacentGroups(x, y) {
		if group.Player == player {
			toMerge = append(toMerge, group)
		}
	}
	if len(toMerge) == 0 {
		return
	}
	b.mergeGroups(newGroup, toMerge)
}

// captureDeadOpponentGroups captures all opponent groups that have no liberties.
func (b *Board) captureDeadOpponentGroups(current Player) {
	opponents := b.collectGroupsToCapture(current)
	for _, group := range opponents {
		b.captureGroup(group)
	}
}

// collectGroupsToCapture returns opponent groups that currently have no liberties.
func (b *Board) collectGroupsToCapture(current Player) []*Group {
	var toCapture []*Group
	for _, group := range b.Groups {
		if group.Player == current {
			continue
		}
		if len(group.Liberties) == 0 {
			toCapture = append(toCapture, group)
		}
	}
	return toCapture
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

// Clone creates a deep copy of the board.
func (b *Board) Clone() *Board {
	cloned := b.cloneBase()

	groupMap := make(map[int]*Group)
	b.cloneGridAndGroupsInto(cloned, groupMap)

	cloned.recomputeAllGroupLiberties()
	return cloned
}

// cloneBase allocates the clone with basic fields and an empty grid.
func (b *Board) cloneBase() *Board {
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

	for i := range cloned.Grid {
		cloned.Grid[i] = make([]*Stone, b.Size)
	}

	return cloned
}

// cloneGridAndGroupsInto clones stones into the cloned board and reconstructs group memberships.
func (b *Board) cloneGridAndGroupsInto(cloned *Board, groupMap map[int]*Group) {
	for x := 0; x < b.Size; x++ {
		for y := 0; y < b.Size; y++ {
			oldStone := b.Grid[x][y]
			if oldStone == nil {
				continue
			}

			newStone := &Stone{
				X:      oldStone.X,
				Y:      oldStone.Y,
				Player: oldStone.Player,
			}
			cloned.Grid[x][y] = newStone

			if oldStone.Group == nil {
				continue
			}

			newGroup := ensureClonedGroup(cloned, groupMap, oldStone.Group)
			newGroup.Stones = append(newGroup.Stones, newStone)
			newStone.Group = newGroup
		}
	}
}

// ensureClonedGroup returns the cloned group corresponding to oldGroup, creating it if needed.
func ensureClonedGroup(cloned *Board, groupMap map[int]*Group, oldGroup *Group) *Group {
	if g, ok := groupMap[oldGroup.ID]; ok {
		return g
	}

	g := &Group{
		ID:        oldGroup.ID,
		Player:    oldGroup.Player,
		Stones:    []*Stone{},
		Liberties: make(map[[2]int]struct{}),
	}
	groupMap[oldGroup.ID] = g
	cloned.Groups = append(cloned.Groups, g)
	return g
}

// ---------- Legal moves ----------

// GetLegalMoves returns all legal moves for the given player.
// Includes pass move represented as (-1, -1).
func (b *Board) GetLegalMoves(player Player) [][2]int {
	moves := make([][2]int, 0, 1+(b.Size*b.Size))
	moves = append(moves, [2]int{PassMoveCoordinate, PassMoveCoordinate})

	for x := 0; x < b.Size; x++ {
		for y := 0; y < b.Size; y++ {
			if b.isLegalPlacement(x, y, player) {
				moves = append(moves, [2]int{x, y})
			}
		}
	}

	return moves
}

// CloneAndPlay clones the board, applies the move, and returns the new board.
// Returns (cloned board, success).
// Does not modify the original board.
func (b *Board) CloneAndPlay(x, y int, player Player) (*Board, bool) {
	cloned := b.Clone()

	if isPassMove(x, y) {
		cloned.Pass()
		return cloned, true
	}

	if !cloned.isInBounds(x, y) || cloned.isOccupied(x, y) {
		return cloned, false
	}
	if cloned.isSuicide(x, y, player) {
		return cloned, false
	}

	return cloned.applyMoveAsPlayer(x, y, player)
}

// isPassMove reports whether (x, y) represents a pass move.
func isPassMove(x, y int) bool {
	return x == PassMoveCoordinate && y == PassMoveCoordinate
}

// applyMoveAsPlayer places a move while temporarily setting CurrentPlayer to the provided player.
func (b *Board) applyMoveAsPlayer(x, y int, player Player) (*Board, bool) {
	originalPlayer := b.CurrentPlayer
	b.CurrentPlayer = player

	success := b.PlaceStone(x, y)
	if !success {
		b.CurrentPlayer = originalPlayer
	}

	return b, success
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
