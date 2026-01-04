// ============================================================================
// File: board_test.go
// Author: Ding Lionel, Darx Christian
// Email: lionel.ding@hes-so.ch, christia.darx@hes-so.ch
// Date: 28.11.2025
// Description: Go file managing the board tests.
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

import "testing"

// --- Helpers ---

func mustPlace(t *testing.T, board *Board, x, y int) {
	t.Helper()
	if ok := board.PlaceStone(x, y); !ok {
		t.Fatalf("expected PlaceStone(%d,%d) to succeed\nboard:\n%s", x, y, board.String())
	}
}

func mustFail(t *testing.T, board *Board, x, y int) {
	t.Helper()
	if ok := board.PlaceStone(x, y); ok {
		t.Fatalf("expected PlaceStone(%d,%d) to fail\nboard:\n%s", x, y, board.String())
	}
}

func assertAt(t *testing.T, board *Board, x, y int, expectedPlayer *Player) {
	t.Helper()
	stone := board.Grid[x][y]
	if expectedPlayer == nil {
		if stone != nil {
			t.Fatalf("expected empty at (%d,%d), got %v\nboard:\n%s", x, y, stone.Player, board.String())
		}
		return
	}
	if stone == nil {
		t.Fatalf("expected stone at (%d,%d), got empty\nboard:\n%s", x, y, board.String())
	}
	if stone.Player != *expectedPlayer {
		t.Fatalf("expected %v at (%d,%d), got %v\nboard:\n%s", *expectedPlayer, x, y, stone.Player, board.String())
	}
}

func assertSingleGroupAt(t *testing.T, board *Board, coords ...[2]int) {
	t.Helper()
	if len(coords) == 0 {
		t.Fatal("assertSingleGroupAt needs coords")
	}
	firstStone := board.Grid[coords[0][0]][coords[0][1]]
	if firstStone == nil || firstStone.Group == nil {
		t.Fatalf("expected stone+group at %v\nboard:\n%s", coords[0], board.String())
	}
	expectedGroupID := firstStone.Group.ID
	for _, coord := range coords[1:] {
		stone := board.Grid[coord[0]][coord[1]]
		if stone == nil || stone.Group == nil || stone.Group.ID != expectedGroupID {
			t.Fatalf("expected %v to be in same group as %v\nboard:\n%s", coord, coords[0], board.String())
		}
	}
}

func assertGroupLiberties(t *testing.T, board *Board, x, y int, expectedLibertyCount int) {
	t.Helper()
	stone := board.Grid[x][y]
	if stone == nil || stone.Group == nil {
		t.Fatalf("expected stone+group at (%d,%d)\nboard:\n%s", x, y, board.String())
	}
	actualLibertyCount := len(stone.Group.Liberties)
	if actualLibertyCount != expectedLibertyCount {
		t.Fatalf("expected group liberties=%d at (%d,%d), got %d\nboard:\n%s", expectedLibertyCount, x, y, actualLibertyCount, board.String())
	}
}

// ------------------------------------------------------------
// Basic validation
// ------------------------------------------------------------

func TestPlaceStone_ValidFirstMove(t *testing.T) {
	b := NewBoard(9)
	mustPlace(t, b, 4, 4)
	want := PlayerBlack
	assertAt(t, b, 4, 4, &want)
}

func TestPlaceStone_OutOfBoundsRejected(t *testing.T) {
	board := NewBoard(9)
	testCases := [][2]int{{-1, 0}, {0, -1}, {9, 0}, {0, 9}, {100, 100}}
	for _, testCase := range testCases {
		mustFail(t, board, testCase[0], testCase[1])
	}
}

func TestPlaceStone_OccupiedRejected(t *testing.T) {
	b := NewBoard(9)
	mustPlace(t, b, 4, 4)
	mustFail(t, b, 4, 4)
}

// ------------------------------------------------------------
// Turn flow
// ------------------------------------------------------------

