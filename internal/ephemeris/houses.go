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

// sweHouses implements swe_houses from swehouse.c.
//
//	housasp.c
//
// cusps are returned in double cusp[13],
//
//	or cusp[37] with house system 'G'.
//
// cusp[1...12]	houses 1 - 12
// additional points are returned in ascmc[10].
// ascmc[0] = ascendant
// ascmc[1] = mc
// ascmc[2] = armc
// ascmc[3] = vertex
// ascmc[4] = equasc		* "equatorial ascendant" *
// ascmc[5] = coasc1		* "co-ascendant" (W. Koch) *
// ascmc[6] = coasc2		* "co-ascendant" (M. Munkasey) *
// ascmc[7] = polasc		* "polar ascendant" (M. Munkasey) *
func (state *State) sweHouses(jdUT float64, latitude float64, longitude float64, houseSystem int32, cusp []float64, ascmc []float64) int32 {
	var (
		i      int32
		status int32 = 0
		armc   float64
		eps    float64
		nutlo  [2]float64
		tjde   float64 = jdUT + state.sweDeltatEx(jdUT, -1, nil)
	)
	eps = state.swiEpsiln(tjde, 0) * degreesPerRadian
	state.swiNutation(tjde, 0, nutlo[:])
	for i = 0; i < 2; i++ {
		nutlo[i] *= degreesPerRadian
	}
	armc = state.sweDegnorm(state.sweSidtime0(jdUT, eps+nutlo[1], nutlo[0])*15 + longitude)
	if upperByte(houseSystem) == 'I' {
		var (
			flags  int32 = 256 | 2*1024
			xp     [6]float64
			result int32 = state.sweCalcUt(jdUT, 0, flags, xp[:], nil)
		)
		if result < 0 {
			result = state.sweHousesArmcEx2(armc, latitude, eps+nutlo[1], 'O', cusp, ascmc, nil, nil, nil)
			return -1
		}
		ascmc[9] = xp[1]
	}
	status = state.sweHousesArmcEx2(armc, latitude, eps+nutlo[1], houseSystem, cusp, ascmc, nil, nil, nil)
	return status
}

// sweHousesEx implements swe_houses_ex from swehouse.c.
func (state *State) sweHousesEx(jdUT float64, flags int32, latitude float64, longitude float64, houseSystem int32, cusp []float64, ascmc []float64) int32 {
	return state.sweHousesEx2(jdUT, flags, latitude, longitude, houseSystem, cusp, ascmc, nil, nil, nil)
}

// sweHousesEx2 implements swe_houses_ex2 from swehouse.c.
func (state *State) sweHousesEx2(jdUT float64, iflag int32, latitude float64, longitude float64, houseSystem int32, cusp []float64, ascmc []float64, cuspSpeed []float64, ascmcSpeed []float64, message []byte) int32 {
	var (
		i        int32
		status   int32 = 0
		armc     float64
		epsMean  float64
		nutlo    [2]float64
		tjde     float64       = jdUT + state.sweDeltatEx(jdUT, iflag, nil)
		sip      *siderealData = &state.positions.data.sidereal
		xp       [6]float64
		retcMakr int32 = 0
		ito      int32
	)
	if upperByte(houseSystem) == 'G' {
		ito = 36
	} else {
		ito = 12
	}
	if iflag&(64*1024) != 0 && !(state.positions.data.ayanaIsSet != 0) {
		state.sweSetSidMode(0, 0, 0)
	}
	epsMean = state.swiEpsiln(tjde, 0) * degreesPerRadian
	state.swiNutation(tjde, 0, nutlo[:])
	for i = 0; i < 2; i++ {
		nutlo[i] *= degreesPerRadian
	}
	if iflag&64 != 0 {
		for i = 0; i < 2; i++ {
			nutlo[i] = 0
		}
	}
	armc = state.sweDegnorm(state.sweSidtime0(jdUT, epsMean+nutlo[1], nutlo[0])*15 + longitude)
	if upperByte(houseSystem) == 'I' {
		var flags int32 = 256 | 2*1024
		retcMakr = state.sweCalcUt(jdUT, 0, flags, xp[:], nil)
		if retcMakr < 0 {
			houseSystem = 'O'
		}
		ascmc[9] = xp[1]
	}
	if iflag&(64*1024) != 0 {
		if sip.mode&256 != 0 {
			status = state.siderealHousesEclT0(tjde, armc, epsMean+nutlo[1], nutlo[:], latitude, houseSystem, cusp, ascmc, cuspSpeed, ascmcSpeed, message)
		} else if sip.mode&512 != 0 {
			status = state.siderealHousesSsypl(tjde, armc, epsMean+nutlo[1], nutlo[:], latitude, houseSystem, cusp, ascmc, cuspSpeed, ascmcSpeed, message)
		} else {
			status = state.siderealHousesTrad(tjde, iflag, armc, epsMean+nutlo[1], nutlo[0], latitude, houseSystem, cusp, ascmc, cuspSpeed, ascmcSpeed, message)
		}
	} else {
		status = state.sweHousesArmcEx2(armc, latitude, epsMean+nutlo[1], houseSystem, cusp, ascmc, cuspSpeed, ascmcSpeed, message)
		if upperByte(houseSystem) == 'I' {
			ascmc[9] = xp[1]
		}
	}
	if iflag&(8*1024) != 0 {
		for i = 1; i <= ito; i++ {
			cusp[i] *= radiansPerDegree
		}
		for i = 0; i < 8; i++ {
			ascmc[i] *= radiansPerDegree
		}
	}
	if retcMakr < 0 {
		return retcMakr
	}
	return status
}

func (state *State) siderealHousesEclT0(tjde float64, armc float64, eps float64, nutlo []float64, lat float64, houseSystem int32, cusp []float64, ascmc []float64, cuspSpeed []float64, ascmcSpeed []float64, message []byte) int32 {
	var (
		i      int32
		j      int32
		status int32 = 0
		x      [6]float64
		xvpx   [6]float64
		x2     [6]float64
		epst0  float64
		xnorm  [6]float64
		rxy    float64
		rxyz   float64
		c2     float64
		epsx   float64
		sgn    float64
		fac    float64
		dvpx   float64
		dvpxe  float64
		armcx  float64
		sip    *siderealData = &state.positions.data.sidereal
		ito    int32
	)
	if upperByte(houseSystem) == 'G' {
		ito = 36
	} else {
		ito = 12
	}
	epst0 = state.swiEpsiln(sip.epoch, 0)
	x[4] = 1
	x[0] = x[4]
	x[5] = 0
	x[3] = x[5]
	x[2] = x[3]
	x[1] = x[2]
	state.swiCoortrf(x[:], x[:], -epst0)
	state.swiCoortrf(x[3:], x[3:], -epst0)
	state.swiPrecess(x[:], sip.epoch, 0, 1)
	state.swiPrecess(x[:], tjde, 0, -1)
	state.swiPrecess(x[3:], sip.epoch, 0, 1)
	state.swiPrecess(x[3:], tjde, 0, -1)
	state.swiCoortrf(x[:], x[:], (eps-nutlo[1])*radiansPerDegree)
	state.swiCoortrf(x[3:], x[3:], (eps-nutlo[1])*radiansPerDegree)
	state.swiCartpolSp(x[:], x[:])
	x[0] += nutlo[0] * radiansPerDegree
	state.swiPolcartSp(x[:], x[:])
	state.swiCoortrf(x[:], x[:], -eps*radiansPerDegree)
	state.swiCoortrf(x[3:], x[3:], -eps*radiansPerDegree)
	state.swiCrossProd(x[:], x[3:], xnorm[:])
	rxy = xnorm[0]*xnorm[0] + xnorm[1]*xnorm[1]
	c2 = rxy + xnorm[2]*xnorm[2]
	rxyz = math.Sqrt(c2)
	rxy = math.Sqrt(rxy)
	epsx = math.Asin(rxy/rxyz) * degreesPerRadian
	if math.Abs(x[5]) < 1e-15 {
		x[5] = 1e-15
	}
	fac = x[2] / x[5]
	sgn = x[5] / math.Abs(x[5])
	for j = 0; j <= 2; j++ {
		xvpx[j] = (x[j] - fac*x[j+3]) * sgn
	}
	state.swiCartpol(xvpx[:], x2[:])
	dvpx = x2[0] * degreesPerRadian
	armcx = state.sweDegnorm(armc - dvpx)
	status = state.sweHousesArmcEx2(armcx, lat, epsx, houseSystem, cusp, ascmc, cuspSpeed, ascmcSpeed, message)
	dvpxe = math.Acos(state.swiDotProdUnit(x[:], xvpx[:])) * degreesPerRadian
	if tjde < sip.epoch {
		dvpxe = -dvpxe
	}
	for i = 1; i <= ito; i++ {
		cusp[i] = state.sweDegnorm(cusp[i] - dvpxe - sip.ayanamsa)
	}
	for i = 0; i <= 8; i++ {
		if i == 2 {
			continue
		}
		ascmc[i] = state.sweDegnorm(ascmc[i] - dvpxe - sip.ayanamsa)
	}
	if houseSystem == 'N' {
		for i = 1; i <= ito; i++ {
			cusp[i] = float64((i - 1) * 30)
		}
	}
	return status
}

// siderealHousesSsypl implements sidereal_houses_ssypl from swehouse.c.
//
// Be t the birth date and t0 the epoch at which ayanamsa = 0.
//  1. Compute the angle between the solar system rotation plane and
//     the true equator at t.
//     The intersection point of these two circles we call the
//     "auxiliary vernal point", and the angle between them the
//     "auxiliary obliquity".
//  2. Compute the distance of the auxiliary vernal point from the
//     zero point at t. (this is a section on the equator)
//  3. subtract this value from the armc of t = aux. armc.
//  4. Compute the axes and houses for this aux. armc and aux. obliquity.
//  5. Compute the distance between the auxiliary vernal point at t
//     and the zero point of the solar system plane J2000
//     (a section measured on the solar system plane)
//  6. subtract this distance from all house cusps.
//  7. compute the ayanamsa of J2000 on the solar system plane,
//     referred to t0
//  8. subtract ayanamsa_t0 from all house cusps.
//  9. subtract ayanamsa_2000 from all house cusps.
func (state *State) siderealHousesSsypl(tjde float64, armc float64, eps float64, nutlo []float64, lat float64, houseSystem int32, cusp []float64, ascmc []float64, cuspSpeed []float64, ascmcSpeed []float64, message []byte) int32 {
	var (
		i       int32
		j       int32
		status  int32 = 0
		x       [6]float64
		x0      [6]float64
		xvpx    [6]float64
		x2      [6]float64
		xnorm   [6]float64
		rxy     float64
		rxyz    float64
		c2      float64
		epsx    float64
		eps2000 float64
		sgn     float64
		fac     float64
		dvpx    float64
		dvpxe   float64
		x00     float64
		armcx   float64
		sip     *siderealData = &state.positions.data.sidereal
		ito     int32
	)
	if upperByte(houseSystem) == 'G' {
		ito = 36
	} else {
		ito = 12
	}
	eps2000 = state.swiEpsiln(julianDayJ2000, 0)
	x[4] = 1
	x[0] = x[4]
	x[5] = 0
	x[3] = x[5]
	x[2] = x[3]
	x[1] = x[2]
	state.swiCoortrf(x[:], x[:], -0.027553530354526998)
	state.swiCoortrf(x[3:], x[3:], -0.027553530354526998)
	state.swiCartpolSp(x[:], x[:])
	x[0] += 1.8776700468039835
	state.swiPolcartSp(x[:], x[:])
	state.swiCoortrf(x[:], x[:], -eps2000)
	state.swiCoortrf(x[3:], x[3:], -eps2000)
	state.swiPrecess(x[:], tjde, 0, -1)
	state.swiPrecess(x[3:], tjde, 0, -1)
	state.swiCoortrf(x[:], x[:], (eps-nutlo[1])*radiansPerDegree)
	state.swiCoortrf(x[3:], x[3:], (eps-nutlo[1])*radiansPerDegree)
	state.swiCartpolSp(x[:], x[:])
	x[0] += nutlo[0] * radiansPerDegree
	state.swiPolcartSp(x[:], x[:])
	state.swiCoortrf(x[:], x[:], -eps*radiansPerDegree)
	state.swiCoortrf(x[3:], x[3:], -eps*radiansPerDegree)
	state.swiCrossProd(x[:], x[3:], xnorm[:])
	rxy = xnorm[0]*xnorm[0] + xnorm[1]*xnorm[1]
	c2 = rxy + xnorm[2]*xnorm[2]
	rxyz = math.Sqrt(c2)
	rxy = math.Sqrt(rxy)
	epsx = math.Asin(rxy/rxyz) * degreesPerRadian
	if math.Abs(x[5]) < 1e-15 {
		x[5] = 1e-15
	}
	fac = x[2] / x[5]
	sgn = x[5] / math.Abs(x[5])
	for j = 0; j <= 2; j++ {
		xvpx[j] = (x[j] - fac*x[j+3]) * sgn
	}
	state.swiCartpol(xvpx[:], x2[:])
	dvpx = x2[0] * degreesPerRadian
	armcx = state.sweDegnorm(armc - dvpx)
	status = state.sweHousesArmcEx2(armcx, lat, epsx, houseSystem, cusp, ascmc, cuspSpeed, ascmcSpeed, message)
	dvpxe = math.Acos(state.swiDotProdUnit(x[:], xvpx[:])) * degreesPerRadian
	dvpxe -= 107.58883388
	x0[0] = 1
	x0[2] = 0
	x0[1] = x0[2]
	if sip.epoch != julianDayJ2000 {
		state.swiPrecess(x0[:], sip.epoch, 0, 1)
	}
	state.swiCoortrf(x0[:], x0[:], eps2000)
	state.swiCartpol(x0[:], x0[:])
	x0[0] -= 1.8776700468039835
	state.swiPolcart(x0[:], x0[:])
	state.swiCoortrf(x0[:], x0[:], 0.027553530354526998)
	state.swiCartpol(x0[:], x0[:])
	x0[0] += 1.8777793895872257
	x00 = x0[0] * degreesPerRadian
	for i = 1; i <= ito; i++ {
		cusp[i] = state.sweDegnorm(cusp[i] - dvpxe - sip.ayanamsa - x00)
	}
	for i = 0; i <= 8; i++ {
		if i == 2 {
			continue
		}
		ascmc[i] = state.sweDegnorm(ascmc[i] - dvpxe - sip.ayanamsa - x00)
	}
	if houseSystem == 'N' {
		for i = 1; i <= ito; i++ {
			cusp[i] = float64((i - 1) * 30)
		}
	}
	return status
}

