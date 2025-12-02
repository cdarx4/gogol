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

// Imports
import (
	"fmt"
)

// Represents the board
type Board struct {
	Size          int
	Grid          [][]*Stone
	Groups        []*Group
	nextGroupId   int
	currentPlayer Player
}

// Creates a new board for the game with the
// given size.
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

// Function to place stone on board
func (b *Board) PlaceStone(x, y int) bool {
	if x < 0 || x >= b.Size || y < 0 || y >= b.Size {
		return false
	}

	// Check if the spot is already occupied
	if b.Grid[x][y] != nil {
		return false
	}

	// Create the stone and place it
	stone := &Stone{
		X:      x,
		Y:      y,
		Player: b.currentPlayer,
	}

	b.Grid[x][y] = stone

	// Handle Groups
	neighbors := b.getNeighbors(x, y)
	b.handleGroups(stone, neighbors)

	// Update liberties
	b.updateAffectedLiberties(stone, x, y)

	// Switch turns
	b.switchTurn()

	return true
}

// Given a coordinate,
// returns an array of the neighbors
func (b *Board) getNeighbors(x, y int) []*Stone {
	neighbors := []*Stone{}
	directions := []struct{ dx, dy int }{{-1, 0}, {1, 0}, {0, -1}, {0, 1}}
	for _, d := range directions {
		nx, ny := x+d.dx, y+d.dy
		if nx >= 0 && nx < b.Size && ny >= 0 && ny < b.Size {
			neighbor := b.Grid[nx][ny]
			if neighbor != nil && neighbor.Player == b.currentPlayer {
				neighbors = append(neighbors, neighbor)
			}
		}
	}
	return neighbors
}

// Given a stone and its neighbors,
// merges the groups if necessary
func (b *Board) handleGroups(stone *Stone, neighbors []*Stone) {
	// Find unique groups to merge
	groupsToMerge := make(map[int]*Group)
	for _, n := range neighbors {
		for _, g := range b.Groups {
			if g.ID == n.GroupId {
				groupsToMerge[g.ID] = g
				break
			}
		}
	}

	if len(groupsToMerge) == 0 {
		// New Group
		newGroup := &Group{
			ID:        b.nextGroupId,
			Player:    b.currentPlayer,
			Stones:    []*Stone{stone},
			Liberties: 0,
		}
		b.nextGroupId++
		b.Groups = append(b.Groups, newGroup)
		stone.GroupId = newGroup.ID
	} else {
		// Merge Groups
		var targetGroup *Group
		// Pick the first one as target
		for _, g := range groupsToMerge {
			targetGroup = g
			break
		}

		targetGroup.Stones = append(targetGroup.Stones, stone)
		stone.GroupId = targetGroup.ID

		for _, g := range groupsToMerge {
			if g != targetGroup {
				// Move stones to targetGroup
				for _, s := range g.Stones {
					s.GroupId = targetGroup.ID
					targetGroup.Stones = append(targetGroup.Stones, s)
				}
			}
		}

		// Remove merged groups from b.Groups
		newGroups := []*Group{}
		for _, g := range b.Groups {
			keep := true
			for _, merged := range groupsToMerge {
				if g == merged && g != targetGroup {
					keep = false
					break
				}
			}
			if keep {
				newGroups = append(newGroups, g)
			}
		}
		b.Groups = newGroups
	}
}

// Given a group,
// updates the liberties of the group
func (b *Board) UpdateLiberties(g *Group) {
	liberties := make(map[string]bool)
	directions := []struct{ dx, dy int }{{-1, 0}, {1, 0}, {0, -1}, {0, 1}}

	for _, s := range g.Stones {
		for _, d := range directions {
			nx, ny := s.X+d.dx, s.Y+d.dy
			if nx >= 0 && nx < b.Size && ny >= 0 && ny < b.Size {
				if b.Grid[nx][ny] == nil {
					key := fmt.Sprintf("%d,%d", nx, ny)
					liberties[key] = true
				}
			}
		}
	}

	// Delete group if no liberties
	if len(liberties) == 0 {
		b.RemoveGroup(g)
	}

	g.Liberties = len(liberties)
}

// Given a stone and its coordinates,
// updates the liberties of the group
func (b *Board) updateAffectedLiberties(stone *Stone, x, y int) {
	// Update liberties for the current group
	currentGroup := b.getGroup(stone.GroupId)
	if currentGroup != nil {
		b.UpdateLiberties(currentGroup)
	}

	// Update liberties for adjacent opponent groups
	directions := []struct{ dx, dy int }{{-1, 0}, {1, 0}, {0, -1}, {0, 1}}
	for _, d := range directions {
		nx, ny := x+d.dx, y+d.dy
		if nx >= 0 && nx < b.Size && ny >= 0 && ny < b.Size {
			neighbor := b.Grid[nx][ny]
			if neighbor != nil && neighbor.Player != b.currentPlayer {
				opponentGroup := b.getGroup(neighbor.GroupId)
				if opponentGroup != nil {
					b.UpdateLiberties(opponentGroup)
				}
			}
		}
	}
}

// Switches the turn
func (b *Board) switchTurn() {
	if b.currentPlayer == PlayerBlack {
		b.currentPlayer = PlayerWhite
	} else {
		b.currentPlayer = PlayerBlack
	}
}

// Given a group id,
// returns the group
func (b *Board) getGroup(id int) *Group {
	for _, g := range b.Groups {
		if g.ID == id {
			return g
		}
	}
	return nil
}

// Removes a group from the board
func (b *Board) RemoveGroup(g *Group) {
	// Identify neighbors to update liberties later
	neighborsToUpdate := make(map[int]*Group)
	directions := []struct{ dx, dy int }{{-1, 0}, {1, 0}, {0, -1}, {0, 1}}

	for _, s := range g.Stones {
		// Check neighbors before removing stone
		for _, d := range directions {
			nx, ny := s.X+d.dx, s.Y+d.dy
			if nx >= 0 && nx < b.Size && ny >= 0 && ny < b.Size {
				neighbor := b.Grid[nx][ny]
				if neighbor != nil && neighbor.GroupId != g.ID {
					ng := b.getGroup(neighbor.GroupId)
					if ng != nil {
						neighborsToUpdate[ng.ID] = ng
					}
				}
			}
		}
		// Remove stone from grid
		b.Grid[s.X][s.Y] = nil
	}

	// Remove group from list
	newGroups := []*Group{}
	for _, group := range b.Groups {
		if group.ID != g.ID {
			newGroups = append(newGroups, group)
		}
	}
	b.Groups = newGroups

	// Update liberties of neighbors
	for _, ng := range neighborsToUpdate {
		b.UpdateLiberties(ng)
	}
}
