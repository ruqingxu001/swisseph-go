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

// Binary64 values match the corresponding expressions in sweph.h. In
// particular, preserve C's rounded conversion factors when changing formulas.
const (
	radiansPerDegree = 0.017453292519943295
	degreesPerRadian = 57.295779513082323
	julianDayJ2000   = 2451545.0
	daysPerCentury   = 36525.0
	secondsPerDay    = 86400.0
)

// Calculation flags correspond to SEFLG_* in swephexp.h.
const (
	flagJPL            = 1 << 0
	flagSwiss          = 1 << 1
	flagMoshier        = 1 << 2
	flagHeliocentric   = 1 << 3
	flagTruePosition   = 1 << 4
	flagJ2000          = 1 << 5
	flagNoNutation     = 1 << 6
	flagSpeed3         = 1 << 7
	flagSpeed          = 1 << 8
	flagNoDeflection   = 1 << 9
	flagNoAberration   = 1 << 10
	flagEquatorial     = 1 << 11
	flagCartesian      = 1 << 12
	flagRadians        = 1 << 13
	flagBarycentric    = 1 << 14
	flagTopocentric    = 1 << 15
	flagSidereal       = 1 << 16
	flagICRS           = 1 << 17
	flagHorizons       = 1 << 18
	flagHorizonsApprox = 1 << 19
	flagCentreOfBody   = 1 << 20
	flagRawPlanetMoon  = 1 << 21
)

// SEFLG_TEST_PLMOON selects the uncorrected planet-moon file data.
const flagTestPlanetMoon = flagRawPlanetMoon | flagJ2000 | flagICRS | flagHeliocentric | flagTruePosition
