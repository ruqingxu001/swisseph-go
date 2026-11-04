# Source and license

This is a derivative Go implementation of Swiss Ephemeris, preserving its astronomical algorithms.

- Upstream: https://github.com/aloistr/swisseph
- Tag: `v2.10.03`
- Commit: `175e1fcb3108bcd5c0d146c803f51dcf23508012`
- Copyright (C) 1997 - 2021 Astrodienst AG, Switzerland. All rights reserved. Authors: Dieter Koch and Alois Treindl. This attribution is not an endorsement.
- This derivative selects **AGPL-3.0-only**, with the full license in `LICENSE`. The original dual-license notice remains in `upstream/LICENSE`, the unmodified sources, and the derived core files. Translation does not grant a free closed-source commercial license.
- Unmodified inputs and SHA-256 hashes remain in `upstream/`. The original Makefile is retained for provenance; optional examples are not included. The Flutter dataset and DE200 binary retain their provenance and checksums; see `EPHEMERIS.md`.
- Added wrappers, tests and tools are licensed under AGPL-3.0-only.

## Go refactoring, 2026-11-03

The former four platform-specific ccgo outputs and their modernc runtime dependencies have been replaced by one manually maintained core in `internal/ephemeris`. The nine original computational domains and 374-function inventory are retained, using native records, arrays, slices, structured control flow and standard-library file I/O. There is no runtime C compiler, subprocess, ABI emulation or external Go dependency. Public API bridges are generated from the pinned header; the numerical core is maintained directly.

Each Engine operation owns its calculation state. A Session retains that state and its caches for the callback, with method-level synchronization. Read-only coefficient tables are shared, while leap-second and delta-T defaults are copied by value. Binary Swiss/JPL records are decoded explicitly, with buffered seek/tell and byte-order conversion independent of host layout.

Numerical ordering, signed/unsigned conversions and C binary64 constant rounding are preserved. Numeric-prefix parsing follows C-locale atof behavior (1,020 native C comparisons); unspecified NaN signs/payloads are compared by classification. The heliacal arc-visibility implementation uses a three-element azimuth output buffer because swe_azalt writes three values; the original C caller declares only two. Truncated binary records and invalid EOP daily records produce errors. Valid file calculations remain covered by native C comparisons. The test C sources themselves are unchanged.

The C reference builds use `TLSOFF` and `NO_SWE_GLP` only in test executables. The Go library-path API returns the upstream empty-path result because there is no dynamically loaded Swiss library. Public C char codes are represented by their byte values. Fixed references, tolerances and remaining coverage limits are documented in `API_COVERAGE.md` and `CORE_VALIDATION.md`.

The Linux Dart wrapper reference was recaptured with Dart 3.5.1 and unchanged sweph 3.2.1+2.10.3 sources. A standalone package configuration supplies its dependencies without changing the wrapper or algorithms. The existing App fixture was retained and its source hashes rechecked; the Flutter application was not rerun for this revision.

## Bundled ephemeris data

`ephe/` preserves the existing Flutter dataset byte-for-byte, including catalog headers and notes. Swiss Ephemeris attribution and notices remain in `upstream/LICENSE` and `LICENSE`.

`testdata/jpl/de200.eph` is an unchanged NASA/JPL DE200 binary from the official Solar System Dynamics service. Provenance, byte count and SHA-256 are in `testdata/jpl-source.json`. It is third-party scientific data; the wrapper's AGPL declaration does not replace the provider's applicable terms. No NASA/JPL endorsement is implied.
