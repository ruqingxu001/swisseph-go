// SPDX-License-Identifier: AGPL-3.0-only
package ephemeris

import (
	"encoding/binary"
	"io"
	"math"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestBufferedFilePositionAndBinaryReads(t *testing.T) {
	path := filepath.Join(t.TempDir(), "mixed.dat")
	data := append([]byte("header\n"), 0x78, 0x56, 0x34, 0x12)
	if err := os.WriteFile(path, data, 0600); err != nil {
		t.Fatal(err)
	}
	file := openTextFile(path, "r")
	if file == nil {
		t.Fatal("open")
	}
	defer closeTextFile(file)
	line := make([]byte, 32)
	if readLine(line, len(line), file) == nil || textString(line) != "header\n" {
		t.Fatal("header")
	}
	if offset := tellFile(file); offset != 7 {
		t.Fatalf("logical position = %d", offset)
	}
	var word [1]uint32
	if count := readBinary(word[:], 4, 1, file); count != 1 || word[0] != 0x12345678 {
		t.Fatalf("binary field = %#x (count %d)", word[0], count)
	}
	if seekFile(file, -4, io.SeekCurrent) != 0 || tellFile(file) != 7 {
		t.Fatal("relative seek")
	}
	if readBinary(word[:], 4, 1, file) != 1 || word[0] != 0x12345678 {
		t.Fatal("read after seek")
	}
	rewindFile(file)
	if tellFile(file) != 0 || readLine(line, len(line), file) == nil {
		t.Fatal("rewind")
	}
}

func TestSwissPackedIndicesAndDamagedFiles(t *testing.T) {
	for _, bigEndian := range []bool{false, true} {
		t.Run(map[bool]string{false: "little", true: "big"}[bigEndian], func(t *testing.T) {
			data := []byte{0x56, 0x34, 0x12}
			reorder := int32(0)
			if bigEndian {
				data = []byte{0x12, 0x34, 0x56}
				reorder = 2
			}
			path := filepath.Join(t.TempDir(), "index.se1")
			if err := os.WriteFile(path, data, 0600); err != nil {
				t.Fatal(err)
			}
			file := openTextFile(path, "r")
			defer closeTextFile(file)
			state := NewState()
			var index [1]int32
			message := make([]byte, 256)
			if status := state.doFread(index[:], 3, 1, 4, file, 0, reorder, 0, 0, message); status != 0 || index[0] != 0x123456 {
				t.Fatalf("packed index: %#x, status %d", index[0], status)
			}
			if status := state.doFread(index[:], 4, 1, 4, file, 0, reorder, 0, 0, message); status != -1 || textString(message) == "" {
				t.Fatal("short file accepted")
			}
			if status := state.doFread(index[:], 4, math.MaxInt32, 4, file, 0, reorder, 0, 0, message); status != -1 {
				t.Fatal("invalid count accepted")
			}
		})
	}
}

func TestBinaryDecodingAndOppositeByteOrder(t *testing.T) {
	data := make([]byte, 16)
	binary.LittleEndian.PutUint64(data, math.Float64bits(1.25))
	binary.LittleEndian.PutUint64(data[8:], math.Float64bits(-0.5))
	values := make([]float64, 2)
	decodeBinary(values, data)
	if !reflect.DeepEqual(values, []float64{1.25, -0.5}) {
		t.Fatal(values)
	}
	state := NewState()
	state.reorder(values, 8, 2)
	state.reorder(values, 8, 2)
	if !reflect.DeepEqual(values, []float64{1.25, -0.5}) {
		t.Fatal("endian round trip")
	}
}

func TestSearchPathPreservesDriveLetters(t *testing.T) {
	for _, tc := range []struct {
		input string
		want  []string
	}{
		{"/one:/two", []string{"/one", "/two"}},
		{`C:\ephe;D:\other`, []string{`C:\ephe`, `D:\other`}},
		{"C:/ephe;D:/other", []string{"C:/ephe", "D:/other"}},
		{".", []string{"."}},
	} {
		if got := splitEphemerisPath(tc.input); !reflect.DeepEqual(got, tc.want) {
			t.Errorf("%q: %v", tc.input, got)
		}
	}
}

func TestCalendarAndChebyshevInvariants(t *testing.T) {
	state := NewState()
	for _, calendar := range []int32{0, 1} {
		for _, year := range []int32{-4000, -1, 0, 1582, 2000, 2400} {
			for month := int32(1); month <= 12; month++ {
				jd := state.sweJulday(year, month, 15, 12.25, calendar)
				y, m, d, h := reverseJulianDay(jd, calendar)
				if y != year || m != month || d != 15 || math.Abs(h-12.25) > 1e-7 {
					t.Fatalf("calendar round trip: %d-%d calendar %d", year, month, calendar)
				}
			}
		}
	}
	// 2 + 3*T1(x) + 4*T2(x), with the doubled Swiss constant term.
	coefficients := []float64{4, 3, 4}
	for _, x := range []float64{-1, -0.4, 0, 0.5, 1} {
		if got := state.swiEcheb(x, coefficients, 3); math.Abs(got-(2+3*x+4*(2*x*x-1))) > 1e-14 {
			t.Fatal("Chebyshev value", x, got)
		}
		if got := state.swiEdcheb(x, coefficients, 3); math.Abs(got-(3+16*x)) > 1e-14 {
			t.Fatal("Chebyshev derivative", x, got)
		}
	}
}

func TestEarthOrientationFailuresAndIndependentDefaults(t *testing.T) {
	for _, tc := range []struct {
		name, contents string
		want           int32
	}{
		{"missing", "", -1},
		{"incomplete", "2024 1 1 60000\n", -2},
		{"gap", "2024 1 1 60000 0 0 0 0 0.1 0.2\n2024 1 3 60002 0 0 0 0 0.3 0.4\n", -2},
	} {
		t.Run(tc.name, func(t *testing.T) {
			path := t.TempDir()
			if tc.contents != "" {
				if err := os.WriteFile(filepath.Join(path, "eop_1962_today.txt"), []byte(tc.contents), 0600); err != nil {
					t.Fatal(err)
				}
			}
			state := NewState()
			defer state.Close()
			state.SetEphePath(path)
			state.loadDpsiDeps()
			if got := state.positions.data.eopDpsiLoaded; got != tc.want {
				t.Fatalf("EOP status %d, want %d", got, tc.want)
			}
		})
	}
	first, second := NewState(), NewState()
	first.calendar.leapSeconds[0] = 0
	first.time.deltaTTable[0] = 0
	if second.calendar.leapSeconds[0] != 19720630 || second.time.deltaTTable[0] != 124 {
		t.Fatal("state defaults are shared")
	}
}
