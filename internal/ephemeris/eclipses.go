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

// sweSolEclipseWhere implements swe_sol_eclipse_where from swecl.c.
//
//	Computes geographic location and type of solar eclipse
//
// for a given tjd
// iflag:	to indicate ephemeris to be used
//
//	(SEFLG_JPLEPH, SEFLG_SWIEPH, SEFLG_MOSEPH)
//
// Algorithms for the central line is taken from Montenbruck, pp. 179ff.,
// with the exception, that we consider refraction for the maxima of
// partial and noncentral eclipses.
// Geographical positions are referred to sea level / the mean ellipsoid.
//
// Errors:
//   - from uncertainty of JPL-ephemerides (0.01 arcsec):
//     about 40 meters
//   - from displacement of shadow points by atmospheric refraction:
//     a few meters
//   - from deviation of the geoid from the ellipsoid
//     a few meters
//   - from polar motion
//     a few meters
//
// For geographical locations that are interesting for observation,
// the error is always < 100 m.
// However, if the sun is close to the horizon,
// all of these errors can grow up to a km or more.
//
// Function returns:
// -1 (ERR)	on error (e.g. if swe_calc() for sun or moon fails)
// 0		if there is no solar eclipse at tjd
// SE_ECL_TOTAL
// SE_ECL_ANNULAR
// SE_ECL_TOTAL | SE_ECL_CENTRAL
// SE_ECL_TOTAL | SE_ECL_NONCENTRAL
// SE_ECL_ANNULAR | SE_ECL_CENTRAL
// SE_ECL_ANNULAR | SE_ECL_NONCENTRAL
// SE_ECL_PARTIAL
//
// geopos[0]:	geographic longitude of central line
// geopos[1]:	geographic latitude of central line
//
// not implemented so far:
//
// geopos[2]:	geographic longitude of northern limit of umbra
// geopos[3]:	geographic latitude of northern limit of umbra
// geopos[4]:	geographic longitude of southern limit of umbra
// geopos[5]:	geographic latitude of southern limit of umbra
// geopos[6]:	geographic longitude of northern limit of penumbra
// geopos[7]:	geographic latitude of northern limit of penumbra
// geopos[8]:	geographic longitude of southern limit of penumbra
// geopos[9]:	geographic latitude of southern limit of penumbra
//
// Attention: "northern" and "southern" limits of umbra do not
// necessarily correspond to the northernmost or southernmost
// geographic position, where the total, annular, or partial
// phase is visible at a given time.
// Imagine a situation in northern summer, when the sun illuminates
// the northern polar circle. The southernmost point of the core
// shadow may then touch the north pole, and therefore the
// northernmost point will be more in the south.
// Note also that with annular eclipses, the northern edge is
// usually geographically the southern one. With annular-total
// ones, the two lines cross, usually twice. The maximum is always
// total in such cases.
//
// attr[0]	fraction of solar diameter covered by moon (magnitude)
// attr[1]	ratio of lunar diameter to solar one
// attr[2]	fraction of solar disc covered by moon (obscuration)
// attr[3]      diameter of core shadow in km
// attr[4]	azimuth of sun at tjd
// attr[5]	true altitude of sun above horizon at tjd
// attr[6]	apparent altitude of sun above horizon at tjd
// attr[7]	angular distance of moon from sun in degrees
// attr[8]	magnitude acc. to NASA;
//
//	= attr[0] for partial and attr[1] for annular and total eclipses
//
// attr[9]	saros series number
// attr[10]	saros series member number
//
//	declare as attr[20] at least !
func (state *State) sweSolEclipseWhere(jdUT float64, ifl int32, location []float64, attr []float64, message []byte) int32 {
	var (
		retflag  int32
		retflag2 int32
		dcore    [10]float64
	)
	ifl &= 1 | 2 | 4
	state.swiSetTidAcc(jdUT, ifl, 0, message)
	if retflag = state.eclipseWhere(jdUT, 0, nil, ifl, location, dcore[:], message); retflag < 0 {
		return retflag
	}
	if retflag2 = state.eclipseHow(jdUT, 0, nil, ifl, location[0], location[1], 0, attr, message); retflag2 == -1 {
		return retflag2
	}
	attr[3] = dcore[0]
	return retflag
}

// sweLunOccultWhere implements swe_lun_occult_where from swecl.c.
//
//	  double tjd_ut,       time, Jul. day UT
//	  int32 ipl,           planet number
//	  char* starname,      star name, must be NULL or ”” if not a star
//	  int32 ifl,           ephemeris flag
//	  double *geopos,      return array, 2 doubles, geo. long. and lat.
//			       eastern longitude is positive,
//			       western longitude is negative,
//			       northern latitude is positive,
//	  double *attr,        return array, 20 doubles, see below
//	  char *serr           return error string
//
//	  array attr:
//	  attr[0]              fraction of object's diameter covered by moon (magnitude)
//	  attr[1]              ratio of lunar diameter to object's diameter
//	  attr[2]              fraction of object's disc covered by moon (obscuration)
//	  attr[3]              diameter of core shadow in km
//	  attr[4]              azimuth of object at tjd
//	  attr[5]              true altitude of object above horizon at tjd
//	  attr[6]              apparent altitude of object above horizon at tjd
//	  attr[7]              angular distance of moon from object in degrees
func (state *State) sweLunOccultWhere(jdUT float64, body int32, starname []byte, ifl int32, location []float64, attr []float64, message []byte) int32 {
	var (
		retflag  int32
		retflag2 int32
		dcore    [10]float64
	)
	if body < 0 {
		body = 0
	}
	ifl &= 1 | 2 | 4
	state.swiSetTidAcc(jdUT, ifl, 0, message)
	if body == 10000+134340 {
		body = 9
	}
	if retflag = state.eclipseWhere(jdUT, body, starname, ifl, location, dcore[:], message); retflag < 0 {
		return retflag
	}
	if retflag2 = state.eclipseHow(jdUT, body, starname, ifl, location[0], location[1], 0, attr, message); retflag2 == -1 {
		return retflag2
	}
	attr[3] = dcore[0]
	return retflag
}

func (state *State) eclipseWhere(jdUT float64, body int32, starname []byte, ifl int32, location []float64, dcore []float64, message []byte) int32 {
	var (
		i             int32
		status        int32 = 0
		niter         int32 = 0
		e             [6]float64
		et            [6]float64
		rm            [6]float64
		rs            [6]float64
		rmt           [6]float64
		rst           [6]float64
		xs            [6]float64
		xst           [6]float64
		x             [6]float64
		lm            [6]float64
		ls            [6]float64
		lx            [6]float64
		dsm           float64
		dsmt          float64
		d0            float64
		D0            float64
		s0            float64
		r0            float64
		d             float64
		s             float64
		dm            float64
		de            float64 = 4.2635232508025265e-05
		earthobl      float64 = 0.99664718030210375
		deltat        float64
		jd            float64
		sidt          float64
		drad          float64
		sinf1         float64
		sinf2         float64
		cosf1         float64
		cosf2         float64
		rmoon         float64 = 1.1618815106570899e-05
		dmoon         float64 = 2 * rmoon
		flags         int32
		iflag2        int32
		noEclipse     int32      = 0
		obliquityData *obliquity = &state.positions.data.obliquity
	)
	for i = 0; i < 10; i++ {
		dcore[i] = 0
	}
	flags = 256 | 2*1024 | ifl
	iflag2 = flags | (flagRadians)
	flags = flags | (flagCartesian)
	deltat = state.sweDeltatEx(jdUT, ifl, message)
	jd = jdUT + deltat
	if status = state.sweCalc(jd, 1, flags, rm[:], message); status == -1 {
		return status
	}
	if status = state.sweCalc(jd, 1, iflag2, lm[:], message); status == -1 {
		return status
	}
	if status = state.calcPlanetStar(jd, body, starname, flags, rs[:], message); status == -1 {
		return status
	}
	if status = state.calcPlanetStar(jd, body, starname, iflag2, ls[:], message); status == -1 {
		return status
	}
	for i = 0; i <= 2; i++ {
		rst[i] = rs[i]
	}
	for i = 0; i <= 2; i++ {
		rmt[i] = rm[i]
	}
	if flags&flagNoNutation != 0 {
		sidt = state.sweSidtime0(jdUT, obliquityData.angle*degreesPerRadian, 0) * 15 * radiansPerDegree
	} else {
		sidt = state.sweSidtime(jdUT) * 15 * radiansPerDegree
	}
	if starname != nil && int32(starname[0]) != '\000' {
		drad = 0
	} else if body < 20+1 {
		drad = calendarPlaDiam[body] / 2 / 1.49597870700e+11
	} else if body > 10000 {
		drad = state.positions.data.astDiam / 2 * 1000 / 1.49597870700e+11
	} else {
		drad = 0
	}
iterWhere:
	for {
		for i = 0; i <= 2; i++ {
			rs[i] = rst[i]
			rm[i] = rmt[i]
		}
		for i = 0; i <= 2; i++ {
			lx[i] = lm[i]
		}
		state.swiPolcart(lx[:], rm[:])
		rm[2] /= earthobl
		dm = math.Sqrt(rm[0]*rm[0] + rm[1]*rm[1] + rm[2]*rm[2])
		for i = 0; i <= 2; i++ {
			lx[i] = ls[i]
		}
		state.swiPolcart(lx[:], rs[:])
		rs[2] /= earthobl
		for i = 0; i <= 2; i++ {
			e[i] = rm[i] - rs[i]
			et[i] = rmt[i] - rst[i]
		}
		dsm = math.Sqrt(e[0]*e[0] + e[1]*e[1] + e[2]*e[2])
		dsmt = math.Sqrt(et[0]*et[0] + et[1]*et[1] + et[2]*et[2])
		for i = 0; i <= 2; i++ {
			e[i] /= dsm
			et[i] /= dsmt
		}
		sinf1 = (drad - rmoon) / dsm
		cosf1 = math.Sqrt(1 - sinf1*sinf1)
		sinf2 = (drad + rmoon) / dsm
		cosf2 = math.Sqrt(1 - sinf2*sinf2)
		s0 = -(rm[0]*e[0] + rm[1]*e[1] + rm[2]*e[2])
		r0 = math.Sqrt(dm*dm - s0*s0)
		d0 = (s0/dsm*(drad*2-dmoon) - dmoon) / cosf1
		D0 = (s0/dsm*(drad*2+dmoon) + dmoon) / cosf2
		dcore[2] = r0
		dcore[3] = d0
		dcore[4] = D0
		dcore[5] = cosf1
		dcore[6] = cosf2
		for i = 2; i < 5; i++ {
			dcore[i] *= 149597870.69999999
		}
		status = 0
		if de*cosf1 >= r0 {
			status |= 1
		} else if r0 <= de*cosf1+math.Abs(d0)/2 {
			status |= 2
		} else if r0 <= de*cosf2+D0/2 {
			status |= 16 | 2
		} else {
			if message != nil {
				formatText(message, textBytes("no solar eclipse at tjd = %f"), jd)
			}
			for i = 0; i < 10; i++ {
				location[i] = 0
			}
			dcore[0] = 0
			status = 0
			d = 0
			noEclipse = 1
		}
		d = s0*s0 + de*de - dm*dm
		if d > 0 {
			d = math.Sqrt(d)
		} else {
			d = 0
		}
		s = s0 - d
		for i = 0; i <= 2; i++ {
			xs[i] = rm[i] + s*e[i]
		}
		for i = 0; i <= 2; i++ {
			xst[i] = xs[i]
		}
		xst[2] *= earthobl
		state.swiCartpol(xst[:], xst[:])
		if niter <= 0 {
			var (
				cosfi float64 = math.Cos(xst[1])
				sinfi float64 = math.Sin(xst[1])
				eobl  float64 = 0.003352819697896193
				cc    float64 = 1 / math.Sqrt(cosfi*cosfi+(1-eobl)*(1-eobl)*sinfi*sinfi)
				ss    float64 = (1 - eobl) * (1 - eobl) * cc
			)
			earthobl = ss
			niter++
			continue iterWhere
		}
		state.swiPolcart(xst[:], xst[:])
		state.swiCartpol(xs[:], xs[:])
		xs[0] -= sidt
		xs[0] *= degreesPerRadian
		xs[1] *= degreesPerRadian
		xs[0] = state.sweDegnorm(xs[0])
		if xs[0] > 180 {
			xs[0] -= 360
		}
		location[0] = xs[0]
		location[1] = xs[1]
		for i = 0; i <= 2; i++ {
			x[i] = rmt[i] - xst[i]
		}
		s = math.Sqrt(x[0]*x[0] + x[1]*x[1] + x[2]*x[2])
		dcore[0] = (s/dsmt*(drad*2-dmoon) - dmoon) * cosf1
		dcore[0] *= 149597870.69999999
		dcore[1] = (s/dsmt*(drad*2+dmoon) + dmoon) * cosf2
		dcore[1] *= 149597870.69999999
		if !(status&16 != 0) && !(noEclipse != 0) {
			if dcore[0] > 0 {
				status |= 8
			} else {
				status |= 4
			}
		}
		return status
	}
}

// calcPlanetStar implements calc_planet_star from swecl.c.
//
// printf("total\n");
func (state *State) calcPlanetStar(jdTT float64, body int32, starname []byte, flags int32, x []float64, message []byte) int32 {
	var status int32 = 0
	if starname == nil || int32(starname[0]) == '\000' {
		status = state.sweCalc(jdTT, body, flags, x, message)
	} else {
		status = state.sweFixstar(starname, jdTT, flags, x, message)
	}
	return status
}

// sweSolEclipseHow implements swe_sol_eclipse_how from swecl.c.
//
//	Computes attributes of a solar eclipse for given tjd, geo. longitude,
//
// geo. latitude, and geo. height.
//
// retflag	SE_ECL_TOTAL or SE_ECL_ANNULAR or SE_ECL_PARTIAL
//
//	SE_ECL_NONCENTRAL
//	if 0, no eclipse is visible at geogr. position.
//
// attr[0]	fraction of solar diameter covered by moon;
//
//	with total/annular eclipses, it results in magnitude acc. to IMCCE.
//
// attr[1]	ratio of lunar diameter to solar one
// attr[2]	fraction of solar disc covered by moon (obscuration)
// attr[3]      diameter of core shadow in km
// attr[4]	azimuth of sun at tjd
// attr[5]	true altitude of sun above horizon at tjd
// attr[6]	apparent altitude of sun above horizon at tjd
// attr[7]	elongation of moon in degrees
// attr[8]	magnitude acc. to NASA;
//
//	= attr[0] for partial and attr[1] for annular and total eclipses
//
// attr[9]	saros series number
// attr[10]	saros series member number
//
//	declare as attr[20] at least !
func (state *State) sweSolEclipseHow(jdUT float64, ifl int32, location []float64, attr []float64, message []byte) int32 {
	var (
		retflag  int32
		retflag2 int32
		i        int32
		dcore    [10]float64
		ls       [6]float64
		xaz      [6]float64
		geopos2  [20]float64
	)
	for i = 0; i <= 10; i++ {
		attr[i] = 0
	}
	if location[2] < -500.0 || location[2] > 25000.0 {
		if message != nil {
			formatText(message, textBytes("location for eclipses must be between %.0f and %.0f m above sea"), -500.0, 25000.0)
		}
		return -1
	}
	ifl &= 1 | 2 | 4
	state.swiSetTidAcc(jdUT, ifl, 0, message)
	if retflag = state.eclipseHow(jdUT, 0, nil, ifl, location[0], location[1], location[2], attr, message); retflag == -1 {
		return retflag
	}
	if retflag2 = state.eclipseWhere(jdUT, 0, nil, ifl, geopos2[:], dcore[:], message); retflag2 == -1 {
		return retflag2
	}
	if retflag != 0 {
		retflag |= retflag2 & (1 | 2)
	}
	attr[3] = dcore[0]
	state.sweSetTopo(location[0], location[1], location[2])
	if state.sweCalcUt(jdUT, 0, ifl|32*1024|2*1024, ls[:], message) == -1 {
		return -1
	}
	state.sweAzalt(jdUT, 1, location, 0, 10, ls[:], xaz[:])
	attr[4] = xaz[0]
	attr[5] = xaz[1]
	attr[6] = xaz[2]
	if xaz[2] <= 0 {
		retflag = 0
	}
	if retflag == 0 {
		for i = 0; i <= 3; i++ {
			attr[i] = 0
		}
		for i = 8; i <= 10; i++ {
			attr[i] = 0
		}
	}
	return retflag
}

// eclipseHow implements eclipse_how from swecl.c.
func (state *State) eclipseHow(jdUT float64, body int32, starname []byte, ifl int32, longitude float64, latitude float64, geohgt float64, attr []float64, message []byte) int32 {
	var (
		i         int32
		j         int32
		k         int32
		status    int32 = 0
		te        float64
		d         float64
		xs        [6]float64
		xm        [6]float64
		ls        [6]float64
		lm        [6]float64
		x1        [6]float64
		x2        [6]float64
		rmoon     float64
		rsun      float64
		rsplusrm  float64
		rsminusrm float64
		dctr      float64
		drad      float64
		flags     int32 = 2*1024 | 32*1024 | ifl
		iflagcart int32 = flags | (flagCartesian)
		xh        [6]float64
		hminAppr  float64
		lsun      float64
		lmoon     float64
		lctr      float64
		lsunleft  float64
		a         float64
		b         float64
		sc1       float64
		sc2       float64
		location  [3]float64
	)
	for i = 0; i < 10; i++ {
		attr[i] = 0
	}
	location[0] = longitude
	location[1] = latitude
	location[2] = geohgt
	te = jdUT + state.sweDeltatEx(jdUT, ifl, message)
	state.sweSetTopo(longitude, latitude, geohgt)
	if state.calcPlanetStar(te, body, starname, flags, ls[:], message) == -1 {
		return -1
	}
	if state.sweCalc(te, 1, flags, lm[:], message) == -1 {
		return -1
	}
	if state.calcPlanetStar(te, body, starname, iflagcart, xs[:], message) == -1 {
		return -1
	}
	if state.sweCalc(te, 1, iflagcart, xm[:], message) == -1 {
		return -1
	}
	if starname != nil && int32(starname[0]) != '\000' {
		drad = 0
	} else if body < 20+1 {
		drad = calendarPlaDiam[body] / 2 / 1.49597870700e+11
	} else if body > 10000 {
		drad = state.positions.data.astDiam / 2 * 1000 / 1.49597870700e+11
	} else {
		drad = 0
	}
	state.sweAzalt(jdUT, 1, location[:], 0, 10, ls[:], xh[:])
	rmoon = math.Asin(1.1618815106570899e-05/lm[2]) * degreesPerRadian
	rsun = math.Asin(drad/ls[2]) * degreesPerRadian
	rsplusrm = rsun + rmoon
	rsminusrm = rsun - rmoon
	for i = 0; i < 3; i++ {
		x1[i] = xs[i] / ls[2]
		x2[i] = xm[i] / lm[2]
	}
	dctr = math.Acos(state.swiDotProdUnit(x1[:], x2[:])) * degreesPerRadian
	if dctr < rsminusrm {
		status = 8
	} else if dctr < math.Abs(rsminusrm) {
		status = 4
	} else if dctr < rsplusrm {
		status = 16
	} else {
		status = 0
		if message != nil {
			formatText(message, textBytes("no solar eclipse at tjd = %f"), jdUT)
		}
	}
	if rsun > 0 {
		attr[1] = rmoon / rsun
	} else {
		attr[1] = 0
	}
	lsun = math.Asin(rsun/2*radiansPerDegree) * 2
	lsunleft = -dctr + rsun + rmoon
	if lsun > 0 {
		attr[0] = lsunleft / rsun / 2
	} else {
		attr[0] = 1
	}
	lsun = rsun
	lmoon = rmoon
	lctr = dctr
	if status == 0 || lsun == 0 {
		attr[2] = 1
	} else if status == 4 || status == 8 {
		attr[2] = lmoon * lmoon / lsun / lsun
	} else {
		a = 2 * lctr * lmoon
		b = 2 * lctr * lsun
		if a < 1e-9 {
			attr[2] = lmoon * lmoon / lsun / lsun
		} else {
			a = (lctr*lctr + lmoon*lmoon - lsun*lsun) / a
			if a > 1 {
				a = 1
			}
			if a < float64(-1) {
				a = float64(-1)
			}
			b = (lctr*lctr + lsun*lsun - lmoon*lmoon) / b
			if b > 1 {
				b = 1
			}
			if b < float64(-1) {
				b = float64(-1)
			}
			a = math.Acos(a)
			b = math.Acos(b)
			sc1 = a * lmoon * lmoon / 2
			sc2 = b * lsun * lsun / 2
			sc1 -= math.Cos(a) * math.Sin(a) * lmoon * lmoon / 2
			sc2 -= math.Cos(b) * math.Sin(b) * lsun * lsun / 2
			attr[2] = (sc1 + sc2) * 2 / 3.14159265358979323846 / lsun / lsun
		}
	}
	attr[7] = dctr
	hminAppr = -(34.4556 + 2.1200000000000001*math.Sqrt(geohgt)) / 60
	if xh[1]+rsun+math.Abs(hminAppr) >= 0 && status != 0 {
		status |= 128
	}
	attr[4] = xh[0]
	attr[5] = xh[1]
	attr[6] = xh[2]
	if body == 0 && (starname == nil || int32(starname[0]) == '\000') {
		attr[8] = attr[0]
		if status&(4|8) != 0 {
			attr[8] = attr[1]
		}
		for i = 0; i < 181; i++ {
			d = (jdUT - eclipsesSarosDataSolar[i].startEpoch) / 6585.3213
			if d < 0 && d*6585.3213 > float64(-2) {
				d = 0.0000001
			}
			if d < 0 {
				continue
			}
			j = int32(d)
			if (d-float64(j))*6585.3213 < 2 {
				attr[9] = float64(eclipsesSarosDataSolar[i].seriesNo)
				attr[10] = float64(j) + 1
				break
			}
			k = j + 1
			if (float64(k)-d)*6585.3213 < 2 {
				attr[9] = float64(eclipsesSarosDataSolar[i].seriesNo)
				attr[10] = float64(k) + 1
				break
			}
		}
		if i == 181 {
			attr[10] = float64(-99999999)
			attr[9] = attr[10]
		}
	}
	return status
}

