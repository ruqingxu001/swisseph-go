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

import "math"

// swiMoshmoon2 implements swi_moshmoon2 from swemmoon.c.
//
//	Calculate geometric coordinates of Moon
//
// without light time or nutation correction.
func (state *State) swiMoshmoon2(J float64, pol []float64) int32 {
	var i int32
	state.lunar.centuries = (J - julianDayJ2000) / daysPerCentury
	state.lunar.centuries2 = state.lunar.centuries * state.lunar.centuries
	state.meanElements()
	state.meanElementsPl()
	state.moon1()
	state.moon2()
	state.moon3()
	state.moon4()
	for i = 0; i < 3; i++ {
		pol[i] = state.lunar.polar[i]
	}
	return 0
}

// swiMoshmoon implements swi_moshmoon from swemmoon.c.
//
//	Moshier's moom
//
// tjd		julian day
// xpm		array of 6 doubles for moon's position and speed vectors
// serr		pointer to error string
func (state *State) swiMoshmoon(jd float64, saveCache int32, xpmret []float64, message []byte) int32 {
	var (
		i      int32
		a      float64
		b      float64
		x1     [6]float64
		x2     [6]float64
		t      float64
		xx     [6]float64
		xpm    []float64
		planet *planetData = &state.positions.data.planets[1]
		s      [256]byte
	)
	if saveCache != 0 {
		xpm = planet.stateVector[:]
	} else {
		xpm = xx[:]
	}
	if jd < 625000.30000000005 || jd > 2818000.7000000002 {
		if message != nil {
			formatText(s[:], textBytes("jd %f outside Moshier's Moon range %.2f .. %.2f "), jd, 625000.5, 2818000.5)
			if textLength(message)+textLength(s[:]) < 256 {
				appendText(message, s[:])
			}
		}
		return -1
	}
	if jd == planet.evaluationTime && planet.ephemeris == 4 {
		if xpmret != nil {
			for i = 0; i <= 5; i++ {
				xpmret[i] = planet.stateVector[i]
			}
		}
		return 0
	}
	state.swiMoshmoon2(jd, xpm)
	if saveCache != 0 {
		planet.evaluationTime = jd
		planet.cachedFlags = -1
		planet.ephemeris = 4
	}
	state.ecldatEqu2000(jd, xpm)
	t = jd + 0.00005
	state.swiMoshmoon2(t, x1[:])
	state.ecldatEqu2000(t, x1[:])
	t = jd - 0.00005
	state.swiMoshmoon2(t, x2[:])
	state.ecldatEqu2000(t, x2[:])
	for i = 0; i <= 2; i++ {
		b = (x1[i] - x2[i]) / 2
		a = (x1[i]+x2[i])/2 - xpm[i]
		xpm[i+3] = (2*a + b) / 0.00005
	}
	if xpmret != nil {
		for i = 0; i <= 5; i++ {
			xpmret[i] = xpm[i]
		}
	}
	return 0
}