// siderealHousesTrad implements sidereal_houses_trad from swehouse.c.
//
//	common simplified procedure
func (state *State) siderealHousesTrad(tjde float64, flags int32, armc float64, eps float64, nutl float64, lat float64, houseSystem int32, cusp []float64, ascmc []float64, cuspSpeed []float64, ascmcSpeed []float64, message []byte) int32 {
	var (
		i      int32
		status int32 = 0
		ay     [1]float64
		ito    int32
		ihs    int32 = upperByte(houseSystem)
		ihs2   int32 = ihs
	)
	status = state.sweGetAyanamsaEx(tjde, flags, ay[:], nil)
	if ihs == 'G' {
		ito = 36
	} else {
		ito = 12
	}
	if ihs == 'W' {
		ihs2 = 'E'
	}
	status = state.sweHousesArmcEx2(armc, lat, eps, ihs2, cusp, ascmc, cuspSpeed, ascmcSpeed, message)
	for i = 1; i <= ito; i++ {
		cusp[i] = state.sweDegnorm(cusp[i] - ay[0])
		if ihs == 'W' {
			cusp[i] -= math.Mod(cusp[i], 30)
		}
	}
	if ihs == 'N' {
		for i = 1; i <= ito; i++ {
			cusp[i] = float64((i - 1) * 30)
		}
	}
	for i = 0; i < 8; i++ {
		if i == 2 {
			continue
		}
		ascmc[i] = state.sweDegnorm(ascmc[i] - ay[0])
	}
	return status
}

// sweHousesArmc implements swe_houses_armc from swehouse.c.
func (state *State) sweHousesArmc(armc float64, latitude float64, eps float64, houseSystem int32, cusp []float64, ascmc []float64) int32 {
	return state.sweHousesArmcEx2(armc, latitude, eps, houseSystem, cusp, ascmc, nil, nil, nil)
}

// sweHousesArmcEx2 implements swe_houses_armc_ex2 from swehouse.c.
func (state *State) sweHousesArmcEx2(armc float64, latitude float64, eps float64, houseSystem int32, cusp []float64, ascmc []float64, cuspSpeed []float64, ascmcSpeed []float64, message []byte) int32 {
	var (
		h      houseData
		hm1    houseData
		hp1    houseData
		i      int32
		status int32 = 0
		rm1    int32
		rp1    int32
		ito    int32
	)
	if upperByte(houseSystem) == 'G' {
		ito = 36
	} else {
		ito = 12
	}
	armc = state.sweDegnorm(armc)
	h.needSpeed = 0
	h.needCuspSpeed = 0
	if ascmcSpeed != nil || cuspSpeed != nil {
		h.needSpeed = 1
	}
	if cuspSpeed != nil {
		h.needCuspSpeed = 1
	}
	if upperByte(houseSystem) == 'I' {
		if ascmc[9] == 99 {
			h.sunDeclination = 0
			if state.houses.savedSunDeclination != 99 {
				h.sunDeclination = state.houses.savedSunDeclination
			}
		} else {
			h.sunDeclination = ascmc[9]
			state.houses.savedSunDeclination = h.sunDeclination
		}
		if h.sunDeclination < float64(-24) || h.sunDeclination > 24 {
			formatText(message, textBytes("House system I (Sunshine) needs valid Sun declination in ascmc[9]"))
			return -1
		}
	}
	status = state.CalcH(armc, latitude, eps, byte(houseSystem), &h)
	cusp[0] = 0
	if h.needCuspSpeed != 0 {
		cuspSpeed[0] = 0
	}
	if status < 0 {
		ito = 12
		if message != nil {
			copyText(message, h.message[:])
		}
	}
	for i = 1; i <= ito; i++ {
		cusp[i] = h.cusp[i]
		if h.needCuspSpeed != 0 {
			cuspSpeed[i] = h.cuspSpeed[i]
		}
	}
	ascmc[0] = h.ascendant
	ascmc[1] = h.midheaven
	ascmc[2] = armc
	ascmc[3] = h.vertex
	ascmc[4] = h.equatorialAscendant
	ascmc[5] = h.coAscendantKoch
	ascmc[6] = h.coAscendantMunkasey
	ascmc[7] = h.polarAscendant
	for i = 8; i < 10; i++ {
		ascmc[i] = 0
	}
	if upperByte(houseSystem) == 'I' {
		ascmc[9] = h.sunDeclination
	}
	if h.needSpeed != 0 && ascmcSpeed != nil {
		ascmcSpeed[0] = h.ascendantSpeed
		ascmcSpeed[1] = h.midheavenSpeed
		ascmcSpeed[2] = h.armcSpeed
		ascmcSpeed[3] = h.vertexSpeed
		ascmcSpeed[4] = h.equatorialAscendantSpeed
		ascmcSpeed[5] = h.coAscendantKochSpeed
		ascmcSpeed[6] = h.coAscendantMunkaseySpeed
		ascmcSpeed[7] = h.polarAscendantSpeed
		for i = 8; i < 10; i++ {
			ascmcSpeed[i] = 0
		}
	}
	if h.needInterpolation != 0 {
		var (
			dt    float64 = 1.1574074074074073e-05
			darmc float64 = dt * 360.98564733498665
		)
		hm1.needSpeed = 0
		hm1.needCuspSpeed = 0
		hp1.needSpeed = 0
		hp1.needCuspSpeed = 0
		if upperByte(houseSystem) == 'I' {
			hm1.sunDeclination = h.sunDeclination
			hp1.sunDeclination = h.sunDeclination
		}
		rm1 = state.CalcH(armc-darmc, latitude, eps, byte(houseSystem), &hm1)
		rp1 = state.CalcH(armc+darmc, latitude, eps, byte(houseSystem), &hp1)
		if rp1 >= 0 && rm1 >= 0 {
			if math.Abs(state.sweDifdeg2n(hp1.ascendant, h.ascendant)) > 90 {
				hp1 = h
				dt = dt / 2
			} else if math.Abs(state.sweDifdeg2n(hm1.ascendant, h.ascendant)) > 90 {
				hm1 = h
				dt = dt / 2
			}
			for i = 1; i <= 12; i++ {
				var dx float64 = state.sweDifdeg2n(hp1.cusp[i], hm1.cusp[i])
				cuspSpeed[i] = dx / 2 / dt
			}
		}
	}
	return status
}

// apcSector implements apc_sector from swehouse.c.
//
//	n  number of house
//
// ph geographic latitude
// e  ecliptic obliquity
// az armc
func (state *State) apcSector(n int32, ph float64, e float64, az float64) float64 {
	var (
		k          int32
		isBelowHor int32 = 0
		kv         float64
		a          float64
		dasc       float64
		dret       float64
	)
	if math.Abs(ph*degreesPerRadian) > 89.999999999899998 {
		kv = 0
		dasc = 0
	} else {
		kv = math.Atan(math.Tan(ph) * math.Tan(e) * math.Cos(az) / (1 + math.Tan(ph)*math.Tan(e)*math.Sin(az)))
		if math.Abs(ph*degreesPerRadian) < 1e-10 {
			dasc = 1.5707963267931513
			if ph < 0 {
				dasc = -dasc
			}
		} else {
			dasc = math.Atan(math.Sin(kv) / math.Tan(ph))
		}
	}
	if n < 8 {
		isBelowHor = 1
		k = n - 1
	} else {
		k = n - 13
	}
	if isBelowHor != 0 {
		a = kv + az + 1.5707963267948966 + float64(k)*(1.5707963267948966-kv)/3
	} else {
		a = kv + az + 1.5707963267948966 + float64(k)*(1.5707963267948966+kv)/3
	}
	a = state.sweRadnorm(a)
	dret = math.Atan2(math.Tan(dasc)*math.Tan(ph)*math.Sin(az)+math.Sin(a), math.Cos(e)*(math.Tan(dasc)*math.Tan(ph)*math.Cos(az)+math.Cos(a))+math.Sin(e)*math.Tan(ph)*math.Sin(az-a))
	dret = state.sweDegnorm(dret * degreesPerRadian)
	return dret
}

// sweHouseName implements swe_house_name from swehouse.c.
func (state *State) sweHouseName(houseSystem int32) []byte {
	var h int32 = houseSystem
	if h != 'i' {
		h = upperByte(h)
	}
	switch h {
	case 'A':
		return textBytes("equal")
	case 'B':
		return textBytes("Alcabitius")
	case 'C':
		return textBytes("Campanus")
	case 'D':
		return textBytes("equal (MC)")
	case 'E':
		return textBytes("equal")
	case 'F':
		return textBytes("Carter poli-equ.")
	case 'G':
		return textBytes("Gauquelin sectors")
	case 'H':
		return textBytes("horizon/azimut")
	case 'I':
		return textBytes("Sunshine")
	case 'i':
		return textBytes("Sunshine/alt.")
	case 'J':
		return textBytes("Savard-A")
	case 'K':
		return textBytes("Koch")
	case 'L':
		return textBytes("Pullen SD")
	case 'M':
		return textBytes("Morinus")
	case 'N':
		return textBytes("equal/1=Aries")
	case 'O':
		return textBytes("Porphyry")
	case 'Q':
		return textBytes("Pullen SR")
	case 'R':
		return textBytes("Regiomontanus")
	case 'S':
		return textBytes("Sripati")
	case 'T':
		return textBytes("Polich/Page")
	case 'U':
		return textBytes("Krusinski-Pisa-Goelzer")
	case 'V':
		return textBytes("equal/Vehlow")
	case 'W':
		return textBytes("equal/ whole sign")
	case 'X':
		return textBytes("axial rotation system/Meridian houses")
	case 'Y':
		return textBytes("APC houses")
	default:
		return textBytes("Placidus")
	}
}

// swiArmcToMc implements swi_armc_to_mc from swehouse.c.
func (state *State) swiArmcToMc(armc float64, eps float64) float64 {
	var (
		tant float64
		mc   float64
	)
	if math.Abs(armc-90) > 1e-10 && math.Abs(armc-270) > 1e-10 {
		tant = math.Tan(armc * radiansPerDegree)
		mc = math.Atan(tant/math.Cos(eps*radiansPerDegree)) * degreesPerRadian
		if armc > 90 && armc <= 270 {
			mc = state.sweDegnorm(mc + 180)
		}
	} else if math.Abs(armc-90) <= 1e-10 {
		mc = 90
	} else {
		mc = 270
	}
	return mc
}

