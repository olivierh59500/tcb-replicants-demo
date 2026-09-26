package tcbreplicants

import (
	"math"
	"math/rand"
	"testing"

	"github.com/olivierh59500/democonstructionkit/motion"
	"github.com/olivierh59500/democonstructionkit/presets"
)

func TestLayeredStarsPreserveSeededPositionsAndWraps(t *testing.T) {
	const seed = 1989
	oldRandom, newRandom := rand.New(rand.NewSource(seed)), rand.New(rand.NewSource(seed))
	config, err := presets.ReplicantsStars(nil, presets.DefaultReplicantsStarOptions(newRandom.Intn))
	if err != nil {
		t.Fatal(err)
	}
	field, err := motion.NewFrameField(config.Motion)
	if err != nil {
		t.Fatal(err)
	}
	type oldStar struct {
		x, y, speed float64
		image       int
	}
	legacy := make([]oldStar, 0, 105)
	for layer, speed := range []float64{11.2, 5.6, 2.8} {
		for range 35 {
			legacy = append(legacy, oldStar{x: float64(oldRandom.Intn(ScreenWidth)),
				y: float64(oldRandom.Intn(280)), speed: speed, image: layer})
		}
	}
	check := func(tick int) {
		t.Helper()
		for i, particle := range field.Samples() {
			want := legacy[i]
			if math.Abs(particle.X-want.x) > 1e-12 || particle.Y != want.y ||
				particle.VelocityX != want.speed || particle.Image != want.image {
				t.Fatalf("tick %d star %d: %+v, want %+v", tick, i, particle, want)
			}
		}
	}
	check(0)
	for tick := 1; tick <= 3000; tick++ {
		multiplier := 1.4
		if tick >= 1000 {
			multiplier = .8
		}
		if tick >= 2000 {
			multiplier = 2
		}
		if err := field.SetSpeedMultiplier(multiplier); err != nil {
			t.Fatal(err)
		}
		for i := range legacy {
			legacy[i].x += legacy[i].speed * multiplier
			if legacy[i].x > ScreenWidth {
				legacy[i].x -= ScreenWidth
				legacy[i].y = float64(oldRandom.Intn(280))
			}
		}
		if err := field.Step(); err != nil {
			t.Fatal(err)
		}
		check(tick)
	}
}