// sweSolEclipseWhenGlob implements swe_sol_eclipse_when_glob from swecl.c.
//
//	When is the next solar eclipse anywhere on earth?
//
// input parameters:
//
// tjd_start    start time for search (UT)
// ifl		ephemeris to be used (SEFLG_SWIEPH, etc.)
// ifltype	eclipse type to be searched (SE_ECL_TOTAL, etc.)
//
//	0, if any type of eclipse is wanted
//
// return values:
//
// retflag	SE_ECL_TOTAL or SE_ECL_ANNULAR or SE_ECL_PARTIAL
//
//	or SE_ECL_ANNULAR_TOTAL
//	SE_ECL_CENTRAL
//	SE_ECL_NONCENTRAL
//
// tret[0]	time of maximum eclipse
// tret[1]	time, when eclipse takes place at local apparent noon
// tret[2]	time of eclipse begin
// tret[3]	time of eclipse end
// tret[4]	time of totality begin
// tret[5]	time of totality end
// tret[6]	time of center line begin
// tret[7]	time of center line end
// tret[8]	time when annular-total eclipse becomes total
//
//	not implemented so far
//
// tret[9]	time when annular-total eclipse becomes annular again
//
//		   not implemented so far
//	   declare as tret[10] at least!
func (state *State) sweSolEclipseWhenGlob(tjdStart float64, ifl int32, ifltype int32, tret []float64, backward int32, message []byte) int32 {
	var (
		i         int32
		j         int32
		k         int32
		m         int32
		n         int32
		o         int32
		i1        int32   = 0
		i2        int32   = 0
		retflag   int32   = 0
		retflag2  int32   = 0
		de        float64 = 6378.140
		a         float64
		t         float64
		tt        float64
		jd        float64
		tjds      float64
		dt        float64
		dtint     [1]float64
		dta       float64
		dtb       float64
		T         float64
		T2        float64
		T3        float64
		T4        float64
		K         float64
		M         float64
		Mm        float64
		E         float64
		Ff        float64
		xs        [6]float64
		xm        [6]float64
		ls        [6]float64
		lm        [6]float64
		rmoon     float64
		rsun      float64
		dcore     [10]float64
		dc        [3]float64
		dctr      [1]float64
		twohr     float64    = 0.083333333333333329
		tenmin    float64    = 0.0069444444444444449
		dt1       [1]float64 = [1]float64{0}
		dt2       [1]float64 = [1]float64{0}
		location  [20]float64
		attr      [20]float64
		dtstart   float64
		dtdiv     float64
		xa        [6]float64
		xb        [6]float64
		direction int32 = 1
		dontTimes int32 = 0
		flags     int32
		iflagcart int32
	)
	ifl &= 1 | 2 | 4
	state.swiSetTidAcc(tjdStart, ifl, 0, message)
	flags = 2*1024 | ifl
	iflagcart = flags | (flagCartesian)
	if ifltype == 16|1 {
		if message != nil {
			copyText(message, textBytes("central partial eclipses do not exist"))
		}
		return -1
	}
	if ifltype == 32|2 {
		if message != nil {
			copyText(message, textBytes("non-central hybrid (annular-total) eclipses do not exist"))
		}
		return -1
	}
	if ifltype == 0 {
		ifltype = 4 | 8 | 16 | 32 | 2 | 1
	}
	if ifltype == 4 || ifltype == 8 || ifltype == 32 {
		ifltype |= 2 | 1
	}
	if ifltype == 16 {
		ifltype |= 2
	}
	if backward != 0 {
		direction = -1
	}
	K = float64(int32((tjdStart - julianDayJ2000) / 365.2425 * 12.3685))
	K -= float64(direction)
nextTry:
	for {
		retflag = 0
		dontTimes = 0
		for i = 0; i <= 9; i++ {
			tret[i] = 0
		}
		T = K / 1236.85
		T2 = T * T
		T3 = T2 * T
		T4 = T3 * T
		Ff = state.sweDegnorm(160.7108 + 390.67050274*K - 0.0016341*T2 - 0.00000227*T3 + 0.000000011*T4)
		if Ff > 180 {
			Ff -= 180
		}
		if Ff > 21 && Ff < 159 {
			K += float64(direction)
			continue nextTry
		}
		jd = 2451550.09765 + 29.530588853*K + 0.0001337*T2 - 0.000000150*T3 + 0.00000000073*T4
		M = state.sweDegnorm(2.5534 + 29.10535669*K - 0.0000218*T2 - 0.00000011*T3)
		Mm = state.sweDegnorm(201.5643 + 385.81693528*K + 0.1017438*T2 + 0.00001239*T3 + 0.000000058*T4)
		E = 1 - 0.002516*T - 0.0000074*T2
		M *= radiansPerDegree
		Mm *= radiansPerDegree
		jd = jd - 0.4075*math.Sin(Mm) + 0.1721*E*math.Sin(M)
		dtstart = 1
		if jd < 2000000 || jd > 2500000 {
			dtstart = 5
		}
		dtdiv = 4
		for dt = dtstart; dt > 0.0001; dt /= dtdiv {
			i = 0
			for t = jd - dt; i <= 2; {
				if state.sweCalc(t, 0, flags, ls[:], message) == -1 {
					return -1
				}
				if state.sweCalc(t, 1, flags, lm[:], message) == -1 {
					return -1
				}
				if state.sweCalc(t, 0, iflagcart, xs[:], message) == -1 {
					return -1
				}
				if state.sweCalc(t, 1, iflagcart, xm[:], message) == -1 {
					return -1
				}
				for m = 0; m < 3; m++ {
					xa[m] = xs[m] / ls[2]
					xb[m] = xm[m] / lm[2]
				}
				dc[i] = math.Acos(state.swiDotProdUnit(xa[:], xb[:])) * degreesPerRadian
				rmoon = math.Asin(1.1618815106570899e-05/lm[2]) * degreesPerRadian
				rsun = math.Asin(0.0046524726370988385/ls[2]) * degreesPerRadian
				dc[i] -= rmoon + rsun
				i++
				t += dt
			}
			state.findMaximum(dc[0], dc[1], dc[2], dt, dtint[:], dctr[:])
			jd += dtint[0] + dt
		}
		tjds = jd - state.sweDeltatEx(jd, ifl, message)
		tjds = jd - state.sweDeltatEx(tjds, ifl, message)
		jd = jd - state.sweDeltatEx(tjds, ifl, message)
		tjds = jd
		if retflag = state.eclipseWhere(jd, 0, nil, ifl, location[:], dcore[:], message); retflag == -1 {
			return retflag
		}
		retflag2 = retflag
		if retflag2 = state.eclipseHow(jd, 0, nil, ifl, location[0], location[1], 0, attr[:], message); retflag2 == -1 {
			return retflag2
		}
		if retflag2 == 0 {
			K += float64(direction)
			continue nextTry
		}
		tret[0] = jd
		if backward != 0 && tret[0] >= tjdStart-0.0001 || !(backward != 0) && tret[0] <= tjdStart+0.0001 {
			K += float64(direction)
			continue nextTry
		}
		if retflag = state.eclipseWhere(jd, 0, nil, ifl, location[:], dcore[:], message); retflag == -1 {
			return retflag
		}
		if retflag == 0 {
			retflag = 16 | 2
			tret[5] = jd
			tret[4] = tret[5]
			dontTimes = 1
		}
		if !(ifltype&2 != 0) && retflag&2 != 0 {
			K += float64(direction)
			continue nextTry
		}
		if !(ifltype&1 != 0) && retflag&1 != 0 {
			K += float64(direction)
			continue nextTry
		}
		if !(ifltype&8 != 0) && retflag&8 != 0 {
			K += float64(direction)
			continue nextTry
		}
		if !(ifltype&16 != 0) && retflag&16 != 0 {
			K += float64(direction)
			continue nextTry
		}
		if !(ifltype&(4|32) != 0) && retflag&4 != 0 {
			K += float64(direction)
			continue nextTry
		}
		if dontTimes != 0 {
			return retflag
		}
		if retflag&16 != 0 {
			o = 0
		} else if retflag&2 != 0 {
			o = 1
		} else {
			o = 2
		}
		dta = twohr
		dtb = tenmin / 3.0
		for n = 0; n <= o; n++ {
			if n == 0 {
				i1 = 2
				i2 = 3
			} else if n == 1 {
				if retflag&16 != 0 {
					continue
				}
				i1 = 4
				i2 = 5
			} else if n == 2 {
				if retflag&2 != 0 {
					continue
				}
				i1 = 6
				i2 = 7
			}
			i = 0
			for t = jd - dta; i <= 2; {
				if retflag2 = state.eclipseWhere(t, 0, nil, ifl, location[:], dcore[:], message); retflag2 == -1 {
					return retflag2
				}
				if n == 0 {
					dc[i] = dcore[4]/2 + de/dcore[5] - dcore[2]
				} else if n == 1 {
					dc[i] = math.Abs(dcore[3])/2 + de/dcore[6] - dcore[2]
				} else if n == 2 {
					dc[i] = de/dcore[6] - dcore[2]
				}
				i += 1
				t += dta
			}
			state.findZero(dc[0], dc[1], dc[2], dta, dt1[:], dt2[:])
			tret[i1] = jd + dt1[0] + dta
			tret[i2] = jd + dt2[0] + dta
			m = 0
			for dt = dtb; m < 3; {
				for j = i1; j <= i2; j += i2 - i1 {
					i = 0
					for t = tret[j] - dt; i < 2; {
						if retflag2 = state.eclipseWhere(t, 0, nil, ifl, location[:], dcore[:], message); retflag2 == -1 {
							return retflag2
						}
						if n == 0 {
							dc[i] = dcore[4]/2 + de/dcore[5] - dcore[2]
						} else if n == 1 {
							dc[i] = math.Abs(dcore[3])/2 + de/dcore[6] - dcore[2]
						} else if n == 2 {
							dc[i] = de/dcore[6] - dcore[2]
						}
						i++
						t += dt
					}
					dt1[0] = dc[1] / ((dc[1] - dc[0]) / dt)
					tret[j] -= dt1[0]
				}
				m++
				dt /= 3
			}
		}
		if retflag&4 != 0 {
			if retflag2 = state.eclipseWhere(tret[0], 0, nil, ifl, location[:], dcore[:], message); retflag2 == -1 {
				return retflag2
			}
			dc[0] = dcore[0]
			if retflag2 = state.eclipseWhere(tret[4], 0, nil, ifl, location[:], dcore[:], message); retflag2 == -1 {
				return retflag2
			}
			dc[1] = dcore[0]
			if retflag2 = state.eclipseWhere(tret[5], 0, nil, ifl, location[:], dcore[:], message); retflag2 == -1 {
				return retflag2
			}
			dc[2] = dcore[0]
			if dc[0]*dc[1] < 0 || dc[0]*dc[2] < 0 {
				retflag |= 32
				retflag &= ^4
			}
		}
		if !(ifltype&4 != 0) && retflag&4 != 0 {
			K += float64(direction)
			continue nextTry
		}
		if !(ifltype&32 != 0) && retflag&32 != 0 {
			K += float64(direction)
			continue nextTry
		}
		k = 2
		for i = 0; i < 2; i++ {
			j = i + k
			tt = tret[j] + state.sweDeltatEx(tret[j], ifl, message)
			if state.sweCalc(tt, 0, flags, ls[:], message) == -1 {
				return -1
			}
			if state.sweCalc(tt, 1, flags, lm[:], message) == -1 {
				return -1
			}
			dc[i] = state.sweDegnorm(ls[0] - lm[0])
			if dc[i] > 180 {
				dc[i] -= 360
			}
		}
		if dc[0]*dc[1] >= 0 {
			tret[1] = 0
		} else {
			jd = tjds
			dt = 0.1
			dt1[0] = (tret[3] - tret[2]) / 2.0
			if dt1[0] < dt {
				dt = dt1[0] / 2.0
			}
			for j = 0; dt > 0.01; {
				i = 0
				for t = jd; i <= 1; {
					tt = t + state.sweDeltatEx(t, ifl, message)
					if state.sweCalc(tt, 0, flags, ls[:], message) == -1 {
						return -1
					}
					if state.sweCalc(tt, 1, flags, lm[:], message) == -1 {
						return -1
					}
					dc[i] = state.sweDegnorm(ls[0] - lm[0])
					if dc[i] > 180 {
						dc[i] -= 360
					}
					if dc[i] > 180 {
						dc[i] -= 360
					}
					i++
					t -= dt
				}
				a = (dc[1] - dc[0]) / dt
				if a < 1e-10 {
					break
				}
				dt1[0] = dc[0] / a
				jd += dt1[0]
				j++
				dt /= 3
			}
			tret[1] = jd
		}
		return retflag
	}
}

func (state *State) sweLunOccultWhenGlob(tjdStart float64, body int32, starname []byte, ifl int32, ifltype int32, tret []float64, backward int32, message []byte) int32 {
	var (
		i         int32
		j         int32
		k         int32
		m         int32
		n         int32
		o         int32
		i1        int32
		i2        int32
		retflag   int32   = 0
		retflag2  int32   = 0
		de        float64 = 6378.140
		a         float64
		t         float64
		tt        float64
		jd        float64 = 0
		tjds      float64
		dt        float64
		dtint     [1]float64
		dta       float64
		dtb       float64
		drad      float64
		dl        float64
		xs        [6]float64
		xm        [6]float64
		ls        [6]float64
		lm        [6]float64
		rmoon     float64
		rsun      float64
		dcore     [10]float64
		dc        [20]float64
		dctr      [1]float64
		twohr     float64    = 0.083333333333333329
		tenmin    float64    = 0.0069444444444444449
		dt1       [1]float64 = [1]float64{0}
		dt2       [1]float64 = [1]float64{0}
		dadd2     float64    = 1
		location  [20]float64
		dtstart   float64
		dtdiv     float64
		direction int32 = 1
		ifltype2  int32
		flags     int32
		iflagcart int32
		dontTimes int32 = 0
		oneTry    int32 = backward & (32 * 1024)
	)
	if body < 0 {
		body = 0
	}
	if body == 10000+134340 {
		body = 9
	}
	ifl &= 1 | 2 | 4
	state.swiSetTidAcc(tjdStart, ifl, 0, message)
	flags = 2*1024 | ifl
	iflagcart = flags | (flagCartesian)
	backward &= 1
	if ifltype == 16|1 {
		if message != nil {
			copyText(message, textBytes("central partial eclipses do not exist"))
		}
		return -1
	}
	if body != 0 {
		ifltype2 = ifltype & ^(2 | 1)
		if ifltype2 == 8 || ifltype == 32 {
			if message != nil {
				formatText(message, textBytes("annular occulation do not exist for object %d %s\n"), body, starname)
			}
			return -1
		}
	}
	if body != 0 && ifltype&(8|32) != 0 {
		ifltype &= ^(8 | 32)
	}
	if ifltype == 0 {
		ifltype = 4 | 16 | 2 | 1
		if body == 0 {
			ifltype |= 8 | 32
		}
	}
	if ifltype&(4|8|32) != 0 {
		ifltype |= 2 | 1
	}
	if ifltype&16 != 0 {
		ifltype |= 2
	}
	retflag = 0
	for i = 0; i <= 9; i++ {
		tret[i] = 0
	}
	if backward != 0 {
		direction = -1
	}
	t = tjdStart
	jd = t
nextTry:
	for {
		if state.calcPlanetStar(t, body, starname, ifl, ls[:], message) == -1 {
			return -1
		}
		if math.Abs(ls[1]) > 7 && starname != nil && int32(starname[0]) != '\000' {
			if message != nil {
				formatText(message, textBytes("occultation never occurs: star %s has ecl. lat. %.1f"), starname, ls[1])
			}
			return -1
		}
		if state.sweCalc(t, 1, ifl, lm[:], message) == -1 {
			return -1
		}
		dl = state.sweDegnorm(ls[0] - lm[0])
		if direction < 0 {
			dl -= 360
		}
		for math.Abs(dl) > 0.1 {
			t += dl / 13
			if state.calcPlanetStar(t, body, starname, ifl, ls[:], message) == -1 {
				return -1
			}
			if state.sweCalc(t, 1, ifl, lm[:], message) == -1 {
				return -1
			}
			dl = state.sweDegnorm(ls[0] - lm[0])
			if dl > 180 {
				dl -= 360
			}
		}
		jd = t
		drad = math.Abs(ls[1] - lm[1])
		if drad > 2 {
			if oneTry != 0 {
				tret[0] = t + float64(direction)
				return 0
			}
			t += float64(direction * 20)
			jd = t
			continue nextTry
		}
		if starname != nil && int32(starname[0]) != '\000' {
			drad = 0
		} else if body < 20+1 {
			drad = calendarPlaDiam[body] / 2 / 1.49597870700e+11
		} else if body > 10000 {
			drad = state.positions.data.astDiam / 2 * 1000 / 1.49597870700e+11
		} else {
			drad = 0
		}
		dtstart = dadd2
		dtdiv = 3
		for dt = dtstart; dt > 0.0001; dt /= dtdiv {
			i = 0
			for t = jd - dt; i <= 2; {
				if state.calcPlanetStar(t, body, starname, flags, ls[:], message) == -1 {
					return -1
				}
				if state.sweCalc(t, 1, flags, lm[:], message) == -1 {
					return -1
				}
				if state.calcPlanetStar(t, body, starname, iflagcart, xs[:], message) == -1 {
					return -1
				}
				if state.sweCalc(t, 1, iflagcart, xm[:], message) == -1 {
					return -1
				}
				dc[i] = math.Acos(state.swiDotProdUnit(xs[:], xm[:])) * degreesPerRadian
				rmoon = math.Asin(1.1618815106570899e-05/lm[2]) * degreesPerRadian
				rsun = math.Asin(drad/ls[2]) * degreesPerRadian
				dc[i] -= rmoon + rsun
				i++
				t += dt
			}
			state.findMaximum(dc[0], dc[1], dc[2], dt, dtint[:], dctr[:])
			jd += dtint[0] + dt
		}
		jd -= state.sweDeltatEx(jd, ifl, message)
		tjds = jd
		if retflag = state.eclipseWhere(jd, body, starname, ifl, location[:], dcore[:], message); retflag == -1 {
			return retflag
		}
		retflag2 = retflag
		if retflag2 == 0 {
			if oneTry != 0 {
				tret[0] = jd
				return 0
			}
			t = jd + float64(direction*20)
			jd = t
			continue nextTry
		}
		tret[0] = jd
		if backward != 0 && tret[0] >= tjdStart-0.0001 || !(backward != 0) && tret[0] <= tjdStart+0.0001 {
			t = jd + float64(direction*20)
			jd = t
			continue nextTry
		}
		if retflag = state.eclipseWhere(jd, body, starname, ifl, location[:], dcore[:], message); retflag == -1 {
			return retflag
		}
		if retflag == 0 {
			retflag = 16 | 2
			tret[5] = jd
			tret[4] = tret[5]
			dontTimes = 1
		}
		if !(ifltype&2 != 0) && retflag&2 != 0 {
			t = jd + float64(direction*20)
			if oneTry != 0 {
				tret[0] = jd
				return 0
			}
			jd = t
			continue nextTry
		}
		if !(ifltype&1 != 0) && retflag&1 != 0 {
			t = jd + float64(direction*20)
			if oneTry != 0 {
				tret[0] = jd
				return 0
			}
			jd = t
			continue nextTry
		}
		if !(ifltype&8 != 0) && retflag&8 != 0 {
			t = jd + float64(direction*20)
			if oneTry != 0 {
				tret[0] = jd
				return 0
			}
			jd = t
			continue nextTry
		}
		if !(ifltype&16 != 0) && retflag&16 != 0 {
			t = jd + float64(direction*20)
			if oneTry != 0 {
				tret[0] = jd
				return 0
			}
			jd = t
			continue nextTry
		}
		if !(ifltype&(4|32) != 0) && retflag&4 != 0 {
			t = jd + float64(direction*20)
			if oneTry != 0 {
				tret[0] = jd
				return 0
			}
			jd = t
			continue nextTry
		}
		if dontTimes != 0 {
			return retflag
		}
		if retflag&16 != 0 {
			o = 0
		} else if retflag&2 != 0 {
			o = 1
		} else {
			o = 2
		}
		dta = twohr
		dtb = tenmin
		for n = 0; n <= o; n++ {
			if n == 0 {
				i1 = 2
				i2 = 3
			} else if n == 1 {
				if retflag&16 != 0 {
					continue
				}
				i1 = 4
				i2 = 5
			} else if n == 2 {
				if retflag&2 != 0 {
					continue
				}
				i1 = 6
				i2 = 7
			}
			i = 0
			for t = jd - dta; i <= 2; {
				if retflag2 = state.eclipseWhere(t, body, starname, ifl, location[:], dcore[:], message); retflag2 == -1 {
					return retflag2
				}
				if n == 0 {
					dc[i] = dcore[4]/2 + de/dcore[5] - dcore[2]
				} else if n == 1 {
					dc[i] = math.Abs(dcore[3])/2 + de/dcore[6] - dcore[2]
				} else if n == 2 {
					dc[i] = de/dcore[6] - dcore[2]
				}
				i += 1
				t += dta
			}
			state.findZero(dc[0], dc[1], dc[2], dta, dt1[:], dt2[:])
			tret[i1] = jd + dt1[0] + dta
			tret[i2] = jd + dt2[0] + dta
			m = 0
			for dt = dtb; m < 3; {
				for j = i1; j <= i2; j += i2 - i1 {
					i = 0
					for t = tret[j] - dt; i < 2; {
						if retflag2 = state.eclipseWhere(t, body, starname, ifl, location[:], dcore[:], message); retflag2 == -1 {
							return retflag2
						}
						if n == 0 {
							dc[i] = dcore[4]/2 + de/dcore[5] - dcore[2]
						} else if n == 1 {
							dc[i] = math.Abs(dcore[3])/2 + de/dcore[6] - dcore[2]
						} else if n == 2 {
							dc[i] = de/dcore[6] - dcore[2]
						}
						i++
						t += dt
					}
					dt1[0] = dc[1] / ((dc[1] - dc[0]) / dt)
					tret[j] -= dt1[0]
				}
				m++
				dt /= 3
			}
		}
		if retflag&4 != 0 {
			if retflag2 = state.eclipseWhere(tret[0], body, starname, ifl, location[:], dcore[:], message); retflag2 == -1 {
				return retflag2
			}
			dc[0] = dcore[0]
			if retflag2 = state.eclipseWhere(tret[4], body, starname, ifl, location[:], dcore[:], message); retflag2 == -1 {
				return retflag2
			}
			dc[1] = dcore[0]
			if retflag2 = state.eclipseWhere(tret[5], body, starname, ifl, location[:], dcore[:], message); retflag2 == -1 {
				return retflag2
			}
			dc[2] = dcore[0]
			if dc[0]*dc[1] < 0 || dc[0]*dc[2] < 0 {
				retflag |= 32
				retflag &= ^4
			}
		}
		if !(ifltype&4 != 0) && retflag&4 != 0 {
			t = jd + float64(direction*20)
			if oneTry != 0 {
				tret[0] = jd
				return 0
			}
			jd = t
			continue nextTry
		}
		if !(ifltype&32 != 0) && retflag&32 != 0 {
			t = jd + float64(direction*20)
			if oneTry != 0 {
				tret[0] = jd
				return 0
			}
			jd = t
			continue nextTry
		}
		k = 2
		for i = 0; i < 2; i++ {
			j = i + k
			tt = tret[j] + state.sweDeltatEx(tret[j], ifl, message)
			if state.calcPlanetStar(tt, body, starname, flags, ls[:], message) == -1 {
				return -1
			}
			if state.sweCalc(tt, 1, flags, lm[:], message) == -1 {
				return -1
			}
			dc[i] = state.sweDegnorm(ls[0] - lm[0])
			if dc[i] > 180 {
				dc[i] -= 360
			}
		}
		if dc[0]*dc[1] >= 0 {
			tret[1] = 0
		} else {
			jd = tjds
			dt = 0.1
			dt1[0] = (tret[3] - tret[2]) / 2.0
			if dt1[0] < dt {
				dt = dt1[0] / 2.0
			}
			for j = 0; dt > 0.01; {
				i = 0
				for t = jd; i <= 1; {
					tt = t + state.sweDeltatEx(t, ifl, message)
					if state.calcPlanetStar(tt, body, starname, flags, ls[:], message) == -1 {
						return -1
					}
					if state.sweCalc(tt, 1, flags, lm[:], message) == -1 {
						return -1
					}
					dc[i] = state.sweDegnorm(ls[0] - lm[0])
					if dc[i] > 180 {
						dc[i] -= 360
					}
					if dc[i] > 180 {
						dc[i] -= 360
					}
					i++
					t -= dt
				}
				a = (dc[1] - dc[0]) / dt
				if a < 1e-10 {
					break
				}
				dt1[0] = dc[0] / a
				jd += dt1[0]
				j++
				dt /= 3
			}
			tret[1] = jd
		}
		return retflag
	}
}

// sweSolEclipseWhenLoc implements swe_sol_eclipse_when_loc from swecl.c.
//
//	When is the next solar eclipse at a given geographical position?
//
// Note the uncertainty of Delta T for the remote past and for
// the future.
//
// retflag	SE_ECL_TOTAL or SE_ECL_ANNULAR or SE_ECL_PARTIAL
//
//	SE_ECL_VISIBLE,
//	SE_ECL_MAX_VISIBLE,
//	SE_ECL_1ST_VISIBLE, SE_ECL_2ND_VISIBLE
//	SE_ECL_3ST_VISIBLE, SE_ECL_4ND_VISIBLE
//
// tret[0]	time of maximum eclipse
// tret[1]	time of first contact
// tret[2]	time of second contact
// tret[3]	time of third contact
// tret[4]	time of forth contact
// tret[5]	time of sun rise between first and forth contact
// tret[6]	time of sun set beween first and forth contact
//
// attr[0]	fraction of solar diameter covered by moon;
//
//	with total/annular eclipses, it results in magnitude acc. to IMCCE.
//
// attr[1]	ratio of lunar diameter to solar one
// attr[2]	fraction of solar disc covered by moon (obscuration)
// attr[3]      diameter of core shadow in km
// attr[4]	azimuth of sun at tjd
// attr[5]	true altitude of sun above horizon at tjd
// attr[6]	apparent altitude of sun above horizon at tjd
// attr[7]	elongation of moon in degrees
// attr[8]	magnitude acc. to NASA;
//
//	= attr[0] for partial and attr[1] for annular and total eclipses
//
// attr[9]	saros series number
// attr[10]	saros series member number
//
//	declare as attr[20] at least !
func (state *State) sweSolEclipseWhenLoc(tjdStart float64, ifl int32, location []float64, tret []float64, attr []float64, backward int32, message []byte) int32 {
	var (
		retflag  int32 = 0
		retflag2 int32 = 0
		geopos2  [20]float64
		dcore    [10]float64
	)
	if location[2] < -500.0 || location[2] > 25000.0 {
		if message != nil {
			formatText(message, textBytes("location for eclipses must be between %.0f and %.0f m above sea"), -500.0, 25000.0)
		}
		return -1
	}
	ifl &= 1 | 2 | 4
	state.swiSetTidAcc(tjdStart, ifl, 0, message)
	if retflag = state.eclipseWhenLoc(tjdStart, ifl, location, tret, attr, backward, message); retflag <= 0 {
		return retflag
	}
	if retflag2 = state.eclipseWhere(tret[0], 0, nil, ifl, geopos2[:], dcore[:], message); retflag2 == -1 {
		return retflag2
	}
	retflag |= retflag2 & 2
	attr[3] = dcore[0]
	return retflag
}

