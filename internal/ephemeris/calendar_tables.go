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

var calendarPlaDiam = [20 + 1]float64{
	1392000000.0, 3475000.0, 4878800, 12103600, 6779000, 139822000, 116464000, 50724000,
	49244000, 2376600, 0, 0, 0, 0, 12742016.800000001, 271370.0,
	290000.0, 939400.0, 545000.0, 246596.0, 525400.0,
}

var calendarAyanamsa = [47]ayanamsaEpoch{
	{t0: 2433282.42346, ayanT0: 24.042044444, t0IsUT: 0, precOffset: 11},
	{t0: 2435553.5, ayanT0: 23.245524743000001, t0IsUT: 0, precOffset: 1},
	{t0: 1721057.5, ayanT0: 0, t0IsUT: 1, precOffset: 0},
	{t0: 2415020.0, ayanT0: 21.014439999999979, t0IsUT: 0, precOffset: 11},
	{t0: 2415020.0, ayanT0: 18.660959999999989, t0IsUT: 0, precOffset: -1},
	{t0: 2415020.0, ayanT0: 22.363888999999972, t0IsUT: 0, precOffset: 11},
	{t0: 2415020.0, ayanT0: 26.963097600000026, t0IsUT: 0, precOffset: 0},
	{t0: 2415020.0, ayanT0: 21.082222000000002, t0IsUT: 0, precOffset: -1},
	{t0: 2415020.0, ayanT0: 21.365556000000026, t0IsUT: 0, precOffset: -1},
	{t0: 1684532.5, ayanT0: -5.66667, t0IsUT: 1, precOffset: -1},
	{t0: 1684532.5, ayanT0: -4.26667, t0IsUT: 1, precOffset: -1},
	{t0: 1684532.5, ayanT0: -3.41667, t0IsUT: 1, precOffset: -1},
	{t0: 1684532.5, ayanT0: -4.46667, t0IsUT: 1, precOffset: -1},
	{t0: 1673941, ayanT0: -5.079167, t0IsUT: 1, precOffset: -1},
	{t0: 1684532.5, ayanT0: -4.44138598, t0IsUT: 1, precOffset: 0},
	{t0: 1674484.0, ayanT0: -9.33333, t0IsUT: 1, precOffset: -1},
	{t0: 1927135.8747793, ayanT0: 0, t0IsUT: 1, precOffset: -1},
	{t0: 0, ayanT0: 0, t0IsUT: 0, precOffset: 0},
	{t0: julianDayJ2000, ayanT0: 0, t0IsUT: 0, precOffset: 0},
	{t0: 2415020.0, ayanT0: 0, t0IsUT: 0, precOffset: 0},
	{t0: 2433282.42345905, ayanT0: 0, t0IsUT: 0, precOffset: 0},
	{t0: 1903396.8128654, ayanT0: 0, t0IsUT: 1, precOffset: 0},
	{t0: 1903396.8128654, ayanT0: -0.21463395, t0IsUT: 1, precOffset: 0},
	{t0: 1903396.7895321, ayanT0: 0, t0IsUT: 1, precOffset: 0},
	{t0: 1903396.7895321, ayanT0: -0.23763238, t0IsUT: 1, precOffset: 0},
	{t0: 1903396.8128654, ayanT0: -0.79167046, t0IsUT: 1, precOffset: 0},
	{t0: 1903396.8128654, ayanT0: 2.11070444, t0IsUT: 1, precOffset: 0},
	{t0: 0, ayanT0: 0, t0IsUT: 0, precOffset: 0},
	{t0: 0, ayanT0: 0, t0IsUT: 0, precOffset: 0},
	{t0: 0, ayanT0: 0, t0IsUT: 0, precOffset: 0},
	{t0: 0, ayanT0: 0, t0IsUT: 0, precOffset: 0},
	{t0: 0, ayanT0: 0, t0IsUT: 0, precOffset: 0},
	{t0: 0, ayanT0: 0, t0IsUT: 0, precOffset: 0},
	{t0: 0, ayanT0: 0, t0IsUT: 0, precOffset: 0},
	{t0: 2451079.734892000, ayanT0: 30, t0IsUT: 0, precOffset: 0},
	{t0: 0, ayanT0: 0, t0IsUT: 0, precOffset: 0},
	{t0: 0, ayanT0: 0, t0IsUT: 0, precOffset: 0},
	{t0: 1911797.740782065, ayanT0: 0, t0IsUT: 1, precOffset: 0},
	{t0: 1721057.5, ayanT0: -3.2, t0IsUT: 1, precOffset: -1},
	{t0: 0, ayanT0: 0, t0IsUT: 0, precOffset: 0},
	{t0: 0, ayanT0: 0, t0IsUT: 0, precOffset: 0},
	{t0: 2451544.5, ayanT0: 25.0, t0IsUT: 1, precOffset: 0},
	{t0: 1775845.5, ayanT0: -2.9422, t0IsUT: 1, precOffset: -1},
	{t0: 2415020.0, ayanT0: 22.44597222, t0IsUT: 0, precOffset: 11},
	{t0: 1825235.2458513028, ayanT0: 0.0, t0IsUT: 0, precOffset: 0},
	{t0: 1827424.752255678, ayanT0: 0.0, t0IsUT: 0, precOffset: 0},
	{t0: 2435553.5, ayanT0: 23.245357930000001, t0IsUT: 0, precOffset: 11},
}
