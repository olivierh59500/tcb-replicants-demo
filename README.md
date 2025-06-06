# TCB-Replicants Demo - Go/Ebiten Port

A faithful port of the classic Atari ST demo "Weird Dream" by TCB-Replicants to Go using the Ebiten game engine.

## Features

- **Classic scrolling text** with sine wave deformation effects
- **Bouncing logos** with 3D depth simulation
- **Multi-layer starfield** with parallax scrolling
- **YM music playback** using the ym-player library
- **Smooth animations** optimized for modern systems

## Requirements

- Go 1.19 or higher
- Ebiten v2
- ym-player library for YM/SNDH music playback

## Installation

1. Clone the repository:
```bash
git clone https://github.com/yourusername/tcb-replicants-demo
cd tcb-replicants-demo
```

2. Install dependencies:
```bash
go get github.com/hajimehoshi/ebiten/v2
go get github.com/olivierh59500/ym-player
```

3. Place the required assets in the `assets/` directory:
   - `union_sprite.png` - Bouncing sprite image
   - `tcb_logo.png` - TCB logo
   - `rep_logo.png` - Replicants logo
   - `tcb_rep_font.png` - Bitmap font for scrolling text (640x300 pixels, 6 rows of 10 characters, 64x50 pixels each)
   - `tcb_rep_splash.png` - Splash screen image
   - `Rollout.ym` - YM music file

## Building and Running

```bash
go build -o demo
./demo
```

Or simply:
```bash
go run main.go
```

## Controls

- **F1** - Switch to song 1
- **F2** - Switch to song 2
- **Up/Down** - Adjust volume
- **ESC** - Exit demo

## Technical Details

### Optimizations

1. **Pre-rendered logo frames**: Instead of scaling logos in real-time, all scale variations are pre-rendered during initialization for better performance.

2. **Efficient scroll buffer**: Uses a wider work buffer to handle text deformation without clipping at screen edges.

3. **Optimized starfield**: Simple rectangle drawing instead of image blitting for stars.

4. **Frame-based animation**: All animations are tied to a virtual blanking (VBL) counter for consistent timing.

### Key Differences from JavaScript Version

- **Slower animation speed**: The original JavaScript version runs very fast on modern browsers. This port intentionally slows down scrolling and bounce effects for better visibility.

- **Wider scroll buffer**: Ensures characters don't disappear at screen edges during deformation.

- **Simplified audio**: Uses the ym-player library for direct YM/SNDH playback instead of web audio APIs.

## Architecture

```
main.go
├── Game struct - Main game state
├── Star struct - Starfield stars
├── ScrollText struct - Scrolling text manager
├── Logo struct - Bouncing logo state
└── Various animation functions
```

### Main Components

1. **Splash Screen**: Displays the intro image line by line
2. **Starfield**: Three layers of stars with different speeds and colors
3. **Scrolling Text**: Deformed text with sine wave effects
4. **Bouncing Logos**: TCB and Replicants logos with 3D scaling
5. **Sprites**: Union sprites bouncing at screen edges

## Credits

- Original demo by TCB-Replicants (1989)
- JavaScript version by DrSkull (2015)
- Go/Ebiten port by Olivier Houte
- YM player library by olivierh59500

## License

This port respects the original demo's legacy. The original demo code was released under the MIT License. Please check individual asset licenses.

## Notes

- The demo requires proper YM/SNDH files for music playback
- Performance may vary depending on system capabilities
- The deformation effects are computationally intensive but optimized for modern hardware
