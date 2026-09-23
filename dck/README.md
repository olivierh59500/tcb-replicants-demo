# DCK version

This directory contains the construction-kit version of tcb-replicants-demo. The original Go sources are preserved at their original paths (revision `1e5e55e65d575bdb091d25d5f73b1e5e0fa09258`), with small asset accessors so both versions use the same embedded resources.

Run the original with `go run ./cmd/tcbreplicants` and this version with `go run ./dck/cmd/tcbreplicants` from the repository root.

The choreography and assets remain in this repository. Reusable rendering and
effects come from the published `github.com/olivierh59500/democonstructionkit`
module pinned in `go.mod`. Music is opened with `sound.Open`; DCK selects the decoder from the asset and
provides the configured stereo PCM format. The demo keeps its playback level and loop settings.
