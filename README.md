# TCB-Replicants Demo — Go/Ebitengine port

A port of the classic Atari ST **Weird Dream** intro by TCB and The Replicants.
The same game package runs on desktop and Android.

## Highlights

- sine-deformed bitmap scroll text;
- bouncing TCB and Replicants logos;
- three-layer parallax starfield;
- embedded YM music synthesized at 48 kHz;
- wide-screen Android layout with touch controls in the side areas;
- allocation-free audio callback and cached render resources.

## Requirements

- Go 1.25 or newer;
- Ebitengine 2.9.11;
- `ym-player` revision `v0.0.0-20260913215440-3f73bdca82e5`.

For Android builds, the supplied configuration uses Java 17, Android SDK 36,
NDK 28.2.13676358, Gradle 8.11.1 and Android Gradle Plugin 8.10.1.

## Desktop

Run directly:

```sh
go run ./cmd/tcbreplicants
```

Or build a binary:

```sh
go build -o tcbreplicants ./cmd/tcbreplicants
./tcbreplicants
```

Keyboard controls:

- `↑` / `↓`: volume;
- `+` / `-`: animation speed.

When the desktop window is wider than the original 16:10 scene, the Android
touch controls are also shown and can be previewed with the mouse.

## Android / Pixel

With one authorized arm64 Android device connected over USB:

```sh
./scripts/run-android.sh
```

The script:

1. finds the Android SDK and Java 17;
2. builds `android/app/libs/tcbreplicants.aar` with the Ebitengine 2.9.11 tool;
3. assembles the debug APK;
4. verifies that exactly one device is authorized;
5. installs and launches `com.olivierh.tcbreplicants/.MainActivity`.

The resulting APK is at:

```text
android/app/build/outputs/apk/debug/app-debug.apk
```

The landscape layout keeps the original 640×400 scene centered and uses the
side areas for four multitouch buttons:

- `S+` / `S-`: animation speed;
- `V+` / `V-`: volume.

## Validation

```sh
go test ./...
go test -race ./...
go vet ./...
go test -run '^$' -bench BenchmarkYMPlayerRead4096 -benchmem
```

## Architecture

```text
game.go, audio.go, controls.go   shared game package
cmd/tcbreplicants/              desktop launcher and redundant-draw guard
mobile/                          ebitenmobile bridge
android/                         Java/Gradle Android shell
scripts/run-android.sh           AAR → APK → ADB workflow
assets/                          resources embedded by Go
```

The YM stream uses one 4096-sample mono buffer, writes 16-bit little-endian
stereo PCM directly into Ebitengine's destination buffer, and opens the audio
device only from the first game update so Android's activity is ready.

## Credits

- Original demo by TCB and The Replicants (1989)
- JavaScript version by DrSkull (2015)
- Go/Ebitengine port by Olivier Houte
- YM player library by Olivier Houte

## Optional DCK version

The original implementation remains at its original paths. Run it with `go run ./cmd/tcbreplicants`.

The construction-kit version is in [dck/](dck/README.md). Run `go run ./dck/cmd/tcbreplicants` from this directory. Both versions share the original assets.
