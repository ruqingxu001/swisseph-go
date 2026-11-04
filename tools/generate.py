#!/usr/bin/env python3
# SPDX-License-Identifier: AGPL-3.0-only
"""Verify pinned sources. The native Go core is maintained by hand."""
import hashlib
from pathlib import Path

ROOT = Path(__file__).resolve().parents[1]

def verify():
    for line in (ROOT / 'upstream/SHA256SUMS').read_text().splitlines():
        digest, name = line.split('  ')
        if hashlib.sha256((ROOT / 'upstream' / name).read_bytes()).hexdigest() != digest:
            raise SystemExit('Upstream checksum mismatch: ' + name)
    print('Pinned Swiss Ephemeris sources verified; edit internal/ephemeris directly.')

if __name__ == '__main__':
    verify()
