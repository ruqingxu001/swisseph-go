//go:build differential

// SPDX-License-Identifier: AGPL-3.0-only
package ephemeris

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// These oracles include one pinned translation unit so that its private
// functions can be tested directly. The remaining units are linked normally.
// No C code is part of the Go runtime or used to produce the Go result.
func TestPrivateModulesAgainstC(t *testing.T) {
	modules := []struct{ source, body string }{
		{"swephlib", `
for(int i=0;i<7;i++){double a=-6.0+i*2.0;row("angle",&a,1);double r=swi_angnorm(a);values(&r,1);}
for(int i=0;i<4;i++){double a=2415020.5+i*12000;double x[6]={0.3,-0.4,0.5,0.001,-0.002,0.003};row("fk4",&a,1);swi_FK4_FK5(x,a);values(x,6);}
for(int i=0;i<4;i++){double a=2415020.5+i*12000;double x[6]={0.3,-0.4,0.5,0.001,-0.002,0.003};row("fk5",&a,1);swi_FK5_FK4(x,a);values(x,6);}
for(int i=0;i<2;i++){double a=i;double x[6]={0.3,-0.4,0.5,0.001,-0.002,0.003};row("icrs",&a,1);swi_icrs2fk5(x,256,i);values(x,6);}
for(int i=0;i<15;i++){double a=i/2.0,v[8]={0.2,0.9,-0.4,0.7,1.2,-0.1,0.6,0.8};row("bessel",&a,1);double r=bessel(v,8,a);values(&r,1);}
`},
		{"swehouse", `
for(int i=0;i<12;i++){double a=i*30.0;row("armc",&a,1);double r=swi_armc_to_mc(a,23.4392911);values(&r,1);row("armcPrivate",&a,1);r=armc_to_mc(a,23.4392911);values(&r,1);}
for(int i=0;i<5;i++){double a=-80+i*40.0;row("polar",&a,1);double r=fix_asc_polar(80,123.45,23.4392911,a);values(&r,1);}
`},
		{"swemmoon", `
for(int i=0;i<6;i++){double a=2300000.5+i*30000;row("elements",&a,1);double r[4];swi_mean_lunar_elements(a,r,r+1,r+2,r+3);values(r,4);}
for(int k=4;k<=5;k++)for(int i=0;i<4;i++){double a[2]={2415020.5+i*15000,k},r[3];row("apsides",a,2);swi_intp_apsides(a[0],r,k);values(r,3);}
`},
		{"swehel", `
for(int i=0;i<6;i++){double a=i*10.0;row("hourAngle",&a,1);double r=HourAngle(a,12.5,31.23);values(&r,1);row("airmass",&a,1);r=Airmass(a,1013.25);values(&r,1);}
for(int i=0;i<5;i++){double a=5+i*3.0;row("width",&a,1);double r=WidthMoon(a,100,-5,110,0.95);values(&r,1);row("length",&a,1);r=LengthMoon(a/10,0.5);values(&r,1);row("yallop",&a,1);r=qYallop(a/10,15);values(&r,1);}
{double a=2451545.0,dgeo[3]={8.55,47.37,408},datm[4]={1013.25,15,40,0.25},r[3]={0};char msg[256],name[256]="Sirius";row("azalt",&a,1);azalt_cart(a,dgeo,datm,name,2,r,msg);values(r,3);}
{double a=2460310.5,dgeo[3]={8.55,47.37,408},datm[4]={1013.25,15,40,0.25},dobs[6]={36,1,1,1,0,0},r[2]={0};char msg[256],name[256]="sirius";row("acronychal",&a,1);r[0]=get_acronychal_day(a,dgeo,datm,dobs,name,2|2048,5,r+1,msg);values(r,2);}
{double a=2460310.5,dgeo[3]={8.55,47.37,408},r[2]={0};char msg[256];row("riseHelper",&a,1);r[0]=call_swe_rise_trans(a,3,"",2,1,dgeo,1013.25,15,r+1,msg);values(r,2);}
`},
		{"swemplan", `
const char *terms[]={"42.5","1+2*T","3.5-0.1*T2","-2*T3+0.01*T4","0.25*T0"};
for(int i=0;i<5;i++){double a=i,out;char term[256];strcpy(term,terms[i]);row("terms",&a,1);check_t_terms(3652.5,term,&out);values(&out,1);}
for(int i=0;i<15;i++){double a=i,out[256];char name[256];swi_get_fict_name(i,name);row("fictName",&a,1);int n=strlen(name);for(int j=0;j<n;j++)out[j]=(unsigned char)name[j];values(out,n);}
`},
		{"swecl", `
for(int i=0;i<8;i++){double a=i*1024.0;row("twilight",&a,1);double r=rdi_twilight((int)a);values(&r,1);}
`},
		{"sweph", `
for(int i=0;i<14;i++){double a=-0.1+i*0.1;row("effectiveMass",&a,1);double r=meff(a);values(&r,1);}
{double a=0,r[7];swe_set_ephe_path(argv[2]);load_dpsi_deps();row("eop",&a,1);r[0]=swed.eop_dpsi_loaded;r[1]=swed.eop_tjd_beg;r[2]=swed.eop_tjd_end;r[3]=swed.dpsi[0];r[4]=swed.deps[0];r[5]=swed.dpsi[2];r[6]=swed.deps[2];values(r,7);}
`},
	}
	for _, module := range modules {
		t.Run(module.source, func(t *testing.T) {
			tmp := t.TempDir()
			writeEOPSample(t, tmp)
			source, binaryPath := filepath.Join(tmp, "oracle.c"), filepath.Join(tmp, "oracle")
			program := `#include <stdio.h>
#include <string.h>
#include "` + module.source + `.c"
static void row(const char *name,double *v,int n){printf("{\"function\":\"%s\",\"input\":[",name);for(int i=0;i<n;i++)printf("%s%.17g",i?",":"",v[i]);printf("],\"value\":[");}
static void values(double *v,int n){for(int i=0;i<n;i++)printf("%s%.17g",i?",":"",v[i]);puts("]}");}
int main(int argc,char **argv){if(argc!=3)return 2;swe_set_ephe_path(argv[1]);` + module.body + `swe_close();return 0;}`
			if err := os.WriteFile(source, []byte(program), 0600); err != nil {
				t.Fatal(err)
			}
			args := []string{"-O2", "-DTLSOFF", "-DNO_SWE_GLP", "-I../../upstream", source}
			for _, unit := range []string{"swedate", "swehouse", "swejpl", "swemmoon", "swemplan", "sweph", "swephlib", "swecl", "swehel"} {
				if unit != module.source {
					args = append(args, "../../upstream/"+unit+".c")
				}
			}
			args = append(args, "-lm", "-o", binaryPath)
			cc := os.Getenv("CC")
			if cc == "" {
				cc = "cc"
			}
			if out, err := exec.Command(cc, args...).CombinedOutput(); err != nil {
				t.Fatalf("compile private C oracle: %v\n%s", err, out)
			}
			path, err := filepath.Abs("../../ephe")
			if err != nil {
				t.Fatal(err)
			}
			output, err := exec.Command(binaryPath, path, tmp).Output()
			if err != nil {
				t.Fatal(err)
			}
			scanner := bufio.NewScanner(bytes.NewReader(output))
			state := NewState()
			defer state.Close()
			state.SetEphePath(path)
			if module.source == "sweph" {
				state.SetEphePath(tmp)
			}
			count := 0
			for scanner.Scan() {
				var row struct {
					Function     string
					Input, Value []float64
				}
				if err := json.Unmarshal(scanner.Bytes(), &row); err != nil {
					t.Fatalf("%v: %s", err, scanner.Text())
				}
				got := privateResult(state, row.Function, row.Input)
				if len(got) != len(row.Value) {
					t.Fatalf("%s: output length", row.Function)
				}
				for i, want := range row.Value {
					if math.IsNaN(got[i]) || math.Abs(got[i]-want) > 1e-10+math.Abs(want)*1e-12 {
						t.Errorf("%s(%v)[%d]: %.17g != %.17g", row.Function, row.Input, i, got[i], want)
					}
				}
				count++
			}
			if err := scanner.Err(); err != nil {
				t.Fatal(err)
			}
			if count == 0 {
				t.Fatal("empty oracle")
			}
			t.Logf("%d private module C/Go comparisons", count)
		})
	}
}