// moon1 implements moon1 from swemmoon.c.
func (state *State) moon1() {
	var (
		a float64
		i int32
		j int32
	)
	for i = 0; i < 5; i++ {
		for j = 0; j < 8; j++ {
			state.lunar.sinMultiples[i][j] = 0
			state.lunar.cosMultiples[i][j] = 0
		}
	}
	state.moonSscc(0, 4.8481368110953599359e-6*state.lunar.elongation, 6)
	state.moonSscc(1, 4.8481368110953599359e-6*state.lunar.solarAnomaly, 4)
	state.moonSscc(2, 4.8481368110953599359e-6*state.lunar.lunarAnomaly, 4)
	state.moonSscc(3, 4.8481368110953599359e-6*state.lunar.latitudeArgument, 4)
	state.lunar.polar[0] = 0.0
	state.lunar.polar[1] = 0.0
	state.lunar.polar[2] = 0.0
	state.chewm(moonLRT2[:], 25, 4, 2, state.lunar.polar[:])
	state.chewm(moonBT2[:], 12, 4, 4, state.lunar.polar[:])
	state.lunar.f = 18*state.lunar.venusLongitude - 16*state.lunar.earthLongitude
	state.lunar.g = 4.8481368110953599359e-6 * (state.lunar.f - state.lunar.lunarAnomaly)
	state.lunar.cg = math.Cos(state.lunar.g)
	state.lunar.sg = math.Sin(state.lunar.g)
	state.lunar.longitude = 6.367278*state.lunar.cg + 12.747036*state.lunar.sg
	state.lunar.l1 = 23123.70*state.lunar.cg - 10570.02*state.lunar.sg
	state.lunar.l2 = moonZ[12]*state.lunar.cg + moonZ[13]*state.lunar.sg
	state.lunar.polar[2] += 5.01*state.lunar.cg + 2.72*state.lunar.sg
	state.lunar.g = 4.8481368110953599359e-6 * (10.*state.lunar.venusLongitude - 3.*state.lunar.earthLongitude - state.lunar.lunarAnomaly)
	state.lunar.cg = math.Cos(state.lunar.g)
	state.lunar.sg = math.Sin(state.lunar.g)
	state.lunar.longitude += -0.253102*state.lunar.cg + 0.503359*state.lunar.sg
	state.lunar.l1 += 1258.46*state.lunar.cg + 707.29*state.lunar.sg
	state.lunar.l2 += moonZ[14]*state.lunar.cg + moonZ[15]*state.lunar.sg
	state.lunar.g = 4.8481368110953599359e-6 * (8.*state.lunar.venusLongitude - 13.*state.lunar.earthLongitude)
	state.lunar.cg = math.Cos(state.lunar.g)
	state.lunar.sg = math.Sin(state.lunar.g)
	state.lunar.longitude += -0.187231*state.lunar.cg - 0.127481*state.lunar.sg
	state.lunar.l1 += -319.87*state.lunar.cg - 18.34*state.lunar.sg
	state.lunar.l2 += moonZ[16]*state.lunar.cg + moonZ[17]*state.lunar.sg
	a = 4.0*state.lunar.earthLongitude - 8.0*state.lunar.marsLongitude + 3.0*state.lunar.jupiterLongitude
	state.lunar.g = 4.8481368110953599359e-6 * a
	state.lunar.cg = math.Cos(state.lunar.g)
	state.lunar.sg = math.Sin(state.lunar.g)
	state.lunar.longitude += -0.866287*state.lunar.cg + 0.248192*state.lunar.sg
	state.lunar.l1 += 41.87*state.lunar.cg + 1053.97*state.lunar.sg
	state.lunar.l2 += moonZ[18]*state.lunar.cg + moonZ[19]*state.lunar.sg
	state.lunar.g = 4.8481368110953599359e-6 * (a - state.lunar.lunarAnomaly)
	state.lunar.cg = math.Cos(state.lunar.g)
	state.lunar.sg = math.Sin(state.lunar.g)
	state.lunar.longitude += -0.165009*state.lunar.cg + 0.044176*state.lunar.sg
	state.lunar.l1 += 4.67*state.lunar.cg + 201.55*state.lunar.sg
	state.lunar.g = 4.8481368110953599359e-6 * state.lunar.f
	state.lunar.cg = math.Cos(state.lunar.g)
	state.lunar.sg = math.Sin(state.lunar.g)
	state.lunar.longitude += 0.330401*state.lunar.cg + 0.661362*state.lunar.sg
	state.lunar.l1 += 1202.67*state.lunar.cg - 555.59*state.lunar.sg
	state.lunar.l2 += moonZ[20]*state.lunar.cg + moonZ[21]*state.lunar.sg
	state.lunar.g = 4.8481368110953599359e-6 * (state.lunar.f - 2.0*state.lunar.lunarAnomaly)
	state.lunar.cg = math.Cos(state.lunar.g)
	state.lunar.sg = math.Sin(state.lunar.g)
	state.lunar.longitude += 0.352185*state.lunar.cg + 0.705041*state.lunar.sg
	state.lunar.l1 += 1283.59*state.lunar.cg - 586.43*state.lunar.sg
	state.lunar.g = 4.8481368110953599359e-6 * (2.0*state.lunar.jupiterLongitude - 5.0*state.lunar.saturnLongitude)
	state.lunar.cg = math.Cos(state.lunar.g)
	state.lunar.sg = math.Sin(state.lunar.g)
	state.lunar.longitude += -0.034700*state.lunar.cg + 0.160041*state.lunar.sg
	state.lunar.l2 += moonZ[22]*state.lunar.cg + moonZ[23]*state.lunar.sg
	state.lunar.g = 4.8481368110953599359e-6 * (state.lunar.meanLongitude - state.lunar.latitudeArgument)
	state.lunar.cg = math.Cos(state.lunar.g)
	state.lunar.sg = math.Sin(state.lunar.g)
	state.lunar.longitude += 0.000116*state.lunar.cg + 7.063040*state.lunar.sg
	state.lunar.l1 += 298.8 * state.lunar.sg
	state.lunar.sg = math.Sin(4.8481368110953599359e-6 * state.lunar.solarAnomaly)
	state.lunar.l3 = moonZ[24] * state.lunar.sg
	state.lunar.l4 = 0
	state.lunar.g = 4.8481368110953599359e-6 * (2.0*state.lunar.elongation - state.lunar.solarAnomaly)
	state.lunar.sg = math.Sin(state.lunar.g)
	state.lunar.cg = math.Cos(state.lunar.g)
	state.lunar.polar[2] += -0.2655 * state.lunar.cg * state.lunar.centuries
	state.lunar.g = 4.8481368110953599359e-6 * (state.lunar.solarAnomaly - state.lunar.lunarAnomaly)
	state.lunar.polar[2] += -0.1568 * math.Cos(state.lunar.g) * state.lunar.centuries
	state.lunar.g = 4.8481368110953599359e-6 * (state.lunar.solarAnomaly + state.lunar.lunarAnomaly)
	state.lunar.polar[2] += 0.1309 * math.Cos(state.lunar.g) * state.lunar.centuries
	state.lunar.g = 4.8481368110953599359e-6 * (2.0*(state.lunar.elongation+state.lunar.solarAnomaly) - state.lunar.lunarAnomaly)
	state.lunar.sg = math.Sin(state.lunar.g)
	state.lunar.cg = math.Cos(state.lunar.g)
	state.lunar.polar[2] += 0.5568 * state.lunar.cg * state.lunar.centuries
	state.lunar.l2 += state.lunar.polar[0]
	state.lunar.g = 4.8481368110953599359e-6 * (2.0*state.lunar.elongation - state.lunar.solarAnomaly - state.lunar.lunarAnomaly)
	state.lunar.polar[2] += -0.1910 * math.Cos(state.lunar.g) * state.lunar.centuries
	state.lunar.polar[1] *= state.lunar.centuries
	state.lunar.polar[2] *= state.lunar.centuries
	state.lunar.polar[0] = 0.0
	state.chewm(moonBT[:], 16, 4, 4, state.lunar.polar[:])
	state.chewm(moonLRT[:], 38, 4, 1, state.lunar.polar[:])
	state.lunar.g = 4.8481368110953599359e-6 * (state.lunar.f - state.lunar.lunarAnomaly - state.lunar.latitudeArgument - 2355767.6)
	state.lunar.polar[1] += -1127. * math.Sin(state.lunar.g)
	state.lunar.g = 4.8481368110953599359e-6 * (state.lunar.f - state.lunar.lunarAnomaly + state.lunar.latitudeArgument - 235353.6)
	state.lunar.polar[1] += -1123. * math.Sin(state.lunar.g)
	state.lunar.g = 4.8481368110953599359e-6 * (state.lunar.earthLongitude + state.lunar.elongation + 51987.6)
	state.lunar.polar[1] += 1303. * math.Sin(state.lunar.g)
	state.lunar.g = 4.8481368110953599359e-6 * state.lunar.meanLongitude
	state.lunar.polar[1] += 342. * math.Sin(state.lunar.g)
	state.lunar.g = 4.8481368110953599359e-6 * (2.*state.lunar.venusLongitude - 3.*state.lunar.earthLongitude)
	state.lunar.cg = math.Cos(state.lunar.g)
	state.lunar.sg = math.Sin(state.lunar.g)
	state.lunar.longitude += -0.343550*state.lunar.cg - 0.000276*state.lunar.sg
	state.lunar.l1 += 105.90*state.lunar.cg + 336.53*state.lunar.sg
	state.lunar.g = 4.8481368110953599359e-6 * (state.lunar.f - 2.*state.lunar.elongation)
	state.lunar.cg = math.Cos(state.lunar.g)
	state.lunar.sg = math.Sin(state.lunar.g)
	state.lunar.longitude += 0.074668*state.lunar.cg + 0.149501*state.lunar.sg
	state.lunar.l1 += 271.77*state.lunar.cg - 124.20*state.lunar.sg
	state.lunar.g = 4.8481368110953599359e-6 * (state.lunar.f - 2.*state.lunar.elongation - state.lunar.lunarAnomaly)
	state.lunar.cg = math.Cos(state.lunar.g)
	state.lunar.sg = math.Sin(state.lunar.g)
	state.lunar.longitude += 0.073444*state.lunar.cg + 0.147094*state.lunar.sg
	state.lunar.l1 += 265.24*state.lunar.cg - 121.16*state.lunar.sg
	state.lunar.g = 4.8481368110953599359e-6 * (state.lunar.f + 2.*state.lunar.elongation - state.lunar.lunarAnomaly)
	state.lunar.cg = math.Cos(state.lunar.g)
	state.lunar.sg = math.Sin(state.lunar.g)
	state.lunar.longitude += 0.072844*state.lunar.cg + 0.145829*state.lunar.sg
	state.lunar.l1 += 265.18*state.lunar.cg - 121.29*state.lunar.sg
	state.lunar.g = 4.8481368110953599359e-6 * (state.lunar.f + 2.*(state.lunar.elongation-state.lunar.lunarAnomaly))
	state.lunar.cg = math.Cos(state.lunar.g)
	state.lunar.sg = math.Sin(state.lunar.g)
	state.lunar.longitude += 0.070201*state.lunar.cg + 0.140542*state.lunar.sg
	state.lunar.l1 += 255.36*state.lunar.cg - 116.79*state.lunar.sg
	state.lunar.g = 4.8481368110953599359e-6 * (state.lunar.earthLongitude + state.lunar.elongation - state.lunar.latitudeArgument)
	state.lunar.cg = math.Cos(state.lunar.g)
	state.lunar.sg = math.Sin(state.lunar.g)
	state.lunar.longitude += 0.288209*state.lunar.cg - 0.025901*state.lunar.sg
	state.lunar.l1 += -63.51*state.lunar.cg - 240.14*state.lunar.sg
	state.lunar.g = 4.8481368110953599359e-6 * (2.*state.lunar.earthLongitude - 3.*state.lunar.jupiterLongitude + 2.*state.lunar.elongation - state.lunar.lunarAnomaly)
	state.lunar.cg = math.Cos(state.lunar.g)
	state.lunar.sg = math.Sin(state.lunar.g)
	state.lunar.longitude += 0.077865*state.lunar.cg + 0.438460*state.lunar.sg
	state.lunar.l1 += 210.57*state.lunar.cg + 124.84*state.lunar.sg
	state.lunar.g = 4.8481368110953599359e-6 * (state.lunar.earthLongitude - 2.*state.lunar.marsLongitude)
	state.lunar.cg = math.Cos(state.lunar.g)
	state.lunar.sg = math.Sin(state.lunar.g)
	state.lunar.longitude += -0.216579*state.lunar.cg + 0.241702*state.lunar.sg
	state.lunar.l1 += 197.67*state.lunar.cg + 125.23*state.lunar.sg
	state.lunar.g = 4.8481368110953599359e-6 * (a + state.lunar.lunarAnomaly)
	state.lunar.cg = math.Cos(state.lunar.g)
	state.lunar.sg = math.Sin(state.lunar.g)
	state.lunar.longitude += -0.165009*state.lunar.cg + 0.044176*state.lunar.sg
	state.lunar.l1 += 4.67*state.lunar.cg + 201.55*state.lunar.sg
	state.lunar.g = 4.8481368110953599359e-6 * (a + 2.*state.lunar.elongation - state.lunar.lunarAnomaly)
	state.lunar.cg = math.Cos(state.lunar.g)
	state.lunar.sg = math.Sin(state.lunar.g)
	state.lunar.longitude += -0.133533*state.lunar.cg + 0.041116*state.lunar.sg
	state.lunar.l1 += 6.95*state.lunar.cg + 187.07*state.lunar.sg
	state.lunar.g = 4.8481368110953599359e-6 * (a - 2.*state.lunar.elongation + state.lunar.lunarAnomaly)
	state.lunar.cg = math.Cos(state.lunar.g)
	state.lunar.sg = math.Sin(state.lunar.g)
	state.lunar.longitude += -0.133430*state.lunar.cg + 0.041079*state.lunar.sg
	state.lunar.l1 += 6.28*state.lunar.cg + 169.08*state.lunar.sg
	state.lunar.g = 4.8481368110953599359e-6 * (3.*state.lunar.venusLongitude - 4.*state.lunar.earthLongitude)
	state.lunar.cg = math.Cos(state.lunar.g)
	state.lunar.sg = math.Sin(state.lunar.g)
	state.lunar.longitude += -0.175074*state.lunar.cg + 0.003035*state.lunar.sg
	state.lunar.l1 += 49.17*state.lunar.cg + 150.57*state.lunar.sg
	state.lunar.g = 4.8481368110953599359e-6 * (2.*(state.lunar.earthLongitude+state.lunar.elongation-state.lunar.lunarAnomaly) - 3.*state.lunar.jupiterLongitude + 213534.)
	state.lunar.l1 += 158.4 * math.Sin(state.lunar.g)
	state.lunar.l1 += state.lunar.polar[0]
	a = 0.1 * state.lunar.centuries
	state.lunar.polar[1] *= a
	state.lunar.polar[2] *= a
}