func (state *State) CalcH(th float64, fi float64, ekl float64, hsy byte, hsp *houseData) int32 {
	var (
		tane         float64
		tanfi        float64
		cosfi        float64
		sinfi        float64
		tant         float64
		sina         float64
		cosa         float64
		th2          float64
		a            float64
		c            float64
		f            float64
		fh1          float64
		fh2          float64
		xh1          float64
		xh2          float64
		xs1          float64
		xs2          float64
		rectasc      float64
		ad3          float64
		acmc         float64
		vemc         float64
		i            int32
		ih           int32
		ih2          int32
		status       int32 = 0
		sine         float64
		cose         float64
		x            [3]float64
		krHorizonLon float64
		niterMax     int32 = 100
		cuspsv       float64
	)
	hsp.message[0] = byte('\000')
	hsp.needInterpolation = 0
	cose = math.Cos(ekl * radiansPerDegree)
	sine = math.Sin(ekl * radiansPerDegree)
	tane = math.Tan(ekl * radiansPerDegree)
	if math.Abs(math.Abs(fi)-90) < 1e-10 {
		if fi < 0 {
			fi = -89.999999999899998
		} else {
			fi = 89.999999999899998
		}
	}
	tanfi = math.Tan(fi * radiansPerDegree)
	if math.Abs(th-90) > 1e-10 && math.Abs(th-270) > 1e-10 {
		tant = math.Tan(th * radiansPerDegree)
		hsp.midheaven = math.Atan(tant/cose) * degreesPerRadian
		if th > 90 && th <= 270 {
			hsp.midheaven = state.sweDegnorm(hsp.midheaven + 180)
		}
	} else if math.Abs(th-90) <= 1e-10 {
		hsp.midheaven = 90
	} else {
		hsp.midheaven = 270
	}
	hsp.midheaven = state.sweDegnorm(hsp.midheaven)
	if hsp.needSpeed != 0 {
		hsp.midheavenSpeed = state.AscDash(th, 0, sine, cose)
	}
	hsp.ascendant = state.Asc1(th+90, fi, sine, cose)
	if hsp.needSpeed != 0 {
		hsp.ascendantSpeed = state.AscDash(th+90, fi, sine, cose)
	}
	if hsp.needCuspSpeed != 0 {
		for i = 0; i <= 12; i++ {
			hsp.cuspSpeed[i] = 0
		}
	}
	hsp.armcSpeed = 360.98564733498665
	hsp.cusp[1] = hsp.ascendant
	hsp.cusp[10] = hsp.midheaven
	if hsp.needCuspSpeed != 0 {
		hsp.cuspSpeed[1] = hsp.ascendantSpeed
		hsp.cuspSpeed[10] = hsp.midheavenSpeed
	}
	if int32(hsy) > 95 && int32(hsy) != 'i' {
		formatText(hsp.message[:], textBytes("use of lower case letters like %c for house systems is deprecated"), hsy)
		hsy = byte(int32(hsy) - 32)
	}
	selection1 := hsy
selectCase1:
	for {
		switch selection1 {
		case 'A':
			fallthrough
		case 'E':
			acmc = state.sweDifdeg2n(hsp.ascendant, hsp.midheaven)
			if acmc < 0 {
				hsp.ascendant = state.sweDegnorm(hsp.ascendant + 180)
				hsp.cusp[1] = hsp.ascendant
			}
			i = 2
			for ; i <= 12; i++ {
				hsp.cusp[i] = state.sweDegnorm(hsp.cusp[1] + float64((i-1)*30))
			}
			if hsp.needCuspSpeed != 0 {
				for i = 1; i <= 12; i++ {
					hsp.cuspSpeed[i] = hsp.ascendantSpeed
				}
			}
		case 'D':
			acmc = state.sweDifdeg2n(hsp.ascendant, hsp.midheaven)
			if acmc < 0 {
				hsp.ascendant = state.sweDegnorm(hsp.ascendant + 180)
			}
			hsp.cusp[10] = hsp.midheaven
			i = 11
			for ; i <= 12; i++ {
				hsp.cusp[i] = state.sweDegnorm(hsp.cusp[10] + float64((i-10)*30))
			}
			i = 1
			for ; i <= 9; i++ {
				hsp.cusp[i] = state.sweDegnorm(hsp.cusp[10] + float64((i+2)*30))
			}
			if hsp.needCuspSpeed != 0 {
				for i = 1; i <= 12; i++ {
					hsp.cuspSpeed[i] = hsp.midheavenSpeed
				}
			}
		case 'C':
			fh1 = math.Asin(math.Sin(fi*radiansPerDegree)/2) * degreesPerRadian
			fh2 = math.Asin(math.Sqrt(3.0)/2*math.Sin(fi*radiansPerDegree)) * degreesPerRadian
			cosfi = math.Cos(fi * radiansPerDegree)
			if math.Abs(cosfi) == 0 {
				if fi > 0 {
					xh2 = 90
					xh1 = xh2
				} else {
					xh2 = 270
					xh1 = xh2
				}
			} else {
				xh1 = math.Atan(math.Sqrt(3.0)/cosfi) * degreesPerRadian
				xh2 = math.Atan(1/math.Sqrt(3.0)/cosfi) * degreesPerRadian
			}
			hsp.cusp[11] = state.Asc1(th+90-xh1, fh1, sine, cose)
			hsp.cusp[12] = state.Asc1(th+90-xh2, fh2, sine, cose)
			hsp.cusp[2] = state.Asc1(th+90+xh2, fh2, sine, cose)
			hsp.cusp[3] = state.Asc1(th+90+xh1, fh1, sine, cose)
			if hsp.needCuspSpeed != 0 {
				hsp.cuspSpeed[11] = state.AscDash(th+90-xh1, fh1, sine, cose)
				hsp.cuspSpeed[12] = state.AscDash(th+90-xh2, fh2, sine, cose)
				hsp.cuspSpeed[2] = state.AscDash(th+90+xh2, fh2, sine, cose)
				hsp.cuspSpeed[3] = state.AscDash(th+90+xh1, fh1, sine, cose)
			}
			if math.Abs(fi) >= 90-ekl {
				acmc = state.sweDifdeg2n(hsp.ascendant, hsp.midheaven)
				if acmc < 0 {
					hsp.ascendant = state.sweDegnorm(hsp.ascendant + 180)
					hsp.midheaven = state.sweDegnorm(hsp.midheaven + 180)
					for i = 1; i <= 12; i++ {
						if i >= 4 && i < 10 {
							continue
						}
						hsp.cusp[i] = state.sweDegnorm(hsp.cusp[i] + 180)
					}
				}
			}
		case 'H':
			if fi > 0 {
				fi = 90 - fi
			} else {
				fi = float64(-90) - fi
			}
			if math.Abs(math.Abs(fi)-90) < 1e-10 {
				if fi < 0 {
					fi = -89.999999999899998
				} else {
					fi = 89.999999999899998
				}
			}
			th = state.sweDegnorm(th + 180)
			fh1 = math.Asin(math.Sin(fi*radiansPerDegree)/2) * degreesPerRadian
			fh2 = math.Asin(math.Sqrt(3.0)/2*math.Sin(fi*radiansPerDegree)) * degreesPerRadian
			cosfi = math.Cos(fi * radiansPerDegree)
			if math.Abs(cosfi) == 0 {
				if fi > 0 {
					xh2 = 90
					xh1 = xh2
				} else {
					xh2 = 270
					xh1 = xh2
				}
			} else {
				xh1 = math.Atan(math.Sqrt(3.0)/cosfi) * degreesPerRadian
				xh2 = math.Atan(1/math.Sqrt(3.0)/cosfi) * degreesPerRadian
			}
			hsp.cusp[11] = state.Asc1(th+90-xh1, fh1, sine, cose)
			hsp.cusp[12] = state.Asc1(th+90-xh2, fh2, sine, cose)
			hsp.cusp[1] = state.Asc1(th+90, fi, sine, cose)
			hsp.cusp[2] = state.Asc1(th+90+xh2, fh2, sine, cose)
			hsp.cusp[3] = state.Asc1(th+90+xh1, fh1, sine, cose)
			if hsp.needCuspSpeed != 0 {
				hsp.cuspSpeed[11] = state.AscDash(th+90-xh1, fh1, sine, cose)
				hsp.cuspSpeed[12] = state.AscDash(th+90-xh2, fh2, sine, cose)
				hsp.cuspSpeed[1] = state.AscDash(th+90, fi, sine, cose)
				hsp.cuspSpeed[2] = state.AscDash(th+90+xh2, fh2, sine, cose)
				hsp.cuspSpeed[3] = state.AscDash(th+90+xh1, fh1, sine, cose)
			}
			if math.Abs(fi) >= 90-ekl {
				acmc = state.sweDifdeg2n(hsp.ascendant, hsp.midheaven)
				if acmc < 0 {
					hsp.ascendant = state.sweDegnorm(hsp.ascendant + 180)
					hsp.midheaven = state.sweDegnorm(hsp.midheaven + 180)
					for i = 1; i <= 12; i++ {
						if i >= 4 && i < 10 {
							continue
						}
						hsp.cusp[i] = state.sweDegnorm(hsp.cusp[i] + 180)
					}
				}
			}
			i = 1
			for ; i <= 3; i++ {
				hsp.cusp[i] = state.sweDegnorm(hsp.cusp[i] + 180)
			}
			i = 11
			for ; i <= 12; i++ {
				hsp.cusp[i] = state.sweDegnorm(hsp.cusp[i] + 180)
			}
			if fi > 0 {
				fi = 90 - fi
			} else {
				fi = float64(-90) - fi
			}
			th = state.sweDegnorm(th + 180)
			acmc = state.sweDifdeg2n(hsp.ascendant, hsp.midheaven)
			if acmc < 0 {
				hsp.ascendant = state.sweDegnorm(hsp.ascendant + 180)
			}
		case 'I':
			fallthrough
		case 'i':
			acmc = state.sweDifdeg2n(hsp.ascendant, hsp.midheaven)
			if acmc < 0 {
				hsp.ascendant = state.sweDegnorm(hsp.ascendant + 180)
				hsp.cusp[1] = hsp.ascendant
				if !(0 != 0) && int32(hsy) == 'I' {
					hsp.midheaven = state.sweDegnorm(hsp.midheaven + 180)
					hsp.cusp[10] = hsp.midheaven
				}
			}
			hsp.cusp[4] = state.sweDegnorm(hsp.cusp[10] + 180)
			hsp.cusp[7] = state.sweDegnorm(hsp.cusp[1] + 180)
			if int32(hsy) == 'I' {
				status = state.sunshineSolutionTreindl(th, fi, ekl, hsp)
			} else {
				status = state.sunshineSolutionMakransky(th, fi, ekl, hsp)
			}
			if status == -1 {
				copyText(hsp.message[:], textBytes("within polar circle, switched to Porphyry"))
				hsy = byte('O')
				selection1 = 'O'
				continue selectCase1
			}
			hsp.needInterpolation = hsp.needCuspSpeed
		case 'J':
			sinfi = math.Sin(fi * radiansPerDegree)
			cosfi = math.Cos(fi * radiansPerDegree)
			if math.Abs(fi) < 1e-10 {
				xs2 = 0.33333333333333331
				xs1 = 0.66666666666666663
			} else {
				xs2 = math.Sin(fi/3*radiansPerDegree) / sinfi
				xs1 = math.Sin(2*fi/3*radiansPerDegree) / sinfi
			}
			xs2 = math.Asin(xs2) * degreesPerRadian
			xs1 = math.Asin(xs1) * degreesPerRadian
			if cosfi == 0 {
				if fi > 0 {
					xh2 = 90
					xh1 = xh2
				} else {
					xh2 = 270
					xh1 = xh2
				}
			} else {
				xh1 = math.Atan(math.Tan(xs1*radiansPerDegree)/cosfi) * degreesPerRadian
				xh2 = math.Atan(math.Tan(xs2*radiansPerDegree)/cosfi) * degreesPerRadian
			}
			fh1 = math.Asin(math.Sin(fi*radiansPerDegree)*math.Sin((90-xs1)*radiansPerDegree)) * degreesPerRadian
			fh2 = math.Asin(math.Sin(fi*radiansPerDegree)*math.Sin((90-xs2)*radiansPerDegree)) * degreesPerRadian
			hsp.cusp[12] = state.Asc1(th+90-xh2, fh2, sine, cose)
			hsp.cusp[11] = state.Asc1(th+90-xh1, fh1, sine, cose)
			hsp.cusp[2] = state.Asc1(th+90+xh2, fh2, sine, cose)
			hsp.cusp[3] = state.Asc1(th+90+xh1, fh1, sine, cose)
			if hsp.needCuspSpeed != 0 {
				hsp.cuspSpeed[11] = state.AscDash(th+90-xh1, fh1, sine, cose)
				hsp.cuspSpeed[12] = state.AscDash(th+90-xh2, fh2, sine, cose)
				hsp.cuspSpeed[3] = state.AscDash(th+90+xh1, fh1, sine, cose)
				hsp.cuspSpeed[2] = state.AscDash(th+90+xh2, fh2, sine, cose)
			}
			if math.Abs(fi) >= 90-ekl {
				acmc = state.sweDifdeg2n(hsp.ascendant, hsp.midheaven)
				if acmc < 0 {
					hsp.ascendant = state.sweDegnorm(hsp.ascendant + 180)
					hsp.midheaven = state.sweDegnorm(hsp.midheaven + 180)
					for i = 1; i <= 12; i++ {
						if i >= 4 && i < 10 {
							continue
						}
						hsp.cusp[i] = state.sweDegnorm(hsp.cusp[i] + 180)
					}
				}
			}
		case 'K':
			if math.Abs(fi) >= 90-ekl {
				status = -1
				copyText(hsp.message[:], textBytes("within polar circle, switched to Porphyry"))
				selection1 = 'O'
				continue selectCase1
			}
			sina = math.Sin(hsp.midheaven*radiansPerDegree) * sine / math.Cos(fi*radiansPerDegree)
			if sina > 1 {
				sina = 1
			}
			if sina < float64(-1) {
				sina = float64(-1)
			}
			cosa = math.Sqrt(1 - sina*sina)
			c = math.Atan(tanfi/cosa) * degreesPerRadian
			ad3 = math.Asin(math.Sin(c*radiansPerDegree)*sina) * degreesPerRadian / 3.0
			hsp.cusp[11] = state.Asc1(th+30-2*ad3, fi, sine, cose)
			hsp.cusp[12] = state.Asc1(th+60-ad3, fi, sine, cose)
			hsp.cusp[2] = state.Asc1(th+120+ad3, fi, sine, cose)
			hsp.cusp[3] = state.Asc1(th+150+2*ad3, fi, sine, cose)
			if hsp.needCuspSpeed != 0 {
				hsp.cuspSpeed[11] = state.AscDash(th+30-2*ad3, fi, sine, cose)
				hsp.cuspSpeed[12] = state.AscDash(th+60-ad3, fi, sine, cose)
				hsp.cuspSpeed[2] = state.AscDash(th+120+ad3, fi, sine, cose)
				hsp.cuspSpeed[3] = state.AscDash(th+150+2*ad3, fi, sine, cose)
			}
		case 'L':
			{
				var (
					d  float64
					q1 float64
				)
				acmc = state.sweDifdeg2n(hsp.ascendant, hsp.midheaven)
				if acmc < 0 {
					hsp.ascendant = state.sweDegnorm(hsp.ascendant + 180)
					hsp.cusp[1] = hsp.ascendant
					acmc = state.sweDifdeg2n(hsp.ascendant, hsp.midheaven)
				}
				q1 = 180 - acmc
				d = (acmc - 90) / 4.0
				if acmc <= 30 {
					hsp.cusp[12] = state.sweDegnorm(hsp.midheaven + acmc/2)
					hsp.cusp[11] = hsp.cusp[12]
				} else {
					hsp.cusp[11] = state.sweDegnorm(hsp.midheaven + 30 + d)
					hsp.cusp[12] = state.sweDegnorm(hsp.midheaven + 60 + 3*d)
				}
				d = (q1 - 90) / 4.0
				if q1 <= 30 {
					hsp.cusp[3] = state.sweDegnorm(hsp.ascendant + q1/2)
					hsp.cusp[2] = hsp.cusp[3]
				} else {
					hsp.cusp[2] = state.sweDegnorm(hsp.ascendant + 30 + d)
					hsp.cusp[3] = state.sweDegnorm(hsp.ascendant + 60 + 3*d)
				}
			}
			hsp.needInterpolation = hsp.needCuspSpeed
		case 'N':
			acmc = state.sweDifdeg2n(hsp.ascendant, hsp.midheaven)
			if acmc < 0 {
				hsp.ascendant = state.sweDegnorm(hsp.ascendant + 180)
			}
			i = 1
			for ; i <= 12; i++ {
				hsp.cusp[i] = float64(i-1) * 30.0
			}
		case 'O':
			acmc = state.sweDifdeg2n(hsp.ascendant, hsp.midheaven)
			if acmc < 0 {
				hsp.ascendant = state.sweDegnorm(hsp.ascendant + 180)
				hsp.cusp[1] = hsp.ascendant
				acmc = state.sweDifdeg2n(hsp.ascendant, hsp.midheaven)
			}
			hsp.cusp[1] = hsp.ascendant
			hsp.cusp[10] = hsp.midheaven
			hsp.cusp[2] = state.sweDegnorm(hsp.ascendant + (180-acmc)/3)
			hsp.cusp[3] = state.sweDegnorm(hsp.ascendant + (180-acmc)/3*2)
			hsp.cusp[11] = state.sweDegnorm(hsp.midheaven + acmc/3)
			hsp.cusp[12] = state.sweDegnorm(hsp.midheaven + acmc/3*2)
			if hsp.needCuspSpeed != 0 {
				var q1Speed float64 = hsp.ascendantSpeed - hsp.midheavenSpeed
				hsp.cuspSpeed[1] = hsp.ascendantSpeed
				hsp.cuspSpeed[10] = hsp.midheavenSpeed
				hsp.cuspSpeed[2] = hsp.ascendantSpeed - q1Speed/3
				hsp.cuspSpeed[3] = hsp.ascendantSpeed - q1Speed/3*2
				hsp.cuspSpeed[11] = hsp.ascendantSpeed + q1Speed/3
				hsp.cuspSpeed[12] = hsp.ascendantSpeed + q1Speed/3*2
			}
		case 'Q':
			{
				var (
					q     float64
					c     float64
					csq   float64
					ccr   float64
					cqx   float64
					two23 float64
					third float64
					r     float64
					r1    float64
					r2    float64
					x     float64
					xr    float64
					xr3   float64
					xr4   float64
				)
				third = 0.33333333333333331
				two23 = math.Pow(4, third)
				acmc = state.sweDifdeg2n(hsp.ascendant, hsp.midheaven)
				if acmc < 0 {
					hsp.ascendant = state.sweDegnorm(hsp.ascendant + 180)
					hsp.cusp[1] = hsp.ascendant
					acmc = state.sweDifdeg2n(hsp.ascendant, hsp.midheaven)
				}
				q = acmc
				if q > 90 {
					q = 180 - q
				}
				if q < 1e-30 {
					xr3 = 0
					xr = xr3
					x = xr
					xr4 = 180
				} else {
					c = (180 - q) / q
					csq = c * c
					ccr = math.Pow(csq-c, third)
					cqx = math.Sqrt(two23*ccr + 1.0)
					r1 = 0.5 * cqx
					r2 = 0.5 * math.Sqrt(float64(-2)*(1-2*c)/cqx-two23*ccr+2)
					r = r1 + r2 - 0.5
					x = q / (2*r + 1)
					xr = r * x
					xr3 = xr * r * r
					xr4 = xr3 * r
				}
				if acmc > 90 {
					hsp.cusp[11] = state.sweDegnorm(hsp.midheaven + xr3)
					hsp.cusp[12] = state.sweDegnorm(hsp.cusp[11] + xr4)
					hsp.cusp[2] = state.sweDegnorm(hsp.ascendant + xr)
					hsp.cusp[3] = state.sweDegnorm(hsp.cusp[2] + x)
				} else {
					hsp.cusp[11] = state.sweDegnorm(hsp.midheaven + xr)
					hsp.cusp[12] = state.sweDegnorm(hsp.cusp[11] + x)
					hsp.cusp[2] = state.sweDegnorm(hsp.ascendant + xr3)
					hsp.cusp[3] = state.sweDegnorm(hsp.cusp[2] + xr4)
				}
			}
			hsp.needInterpolation = hsp.needCuspSpeed
		case 'R':
			fh1 = math.Atan(tanfi*0.5) * degreesPerRadian
			fh2 = math.Atan(tanfi*math.Cos(0.52359877559829882)) * degreesPerRadian
			hsp.cusp[11] = state.Asc1(30+th, fh1, sine, cose)
			hsp.cusp[12] = state.Asc1(60+th, fh2, sine, cose)
			hsp.cusp[2] = state.Asc1(120+th, fh2, sine, cose)
			hsp.cusp[3] = state.Asc1(150+th, fh1, sine, cose)
			if hsp.needCuspSpeed != 0 {
				hsp.cuspSpeed[11] = state.AscDash(30+th, fh1, sine, cose)
				hsp.cuspSpeed[12] = state.AscDash(60+th, fh2, sine, cose)
				hsp.cuspSpeed[2] = state.AscDash(120+th, fh2, sine, cose)
				hsp.cuspSpeed[3] = state.AscDash(150+th, fh1, sine, cose)
			}
			if math.Abs(fi) >= 90-ekl {
				acmc = state.sweDifdeg2n(hsp.ascendant, hsp.midheaven)
				if acmc < 0 {
					hsp.ascendant = state.sweDegnorm(hsp.ascendant + 180)
					hsp.midheaven = state.sweDegnorm(hsp.midheaven + 180)
					for i = 1; i <= 12; i++ {
						if i >= 4 && i < 10 {
							continue
						}
						hsp.cusp[i] = state.sweDegnorm(hsp.cusp[i] + 180)
					}
				}
			}
		case 'S':
			{
				var (
					s1 float64
					s4 float64
					q1 float64
				)
				acmc = state.sweDifdeg2n(hsp.ascendant, hsp.midheaven)
				if acmc < 0 {
					hsp.ascendant = state.sweDegnorm(hsp.ascendant + 180)
					acmc = state.sweDifdeg2n(hsp.ascendant, hsp.midheaven)
				}
				q1 = 180 - acmc
				s1 = q1 / 3.0
				s4 = acmc / 3.0
				hsp.cusp[1] = state.sweDegnorm(hsp.ascendant - s4*0.5)
				hsp.cusp[2] = state.sweDegnorm(hsp.ascendant + s1*0.5)
				hsp.cusp[3] = state.sweDegnorm(hsp.ascendant + s1*1.5)
				hsp.cusp[10] = state.sweDegnorm(hsp.midheaven - s1*0.5)
				hsp.cusp[11] = state.sweDegnorm(hsp.midheaven + s4*0.5)
				hsp.cusp[12] = state.sweDegnorm(hsp.midheaven + s4*1.5)
			}
			hsp.needInterpolation = hsp.needCuspSpeed
		case 'T':
			fh1 = math.Atan(tanfi/3.0) * degreesPerRadian
			fh2 = math.Atan(tanfi*2.0/3.0) * degreesPerRadian
			hsp.cusp[11] = state.Asc1(30+th, fh1, sine, cose)
			hsp.cusp[12] = state.Asc1(60+th, fh2, sine, cose)
			hsp.cusp[2] = state.Asc1(120+th, fh2, sine, cose)
			hsp.cusp[3] = state.Asc1(150+th, fh1, sine, cose)
			if hsp.needCuspSpeed != 0 {
				hsp.cuspSpeed[11] = state.AscDash(30+th, fh1, sine, cose)
				hsp.cuspSpeed[12] = state.AscDash(60+th, fh2, sine, cose)
				hsp.cuspSpeed[2] = state.AscDash(120+th, fh2, sine, cose)
				hsp.cuspSpeed[3] = state.AscDash(150+th, fh1, sine, cose)
			}
			if math.Abs(fi) >= 90-ekl {
				acmc = state.sweDifdeg2n(hsp.ascendant, hsp.midheaven)
				if acmc < 0 {
					hsp.ascendant = state.sweDegnorm(hsp.ascendant + 180)
					hsp.midheaven = state.sweDegnorm(hsp.midheaven + 180)
					for i = 1; i <= 12; i++ {
						hsp.cusp[i] = state.sweDegnorm(hsp.cusp[i] + 180)
					}
				}
			}
		case 'V':
			acmc = state.sweDifdeg2n(hsp.ascendant, hsp.midheaven)
			if acmc < 0 {
				hsp.ascendant = state.sweDegnorm(hsp.ascendant + 180)
			}
			hsp.cusp[1] = state.sweDegnorm(hsp.ascendant - 15)
			i = 2
			for ; i <= 12; i++ {
				hsp.cusp[i] = state.sweDegnorm(hsp.cusp[1] + float64((i-1)*30))
			}
			if hsp.needCuspSpeed != 0 {
				for i = 1; i <= 12; i++ {
					hsp.cuspSpeed[i] = hsp.ascendantSpeed
				}
			}
		case 'W':
			acmc = state.sweDifdeg2n(hsp.ascendant, hsp.midheaven)
			if acmc < 0 {
				hsp.ascendant = state.sweDegnorm(hsp.ascendant + 180)
				hsp.cusp[1] = hsp.ascendant
			}
			hsp.cusp[1] = hsp.ascendant - math.Mod(hsp.ascendant, 30)
			i = 2
			for ; i <= 12; i++ {
				hsp.cusp[i] = state.sweDegnorm(hsp.cusp[1] + float64((i-1)*30))
			}
		case 'X':
			{
				var (
					j int32
					a float64 = th
				)
				for i = 1; i <= 12; i++ {
					j = i + 10
					if j > 12 {
						j -= 12
					}
					a = state.sweDegnorm(a + 30)
					if math.Abs(a-90) > 1e-10 && math.Abs(a-270) > 1e-10 {
						tant = math.Tan(a * radiansPerDegree)
						hsp.cusp[j] = math.Atan(tant/cose) * degreesPerRadian
						if a > 90 && a <= 270 {
							hsp.cusp[j] = state.sweDegnorm(hsp.cusp[j] + 180)
						}
					} else if math.Abs(a-90) <= 1e-10 {
						hsp.cusp[j] = 90
					} else {
						hsp.cusp[j] = 270
					}
					hsp.cusp[j] = state.sweDegnorm(hsp.cusp[j])
				}
				acmc = state.sweDifdeg2n(hsp.ascendant, hsp.midheaven)
				if acmc < 0 {
					hsp.ascendant = state.sweDegnorm(hsp.ascendant + 180)
				}
				hsp.needInterpolation = hsp.needCuspSpeed
				break
			}
		case 'M':
			{
				var (
					j int32
					a float64 = th
					x [3]float64
				)
				for i = 1; i <= 12; i++ {
					j = i + 10
					if j > 12 {
						j -= 12
					}
					a = state.sweDegnorm(a + 30)
					x[0] = a
					x[1] = 0
					state.sweCotrans(x[:], x[:], ekl)
					hsp.cusp[j] = x[0]
				}
				acmc = state.sweDifdeg2n(hsp.ascendant, hsp.midheaven)
				if acmc < 0 {
					hsp.ascendant = state.sweDegnorm(hsp.ascendant + 180)
				}
				hsp.needInterpolation = hsp.needCuspSpeed
				break
			}
		case 'F':
			{
				var (
					a  float64
					ra float64
					x  [3]float64
				)
				acmc = state.sweDifdeg2n(hsp.ascendant, hsp.midheaven)
				if acmc < 0 {
					hsp.ascendant = state.sweDegnorm(hsp.ascendant + 180)
					hsp.cusp[1] = hsp.ascendant
				}
				x[0] = hsp.ascendant
				x[1] = 0
				state.sweCotrans(x[:], x[:], -ekl)
				a = x[0]
				for i = 2; i <= 12; i++ {
					if i <= 3 || i >= 10 {
						ra = state.sweDegnorm(a + float64((i-1)*30))
						if math.Abs(ra-90) > 1e-10 && math.Abs(ra-270) > 1e-10 {
							tant = math.Tan(ra * radiansPerDegree)
							hsp.cusp[i] = math.Atan(tant/cose) * degreesPerRadian
							if ra > 90 && ra <= 270 {
								hsp.cusp[i] = state.sweDegnorm(hsp.cusp[i] + 180)
							}
						} else if math.Abs(ra-90) <= 1e-10 {
							hsp.cusp[i] = 90
						} else {
							hsp.cusp[i] = 270
						}
						hsp.cusp[i] = state.sweDegnorm(hsp.cusp[i])
					}
				}
				hsp.needInterpolation = hsp.needCuspSpeed
				break
			}
		case 'B':
			{
				var (
					dek float64
					r   float64
					sna float64
					sda float64
					sn3 float64
					sd3 float64
				)
				acmc = state.sweDifdeg2n(hsp.ascendant, hsp.midheaven)
				if acmc < 0 {
					hsp.ascendant = state.sweDegnorm(hsp.ascendant + 180)
					hsp.cusp[1] = hsp.ascendant
					acmc = state.sweDifdeg2n(hsp.ascendant, hsp.midheaven)
				}
				dek = math.Asin(math.Sin(hsp.ascendant*radiansPerDegree)*sine) * degreesPerRadian
				r = -tanfi * math.Tan(dek*radiansPerDegree)
				if r > 1 {
					r = 1
				}
				if r < float64(-1) {
					r = float64(-1)
				}
				sda = math.Acos(r) * degreesPerRadian
				sna = 180 - sda
				sd3 = sda / 3
				sn3 = sna / 3
				rectasc = state.sweDegnorm(th + sd3)
				hsp.cusp[11] = state.Asc1(rectasc, 0, sine, cose)
				rectasc = state.sweDegnorm(th + 2*sd3)
				hsp.cusp[12] = state.Asc1(rectasc, 0, sine, cose)
				rectasc = state.sweDegnorm(th + 180 - 2*sn3)
				hsp.cusp[2] = state.Asc1(rectasc, 0, sine, cose)
				rectasc = state.sweDegnorm(th + 180 - sn3)
				hsp.cusp[3] = state.Asc1(rectasc, 0, sine, cose)
			}
			hsp.needInterpolation = hsp.needCuspSpeed
		case 'G':
			i = 1
			for ; i <= 36; i++ {
				hsp.cusp[i] = 0
				hsp.cuspSpeed[i] = 0
			}
			if math.Abs(fi) >= 90-ekl {
				status = -1
				copyText(hsp.message[:], textBytes("within polar circle, switched to Porphyry"))
				hsy = byte('O')
				selection1 = 'O'
				continue selectCase1
			}
			a = math.Asin(math.Tan(fi*radiansPerDegree)*tane) * degreesPerRadian
			ih = 2
			for ; ih <= 9; ih++ {
				ih2 = 10 - ih
				fh1 = math.Atan(math.Sin(a*float64(ih2)/9*radiansPerDegree)/tane) * degreesPerRadian
				rectasc = state.sweDegnorm(float64(90/9*ih2) + th)
				tant = math.Tan(math.Asin(sine*math.Sin(state.Asc1(rectasc, fh1, sine, cose)*radiansPerDegree)) * degreesPerRadian * radiansPerDegree)
				if math.Abs(tant) < 1e-10 {
					hsp.cusp[ih] = rectasc
					if hsp.needCuspSpeed != 0 {
						hsp.cuspSpeed[ih] = hsp.armcSpeed
					}
				} else {
					f = math.Atan(math.Sin(math.Asin(tanfi*tant)*degreesPerRadian*float64(ih2)/9*radiansPerDegree)/tant) * degreesPerRadian
					hsp.cusp[ih] = state.Asc1(rectasc, f, sine, cose)
					cuspsv = 0
					for i = 1; i <= niterMax; i++ {
						tant = math.Tan(math.Asin(sine*math.Sin(hsp.cusp[ih]*radiansPerDegree)) * degreesPerRadian * radiansPerDegree)
						if math.Abs(tant) < 1e-10 {
							hsp.cusp[ih] = rectasc
							if hsp.needCuspSpeed != 0 {
								hsp.cuspSpeed[ih] = hsp.armcSpeed
							}
							break
						}
						f = math.Atan(math.Sin(math.Asin(tanfi*tant)*degreesPerRadian*float64(ih2)/9*radiansPerDegree)/tant) * degreesPerRadian
						hsp.cusp[ih] = state.Asc1(rectasc, f, sine, cose)
						if i > 1 && math.Abs(state.sweDifdeg2n(hsp.cusp[ih], cuspsv)) < 2.7777777777777779e-06 {
							break
						}
						cuspsv = hsp.cusp[ih]
					}
					if i >= niterMax {
						status = -1
						hsy = byte('O')
						copyText(hsp.message[:], textBytes("very close to polar circle, switched to Porphyry"))
						selection1 = 'O'
						continue selectCase1
					}
					if hsp.needCuspSpeed != 0 {
						hsp.cuspSpeed[ih] = state.AscDash(rectasc, f, sine, cose)
					}
				}
				hsp.cusp[ih+18] = state.sweDegnorm(hsp.cusp[ih] + 180)
				if hsp.needCuspSpeed != 0 {
					hsp.cuspSpeed[ih+18] = hsp.cuspSpeed[ih]
				}
			}
			ih = 29
			for ; ih <= 36; ih++ {
				ih2 = ih - 28
				fh1 = math.Atan(math.Sin(a*float64(ih2)/9*radiansPerDegree)/tane) * degreesPerRadian
				rectasc = state.sweDegnorm(float64(180-ih2*90/9) + th)
				tant = math.Tan(math.Asin(sine*math.Sin(state.Asc1(rectasc, fh1, sine, cose)*radiansPerDegree)) * degreesPerRadian * radiansPerDegree)
				if math.Abs(tant) < 1e-10 {
					hsp.cusp[ih] = rectasc
					if hsp.needCuspSpeed != 0 {
						hsp.cuspSpeed[ih] = hsp.armcSpeed
					}
				} else {
					f = math.Atan(math.Sin(math.Asin(tanfi*tant)*degreesPerRadian*float64(ih2)/9*radiansPerDegree)/tant) * degreesPerRadian
					hsp.cusp[ih] = state.Asc1(rectasc, f, sine, cose)
					cuspsv = 0
					for i = 1; i <= niterMax; i++ {
						tant = math.Tan(math.Asin(sine*math.Sin(hsp.cusp[ih]*radiansPerDegree)) * degreesPerRadian * radiansPerDegree)
						if math.Abs(tant) < 1e-10 {
							hsp.cusp[ih] = rectasc
							if hsp.needCuspSpeed != 0 {
								hsp.cuspSpeed[ih] = hsp.armcSpeed
							}
							break
						}
						f = math.Atan(math.Sin(math.Asin(tanfi*tant)*degreesPerRadian*float64(ih2)/9*radiansPerDegree)/tant) * degreesPerRadian
						hsp.cusp[ih] = state.Asc1(rectasc, f, sine, cose)
						if i > 1 && math.Abs(state.sweDifdeg2n(hsp.cusp[ih], cuspsv)) < 2.7777777777777779e-06 {
							break
						}
						cuspsv = hsp.cusp[ih]
					}
					if i >= niterMax {
						status = -1
						hsy = byte('O')
						copyText(hsp.message[:], textBytes("very close to polar circle, switched to Porphyry"))
						selection1 = 'O'
						continue selectCase1
					}
					if hsp.needCuspSpeed != 0 {
						hsp.cuspSpeed[ih] = state.AscDash(rectasc, f, sine, cose)
					}
				}
				hsp.cusp[ih-18] = state.sweDegnorm(hsp.cusp[ih] + 180)
				if hsp.needCuspSpeed != 0 {
					hsp.cuspSpeed[ih-18] = hsp.cuspSpeed[ih]
				}
			}
			hsp.cusp[1] = hsp.ascendant
			hsp.cusp[10] = hsp.midheaven
			hsp.cusp[19] = state.sweDegnorm(hsp.ascendant + 180)
			hsp.cusp[28] = state.sweDegnorm(hsp.midheaven + 180)
			if hsp.needCuspSpeed != 0 {
				hsp.cuspSpeed[1] = hsp.ascendantSpeed
				hsp.cuspSpeed[10] = hsp.midheavenSpeed
				hsp.cuspSpeed[19] = hsp.ascendantSpeed
				hsp.cuspSpeed[28] = hsp.midheavenSpeed
			}
		case 'U':
			acmc = state.sweDifdeg2n(hsp.ascendant, hsp.midheaven)
			if acmc < 0 {
				hsp.ascendant = state.sweDegnorm(hsp.ascendant + 180)
			}
			x[0] = hsp.ascendant
			x[1] = 0.0
			x[2] = 1.0
			state.sweCotrans(x[:], x[:], -ekl)
			x[0] = x[0] - (th - 90)
			state.sweCotrans(x[:], x[:], -(90 - fi))
			krHorizonLon = x[0]
			x[0] = x[0] - x[0]
			state.sweCotrans(x[:], x[:], float64(-90))
			i = 0
			for ; i < 6; i++ {
				x[0] = 30.0 * float64(i)
				x[1] = 0.0
				state.sweCotrans(x[:], x[:], 90)
				x[0] = x[0] + krHorizonLon
				state.sweCotrans(x[:], x[:], 90-fi)
				x[0] = state.sweDegnorm(x[0] + (th - 90))
				hsp.cusp[i+1] = math.Atan(math.Tan(x[0]*radiansPerDegree)/math.Cos(ekl*radiansPerDegree)) * degreesPerRadian
				if x[0] > 90 && x[0] <= 270 {
					hsp.cusp[i+1] = state.sweDegnorm(hsp.cusp[i+1] + 180)
				}
				hsp.cusp[i+1] = state.sweDegnorm(hsp.cusp[i+1])
				hsp.cusp[i+7] = state.sweDegnorm(hsp.cusp[i+1] + 180)
			}
		case 'Y':
			i = 1
			for ; i <= 12; i++ {
				hsp.cusp[i] = state.apcSector(i, fi*radiansPerDegree, ekl*radiansPerDegree, th*radiansPerDegree)
			}
			hsp.cusp[10] = hsp.midheaven
			hsp.cusp[4] = state.sweDegnorm(hsp.midheaven + 180)
			if math.Abs(fi) >= 90-ekl {
				acmc = state.sweDifdeg2n(hsp.ascendant, hsp.midheaven)
				if acmc < 0 {
					hsp.ascendant = state.sweDegnorm(hsp.ascendant + 180)
					hsp.midheaven = state.sweDegnorm(hsp.midheaven + 180)
					for i = 1; i <= 12; i++ {
						hsp.cusp[i] = state.sweDegnorm(hsp.cusp[i] + 180)
					}
				}
			}
			hsp.needInterpolation = hsp.needCuspSpeed
		default:
			if math.Abs(fi) >= 90-ekl {
				status = -1
				copyText(hsp.message[:], textBytes("within polar circle, switched to Porphyry"))
				selection1 = 'O'
				continue selectCase1
			}
			a = math.Asin(math.Tan(fi*radiansPerDegree)*tane) * degreesPerRadian
			fh1 = math.Atan(math.Sin(a/3*radiansPerDegree)/tane) * degreesPerRadian
			fh2 = math.Atan(math.Sin(a*2/3*radiansPerDegree)/tane) * degreesPerRadian
			rectasc = state.sweDegnorm(30 + th)
			tant = math.Tan(math.Asin(sine*math.Sin(state.Asc1(rectasc, fh1, sine, cose)*radiansPerDegree)) * degreesPerRadian * radiansPerDegree)
			ih = 11
			if math.Abs(tant) < 1e-10 {
				hsp.cusp[ih] = rectasc
				if hsp.needCuspSpeed != 0 {
					hsp.cuspSpeed[ih] = hsp.armcSpeed
				}
			} else {
				f = math.Atan(math.Sin(math.Asin(tanfi*tant)*degreesPerRadian/3*radiansPerDegree)/tant) * degreesPerRadian
				hsp.cusp[ih] = state.Asc1(rectasc, f, sine, cose)
				cuspsv = 0
				for i = 1; i <= niterMax; i++ {
					tant = math.Tan(math.Asin(sine*math.Sin(hsp.cusp[ih]*radiansPerDegree)) * degreesPerRadian * radiansPerDegree)
					if math.Abs(tant) < 1e-10 {
						hsp.cusp[ih] = rectasc
						if hsp.needCuspSpeed != 0 {
							hsp.cuspSpeed[ih] = hsp.armcSpeed
						}
						break
					}
					f = math.Atan(math.Sin(math.Asin(tanfi*tant)*degreesPerRadian/3*radiansPerDegree)/tant) * degreesPerRadian
					hsp.cusp[ih] = state.Asc1(rectasc, f, sine, cose)
					if i > 1 && math.Abs(state.sweDifdeg2n(hsp.cusp[ih], cuspsv)) < 2.7777777777777779e-06 {
						break
					}
					cuspsv = hsp.cusp[ih]
				}
				if i >= niterMax {
					status = -1
					copyText(hsp.message[:], textBytes("very close to polar circle, switched to Porphyry"))
					selection1 = 'O'
					continue selectCase1
				}
				if hsp.needCuspSpeed != 0 {
					hsp.cuspSpeed[ih] = state.AscDash(rectasc, f, sine, cose)
				}
			}
			rectasc = state.sweDegnorm(60 + th)
			tant = math.Tan(math.Asin(sine*math.Sin(state.Asc1(rectasc, fh2, sine, cose)*radiansPerDegree)) * degreesPerRadian * radiansPerDegree)
			ih = 12
			if math.Abs(tant) < 1e-10 {
				hsp.cusp[ih] = rectasc
				if hsp.needCuspSpeed != 0 {
					hsp.cuspSpeed[ih] = hsp.armcSpeed
				}
			} else {
				f = math.Atan(math.Sin(math.Asin(tanfi*tant)*degreesPerRadian/1.5*radiansPerDegree)/tant) * degreesPerRadian
				hsp.cusp[ih] = state.Asc1(rectasc, f, sine, cose)
				cuspsv = 0
				for i = 1; i <= niterMax; i++ {
					tant = math.Tan(math.Asin(sine*math.Sin(hsp.cusp[ih]*radiansPerDegree)) * degreesPerRadian * radiansPerDegree)
					if math.Abs(tant) < 1e-10 {
						hsp.cusp[ih] = rectasc
						if hsp.needCuspSpeed != 0 {
							hsp.cuspSpeed[ih] = hsp.armcSpeed
						}
						break
					}
					f = math.Atan(math.Sin(math.Asin(tanfi*tant)*degreesPerRadian/1.5*radiansPerDegree)/tant) * degreesPerRadian
					hsp.cusp[ih] = state.Asc1(rectasc, f, sine, cose)
					if i > 1 && math.Abs(state.sweDifdeg2n(hsp.cusp[ih], cuspsv)) < 2.7777777777777779e-06 {
						break
					}
					cuspsv = hsp.cusp[ih]
				}
				if i >= niterMax {
					status = -1
					copyText(hsp.message[:], textBytes("very close to polar circle, switched to Porphyry"))
					selection1 = 'O'
					continue selectCase1
				}
				if hsp.needCuspSpeed != 0 {
					hsp.cuspSpeed[ih] = state.AscDash(rectasc, f, sine, cose)
				}
			}
			rectasc = state.sweDegnorm(120 + th)
			tant = math.Tan(math.Asin(sine*math.Sin(state.Asc1(rectasc, fh2, sine, cose)*radiansPerDegree)) * degreesPerRadian * radiansPerDegree)
			ih = 2
			if math.Abs(tant) < 1e-10 {
				hsp.cusp[ih] = rectasc
				if hsp.needCuspSpeed != 0 {
					hsp.cuspSpeed[ih] = hsp.armcSpeed
				}
			} else {
				f = math.Atan(math.Sin(math.Asin(tanfi*tant)*degreesPerRadian/1.5*radiansPerDegree)/tant) * degreesPerRadian
				hsp.cusp[ih] = state.Asc1(rectasc, f, sine, cose)
				cuspsv = 0
				for i = 1; i <= niterMax; i++ {
					tant = math.Tan(math.Asin(sine*math.Sin(hsp.cusp[ih]*radiansPerDegree)) * degreesPerRadian * radiansPerDegree)
					if math.Abs(tant) < 1e-10 {
						hsp.cusp[ih] = rectasc
						if hsp.needCuspSpeed != 0 {
							hsp.cuspSpeed[ih] = hsp.armcSpeed
						}
						break
					}
					f = math.Atan(math.Sin(math.Asin(tanfi*tant)*degreesPerRadian/1.5*radiansPerDegree)/tant) * degreesPerRadian
					hsp.cusp[ih] = state.Asc1(rectasc, f, sine, cose)
					if i > 1 && math.Abs(state.sweDifdeg2n(hsp.cusp[ih], cuspsv)) < 2.7777777777777779e-06 {
						break
					}
					cuspsv = hsp.cusp[ih]
				}
				if i >= niterMax {
					status = -1
					copyText(hsp.message[:], textBytes("very close to polar circle, switched to Porphyry"))
					selection1 = 'O'
					continue selectCase1
				}
				if hsp.needCuspSpeed != 0 {
					hsp.cuspSpeed[ih] = state.AscDash(rectasc, f, sine, cose)
				}
			}
			rectasc = state.sweDegnorm(150 + th)
			tant = math.Tan(math.Asin(sine*math.Sin(state.Asc1(rectasc, fh1, sine, cose)*radiansPerDegree)) * degreesPerRadian * radiansPerDegree)
			ih = 3
			if math.Abs(tant) < 1e-10 {
				hsp.cusp[ih] = rectasc
				if hsp.needCuspSpeed != 0 {
					hsp.cuspSpeed[ih] = hsp.armcSpeed
				}
			} else {
				f = math.Atan(math.Sin(math.Asin(tanfi*tant)*degreesPerRadian/3*radiansPerDegree)/tant) * degreesPerRadian
				hsp.cusp[ih] = state.Asc1(rectasc, f, sine, cose)
				cuspsv = 0
				for i = 1; i <= niterMax; i++ {
					tant = math.Tan(math.Asin(sine*math.Sin(hsp.cusp[ih]*radiansPerDegree)) * degreesPerRadian * radiansPerDegree)
					if math.Abs(tant) < 1e-10 {
						hsp.cusp[ih] = rectasc
						if hsp.needCuspSpeed != 0 {
							hsp.cuspSpeed[ih] = hsp.armcSpeed
						}
						break
					}
					f = math.Atan(math.Sin(math.Asin(tanfi*tant)*degreesPerRadian/3*radiansPerDegree)/tant) * degreesPerRadian
					hsp.cusp[ih] = state.Asc1(rectasc, f, sine, cose)
					if i > 1 && math.Abs(state.sweDifdeg2n(hsp.cusp[ih], cuspsv)) < 2.7777777777777779e-06 {
						break
					}
					cuspsv = hsp.cusp[ih]
				}
				if i >= niterMax {
					status = -1
					copyText(hsp.message[:], textBytes("very close to polar circle, switched to Porphyry"))
					selection1 = 'O'
					continue selectCase1
				}
				if hsp.needCuspSpeed != 0 {
					hsp.cuspSpeed[ih] = state.AscDash(rectasc, f, sine, cose)
				}
			}
		}
		break
	}
	if int32(hsy) != 'G' && int32(hsy) != 'Y' && upperByte(hsy) != 'I' {
		hsp.cusp[4] = state.sweDegnorm(hsp.cusp[10] + 180)
		hsp.cusp[5] = state.sweDegnorm(hsp.cusp[11] + 180)
		hsp.cusp[6] = state.sweDegnorm(hsp.cusp[12] + 180)
		hsp.cusp[7] = state.sweDegnorm(hsp.cusp[1] + 180)
		hsp.cusp[8] = state.sweDegnorm(hsp.cusp[2] + 180)
		hsp.cusp[9] = state.sweDegnorm(hsp.cusp[3] + 180)
		if hsp.needCuspSpeed != 0 && !(hsp.needInterpolation != 0) {
			hsp.cuspSpeed[4] = hsp.cuspSpeed[10]
			hsp.cuspSpeed[5] = hsp.cuspSpeed[11]
			hsp.cuspSpeed[6] = hsp.cuspSpeed[12]
			hsp.cuspSpeed[7] = hsp.cuspSpeed[1]
			hsp.cuspSpeed[8] = hsp.cuspSpeed[2]
			hsp.cuspSpeed[9] = hsp.cuspSpeed[3]
		}
	}
	if fi >= 0 {
		f = 90 - fi
	} else {
		f = float64(-90) - fi
	}
	hsp.vertex = state.Asc1(th-90, f, sine, cose)
	if hsp.needSpeed != 0 {
		hsp.vertexSpeed = state.AscDash(th-90, f, sine, cose)
	}
	if math.Abs(fi) <= ekl {
		vemc = state.sweDifdeg2n(hsp.vertex, hsp.midheaven)
		if vemc > 0 {
			hsp.vertex = state.sweDegnorm(hsp.vertex + 180)
		}
	}
	th2 = state.sweDegnorm(th + 90)
	if math.Abs(th2-90) > 1e-10 && math.Abs(th2-270) > 1e-10 {
		tant = math.Tan(th2 * radiansPerDegree)
		hsp.equatorialAscendant = math.Atan(tant/cose) * degreesPerRadian
		if th2 > 90 && th2 <= 270 {
			hsp.equatorialAscendant = state.sweDegnorm(hsp.equatorialAscendant + 180)
		}
	} else if math.Abs(th2-90) <= 1e-10 {
		hsp.equatorialAscendant = 90
	} else {
		hsp.equatorialAscendant = 270
	}
	hsp.equatorialAscendant = state.sweDegnorm(hsp.equatorialAscendant)
	if hsp.needSpeed != 0 {
		hsp.equatorialAscendantSpeed = state.AscDash(th+90, 0, sine, cose)
	}
	hsp.coAscendantKoch = state.sweDegnorm(state.Asc1(th-90, fi, sine, cose) + 180)
	if hsp.needSpeed != 0 {
		hsp.coAscendantKochSpeed = state.AscDash(th-90, fi, sine, cose)
	}
	if fi >= 0 {
		hsp.coAscendantMunkasey = state.Asc1(th+90, 90-fi, sine, cose)
		if hsp.needSpeed != 0 {
			hsp.coAscendantMunkaseySpeed = state.AscDash(th+90, 90-fi, sine, cose)
		}
	} else {
		hsp.coAscendantMunkasey = state.Asc1(th+90, float64(-90)-fi, sine, cose)
		if hsp.needSpeed != 0 {
			hsp.coAscendantMunkaseySpeed = state.AscDash(th+90, float64(-90)-fi, sine, cose)
		}
	}
	hsp.polarAscendant = state.Asc1(th-90, fi, sine, cose)
	if hsp.needSpeed != 0 {
		hsp.polarAscendantSpeed = state.AscDash(th-90, fi, sine, cose)
	}
	return status
}

