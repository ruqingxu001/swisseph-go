// SPDX-License-Identifier: AGPL-3.0-only
package swisseph

import (
	"encoding/json"
	"math"
	"os"
	"reflect"
	"sync"
	"testing"
)

func TestEveryAPIBoundary(t *testing.T) {
	var defs []struct {
		Name   string `json:"go_name"`
		Params []struct {
			Name, GoType string
			Size         int
		}
	}
	data, err := os.ReadFile("api.json")
	if err != nil {
		t.Fatal(err)
	}
	if err = json.Unmarshal(data, &defs); err != nil {
		t.Fatal(err)
	}
	e := testEngine(t, Config{})
	var expired *Session
	checks := 0
	err = e.WithSession(func(s *Session) error {
		expired = s
		for _, def := range defs {
			method := reflect.ValueOf(s).MethodByName(def.Name)
			if !method.IsValid() {
				t.Fatalf("missing %s", def.Name)
			}
			args := make([]reflect.Value, len(def.Params))
			for i, p := range def.Params {
				switch p.GoType {
				case "float64":
					args[i] = reflect.ValueOf(float64(0))
				case "int32":
					args[i] = reflect.ValueOf(int32(0))
				case "int8":
					args[i] = reflect.ValueOf(int8(0))
				case "string":
					args[i] = reflect.ValueOf("valid")
				case "[]byte":
					args[i] = reflect.ValueOf(make([]byte, p.Size))
				case "[]float64":
					args[i] = reflect.ValueOf(make([]float64, p.Size))
				case "[]int32":
					args[i] = reflect.ValueOf(make([]int32, p.Size))
				default:
					t.Fatalf("unknown type %s", p.GoType)
				}
			}
			for i, p := range def.Params {
				good := args[i]
				var bad []reflect.Value
				switch p.GoType {
				case "float64":
					bad = []reflect.Value{reflect.ValueOf(math.NaN()), reflect.ValueOf(math.Inf(1))}
				case "string":
					bad = []reflect.Value{reflect.ValueOf("bad\x00value")}
				case "[]byte", "[]float64", "[]int32":
					bad = []reflect.Value{good.Slice(0, p.Size-1)}
				}
				for _, v := range bad {
					args[i] = v
					out := method.Call(args)
					if out[len(out)-1].IsNil() {
						t.Errorf("%s failed to reject %s", def.Name, p.Name)
					}
					checks++
				}
				args[i] = good
			}
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	for _, def := range defs {
		method := reflect.ValueOf(expired).MethodByName(def.Name)
		args := make([]reflect.Value, method.Type().NumIn())
		for i := range args {
			args[i] = reflect.Zero(method.Type().In(i))
		}
		out := method.Call(args)
		if out[len(out)-1].IsNil() {
			t.Errorf("expired %s accepted", def.Name)
		}
	}
	if len(defs) != 106 || checks < 300 {
		t.Fatalf("incomplete boundaries %d/%d", len(defs), checks)
	}
	t.Logf("106 APIs reject expired sessions; %d argument boundary checks", checks)
}

func TestSessionPanicCleanupAndIsolation(t *testing.T) {
	e := testEngine(t, Config{SiderealMode: Lahiri, Longitude: 121, Latitude: 31})
	flags := MoshierEphemeris | Speed | Sidereal | Topocentric
	want, err := e.CalcUT(2415020.5, Moon, flags)
	if err != nil {
		t.Fatal(err)
	}
	var escaped *Session
	func() {
		defer func() {
			if recover() == nil {
				t.Error("missing test panic")
			}
		}()
		_ = e.WithSession(func(s *Session) error {
			escaped = s
			for _, err := range []error{s.SetDeltaTUserdef(1), s.SetTidAcc(-20), s.SetInterpolateNut(1), s.SetLapseRate(.001), s.SetSidMode(3, 0, 0), s.SetTopo(-73, 40, 100)} {
				if err != nil {
					t.Fatal(err)
				}
			}
			panic("exercise cleanup")
		})
	}()
	if _, err := escaped.Degnorm(360); err == nil {
		t.Fatal("escaped session live")
	}
	got, err := e.CalcUT(2415020.5, Moon, flags)
	if err != nil || got != want {
		t.Fatalf("session leaked state: %+v != %+v; %v", got, want, err)
	}
	if err := e.WithSession(nil); err == nil {
		t.Fatal("nil callback accepted")
	}
}

func TestConcurrentSessionAndBufferCopyback(t *testing.T) {
	e := testEngine(t, Config{})
	err := e.WithSession(func(s *Session) error {
		var wg sync.WaitGroup
		for i := 0; i < 8; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for j := 0; j < 20; j++ {
					x := make([]float64, 6)
					msg := make([]byte, 256)
					code, err := s.CalcUt(2451545, Sun, MoshierEphemeris|Speed, x, msg)
					if err != nil || code < 0 || math.Abs(x[0]-280.36892) > .0001 {
						t.Errorf("copyback %v %d %v", x, code, err)
						return
					}
				}
			}()
		}
		wg.Wait()
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestDateLastMicrosecondAndHouseCodes(t *testing.T) {
	e := testEngine(t, Config{})
	jd, err := e.JulianDay(2024, 2, 29, 24-1/3.6e9, Gregorian)
	if err != nil || jd != 2460370.5 {
		t.Fatalf("valid date rounded to midnight: %v %v", jd, err)
	}
	p, err := e.Houses(2451545, 0, 31.2304, 121.4737, 'P')
	if err != nil {
		t.Fatal(err)
	}
	for _, sys := range []byte{'X', 'Y'} {
		h, err := e.Houses(2451545, 0, 31.2304, 121.4737, sys)
		if err != nil || h.Cusps == p.Cusps {
			t.Fatalf("%c must use its real house algorithm: %v", sys, err)
		}
	}
}
