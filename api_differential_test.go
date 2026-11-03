//go:build differential

// SPDX-License-Identifier: AGPL-3.0-only
package swisseph

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

type apiParam struct {
	Name, GoType string
	Pointer      bool
	Size         int
}
type apiDefinition struct {
	CName  string `json:"c_name"`
	GoName string `json:"go_name"`
	Return string
	Params []apiParam
}
type apiStep struct {
	Function string
	Args     map[string]any
}
type apiCase struct {
	ID    string
	Steps []apiStep
	// Additional velocity samples use the Engine suite's established bound.
	// The original 256 cases retain their stricter field-specific tolerances.
	VelocityTolerance float64 `json:"velocity_tolerance,omitempty"`
}
type apiReference struct {
	ID, Function string
	Step         int
	Result       any
	Outputs      map[string]any
}

func loadJSON(t *testing.T, path string, v any) {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if err = json.Unmarshal(data, v); err != nil {
		t.Fatal(err)
	}
}

func TestAllPublicAPIsAgainstC(t *testing.T) {
	path := os.Getenv("SWISSGO_TEST_EPHE_PATH")
	if path == "" {
		t.Fatal("SWISSGO_TEST_EPHE_PATH required")
	}
	var defs []apiDefinition
	loadJSON(t, "api.json", &defs)
	byName := map[string]apiDefinition{}
	for _, a := range defs {
		byName[a.CName] = a
	}
	var cases []apiCase
	loadJSON(t, "testdata/api_cases.json", &cases)
	tmp := t.TempDir()
	source := filepath.Join(tmp, "api_oracle.c")
	binary := filepath.Join(tmp, "api_oracle")
	if out, err := exec.Command("python3", "tools/api_oracle.py", source).CombinedOutput(); err != nil {
		t.Fatalf("generate oracle %v %s", err, out)
	}
	args := []string{"-O2", "-DTLSOFF", "-DNO_SWE_GLP", "-Iupstream", source}
	for _, n := range []string{"swedate", "swehouse", "swejpl", "swemmoon", "swemplan", "sweph", "swephlib", "swecl", "swehel"} {
		args = append(args, "upstream/"+n+".c")
	}
	args = append(args, "-lm", "-o", binary)
	cc := os.Getenv("CC")
	if cc == "" {
		cc = "cc"
	}
	if out, err := exec.Command(cc, args...).CombinedOutput(); err != nil {
		t.Fatalf("compile reference %v %s", err, out)
	}
	output, err := exec.Command(binary, path).Output()
	if err != nil {
		t.Fatalf("run C oracle %v", err)
	}
	refs := map[string]apiReference{}
	scanner := bufio.NewScanner(bytes.NewReader(output))
	scanner.Buffer(make([]byte, 4096), 1<<20)
	for scanner.Scan() {
		var r apiReference
		if err := json.Unmarshal(scanner.Bytes(), &r); err != nil {
			t.Fatalf("oracle JSON %v: %s", err, scanner.Text())
		}
		refs[fmt.Sprintf("%s/%d", r.ID, r.Step)] = r
	}
	if err := scanner.Err(); err != nil {
		t.Fatal(err)
	}
	seen := map[string]int{}
	maxDelta := map[string]float64{}
	for _, tc := range cases {
		t.Run(tc.ID, func(t *testing.T) {
			e := testEngine(t, Config{EphemerisPath: path, AllowFallback: true})
			err := e.WithSession(func(s *Session) error {
				for i, step := range tc.Steps {
					a := byName[step.Function]
					method := reflect.ValueOf(s).MethodByName(a.GoName)
					if !method.IsValid() {
						t.Fatalf("missing Go API %s", a.GoName)
					}
					in := []reflect.Value{}
					outputs := map[string]any{}
					for _, p := range a.Params {
						v := step.Args[p.Name]
						var arg any
						switch p.GoType {
						case "float64":
							arg = v.(float64)
						case "int32":
							arg = int32(v.(float64))
						case "int8":
							arg = int8(v.(float64))
						case "string":
							str := v.(string)
							if str == "@ephe" {
								str = path
							}
							arg = str
						case "[]byte":
							b := make([]byte, p.Size)
							copy(b, v.(string))
							arg = b
							outputs[p.Name] = b
						case "[]float64":
							arr := v.([]any)
							b := make([]float64, len(arr))
							for j, x := range arr {
								b[j] = x.(float64)
							}
							arg = b
							outputs[p.Name] = b
						case "[]int32":
							arr := v.([]any)
							b := make([]int32, len(arr))
							for j, x := range arr {
								b[j] = int32(x.(float64))
							}
							arg = b
							outputs[p.Name] = b
						default:
							t.Fatalf("unknown ABI type %s", p.GoType)
						}
						in = append(in, reflect.ValueOf(arg))
					}
					out := method.Call(in)
					if last := out[len(out)-1]; !last.IsNil() {
						return fmt.Errorf("%s: %v", a.CName, last.Interface())
					}
					var result any
					if a.Return != "void" {
						result = out[0].Interface()
					}
					ref, ok := refs[fmt.Sprintf("%s/%d", tc.ID, i)]
					if !ok {
						t.Fatal("missing native reference")
					}
					compareAPI(t, a.CName, "return", result, ref.Result, maxDelta)
					for name, v := range outputs {
						if b, ok := v.([]byte); ok {
							end := bytes.IndexByte(b, 0)
							if end < 0 {
								t.Fatal("unterminated output")
							}
							v = string(b[:end])
						}
						compareAPIWithVelocity(t, a.CName, name, v, ref.Outputs[name], maxDelta, tc.VelocityTolerance)
					}
					seen[a.CName]++
				}
				return nil
			})
			if err != nil {
				t.Fatal(err)
			}
		})
	}
	for _, a := range defs {
		if seen[a.CName] == 0 {
			t.Errorf("untested public API: %s", a.CName)
		}
	}
	t.Logf("API coverage: %d/%d; cases=%d; max deltas=%v", len(seen), len(defs), len(cases), maxDelta)
}

