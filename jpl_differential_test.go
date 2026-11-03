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

func TestJPLFileAgainstC(t *testing.T) {
	file := os.Getenv("SWISSGO_TEST_JPL_FILE")
	if file == "" {
		t.Skip("optional DE200 binary: set SWISSGO_TEST_JPL_FILE")
	}
	file, err := filepath.Abs(file)
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	if err = os.Symlink(file, filepath.Join(dir, "de200.eph")); err != nil {
		t.Fatal(err)
	}
	source := filepath.Join(dir, "oracle.c")
	binary := filepath.Join(dir, "oracle")
	c := `#include "swephexp.h"
#include <stdio.h>
int main(int argc,char **argv){
 if(argc!=2)return 2; swe_set_ephe_path(argv[1]);swe_set_jpl_file("de200.eph");
 double dates[]={2305448.5,2415020.5,2451545.,2460409.25,2500000.5,2400000.5,2460310.5,2350000.5};
 int flags[]={257,257|2048,257|4096,257|16384,257|32|64};
 for(int d=0;d<8;d++)for(int p=0;p<10;p++)for(int f=0;f<5;f++){
  double x[6]={0};char msg[256]={0};int code=swe_calc_ut(dates[d],p,flags[f],x,msg);
  printf("{\"JD\":%.17g,\"Body\":%d,\"Flags\":%d,\"Code\":%d,\"X\":[",dates[d],p,flags[f],code);
  for(int i=0;i<6;i++)printf("%s%.17g",i?",":"",x[i]);puts("]}");
 } swe_close();return 0;
}`
	if err = os.WriteFile(source, []byte(c), 0600); err != nil {
		t.Fatal(err)
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
		t.Fatalf("compile %v %s", err, out)
	}
	out, err := exec.Command(binary, dir).Output()
	if err != nil {
		t.Fatal(err)
	}
	e := testEngine(t, Config{EphemerisPath: dir})
	count := 0
	maxPosition, maxSpeed := 0.0, 0.0
	err = e.WithSession(func(s *Session) error {
		if err := s.SetJplFile("de200.eph"); err != nil {
			return err
		}
		scanner := bufio.NewScanner(bytes.NewReader(out))
		for scanner.Scan() {
			var r struct {
				JD                float64
				Body, Flags, Code int32
				X                 [6]float64
			}
			if err := json.Unmarshal(scanner.Bytes(), &r); err != nil {
				return err
			}
			x := make([]float64, 6)
			code, err := s.CalcUt(r.JD, r.Body, r.Flags, x, make([]byte, 256))
			if err != nil {
				return err
			}
			if code < 0 || code&7 != JPLEphemeris || code != r.Code {
				t.Fatalf("JPL fallback or status mismatch %d vs %d", code, r.Code)
			}
			for i := range x {
				delta := math.Abs(x[i] - r.X[i])
				tol := 1e-8
				if i >= 3 {
					tol = 1e-6
					maxSpeed = math.Max(delta, maxSpeed)
				} else {
					maxPosition = math.Max(delta, maxPosition)
				}
				if !finite(x[i]) || delta > tol {
					t.Errorf("JPL jd=%v body=%d flags=%d field=%d delta=%g", r.JD, r.Body, r.Flags, i, delta)
				}
			}
			count++
		}
		return scanner.Err()
	})
	if err != nil {
		t.Fatal(err)
	}
	if count != 400 {
		t.Fatalf("incomplete JPL samples %d", count)
	}
	t.Logf("400 real DE200 JPL file cases; max position delta=%g, max speed delta=%g", maxPosition, maxSpeed)
}