// sweLunOccultWhenLoc implements swe_lun_occult_when_loc from swecl.c.
//
//	When is the next solar eclipse at a given geographical position?
//
// Note the uncertainty of Delta T for the remote past and for
// the future.
//
// retflag	SE_ECL_TOTAL or SE_ECL_ANNULAR or SE_ECL_PARTIAL
//
//	SE_ECL_VISIBLE,
//	SE_ECL_MAX_VISIBLE,
//	SE_ECL_1ST_VISIBLE, SE_ECL_2ND_VISIBLE
//	SE_ECL_3ST_VISIBLE, SE_ECL_4ND_VISIBLE
//	SE_ECL_OCC_BEG_DAYLIGHT, SE_ECL_OCC_END_DAYLIGHT
//
// The latter two indicate that the beginning or end of the occultation takes
// place during the day. If Venus is occulted, it may be observable with the
// naked eye; if other objects, it may be observable with telescopes.
//
// int32 ipl          planet number of occulted body
// char* starname     name of occulted star. Must be NULL or "", if a planetary
//
//	occultation is to be calculated. For the use of this
//	field, also see swe_fixstar().
//
// int32 ifl	      ephemeris flag. If you want to have only one conjunction
//
//	of the moon with the body tested, add the following flag:
//	backward |= SE_ECL_ONE_TRY. If this flag is not set,
//	the function will search for an occultation until it
//	finds one. For bodies with ecliptical latitudes > 5,
//	the function may search unsuccessfully until it reaches
//	the end of the ephemeris.
//
// for all other parameters, see function swe_sol_eclipse_when_loc().
func (state *State) sweLunOccultWhenLoc(tjdStart float64, body int32, starname []byte, ifl int32, location []float64, tret []float64, attr []float64, backward int32, message []byte) int32 {
	var (
		retflag  int32 = 0
		retflag2 int32 = 0
		geopos2  [20]float64
		dcore    [10]float64
	)
	if location[2] < -500.0 || location[2] > 25000.0 {
		if message != nil {
			formatText(message, textBytes("location for occultations must be between %.0f and %.0f m above sea"), -500.0, 25000.0)
		}
		return -1
	}
	if body < 0 {
		body = 0
	}
	if body == 10000+134340 {
		body = 9
	}
	ifl &= 1 | 2 | 4
	state.swiSetTidAcc(tjdStart, ifl, 0, message)
	if retflag = state.occultWhenLoc(tjdStart, body, starname, ifl, location, tret, attr, backward, message); retflag <= 0 {
		return retflag
	}
	if retflag2 = state.eclipseWhere(tret[0], body, starname, ifl, geopos2[:], dcore[:], message); retflag2 == -1 {
		return retflag2
	}
	retflag |= retflag2 & 2
	attr[3] = dcore[0]
	return retflag
}

// eclipseWhenLoc implements eclipse_when_loc from swecl.c.
func (state *State) eclipseWhenLoc(tjdStart float64, ifl int32, location []float64, tret []float64, attr []float64, backward int32, message []byte) int32 {
	var (
		i         int32
		j         int32
		k         int32
		m         int32
		retflag   int32 = 0
		status    int32
		t         float64
		jd        float64
		dt        float64
		dtint     [1]float64
		K         float64
		T         float64
		T2        float64
		T3        float64
		T4        float64
		F         float64
		M         float64
		Mm        float64
		tjdr      [1]float64
		tjds      [1]float64
		E         float64
		Ff        float64
		A1        float64
		Om        float64
		xs        [6]float64
		xm        [6]float64
		ls        [6]float64
		lm        [6]float64
		x1        [6]float64
		x2        [6]float64
		dm        float64
		ds        float64
		rmoon     float64
		rsun      float64
		rsplusrm  float64
		rsminusrm float64
		dc        [3]float64
		dctr      [1]float64
		dctrmin   float64
		twomin    float64    = 0.0013888888888888887
		tensec    float64    = 0.00011574074074074075
		twohr     float64    = 0.083333333333333329
		tenmin    float64    = 0.0069444444444444449
		dt1       [1]float64 = [1]float64{0}
		dt2       [1]float64 = [1]float64{0}
		dtdiv     float64
		dtstart   float64
		flags     int32 = 2*1024 | 32*1024 | ifl
		iflagcart int32 = flags | (flagCartesian)
	)
	state.sweSetTopo(location[0], location[1], location[2])
	K = float64(int32((tjdStart - julianDayJ2000) / 365.2425 * 12.3685))
	if backward != 0 {
		K++
	} else {
		K--
	}
nextTry:
	for {
		T = K / 1236.85
		T2 = T * T
		T3 = T2 * T
		T4 = T3 * T
		F = state.sweDegnorm(160.7108 + 390.67050274*K - 0.0016341*T2 - 0.00000227*T3 + 0.000000011*T4)
		Ff = F
		if Ff > 180 {
			Ff -= 180
		}
		if Ff > 21 && Ff < 159 {
			if backward != 0 {
				K--
			} else {
				K++
			}
			continue nextTry
		}
		jd = 2451550.09765 + 29.530588853*K + 0.0001337*T2 - 0.000000150*T3 + 0.00000000073*T4
		M = state.sweDegnorm(2.5534 + 29.10535669*K - 0.0000218*T2 - 0.00000011*T3)
		Mm = state.sweDegnorm(201.5643 + 385.81693528*K + 0.1017438*T2 + 0.00001239*T3 + 0.000000058*T4)
		Om = state.sweDegnorm(124.7746 - 1.56375580*K + 0.0020691*T2 + 0.00000215*T3)
		E = 1 - 0.002516*T - 0.0000074*T2
		A1 = state.sweDegnorm(299.77 + 0.107408*K - 0.009173*T2)
		M *= radiansPerDegree
		Mm *= radiansPerDegree
		F *= radiansPerDegree
		Om *= radiansPerDegree
		A1 *= radiansPerDegree
		jd = jd - 0.4075*math.Sin(Mm) + 0.1721*E*math.Sin(M)
		state.sweSetTopo(location[0], location[1], location[2])
		dtdiv = 2
		dtstart = 0.5
		if jd < 1900000 || jd > 2500000 {
			dtstart = 2
		}
		for dt = dtstart; dt > 0.00001; dt /= dtdiv {
			if dt < 0.1 {
				dtdiv = 3
			}
			i = 0
			for t = jd - dt; i <= 2; {
				if state.sweCalc(t, 0, iflagcart, xs[:], message) == -1 {
					return -1
				}
				if state.sweCalc(t, 0, flags, ls[:], message) == -1 {
					return -1
				}
				if state.sweCalc(t, 1, iflagcart, xm[:], message) == -1 {
					return -1
				}
				if state.sweCalc(t, 1, flags, lm[:], message) == -1 {
					return -1
				}
				dm = math.Sqrt(xm[0]*xm[0] + xm[1]*xm[1] + xm[2]*xm[2])
				ds = math.Sqrt(xs[0]*xs[0] + xs[1]*xs[1] + xs[2]*xs[2])
				for k = 0; k < 3; k++ {
					x1[k] = xs[k] / ds
					x2[k] = xm[k] / dm
				}
				dc[i] = math.Acos(state.swiDotProdUnit(x1[:], x2[:])) * degreesPerRadian
				i++
				t += dt
			}
			state.findMaximum(dc[0], dc[1], dc[2], dt, dtint[:], dctr[:])
			jd += dtint[0] + dt
		}
		if state.sweCalc(jd, 0, iflagcart, xs[:], message) == -1 {
			return -1
		}
		if state.sweCalc(jd, 0, flags, ls[:], message) == -1 {
			return -1
		}
		if state.sweCalc(jd, 1, iflagcart, xm[:], message) == -1 {
			return -1
		}
		if state.sweCalc(jd, 1, flags, lm[:], message) == -1 {
			return -1
		}
		dctr[0] = math.Acos(state.swiDotProdUnit(xs[:], xm[:])) * degreesPerRadian
		rmoon = math.Asin(1.1618815106570899e-05/lm[2]) * degreesPerRadian
		rsun = math.Asin(0.0046524726370988385/ls[2]) * degreesPerRadian
		rsplusrm = rsun + rmoon
		rsminusrm = rsun - rmoon
		if dctr[0] > rsplusrm {
			if backward != 0 {
				K--
			} else {
				K++
			}
			continue nextTry
		}
		tret[0] = jd - state.sweDeltatEx(jd, ifl, message)
		tret[0] = jd - state.sweDeltatEx(tret[0], ifl, message)
		if backward != 0 && tret[0] >= tjdStart-0.0001 || !(backward != 0) && tret[0] <= tjdStart+0.0001 {
			if backward != 0 {
				K--
			} else {
				K++
			}
			continue nextTry
		}
		if dctr[0] < rsminusrm {
			retflag = 8
		} else if dctr[0] < math.Abs(rsminusrm) {
			retflag = 4
		} else if dctr[0] <= rsplusrm {
			retflag = 16
		}
		dctrmin = dctr[0]
		if dctr[0] > math.Abs(rsminusrm) {
			tret[3] = 0
			tret[2] = tret[3]
		} else {
			dc[1] = math.Abs(rsminusrm) - dctrmin
			i = 0
			for t = jd - twomin; i <= 2; {
				if state.sweCalc(t, 0, iflagcart, xs[:], message) == -1 {
					return -1
				}
				if state.sweCalc(t, 1, iflagcart, xm[:], message) == -1 {
					return -1
				}
				dm = math.Sqrt(xm[0]*xm[0] + xm[1]*xm[1] + xm[2]*xm[2])
				ds = math.Sqrt(xs[0]*xs[0] + xs[1]*xs[1] + xs[2]*xs[2])
				rmoon = math.Asin(1.1618815106570899e-05/dm) * degreesPerRadian
				rmoon *= 0.99916
				rsun = math.Asin(0.0046524726370988385/ds) * degreesPerRadian
				rsminusrm = rsun - rmoon
				for k = 0; k < 3; k++ {
					x1[k] = xs[k] / ds
					x2[k] = xm[k] / dm
				}
				dctr[0] = math.Acos(state.swiDotProdUnit(x1[:], x2[:])) * degreesPerRadian
				dc[i] = math.Abs(rsminusrm) - dctr[0]
				i += 2
				t = jd + twomin
			}
			state.findZero(dc[0], dc[1], dc[2], twomin, dt1[:], dt2[:])
			tret[2] = jd + dt1[0] + twomin
			tret[3] = jd + dt2[0] + twomin
			m = 0
			for dt = tensec; m < 2; {
				for j = 2; j <= 3; j++ {
					if state.sweCalc(tret[j], 0, iflagcart|256, xs[:], message) == -1 {
						return -1
					}
					if state.sweCalc(tret[j], 1, iflagcart|256, xm[:], message) == -1 {
						return -1
					}
					for i = 0; i < 2; i++ {
						if i == 1 {
							for k = 0; k < 3; k++ {
								xs[k] -= xs[k+3] * dt
								xm[k] -= xm[k+3] * dt
							}
						}
						dm = math.Sqrt(xm[0]*xm[0] + xm[1]*xm[1] + xm[2]*xm[2])
						ds = math.Sqrt(xs[0]*xs[0] + xs[1]*xs[1] + xs[2]*xs[2])
						rmoon = math.Asin(1.1618815106570899e-05/dm) * degreesPerRadian
						rmoon *= 0.99916
						rsun = math.Asin(0.0046524726370988385/ds) * degreesPerRadian
						rsminusrm = rsun - rmoon
						for k = 0; k < 3; k++ {
							x1[k] = xs[k] / ds
							x2[k] = xm[k] / dm
						}
						dctr[0] = math.Acos(state.swiDotProdUnit(x1[:], x2[:])) * degreesPerRadian
						dc[i] = math.Abs(rsminusrm) - dctr[0]
					}
					dt1[0] = -dc[0] / ((dc[0] - dc[1]) / dt)
					tret[j] += dt1[0]
				}
				m++
				dt /= 10
			}
			tret[2] -= state.sweDeltatEx(tret[2], ifl, message)
			tret[3] -= state.sweDeltatEx(tret[3], ifl, message)
		}
		dc[1] = rsplusrm - dctrmin
		i = 0
		for t = jd - twohr; i <= 2; {
			if state.sweCalc(t, 0, iflagcart, xs[:], message) == -1 {
				return -1
			}
			if state.sweCalc(t, 1, iflagcart, xm[:], message) == -1 {
				return -1
			}
			dm = math.Sqrt(xm[0]*xm[0] + xm[1]*xm[1] + xm[2]*xm[2])
			ds = math.Sqrt(xs[0]*xs[0] + xs[1]*xs[1] + xs[2]*xs[2])
			rmoon = math.Asin(1.1618815106570899e-05/dm) * degreesPerRadian
			rsun = math.Asin(0.0046524726370988385/ds) * degreesPerRadian
			rsplusrm = rsun + rmoon
			for k = 0; k < 3; k++ {
				x1[k] = xs[k] / ds
				x2[k] = xm[k] / dm
			}
			dctr[0] = math.Acos(state.swiDotProdUnit(x1[:], x2[:])) * degreesPerRadian
			dc[i] = rsplusrm - dctr[0]
			i += 2
			t = jd + twohr
		}
		state.findZero(dc[0], dc[1], dc[2], twohr, dt1[:], dt2[:])
		tret[1] = jd + dt1[0] + twohr
		tret[4] = jd + dt2[0] + twohr
		m = 0
		for dt = tenmin; m < 3; {
			for j = 1; j <= 4; j += 3 {
				if state.sweCalc(tret[j], 0, iflagcart|256, xs[:], message) == -1 {
					return -1
				}
				if state.sweCalc(tret[j], 1, iflagcart|256, xm[:], message) == -1 {
					return -1
				}
				for i = 0; i < 2; i++ {
					if i == 1 {
						for k = 0; k < 3; k++ {
							xs[k] -= xs[k+3] * dt
							xm[k] -= xm[k+3] * dt
						}
					}
					dm = math.Sqrt(xm[0]*xm[0] + xm[1]*xm[1] + xm[2]*xm[2])
					ds = math.Sqrt(xs[0]*xs[0] + xs[1]*xs[1] + xs[2]*xs[2])
					rmoon = math.Asin(1.1618815106570899e-05/dm) * degreesPerRadian
					rsun = math.Asin(0.0046524726370988385/ds) * degreesPerRadian
					rsplusrm = rsun + rmoon
					for k = 0; k < 3; k++ {
						x1[k] = xs[k] / ds
						x2[k] = xm[k] / dm
					}
					dctr[0] = math.Acos(state.swiDotProdUnit(x1[:], x2[:])) * degreesPerRadian
					dc[i] = math.Abs(rsplusrm) - dctr[0]
				}
				dt1[0] = -dc[0] / ((dc[0] - dc[1]) / dt)
				tret[j] += dt1[0]
			}
			m++
			dt /= 10
		}
		tret[1] -= state.sweDeltatEx(tret[1], ifl, message)
		tret[4] -= state.sweDeltatEx(tret[4], ifl, message)
		for i = 4; i >= 0; i-- {
			if tret[i] == 0 {
				continue
			}
			if state.eclipseHow(tret[i], 0, nil, ifl, location[0], location[1], location[2], attr, message) == -1 {
				return -1
			}
			if attr[6] > 0 {
				retflag |= 128
				switch i {
				case 0:
					retflag |= 256
				case 1:
					retflag |= 512
				case 2:
					retflag |= 1024
				case 3:
					retflag |= 2048
				case 4:
					retflag |= 4096
				default:
				}
			}
		}
		if !(retflag&128 != 0) {
			if backward != 0 {
				K--
			} else {
				K++
			}
			continue nextTry
		}
		if status = state.sweRiseTrans(tret[1]-0.001, 0, nil, flags, 1|8192, location, 0, 0, tjdr[:], message); status == -1 {
			return -1
		}
		if status == -2 {
			return retflag
		}
		if status = state.sweRiseTrans(tret[1]-0.001, 0, nil, flags, 2|8192, location, 0, 0, tjds[:], message); status == -1 {
			return -1
		}
		if status == -2 {
			return retflag
		}
		if tjds[0] < tret[1] || tjds[0] > tjdr[0] && tjdr[0] > tret[4] {
			if backward != 0 {
				K--
			} else {
				K++
			}
			continue nextTry
		}
		if tjdr[0] > tret[1] && tjdr[0] < tret[4] {
			tret[5] = tjdr[0]
			if !(retflag&256 != 0) {
				tret[0] = tjdr[0]
				if status = state.eclipseHow(tret[5], 0, nil, ifl, location[0], location[1], location[2], attr, message); status == -1 {
					return -1
				}
				retflag &= ^(4 | 8 | 16)
				retflag |= status & (4 | 8 | 16)
			}
		}
		if tjds[0] > tret[1] && tjds[0] < tret[4] {
			tret[6] = tjds[0]
			if !(retflag&256 != 0) {
				tret[0] = tjds[0]
				if status = state.eclipseHow(tret[6], 0, nil, ifl, location[0], location[1], location[2], attr, message); status == -1 {
					return -1
				}
				retflag &= ^(4 | 8 | 16)
				retflag |= status & (4 | 8 | 16)
			}
		}
		return retflag
	}
}

