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

var positionsAyanamsaName = [47][]byte{
	textBytes("Fagan/Bradley"), textBytes("Lahiri"), textBytes("De Luce"), textBytes("Raman"),
	textBytes("Usha/Shashi"), textBytes("Krishnamurti"), textBytes("Djwhal Khul"), textBytes("Yukteshwar"),
	textBytes("J.N. Bhasin"), textBytes("Babylonian/Kugler 1"), textBytes("Babylonian/Kugler 2"),
	textBytes("Babylonian/Kugler 3"), textBytes("Babylonian/Huber"), textBytes("Babylonian/Eta Piscium"),
	textBytes("Babylonian/Aldebaran = 15 Tau"), textBytes("Hipparchos"), textBytes("Sassanian"),
	textBytes("Galact. Center = 0 Sag"), textBytes("J2000"), textBytes("J1900"), textBytes("B1950"),
	textBytes("Suryasiddhanta"), textBytes("Suryasiddhanta, mean Sun"), textBytes("Aryabhata"),
	textBytes("Aryabhata, mean Sun"), textBytes("SS Revati"), textBytes("SS Citra"), textBytes("True Citra"),
	textBytes("True Revati"), textBytes("True Pushya (PVRN Rao)"), textBytes("Galactic Center (Gil Brand)"),
	textBytes("Galactic Equator (IAU1958)"), textBytes("Galactic Equator"),
	textBytes("Galactic Equator mid-Mula"), textBytes("Skydram (Mardyks)"),
	textBytes("True Mula (Chandra Hari)"), textBytes("Dhruva/Gal.Center/Mula (Wilhelm)"),
	textBytes("Aryabhata 522"), textBytes("Babylonian/Britton"), textBytes("\"Vedic\"/Sheoran"),
	textBytes("Cochrane (Gal.Center = 0 Cap)"), textBytes("Galactic Equator (Fiorenza)"),
	textBytes("Vettius Valens"), textBytes("Lahiri 1940"), textBytes("Lahiri VP285"),
	textBytes("Krishnamurti-Senthilathiban"), textBytes("Lahiri ICRC"),
}

var positionsPnoint2jpl = [11]int32{
	2, 9, 0, 1, 3, 4, 5, 6,
	7, 8, 10,
}

var positionsPnoext2int = [21]int32{
	0, 1, 2, 3, 4, 5, 6, 7,
	8, 9, 0, 0, 0, 0, 0, 12,
	13, 14, 15, 16, 17,
}

