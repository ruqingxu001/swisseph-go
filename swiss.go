// SPDX-License-Identifier: AGPL-3.0-only

// Package swisseph provides a Go interface to a direct Go translation of Swiss
// Ephemeris 2.10.03. Runtime calculations do not use CGO or a C subprocess.
package swisseph

import (
	"errors"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"unsafe"

	"github.com/ruqingxu001/swisseph-go/internal/raw"
	"modernc.org/libc"
)

const (
	Julian    int32 = 0
	Gregorian int32 = 1
)

const (
	Sun int32 = iota
	Moon
	Mercury
	Venus
	Mars
	Jupiter
	Saturn
	Uranus
	Neptune
	Pluto
	MeanNode
	TrueNode
	MeanApogee
	OsculatingApogee
	Earth
	Chiron
	Pholus
	Ceres
	Pallas
	Juno
	Vesta
)

const (
	JPLEphemeris     int32 = 1
	SwissEphemeris   int32 = 2
	MoshierEphemeris int32 = 4
	Heliocentric     int32 = 8
	TruePosition     int32 = 16
	J2000            int32 = 32
	NoNutation       int32 = 64
	Speed            int32 = 256
	Equatorial       int32 = 2048
	Cartesian        int32 = 4096
	Radians          int32 = 8192
	Barycentric      int32 = 16384
	Topocentric      int32 = 32768
	Sidereal         int32 = 65536
	Lahiri           int32 = 1
)

// Config is copied by New. Coordinates use east/north positive degrees and
// altitude in metres. SiderealMode is the original SE_SIDM_* value (0 = Fagan).
// EphemerisPath is one directory, not a path search list. An empty path means
// the current directory at New time. Missing files are errors unless
// AllowFallback is explicitly enabled; CalcResult always includes actual flags.
type Config struct {
	EphemerisPath                 string
	SiderealMode                  int32
	SiderealEpoch, Ayanamsa       float64
	Longitude, Latitude, Altitude float64
	AllowFallback                 bool
}

type Engine struct{ config Config }

// All upstream global state (including static caches) is protected for the
// complete configure/calculate/close operation, not just individual setters.
var engineMu sync.Mutex

func New(config Config) (*Engine, error) {
	if !finite(config.SiderealEpoch, config.Ayanamsa, config.Longitude, config.Latitude, config.Altitude) || math.Abs(config.Latitude) > 90 || math.Abs(config.Longitude) > 180 {
		return nil, errors.New("invalid configuration coordinates or non-finite value")
	}
	path, err := filepath.Abs(config.EphemerisPath)
	if err != nil {
		return nil, err
	}
	if strings.ContainsAny(path, "\x00:;") || len(path) > 230 {
		return nil, errors.New("ephemeris directory must be one path of at most 230 bytes")
	}
	info, err := os.Stat(path)
	if err != nil {
		return nil, err
	}
	if !info.IsDir() {
		return nil, errors.New("ephemeris path is not a directory")
	}
	config.EphemerisPath = path
	return &Engine{config: config}, nil
}

type call struct {
	tls    *libc.TLS
	memory uintptr
}

// The translated ABI accepts libc-allocated memory only. Never pass pointers to
// Go arrays/slices as uintptr: stack movement and GC would invalidate them.
func (e *Engine) run(fn func(*call) error) error {
	if e == nil {
		return errors.New("nil ephemeris engine")
	}
	engineMu.Lock()
	defer engineMu.Unlock()
	if os.Getenv("SE_EPHE_PATH") != "" {
		return errors.New("unset SE_EPHE_PATH: it overrides the configured ephemeris directory")
	}
	tls := libc.NewTLS()
	defer tls.Close()
	ptr := libc.Xcalloc(tls, 1, 2048)
	if ptr == 0 {
		return errors.New("ephemeris allocation failed")
	}
	defer libc.Xfree(tls, ptr)
	c := &call{tls: tls, memory: ptr}
	// Reserve the first 256 bytes for the path and the next 256 for messages.
	copy(unsafe.Slice((*byte)(unsafe.Pointer(ptr)), 256), e.config.EphemerisPath)
	raw.Xswe_close(tls)
	raw.Xswe_set_delta_t_userdef(tls, -1e-10)
	raw.Xswe_set_interpolate_nut(tls, 0)
	raw.Xswe_set_lapse_rate(tls, 0.0065)
	raw.Xswe_set_ephe_path(tls, ptr)
	copy(unsafe.Slice((*byte)(unsafe.Pointer(ptr+1536)), 256), "de431.eph")
	raw.Xswe_set_jpl_file(tls, ptr+1536)
	// The JPL setter clears lunar-file metadata. Restore it before UT delta-T
	// selection, exactly as a fresh swe_set_ephe_path in the Flutter wrapper.
	raw.Xswe_set_ephe_path(tls, ptr)
	defer raw.Xswe_close(tls)
	raw.Xswe_set_sid_mode(tls, e.config.SiderealMode, e.config.SiderealEpoch, e.config.Ayanamsa)
	raw.Xswe_set_topo(tls, e.config.Longitude, e.config.Latitude, e.config.Altitude)
	return fn(c)
}