// occultWhenLoc implements occult_when_loc from swecl.c.
func (state *State) occultWhenLoc(tjdStart float64, body int32, starname []byte, ifl int32, location []float64, tret []float64, attr []float64, backward int32, message []byte) int32 {
	var (
		i             int32
		j             int32
		k             int32
		m             int32
		retflag       int32 = 0
		status        int32
		t             float64
		jd            float64
		dt            float64
		dtint         [1]float64
		tjdr          [1]float64
		tjds          [1]float64
		xs            [6]float64
		xm            [6]float64
		ls            [6]float64
		lm            [6]float64
		x1            [6]float64
		x2            [6]float64
		dm            float64
		ds            float64
		rmoon         float64
		rsun          float64
		rsplusrm      float64
		rsminusrm     float64
		dc            [20]float64
		dctr          [1]float64
		dctrmin       float64
		twomin        float64    = 0.0013888888888888887
		tensec        float64    = 0.00011574074074074075
		twohr         float64    = 0.083333333333333329
		tenmin        float64    = 0.0069444444444444449
		dt1           [1]float64 = [1]float64{0}
		dt2           [1]float64 = [1]float64{0}
		dtdiv         float64
		dtstart       float64
		dadd2         float64 = 1
		drad          float64
		dl            float64
		flags         int32 = 32*1024 | ifl
		iflaggeo      int32 = flags & ^(32 * 1024)
		iflagcart     int32 = flags | (flagCartesian)
		direction     int32 = 1
		oneTry        int32 = backward & (32 * 1024)
		stopAfterThis int32 = 0
	)
	backward &= 1
	retflag = 0
	state.sweSetTopo(location[0], location[1], location[2])
	for i = 0; i <= 9; i++ {
		tret[i] = 0
	}
	if backward != 0 {
		direction = -1
	}
	t = tjdStart
	jd = tjdStart
nextTry:
	for {
		if state.calcPlanetStar(t, body, starname, iflaggeo, ls[:], message) == -1 {
			return -1
		}
		if math.Abs(ls[1]) > 7 && starname != nil && int32(starname[0]) != '\000' {
			if message != nil {
				formatText(message, textBytes("occultation never occurs: star %s has ecl. lat. %.1f"), starname, ls[1])
			}
			return -1
		}
		if state.sweCalc(t, 1, iflaggeo, lm[:], message) == -1 {
			return -1
		}
		dl = state.sweDegnorm(ls[0] - lm[0])
		if direction < 0 {
			dl -= 360
		}
		for math.Abs(dl) > 0.1 {
			t += dl / 13
			if state.calcPlanetStar(t, body, starname, iflaggeo, ls[:], message) == -1 {
				return -1
			}
			if state.sweCalc(t, 1, iflaggeo, lm[:], message) == -1 {
				return -1
			}
			dl = state.sweDegnorm(ls[0] - lm[0])
			if dl > 180 {
				dl -= 360
			}
		}
		jd = t
		drad = math.Abs(ls[1] - lm[1])
		if drad > 2 {
			if oneTry != 0 {
				tret[0] = t + float64(direction)
				return 0
			}
			t += float64(direction * 20)
			jd = t
			continue nextTry
		}
		if starname != nil && int32(starname[0]) != '\000' {
			drad = 0
		} else if body < 20+1 {
			drad = calendarPlaDiam[body] / 2 / 1.49597870700e+11
		} else if body > 10000 {
			drad = state.positions.data.astDiam / 2 * 1000 / 1.49597870700e+11
		} else {
			drad = 0
		}
		dtdiv = 2
		dtstart = dadd2
		for dt = dtstart; dt > 0.00001; dt /= dtdiv {
			if dt < 0.01 {
				dtdiv = 2
			}
			i = 0
			for t = jd - dt; i <= 2; {
				if state.calcPlanetStar(t, body, starname, iflagcart, xs[:], message) == -1 {
					return -1
				}
				if state.calcPlanetStar(t, body, starname, flags, ls[:], message) == -1 {
					return -1
				}
				if state.sweCalc(t, 1, iflagcart, xm[:], message) == -1 {
					return -1
				}
				if state.sweCalc(t, 1, flags, lm[:], message) == -1 {
					return -1
				}
				if dt < 0.1 && math.Abs(ls[1]-lm[1]) > 2 {
					if oneTry != 0 || stopAfterThis != 0 {
						stopAfterThis = 1
					} else {
						t = jd + float64(direction*20)
						jd = t
						continue nextTry
					}
				}
				dc[i] = math.Acos(state.swiDotProdUnit(xs[:], xm[:])) * degreesPerRadian
				rmoon = math.Asin(1.1618815106570899e-05/lm[2]) * degreesPerRadian
				rsun = math.Asin(drad/ls[2]) * degreesPerRadian
				dc[i] -= rmoon + rsun
				i++
				t += dt
			}
			state.findMaximum(dc[0], dc[1], dc[2], dt, dtint[:], dctr[:])
			jd += dtint[0] + dt
		}
		if stopAfterThis != 0 {
			tret[0] = jd + float64(direction)
			return 0
		}
		if state.calcPlanetStar(jd, body, starname, iflagcart, xs[:], message) == -1 {
			return -1
		}
		if state.calcPlanetStar(jd, body, starname, flags, ls[:], message) == -1 {
			return -1
		}
		if state.sweCalc(jd, 1, iflagcart, xm[:], message) == -1 {
			return -1
		}
		if state.sweCalc(jd, 1, flags, lm[:], message) == -1 {
			return -1
		}
		dctr[0] = math.Acos(state.swiDotProdUnit(xs[:], xm[:])) * degreesPerRadian
		rmoon = math.Asin(1.1618815106570899e-05/lm[2]) * degreesPerRadian
		rsun = math.Asin(drad/ls[2]) * degreesPerRadian
		rsplusrm = rsun + rmoon
		rsminusrm = rsun - rmoon
		if dctr[0] > rsplusrm {
			if oneTry != 0 {
				tret[0] = jd + float64(direction)
				return 0
			}
			t = jd + float64(direction*20)
			jd = t
			continue nextTry
		}
		tret[0] = jd - state.sweDeltatEx(jd, ifl, message)
		tret[0] = jd - state.sweDeltatEx(tret[0], ifl, message)
		if backward != 0 && tret[0] >= tjdStart-0.0001 || !(backward != 0) && tret[0] <= tjdStart+0.0001 {
			if oneTry != 0 {
				tret[0] = jd + float64(direction)
				return 0
			}
			t = jd + float64(direction*20)
			jd = t
			continue nextTry
		}
		if dctr[0] < rsminusrm {
			retflag = 8
		} else if dctr[0] < math.Abs(rsminusrm) {
			retflag = 4
		} else if dctr[0] <= rsplusrm {
			retflag = 16
		}
		dctrmin = dctr[0]
		if dctr[0] > math.Abs(rsminusrm) {
			tret[3] = 0
			tret[2] = tret[3]
		} else {
			dc[1] = math.Abs(rsminusrm) - dctrmin
			i = 0
			for t = jd - twomin; i <= 2; {
				if state.calcPlanetStar(t, body, starname, iflagcart, xs[:], message) == -1 {
					return -1
				}
				if state.sweCalc(t, 1, iflagcart, xm[:], message) == -1 {
					return -1
				}
				dm = math.Sqrt(xm[0]*xm[0] + xm[1]*xm[1] + xm[2]*xm[2])
				ds = math.Sqrt(xs[0]*xs[0] + xs[1]*xs[1] + xs[2]*xs[2])
				rmoon = math.Asin(1.1618815106570899e-05/dm) * degreesPerRadian
				rmoon *= 0.99916
				rsun = math.Asin(drad/ds) * degreesPerRadian
				rsminusrm = rsun - rmoon
				for k = 0; k < 3; k++ {
					x1[k] = xs[k] / ds
					x2[k] = xm[k] / dm
				}
				dctr[0] = math.Acos(state.swiDotProdUnit(x1[:], x2[:])) * degreesPerRadian
				dc[i] = math.Abs(rsminusrm) - dctr[0]
				i += 2
				t = jd + twomin
			}
			state.findZero(dc[0], dc[1], dc[2], twomin, dt1[:], dt2[:])
			tret[2] = jd + dt1[0] + twomin
			tret[3] = jd + dt2[0] + twomin
			m = 0
			for dt = tensec; m < 2; {
				for j = 2; j <= 3; j++ {
					if state.calcPlanetStar(tret[j], body, starname, iflagcart|256, xs[:], message) == -1 {
						return -1
					}
					if state.sweCalc(tret[j], 1, iflagcart|256, xm[:], message) == -1 {
						return -1
					}
					for i = 0; i < 2; i++ {
						if i == 1 {
							for k = 0; k < 3; k++ {
								xs[k] -= xs[k+3] * dt
								xm[k] -= xm[k+3] * dt
							}
						}
						dm = math.Sqrt(xm[0]*xm[0] + xm[1]*xm[1] + xm[2]*xm[2])
						ds = math.Sqrt(xs[0]*xs[0] + xs[1]*xs[1] + xs[2]*xs[2])
						rmoon = math.Asin(1.1618815106570899e-05/dm) * degreesPerRadian
						rmoon *= 0.99916
						rsun = math.Asin(drad/ds) * degreesPerRadian
						rsminusrm = rsun - rmoon
						for k = 0; k < 3; k++ {
							x1[k] = xs[k] / ds
							x2[k] = xm[k] / dm
						}
						dctr[0] = math.Acos(state.swiDotProdUnit(x1[:], x2[:])) * degreesPerRadian
						dc[i] = math.Abs(rsminusrm) - dctr[0]
					}
					dt1[0] = -dc[0] / ((dc[0] - dc[1]) / dt)
					tret[j] += dt1[0]
				}
				m++
				dt /= 10
			}
			tret[2] -= state.sweDeltatEx(tret[2], ifl, message)
			tret[3] -= state.sweDeltatEx(tret[3], ifl, message)
		}
		dc[1] = rsplusrm - dctrmin
		if starname == nil || int32(starname[0]) == '\000' {
			i = 0
			for t = jd - twohr; i <= 2; {
				if state.calcPlanetStar(t, body, starname, iflagcart, xs[:], message) == -1 {
					return -1
				}
				if state.sweCalc(t, 1, iflagcart, xm[:], message) == -1 {
					return -1
				}
				dm = math.Sqrt(xm[0]*xm[0] + xm[1]*xm[1] + xm[2]*xm[2])
				ds = math.Sqrt(xs[0]*xs[0] + xs[1]*xs[1] + xs[2]*xs[2])
				rmoon = math.Asin(1.1618815106570899e-05/dm) * degreesPerRadian
				rsun = math.Asin(drad/ds) * degreesPerRadian
				rsplusrm = rsun + rmoon
				for k = 0; k < 3; k++ {
					x1[k] = xs[k] / ds
					x2[k] = xm[k] / dm
				}
				dctr[0] = math.Acos(state.swiDotProdUnit(x1[:], x2[:])) * degreesPerRadian
				dc[i] = rsplusrm - dctr[0]
				i += 2
				t = jd + twohr
			}
			state.findZero(dc[0], dc[1], dc[2], twohr, dt1[:], dt2[:])
			tret[1] = jd + dt1[0] + twohr
			tret[4] = jd + dt2[0] + twohr
			m = 0
			for dt = tenmin; m < 3; {
				for j = 1; j <= 4; j += 3 {
					if state.calcPlanetStar(tret[j], body, starname, iflagcart|256, xs[:], message) == -1 {
						return -1
					}
					if state.sweCalc(tret[j], 1, iflagcart|256, xm[:], message) == -1 {
						return -1
					}
					for i = 0; i < 2; i++ {
						if i == 1 {
							for k = 0; k < 3; k++ {
								xs[k] -= xs[k+3] * dt
								xm[k] -= xm[k+3] * dt
							}
						}
						dm = math.Sqrt(xm[0]*xm[0] + xm[1]*xm[1] + xm[2]*xm[2])
						ds = math.Sqrt(xs[0]*xs[0] + xs[1]*xs[1] + xs[2]*xs[2])
						rmoon = math.Asin(1.1618815106570899e-05/dm) * degreesPerRadian
						rsun = math.Asin(drad/ds) * degreesPerRadian
						rsplusrm = rsun + rmoon
						for k = 0; k < 3; k++ {
							x1[k] = xs[k] / ds
							x2[k] = xm[k] / dm
						}
						dctr[0] = math.Acos(state.swiDotProdUnit(x1[:], x2[:])) * degreesPerRadian
						dc[i] = math.Abs(rsplusrm) - dctr[0]
					}
					dt1[0] = -dc[0] / ((dc[0] - dc[1]) / dt)
					tret[j] += dt1[0]
				}
				m++
				dt /= 10
			}
			tret[1] -= state.sweDeltatEx(tret[1], ifl, message)
			tret[4] -= state.sweDeltatEx(tret[4], ifl, message)
		} else {
			tret[1] = tret[2]
			tret[4] = tret[3]
		}
		for i = 4; i >= 0; i-- {
			if tret[i] == 0 {
				continue
			}
			if state.eclipseHow(tret[i], body, starname, ifl, location[0], location[1], location[2], attr, message) == -1 {
				return -1
			}
			if attr[6] > 0 {
				retflag |= 128
				switch i {
				case 0:
					retflag |= 256
				case 1:
					retflag |= 512
				case 2:
					retflag |= 1024
				case 3:
					retflag |= 2048
				case 4:
					retflag |= 4096
				default:
				}
			}
		}
		if !(retflag&128 != 0) {
			if oneTry != 0 {
				tret[0] = jd + float64(direction)
				return 0
			}
			t = jd + float64(direction*20)
			jd = t
			continue nextTry
		}
		if status = state.sweRiseTrans(tret[1]-0.1, body, starname, flags, 1|8192, location, 0, 0, tjdr[:], message); status == -1 {
			return -1
		}
		if status >= 0 {
			status = state.sweRiseTrans(tret[1]-0.1, body, starname, flags, 2|8192, location, 0, 0, tjds[:], message)
			if status == -1 {
				return -1
			}
		}
		if status >= 0 {
			if tjdr[0] > tret[1] && tjdr[0] < tret[4] {
				tret[5] = tjdr[0]
			}
			if tjds[0] > tret[1] && tjds[0] < tret[4] {
				tret[6] = tjds[0]
			}
		}
		if status = state.sweRiseTrans(tret[1], 0, nil, flags, 1, location, 0, 0, tjdr[:], message); status == -1 {
			return -1
		}
		if status >= 0 {
			status = state.sweRiseTrans(tret[1], 0, nil, flags, 2, location, 0, 0, tjds[:], message)
			if status == -1 {
				return -1
			}
		}
		if status >= 0 {
			if tjds[0] < tjdr[0] {
				retflag |= 8192
			}
		}
		if status = state.sweRiseTrans(tret[4], 0, nil, flags, 1, location, 0, 0, tjdr[:], message); status == -1 {
			return -1
		}
		if status >= 0 {
			status = state.sweRiseTrans(tret[4], 0, nil, flags, 2, location, 0, 0, tjds[:], message)
			if status == -1 {
				return -1
			}
		}
		if status >= 0 {
			if tjds[0] < tjdr[0] {
				retflag |= 16384
			}
		}
		return retflag
	}
}

func (state *State) sweAzalt(jdUT float64, calcFlag int32, location []float64, pressure float64, temperature float64, xin []float64, xaz []float64) {
	var (
		i       int32
		x       [6]float64
		xra     [3]float64
		armc    float64 = state.sweDegnorm(state.sweSidtime(jdUT)*15 + location[0])
		mdd     float64
		epsTrue float64
	)
	for i = 0; i < 2; i++ {
		xra[i] = xin[i]
	}
	xra[2] = 1
	if calcFlag == 0 {
		state.sweCalc(jdUT+state.sweDeltatEx(jdUT, -1, nil), -1, 0, x[:], nil)
		epsTrue = x[0]
		state.sweCotrans(xra[:], xra[:], -epsTrue)
	}
	mdd = state.sweDegnorm(xra[0] - armc)
	x[0] = state.sweDegnorm(mdd - 90)
	x[1] = xra[1]
	x[2] = 1
	state.sweCotrans(x[:], x[:], 90-location[1])
	x[0] = state.sweDegnorm(x[0] + 90)
	xaz[0] = 360 - x[0]
	xaz[1] = x[1]
	if pressure == 0 {
		pressure = 1013.25 * math.Pow(1-0.0065*location[2]/288, 5.255)
	}
	xaz[2] = state.sweRefracExtended(x[1], location[2], pressure, temperature, state.eclipse.lapseRate, 0, nil)
}

// sweAzaltRev implements swe_azalt_rev from swecl.c.
//
//	xaz[2] = swe_refrac_extended(xaz[2], geopos[2], atpress, attemp, const_lapse_rate, SE_APP_TO_TRUE, NULL);
func (state *State) sweAzaltRev(jdUT float64, calcFlag int32, location []float64, xin []float64, xout []float64) {
	var (
		i         int32
		x         [6]float64
		xaz       [3]float64
		longitude float64 = location[0]
		latitude  float64 = location[1]
		armc      float64 = state.sweDegnorm(state.sweSidtime(jdUT)*15 + longitude)
		epsTrue   float64
		dang      float64
	)
	for i = 0; i < 2; i++ {
		xaz[i] = xin[i]
	}
	xaz[2] = 1
	xaz[0] = 360 - xaz[0]
	xaz[0] = state.sweDegnorm(xaz[0] - 90)
	dang = latitude - 90
	state.sweCotrans(xaz[:], xaz[:], dang)
	xaz[0] = state.sweDegnorm(xaz[0] + armc + 90)
	xout[0] = xaz[0]
	xout[1] = xaz[1]
	if calcFlag == 0 {
		state.sweCalc(jdUT+state.sweDeltatEx(jdUT, -1, nil), -1, 0, x[:], nil)
		epsTrue = x[0]
		state.sweCotrans(xaz[:], x[:], epsTrue)
		xout[0] = x[0]
		xout[1] = x[1]
	}
}

// sweRefrac implements swe_refrac from swecl.c.
func (state *State) sweRefrac(inalt float64, pressure float64, temperature float64, calcFlag int32) float64 {
	var (
		a        float64
		refr     float64
		ptFactor float64 = pressure / 1010.0 * 283.0 / (273.0 + temperature)
		trualt   float64
		appalt   float64
	)
	if calcFlag == 0 {
		trualt = inalt
		if trualt > 15 {
			a = math.Tan((90 - trualt) * radiansPerDegree)
			refr = 58.276*a - 0.0824*a*a*a
			refr *= ptFactor / 3600.0
		} else if trualt > float64(-5) {
			a = trualt + 10.3/(trualt+5.11)
			if a+1e-10 >= 90 {
				refr = 0
			} else {
				refr = 1.02 / math.Tan(a*radiansPerDegree)
			}
			refr *= ptFactor / 60.0
		} else {
			refr = 0
		}
		appalt = trualt
		if appalt+refr > 0 {
			appalt += refr
		}
		return appalt
	} else {
		appalt = inalt
		a = appalt + 7.31/(appalt+4.4)
		if a+1e-10 >= 90 {
			refr = 0
		} else {
			refr = 1.00 / math.Tan(a*radiansPerDegree)
			refr -= 0.06 * math.Sin(14.7*refr+13)
		}
		refr *= ptFactor / 60.0
		trualt = appalt
		if appalt-refr > 0 {
			trualt = appalt - refr
		}
		return trualt
	}
}

// sweSetLapseRate implements swe_set_lapse_rate from swecl.c.
func (state *State) sweSetLapseRate(lapseRate float64) {
	state.eclipse.lapseRate = lapseRate
}

// sweRefracExtended implements swe_refrac_extended from swecl.c.
//
//	swe_refrac_extended()
//
// This function was created thanks to and with the help of the
// archaeoastronomer Victor Reijs.
// It is more correct and more skilled than the old function swe_refrac():
//   - it allows correct calculation of refraction for altitudes above sea > 0,
//     where the ideal horizon and planets that are visible may have a
//     negative height. (for swe_refrac(), negative apparent heights do not
//     exist!)
//   - it allows to manipulate the refraction constant
//
// Transforms apparent to true altitude and vice-versa.
// input:
// double inalt;        * altitude of object above geometric horizon in degrees*
//   - geometric horizon = plane perpendicular to gravity *
//
// double geoalt;       * altitude of observer above sea level in meters *
// double atpress;      * millibars (hectopascal) *
// double lapse_rate;    * (dT/dh) [deg K/m]
// double attemp;       * degrees C *
// int32  calc_flag;    * either SE_APP_TO_TRUE or
//   - SE_TRUE_TO_APP
//
// function returns:
// double *dret;        * array of 4 doubles; declare 20 doubles !
// - dret[0] true altitude, if possible; otherwise input value
// - dret[1] apparent altitude, if possible; otherwise input value
// - dret[2] refraction
// - dret[3] dip of the horizon
//
// The body is above the horizon if the dret[0] != dret[1]
//
// case 1, conversion from true altitude to apparent altitude
//   - apparent altitude, if body is observable above ideal horizon
//   - true altitude (the input value), otherwise
//     "ideal horizon" is the horizon as seen above an ideal sphere (as seen
//     from a plane over the ocean with a clear sky)
//
// case 2, conversion from apparent altitude to true altitude
//   - the true altitude resulting from the input apparent altitude, if this value
//     is a plausible apparent altitude, i.e. if it is a position above the ideal
//     horizon, taking into account the dip of the horizon
//   - the input altitude otherwise
//
// The body is above the horizon if the dret[0] != dret[1]
func (state *State) sweRefracExtended(inalt float64, altitude float64, pressure float64, temperature float64, lapseRate float64, calcFlag int32, dret []float64) float64 {
	var (
		refr   float64
		trualt float64
		dip    float64 = state.calcDip(altitude, pressure, temperature, lapseRate)
		D      float64
		D0     float64
		N      float64
		y      float64
		yy0    float64
		i      int32
	)
	if inalt > 90 {
		inalt = 180 - inalt
	}
	if calcFlag == 0 {
		if inalt < float64(-10) {
			if dret != nil {
				dret[0] = inalt
				dret[1] = inalt
				dret[2] = 0
				dret[3] = dip
			}
			return inalt
		}
		y = inalt
		D = 0.0
		yy0 = 0
		D0 = D
		for i = 0; i < 5; i++ {
			D = state.calcAstronomicalRefr(y, pressure, temperature)
			N = y - yy0
			yy0 = D - D0 - N
			if N != 0.0 && yy0 != 0.0 {
				N = y - N*(inalt+D-y)/yy0
			} else {
				N = inalt + D
			}
			yy0 = y
			D0 = D
			y = N
		}
		refr = D
		if inalt+refr < dip {
			if dret != nil {
				dret[0] = inalt
				dret[1] = inalt
				dret[2] = 0
				dret[3] = dip
			}
			return inalt
		}
		if dret != nil {
			dret[0] = inalt
			dret[1] = inalt + refr
			dret[2] = refr
			dret[3] = dip
		}
		return inalt + refr
	} else {
		refr = state.calcAstronomicalRefr(inalt, pressure, temperature)
		trualt = inalt - refr
		if dret != nil {
			if inalt > dip {
				dret[0] = trualt
				dret[1] = inalt
				dret[2] = refr
				dret[3] = dip
			} else {
				dret[0] = inalt
				dret[1] = inalt
				dret[2] = 0
				dret[3] = dip
			}
		}
		if inalt >= dip {
			return trualt
		} else {
			return inalt
		}
	}
}

func (state *State) calcAstronomicalRefr(inalt float64, pressure float64, temperature float64) float64 {
	var r float64
	if inalt > 17.904104638432 {
		r = 0.97 / math.Tan(inalt*radiansPerDegree)
	} else {
		r = (34.46 + 4.23*inalt + 0.004*inalt*inalt) / (1 + 0.505*inalt + 0.0845*inalt*inalt)
	}
	r = (pressure - 80) / 930 / (1 + 0.00008*(r+39)*(temperature-10)) * r / 60.0
	return r
}

func (state *State) calcDip(altitude float64, pressure float64, temperature float64, lapseRate float64) float64 {
	var (
		krefr float64 = (0.0342 + lapseRate) / 0.0036652000000000004
		d     float64 = 1 - 1.8480*krefr*pressure/(273.15+temperature)/(273.15+temperature)
	)
	return -degreesPerRadian * math.Acos(1/(1+altitude/6378136.6)) * math.Sqrt(d)
}

// sweLunEclipseHow implements swe_lun_eclipse_how from swecl.c.
//
//	Computes attributes of a lunar eclipse for given tjd and geopos
//
// retflag	SE_ECL_TOTAL or SE_ECL_PARTIAL
//
//	SE_ECL_PENUMBRAL
//	if 0, there is no eclipse
//
// attr[0]	umbral magnitude at tjd
// attr[1]      penumbral magnitude
// attr[4]	azimuth of moon at tjd
// attr[5]	true altitude of moon above horizon at tjd
// attr[6]	apparent altitude of moon above horizon at tjd
// attr[7]	distance of moon from opposition in degrees
// attr[8]	umbral magnitude at tjd (= attr[0])
// attr[9]	saros series number
// attr[10]	saros series member number
//
//	declare as attr[20] at least !
func (state *State) sweLunEclipseHow(jdUT float64, ifl int32, location []float64, attr []float64, message []byte) int32 {
	var (
		dcore  [10]float64
		lm     [6]float64
		xaz    [6]float64
		status int32
	)
	if location != nil {
	}
	if location != nil && (location[2] < -500.0 || location[2] > 25000.0) {
		if message != nil {
			formatText(message, textBytes("location for eclipses must be between %.0f and %.0f m above sea"), -500.0, 25000.0)
		}
		return -1
	}
	ifl = ifl & ^(32 * 1024)
	ifl &= ^(256*1024 | 512*1024)
	state.swiSetTidAcc(jdUT, ifl, 0, message)
	status = state.lunEclipseHow(jdUT, ifl, attr, dcore[:], message)
	if location == nil {
		return status
	}
	state.sweSetTopo(location[0], location[1], location[2])
	if state.sweCalcUt(jdUT, 1, ifl|32*1024|2*1024, lm[:], message) == -1 {
		return -1
	}
	state.sweAzalt(jdUT, 1, location, 0, 10, lm[:], xaz[:])
	attr[4] = xaz[0]
	attr[5] = xaz[1]
	attr[6] = xaz[2]
	if xaz[2] <= 0 {
		status = 0
	}
	return status
}

// lunEclipseHow implements lun_eclipse_how from swecl.c.
//
// attr[]: 	see swe_lun_eclipse_how()
//
// dcore[0]:	distance of shadow axis from geocenter r0
// dcore[1]:	diameter of core shadow on fundamental plane d0
// dcore[2]:	diameter of half-shadow on fundamental plane D0
func (state *State) lunEclipseHow(jdUT float64, ifl int32, attr []float64, dcore []float64, message []byte) int32 {
	var (
		i      int32
		j      int32
		k      int32
		status int32 = 0
		e      [6]float64
		rm     [6]float64
		rs     [6]float64
		dsm    float64
		d0     float64
		D0     float64
		s0     float64
		r0     float64
		ds     float64
		dm     float64
		dctr   float64
		x1     [6]float64
		x2     [6]float64
		f1     float64
		f2     float64
		deltat float64
		jd     float64
		d      float64
		cosf1  float64
		cosf2  float64
		rmoon  float64 = 1.1618815106570899e-05
		dmoon  float64 = 2 * rmoon
		flags  int32
	)
	for i = 0; i < 10; i++ {
		dcore[i] = 0
	}
	for i = 0; i < 20; i++ {
		attr[i] = 0
	}
	flags = 256 | 2*1024 | ifl
	flags = flags | (flagCartesian)
	deltat = state.sweDeltatEx(jdUT, ifl, message)
	jd = jdUT + deltat
	if state.sweCalc(jd, 1, flags, rm[:], message) == -1 {
		return -1
	}
	dm = math.Sqrt(rm[0]*rm[0] + rm[1]*rm[1] + rm[2]*rm[2])
	if state.sweCalc(jd, 0, flags, rs[:], message) == -1 {
		return -1
	}
	ds = math.Sqrt(rs[0]*rs[0] + rs[1]*rs[1] + rs[2]*rs[2])
	for i = 0; i < 3; i++ {
		x1[i] = rs[i] / ds
		x2[i] = rm[i] / dm
	}
	dctr = math.Acos(state.swiDotProdUnit(x1[:], x2[:])) * degreesPerRadian
	for i = 0; i <= 2; i++ {
		rs[i] -= rm[i]
	}
	for i = 0; i <= 2; i++ {
		rm[i] = -rm[i]
	}
	for i = 0; i <= 2; i++ {
		e[i] = rm[i] - rs[i]
	}
	dsm = math.Sqrt(e[0]*e[0] + e[1]*e[1] + e[2]*e[2])
	for i = 0; i <= 2; i++ {
		e[i] /= dsm
	}
	f1 = 0.0046098374045908134 / dsm
	cosf1 = math.Sqrt(1 - f1*f1)
	f2 = 0.0046951078696068635 / dsm
	cosf2 = math.Sqrt(1 - f2*f2)
	s0 = -(rm[0]*e[0] + rm[1]*e[1] + rm[2]*e[2])
	r0 = math.Sqrt(dm*dm - s0*s0)
	d0 = math.Abs(s0/dsm*0.0092196748091816268-8.5270465016050529e-05) * 1.02 / cosf1
	D0 = (s0/dsm*0.0093902157392137271 + 8.5270465016050529e-05) * 1.02 / cosf2
	d0 /= cosf1
	D0 /= cosf2
	d0 *= 0.99405
	D0 *= 0.98813
	dcore[0] = r0
	dcore[1] = d0
	dcore[2] = D0
	dcore[3] = cosf1
	dcore[4] = cosf2
	status = 0
	if d0/2 >= r0+rmoon/cosf1 {
		status = 4
		attr[0] = (d0/2 - r0 + rmoon) / dmoon
	} else if d0/2 >= r0-rmoon/cosf1 {
		status = 16
		attr[0] = (d0/2 - r0 + rmoon) / dmoon
	} else if D0/2 >= r0-rmoon/cosf2 {
		status = 64
		attr[0] = 0
	} else if message != nil {
		formatText(message, textBytes("no lunar eclipse at tjd = %f"), jd)
	}
	attr[8] = attr[0]
	attr[1] = (D0/2 - r0 + rmoon) / dmoon
	if status != 0 {
		attr[7] = 180 - math.Abs(dctr)
	}
	for i = 0; i < 180; i++ {
		d = (jdUT - eclipsesSarosDataLunar[i].startEpoch) / 6585.3213
		if d < 0 && d*6585.3213 > float64(-2) {
			d = 0.0000001
		}
		if d < 0 {
			continue
		}
		j = int32(d)
		if (d-float64(j))*6585.3213 < 2 {
			attr[9] = float64(eclipsesSarosDataLunar[i].seriesNo)
			attr[10] = float64(j) + 1
			break
		}
		k = j + 1
		if (float64(k)-d)*6585.3213 < 2 {
			attr[9] = float64(eclipsesSarosDataLunar[i].seriesNo)
			attr[10] = float64(k) + 1
			break
		}
	}
	if i == 180 {
		attr[10] = float64(-99999999)
		attr[9] = attr[10]
	}
	return status
}

