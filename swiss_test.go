// SPDX-License-Identifier: AGPL-3.0-only
package swisseph

import (
	"math"
	"strings"
	"sync"
	"testing"
)

func testEngine(t *testing.T, config Config) *Engine {
	t.Helper()
	if config.EphemerisPath == "" {
		config.EphemerisPath = t.TempDir()
	}
	e, err := New(config)
	if err != nil {
		t.Fatal(err)
	}
	return e
}

func TestBasicCalculation(t *testing.T) {
	e := testEngine(t, Config{})
	version, err := e.Version()
	if err != nil || version != "2.10.03" {
		t.Fatalf("version %q %v", version, err)
	}
	jd, err := e.JulianDay(2000, 1, 1, 12, Gregorian)
	if err != nil || jd != 2451545 {
		t.Fatalf("J2000 %v %v", jd, err)
	}
	got, err := e.CalcUT(jd, Sun, MoshierEphemeris|Speed)
	if err != nil {
		t.Fatal(err)
	}
	if got.Flags&7 != MoshierEphemeris || math.Abs(got.Position[0]-280.36892) > 0.0001 || got.Position[3] <= 0 {
		t.Fatalf("unexpected Sun %+v", got)
	}
	houses, err := e.Houses(jd, 0, 31.2304, 121.4737, 'P')
	if err != nil || houses.Count != 12 || houses.Cusps[1] <= 0 {
		t.Fatalf("houses %+v %v", houses, err)
	}
}

func TestFallbackAndErrors(t *testing.T) {
	e := testEngine(t, Config{})
	got, err := e.CalcUT(2451545, Sun, SwissEphemeris|Speed)
	if err == nil || !strings.Contains(err.Error(), "fallback") || got.Flags&7 != MoshierEphemeris || got.Warning == "" {
		t.Fatalf("missing file must be explicit: %+v %v", got, err)
	}
	config := e.config
	config.AllowFallback = true
	allowed := testEngine(t, config)
	if _, err := allowed.CalcUT(2451545, Sun, SwissEphemeris|Speed); err != nil {
		t.Fatal(err)
	}
	if _, err := e.CalcUT(2451545, 1000000000, MoshierEphemeris); err == nil {
		t.Fatal("invalid planet accepted")
	}
	if _, err := e.CalcUT(math.NaN(), Sun, MoshierEphemeris); err == nil {
		t.Fatal("NaN accepted")
	}
	if _, err := e.JulianDay(2023, 2, 29, 0, Gregorian); err == nil {
		t.Fatal("invalid leap date accepted")
	}
	if _, err := e.UTCToJD(UTCDate{2024, 1, 1, 23, 59, 60}, Gregorian); err == nil {
		t.Fatal("invalid leap second accepted")
	}
	if _, err := e.Houses(2451545, 0, 70, 0, 'P'); err == nil {
		t.Fatal("polar fallback must not be hidden")
	}
	if _, err := e.NextEclipse(2451545, MoshierEphemeris, 64, true, false); err == nil {
		t.Fatal("invalid solar eclipse type accepted")
	}
	if _, err := e.Houses(2451545, 0, 0, 0, '?'); err == nil {
		t.Fatal("unknown houses accepted")
	}
	if _, err := New(Config{EphemerisPath: "bad\x00path"}); err == nil {
		t.Fatal("NUL path accepted")
	}
}

func TestConcurrentConfigurationIsolation(t *testing.T) {
	path := t.TempDir()
	engines := []*Engine{
		testEngine(t, Config{EphemerisPath: path, SiderealMode: 0, Longitude: 0, Latitude: 0}),
		testEngine(t, Config{EphemerisPath: path, SiderealMode: Lahiri, Longitude: 121, Latitude: 31, Altitude: 20}),
		testEngine(t, Config{EphemerisPath: path, SiderealMode: 3, Longitude: -73, Latitude: 40, Altitude: 100}),
	}
	flags := MoshierEphemeris | Speed | Sidereal | Topocentric
	want := make([]CalcResult, len(engines))
	for i, e := range engines {
		var err error
		want[i], err = e.CalcUT(2460409.25, Moon, flags)
		if err != nil {
			t.Fatal(err)
		}
	}
	var wg sync.WaitGroup
	for worker := 0; worker < 9; worker++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			idx := i % len(engines)
			for n := 0; n < 20; n++ {
				got, err := engines[idx].CalcUT(2460409.25, Moon, flags)
				if err != nil || got != want[idx] {
					t.Errorf("configuration leaked: %d %+v %v", idx, got, err)
					return
				}
			}
		}(worker)
	}
	wg.Wait()
}

func TestEnvironmentOverrideRejected(t *testing.T) {
	e := testEngine(t, Config{})
	t.Setenv("SE_EPHE_PATH", t.TempDir())
	if _, err := e.CalcUT(2451545, Sun, MoshierEphemeris); err == nil {
		t.Fatal("environment override accepted")
	}
}

func ExampleEngine_CalcUT() {
	engine, err := New(Config{SiderealMode: Lahiri})
	if err != nil {
		panic(err)
	}
	_, err = engine.CalcUT(2451545, Sun, MoshierEphemeris|Speed)
	if err != nil {
		panic(err)
	}
	// For SWIEPH use Config{EphemerisPath: "/absolute/ephe"} and SwissEphemeris|Speed.
}
