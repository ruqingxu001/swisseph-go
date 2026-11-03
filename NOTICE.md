# Source and license

This is a local Go translation of Swiss Ephemeris, not an independent algorithm implementation.

- Upstream: https://github.com/aloistr/swisseph
- Tag: `v2.10.03`
- Commit: `175e1fcb3108bcd5c0d146c803f51dcf23508012`
- Original copyright: Astrodienst AG; authors Dieter Koch and Alois Treindl. This attribution is not an endorsement.
- This local derivative selects **AGPL-3.0**, with the full license in `LICENSE`. The original dual-license notice is preserved in `upstream/LICENSE`, all original source headers, and the generated Go files. Translation does not grant a free closed-source commercial license.
- Unmodified input files and SHA-256 hashes are in `upstream/`. The original Makefile is retained for provenance; its optional examples are not included. No ephemeris binary data is redistributed here.
- Our added wrapper, harness, and build scripts are licensed under AGPL-3.0-only.

## Translation changes

`tools/generate.py` uses `modernc.org/ccgo/v4 v4.26.0`; generated code uses `modernc.org/libc v1.62.1`. Their own BSD-style license notices remain in their Go modules (use `go mod vendor` if distributing vendored dependencies).

The nine library units in upstream `SWEOBJ` are translated without replacing their astronomical algorithms. The build defines `TLSOFF`: C thread-local storage becomes shared state, protected by the public Go wrapper's process-wide mutex for the entire operation. `NO_SWE_GLP` uses the upstream empty-library-path implementation because there is no dynamic Swiss library.

The temporary translation copy changes the two `swejpl.c` macros `FSEEK fseeko` / `FTELL ftello` to `fseek` / `ftell`. All four supported targets have 64-bit `long` and file offsets. This resolves missing Darwin libc adapter symbols without modifying numerical logic. Upstream files themselves remain unchanged. JPL binary-file seek/read behavior is covered by 400 DE200 differential cases; other JPL versions and damaged inputs are not exhaustively covered.

Generated layouts are platform-specific. The initial Darwin generation used Xcode SDK headers on the local Mac; regenerating with a different SDK/toolchain can change platform declarations. Build scripts and source hashes are reproducible inputs, not a claim of byte-identical output across SDK versions.

A deterministic post-processing step replaces generated `libc.Xatof` calls with the local `cAtof` adapter. libc v1.62.1 aborts on nonnumeric fixed-star fields such as `ICRS`; C atof returns zero. The adapter implements C-locale numeric-prefix parsing and is checked against native C in 1,020 cases. It changes no astronomical formulas.

The public API marshaling layer is generated from the unmodified `swephexp.h`; it copies slices to libc-managed allocations and preserves native statuses. Engine/Session state handling, validation, fixtures and documentation are local additions dated 2026-11-02.

Public scalar C `char` arguments use a target-specific alias: signed on Darwin and Linux amd64, unsigned on Linux arm64. The Go-facing int8 parameter preserves the same eight bits.
