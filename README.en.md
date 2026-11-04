# Swiss Ephemeris Go

[简体中文](README.md) | **English**

A Go port of Swiss Ephemeris **2.10.03**, preserving its algorithms and AGPL license. All nine core modules now share one native Go implementation using only the standard library. Records, slices, explicit module state and binary decoding replace `unsafe`, `modernc.org/libc`, C address/layout emulation and `goto`.

**All 106 public APIs have Go interfaces and sample-based comparisons against the original C implementation.** Validation also runs the actual Flutter sweph wrapper and the application's CalcTools. Per-function results, tolerances, and coverage limits are listed in [API_COVERAGE.md](API_COVERAGE.md) (Chinese).

The module has no dependency on the application. Repository: [ruqingxu001/swisseph-go](https://github.com/ruqingxu001/swisseph-go). Module path: `github.com/ruqingxu001/swisseph-go`. This is a port under validation and has not yet been integrated into the product.

## Usage

Requires Go 1.24.2 or later. This revision was tested on Windows/amd64 and Linux/amd64. Six targets (macOS/Linux/Windows × amd64/arm64) are cross-built; cross-building does not establish runtime correctness on the other machines.

```go
import swiss "github.com/ruqingxu001/swisseph-go"

engine, err := swiss.New(swiss.Config{
    EphemerisPath: "./ephe", // Run from the repository root; use your data directory in deployment.
    SiderealMode: swiss.Lahiri,
})
if err != nil { return err }
jd, err := engine.JulianDay(2000, 1, 1, 12, swiss.Gregorian)
if err != nil { return err }
sun, err := engine.CalcUT(jd, swiss.Sun, swiss.SwissEphemeris|swiss.Speed)
if err != nil { return err }
_ = sun.Position[0] // Ecliptic longitude in degrees; Position[3] is its speed in degrees/day.
```

There are two interface layers:

| Interface | Purpose |
| --- | --- |
| `Engine` | Common calculations: positions, houses, dates, UTC conversion, and eclipses, with error and ephemeris fallback checks |
| `Engine.WithSession` → `Session` | All 106 C APIs, including continuous batches, fixed stars, occultations, heliacal phenomena, and JPL operations |

See [api_generated.go](api_generated.go) for `Session` methods and buffer lengths, or [api.json](api.json) for the machine-readable inventory. For example:

```go
err = engine.WithSession(func(s *swiss.Session) error {
    name, err := s.GetPlanetName(swiss.Mars, make([]byte, 256))
    if err != nil { return err }
    _ = name
    return nil
})
```

- UT means UT1; ET means TT. Use `UTCToJD` for a strict UTC-to-UT1/TT conversion. The current Flutter birth-date flow converts a UTC calendar date directly to JD; use `JulianDay` to match that flow.
- `Engine` rejects ephemeris fallback caused by missing files by default. Set `AllowFallback: true` to allow it; returned `Flags` and `Warning` retain the actual source information. House calculation errors still preserve the fallback results supplied by C.
- `Session` preserves C return codes, diagnostics, and fallback rules. A Go `error` reports session or input-buffer errors; callers must also check the C status code.
- Each `Engine` operation owns independent state and may run concurrently. A `Session` reuses its callback’s caches and serializes its methods. Match the session and call order when comparing continuous Flutter calculations, especially for historical dates.
- Nested `Engine` operations use separate state. The session expires when its callback returns. Callers synchronize shared output buffers; slices passed as separate parameters must not rely on C pointer aliasing.
- Use trusted ephemeris files that remain unchanged during calculations. The path must identify a single directory; the `SE_EPHE_PATH` environment variable must not override it.

## Validation

```sh
CGO_ENABLED=0 go test ./...

# Use the bundled data from the repository root.
export SWISSGO_TEST_EPHE_PATH="$PWD/ephe"
export SWISSGO_TEST_JPL_FILE="$PWD/testdata/jpl/de200.eph"
CGO_ENABLED=0 go test -tags differential,fluttercompare -v ./... -timeout 10m

go test -race -tags differential,fluttercompare ./... -timeout 10m
go vet ./...
```

A C compiler and Python 3 are required only for reference tests. The race detector requires CGO; calculations still use the same Go implementation. Default `go vet ./...` runs with every standard analyzer enabled. Use `py -3` for Python scripts on Windows; run the complete C differential suite on Linux/macOS.

Validation covers all 106 APIs / 357 scenarios, 2,260 Engine C comparisons, 400 real JPL cases, 148 private-module C comparisons, 497 argument boundaries and 1,020 numeric-prefix parsing cases. All 374 original algorithm functions execute in the suite; per-function statement coverage is recorded in [CORE_VALIDATION.md](CORE_VALIDATION.md). The existing 1,885 Dart/App rows remain, with an additional 1,335 real Dart wrapper rows recaptured on Linux. The 550 App rows were retained with their source hashes rechecked, rather than recaptured. Full differential tests, race checks and default vet pass.

The repository includes the 36 ephemeris and supporting files used by Flutter in `ephe/`, plus DE200 for validation in `testdata/jpl/de200.eph`. See [EPHEMERIS.md](EPHEMERIS.md) (Chinese) for provenance, sizes, and usage. Checksums are recorded in `testdata/ephemeris.sha256` and `testdata/jpl-source.json`. The JPL test explicitly skips if its environment variable is not set.

## Refreshing Flutter references

`testdata/flutter/` contains outputs from actual runs with synthetic inputs and source-file hashes, with no user profiles. Tests use these saved results by default. Refreshing them requires the corresponding Flutter application and its pinned sweph version, 3.2.1+2.10.3:

```sh
python3 tools/run_flutter_reference.py \
  --dart /path/to/flutter/bin/dart \
  --packages /path/to/app/.dart_tool/package_config.json \
  --ephe /path/to/ephe \
  --output testdata/flutter/wrapper.jsonl \
  --app /path/to/app \
  --flutter /path/to/flutter/bin/flutter \
  --app-output testdata/flutter/app.jsonl
```

The script checks that the Dart package's C sources match this library's pinned version, compiles those sources, and invokes the actual wrapper and application CalcTools. Building this standalone library and running its ordinary tests do not require the Flutter application.

The Flutter package has an identified `Hsys.X/Y` encoding defect: it uses 58/59 instead of 88/89. The comparison records the original wrapper's behavior; Go's `Engine.Houses` uses the correct codes. The application and package cache were not modified.


The Linux capture is saved as `testdata/flutter/linux-wrapper.jsonl`; its manifest distinguishes the new wrapper capture from the retained App fixture. To select it:

```sh
SWISSGO_FLUTTER_REFERENCE=testdata/flutter/linux-wrapper.jsonl \
  go test -tags fluttercompare ./...
```

## Generation and contributions

```sh
python3 tools/generate.py
python3 tools/generate_api.py
python3 tools/build_api_cases.py
# Generate both function matrices after the full suite and coverage run:
# go test -tags differential,fluttercompare -v -coverpkg=./internal/ephemeris -coverprofile=core.cover ./...
# go tool cover -func=core.cover > core-functions.txt
python3 tools/write_coverage.py --log /path/to/full-test.log \
  --functions /path/to/core-functions.txt
```

See [NOTICE.md](NOTICE.md) for provenance, the pinned upstream commit, translation changes, and licensing. Contribution guidance is in [CONTRIBUTING.md](CONTRIBUTING.md) (Chinese). Maintain `internal/ephemeris` directly. `tools/generate.py` verifies upstream hashes; the API generator does not overwrite algorithms. CI runs unit tests, full differential and saved Flutter comparisons, race, default vet and six cross-builds.

This remains a port under validation. Inputs and data files have not been exhaustively tested; the application's full synastry, progression, Arabic-parts, and report logic has not been ported, and production load testing has not been performed. Passing samples does not establish correctness for every possible scenario.
