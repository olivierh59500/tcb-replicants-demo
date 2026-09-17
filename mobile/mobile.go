// Package mobile exposes the game to ebitenmobile.
package mobile

import (
	enginemobile "github.com/hajimehoshi/ebiten/v2/mobile"

	demo "tcb-replicants-demo"
)

func init() {
	enginemobile.SetGame(demo.NewGame())
}

// Dummy forces gomobile to include this package in the Android binding.
func Dummy() {}
