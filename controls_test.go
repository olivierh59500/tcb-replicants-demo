package tcbreplicants

import (
	"runtime"
	"testing"
)

func TestLogicalWidth(t *testing.T) {
	tests := []struct {
		name          string
		outsideWidth  int
		outsideHeight int
		want          int
	}{
		{name: "invalid dimensions", want: ScreenWidth},
		{name: "portrait keeps minimum", outsideWidth: 1080, outsideHeight: 2424, want: ScreenWidth},
		{name: "Pixel 10a landscape", outsideWidth: 2424, outsideHeight: 1080, want: 898},
		{name: "ultrawide is capped", outsideWidth: 4000, outsideHeight: 1000, want: maxLogicalWidth},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := logicalWidth(tt.outsideWidth, tt.outsideHeight); got != tt.want {
				t.Fatalf("logicalWidth(%d, %d) = %d, want %d", tt.outsideWidth, tt.outsideHeight, got, tt.want)
			}
		})
	}
}

func TestWideControlLayoutUsesSideAreas(t *testing.T) {
	const width = 898
	layout := makeControlLayout(width, ScreenHeight)
	sceneLeft := float64(width-ScreenWidth) / 2
	sceneRight := sceneLeft + ScreenWidth

	for _, button := range []controlButton{layout.SpeedUp, layout.SpeedDown} {
		if button.X-button.Radius < 0 || button.X+button.Radius > sceneLeft {
			t.Fatalf("speed button is outside the left side area: %#v", button)
		}
	}
	for _, button := range []controlButton{layout.VolumeUp, layout.VolumeDown} {
		if button.X-button.Radius < sceneRight || button.X+button.Radius > width {
			t.Fatalf("volume button is outside the right side area: %#v", button)
		}
	}
}

func TestControlStateSupportsMultitouch(t *testing.T) {
	layout := makeControlLayout(898, ScreenHeight)
	state := controlState{}
	state.press(layout, int(layout.SpeedUp.X), int(layout.SpeedUp.Y))
	state.press(layout, int(layout.VolumeDown.X), int(layout.VolumeDown.Y))

	if !state.SpeedUp || state.SpeedDown || state.VolumeUp || !state.VolumeDown {
		t.Fatalf("unexpected control state: %#v", state)
	}
}

func TestVirtualControlsVisibilityOnDesktop(t *testing.T) {
	if runtime.GOOS == "android" || runtime.GOOS == "ios" {
		t.Skip("desktop-only assertion")
	}
	game := &Game{layoutWidth: ScreenWidth}
	if game.virtualControlsVisible() {
		t.Fatal("controls must stay hidden at the native desktop aspect ratio")
	}
	game.layoutWidth++
	if !game.virtualControlsVisible() {
		t.Fatal("controls should be available for desktop preview in a wide window")
	}
}
