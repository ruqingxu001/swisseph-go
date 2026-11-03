//go:build fluttercompare

// SPDX-License-Identifier: AGPL-3.0-only
package swisseph

import (
	"bufio"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"testing"
	"time"
)

type flutterReference struct {
	Kind, Version, Instant, SystemName string
	JD                                 float64
	Body, Flags, Year                  int32
	System                             byte
	Lat, Lon                           float64
	Solar, Back, Error                 bool
	Input                              []int32
	Value                              json.RawMessage
}

func TestFlutterActualOutputs(t *testing.T) {
	fixtures := []string{os.Getenv("SWISSGO_FLUTTER_REFERENCE"), os.Getenv("SWISSGO_FLUTTER_APP_REFERENCE")}
	if fixtures[0] == "" {
		fixtures[0] = "testdata/flutter/wrapper.jsonl"
	}
	if fixtures[1] == "" {
		fixtures[1] = "testdata/flutter/app.jsonl"
	}
	path := os.Getenv("SWISSGO_TEST_EPHE_PATH")
	if path == "" {
		t.Fatal("SWISSGO_TEST_EPHE_PATH required")
	}
	var manifest struct {
		Fixtures map[string]string `json:"fixtures_sha256"`
		Data     map[string]string `json:"data_sha256"`
	}
	manifestFile := fixtures[0][:len(fixtures[0])-len(filepath.Ext(fixtures[0]))] + ".manifest.json"
	manifestBytes, err := os.ReadFile(manifestFile)
	if err != nil {
		t.Fatal(err)
	}
	if err = json.Unmarshal(manifestBytes, &manifest); err != nil {
		t.Fatal(err)
	}
	for _, file := range fixtures {
		b, er := os.ReadFile(file)
		if er != nil {
			t.Fatal(er)
		}
		if fmt.Sprintf("%x", sha256.Sum256(b)) != manifest.Fixtures[filepath.Base(file)] {
			t.Fatal("fixture hash mismatch", file)
		}
	}
	for file, want := range manifest.Data {
		b, er := os.ReadFile(filepath.Join(path, file))
		if er != nil {
			t.Fatal(er)
		}
		if fmt.Sprintf("%x", sha256.Sum256(b)) != want {
			t.Fatal("ephemeris hash mismatch", file)
		}
	}
	e := testEngine(t, Config{EphemerisPath: path, SiderealMode: Lahiri, AllowFallback: true})
	counts := map[string]int{}
	enumMismatches := map[string]byte{}
	errorsChecked := 0
	maxPosition, maxSpeed, maxDistance, maxRadialSpeed := 0.0, 0.0, 0.0, 0.0
	compare := func(label string, got, want []float64, tol float64) {
		t.Helper()
		if len(got) != len(want) {
			t.Fatalf("%s length %d != %d", label, len(got), len(want))
		}
		for i, x := range got {
			delta := math.Abs(x - want[i])
			if !finite(x) || delta > tol {
				t.Errorf("%s[%d] %.17g != %.17g (delta %.3g)", label, i, x, want[i], delta)
			}
		}
	}
	for _, file := range fixtures {
		f, err := os.Open(file)
		if err != nil {
			t.Fatal(err)
		}
		err = e.WithSession(func(session *Session) error {
			api := flutterSession{session}
			scanner := bufio.NewScanner(f)
			row := 0
			for scanner.Scan() {
				row++
				var r flutterReference
				if err = json.Unmarshal(scanner.Bytes(), &r); err != nil {
					t.Fatalf("%s row %d: %v", file, row, err)
				}
				counts[r.Kind]++
				var gotErr error
				label := fmt.Sprintf("%s row=%d %s", file, row, r.Kind)
				switch r.Kind {
				case "metadata":
					if r.Version != "2.10.03" {
						t.Fatalf("unmatched engine version %s", r.Version)
					}
					continue
				case "calc", "et":
					var result CalcResult
					if r.Kind == "calc" {
						result, gotErr = api.CalcUT(r.JD, r.Body, r.Flags)
					} else {
						result, gotErr = api.CalcET(r.JD, r.Body, r.Flags)
					}
					if !r.Error && gotErr == nil {
						var want []float64
						if err = json.Unmarshal(r.Value, &want); err != nil {
							t.Fatal(err)
						}
						positionTolerance := 1e-8
						if r.Body == OsculatingApogee {
							positionTolerance = 1e-7
						}
						compare(label+" position", result.Position[:3], want[:3], positionTolerance)
						compare(label+" speed", result.Position[3:], want[3:], 1e-6)
						for i := 0; i < 3; i++ {
							if i < 2 {
								maxPosition = math.Max(maxPosition, math.Abs(result.Position[i]-want[i]))
								maxSpeed = math.Max(maxSpeed, math.Abs(result.Position[i+3]-want[i+3]))
							} else {
								maxDistance = math.Max(maxDistance, math.Abs(result.Position[i]-want[i]))
								maxRadialSpeed = math.Max(maxRadialSpeed, math.Abs(result.Position[i+3]-want[i+3]))
							}
						}
					}
				case "houses", "appHouse":
					// Preserve the actual Dart wire value when checking the wrapper. X/Y in
					// this particular Dart package incorrectly encode 58/59, and C falls back
					// to Placidus. Engine.Houses deliberately accepts the correct codes only.
					if r.SystemName != "" && r.System != r.SystemName[0] {
						enumMismatches[r.SystemName] = r.System
					}
					var h HousesResult
					gotErr = func(s *Session) error {
						var er error
						msg := make([]byte, 256)
						h.Status, er = s.HousesEx2(r.JD, r.Flags, r.Lat, r.Lon, int32(r.System), h.Cusps[:], h.Angles[:], h.CuspSpeeds[:], h.AngleSpeeds[:], msg)
						if er != nil {
							return er
						}
						if h.Status < 0 {
							return fmt.Errorf("houses status %d", h.Status)
						}
						return nil
					}(session)
					if !r.Error && gotErr == nil {
						var want struct{ Cusps, Angles, CuspSpeeds, AngleSpeeds []float64 }
						if err = json.Unmarshal(r.Value, &want); err != nil {
							t.Fatal(err)
						}
						compare(label+" cusps", h.Cusps[:len(want.Cusps)], want.Cusps, 1e-8)
						if r.Kind == "appHouse" {
							compare(label+" derived angles", []float64{h.Angles[0], h.Angles[1], math.Mod(h.Angles[0]+180, 360), math.Mod(h.Angles[1]+180, 360), h.Angles[3], h.Angles[4]}, want.Angles, 1e-8)
						} else {
							compare(label+" angles", h.Angles[:], want.Angles, 1e-8)
							compare(label+" cusp speeds", h.CuspSpeeds[:len(want.CuspSpeeds)], want.CuspSpeeds, 1e-6)
							compare(label+" angle speeds", h.AngleSpeeds[:], want.AngleSpeeds, 1e-6)
						}
					}
				case "julday", "appJulday":
					var got float64
					if r.Kind == "julday" {
						got, gotErr = api.JulianDay(r.Year, 3, 1, 12.345, Gregorian)
					} else {
						instant, er := time.Parse(time.RFC3339Nano, r.Instant)
						if er != nil {
							t.Fatal(er)
						}
						hour := float64(instant.Hour()) + float64(instant.Minute())/60 + float64(instant.Second())/3600 + float64(instant.Nanosecond()/1000)/3.6e9
						got, gotErr = api.JulianDay(int32(instant.Year()), int32(instant.Month()), int32(instant.Day()), hour, Gregorian)
					}
					if !r.Error && gotErr == nil {
						var want float64
						json.Unmarshal(r.Value, &want)
						compare(label, []float64{got}, []float64{want}, 1e-9)
					}
				case "revjul":
					date, er := api.ReverseJulianDay(r.JD, Gregorian)
					gotErr = er
					if !r.Error && er == nil {
						hour := math.Floor(date.Hour)
						minutes := (date.Hour - hour) * 60
						minute := math.Floor(minutes)
						seconds := (minutes - minute) * 60
						second := math.Floor(seconds)
						milliseconds := (seconds - second) * 1000
						millisecond := math.Floor(milliseconds)
						microsecond := math.Floor((milliseconds - millisecond) * 1000)
						got := time.Date(int(date.Year), time.Month(date.Month), int(date.Day), int(hour), int(minute), int(second), int(millisecond*1e6+microsecond*1000), time.UTC).UnixMicro()
						var want int64
						json.Unmarshal(r.Value, &want)
						if got != want {
							t.Errorf("%s reversed time %d != %d microseconds", label, got, want)
						}
					}
				case "utc":
					v := r.Input
					got, er := api.UTCToJD(UTCDate{v[0], v[1], v[2], v[3], v[4], float64(v[5])}, Gregorian)
					gotErr = er
					if !r.Error && er == nil {
						var want []float64
						json.Unmarshal(r.Value, &want)
						compare(label, []float64{got.ET, got.UT1}, want, 1e-9)
					}
				case "eclipse":
					typ := int32(84)
					if r.Solar {
						typ = 63
					}
					got, er := api.NextEclipse(r.JD, SwissEphemeris, typ, r.Solar, r.Back)
					gotErr = er
					if !r.Error && er == nil {
						var want struct {
							Times []float64
							Type  int32
						}
						json.Unmarshal(r.Value, &want)
						if got.Type != want.Type {
							t.Errorf("eclipse type %d != %d", got.Type, want.Type)
						}
						compare(label, got.Times[:], want.Times, 1e-8)
					}
				default:
					t.Fatalf("unknown Flutter reference kind %s", r.Kind)
				}
				if (gotErr != nil) != r.Error {
					t.Errorf("%s error mismatch: Dart=%v Go=%v", label, r.Error, gotErr)
				}
				if r.Error {
					errorsChecked++
				}
			}
			if err = scanner.Err(); err != nil {
				t.Fatal(err)
			}
			return nil
		})
		if err != nil {
			t.Fatal(err)
		}
		f.Close()
	}
	if counts["calc"] < 500 || counts["appHouse"] < 100 || counts["metadata"] != 1 {
		t.Fatal("incomplete actual Flutter fixtures")
	}
	if enumMismatches["X"] != 58 || enumMismatches["Y"] != 59 || len(enumMismatches) != 2 {
		t.Fatalf("Flutter enum changed: re-evaluate adapter diagnosis: %v", enumMismatches)
	}
	t.Logf("actual Flutter/Dart rows=%v; error cases=%d; max angle delta %.3g deg; max angular speed delta %.3g deg/day; max distance delta %.3g AU; max radial speed delta %.3g AU/day; known Flutter X/Y encoding defects=%v", counts, errorsChecked, maxPosition, maxSpeed, maxDistance, maxRadialSpeed, enumMismatches)
}

