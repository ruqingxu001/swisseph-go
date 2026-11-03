//go:build differential

// SPDX-License-Identifier: AGPL-3.0-only
package ephemeris

import (
	"bytes"
	"fmt"
	"math"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

func TestCFloatAgainstNative(t *testing.T) {
	cases := []string{"ICRS", "", "+", "-", ".", "-.", "-0", "-0xjunk", "0x.p1", "-0x.p1", "0x1.8", "0x1.8p+2tail", "1e+bad", "1e309", "-1e-9999", "inf", "-infinity", "nan", "-nan(payload)", " \t\v\f-123.5 suffix"}
	rng := rand.New(rand.NewSource(420))
	for i := 0; i < 1000; i++ {
		v := math.Ldexp(rng.Float64()*2-1, rng.Intn(2000)-1000)
		cases = append(cases, strconv.FormatFloat(v, 'g', 17, 64)+"suffix")
	}
	dir := t.TempDir()
	source := filepath.Join(dir, "atof.c")
	bin := filepath.Join(dir, "atof")
	c := `#include <stdio.h>
#include <stdlib.h>
#include <math.h>
#include <locale.h>
int main(void){char s[1024];setlocale(LC_NUMERIC,"C");while(fgets(s,sizeof(s),stdin)){double v=atof(s);printf("%d %.17g\n",!!signbit(v),v);}return 0;}`
	if err := os.WriteFile(source, []byte(c), 0600); err != nil {
		t.Fatal(err)
	}
	cc := os.Getenv("CC")
	if cc == "" {
		cc = "cc"
	}
	if out, err := exec.Command(cc, source, "-lm", "-o", bin).CombinedOutput(); err != nil {
		t.Fatalf("%v %s", err, out)
	}
	cmd := exec.Command(bin)
	cmd.Stdin = strings.NewReader(strings.Join(cases, "\n") + "\n")
	out, err := cmd.Output()
	if err != nil {
		t.Fatal(err)
	}
	lines := bytes.Split(bytes.TrimSpace(out), []byte("\n"))
	if len(lines) != len(cases) {
		t.Fatal("count")
	}
	for i, line := range lines {
		parts := strings.Fields(string(line))
		negative := parts[0] == "1"
		var want float64
		if strings.Contains(strings.ToLower(parts[1]), "nan") {
			want = math.NaN()
		} else {
			want, err = strconv.ParseFloat(parts[1], 64)
			if err != nil {
				t.Fatal(err)
			}
		}
		got := parseCFloat(cases[i])
		// NaN payload and sign are not specified by C.
		if math.IsNaN(want) {
			if !math.IsNaN(got) {
				t.Errorf("%q: expected NaN", cases[i])
			}
			continue
		}
		if got != want || math.Signbit(got) != negative {
			t.Error(fmt.Sprintf("%q: got %v sign=%v native %s", cases[i], got, math.Signbit(got), line))
		}
	}
	t.Logf("%d numeric prefix comparisons", len(cases))
}
