//go:build differential

// SPDX-License-Identifier: AGPL-3.0-only
package swisseph

import (
	"bufio"
	"bytes"
	"encoding/json"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

type reference struct {
	Kind                                     string
	JD                                       float64
	Body, Flags, Sid, Status, Calendar, Year int32
	Lat, Lon                                 float64
	System                                   byte
	Solar, Back                              int
	V, A, Cs, As, Input                      []float64
}

func TestNativeDifferential(t *testing.T) {
	path := os.Getenv("SWISSGO_TEST_EPHE_PATH")
	if path == "" {
		t.Fatal("set SWISSGO_TEST_EPHE_PATH to the Flutter ephe directory (required, no skipped file coverage)")
	}
	for _, f := range []string{"sepl_18.se1", "semo_18.se1", "seas_18.se1"} {
		if _, err := os.Stat(filepath.Join(path, f)); err != nil {
			t.Fatal(err)
		}
	}
	binary := filepath.Join(t.TempDir(), "oracle")
	args := []string{"-O2", "-DTLSOFF", "-DNO_SWE_GLP", "-Iupstream", "testdata/oracle.c"}
	for _, n := range []string{"swedate", "swehouse", "swejpl", "swemmoon", "swemplan", "sweph", "swephlib", "swecl", "swehel"} {
		args = append(args, "upstream/"+n+".c")
	}
	args = append(args, "-lm", "-o", binary)
	cc := os.Getenv("CC")
	if cc == "" {
		cc = "cc"
	}
	if out, err := exec.Command(cc, args...).CombinedOutput(); err != nil {
		t.Fatalf("C oracle compile: %v\n%s", err, out)
	}
	output, err := exec.Command(binary, path).Output()
	if err != nil {
		t.Fatal(err)
	}
	scanner := bufio.NewScanner(bytes.NewReader(output))
	count := 0
	fileSuccess := 0
	maxDifference := 0.0
	maxima := map[string]float64{}
	maxCases := map[string]int{}
	violations := 0
	compare := func(label string, got, want []float64, tolerance float64) {
		t.Helper()
		if len(got) != len(want) {
			t.Fatalf("%s lengths", label)
		}
		for i, v := range got {
			diff := math.Abs(v - want[i])
			if math.IsNaN(v) || math.IsInf(v, 0) || diff > tolerance {
				violations++
				if violations <= 40 {
					t.Errorf("case %d %s[%d]: got %.17g want %.17g diff %.3g tolerance %.3g", count, label, i, v, want[i], diff, tolerance)
				}
			}
			if diff > maxima[label] {
				maxima[label] = diff
				maxCases[label] = count
			}
			if diff > maxDifference {
				maxDifference = diff
			}
		}
	}
	for scanner.Scan() {
		count++
		var r reference
		if err := json.Unmarshal(scanner.Bytes(), &r); err != nil {
			t.Fatalf("row %d %v: %s", count, err, scanner.Text())
		}
		e, err := New(Config{EphemerisPath: path, SiderealMode: r.Sid, Longitude: 121.4737, Latitude: 31.2304, Altitude: 20, AllowFallback: true})
		if err != nil {
			t.Fatal(err)
		}
		var actualErr error
		switch r.Kind {
		case "calc", "et":
			var got CalcResult
			if r.Kind == "et" {
				got, actualErr = e.CalcET(r.JD, r.Body, r.Flags)
			} else {
				got, actualErr = e.CalcUT(r.JD, r.Body, r.Flags)
			}
			if got.Flags != r.Status {
				t.Fatalf("case %d flags %d != %d", count, got.Flags, r.Status)
			}
			if got.Flags >= 0 && got.Flags&7 == SwissEphemeris {
				fileSuccess++
			}
			// Osculating apogee is sensitive to floating-point cancellation;
			// topocentric velocity also includes finite-difference observer terms.
			// Keep tighter thresholds for all other bodies/flags.
			positionTolerance, speedTolerance := 1e-8, 1e-6
			if r.Body == OsculatingApogee {
				positionTolerance = 1e-7
			}
			if r.Flags&Topocentric != 0 {
				speedTolerance = 1e-5
			}
			compare("position", got.Position[:3], r.V[:3], positionTolerance)
			compare("speed", got.Position[3:], r.V[3:], speedTolerance)
		case "houses":
			got, err := e.Houses(r.JD, r.Flags, r.Lat, r.Lon, r.System)
			actualErr = err
			if got.Status != r.Status {
				t.Fatalf("case %d house status %d != %d", count, got.Status, r.Status)
			}
			compare("cusps", got.Cusps[:got.Count+1], r.V, 1e-8)
			compare("angles", got.Angles[:], r.A, 1e-8)
			compare("cusp speed", got.CuspSpeeds[:got.Count+1], r.Cs, 1e-6)
			compare("angle speed", got.AngleSpeeds[:], r.As, 1e-6)
		case "date":
			jd, err := e.JulianDay(r.Year, 3, 1, 12.345, r.Calendar)
			if err != nil {
				t.Fatal(err)
			}
			compare("jd", []float64{jd}, []float64{r.JD}, 1e-10)
			got, err := e.ReverseJulianDay(r.JD, r.Calendar)
			actualErr = err
			compare("date", []float64{float64(got.Year), float64(got.Month), float64(got.Day), got.Hour}, r.V, 1e-10)
		case "utc":
			v := r.Input
			got, err := e.UTCToJD(UTCDate{int32(v[0]), int32(v[1]), int32(v[2]), int32(v[3]), int32(v[4]), v[5]}, Gregorian)
			actualErr = err
			compare("utc", []float64{got.ET, got.UT1}, r.V, 1e-9)
		case "eclipse":
			got, err := e.NextEclipse(r.JD, MoshierEphemeris, 0, r.Solar == 1, r.Back == 1)
			actualErr = err
			if got.Type != r.Status {
				t.Fatalf("case %d eclipse type %d != %d", count, got.Type, r.Status)
			}
			compare("eclipse", got.Times[:], r.V, 1e-8)
		default:
			t.Fatalf("unknown reference kind %s", r.Kind)
		}
		if (actualErr != nil) != (r.Status < 0) {
			t.Fatalf("case %d %s status=%d err=%v", count, r.Kind, r.Status, actualErr)
		}
	}
	if err := scanner.Err(); err != nil {
		t.Fatal(err)
	}
	if count < 2000 || fileSuccess < 100 {
		t.Fatalf("insufficient coverage: %d cases, %d file successes", count, fileSuccess)
	}
	t.Logf("maxima by component: %v; violations=%d", maxima, violations)
	t.Logf("maximum case indices: %v", maxCases)
	t.Logf("%d C/Go cases; %d confirmed Swiss-file calculations; max absolute numeric difference %.3g (mixed units)", count, fileSuccess, maxDifference)
}