// Asc1 implements Asc1 from swehouse.c.
//
// ***
// oblique triangle formed by: great circle with pole height f, ecliptic and equator,
// x = intersection equator - great circle.
// return crossing of ecliptic with great circle.
// Prepare quadrants before doing the work in Asc2.
func (state *State) Asc1(x1 float64, f float64, sine float64, cose float64) float64 {
	var (
		n   int32
		ass float64
	)
	x1 = state.sweDegnorm(x1)
	n = int32(x1/90 + 1)
	if math.Abs(90-f) < 1e-10 {
		return 180
	}
	if math.Abs(90+f) < 1e-10 {
		return 0
	}
	if n == 1 {
		ass = state.Asc2(x1, f, sine, cose)
	} else if n == 2 {
		ass = 180 - state.Asc2(180-x1, -f, sine, cose)
	} else if n == 3 {
		ass = 180 + state.Asc2(x1-180, -f, sine, cose)
	} else {
		ass = 360 - state.Asc2(360-x1, f, sine, cose)
	}
	ass = state.sweDegnorm(ass)
	if math.Abs(ass-90) < 1e-10 {
		ass = 90
	}
	if math.Abs(ass-180) < 1e-10 {
		ass = 180
	}
	if math.Abs(ass-270) < 1e-10 {
		ass = 270
	}
	if math.Abs(ass-360) < 1e-10 {
		ass = 0
	}
	return ass
}

