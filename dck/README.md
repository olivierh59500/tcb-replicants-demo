# DCK version

This directory contains the construction-kit version of tcb-replicants-demo. The original Go sources are preserved at their original paths (revision `1e5e55e65d575bdb091d25d5f73b1e5e0fa09258`), with small asset accessors so both versions use the same embedded resources.

Run the original with `go run ./cmd/tcbreplicants` and this version with `go run ./dck/cmd/tcbreplicants` from the repository root.

The choreography and assets stay local; reusable rendering and effects live in `../../lib/democonstructionkit`. Second Reality retains its original ST3 music synchronization.