// moon2 implements moon2 from swemmoon.c.
//
//	MOSH_MOON_200
func (state *State) moon2() {
	state.lunar.g = 4.8481368110953599359e-6 * (2*(state.lunar.earthLongitude-state.lunar.jupiterLongitude+state.lunar.elongation) - state.lunar.lunarAnomaly + 648431.172)
	state.lunar.longitude += 1.14307 * math.Sin(state.lunar.g)
	state.lunar.g = 4.8481368110953599359e-6 * (state.lunar.venusLongitude - state.lunar.earthLongitude + 648035.568)
	state.lunar.longitude += 0.82155 * math.Sin(state.lunar.g)
	state.lunar.g = 4.8481368110953599359e-6 * (3*(state.lunar.venusLongitude-state.lunar.earthLongitude) + 2*state.lunar.elongation - state.lunar.lunarAnomaly + 647933.184)
	state.lunar.longitude += 0.64371 * math.Sin(state.lunar.g)
	state.lunar.g = 4.8481368110953599359e-6 * (state.lunar.earthLongitude - state.lunar.jupiterLongitude + 4424.04)
	state.lunar.longitude += 0.63880 * math.Sin(state.lunar.g)
	state.lunar.g = 4.8481368110953599359e-6 * (state.lunar.meanLongitude + state.lunar.lunarAnomaly - state.lunar.latitudeArgument + 4.68)
	state.lunar.longitude += 0.49331 * math.Sin(state.lunar.g)
	state.lunar.g = 4.8481368110953599359e-6 * (state.lunar.meanLongitude - state.lunar.lunarAnomaly - state.lunar.latitudeArgument + 4.68)
	state.lunar.longitude += 0.4914 * math.Sin(state.lunar.g)
	state.lunar.g = 4.8481368110953599359e-6 * (state.lunar.meanLongitude + state.lunar.latitudeArgument + 2.52)
	state.lunar.longitude += 0.36061 * math.Sin(state.lunar.g)
	state.lunar.g = 4.8481368110953599359e-6 * (2.*state.lunar.venusLongitude - 2.*state.lunar.earthLongitude + 736.2)
	state.lunar.longitude += 0.30154 * math.Sin(state.lunar.g)
	state.lunar.g = 4.8481368110953599359e-6 * (2.*state.lunar.earthLongitude - 3.*state.lunar.jupiterLongitude + 2.*state.lunar.elongation - 2.*state.lunar.lunarAnomaly + 36138.2)
	state.lunar.longitude += 0.28282 * math.Sin(state.lunar.g)
	state.lunar.g = 4.8481368110953599359e-6 * (2.*state.lunar.earthLongitude - 2.*state.lunar.jupiterLongitude + 2.*state.lunar.elongation - 2.*state.lunar.lunarAnomaly + 311.0)
	state.lunar.longitude += 0.24516 * math.Sin(state.lunar.g)
	state.lunar.g = 4.8481368110953599359e-6 * (state.lunar.earthLongitude - state.lunar.jupiterLongitude - 2.*state.lunar.elongation + state.lunar.lunarAnomaly + 6275.88)
	state.lunar.longitude += 0.21117 * math.Sin(state.lunar.g)
	state.lunar.g = 4.8481368110953599359e-6 * (2.*(state.lunar.earthLongitude-state.lunar.marsLongitude) - 846.36)
	state.lunar.longitude += 0.19444 * math.Sin(state.lunar.g)
	state.lunar.g = 4.8481368110953599359e-6 * (2.*(state.lunar.earthLongitude-state.lunar.jupiterLongitude) + 1569.96)
	state.lunar.longitude -= 0.18457 * math.Sin(state.lunar.g)
	state.lunar.g = 4.8481368110953599359e-6 * (2.*(state.lunar.earthLongitude-state.lunar.jupiterLongitude) - state.lunar.lunarAnomaly - 55.8)
	state.lunar.longitude += 0.18256 * math.Sin(state.lunar.g)
	state.lunar.g = 4.8481368110953599359e-6 * (state.lunar.earthLongitude - state.lunar.jupiterLongitude - 2.*state.lunar.elongation + 6490.08)
	state.lunar.longitude += 0.16499 * math.Sin(state.lunar.g)
	state.lunar.g = 4.8481368110953599359e-6 * (state.lunar.earthLongitude - 2.*state.lunar.jupiterLongitude - 212378.4)
	state.lunar.longitude += 0.16427 * math.Sin(state.lunar.g)
	state.lunar.g = 4.8481368110953599359e-6 * (2.*(state.lunar.venusLongitude-state.lunar.earthLongitude-state.lunar.elongation) + state.lunar.lunarAnomaly + 1122.48)
	state.lunar.longitude += 0.16088 * math.Sin(state.lunar.g)
	state.lunar.g = 4.8481368110953599359e-6 * (state.lunar.venusLongitude - state.lunar.earthLongitude - state.lunar.lunarAnomaly + 32.04)
	state.lunar.longitude -= 0.15350 * math.Sin(state.lunar.g)
	state.lunar.g = 4.8481368110953599359e-6 * (state.lunar.earthLongitude - state.lunar.jupiterLongitude - state.lunar.lunarAnomaly + 4488.88)
	state.lunar.longitude += 0.14346 * math.Sin(state.lunar.g)
	state.lunar.g = 4.8481368110953599359e-6 * (2.*(state.lunar.venusLongitude-state.lunar.earthLongitude+state.lunar.elongation) - state.lunar.lunarAnomaly - 8.64)
	state.lunar.longitude += 0.13594 * math.Sin(state.lunar.g)
	state.lunar.g = 4.8481368110953599359e-6 * (2.*(state.lunar.venusLongitude-state.lunar.earthLongitude-state.lunar.elongation) + 1319.76)
	state.lunar.longitude += 0.13432 * math.Sin(state.lunar.g)
	state.lunar.g = 4.8481368110953599359e-6 * (state.lunar.venusLongitude - state.lunar.earthLongitude - 2.*state.lunar.elongation + state.lunar.lunarAnomaly - 56.16)
	state.lunar.longitude -= 0.13122 * math.Sin(state.lunar.g)
	state.lunar.g = 4.8481368110953599359e-6 * (state.lunar.venusLongitude - state.lunar.earthLongitude + state.lunar.lunarAnomaly + 54.36)
	state.lunar.longitude -= 0.12722 * math.Sin(state.lunar.g)
	state.lunar.g = 4.8481368110953599359e-6 * (3.*(state.lunar.venusLongitude-state.lunar.earthLongitude) - state.lunar.lunarAnomaly + 433.8)
	state.lunar.longitude += 0.12539 * math.Sin(state.lunar.g)
	state.lunar.g = 4.8481368110953599359e-6 * (state.lunar.earthLongitude - state.lunar.jupiterLongitude + state.lunar.lunarAnomaly + 4002.12)
	state.lunar.longitude += 0.10994 * math.Sin(state.lunar.g)
	state.lunar.g = 4.8481368110953599359e-6 * (20.*state.lunar.venusLongitude - 21.*state.lunar.earthLongitude - 2.*state.lunar.elongation + state.lunar.lunarAnomaly - 317511.72)
	state.lunar.longitude += 0.10652 * math.Sin(state.lunar.g)
	state.lunar.g = 4.8481368110953599359e-6 * (26.*state.lunar.venusLongitude - 29.*state.lunar.earthLongitude - state.lunar.lunarAnomaly + 270002.52)
	state.lunar.longitude += 0.10490 * math.Sin(state.lunar.g)
	state.lunar.g = 4.8481368110953599359e-6 * (3.*state.lunar.venusLongitude - 4.*state.lunar.earthLongitude + state.lunar.elongation - state.lunar.lunarAnomaly - 322765.56)
	state.lunar.longitude += 0.10386 * math.Sin(state.lunar.g)
	state.lunar.g = 4.8481368110953599359e-6 * (state.lunar.meanLongitude + 648002.556)
	state.lunar.latitude = 8.04508 * math.Sin(state.lunar.g)
	state.lunar.g = 4.8481368110953599359e-6 * (state.lunar.earthLongitude + state.lunar.elongation + 996048.252)
	state.lunar.latitude += 1.51021 * math.Sin(state.lunar.g)
	state.lunar.g = 4.8481368110953599359e-6 * (state.lunar.f - state.lunar.lunarAnomaly + state.lunar.latitudeArgument + 95554.332)
	state.lunar.latitude += 0.63037 * math.Sin(state.lunar.g)
	state.lunar.g = 4.8481368110953599359e-6 * (state.lunar.f - state.lunar.lunarAnomaly - state.lunar.latitudeArgument + 95553.792)
	state.lunar.latitude += 0.63014 * math.Sin(state.lunar.g)
	state.lunar.g = 4.8481368110953599359e-6 * (state.lunar.meanLongitude - state.lunar.lunarAnomaly + 2.9)
	state.lunar.latitude += 0.45587 * math.Sin(state.lunar.g)
	state.lunar.g = 4.8481368110953599359e-6 * (state.lunar.meanLongitude + state.lunar.lunarAnomaly + 2.5)
	state.lunar.latitude += -0.41573 * math.Sin(state.lunar.g)
	state.lunar.g = 4.8481368110953599359e-6 * (state.lunar.meanLongitude - 2.0*state.lunar.latitudeArgument + 3.2)
	state.lunar.latitude += 0.32623 * math.Sin(state.lunar.g)
	state.lunar.g = 4.8481368110953599359e-6 * (state.lunar.meanLongitude - 2.0*state.lunar.elongation + 2.5)
	state.lunar.latitude += 0.29855 * math.Sin(state.lunar.g)
}

