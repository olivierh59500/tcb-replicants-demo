//go:build dck_fidelity_rendercheck

package tcbreplicants

import (
	"fmt"
	"math/rand"
	"os"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
	capture "github.com/olivierh59500/democonstructionkit/fidelity/ebiten"
)

// This opt-in GPU check captures the splash boundary and the running scene.
// Fixed star seeds make separate DCK revisions comparable pixel for pixel.
func TestMain(m *testing.M) {
	if code := m.Run(); code != 0 {
		os.Exit(code)
	}
	directory := os.Getenv("DCK_REPLICANTS_CAPTURE_DIR")
	if directory == "" {
		var err error
		directory, err = os.MkdirTemp("", "replicants-dck-capture-")
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	}
	var demo *Game
	err := capture.Run(capture.Config{
		Directory: directory,
		Frames:    []int{0, 1, 99, 100, 101, 240, 600, 1200, 2400},
		Width:     ScreenWidth,
		Height:    ScreenHeight,
	}, func() (ebiten.Game, error) {
		demo = NewGame()
		demo.rng = rand.New(rand.NewSource(42))
		// Audio playback does not control image motion in this production.
		demo.audioReady = true
		if err := demo.Init(); err != nil {
			return nil, err
		}
		return demo, nil
	})
	if demo != nil {
		demo.Cleanup()
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Printf("Replicants DCK captures: %s\n", directory)
}
