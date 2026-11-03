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

// Defaults are copied by value, so loading a local table affects only its State.
var defaultLeapSeconds = [100]int32{
	19720630, 19721231, 19731231, 19741231, 19751231, 19761231, 19771231, 19781231,
	19791231, 19810630, 19820630, 19830630, 19850630, 19871231, 19891231, 19901231,
	19920630, 19930630, 19940630, 19951231, 19970630, 19981231, 20051231, 20081231,
	20120630, 20150630, 20161231, 0,
}

var defaultDeltaT = [2028 - 1620 + 1 + 100]float64{
	124.00, 119.00, 115.00, 110.00, 106.00, 102.00, 98.00, 95.00,
	91.00, 88.00, 85.00, 82.00, 79.00, 77.00, 74.00, 72.00,
	70.00, 67.00, 65.00, 63.00, 62.00, 60.00, 58.00, 57.00,
	55.00, 54.00, 53.00, 51.00, 50.00, 49.00, 48.00, 47.00,
	46.00, 45.00, 44.00, 43.00, 42.00, 41.00, 40.00, 38.00,
	37.00, 36.00, 35.00, 34.00, 33.00, 32.00, 31.00, 30.00,
	28.00, 27.00, 26.00, 25.00, 24.00, 23.00, 22.00, 21.00,
	20.00, 19.00, 18.00, 17.00, 16.00, 15.00, 14.00, 14.00,
	13.00, 12.00, 12.00, 11.00, 11.00, 10.00, 10.00, 10.00,
	9.00, 9.00, 9.00, 9.00, 9.00, 9.00, 9.00, 9.00,
	9.00, 9.00, 9.00, 9.00, 9.00, 9.00, 9.00, 9.00,
	10.00, 10.00, 10.00, 10.00, 10.00, 10.00, 10.00, 10.00,
	10.00, 11.00, 11.00, 11.00, 11.00, 11.00, 11.00, 11.00,
	11.00, 11.00, 11.00, 11.00, 11.00, 11.00, 11.00, 11.00,
	11.00, 11.00, 12.00, 12.00, 12.00, 12.00, 12.00, 12.00,
	12.00, 12.00, 12.00, 12.00, 13.00, 13.00, 13.00, 13.00,
	13.00, 13.00, 13.00, 14.00, 14.00, 14.00, 14.00, 14.00,
	14.00, 14.00, 15.00, 15.00, 15.00, 15.00, 15.00, 15.00,
	15.00, 16.00, 16.00, 16.00, 16.00, 16.00, 16.00, 16.00,
	16.00, 16.00, 16.00, 17.00, 17.00, 17.00, 17.00, 17.00,
	17.00, 17.00, 17.00, 17.00, 17.00, 17.00, 17.00, 17.00,
	17.00, 17.00, 17.00, 17.00, 16.00, 16.00, 16.00, 16.00,
	15.00, 15.00, 14.00, 14.00, 13.70, 13.40, 13.10, 12.90,
	12.70, 12.60, 12.50, 12.50, 12.50, 12.50, 12.50, 12.50,
	12.50, 12.50, 12.50, 12.50, 12.50, 12.40, 12.30, 12.20,
	12.00, 11.70, 11.40, 11.10, 10.60, 10.20, 9.60, 9.10,
	8.60, 8.00, 7.50, 7.00, 6.60, 6.30, 6.00, 5.80,
	5.70, 5.60, 5.60, 5.60, 5.70, 5.80, 5.90, 6.10,
	6.20, 6.30, 6.50, 6.60, 6.80, 6.90, 7.10, 7.20,
	7.30, 7.40, 7.50, 7.60, 7.70, 7.70, 7.80, 7.80,
	7.88, 7.82, 7.54, 6.97, 6.40, 6.02, 5.41, 4.10,
	2.92, 1.82, 1.61, .10, -1.02, -1.28, -2.69, -3.24,
	-3.64, -4.54, -4.71, -5.11, -5.40, -5.42, -5.20, -5.46,
	-5.46, -5.79, -5.63, -5.64, -5.80, -5.66, -5.87, -6.01,
	-6.19, -6.64, -6.44, -6.47, -6.09, -5.76, -4.66, -3.74,
	-2.72, -1.54, -.02, 1.24, 2.64, 3.86, 5.37, 6.14,
	7.75, 9.13, 10.46, 11.53, 13.36, 14.65, 16.01, 17.20,
	18.24, 19.06, 20.25, 20.95, 21.16, 22.25, 22.41, 23.03,
	23.49, 23.62, 23.86, 24.49, 24.34, 24.08, 24.02, 24.00,
	23.87, 23.95, 23.86, 23.93, 23.73, 23.92, 23.96, 24.02,
	24.33, 24.83, 25.30, 25.70, 26.24, 26.77, 27.28, 27.78,
	28.25, 28.71, 29.15, 29.57, 29.97, 30.36, 30.72, 31.07,
	31.35, 31.68, 32.18, 32.68, 33.15, 33.59, 34.00, 34.47,
	35.03, 35.73, 36.54, 37.43, 38.29, 39.20, 40.18, 41.17,
	42.23, 43.37, 44.4841, 45.4761, 46.4567, 47.5214, 48.5344, 49.5862,
	50.5387, 51.3808, 52.1668, 52.9565, 53.7882, 54.3427, 54.8713, 55.3222,
	55.8197, 56.3000, 56.8553, 57.5653, 58.3092, 59.1218, 59.9845, 60.7854,
	61.6287, 62.2951, 62.9659, 63.4673, 63.8285, 64.0908, 64.2998, 64.4734,
	64.5736, 64.6876, 64.8452, 65.1464, 65.4574, 65.7768, 66.0699, 66.3246,
	66.6030, 66.9069, 67.2810, 67.6439, 68.1024, 68.5927, 68.9676, 69.2202,
	69.3612, 69.3593, 69.2945, 69.1833, 69.10, 69.00, 68.90, 68.80,
	68.80,
}
