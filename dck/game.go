// Package tcbreplicants implements the TCB-Replicants demo remake.
package tcbreplicants

import (
	"bytes"
	"fmt"
	originalassets "tcb-replicants-demo"

	"github.com/olivierh59500/democonstructionkit/sound"

	"image"
	"image/color"

	kit "github.com/olivierh59500/democonstructionkit"
	"github.com/olivierh59500/democonstructionkit/composite"
	"github.com/olivierh59500/democonstructionkit/scrolling"

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
	starSpeed    = 1.0
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

type Star struct {
	x     float64
	y     float64
	speed float64
	image *ebiten.Image
}

type ScrollText struct {
	renderer      *scrolling.Scrolling
	tiles         []int
	x             float64
	totalWidth    float64
	glyphs        []*ebiten.Image
	charWidth     int
	charHeight    int
	workBuffer    *ebiten.Image
	deformBuffer  *ebiten.Image
	deformRows    []*ebiten.Image
	deformRowMin  int
	deformRowSpan int
	deformColumns []*ebiten.Image
}

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

	rng        *rand.Rand
	stars      []Star
	scrollText *ScrollText
	scrollX    []float64
	scrollXMod int

	vbl        int
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
		stars:           make([]Star, 0, 105),
		speedMultiplier: 1.4,
		layoutWidth:     ScreenWidth,
	}
	g.initScrollX()
	return g
}