// Asc2 implements Asc2 from swehouse.c.
//
// x in range 0..90
// f in range -90 .. +90
// sine, cose around e=23°
// oblique triangle formed by: great circle with pole height f, ecliptic and equator,
// x = intersection equator - great circle.
// return crossing of ecliptic with great circle.
func (state *State) Asc2(x float64, f float64, sine float64, cose float64) float64 {
	var (
		ass  float64
		sinx float64
	)
	ass = -math.Tan(f*radiansPerDegree)*sine + cose*math.Cos(x*radiansPerDegree)
	if math.Abs(ass) < 1e-10 {
		ass = 0
	}
	sinx = math.Sin(x * radiansPerDegree)
	if math.Abs(sinx) < 1e-10 {
		sinx = 0
	}
	if sinx == 0 {
		if ass < 0 {
			ass = -1e-10
		} else {
			ass = 1e-10
		}
	} else if ass == 0 {
		if sinx < 0 {
			ass = float64(-90)
		} else {
			ass = 90
		}
	} else {
		ass = math.Atan(sinx/ass) * degreesPerRadian
	}
	if ass < 0 {
		ass = 180 + ass
	}
	return ass
}

// AscDash implements AscDash from swehouse.c.
//
//	Asc2
func (state *State) AscDash(x float64, f float64, sine float64, cose float64) float64 {
	var (
		cosx  float64 = math.Cos(x * radiansPerDegree)
		sinx  float64 = math.Sin(x * radiansPerDegree)
		sinx2 float64 = sinx * sinx
		c     float64 = cose*cosx - math.Tan(f*radiansPerDegree)*sine
		d     float64 = sinx2 + c*c
		dudt  float64
	)
	if d > 1e-10 {
		dudt = (cosx*c + cose*sinx2) / d
	} else {
		dudt = 0.0
	}
	return dudt * 360.98564733498665
}

