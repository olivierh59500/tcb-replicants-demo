package main

import (
	"bytes"
	_ "embed"
	"fmt"
	"image"
	"image/color"
	_ "image/jpeg"
	_ "image/png"
	"log"
	"math"
	"math/rand"

	//	"os"
	//	"path/filepath"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/audio"
	"github.com/hajimehoshi/ebiten/v2/audio/mp3"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

const (
	screenWidth  = 640
	screenHeight = 400
	scrollHeight = 50
	scrollSpeed  = 4.0 // Doubled from 2.0
	starSpeed    = 1.0 // Doubled from 0.5
)

// Embed assets
//
//go:embed assets/union_sprite.png
var spriteData []byte

//go:embed assets/tcb_logo.png
var tcbLogoData []byte

//go:embed assets/rep_logo.png
var repLogoData []byte

//go:embed assets/tcb_rep_font.png
var scrollFontData []byte

//go:embed assets/tcb_rep_splash.png
var splashData []byte

//go:embed assets/music.mp3
var musicData []byte

// Star represents a star in the starfield
type Star struct {
	x, y  float64
	speed float64
	color color.Color
	size  float64
}

// ScrollText manages the scrolling text
type ScrollText struct {
	text         string
	x            float64
	fontImage    *ebiten.Image
	charWidth    int
	charHeight   int
	charsPerRow  int
	scrollBuffer *ebiten.Image
	workBuffer   *ebiten.Image // Wider buffer for deformation
	deformBuffer *ebiten.Image
}

// Logo represents a bouncing logo
type Logo struct {
	image     *ebiten.Image
	x, y      float64
	scale     float64
	angle     float64
	angleInc  float64
	scaleBase float64
}

// Game represents the game state
type Game struct {
	// Images
	sprite     *ebiten.Image
	tcbLogo    *ebiten.Image
	repLogo    *ebiten.Image
	scrollFont *ebiten.Image
	splash     *ebiten.Image

	// Pre-rendered logo frames
	tcbFrames []*ebiten.Image
	repFrames []*ebiten.Image

	// Audio
	audioContext *audio.Context
	audioPlayer  *audio.Player
	currentSong  int

	// Effects
	stars      []*Star
	scrollText *ScrollText
	scrollX    []float64
	scrollXMod int

	// Animation state
	vbl        int
	splashTime int
	splashLine int
	offsetScr  float64
	offRep     float64
	offTcb     float64

	// Speed multiplier
	speedMultiplier float64

	// Game state
	initialized bool
	splashDone  bool
}

// NewGame creates a new game instance
func NewGame() *Game {
	g := &Game{
		stars:           make([]*Star, 0),
		splashTime:      0,
		splashLine:      0,
		speedMultiplier: 1.4, // Default speed at 1.4x
	}

	// Initialize scroll X positions
	g.initScrollX()

	// Initialize audio context
	g.audioContext = audio.NewContext(44100)

	return g
}

// initScrollX initializes the scroll deformation positions
func (g *Game) initScrollX() {
	g.scrollX = make([]float64, 0)

	// First wave pattern
	stp1 := 7.0 / 180.0 * math.Pi
	stp2 := 3.0 / 180.0 * math.Pi
	for i := 0; i < 389; i++ {
		x := 20*math.Sin(float64(i)*stp1) + 30*math.Cos(float64(i)*stp2)
		g.scrollX = append(g.scrollX, x)
	}

	// Second wave pattern
	stp1 = 72.0 / 180.0 * math.Pi
	for i := 0; i < 120; i++ {
		x := 4 * math.Sin(float64(i)*stp1)
		g.scrollX = append(g.scrollX, x)
	}

	// Third wave pattern
	stp1 = 8.0 / 180.0 * math.Pi
	for i := 0; i < 68; i++ {
		x := 40 * math.Sin(float64(i)*stp1)
		g.scrollX = append(g.scrollX, x)
	}

	// Repeat first pattern
	stp1 = 7.0 / 180.0 * math.Pi
	stp2 = 3.0 / 180.0 * math.Pi
	for i := 0; i < 389; i++ {
		x := 20*math.Sin(float64(i)*stp1) + 30*math.Cos(float64(i)*stp2)
		g.scrollX = append(g.scrollX, x)
	}

	// Small wave
	stp1 = 72.0 / 180.0 * math.Pi
	for i := 0; i < 36; i++ {
		x := 4 * math.Sin(float64(i)*stp1)
		g.scrollX = append(g.scrollX, x)
	}

	// Final wave
	stp1 = 8.0 / 180.0 * math.Pi
	for i := 0; i < 189; i++ {
		x := 30 * math.Sin(float64(i)*stp1)
		g.scrollX = append(g.scrollX, x)
	}

	g.scrollXMod = len(g.scrollX)
}

// loadImages loads all image assets
func (g *Game) loadImages() error {
	var err error

	// Check if assets are embedded
	if len(spriteData) == 0 {
		return fmt.Errorf("sprite data is empty - make sure assets/union_sprite.png exists")
	}

	// Load sprite
	img, format, err := image.Decode(bytes.NewReader(spriteData))
	if err != nil {
		return fmt.Errorf("failed to load sprite (format: %s): %v", format, err)
	}
	g.sprite = ebiten.NewImageFromImage(img)

	// Load TCB logo
	if len(tcbLogoData) == 0 {
		return fmt.Errorf("TCB logo data is empty - make sure assets/tcb_logo.png exists")
	}
	img, _, err = image.Decode(bytes.NewReader(tcbLogoData))
	if err != nil {
		return fmt.Errorf("failed to load TCB logo: %v", err)
	}
	g.tcbLogo = ebiten.NewImageFromImage(img)

	// Load REP logo
	if len(repLogoData) == 0 {
		return fmt.Errorf("REP logo data is empty - make sure assets/rep_logo.png exists")
	}
	img, _, err = image.Decode(bytes.NewReader(repLogoData))
	if err != nil {
		return fmt.Errorf("failed to load REP logo: %v", err)
	}
	g.repLogo = ebiten.NewImageFromImage(img)

	// Load scroll font
	if len(scrollFontData) == 0 {
		return fmt.Errorf("scroll font data is empty - make sure assets/tcb_rep_font.png exists")
	}
	img, _, err = image.Decode(bytes.NewReader(scrollFontData))
	if err != nil {
		return fmt.Errorf("failed to load scroll font: %v", err)
	}
	g.scrollFont = ebiten.NewImageFromImage(img)

	// Load splash
	if len(splashData) == 0 {
		return fmt.Errorf("splash data is empty - make sure assets/tcb_rep_splash.png exists")
	}
	img, _, err = image.Decode(bytes.NewReader(splashData))
	if err != nil {
		return fmt.Errorf("failed to load splash: %v", err)
	}
	g.splash = ebiten.NewImageFromImage(img)

	return nil
}

// initStarfield initializes the starfield
func (g *Game) initStarfield() {
	// Three layers of stars with different speeds and colors
	starParams := []struct {
		count int
		speed float64
		color color.Color
		size  float64
	}{
		{35, 11.2 * starSpeed, color.RGBA{0xE0, 0xA0, 0xA0, 0xFF}, 2},
		{35, 5.6 * starSpeed, color.RGBA{0xC0, 0x60, 0x60, 0xFF}, 2},
		{35, 2.8 * starSpeed, color.RGBA{0x80, 0x40, 0x40, 0xFF}, 2},
	}

	for _, param := range starParams {
		for i := 0; i < param.count; i++ {
			star := &Star{
				x:     float64(rand.Intn(screenWidth)),
				y:     float64(rand.Intn(280)), // Stars only in upper part
				speed: param.speed,
				color: param.color,
				size:  param.size,
			}
			g.stars = append(g.stars, star)
		}
	}
}

// initScrollText initializes the scrolling text
func (g *Game) initScrollText() {
	scrollText := `      YO, SHITY-FUCKY-LAMEEUUUUURS !!!  AFTER HARD LABOUR, THE MEGAMIGHTY CAREBEARS AND THE FAMOUS REPLICANTS ARE PROUD TO PRESENT    - WEIRD DREAM -   CRACKED BY RATBOY.  THIS INTRO WAS CODED BY NICK, JAS AND AN THE MOTHERFUCKIN COOL AT THE FIRST MEETING TCB - REPLICANTS...    OK, NOW ALL THE MEMBERS OF THE WILL WRITE A PART OF THIS SCROLLTEXT...          HEY, IT'S RATBOY ON THE KEYBOARD, I DON'T KNOW WHAT TO WRITE AND I HATE WRITING SCROLLTEXT.  I'LL TELL YOU MORE DETAILS ABOUT THIS MEETING. TCB ARRIVED FIVE DAYS AGO. SO, THEY DECIDED TO CODE THIS FANTASTIC INTRO. AFTER 25 LITRES OF COKE, 1 MONOPOLY PLAY, 1 BOTTLE OF WHISKY, 20 BIG TOM, SOME PING-PONG MATCHES (OK, JAS !!  YOU'RE BETTER THAN ME, BUT THE REVENGE OF RATBOY WILL BE TERRIBLE !), THIS INTRO IS FINISHED...`

	g.scrollText = &ScrollText{
		text:         scrollText,
		x:            0,
		fontImage:    g.scrollFont,
		charWidth:    64,
		charHeight:   50,
		charsPerRow:  10,                                             // 10 characters per row in the font
		scrollBuffer: ebiten.NewImage(screenWidth+256, scrollHeight), // Extra width for smooth scrolling
		workBuffer:   ebiten.NewImage(screenWidth+512, scrollHeight), // Even wider for deformation
		deformBuffer: ebiten.NewImage(screenWidth, scrollHeight),
	}
}

// preRenderLogoFrames pre-renders logo frames for optimization
func (g *Game) preRenderLogoFrames() {
	// Pre-render TCB logo frames (40 different scales)
	g.tcbFrames = make([]*ebiten.Image, 40)
	for i := 0; i < 40; i++ {
		scale := float64(i+1) / 40.0
		w := int(float64(g.tcbLogo.Bounds().Dx()) * scale)
		h := int(float64(g.tcbLogo.Bounds().Dy()) * scale)
		if w < 1 || h < 1 {
			continue
		}

		frame := ebiten.NewImage(w, h)
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Scale(scale, scale)
		frame.DrawImage(g.tcbLogo, op)
		g.tcbFrames[i] = frame
	}

	// Pre-render REP logo frames (35 different scales)
	g.repFrames = make([]*ebiten.Image, 35)
	for i := 0; i < 35; i++ {
		scale := float64(i+1) / 35.0
		w := int(float64(g.repLogo.Bounds().Dx()) * scale)
		h := int(float64(g.repLogo.Bounds().Dy()) * scale)
		if w < 1 || h < 1 {
			continue
		}

		frame := ebiten.NewImage(w, h)
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Scale(scale, scale)
		frame.DrawImage(g.repLogo, op)
		g.repFrames[i] = frame
	}
}

// Init initializes the game
func (g *Game) Init() error {
	if g.initialized {
		return nil
	}

	// Load images
	if err := g.loadImages(); err != nil {
		return err
	}

	// Initialize components
	g.initStarfield()
	g.initScrollText()
	g.preRenderLogoFrames()

	// Initialize audio
	if err := g.loadMusic(); err != nil {
		log.Printf("Failed to load music: %v", err)
		// Continue without music
	}

	g.initialized = true
	return nil
}

// loadMusic loads and plays the MP3 music
func (g *Game) loadMusic() error {
	// Decode MP3
	d, err := mp3.DecodeWithSampleRate(g.audioContext.SampleRate(), bytes.NewReader(musicData))
	if err != nil {
		return fmt.Errorf("failed to decode MP3: %w", err)
	}

	// Create infinite loop
	loop := audio.NewInfiniteLoop(d, d.Length())

	// Create player
	g.audioPlayer, err = g.audioContext.NewPlayer(loop)
	if err != nil {
		return fmt.Errorf("failed to create audio player: %w", err)
	}

	g.audioPlayer.SetVolume(0.5) // Start at 50% volume
	g.audioPlayer.Play()
	return nil
}

// Update updates the game state
func (g *Game) Update() error {
	if !g.initialized {
		return g.Init()
	}

	// Handle input
	if g.audioPlayer != nil {
		// Volume control
		if ebiten.IsKeyPressed(ebiten.KeyUp) {
			vol := g.audioPlayer.Volume() + 0.01
			if vol > 1.0 {
				vol = 1.0
			}
			g.audioPlayer.SetVolume(vol)
		}
		if ebiten.IsKeyPressed(ebiten.KeyDown) {
			vol := g.audioPlayer.Volume() - 0.01
			if vol < 0 {
				vol = 0
			}
			g.audioPlayer.SetVolume(vol)
		}
	}

	// Speed control with +/- keys
	if inpututil.IsKeyJustPressed(ebiten.KeyEqual) || inpututil.IsKeyJustPressed(ebiten.KeyKPAdd) {
		g.speedMultiplier += 0.1
		if g.speedMultiplier > 2.0 {
			g.speedMultiplier = 2.0
		}
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyMinus) || inpututil.IsKeyJustPressed(ebiten.KeyKPSubtract) {
		g.speedMultiplier -= 0.1
		if g.speedMultiplier < 0.5 {
			g.speedMultiplier = 0.5
		}
	}

	// Update splash screen
	if !g.splashDone {
		g.splashTime++
		if g.splashTime >= 100 { // Reduced from 200 for faster splash
			g.splashDone = true
		} else if g.splashTime%3 == 0 && g.splashLine < 360 { // Faster splash drawing
			g.splashLine += 40
		}
		return nil
	}

	// Update starfield
	for _, star := range g.stars {
		star.x += star.speed * g.speedMultiplier
		if star.x > screenWidth {
			star.x = 0
			star.y = float64(rand.Intn(280))
		}
	}

	// Update scroll text
	g.scrollText.x -= scrollSpeed * g.speedMultiplier
	textWidth := float64(len(g.scrollText.text) * g.scrollText.charWidth)
	if g.scrollText.x < -textWidth {
		g.scrollText.x = float64(screenWidth)
	}

	// Update animation counters
	g.vbl++
	g.offsetScr += 0.1 * g.speedMultiplier
	g.offRep += (2.0 / 180.0 * math.Pi) * g.speedMultiplier
	g.offTcb += (7.0 / 180.0 * math.Pi) * g.speedMultiplier

	return nil
}

// drawSplash draws the splash screen
func (g *Game) drawSplash(screen *ebiten.Image) {
	screen.Fill(color.RGBA{0xE0, 0xE0, 0xE0, 0xFF})

	// Draw splash image line by line
	for y := 0; y < g.splashLine; y += 40 {
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(0, float64(y))

		// Draw a 640x40 section
		subImg := g.splash.SubImage(image.Rect(0, y, 640, y+40)).(*ebiten.Image)
		screen.DrawImage(subImg, op)
	}
}

// drawStarfield draws the starfield
func (g *Game) drawStarfield(screen *ebiten.Image) {
	for _, star := range g.stars {
		vector.DrawFilledRect(screen, float32(star.x), float32(star.y),
			float32(star.size), float32(star.size), star.color, false)
	}
}

// charToFontIndex converts a character to its position in the font bitmap
func charToFontIndex(ch rune) (int, bool) {
	// Font layout (6 rows of 10 characters):
	// Row 0: [EMPTY]!"[EMPTY][EMPTY][EMPTY][EMPTY]'()
	// Row 1: [EMPTY][EMPTY],-.[EMPTY]0123
	// Row 2: 4567890:;[EMPTY][EMPTY]
	// Row 3: [EMPTY]?[EMPTY]ABCDEFG
	// Row 4: HIJKLMNOPQ
	// Row 5: RSTUVWXYZ[EMPTY]

	switch ch {
	case '!':
		return 1, true
	case '"':
		return 2, true
	case '\'':
		return 7, true
	case '(':
		return 8, true
	case ')':
		return 9, true
	case ',':
		return 12, true
	case '-':
		return 13, true
	case '.':
		return 14, true
	case '0':
		return 16, true
	case '1':
		return 17, true
	case '2':
		return 18, true
	case '3':
		return 19, true
	case '4':
		return 20, true
	case '5':
		return 21, true
	case '6':
		return 22, true
	case '7':
		return 23, true
	case '8':
		return 24, true
	case '9':
		return 25, true
	case ':':
		return 27, true
	case ';':
		return 28, true
	case '?':
		return 31, true
	case 'A':
		return 33, true
	case 'B':
		return 34, true
	case 'C':
		return 35, true
	case 'D':
		return 36, true
	case 'E':
		return 37, true
	case 'F':
		return 38, true
	case 'G':
		return 39, true
	case 'H':
		return 40, true
	case 'I':
		return 41, true
	case 'J':
		return 42, true
	case 'K':
		return 43, true
	case 'L':
		return 44, true
	case 'M':
		return 45, true
	case 'N':
		return 46, true
	case 'O':
		return 47, true
	case 'P':
		return 48, true
	case 'Q':
		return 49, true
	case 'R':
		return 50, true
	case 'S':
		return 51, true
	case 'T':
		return 52, true
	case 'U':
		return 53, true
	case 'V':
		return 54, true
	case 'W':
		return 55, true
	case 'X':
		return 56, true
	case 'Y':
		return 57, true
	case 'Z':
		return 58, true
	default:
		return 0, false
	}
}

// drawScrollText draws the scrolling text with deformation
func (g *Game) drawScrollText(screen *ebiten.Image) {
	// Clear buffers
	g.scrollText.workBuffer.Clear()
	g.scrollText.deformBuffer.Clear()

	// Draw text to work buffer
	x := g.scrollText.x
	for _, ch := range g.scrollText.text {
		if ch == ' ' {
			x += float64(g.scrollText.charWidth)
			continue
		}

		// Get character position in font
		charIndex, found := charToFontIndex(ch)
		if !found {
			// Character not in font, skip
			x += float64(g.scrollText.charWidth)
			continue
		}

		row := charIndex / g.scrollText.charsPerRow
		col := charIndex % g.scrollText.charsPerRow

		sx := col * g.scrollText.charWidth
		sy := row * g.scrollText.charHeight

		if x > -float64(g.scrollText.charWidth) && x < float64(g.scrollText.workBuffer.Bounds().Dx()) {
			op := &ebiten.DrawImageOptions{}
			op.GeoM.Translate(x, 0)

			subImg := g.scrollText.fontImage.SubImage(
				image.Rect(sx, sy, sx+g.scrollText.charWidth, sy+g.scrollText.charHeight),
			).(*ebiten.Image)

			g.scrollText.workBuffer.DrawImage(subImg, op)
		}

		x += float64(g.scrollText.charWidth)
	}

	// Apply deformation line by line
	for y := 0; y < 25; y++ {
		offsetX := g.scrollX[(g.vbl+y)%g.scrollXMod] + 64

		// Draw each line with horizontal offset
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(-offsetX, 0)

		srcRect := image.Rect(int(offsetX), y*2, int(offsetX)+screenWidth, (y+1)*2)
		if srcRect.Min.X < 0 {
			srcRect.Min.X = 0
		}
		if srcRect.Max.X > g.scrollText.workBuffer.Bounds().Dx() {
			srcRect.Max.X = g.scrollText.workBuffer.Bounds().Dx()
		}

		subImg := g.scrollText.workBuffer.SubImage(srcRect).(*ebiten.Image)

		dstOp := &ebiten.DrawImageOptions{}
		dstOp.GeoM.Translate(0, float64(y*2))
		g.scrollText.deformBuffer.DrawImage(subImg, dstOp)
	}

	// Draw deformed scroll with vertical wave
	for x := 0; x < 40; x++ {
		yOffset := 35 + math.Cos(g.offsetScr+float64(x)*0.1)*35

		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(float64(x*16), 280+yOffset)

		subImg := g.scrollText.deformBuffer.SubImage(
			image.Rect(x*16, 0, (x+1)*16, scrollHeight),
		).(*ebiten.Image)

		screen.DrawImage(subImg, op)
	}
}

// drawLogos draws the bouncing logos
func (g *Game) drawLogos(screen *ebiten.Image) {
	// Calculate logo positions and scales
	zRep := (1 + math.Sin(g.offRep)) / 2
	if zRep < 0 {
		zRep = 0
	}

	yRep := 140 + (math.Cos(g.offRep)+math.Sin(g.offRep))*40
	zTcb := zRep + math.Cos(g.offTcb)/8
	yTcb := yRep + math.Sin(g.offTcb)*(70*zRep)

	// Get frame indices
	repFrame := int(zRep * 35)
	tcbFrame := int(4 + zTcb*32)

	if repFrame >= 35 {
		repFrame = 34
	}
	if tcbFrame >= 40 {
		tcbFrame = 39
	}

	// Draw logos in correct order based on Z
	if zTcb >= zRep {
		// Draw REP first, then TCB
		if repFrame > 0 && g.repFrames[repFrame] != nil {
			op := &ebiten.DrawImageOptions{}
			w := g.repFrames[repFrame].Bounds().Dx()
			h := g.repFrames[repFrame].Bounds().Dy()
			op.GeoM.Translate(float64(320-w/2), yRep-float64(h/2))
			screen.DrawImage(g.repFrames[repFrame], op)
		}

		if tcbFrame > 0 && g.tcbFrames[tcbFrame] != nil {
			op := &ebiten.DrawImageOptions{}
			w := g.tcbFrames[tcbFrame].Bounds().Dx()
			h := g.tcbFrames[tcbFrame].Bounds().Dy()
			op.GeoM.Translate(float64(320-w/2), yTcb-float64(h/2)) // Changed from 225 to 320 to center
			screen.DrawImage(g.tcbFrames[tcbFrame], op)
		}
	} else {
		// Draw TCB first, then REP
		if tcbFrame > 0 && g.tcbFrames[tcbFrame] != nil {
			op := &ebiten.DrawImageOptions{}
			w := g.tcbFrames[tcbFrame].Bounds().Dx()
			h := g.tcbFrames[tcbFrame].Bounds().Dy()
			op.GeoM.Translate(float64(320-w/2), yTcb-float64(h/2)) // Changed from 225 to 320 to center
			screen.DrawImage(g.tcbFrames[tcbFrame], op)
		}

		if repFrame > 0 && g.repFrames[repFrame] != nil {
			op := &ebiten.DrawImageOptions{}
			w := g.repFrames[repFrame].Bounds().Dx()
			h := g.repFrames[repFrame].Bounds().Dy()
			op.GeoM.Translate(float64(320-w/2), yRep-float64(h/2))
			screen.DrawImage(g.repFrames[repFrame], op)
		}
	}
}

// drawSprites draws the bouncing sprites
func (g *Game) drawSprites(screen *ebiten.Image) {
	// Left sprite
	y1 := 326 - math.Abs(math.Cos(g.offsetScr)*24)
	op1 := &ebiten.DrawImageOptions{}
	op1.GeoM.Translate(32, y1)
	screen.DrawImage(g.sprite, op1)

	// Right sprite
	y2 := 326 - math.Abs(math.Sin(g.offsetScr)*24)
	op2 := &ebiten.DrawImageOptions{}
	op2.GeoM.Translate(512, y2)
	screen.DrawImage(g.sprite, op2)
}

// Draw draws the game
func (g *Game) Draw(screen *ebiten.Image) {
	if !g.initialized {
		return
	}

	if !g.splashDone {
		g.drawSplash(screen)
		return
	}

	// Clear screen
	screen.Fill(color.Black)

	// Draw starfield
	g.drawStarfield(screen)

	// Draw logos
	g.drawLogos(screen)

	// Draw scrolling text
	g.drawScrollText(screen)

	// Draw sprites
	g.drawSprites(screen)
}

// Layout returns the game's logical screen size
func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return screenWidth, screenHeight
}

// Cleanup cleans up resources
func (g *Game) Cleanup() {
	if g.audioPlayer != nil {
		g.audioPlayer.Close()
	}
}

func main() {
	ebiten.SetWindowSize(screenWidth*2, screenHeight*2)
	ebiten.SetWindowTitle("TCB-Replicants Demo - Go/Ebiten Port")
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)

	game := NewGame()

	// Ensure cleanup on exit
	defer game.Cleanup()

	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}
