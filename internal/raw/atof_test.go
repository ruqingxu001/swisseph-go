// SPDX-License-Identifier: AGPL-3.0-only
package raw

import (
	"math"
	"testing"
)

func TestCFloatPrefix(t *testing.T) {
	for _, tc := range []struct {
		input string
		want  float64
	}{
		{"ICRS", 0}, {"", 0}, {"  -12.5tail", -12.5}, {"+1.25e+2 rest", 125}, {"1e+bad", 1}, {".25", .25}, {"-0x1.8p+2tail", -6}, {"0x10", 16}, {"0xjunk", 0}, {"+", 0}, {"1e9999", math.Inf(1)}, {"-1e-9999", math.Copysign(0, -1)},
	} {
		got := parseCFloat(tc.input)
		if got != tc.want || math.Signbit(got) != math.Signbit(tc.want) {
			t.Errorf("%q: %v != %v", tc.input, got, tc.want)
		}
	}
	if !math.IsNaN(parseCFloat("nan(payload)")) {
		t.Fatal("NaN")
	}
}