// moon3 implements moon3 from swemmoon.c.
func (state *State) moon3() {
	state.lunar.polar[0] = 0.0
	state.chewm(moonLR[:], 118, 4, 1, state.lunar.polar[:])
	state.chewm(moonMB[:], 77, 4, 3, state.lunar.polar[:])
	state.lunar.longitude += (((state.lunar.l4*state.lunar.centuries+state.lunar.l3)*state.lunar.centuries+state.lunar.l2)*state.lunar.centuries + state.lunar.l1) * state.lunar.centuries * 1.0e-5
	state.lunar.polar[0] = state.lunar.meanLongitude + state.lunar.longitude + 1.0e-4*state.lunar.polar[0]
	state.lunar.polar[1] = 1.0e-4*state.lunar.polar[1] + state.lunar.latitude
	state.lunar.polar[2] = 1.0e-4*state.lunar.polar[2] + 385000.52899
}

// moon4 implements moon4 from swemmoon.c.
//
//	Compute final ecliptic polar coordinates
func (state *State) moon4() {
	state.lunar.polar[2] /= 149597870.69999999
	state.lunar.polar[0] = 4.8481368110953599359e-6 * state.mods3600(state.lunar.polar[0])
	state.lunar.polar[1] = 4.8481368110953599359e-6 * state.lunar.polar[1]
	state.lunar.latitude = state.lunar.polar[1]
}

