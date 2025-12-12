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

func mustPlace(t *testing.T, b *Board, x, y int) {
	t.Helper()
	if ok := b.PlaceStone(x, y); !ok {
		t.Fatalf("expected PlaceStone(%d,%d) to succeed\nboard:\n%s", x, y, b.String())
	}
}

func mustFail(t *testing.T, b *Board, x, y int) {
	t.Helper()
	if ok := b.PlaceStone(x, y); ok {
		t.Fatalf("expected PlaceStone(%d,%d) to fail\nboard:\n%s", x, y, b.String())
	}
}

func assertAt(t *testing.T, b *Board, x, y int, want *Player) {
	t.Helper()
	s := b.Grid[x][y]
	if want == nil {
		if s != nil {
			t.Fatalf("expected empty at (%d,%d), got %v\nboard:\n%s", x, y, s.Player, b.String())
		}
		return
	}
	if s == nil {
		t.Fatalf("expected stone at (%d,%d), got empty\nboard:\n%s", x, y, b.String())
	}
	if s.Player != *want {
		t.Fatalf("expected %v at (%d,%d), got %v\nboard:\n%s", *want, x, y, s.Player, b.String())
	}
}

func assertSingleGroupAt(t *testing.T, b *Board, coords ...[2]int) {
	t.Helper()
	if len(coords) == 0 {
		t.Fatal("assertSingleGroupAt needs coords")
	}
	first := b.Grid[coords[0][0]][coords[0][1]]
	if first == nil || first.Group == nil {
		t.Fatalf("expected stone+group at %v\nboard:\n%s", coords[0], b.String())
	}
	gid := first.Group.ID
	for _, c := range coords[1:] {
		s := b.Grid[c[0]][c[1]]
		if s == nil || s.Group == nil || s.Group.ID != gid {
			t.Fatalf("expected %v to be in same group as %v\nboard:\n%s", c, coords[0], b.String())
		}
	}
}

func assertGroupLiberties(t *testing.T, b *Board, x, y int, want int) {
	t.Helper()
	s := b.Grid[x][y]
	if s == nil || s.Group == nil {
		t.Fatalf("expected stone+group at (%d,%d)\nboard:\n%s", x, y, b.String())
	}
	got := len(s.Group.Liberties)
	if got != want {
		t.Fatalf("expected group liberties=%d at (%d,%d), got %d\nboard:\n%s", want, x, y, got, b.String())
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
	b := NewBoard(9)
	cases := [][2]int{{-1, 0}, {0, -1}, {9, 0}, {0, 9}, {100, 100}}
	for _, c := range cases {
		mustFail(t, b, c[0], c[1])
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
	b := NewBoard(9)

	mustPlace(t, b, 0, 0) // B
	mustPlace(t, b, 0, 1) // W
	mustPlace(t, b, 0, 2) // B

	wantB := PlayerBlack
	wantW := PlayerWhite
	assertAt(t, b, 0, 0, &wantB)
	assertAt(t, b, 0, 1, &wantW)
	assertAt(t, b, 0, 2, &wantB)
}

// ------------------------------------------------------------
// Group creation / merge
// ------------------------------------------------------------

func TestGroups_MergeFriendlyGroupsThroughBridge(t *testing.T) {
	b := NewBoard(9)

	// Build:
	// B . B   (row 0)
	// then bridge at (1,0)
	mustPlace(t, b, 0, 0) // B
	mustPlace(t, b, 8, 8) // W (far)
	mustPlace(t, b, 2, 0) // B
	mustPlace(t, b, 8, 7) // W (far)
	mustPlace(t, b, 1, 0) // B bridges

	assertSingleGroupAt(t, b, [2]int{0, 0}, [2]int{1, 0}, [2]int{2, 0})
}

func TestGroups_UniqueLibertiesNotDoubleCounted(t *testing.T) {
	b := NewBoard(9)

	// Make a 2-stone vertical black group at (4,4) and (4,5).
	// Liberties should be 6 (not 8):
	// (4,4) has 4 liberties, (4,5) has 4, but they share (4,4)/(4,5) adjacency not a liberty,
	// and they also share none as liberties; actual unique empties = 6.
	mustPlace(t, b, 4, 4) // B
	mustPlace(t, b, 8, 8) // W far
	mustPlace(t, b, 4, 5) // B

	assertSingleGroupAt(t, b, [2]int{4, 4}, [2]int{4, 5})
	assertGroupLiberties(t, b, 4, 4, 6)
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