// armcToMc implements armc_to_mc from swehouse.c.
func (state *State) armcToMc(armc float64, eps float64) float64 {
	var (
		cose float64 = math.Cos(eps * radiansPerDegree)
		mc   float64
		tant float64
	)
	if math.Abs(armc-90) > 1e-10 && math.Abs(armc-270) > 1e-10 {
		tant = math.Tan(armc * radiansPerDegree)
		mc = state.sweDegnorm(math.Atan(tant/cose) * degreesPerRadian)
		if armc > 90 && armc <= 270 {
			mc = state.sweDegnorm(mc + 180)
		}
	} else if math.Abs(armc-90) <= 1e-10 {
		mc = 90
	} else {
		mc = 270
	}
	return mc
}

// fixAscPolar implements fix_asc_polar from swehouse.c.
//
//	if ascendant is on western half of horizon, add 180°
func (state *State) fixAscPolar(asc float64, armc float64, eps float64, latitude float64) float64 {
	var demc float64 = math.Atan(math.Sin(armc*radiansPerDegree)*math.Tan(eps*radiansPerDegree)) * degreesPerRadian
	if latitude >= 0 && 90-latitude+demc < 0 {
		asc = state.sweDegnorm(asc + 180)
	}
	if latitude < 0 && float64(-90)-latitude+demc > 0 {
		asc = state.sweDegnorm(asc + 180)
	}
	return asc
}