// corrMeanNode implements corr_mean_node from swemmoon.c.
//
//	1 jan  17000 greg.
func (state *State) corrMeanNode(J float64) float64 {
	var (
		J0      float64
		dJ      float64
		dayscty float64
		dcor    float64
		dcor0   float64
		dcor1   float64
		dfrac   float64
		i       int32
	)
	J0 = -3063616.5
	dayscty = 36524.25
	if J < -3027215.5 {
		return 0
	}
	if J > 7930192.5 {
		return 0
	}
	dJ = J - J0
	i = int32(math.Floor(dJ / dayscty))
	dfrac = (dJ - float64(i)*dayscty) / dayscty
	dcor0 = moonMeanNodeCorr[i]
	dcor1 = moonMeanNodeCorr[i+1]
	dcor = dcor0 + dfrac*(dcor1-dcor0)
	return dcor
}

// swiMeanNode implements swi_mean_node from swemmoon.c.
//
//	mean lunar node
//
// J		julian day
// pol		return array for position and velocity
//
//	(polar coordinates of ecliptic of date)
func (state *State) swiMeanNode(J float64, pol []float64, message []byte) int32 {
	var (
		s    [256]byte
		dcor float64
	)
	state.lunar.centuries = (J - julianDayJ2000) / daysPerCentury
	state.lunar.centuries2 = state.lunar.centuries * state.lunar.centuries
	state.lunar.centuries3 = state.lunar.centuries * state.lunar.centuries2
	state.lunar.centuries4 = state.lunar.centuries2 * state.lunar.centuries2
	if J < -3100015.5 || J > 8000016.5 {
		if message != nil {
			formatText(s[:], textBytes("jd %f outside mean node range %.2f .. %.2f "), J, -3100015.5, 8000016.5)
			if textLength(message)+textLength(s[:]) < 256 {
				appendText(message, s[:])
			}
		}
		return -1
	}
	state.meanElements()
	dcor = state.corrMeanNode(J) * 3600
	pol[0] = state.swiMod2PI((state.lunar.meanLongitude - state.lunar.latitudeArgument - dcor) * 4.8481368110953599359e-6)
	pol[1] = 0.0
	pol[2] = 0.0025695552897999903
	return 0
}