// sweLunEclipseWhen implements swe_lun_eclipse_when from swecl.c.
//
//	When is the next lunar eclipse?
//
// retflag	SE_ECL_TOTAL or SE_ECL_PENUMBRAL or SE_ECL_PARTIAL
//
// tret[0]	time of maximum eclipse
// tret[1]
// tret[2]	time of partial phase begin (indices consistent with solar eclipses)
// tret[3]	time of partial phase end
// tret[4]	time of totality begin
// tret[5]	time of totality end
// tret[6]	time of penumbral phase begin
// tret[7]	time of penumbral phase end
func (state *State) sweLunEclipseWhen(tjdStart float64, ifl int32, ifltype int32, tret []float64, backward int32, message []byte) int32 {
	var (
		i         int32
		j         int32
		m         int32
		n         int32
		o         int32
		i1        int32 = 0
		i2        int32 = 0
		retflag   int32 = 0
		retflag2  int32 = 0
		t         float64
		jd        float64
		nextJD    float64
		dt        float64
		dtint     [1]float64
		dta       float64
		dtb       float64
		T         float64
		T2        float64
		T3        float64
		T4        float64
		K         float64
		F         float64
		M         float64
		Mm        float64
		E         float64
		Ff        float64
		F1        float64
		A1        float64
		Om        float64
		xs        [6]float64
		xm        [6]float64
		dm        float64
		ds        float64
		rsun      float64
		rearth    float64
		dcore     [10]float64
		dc        [3]float64
		dctr      [1]float64
		twohr     float64    = 0.083333333333333329
		tenmin    float64    = 0.0069444444444444449
		dt1       [1]float64 = [1]float64{0}
		dt2       [1]float64 = [1]float64{0}
		kk        float64
		attr      [20]float64
		dtstart   float64
		dtdiv     float64
		xa        [6]float64
		xb        [6]float64
		direction int32 = 1
		flags     int32
		iflagcart int32
	)
	ifl &= 1 | 2 | 4
	state.swiSetTidAcc(tjdStart, ifl, 0, message)
	flags = 2*1024 | ifl
	iflagcart = flags | (flagCartesian)
	ifltype &= ^(1 | 2)
	if ifltype&(8|32) != 0 {
		ifltype &= ^(8 | 32)
		if ifltype == 0 {
			if message != nil {
				copyText(message, textBytes("annular lunar eclipses don't exist"))
			}
			return -1
		}
	}
	if ifltype == 0 {
		ifltype = 4 | 64 | 16
	}
	if backward != 0 {
		direction = -1
	}
	K = float64(int32((tjdStart - julianDayJ2000) / 365.2425 * 12.3685))
	K -= float64(direction)
nextTry:
	for {
		retflag = 0
		for i = 0; i <= 9; i++ {
			tret[i] = 0
		}
		kk = K + 0.5
		T = kk / 1236.85
		T2 = T * T
		T3 = T2 * T
		T4 = T3 * T
		F = state.sweDegnorm(160.7108 + 390.67050274*kk - 0.0016341*T2 - 0.00000227*T3 + 0.000000011*T4)
		Ff = F
		if Ff > 180 {
			Ff -= 180
		}
		if Ff > 21 && Ff < 159 {
			K += float64(direction)
			continue nextTry
		}
		jd = 2451550.09765 + 29.530588853*kk + 0.0001337*T2 - 0.000000150*T3 + 0.00000000073*T4
		M = state.sweDegnorm(2.5534 + 29.10535669*kk - 0.0000218*T2 - 0.00000011*T3)
		Mm = state.sweDegnorm(201.5643 + 385.81693528*kk + 0.1017438*T2 + 0.00001239*T3 + 0.000000058*T4)
		Om = state.sweDegnorm(124.7746 - 1.56375580*kk + 0.0020691*T2 + 0.00000215*T3)
		E = 1 - 0.002516*T - 0.0000074*T2
		A1 = state.sweDegnorm(299.77 + 0.107408*kk - 0.009173*T2)
		M *= radiansPerDegree
		Mm *= radiansPerDegree
		F *= radiansPerDegree
		Om *= radiansPerDegree
		F1 = F - 0.02665*math.Sin(Om)*radiansPerDegree
		A1 *= radiansPerDegree
		jd = jd - 0.4075*math.Sin(Mm) + 0.1721*E*math.Sin(M) + 0.0161*math.Sin(2*Mm) - 0.0097*math.Sin(2*F1) + 0.0073*E*math.Sin(Mm-M) - 0.0050*E*math.Sin(Mm+M) - 0.0023*math.Sin(Mm-2*F1) + 0.0021*E*math.Sin(2*M) + 0.0012*math.Sin(Mm+2*F1) + 0.0006*E*math.Sin(2*Mm+M) - 0.0004*math.Sin(3*Mm) - 0.0003*E*math.Sin(M+2*F1) + 0.0003*math.Sin(A1) - 0.0002*E*math.Sin(M-2*F1) - 0.0002*E*math.Sin(2*Mm-M) - 0.0002*math.Sin(Om)
		dtstart = 0.1
		if jd < 2100000 || jd > 2500000 {
			dtstart = 5
		}
		dtdiv = 4
		j = 0
		for dt = dtstart; dt > 0.001; {
			i = 0
			for t = jd - dt; i <= 2; {
				if state.sweCalc(t, 0, iflagcart, xs[:], message) == -1 {
					return -1
				}
				if state.sweCalc(t, 1, iflagcart, xm[:], message) == -1 {
					return -1
				}
				for m = 0; m < 3; m++ {
					xs[m] -= xm[m]
					xm[m] = -xm[m]
				}
				ds = math.Sqrt(xs[0]*xs[0] + xs[1]*xs[1] + xs[2]*xs[2])
				dm = math.Sqrt(xm[0]*xm[0] + xm[1]*xm[1] + xm[2]*xm[2])
				for m = 0; m < 3; m++ {
					xa[m] = xs[m] / ds
					xb[m] = xm[m] / dm
				}
				dc[i] = math.Acos(state.swiDotProdUnit(xa[:], xb[:])) * degreesPerRadian
				rearth = math.Asin(4.2635232508025265e-05/dm) * degreesPerRadian
				rsun = math.Asin(0.0046524726370988385/ds) * degreesPerRadian
				dc[i] -= rearth + rsun
				i++
				t += dt
			}
			state.findMaximum(dc[0], dc[1], dc[2], dt, dtint[:], dctr[:])
			jd += dtint[0] + dt
			j++
			dt /= dtdiv
		}
		nextJD = jd - state.sweDeltatEx(jd, ifl, message)
		nextJD = jd - state.sweDeltatEx(nextJD, ifl, message)
		jd = jd - state.sweDeltatEx(nextJD, ifl, message)
		if retflag = state.sweLunEclipseHow(jd, ifl, nil, attr[:], message); retflag == -1 {
			return retflag
		}
		if retflag == 0 {
			K += float64(direction)
			continue nextTry
		}
		tret[0] = jd
		if backward != 0 && tret[0] >= tjdStart-0.0001 || !(backward != 0) && tret[0] <= tjdStart+0.0001 {
			K += float64(direction)
			continue nextTry
		}
		if !(ifltype&64 != 0) && retflag&64 != 0 {
			K += float64(direction)
			continue nextTry
		}
		if !(ifltype&16 != 0) && retflag&16 != 0 {
			K += float64(direction)
			continue nextTry
		}
		if !(ifltype&4 != 0) && retflag&4 != 0 {
			K += float64(direction)
			continue nextTry
		}
		if retflag&64 != 0 {
			o = 0
		} else if retflag&16 != 0 {
			o = 1
		} else {
			o = 2
		}
		dta = twohr
		dtb = tenmin
		for n = 0; n <= o; n++ {
			if n == 0 {
				i1 = 6
				i2 = 7
			} else if n == 1 {
				i1 = 2
				i2 = 3
			} else if n == 2 {
				i1 = 4
				i2 = 5
			}
			i = 0
			for t = jd - dta; i <= 2; {
				if retflag2 = state.lunEclipseHow(t, ifl, attr[:], dcore[:], message); retflag2 == -1 {
					return retflag2
				}
				if n == 0 {
					dc[i] = dcore[2]/2 + 1.1618815106570899e-05/dcore[4] - dcore[0]
				} else if n == 1 {
					dc[i] = dcore[1]/2 + 1.1618815106570899e-05/dcore[3] - dcore[0]
				} else if n == 2 {
					dc[i] = dcore[1]/2 - 1.1618815106570899e-05/dcore[3] - dcore[0]
				}
				i += 1
				t += dta
			}
			state.findZero(dc[0], dc[1], dc[2], dta, dt1[:], dt2[:])
			dtb = (dt1[0] + dta) / 2
			tret[i1] = jd + dt1[0] + dta
			tret[i2] = jd + dt2[0] + dta
			m = 0
			for dt = dtb / 2; m < 3; {
				for j = i1; j <= i2; j += i2 - i1 {
					i = 0
					for t = tret[j] - dt; i < 2; {
						if retflag2 = state.lunEclipseHow(t, ifl, attr[:], dcore[:], message); retflag2 == -1 {
							return retflag2
						}
						if n == 0 {
							dc[i] = dcore[2]/2 + 1.1618815106570899e-05/dcore[4] - dcore[0]
						} else if n == 1 {
							dc[i] = dcore[1]/2 + 1.1618815106570899e-05/dcore[3] - dcore[0]
						} else if n == 2 {
							dc[i] = dcore[1]/2 - 1.1618815106570899e-05/dcore[3] - dcore[0]
						}
						i++
						t += dt
					}
					dt1[0] = dc[1] / ((dc[1] - dc[0]) / dt)
					tret[j] -= dt1[0]
				}
				m++
				dt /= 2
			}
		}
		return retflag
	}
}

// sweLunEclipseWhenLoc implements swe_lun_eclipse_when_loc from swecl.c.
//
//	When is the next lunar eclipse, observable at a geographic position?
//
// retflag	SE_ECL_TOTAL or SE_ECL_PENUMBRAL or SE_ECL_PARTIAL
//
// tret[0]	time of maximum eclipse
// tret[1]
// tret[2]	time of partial phase begin (indices consistent with solar eclipses)
// tret[3]	time of partial phase end
// tret[4]	time of totality begin
// tret[5]	time of totality end
// tret[6]	time of penumbral phase begin
// tret[7]	time of penumbral phase end
// tret[8]	time of moonrise, if it occurs during the eclipse
// tret[9]	time of moonset, if it occurs during the eclipse
//
// attr[0]      umbral magnitude at tjd
// attr[1]      penumbral magnitude
// attr[4]      azimuth of moon at tjd
// attr[5]      true altitude of moon above horizon at tjd
// attr[6]      apparent altitude of moon above horizon at tjd
// attr[7]      distance of moon from opposition in degrees
// attr[8]      umbral magnitude at tjd (= attr[0])
// attr[9]      saros series number
// attr[10]     saros series member number
//
//	declare as attr[20] at least !
func (state *State) sweLunEclipseWhenLoc(tjdStart float64, ifl int32, location []float64, tret []float64, attr []float64, backward int32, message []byte) int32 {
	var (
		retflag  int32 = 0
		retflag2 int32 = 0
		status   int32
		tjdr     [1]float64
		tjds     [1]float64
		tjdMax   float64 = 0
		i        int32
	)
	if location != nil && (location[2] < -500.0 || location[2] > 25000.0) {
		if message != nil {
			formatText(message, textBytes("location for eclipses must be between %.0f and %.0f m above sea"), -500.0, 25000.0)
		}
		return -1
	}
	ifl &= ^(256*1024 | 512*1024)
nextLunEcl:
	for {
		if retflag = state.sweLunEclipseWhen(tjdStart, ifl, 0, tret, backward, message); retflag == -1 {
			return -1
		}
		retflag = 0
		for i = 7; i >= 0; i-- {
			if i == 1 {
				continue
			}
			if tret[i] == 0 {
				continue
			}
			if retflag2 = state.sweLunEclipseHow(tret[i], ifl, location, attr, message); retflag2 == -1 {
				return -1
			}
			if attr[6] > 0 {
				retflag |= 128
				switch i {
				case 0:
					retflag |= 256
				case 2:
					retflag |= 512
				case 3:
					retflag |= 4096
				case 4:
					retflag |= 1024
				case 5:
					retflag |= 2048
				case 6:
					retflag |= 8192
				case 7:
					retflag |= 16384
				default:
				}
			}
		}
		if !(retflag&128 != 0) {
			if backward != 0 {
				tjdStart = tret[0] - 25
			} else {
				tjdStart = tret[0] + 25
			}
			continue nextLunEcl
		}
		tjdMax = tret[0]
		if status = state.sweRiseTrans(tret[6]-0.001, 1, nil, ifl, 1|8192, location, 0, 0, tjdr[:], message); status == -1 {
			return -1
		}
		if status >= 0 {
			status = state.sweRiseTrans(tret[6]-0.001, 1, nil, ifl, 2|8192, location, 0, 0, tjds[:], message)
			if status == -1 {
				return -1
			}
		}
		if status >= 0 {
			if tjds[0] < tret[6] || tjds[0] > tjdr[0] && tjdr[0] > tret[7] {
				if backward != 0 {
					tjdStart = tret[0] - 25
				} else {
					tjdStart = tret[0] + 25
				}
				continue nextLunEcl
			}
			if tjdr[0] > tret[6] && tjdr[0] < tret[7] {
				tret[6] = 0
				for i = 2; i <= 5; i++ {
					if tjdr[0] > tret[i] {
						tret[i] = 0
					}
				}
				tret[8] = tjdr[0]
				if tjdr[0] > tret[0] {
					tjdMax = tjdr[0]
				}
			}
			if tjds[0] > tret[6] && tjds[0] < tret[7] {
				tret[7] = 0
				for i = 2; i <= 5; i++ {
					if tjds[0] < tret[i] {
						tret[i] = 0
					}
				}
				tret[9] = tjds[0]
				if tjds[0] < tret[0] {
					tjdMax = tjds[0]
				}
			}
		}
		tret[0] = tjdMax
		if retflag2 = state.sweLunEclipseHow(tjdMax, ifl, location, attr, message); retflag2 == -1 {
			return -1
		}
		if retflag2 == 0 {
			if backward != 0 {
				tjdStart = tret[0] - 25
			} else {
				tjdStart = tret[0] + 25
			}
			continue nextLunEcl
		}
		retflag |= retflag2 & (4 | 16 | 64)
		return retflag
	}
}

// swePheno implements swe_pheno from swecl.c.
//
//	Vesta
func (state *State) swePheno(jd float64, body int32, flags int32, attr []float64, message []byte) int32 {
	var (
		i         int32
		xx        [6]float64
		xx2       [6]float64
		xxs       [6]float64
		lbr       [6]float64
		lbr2      [6]float64
		dt        float64 = 0
		dd        float64
		fac       float64
		T         float64
		in        float64
		om        float64
		sinB      float64
		ph1       float64
		ph2       float64
		me        [2]float64
		iflagp    int32
		ephemeris int32
		retflag   int32
		epheflag2 int32
		serr2     [256]byte
	)
	serr2[0] = byte('\000')
	flags &= ^(256*1024 | 512*1024)
	if body == 10000+134340 {
		body = 9
	}
	for i = 0; i < 20; i++ {
		attr[i] = 0
	}
	if body > 10000 && body <= 10000+4 {
		body = body - 10000 - 1 + 17
	}
	flags = flags & (flagJPL | flagSwiss | flagMoshier | flagTruePosition | flagJ2000 | flagNoNutation | flagNoDeflection | flagNoAberration | flagTopocentric)
	iflagp = flags & (flagJPL | flagSwiss | flagMoshier | flagTruePosition | flagJ2000 | flagNoNutation | flagNoAberration)
	iflagp |= 8
	ephemeris = flags & (flagJPL | flagSwiss | flagMoshier)
	if retflag = state.sweCalc(jd, body, flags|(flagCartesian), xx[:], message); retflag == -1 {
		return -1
	}
	epheflag2 = retflag & (1 | 2 | 4)
	if ephemeris != epheflag2 {
		flags &= ^ephemeris
		iflagp &= ^ephemeris
		flags |= epheflag2
		iflagp |= epheflag2
		ephemeris = epheflag2
	}
	if state.sweCalc(jd, body, flags, lbr[:], message) == -1 {
		return -1
	}
	if body == 1 {
		if state.sweCalc(jd, 0, flags|(flagCartesian), xxs[:], message) == -1 {
			return -1
		}
	}
	if body != 0 && body != 14 && body != 10 && body != 11 && body != 12 && body != 13 {
		dt = lbr[2] * 1.49597870700e+11 / 2.99792458e+8 / secondsPerDay
		if flags&flagTruePosition != 0 {
			dt = 0
		}
		if state.sweCalc(jd-dt, body, iflagp|4*1024, xx2[:], message) == -1 {
			return -1
		}
		if state.sweCalc(jd-dt, body, iflagp, lbr2[:], message) == -1 {
			return -1
		}
		attr[0] = math.Acos(state.swiDotProdUnit(xx[:], xx2[:])) * degreesPerRadian
		attr[1] = (1 + math.Cos(attr[0]*radiansPerDegree)) / 2
	}
	if body < 20+1 {
		dd = calendarPlaDiam[body]
	} else if body > 10000 {
		dd = state.positions.data.astDiam * 1000
	} else {
		dd = 0
	}
	if lbr[2] < dd/2/1.49597870700e+11 {
		attr[3] = 180
	} else {
		attr[3] = math.Asin(dd/2/1.49597870700e+11/lbr[2]) * 2 * degreesPerRadian
	}
	if body > 10000 || body < 20+1 && eclipsesMagElem[body][0] < 99 {
		if body == 0 {
			fac = attr[3] / (math.Asin(calendarPlaDiam[0]/2.0/1.49597870700e+11) * 2 * degreesPerRadian)
			fac *= fac
			attr[4] = eclipsesMagElem[body][0] - 2.5*math.Log10(fac)
		} else if body == 1 {
			var a float64 = attr[0]
			if a <= 147.1385465 {
				attr[4] = -21.62 + 0.026*math.Abs(a) + 0.000000004*math.Pow(a, 4)
				attr[4] += 5 * math.Log10(lbr[2]*lbr2[2]*1.49597870700e+11/6378136.6)
			} else {
				attr[4] = -4.5444 - 2.5*math.Log10(math.Pow(180-a, 3))
				attr[4] += 5 * math.Log10(lbr[2]*lbr2[2]*1.49597870700e+11/6378136.6)
			}
		} else if body == 2 {
			var (
				a  float64 = attr[0]
				a2 float64 = a * a
				a3 float64 = a2 * a
				a4 float64 = a3 * a
				a5 float64 = a4 * a
				a6 float64 = a5 * a
			)
			attr[4] = -0.613 + a*6.3280e-02 - a2*1.6336e-03 + a3*3.3644e-05 - a4*3.4265e-07 + a5*1.6893e-09 - a6*3.0334e-12
			attr[4] += 5 * math.Log10(lbr2[2]*lbr[2])
		} else if body == 3 {
			var (
				a  float64 = attr[0]
				a2 float64 = a * a
				a3 float64 = a2 * a
				a4 float64 = a3 * a
			)
			if a <= 163.7 {
				attr[4] = -4.384 - a*1.044e-03 + a2*3.687e-04 - a3*2.814e-06 + a4*8.938e-09
			} else {
				attr[4] = 236.05828 - a*2.81914e+00 + a2*8.39034e-03
			}
			attr[4] += 5 * math.Log10(lbr2[2]*lbr[2])
			if attr[0] > 179.0 {
				formatText(serr2[:], textBytes("magnitude value for Venus at phase angle i=%.1f is bad; formula is valid only for i < 179.0"), attr[0])
			}
		} else if body == 4 {
			var (
				a  float64 = attr[0]
				a2 float64 = a * a
			)
			if a <= 50.0 {
				attr[4] = -1.601 + a*0.02267 - a2*0.0001302
			} else {
				attr[4] = -0.367 - a*0.02573 + a2*0.0003445
			}
			attr[4] += 5 * math.Log10(lbr2[2]*lbr[2])
		} else if body == 5 {
			var (
				a  float64 = attr[0]
				a2 float64 = a * a
			)
			attr[4] = -9.395 - a*3.7e-04 + a2*6.16e-04
			attr[4] += 5 * math.Log10(lbr2[2]*lbr[2])
		} else if body == 6 {
			var (
				a     float64 = attr[0]
				sinB2 float64
			)
			T = (jd - dt - julianDayJ2000) / daysPerCentury
			in = (28.075216 - 0.012998*T + 0.000004*T*T) * radiansPerDegree
			om = (169.508470 + 1.394681*T + 0.000412*T*T) * radiansPerDegree
			sinB = math.Sin(in)*math.Cos(lbr[1]*radiansPerDegree)*math.Sin(lbr[0]*radiansPerDegree-om) - math.Cos(in)*math.Sin(lbr[1]*radiansPerDegree)
			sinB2 = math.Sin(in)*math.Cos(lbr2[1]*radiansPerDegree)*math.Sin(lbr2[0]*radiansPerDegree-om) - math.Cos(in)*math.Sin(lbr2[1]*radiansPerDegree)
			sinB = math.Abs(math.Sin((math.Asin(sinB) + math.Asin(sinB2)) / 2.0))
			attr[4] = -8.914 - 1.825*sinB + 0.026*a - 0.378*sinB*math.Pow(2.7182818, -2.25*a)
			attr[4] += 5 * math.Log10(lbr2[2]*lbr[2])
		} else if body == 7 {
			var (
				a   float64 = attr[0]
				a2  float64 = a * a
				fi_ float64 = 0
			)
			attr[4] = -7.110 - 8.4e-04*fi_ + a*6.587e-3 + a2*1.045e-4
			attr[4] += 5 * math.Log10(lbr2[2]*lbr[2])
			attr[4] -= 0.05
		} else if body == 8 {
			if jd < 2444239.5 {
				attr[4] = -6.89
			} else if jd <= 2451544.5 {
				attr[4] = -6.89 - 0.0055*(jd-2444239.5)/365.25
			} else {
				attr[4] = -7.00
			}
			attr[4] += 5 * math.Log10(lbr2[2]*lbr[2])
		} else if body < 15 {
			attr[4] = 5*math.Log10(lbr2[2]*lbr[2]) + eclipsesMagElem[body][1]*attr[0]/100.0 + eclipsesMagElem[body][2]*attr[0]*attr[0]/10000.0 + eclipsesMagElem[body][3]*attr[0]*attr[0]*attr[0]/1000000.0 + eclipsesMagElem[body][0]
		} else if body < 20+1 || body > 10000 {
			ph1 = math.Pow(2.718281828459, -3.33*math.Pow(math.Tan(attr[0]*radiansPerDegree/2), 0.63))
			ph2 = math.Pow(2.718281828459, -1.87*math.Pow(math.Tan(attr[0]*radiansPerDegree/2), 1.22))
			if body < 20+1 {
				me[0] = eclipsesMagElem[body][0]
				me[1] = eclipsesMagElem[body][1]
			} else if body == 10000+1566 {
				me[0] = 16.9
				me[1] = 0.15
			} else {
				me[0] = state.positions.data.astH
				me[1] = state.positions.data.astG
			}
			attr[4] = 5*math.Log10(lbr2[2]*lbr[2]) + me[0] - 2.5*math.Log10((1-me[1])*ph1+me[1]*ph2)
		} else {
			attr[4] = 0
		}
	}
	if body != 0 && body != 14 {
		if state.sweCalc(jd, 0, flags|(flagCartesian), xx2[:], message) == -1 {
			return -1
		}
		if state.sweCalc(jd, 0, flags, lbr2[:], message) == -1 {
			return -1
		}
		attr[2] = math.Acos(state.swiDotProdUnit(xx[:], xx2[:])) * degreesPerRadian
	}
	if body == 1 {
		var (
			sinhp float64
			xm    [6]float64
		)
		if state.sweCalc(jd, body, ephemeris|16|2*1024|8*1024, xm[:], message) == -1 {
			return -1
		}
		sinhp = 6378136.6 / xm[2] / 1.49597870700e+11
		attr[5] = math.Asin(sinhp) / radiansPerDegree
		if flags&flagTopocentric != 0 {
			if state.sweCalc(jd, body, ephemeris|4*1024|32*1024, xm[:], message) == -1 {
				return -1
			}
			if state.sweCalc(jd, body, ephemeris|4*1024, xx[:], message) == -1 {
				return -1
			}
			attr[5] = math.Acos(state.swiDotProdUnit(xm[:], xx[:])) / radiansPerDegree
		}
	}
	if int32(serr2[0]) != '\000' && message != nil {
		copyText(message, serr2[:])
	}
	return flags
}