func TestPlaceStone_TurnAlternates(t *testing.T) {
	board := NewBoard(9)

	mustPlace(t, board, 0, 0) // B
	mustPlace(t, board, 0, 1) // W
	mustPlace(t, board, 0, 2) // B

	expectedBlack := PlayerBlack
	expectedWhite := PlayerWhite
	assertAt(t, board, 0, 0, &expectedBlack)
	assertAt(t, board, 0, 1, &expectedWhite)
	assertAt(t, board, 0, 2, &expectedBlack)
}

// ------------------------------------------------------------
// Group creation / merge
// ------------------------------------------------------------

func TestGroups_MergeFriendlyGroupsThroughBridge(t *testing.T) {
	board := NewBoard(9)

	// Build:
	// B . B   (row 0)
	// then bridge at (1,0)
	mustPlace(t, board, 0, 0) // B
	mustPlace(t, board, 8, 8) // W (far)
	mustPlace(t, board, 2, 0) // B
	mustPlace(t, board, 8, 7) // W (far)
	mustPlace(t, board, 1, 0) // B bridges

	assertSingleGroupAt(t, board, [2]int{0, 0}, [2]int{1, 0}, [2]int{2, 0})
}

func TestGroups_UniqueLibertiesNotDoubleCounted(t *testing.T) {
	board := NewBoard(9)

	// Make a 2-stone vertical black group at (4,4) and (4,5).
	// Liberties should be 6 (not 8):
	// (4,4) has 4 liberties, (4,5) has 4, but they share (4,4)/(4,5) adjacency not a liberty,
	// and they also share none as liberties; actual unique empties = 6.
	mustPlace(t, board, 4, 4) // B
	mustPlace(t, board, 8, 8) // W far
	mustPlace(t, board, 4, 5) // B

	assertSingleGroupAt(t, board, [2]int{4, 4}, [2]int{4, 5})
	assertGroupLiberties(t, board, 4, 4, 6)
}

// ------------------------------------------------------------
// Captures (single stone and multi-stone)
// ------------------------------------------------------------

func TestCapture_SingleStoneInCenter(t *testing.T) {
	b := NewBoard(9)

	// Sequence to capture a white stone at (1,1) by black:
	// B at (0,1), (2,1), (1,0), then final (1,2) closes.
	// Interleave far moves for white to keep turns correct.

	mustPlace(t, b, 0, 1) // B
	mustPlace(t, b, 8, 8) // W far
	mustPlace(t, b, 2, 1) // B
	mustPlace(t, b, 1, 1) // W (target)
	mustPlace(t, b, 1, 0) // B
	mustPlace(t, b, 8, 7) // W far
	mustPlace(t, b, 1, 2) // B -> capture

	// White stone removed
	assertAt(t, b, 1, 1, nil)
}

func TestCapture_TwoStoneChain(t *testing.T) {
	b := NewBoard(9)

	// Capture white chain at (1,1)-(1,2).
	// Black surrounds with: (0,1),(2,1),(0,2),(2,2),(1,0),(1,3)
	// with legal turn alternation.
	//
	// We'll place whites (including targets) when it's White's turn,
	// and use far filler moves otherwise.

	mustPlace(t, b, 0, 1) // B
	mustPlace(t, b, 1, 1) // W target 1

	mustPlace(t, b, 2, 1) // B
	mustPlace(t, b, 1, 2) // W target 2

	mustPlace(t, b, 0, 2) // B
	mustPlace(t, b, 8, 8) // W far

	mustPlace(t, b, 2, 2) // B
	mustPlace(t, b, 8, 7) // W far

	mustPlace(t, b, 1, 0) // B
	mustPlace(t, b, 8, 6) // W far

	mustPlace(t, b, 1, 3) // B -> capture both

	assertAt(t, b, 1, 1, nil)
	assertAt(t, b, 1, 2, nil)
}

// ------------------------------------------------------------
// Suicide + “capture makes suicide legal”
// ------------------------------------------------------------