// corrMeanApog implements corr_mean_apog from swemmoon.c.
//
//	1 jan  17000 greg.
func (state *State) corrMeanApog(J float64) float64 {
	var (
		J0      float64
		dJ      float64
		dayscty float64
		dcor    float64
		dcor0   float64
		dcor1   float64
		dfrac   float64
		i       int32
	)
	J0 = -3063616.5
	dayscty = 36524.25
	if J < -3027215.5 {
		return 0
	}
	if J > 7930192.5 {
		return 0
	}
	dJ = J - J0
	i = int32(math.Floor(dJ / dayscty))
	dfrac = (dJ - float64(i)*dayscty) / dayscty
	dcor0 = moonMeanApsisCorr[i]
	dcor1 = moonMeanApsisCorr[i+1]
	dcor = dcor0 + dfrac*(dcor1-dcor0)
	return dcor
}

// swiMeanApog implements swi_mean_apog from swemmoon.c.
//
//	mean lunar apogee ('dark moon', 'lilith')
//
// J		julian day
// pol		return array for position
//
//	(polar coordinates of ecliptic of date)
//
// serr         error return string
func (state *State) swiMeanApog(J float64, pol []float64, message []byte) int32 {
	var (
		node float64
		dcor float64
		s    [256]byte
	)
	state.lunar.centuries = (J - julianDayJ2000) / daysPerCentury
	state.lunar.centuries2 = state.lunar.centuries * state.lunar.centuries
	state.lunar.centuries3 = state.lunar.centuries * state.lunar.centuries2
	state.lunar.centuries4 = state.lunar.centuries2 * state.lunar.centuries2
	if J < -3100015.5 || J > 8000016.5 {
		if message != nil {
			formatText(s[:], textBytes("jd %f outside mean apogee range %.2f .. %.2f "), J, -3100015.5, 8000016.5)
			if textLength(message)+textLength(s[:]) < 256 {
				appendText(message, s[:])
			}
		}
		return -1
	}
	state.meanElements()
	pol[0] = state.swiMod2PI((state.lunar.meanLongitude-state.lunar.lunarAnomaly)*4.8481368110953599359e-6 + 3.14159265358979323846)
	pol[1] = 0
	pol[2] = 0.0027106251317225464
	dcor = state.corrMeanApog(J) * radiansPerDegree
	pol[0] = state.swiMod2PI(pol[0] - dcor)
	node = (state.lunar.meanLongitude - state.lunar.latitudeArgument) * 4.8481368110953599359e-6
	dcor = state.corrMeanNode(J) * radiansPerDegree
	node = state.swiMod2PI(node - dcor)
	pol[0] = state.swiMod2PI(pol[0] - node)
	state.swiPolcart(pol, pol)
	state.swiCoortrf(pol, pol, -0.089804108500263158)
	state.swiCartpol(pol, pol)
	pol[0] = state.swiMod2PI(pol[0] + node)
	return 0
}

// chewm implements chewm from swemmoon.c.
//
//	Program to step through the perturbation table
func (state *State) chewm(pt []int16, nlines int32, nangles int32, typflg int32, ans []float64) {
	var (
		i  int32
		j  int32
		k  int32
		k1 int32
		m  int32
		cu float64
		su float64
		cv float64
		sv float64
		ff float64
	)
	for i = 0; i < nlines; i++ {
		k1 = 0
		sv = 0.0
		cv = 0.0
		for m = 0; m < nangles; m++ {
			j = int32(nextValue(&pt))
			if j != 0 {
				k = j
				if j < 0 {
					k = -k
				}
				su = state.lunar.sinMultiples[m][k-1]
				cu = state.lunar.cosMultiples[m][k-1]
				if j < 0 {
					su = -su
				}
				if k1 == 0 {
					sv = su
					cv = cu
					k1 = 1
				} else {
					ff = su*cv + cu*sv
					cv = cu*cv - su*sv
					sv = ff
				}
			}
		}
		switch typflg {
		case 1:
			j = int32(nextValue(&pt))
			k = int32(nextValue(&pt))
			ans[0] += (10000.0*float64(j) + float64(k)) * sv
			j = int32(nextValue(&pt))
			k = int32(nextValue(&pt))
			if k != 0 {
				ans[2] += (10000.0*float64(j) + float64(k)) * cv
			}
		case 2:
			j = int32(nextValue(&pt))
			k = int32(nextValue(&pt))
			ans[0] += float64(j) * sv
			ans[2] += float64(k) * cv
		case 3:
			j = int32(nextValue(&pt))
			k = int32(nextValue(&pt))
			ans[1] += (10000.0*float64(j) + float64(k)) * sv
		case 4:
			j = int32(nextValue(&pt))
			ans[1] += float64(j) * sv
		}
	}
}