func privateResult(state *State, function string, input []float64) []float64 {
	x := []float64{0.3, -0.4, 0.5, 0.001, -0.002, 0.003}
	a := input[0]
	switch function {
	case "angle":
		return []float64{state.swiAngnorm(a)}
	case "fk4":
		state.swiFK4FK5(x, a)
		return x
	case "fk5":
		state.swiFK5FK4(x, a)
		return x
	case "icrs":
		state.swiIcrs2fk5(x, 256, int32(a))
		return x
	case "bessel":
		return []float64{state.bessel([]float64{0.2, 0.9, -0.4, 0.7, 1.2, -0.1, 0.6, 0.8}, 8, a)}
	case "armc":
		return []float64{state.swiArmcToMc(a, 23.4392911)}
	case "armcPrivate":
		return []float64{state.armcToMc(a, 23.4392911)}
	case "polar":
		return []float64{state.fixAscPolar(80, 123.45, 23.4392911, a)}
	case "elements":
		r := make([]float64, 4)
		state.swiMeanLunarElements(a, r, r[1:], r[2:], r[3:])
		return r
	case "apsides":
		r := make([]float64, 3)
		state.swiIntpApsides(a, r, int32(input[1]))
		return r
	case "hourAngle":
		return []float64{state.HourAngle(a, 12.5, 31.23)}
	case "airmass":
		return []float64{state.Airmass(a, 1013.25)}
	case "width":
		return []float64{state.WidthMoon(a, 100, -5, 110, 0.95)}
	case "length":
		return []float64{state.LengthMoon(a/10, 0.5)}
	case "yallop":
		return []float64{state.qYallop(a/10, 15)}
	case "azalt":
		r := make([]float64, 3)
		state.azaltCart(a, []float64{8.55, 47.37, 408}, []float64{1013.25, 15, 40, 0.25}, textBytes("Sirius"), 2, r, make([]byte, 256))
		return r
	case "terms":
		r := make([]float64, 1)
		terms := []string{"42.5", "1+2*T", "3.5-0.1*T2", "-2*T3+0.01*T4", "0.25*T0"}
		state.checkTTerms(3652.5, textBytes(terms[int(a)]), r)
		return r
	case "fictName":
		name := textString(state.swiGetFictName(int32(a), make([]byte, 256)))
		r := make([]float64, len(name))
		for i := range name {
			r[i] = float64(name[i])
		}
		return r
	case "twilight":
		return []float64{state.rdiTwilight(int32(a))}
	case "effectiveMass":
		return []float64{state.meff(a)}
	case "eop":
		state.loadDpsiDeps()
		d := &state.positions.data
		return []float64{float64(d.eopDpsiLoaded), d.eopTjdBeg, d.eopTjdEnd, d.dpsi[0], d.deps[0], d.dpsi[2], d.deps[2]}
	case "acronychal":
		r := make([]float64, 2)
		r[0] = float64(state.getAcronychalDay(a, []float64{8.55, 47.37, 408}, []float64{1013.25, 15, 40, 0.25}, []float64{36, 1, 1, 1, 0, 0}, textBytes("sirius"), 2|2048, 5, r[1:], make([]byte, 256)))
		return r
	case "riseHelper":
		r := make([]float64, 2)
		r[0] = float64(state.callSweRiseTrans(a, 3, textBytes(""), 2, 1, []float64{8.55, 47.37, 408}, 1013.25, 15, r[1:], make([]byte, 256)))
		return r
	default:
		panic(fmt.Sprintf("unknown private oracle function %q", function))
	}
}

// writeEOPSample uses the real text layouts with synthetic daily corrections.
// Both C and Go consume these files, including an appended finals record.
func writeEOPSample(t *testing.T, path string) {
	t.Helper()
	daily := "header\n2024 1 1 60000 0 0 0 0 0.1 0.2\n2024 1 2 60001 0 0 0 0 0.3 0.4\n"
	final := bytes.Repeat([]byte{' '}, 200)
	copy(final[7:], "60002 ")
	copy(final[168:], " 500.000  ")
	copy(final[178:], " 600.000  ")
	final = append(final, '\n')
	for name, data := range map[string][]byte{"eop_1962_today.txt": []byte(daily), "eop_finals.txt": final} {
		if err := os.WriteFile(filepath.Join(path, name), data, 0600); err != nil {
			t.Fatal(err)
		}
	}
}
