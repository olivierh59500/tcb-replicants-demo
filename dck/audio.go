package tcbreplicants

import (
	"log"

	"github.com/olivierh59500/democonstructionkit/sound"

	audio "github.com/olivierh59500/democonstructionkit/sound/output"
)

const sampleRate = 48000

func (g *Game) initAudio() {
	g.audioContext = audio.NewContext(sampleRate)

	music, err := sound.Open("music.ym", ymData, sound.Options{SampleRate: sampleRate, Loop: true, PCMFormat: sound.PCM16, Gain: 0.5})
	if err != nil {
		log.Printf("open music: %v", err)
		return
	}
	g.musicStream = music

	player, err := g.audioContext.NewPlayer(music)
	if err != nil {
		log.Printf("create audio player: %v", err)
		if closeErr := music.Close(); closeErr != nil {
			log.Printf("close music stream: %v", closeErr)
		}
		g.musicStream = nil
		return
	}
	g.audioPlayer = player
}

func (g *Game) startMusic() {
	if g.musicStarted || g.audioPlayer == nil {
		return
	}
	g.audioPlayer.Play()
	g.musicStarted = true
}