func (state *State) sweHousePos(armc float64, latitude float64, eps float64, houseSystem int32, xpin []float64, message []byte) float64 {
	var (
		xp            [6]float64
		xeq           [6]float64
		ra            float64
		de            float64
		mdd           float64
		mdn           float64
		sad           float64
		san           float64
		hpos          float64
		sinad         float64
		ad            float64
		a             float64
		admc          float64
		adp           float64
		samc          float64
		asc           float64
		mc            float64
		acmc          float64
		tant          float64
		fh            float64
		ra0           float64
		tanfi         float64
		sinfi         float64
		fac           float64
		dfac          float64
		tanx          float64
		x             [3]float64
		xasc          [3]float64
		xs1           float64
		xs2           float64
		raep          float64
		raaz          float64
		oblaz         float64
		xtemp         float64
		hcusp         [37]float64
		ascmc         [10]float64
		sine          float64 = math.Sin(eps * radiansPerDegree)
		cose          float64 = math.Cos(eps * radiansPerDegree)
		c1            float64
		c2            float64
		d             float64
		hsize         float64
		i             int32
		j             int32
		nloop         int32
		dsun          float64 = 0
		darmc         float64
		harmc         float64
		y             float64
		sinpsi        float64
		sa            float64
		isWesternHalf int32 = 0
	)
	houseSystem = upperByte(houseSystem)
	if 1 != 0 {
		ascmc[9] = 99
		if state.sweHousesArmcEx2(armc, latitude, eps, houseSystem, hcusp[:], ascmc[:], nil, nil, message) == -1 {
			if message != nil {
				formatText(message, textBytes("swe_house_pos(): failed for system %c"), houseSystem)
			}
		} else {
			hpos = 0
			for i = 1; i <= 12; i++ {
				if math.Abs(state.sweDifdeg2n(xpin[0], hcusp[i])) < 2.7777777777777776e-07 && xpin[1] == 0 {
					hpos = float64(i)
				}
			}
			for i = 1; i <= 12; i += 3 {
				if math.Abs(state.sweDifdeg2n(xpin[0], hcusp[i])) < 2.7777777777777776e-07 && xpin[1] == 0 {
					xp[0] = float64(i)
				}
			}
			if hpos > 0 {
				return hpos
			}
			if houseSystem == 'I' {
				dsun = ascmc[9]
			}
			if houseSystem == 'Y' {
				xeq[0] = ascmc[0]
				xeq[1] = 0
				xeq[2] = 1
				state.sweCotrans(xeq[:], xeq[:], -eps)
				dsun = xeq[1]
			}
		}
	}
	var (
		isAboveHor    int32 = 0
		isInvalid     int32 = 0
		isCircumpolar int32 = 0
	)
	if message != nil {
		message[0] = byte('\000')
	}
	xeq[0] = xpin[0]
	xeq[1] = xpin[1]
	xeq[2] = 1
	state.sweCotrans(xeq[:], xeq[:], -eps)
	ra = xeq[0]
	de = xeq[1]
	mdd = state.sweDegnorm(ra - armc)
	mdn = state.sweDegnorm(mdd + 180)
	if mdd >= 180 {
		mdd -= 360
	}
	if mdn >= 180 {
		mdn -= 360
	}
	switch houseSystem {
	case 'N':
		xp[0] = xpin[0]
		hpos = xp[0]/30.0 + 1
	case 'A':
		fallthrough
	case 'E':
		fallthrough
	case 'D':
		fallthrough
	case 'V':
		fallthrough
	case 'W':
		asc = state.Asc1(state.sweDegnorm(armc+90), latitude, sine, cose)
		mc = state.armcToMc(armc, eps)
		asc = state.fixAscPolar(asc, armc, eps, latitude)
		xp[0] = state.sweDegnorm(xpin[0] - asc)
		if houseSystem == 'V' {
			xp[0] = state.sweDegnorm(xp[0] + 15)
		}
		if houseSystem == 'W' {
			xp[0] = state.sweDegnorm(xp[0] + math.Mod(asc, 30))
		}
		if houseSystem == 'D' {
			xp[0] = state.sweDegnorm(xpin[0] - mc - 90)
		}
		xp[0] = state.sweDegnorm(xp[0] + 2.7777777777777776e-07)
		hpos = xp[0]/30.0 + 1
	case 'O':
		fallthrough
	case 'B':
		fallthrough
	case 'S':
		asc = state.Asc1(state.sweDegnorm(armc+90), latitude, sine, cose)
		mc = state.armcToMc(armc, eps)
		asc = state.fixAscPolar(asc, armc, eps, latitude)
		if houseSystem == 'O' || houseSystem == 'S' {
			xp[0] = state.sweDegnorm(xpin[0] - asc)
			xp[0] = state.sweDegnorm(xp[0] + 2.7777777777777776e-07)
			if xp[0] < 180 {
				hpos = 1
			} else {
				hpos = 7
				xp[0] -= 180
			}
			acmc = state.sweDifdeg2n(asc, mc)
			if xp[0] < 180-acmc {
				hpos += xp[0] * 3 / (180 - acmc)
			} else {
				hpos += 3 + (xp[0]-180+acmc)*3/acmc
			}
			if houseSystem == 'S' {
				hpos += 0.5
				if hpos > 12 {
					hpos = 1
				}
			}
		} else {
			var (
				dek float64
				r   float64
				sna float64
				sda float64
			)
			dek = math.Asin(math.Sin(asc*radiansPerDegree)*sine) * degreesPerRadian
			tanfi = math.Tan(latitude * radiansPerDegree)
			r = -tanfi * math.Tan(dek*radiansPerDegree)
			sda = math.Acos(r) * degreesPerRadian
			sna = 180 - sda
			if mdd > 0 {
				if mdd < sda {
					hpos = mdd * 90 / sda
				} else {
					hpos = 90 + (mdd-sda)*90/sna
				}
			} else if mdd > -sna {
				hpos = 360 + mdd*90/sna
			} else {
				hpos = 270 + (mdd+sna)*90/sda
			}
			hpos = state.sweDegnorm(hpos-90)/30.0 + 1.0
			if hpos >= 13.0 {
				hpos -= 12
			}
		}
	case 'X':
		hpos = state.sweDegnorm(mdd-90)/30.0 + 1.0
	case 'F':
		x[0] = state.Asc1(state.sweDegnorm(armc+90), latitude, sine, cose)
		x[0] = state.fixAscPolar(x[0], armc, eps, latitude)
		x[1] = 0
		state.sweCotrans(x[:], x[:], -eps)
		hpos = state.sweDegnorm(ra-x[0])/30.0 + 1
	case 'M':
		{
			var a float64 = xpin[0]
			if math.Abs(a-90) > 1e-10 && math.Abs(a-270) > 1e-10 {
				tant = math.Tan(a * radiansPerDegree)
				hpos = math.Atan(tant/cose) * degreesPerRadian
				if a > 90 && a <= 270 {
					hpos = state.sweDegnorm(hpos + 180)
				}
			} else if math.Abs(a-90) <= 1e-10 {
				hpos = 90
			} else {
				hpos = 270
			}
			hpos = state.sweDegnorm(hpos - armc - 90)
			hpos = hpos/30.0 + 1
		}
	case 'K':
		isInvalid = 0
		isCircumpolar = 0
		if 90-latitude < de || float64(-90)-latitude > de {
			adp = 90
			isCircumpolar = 1
		} else if latitude-90 > de || latitude+90 < de {
			adp = float64(-90)
			isCircumpolar = 1
		} else {
			adp = math.Asin(math.Tan(latitude*radiansPerDegree)*math.Tan(de*radiansPerDegree)) * degreesPerRadian
		}
		admc = math.Tan(eps*radiansPerDegree) * math.Tan(latitude*radiansPerDegree) * math.Sin(armc*radiansPerDegree)
		if math.Abs(admc) > 1 {
			if admc > 1 {
				admc = 1
			} else {
				admc = float64(-1)
			}
			isCircumpolar = 1
		}
		admc = math.Asin(admc) * degreesPerRadian
		samc = 90 + admc
		if samc == 0 {
			isInvalid = 1
		}
		if math.Abs(samc) > 0 {
			if mdd >= 0 {
				dfac = (mdd - adp + admc) / samc
				xp[0] = state.sweDegnorm((dfac - 1) * 90)
				xp[0] = state.sweDegnorm(xp[0] + 2.7777777777777776e-07)
				if dfac > 2 || dfac < 0 {
					isInvalid = 1
				}
			} else {
				dfac = (mdd + 180 + adp + admc) / samc
				xp[0] = state.sweDegnorm((dfac + 1) * 90)
				xp[0] = state.sweDegnorm(xp[0] + 2.7777777777777776e-07)
				if dfac > 2 || dfac < 0 {
					isInvalid = 1
				}
			}
		}
		if isInvalid != 0 {
			xp[0] = 0
			hpos = 0
			if message != nil {
				copyText(message, textBytes("Koch house position failed in circumpolar area"))
			}
			break
		}
		if isCircumpolar != 0 {
			if message != nil {
				copyText(message, textBytes("Koch house position, doubtful result in circumpolar area"))
			}
		}
		hpos = xp[0]/30.0 + 1
	case 'C':
		xeq[0] = state.sweDegnorm(mdd - 90)
		state.sweCotrans(xeq[:], xp[:], -latitude)
		xp[0] = state.sweDegnorm(xp[0] + 2.7777777777777776e-07)
		hpos = xp[0]/30.0 + 1
	case 'J':
		sinfi = math.Sin(latitude * radiansPerDegree)
		if math.Abs(latitude) < 1e-10 {
			xs2 = 0.33333333333333331
			xs1 = 0.66666666666666663
		} else {
			xs2 = math.Sin(latitude/3*radiansPerDegree) / sinfi
			xs1 = math.Sin(2*latitude/3*radiansPerDegree) / sinfi
		}
		xs2 = math.Asin(xs2) * degreesPerRadian
		xs1 = math.Asin(xs1) * degreesPerRadian
		hcusp[1] = 0
		hcusp[2] = xs2
		hcusp[3] = xs1
		hcusp[4] = 90
		hcusp[5] = 180 - xs1
		hcusp[6] = 180 - xs2
		hcusp[7] = 180
		hcusp[8] = 180 + xs2
		hcusp[9] = 180 + xs1
		hcusp[10] = 270
		hcusp[11] = 360 - xs1
		hcusp[12] = 360 - xs2
		xeq[0] = state.sweDegnorm(mdd - 90)
		state.sweCotrans(xeq[:], xp[:], -latitude)
		a = xp[0]
		if state.sweDifdeg2n(hcusp[6], hcusp[1]) > 0 {
			d = state.sweDegnorm(a - hcusp[1])
			for i = 1; i <= 12; i++ {
				j = i + 1
				if j > 12 {
					c2 = 360
				} else {
					c2 = state.sweDegnorm(hcusp[j] - hcusp[1])
				}
				if d < c2 {
					break
				}
			}
			c1 = state.sweDegnorm(hcusp[i] - hcusp[1])
		} else {
			d = state.sweDegnorm(hcusp[1] - a)
			for i = 1; i <= 12; i++ {
				j = i + 1
				if j > 12 {
					c2 = 360
				} else {
					c2 = state.sweDegnorm(hcusp[1] - hcusp[j])
				}
				if d < c2 {
					break
				}
			}
			c1 = state.sweDegnorm(hcusp[1] - hcusp[i])
		}
		hsize = c2 - c1
		if hsize == 0 {
			hpos = float64(i)
		} else {
			hpos = float64(i) + (d-c1)/hsize
		}
	case 'U':
		if math.Abs(latitude) < 1e-10 {
			if latitude >= 0 {
				latitude = 1e-10
			} else {
				latitude = -1e-10
			}
		}
		asc = state.Asc1(state.sweDegnorm(armc+90), latitude, sine, cose)
		asc = state.fixAscPolar(asc, armc, eps, latitude)
		x[0] = asc
		x[1] = 0.0
		x[2] = 1.0
		state.sweCotrans(x[:], x[:], -eps)
		raep = state.sweDegnorm(armc + 90)
		x[0] = state.sweDegnorm(raep - x[0])
		state.sweCotrans(x[:], x[:], -(90 - latitude))
		tanx = math.Tan(x[0] * radiansPerDegree)
		if latitude == 0 {
			xtemp = 90
			if tanx < 0 {
				xtemp = -90
			}
		} else {
			xtemp = math.Atan(tanx/math.Cos((90-latitude)*radiansPerDegree)) * degreesPerRadian
		}
		if x[0] > 90 && x[0] <= 270 {
			xtemp = state.sweDegnorm(xtemp + 180)
		}
		x[0] = state.sweDegnorm(xtemp)
		raaz = state.sweDegnorm(raep - x[0])
		x[0] = raaz
		x[1] = 0.0
		x[0] = state.sweDegnorm(raep - x[0])
		state.sweCotrans(x[:], x[:], -(90 - latitude))
		x[1] = x[1] + 90
		state.sweCotrans(x[:], x[:], 90-latitude)
		oblaz = x[1]
		xasc[0] = asc
		xasc[1] = 0.0
		xasc[2] = 1.0
		state.sweCotrans(xasc[:], xasc[:], -eps)
		xasc[0] = state.sweDegnorm(xasc[0] - raaz)
		xtemp = math.Atan(math.Tan(xasc[0]*radiansPerDegree)/math.Cos(oblaz*radiansPerDegree)) * degreesPerRadian
		if xasc[0] > 90 && xasc[0] <= 270 {
			xtemp = state.sweDegnorm(xtemp + 180)
		}
		xasc[0] = state.sweDegnorm(xtemp)
		xp[0] = state.sweDegnorm(xeq[0] - raaz)
		xtemp = math.Atan(math.Tan(xp[0]*radiansPerDegree)/math.Cos(oblaz*radiansPerDegree)) * degreesPerRadian
		if xp[0] > 90 && xp[0] <= 270 {
			xtemp = state.sweDegnorm(xtemp + 180)
		}
		xp[0] = state.sweDegnorm(xtemp)
		xp[0] = state.sweDegnorm(xp[0] - xasc[0])
		x[0] = xeq[0]
		x[1] = xeq[1]
		state.sweCotrans(x[:], x[:], oblaz)
		xp[1] = xeq[1] - x[1]
		xp[0] = state.sweDegnorm(xp[0] + 2.7777777777777776e-07)
		hpos = xp[0]/30.0 + 1
	case 'H':
		xeq[0] = state.sweDegnorm(mdd - 90)
		state.sweCotrans(xeq[:], xp[:], 90-latitude)
		xp[0] = state.sweDegnorm(xp[0] + 2.7777777777777776e-07)
		hpos = xp[0]/30.0 + 1
	case 'R':
		if math.Abs(mdd) < 1e-10 {
			xp[0] = 270
		} else if 180-math.Abs(mdd) < 1e-10 {
			xp[0] = 90
		} else {
			if 90-math.Abs(latitude) < 1e-10 {
				if latitude > 0 {
					latitude = 89.999999999899998
				} else {
					latitude = -89.999999999899998
				}
			}
			if 90-math.Abs(de) < 1e-10 {
				if de > 0 {
					de = 89.999999999899998
				} else {
					de = -89.999999999899998
				}
			}
			a = math.Tan(latitude*radiansPerDegree)*math.Tan(de*radiansPerDegree) + math.Cos(mdd*radiansPerDegree)
			xp[0] = state.sweDegnorm(math.Atan(-a/math.Sin(mdd*radiansPerDegree)) * degreesPerRadian)
			if mdd < 0 {
				xp[0] += 180
			}
			xp[0] = state.sweDegnorm(xp[0])
			xp[0] = state.sweDegnorm(xp[0] + 2.7777777777777776e-07)
		}
		hpos = xp[0]/30.0 + 1
	case 'I':
		fallthrough
	case 'i':
		fallthrough
	case 'Y':
		if latitude > 89.999999722222228 {
			latitude = 89.999999722222228
		}
		if latitude < -89.999999722222228 {
			latitude = -89.999999722222228
		}
		if 90-math.Abs(de) < 1e-10 {
			if de > 0 {
				de = 89.999999999899998
			} else {
				de = -89.999999999899998
			}
		}
		a = math.Tan(latitude*radiansPerDegree)*math.Tan(de*radiansPerDegree) + math.Cos(mdd*radiansPerDegree)
		xp[0] = state.sweDegnorm(math.Atan(-a/math.Sin(mdd*radiansPerDegree)) * degreesPerRadian)
		if mdd < 0 {
			xp[0] += 180
		}
		xp[0] = state.sweDegnorm(xp[0])
		sinad = math.Tan(de*radiansPerDegree) * math.Tan(latitude*radiansPerDegree)
		a = sinad + math.Cos(mdd*radiansPerDegree)
		if a >= 0 {
			isAboveHor = 1
		}
		harmc = 90 - latitude
		if latitude < 0 {
			harmc = 90 + latitude
		}
		darmc = state.sweDegnorm(xp[0] - 270)
		if darmc > 180 {
			isWesternHalf = 1
			darmc = 360 - darmc
		}
		sinad = math.Tan(dsun*radiansPerDegree) * math.Tan(latitude*radiansPerDegree)
		if sinad >= 1 {
			ad = 90
		} else if sinad <= float64(-1) {
			ad = float64(-90)
		} else {
			ad = math.Asin(sinad) * degreesPerRadian
		}
		sad = 90 + ad
		san = 90 - ad
		if sad == 0 && isAboveHor != 0 {
			xp[0] = 270
		} else if san == 0 && !(isAboveHor != 0) {
			xp[0] = 90
		} else {
			sa = sad
			if !(isAboveHor != 0) {
				dsun = -dsun
				sa = san
				darmc = 180 - darmc
				isWesternHalf = int32(boolInt(!(isWesternHalf != 0)))
			}
			a = math.Acos(math.Cos(harmc*radiansPerDegree)*math.Cos(darmc*radiansPerDegree)) * degreesPerRadian
			if a < 1e-10 {
				a = 1e-10
			}
			sinpsi = math.Sin(harmc*radiansPerDegree) / math.Sin(a*radiansPerDegree)
			if sinpsi > 1 {
				sinpsi = 1
			}
			if sinpsi < float64(-1) {
				sinpsi = float64(-1)
			}
			y = math.Sin(dsun*radiansPerDegree) / sinpsi
			if y > 1 {
				y = 89.999999999899998
			} else if y < float64(-1) {
				y = -89.999999999899998
			} else {
				y = math.Asin(y) * degreesPerRadian
			}
			d = math.Acos(math.Cos(y*radiansPerDegree)/math.Cos(dsun*radiansPerDegree)) * degreesPerRadian
			if dsun < 0 {
				d = -d
			}
			if latitude < 0 {
				d = -d
			}
			darmc += d
			if isWesternHalf != 0 {
				xp[0] = 270 - darmc/sa*90
			} else {
				xp[0] = 270 + darmc/sa*90
			}
			if !(isAboveHor != 0) {
				xp[0] = state.sweDegnorm(xp[0] + 180)
			}
		}
		xp[0] = state.sweDegnorm(xp[0] + 2.7777777777777776e-07)
		hpos = xp[0]/30.0 + 1
	case 'T':
		fh = latitude
		if fh > 89.999 {
			fh = 89.999
		}
		if fh < -89.999 {
			fh = -89.999
		}
		mdd = state.sweDegnorm(mdd)
		if de > 89.999999999899998 {
			de = 89.999999999899998
		}
		if de < -89.999999999899998 {
			de = -89.999999999899998
		}
		sinad = math.Tan(de*radiansPerDegree) * math.Tan(fh*radiansPerDegree)
		if sinad > 1.0 {
			sinad = 1.0
		}
		if sinad < -1.0 {
			sinad = -1.0
		}
		a = sinad + math.Cos(mdd*radiansPerDegree)
		if a >= 0 {
			isAboveHor = 1
		}
		if !(isAboveHor != 0) {
			ra = state.sweDegnorm(ra + 180)
			de = -de
			mdd = state.sweDegnorm(mdd + 180)
		}
		if mdd > 180 {
			ra = state.sweDegnorm(armc - mdd)
		}
		tanfi = math.Tan(fh * radiansPerDegree)
		ra0 = state.sweDegnorm(armc + 90)
		xp[1] = 1
		xeq[1] = de
		fac = 2
		nloop = 0
		for math.Abs(xp[1]) > 0.000001 && nloop < 1000 {
			if xp[1] > 0 {
				fh = math.Atan(math.Tan(fh*radiansPerDegree)-tanfi/fac) * degreesPerRadian
				ra0 -= 90 / fac
			} else {
				fh = math.Atan(math.Tan(fh*radiansPerDegree)+tanfi/fac) * degreesPerRadian
				ra0 += 90 / fac
			}
			xeq[0] = state.sweDegnorm(ra - ra0)
			state.sweCotrans(xeq[:], xp[:], 90-fh)
			fac *= 2
			nloop++
		}
		hpos = state.sweDegnorm(ra0 - armc)
		if mdd > 180 {
			hpos = state.sweDegnorm(-hpos)
		}
		if !(isAboveHor != 0) {
			hpos = state.sweDegnorm(hpos + 180)
		}
		hpos = state.sweDegnorm(hpos-90)/30 + 1
	case 'P':
		fallthrough
	case 'G':
		if 90-math.Abs(de) <= math.Abs(latitude) {
			if de*latitude < 0 {
				xp[0] = state.sweDegnorm(90 + mdn/2)
			} else {
				xp[0] = state.sweDegnorm(270 + mdd/2)
			}
			if message != nil {
				copyText(message, textBytes("Otto Ludwig procedure within circumpolar regions."))
			}
		} else {
			sinad = math.Tan(de*radiansPerDegree) * math.Tan(latitude*radiansPerDegree)
			ad = math.Asin(sinad) * degreesPerRadian
			a = sinad + math.Cos(mdd*radiansPerDegree)
			if a >= 0 {
				isAboveHor = 1
			}
			sad = 90 + ad
			san = 90 - ad
			if isAboveHor != 0 {
				xp[0] = (mdd/sad + 3) * 90
			} else {
				xp[0] = (mdn/san + 1) * 90
			}
			xp[0] = state.sweDegnorm(xp[0] + 2.7777777777777776e-07)
		}
		if houseSystem == 'G' {
			xp[0] = 360 - xp[0]
			hpos = xp[0]/10.0 + 1
		} else {
			hpos = xp[0]/30.0 + 1
		}
	default:
		hpos = 0
		if state.sweHousesArmcEx2(armc, latitude, eps, houseSystem, hcusp[:], ascmc[:], nil, nil, message) == -1 {
			if message != nil {
				formatText(message, textBytes("swe_house_pos(): failed for system %c"), houseSystem)
			}
			break
		}
		if state.sweDifdeg2n(hcusp[6], hcusp[1]) > 0 {
			d = state.sweDegnorm(xpin[0] - hcusp[1])
			for i = 1; i <= 12; i++ {
				j = i + 1
				if j > 12 {
					c2 = 360
				} else {
					c2 = state.sweDegnorm(hcusp[j] - hcusp[1])
				}
				if d < c2 {
					break
				}
			}
			c1 = state.sweDegnorm(hcusp[i] - hcusp[1])
		} else {
			d = state.sweDegnorm(hcusp[1] - xpin[0])
			for i = 1; i <= 12; i++ {
				j = i + 1
				if j > 12 {
					c2 = 360
				} else {
					c2 = state.sweDegnorm(hcusp[1] - hcusp[j])
				}
				if d < c2 {
					break
				}
			}
			c1 = state.sweDegnorm(hcusp[1] - hcusp[i])
		}
		hsize = c2 - c1
		if hsize == 0 {
			hpos = float64(i)
		} else {
			hpos = float64(i) + (d-c1)/hsize
		}
		if message != nil {
			formatText(message, textBytes("swe_house_pos(): using simplified algorithm for system %c\n"), houseSystem)
		}
	}
	return hpos
}

