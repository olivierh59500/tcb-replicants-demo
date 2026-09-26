// Package tcbreplicants implements the TCB-Replicants demo remake.
package tcbreplicants

import (
	"bytes"
	"fmt"
	"github.com/olivierh59500/democonstructionkit/presets"
	originalassets "tcb-replicants-demo"

	"github.com/olivierh59500/democonstructionkit/sound"

	"image"
	"image/color"

	kit "github.com/olivierh59500/democonstructionkit"
	"github.com/olivierh59500/democonstructionkit/composite"
	"github.com/olivierh59500/democonstructionkit/scrolling"
	"github.com/olivierh59500/democonstructionkit/sprites"

	_ "image/png"
	"log"
	"math/rand"
	"time"

	"github.com/hajimehoshi/ebiten/v2"

	audio "github.com/olivierh59500/democonstructionkit/sound/output"
)

const (
	ScreenWidth  = 640
	ScreenHeight = 400

	scrollHeight = 50
	scrollSpeed  = 4.0
)

var spriteData = originalassets.
	DCKAssetSpriteData()

var tcbLogoData = originalassets.
	DCKAssetTcbLogoData()

var repLogoData = originalassets.
	DCKAssetRepLogoData()

var scrollFontData = originalassets.
	DCKAssetScrollFontData()

var splashData = originalassets.
	DCKAssetSplashData()

var ymData = originalassets.DCKAssetYmData()

const replicantsMessage = `      YO, SHITY-FUCKY-LAMEEUUUUURS !!!  AFTER HARD LABOUR, THE MEGAMIGHTY CAREBEARS AND THE FAMOUS REPLICANTS ARE PROUD TO PRESENT    - WEIRD DREAM -   CRACKED BY RATBOY.  THIS INTRO WAS CODED BY NICK, JAS AND AN THE MOTHERFUCKIN COOL AT THE FIRST MEETING TCB - REPLICANTS...    OK, NOW ALL THE MEMBERS OF THE WILL WRITE A PART OF THIS SCROLLTEXT...          HEY, IT'S RATBOY ON THE KEYBOARD, I DON'T KNOW WHAT TO WRITE AND I HATE WRITING SCROLLTEXT.  I'LL TELL YOU MORE DETAILS ABOUT THIS MEETING. TCB ARRIVED FIVE DAYS AGO. SO, THEY DECIDED TO CODE THIS FANTASTIC INTRO. AFTER 25 LITRES OF COKE, 1 MONOPOLY PLAY, 1 BOTTLE OF WHISKY, 20 BIG TOM, SOME PING-PONG MATCHES (OK, JAS !!  YOU'RE BETTER THAN ME, BUT THE REVENGE OF RATBOY WILL BE TERRIBLE !), THIS INTRO IS FINISHED...`

// Game contains the shared desktop and mobile game state.
type Game struct {
	sprite     *ebiten.Image
	tcbLogo    *ebiten.Image
	repLogo    *ebiten.Image
	scrollFont *ebiten.Image
	splash     *ebiten.Image
	scene      *ebiten.Image

	splashReveal *composite.BlockReveal

	audioContext *audio.Context
	audioPlayer  *audio.Player
	musicStream  *sound.Stream
	audioReady   bool
	musicStarted bool

	rng          *rand.Rand
	stars        *sprites.AnimatedField
	starFrames   []*ebiten.Image
	logos        *sprites.CoupledLogoPair
	bouncing     *sprites.Train
	scrollEffect *scrolling.Scrolling

	speedMultiplier float64

	layoutWidth      int
	touchIDs         []ebiten.TouchID
	controls         controlState
	previousControls controlState
	controlUI        controlSprites

	initialized bool
}

// NewGame constructs state without opening platform services. In particular,
// audio is deferred until the first Update so Android has a ready Activity.
func NewGame() *Game {
	g := &Game{
		rng:             rand.New(rand.NewSource(time.Now().UnixNano())),
		speedMultiplier: 1.4,
		layoutWidth:     ScreenWidth,
	}
	return g
}

func (g *Game) loadImages() error {
	var err error
	if g.sprite, err = decodeImage("union sprite", spriteData); err != nil {
		return err
	}
	if g.tcbLogo, err = decodeImage("TCB logo", tcbLogoData); err != nil {
		return err
	}
	if g.repLogo, err = decodeImage("Replicants logo", repLogoData); err != nil {
		return err
	}
	if g.scrollFont, err = decodeImage("scroll font", scrollFontData); err != nil {
		return err
	}
	if g.splash, err = decodeImage("splash", splashData); err != nil {
		return err
	}
	return nil
}

func decodeImage(name string, data []byte) (*ebiten.Image, error) {
	if len(data) == 0 {
		return nil, fmt.Errorf("%s data is empty", name)
	}
	decoded, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("decode %s: %w", name, err)
	}
	return ebiten.NewImageFromImage(decoded), nil
}

