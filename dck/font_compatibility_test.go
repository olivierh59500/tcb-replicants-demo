package tcbreplicants

import (
	"github.com/olivierh59500/democonstructionkit/presets"
	"testing"
)

func legacyAtlasIndex(ch rune) (int, bool) {
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
func TestSharedAtlasIndicesMatchOriginalAlphabet(t *testing.T) {
	lookup, err := presets.TileLookup("tcb-replicants-demo", true)
	if err != nil {
		t.Fatal(err)
	}
	for r := rune(0); r < 256; r++ {
		got, ok := lookup(r)
		want, found := legacyAtlasIndex(r)
		if got != want || ok != found {
			t.Fatalf("rune %U: got (%d,%t), want (%d,%t)", r, got, ok, want, found)
		}
	}
}
