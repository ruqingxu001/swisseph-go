# Swiss Ephemeris Go

[简体中文](README.md) | **English**

An independent Go port of Swiss Ephemeris **2.10.03**, preserving the original algorithms and AGPL license. All nine core C translation units have been translated. Runtime calculations require no CGO, libswe, or C subprocess; they use the pure-Go `modernc.org/libc` package and `unsafe`. The generated code follows the original C structure.

**All 106 public APIs have Go interfaces and sample-based comparisons against the original C implementation.** Validation also runs the actual Flutter sweph wrapper and the application's CalcTools. Per-function results, tolerances, and coverage limits are listed in [API_COVERAGE.md](API_COVERAGE.md) (Chinese).

The module has no dependency on the application. Repository: [ruqingxu001/swisseph-go](https://github.com/ruqingxu001/swisseph-go). Module path: `github.com/ruqingxu001/swisseph-go`. This is a port under validation and has not yet been integrated into the product.

## Usage

Requires Go 1.24.2 or later. Validation was performed with Go 1.27.0 on macOS/amd64. Generated targets cover macOS/Linux × amd64/arm64; the other three platforms have only been cross-built so far.

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
- Calculations within a process are serialized. `Engine` resets configuration for each operation; `Session` retains settings and caches within its callback. Match the session and call order when comparing continuous Flutter calculations, especially for historical dates.
- Do not call `Engine` methods or nest sessions inside a `WithSession` callback. The session expires when the callback returns. Slices passed as different parameters must not rely on C pointer aliasing.
- Use trusted ephemeris files that remain unchanged during calculations. The path must identify a single directory; the `SE_EPHE_PATH` environment variable must not override it.

## Validation

```sh
CGO_ENABLED=0 go test ./...

# Use the bundled data from the repository root.
export SWISSGO_TEST_EPHE_PATH="$PWD/ephe"
export SWISSGO_TEST_JPL_FILE="$PWD/testdata/jpl/de200.eph"
CGO_ENABLED=0 go test -tags differential,fluttercompare -v ./... -timeout 5m

go test -race -tags differential,fluttercompare ./... -timeout 5m
go vet -unsafeptr=false -unreachable=false -assign=false ./...
```

A C compiler and Python 3 are used only to build test references. The race detector itself requires CGO; the calculation implementation stays the same. Standard `go vet` reports pointer conversions, unreachable statements, and self-assignments in translated code, so the command above disables only those three analyzers.

Passed validation includes 106 APIs across 256 scenarios, 2,260 common-interface cases, 400 real JPL-file cases, 1,885 actual Dart/application outputs, 497 invalid-argument checks, 1,020 `atof` comparisons, and race checks. All four targets build successfully. Regenerating the four platform files and API wrappers reproduced the existing files.

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

## Generation and contributions

```sh
python3 tools/generate.py
python3 tools/generate_api.py
python3 tools/build_api_cases.py
# Generate the per-function matrix after a successful full test run:
python3 tools/write_coverage.py --log /path/to/full-test.log
```

See [NOTICE.md](NOTICE.md) for provenance, the pinned upstream commit, translation changes, and licensing. Contribution guidance is in [CONTRIBUTING.md](CONTRIBUTING.md) (Chinese). CI runs basic validation; use the `differential,fluttercompare` commands above for the complete data-backed suite.

This remains a port under validation. Inputs and data files have not been exhaustively tested; the application's full synastry, progression, Arabic-parts, and report logic has not been ported, and production load testing has not been performed. Passing samples does not establish correctness for every possible scenario.
