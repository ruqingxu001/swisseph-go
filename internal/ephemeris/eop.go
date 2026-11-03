// SPDX-License-Identifier: AGPL-3.0-only
// Derived from Swiss Ephemeris 2.10.03; see NOTICE.md and upstream/LICENSE.
// Modified for native Go records, slices and isolated state, 2026-11-03.

/* Copyright (C) 1997 - 2021 Astrodienst AG, Switzerland.  All rights reserved.

License conditions
------------------

This file is part of Swiss Ephemeris.

Swiss Ephemeris is distributed with NO WARRANTY OF ANY KIND.  No author
or distributor accepts any responsibility for the consequences of using it,
or for whether it serves any particular purpose or works at all, unless he
or she says so in writing.

Swiss Ephemeris is made available by its authors under a dual licensing
system. The software developer, who uses any part of Swiss Ephemeris
in his or her software, must choose between one of the two license models,
which are
a) GNU Affero General Public License (AGPL)
b) Swiss Ephemeris Professional License

The choice must be made before the software developer distributes software
containing parts of Swiss Ephemeris to others, and before any public
service using the developed software is activated.

If the developer choses the AGPL software license, he or she must fulfill
the conditions of that license, which includes the obligation to place his
or her whole software project under the AGPL or a compatible license.
See https://www.gnu.org/licenses/agpl-3.0.html

If the developer choses the Swiss Ephemeris Professional license,
he must follow the instructions as found in http://www.astro.com/swisseph/
and purchase the Swiss Ephemeris Professional Edition from Astrodienst
and sign the corresponding license contract.

The License grants you the right to use, copy, modify and redistribute
Swiss Ephemeris, but only under certain conditions described in the License.
Among other things, the License requires that the copyright notices and
this notice be preserved on all copies.

Authors of the Swiss Ephemeris: Dieter Koch and Alois Treindl

The authors of Swiss Ephemeris have no control or influence over any of
the derived works, i.e. over software or services created by other
programmers which use Swiss Ephemeris functions.

The names of the authors or of the copyright holder (Astrodienst) must not
be used for promoting any software, product or service which uses or contains
the Swiss Ephemeris. This copyright notice is the ONLY place where the
names of the authors can legally appear, except in cases where they have
given special permission in writing.
*/
package ephemeris

import "strings"

// loadDpsiDeps implements load_dpsi_deps from sweph.c. Daily IERS corrections
// precede optional fixed-column finals. Dates must remain consecutive across
// both files. Corrections in the daily file are arcseconds; finals are mas.
func (state *State) loadDpsiDeps() {
	data := &state.positions.data
	if data.eopDpsiLoaded > 0 {
		return
	}
	file := state.swiFopen(-1, textBytes("eop_1962_today.txt"), data.ephemerisPath[:], nil)
	if file == nil {
		data.eopDpsiLoaded = -1
		return
	}
	data.dpsi = make([]float64, 36525)
	data.deps = make([]float64, 36525)
	data.eopTjdBegHorizons = 2437684.5
	const modifiedJDOffset = 2400000.5
	var line [256]byte
	count, previousDay, day := 0, int32(0), int32(0)
	for readLine(line[:], len(line), file) != nil {
		fields := strings.Fields(textString(line[:]))
		if len(fields) == 0 || parseInt(fields[0]) == 0 {
			continue
		}
		if len(fields) < 10 || count >= len(data.dpsi) {
			data.eopDpsiLoaded = -2
			closeTextFile(file)
			return
		}
		day = parseInt(fields[3])
		if previousDay > 0 && day-previousDay != 1 {
			data.eopDpsiLoaded = -2
			closeTextFile(file)
			return
		}
		if count == 0 {
			data.eopTjdBeg = float64(day) + modifiedJDOffset
		}
		data.dpsi[count] = parseFloat(fields[8])
		data.deps[count] = parseFloat(fields[9])
		count++
		previousDay = day
	}
	data.eopTjdEnd = float64(day) + modifiedJDOffset
	data.eopDpsiLoaded = 1
	closeTextFile(file)
	file = state.swiFopen(-1, textBytes("eop_finals.txt"), data.ephemerisPath[:], nil)
	if file == nil {
		return
	}
	defer closeTextFile(file)
	for readLine(line[:], len(line), file) != nil {
		day = parseInt(line[7:])
		if float64(day)+modifiedJDOffset <= data.eopTjdEnd {
			continue
		}
		if count >= len(data.dpsi) {
			return
		}
		if previousDay > 0 && day-previousDay != 1 {
			data.eopDpsiLoaded = -3
			return
		}
		if textLength(line[:]) < 178 {
			data.eopDpsiLoaded = 2
			return
		}
		longitude := parseFloat(line[168:])
		obliquity := parseFloat(line[178:])
		if longitude == 0 {
			longitude = parseFloat(line[99:])
			obliquity = parseFloat(line[118:])
		}
		if longitude == 0 {
			data.eopDpsiLoaded = 2
			return
		}
		data.eopTjdEnd = float64(day) + modifiedJDOffset
		data.dpsi[count] = longitude / 1000.0
		data.deps[count] = obliquity / 1000.0
		count++
		previousDay = day
	}
	data.eopDpsiLoaded = 2
}