// Init creates GPU resources. It is idempotent and runs from Update.
func (g *Game) Init() error {
	if g.initialized {
		return nil
	}
	if err := g.loadImages(); err != nil {
		return err
	}

	g.scene = ebiten.NewImage(ScreenWidth, ScreenHeight)
	var err error
	g.starFrames, err = sprites.NewSolidFrames(presets.ReplicantsStarMaterials())
	if err != nil {
		return err
	}
	starConfig, err := presets.ReplicantsStars(g.starFrames, presets.DefaultReplicantsStarOptions(g.rng.Intn))
	if err != nil {
		return err
	}
	g.stars, err = sprites.NewAnimatedField(starConfig)
	if err != nil {
		return err
	}
	atlas, err := presets.FontAtlas("tcb-replicants-demo", g.scrollFont)
	if err != nil {
		return err
	}
	config := presets.ReplicantsRowColumn(atlas, replicantsMessage)
	g.scrollEffect, err = scrolling.New(scrolling.Config{RowColumn: &config})
	if err != nil {
		return err
	}
	g.logos, err = sprites.NewCoupledLogoPair(presets.ReplicantsLogoPair(g.repLogo, g.tcbLogo))
	if err != nil {
		return err
	}
	g.splashReveal, err = composite.NewBlockReveal(presets.ReplicantsSplash(g.splash))
	if err != nil {
		return err
	}
	g.bouncing, err = sprites.NewTrain(presets.ReplicantsBouncingSprites(g.sprite))
	if err != nil {
		return err
	}
	g.controlUI = newControlSprites()
	g.initialized = true
	return nil
}

// Update advances the demo by one tick.
func (g *Game) Update() error {
	if err := g.Init(); err != nil {
		return err
	}

	if !g.audioReady {
		g.audioReady = true
		g.initAudio()
		g.startMusic()
	}

	g.updateInput()

	if !g.splashReveal.Done() {
		if err := g.splashReveal.Update(kit.Frame{}); err != nil {
			return err
		}
		return nil
	}

	if err := g.stars.Motion().SetSpeedMultiplier(g.speedMultiplier); err != nil {
		return err
	}
	if err := g.stars.Update(kit.Frame{}); err != nil {
		return err
	}

	if err := g.scrollEffect.SetTransportMultiplier(g.speedMultiplier); err != nil {
		return err
	}
	if err := g.scrollEffect.Update(kit.Frame{}); err != nil {
		return err
	}

	if err := g.bouncing.SetSpeedMultiplier(g.speedMultiplier); err != nil {
		return err
	}
	if err := g.bouncing.Update(kit.Frame{}); err != nil {
		return err
	}
	if err := g.logos.SetSpeedMultiplier(g.speedMultiplier); err != nil {
		return err
	}
	if err := g.logos.Update(kit.Frame{}); err != nil {
		return err
	}
	return nil
}

func drawImageAt(dst, source *ebiten.Image, x, y int) {
	drawImageAtFloat(dst, source, float64(x), float64(y))
}

func drawImageAtFloat(dst, source *ebiten.Image, x, y float64) {
	op := ebiten.DrawImageOptions{}
	op.GeoM.Translate(x, y)
	composite.Instance{Image: source, Options: op}.Draw(dst)
}

func (g *Game) drawScene(dst *ebiten.Image) {
	if !g.splashReveal.Done() {
		g.splashReveal.Draw(dst)
		return
	}

	dst.Fill(color.Black)
	g.stars.Draw(dst)
	g.logos.Draw(dst)
	g.scrollEffect.Draw(dst)
	g.bouncing.Draw(dst)
}

// Draw renders the fixed-size scene centered within the current logical view.
func (g *Game) Draw(screen *ebiten.Image) {
	if !g.initialized {
		return
	}

	if screen.Bounds().Dx() == ScreenWidth && screen.Bounds().Dy() == ScreenHeight {
		g.drawScene(screen)
	} else {
		g.drawScene(g.scene)
		screen.Fill(color.Black)
		drawImageAt(screen, g.scene, (screen.Bounds().Dx()-ScreenWidth)/2, 0)
	}

	if g.virtualControlsVisible() {
		g.drawVirtualControls(screen)
	}
}

// Layout preserves the original pixels and uses wide side areas on phones.
func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	g.layoutWidth = logicalWidth(outsideWidth, outsideHeight)
	return g.layoutWidth, ScreenHeight
}

// Cleanup releases audio resources owned by the desktop game instance.
func (g *Game) Cleanup() {
	if g.splashReveal != nil {
		_ = g.splashReveal.Close()
		g.splashReveal = nil
	}
	if g.logos != nil {
		_ = g.logos.Close()
		g.logos = nil
	}
	for _, frame := range g.starFrames {
		frame.Deallocate()
	}
	g.starFrames = nil
	if g.scrollEffect != nil {
		g.scrollEffect.Close()
	}
	if g.audioPlayer != nil {
		if err := g.audioPlayer.Close(); err != nil {
			log.Printf("close audio player: %v", err)
		}
		g.audioPlayer = nil
	}
	if g.musicStream != nil {
		if err := g.musicStream.Close(); err != nil {
			log.Printf("close music stream: %v", err)
		}
		g.musicStream = nil
	}
}
