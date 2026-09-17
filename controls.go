package tcbreplicants

import (
	"image/color"
	"runtime"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

const (
	maxLogicalWidth = 1280
	controlRadius   = 34
)

type controlState struct {
	SpeedUp    bool
	SpeedDown  bool
	VolumeUp   bool
	VolumeDown bool
}

type controlButton struct {
	X      float64
	Y      float64
	Radius float64
}

func (b controlButton) contains(x, y int) bool {
	dx := float64(x) - b.X
	dy := float64(y) - b.Y
	return dx*dx+dy*dy <= b.Radius*b.Radius
}

type controlLayout struct {
	SpeedUp    controlButton
	SpeedDown  controlButton
	VolumeUp   controlButton
	VolumeDown controlButton
}

func (s *controlState) press(layout controlLayout, x, y int) {
	if layout.SpeedUp.contains(x, y) {
		s.SpeedUp = true
	}
	if layout.SpeedDown.contains(x, y) {
		s.SpeedDown = true
	}
	if layout.VolumeUp.contains(x, y) {
		s.VolumeUp = true
	}
	if layout.VolumeDown.contains(x, y) {
		s.VolumeDown = true
	}
}

type controlSprites struct {
	speedUp    [2]*ebiten.Image
	speedDown  [2]*ebiten.Image
	volumeUp   [2]*ebiten.Image
	volumeDown [2]*ebiten.Image
}

func newControlSprites() controlSprites {
	return controlSprites{
		speedUp:    newControlSpritePair("S+"),
		speedDown:  newControlSpritePair("S-"),
		volumeUp:   newControlSpritePair("V+"),
		volumeDown: newControlSpritePair("V-"),
	}
}

func newControlSpritePair(label string) [2]*ebiten.Image {
	return [2]*ebiten.Image{
		newControlSprite(label, false),
		newControlSprite(label, true),
	}
}

func newControlSprite(label string, pressed bool) *ebiten.Image {
	const padding = 6
	size := controlRadius*2 + padding*2
	center := float64(size) / 2
	image := ebiten.NewImage(size, size)
	drawRoundButton(image, controlButton{X: center, Y: center, Radius: controlRadius}, pressed)
	ebitenutil.DebugPrintAt(image, label, int(center)-len(label)*3, int(center)-6)
	return image
}

func logicalWidth(outsideWidth, outsideHeight int) int {
	if outsideWidth <= 0 || outsideHeight <= 0 {
		return ScreenWidth
	}
	width := (outsideWidth*ScreenHeight + outsideHeight - 1) / outsideHeight
	return min(max(width, ScreenWidth), maxLogicalWidth)
}

func makeControlLayout(width, height int) controlLayout {
	sideWidth := float64(width-ScreenWidth) / 2
	leftX := sideWidth / 2
	rightX := float64(width) - sideWidth/2
	if sideWidth < controlRadius*2+16 {
		leftX = controlRadius + 8
		rightX = float64(width) - controlRadius - 8
	}
	upperY := float64(height) * 0.34
	lowerY := float64(height) * 0.66
	return controlLayout{
		SpeedUp:    controlButton{X: leftX, Y: upperY, Radius: controlRadius},
		SpeedDown:  controlButton{X: leftX, Y: lowerY, Radius: controlRadius},
		VolumeUp:   controlButton{X: rightX, Y: upperY, Radius: controlRadius},
		VolumeDown: controlButton{X: rightX, Y: lowerY, Radius: controlRadius},
	}
}

func (g *Game) readVirtualControls() controlState {
	if !g.virtualControlsVisible() {
		g.controls = controlState{}
		return g.controls
	}

	layout := makeControlLayout(g.layoutWidth, ScreenHeight)
	state := controlState{}
	g.touchIDs = ebiten.AppendTouchIDs(g.touchIDs[:0])
	for _, id := range g.touchIDs {
		x, y := ebiten.TouchPosition(id)
		state.press(layout, x, y)
	}
	if ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
		x, y := ebiten.CursorPosition()
		state.press(layout, x, y)
	}

	g.controls = state
	return state
}

func (g *Game) updateInput() {
	controls := g.readVirtualControls()

	if g.audioPlayer != nil {
		if ebiten.IsKeyPressed(ebiten.KeyUp) || controls.VolumeUp {
			g.audioPlayer.SetVolume(min(g.audioPlayer.Volume()+0.01, 1))
		}
		if ebiten.IsKeyPressed(ebiten.KeyDown) || controls.VolumeDown {
			g.audioPlayer.SetVolume(max(g.audioPlayer.Volume()-0.01, 0))
		}
	}

	speedUp := inpututil.IsKeyJustPressed(ebiten.KeyEqual) ||
		inpututil.IsKeyJustPressed(ebiten.KeyKPAdd) ||
		(controls.SpeedUp && !g.previousControls.SpeedUp)
	speedDown := inpututil.IsKeyJustPressed(ebiten.KeyMinus) ||
		inpututil.IsKeyJustPressed(ebiten.KeyKPSubtract) ||
		(controls.SpeedDown && !g.previousControls.SpeedDown)
	if speedUp {
		g.speedMultiplier = min(g.speedMultiplier+0.1, 2)
	}
	if speedDown {
		g.speedMultiplier = max(g.speedMultiplier-0.1, 0.5)
	}
	g.previousControls = controls
}

func (g *Game) virtualControlsVisible() bool {
	return runtime.GOOS == "android" || runtime.GOOS == "ios" || g.layoutWidth > ScreenWidth
}

func (g *Game) drawVirtualControls(dst *ebiten.Image) {
	layout := makeControlLayout(dst.Bounds().Dx(), dst.Bounds().Dy())
	drawControlSprite(dst, g.controlUI.speedUp[boolIndex(g.controls.SpeedUp)], layout.SpeedUp)
	drawControlSprite(dst, g.controlUI.speedDown[boolIndex(g.controls.SpeedDown)], layout.SpeedDown)
	drawControlSprite(dst, g.controlUI.volumeUp[boolIndex(g.controls.VolumeUp)], layout.VolumeUp)
	drawControlSprite(dst, g.controlUI.volumeDown[boolIndex(g.controls.VolumeDown)], layout.VolumeDown)
}

func boolIndex(value bool) int {
	if value {
		return 1
	}
	return 0
}

func drawControlSprite(dst, sprite *ebiten.Image, button controlButton) {
	var op ebiten.DrawImageOptions
	op.GeoM.Translate(
		button.X-float64(sprite.Bounds().Dx())/2,
		button.Y-float64(sprite.Bounds().Dy())/2,
	)
	dst.DrawImage(sprite, &op)
}

func drawRoundButton(dst *ebiten.Image, button controlButton, pressed bool) {
	fill := color.RGBA{R: 56, G: 16, B: 22, A: 180}
	border := color.RGBA{R: 224, G: 128, B: 128, A: 230}
	if pressed {
		fill = color.RGBA{R: 160, G: 38, B: 48, A: 230}
		border = color.RGBA{R: 255, G: 255, B: 255, A: 255}
	}
	vector.FillCircle(dst, float32(button.X), float32(button.Y), float32(button.Radius), fill, true)
	vector.StrokeCircle(dst, float32(button.X), float32(button.Y), float32(button.Radius), 3, border, true)
}