// swePhenoUt implements swe_pheno_ut from swecl.c.
func (state *State) swePhenoUt(jdUT float64, body int32, flags int32, attr []float64, message []byte) int32 {
	var (
		deltat    float64
		retflag   int32 = 0
		ephemeris int32 = flags & (flagJPL | flagSwiss | flagMoshier)
	)
	if ephemeris == 0 {
		ephemeris = 2
		flags |= flagSwiss
	}
	deltat = state.sweDeltatEx(jdUT, flags, message)
	retflag = state.swePheno(jdUT+deltat, body, flags, attr, message)
	if retflag&(1|2|4) != ephemeris {
		deltat = state.sweDeltatEx(jdUT, retflag, message)
		retflag = state.swePheno(jdUT+deltat, body, flags, attr, message)
	}
	return retflag
}

// findMaximum implements find_maximum from swecl.c.
func (state *State) findMaximum(y00 float64, y11 float64, y2 float64, dx float64, dxret []float64, yret []float64) int32 {
	var (
		a float64
		b float64
		c float64
		x float64
		y float64
	)
	c = y11
	b = (y2 - y00) / 2.0
	a = (y2+y00)/2.0 - c
	x = -b / 2 / a
	y = (4*a*c - b*b) / 4 / a
	dxret[0] = (x - 1) * dx
	if yret != nil {
		yret[0] = y
	}
	return 0
}

// findZero implements find_zero from swecl.c.
func (state *State) findZero(y00 float64, y11 float64, y2 float64, dx float64, dxret []float64, dxret2 []float64) int32 {
	var (
		a  float64
		b  float64
		c  float64
		x1 float64
		x2 float64
	)
	c = y11
	b = (y2 - y00) / 2.0
	a = (y2+y00)/2.0 - c
	if b*b-4*a*c < 0 {
		return -1
	}
	x1 = (-b + math.Sqrt(b*b-4*a*c)) / 2 / a
	x2 = (-b - math.Sqrt(b*b-4*a*c)) / 2 / a
	dxret[0] = (x1 - 1) * dx
	dxret2[0] = (x2 - 1) * dx
	return 0
}

// rdiTwilight implements rdi_twilight from swecl.c.
func (state *State) rdiTwilight(eventFlags int32) float64 {
	var rdi float64 = 0
	if eventFlags&1024 != 0 {
		rdi = 6
	}
	if eventFlags&2048 != 0 {
		rdi = 12
	}
	if eventFlags&4096 != 0 {
		rdi = 18
	}
	return rdi
}

// getSunRadPlusRefr implements get_sun_rad_plus_refr from swecl.c.
func (state *State) getSunRadPlusRefr(body int32, dd float64, eventFlags int32, refr float64) float64 {
	var rdi float64 = 0
	if eventFlags&16384 != 0 {
		if body == 0 {
			dd = 1.0
		} else if body == 1 {
			dd = 0.00257
		}
	}
	if !(eventFlags&256 != 0) {
		rdi = math.Asin(calendarPlaDiam[body]/2.0/1.49597870700e+11/dd) * degreesPerRadian
	}
	if eventFlags&8192 != 0 {
		rdi = -rdi
	}
	if !(eventFlags&512 != 0) {
		rdi += refr
	}
	return rdi
}

// riseSetFast implements rise_set_fast from swecl.c.
//
//	Simple fast algorithm for risings and settings of
//
// - planets Sun, Moon, Mercury - Pluto + Lunar Nodes and Fixed stars
// Does not work well for geographic latitudes
// > 65 N/S for the Sun
// > 60 N/S for the Moon and the planets
// and is called only for latitudes smaller than this.
func (state *State) riseSetFast(jdUT float64, body int32, ephemeris int32, eventFlags int32, location []float64, pressure float64, temperature float64, tret []float64, message []byte) int32 {
	var (
		i           int32
		xx          [6]float64
		xaz         [6]float64
		xaz2        [6]float64
		dd          float64
		dt          float64
		refr        float64
		dtsum       float64 = 0
		flags       int32   = ephemeris & (1 | 2 | 4)
		iflagtopo   int32   = flags | (flagEquatorial)
		sda         float64
		armc        float64
		md          float64
		dmd         float64
		mdrise      float64
		rdi         float64
		tr          float64
		dalt        float64
		decl        float64
		tjdUt0      float64 = jdUT
		facrise     int32   = 1
		tohorFlag   int32   = 1
		isSecondRun int32   = 0
		nloop       int32   = 2
	)
	tret[0] = 0
	if body == 1 {
		nloop = 4
	}
	if eventFlags&2 != 0 {
		facrise = -1
	}
	if !(eventFlags&128 != 0) {
		iflagtopo |= 32 * 1024
		state.sweSetTopo(location[0], location[1], location[2])
	}
runRiseAgain:
	for {
		if state.sweCalcUt(jdUT, body, iflagtopo, xx[:], message) == -1 {
			return -1
		}
		decl = xx[1]
		sda = -math.Tan(location[1]*radiansPerDegree) * math.Tan(decl*radiansPerDegree)
		if sda >= 1 {
			sda = 10
		} else if sda <= float64(-1) {
			sda = 180
		} else {
			sda = math.Acos(sda) * degreesPerRadian
		}
		armc = state.sweDegnorm(state.sweSidtime(jdUT)*15 + location[0])
		md = state.sweDegnorm(xx[0] - armc)
		mdrise = state.sweDegnorm(sda * float64(facrise))
		dmd = state.sweDegnorm(md - mdrise)
		if dmd > 358 {
			dmd -= 360
		}
		tr = jdUT + dmd/360
		rdi = 0
		if pressure == 0 {
			pressure = 1013.25 * math.Pow(1-0.0065*location[2]/288, 5.255)
		}
		state.sweRefracExtended(0.000001, 0, pressure, temperature, state.eclipse.lapseRate, 1, xx[:])
		refr = xx[1] - xx[0]
		if eventFlags&128 != 0 {
			tohorFlag = 0
			iflagtopo = flags
		} else {
			tohorFlag = 1
			iflagtopo = flags | (flagEquatorial)
			iflagtopo |= 32 * 1024
			state.sweSetTopo(location[0], location[1], location[2])
		}
		for i = 0; i < nloop; i++ {
			if state.sweCalcUt(tr, body, iflagtopo, xx[:], message) == -1 {
				return -1
			}
			if eventFlags&128 != 0 {
				xx[1] = 0
			}
			rdi = state.getSunRadPlusRefr(body, xx[2], eventFlags, refr)
			state.sweAzalt(tr, tohorFlag, location, pressure, temperature, xx[:], xaz[:])
			state.sweAzalt(tr+0.001, tohorFlag, location, pressure, temperature, xx[:], xaz2[:])
			dd = xaz2[1] - xaz[1]
			dalt = xaz[1] + rdi
			dt = dalt / dd / 1000.0
			if dt > 0.1 {
				dt = 0.1
			} else if dt < -0.1 {
				dt = -0.1
			}
			dtsum += dt
			if 0 != 0 && math.Abs(dt) > 5.7870370370370373e-05 && nloop < 20 {
				nloop++
			}
			tr -= dt
		}
		if tr < tjdUt0 && !(isSecondRun != 0) {
			jdUT += 0.5
			isSecondRun = 1
			continue runRiseAgain
		}
		tret[0] = tr
		return 0
	}
}

// sweRiseTrans implements swe_rise_trans from swecl.c.
//
//	rise, set, and meridian transits of sun, moon, planets, and stars
//
// tjd_ut	universal time from when on search ought to start
// ipl		planet number, neglected, if starname is given
// starname	pointer to string. if a planet, not a star, is
//
//	wanted, starname must be NULL or ""
//
// epheflag	used for ephemeris only
// rsmi		SE_CALC_RISE, SE_CALC_SET, SE_CALC_MTRANSIT, SE_CALC_ITRANSIT
//
//	| SE_BIT_DISC_CENTER       for rises of disc center of body
//	| SE_BIT_DISC_BOTTOM       for rises of disc bottom of body
//	| SE_BIT_GEOCTR_NO_ECL_LAT use geocentric position of object
//	                           and ignore ecliptic latitude
//	| SE_BIT_NO_REFRACTION     to neglect refraction
//	| SE_BIT_CIVIL_TWILIGHT    calculate civil twilight
//	| SE_BIT_NAUTIC_TWILIGHT   calculate nautical twilight
//	| SE_BIT_ASTRO_TWILIGHT    calculate astronomical twilight
//	| SE_BIT_FIXED_DISC_SIZE   neglect the effect of distance on disc size
//	| SE_BIT_HINDU_RISING      risings according to Hindu astrology
//	                           = (SE_BIT_DISC_CENTER|SE_BIT_NO_REFRACTION|SE_BIT_GEOCTR_NO_ECL_LAT)
//
// geopos	array of doubles for geogr. long., lat. and height above sea
// atpress	atmospheric pressure
// attemp	atmospheric temperature
//
// return variables:
// tret         time of rise, set, meridian transits
// serr[256]	error string
// function return value -2 means that the body does not rise or set
func (state *State) sweRiseTrans(jdUT float64, body int32, starname []byte, ephemeris int32, eventFlags int32, location []float64, pressure float64, temperature float64, tret []float64, message []byte) int32 {
	var (
		retval    int32 = 0
		doFixstar int32 = int32(boolInt(starname != nil && int32(starname[0]) != '\000'))
	)
	if !(doFixstar != 0) && eventFlags&(1|2) != 0 && !(eventFlags&32768 != 0) && !(eventFlags&(1024|2048|4096) != 0) && (body >= 0 && body <= 11) && (math.Abs(location[1]) <= 60 || body == 0 && math.Abs(location[1]) <= 65) {
		retval = state.riseSetFast(jdUT, body, ephemeris, eventFlags, location, pressure, temperature, tret, message)
		return retval
	}
	return state.sweRiseTransTrueHor(jdUT, body, starname, ephemeris, eventFlags, location, pressure, temperature, 0, tret, message)
}

func (state *State) sweRiseTransTrueHor(jdUT float64, body int32, starname []byte, ephemeris int32, eventFlags int32, location []float64, pressure float64, temperature float64, horhgt float64, tret []float64, message []byte) int32 {
	var (
		i         int32
		j         int32
		k         int32
		ii        int32
		calcCulm  int32
		nculm     int32   = -1
		jdTT      float64 = jdUT + state.sweDeltatEx(jdUT, ephemeris, message)
		xc        [6]float64
		xh        [20][6]float64
		ah        [6]float64
		aha       float64
		tculm     [4]float64
		tcu       float64
		tc        [20]float64
		h         [20]float64
		t2        [6]float64
		dc        [6]float64
		dtint     [1]float64
		dx        [1]float64
		rdi       float64
		dd        float64 = 0
		flags     int32   = ephemeris
		jmax      int32   = 14
		t         float64
		te        float64
		tt        float64
		dt        float64
		twohrs    float64 = 0.083333333333333329
		curdist   float64
		tohorFlag int32 = 1
		nazalt    int32 = 0
		ncalc     int32 = 0
		doFixstar int32 = int32(boolInt(starname != nil && int32(starname[0]) != '\000'))
	)
	if location[2] < -500.0 || location[2] > 25000.0 {
		if message != nil {
			formatText(message, textBytes("location for swe_rise_trans() must be between %.0f and %.0f m above sea"), -500.0, 25000.0)
		}
		return -1
	}
	if horhgt == float64(-100) {
		horhgt = 0.0001 + state.calcDip(location[2], pressure, temperature, state.eclipse.lapseRate)
	}
	if body == 10000+134340 {
		body = 9
	}
	xh[0][0] = 0
	flags &= flagJPL | flagSwiss | flagMoshier | flagTruePosition | flagNoNutation
	tret[0] = 0
	if eventFlags&128 != 0 {
		tohorFlag = 0
	} else {
		tohorFlag = 1
		flags |= flagEquatorial
		flags |= flagTopocentric
		state.sweSetTopo(location[0], location[1], location[2])
	}
	if eventFlags&(4|8) != 0 {
		return state.calcMerTrans(jdUT, body, ephemeris, eventFlags, location, starname, tret, message)
	}
	if !(eventFlags&(1|2) != 0) {
		eventFlags |= 1
	}
	if body == 0 && eventFlags&(1024|2048|4096) != 0 {
		eventFlags |= 512 | 256
		horhgt = -state.rdiTwilight(eventFlags)
	}
	if doFixstar != 0 {
		if state.sweFixstar(starname, jdTT, flags, xc[:], message) == -1 {
			return -1
		}
	}
	ii = 0
	for t = jdUT - twohrs; ii <= jmax; {
		tc[ii] = t
		if !(doFixstar != 0) {
			te = t + state.sweDeltatEx(t, ephemeris, message)
			if state.sweCalc(te, body, flags, xc[:], message) == -1 {
				return -1
			}
			ncalc++
		}
		if eventFlags&128 != 0 {
			xc[1] = 0
		}
		if ii == 0 {
			if doFixstar != 0 {
				dd = 0
			} else if eventFlags&256 != 0 {
				dd = 0
			} else if body < 20+1 {
				dd = calendarPlaDiam[body]
			} else if body > 10000 {
				dd = state.positions.data.astDiam * 1000
			} else {
				dd = 0
			}
		}
		curdist = xc[2]
		if eventFlags&16384 != 0 {
			if body == 0 {
				curdist = 1.0
			} else if body == 1 {
				curdist = 0.00257
			}
		}
		rdi = math.Asin(dd/2/1.49597870700e+11/curdist) * degreesPerRadian
		state.sweAzalt(t, tohorFlag, location, pressure, temperature, xc[:], xh[ii][:])
		nazalt++
		if eventFlags&8192 != 0 {
			xh[ii][1] -= rdi
		} else {
			xh[ii][1] += rdi
		}
		if eventFlags&512 != 0 {
			xh[ii][1] -= horhgt
			h[ii] = xh[ii][1]
		} else {
			state.sweAzaltRev(t, 1, location, xh[ii][:], xc[:])
			nazalt++
			state.sweAzalt(t, 1, location, pressure, temperature, xc[:], xh[ii][:])
			nazalt++
			xh[ii][1] -= horhgt
			xh[ii][2] -= horhgt
			h[ii] = xh[ii][2]
		}
		calcCulm = 0
		if ii > 1 {
			dc[0] = xh[ii-2][1]
			dc[1] = xh[ii-1][1]
			dc[2] = xh[ii][1]
			if dc[1] > dc[0] && dc[1] > dc[2] {
				calcCulm = 1
			}
			if dc[1] < dc[0] && dc[1] < dc[2] {
				calcCulm = 2
			}
		}
		if calcCulm != 0 {
			dt = twohrs
			tcu = t - dt
			state.findMaximum(dc[0], dc[1], dc[2], dt, dtint[:], dx[:])
			tcu += dtint[0] + dt
			dt /= 3
			for ; dt > 0.0001; dt /= 3 {
				i = 0
				for tt = tcu - dt; i < 3; {
					te = tt + state.sweDeltatEx(tt, ephemeris, message)
					if !(doFixstar != 0) {
						if state.sweCalc(te, body, flags, xc[:], message) == -1 {
							return -1
						}
					}
					if eventFlags&128 != 0 {
						xc[1] = 0
					}
					ncalc++
					state.sweAzalt(tt, tohorFlag, location, pressure, temperature, xc[:], ah[:])
					nazalt++
					ah[1] -= horhgt
					dc[i] = ah[1]
					tt += dt
					i++
				}
				state.findMaximum(dc[0], dc[1], dc[2], dt, dtint[:], dx[:])
				tcu += dtint[0] + dt
			}
			nculm++
			tculm[nculm] = tcu
		}
		ii++
		t += twohrs
	}
	for i = 0; i <= nculm; i++ {
		for j = 1; j <= jmax; j++ {
			if tculm[i] < tc[j] {
				for k = jmax; k >= j; k-- {
					tc[k+1] = tc[k]
					h[k+1] = h[k]
				}
				tc[j] = tculm[i]
				if !(doFixstar != 0) {
					te = tc[j] + state.sweDeltatEx(tc[j], ephemeris, message)
					if state.sweCalc(te, body, flags, xc[:], message) == -1 {
						return -1
					}
					if eventFlags&128 != 0 {
						xc[1] = 0
					}
					ncalc++
				}
				curdist = xc[2]
				if eventFlags&16384 != 0 {
					if body == 0 {
						curdist = 1.0
					} else if body == 1 {
						curdist = 0.00257
					}
				}
				rdi = math.Asin(dd/2/1.49597870700e+11/curdist) * degreesPerRadian
				state.sweAzalt(tc[j], tohorFlag, location, pressure, temperature, xc[:], ah[:])
				nazalt++
				if eventFlags&8192 != 0 {
					ah[1] -= rdi
				} else {
					ah[1] += rdi
				}
				if eventFlags&512 != 0 {
					ah[1] -= horhgt
					h[j] = ah[1]
				} else {
					state.sweAzaltRev(tc[j], 1, location, ah[:], xc[:])
					nazalt++
					state.sweAzalt(tc[j], 1, location, pressure, temperature, xc[:], ah[:])
					nazalt++
					ah[1] -= horhgt
					ah[2] -= horhgt
					h[j] = ah[2]
				}
				jmax++
				break
			}
		}
	}
	tret[0] = 0
	for ii = 1; ii <= jmax; ii++ {
		if h[ii-1]*h[ii] >= 0 {
			continue
		}
		if h[ii-1] < h[ii] && !(eventFlags&1 != 0) {
			continue
		}
		if h[ii-1] > h[ii] && !(eventFlags&2 != 0) {
			continue
		}
		dc[0] = h[ii-1]
		dc[1] = h[ii]
		t2[0] = tc[ii-1]
		t2[1] = tc[ii]
		for i = 0; i < 20; i++ {
			t = (t2[0] + t2[1]) / 2
			if !(doFixstar != 0) {
				te = t + state.sweDeltatEx(t, ephemeris, message)
				if state.sweCalc(te, body, flags, xc[:], message) == -1 {
					return -1
				}
				if eventFlags&128 != 0 {
					xc[1] = 0
				}
				ncalc++
			}
			curdist = xc[2]
			if eventFlags&16384 != 0 {
				if body == 0 {
					curdist = 1.0
				} else if body == 1 {
					curdist = 0.00257
				}
			}
			rdi = math.Asin(dd/2/1.49597870700e+11/curdist) * degreesPerRadian
			state.sweAzalt(t, tohorFlag, location, pressure, temperature, xc[:], ah[:])
			nazalt++
			if eventFlags&8192 != 0 {
				ah[1] -= rdi
			} else {
				ah[1] += rdi
			}
			if eventFlags&512 != 0 {
				ah[1] -= horhgt
				aha = ah[1]
			} else {
				state.sweAzaltRev(t, 1, location, ah[:], xc[:])
				nazalt++
				state.sweAzalt(t, 1, location, pressure, temperature, xc[:], ah[:])
				nazalt++
				ah[1] -= horhgt
				ah[2] -= horhgt
				aha = ah[2]
			}
			if aha*dc[0] <= 0 {
				dc[1] = aha
				t2[1] = t
			} else {
				dc[0] = aha
				t2[0] = t
			}
		}
		if t > jdUT {
			tret[0] = t
			return 0
		}
	}
	if message != nil {
		formatText(message, textBytes("rise or set not found for planet %d"), body)
	}
	return -2
}

// calcMerTrans implements calc_mer_trans from swecl.c.
//
//	no t of rise or set found
func (state *State) calcMerTrans(jdUT float64, body int32, ephemeris int32, eventFlags int32, location []float64, starname []byte, tret []float64, message []byte) int32 {
	var (
		i         int32
		jdTT      float64 = jdUT + state.sweDeltatEx(jdUT, ephemeris, message)
		armc      float64
		armc0     float64
		arxc      float64
		x0        [6]float64
		x         [6]float64
		t         float64
		te        float64
		mdd       float64
		flags     int32 = ephemeris
		doFixstar int32 = int32(boolInt(starname != nil && int32(starname[0]) != '\000'))
	)
	flags &= flagJPL | flagSwiss | flagMoshier
	tret[0] = 0
	flags |= flagEquatorial | flagTopocentric
	armc0 = state.sweSidtime(jdUT) + location[0]/15
	if armc0 >= 24 {
		armc0 -= 24
	}
	if armc0 < 0 {
		armc0 += 24
	}
	armc0 *= 15
	if doFixstar != 0 {
		if state.sweFixstar(starname, jdTT, flags, x0[:], message) == -1 {
			return -1
		}
	} else if state.sweCalc(jdTT, body, flags, x0[:], message) == -1 {
		return -1
	}
	x[0] = x0[0]
	x[1] = x0[1]
	t = jdUT
	arxc = armc0
	if eventFlags&8 != 0 {
		arxc = state.sweDegnorm(arxc + 180)
	}
	for i = 0; i < 4; i++ {
		mdd = state.sweDegnorm(x[0] - arxc)
		if i > 0 && mdd > 180 {
			mdd -= 360
		}
		t += mdd / 361
		armc = state.sweSidtime(t) + location[0]/15
		if armc >= 24 {
			armc -= 24
		}
		if armc < 0 {
			armc += 24
		}
		armc *= 15
		arxc = armc
		if eventFlags&8 != 0 {
			arxc = state.sweDegnorm(arxc + 180)
		}
		if !(doFixstar != 0) {
			te = t + state.sweDeltatEx(t, ephemeris, message)
			if state.sweCalc(te, body, flags, x[:], message) == -1 {
				return -1
			}
		}
	}
	tret[0] = t
	return 0
}