// Match the actual Dart wrapper's continuous C session, including caches and
// its hidden fallback flags. High-level Engine policies have separate C tests.
type flutterSession struct{ s *Session }

func nativeResult(code int32, err error) error {
	if err != nil {
		return err
	}
	if code < 0 {
		return fmt.Errorf("C status %d", code)
	}
	return nil
}
func (a flutterSession) CalcUT(jd float64, body, flags int32) (r CalcResult, err error) {
	r.Flags, err = a.s.CalcUt(jd, body, flags, r.Position[:], make([]byte, 256))
	return r, nativeResult(r.Flags, err)
}
func (a flutterSession) CalcET(jd float64, body, flags int32) (r CalcResult, err error) {
	r.Flags, err = a.s.Calc(jd, body, flags, r.Position[:], make([]byte, 256))
	return r, nativeResult(r.Flags, err)
}
func (a flutterSession) JulianDay(y, m, d int32, h float64, cal int32) (float64, error) {
	return a.s.Julday(y, m, d, h, cal)
}
func (a flutterSession) ReverseJulianDay(jd float64, cal int32) (r CalendarDate, err error) {
	y, m, d, h := make([]int32, 1), make([]int32, 1), make([]int32, 1), make([]float64, 1)
	err = a.s.Revjul(jd, cal, y, m, d, h)
	return CalendarDate{y[0], m[0], d[0], h[0]}, err
}
func (a flutterSession) UTCToJD(d UTCDate, cal int32) (r JulianTimes, err error) {
	v := make([]float64, 2)
	code, err := a.s.UtcToJd(d.Year, d.Month, d.Day, d.Hour, d.Minute, d.Second, cal, v, make([]byte, 256))
	r.ET, r.UT1 = v[0], v[1]
	return r, nativeResult(code, err)
}
func (a flutterSession) NextEclipse(jd float64, flags, typ int32, solar, back bool) (r EclipseResult, err error) {
	b := int32(0)
	if back {
		b = 1
	}
	msg := make([]byte, 256)
	if solar {
		r.Type, err = a.s.SolEclipseWhenGlob(jd, flags, typ, r.Times[:], b, msg)
	} else {
		r.Type, err = a.s.LunEclipseWhen(jd, flags, typ, r.Times[:], b, msg)
	}
	return r, nativeResult(r.Type, err)
}