func finite(values ...float64) bool {
	for _, v := range values {
		if math.IsNaN(v) || math.IsInf(v, 0) {
			return false
		}
	}
	return true
}
func (c *call) message() string { return libc.GoString(c.memory + 256) }
func (c *call) doubles(offset uintptr, n int) []float64 {
	return unsafe.Slice((*float64)(unsafe.Pointer(c.memory+offset)), n)
}
func (c *call) status(code int32) error {
	if code < 0 {
		return fmt.Errorf("Swiss Ephemeris (%d): %s", code, c.message())
	}
	return nil
}

type CalcResult struct {
	// Ecliptic longitude, latitude, distance (AU), then their daily speeds.
	// Equatorial/Cartesian/Radians flags change these units as in the C API.
	Position [6]float64
	Flags    int32
	Warning  string
}

func (e *Engine) CalcUT(jd float64, body, flags int32) (CalcResult, error) {
	return e.calc(jd, body, flags, false)
}
func (e *Engine) CalcET(jd float64, body, flags int32) (CalcResult, error) {
	return e.calc(jd, body, flags, true)
}
func (e *Engine) calc(jd float64, body, flags int32, et bool) (result CalcResult, err error) {
	if !finite(jd) {
		return result, errors.New("Julian day must be finite")
	}
	err = e.run(func(c *call) error {
		if et {
			result.Flags = raw.Xswe_calc(c.tls, jd, body, flags, c.memory+512, c.memory+256)
		} else {
			result.Flags = raw.Xswe_calc_ut(c.tls, jd, body, flags, c.memory+512, c.memory+256)
		}
		copy(result.Position[:], c.doubles(512, 6))
		result.Warning = c.message()
		if err := c.status(result.Flags); err != nil {
			return err
		}
		requested := flags & 7
		if requested == 0 {
			requested = SwissEphemeris
		}
		// SE_ECL_NUT (-1) returns obliquity/nutation and has no ephemeris flag.
		if body != -1 && !e.config.AllowFallback && result.Flags&7 != requested {
			return fmt.Errorf("ephemeris fallback: requested %d, used %d: %s", requested, result.Flags&7, result.Warning)
		}
		return nil
	})
	return
}

type HousesResult struct {
	// Cusp zero is unused. Gauquelin (G) has 36 sectors; others have 12 cusps.
	Cusps       [37]float64
	Angles      [10]float64
	CuspSpeeds  [37]float64
	AngleSpeeds [10]float64
	Count       int
	Status      int32
	Warning     string
}

func (e *Engine) Houses(jd float64, flags int32, latitude, longitude float64, system byte) (result HousesResult, err error) {
	if !finite(jd, latitude, longitude) || math.Abs(latitude) > 90 || math.Abs(longitude) > 180 {
		return result, errors.New("invalid house time or coordinates")
	}
	if !strings.ContainsRune("PKORCAEVXHTBMGUWFIiJLYSQDN", rune(system)) {
		return result, fmt.Errorf("unsupported house system %q", system)
	}
	result.Count = 12
	if system == 'G' {
		result.Count = 36
	}
	err = e.run(func(c *call) error {
		result.Status = raw.Xswe_houses_ex2(c.tls, jd, flags, latitude, longitude, int32(system), c.memory+512, c.memory+808, c.memory+888, c.memory+1184, c.memory+256)
		copy(result.Cusps[:], c.doubles(512, 37))
		copy(result.Angles[:], c.doubles(808, 10))
		copy(result.CuspSpeeds[:], c.doubles(888, 37))
		copy(result.AngleSpeeds[:], c.doubles(1184, 10))
		result.Warning = c.message()
		return c.status(result.Status) // Preserve partial/fallback output, but never hide polar errors.
	})
	return
}

type CalendarDate struct {
	Year, Month, Day int32
	Hour             float64
}