// sweNodAps implements swe_nod_aps from swecl.c.
//
//	Pluto
func (state *State) sweNodAps(jdTT float64, body int32, flags int32, method int32, xnasc []float64, xndsc []float64, xperi []float64, xaphe []float64, message []byte) int32 {
	var (
		ij             int32
		i              int32
		j              int32
		iplx           int32
		bodyIndex      int32
		istart         int32
		iend           int32
		iflJ2000       int32
		daya           [1]float64
		plm            float64
		t              float64 = (jdTT - julianDayJ2000) / 36525
		dt             float64
		x              [6]float64
		xx             [24]float64
		xp             []float64
		observerVector [6]float64
		x2000          [6]float64
		xpos           [3][6]float64
		xnorm          [6]float64
		xposm          [6]float64
		xn             [3][6]float64
		xs             [3][6]float64
		xq             [3][6]float64
		xa             [3][6]float64
		xobs2          [6]float64
		x2             [6]float64
		xna            []float64
		xnd            []float64
		xpe            []float64
		xap            []float64
		incl           float64
		sema           float64
		ecce           float64
		parg           float64
		ea             float64
		vincl          float64
		vsema          float64
		vecce          float64
		pargx          float64
		eax            float64
		earthData      *planetData = &state.positions.data.planets[0]
		psbdp          *planetData = &state.positions.data.planets[10]
		pldat          planetData
		sunVector      []float64 = psbdp.stateVector[:]
		xear           []float64 = earthData.stateVector[:]
		ep             []float64
		Gmsm           float64
		dzmin          float64
		rxy            float64
		rxyz           float64
		fac            float64
		sgn            float64
		sinnode        float64
		cosnode        float64
		sinincl        float64
		cosincl        float64
		sinu           float64
		cosu           float64
		sinE           float64
		cosE           float64
		cosE2          float64
		uu             float64
		ny             float64
		ny2            float64
		c2             float64
		v2             float64
		pp             float64
		ro             float64
		ro2            float64
		rn             float64
		rn2            float64
		obliquityData  *obliquity
		isTrueNodaps   int32 = 0
		doAberr        int32 = int32(boolInt(!(flags&(flagTruePosition|flagNoAberration) != 0)))
		doDefl         int32 = int32(boolInt(flags&flagTruePosition == 0 && flags&flagNoDeflection == 0))
		doFocalPoint   int32 = method & 256
		ellipseIsBary  int32 = 0
		iflg0          int32
	)
	flags &= ^(256*1024 | 512*1024)
	if body == 10000+134340 {
		body = 9
	}
	xna = xx[:]
	xnd = xx[6:]
	xpe = xx[12:]
	xap = xx[18:]
	xpos[0][0] = 0
	state.swiForceAppPosEtc()
	method %= 256
	bodyIndex = body
	if body == 0 {
		bodyIndex = 14
	}
	if body == 1 {
		doDefl = 0
		if flags&flagHeliocentric == 0 {
			doAberr = 0
		}
	}
	iflg0 = flags&(flagJPL|flagSwiss|flagMoshier|flagNoNutation) | 256 | 16
	if bodyIndex != 1 {
		iflg0 |= 8
	}
	if body == 10 || body == 11 || body == 12 || body == 13 || body < 0 || body >= 23 && body <= 10000 {
		if message != nil {
			formatText(message, textBytes("nodes/apsides for planet %5.0f are not implemented"), float64(body))
		}
		if xnasc != nil {
			for i = 0; i <= 5; i++ {
				xnasc[i] = 0
			}
		}
		if xndsc != nil {
			for i = 0; i <= 5; i++ {
				xndsc[i] = 0
			}
		}
		if xaphe != nil {
			for i = 0; i <= 5; i++ {
				xaphe[i] = 0
			}
		}
		if xperi != nil {
			for i = 0; i <= 5; i++ {
				xperi[i] = 0
			}
		}
		return -1
	}
	for i = 0; i < 24; i++ {
		xx[i] = 0
	}
	if (method == 0 || method&1 != 0) && (body >= 0 && body <= 8 || body == 14) {
		if body == 1 {
			state.swiMeanLunarElements(jdTT, xna[:], xna[3:], xpe[:], xpe[3:])
			incl = 5.1453964
			vincl = 0
			ecce = 0.054900489
			vecce = 0
			sema = 0.0025695552897999903
			vsema = 0
		} else {
			iplx = eclipsesIplToElem[body]
			ep = eclipsesElIncl[iplx][:]
			incl = ep[0] + ep[1]*t + ep[2]*t*t + ep[3]*t*t*t
			vincl = ep[1] / 36525
			ep = eclipsesElSema[iplx][:]
			sema = ep[0] + ep[1]*t + ep[2]*t*t + ep[3]*t*t*t
			vsema = ep[1] / 36525
			ep = eclipsesElEcce[iplx][:]
			ecce = ep[0] + ep[1]*t + ep[2]*t*t + ep[3]*t*t*t
			vecce = ep[1] / 36525
			ep = eclipsesElNode[iplx][:]
			xna[0] = ep[0] + ep[1]*t + ep[2]*t*t + ep[3]*t*t*t
			xna[3] = ep[1] / 36525
			ep = eclipsesElPeri[iplx][:]
			xpe[0] = ep[0] + ep[1]*t + ep[2]*t*t + ep[3]*t*t*t
			xpe[3] = ep[1] / 36525
		}
		xnd[0] = state.sweDegnorm(xna[0] + 180)
		xnd[3] = xna[3]
		xpe[0] = state.sweDegnorm(xpe[0] - xna[0])
		parg = xpe[0]
		xpe[3] = state.sweDegnorm(xpe[0] + xpe[3] - xna[3])
		pargx = xpe[3]
		state.sweCotrans(xpe, xpe, -incl)
		state.sweCotrans(xpe[3:], xpe[3:], -incl-vincl)
		xpe[0] = state.sweDegnorm(xpe[0] + xna[0])
		xpe[3] = state.sweDegnorm(xpe[3] + xna[0] + xna[3])
		xpe[3] = state.sweDegnorm(xpe[3] - xpe[0])
		xpe[2] = sema * (1 - ecce)
		xpe[5] = (sema+vsema)*(1-ecce-vecce) - xpe[2]
		xap[0] = state.sweDegnorm(xpe[0] + 180)
		xap[1] = -xpe[1]
		xap[3] = xpe[3]
		xap[4] = -xpe[4]
		if doFocalPoint != 0 {
			xap[2] = sema * ecce * 2
			xap[5] = (sema+vsema)*(ecce+vecce)*2 - xap[2]
		} else {
			xap[2] = sema * (1 + ecce)
			xap[5] = (sema+vsema)*(1+ecce+vecce) - xap[2]
		}
		ea = math.Atan(math.Tan(-parg*radiansPerDegree/2)*math.Sqrt((1-ecce)/(1+ecce))) * 2
		eax = math.Atan(math.Tan(-pargx*radiansPerDegree/2)*math.Sqrt((1-ecce-vecce)/(1+ecce+vecce))) * 2
		xna[2] = sema * (math.Cos(ea) - ecce) / math.Cos(parg*radiansPerDegree)
		xna[5] = (sema + vsema) * (math.Cos(eax) - ecce - vecce) / math.Cos(pargx*radiansPerDegree)
		xna[5] -= xna[2]
		ea = math.Atan(math.Tan((180-parg)*radiansPerDegree/2)*math.Sqrt((1-ecce)/(1+ecce))) * 2
		eax = math.Atan(math.Tan((180-pargx)*radiansPerDegree/2)*math.Sqrt((1-ecce-vecce)/(1+ecce+vecce))) * 2
		xnd[2] = sema * (math.Cos(ea) - ecce) / math.Cos((180-parg)*radiansPerDegree)
		xnd[5] = (sema + vsema) * (math.Cos(eax) - ecce - vecce) / math.Cos((180-pargx)*radiansPerDegree)
		xnd[5] -= xnd[2]
		i = 0
		for xp = xx[:]; i < 4; {
			xp[0] *= radiansPerDegree
			xp[1] *= radiansPerDegree
			xp[3] *= radiansPerDegree
			xp[4] *= radiansPerDegree
			state.swiPolcartSp(xp, xp)
			i++
			xp = xp[6:]
		}
	} else {
		if state.sweCalc(jdTT, bodyIndex, iflg0, x[:], message) == -1 {
			return -1
		}
		iflJ2000 = flags&(flagJPL|flagSwiss|flagMoshier) | 32 | 2*1024 | 4*1024 | 16 | 64 | 256
		ellipseIsBary = 0
		if bodyIndex != 1 {
			if method&4 != 0 && x[2] > 6 {
				iflJ2000 |= 16 * 1024
				ellipseIsBary = 1
			} else {
				iflJ2000 |= 8
			}
		}
		if bodyIndex == 1 {
			dt = 0.0001
			dzmin = 1e-15
			Gmsm = 8.9970116861855104e-10
		} else {
			if bodyIndex >= 2 && bodyIndex <= 9 || bodyIndex == 14 {
				plm = 1 / eclipsesPlmass[eclipsesIplToElem[body]]
			} else {
				plm = 0
			}
			dt = 0.001 * x[2]
			dzmin = 1e-15 * dt / 0.0001
			Gmsm = 1.32712440017987e+20 * (1 + plm) / 1.49597870700e+11 / 1.49597870700e+11 / 1.49597870700e+11 * secondsPerDay * secondsPerDay
		}
		if flags&flagSpeed != 0 {
			istart = 0
			iend = 2
		} else {
			iend = 0
			istart = iend
			dt = 0
		}
		i = istart
		for t = jdTT - dt; i <= iend; {
			if istart == iend {
				t = jdTT
			}
			if state.sweCalc(t, bodyIndex, iflJ2000, xpos[i][:], message) == -1 {
				return -1
			}
			if bodyIndex == 14 {
				if state.sweCalc(t, 1, iflJ2000 & ^(16*1024|8), xposm[:], message) == -1 {
					return -1
				}
				for j = 0; j <= 5; j++ {
					xpos[i][j] += xposm[j] / 82.30055985272827
				}
			}
			state.swiPlanForOscElem(iflg0, t, xpos[i][:])
			i++
			t += dt
		}
		for i = istart; i <= iend; i++ {
			if math.Abs(xpos[i][5]) < dzmin {
				xpos[i][5] = dzmin
			}
			fac = xpos[i][2] / xpos[i][5]
			sgn = xpos[i][5] / math.Abs(xpos[i][5])
			for j = 0; j <= 2; j++ {
				xn[i][j] = (xpos[i][j] - fac*xpos[i][j+3]) * sgn
				xs[i][j] = -xn[i][j]
			}
		}
		for i = istart; i <= iend; i++ {
			rxy = math.Sqrt(xn[i][0]*xn[i][0] + xn[i][1]*xn[i][1])
			cosnode = xn[i][0] / rxy
			sinnode = xn[i][1] / rxy
			state.swiCrossProd(xpos[i][:], xpos[i][3:], xnorm[:])
			rxy = xnorm[0]*xnorm[0] + xnorm[1]*xnorm[1]
			c2 = rxy + xnorm[2]*xnorm[2]
			rxyz = math.Sqrt(c2)
			rxy = math.Sqrt(rxy)
			sinincl = rxy / rxyz
			cosincl = math.Sqrt(1 - sinincl*sinincl)
			if xnorm[2] < 0 {
				cosincl = -cosincl
			}
			cosu = xpos[i][0]*cosnode + xpos[i][1]*sinnode
			sinu = xpos[i][2] / sinincl
			uu = math.Atan2(sinu, cosu)
			rxyz = math.Sqrt(xpos[i][0]*xpos[i][0] + xpos[i][1]*xpos[i][1] + xpos[i][2]*xpos[i][2])
			v2 = xpos[i][3]*xpos[i][3] + xpos[i][3:][1]*xpos[i][3:][1] + xpos[i][3:][2]*xpos[i][3:][2]
			sema = 1 / (2/rxyz - v2/Gmsm)
			pp = c2 / Gmsm
			ecce = math.Sqrt(1 - pp/sema)
			cosE = 1 / ecce * (1 - rxyz/sema)
			sinE = 1 / ecce / math.Sqrt(sema*Gmsm) * (xpos[i][0]*xpos[i][3] + xpos[i][1]*xpos[i][3:][1] + xpos[i][2]*xpos[i][3:][2])
			ny = 2 * math.Atan(math.Sqrt((1+ecce)/(1-ecce))*sinE/(1+cosE))
			xq[i][0] = state.swiMod2PI(uu - ny)
			xq[i][1] = 0
			xq[i][2] = sema * (1 - ecce)
			state.swiPolcart(xq[i][:], xq[i][:])
			state.swiCoortrf2(xq[i][:], xq[i][:], -sinincl, cosincl)
			state.swiCartpol(xq[i][:], xq[i][:])
			xq[i][0] += math.Atan2(sinnode, cosnode)
			xa[i][0] = state.swiMod2PI(xq[i][0] + 3.14159265358979323846)
			xa[i][1] = -xq[i][1]
			if doFocalPoint != 0 {
				xa[i][2] = sema * ecce * 2
			} else {
				xa[i][2] = sema * (1 + ecce)
			}
			state.swiPolcart(xq[i][:], xq[i][:])
			state.swiPolcart(xa[i][:], xa[i][:])
			ny = state.swiMod2PI(ny - uu)
			ny2 = state.swiMod2PI(ny + 3.14159265358979323846)
			cosE = math.Cos(2 * math.Atan(math.Tan(ny/2)/math.Sqrt((1+ecce)/(1-ecce))))
			cosE2 = math.Cos(2 * math.Atan(math.Tan(ny2/2)/math.Sqrt((1+ecce)/(1-ecce))))
			rn = sema * (1 - ecce*cosE)
			rn2 = sema * (1 - ecce*cosE2)
			ro = math.Sqrt(xn[i][0]*xn[i][0] + xn[i][1]*xn[i][1] + xn[i][2]*xn[i][2])
			ro2 = math.Sqrt(xs[i][0]*xs[i][0] + xs[i][1]*xs[i][1] + xs[i][2]*xs[i][2])
			for j = 0; j <= 2; j++ {
				xn[i][j] *= rn / ro
				xs[i][j] *= rn2 / ro2
			}
		}
		for i = 0; i <= 2; i++ {
			if flags&flagSpeed != 0 {
				xpe[i] = xq[1][i]
				xpe[i+3] = (xq[2][i] - xq[0][i]) / dt / 2
				xap[i] = xa[1][i]
				xap[i+3] = (xa[2][i] - xa[0][i]) / dt / 2
				xna[i] = xn[1][i]
				xna[i+3] = (xn[2][i] - xn[0][i]) / dt / 2
				xnd[i] = xs[1][i]
				xnd[i+3] = (xs[2][i] - xs[0][i]) / dt / 2
			} else {
				xpe[i] = xq[0][i]
				xpe[i+3] = 0
				xap[i] = xa[0][i]
				xap[i+3] = 0
				xna[i] = xn[0][i]
				xna[i+3] = 0
				xnd[i] = xs[0][i]
				xnd[i+3] = 0
			}
		}
		isTrueNodaps = 1
	}
	if bodyIndex == 1 && flags&(flagHeliocentric|flagBarycentric) != 0 {
		state.swiForceAppPosEtc()
		if state.sweCalc(jdTT, 0, iflg0, x[:], message) == -1 {
			return -1
		}
	} else if state.sweCalc(jdTT, bodyIndex, iflg0|flags&flagTopocentric, x[:], message) == -1 {
		return -1
	}
	if flags&flagTopocentric != 0 {
		if state.swiGetObserver(jdTT, flags, 0, observerVector[:], message) != 0 {
			return -1
		}
	} else {
		for i = 0; i <= 5; i++ {
			observerVector[i] = 0
		}
	}
	if flags&(flagHeliocentric|flagBarycentric) != 0 {
		if flags&flagHeliocentric != 0 && flags&flagMoshier == 0 {
			for i = 0; i <= 5; i++ {
				observerVector[i] = sunVector[i]
			}
		}
	} else if body == 0 && flags&flagMoshier == 0 {
		for i = 0; i <= 5; i++ {
			observerVector[i] = sunVector[i]
		}
	} else {
		for i = 0; i <= 5; i++ {
			observerVector[i] += xear[i]
		}
	}
	if flags&flagJ2000 != 0 {
		obliquityData = &state.positions.data.obliquityJ2000
	} else {
		obliquityData = &state.positions.data.obliquity
	}
	ij = 0
	for xp = xx[:]; ij < 4; {
		if bodyIndex == 14 && ij <= 1 {
			for i = 0; i <= 5; i++ {
				xp[i] = 0
			}
			ij++
			xp = xp[6:]
			continue
		}
		if isTrueNodaps != 0 && flags&flagNoNutation == 0 {
			state.swiCoortrf2(xp, xp, -state.positions.data.nutation.sine, state.positions.data.nutation.cosine)
			if flags&flagSpeed != 0 {
				state.swiCoortrf2(xp[3:], xp[3:], -state.positions.data.nutation.sine, state.positions.data.nutation.cosine)
			}
		}
		state.swiCoortrf2(xp, xp, -obliquityData.sine, obliquityData.cosine)
		state.swiCoortrf2(xp[3:], xp[3:], -obliquityData.sine, obliquityData.cosine)
		if isTrueNodaps != 0 {
			if flags&flagNoNutation == 0 {
				state.swiNutate(xp, flags, 1)
			}
		}
		state.swiPrecess(xp, jdTT, flags, 1)
		if flags&flagSpeed != 0 {
			state.swiPrecessSpeed(xp, jdTT, flags, 1)
		}
		if bodyIndex == 1 {
			for i = 0; i <= 5; i++ {
				xp[i] += xear[i]
			}
		} else if flags&flagMoshier == 0 && !(ellipseIsBary != 0) {
			for j = 0; j <= 5; j++ {
				xp[j] += sunVector[j]
			}
		}
		for j = 0; j <= 5; j++ {
			xp[j] -= observerVector[j]
		}
		if body == 0 && !(flags&(flagHeliocentric|flagBarycentric) != 0) {
			for j = 0; j <= 5; j++ {
				xp[j] = -xp[j]
			}
		}
		dt = math.Sqrt(xp[0]*xp[0]+xp[1]*xp[1]+xp[2]*xp[2]) * 1.49597870700e+11 / 2.99792458e+8 / secondsPerDay
		if doDefl != 0 {
			state.swiDeflectLight(xp, dt, flags)
		}
		if doAberr != 0 {
			state.swiAberrLight(xp, observerVector[:], flags)
			if flags&flagSpeed != 0 {
				if state.sweCalc(jdTT-dt, bodyIndex, iflg0|flags&flagTopocentric, x2[:], message) == -1 {
					return -1
				}
				if flags&flagTopocentric != 0 {
					for i = 0; i <= 5; i++ {
						xobs2[i] = state.positions.data.observer.stateVector[i]
					}
				} else {
					for i = 0; i <= 5; i++ {
						xobs2[i] = 0
					}
				}
				if flags&(flagHeliocentric|flagBarycentric) != 0 {
					if flags&flagHeliocentric != 0 && flags&flagMoshier == 0 {
						for i = 0; i <= 5; i++ {
							xobs2[i] = sunVector[i]
						}
					}
				} else if body == 0 && flags&flagMoshier == 0 {
					for i = 0; i <= 5; i++ {
						xobs2[i] = sunVector[i]
					}
				} else {
					for i = 0; i <= 5; i++ {
						xobs2[i] += xear[i]
					}
				}
				for i = 3; i <= 5; i++ {
					xp[i] += observerVector[i] - xobs2[i]
				}
				if state.sweCalc(jdTT, 0, iflg0|flags&flagTopocentric, x2[:], message) == -1 {
					return -1
				}
			}
		}
		for j = 0; j <= 5; j++ {
			x2000[j] = xp[j]
		}
		if flags&flagJ2000 == 0 {
			state.swiPrecess(xp, jdTT, flags, -1)
			if flags&flagSpeed != 0 {
				state.swiPrecessSpeed(xp, jdTT, flags, -1)
			}
		}
		if flags&flagNoNutation == 0 {
			state.swiNutate(xp, flags, 0)
		}
		for j = 0; j <= 5; j++ {
			pldat.coordinates[18+j] = xp[j]
		}
		state.swiCoortrf2(xp, xp, obliquityData.sine, obliquityData.cosine)
		if flags&flagSpeed != 0 {
			state.swiCoortrf2(xp[3:], xp[3:], obliquityData.sine, obliquityData.cosine)
		}
		if flags&flagNoNutation == 0 {
			state.swiCoortrf2(xp, xp, state.positions.data.nutation.sine, state.positions.data.nutation.cosine)
			if flags&flagSpeed != 0 {
				state.swiCoortrf2(xp[3:], xp[3:], state.positions.data.nutation.sine, state.positions.data.nutation.cosine)
			}
		}
		for j = 0; j <= 5; j++ {
			pldat.coordinates[6+j] = xp[j]
		}
		if flags&flagSidereal != 0 {
			if state.positions.data.sidereal.mode&256 != 0 {
				if state.swiTropRa2sidLon(x2000[:], pldat.coordinates[6:], pldat.coordinates[18:], flags) != 0 {
					return -1
				}
			} else if state.positions.data.sidereal.mode&512 != 0 {
				if state.swiTropRa2sidLonSosy(x2000[:], pldat.coordinates[6:], flags) != 0 {
					return -1
				}
			} else {
				state.swiCartpolSp(pldat.coordinates[6:], pldat.coordinates[:])
				if state.swiGetAyanamsaEx(jdTT, flags, daya[:], message) == -1 {
					return -1
				}
				pldat.coordinates[0] -= daya[0] * radiansPerDegree
				state.swiPolcartSp(pldat.coordinates[:], pldat.coordinates[6:])
			}
		}
		if flags&flagCartesian != 0 && flags&flagEquatorial != 0 {
			for j = 0; j <= 5; j++ {
				xp[j] = pldat.coordinates[18+j]
			}
			ij++
			xp = xp[6:]
			continue
		}
		if flags&flagCartesian != 0 {
			for j = 0; j <= 5; j++ {
				xp[j] = pldat.coordinates[6+j]
			}
			ij++
			xp = xp[6:]
			continue
		}
		state.swiCartpolSp(pldat.coordinates[18:], pldat.coordinates[12:])
		state.swiCartpolSp(pldat.coordinates[6:], pldat.coordinates[:])
		if flags&flagRadians == 0 {
			for j = 0; j < 2; j++ {
				pldat.coordinates[j] *= degreesPerRadian
				pldat.coordinates[j+3] *= degreesPerRadian
				pldat.coordinates[j+12] *= degreesPerRadian
				pldat.coordinates[j+15] *= degreesPerRadian
			}
		}
		if flags&flagEquatorial != 0 {
			for j = 0; j <= 5; j++ {
				xp[j] = pldat.coordinates[12+j]
			}
			ij++
			xp = xp[6:]
			continue
		} else {
			for j = 0; j <= 5; j++ {
				xp[j] = pldat.coordinates[j]
			}
			ij++
			xp = xp[6:]
			continue
		}
	}
	for i = 0; i <= 5; i++ {
		if i > 2 && flags&flagSpeed == 0 {
			xap[i] = 0
			xpe[i] = xap[i]
			xnd[i] = xpe[i]
			xna[i] = xnd[i]
		}
		if xnasc != nil {
			xnasc[i] = xna[i]
		}
		if xndsc != nil {
			xndsc[i] = xnd[i]
		}
		if xperi != nil {
			xperi[i] = xpe[i]
		}
		if xaphe != nil {
			xaphe[i] = xap[i]
		}
	}
	return 0
}

// sweNodApsUt implements swe_nod_aps_ut from swecl.c.
func (state *State) sweNodApsUt(jdUT float64, body int32, flags int32, method int32, xnasc []float64, xndsc []float64, xperi []float64, xaphe []float64, message []byte) int32 {
	return state.sweNodAps(jdUT+state.sweDeltatEx(jdUT, flags, message), body, flags, method, xnasc, xndsc, xperi, xaphe, message)
}

// getGmsm implements get_gmsm from swecl.c.
//
//	Pluto
func (state *State) getGmsm(jdTT float64, body int32, flags int32, r float64, gmsm []float64, message []byte) int32 {
	var (
		j         int32
		Gmsm      float64 = 0
		plm       float64 = 0
		x         [6]float64
		iflJ2000p int32 = flags&(flagJPL|flagSwiss|flagMoshier|flagHeliocentric|flagBarycentric) | 32 | 16 | 64
	)
	if !(iflJ2000p&(8|16*1024) != 0) {
		iflJ2000p |= 8
	}
	if body == 1 {
		Gmsm = 8.9970116861855104e-10
	} else if body >= 2 && body <= 9 || body == 14 {
		plm = 0
		if flags&flagTopocentric != 0 {
			if body == 14 {
				plm = 1.0 / eclipsesPlmass[eclipsesIplToElem[body]]
				plm += 1.0 / eclipsesPlmass[eclipsesIplToElem[3]]
				plm += 1.0 / eclipsesPlmass[eclipsesIplToElem[2]]
			} else {
				for j = body; j >= 2; j-- {
					plm += 1.0 / eclipsesPlmass[eclipsesIplToElem[j]]
				}
				if body >= 4 {
					plm += 1.0 / eclipsesPlmass[eclipsesIplToElem[14]]
				}
			}
		} else {
			plm = 1.0 / eclipsesPlmass[eclipsesIplToElem[body]]
		}
		Gmsm = 1.32712440017987e+20 * (1 + plm) / 1.49597870700e+11 / 1.49597870700e+11 / 1.49597870700e+11 * secondsPerDay * secondsPerDay
	} else {
		plm = 0
		if flags&flagTopocentric != 0 {
			for j = 2; j <= 9; j++ {
				if state.sweCalc(jdTT, j, iflJ2000p, x[:], message) == -1 {
					return -1
				}
				if r > x[2] {
					plm += 1.0 / eclipsesPlmass[eclipsesIplToElem[j]]
				}
			}
			if state.sweCalc(jdTT, 14, iflJ2000p, x[:], message) == -1 {
				return -1
			}
			if r > x[2] {
				plm += 1.0 / eclipsesPlmass[eclipsesIplToElem[14]]
			}
		}
		Gmsm = 1.32712440017987e+20 * (1 + plm) / 1.49597870700e+11 / 1.49597870700e+11 / 1.49597870700e+11 * secondsPerDay * secondsPerDay
	}
	gmsm[0] = Gmsm
	return 0
}