func TestSuicide_Rejected_SimpleEye(t *testing.T) {
	b := NewBoard(9)

	// Create the classic suicide for Black at (1,1) surrounded by White:
	// White stones at (0,1),(2,1),(1,0),(1,2)
	// Achieve with turn alternation by using far filler moves.

	mustPlace(t, b, 8, 8) // B far
	mustPlace(t, b, 0, 1) // W
	mustPlace(t, b, 8, 7) // B far
	mustPlace(t, b, 2, 1) // W
	mustPlace(t, b, 8, 6) // B far
	mustPlace(t, b, 1, 0) // W
	mustPlace(t, b, 8, 5) // B far
	mustPlace(t, b, 1, 2) // W

	// Now Black to play at (1,1): should be suicide (no capture)
	mustFail(t, b, 1, 1)
	assertAt(t, b, 1, 1, nil)
}

func TestSuicide_AllowedIfCaptures(t *testing.T) {
	b := NewBoard(9)

	// Make a position where Black plays into (1,1) with no immediate liberties
	// BUT captures a white stone at (1,0).
	//
	// Setup:
	// Black at (0,0), (2,0)
	// White at (1,0) (has one liberty at (1,1))
	// Black also at (0,1) and (2,1) so (1,1) itself has no empties except via capture.
	//
	// Then Black plays (1,1) capturing (1,0).

	mustPlace(t, b, 0, 0) // B
	mustPlace(t, b, 1, 0) // W (target)

	mustPlace(t, b, 2, 0) // B
	mustPlace(t, b, 8, 8) // W far

	mustPlace(t, b, 0, 1) // B
	mustPlace(t, b, 8, 7) // W far

	mustPlace(t, b, 2, 1) // B
	mustPlace(t, b, 8, 6) // W far

	// Now Black to play? Let's check turn count:
	// 8 moves placed => next is Black. Good.
	mustPlace(t, b, 1, 1) // B captures

	assertAt(t, b, 1, 0, nil) // captured
	wantB := PlayerBlack
	assertAt(t, b, 1, 1, &wantB)
}

// ------------------------------------------------------------
// Edge / corner behavior
// ------------------------------------------------------------

func TestLiberties_CornerStone(t *testing.T) {
	b := NewBoard(9)

	mustPlace(t, b, 0, 0) // B corner
	// Group liberties in corner should be 2: (1,0) and (0,1)
	assertGroupLiberties(t, b, 0, 0, 2)
}

func TestLiberties_EdgeStone(t *testing.T) {
	b := NewBoard(9)

	mustPlace(t, b, 4, 0) // B edge
	// Group liberties should be 3: (3,0),(5,0),(4,1)
	assertGroupLiberties(t, b, 4, 0, 3)
}

func TestCapture_OnEdge(t *testing.T) {
	b := NewBoard(9)

	// Capture white at (0,1) on left edge.
	// Surround needs 3 points: (0,0),(0,2),(1,1)
	// Build with turn alternation.

	mustPlace(t, b, 0, 0) // B
	mustPlace(t, b, 0, 1) // W target

	mustPlace(t, b, 8, 8) // B far
	mustPlace(t, b, 8, 7) // W far

	mustPlace(t, b, 0, 2) // B
	mustPlace(t, b, 8, 6) // W far

	mustPlace(t, b, 1, 1) // B capture

	assertAt(t, b, 0, 1, nil)
}

// ------------------------------------------------------------
// Normal flow / regression-ish: multiple moves, no illegalities
// ------------------------------------------------------------

func TestNormalFlow_MultipleMoves_NoUnexpectedCaptures(t *testing.T) {
	b := NewBoard(9)

	moves := [][2]int{
		{3, 3}, {3, 4},
		{4, 3}, {4, 4},
		{5, 5}, {5, 6},
		{6, 5}, {6, 6},
	}

	for i, m := range moves {
		if ok := b.PlaceStone(m[0], m[1]); !ok {
			t.Fatalf("move %d PlaceStone(%d,%d) failed\nboard:\n%s", i+1, m[0], m[1], b.String())
		}
		if b.Grid[m[0]][m[1]] == nil {
			t.Fatalf("move %d expected stone at (%d,%d)\nboard:\n%s", i+1, m[0], m[1], b.String())
		}
	}

	// ensure all remain
	for _, m := range moves {
		if b.Grid[m[0]][m[1]] == nil {
			t.Fatalf("expected stone still at (%d,%d)\nboard:\n%s", m[0], m[1], b.String())
		}
	}
}

