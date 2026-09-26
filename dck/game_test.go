package tcbreplicants

import (
	"math"
	"testing"

	"github.com/olivierh59500/democonstructionkit/presets"
)

func TestForegroundTrainMatchesAuthoredRectifiedPhases(t *testing.T) {
	config := presets.ReplicantsBouncingSprites(nil)
	if config.Count != 2 || config.X.Offset != 32 || config.Spacing.X != 480 || config.Y.Wave == nil {
		t.Fatal("foreground sprite geometry changed")
	}
	phase := 0.0
	for tick := 0; tick < 1000; tick++ {
		phase += .1 * 1.4
		first := config.Y.Offset + config.Y.Wave.At(0, phase)
		second := config.Y.Offset + config.Y.Wave.At(1, phase)
		if math.Abs(first-(326-math.Abs(math.Cos(phase)*24))) > 1e-12 ||
			math.Abs(second-(326-math.Abs(math.Sin(phase)*24))) > 1e-12 {
			t.Fatalf("tick %d: foreground sprite heights %.12f/%.12f", tick, first, second)
		}
	}
}

func TestTextTilesPreservesSpacingAndUnsupportedCharacters(t *testing.T) {
	lookup, err := presets.TileLookup("tcb-replicants-demo", true)
	if err != nil {
		t.Fatal(err)
	}
	got := make([]int, 0, 5)
	for _, r := range " A#Z " {
		index, ok := lookup(r)
		if !ok {
			index = -1
		}
		got = append(got, index)
	}
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
	if got, want := len(presets.ReplicantsRowWave()), 1191; got != want {
		t.Fatalf("scroll lookup length = %d, want %d", got, want)
	}
}