func (state *State) moonSscc(k int32, arg float64, n int32) {
	var (
		cu float64
		su float64
		cv float64
		sv float64
		s  float64
		i  int32
	)
	su = math.Sin(arg)
	cu = math.Cos(arg)
	state.lunar.sinMultiples[k][0] = su
	state.lunar.cosMultiples[k][0] = cu
	sv = 2.0 * su * cu
	cv = cu*cu - su*su
	state.lunar.sinMultiples[k][1] = sv
	state.lunar.cosMultiples[k][1] = cv
	for i = 2; i < n; i++ {
		s = su*cv + cu*sv
		cv = cu*cv - su*sv
		sv = s
		state.lunar.sinMultiples[k][i] = sv
		state.lunar.cosMultiples[k][i] = cv
	}
}

// ecldatEqu2000 implements ecldat_equ2000 from swemmoon.c.
//
//	converts from polar coordinates of ecliptic of date
//	        to   cartesian coordinates of equator 2000
//
// tjd		date
// x 		array of position
func (state *State) ecldatEqu2000(jd float64, xpm []float64) {
	state.swiPolcart(xpm, xpm)
	state.swiCoortrf2(xpm, xpm, -state.positions.data.obliquity.sine, state.positions.data.obliquity.cosine)
	state.swiPrecess(xpm, jd, 0, 1)
}

// mods3600 implements mods3600 from swemmoon.c.
//
//	Reduce arc seconds modulo 360 degrees
//
// answer in arc seconds
func (state *State) mods3600(x float64) float64 {
	var lx float64
	lx = x
	lx = lx - 1296000.0*math.Floor(lx/1296000.0)
	return lx
}

// swiMeanLunarElements implements swi_mean_lunar_elements from swemmoon.c.
//
//	Reduce arc seconds modulo 360 degrees
//
// answer in arc seconds
func (state *State) swiMeanLunarElements(jd float64, node []float64, dnode []float64, peri []float64, dperi []float64) {
	var dcor float64
	state.lunar.centuries = (jd - julianDayJ2000) / daysPerCentury
	state.lunar.centuries2 = state.lunar.centuries * state.lunar.centuries
	state.meanElements()
	node[0] = state.sweDegnorm((state.lunar.meanLongitude - state.lunar.latitudeArgument) * 4.8481368110953599359e-6 * degreesPerRadian)
	peri[0] = state.sweDegnorm((state.lunar.meanLongitude - state.lunar.lunarAnomaly) * 4.8481368110953599359e-6 * degreesPerRadian)
	state.lunar.centuries -= 2.7378507871321012e-05
	state.meanElements()
	dnode[0] = state.sweDegnorm(node[0] - (state.lunar.meanLongitude-state.lunar.latitudeArgument)*4.8481368110953599359e-6*degreesPerRadian)
	dnode[0] -= 360
	dperi[0] = state.sweDegnorm(peri[0] - (state.lunar.meanLongitude-state.lunar.lunarAnomaly)*4.8481368110953599359e-6*degreesPerRadian)
	dcor = state.corrMeanNode(jd)
	node[0] = state.sweDegnorm(node[0] - dcor)
	dcor = state.corrMeanApog(jd)
	peri[0] = state.sweDegnorm(peri[0] - dcor)
}

// meanElements implements mean_elements from swemmoon.c.
func (state *State) meanElements() {
	var fracT float64 = math.Mod(state.lunar.centuries, 1)
	state.lunar.solarAnomaly = state.mods3600(129600000.0*fracT - 3418.961646*state.lunar.centuries + 1287104.76154)
	state.lunar.solarAnomaly += ((((((((1.62e-20*state.lunar.centuries-1.0390e-17)*state.lunar.centuries-3.83508e-15)*state.lunar.centuries+4.237343e-13)*state.lunar.centuries+8.8555011e-11)*state.lunar.centuries-4.77258489e-8)*state.lunar.centuries-1.1297037031e-5)*state.lunar.centuries+1.4732069041e-4)*state.lunar.centuries - 0.552891801772) * state.lunar.centuries2
	state.lunar.latitudeArgument = state.mods3600(1739232000.0*fracT + 295263.0983*state.lunar.centuries - 2.079419901760e-01*state.lunar.centuries + 335779.55755)
	state.lunar.lunarAnomaly = state.mods3600(1717200000.0*fracT + 715923.4728*state.lunar.centuries - 2.035946368532e-01*state.lunar.centuries + 485868.28096)
	state.lunar.elongation = state.mods3600(1601856000.0*fracT + 1105601.4603*state.lunar.centuries + 3.962893294503e-01*state.lunar.centuries + 1072260.73512)
	state.lunar.meanLongitude = state.mods3600(1731456000.0*fracT + 1108372.83264*state.lunar.centuries - 6.784914260953e-01*state.lunar.centuries + 785939.95571)
	state.lunar.latitudeArgument += ((moonZ[2]*state.lunar.centuries+moonZ[1])*state.lunar.centuries + moonZ[0]) * state.lunar.centuries2
	state.lunar.lunarAnomaly += ((moonZ[5]*state.lunar.centuries+moonZ[4])*state.lunar.centuries + moonZ[3]) * state.lunar.centuries2
	state.lunar.elongation += ((moonZ[8]*state.lunar.centuries+moonZ[7])*state.lunar.centuries + moonZ[6]) * state.lunar.centuries2
	state.lunar.meanLongitude += ((moonZ[11]*state.lunar.centuries+moonZ[10])*state.lunar.centuries + moonZ[9]) * state.lunar.centuries2
}

