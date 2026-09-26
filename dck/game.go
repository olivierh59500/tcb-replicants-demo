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
	"math"
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

	tcbFrames  []*ebiten.Image
	repFrames  []*ebiten.Image
	splashRows []*ebiten.Image

	audioContext *audio.Context
	audioPlayer  *audio.Player
	musicStream  *sound.Stream
	audioReady   bool
	musicStarted bool

	rng          *rand.Rand
	stars        *sprites.AnimatedField
	starFrames   []*ebiten.Image
	scrollEffect *scrolling.Scrolling

	splashTime int
	splashLine int
	offsetScr  float64
	offRep     float64
	offTcb     float64

	speedMultiplier float64

	layoutWidth      int
	touchIDs         []ebiten.TouchID
	controls         controlState
	previousControls controlState
	controlUI        controlSprites

	initialized bool
	splashDone  bool
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

func (g *Game) preRenderLogoFrames() {
	g.tcbFrames = preRenderFrames(g.tcbLogo, 40)
	g.repFrames = preRenderFrames(g.repLogo, 35)
}

func preRenderFrames(source *ebiten.Image, count int) []*ebiten.Image {
	frames := make([]*ebiten.Image, count)
	for i := range frames {
		scale := float64(i+1) / float64(count)
		width := max(1, int(float64(source.Bounds().Dx())*scale))
		height := max(1, int(float64(source.Bounds().Dy())*scale))
		frame := ebiten.NewImage(width, height)
		var op ebiten.DrawImageOptions
		op.GeoM.Scale(scale, scale)
		frame.DrawImage(source, &op)
		frames[i] = frame
	}
	return frames
}

func (g *Game) cacheSplashRows() {
	const rowHeight = 40
	rowCount := (g.splash.Bounds().Dy() + rowHeight - 1) / rowHeight
	g.splashRows = make([]*ebiten.Image, rowCount)
	for row := range g.splashRows {
		y := row * rowHeight
		bottom := min(y+rowHeight, g.splash.Bounds().Dy())
		g.splashRows[row] = g.splash.SubImage(image.Rect(0, y, g.splash.Bounds().Dx(), bottom)).(*ebiten.Image)
	}
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
	g.preRenderLogoFrames()
	g.cacheSplashRows()
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

	if !g.splashDone {
		g.splashTime++
		if g.splashTime >= 100 {
			g.splashDone = true
		} else if g.splashTime%3 == 0 && g.splashLine < ScreenHeight {
			g.splashLine = min(g.splashLine+40, ScreenHeight)
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

	g.offsetScr += 0.1 * g.speedMultiplier
	g.offRep += (2.0 / 180.0 * math.Pi) * g.speedMultiplier
	g.offTcb += (7.0 / 180.0 * math.Pi) * g.speedMultiplier
	return nil
}

func (g *Game) drawSplash(dst *ebiten.Image) {
	dst.Fill(color.RGBA{R: 0xE0, G: 0xE0, B: 0xE0, A: 0xFF})
	visibleRows := min((g.splashLine+39)/40, len(g.splashRows))
	for row := 0; row < visibleRows; row++ {
		drawImageAt(dst, g.splashRows[row], 0, row*40)
	}
}

func (g *Game) drawLogos(dst *ebiten.Image) {
	zRep := (1 + math.Sin(g.offRep)) / 2
	yRep := 140 + (math.Cos(g.offRep)+math.Sin(g.offRep))*40
	zTCB := zRep + math.Cos(g.offTcb)/8
	yTCB := yRep + math.Sin(g.offTcb)*(70*zRep)

	repFrame := min(int(zRep*35), len(g.repFrames)-1)
	tcbFrame := min(int(4+zTCB*32), len(g.tcbFrames)-1)
	repFrame = max(repFrame, 0)
	tcbFrame = max(tcbFrame, 0)

	if zTCB >= zRep {
		if repFrame > 0 {
			drawCentered(dst, g.repFrames[repFrame], 320, yRep)
		}
		if tcbFrame > 0 {
			drawCentered(dst, g.tcbFrames[tcbFrame], 320, yTCB)
		}
		return
	}
	if tcbFrame > 0 {
		drawCentered(dst, g.tcbFrames[tcbFrame], 320, yTCB)
	}
	if repFrame > 0 {
		drawCentered(dst, g.repFrames[repFrame], 320, yRep)
	}
}

func drawCentered(dst, source *ebiten.Image, centerX int, centerY float64) {
	op := ebiten.DrawImageOptions{}
	op.GeoM.Translate(float64(centerX-source.Bounds().Dx()/2), centerY-float64(source.Bounds().Dy())/2)
	composite.Instance{Image: source, Options: op}.Draw(dst)
}

func (g *Game) drawSprites(dst *ebiten.Image) {
	drawImageAtFloat(dst, g.sprite, 32, 326-math.Abs(math.Cos(g.offsetScr)*24))
	drawImageAtFloat(dst, g.sprite, 512, 326-math.Abs(math.Sin(g.offsetScr)*24))
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
	if !g.splashDone {
		g.drawSplash(dst)
		return
	}

	dst.Fill(color.Black)
	g.stars.Draw(dst)
	g.drawLogos(dst)
	g.scrollEffect.Draw(dst)
	g.drawSprites(dst)
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