var positionsEffArr = [101]massDistribution{
	{radiusFraction: 1.000, massFraction: 1.000000},
	{radiusFraction: 0.990, massFraction: 0.999979},
	{radiusFraction: 0.980, massFraction: 0.999940},
	{radiusFraction: 0.970, massFraction: 0.999881},
	{radiusFraction: 0.960, massFraction: 0.999811},
	{radiusFraction: 0.950, massFraction: 0.999724},
	{radiusFraction: 0.940, massFraction: 0.999622},
	{radiusFraction: 0.930, massFraction: 0.999497},
	{radiusFraction: 0.920, massFraction: 0.999354},
	{radiusFraction: 0.910, massFraction: 0.999192},
	{radiusFraction: 0.900, massFraction: 0.999000},
	{radiusFraction: 0.890, massFraction: 0.998786},
	{radiusFraction: 0.880, massFraction: 0.998535},
	{radiusFraction: 0.870, massFraction: 0.998242},
	{radiusFraction: 0.860, massFraction: 0.997919},
	{radiusFraction: 0.850, massFraction: 0.997571},
	{radiusFraction: 0.840, massFraction: 0.997198},
	{radiusFraction: 0.830, massFraction: 0.996792},
	{radiusFraction: 0.820, massFraction: 0.996316},
	{radiusFraction: 0.810, massFraction: 0.995791},
	{radiusFraction: 0.800, massFraction: 0.995226},
	{radiusFraction: 0.790, massFraction: 0.994625},
	{radiusFraction: 0.780, massFraction: 0.993991},
	{radiusFraction: 0.770, massFraction: 0.993326},
	{radiusFraction: 0.760, massFraction: 0.992598},
	{radiusFraction: 0.750, massFraction: 0.991770},
	{radiusFraction: 0.740, massFraction: 0.990873},
	{radiusFraction: 0.730, massFraction: 0.989919},
	{radiusFraction: 0.720, massFraction: 0.988912},
	{radiusFraction: 0.710, massFraction: 0.987856},
	{radiusFraction: 0.700, massFraction: 0.986755},
	{radiusFraction: 0.690, massFraction: 0.985610},
	{radiusFraction: 0.680, massFraction: 0.984398},
	{radiusFraction: 0.670, massFraction: 0.982986},
	{radiusFraction: 0.660, massFraction: 0.981437},
	{radiusFraction: 0.650, massFraction: 0.979779},
	{radiusFraction: 0.640, massFraction: 0.978024},
	{radiusFraction: 0.630, massFraction: 0.976182},
	{radiusFraction: 0.620, massFraction: 0.974256},
	{radiusFraction: 0.610, massFraction: 0.972253},
	{radiusFraction: 0.600, massFraction: 0.970174},
	{radiusFraction: 0.590, massFraction: 0.968024},
	{radiusFraction: 0.580, massFraction: 0.965594},
	{radiusFraction: 0.570, massFraction: 0.962797},
	{radiusFraction: 0.560, massFraction: 0.959758},
	{radiusFraction: 0.550, massFraction: 0.956515},
	{radiusFraction: 0.540, massFraction: 0.953088},
	{radiusFraction: 0.530, massFraction: 0.949495},
	{radiusFraction: 0.520, massFraction: 0.945741},
	{radiusFraction: 0.510, massFraction: 0.941838},
	{radiusFraction: 0.500, massFraction: 0.937790},
	{radiusFraction: 0.490, massFraction: 0.933563},
	{radiusFraction: 0.480, massFraction: 0.928668},
	{radiusFraction: 0.470, massFraction: 0.923288},
	{radiusFraction: 0.460, massFraction: 0.917527},
	{radiusFraction: 0.450, massFraction: 0.911432},
	{radiusFraction: 0.440, massFraction: 0.905035},
	{radiusFraction: 0.430, massFraction: 0.898353},
	{radiusFraction: 0.420, massFraction: 0.891022},
	{radiusFraction: 0.410, massFraction: 0.882940},
	{radiusFraction: 0.400, massFraction: 0.874312},
	{radiusFraction: 0.390, massFraction: 0.865206},
	{radiusFraction: 0.380, massFraction: 0.855423},
	{radiusFraction: 0.370, massFraction: 0.844619},
	{radiusFraction: 0.360, massFraction: 0.833074},
	{radiusFraction: 0.350, massFraction: 0.820876},
	{radiusFraction: 0.340, massFraction: 0.808031},
	{radiusFraction: 0.330, massFraction: 0.793962},
	{radiusFraction: 0.320, massFraction: 0.778931},
	{radiusFraction: 0.310, massFraction: 0.763021},
	{radiusFraction: 0.300, massFraction: 0.745815},
	{radiusFraction: 0.290, massFraction: 0.727557},
	{radiusFraction: 0.280, massFraction: 0.708234},
	{radiusFraction: 0.270, massFraction: 0.687583},
	{radiusFraction: 0.260, massFraction: 0.665741},
	{radiusFraction: 0.250, massFraction: 0.642597},
	{radiusFraction: 0.240, massFraction: 0.618252},
	{radiusFraction: 0.230, massFraction: 0.592586},
	{radiusFraction: 0.220, massFraction: 0.565747},
	{radiusFraction: 0.210, massFraction: 0.537697},
	{radiusFraction: 0.200, massFraction: 0.508554},
	{radiusFraction: 0.190, massFraction: 0.478420},
	{radiusFraction: 0.180, massFraction: 0.447322},
	{radiusFraction: 0.170, massFraction: 0.415454},
	{radiusFraction: 0.160, massFraction: 0.382892},
	{radiusFraction: 0.150, massFraction: 0.349955},
	{radiusFraction: 0.140, massFraction: 0.316691},
	{radiusFraction: 0.130, massFraction: 0.283565},
	{radiusFraction: 0.120, massFraction: 0.250431},
	{radiusFraction: 0.110, massFraction: 0.218327},
	{radiusFraction: 0.100, massFraction: 0.186794},
	{radiusFraction: 0.090, massFraction: 0.156287},
	{radiusFraction: 0.080, massFraction: 0.128421},
	{radiusFraction: 0.070, massFraction: 0.102237},
	{radiusFraction: 0.060, massFraction: 0.077393},
	{radiusFraction: 0.050, massFraction: 0.054833},
	{radiusFraction: 0.040, massFraction: 0.036361},
	{radiusFraction: 0.030, massFraction: 0.020953},
	{radiusFraction: 0.020, massFraction: 0.009645},
	{radiusFraction: 0.010, massFraction: 0.002767},
	{radiusFraction: 0.000, massFraction: 0.000000},
}
