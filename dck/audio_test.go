package tcbreplicants

import (
	"encoding/binary"
	"testing"

	"github.com/olivierh59500/democonstructionkit/sound"
)

func TestMusicStreamReadProducesStereoWithoutAllocating(t *testing.T) {
	player, err := sound.Open("music.ym", ymData, sound.Options{SampleRate: sampleRate, Loop: true, PCMFormat: sound.PCM16, Gain: 0.5})
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := player.Close(); err != nil {
			t.Errorf("Close: %v", err)
		}
	}()

	buffer := make([]byte, 4096*4)
	read := func() {
		n, err := player.Read(buffer)
		if err != nil {
			t.Fatalf("Read: %v", err)
		}
		if n != len(buffer) {
			t.Fatalf("Read bytes = %d, want %d", n, len(buffer))
		}
	}

	read()
	for i := 0; i < len(buffer); i += 4 {
		left := binary.LittleEndian.Uint16(buffer[i : i+2])
		right := binary.LittleEndian.Uint16(buffer[i+2 : i+4])
		if left != right {
			t.Fatalf("frame %d is not mono duplicated to stereo: %d != %d", i/4, left, right)
		}
	}

	if allocations := testing.AllocsPerRun(20, read); allocations != 0 {
		t.Fatalf("Read allocations = %v, want 0", allocations)
	}
}

func BenchmarkMusicStreamRead4096(b *testing.B) {
	player, err := sound.Open("music.ym", ymData, sound.Options{SampleRate: sampleRate, Loop: true, PCMFormat: sound.PCM16, Gain: 0.5})
	if err != nil {
		b.Fatal(err)
	}
	defer func() {
		if err := player.Close(); err != nil {
			b.Errorf("Close: %v", err)
		}
	}()

	buffer := make([]byte, 4096*4)
	b.ReportAllocs()
	b.SetBytes(int64(len(buffer)))
	b.ResetTimer()
	for b.Loop() {
		if _, err := player.Read(buffer); err != nil {
			b.Fatal(err)
		}
	}
}