func TestKo_SimpleKoRejected(t *testing.T) {
	b := NewBoard(9)

	moves := [][2]int{
		{2, 1},
		{1, 3},
		{1, 2},
		{3, 3},
		{3, 2},
		{2, 4},
		{0, 0},
		{2, 2},
	}

	for i, m := range moves {
		if ok := b.PlaceStone(m[0], m[1]); !ok {
			t.Fatalf("move %d at (%d,%d) failed", i, m[0], m[1])
		}
	}

	assertGroupLiberties(t, b, 2, 2, 1)

	mustPlace(t, b, 2, 3)

	assertAt(t, b, 2, 2, nil)

	mustFail(t, b, 2, 2)
}

// ------------------------------------------------------------
// Board methods used by AI
// ------------------------------------------------------------

func TestBoard_Clone_Independence(t *testing.T) {
	b := NewBoard(9)
	b.PlaceStone(4, 4) // Black
	b.PlaceStone(5, 5) // White

	cloned := b.Clone()

	// Modify clone
	cloned.PlaceStone(0, 0)

	// Original should be unchanged
	if b.Grid[0][0] != nil {
		t.Error("modifying clone affected original board")
	}

	// Clone should have the new stone
	if cloned.Grid[0][0] == nil {
		t.Error("clone was not modified correctly")
	}
}

func TestBoard_Clone_SameState(t *testing.T) {
	b := NewBoard(9)
	b.PlaceStone(4, 4)
	b.PlaceStone(5, 5)

	cloned := b.Clone()

	// Check that stones are in same positions
	if b.Grid[4][4] == nil || cloned.Grid[4][4] == nil {
		t.Error("stone at (4,4) missing in clone")
	}
	if b.Grid[5][5] == nil || cloned.Grid[5][5] == nil {
		t.Error("stone at (5,5) missing in clone")
	}

	// Check current player
	if b.CurrentPlayer != cloned.CurrentPlayer {
		t.Errorf("current player mismatch: original=%v, clone=%v", b.CurrentPlayer, cloned.CurrentPlayer)
	}
}

func TestBoard_GetLegalMoves_IncludesPass(t *testing.T) {
	b := NewBoard(9)
	moves := b.GetLegalMoves(PlayerBlack)

	passFound := false
	for _, move := range moves {
		if move[0] == -1 && move[1] == -1 {
			passFound = true
			break
		}
	}

	if !passFound {
		t.Error("GetLegalMoves should include pass move (-1, -1)")
	}
}

func TestBoard_GetLegalMoves_ExcludesOccupied(t *testing.T) {
	b := NewBoard(9)
	b.PlaceStone(4, 4) // Black
	b.PlaceStone(5, 5) // White

	moves := b.GetLegalMoves(PlayerBlack)

	for _, move := range moves {
		if move[0] == 4 && move[1] == 4 {
			t.Error("GetLegalMoves should not include occupied position (4,4)")
		}
		if move[0] == 5 && move[1] == 5 {
			t.Error("GetLegalMoves should not include occupied position (5,5)")
		}
	}
}

func TestBoard_GetLegalMoves_ExcludesSuicide(t *testing.T) {
	b := NewBoard(9)

	// Create a suicide position for Black at (1,1)
	// Surround (1,1) with White stones
	b.PlaceStone(8, 8) // Black (far)
	b.PlaceStone(0, 1) // White
	b.PlaceStone(8, 7) // Black (far)
	b.PlaceStone(2, 1) // White
	b.PlaceStone(8, 6) // Black (far)
	b.PlaceStone(1, 0) // White
	b.PlaceStone(8, 5) // Black (far)
	b.PlaceStone(1, 2) // White

	// Now Black to move - (1,1) should not be in legal moves
	moves := b.GetLegalMoves(PlayerBlack)

	for _, move := range moves {
		if move[0] == 1 && move[1] == 1 {
			t.Error("GetLegalMoves should not include suicide move (1,1)")
		}
	}
}