func validCalendar(calendar int32) bool { return calendar == Julian || calendar == Gregorian }
func (e *Engine) JulianDay(year, month, day int32, hour float64, calendar int32) (jd float64, err error) {
	if !validCalendar(calendar) || !finite(hour) || hour < 0 || hour >= 24 || month < 1 || month > 12 || day < 1 || day > 31 {
		return 0, errors.New("invalid calendar date")
	}
	err = e.run(func(c *call) error {
		jd = raw.Xswe_julday(c.tls, year, month, day, hour, calendar)
		// Validate at noon: the final microsecond of a valid day can round to
		// the next midnight in a float64 Julian day.
		date := c.reverse(raw.Xswe_julday(c.tls, year, month, day, 12, calendar), calendar)
		if date.Year != year || date.Month != month || date.Day != day {
			return errors.New("invalid day in calendar month")
		}
		return nil
	})
	return
}
func (c *call) reverse(jd float64, calendar int32) CalendarDate {
	raw.Xswe_revjul(c.tls, jd, calendar, c.memory+512, c.memory+516, c.memory+520, c.memory+528)
	return CalendarDate{*(*int32)(unsafe.Pointer(c.memory + 512)), *(*int32)(unsafe.Pointer(c.memory + 516)), *(*int32)(unsafe.Pointer(c.memory + 520)), c.doubles(528, 1)[0]}
}
func (e *Engine) ReverseJulianDay(jd float64, calendar int32) (date CalendarDate, err error) {
	if !validCalendar(calendar) || !finite(jd) {
		return date, errors.New("invalid Julian day or calendar")
	}
	err = e.run(func(c *call) error { date = c.reverse(jd, calendar); return nil })
	return
}

type UTCDate struct {
	Year, Month, Day, Hour, Minute int32
	Second                         float64
}
type JulianTimes struct{ ET, UT1 float64 }

func (e *Engine) UTCToJD(date UTCDate, calendar int32) (times JulianTimes, err error) {
	if !validCalendar(calendar) || !finite(date.Second) {
		return times, errors.New("invalid calendar or second")
	}
	err = e.run(func(c *call) error {
		status := raw.Xswe_utc_to_jd(c.tls, date.Year, date.Month, date.Day, date.Hour, date.Minute, date.Second, calendar, c.memory+512, c.memory+256)
		times = JulianTimes{c.doubles(512, 2)[0], c.doubles(512, 2)[1]}
		return c.status(status)
	})
	return
}

type EclipseResult struct {
	Type    int32
	Times   [10]float64
	Warning string
}

// NextEclipse finds the next/previous global solar or lunar eclipse. Type zero
// requests any eclipse; returned Times have the original C API's semantics.
func (e *Engine) NextEclipse(jd float64, flags, eclipseType int32, solar, backward bool) (result EclipseResult, err error) {
	if !finite(jd) {
		return result, errors.New("Julian day must be finite")
	}
	allowed := int32(4 | 16 | 64)
	if solar {
		allowed = 1 | 2 | 4 | 8 | 16 | 32
	}
	if eclipseType < 0 || eclipseType & ^allowed != 0 {
		return result, errors.New("invalid eclipse type mask")
	}
	// A central/noncentral qualifier alone is not a physical solar type.
	if solar && eclipseType != 0 && eclipseType&(4|8|16|32) == 0 {
		return result, errors.New("solar eclipse mask needs a physical type")
	}
	// These C search functions do not report the actual ephemeris flag. Probe Sun
	// and Moon at the search start so a missing data set is not silently accepted.
	for _, body := range []int32{Sun, Moon} {
		if _, err = e.CalcUT(jd, body, flags); err != nil {
			return result, err
		}
	}
	err = e.run(func(c *call) error {
		back := int32(0)
		if backward {
			back = 1
		}
		if solar {
			result.Type = raw.Xswe_sol_eclipse_when_glob(c.tls, jd, flags, eclipseType, c.memory+512, back, c.memory+256)
		} else {
			result.Type = raw.Xswe_lun_eclipse_when(c.tls, jd, flags, eclipseType, c.memory+512, back, c.memory+256)
		}
		copy(result.Times[:], c.doubles(512, 10))
		result.Warning = c.message()
		return c.status(result.Type)
	})
	if err == nil && result.Type > 0 {
		for _, body := range []int32{Sun, Moon} {
			if _, probeErr := e.CalcUT(result.Times[0], body, flags); probeErr != nil {
				return result, probeErr
			}
		}
	}
	return
}

func (e *Engine) Version() (version string, err error) {
	err = e.run(func(c *call) error {
		raw.Xswe_version(c.tls, c.memory+512)
		version = libc.GoString(c.memory + 512)
		return nil
	})
	return
}
