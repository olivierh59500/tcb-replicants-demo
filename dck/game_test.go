package tcbreplicants

import "testing"

func TestTextTilesPreservesSpacingAndUnsupportedCharacters(t *testing.T) {
	got := textTiles(" A#Z ")
	want := []int{-1, 33, -1, 58, -1}
	if len(got) != len(want) {
		t.Fatalf("tiles length = %d, want %d", len(got), len(want))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("tiles[%d] = %d, want %d", i, got[i], want[i])
		}
	}
}

func TestNewGameDefersPlatformInitialization(t *testing.T) {
	game := NewGame()
	if game.initialized {
		t.Fatal("NewGame initialized GPU resources")
	}
	if game.audioContext != nil || game.audioPlayer != nil || game.audioReady {
		t.Fatal("NewGame initialized audio before the game loop")
	}
	if got, want := len(game.scrollX), 1191; got != want {
		t.Fatalf("scroll lookup length = %d, want %d", got, want)
	}
}