func compareAPI(t *testing.T, function, field string, got, want any, maxDelta map[string]float64) {
	t.Helper()
	compareAPIWithVelocity(t, function, field, got, want, maxDelta, 0)
}

func compareAPIWithVelocity(t *testing.T, function, field string, got, want any, maxDelta map[string]float64, velocityTolerance float64) {
	t.Helper()
	if got == nil || want == nil {
		if got != want {
			t.Errorf("%s %s: %v != %v", function, field, got, want)
		}
		return
	}
	if str, ok := want.(string); ok {
		if got != str {
			t.Errorf("%s %s: %q != %q", function, field, got, str)
		}
		return
	}
	gv := reflect.ValueOf(got)
	if gv.Kind() == reflect.Slice {
		ws := want.([]any)
		if gv.Len() != len(ws) {
			t.Fatalf("output length mismatch")
		}
		for i := 0; i < gv.Len(); i++ {
			compareAPIWithVelocity(t, function, fmt.Sprintf("%s[%d]", field, i), gv.Index(i).Interface(), ws[i], maxDelta, velocityTolerance)
		}
		return
	}
	var number float64
	switch x := got.(type) {
	case float64:
		number = x
	case int32:
		number = float64(x)
	default:
		t.Fatalf("unexpected value %T", got)
	}
	expected := want.(float64)
	diff := math.Abs(number - expected)
	// Strict absolute tolerance for coordinates; adaptive relative term for
	// large heliacal quantities. Integral statuses and strings remain exact.
	tolerance := 1e-8
	if velocityTolerance > 0 && function == "swe_calc" && (field == "xx[3]" || field == "xx[4]" || field == "xx[5]") {
		tolerance = velocityTolerance
	}
	// Radial stellar speed differentiates very large distances; native libm
	// and Go rounding are amplified. The bound is AU/day, not angular degrees.
	if (function == "swe_fixstar_ut" || function == "swe_fixstar2_ut") && field == "xx[5]" {
		tolerance = 1e-5
	}
	// Occultation attributes mix km and angular quantities; include a relative
	// term for shadow dimensions and a sub-milliarcsecond angular allowance.
	if function == "swe_lun_occult_when_loc" && strings.HasPrefix(field, "attr[") {
		tolerance = 3e-7 + math.Abs(expected)*1e-10
	}
	if strings.Contains(function, "heliacal") || strings.Contains(function, "vis_limit") || strings.Contains(function, "arcus") {
		tolerance = 1e-7 + math.Abs(expected)*1e-10
	}
	if _, ok := got.(int32); ok {
		tolerance = 0
	}
	if diff > maxDelta[function] {
		maxDelta[function] = diff
	}
	if !finite(number) || diff > tolerance {
		t.Errorf("%s %s: %.17g != %.17g delta=%.3g tol=%.3g", function, field, number, expected, diff, tolerance)
	}
}