// sunshineInit implements sunshine_init from swehouse.c.
func (state *State) sunshineInit(lat float64, dec float64, xh []float64) int32 {
	var (
		ad  float64
		nsa float64
		dsa float64
		arg float64
	)
	arg = math.Tan(dec*radiansPerDegree) * math.Tan(lat*radiansPerDegree)
	if arg >= 1 {
		ad = 89.999999999899998
	} else if arg <= float64(-1) {
		ad = -89.999999999899998
	} else {
		ad = math.Asin(arg) * degreesPerRadian
	}
	nsa = 90 - ad
	dsa = 90 + ad
	xh[2] = float64(-2) * nsa / 3
	xh[3] = float64(-1) * nsa / 3
	xh[5] = 1 * nsa / 3
	xh[6] = 2 * nsa / 3
	xh[8] = float64(-2) * dsa / 3
	xh[9] = float64(-1) * dsa / 3
	xh[11] = 1 * dsa / 3
	xh[12] = 2 * dsa / 3
	if math.Abs(arg) >= 1 {
		return -1
	}
	return 0
}

// sunshineSolutionMakransky implements sunshine_solution_makransky from swehouse.c.
func (state *State) sunshineSolutionMakransky(ramc float64, lat float64, ecl float64, hsp *houseData) int32 {
	var (
		xh     [13]float64
		md     float64
		zd     float64
		pole   float64
		q      float64
		w      float64
		a      float64
		b      float64
		c      float64
		f      float64
		cu     float64
		r      float64 = 0
		rah    float64
		sinlat float64
		coslat float64
		tanlat float64
		tandec float64
		sinecl float64
		dec    float64 = hsp.sunDeclination
	)
	sinlat = math.Sin(lat * radiansPerDegree)
	coslat = math.Cos(lat * radiansPerDegree)
	tanlat = math.Tan(lat * radiansPerDegree)
	tandec = math.Tan(dec * radiansPerDegree)
	sinecl = math.Sin(ecl * radiansPerDegree)
	var ih int32
	if state.sunshineInit(lat, dec, xh[:]) == -1 {
		return -1
	}
	for ih = 1; ih <= 12; ih++ {
		var z float64 = 0
		if (ih-1)%3 == 0 {
			continue
		}
		md = math.Abs(xh[ih])
		if ih <= 6 {
			rah = state.sweDegnorm(ramc + 180 + xh[ih])
		} else {
			rah = state.sweDegnorm(ramc + xh[ih])
		}
		if lat < 0 {
			rah = state.sweDegnorm(180 + rah)
		}
		if md == 90 {
			zd = 90.0 - math.Atan(sinlat*tandec)*degreesPerRadian
		} else {
			if md < 90 {
				a = math.Atan(coslat*math.Tan(md*radiansPerDegree)) * degreesPerRadian
			} else {
				a = math.Atan(math.Tan((md-90)*radiansPerDegree)/coslat) * degreesPerRadian
			}
			b = math.Atan(tanlat*math.Cos(md*radiansPerDegree)) * degreesPerRadian
			if ih <= 6 {
				c = b + dec
			} else {
				c = b - dec
			}
			f = math.Atan(sinlat*math.Sin(md*radiansPerDegree)*math.Tan(c*radiansPerDegree)) * degreesPerRadian
			zd = a + f
		}
		pole = math.Asin(math.Sin(zd*radiansPerDegree)*sinlat) * degreesPerRadian
		q = math.Asin(tandec*math.Tan(pole*radiansPerDegree)) * degreesPerRadian
		if ih <= 3 || ih >= 11 {
			w = state.sweDegnorm(rah - q)
		} else {
			w = state.sweDegnorm(rah + q)
		}
		if w == 90 {
			r = math.Atan(math.Sin(ecl*radiansPerDegree)*math.Tan(pole*radiansPerDegree)) * degreesPerRadian
			if ih <= 3 || ih >= 11 {
				cu = 90 + r
			} else {
				cu = 90 - r
			}
		} else if w == 270 {
			r = math.Atan(sinecl*math.Tan(pole*radiansPerDegree)) * degreesPerRadian
			if ih <= 3 || ih >= 11 {
				cu = 270 - r
			} else {
				cu = 270 + r
			}
		} else {
			var m float64
			m = math.Atan(math.Abs(math.Tan(pole*radiansPerDegree)/math.Cos(w*radiansPerDegree))) * degreesPerRadian
			if ih <= 3 || ih >= 11 {
				if w > 90 && w < 270 {
					z = m - ecl
				} else {
					z = m + ecl
				}
			} else if w > 90 && w < 270 {
				z = m + ecl
			} else {
				z = m - ecl
			}
			if z == 90 {
				if w < 180 {
					cu = 90
				} else {
					cu = 270
				}
			} else {
				r = math.Atan(math.Abs(math.Cos(m*radiansPerDegree)*math.Tan(w*radiansPerDegree)/math.Cos(z*radiansPerDegree))) * degreesPerRadian
				if w < 90 {
					cu = r
				} else if w > 90 && w < 180 {
					cu = 180 - r
				} else if w > 180 && w < 270 {
					cu = 180 + r
				} else {
					cu = 360 - r
				}
			}
			if z > 90 {
				if w < 90 {
					cu = 180 - r
				} else if w > 90 && w < 180 {
					cu = +r
				} else if w > 180 && w < 270 {
					cu = 360 - r
				} else {
					cu = 180 + r
				}
			}
			if lat < 0 {
				cu = state.sweDegnorm(cu + 180)
			}
		}
		hsp.cusp[ih] = cu
	}
	return 0
}

// sunshineSolutionTreindl implements sunshine_solution_treindl from swehouse.c.
func (state *State) sunshineSolutionTreindl(ramc float64, lat float64, ecl float64, hsp *houseData) int32 {
	var (
		xh             [13]float64
		mcdec          float64
		sinlat         float64
		coslat         float64
		cosdec         float64
		tandec         float64
		sinecl         float64
		cosecl         float64
		xhs            float64
		pole           float64
		a              float64
		cosa           float64
		alph           float64
		alpha2         float64
		c              float64
		cosc           float64
		b              float64
		sinzd          float64
		zd             float64
		rax            float64
		hc             float64
		ih             int32
		retval         int32 = 0
		mcUnderHorizon int32
		dec            float64 = hsp.sunDeclination
	)
	sinlat = math.Sin(lat * radiansPerDegree)
	coslat = math.Cos(lat * radiansPerDegree)
	cosdec = math.Cos(dec * radiansPerDegree)
	tandec = math.Tan(dec * radiansPerDegree)
	sinecl = math.Sin(ecl * radiansPerDegree)
	cosecl = math.Cos(ecl * radiansPerDegree)
	state.sunshineInit(lat, dec, xh[:])
	mcdec = math.Atan(math.Sin(ramc*radiansPerDegree)*math.Tan(ecl*radiansPerDegree)) * degreesPerRadian
	mcUnderHorizon = int32(boolInt(math.Abs(lat-mcdec) > 90))
	if mcUnderHorizon != 0 && 0 != 0 {
		for ih = 2; ih <= 12; ih++ {
			xh[ih] = -xh[ih]
		}
	}
	for ih = 1; ih <= 12; ih++ {
		if (ih-1)%3 == 0 {
			continue
		}
		xhs = 2 * (math.Asin(cosdec*math.Sin(xh[ih]/2*radiansPerDegree)) * degreesPerRadian)
		cosa = tandec * math.Tan(xhs/2*radiansPerDegree)
		alph = math.Acos(cosa) * degreesPerRadian
		if ih > 7 {
			alpha2 = 180 - alph
			b = 90 - lat + dec
		} else {
			alpha2 = alph
			b = 90 - lat - dec
		}
		cosc = math.Cos(xhs*radiansPerDegree)*math.Cos(b*radiansPerDegree) + math.Sin(xhs*radiansPerDegree)*math.Sin(b*radiansPerDegree)*math.Cos(alpha2*radiansPerDegree)
		c = math.Acos(cosc) * degreesPerRadian
		if c < 1e-6 {
			formatText(hsp.message[:], textBytes("Sunshine house %d c=%le very small"), ih, c)
			retval = -1
		}
		sinzd = math.Sin(xhs*radiansPerDegree) * math.Sin(alpha2*radiansPerDegree) / math.Sin(c*radiansPerDegree)
		zd = math.Asin(sinzd) * degreesPerRadian
		rax = math.Atan(coslat*math.Tan(zd*radiansPerDegree)) * degreesPerRadian
		pole = math.Asin(sinzd*sinlat) * degreesPerRadian
		if ih <= 6 {
			pole = -pole
			a = state.sweDegnorm(rax + ramc + 180)
		} else {
			a = state.sweDegnorm(ramc + rax)
		}
		hc = state.Asc1(a, pole, sinecl, cosecl)
		hsp.cusp[ih] = hc
	}
	if mcUnderHorizon != 0 && !(0 != 0) {
		for ih = 2; ih <= 12; ih++ {
			if (ih-1)%3 == 0 {
				continue
			}
			hsp.cusp[ih] = state.sweDegnorm(hsp.cusp[ih] + 180)
		}
	}
	return retval
}