// sweGetOrbitalElements implements swe_get_orbital_elements from swecl.c.
//
//	Function calculates osculating orbital elements (Kepler elements) of a planet
//
// or asteroid or the Earth-Moon barycentre.
// The function returns error if called for the Sun, the lunar nodes, or the apsides.
// Input parameters:
// tjd_et	Julian day number, in TT (ET)
// ipl		object number
// iflag	can contain
//   - ephemeris flag: SEFLG_JPLEPH, SEFLG_SWIEPH, SEFLG_MOSEPH
//   - center:
//     Sun:            SEFLG_HELCTR (assumed as default) or
//     SS Barycentre:  SEFLG_BARYCTR (rel. to solar system barycentre)
//     (only possible for planets beyond Jupiter)
//     For elements of the Moon, the calculation is geocentric.
//   - sum all masses inside the orbit to be computed (method
//     of Astronomical Almanac):
//     SEFLG_ORBEL_AA
//   - reference ecliptic: SEFLG_J2000;
//     if missing, mean ecliptic of date is chosen (still not implemented)
//
// output parameters:
// dret[]       array of return values, declare as dret[50]
// dret[0]      semimajor axis (a)
// dret[1]      eccentricity (e)
// dret[2]      inclination (in)
// dret[3]      longitude of ascending node (upper case omega OM)
// dret[4]      argument of periapsis (lower case omega om)
// dret[5]      longitude of periapsis (peri)
// dret[6]      mean anomaly at epoch (M0)
// dret[7]      true anomaly at epoch (N0)
// dret[8]      eccentric anomaly at epoch (E0)
// dret[9]      mean longitude at epoch (LM)
// dret[10]     sidereal orbital period in tropical years
// dret[11]     mean daily motion
// dret[12]     tropical period in years
// dret[13]     synodic period in days,
//
//	negative, if inner planet (Venus, Mercury, Aten asteroids) or Moon
//
// dret[14]     time of perihelion passage
// dret[15]     perihelion distance
// dret[16]     aphelion distance
func (state *State) sweGetOrbitalElements(jdTT float64, body int32, flags int32, dret []float64, message []byte) int32 {
	var (
		j         int32
		x         [6]float64
		xpos      [6]float64
		xposm     [6]float64
		xn        [6]float64
		xs        [6]float64
		xnorm     [6]float64
		xq        [6]float64
		xa        [6]float64
		iflJ2000  int32 = flags&(flagJPL|flagSwiss|flagMoshier) | 32 | 4*1024 | 16 | 64 | 256
		iflJ2000p int32 = flags&(flagJPL|flagSwiss|flagMoshier) | 32 | 16 | 64 | 256
		Gmsm      [1]float64
		iflg0     int32 = 0
		fac       float64
		sgn       float64
		rxy       float64
		rxyz      float64
		c2        float64
		cosnode   float64
		sinnode   float64
		incl      float64
		node      float64
		parg      float64
		peri      float64
		mlon      float64
		csid      float64
		ctro      float64
		csyn      float64
		dmot      float64
		pa        float64
		ytrop     float64
		ysid      float64
		T         float64
		T2        float64
		T3        float64
		T4        float64
		T5        float64
		sinincl   float64
		cosincl   float64
		cosu      float64
		sinu      float64
		uu        float64
		eanom     float64
		tanom     float64
		manom     float64
		v2        float64
		sema      float64
		pp        float64
		ecce      float64
		cosE      float64
		sinE      float64
		ny        float64
		ny2       float64
		rn        float64
		rn2       float64
		ro        float64
		ro2       float64
		cosE2     float64
		r         float64
		ecce2     float64
	)
	if body <= 0 || body == 10 || body == 11 || body == 12 || body == 13 || body == 21 || body == 22 {
		if message != nil {
			formatText(message, textBytes("error in swe_get_orbital_elements(): object %d not valid\n"), body)
		}
		return -1
	}
	if body != 1 {
		iflg0 |= 8
	}
	if state.sweCalc(jdTT, body, iflJ2000p, x[:], message) == -1 {
		return -1
	}
	r = x[2]
	if body != 1 {
		if flags&flagBarycentric != 0 && r > 6 {
			iflJ2000 |= 16 * 1024
		} else {
			iflJ2000 |= 8
		}
	}
	if state.getGmsm(jdTT, body, flags, r, Gmsm[:], message) != 0 {
		return -1
	}
	if state.sweCalc(jdTT, body, iflJ2000, xpos[:], message) == -1 {
		return -1
	}
	if body == 14 {
		if state.sweCalc(jdTT, 1, iflJ2000 & ^(16*1024|8), xposm[:], message) == -1 {
			return -1
		}
		for j = 0; j <= 5; j++ {
			xpos[j] += xposm[j] / 82.30055985272827
		}
	}
	fac = xpos[2] / xpos[5]
	sgn = xpos[5] / math.Abs(xpos[5])
	for j = 0; j <= 2; j++ {
		xn[j] = (xpos[j] - fac*xpos[j+3]) * sgn
		xs[j] = -xn[j]
	}
	rxy = math.Sqrt(xn[0]*xn[0] + xn[1]*xn[1])
	cosnode = xn[0] / rxy
	sinnode = xn[1] / rxy
	state.swiCrossProd(xpos[:], xpos[3:], xnorm[:])
	rxy = xnorm[0]*xnorm[0] + xnorm[1]*xnorm[1]
	c2 = rxy + xnorm[2]*xnorm[2]
	rxyz = math.Sqrt(c2)
	rxy = math.Sqrt(rxy)
	sinincl = rxy / rxyz
	cosincl = math.Sqrt(1 - sinincl*sinincl)
	if xnorm[2] < 0 {
		cosincl = -cosincl
	}
	incl = math.Acos(cosincl) * degreesPerRadian
	cosu = xpos[0]*cosnode + xpos[1]*sinnode
	sinu = xpos[2] / sinincl
	uu = math.Atan2(sinu, cosu)
	rxyz = math.Sqrt(xpos[0]*xpos[0] + xpos[1]*xpos[1] + xpos[2]*xpos[2])
	v2 = xpos[3]*xpos[3] + xpos[3:][1]*xpos[3:][1] + xpos[3:][2]*xpos[3:][2]
	sema = 1.0 / (2.0/rxyz - v2/Gmsm[0])
	pp = c2 / Gmsm[0]
	ecce = pp / sema
	if ecce > 1 {
		ecce = 1
	}
	ecce = math.Sqrt(1 - ecce)
	ecce2 = ecce
	if ecce2 == 0 {
		ecce2 = 0.0000000001
	}
	cosE = 1 / ecce2 * (1 - rxyz/sema)
	sinE = 1 / ecce2 / math.Sqrt(sema*Gmsm[0]) * (xpos[0]*xpos[3] + xpos[1]*xpos[3:][1] + xpos[2]*xpos[3:][2])
	eanom = state.sweDegnorm(math.Atan2(sinE, cosE) * degreesPerRadian)
	ny = 2 * math.Atan(math.Sqrt((1+ecce)/(1-ecce))*sinE/(1+cosE))
	tanom = state.sweDegnorm(ny * degreesPerRadian)
	if eanom > 180 && tanom < 180 {
		tanom += 180
	}
	if eanom < 180 && tanom > 180 {
		tanom -= 180
	}
	manom = state.sweDegnorm(eanom - ecce*degreesPerRadian*math.Sin(eanom*radiansPerDegree))
	xq[0] = state.swiMod2PI(uu - ny)
	parg = xq[0] * degreesPerRadian
	xq[1] = 0
	xq[2] = sema * (1 - ecce)
	state.swiPolcart(xq[:], xq[:])
	state.swiCoortrf2(xq[:], xq[:], -sinincl, cosincl)
	state.swiCartpol(xq[:], xq[:])
	xq[0] += math.Atan2(sinnode, cosnode)
	xa[0] = state.swiMod2PI(xq[0] + 3.14159265358979323846)
	xa[1] = -xq[1]
	xa[2] = sema * (1 + ecce)
	state.swiPolcart(xq[:], xq[:])
	state.swiPolcart(xa[:], xa[:])
	ny = state.swiMod2PI(ny - uu)
	ny2 = state.swiMod2PI(ny + 3.14159265358979323846)
	cosE = math.Cos(2 * math.Atan(math.Tan(ny/2)/math.Sqrt((1+ecce)/(1-ecce))))
	cosE2 = math.Cos(2 * math.Atan(math.Tan(ny2/2)/math.Sqrt((1+ecce)/(1-ecce))))
	rn = sema * (1 - ecce*cosE)
	rn2 = sema * (1 - ecce*cosE2)
	ro = math.Sqrt(xn[0]*xn[0] + xn[1]*xn[1] + xn[2]*xn[2])
	ro2 = math.Sqrt(xs[0]*xs[0] + xs[1]*xs[1] + xs[2]*xs[2])
	for j = 0; j <= 2; j++ {
		xn[j] *= rn / ro
		xs[j] *= rn2 / ro2
	}
	state.swiCartpol(xn[:], xn[:])
	state.swiCartpol(xq[:], xq[:])
	node = xn[0] * degreesPerRadian
	peri = state.sweDegnorm(node + parg)
	mlon = state.sweDegnorm(manom + peri)
	csid = sema * math.Sqrt(sema)
	if body == 1 {
		var semam float64 = sema * 1.49597870700e+11 / 383397772.5
		csid = semam * math.Sqrt(semam)
		csid *= 0.074801325227015963
	}
	dmot = 0.9856076686 / csid
	csid *= 1.0000387989132329
	T = (jdTT - julianDayJ2000) / 365250.0
	T2 = T * T
	T3 = T2 * T
	T4 = T3 * T
	T5 = T4 * T
	pa = (50288.200 + 222.4045*T + 0.2095*T2 - 0.9408*T3 - 0.0090*T4 + 0.0010*T5) / 3600.0 / 365250.0
	ysid = (1295977422.83429 - 4.0881999999999996*T - 0.015690000000000003*T*T) / 3600.0 / 365250.0
	ysid = 360.0 / ysid
	ytrop = (1296027711.03429 + 218.31618*T + 0.21620999999999999*T2 - 0.94120000000000004*T3 - 0.0089999999999999993*T4 + 0.0012000000000000001*T5) / 3600.0 / 365250.0
	ytrop = 360.0 / ytrop
	ctro = 360.0 / (dmot + pa) / 365.242189
	ctro *= ysid / ytrop
	if body == 14 {
		csyn = 0
	} else {
		csyn = 360.0 / (0.9856076686 - dmot)
	}
	dret[0] = sema
	dret[1] = ecce
	dret[2] = incl
	dret[3] = node
	dret[4] = parg
	dret[5] = peri
	dret[6] = manom
	dret[7] = tanom
	dret[8] = eanom
	dret[9] = mlon
	dret[10] = csid
	dret[11] = dmot
	dret[12] = ctro
	dret[13] = csyn
	dret[14] = jdTT - dret[6]/dmot
	dret[15] = sema * (1 - ecce)
	dret[16] = sema * (1 + ecce)
	return 0
}

// oscGetOrbitConstants implements osc_get_orbit_constants from swecl.c.
func (state *State) oscGetOrbitConstants(dp []float64, pqr []float64) {
	var (
		sema    float64 = dp[0]
		ecce    float64 = dp[1]
		incl    float64 = dp[2]
		node    float64 = dp[3]
		parg    float64 = dp[4]
		cosnode float64 = math.Cos(node * radiansPerDegree)
		sinnode float64 = math.Sin(node * radiansPerDegree)
		cosincl float64 = math.Cos(incl * radiansPerDegree)
		sinincl float64 = math.Sin(incl * radiansPerDegree)
		cosparg float64 = math.Cos(parg * radiansPerDegree)
		sinparg float64 = math.Sin(parg * radiansPerDegree)
		fac     float64 = math.Sqrt((1 - ecce) * (1 + ecce))
	)
	pqr[0] = cosparg*cosnode - sinparg*cosincl*sinnode
	pqr[1] = -sinparg*cosnode - cosparg*cosincl*sinnode
	pqr[2] = sinincl * sinnode
	pqr[3] = cosparg*sinnode + sinparg*cosincl*cosnode
	pqr[4] = -sinparg*sinnode + cosparg*cosincl*cosnode
	pqr[5] = -sinincl * cosnode
	pqr[6] = sinparg * sinincl
	pqr[7] = cosparg * sinincl
	pqr[8] = cosincl
	pqr[9] = sema
	pqr[10] = ecce
	pqr[11] = fac
}

// oscGetEclPos implements osc_get_ecl_pos from swecl.c.
func (state *State) oscGetEclPos(ean float64, pqr []float64, xp []float64) {
	var (
		x    [2]float64
		cose float64 = math.Cos(ean * radiansPerDegree)
		sine float64 = math.Sin(ean * radiansPerDegree)
		sema float64 = pqr[9]
		ecce float64 = pqr[10]
		fac  float64 = pqr[11]
	)
	x[0] = sema * (cose - ecce)
	x[1] = sema * fac * sine
	xp[0] = pqr[0]*x[0] + pqr[1]*x[1]
	xp[1] = pqr[3]*x[0] + pqr[4]*x[1]
	xp[2] = pqr[6]*x[0] + pqr[7]*x[1]
}

// getDistFrom2Vectors implements get_dist_from_2_vectors from swecl.c.
//
//	transformation to ecliptic
func (state *State) getDistFrom2Vectors(x1 []float64, x2 []float64) float64 {
	var (
		r0 float64
		r1 float64
		r2 float64
	)
	r0 = x1[0] - x2[0]
	r1 = x1[1] - x2[1]
	r2 = x1[2] - x2[2]
	return math.Sqrt(r0*r0 + r1*r1 + r2*r2)
}

// oscIterateMaxDist implements osc_iterate_max_dist from swecl.c.
func (state *State) oscIterateMaxDist(ean float64, pqr []float64, xa []float64, xb []float64, deanopt []float64, drmax []float64, highPrec int32) {
	var (
		i        int32
		r        float64
		rmax     float64
		eansv    float64 = 0
		dstep    float64
		dstepMin float64 = 1
	)
	if highPrec != 0 {
		dstepMin = 0.000001
	}
	ean = 0
	state.oscGetEclPos(ean, pqr, xa)
	r = state.getDistFrom2Vectors(xb, xa)
	rmax = r
	for dstep = 1; dstep >= dstepMin; {
		for i = 0; i < 2; i++ {
			for r >= rmax {
				eansv = ean
				if i == 0 {
					ean += dstep
				} else {
					ean -= dstep
				}
				state.oscGetEclPos(ean, pqr, xa)
				r = state.getDistFrom2Vectors(xb, xa)
				if r > rmax {
					rmax = r
				}
			}
			ean = eansv
			r = rmax
		}
		ean = eansv
		r = rmax
		dstep /= 10
	}
	drmax[0] = rmax
	deanopt[0] = eansv
}

// oscIterateMinDist implements osc_iterate_min_dist from swecl.c.
func (state *State) oscIterateMinDist(ean float64, pqr []float64, xa []float64, xb []float64, deanopt []float64, drmin []float64, highPrec int32) {
	var (
		i        int32
		r        float64
		rmin     float64
		eansv    float64 = 0
		dstep    float64
		dstepMin float64 = 1
	)
	if highPrec != 0 {
		dstepMin = 0.000001
	}
	ean = 0
	state.oscGetEclPos(ean, pqr, xa)
	r = state.getDistFrom2Vectors(xb, xa)
	rmin = r
	for dstep = 1; dstep >= dstepMin; {
		for i = 0; i < 2; i++ {
			for r <= rmin {
				eansv = ean
				if i == 0 {
					ean += dstep
				} else {
					ean -= dstep
				}
				state.oscGetEclPos(ean, pqr, xa)
				r = state.getDistFrom2Vectors(xb, xa)
				if r < rmin {
					rmin = r
				}
			}
			ean = eansv
			r = rmin
		}
		ean = eansv
		r = rmin
		dstep /= 10
	}
	drmin[0] = rmin
	deanopt[0] = eansv
}

// orbitMaxMinTrueDistanceHelio implements orbit_max_min_true_distance_helio from swecl.c.
//
//	function calculates maximum distance, minimum distance and true distance between
//
// the Sun and the Earth-Moon barycentre. Maximum and minimum distance are derived
// from Kepler elements.
func (state *State) orbitMaxMinTrueDistanceHelio(jdTT float64, body int32, flags int32, dmax []float64, dmin []float64, dtrue []float64, message []byte) int32 {
	var (
		xinner    [3]float64
		pqri      [20]float64
		eani      float64
		de        [50]float64
		retval    int32
		bodyIndex int32 = body
		iflagi    int32 = flags & (flagJPL | flagSwiss | flagMoshier | flagHeliocentric | flagBarycentric)
	)
	if body == 0 {
		bodyIndex = 14
	}
	if retval = state.sweGetOrbitalElements(jdTT, bodyIndex, iflagi, de[:], message); retval == -1 {
		return -1
	}
	dmax[0] = de[16]
	dmin[0] = de[15]
	state.oscGetOrbitConstants(de[:], pqri[:])
	eani = de[8]
	state.oscGetEclPos(eani, pqri[:], xinner[:])
	dtrue[0] = math.Sqrt(xinner[0]*xinner[0] + xinner[1]*xinner[1] + xinner[2]*xinner[2])
	return retval
}

func (state *State) sweOrbitMaxMinTrueDistance(jdTT float64, body int32, flags int32, dmax []float64, dmin []float64, dtrue []float64, message []byte) int32 {
	var (
		i         int32
		j         int32
		k         int32
		retval    int32
		iflagi    int32 = flags & (flagJPL | flagSwiss | flagMoshier | flagHeliocentric | flagBarycentric)
		dp        [50]float64
		de        [50]float64
		xouter    [3]float64
		xinner    [3]float64
		maxXouter [3]float64
		maxXinner [3]float64
		minXouter [3]float64
		minXinner [3]float64
		pqro      [20]float64
		pqri      [20]float64
		eano      [1]float64
		eani      [1]float64
		douter    []float64
		dinner    []float64
		r         float64
		rtrue     float64
		rmax      [1]float64 = [1]float64{0}
		rmin      [1]float64 = [1]float64{100000000}
		rminsv    float64    = 0
		rmaxsv    float64    = 0
		minEanisv float64    = 0
		minEanosv float64    = 0
		maxEanisv float64    = 0
		maxEanosv float64    = 0
		ncnt      int32
		dstep     float64
		nitermax  float64 = 300
	)
	if body == 0 || body == 1 || iflagi&(8|16*1024) != 0 {
		retval = state.orbitMaxMinTrueDistanceHelio(jdTT, body, iflagi, dmax, dmin, dtrue, message)
		return retval
	}
	if retval = state.sweGetOrbitalElements(jdTT, body, iflagi, dp[:], message); retval == -1 {
		return -1
	}
	if retval = state.sweGetOrbitalElements(jdTT, 14, iflagi, de[:], message); retval == -1 {
		return -1
	}
	if de[0] > dp[0] {
		douter = de[:]
		dinner = dp[:]
	} else {
		douter = dp[:]
		dinner = de[:]
	}
	state.oscGetOrbitConstants(douter, pqro[:])
	state.oscGetOrbitConstants(dinner, pqri[:])
	eano[0] = douter[8]
	eani[0] = dinner[8]
	state.oscGetEclPos(eano[0], pqro[:], xouter[:])
	state.oscGetEclPos(eani[0], pqri[:], xinner[:])
	rtrue = state.getDistFrom2Vectors(xouter[:], xinner[:])
	ncnt = 182
	dstep = 2
	for i = 0; i < 3; i++ {
		maxXouter[i] = 0
		maxXinner[i] = 0
		minXouter[i] = 0
		minXinner[i] = 0
	}
	for j = 0; j < ncnt; j++ {
		eano[0] = float64(j) * dstep
		state.oscGetEclPos(eano[0], pqro[:], xouter[:])
		for i = 0; i < ncnt; i++ {
			eani[0] = float64(i)
			state.oscGetEclPos(eani[0], pqri[:], xinner[:])
			r = state.getDistFrom2Vectors(xouter[:], xinner[:])
			if r > rmax[0] {
				rmax[0] = r
				maxEanisv = eani[0]
				maxEanosv = eano[0]
				for k = 0; k < 3; k++ {
					maxXouter[k] = xouter[k]
					maxXinner[k] = xinner[k]
				}
			}
			if r < rmin[0] {
				rmin[0] = r
				minEanisv = eani[0]
				minEanosv = eano[0]
				for k = 0; k < 3; k++ {
					minXouter[k] = xouter[k]
					minXinner[k] = xinner[k]
				}
			}
		}
	}
	eani[0] = maxEanisv
	eano[0] = maxEanosv
	for k = 0; k < 3; k++ {
		xouter[k] = maxXouter[k]
		xinner[k] = maxXinner[k]
	}
	for k = 0; float64(k) <= nitermax; k++ {
		state.oscIterateMaxDist(eani[0], pqri[:], xinner[:], xouter[:], eani[:], rmax[:], 1)
		state.oscIterateMaxDist(eano[0], pqro[:], xouter[:], xinner[:], eano[:], rmax[:], 1)
		if k > 0 && math.Abs(rmax[0]-rmaxsv) < 0.00000001 {
			break
		}
		rmaxsv = rmax[0]
	}
	eani[0] = minEanisv
	eano[0] = minEanosv
	for k = 0; k < 3; k++ {
		xouter[k] = minXouter[k]
		xinner[k] = minXinner[k]
	}
	for k = 0; float64(k) <= nitermax; k++ {
		state.oscIterateMinDist(eani[0], pqri[:], xinner[:], xouter[:], eani[:], rmin[:], 1)
		state.oscIterateMinDist(eano[0], pqro[:], xouter[:], xinner[:], eano[:], rmin[:], 1)
		if k > 0 && math.Abs(rmin[0]-rminsv) < 0.00000001 {
			break
		}
		rminsv = rmin[0]
	}
	dmax[0] = rmax[0]
	dmin[0] = rmin[0]
	dtrue[0] = rtrue
	return retval
}

func (state *State) sweGauquelinSector(tUt float64, body int32, starname []byte, flags int32, imeth int32, location []float64, pressure float64, temperature float64, dgsect []float64, message []byte) int32 {
	var (
		riseFound    int32 = 1
		setFound     int32 = 1
		retval       int32
		tret         [3]float64
		tEt          float64
		t            float64
		x0           [6]float64
		eps          float64
		nutlo        [2]float64
		armc         float64
		ephemeris    int32 = flags & (flagJPL | flagSwiss | flagMoshier)
		doFixstar    int32 = int32(boolInt(starname != nil && int32(starname[0]) != '\000'))
		risemeth     int32 = 0
		aboveHorizon int32 = 0
	)
	if imeth < 0 || imeth > 5 {
		if message != nil {
			formatText(message, textBytes("invalid method: %d"), imeth)
		}
		return -1
	}
	if body == 10000+134340 {
		body = 9
	}
	if imeth == 0 || imeth == 1 {
		tEt = tUt + state.sweDeltatEx(tUt, flags, message)
		eps = state.swiEpsiln(tEt, flags) * degreesPerRadian
		state.swiNutation(tEt, flags, nutlo[:])
		nutlo[0] *= degreesPerRadian
		nutlo[1] *= degreesPerRadian
		armc = state.sweDegnorm(state.sweSidtime0(tUt, eps+nutlo[1], nutlo[0])*15 + location[0])
		if doFixstar != 0 {
			if state.sweFixstar(starname, tEt, flags, x0[:], message) == -1 {
				return -1
			}
		} else if state.sweCalc(tEt, body, flags, x0[:], message) == -1 {
			return -1
		}
		if imeth == 1 {
			x0[1] = 0
		}
		dgsect[0] = state.sweHousePos(armc, location[1], eps+nutlo[1], 'G', x0[:], nil)
		return 0
	}
	if imeth == 2 || imeth == 4 {
		risemeth |= 512
	}
	if imeth == 2 || imeth == 3 {
		risemeth |= 256
	}
	retval = state.sweRiseTrans(tUt, body, starname, ephemeris, 1|risemeth, location, pressure, temperature, tret[:], message)
	if retval == -1 {
		return -1
	} else if retval == -2 {
		riseFound = 0
	}
	retval = state.sweRiseTrans(tUt, body, starname, ephemeris, 2|risemeth, location, pressure, temperature, tret[1:], message)
	if retval == -1 {
		return -1
	} else if retval == -2 {
		setFound = 0
	}
	if tret[0] < tret[1] && riseFound == 1 {
		aboveHorizon = 0
		t = tUt - 1.2
		if setFound != 0 {
			t = tret[1] - 1.2
		}
		setFound = 1
		retval = state.sweRiseTrans(t, body, starname, ephemeris, 2|risemeth, location, pressure, temperature, tret[1:], message)
		if retval == -1 {
			return -1
		} else if retval == -2 {
			setFound = 0
		}
	} else if tret[0] >= tret[1] && setFound == 1 {
		aboveHorizon = 1
		t = tUt - 1.2
		if riseFound != 0 {
			t = tret[0] - 1.2
		}
		riseFound = 1
		retval = state.sweRiseTrans(t, body, starname, ephemeris, 1|risemeth, location, pressure, temperature, tret[:], message)
		if retval == -1 {
			return -1
		} else if retval == -2 {
			riseFound = 0
		}
	}
	if riseFound != 0 && setFound != 0 {
		if aboveHorizon != 0 {
			dgsect[0] = (tUt-tret[0])/(tret[1]-tret[0])*18 + 1
		} else {
			dgsect[0] = (tUt-tret[1])/(tret[0]-tret[1])*18 + 19
		}
		return 0
	} else {
		dgsect[0] = 0
		if message != nil {
			formatText(message, textBytes("rise or set not found for planet %d"), body)
		}
		return -1
	}
}
