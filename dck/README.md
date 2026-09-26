# DCK version

This directory contains the construction-kit version of tcb-replicants-demo. The original Go sources are preserved at their original paths (revision `1e5e55e65d575bdb091d25d5f73b1e5e0fa09258`), with small asset accessors so both versions use the same embedded resources.

Run the original with `go run ./cmd/tcbreplicants` and this version with `go run ./dck/cmd/tcbreplicants` from the repository root.

The choreography and assets remain in this repository. Reusable rendering and
effects come from the published `github.com/olivierh59500/democonstructionkit`
module pinned in `go.mod`. Music is opened with `sound.Open`; DCK selects the decoder from the asset and
provides the configured stereo PCM format. The demo keeps its playback level and loop settings.

The three layers of 35 stars now use `sprites.AnimatedField` and the editable
`presets.ReplicantsStars` recipe. DCK keeps per-star speed, the strict right-edge
wrap, preserved overshoot and a newly sampled height. The three solid-color
materials are created once by `sprites.NewSolidFrames`. The original source
remains at the repository root.

The two logos use `sprites.CoupledLogoPair`. It caches 35 and 40 quantized
sizes once, then selects frames from two linked depth paths and paints the
farther logo first. `presets.ReplicantsLogoPair` retains the original phase
steps, motion amplitudes, frame rules and centered placement while exposing
them for reuse with other artwork.

The splash screen now uses `composite.BlockReveal`: it reveals cached 40-pixel
rows every three ticks, holds the complete image, then hands off at tick 100.
The grid geometry, order, cadence and hold are editable for another image.

The two foreground sprites now use `sprites.Train` with
`presets.ReplicantsBouncingSprites`: one shared rectified wave, a quarter-cycle
phase difference and editable spacing replace the local sine/cosine draw code.
The train now owns its 0.1-step phase and accepts the live speed multiplier;
the demo no longer stores a separate foreground phase. `OwnTime` can be turned
off for a production that supplies an absolute or music-driven `Frame.Time`.
Nine deterministic complete-frame GPU captures around the splash handoff and
through frame 2,400 match the preceding DCK revision pixel for pixel. Run the
opt-in capture with `go test -tags dck_fidelity_rendercheck -run '^$' ./dck`.
