package main

import (
	"log"

	"github.com/hajimehoshi/ebiten/v2"

	demo "tcb-replicants-demo/dck"
)

func main() {
	ebiten.SetWindowSize(demo.ScreenWidth*2, demo.ScreenHeight*2)
	ebiten.SetWindowTitle("TCB-Replicants Demo - Go/Ebitengine Port")
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)
	ebiten.SetScreenClearedEveryFrame(false)

	game := demo.NewGame()
	defer game.Cleanup()
	if err := ebiten.RunGame(newDrawOnUpdateGame(game)); err != nil {
		log.Fatal(err)
	}
}