func (g *Game) initScrollX() {
	const values = 389 + 120 + 68 + 389 + 36 + 189
	g.scrollX = make([]float64, 0, values)

	stp1 := 7.0 / 180.0 * math.Pi
	stp2 := 3.0 / 180.0 * math.Pi
	for i := 0; i < 389; i++ {
		g.scrollX = append(g.scrollX, 20*math.Sin(float64(i)*stp1)+30*math.Cos(float64(i)*stp2))
	}

	stp1 = 72.0 / 180.0 * math.Pi
	for i := 0; i < 120; i++ {
		g.scrollX = append(g.scrollX, 4*math.Sin(float64(i)*stp1))
	}

	stp1 = 8.0 / 180.0 * math.Pi
	for i := 0; i < 68; i++ {
		g.scrollX = append(g.scrollX, 40*math.Sin(float64(i)*stp1))
	}

	stp1 = 7.0 / 180.0 * math.Pi
	stp2 = 3.0 / 180.0 * math.Pi
	for i := 0; i < 389; i++ {
		g.scrollX = append(g.scrollX, 20*math.Sin(float64(i)*stp1)+30*math.Cos(float64(i)*stp2))
	}

	stp1 = 72.0 / 180.0 * math.Pi
	for i := 0; i < 36; i++ {
		g.scrollX = append(g.scrollX, 4*math.Sin(float64(i)*stp1))
	}

	stp1 = 8.0 / 180.0 * math.Pi
	for i := 0; i < 189; i++ {
		g.scrollX = append(g.scrollX, 30*math.Sin(float64(i)*stp1))
	}

	g.scrollXMod = len(g.scrollX)
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

func (g *Game) initStarfield() {
	params := [...]struct {
		count int
		speed float64
		color color.Color
		size  int
	}{
		{35, 11.2 * starSpeed, color.RGBA{R: 0xE0, G: 0xA0, B: 0xA0, A: 0xFF}, 2},
		{35, 5.6 * starSpeed, color.RGBA{R: 0xC0, G: 0x60, B: 0x60, A: 0xFF}, 2},
		{35, 2.8 * starSpeed, color.RGBA{R: 0x80, G: 0x40, B: 0x40, A: 0xFF}, 2},
	}

	for _, param := range params {
		starImage := ebiten.NewImage(param.size, param.size)
		starImage.Fill(param.color)
		for range param.count {
			g.stars = append(g.stars, Star{
				x:     float64(g.rng.Intn(ScreenWidth)),
				y:     float64(g.rng.Intn(280)),
				speed: param.speed,
				image: starImage,
			})
		}
	}
}

func (g *Game) initScrollText() {
	const text = `      YO, SHITY-FUCKY-LAMEEUUUUURS !!!  AFTER HARD LABOUR, THE MEGAMIGHTY CAREBEARS AND THE FAMOUS REPLICANTS ARE PROUD TO PRESENT    - WEIRD DREAM -   CRACKED BY RATBOY.  THIS INTRO WAS CODED BY NICK, JAS AND AN THE MOTHERFUCKIN COOL AT THE FIRST MEETING TCB - REPLICANTS...    OK, NOW ALL THE MEMBERS OF THE WILL WRITE A PART OF THIS SCROLLTEXT...          HEY, IT'S RATBOY ON THE KEYBOARD, I DON'T KNOW WHAT TO WRITE AND I HATE WRITING SCROLLTEXT.  I'LL TELL YOU MORE DETAILS ABOUT THIS MEETING. TCB ARRIVED FIVE DAYS AGO. SO, THEY DECIDED TO CODE THIS FANTASTIC INTRO. AFTER 25 LITRES OF COKE, 1 MONOPOLY PLAY, 1 BOTTLE OF WHISKY, 20 BIG TOM, SOME PING-PONG MATCHES (OK, JAS !!  YOU'RE BETTER THAN ME, BUT THE REVENGE OF RATBOY WILL BE TERRIBLE !), THIS INTRO IS FINISHED...`

	const (
		charWidth   = 64
		charHeight  = 50
		charsPerRow = 10
	)
	glyphCount := (g.scrollFont.Bounds().Dy() / charHeight) * charsPerRow
	glyphs := make([]*ebiten.Image, glyphCount)
	for index := range glyphs {
		row := index / charsPerRow
		column := index % charsPerRow
		rect := image.Rect(column*charWidth, row*charHeight, (column+1)*charWidth, (row+1)*charHeight)
		glyphs[index] = g.scrollFont.SubImage(rect).(*ebiten.Image)
	}

	tiles := textTiles(text)
	workBuffer := ebiten.NewImage(ScreenWidth+512, scrollHeight)
	deformBuffer := ebiten.NewImage(ScreenWidth, scrollHeight)

	firstOffset := int(g.scrollX[0] + 64)
	minOffset, maxOffset := firstOffset, firstOffset
	for _, value := range g.scrollX {
		offset := int(value + 64)
		minOffset = min(minOffset, offset)
		maxOffset = max(maxOffset, offset)
	}
	rowSpan := maxOffset - minOffset + 1
	rows := make([]*ebiten.Image, charHeight/2*rowSpan)
	for y := 0; y < charHeight/2; y++ {
		for offset := minOffset; offset <= maxOffset; offset++ {
			rect := image.Rect(offset, y*2, offset+ScreenWidth, (y+1)*2)
			rows[y*rowSpan+offset-minOffset] = workBuffer.SubImage(rect).(*ebiten.Image)
		}
	}

	columns := make([]*ebiten.Image, ScreenWidth/16)
	for column := range columns {
		x := column * 16
		columns[column] = deformBuffer.SubImage(image.Rect(x, 0, x+16, scrollHeight)).(*ebiten.Image)
	}

	g.scrollText = &ScrollText{
		tiles:         tiles,
		totalWidth:    float64(len(tiles) * charWidth),
		glyphs:        glyphs,
		charWidth:     charWidth,
		charHeight:    charHeight,
		workBuffer:    workBuffer,
		deformBuffer:  deformBuffer,
		deformRows:    rows,
		deformRowMin:  minOffset,
		deformRowSpan: rowSpan,
		deformColumns: columns,
	}
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
	g.initStarfield()
	g.initScrollText()
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

	for i := range g.stars {
		star := &g.stars[i]
		star.x += star.speed * g.speedMultiplier
		if star.x > ScreenWidth {
			star.x -= ScreenWidth
			star.y = float64(g.rng.Intn(280))
		}
	}

	g.scrollText.x -= scrollSpeed * g.speedMultiplier
	if g.scrollText.x < -g.scrollText.totalWidth {
		g.scrollText.x = ScreenWidth
	}

	g.vbl++
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

func (g *Game) drawStarfield(dst *ebiten.Image) {
	for i := range g.stars {
		star := &g.stars[i]
		var op ebiten.DrawImageOptions
		op.GeoM.Translate(star.x, star.y)
		dst.DrawImage(star.image, &op)
	}
}

func charToFontIndex(ch rune) (int, bool) {
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
	case 'A', 'B', 'C', 'D', 'E', 'F', 'G':
		return 33 + int(ch-'A'), true
	case 'H', 'I', 'J', 'K', 'L', 'M', 'N', 'O', 'P', 'Q':
		return 40 + int(ch-'H'), true
	case 'R', 'S', 'T', 'U', 'V', 'W', 'X', 'Y', 'Z':
		return 50 + int(ch-'R'), true
	default:
		return 0, false
	}
}

func textTiles(text string) []int {
	tiles := make([]int, 0, len(text))
	for _, ch := range text {
		index, ok := charToFontIndex(ch)
		if !ok {
			index = -1
		}
		tiles = append(tiles, index)
	}
	return tiles
}

func (g *Game) drawScrollText(dst *ebiten.Image) {
	st := g.scrollText
	st.workBuffer.Clear()
	st.deformBuffer.Clear()
	if st.renderer == nil {
		images := make([]*ebiten.Image, len(st.tiles))
		for i, tile := range st.tiles {
			if tile >= 0 && tile < len(st.glyphs) {
				images[i] = st.glyphs[tile]
			}
		}
		var err error
		st.renderer, err = scrolling.FromImages(images, float64(st.charWidth))
		if err != nil {
			panic(err)
		}
	}
	state := scrolling.IdentityState()
	state.X = st.x
	state.Map = func(s scrolling.Sample, op *ebiten.DrawImageOptions) bool {
		return s.X > -float64(st.charWidth) && s.X < float64(st.workBuffer.Bounds().Dx())
	}
	st.renderer.DrawAt(st.workBuffer, state)
	frame := kit.Frame{Tick: uint64(g.vbl)}
	composite.Strips{Thickness: 2, Count: st.charHeight / 2, Map: func(i int, r image.Rectangle, f kit.Frame) composite.Strip {
		x := int(g.scrollX[(int(f.Tick)+i)%g.scrollXMod] + 64)
		op := ebiten.DrawImageOptions{}
		op.GeoM.Translate(0, float64(i*2))
		return composite.Strip{Source: image.Rect(x, i*2, x+ScreenWidth, (i+1)*2), Options: op}
	}}.Draw(st.deformBuffer, st.workBuffer, frame)
	composite.Strips{Axis: composite.Columns, Thickness: 16, Count: ScreenWidth / 16, Map: func(i int, r image.Rectangle, f kit.Frame) composite.Strip {
		yOffset := 35 + math.Cos(g.offsetScr+float64(i)*.1)*35
		op := ebiten.DrawImageOptions{}
		op.GeoM.Translate(float64(i*16), 280+yOffset)
		return composite.Strip{Source: r, Options: op}
	}}.Draw(dst, st.deformBuffer, frame)
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
	g.drawStarfield(dst)
	g.drawLogos(dst)
	g.drawScrollText(dst)
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