// meanElementsPl implements mean_elements_pl from swemmoon.c.
//
//	sensitivity of mean elements
//	  delta argument = scale factor times delta amplitude (arcsec)
//
// cos l  9.0019 = mean eccentricity
// cos 2D 43.6
// cos F  11.2 (latitude term)
func (state *State) meanElementsPl() {
	state.lunar.venusLongitude = state.mods3600(210664136.4335482*state.lunar.centuries + 655127.283046)
	state.lunar.venusLongitude += ((((((((-9.36e-023*state.lunar.centuries-1.95e-20)*state.lunar.centuries+6.097e-18)*state.lunar.centuries+4.43201e-15)*state.lunar.centuries+2.509418e-13)*state.lunar.centuries-3.0622898e-10)*state.lunar.centuries-2.26602516e-9)*state.lunar.centuries-1.4244812531e-5)*state.lunar.centuries + 0.005871373088) * state.lunar.centuries2
	state.lunar.earthLongitude = state.mods3600(129597742.26669231*state.lunar.centuries + 361679.214649)
	state.lunar.earthLongitude += ((((((((-1.16e-22*state.lunar.centuries+2.976e-19)*state.lunar.centuries+2.8460e-17)*state.lunar.centuries-1.08402e-14)*state.lunar.centuries-1.226182e-12)*state.lunar.centuries+1.7228268e-10)*state.lunar.centuries+1.515912254e-7)*state.lunar.centuries+8.863982531e-6)*state.lunar.centuries - 2.0199859001e-2) * state.lunar.centuries2
	state.lunar.marsLongitude = state.mods3600(68905077.59284*state.lunar.centuries + 1279559.78866)
	state.lunar.marsLongitude += (-1.043e-5*state.lunar.centuries + 9.38012e-3) * state.lunar.centuries2
	state.lunar.jupiterLongitude = state.mods3600(10925660.428608*state.lunar.centuries + 123665.342120)
	state.lunar.jupiterLongitude += (1.543273e-5*state.lunar.centuries - 3.06037836351e-1) * state.lunar.centuries2
	state.lunar.saturnLongitude = state.mods3600(4399609.65932*state.lunar.centuries + 180278.89694)
	state.lunar.saturnLongitude += ((4.475946e-8*state.lunar.centuries-6.874806e-5)*state.lunar.centuries + 7.56161437443e-1) * state.lunar.centuries2
}

// swiIntpApsides implements swi_intp_apsides from swemmoon.c.
//
//	Calculate geometric coordinates of true interpolated Moon apsides
func (state *State) swiIntpApsides(J float64, pol []float64, bodyIndex int32) int32 {
	var (
		dd    float64
		rsv   [3]float64
		sNF   float64
		sD    float64
		sLP   float64
		sMP   float64
		sM    float64
		sVe   float64
		sEa   float64
		sMa   float64
		sJu   float64
		sSa   float64
		fM    float64
		fVe   float64
		fEa   float64
		fMa   float64
		fJu   float64
		fSa   float64
		cMP   float64
		zMP   float64
		fNF   float64
		fD    float64
		fLP   float64
		dMP   float64
		mLP   float64
		mNF   float64
		mD    float64
		mMP   float64
		i     int32
		ii    int32
		iii   int32
		niter int32 = 4
	)
	ii = 1
	zMP = 27.55454988
	fNF = 27.212220817 / zMP
	fD = 29.530588835 / zMP
	fLP = 27.321582 / zMP
	fM = 365.2596359 / zMP
	fVe = 224.7008001 / zMP
	fEa = 365.2563629 / zMP
	fMa = 686.9798519 / zMP
	fJu = 4332.589348 / zMP
	fSa = 10759.22722 / zMP
	state.lunar.centuries = (J - julianDayJ2000) / daysPerCentury
	state.lunar.centuries2 = state.lunar.centuries * state.lunar.centuries
	state.lunar.centuries4 = state.lunar.centuries2 * state.lunar.centuries2
	state.meanElements()
	state.meanElementsPl()
	sNF = state.lunar.latitudeArgument
	sD = state.lunar.elongation
	sLP = state.lunar.meanLongitude
	sMP = state.lunar.lunarAnomaly
	sM = state.lunar.solarAnomaly
	sVe = state.lunar.venusLongitude
	sEa = state.lunar.earthLongitude
	sMa = state.lunar.marsLongitude
	sJu = state.lunar.jupiterLongitude
	sSa = state.lunar.saturnLongitude
	sNF = state.mods3600(state.lunar.latitudeArgument)
	sD = state.mods3600(state.lunar.elongation)
	sLP = state.mods3600(state.lunar.meanLongitude)
	sMP = state.mods3600(state.lunar.lunarAnomaly)
	if bodyIndex == 5 {
		state.lunar.lunarAnomaly = 0.0
		niter = 5
	}
	if bodyIndex == 4 {
		state.lunar.lunarAnomaly = 648000.0
		niter = 4
	}
	cMP = 0
	dd = 18000.0
	for iii = 0; iii <= niter; iii++ {
		dMP = sMP - state.lunar.lunarAnomaly
		mLP = sLP - dMP
		mNF = sNF - dMP
		mD = sD - dMP
		mMP = sMP - dMP
		for ii = 0; ii <= 2; ii++ {
			state.lunar.lunarAnomaly = mMP + float64(ii-1)*dd
			state.lunar.latitudeArgument = mNF + float64(ii-1)*dd/fNF
			state.lunar.elongation = mD + float64(ii-1)*dd/fD
			state.lunar.meanLongitude = mLP + float64(ii-1)*dd/fLP
			state.lunar.solarAnomaly = sM + float64(ii-1)*dd/fM
			state.lunar.venusLongitude = sVe + float64(ii-1)*dd/fVe
			state.lunar.earthLongitude = sEa + float64(ii-1)*dd/fEa
			state.lunar.marsLongitude = sMa + float64(ii-1)*dd/fMa
			state.lunar.jupiterLongitude = sJu + float64(ii-1)*dd/fJu
			state.lunar.saturnLongitude = sSa + float64(ii-1)*dd/fSa
			state.moon1()
			state.moon2()
			state.moon3()
			state.moon4()
			if ii == 1 {
				for i = 0; i < 3; i++ {
					pol[i] = state.lunar.polar[i]
				}
			}
			rsv[ii] = state.lunar.polar[2]
		}
		cMP = (1.5*rsv[0] - 2*rsv[1] + 0.5*rsv[2]) / (rsv[0] + rsv[2] - 2*rsv[1])
		cMP *= dd
		cMP = cMP - dd
		mMP += cMP
		state.lunar.lunarAnomaly = mMP
		dd /= 10
	}
	return 0
}