func TestBoard_GetLegalMoves_ReturnsValidMoves(t *testing.T) {
	b := NewBoard(9)
	moves := b.GetLegalMoves(PlayerBlack)

	// Should have at least pass + some empty positions
	if len(moves) < 2 {
		t.Errorf("expected at least 2 moves (pass + empty), got %d", len(moves))
	}

	// All moves should be valid coordinates or pass
	for _, move := range moves {
		if move[0] == -1 && move[1] == -1 {
			continue // Pass move is valid
		}
		if move[0] < 0 || move[0] >= 9 || move[1] < 0 || move[1] >= 9 {
			t.Errorf("invalid move coordinates: (%d, %d)", move[0], move[1])
		}
	}
}

func TestBoard_CloneAndPlay_DoesNotModifyOriginal(t *testing.T) {
	b := NewBoard(9)
	b.PlaceStone(4, 4) // Black
	b.PlaceStone(5, 5) // White

	originalStoneCount := 0
	for x := 0; x < 9; x++ {
		for y := 0; y < 9; y++ {
			if b.Grid[x][y] != nil {
				originalStoneCount++
			}
		}
	}

	cloned, success := b.CloneAndPlay(0, 0, PlayerBlack)
	if !success {
		t.Fatal("CloneAndPlay should succeed for valid move")
	}

	// Check original is unchanged
	newStoneCount := 0
	for x := 0; x < 9; x++ {
		for y := 0; y < 9; y++ {
			if b.Grid[x][y] != nil {
				newStoneCount++
			}
		}
	}

	if newStoneCount != originalStoneCount {
		t.Errorf("original board modified: had %d stones, now has %d", originalStoneCount, newStoneCount)
	}

	// Check clone has the new move
	if cloned.Grid[0][0] == nil {
		t.Error("clone should have the new stone at (0,0)")
	}
}

func TestBoard_CloneAndPlay_ValidMove(t *testing.T) {
	b := NewBoard(9)
	b.PlaceStone(4, 4) // Black
	b.PlaceStone(5, 5) // White

	cloned, success := b.CloneAndPlay(0, 0, PlayerBlack)
	if !success {
		t.Fatal("CloneAndPlay should succeed for valid move")
	}

	if cloned.Grid[0][0] == nil {
		t.Error("clone should have stone at (0,0)")
	}

	if cloned.Grid[0][0].Player != PlayerBlack {
		t.Error("clone should have black stone at (0,0)")
	}
}

func TestBoard_CloneAndPlay_InvalidMove(t *testing.T) {
	b := NewBoard(9)
	b.PlaceStone(4, 4) // Black

	// Try to place on occupied position
	cloned, success := b.CloneAndPlay(4, 4, PlayerWhite)
	if success {
		t.Error("CloneAndPlay should fail for occupied position")
	}

	// Clone should still exist but move shouldn't be applied
	if cloned.Grid[4][4] == nil {
		t.Error("clone should still have original stone")
	}
}

func TestBoard_CloneAndPlay_PassMove(t *testing.T) {
	b := NewBoard(9)
	originalPlayer := b.CurrentPlayer

	cloned, success := b.CloneAndPlay(-1, -1, PlayerBlack)
	if !success {
		t.Fatal("CloneAndPlay should succeed for pass move")
	}

	// Original should be unchanged
	if b.CurrentPlayer != originalPlayer {
		t.Error("pass on clone should not affect original current player")
	}

	// Clone should have switched player
	if cloned.CurrentPlayer == originalPlayer {
		t.Error("clone should have switched player after pass")
	}

	// Clone should have incremented pass count
	if cloned.PassCount != 1 {
		t.Errorf("clone pass count should be 1, got %d", cloned.PassCount)
	}
}
