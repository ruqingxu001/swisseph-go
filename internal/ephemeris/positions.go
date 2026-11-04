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

// sweVersion implements swe_version from sweph.c.
func (state *State) sweVersion(s []byte) []byte {
	copyText(s, textBytes("2.10.03"))
	return s
}

// sweGetLibraryPath implements swe_get_library_path from sweph.c.
func (state *State) sweGetLibraryPath(s []byte) []byte {
	s[0] = byte('\000')
	return s
}

// sweCalc implements swe_calc from sweph.c.
func (state *State) sweCalc(jd float64, body int32, flags int32, xx []float64, message []byte) int32 {
	var (
		i          int32
		j          int32
		moonIndex  int32 = 0
		savedFlags int32 = flags
		ephemeris  int32
		useSpeed3  int32 = 0
		sd         *savedPosition
		x          [6]float64
		xs         []float64
		x0         [24]float64
		x2         [24]float64
		dt         float64
	)
endSweCalc:
	for {
		if message != nil {
			message[0] = byte('\000')
		}
		if body == 10000+134340 {
			body = 9
		}
		ephemeris = flags & (flagJPL | flagSwiss | flagMoshier)
		if ephemeris&4 != 0 {
			ephemeris = 4
		} else if ephemeris&1 != 0 {
			ephemeris = 1
		} else {
			ephemeris = 2
		}
		if state.swiInitSwedIfStart() == 1 && !(ephemeris&4 != 0) && message != nil {
			copyText(message, textBytes("Please call swe_set_ephe_path() or swe_set_jplfile() before calling swe_calc() or swe_calc_ut()"))
		}
		if state.positions.data.lastEpheflag != ephemeris {
			state.freePlanets()
			if body != -1 {
				if state.positions.data.jplFileIsOpen != 0 {
					state.swiCloseJplFile()
					state.positions.data.jplFileIsOpen = 0
				}
				for i = 0; i < 7; i++ {
					if state.positions.data.files[i].handle != nil {
						closeTextFile(state.positions.data.files[i].handle)
					}
					state.positions.data.files[i] = ephemerisFile{}
				}
				state.positions.data.lastEpheflag = ephemeris
			}
		}
		if flags&flagSpeed3 != 0 && flags&flagSpeed != 0 {
			flags &^= flagSpeed3
		}
		if flags&flagSpeed3 != 0 {
			useSpeed3 = 1
		}
		if flags&flagSpeed != 0 && flags&flagTopocentric != 0 && flags&flagNoAberration == 0 {
			useSpeed3 = 1
		}
		if flags&flagCartesian != 0 && flags&flagRadians != 0 {
			flags &^= flagRadians
		}
		if flags&flagCentreOfBody != 0 && body <= 9 && flags&(flagHeliocentric|flagTruePosition|flagJ2000|flagICRS|flagRawPlanetMoon) != flagTestPlanetMoon {
			moonIndex = body*100 + 9099
		}
		if body >= 9000 && body < 10000 && flags&(flagHeliocentric|flagTruePosition|flagJ2000|flagICRS|flagRawPlanetMoon) != flagTestPlanetMoon {
			moonIndex = body
			body = (body - 9000) / 100
			flags |= flagCentreOfBody
		}
		if flags&flagCentreOfBody != 0 && body <= 4 && moonIndex%100 == 99 {
			moonIndex = 0
			flags &^= flagCentreOfBody
		}
		if flags&flagCentreOfBody != 0 || moonIndex > 0 {
			state.swiForceAppPosEtc()
		}
		if body < 23 && body >= 0 {
			sd = &state.positions.data.savedPositions[body]
		} else {
			sd = &state.positions.data.savedPositions[23]
		}
		if sd.epoch == jd && jd != 0 && body == sd.body && moonIndex == 0 {
			if sd.flags & ^(2*1024|4*1024|8*1024) == flags & ^(2*1024|4*1024|8*1024) {
				break endSweCalc
			}
		}
		if !(useSpeed3 != 0) {
			sd.epoch = jd
			sd.body = body
			if sd.flags = state.swecalc(jd, body, moonIndex, flags, sd.coordinates[:], message); sd.flags == -1 {
				i = 0
				for i = 0; i <= 5; i++ {
					xx[i] = 0
				}
				return -1
			}
		} else {
			sd.epoch = jd
			sd.body = body
			switch body {
			case 1:
				dt = 0.00005
			case 13:
				fallthrough
			case 11:
				dt = 0.1
			default:
				dt = 0.0001
			}
			if sd.flags = state.swecalc(jd-dt, body, moonIndex, flags, x0[:], message); sd.flags == -1 {
				i = 0
				for i = 0; i <= 5; i++ {
					xx[i] = 0
				}
				return -1
			}
			if sd.flags = state.swecalc(jd+dt, body, moonIndex, flags, x2[:], message); sd.flags == -1 {
				i = 0
				for i = 0; i <= 5; i++ {
					xx[i] = 0
				}
				return -1
			}
			if sd.flags = state.swecalc(jd, body, moonIndex, flags, sd.coordinates[:], message); sd.flags == -1 {
				i = 0
				for i = 0; i <= 5; i++ {
					xx[i] = 0
				}
				return -1
			}
			state.denormalizePositions(x0[:], sd.coordinates[:], x2[:])
			state.calcSpeed(x0[:], sd.coordinates[:], x2[:], dt)
		}
		break
	}
	if flags&flagEquatorial != 0 {
		xs = sd.coordinates[12:]
	} else {
		xs = sd.coordinates[:]
	}
	if flags&flagCartesian != 0 {
		xs = xs[6:]
	}
	if body == -1 {
		i = 4
	} else {
		i = 3
	}
	for j = 0; j < i; j++ {
		x[j] = xs[j]
	}
	for j = i; j < 6; j++ {
		x[j] = 0
	}
	if flags&(flagSpeed3|flagSpeed) != 0 {
		for j = 3; j < 6; j++ {
			x[j] = xs[j]
		}
	}
	if flags&flagRadians != 0 {
		if body == -1 {
			for j = 0; j < 4; j++ {
				x[j] *= radiansPerDegree
			}
		} else {
			for j = 0; j < 2; j++ {
				x[j] *= radiansPerDegree
			}
			if flags&(flagSpeed3|flagSpeed) != 0 {
				for j = 3; j < 5; j++ {
					x[j] *= radiansPerDegree
				}
			}
		}
	}
	for i = 0; i <= 5; i++ {
		xx[i] = x[i]
	}
	flags = sd.flags & ^(2*1024 | 4*1024 | 8*1024)
	flags |= savedFlags & (flagEquatorial | flagCartesian | flagRadians)
	if savedFlags&(flagJPL|flagSwiss|flagMoshier) == 0 {
		flags = flags & ^2
	}
	return flags
}

// sweCalcUt implements swe_calc_ut from sweph.c.
func (state *State) sweCalcUt(jdUT float64, body int32, flags int32, xx []float64, message []byte) int32 {
	var (
		deltat    float64
		retval    int32 = 0
		ephemeris int32 = 0
	)
	flags = state.plausIflag(flags, body, jdUT, message)
	ephemeris = flags & (flagJPL | flagSwiss | flagMoshier)
	if ephemeris == 0 {
		ephemeris = 2
		flags |= flagSwiss
	}
	deltat = state.sweDeltatEx(jdUT, flags, message)
	retval = state.sweCalc(jdUT+deltat, body, flags, xx, message)
	if retval&(1|2|4) != ephemeris {
		deltat = state.sweDeltatEx(jdUT, retval, nil)
		retval = state.sweCalc(jdUT+deltat, body, flags, xx, nil)
	}
	return retval
}

// swecalc implements swecalc from sweph.c.
func (state *State) swecalc(jd float64, body int32, moonIndex int32, flags int32, x []float64, message []byte) int32 {
	var (
		i         int32
		bodyIndex int32
		ipliAst   int32
		fileIndex int32
		status    int32
		ephemeris int32 = 2
		planet    *planetData
		earthData *planetData = &state.positions.data.planets[0]
		sunData   *planetData = &state.positions.data.planets[10]
		node      *planetData
		xp        []float64
		xp2       []float64
		ss        [3]float64
		serr2     [256]byte
	)
	serr2[0] = byte('\000')
	flags = state.plausIflag(flags, body, jd, message)
	if flags&flagMoshier != 0 {
		ephemeris = 4
	}
	if flags&flagSwiss != 0 {
		ephemeris = 2
	}
	if flags&flagJPL != 0 {
		ephemeris = 1
	}
	if flags&flagBarycentric != 0 && flags&flagMoshier != 0 {
		if message != nil {
			copyText(message, textBytes("barycentric Moshier positions are not supported."))
		}
		return -1
	}
	if ephemeris != 4 && !(state.positions.data.ephePathIsSet != 0) && !(state.positions.data.jplFileIsOpen != 0) {
		state.sweSetEphePath(nil)
	}
	if flags&flagSidereal != 0 && !(state.positions.data.ayanaIsSet != 0) {
		state.sweSetSidMode(0, 0, 0)
	}
	state.swiCheckEcliptic(jd, flags)
	state.swiCheckNutation(jd, flags)
	if body == -1 {
		x[0] = state.positions.data.obliquity.angle + state.positions.data.nutation.angles[1]
		x[1] = state.positions.data.obliquity.angle
		x[2] = state.positions.data.nutation.angles[0]
		x[3] = state.positions.data.nutation.angles[1]
		for i = 0; i <= 3; i++ {
			x[i] *= degreesPerRadian
		}
		return flags
	} else if body == 1 {
		bodyIndex = 1
		planet = &state.positions.data.planets[bodyIndex]
		xp = planet.coordinates[:]
		selection1 := ephemeris
	selectCase1:
		for {
			switch selection1 {
			case 1:
				status = state.jplplan(jd, bodyIndex, flags, 1, nil, nil, nil, message)
				if status == -1 {
					i = 0
					for i = 0; i < 24; i++ {
						x[i] = 0
					}
					return -1
				}
				if status == -2 {
					flags = flags & ^1 | 2
					if message != nil && textLength(message)+30 < 256 {
						appendText(message, textBytes(" \ntrying Swiss Eph; "))
					}
					selection1 = 2
					continue selectCase1
				} else if status == -3 {
					if jd > 625000.5 && jd < 2818000.5 {
						flags = flags & ^1 | 4
						if message != nil && textLength(message)+30 < 256 {
							appendText(message, textBytes(" \nusing Moshier Eph; "))
						}
						selection1 = 4
						continue selectCase1
					} else {
						i = 0
						for i = 0; i < 24; i++ {
							x[i] = 0
						}
						return -1
					}
				}
			case 2:
				status = state.sweplan(jd, bodyIndex, 1, flags, 1, nil, nil, nil, nil, message)
				if status == -1 {
					i = 0
					for i = 0; i < 24; i++ {
						x[i] = 0
					}
					return -1
				}
				if status == -2 {
					if jd > 625000.5 && jd < 2818000.5 {
						flags = flags & ^2 | 4
						if message != nil && textLength(message)+30 < 256 {
							appendText(message, textBytes(" \nusing Moshier eph.; "))
						}
						selection1 = 4
						continue selectCase1
					} else {
						i = 0
						for i = 0; i < 24; i++ {
							x[i] = 0
						}
						return -1
					}
				}
			case 4:
				status = state.swiMoshmoon(jd, 1, nil, message)
				if status == -1 {
					i = 0
					for i = 0; i < 24; i++ {
						x[i] = 0
					}
					return -1
				}
				status = state.swiMoshplan(jd, 0, 1, nil, nil, message)
				if status == -1 {
					i = 0
					for i = 0; i < 24; i++ {
						x[i] = 0
					}
					return -1
				}
			default:
			}
			break
		}
		if status = state.appPosEtcMoon(flags, message); status != 0 {
			i = 0
			for i = 0; i < 24; i++ {
				x[i] = 0
			}
			return -1
		}
	} else if body == 0 && flags&flagBarycentric != 0 {
		bodyIndex = 0
		xp = earthData.coordinates[:]
		selection2 := ephemeris
	selectCase2:
		for {
			switch selection2 {
			case 1:
				if !(state.positions.data.jplFileIsOpen != 0) {
					status = state.openJplFile(ss[:], state.positions.data.jplFilename[:], state.positions.data.ephemerisPath[:], message)
					if status != 0 {
						selection2 = 2
						continue selectCase2
					}
				}
				status = state.swiPleph(jd, 10, 11, sunData.stateVector[:], message)
				if status == -1 || status == -3 {
					state.swiCloseJplFile()
					state.positions.data.jplFileIsOpen = 0
					i = 0
					for i = 0; i < 24; i++ {
						x[i] = 0
					}
					return -1
				}
				if status == -2 {
					flags = flags & ^1 | 2
					if message != nil && textLength(message)+30 < 256 {
						appendText(message, textBytes(" \ntrying Swiss Eph; "))
					}
					selection2 = 2
					continue selectCase2
				}
				sunData.evaluationTime = jd
			case 2:
				status = state.sweplan(jd, 0, 0, flags, 1, nil, nil, nil, nil, message)
				if status == -1 || status == -2 {
					i = 0
					for i = 0; i < 24; i++ {
						x[i] = 0
					}
					return -1
				}
				sunData.evaluationTime = jd
			default:
				return -1
			}
			break
		}
		if status = state.appPosEtcSbar(flags, message); status != 0 {
			i = 0
			for i = 0; i < 24; i++ {
				x[i] = 0
			}
			return -1
		}
		flags = earthData.cachedFlags
		earthData.cachedFlags = -1
	} else if body == 0 || body == 2 || body == 3 || body == 4 || body == 5 || body == 6 || body == 7 || body == 8 || body == 9 || body == 14 {
		if flags&flagHeliocentric != 0 {
			if body == 0 {
				for i = 0; i < 24; i++ {
					x[i] = 0
				}
				return flags
			}
		} else if flags&flagBarycentric != 0 {
		} else if body == 14 {
			for i = 0; i < 24; i++ {
				x[i] = 0
			}
			return flags
		}
		bodyIndex = positionsPnoext2int[body]
		planet = &state.positions.data.planets[bodyIndex]
		xp = planet.coordinates[:]
		status = state.mainPlanet(jd, bodyIndex, moonIndex, ephemeris, flags, message)
		if status == -1 {
			i = 0
			for i = 0; i < 24; i++ {
				x[i] = 0
			}
			return -1
		}
		flags = planet.cachedFlags
	} else if body == 10 {
		if flags&flagHeliocentric != 0 || flags&flagBarycentric != 0 {
			for i = 0; i < 24; i++ {
				x[i] = 0
			}
			return flags
		}
		node = &state.positions.data.lunarNodes[0]
		xp = node.coordinates[:]
		xp2 = node.stateVector[:]
		status = state.swiMeanNode(jd, xp2, message)
		if status == -1 {
			i = 0
			for i = 0; i < 24; i++ {
				x[i] = 0
			}
			return -1
		}
		status = state.swiMeanNode(jd-0.001, xp2[3:], message)
		if status == -1 {
			i = 0
			for i = 0; i < 24; i++ {
				x[i] = 0
			}
			return -1
		}
		xp2[3] = state.sweDifrad2n(xp2[0], xp2[3]) / 0.001
		xp2[5] = 0
		xp2[4] = xp2[5]
		node.evaluationTime = jd
		node.cachedFlags = -1
		if status = state.appPosEtcMean(0, flags, message); status != 0 {
			i = 0
			for i = 0; i < 24; i++ {
				x[i] = 0
			}
			return -1
		}
		if flags&flagSidereal == 0 && flags&flagJ2000 == 0 {
			node.coordinates[1] = 0.0
			node.coordinates[4] = 0.0
			node.coordinates[5] = 0.0
			node.coordinates[8] = 0.0
			node.coordinates[11] = 0.0
		}
		if status == -1 {
			i = 0
			for i = 0; i < 24; i++ {
				x[i] = 0
			}
			return -1
		}
	} else if body == 12 {
		if flags&flagHeliocentric != 0 || flags&flagBarycentric != 0 {
			for i = 0; i < 24; i++ {
				x[i] = 0
			}
			return flags
		}
		node = &state.positions.data.lunarNodes[2]
		xp = node.coordinates[:]
		xp2 = node.stateVector[:]
		status = state.swiMeanApog(jd, xp2, message)
		if status == -1 {
			i = 0
			for i = 0; i < 24; i++ {
				x[i] = 0
			}
			return -1
		}
		status = state.swiMeanApog(jd-0.001, xp2[3:], message)
		if status == -1 {
			i = 0
			for i = 0; i < 24; i++ {
				x[i] = 0
			}
			return -1
		}
		for i = 0; i <= 1; i++ {
			xp2[3+i] = state.sweDifrad2n(xp2[i], xp2[3+i]) / 0.001
		}
		xp2[5] = 0
		node.evaluationTime = jd
		node.cachedFlags = -1
		if status = state.appPosEtcMean(2, flags, message); status != 0 {
			i = 0
			for i = 0; i < 24; i++ {
				x[i] = 0
			}
			return -1
		}
		node.coordinates[5] = 0.0
		if status == -1 {
			i = 0
			for i = 0; i < 24; i++ {
				x[i] = 0
			}
			return -1
		}
	} else if body == 11 {
		if flags&flagHeliocentric != 0 || flags&flagBarycentric != 0 {
			for i = 0; i < 24; i++ {
				x[i] = 0
			}
			return flags
		}
		node = &state.positions.data.lunarNodes[1]
		xp = node.coordinates[:]
		status = state.lunarOscElem(jd, 1, flags, message)
		flags = node.cachedFlags
		if flags&flagSidereal == 0 && flags&flagJ2000 == 0 {
			node.coordinates[1] = 0.0
			node.coordinates[4] = 0.0
			node.coordinates[8] = 0.0
			node.coordinates[11] = 0.0
		}
		if status == -1 {
			i = 0
			for i = 0; i < 24; i++ {
				x[i] = 0
			}
			return -1
		}
	} else if body == 13 {
		if flags&flagHeliocentric != 0 || flags&flagBarycentric != 0 {
			for i = 0; i < 24; i++ {
				x[i] = 0
			}
			return flags
		}
		node = &state.positions.data.lunarNodes[3]
		xp = node.coordinates[:]
		status = state.lunarOscElem(jd, 3, flags, message)
		flags = node.cachedFlags
		if status == -1 {
			i = 0
			for i = 0; i < 24; i++ {
				x[i] = 0
			}
			return -1
		}
	} else if body == 21 {
		if flags&flagHeliocentric != 0 || flags&flagBarycentric != 0 {
			for i = 0; i < 24; i++ {
				x[i] = 0
			}
			return flags
		}
		if jd < 625000.5 || jd > 2818000.5 {
			for i = 0; i < 24; i++ {
				x[i] = 0
			}
			if message != nil {
				formatText(message, textBytes("Interpolated apsides are restricted to JD %8.1f - JD %8.1f"), 625000.5, 2818000.5)
			}
			return -1
		}
		node = &state.positions.data.lunarNodes[4]
		xp = node.coordinates[:]
		status = state.intpApsides(jd, 4, flags, message)
		flags = node.cachedFlags
		if status == -1 {
			i = 0
			for i = 0; i < 24; i++ {
				x[i] = 0
			}
			return -1
		}
	} else if body == 22 {
		if flags&flagHeliocentric != 0 || flags&flagBarycentric != 0 {
			for i = 0; i < 24; i++ {
				x[i] = 0
			}
			return flags
		}
		if jd < 625000.5 || jd > 2818000.5 {
			for i = 0; i < 24; i++ {
				x[i] = 0
			}
			if message != nil {
				formatText(message, textBytes("Interpolated apsides are restricted to JD %8.1f - JD %8.1f"), 625000.5, 2818000.5)
			}
			return -1
		}
		node = &state.positions.data.lunarNodes[5]
		xp = node.coordinates[:]
		status = state.intpApsides(jd, 5, flags, message)
		flags = node.cachedFlags
		if status == -1 {
			i = 0
			for i = 0; i < 24; i++ {
				x[i] = 0
			}
			return -1
		}
	} else if body == 15 || body == 16 || body == 17 || body == 18 || body == 19 || body == 20 || body > 9000 || body > 10000 {
		if body < 23 {
			bodyIndex = positionsPnoext2int[body]
		} else if body <= 10000+4 && body > 10000 {
			bodyIndex = 14 + body - 10000 - 1
			body = 17 + body - 10000 - 1
		} else {
			bodyIndex = 11
		}
		if bodyIndex == 11 {
			ipliAst = body
		} else {
			ipliAst = bodyIndex
		}
		planet = &state.positions.data.planets[bodyIndex]
		xp = planet.coordinates[:]
		if ipliAst > 10000 {
			fileIndex = 3
		} else if ipliAst > 9000 {
			fileIndex = 3
		} else {
			fileIndex = 2
		}
		if bodyIndex == 12 && (jd < 1967601.5 || jd > 3419437.5) {
			if message != nil {
				formatText(message, textBytes("Chiron's ephemeris is restricted to JD %8.1f - JD %8.1f"), 1967601.5, 3419437.5)
			}
			return -1
		}
		if bodyIndex == 13 && (jd < 640648.5 || jd > 4390617.5) {
			if message != nil {
				formatText(message, textBytes("Pholus's ephemeris is restricted to JD %8.1f - JD %8.1f"), 640648.5, 4390617.5)
			}
			return -1
		}
	doAsteroid:
		for {
			status = state.mainPlanet(jd, 0, 0, ephemeris, flags, message)
			if status == -1 {
				i = 0
				for i = 0; i < 24; i++ {
					x[i] = 0
				}
				return -1
			}
			flags = state.positions.data.planets[0].cachedFlags
			if message != nil {
				copyText(serr2[:], message)
				message[0] = byte('\000')
			}
			status = state.sweph(jd, ipliAst, fileIndex, flags, sunData.stateVector[:], 1, nil, message)
			if status == -1 || status == -2 {
				i = 0
				for i = 0; i < 24; i++ {
					x[i] = 0
				}
				return -1
			}
			status = state.appPosEtcPlan(ipliAst, 0, flags, message)
			if status == -1 {
				i = 0
				for i = 0; i < 24; i++ {
					x[i] = 0
				}
				return -1
			}
			if status == -2 || status == -3 {
				if ephemeris != 4 {
					flags = flags & ^(1|2|4) | 4
					ephemeris = 4
					if message != nil && textLength(message)+30 < 256 {
						appendText(message, textBytes("\nusing Moshier eph.; "))
					}
					continue doAsteroid
				} else {
					i = 0
					for i = 0; i < 24; i++ {
						x[i] = 0
					}
					return -1
				}
			}
			if message != nil && int32(message[0]) == '\000' && int32(serr2[0]) != '\000' {
				copyText(message, textBytes("sun: "))
				serr2[256-5] = byte('\000')
				appendText(message, serr2[:])
			}
			break
		}
	} else if body >= 40 && body <= 999 {
		bodyIndex = 11
		planet = &state.positions.data.planets[bodyIndex]
		xp = planet.coordinates[:]
	doFictPlan:
		for {
			status = state.mainPlanet(jd, 0, 0, ephemeris, flags, message)
			flags = state.positions.data.planets[0].cachedFlags
			if state.swiOscElPlan(jd, planet.stateVector[:], body-40, bodyIndex, earthData.stateVector[:], sunData.stateVector[:], message) != 0 {
				i = 0
				for i = 0; i < 24; i++ {
					x[i] = 0
				}
				return -1
			}
			if status == -1 {
				i = 0
				for i = 0; i < 24; i++ {
					x[i] = 0
				}
				return -1
			}
			status = state.appPosEtcPlanOsc(body, bodyIndex, flags, message)
			if status == -1 {
				i = 0
				for i = 0; i < 24; i++ {
					x[i] = 0
				}
				return -1
			}
			if status == -2 || status == -3 {
				if ephemeris != 4 {
					flags = flags & ^(1|2|4) | 4
					ephemeris = 4
					if message != nil && textLength(message)+30 < 256 {
						appendText(message, textBytes("\nusing Moshier eph.; "))
					}
					continue doFictPlan
				} else {
					i = 0
					for i = 0; i < 24; i++ {
						x[i] = 0
					}
					return -1
				}
			}
			break
		}
	} else {
		if message != nil {
			formatText(message, textBytes("illegal planet number %d."), body)
		}
		i = 0
		for i = 0; i < 24; i++ {
			x[i] = 0
		}
		return -1
	}
	for i = 0; i < 24; i++ {
		x[i] = xp[i]
	}
	return flags
}

// freePlanets implements free_planets from sweph.c.
func (state *State) freePlanets() {
	var i int32
	for i = 0; i < 18; i++ {
		if state.positions.data.planets[i].coefficients != nil {
		}
		if state.positions.data.planets[i].referenceCoefficients != nil {
		}
		state.positions.data.planets[i] = planetData{}
	}
	for i = 0; i <= 23; i++ {
		state.positions.data.savedPositions[i] = savedPosition{}
	}
	for i = 0; i < 6; i++ {
		state.positions.data.lunarNodes[i] = planetData{}
	}
}

// swiInitSwedIfStart implements swi_init_swed_if_start from sweph.c.
func (state *State) swiInitSwedIfStart() int32 {
	if !(state.positions.data.swedIsInitialised != 0) {
		state.positions.data = calculationData{}
		copyText(state.positions.data.ephemerisPath[:], textBytes(".:/users/ephe2/:/users/ephe/"))
		copyText(state.positions.data.jplFilename[:], textBytes("de431.eph"))
		state.sweSetTidAcc(999999)
		state.positions.data.swedIsInitialised = 1
		return 1
	}
	return 0
}

// swiCloseKeepTopoEtc implements swi_close_keep_topo_etc from sweph.c.
//
//	closes all open files, frees space of planetary data,
//
// deletes memory of all computed positions
func (state *State) swiCloseKeepTopoEtc() {
	var i int32
	for i = 0; i < 7; i++ {
		if state.positions.data.files[i].handle != nil {
			closeTextFile(state.positions.data.files[i].handle)
		}
		state.positions.data.files[i] = ephemerisFile{}
	}
	state.freePlanets()
	state.positions.data.obliquity = obliquity{}
	state.positions.data.obliquityJ2000 = obliquity{}
	state.positions.data.nutation = nutation{}
	state.positions.data.nutationJ2000 = nutation{}
	state.positions.data.nutationWithSpeed = nutation{}
	clear(state.positions.data.astroModels[:])
	state.swiCloseJplFile()
	state.positions.data.jplFileIsOpen = 0
	state.positions.data.jplDENumber = 0
	if state.positions.data.starFile != nil {
		closeTextFile(state.positions.data.starFile)
		state.positions.data.starFile = nil
	}
	state.sweSetTidAcc(999999)
	state.positions.data.isOldStarfile = 0
	state.positions.data.iSavedPlanetName = 0
	state.positions.data.savedPlanetName[0] = byte('\000')
	state.positions.data.timeout = 0
}

// sweClose implements swe_close from sweph.c.
//
//	closes all open files, frees space of planetary data,
//
// deletes memory of all computed positions
func (state *State) sweClose() {
	var i int32
	for i = 0; i < 7; i++ {
		if state.positions.data.files[i].handle != nil {
			closeTextFile(state.positions.data.files[i].handle)
		}
		state.positions.data.files[i] = ephemerisFile{}
	}
	state.freePlanets()
	state.positions.data.obliquity = obliquity{}
	state.positions.data.obliquityJ2000 = obliquity{}
	state.positions.data.nutation = nutation{}
	state.positions.data.nutationJ2000 = nutation{}
	state.positions.data.nutationWithSpeed = nutation{}
	clear(state.positions.data.astroModels[:])
	state.swiCloseJplFile()
	state.positions.data.jplFileIsOpen = 0
	state.positions.data.jplDENumber = 0
	if state.positions.data.starFile != nil {
		closeTextFile(state.positions.data.starFile)
		state.positions.data.starFile = nil
	}
	state.sweSetTidAcc(999999)
	state.positions.data.geoposIsSet = 0
	state.positions.data.ayanaIsSet = 0
	state.positions.data.isOldStarfile = 0
	state.positions.data.iSavedPlanetName = 0
	state.positions.data.savedPlanetName[0] = byte('\000')
	state.positions.data.observer = observerData{}
	state.positions.data.sidereal = siderealData{}
	state.positions.data.timeout = 0
	state.positions.data.lastEpheflag = 0
	if state.positions.data.dpsi != nil {
		state.positions.data.dpsi = nil
	}
	if state.positions.data.deps != nil {
		state.positions.data.deps = nil
	}
	if state.positions.data.nFixstarsRecords > 0 {
		state.positions.data.fixedStars = nil
		state.positions.data.nFixstarsReal = 0
		state.positions.data.nFixstarsNamed = 0
		state.positions.data.nFixstarsRecords = 0
	}
}

// sweSetEphePath implements swe_set_ephe_path from sweph.c.
func (state *State) sweSetEphePath(path []byte) {
	var (
		i     int32
		flags int32
		s     [256]byte
		sp    []byte
		xx    [6]float64
	)
	state.swiCloseKeepTopoEtc()
	state.swiInitSwedIfStart()
	state.positions.data.ephePathIsSet = 1
	if sp = environmentText(textBytes("SE_EPHE_PATH")); sp != nil && textLength(sp) != 0 && textLength(sp) <= uint64(256-1-13) {
		copyText(s[:], sp)
	} else if path == nil || int32(path[0]) == '\000' {
		copyText(s[:], textBytes(".:/users/ephe2/:/users/ephe/"))
	} else if textLength(path) <= uint64(256-1-13) {
		copyText(s[:], path)
	} else {
		copyText(s[:], textBytes(".:/users/ephe2/:/users/ephe/"))
	}
	i = int32(textLength(s[:]))
	if int32(s[i-1]) != int32(textBytes("/")[0]) && int32(s[0]) != '\000' {
		appendText(s[:], textBytes("/"))
	}
	copyText(state.positions.data.ephemerisPath[:], s[:])
	flags = flagSwiss | flagTruePosition | flagJ2000 | flagICRS
	state.positions.data.lastEpheflag = 2
	state.sweCalc(julianDayJ2000, 1, flags, xx[:], nil)
	if state.positions.data.files[1].handle != nil {
		state.swiSetTidAcc(0, 0, state.positions.data.files[1].deNumber, nil)
	}
}

// sweSetJplFile implements swe_set_jpl_file from sweph.c.
func (state *State) sweSetJplFile(fname []byte) {
	var (
		sp     []byte
		s      [256]byte
		status int32
		ss     [3]float64
	)
	state.swiCloseKeepTopoEtc()
	state.swiInitSwedIfStart()
	if textLength(fname) >= 256 {
		copyTextN(s[:], fname, 256-1)
		s[256-1] = byte('\000')
	} else {
		copyText(s[:], fname)
	}
	sp = findLastByte(s[:], int32(textBytes("/")[0]))
	if sp == nil {
		sp = s[:]
	} else {
		sp = sp[1:]
	}
	if textLength(sp) >= 256 {
		sp[256-1] = byte('\000')
	}
	copyText(state.positions.data.jplFilename[:], sp)
	status = state.openJplFile(ss[:], state.positions.data.jplFilename[:], state.positions.data.ephemerisPath[:], nil)
	if status == 0 {
		if state.positions.data.jplDENumber >= 403 {
			state.loadDpsiDeps()
		}
	}
}

// calcEpsilon implements calc_epsilon from sweph.c.
//
//	calculates obliquity of ecliptic and stores it together
//
// with its date, sine, and cosine
func (state *State) calcEpsilon(jd float64, flags int32, e *obliquity) {
	e.epoch = jd
	e.angle = state.swiEpsiln(jd, flags)
	e.sine = math.Sin(e.angle)
	e.cosine = math.Cos(e.angle)
}

// mainPlanet implements main_planet from sweph.c.
//
//	computes a main planet from any ephemeris, if it
//
// has not yet been computed for this date.
// since a geocentric position requires the earth, the
// earth's position will be computed as well. With SWISSEPH
// files the barycentric sun will be done as well.
// With Moshier, the moon will be done as well.
//
// tjd 		= julian day
// ipli		= body number
// epheflag	= which ephemeris? JPL, SWISSEPH, Moshier?
// iflag	= other flags
//
// the geocentric apparent position of ipli (or whatever has
// been specified in iflag) will be saved in
// &swed.pldat[ipli].xreturn[];
//
// the barycentric (heliocentric with Moshier) position J2000
// will be kept in
// &swed.pldat[ipli].x[];
func (state *State) mainPlanet(jd float64, bodyIndex int32, moonIndex int32, ephemeris int32, flags int32, message []byte) int32 {
	var status int32
	if flags&flagCentreOfBody != 0 && bodyIndex >= 4 && bodyIndex <= 9 {
		status = state.sweph(jd, moonIndex, 3, flags, nil, 1, nil, message)
		if status == -1 || status == -2 {
			return -1
		}
	}
	selection1 := ephemeris
selectCase1:
	for {
		switch selection1 {
		case 1:
			status = state.jplplan(jd, bodyIndex, flags, 1, nil, nil, nil, message)
			if status == -1 {
				return -1
			}
			if status == -2 {
				flags = flags & ^1 | 2
				if message != nil && textLength(message)+30 < 256 {
					appendText(message, textBytes(" \ntrying Swiss Eph; "))
				}
				selection1 = 2
				continue selectCase1
			} else if status == -3 {
				if jd > 625000.5 && jd < 2818000.5 {
					flags = flags & ^1 | 4
					if message != nil && textLength(message)+30 < 256 {
						appendText(message, textBytes(" \nusing Moshier Eph; "))
					}
					selection1 = 4
					continue selectCase1
				} else {
					return -1
				}
			}
			if bodyIndex == 0 {
				status = state.appPosEtcSun(flags, message)
			} else {
				status = state.appPosEtcPlan(bodyIndex, moonIndex, flags, message)
			}
			if status == -1 {
				return -1
			}
			if status == -2 {
				flags = flags & ^1 | 2
				if message != nil && textLength(message)+30 < 256 {
					appendText(message, textBytes(" \ntrying Swiss Eph; "))
				}
				selection1 = 2
				continue selectCase1
			} else if status == -3 {
				if jd > 625000.5 && jd < 2818000.5 {
					flags = flags & ^1 | 4
					if message != nil && textLength(message)+30 < 256 {
						appendText(message, textBytes(" \nusing Moshier Eph; "))
					}
					selection1 = 4
					continue selectCase1
				} else {
					return -1
				}
			}
		case 2:
			status = state.sweplan(jd, bodyIndex, 0, flags, 1, nil, nil, nil, nil, message)
			if status == -1 {
				return -1
			}
			if status == -2 {
				if jd > 625000.5 && jd < 2818000.5 {
					flags = flags & ^2 | 4
					if message != nil && textLength(message)+30 < 256 {
						appendText(message, textBytes(" \nusing Moshier eph.; "))
					}
					selection1 = 4
					continue selectCase1
				} else {
					return -1
				}
			}
			if bodyIndex == 0 {
				status = state.appPosEtcSun(flags, message)
			} else {
				status = state.appPosEtcPlan(bodyIndex, moonIndex, flags, message)
			}
			if status == -1 {
				return -1
			}
			if status == -2 {
				if jd > 625000.5 && jd < 2818000.5 {
					flags = flags & ^2 | 4
					if message != nil && textLength(message)+30 < 256 {
						appendText(message, textBytes(" \nusing Moshier eph.; "))
					}
					selection1 = 4
					continue selectCase1
				} else {
					return -1
				}
			}
		case 4:
			status = state.swiMoshplan(jd, bodyIndex, 1, nil, nil, message)
			if status == -1 {
				return -1
			}
			if bodyIndex == 0 {
				status = state.appPosEtcSun(flags, message)
			} else {
				status = state.appPosEtcPlan(bodyIndex, moonIndex, flags, message)
			}
			if status == -1 {
				return -1
			}
		default:
		}
		break
	}
	return 0
}

// mainPlanetBary implements main_planet_bary from sweph.c.
func (state *State) mainPlanetBary(jd float64, bodyIndex int32, ephemeris int32, flags int32, saveCache int32, xp []float64, xe []float64, xs []float64, xm []float64, message []byte) int32 {
	var (
		i      int32
		status int32
	)
	selection1 := ephemeris
selectCase1:
	for {
		switch selection1 {
		case 1:
			status = state.jplplan(jd, bodyIndex, flags, saveCache, xp, xe, xs, message)
			if status == -1 || status == -3 {
				return status
			}
			if status == -2 {
				flags = flags & ^1 | 2
				if message != nil && textLength(message)+30 < 256 {
					appendText(message, textBytes(" \ntrying Swiss Eph; "))
				}
				selection1 = 2
				continue selectCase1
			}
		case 2:
			status = state.sweplan(jd, bodyIndex, 0, flags, saveCache, xp, xe, xs, xm, message)
			if status == -1 {
				return -1
			}
			if status == -2 {
				if jd > 625000.5 && jd < 2818000.5 {
					flags = flags & ^2 | 4
					if message != nil && textLength(message)+30 < 256 {
						appendText(message, textBytes(" \nusing Moshier eph.; "))
					}
					selection1 = 4
					continue selectCase1
				} else {
					return -1
				}
			}
		case 4:
			status = state.swiMoshplan(jd, bodyIndex, saveCache, xp, xe, message)
			if status == -1 {
				return -1
			}
			i = 0
			for ; i <= 5; i++ {
				xs[i] = 0
			}
		default:
		}
		break
	}
	return 0
}

// swemoon implements swemoon from sweph.c.
//
//	SWISSEPH
//
// this routine computes heliocentric cartesian equatorial coordinates
// of equinox 2000 of
// geocentric moon
//
// tjd 		julian date
// iflag	flag
// do_save	save J2000 position in save area pdp->x ?
// xp		array of 6 doubles for lunar position and speed
// serr		error string
func (state *State) swemoon(jd float64, flags int32, saveCache int32, xpret []float64, message []byte) int32 {
	var (
		i       int32
		status  int32
		planet  *planetData = &state.positions.data.planets[1]
		speedf1 int32
		speedf2 int32
		xx      [6]float64
		xp      []float64
	)
	if saveCache != 0 {
		xp = planet.stateVector[:]
	} else {
		xp = xx[:]
	}
	speedf1 = planet.cachedFlags & 256
	speedf2 = flags & (flagSpeed)
	if jd == planet.evaluationTime && planet.ephemeris == 2 && (!(speedf2 != 0) || speedf1 != 0) {
		xp = planet.stateVector[:]
	} else {
		status = state.sweph(jd, 1, 1, flags, nil, saveCache, xp, message)
		if status != 0 {
			return status
		}
		if saveCache != 0 {
			planet.evaluationTime = jd
			planet.cachedFlags = -1
			planet.ephemeris = 2
		}
	}
	if xpret != nil {
		for i = 0; i <= 5; i++ {
			xpret[i] = xp[i]
		}
	}
	return 0
}

// sweplan implements sweplan from sweph.c.
//
//	SWISSEPH
//
// this function computes
// 1. a barycentric planet
// plus, under certain conditions,
// 2. the barycentric sun,
// 3. the barycentric earth, and
// 4. the geocentric moon,
// in barycentric cartesian equatorial coordinates J2000.
//
// these are the data needed for calculation of light-time etc.
//
// tjd 		julian date
// ipli		SEI_ planet number
// ifno		ephemeris file number
// do_save	write new positions in save area
// xp		array of 6 doubles for planet's position and velocity
// xpe                                 earth's
// xps                                 sun's
// xpm                                 moon's
// serr		error string
//
// xp - xpm can be NULL. if do_save is TRUE, all of them can be NULL.
// the positions will be written into the save area (swed.pldat[ipli].x)
func (state *State) sweplan(jd float64, bodyIndex int32, fileIndex int32, flags int32, saveCache int32, xpret []float64, xperet []float64, xpsret []float64, xpmret []float64, message []byte) int32 {
	var (
		i         int32
		status    int32
		doEarth   int32       = 0
		doMoon    int32       = 0
		doSunbary int32       = 0
		planet    *planetData = &state.positions.data.planets[bodyIndex]
		pebdp     *planetData = &state.positions.data.planets[0]
		psbdp     *planetData = &state.positions.data.planets[10]
		pmdp      *planetData = &state.positions.data.planets[1]
		xxp       [6]float64
		xxm       [6]float64
		xxs       [6]float64
		xxe       [6]float64
		xp        []float64
		xpe       []float64
		xpm       []float64
		xps       []float64
		speedf1   int32
		speedf2   int32
	)
	if saveCache != 0 || bodyIndex == 10 || planet.fileFlags&1 != 0 || xpsret != nil || flags&flagHeliocentric != 0 {
		doSunbary = 1
	}
	if saveCache != 0 || bodyIndex == 0 || xperet != nil {
		doEarth = 1
	}
	if bodyIndex == 1 {
		doEarth = 1
		doSunbary = 1
	}
	if saveCache != 0 || bodyIndex == 1 || bodyIndex == 0 || xperet != nil || xpmret != nil {
		doMoon = 1
	}
	if saveCache != 0 {
		xp = planet.stateVector[:]
		xpe = pebdp.stateVector[:]
		xps = psbdp.stateVector[:]
		xpm = pmdp.stateVector[:]
	} else {
		xp = xxp[:]
		xpe = xxe[:]
		xps = xxs[:]
		xpm = xxm[:]
	}
	speedf2 = flags & (flagSpeed)
	if doSunbary != 0 {
		speedf1 = psbdp.cachedFlags & 256
		if jd == psbdp.evaluationTime && psbdp.ephemeris == 2 && (!(speedf2 != 0) || speedf1 != 0) {
			for i = 0; i <= 5; i++ {
				xps[i] = psbdp.stateVector[i]
			}
		} else {
			status = state.sweph(jd, 10, 0, flags, nil, saveCache, xps, message)
			if status != 0 {
				return status
			}
		}
		if xpsret != nil {
			for i = 0; i <= 5; i++ {
				xpsret[i] = xps[i]
			}
		}
	}
	if doMoon != 0 {
		speedf1 = pmdp.cachedFlags & 256
		if jd == pmdp.evaluationTime && pmdp.ephemeris == 2 && (!(speedf2 != 0) || speedf1 != 0) {
			for i = 0; i <= 5; i++ {
				xpm[i] = pmdp.stateVector[i]
			}
		} else {
			status = state.sweph(jd, 1, 1, flags, nil, saveCache, xpm, message)
			if status == -1 {
				return status
			}
			if state.positions.data.files[1].handle == nil {
				if message != nil && textLength(message)+35 < 256 {
					appendText(message, textBytes(" \nusing Moshier eph. for moon; "))
				}
				status = state.swiMoshmoon(jd, saveCache, xpm, message)
				if status != 0 {
					return status
				}
			}
		}
		if xpmret != nil {
			for i = 0; i <= 5; i++ {
				xpmret[i] = xpm[i]
			}
		}
	}
	if doEarth != 0 {
		speedf1 = pebdp.cachedFlags & 256
		if jd == pebdp.evaluationTime && pebdp.ephemeris == 2 && (!(speedf2 != 0) || speedf1 != 0) {
			for i = 0; i <= 5; i++ {
				xpe[i] = pebdp.stateVector[i]
			}
		} else {
			status = state.sweph(jd, 0, 0, flags, nil, saveCache, xpe, message)
			if status != 0 {
				return status
			}
			state.embofs(xpe, xpm)
			if sameSlice(xpe, pebdp.stateVector[:]) || flags&flagSpeed != 0 {
				state.embofs(xpe[3:], xpm[3:])
			}
		}
		if xperet != nil {
			for i = 0; i <= 5; i++ {
				xperet[i] = xpe[i]
			}
		}
	}
	if bodyIndex == 1 {
		for i = 0; i <= 5; i++ {
			xp[i] = xpm[i]
		}
	} else if bodyIndex == 0 {
		for i = 0; i <= 5; i++ {
			xp[i] = xpe[i]
		}
	} else if bodyIndex == 0 {
		for i = 0; i <= 5; i++ {
			xp[i] = xps[i]
		}
	} else {
		speedf1 = planet.cachedFlags & 256
		if jd == planet.evaluationTime && planet.ephemeris == 2 && (!(speedf2 != 0) || speedf1 != 0) {
			for i = 0; i <= 5; i++ {
				xp[i] = planet.stateVector[i]
			}
			return 0
		} else {
			status = state.sweph(jd, bodyIndex, fileIndex, flags, nil, saveCache, xp, message)
			if status != 0 {
				return status
			}
			if planet.fileFlags&1 != 0 {
				for i = 0; i <= 2; i++ {
					xp[i] += xps[i]
				}
				if saveCache != 0 || flags&flagSpeed != 0 {
					for i = 3; i <= 5; i++ {
						xp[i] += xps[i]
					}
				}
			}
		}
	}
	if xpret != nil {
		for i = 0; i <= 5; i++ {
			xpret[i] = xp[i]
		}
	}
	return 0
}

// jplplan implements jplplan from sweph.c.
func (state *State) jplplan(jd float64, bodyIndex int32, flags int32, saveCache int32, xpret []float64, xperet []float64, xpsret []float64, message []byte) int32 {
	var (
		i         int32
		status    int32
		doEarth   int32 = 0
		doSunbary int32 = 0
		ss        [3]float64
		xxp       [6]float64
		xxe       [6]float64
		xxs       [6]float64
		xp        []float64
		xpe       []float64
		xps       []float64
		ictr      int32       = 11
		planet    *planetData = &state.positions.data.planets[bodyIndex]
		earthData *planetData = &state.positions.data.planets[0]
		sunData   *planetData = &state.positions.data.planets[10]
	)
	flags = flagJPL
	if saveCache != 0 {
		xp = planet.stateVector[:]
		xpe = earthData.stateVector[:]
		xps = sunData.stateVector[:]
	} else {
		xp = xxp[:]
		xpe = xxe[:]
		xps = xxs[:]
	}
	if saveCache != 0 || bodyIndex == 0 || xperet != nil || bodyIndex == 1 {
		doEarth = 1
	}
	if saveCache != 0 || bodyIndex == 10 || xpsret != nil || bodyIndex == 1 {
		doSunbary = 1
	}
	if bodyIndex == 1 {
		ictr = 2
	}
	if !(state.positions.data.jplFileIsOpen != 0) {
		status = state.openJplFile(ss[:], state.positions.data.jplFilename[:], state.positions.data.ephemerisPath[:], message)
		if status != 0 {
			return status
		}
	}
	if doEarth != 0 {
		if jd != earthData.evaluationTime || jd == 0 {
			status = state.swiPleph(jd, 2, 11, xpe, message)
			if saveCache != 0 {
				earthData.evaluationTime = jd
				earthData.cachedFlags = -1
				earthData.ephemeris = 1
			}
			if status != 0 {
				state.swiCloseJplFile()
				state.positions.data.jplFileIsOpen = 0
				return status
			}
		} else {
			xpe = earthData.stateVector[:]
		}
		if xperet != nil {
			for i = 0; i <= 5; i++ {
				xperet[i] = xpe[i]
			}
		}
	}
	if doSunbary != 0 {
		if jd != sunData.evaluationTime || jd == 0 {
			status = state.swiPleph(jd, 10, 11, xps, message)
			if saveCache != 0 {
				sunData.evaluationTime = jd
				sunData.cachedFlags = -1
				sunData.ephemeris = 1
			}
			if status != 0 {
				state.swiCloseJplFile()
				state.positions.data.jplFileIsOpen = 0
				return status
			}
		} else {
			xps = sunData.stateVector[:]
		}
		if xpsret != nil {
			for i = 0; i <= 5; i++ {
				xpsret[i] = xps[i]
			}
		}
	}
	if bodyIndex == 0 {
		for i = 0; i <= 5; i++ {
			xp[i] = xpe[i]
		}
	}
	if bodyIndex == 10 {
		for i = 0; i <= 5; i++ {
			xp[i] = xps[i]
		}
	} else if jd == planet.evaluationTime && planet.ephemeris == 1 {
		xp = planet.stateVector[:]
	} else {
		status = state.swiPleph(jd, positionsPnoint2jpl[bodyIndex], ictr, xp, message)
		if saveCache != 0 {
			planet.evaluationTime = jd
			planet.cachedFlags = -1
			planet.ephemeris = 1
		}
		if status != 0 {
			state.swiCloseJplFile()
			state.positions.data.jplFileIsOpen = 0
			return status
		}
	}
	if xpret != nil {
		for i = 0; i <= 5; i++ {
			xpret[i] = xp[i]
		}
	}
	return 0
}

// sweph implements sweph from sweph.c.
func (state *State) sweph(jd float64, bodyIndex int32, fileIndex int32, flags int32, xsunb []float64, saveCache int32, xpret []float64, message []byte) int32 {
	var (
		i         int32
		body      int32
		status    int32
		subdirlen int32
		s         [2 * 256]byte
		subdirnam [256]byte
		fname     [256]byte
		sp        []byte
		t         float64
		tsv       float64
		xemb      [6]float64
		xx        [6]float64
		xp        []float64
		planet    *planetData
		earthData *planetData    = &state.positions.data.planets[0]
		sunData   *planetData    = &state.positions.data.planets[10]
		fileData  *ephemerisFile = &state.positions.data.files[fileIndex]
		speedf1   int32
		speedf2   int32
		needSpeed int32
	)
	body = bodyIndex
	if bodyIndex > 10000 {
		body = 11
	}
	if bodyIndex > 9000 {
		body = 11
	}
	planet = &state.positions.data.planets[body]
	if saveCache != 0 {
		xp = planet.stateVector[:]
	} else {
		xp = xx[:]
	}
	speedf1 = planet.cachedFlags & 256
	speedf2 = flags & (flagSpeed)
	if jd == planet.evaluationTime && planet.ephemeris == 2 && (!(speedf2 != 0) || speedf1 != 0) && body < 11 {
		if xpret != nil {
			for i = 0; i <= 5; i++ {
				xpret[i] = planet.stateVector[i]
			}
		}
		return 0
	}
	if fileData.handle != nil {
		if jd < fileData.start || jd > fileData.end || body == 11 && bodyIndex != planet.bodyID {
			closeTextFile(fileData.handle)
			fileData.handle = nil
			if planet.referenceCoefficients != nil {
			}
			planet.referenceCoefficients = nil
			if planet.coefficients != nil {
			}
			planet.coefficients = nil
		}
	}
	if fileData.handle == nil {
		state.swiGenFilename(jd, bodyIndex, fname[:])
		copyText(subdirnam[:], fname[:])
		sp = findLastByte(subdirnam[:], int32(textBytes("/")[0]))
		if sp != nil {
			sp[0] = byte('\000')
			subdirlen = int32(textLength(subdirnam[:]))
		} else {
			subdirlen = 0
		}
		copyText(s[:], fname[:])
	again:
		for {
			fileData.handle = state.swiFopen(fileIndex, s[:], state.positions.data.ephemerisPath[:], message)
			if fileData.handle == nil {
				if bodyIndex > 9000 && bodyIndex < 10000 {
					if subdirlen > 0 && compareTextN(s[:], subdirnam[:], uint64(subdirlen)) == 0 {
						state.swiStrcpy(s[:], s[subdirlen+1:])
						continue again
					}
				} else if bodyIndex > 10000 {
					var spp []byte
					spp = findByte(s[:], '.')
					dot := -1
					if spp != nil {
						dot = len(s) - len(spp)
					}
					if dot > 0 && s[dot-1] != 's' {
						formatText(s[dot:], "s.%s", "se1")
						continue again
					}
					if dot <= 0 {
						return -2
					}
					spp = s[dot-1:]
					state.swiStrcpy(spp, spp[1:])
					if subdirlen > 0 && compareTextN(s[:], subdirnam[:], uint64(subdirlen)) == 0 {
						state.swiStrcpy(s[:], s[subdirlen+1:])
						continue again
					}
				}
				return -2
			}
			if message != nil {
				message[0] = byte('\000')
			}
			status = state.readConst(fileIndex, message)
			if status != 0 {
				return status
			}
			break
		}
	}
	if jd < fileData.start || jd > fileData.end {
		if message != nil {
			sp = findLastByte(fname[:], int32(textBytes("/")[0]))
			if sp != nil {
				sp = sp[1:]
			} else {
				sp = fname[:]
			}
			if bodyIndex > 10000 {
				formatText(s[:], textBytes("asteroid No. %d (%s): "), bodyIndex-10000, sp)
			} else if bodyIndex > 9000 {
				if findText(fname[:], textBytes("99.")) != nil {
					formatText(s[:], textBytes("plan. COB No. %d (%s): "), bodyIndex, sp)
				} else {
					formatText(s[:], textBytes("plan. moon No. %d (%s): "), bodyIndex, sp)
				}
			} else if bodyIndex > 9 {
				formatText(s[:], textBytes("asteroid eph. file (%s): "), sp)
			} else if bodyIndex != 1 {
				formatText(s[:], textBytes("planets eph. file (%s): "), sp)
			} else {
				formatText(s[:], textBytes("moon eph. file (%s): "), sp)
			}
			if jd < fileData.start {
				formatText(s[textLength(s[:]):], textBytes("jd %f < lower limit %f;"), jd, fileData.start)
			} else {
				formatText(s[textLength(s[:]):], textBytes("jd %f > upper limit %f;"), jd, fileData.end)
			}
			if textLength(message)+textLength(s[:]) < 256 {
				appendText(message, s[:])
			}
		}
		return -2
	}
	if planet.coefficients == nil || jd < planet.segmentStart || jd > planet.segmentEnd {
		status = state.getNewSegment(jd, body, fileIndex, message)
		if status != 0 {
			return status
		}
		if planet.fileFlags&2 != 0 {
			state.rotBack(body)
		} else {
			planet.evaluationCount = planet.coefficientCount
		}
	}
	t = (jd - planet.segmentStart) / planet.segmentSpan
	t = t*2 - 1
	needSpeed = int32(boolInt(saveCache != 0 || flags&flagSpeed != 0))
	for i = 0; i <= 2; i++ {
		xp[i] = state.swiEcheb(t, planet.coefficients[i*planet.coefficientCount:], planet.evaluationCount)
		if needSpeed != 0 {
			xp[i+3] = state.swiEdcheb(t, planet.coefficients[i*planet.coefficientCount:], planet.evaluationCount) / planet.segmentSpan * 2
		} else {
			xp[i+3] = 0
		}
	}
	if body == 10 && planet.fileFlags&8 != 0 {
		tsv = earthData.evaluationTime
		earthData.evaluationTime = 0
		status = state.sweph(jd, 0, fileIndex, flags|(flagSpeed), nil, 0, xemb[:], message)
		if status != 0 {
			return status
		}
		earthData.evaluationTime = tsv
		for i = 0; i <= 2; i++ {
			xp[i] = xemb[i] - xp[i]
		}
		if needSpeed != 0 {
			for i = 3; i <= 5; i++ {
				xp[i] = xemb[i] - xp[i]
			}
		}
	}
	if xsunb != nil && (flags&flagJPL != 0 || flags&flagSwiss != 0) {
		if body >= 11 {
			for i = 0; i <= 2; i++ {
				xp[i] += xsunb[i]
			}
			if needSpeed != 0 {
				for i = 3; i <= 5; i++ {
					xp[i] += xsunb[i]
				}
			}
		}
	}
	if saveCache != 0 {
		planet.evaluationTime = jd
		planet.cachedFlags = -1
		if fileIndex == 0 || fileIndex == 1 {
			planet.ephemeris = 2
		} else {
			planet.ephemeris = sunData.ephemeris
		}
	}
	if xpret != nil {
		for i = 0; i <= 5; i++ {
			xpret[i] = xp[i]
		}
	}
	return 0
}

// swiFopen searches the configured directories and records the selected path.
// The search order and diagnostics follow swi_fopen. Drive letters are kept
// intact on Windows instead of being interpreted as Unix path separators.
func (state *State) swiFopen(fileIndex int32, name, searchPath, message []byte) *textFile {
	filename, path := textString(name), textString(searchPath)
	var localName [256]byte
	selectedName := localName[:]
	if fileIndex >= 0 {
		selectedName = state.positions.data.files[fileIndex].name[:]
	}
	directories := splitEphemerisPath(path)
	for _, directory := range directories {
		fullName := filename
		if directory != "" && directory != "." {
			fullName = directory
			if fullName[len(fullName)-1] != '/' {
				fullName += "/"
			}
			fullName += filename
		}
		if len(fullName) >= 256 {
			formatText(message, "error: file path and name must be shorter than %d.", 256)
			return nil
		}
		copyText(selectedName, fullName)
		if file := openTextFile(selectedName, "r"); file != nil {
			return file
		}
	}
	formatText(message, "SwissEph file '%s' not found in PATH '%s'", filename, path)
	return nil
}

// swiGetDenum implements swi_get_denum from sweph.c.
//
//	s must not be longer then AS_MAXCH
func (state *State) swiGetDenum(bodyIndex int32, flags int32) int32 {
	var fileData *ephemerisFile = nil
	if flags&flagMoshier != 0 {
		return 403
	}
	if flags&flagJPL != 0 {
		if state.positions.data.jplDENumber > 0 {
			return state.positions.data.jplDENumber
		} else {
			return 431
		}
	}
	if bodyIndex > 10000 {
		fileData = &state.positions.data.files[3]
	} else if bodyIndex > 9000 {
		fileData = &state.positions.data.files[3]
	} else if bodyIndex == 12 || bodyIndex == 13 || bodyIndex == 14 || bodyIndex == 15 || bodyIndex == 16 || bodyIndex == 17 {
		fileData = &state.positions.data.files[2]
	} else if bodyIndex == 1 {
		fileData = &state.positions.data.files[1]
	} else {
		fileData = &state.positions.data.files[0]
	}
	if fileData != nil {
		if fileData.deNumber != 0 {
			return fileData.deNumber
		} else {
			return 431
		}
	}
	return 431
}

// calcCenterBody implements calc_center_body from sweph.c.
func (state *State) calcCenterBody(bodyIndex int32, flags int32, xx []float64, xcom []float64, message []byte) int32 {
	var i int32
	if flags&flagCentreOfBody == 0 {
		return 0
	}
	if bodyIndex < 4 || bodyIndex > 9 {
		return 0
	}
	for i = 0; i <= 5; i++ {
		xx[i] += xcom[i]
	}
	return 0
}

// appPosEtcPlan implements app_pos_etc_plan from sweph.c.
func (state *State) appPosEtcPlan(bodyIndex int32, moonIndex int32, flags int32, message []byte) int32 {
	var (
		i              int32
		j              int32
		niter          int32
		status         int32 = 0
		body           int32
		fileIndex      int32
		ibody          int32
		flg1           int32
		flg2           int32
		xx             [6]float64
		xx0            [6]float64
		dx             [3]float64
		dt             float64
		t              float64
		dtsaveForDefl  float64
		observerVector [6]float64
		xobs2          [6]float64
		earthVector    [6]float64
		sunVector      [6]float64
		xcom           [6]float64
		xxsp           [6]float64
		xxsv           [6]float64
		earthData      *planetData = &state.positions.data.planets[0]
		planet         *planetData
		obliquityData  *obliquity = &state.positions.data.obliquityJ2000
		ephemeris      int32      = flags & (flagJPL | flagSwiss | flagMoshier)
	)
	dtsaveForDefl = 0
	if bodyIndex > 9000 || bodyIndex > 10000 {
		fileIndex = 3
		ibody = 2
		planet = &state.positions.data.planets[11]
	} else if bodyIndex == 12 || bodyIndex == 13 || bodyIndex == 14 || bodyIndex == 15 || bodyIndex == 16 || bodyIndex == 17 {
		fileIndex = 2
		ibody = 3
		planet = &state.positions.data.planets[bodyIndex]
	} else {
		fileIndex = 0
		ibody = 0
		planet = &state.positions.data.planets[bodyIndex]
	}
	t = planet.evaluationTime
	flg1 = flags & ^(2 * 1024) & ^(4 * 1024)
	flg2 = planet.cachedFlags & ^(2 * 1024) & ^(4 * 1024)
	if flg1 == flg2 {
		planet.cachedFlags = flags
		planet.ephemeris = flags & (flagJPL | flagSwiss | flagMoshier)
		return 0
	}
	for i = 0; i <= 5; i++ {
		xx[i] = planet.stateVector[i]
	}
	state.calcCenterBody(bodyIndex, flags, xx[:], state.positions.data.planets[11].stateVector[:], message)
	for i = 0; i <= 5; i++ {
		xx0[i] = xx[i]
	}
	if flags&flagHeliocentric != 0 {
		if planet.ephemeris == 1 || planet.ephemeris == 2 {
			for i = 0; i <= 5; i++ {
				xx[i] -= state.positions.data.planets[10].stateVector[i]
			}
		}
	}
	if flags&flagTopocentric != 0 {
		if state.positions.data.observer.evaluationTime != earthData.evaluationTime || state.positions.data.observer.evaluationTime == 0 {
			if state.swiGetObserver(earthData.evaluationTime, flags|(flagNoNutation), 1, observerVector[:], message) != 0 {
				return -1
			}
		} else {
			for i = 0; i <= 5; i++ {
				observerVector[i] = state.positions.data.observer.stateVector[i]
			}
		}
		for i = 0; i <= 5; i++ {
			observerVector[i] = observerVector[i] + earthData.stateVector[i]
		}
	} else {
		for i = 0; i <= 5; i++ {
			observerVector[i] = earthData.stateVector[i]
		}
	}
	if flags&flagTruePosition == 0 {
		if planet.ephemeris == 1 || planet.ephemeris == 2 {
			niter = 1
		} else {
			niter = 0
		}
		if flags&flagSpeed != 0 {
			for i = 0; i <= 2; i++ {
				xxsp[i] = xx[i] - xx[i+3]
				xxsv[i] = xxsp[i]
			}
			for j = 0; j <= niter; j++ {
				for i = 0; i <= 2; i++ {
					dx[i] = xxsp[i]
					if flags&flagHeliocentric == 0 && flags&flagBarycentric == 0 {
						dx[i] -= observerVector[i] - observerVector[i+3]
					}
				}
				dt = math.Sqrt(dx[0]*dx[0]+dx[1]*dx[1]+dx[2]*dx[2]) * 1.49597870700e+11 / 2.99792458e+8 / secondsPerDay
				for i = 0; i <= 2; i++ {
					xxsp[i] = xxsv[i] - dt*xx0[i+3]
				}
			}
			for i = 0; i <= 2; i++ {
				xxsp[i] = xxsv[i] - xxsp[i]
			}
		}
		for j = 0; j <= niter; j++ {
			for i = 0; i <= 2; i++ {
				dx[i] = xx[i]
				if flags&flagHeliocentric == 0 && flags&flagBarycentric == 0 {
					dx[i] -= observerVector[i]
				}
			}
			dt = math.Sqrt(dx[0]*dx[0]+dx[1]*dx[1]+dx[2]*dx[2]) * 1.49597870700e+11 / 2.99792458e+8 / secondsPerDay
			t = planet.evaluationTime - dt
			dtsaveForDefl = dt
			for i = 0; i <= 2; i++ {
				xx[i] = xx0[i] - dt*xx0[i+3]
			}
		}
		if flags&flagSpeed != 0 {
			for i = 0; i <= 2; i++ {
				xxsp[i] = xx0[i] - xx[i] - xxsp[i]
			}
		}
		if flags&flagCentreOfBody != 0 && bodyIndex >= 4 && bodyIndex <= 9 {
			status = state.sweph(t, moonIndex, 3, flags, nil, 0, xcom[:], message)
			if status == -1 || status == -2 {
				return -1
			}
		}
		switch ephemeris {
		case 1:
			if ibody >= 2 {
				body = -1
			} else {
				body = positionsPnoint2jpl[bodyIndex]
			}
			if ibody == 0 {
				status = state.swiPleph(t, body, 11, xx[:], message)
				if status != 0 {
					state.swiCloseJplFile()
					state.positions.data.jplFileIsOpen = 0
				}
			} else {
				status = state.swiPleph(t, 10, 11, sunVector[:], message)
				if status != 0 {
					state.swiCloseJplFile()
					state.positions.data.jplFileIsOpen = 0
				}
				status = state.sweph(t, bodyIndex, fileIndex, flags, sunVector[:], 0, xx[:], message)
			}
			if status != 0 {
				return status
			}
			if flags&flagSpeed != 0 && flags&flagHeliocentric == 0 && flags&flagBarycentric == 0 {
				status = state.swiPleph(t, 2, 11, earthVector[:], message)
				if status != 0 {
					state.swiCloseJplFile()
					state.positions.data.jplFileIsOpen = 0
					return status
				}
			}
		case 2:
			if ibody == 0 {
				status = state.sweplan(t, bodyIndex, fileIndex, flags, 0, xx[:], earthVector[:], sunVector[:], nil, message)
			} else {
				status = state.sweplan(t, 0, 0, flags, 0, earthVector[:], nil, sunVector[:], nil, message)
				if status == 0 {
					status = state.sweph(t, bodyIndex, fileIndex, flags, sunVector[:], 0, xx[:], message)
				}
			}
			if status != 0 {
				return status
			}
		case 4:
			fallthrough
		default:
			if flags&flagSpeed != 0 && !(flags&(flagHeliocentric|flagBarycentric) != 0) {
				if ibody == 0 {
					status = state.swiMoshplan(t, bodyIndex, 0, xxsv[:], earthVector[:], message)
				} else {
					status = state.sweph(t, bodyIndex, fileIndex, flags, nil, 0, xxsv[:], message)
					if status == 0 {
						status = state.swiMoshplan(t, 0, 0, earthVector[:], earthVector[:], message)
					}
				}
				if status != 0 {
					return status
				}
				for i = 3; i <= 5; i++ {
					xx[i] = xxsv[i]
				}
			}
		}
		state.calcCenterBody(bodyIndex, flags, xx[:], xcom[:], message)
		if flags&flagHeliocentric != 0 {
			if planet.ephemeris == 1 || planet.ephemeris == 2 {
				for i = 0; i <= 5; i++ {
					xx[i] -= state.positions.data.planets[10].stateVector[i]
				}
			}
		}
		if flags&flagSpeed != 0 {
			if flags&flagTopocentric != 0 {
				if state.swiGetObserver(t, flags|(flagNoNutation), 0, xobs2[:], message) != 0 {
					return -1
				}
				for i = 0; i <= 5; i++ {
					xobs2[i] += earthVector[i]
				}
			} else {
				for i = 0; i <= 5; i++ {
					xobs2[i] = earthVector[i]
				}
			}
		}
	}
	if flags&flagHeliocentric == 0 && flags&flagBarycentric == 0 {
		for i = 0; i <= 5; i++ {
			xx[i] -= observerVector[i]
		}
		if flags&flagTruePosition == 0 {
			if flags&flagSpeed != 0 {
				for i = 3; i <= 5; i++ {
					xx[i] -= xxsp[i-3]
				}
			}
		}
	}
	if flags&flagSpeed == 0 {
		for i = 3; i <= 5; i++ {
			xx[i] = 0
		}
	}
	if flags&flagTruePosition == 0 && flags&flagNoDeflection == 0 {
		state.swiDeflectLight(xx[:], dtsaveForDefl, flags)
	}
	if flags&flagTruePosition == 0 && flags&flagNoAberration == 0 {
		state.swiAberrLight(xx[:], observerVector[:], flags)
		if flags&flagSpeed != 0 {
			for i = 3; i <= 5; i++ {
				xx[i] += observerVector[i] - xobs2[i]
			}
		}
	}
	if flags&flagSpeed == 0 {
		for i = 3; i <= 5; i++ {
			xx[i] = 0
		}
	}
	if flags&flagICRS == 0 && state.swiGetDenum(bodyIndex, ephemeris) >= 403 {
		state.swiBias(xx[:], t, flags, 0)
	}
	for i = 0; i <= 5; i++ {
		xxsv[i] = xx[i]
	}
	if flags&flagJ2000 == 0 {
		state.swiPrecess(xx[:], planet.evaluationTime, flags, -1)
		if flags&flagSpeed != 0 {
			state.swiPrecessSpeed(xx[:], planet.evaluationTime, flags, -1)
		}
		obliquityData = &state.positions.data.obliquity
	} else {
		obliquityData = &state.positions.data.obliquityJ2000
	}
	return state.appPosRest(planet, flags, xx[:], xxsv[:], obliquityData, message)
}

// appPosRest implements app_pos_rest from sweph.c.
func (state *State) appPosRest(planet *planetData, flags int32, xx []float64, x2000 []float64, obliquityData *obliquity, message []byte) int32 {
	var (
		i    int32
		daya [2]float64
		xxsv [24]float64
	)
	if flags&flagNoNutation == 0 {
		state.swiNutate(xx, flags, 0)
	}
	for i = 0; i <= 5; i++ {
		planet.coordinates[18+i] = xx[i]
	}
	state.swiCoortrf2(xx, xx, obliquityData.sine, obliquityData.cosine)
	if flags&flagSpeed != 0 {
		state.swiCoortrf2(xx[3:], xx[3:], obliquityData.sine, obliquityData.cosine)
	}
	if flags&flagNoNutation == 0 {
		state.swiCoortrf2(xx, xx, state.positions.data.nutation.sine, state.positions.data.nutation.cosine)
		if flags&flagSpeed != 0 {
			state.swiCoortrf2(xx[3:], xx[3:], state.positions.data.nutation.sine, state.positions.data.nutation.cosine)
		}
	}
	for i = 0; i <= 5; i++ {
		planet.coordinates[6+i] = xx[i]
	}
	if flags&flagSidereal != 0 {
		if state.positions.data.sidereal.mode&256 != 0 {
			if state.swiTropRa2sidLon(x2000, planet.coordinates[6:], planet.coordinates[18:], flags) != 0 {
				return -1
			}
		} else if state.positions.data.sidereal.mode&512 != 0 {
			if state.swiTropRa2sidLonSosy(x2000, planet.coordinates[6:], flags) != 0 {
				return -1
			}
		} else {
			state.swiCartpolSp(planet.coordinates[6:], planet.coordinates[:])
			for i = 0; i < 24; i++ {
				xxsv[i] = planet.coordinates[i]
			}
			if state.swiGetAyanamsaWithSpeed(planet.evaluationTime, flags, daya[:], message) == -1 {
				return -1
			}
			for i = 0; i < 24; i++ {
				planet.coordinates[i] = xxsv[i]
			}
			planet.coordinates[0] -= daya[0] * radiansPerDegree
			planet.coordinates[3] -= daya[1] * radiansPerDegree
			state.swiPolcartSp(planet.coordinates[:], planet.coordinates[6:])
		}
	}
	state.swiCartpolSp(planet.coordinates[18:], planet.coordinates[12:])
	state.swiCartpolSp(planet.coordinates[6:], planet.coordinates[:])
	for i = 0; i < 2; i++ {
		planet.coordinates[i] *= degreesPerRadian
		planet.coordinates[i+3] *= degreesPerRadian
		planet.coordinates[i+12] *= degreesPerRadian
		planet.coordinates[i+15] *= degreesPerRadian
	}
	planet.cachedFlags = flags
	planet.ephemeris = flags & (flagJPL | flagSwiss | flagMoshier)
	return 0
}

// sweSetSidMode implements swe_set_sid_mode from sweph.c.
//
//	save, what has been done
func (state *State) sweSetSidMode(sidMode int32, t0 float64, ayanT0 float64) {
	var sip *siderealData = &state.positions.data.sidereal
	state.swiInitSwedIfStart()
	if sidMode < 0 {
		sidMode = 0
	}
	sip.mode = sidMode
	if sidMode >= 256 {
		sidMode %= 256
	}
	if sidMode == 18 || sidMode == 19 || sidMode == 20 || sidMode == 34 {
		sip.mode = sidMode
		sip.mode |= 256
	}
	if sidMode == 27 || sidMode == 28 || sidMode == 29 || sidMode == 39 || sidMode == 35 || sidMode == 17 || sidMode == 40 || sidMode == 30 || sidMode == 36 || sidMode == 31 || sidMode == 32 || sidMode == 33 {
		sip.mode = sidMode
	}
	if sidMode >= 47 && sidMode != 255 {
		sidMode = 0
		sip.mode = sidMode
	}
	state.positions.data.ayanaIsSet = 1
	if sidMode == 255 {
		sip.epoch = t0
		sip.ayanamsa = ayanT0
		sip.epochIsUT = 0
		if sip.mode&1024 != 0 {
			sip.epochIsUT = 1
		}
	} else {
		sip.epoch = calendarAyanamsa[sidMode].t0
		sip.ayanamsa = calendarAyanamsa[sidMode].ayanT0
		sip.epochIsUT = calendarAyanamsa[sidMode].t0IsUT
	}
	if sidMode < 47 && sip.mode&8192 != 0 && calendarAyanamsa[sidMode].precOffset > 0 {
		state.positions.data.astroModels[1] = calendarAyanamsa[sidMode].precOffset
		state.positions.data.astroModels[2] = calendarAyanamsa[sidMode].precOffset
		switch calendarAyanamsa[sidMode].precOffset {
		case 11:
			state.positions.data.astroModels[3] = 5
		case 1:
			state.positions.data.astroModels[3] = 1
		default:
		}
	}
	state.swiForceAppPosEtc()
}

// sweGetAyanamsaEx implements swe_get_ayanamsa_ex from sweph.c.
func (state *State) sweGetAyanamsaEx(jdTT float64, flags int32, daya []float64, message []byte) int32 {
	var (
		nuttmp       nutation
		nutationData *nutation = &nuttmp
		retval       int32     = state.swiGetAyanamsaEx(jdTT, flags, daya, message)
	)
	if flags&flagNoNutation == 0 {
		if jdTT == state.positions.data.nutation.epoch {
			nutationData = &state.positions.data.nutation
		} else {
			nutationData = &nuttmp
			state.swiNutation(jdTT, flags, nutationData.angles[:])
		}
		daya[0] += nutationData.angles[0] * degreesPerRadian
		retval &= ^64
	}
	return retval
}

// getAyaCorrection implements get_aya_correction from sweph.c.
//
// Function calculates a correction for ayanamsha if the ayanamsha was
// defined using a different precession model than our standard one.
// This allows us to use this ayanamsha with our standard precession
// model and still remain very accurate in ephemeris positions.
// It has the effect that ayanamsha values change depending on the precession
// model used, but sidereal planetary positions remain the same.
//
// For this function to work correctly, our standard precession model
// must be relative to J2000. Any future precession model should not
// use a different starting epoch.
func (state *State) getAyaCorrection(flags int32, corr []float64, message []byte) int32 {
	var (
		x              [6]float64
		eps            float64
		t0             float64
		sip            *siderealData = &state.positions.data.sidereal
		precModel      int32         = state.positions.data.astroModels[1]
		precModelShort int32         = state.positions.data.astroModels[2]
		precOffset     int32         = 0
		sidMode        int32         = sip.mode
	)
	sidMode %= 256
	corr[0] = 0
	if sip.epoch == julianDayJ2000 {
		return 0
	}
	if sip.mode&4096 != 0 {
		return 0
	}
	if sidMode < 47 {
		precOffset = calendarAyanamsa[sidMode].precOffset
	}
	if precOffset < 0 {
		precOffset = 0
	}
	if precModel == precOffset {
		return 0
	}
	t0 = sip.epoch
	if sip.epochIsUT != 0 {
		t0 += state.sweDeltatEx(t0, flags, message)
	}
	x[0] = 1
	x[2] = 0
	x[1] = x[2]
	state.swiPrecess(x[:], t0, 0, 1)
	state.positions.data.astroModels[1] = precOffset
	state.positions.data.astroModels[2] = precOffset
	state.swiPrecess(x[:], t0, 0, -1)
	state.positions.data.astroModels[1] = precModel
	state.positions.data.astroModels[2] = precModelShort
	eps = state.swiEpsiln(t0, 0)
	state.swiCoortrf(x[:], x[:], eps)
	state.swiCartpol(x[:], x[:])
	corr[0] = x[0] * degreesPerRadian
	if corr[0] > 350 {
		corr[0] -= 360
	}
	return 0
}

// swiGetAyanamsaEx implements swi_get_ayanamsa_ex from sweph.c.
//
// correct!
func (state *State) swiGetAyanamsaEx(jdTT float64, flags int32, daya []float64, message []byte) int32 {
	var (
		x           [6]float64
		eps         float64
		t0          float64
		corr        [1]float64
		sip         *siderealData = &state.positions.data.sidereal
		star        [256]byte
		ephemeris   int32
		otherflag   int32
		retflag     int32
		iflagTrue   int32
		iflagGalequ int32
		sidMode     int32 = sip.mode
	)
	flags = state.plausIflag(flags, -1, jdTT, message)
	ephemeris = flags & (flagJPL | flagSwiss | flagMoshier)
	otherflag = flags & ^(1 | 2 | 4)
	daya[0] = 0.0
	flags &= flagJPL | flagSwiss | flagMoshier
	flags |= flagNoNutation
	sidMode %= 256
	iflagGalequ = flags | (flagTruePosition)
	iflagTrue = flags
	if otherflag&16 != 0 {
		iflagTrue |= 16
	}
	if otherflag&1024 != 0 {
		iflagTrue |= 1024
	}
	if otherflag&512 != 0 {
		iflagTrue |= 512
	}
	if state.swiInitSwedIfStart() == 1 && !(ephemeris&4 != 0) && (sidMode == 27 || sidMode == 28 || sidMode == 29 || sip.mode == 39 || sidMode == 35 || sidMode == 17 || sidMode == 40 || sidMode == 30 || sidMode == 36 || sidMode == 31 || sidMode == 32 || sidMode == 33) && message != nil {
		copyText(message, textBytes("Please call swe_set_ephe_path() or swe_set_jplfile() before calling swe_get_ayanamsa_ex()"))
	}
	if !(state.positions.data.ayanaIsSet != 0) {
		state.sweSetSidMode(0, 0, 0)
	}
	if sidMode == 27 {
		copyText(star[:], textBytes("Spica"))
		if retflag = state.sweFixstar(star[:], jdTT, iflagTrue, x[:], message); retflag == -1 {
			return -1
		}
		daya[0] = state.sweDegnorm(x[0] - 180)
		return retflag & (1 | 2 | 4)
	}
	if sidMode == 28 {
		copyText(star[:], textBytes(",zePsc"))
		if retflag = state.sweFixstar(star[:], jdTT, iflagTrue, x[:], message); retflag == -1 {
			return -1
		}
		daya[0] = state.sweDegnorm(x[0] - 359.8333333333)
		return retflag & (1 | 2 | 4)
	}
	if sidMode == 29 {
		copyText(star[:], textBytes(",deCnc"))
		if retflag = state.sweFixstar(star[:], jdTT, iflagTrue, x[:], message); retflag == -1 {
			return -1
		}
		daya[0] = state.sweDegnorm(x[0] - 106)
		return retflag & (1 | 2 | 4)
	}
	if sidMode == 39 {
		copyText(star[:], textBytes(",deCnc"))
		if retflag = state.sweFixstar(star[:], jdTT, iflagTrue, x[:], message); retflag == -1 {
			return -1
		}
		daya[0] = state.sweDegnorm(x[0] - 103.49264221625)
		return retflag & (1 | 2 | 4)
	}
	if sidMode == 35 {
		copyText(star[:], textBytes(",laSco"))
		if retflag = state.sweFixstar(star[:], jdTT, iflagTrue, x[:], message); retflag == -1 {
			return -1
		}
		daya[0] = state.sweDegnorm(x[0] - 240)
		return retflag & (1 | 2 | 4)
	}
	if sidMode == 17 {
		copyText(star[:], textBytes(",SgrA*"))
		if retflag = state.sweFixstar(star[:], jdTT, iflagTrue, x[:], message); retflag == -1 {
			return -1
		}
		daya[0] = state.sweDegnorm(x[0] - 240.0)
		return retflag & (1 | 2 | 4)
	}
	if sidMode == 40 {
		copyText(star[:], textBytes(",SgrA*"))
		if retflag = state.sweFixstar(star[:], jdTT, iflagTrue, x[:], message); retflag == -1 {
			return -1
		}
		daya[0] = state.sweDegnorm(x[0] - 270.0)
		return retflag & (1 | 2 | 4)
	}
	if sidMode == 30 {
		copyText(star[:], textBytes(",SgrA*"))
		if retflag = state.sweFixstar(star[:], jdTT, iflagTrue, x[:], message); retflag == -1 {
			return -1
		}
		daya[0] = state.sweDegnorm(x[0] - 210.0 - 34.376941017)
		return retflag & (1 | 2 | 4)
	}
	if sidMode == 36 {
		copyText(star[:], textBytes(",SgrA*"))
		if retflag = state.sweFixstar(star[:], jdTT, iflagTrue|2*1024, x[:], message); retflag == -1 {
			return -1
		}
		eps = state.swiEpsiln(jdTT, flags) * degreesPerRadian
		daya[0] = state.swiArmcToMc(x[0], eps)
		daya[0] = state.sweDegnorm(daya[0] - 246.6666666667)
		return retflag & (1 | 2 | 4)
	}
	if sidMode == 31 {
		copyText(star[:], textBytes(",GP1958"))
		if retflag = state.sweFixstar(star[:], jdTT, iflagGalequ, x[:], message); retflag == -1 {
			return -1
		}
		daya[0] = state.sweDegnorm(x[0] - 150)
		return retflag & (1 | 2 | 4)
	}
	if sidMode == 32 {
		copyText(star[:], textBytes(",GPol"))
		if retflag = state.sweFixstar(star[:], jdTT, iflagGalequ, x[:], message); retflag == -1 {
			return -1
		}
		daya[0] = state.sweDegnorm(x[0] - 150)
		return retflag & (1 | 2 | 4)
	}
	if sidMode == 33 {
		copyText(star[:], textBytes(",GPol"))
		if retflag = state.sweFixstar(star[:], jdTT, iflagGalequ, x[:], message); retflag == -1 {
			return -1
		}
		daya[0] = state.sweDegnorm(x[0] - 150 - 6.6666666667)
		return retflag & (1 | 2 | 4)
	}
	if !(sip.mode&2048 != 0) {
		x[0] = 1
		x[5] = 0
		x[4] = x[5]
		x[3] = x[4]
		x[2] = x[3]
		x[1] = x[2]
		if jdTT != julianDayJ2000 {
			state.swiPrecess(x[:], jdTT, 0, 1)
		}
		t0 = sip.epoch
		if sip.epochIsUT != 0 {
			t0 += state.sweDeltatEx(t0, flags, message)
		}
		state.swiPrecess(x[:], t0, 0, -1)
		eps = state.swiEpsiln(t0, 0)
		state.swiCoortrf(x[:], x[:], eps)
		state.swiCartpol(x[:], x[:])
		x[0] = -x[0]*degreesPerRadian + sip.ayanamsa
	} else {
		x[0] = state.sweDegnorm(sip.ayanamsa) * radiansPerDegree
		x[1] = 0
		x[2] = 1
		t0 = sip.epoch
		if sip.epochIsUT != 0 {
			t0 += state.sweDeltatEx(t0, flags, message)
		}
		eps = state.swiEpsiln(t0, 0)
		state.swiPolcart(x[:], x[:])
		state.swiCoortrf(x[:], x[:], -eps)
		if t0 != julianDayJ2000 {
			state.swiPrecess(x[:], t0, 0, 1)
		}
		state.swiPrecess(x[:], jdTT, 0, -1)
		eps = state.swiEpsiln(jdTT, 0)
		state.swiCoortrf(x[:], x[:], eps)
		state.swiCartpol(x[:], x[:])
		x[0] = state.sweDegnorm(x[0] * degreesPerRadian)
	}
	state.getAyaCorrection(flags, corr[:], message)
	daya[0] = state.sweDegnorm(x[0] - corr[0])
	return flags
}

// swiGetAyanamsaWithSpeed implements swi_get_ayanamsa_with_speed from sweph.c.
//
//	get ayanamsa
func (state *State) swiGetAyanamsaWithSpeed(jdTT float64, flags int32, daya []float64, message []byte) int32 {
	var (
		dayaT2  [1]float64
		t2      float64
		tintv   float64 = 0.001
		retflag int32
	)
	t2 = jdTT - tintv
	retflag = state.swiGetAyanamsaEx(t2, flags, dayaT2[:], message)
	if retflag == -1 {
		return -1
	}
	retflag = state.swiGetAyanamsaEx(jdTT, flags, daya, message)
	if retflag == -1 {
		return -1
	}
	daya[1] = (daya[0] - dayaT2[0]) / tintv
	return retflag
}

// sweGetAyanamsaExUt implements swe_get_ayanamsa_ex_ut from sweph.c.
func (state *State) sweGetAyanamsaExUt(jdUT float64, flags int32, daya []float64, message []byte) int32 {
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
	retflag = state.sweGetAyanamsaEx(jdUT+deltat, flags, daya, message)
	if retflag&(1|2|4) != ephemeris {
		deltat = state.sweDeltatEx(jdUT, retflag, message)
		retflag = state.sweGetAyanamsaEx(jdUT+deltat, flags, daya, message)
	}
	return retflag
}

// sweGetAyanamsa implements swe_get_ayanamsa from sweph.c.
//
//	the ayanamsa (precession in longitude)
//
// according to Newcomb's definition: 360 -
// longitude of the vernal point of t referred to the
// ecliptic of t0.
func (state *State) sweGetAyanamsa(jdTT float64) float64 {
	var (
		daya  [1]float64
		flags int32 = state.swiGuessEpheFlag()
	)
	state.swiGetAyanamsaEx(jdTT, flags, daya[:], nil)
	return daya[0]
}

// sweGetAyanamsaUt implements swe_get_ayanamsa_ut from sweph.c.
func (state *State) sweGetAyanamsaUt(jdUT float64) float64 {
	var (
		daya  [1]float64
		flags int32 = state.swiGuessEpheFlag()
	)
	state.swiGetAyanamsaEx(jdUT+state.sweDeltatEx(jdUT, flags, nil), 0, daya[:], nil)
	return daya[0]
}

// swiTropRa2sidLon implements swi_trop_ra2sid_lon from sweph.c.
func (state *State) swiTropRa2sidLon(xin []float64, xout []float64, xoutr []float64, flags int32) int32 {
	var (
		x      [6]float64
		corr   [1]float64
		i      int32
		sip    *siderealData = &state.positions.data.sidereal
		oectmp obliquity
	)
	for i = 0; i <= 5; i++ {
		x[i] = xin[i]
	}
	if sip.epoch != julianDayJ2000 {
		state.swiPrecess(x[:], sip.epoch, 0, -1)
		state.swiPrecess(x[3:], sip.epoch, 0, -1)
	}
	for i = 0; i <= 5; i++ {
		xoutr[i] = x[i]
	}
	state.calcEpsilon(state.positions.data.sidereal.epoch, flags, &oectmp)
	state.swiCoortrf2(x[:], x[:], oectmp.sine, oectmp.cosine)
	if flags&flagSpeed != 0 {
		state.swiCoortrf2(x[3:], x[3:], oectmp.sine, oectmp.cosine)
	}
	state.swiCartpolSp(x[:], x[:])
	state.getAyaCorrection(flags, corr[:], nil)
	x[0] -= sip.ayanamsa * radiansPerDegree
	x[0] = state.sweRadnorm(x[0] + corr[0]*radiansPerDegree)
	state.swiPolcartSp(x[:], xout)
	return 0
}

// swiTropRa2sidLonSosy implements swi_trop_ra2sid_lon_sosy from sweph.c.
//
// input coordinates are J2000, cartesian.
// xout 	ecliptical sidereal position
// xoutr 	equatorial sidereal position
func (state *State) swiTropRa2sidLonSosy(xin []float64, xout []float64, flags int32) int32 {
	var (
		x             [6]float64
		x0            [6]float64
		corr          [1]float64
		i             int32
		sip           *siderealData = &state.positions.data.sidereal
		obliquityData *obliquity    = &state.positions.data.obliquityJ2000
		planeNode     float64       = 1.8776700468039835
		planeIncl     float64       = 0.027553530354526998
	)
	for i = 0; i <= 5; i++ {
		x[i] = xin[i]
	}
	state.swiCoortrf2(x[:], x[:], obliquityData.sine, obliquityData.cosine)
	if flags&flagSpeed != 0 {
		state.swiCoortrf2(x[3:], x[3:], obliquityData.sine, obliquityData.cosine)
	}
	state.swiCartpolSp(x[:], x[:])
	x[0] -= planeNode
	state.swiPolcartSp(x[:], x[:])
	state.swiCoortrf(x[:], x[:], planeIncl)
	state.swiCoortrf(x[3:], x[3:], planeIncl)
	state.swiCartpolSp(x[:], x[:])
	x0[0] = 1
	x0[2] = 0
	x0[1] = x0[2]
	if sip.epoch != julianDayJ2000 {
		state.swiPrecess(x0[:], sip.epoch, 0, 1)
	}
	state.swiCoortrf2(x0[:], x0[:], obliquityData.sine, obliquityData.cosine)
	state.swiCartpol(x0[:], x0[:])
	x0[0] -= planeNode
	state.swiPolcart(x0[:], x0[:])
	state.swiCoortrf(x0[:], x0[:], planeIncl)
	state.swiCartpol(x0[:], x0[:])
	x[0] -= x0[0]
	x[0] *= degreesPerRadian
	state.getAyaCorrection(flags, corr[:], nil)
	x[0] -= sip.ayanamsa
	x[0] = state.sweDegnorm(x[0]+corr[0]) * radiansPerDegree
	state.swiPolcartSp(x[:], xout)
	return 0
}

// appPosEtcPlanOsc implements app_pos_etc_plan_osc from sweph.c.
func (state *State) appPosEtcPlanOsc(body int32, bodyIndex int32, flags int32, message []byte) int32 {
	var (
		i              int32
		j              int32
		niter          int32
		status         int32
		xx             [6]float64
		dx             [3]float64
		dt             float64
		dtsaveForDefl  float64
		earthVector    [6]float64
		sunVector      [6]float64
		moonVector     [6]float64
		xxsv           [6]float64
		xxsp           [3]float64 = [3]float64{0}
		observerVector [6]float64
		xobs2          [6]float64
		t              float64
		planet         *planetData = &state.positions.data.planets[bodyIndex]
		earthData      *planetData = &state.positions.data.planets[0]
		sunData        *planetData = &state.positions.data.planets[10]
		obliquityData  *obliquity  = &state.positions.data.obliquityJ2000
		ephemeris      int32       = 2
	)
	dtsaveForDefl = 0
	dt = dtsaveForDefl
	if flags&flagMoshier != 0 {
		ephemeris = 4
	} else if flags&flagSwiss != 0 {
		ephemeris = 2
	} else if flags&flagJPL != 0 {
		ephemeris = 1
	}
	for i = 0; i <= 5; i++ {
		xx[i] = planet.stateVector[i]
	}
	if flags&flagTopocentric != 0 {
		if state.positions.data.observer.evaluationTime != earthData.evaluationTime || state.positions.data.observer.evaluationTime == 0 {
			if state.swiGetObserver(earthData.evaluationTime, flags|(flagNoNutation), 1, observerVector[:], message) != 0 {
				return -1
			}
		} else {
			for i = 0; i <= 5; i++ {
				observerVector[i] = state.positions.data.observer.stateVector[i]
			}
		}
		for i = 0; i <= 5; i++ {
			observerVector[i] = observerVector[i] + earthData.stateVector[i]
		}
	} else if flags&flagBarycentric != 0 {
		for i = 0; i <= 5; i++ {
			observerVector[i] = 0
		}
	} else if flags&flagHeliocentric != 0 {
		if flags&flagMoshier != 0 {
			for i = 0; i <= 5; i++ {
				observerVector[i] = 0
			}
		} else {
			for i = 0; i <= 5; i++ {
				observerVector[i] = sunData.stateVector[i]
			}
		}
	} else {
		for i = 0; i <= 5; i++ {
			observerVector[i] = earthData.stateVector[i]
		}
	}
	if flags&flagTruePosition == 0 {
		niter = 1
		if flags&flagSpeed != 0 {
			for i = 0; i <= 2; i++ {
				xxsp[i] = xx[i] - xx[i+3]
				xxsv[i] = xxsp[i]
			}
			for j = 0; j <= niter; j++ {
				for i = 0; i <= 2; i++ {
					dx[i] = xxsp[i]
					if flags&flagHeliocentric == 0 && flags&flagBarycentric == 0 {
						dx[i] -= observerVector[i] - observerVector[i+3]
					}
				}
				dt = math.Sqrt(dx[0]*dx[0]+dx[1]*dx[1]+dx[2]*dx[2]) * 1.49597870700e+11 / 2.99792458e+8 / secondsPerDay
				for i = 0; i <= 2; i++ {
					xxsp[i] = xxsv[i] - dt*planet.stateVector[i+3]
				}
			}
			for i = 0; i <= 2; i++ {
				xxsp[i] = xxsv[i] - xxsp[i]
			}
		}
		for j = 0; j <= niter; j++ {
			for i = 0; i <= 2; i++ {
				dx[i] = xx[i]
				if flags&flagHeliocentric == 0 && flags&flagBarycentric == 0 {
					dx[i] -= observerVector[i]
				}
			}
			dt = math.Sqrt(dx[0]*dx[0]+dx[1]*dx[1]+dx[2]*dx[2]) * 1.49597870700e+11 / 2.99792458e+8 / secondsPerDay
			dtsaveForDefl = dt
			for i = 0; i <= 2; i++ {
				xx[i] = planet.stateVector[i] - dt*planet.stateVector[i+3]
				xx[i+3] = planet.stateVector[i+3]
			}
		}
		if flags&flagSpeed != 0 {
			for i = 0; i <= 2; i++ {
				xxsp[i] = planet.stateVector[i] - xx[i] - xxsp[i]
			}
			t = planet.evaluationTime - dt
			status = state.mainPlanetBary(t, 0, ephemeris, flags, 0, earthVector[:], earthVector[:], sunVector[:], moonVector[:], message)
			if state.swiOscElPlan(t, xx[:], body-40, bodyIndex, earthVector[:], sunVector[:], message) != 0 {
				return -1
			}
			if status != 0 {
				return status
			}
			if flags&flagTopocentric != 0 {
				if state.swiGetObserver(t, flags|(flagNoNutation), 0, xobs2[:], message) != 0 {
					return -1
				}
				for i = 0; i <= 5; i++ {
					xobs2[i] += earthVector[i]
				}
			} else {
				for i = 0; i <= 5; i++ {
					xobs2[i] = earthVector[i]
				}
			}
		}
	}
	for i = 0; i <= 5; i++ {
		xx[i] -= observerVector[i]
	}
	if flags&flagTruePosition == 0 {
		if flags&flagSpeed != 0 {
			for i = 3; i <= 5; i++ {
				xx[i] -= xxsp[i-3]
			}
		}
	}
	if flags&flagSpeed == 0 {
		for i = 3; i <= 5; i++ {
			xx[i] = 0
		}
	}
	if flags&flagTruePosition == 0 && flags&flagNoDeflection == 0 {
		state.swiDeflectLight(xx[:], dtsaveForDefl, flags)
	}
	if flags&flagTruePosition == 0 && flags&flagNoAberration == 0 {
		state.swiAberrLight(xx[:], observerVector[:], flags)
		if flags&flagSpeed != 0 {
			for i = 3; i <= 5; i++ {
				xx[i] += observerVector[i] - xobs2[i]
			}
		}
	}
	for i = 0; i <= 5; i++ {
		xxsv[i] = xx[i]
	}
	if flags&flagJ2000 == 0 {
		state.swiPrecess(xx[:], planet.evaluationTime, flags, -1)
		if flags&flagSpeed != 0 {
			state.swiPrecessSpeed(xx[:], planet.evaluationTime, flags, -1)
		}
		obliquityData = &state.positions.data.obliquity
	} else {
		obliquityData = &state.positions.data.obliquityJ2000
	}
	return state.appPosRest(planet, flags, xx[:], xxsv[:], obliquityData, message)
}

// swiPrecessSpeed implements swi_precess_speed from sweph.c.
//
//	influence of precession on speed
//
// xx		position and speed of planet in equatorial cartesian
//
//	coordinates
func (state *State) swiPrecessSpeed(xx []float64, t float64, flags int32, direction int32) {
	var (
		obliquityData *obliquity
		fac           float64
		dpre          [1]float64
		dpre2         [1]float64
		tprec         float64 = (t - julianDayJ2000) / daysPerCentury
		precModel     int32   = state.positions.data.astroModels[1]
	)
	if precModel == 0 {
		precModel = 9
	}
	if direction == -1 {
		fac = 1
		obliquityData = &state.positions.data.obliquity
	} else {
		fac = float64(-1)
		obliquityData = &state.positions.data.obliquityJ2000
	}
	state.swiPrecess(xx[3:], t, flags, direction)
	state.swiCoortrf2(xx, xx, obliquityData.sine, obliquityData.cosine)
	state.swiCoortrf2(xx[3:], xx[3:], obliquityData.sine, obliquityData.cosine)
	state.swiCartpolSp(xx, xx)
	if 1 != 0 {
		if precModel == 9 {
			state.swiLdpPeps(t, dpre[:], nil)
			state.swiLdpPeps(t+1, dpre2[:], nil)
			xx[3] += (dpre2[0] - dpre[0]) * fac
		} else {
			xx[3] += (50.290966 + 0.0222226*tprec) / 3600 / 365.25 * radiansPerDegree * fac
		}
	}
	state.swiPolcartSp(xx, xx)
	state.swiCoortrf2(xx, xx, -obliquityData.sine, obliquityData.cosine)
	state.swiCoortrf2(xx[3:], xx[3:], -obliquityData.sine, obliquityData.cosine)
}

// swiNutate implements swi_nutate from sweph.c.
//
//	multiplies cartesian equatorial coordinates with previously
//
// calculated nutation matrix. also corrects speed.
func (state *State) swiNutate(xx []float64, flags int32, backward int32) {
	var (
		i  int32
		x  [6]float64
		xv [6]float64
	)
	for i = 0; i <= 2; i++ {
		if backward != 0 {
			x[i] = xx[0]*state.positions.data.nutation.matrix[i][0] + xx[1]*state.positions.data.nutation.matrix[i][1] + xx[2]*state.positions.data.nutation.matrix[i][2]
		} else {
			x[i] = xx[0]*state.positions.data.nutation.matrix[0][i] + xx[1]*state.positions.data.nutation.matrix[1][i] + xx[2]*state.positions.data.nutation.matrix[2][i]
		}
	}
	if flags&flagSpeed != 0 {
		for i = 0; i <= 2; i++ {
			if backward != 0 {
				x[i+3] = xx[3]*state.positions.data.nutation.matrix[i][0] + xx[4]*state.positions.data.nutation.matrix[i][1] + xx[5]*state.positions.data.nutation.matrix[i][2]
			} else {
				x[i+3] = xx[3]*state.positions.data.nutation.matrix[0][i] + xx[4]*state.positions.data.nutation.matrix[1][i] + xx[5]*state.positions.data.nutation.matrix[2][i]
			}
		}
		for i = 0; i <= 2; i++ {
			if backward != 0 {
				xv[i] = xx[0]*state.positions.data.nutationWithSpeed.matrix[i][0] + xx[1]*state.positions.data.nutationWithSpeed.matrix[i][1] + xx[2]*state.positions.data.nutationWithSpeed.matrix[i][2]
			} else {
				xv[i] = xx[0]*state.positions.data.nutationWithSpeed.matrix[0][i] + xx[1]*state.positions.data.nutationWithSpeed.matrix[1][i] + xx[2]*state.positions.data.nutationWithSpeed.matrix[2][i]
			}
			xx[3+i] = x[3+i] + (x[i]-xv[i])/0.0001
		}
	}
	for i = 0; i <= 2; i++ {
		xx[i] = x[i]
	}
}

// aberrLight implements aberr_light from sweph.c.
//
//	computes 'annual' aberration
//
// xx		planet's position accounted for light-time
//
//	and gravitational light deflection
//
// xe    	earth's position and speed
func (state *State) aberrLight(xx []float64, xe []float64) {
	var (
		i   int32
		xxs [6]float64
		v   [6]float64
		u   [6]float64
		ru  float64
		b1  float64
		f1  float64
		f2  float64
		v2  float64
	)
	for i = 0; i <= 5; i++ {
		xxs[i] = xx[i]
		u[i] = xxs[i]
	}
	ru = math.Sqrt(u[0]*u[0] + u[1]*u[1] + u[2]*u[2])
	for i = 0; i <= 2; i++ {
		v[i] = xe[i+3] / 24.0 / 3600.0 / 2.99792458e+8 * 1.49597870700e+11
	}
	v2 = v[0]*v[0] + v[1]*v[1] + v[2]*v[2]
	b1 = math.Sqrt(1 - v2)
	f1 = (u[0]*v[0] + u[1]*v[1] + u[2]*v[2]) / ru
	f2 = 1.0 + f1/(1.0+b1)
	for i = 0; i <= 2; i++ {
		xx[i] = (b1*xx[i] + f2*ru*v[i]) / (1.0 + f1)
	}
}

// swiAberrLightEx implements swi_aberr_light_ex from sweph.c.
//
//	computes 'annual' aberration
//
// xx		planet's position accounted for light-time
//
//	and gravitational light deflection
//
// xe    	earth's position and speed
// xe_dt    	earth's position and speed at t - dt
// dt    	time difference for which xe_dt is given
func (state *State) swiAberrLightEx(xx []float64, xe []float64, xeDt []float64, dt float64, flags int32) {
	var (
		i   int32
		xxs [6]float64
		xx2 [6]float64
	)
	for i = 0; i <= 5; i++ {
		xxs[i] = xx[i]
	}
	state.aberrLight(xx, xe)
	if flags&flagSpeed != 0 {
		for i = 0; i <= 2; i++ {
			xx2[i] = xxs[i] - dt*xxs[i+3]
		}
		state.aberrLight(xx2[:], xeDt)
		for i = 0; i <= 2; i++ {
			xx[i+3] = (xx[i] - xx2[i]) / dt
		}
	}
}

// swiAberrLight implements swi_aberr_light from sweph.c.
//
//	computes 'annual' aberration
//
// xx		planet's position accounted for light-time
//
//	and gravitational light deflection
//
// xe    	earth's position and speed
func (state *State) swiAberrLight(xx []float64, xe []float64, flags int32) {
	var (
		i    int32
		xxs  [6]float64
		v    [6]float64
		u    [6]float64
		ru   float64
		xx2  [6]float64
		dx1  float64
		dx2  float64
		b1   float64
		f1   float64
		f2   float64
		v2   float64
		intv float64 = 0.0001
	)
	for i = 0; i <= 5; i++ {
		xxs[i] = xx[i]
		u[i] = xxs[i]
	}
	ru = math.Sqrt(u[0]*u[0] + u[1]*u[1] + u[2]*u[2])
	for i = 0; i <= 2; i++ {
		v[i] = xe[i+3] / 24.0 / 3600.0 / 2.99792458e+8 * 1.49597870700e+11
	}
	v2 = v[0]*v[0] + v[1]*v[1] + v[2]*v[2]
	b1 = math.Sqrt(1 - v2)
	f1 = (u[0]*v[0] + u[1]*v[1] + u[2]*v[2]) / ru
	f2 = 1.0 + f1/(1.0+b1)
	for i = 0; i <= 2; i++ {
		xx[i] = (b1*xx[i] + f2*ru*v[i]) / (1.0 + f1)
	}
	if flags&flagSpeed != 0 {
		for i = 0; i <= 2; i++ {
			u[i] = xxs[i] - intv*xxs[i+3]
		}
		ru = math.Sqrt(u[0]*u[0] + u[1]*u[1] + u[2]*u[2])
		f1 = (u[0]*v[0] + u[1]*v[1] + u[2]*v[2]) / ru
		f2 = 1.0 + f1/(1.0+b1)
		for i = 0; i <= 2; i++ {
			xx2[i] = (b1*u[i] + f2*ru*v[i]) / (1.0 + f1)
		}
		for i = 0; i <= 2; i++ {
			dx1 = xx[i] - xxs[i]
			dx2 = xx2[i] - u[i]
			dx1 -= dx2
			xx[i+3] += dx1 / intv
		}
	}
}

// swiDeflectLight implements swi_deflect_light from sweph.c.
//
//	computes relativistic light deflection by the sun
//
// ipli 	sweph internal planet number
// xx		planet's position accounted for light-time
// dt		dt of light-time
func (state *State) swiDeflectLight(xx []float64, dt float64, flags int32) {
	var (
		i           int32
		xx2         [6]float64
		u           [6]float64
		e           [6]float64
		q           [6]float64
		ru          float64
		re          float64
		rq          float64
		uq          float64
		ue          float64
		qe          float64
		g1          float64
		g2          float64
		xx3         [6]float64
		dx1         float64
		dx2         float64
		dtsp        float64
		sunVector   [6]float64
		earthVector [6]float64
		sina        float64
		sinSunr     float64
		meffFact    float64
		earthData   *planetData = &state.positions.data.planets[0]
		sunData     *planetData = &state.positions.data.planets[10]
		iephe       int32       = earthData.ephemeris
	)
	for i = 0; i <= 5; i++ {
		earthVector[i] = earthData.stateVector[i]
	}
	if flags&flagTopocentric != 0 {
		for i = 0; i <= 5; i++ {
			earthVector[i] += state.positions.data.observer.stateVector[i]
		}
	}
	for i = 0; i <= 2; i++ {
		u[i] = xx[i]
	}
	if iephe == 1 || iephe == 2 {
		for i = 0; i <= 2; i++ {
			e[i] = earthVector[i] - sunData.stateVector[i]
		}
	} else {
		for i = 0; i <= 2; i++ {
			e[i] = earthVector[i]
		}
	}
	if iephe == 1 || iephe == 2 {
		for i = 0; i <= 2; i++ {
			sunVector[i] = sunData.stateVector[i] - dt*sunData.stateVector[i+3]
		}
		for i = 3; i <= 5; i++ {
			sunVector[i] = sunData.stateVector[i]
		}
	} else {
		for i = 0; i <= 5; i++ {
			sunVector[i] = sunData.stateVector[i]
		}
	}
	for i = 0; i <= 2; i++ {
		q[i] = xx[i] + earthVector[i] - sunVector[i]
	}
	ru = math.Sqrt(u[0]*u[0] + u[1]*u[1] + u[2]*u[2])
	rq = math.Sqrt(q[0]*q[0] + q[1]*q[1] + q[2]*q[2])
	re = math.Sqrt(e[0]*e[0] + e[1]*e[1] + e[2]*e[2])
	for i = 0; i <= 2; i++ {
		u[i] /= ru
		q[i] /= rq
		e[i] /= re
	}
	uq = u[0]*q[0] + u[1]*q[1] + u[2]*q[2]
	ue = u[0]*e[0] + u[1]*e[1] + u[2]*e[2]
	qe = q[0]*e[0] + q[1]*e[1] + q[2]*e[2]
	sina = math.Sqrt(1 - ue*ue)
	sinSunr = 0.0046524175280314396 / re
	if sina < sinSunr {
		meffFact = state.meff(sina / sinSunr)
	} else {
		meffFact = 1
	}
	g1 = 2.6542488003597399e+20 * meffFact / 2.99792458e+8 / 2.99792458e+8 / 1.49597870700e+11 / re
	g2 = 1.0 + qe
	for i = 0; i <= 2; i++ {
		xx2[i] = ru * (u[i] + g1/g2*(uq*e[i]-ue*q[i]))
	}
	if flags&flagSpeed != 0 {
		dtsp = -0.0000005
		for i = 0; i <= 2; i++ {
			u[i] = xx[i] - dtsp*xx[i+3]
		}
		if iephe == 1 || iephe == 2 {
			for i = 0; i <= 2; i++ {
				e[i] = earthVector[i] - sunData.stateVector[i] - dtsp*(earthVector[i+3]-sunData.stateVector[i+3])
			}
		} else {
			for i = 0; i <= 2; i++ {
				e[i] = earthVector[i] - dtsp*earthVector[i+3]
			}
		}
		for i = 0; i <= 2; i++ {
			q[i] = u[i] + earthVector[i] - sunVector[i] - dtsp*(earthVector[i+3]-sunVector[i+3])
		}
		ru = math.Sqrt(u[0]*u[0] + u[1]*u[1] + u[2]*u[2])
		rq = math.Sqrt(q[0]*q[0] + q[1]*q[1] + q[2]*q[2])
		re = math.Sqrt(e[0]*e[0] + e[1]*e[1] + e[2]*e[2])
		for i = 0; i <= 2; i++ {
			u[i] /= ru
			q[i] /= rq
			e[i] /= re
		}
		uq = u[0]*q[0] + u[1]*q[1] + u[2]*q[2]
		ue = u[0]*e[0] + u[1]*e[1] + u[2]*e[2]
		qe = q[0]*e[0] + q[1]*e[1] + q[2]*e[2]
		sina = math.Sqrt(1 - ue*ue)
		sinSunr = 0.0046524175280314396 / re
		if sina < sinSunr {
			meffFact = state.meff(sina / sinSunr)
		} else {
			meffFact = 1
		}
		g1 = 2.6542488003597399e+20 * meffFact / 2.99792458e+8 / 2.99792458e+8 / 1.49597870700e+11 / re
		g2 = 1.0 + qe
		for i = 0; i <= 2; i++ {
			xx3[i] = ru * (u[i] + g1/g2*(uq*e[i]-ue*q[i]))
		}
		for i = 0; i <= 2; i++ {
			dx1 = xx2[i] - xx[i]
			dx2 = xx3[i] - u[i]*ru
			dx1 -= dx2
			xx[i+3] += dx1 / dtsp
		}
	}
	for i = 0; i <= 2; i++ {
		xx[i] = xx2[i]
	}
}

// appPosEtcSun implements app_pos_etc_sun from sweph.c.
func (state *State) appPosEtcSun(flags int32, message []byte) int32 {
	var (
		i              int32
		j              int32
		niter          int32
		status         int32 = 0
		flg1           int32
		flg2           int32
		xx             [6]float64
		xxsv           [6]float64
		dx             [3]float64
		dt             float64
		t              float64 = 0
		earthVector    [6]float64
		sunVector      [6]float64
		observerVector [6]float64
		earthData      *planetData = &state.positions.data.planets[0]
		sunData        *planetData = &state.positions.data.planets[10]
		obliquityData  *obliquity  = &state.positions.data.obliquityJ2000
	)
	flg1 = flags & ^(2 * 1024) & ^(4 * 1024)
	flg2 = earthData.cachedFlags & ^(2 * 1024) & ^(4 * 1024)
	if flg1 == flg2 {
		earthData.cachedFlags = flags
		earthData.ephemeris = flags & (flagJPL | flagSwiss | flagMoshier)
		return 0
	}
	if flags&flagTopocentric != 0 {
		if state.positions.data.observer.evaluationTime != earthData.evaluationTime || state.positions.data.observer.evaluationTime == 0 {
			if state.swiGetObserver(earthData.evaluationTime, flags|(flagNoNutation), 1, observerVector[:], message) != 0 {
				return -1
			}
		} else {
			for i = 0; i <= 5; i++ {
				observerVector[i] = state.positions.data.observer.stateVector[i]
			}
		}
		for i = 0; i <= 5; i++ {
			observerVector[i] = observerVector[i] + earthData.stateVector[i]
		}
	} else {
		for i = 0; i <= 5; i++ {
			observerVector[i] = earthData.stateVector[i]
		}
	}
	if earthData.ephemeris == 4 || flags&flagBarycentric != 0 {
		for i = 0; i <= 5; i++ {
			xx[i] = observerVector[i]
		}
	} else {
		for i = 0; i <= 5; i++ {
			xx[i] = observerVector[i] - sunData.stateVector[i]
		}
	}
	if flags&flagTruePosition == 0 {
		if earthData.ephemeris == 1 || earthData.ephemeris == 2 || flags&flagHeliocentric != 0 || flags&flagBarycentric != 0 {
			for i = 0; i <= 5; i++ {
				earthVector[i] = observerVector[i]
				if earthData.ephemeris == 4 {
					sunVector[i] = 0
				} else {
					sunVector[i] = sunData.stateVector[i]
				}
			}
			niter = 1
			for j = 0; j <= niter; j++ {
				for i = 0; i <= 2; i++ {
					dx[i] = earthVector[i]
					if flags&flagBarycentric == 0 {
						dx[i] -= sunVector[i]
					}
				}
				dt = math.Sqrt(dx[0]*dx[0]+dx[1]*dx[1]+dx[2]*dx[2]) * 1.49597870700e+11 / 2.99792458e+8 / secondsPerDay
				t = earthData.evaluationTime - dt
				switch earthData.ephemeris {
				case 1:
					if flags&flagHeliocentric != 0 || flags&flagBarycentric != 0 {
						status = state.swiPleph(t, 2, 11, earthVector[:], message)
					} else {
						status = state.swiPleph(t, 10, 11, sunVector[:], message)
					}
					if status != 0 {
						state.swiCloseJplFile()
						state.positions.data.jplFileIsOpen = 0
						return status
					}
				case 2:
					if flags&flagHeliocentric != 0 || flags&flagBarycentric != 0 {
						status = state.sweplan(t, 0, 0, flags, 0, earthVector[:], nil, sunVector[:], nil, message)
					} else {
						status = state.sweph(t, 10, 0, flags, nil, 0, sunVector[:], message)
					}
				case 4:
					if flags&flagHeliocentric != 0 || flags&flagBarycentric != 0 {
						status = state.swiMoshplan(t, 0, 0, earthVector[:], earthVector[:], message)
					}
				default:
					status = -1
				}
				if status != 0 {
					return status
				}
			}
			for i = 0; i <= 5; i++ {
				xx[i] = earthVector[i]
				if flags&flagBarycentric == 0 {
					xx[i] -= sunVector[i]
				}
			}
		}
	}
	if flags&flagSpeed == 0 {
		for i = 3; i <= 5; i++ {
			xx[i] = 0
		}
	}
	if flags&flagHeliocentric == 0 && flags&flagBarycentric == 0 {
		for i = 0; i <= 5; i++ {
			xx[i] = -xx[i]
		}
	}
	if flags&flagTruePosition == 0 && flags&flagNoAberration == 0 {
		state.swiAberrLight(xx[:], observerVector[:], flags)
	}
	if flags&flagSpeed == 0 {
		for i = 3; i <= 5; i++ {
			xx[i] = 0
		}
	}
	if flags&flagICRS == 0 && state.swiGetDenum(0, flags) >= 403 {
		state.swiBias(xx[:], t, flags, 0)
	}
	for i = 0; i <= 5; i++ {
		xxsv[i] = xx[i]
	}
	if flags&flagJ2000 == 0 {
		state.swiPrecess(xx[:], earthData.evaluationTime, flags, -1)
		if flags&flagSpeed != 0 {
			state.swiPrecessSpeed(xx[:], earthData.evaluationTime, flags, -1)
		}
		obliquityData = &state.positions.data.obliquity
	} else {
		obliquityData = &state.positions.data.obliquityJ2000
	}
	return state.appPosRest(earthData, flags, xx[:], xxsv[:], obliquityData, message)
}

// appPosEtcMoon implements app_pos_etc_moon from sweph.c.
func (state *State) appPosEtcMoon(flags int32, message []byte) int32 {
	var (
		i              int32
		flg1           int32
		flg2           int32
		xx             [6]float64
		xxsv           [6]float64
		observerVector [6]float64
		xxm            [6]float64
		xs             [6]float64
		xe             [6]float64
		xobs2          [6]float64
		dt             float64
		earthData      *planetData = &state.positions.data.planets[0]
		sunData        *planetData = &state.positions.data.planets[10]
		planet         *planetData = &state.positions.data.planets[1]
		obliquityData  *obliquity  = &state.positions.data.obliquity
		t              float64     = 0
		status         int32
	)
	flg1 = flags & ^(2 * 1024) & ^(4 * 1024)
	flg2 = planet.cachedFlags & ^(2 * 1024) & ^(4 * 1024)
	if flg1 == flg2 {
		planet.cachedFlags = flags
		planet.ephemeris = flags & (flagJPL | flagSwiss | flagMoshier)
		return 0
	}
	for i = 0; i <= 5; i++ {
		xx[i] = planet.stateVector[i]
		xxm[i] = xx[i]
	}
	for i = 0; i <= 5; i++ {
		xx[i] += earthData.stateVector[i]
	}
	if flags&flagTopocentric != 0 {
		if state.positions.data.observer.evaluationTime != planet.evaluationTime || state.positions.data.observer.evaluationTime == 0 {
			if state.swiGetObserver(planet.evaluationTime, flags|(flagNoNutation), 1, observerVector[:], message) != 0 {
				return -1
			}
		} else {
			for i = 0; i <= 5; i++ {
				observerVector[i] = state.positions.data.observer.stateVector[i]
			}
		}
		for i = 0; i <= 5; i++ {
			xxm[i] -= observerVector[i]
		}
		for i = 0; i <= 5; i++ {
			observerVector[i] += earthData.stateVector[i]
		}
	} else if flags&flagBarycentric != 0 {
		for i = 0; i <= 5; i++ {
			observerVector[i] = 0
		}
		for i = 0; i <= 5; i++ {
			xxm[i] += earthData.stateVector[i]
		}
	} else if flags&flagHeliocentric != 0 {
		for i = 0; i <= 5; i++ {
			observerVector[i] = sunData.stateVector[i]
		}
		for i = 0; i <= 5; i++ {
			xxm[i] += earthData.stateVector[i] - sunData.stateVector[i]
		}
	} else {
		for i = 0; i <= 5; i++ {
			observerVector[i] = earthData.stateVector[i]
		}
	}
	t = planet.evaluationTime
	if flags&flagTruePosition == 0 {
		dt = math.Sqrt(xxm[0]*xxm[0]+xxm[1]*xxm[1]+xxm[2]*xxm[2]) * 1.49597870700e+11 / 2.99792458e+8 / secondsPerDay
		t = planet.evaluationTime - dt
		switch planet.ephemeris {
		case 1:
			status = state.swiPleph(t, 9, 2, xx[:], message)
			if status == 0 {
				status = state.swiPleph(t, 2, 11, xe[:], message)
			}
			if status == 0 && flags&flagHeliocentric != 0 {
				status = state.swiPleph(t, 10, 11, xs[:], message)
			}
			if status != 0 {
				state.swiCloseJplFile()
				state.positions.data.jplFileIsOpen = 0
			}
			i = 0
			for ; i <= 5; i++ {
				xx[i] += xe[i]
			}
		case 2:
			status = state.sweplan(t, 1, 1, flags, 0, xx[:], xe[:], xs[:], nil, message)
			if status != 0 {
				return status
			}
			i = 0
			for ; i <= 5; i++ {
				xx[i] += xe[i]
			}
		case 4:
			i = 0
			for ; i <= 2; i++ {
				xx[i] -= dt * xx[i+3]
				xe[i] = earthData.stateVector[i] - dt*earthData.stateVector[i+3]
				xe[i+3] = earthData.stateVector[i+3]
				xs[i] = 0
				xs[i+3] = 0
			}
		}
		if flags&flagTopocentric != 0 {
			if state.swiGetObserver(t, flags|(flagNoNutation), 0, xobs2[:], nil) != 0 {
				return -1
			}
			for i = 0; i <= 5; i++ {
				xobs2[i] += xe[i]
			}
		} else if flags&flagBarycentric != 0 {
			for i = 0; i <= 5; i++ {
				xobs2[i] = 0
			}
		} else if flags&flagHeliocentric != 0 {
			for i = 0; i <= 5; i++ {
				xobs2[i] = xs[i]
			}
		} else {
			for i = 0; i <= 5; i++ {
				xobs2[i] = xe[i]
			}
		}
	}
	for i = 0; i <= 5; i++ {
		xx[i] -= observerVector[i]
	}
	if flags&flagTruePosition == 0 && flags&flagNoAberration == 0 {
		state.swiAberrLight(xx[:], observerVector[:], flags)
		if flags&flagSpeed != 0 {
			for i = 3; i <= 5; i++ {
				xx[i] += observerVector[i] - xobs2[i]
			}
		}
	}
	if flags&flagSpeed == 0 {
		for i = 3; i <= 5; i++ {
			xx[i] = 0
		}
	}
	if flags&flagICRS == 0 && state.swiGetDenum(1, flags) >= 403 {
		state.swiBias(xx[:], t, flags, 0)
	}
	for i = 0; i <= 5; i++ {
		xxsv[i] = xx[i]
	}
	if flags&flagJ2000 == 0 {
		state.swiPrecess(xx[:], planet.evaluationTime, flags, -1)
		if flags&flagSpeed != 0 {
			state.swiPrecessSpeed(xx[:], planet.evaluationTime, flags, -1)
		}
		obliquityData = &state.positions.data.obliquity
	} else {
		obliquityData = &state.positions.data.obliquityJ2000
	}
	return state.appPosRest(planet, flags, xx[:], xxsv[:], obliquityData, message)
}

// appPosEtcSbar implements app_pos_etc_sbar from sweph.c.
//
//	transforms the position of the barycentric sun:
//
// precession and nutation
// according to flags
// iflag	flags
// serr         error string
func (state *State) appPosEtcSbar(flags int32, message []byte) int32 {
	var (
		i             int32
		xx            [6]float64
		xxsv          [6]float64
		dt            float64
		sunData       *planetData = &state.positions.data.planets[0]
		psbdp         *planetData = &state.positions.data.planets[10]
		obliquityData *obliquity  = &state.positions.data.obliquity
	)
	for i = 0; i <= 5; i++ {
		xx[i] = psbdp.stateVector[i]
	}
	if flags&flagTruePosition == 0 {
		dt = math.Sqrt(xx[0]*xx[0]+xx[1]*xx[1]+xx[2]*xx[2]) * 1.49597870700e+11 / 2.99792458e+8 / secondsPerDay
		for i = 0; i <= 2; i++ {
			xx[i] -= dt * xx[i+3]
		}
	}
	if flags&flagSpeed == 0 {
		for i = 3; i <= 5; i++ {
			xx[i] = 0
		}
	}
	if flags&flagICRS == 0 && state.swiGetDenum(0, flags) >= 403 {
		state.swiBias(xx[:], sunData.evaluationTime, flags, 0)
	}
	for i = 0; i <= 5; i++ {
		xxsv[i] = xx[i]
	}
	if flags&flagJ2000 == 0 {
		state.swiPrecess(xx[:], psbdp.evaluationTime, flags, -1)
		if flags&flagSpeed != 0 {
			state.swiPrecessSpeed(xx[:], psbdp.evaluationTime, flags, -1)
		}
		obliquityData = &state.positions.data.obliquity
	} else {
		obliquityData = &state.positions.data.obliquityJ2000
	}
	return state.appPosRest(sunData, flags, xx[:], xxsv[:], obliquityData, message)
}

// appPosEtcMean implements app_pos_etc_mean from sweph.c.
func (state *State) appPosEtcMean(body int32, flags int32, message []byte) int32 {
	var (
		i             int32
		flg1          int32
		flg2          int32
		xx            [6]float64
		xxsv          [6]float64
		planet        *planetData = &state.positions.data.lunarNodes[body]
		obliquityData *obliquity
	)
	flg1 = flags & ^(2 * 1024) & ^(4 * 1024)
	flg2 = planet.cachedFlags & ^(2 * 1024) & ^(4 * 1024)
	if flg1 == flg2 {
		planet.cachedFlags = flags
		planet.ephemeris = flags & (flagJPL | flagSwiss | flagMoshier)
		return 0
	}
	for i = 0; i <= 5; i++ {
		xx[i] = planet.stateVector[i]
	}
	state.swiPolcartSp(xx[:], xx[:])
	state.swiCoortrf2(xx[:], xx[:], -state.positions.data.obliquity.sine, state.positions.data.obliquity.cosine)
	state.swiCoortrf2(xx[3:], xx[3:], -state.positions.data.obliquity.sine, state.positions.data.obliquity.cosine)
	if flags&flagSpeed == 0 {
		for i = 3; i <= 5; i++ {
			xx[i] = 0
		}
	}
	if flags&flagSidereal != 0 && state.positions.data.sidereal.mode&256 != 0 || state.positions.data.sidereal.mode&512 != 0 {
		for i = 0; i <= 5; i++ {
			xxsv[i] = xx[i]
		}
		if planet.evaluationTime != julianDayJ2000 {
			state.swiPrecess(xxsv[:], planet.evaluationTime, flags, 1)
			if flags&flagSpeed != 0 {
				state.swiPrecessSpeed(xxsv[:], planet.evaluationTime, flags, 1)
			}
		}
	}
	if flags&flagJ2000 != 0 {
		state.swiPrecess(xx[:], planet.evaluationTime, flags, 1)
		if flags&flagSpeed != 0 {
			state.swiPrecessSpeed(xx[:], planet.evaluationTime, flags, 1)
		}
		obliquityData = &state.positions.data.obliquityJ2000
	} else {
		obliquityData = &state.positions.data.obliquity
	}
	return state.appPosRest(planet, flags, xx[:], xxsv[:], obliquityData, message)
}

// getNewSegment implements get_new_segment from sweph.c.
//
//	fetch chebyshew coefficients from sweph file for
//
// tjd 		time
// ipli		planet number
// ifno		file number
// serr		error string
func (state *State) getNewSegment(jd float64, bodyIndex int32, fileIndex int32, message []byte) int32 {
	var (
		i        int32
		j        int32
		k        int32
		m        int32
		n        int32
		o        int32
		icoord   int32
		status   int32
		iseg     int32
		fpos     [1]int32
		nsizes   int32
		nsize    [6]int32
		nco      int32
		idbl     int32
		c        [4]byte
		planet   *planetData    = &state.positions.data.planets[bodyIndex]
		fileData *ephemerisFile = &state.positions.data.files[fileIndex]
		fp       *textFile      = fileData.handle
		freord   int32          = fileData.flags & 2
		fendian  int32          = fileData.flags & 1
		longs    [40 + 1]uint32
	)
	iseg = int32((jd - planet.fileStart) / planet.segmentSpan)
	planet.segmentStart = planet.fileStart + float64(iseg)*planet.segmentSpan
	planet.segmentEnd = planet.segmentStart + planet.segmentSpan
	fpos[0] = planet.indexOffset + iseg*3
	status = state.doFread(fpos[:], 3, 1, 4, fp, fpos[0], freord, fendian, fileIndex, message)
	if status != 0 {
		closeTextFile(fileData.handle)
		fileData.handle = nil
		state.freePlanets()
		return -1
	}
	seekFile(fp, fpos[0], 0)
	if planet.coefficients == nil {
		planet.coefficients = make([]float64, uint64(planet.coefficientCount)*3*8/8)
	}
	clear(planet.coefficients)
	for icoord = 0; icoord < 3; icoord++ {
		idbl = icoord * planet.coefficientCount
		status = state.doFread(c[:], 1, 2, 1, fp, -1, freord, fendian, fileIndex, message)
		if status != 0 {
			closeTextFile(fileData.handle)
			fileData.handle = nil
			state.freePlanets()
			return -1
		}
		if int32(c[0])&128 != 0 {
			nsizes = 6
			status = state.doFread(c[2:], 1, 2, 1, fp, -1, freord, fendian, fileIndex, message)
			if status != 0 {
				closeTextFile(fileData.handle)
				fileData.handle = nil
				state.freePlanets()
				return -1
			}
			nsize[0] = int32(c[1]) / 16
			nsize[1] = int32(c[1]) % 16
			nsize[2] = int32(c[2]) / 16
			nsize[3] = int32(c[2]) % 16
			nsize[4] = int32(c[3]) / 16
			nsize[5] = int32(c[3]) % 16
			nco = nsize[0] + nsize[1] + nsize[2] + nsize[3] + nsize[4] + nsize[5]
		} else {
			nsizes = 4
			nsize[0] = int32(c[0]) / 16
			nsize[1] = int32(c[0]) % 16
			nsize[2] = int32(c[1]) / 16
			nsize[3] = int32(c[1]) % 16
			nco = nsize[0] + nsize[1] + nsize[2] + nsize[3]
		}
		if nco > planet.coefficientCount {
			if message != nil {
				formatText(message, textBytes("error in ephemeris file: %d coefficients instead of %d. "), nco, planet.coefficientCount)
				if textLength(message)+textLength(fileData.name[:]) < uint64(256-1) {
					formatText(message, textBytes("error in ephemeris file %s: %d coefficients instead of %d. "), fileData.name[:], nco, planet.coefficientCount)
				}
			}
			planet.coefficients = nil
			return -1
		}
		for i = 0; i < nsizes; i++ {
			if nsize[i] == 0 {
				continue
			}
			if i < 4 {
				j = 4 - i
				k = nsize[i]
				status = state.doFread(longs[:], j, k, 4, fp, -1, freord, fendian, fileIndex, message)
				if status != 0 {
					closeTextFile(fileData.handle)
					fileData.handle = nil
					state.freePlanets()
					return -1
				}
				for m = 0; m < k; {
					if longs[m]&1 != 0 {
						planet.coefficients[idbl] = -(float64((longs[m]+1)/2) / 1e+9 * planet.coefficientScale / 2)
					} else {
						planet.coefficients[idbl] = float64(longs[m]/2) / 1e+9 * planet.coefficientScale / 2
					}
					m++
					idbl++
				}
			} else if i == 4 {
				j = 1
				k = (nsize[i] + 1) / 2
				status = state.doFread(longs[:], j, k, 4, fp, -1, freord, fendian, fileIndex, message)
				if status != 0 {
					closeTextFile(fileData.handle)
					fileData.handle = nil
					state.freePlanets()
					return -1
				}
				m = 0
				for j = 0; m < k && j < nsize[i]; m++ {
					n = 0
					for o = 16; n < 2 && j < nsize[i]; {
						if longs[m]&uint32(o) != 0 {
							planet.coefficients[idbl] = -(float64((longs[m]+uint32(o))/uint32(o)/2) * planet.coefficientScale / 2 / 1e+9)
						} else {
							planet.coefficients[idbl] = float64(longs[m]/uint32(o)/2) * planet.coefficientScale / 2 / 1e+9
						}
						n++
						j++
						idbl++
						longs[m] %= uint32(o)
						o /= 16
					}
				}
			} else if i == 5 {
				j = 1
				k = (nsize[i] + 3) / 4
				status = state.doFread(longs[:], j, k, 4, fp, -1, freord, fendian, fileIndex, message)
				if status != 0 {
					closeTextFile(fileData.handle)
					fileData.handle = nil
					state.freePlanets()
					return -1
				}
				m = 0
				for j = 0; m < k && j < nsize[i]; m++ {
					n = 0
					for o = 64; n < 4 && j < nsize[i]; {
						if longs[m]&uint32(o) != 0 {
							planet.coefficients[idbl] = -(float64((longs[m]+uint32(o))/uint32(o)/2) * planet.coefficientScale / 2 / 1e+9)
						} else {
							planet.coefficients[idbl] = float64(longs[m]/uint32(o)/2) * planet.coefficientScale / 2 / 1e+9
						}
						n++
						j++
						idbl++
						longs[m] %= uint32(o)
						o /= 4
					}
				}
			}
		}
	}
	return 0
}

// readConst implements read_const from sweph.c.
//
//	SWISSEPH
//
// reads constants on ephemeris file
// ifno         file #
// serr         error string
func (state *State) readConst(fileIndex int32, message []byte) int32 {
	var (
		sp             []byte
		s              [256 * 2]byte
		s2             [256]byte
		sastnam        [41]byte
		i              int32
		bodyIndex      int32
		kpl            int32
		status         int32
		fendian        int32
		freord         int32
		lastnam        int32 = 19
		fp             *textFile
		lng            [1]int32
		ulng           [1]uint32
		flen           int32
		fpos           int32
		nplan          [1]int16
		testendian     [1]int32
		doubles        [20]float64
		planet         *planetData
		fileData       *ephemerisFile = &state.positions.data.files[fileIndex]
		serrFileDamage []byte         = textBytes("Ephemeris file %s is damaged (0%s). ")
		smsg           []byte         = textBytes("")
		nbytesIpl      int32          = 2
	)
	fp = fileData.handle
	sp = readLine(s[:], 256, fp)
	if sp == nil || findText(sp, textBytes("\r\n")) == nil {
		if message != nil {
			message[0] = byte('\000')
			if textLength(serrFileDamage)+textLength(fileData.name[:])+textLength(smsg) < 256 {
				formatText(message, serrFileDamage, fileData.name[:], smsg)
			}
		}
		closeTextFile(fileData.handle)
		fileData.handle = nil
		state.freePlanets()
		return -1
	}
	sp = findByte(s[:], '\r')
	sp[0] = byte('\000')
	for sp = s[:]; isDigit(int32(sp[0])) == 0 && int32(sp[0]) != '\000'; {
		sp = sp[1:]
	}
	if int32(sp[0]) == '\000' {
		smsg = textBytes("a")
		if message != nil {
			message[0] = byte('\000')
			if textLength(serrFileDamage)+textLength(fileData.name[:])+textLength(smsg) < 256 {
				formatText(message, serrFileDamage, fileData.name[:], smsg)
			}
		}
		closeTextFile(fileData.handle)
		fileData.handle = nil
		state.freePlanets()
		return -1
	}
	fileData.version = parseInt(sp)
	sp = readLine(s[:], 256, fp)
	if sp == nil || findText(sp, textBytes("\r\n")) == nil {
		smsg = textBytes("b")
		if message != nil {
			message[0] = byte('\000')
			if textLength(serrFileDamage)+textLength(fileData.name[:])+textLength(smsg) < 256 {
				formatText(message, serrFileDamage, fileData.name[:], smsg)
			}
		}
		closeTextFile(fileData.handle)
		fileData.handle = nil
		state.freePlanets()
		return -1
	}
	sp = findLastByte(fileData.name[:], int32(textBytes("/")[0]))
	if sp == nil {
		sp = fileData.name[:]
	} else {
		sp = sp[1:]
	}
	copyText(s2[:], sp)
	for sp = s2[:]; int32(sp[0]) != '\000'; sp = sp[1:] {
		sp[0] = byte(lowerByte(int32(sp[0])))
	}
	trimBufferEnd(s[:], "\n\r ")
	for sp = s[:]; int32(sp[0]) != '\000'; sp = sp[1:] {
		sp[0] = byte(lowerByte(int32(sp[0])))
	}
	if compareText(s2[:], s[:]) != 0 {
		if message != nil {
			formatText(message, textBytes("Ephemeris file name '%s' wrong; rename '%s' "), s2[:], s[:])
		}
		closeTextFile(fileData.handle)
		fileData.handle = nil
		state.freePlanets()
		return -1
	}
	sp = readLine(s[:], 256, fp)
	if sp == nil || findText(sp, textBytes("\r\n")) == nil {
		smsg = textBytes("c")
		if message != nil {
			message[0] = byte('\000')
			if textLength(serrFileDamage)+textLength(fileData.name[:])+textLength(smsg) < 256 {
				formatText(message, serrFileDamage, fileData.name[:], smsg)
			}
		}
		closeTextFile(fileData.handle)
		fileData.handle = nil
		state.freePlanets()
		return -1
	}
	if fileIndex == 3 {
		sp = readLine(s[:], 256*2, fp)
		if sp == nil || findText(sp, textBytes("\r\n")) == nil {
			smsg = textBytes("d")
			if message != nil {
				message[0] = byte('\000')
				if textLength(serrFileDamage)+textLength(fileData.name[:])+textLength(smsg) < 256 {
					formatText(message, serrFileDamage, fileData.name[:], smsg)
				}
			}
			closeTextFile(fileData.handle)
			fileData.handle = nil
			state.freePlanets()
			return -1
		}
		for int32(sp[0]) == ' ' {
			sp = sp[1:]
		}
		for isDigit(int32(sp[0])) != 0 {
			sp = sp[1:]
		}
		sp = sp[1:]
		i = sliceOffset(sp, s[:])
		copyTextN(sastnam[:], s[:], lastnam+i)
		sastnam[lastnam+i] = byte('\000')
		copyText(state.positions.data.astelem[:], s[:])
		state.positions.data.astH = parseFloat(s[35+i:])
		state.positions.data.astG = parseFloat(s[42+i:])
		if state.positions.data.astG == 0 {
			state.positions.data.astG = 0.15
		}
		copyTextN(s2[:], s[51+i:], 7)
		s2[7] = byte('\000')
		state.positions.data.astDiam = parseFloat(s2[:])
		if state.positions.data.astDiam == 0 {
			state.positions.data.astDiam = 1329 / math.Sqrt(0.15) * math.Pow(10, -0.2*state.positions.data.astH)
		}
	}
	if readBinary(testendian[:], 4, 1, fp) != 1 {
		smsg = textBytes("e")
		if message != nil {
			message[0] = byte('\000')
			if textLength(serrFileDamage)+textLength(fileData.name[:])+textLength(smsg) < 256 {
				formatText(message, serrFileDamage, fileData.name[:], smsg)
			}
		}
		closeTextFile(fileData.handle)
		fileData.handle = nil
		state.freePlanets()
		return -1
	}
	freord = 0
	fendian = 1
	if uint32(testendian[0]) != 0x00616263 {
		freord = 2
		fendian = 0
		if reverse32(uint32(testendian[0])) != 0x00616263 {
			smsg = textBytes("f")
			if message != nil {
				message[0] = byte('\000')
				if textLength(serrFileDamage)+textLength(fileData.name[:])+textLength(smsg) < 256 {
					formatText(message, serrFileDamage, fileData.name[:], smsg)
				}
			}
			closeTextFile(fileData.handle)
			fileData.handle = nil
			state.freePlanets()
			return -1
		}
	}
	fileData.flags = freord | fendian
	status = state.doFread(lng[:], 4, 1, 4, fp, -1, freord, fendian, fileIndex, message)
	if status != 0 {
		closeTextFile(fileData.handle)
		fileData.handle = nil
		state.freePlanets()
		return -1
	}
	fpos = int32(tellFile(fp))
	if seekFile(fp, 0, 2) != 0 {
		smsg = textBytes("g")
		if message != nil {
			message[0] = byte('\000')
			if textLength(serrFileDamage)+textLength(fileData.name[:])+textLength(smsg) < 256 {
				formatText(message, serrFileDamage, fileData.name[:], smsg)
			}
		}
		closeTextFile(fileData.handle)
		fileData.handle = nil
		state.freePlanets()
		return -1
	}
	flen = int32(tellFile(fp))
	if lng[0] != flen {
		smsg = textBytes("h")
		if message != nil {
			message[0] = byte('\000')
			if textLength(serrFileDamage)+textLength(fileData.name[:])+textLength(smsg) < 256 {
				formatText(message, serrFileDamage, fileData.name[:], smsg)
			}
		}
		closeTextFile(fileData.handle)
		fileData.handle = nil
		state.freePlanets()
		return -1
	}
	status = state.doFread(&fileData.deNumber, 4, 1, 4, fp, fpos, freord, fendian, fileIndex, message)
	if status != 0 {
		closeTextFile(fileData.handle)
		fileData.handle = nil
		state.freePlanets()
		return -1
	}
	status = state.doFread(&fileData.start, 8, 1, 8, fp, -1, freord, fendian, fileIndex, message)
	if status != 0 {
		closeTextFile(fileData.handle)
		fileData.handle = nil
		state.freePlanets()
		return -1
	}
	status = state.doFread(&fileData.end, 8, 1, 8, fp, -1, freord, fendian, fileIndex, message)
	if status != 0 {
		closeTextFile(fileData.handle)
		fileData.handle = nil
		state.freePlanets()
		return -1
	}
	status = state.doFread(nplan[:], 2, 1, 2, fp, -1, freord, fendian, fileIndex, message)
	if status != 0 {
		closeTextFile(fileData.handle)
		fileData.handle = nil
		state.freePlanets()
		return -1
	}
	if int32(nplan[0]) > 256 {
		nbytesIpl = 4
		nplan[0] %= 256
	}
	if int32(nplan[0]) < 1 || int32(nplan[0]) > 20 {
		smsg = textBytes("i")
		if message != nil {
			message[0] = byte('\000')
			if textLength(serrFileDamage)+textLength(fileData.name[:])+textLength(smsg) < 256 {
				formatText(message, serrFileDamage, fileData.name[:], smsg)
			}
		}
		closeTextFile(fileData.handle)
		fileData.handle = nil
		state.freePlanets()
		return -1
	}
	fileData.bodyCount = nplan[0]
	status = state.doFread(fileData.bodies[:], nbytesIpl, int32(nplan[0]), 4, fp, -1, freord, fendian, fileIndex, message)
	if status != 0 {
		closeTextFile(fileData.handle)
		fileData.handle = nil
		state.freePlanets()
		return -1
	}
	if fileIndex == 3 {
		var (
			sastno [12]byte
			j      int32
		)
		for j = 4; int32(sastnam[j]) != ' ' && j < 10; {
			j++
		}
		copyTextN(sastno[:], sastnam[:], j)
		sastno[j] = byte('\000')
		i = int32(parseLong(sastno[:]))
		if i == fileData.bodies[0]-10000 || i == fileData.bodies[0] {
			copyTextN(fileData.asteroidName[:], sastnam[j+1:], lastnam)
			fileData.asteroidName[lastnam] = byte('\000')
			if readBinary(s[:], 30, 1, fp) != 1 {
				smsg = textBytes("j")
				if message != nil {
					message[0] = byte('\000')
					if textLength(serrFileDamage)+textLength(fileData.name[:])+textLength(smsg) < 256 {
						formatText(message, serrFileDamage, fileData.name[:], smsg)
					}
				}
				closeTextFile(fileData.handle)
				fileData.handle = nil
				state.freePlanets()
				return -1
			}
		} else if readBinary(fileData.asteroidName[:], 30, 1, fp) != 1 {
			smsg = textBytes("k")
			if message != nil {
				message[0] = byte('\000')
				if textLength(serrFileDamage)+textLength(fileData.name[:])+textLength(smsg) < 256 {
					formatText(message, serrFileDamage, fileData.name[:], smsg)
				}
			}
			closeTextFile(fileData.handle)
			fileData.handle = nil
			state.freePlanets()
			return -1
		}
		i = int32(textLength(fileData.asteroidName[:]) - 1)
		if i < 0 {
			i = 0
		}
		for i > 0 && fileData.asteroidName[i] == ' ' {
			i--
		}
		fileData.asteroidName[i+1] = 0
		if sp = findText(fileData.asteroidName[:], textBytes("  ")); sp != nil {
			sp[0] = byte('\000')
		}
	}
	fpos = int32(tellFile(fp))
	status = state.doFread(ulng[:], 4, 1, 4, fp, -1, freord, fendian, fileIndex, message)
	if status != 0 {
		closeTextFile(fileData.handle)
		fileData.handle = nil
		state.freePlanets()
		return -1
	}
	seekFile(fp, 0, 0)
	if fpos-1 > 2*256 {
		smsg = textBytes("l")
		if message != nil {
			message[0] = byte('\000')
			if textLength(serrFileDamage)+textLength(fileData.name[:])+textLength(smsg) < 256 {
				formatText(message, serrFileDamage, fileData.name[:], smsg)
			}
		}
		closeTextFile(fileData.handle)
		fileData.handle = nil
		state.freePlanets()
		return -1
	}
	if readBinary(s[:], uint64(fpos), 1, fp) != 1 {
		smsg = textBytes("m")
		if message != nil {
			message[0] = byte('\000')
			if textLength(serrFileDamage)+textLength(fileData.name[:])+textLength(smsg) < 256 {
				formatText(message, serrFileDamage, fileData.name[:], smsg)
			}
		}
		closeTextFile(fileData.handle)
		fileData.handle = nil
		state.freePlanets()
		return -1
	}
	if state.swiCrc32(s[:], fpos) != ulng[0] {
		smsg = textBytes("n")
		if message != nil {
			message[0] = byte('\000')
			if textLength(serrFileDamage)+textLength(fileData.name[:])+textLength(smsg) < 256 {
				formatText(message, serrFileDamage, fileData.name[:], smsg)
			}
		}
		closeTextFile(fileData.handle)
		fileData.handle = nil
		state.freePlanets()
		return -1
	}
	seekFile(fp, fpos+4, 0)
	status = state.doFread(doubles[:], 8, 5, 8, fp, -1, freord, fendian, fileIndex, message)
	if status != 0 {
		closeTextFile(fileData.handle)
		fileData.handle = nil
		state.freePlanets()
		return -1
	}
	state.positions.data.constants.speedOfLight = doubles[0]
	state.positions.data.constants.astronomicalUnit = doubles[1]
	state.positions.data.constants.solarGravity = doubles[2]
	state.positions.data.constants.earthMoonMassRatio = doubles[3]
	state.positions.data.constants.solarRadius = doubles[4]
	for kpl = 0; kpl < int32(fileData.bodyCount); kpl++ {
		bodyIndex = fileData.bodies[kpl]
		if bodyIndex >= 10000 {
			planet = &state.positions.data.planets[11]
		} else if bodyIndex >= 9000 {
			planet = &state.positions.data.planets[11]
		} else {
			planet = &state.positions.data.planets[bodyIndex]
		}
		planet.bodyID = bodyIndex
		status = state.doFread(&planet.indexOffset, 4, 1, 4, fp, -1, freord, fendian, fileIndex, message)
		if status != 0 {
			closeTextFile(fileData.handle)
			fileData.handle = nil
			state.freePlanets()
			return -1
		}
		status = state.doFread(&planet.fileFlags, 1, 1, 4, fp, -1, freord, fendian, fileIndex, message)
		if status != 0 {
			closeTextFile(fileData.handle)
			fileData.handle = nil
			state.freePlanets()
			return -1
		}
		status = state.doFread(&planet.coefficientCount, 1, 1, 4, fp, -1, freord, fendian, fileIndex, message)
		if status != 0 {
			closeTextFile(fileData.handle)
			fileData.handle = nil
			state.freePlanets()
			return -1
		}
		status = state.doFread(lng[:], 4, 1, 4, fp, -1, freord, fendian, fileIndex, message)
		if status != 0 {
			closeTextFile(fileData.handle)
			fileData.handle = nil
			state.freePlanets()
			return -1
		}
		planet.coefficientScale = float64(lng[0]) / 1000.0
		if bodyIndex >= 9000 && bodyIndex < 10000 {
			if bodyIndex%100 == 99 || (bodyIndex-9000)/100 == 4 {
				planet.coefficientScale = float64(lng[0]) / 1000000.0
			}
		}
		status = state.doFread(doubles[:], 8, 10, 8, fp, -1, freord, fendian, fileIndex, message)
		if status != 0 {
			closeTextFile(fileData.handle)
			fileData.handle = nil
			state.freePlanets()
			return -1
		}
		planet.fileStart = doubles[0]
		planet.fileEnd = doubles[1]
		planet.segmentSpan = doubles[2]
		planet.segmentCount = int32((doubles[1] - doubles[0] + 0.1) / doubles[2])
		planet.elementEpoch = doubles[3]
		planet.prot = doubles[4]
		planet.dprot = doubles[5]
		planet.qrot = doubles[6]
		planet.dqrot = doubles[7]
		planet.peri = doubles[8]
		planet.dperi = doubles[9]
		if planet.fileFlags&4 != 0 {
			if planet.referenceCoefficients != nil {
				planet.referenceCoefficients = nil
				if planet.coefficients != nil {
					planet.coefficients = nil
				}
			}
			planet.referenceCoefficients = make([]float64, uint64(planet.coefficientCount)*2*8/8)
			status = state.doFread(planet.referenceCoefficients, 8, 2*planet.coefficientCount, 8, fp, -1, freord, fendian, fileIndex, message)
			if status != 0 {
				planet.referenceCoefficients = nil
				closeTextFile(fileData.handle)
				fileData.handle = nil
				state.freePlanets()
				return -1
			}
		}
	}
	return 0
}

// rotBack implements rot_back from sweph.c.
func (state *State) rotBack(bodyIndex int32) {
	var (
		i        int32
		t        float64
		tdiff    float64
		qav      float64
		pav      float64
		dn       float64
		omtild   float64
		com      float64
		som      float64
		cosih2   float64
		x        [40 + 1][3]float64
		uix      [3]float64
		uiy      [3]float64
		uiz      [3]float64
		xrot     float64
		yrot     float64
		zrot     float64
		chcfx    []float64
		chcfy    []float64
		chcfz    []float64
		refepx   []float64
		refepy   []float64
		seps2000 float64     = 0.39777715572793088
		ceps2000 float64     = 0.91748206215761929
		planet   *planetData = &state.positions.data.planets[bodyIndex]
		nco      int32       = planet.coefficientCount
	)
	t = planet.segmentStart + planet.segmentSpan/2
	chcfx = planet.coefficients
	chcfy = chcfx[nco:]
	chcfz = chcfx[2*nco:]
	tdiff = (t - planet.elementEpoch) / 365250.0
	if bodyIndex == 1 {
		dn = planet.prot + tdiff*planet.dprot
		i = int32(dn / 6.2831853071795862)
		dn -= float64(i) * 6.2831853071795862
		qav = (planet.qrot + tdiff*planet.dqrot) * math.Cos(dn)
		pav = (planet.qrot + tdiff*planet.dqrot) * math.Sin(dn)
	} else {
		qav = planet.qrot + tdiff*planet.dqrot
		pav = planet.prot + tdiff*planet.dprot
	}
	for i = 0; i < nco; i++ {
		x[i][0] = chcfx[i]
		x[i][1] = chcfy[i]
		x[i][2] = chcfz[i]
	}
	if planet.fileFlags&4 != 0 {
		refepx = planet.referenceCoefficients
		refepy = refepx[nco:]
		omtild = planet.peri + tdiff*planet.dperi
		i = int32(omtild / 6.2831853071795862)
		omtild -= float64(i) * 6.2831853071795862
		com = math.Cos(omtild)
		som = math.Sin(omtild)
		for i = 0; i < nco; i++ {
			x[i][0] = chcfx[i] + com*refepx[i] - som*refepy[i]
			x[i][1] = chcfy[i] + com*refepy[i] + som*refepx[i]
		}
	}
	cosih2 = 1.0 / (1.0 + qav*qav + pav*pav)
	uiz[0] = 2.0 * pav * cosih2
	uiz[1] = -2.0 * qav * cosih2
	uiz[2] = (1.0 - qav*qav - pav*pav) * cosih2
	uix[0] = (1.0 + qav*qav - pav*pav) * cosih2
	uix[1] = 2.0 * qav * pav * cosih2
	uix[2] = -2.0 * pav * cosih2
	uiy[0] = 2.0 * qav * pav * cosih2
	uiy[1] = (1.0 - qav*qav + pav*pav) * cosih2
	uiy[2] = 2.0 * qav * cosih2
	for i = 0; i < nco; i++ {
		xrot = x[i][0]*uix[0] + x[i][1]*uiy[0] + x[i][2]*uiz[0]
		yrot = x[i][0]*uix[1] + x[i][1]*uiy[1] + x[i][2]*uiz[1]
		zrot = x[i][0]*uix[2] + x[i][1]*uiy[2] + x[i][2]*uiz[2]
		if math.Abs(xrot)+math.Abs(yrot)+math.Abs(zrot) >= 1e-14 {
			planet.evaluationCount = i
		}
		x[i][0] = xrot
		x[i][1] = yrot
		x[i][2] = zrot
		if bodyIndex == 1 {
			x[i][1] = ceps2000*yrot - seps2000*zrot
			x[i][2] = seps2000*yrot + ceps2000*zrot
		}
	}
	for i = 0; i < nco; i++ {
		chcfx[i] = x[i][0]
		chcfy[i] = x[i][1]
		chcfz[i] = x[i][2]
	}
}

// embofs implements embofs from sweph.c.
func (state *State) embofs(xemb []float64, moonVector []float64) {
	var i int32
	for i = 0; i <= 2; i++ {
		xemb[i] -= moonVector[i] / 82.30055985272827
	}
}

// nutMatrix implements nut_matrix from sweph.c.
//
//	calculates the nutation matrix
//
// nu		pointer to nutation data structure
// oe		pointer to epsilon data structure
func (state *State) nutMatrix(nu *nutation, obliquityData *obliquity) {
	var (
		psi     float64
		eps     float64
		sinpsi  float64
		cospsi  float64
		sineps  float64
		coseps  float64
		sineps0 float64
		coseps0 float64
	)
	psi = nu.angles[0]
	eps = obliquityData.angle + nu.angles[1]
	sinpsi = math.Sin(psi)
	cospsi = math.Cos(psi)
	sineps0 = obliquityData.sine
	coseps0 = obliquityData.cosine
	sineps = math.Sin(eps)
	coseps = math.Cos(eps)
	nu.matrix[0][0] = cospsi
	nu.matrix[0][1] = sinpsi * coseps
	nu.matrix[0][2] = sinpsi * sineps
	nu.matrix[1][0] = -sinpsi * coseps0
	nu.matrix[1][1] = cospsi*coseps*coseps0 + sineps*sineps0
	nu.matrix[1][2] = cospsi*sineps*coseps0 - coseps*sineps0
	nu.matrix[2][0] = -sinpsi * sineps0
	nu.matrix[2][1] = cospsi*coseps*sineps0 - sineps*coseps0
	nu.matrix[2][2] = cospsi*sineps*sineps0 + coseps*coseps0
}

// lunarOscElem implements lunar_osc_elem from sweph.c.
func (state *State) lunarOscElem(jd float64, body int32, flags int32, message []byte) int32 {
	var (
		i             int32
		j             int32
		istart        int32
		bodyIndex     int32 = 1
		ephemeris     int32 = 2
		status        int32 = -1
		flg1          int32
		flg2          int32
		daya          [2]float64
		node          *planetData
		ndnp          *planetData
		ndap          *planetData
		obliquityData *obliquity
		speedIntv     float64 = 0.0001
		a             float64
		b             float64
		xpos          [3][6]float64
		xx            [3][6]float64
		xxa           [3][6]float64
		xnorm         [6]float64
		r             [6]float64
		xp            []float64
		rxy           float64
		rxyz          float64
		t             float64
		dt            float64
		fac           float64
		sgn           float64
		sinnode       float64
		cosnode       float64
		sinincl       float64
		cosincl       float64
		sinu          float64
		cosu          float64
		sinE          float64
		cosE          float64
		uu            float64
		ny            float64
		sema          float64
		ecce          float64
		Gmsm          float64
		c2            float64
		v2            float64
		pp            float64
		speedf1       int32
		speedf2       int32
	)
	obliquityData = &state.positions.data.obliquity
	node = &state.positions.data.lunarNodes[body]
	flg1 = flags & ^(2 * 1024) & ^(4 * 1024)
	flg2 = node.cachedFlags & ^(2 * 1024) & ^(4 * 1024)
	speedf1 = node.cachedFlags & 256
	speedf2 = flags & (flagSpeed)
	if jd == node.evaluationTime && jd != 0 && flg1 == flg2 && (!(speedf2 != 0) || speedf1 != 0) {
		node.cachedFlags = flags
		node.ephemeris = flags & (flagJPL | flagSwiss | flagMoshier)
		return 0
	}
	if flags&flagMoshier != 0 {
		ephemeris = 4
	} else if flags&flagSwiss != 0 {
		ephemeris = 2
	} else if flags&flagJPL != 0 {
		ephemeris = 1
	}
	state.positions.data.planets[1].evaluationTime = 0
	if flags&flagSpeed != 0 {
		istart = 0
	} else {
		istart = 2
	}
	if message != nil {
		message[0] = byte('\000')
	}
threePositions:
	for {
		switch ephemeris {
		case 1:
			speedIntv = 0.0001
			i = istart
			for ; i <= 2; i++ {
				if i == 0 {
					t = jd - speedIntv
				} else if i == 1 {
					t = jd + speedIntv
				} else {
					t = jd
				}
				xp = xpos[i][:]
				status = state.jplplan(t, bodyIndex, flags, 0, xp, nil, nil, message)
				if status == -1 {
					return -1
				}
				if flags&flagTruePosition == 0 && status >= 0 {
					dt = math.Sqrt(xpos[i][0]*xpos[i][0]+xpos[i][1]*xpos[i][1]+xpos[i][2]*xpos[i][2]) * 1.49597870700e+11 / 2.99792458e+8 / secondsPerDay
					status = state.jplplan(t-dt, bodyIndex, flags, 0, xpos[i][:], nil, nil, message)
					if status == -1 {
						return -1
					}
				}
				if status == -2 {
					flags = flags & ^1 | 2
					ephemeris = 2
					if message != nil && textLength(message)+30 < 256 {
						appendText(message, textBytes(" \ntrying Swiss Eph; "))
					}
					break
				} else if status == -3 {
					if jd > 625000.5 && jd < 2818000.5 {
						flags = flags & ^1 | 4
						ephemeris = 4
						if message != nil && textLength(message)+30 < 256 {
							appendText(message, textBytes(" \nusing Moshier Eph; "))
						}
						break
					} else {
						return -1
					}
				}
				status = state.swiPlanForOscElem(flags|(flagSpeed), t, xpos[i][:])
			}
		case 2:
			speedIntv = 0.0001
			i = istart
			for ; i <= 2; i++ {
				if i == 0 {
					t = jd - speedIntv
				} else if i == 1 {
					t = jd + speedIntv
				} else {
					t = jd
				}
				status = state.swemoon(t, flags|(flagSpeed), 0, xpos[i][:], message)
				if status == -1 {
					return -1
				}
				if flags&flagTruePosition == 0 && status >= 0 {
					dt = math.Sqrt(xpos[i][0]*xpos[i][0]+xpos[i][1]*xpos[i][1]+xpos[i][2]*xpos[i][2]) * 1.49597870700e+11 / 2.99792458e+8 / secondsPerDay
					status = state.swemoon(t-dt, flags|(flagSpeed), 0, xpos[i][:], message)
					if status == -1 {
						return -1
					}
				}
				if status == -2 {
					if jd > 625000.5 && jd < 2818000.5 {
						flags = flags & ^2 | 4
						ephemeris = 4
						if message != nil && textLength(message)+30 < 256 {
							appendText(message, textBytes(" \nusing Moshier eph.; "))
						}
						break
					} else {
						return -1
					}
				}
				status = state.swiPlanForOscElem(flags|(flagSpeed), t, xpos[i][:])
			}
		case 4:
			speedIntv = 0.1
			i = istart
			for ; i <= 2; i++ {
				if i == 0 {
					t = jd - speedIntv
				} else if i == 1 {
					t = jd + speedIntv
				} else {
					t = jd
				}
				status = state.swiMoshmoon(t, 0, xpos[i][:], message)
				if status == -1 {
					return status
				}
				status = state.swiPlanForOscElem(flags|(flagSpeed), t, xpos[i][:])
			}
		default:
		}
		if status == -2 || status == -3 {
			continue threePositions
		}
		ndnp = &state.positions.data.lunarNodes[1]
		for i = istart; i <= 2; i++ {
			if math.Abs(xpos[i][5]) < 1e-15 {
				xpos[i][5] = 1e-15
			}
			fac = xpos[i][2] / xpos[i][5]
			sgn = xpos[i][5] / math.Abs(xpos[i][5])
			for j = 0; j <= 2; j++ {
				xx[i][j] = (xpos[i][j] - fac*xpos[i][j+3]) * sgn
			}
		}
		for i = 0; i <= 2; i++ {
			ndnp.stateVector[i] = xx[2][i]
			if flags&flagSpeed != 0 {
				b = (xx[1][i] - xx[0][i]) / 2
				a = (xx[1][i]+xx[0][i])/2 - xx[2][i]
				ndnp.stateVector[i+3] = (2*a + b) / speedIntv
			} else {
				ndnp.stateVector[i+3] = 0
			}
			ndnp.evaluationTime = jd
			ndnp.ephemeris = ephemeris
		}
		ndap = &state.positions.data.lunarNodes[3]
		Gmsm = 8.9970116861855104e-10
		for i = istart; i <= 2; i++ {
			rxy = math.Sqrt(xx[i][0]*xx[i][0] + xx[i][1]*xx[i][1])
			cosnode = xx[i][0] / rxy
			sinnode = xx[i][1] / rxy
			state.swiCrossProd(xpos[i][:], xpos[i][3:], xnorm[:])
			rxy = xnorm[0]*xnorm[0] + xnorm[1]*xnorm[1]
			c2 = rxy + xnorm[2]*xnorm[2]
			rxyz = math.Sqrt(c2)
			rxy = math.Sqrt(rxy)
			sinincl = rxy / rxyz
			cosincl = math.Sqrt(1 - sinincl*sinincl)
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
			xxa[i][0] = state.swiMod2PI(uu - ny + 3.14159265358979323846)
			xxa[i][1] = 0
			xxa[i][2] = sema * (1 + ecce)
			state.swiPolcart(xxa[i][:], xxa[i][:])
			state.swiCoortrf2(xxa[i][:], xxa[i][:], -sinincl, cosincl)
			state.swiCartpol(xxa[i][:], xxa[i][:])
			xxa[i][0] += math.Atan2(sinnode, cosnode)
			state.swiPolcart(xxa[i][:], xxa[i][:])
			ny = state.swiMod2PI(ny - uu)
			cosE = math.Cos(2 * math.Atan(math.Tan(ny/2)/math.Sqrt((1+ecce)/(1-ecce))))
			r[0] = sema * (1 - ecce*cosE)
			r[1] = math.Sqrt(xx[i][0]*xx[i][0] + xx[i][1]*xx[i][1] + xx[i][2]*xx[i][2])
			for j = 0; j <= 2; j++ {
				xx[i][j] *= r[0] / r[1]
			}
		}
		for i = 0; i <= 2; i++ {
			ndap.stateVector[i] = xxa[2][i]
			if flags&flagSpeed != 0 {
				ndap.stateVector[i+3] = (xxa[1][i] - xxa[0][i]) / speedIntv / 2
			} else {
				ndap.stateVector[i+3] = 0
			}
			ndap.evaluationTime = jd
			ndap.ephemeris = ephemeris
			ndnp.stateVector[i] = xx[2][i]
			if flags&flagSpeed != 0 {
				ndnp.stateVector[i+3] = (xx[1][i] - xx[0][i]) / speedIntv / 2
			} else {
				ndnp.stateVector[i+3] = 0
			}
		}
		for j = 0; j <= 1; j++ {
			var x [6]float64
			if j == 0 {
				node = &state.positions.data.lunarNodes[1]
			} else {
				node = &state.positions.data.lunarNodes[3]
			}
			clear(node.coordinates[:])
			for i = 0; i <= 5; i++ {
				node.coordinates[6+i] = node.stateVector[i]
			}
			state.swiCartpolSp(node.coordinates[6:], node.coordinates[:])
			state.swiCoortrf2(node.coordinates[6:], node.coordinates[18:], -obliquityData.sine, obliquityData.cosine)
			if flags&flagSpeed != 0 {
				state.swiCoortrf2(node.coordinates[9:], node.coordinates[21:], -obliquityData.sine, obliquityData.cosine)
			}
			if flags&flagNoNutation == 0 {
				state.swiCoortrf2(node.coordinates[18:], node.coordinates[18:], -state.positions.data.nutation.sine, state.positions.data.nutation.cosine)
				if flags&flagSpeed != 0 {
					state.swiCoortrf2(node.coordinates[21:], node.coordinates[21:], -state.positions.data.nutation.sine, state.positions.data.nutation.cosine)
				}
			}
			state.swiCartpolSp(node.coordinates[18:], node.coordinates[12:])
			node.cachedFlags = flags
			node.ephemeris = flags & (flagJPL | flagSwiss | flagMoshier)
			if flags&flagSidereal != 0 {
				if state.positions.data.sidereal.mode&256 != 0 || state.positions.data.sidereal.mode&512 != 0 {
					for i = 0; i <= 5; i++ {
						x[i] = node.coordinates[18+i]
					}
					if flags&flagNoNutation == 0 {
						state.swiNutate(x[:], flags, 1)
					}
					state.swiPrecess(x[:], jd, flags, 1)
					if flags&flagSpeed != 0 {
						state.swiPrecessSpeed(x[:], jd, flags, 1)
					}
					if state.positions.data.sidereal.mode&256 != 0 {
						state.swiTropRa2sidLon(x[:], node.coordinates[6:], node.coordinates[18:], flags)
					} else if state.positions.data.sidereal.mode&512 != 0 {
						state.swiTropRa2sidLonSosy(x[:], node.coordinates[6:], flags)
					}
					state.swiCartpolSp(node.coordinates[6:], node.coordinates[:])
					state.swiCartpolSp(node.coordinates[18:], node.coordinates[12:])
				} else {
					state.swiCartpolSp(node.coordinates[6:], node.coordinates[:])
					if state.swiGetAyanamsaWithSpeed(node.evaluationTime, flags, daya[:], message) == -1 {
						return -1
					}
					node.coordinates[0] -= daya[0] * radiansPerDegree
					node.coordinates[3] -= daya[1] * radiansPerDegree
					state.swiPolcartSp(node.coordinates[:], node.coordinates[6:])
				}
			} else if flags&flagJ2000 != 0 {
				for i = 0; i <= 5; i++ {
					x[i] = node.coordinates[18+i]
				}
				state.swiPrecess(x[:], jd, flags, 1)
				if flags&flagSpeed != 0 {
					state.swiPrecessSpeed(x[:], jd, flags, 1)
				}
				for i = 0; i <= 5; i++ {
					node.coordinates[18+i] = x[i]
				}
				state.swiCartpolSp(node.coordinates[18:], node.coordinates[12:])
				state.swiCoortrf2(node.coordinates[18:], node.coordinates[6:], state.positions.data.obliquityJ2000.sine, state.positions.data.obliquityJ2000.cosine)
				if flags&flagSpeed != 0 {
					state.swiCoortrf2(node.coordinates[21:], node.coordinates[9:], state.positions.data.obliquityJ2000.sine, state.positions.data.obliquityJ2000.cosine)
				}
				state.swiCartpolSp(node.coordinates[6:], node.coordinates[:])
			}
			for i = 0; i < 2; i++ {
				node.coordinates[i] *= degreesPerRadian
				node.coordinates[i+3] *= degreesPerRadian
				node.coordinates[i+12] *= degreesPerRadian
				node.coordinates[i+15] *= degreesPerRadian
			}
			node.coordinates[0] = state.sweDegnorm(node.coordinates[0])
			node.coordinates[12] = state.sweDegnorm(node.coordinates[12])
		}
		return 0
	}
}

// intpApsides implements intp_apsides from sweph.c.
//
//	lunar osculating elements, i.e.
func (state *State) intpApsides(jd float64, body int32, flags int32, message []byte) int32 {
	var (
		i             int32
		flg1          int32
		flg2          int32
		node          *planetData
		obliquityData *obliquity
		nut           *nutation
		daya          [2]float64
		speedIntv     float64 = 0.1
		t             float64
		dt            float64
		xpos          [3][6]float64
		xx            [6]float64
		x             [6]float64
		speedf1       int32
		speedf2       int32
	)
	obliquityData = &state.positions.data.obliquity
	nut = &state.positions.data.nutation
	node = &state.positions.data.lunarNodes[body]
	flg1 = flags & ^(2 * 1024) & ^(4 * 1024)
	flg2 = node.cachedFlags & ^(2 * 1024) & ^(4 * 1024)
	speedf1 = node.cachedFlags & 256
	speedf2 = flags & (flagSpeed)
	if jd == node.evaluationTime && jd != 0 && flg1 == flg2 && (!(speedf2 != 0) || speedf1 != 0) {
		node.cachedFlags = flags
		node.ephemeris = flags & (flagMoshier)
		return 0
	}
	t = jd - speedIntv
	for i = 0; i < 3; {
		if flags&flagSpeed == 0 && i != 1 {
			t += speedIntv
			i++
			continue
		}
		state.swiIntpApsides(t, xpos[i][:], body)
		t += speedIntv
		i++
	}
	for i = 0; i < 3; i++ {
		xx[i] = xpos[1][i]
		xx[i+3] = 0
	}
	if flags&flagSpeed != 0 {
		xx[3] = state.sweDifrad2n(xpos[2][0], xpos[0][0]) / speedIntv / 2.0
		xx[4] = (xpos[2][1] - xpos[0][1]) / speedIntv / 2.0
		xx[5] = (xpos[2][2] - xpos[0][2]) / speedIntv / 2.0
	}
	clear(node.coordinates[:])
	state.swiPolcartSp(xx[:], xx[:])
	if flags&flagTruePosition == 0 {
		dt = math.Sqrt(xx[0]*xx[0]+xx[1]*xx[1]+xx[2]*xx[2]) * 1.49597870700e+11 / 2.99792458e+8 / secondsPerDay
		for i = 1; i < 3; i++ {
			xx[i] -= dt * xx[i+3]
		}
	}
	for i = 0; i <= 5; i++ {
		node.coordinates[i+6] = xx[i]
	}
	state.swiCoortrf2(node.coordinates[6:], node.coordinates[18:], -obliquityData.sine, obliquityData.cosine)
	if flags&flagSpeed != 0 {
		state.swiCoortrf2(node.coordinates[9:], node.coordinates[21:], -obliquityData.sine, obliquityData.cosine)
	}
	node.evaluationTime = jd
	node.cachedFlags = flags
	node.ephemeris = flags & (flagJPL | flagSwiss | flagMoshier)
	if flags&flagSidereal != 0 {
		if state.positions.data.sidereal.mode&256 != 0 || state.positions.data.sidereal.mode&512 != 0 {
			for i = 0; i <= 5; i++ {
				x[i] = node.coordinates[18+i]
			}
			state.swiPrecess(x[:], jd, flags, 1)
			if flags&flagSpeed != 0 {
				state.swiPrecessSpeed(x[:], jd, flags, 1)
			}
			if state.positions.data.sidereal.mode&256 != 0 {
				state.swiTropRa2sidLon(x[:], node.coordinates[6:], node.coordinates[18:], flags)
			} else if state.positions.data.sidereal.mode&512 != 0 {
				state.swiTropRa2sidLonSosy(x[:], node.coordinates[6:], flags)
			}
			state.swiCartpolSp(node.coordinates[6:], node.coordinates[:])
			state.swiCartpolSp(node.coordinates[18:], node.coordinates[12:])
		} else {
			state.swiCartpolSp(node.coordinates[6:], node.coordinates[:])
			if state.swiGetAyanamsaWithSpeed(node.evaluationTime, flags, daya[:], message) == -1 {
				return -1
			}
			node.coordinates[0] -= daya[0] * radiansPerDegree
			node.coordinates[3] -= daya[1] * radiansPerDegree
			state.swiPolcartSp(node.coordinates[:], node.coordinates[6:])
			state.swiCartpolSp(node.coordinates[18:], node.coordinates[12:])
		}
	} else if flags&flagJ2000 != 0 {
		for i = 0; i <= 5; i++ {
			x[i] = node.coordinates[18+i]
		}
		state.swiPrecess(x[:], jd, flags, 1)
		if flags&flagSpeed != 0 {
			state.swiPrecessSpeed(x[:], jd, flags, 1)
		}
		for i = 0; i <= 5; i++ {
			node.coordinates[18+i] = x[i]
		}
		state.swiCartpolSp(node.coordinates[18:], node.coordinates[12:])
		state.swiCoortrf2(node.coordinates[18:], node.coordinates[6:], state.positions.data.obliquityJ2000.sine, state.positions.data.obliquityJ2000.cosine)
		if flags&flagSpeed != 0 {
			state.swiCoortrf2(node.coordinates[21:], node.coordinates[9:], state.positions.data.obliquityJ2000.sine, state.positions.data.obliquityJ2000.cosine)
		}
		state.swiCartpolSp(node.coordinates[6:], node.coordinates[:])
	} else {
		if flags&flagNoNutation == 0 {
			state.swiNutate(node.coordinates[18:], flags, 0)
		}
		state.swiCartpolSp(node.coordinates[18:], node.coordinates[12:])
		state.swiCoortrf2(node.coordinates[18:], node.coordinates[6:], obliquityData.sine, obliquityData.cosine)
		if flags&flagSpeed != 0 {
			state.swiCoortrf2(node.coordinates[21:], node.coordinates[9:], obliquityData.sine, obliquityData.cosine)
		}
		if flags&flagNoNutation == 0 {
			state.swiCoortrf2(node.coordinates[6:], node.coordinates[6:], nut.sine, nut.cosine)
			if flags&flagSpeed != 0 {
				state.swiCoortrf2(node.coordinates[9:], node.coordinates[9:], nut.sine, nut.cosine)
			}
		}
		state.swiCartpolSp(node.coordinates[6:], node.coordinates[:])
	}
	for i = 0; i < 2; i++ {
		node.coordinates[i] *= degreesPerRadian
		node.coordinates[i+3] *= degreesPerRadian
		node.coordinates[i+12] *= degreesPerRadian
		node.coordinates[i+15] *= degreesPerRadian
	}
	node.coordinates[0] = state.sweDegnorm(node.coordinates[0])
	node.coordinates[12] = state.sweDegnorm(node.coordinates[12])
	return 0
}

// swiPlanForOscElem implements swi_plan_for_osc_elem from sweph.c.
func (state *State) swiPlanForOscElem(flags int32, jd float64, xx []float64) int32 {
	var (
		i             int32
		x             [6]float64
		nuttmp        nutation
		nutationData  *nutation  = &nuttmp
		obliquityData *obliquity = &state.positions.data.obliquity
		oectmp        obliquity
	)
	if flags&flagICRS == 0 && state.swiGetDenum(0, flags) >= 403 {
		state.swiBias(xx, jd, flags, 0)
	}
	state.swiPrecess(xx, jd, flags, -1)
	state.swiPrecess(xx[3:], jd, flags, -1)
	if jd == state.positions.data.obliquity.epoch {
		obliquityData = &state.positions.data.obliquity
	} else if jd == julianDayJ2000 {
		obliquityData = &state.positions.data.obliquityJ2000
	} else {
		state.calcEpsilon(jd, flags, &oectmp)
		obliquityData = &oectmp
	}
	if flags&flagNoNutation == 0 {
		if jd == state.positions.data.nutation.epoch {
			nutationData = &state.positions.data.nutation
		} else if jd == julianDayJ2000 {
			nutationData = &state.positions.data.nutationJ2000
		} else if jd == state.positions.data.nutationWithSpeed.epoch {
			nutationData = &state.positions.data.nutationWithSpeed
		} else {
			nutationData = &nuttmp
			state.swiNutation(jd, flags, nutationData.angles[:])
			nutationData.epoch = jd
			nutationData.sine = math.Sin(nutationData.angles[1])
			nutationData.cosine = math.Cos(nutationData.angles[1])
			state.nutMatrix(nutationData, obliquityData)
		}
		for i = 0; i <= 2; i++ {
			x[i] = xx[0]*nutationData.matrix[0][i] + xx[1]*nutationData.matrix[1][i] + xx[2]*nutationData.matrix[2][i]
		}
		for i = 0; i <= 2; i++ {
			x[i+3] = xx[3]*nutationData.matrix[0][i] + xx[4]*nutationData.matrix[1][i] + xx[5]*nutationData.matrix[2][i]
		}
		for i = 0; i <= 5; i++ {
			xx[i] = x[i]
		}
	}
	state.swiCoortrf2(xx, xx, obliquityData.sine, obliquityData.cosine)
	state.swiCoortrf2(xx[3:], xx[3:], obliquityData.sine, obliquityData.cosine)
	if flags&flagNoNutation == 0 {
		state.swiCoortrf2(xx, xx, nutationData.sine, nutationData.cosine)
		state.swiCoortrf2(xx[3:], xx[3:], nutationData.sine, nutationData.cosine)
	}
	return 0
}

// meff implements meff from sweph.c.
func (state *State) meff(r float64) float64 {
	var (
		f float64
		m float64
		i int32
	)
	if r <= 0 {
		return 0.0
	} else if r >= 1 {
		return 1.0
	}
	for i = 0; positionsEffArr[i].radiusFraction > r; i++ {
	}
	f = (r - positionsEffArr[i-1].radiusFraction) / (positionsEffArr[i].radiusFraction - positionsEffArr[i-1].radiusFraction)
	m = positionsEffArr[i-1].massFraction + f*(positionsEffArr[i].massFraction-positionsEffArr[i-1].massFraction)
	return m
}

// denormalizePositions implements denormalize_positions from sweph.c.
func (state *State) denormalizePositions(x0 []float64, x1 []float64, x2 []float64) {
	var i int32
	for i = 0; i <= 12; i += 12 {
		if x1[i]-x0[i] < float64(-180) {
			x0[i] -= 360
		}
		if x1[i]-x0[i] > 180 {
			x0[i] += 360
		}
		if x1[i]-x2[i] < float64(-180) {
			x2[i] -= 360
		}
		if x1[i]-x2[i] > 180 {
			x2[i] += 360
		}
	}
}

// calcSpeed implements calc_speed from sweph.c.
func (state *State) calcSpeed(x0 []float64, x1 []float64, x2 []float64, dt float64) {
	var (
		i int32
		j int32
		k int32
		a float64
		b float64
	)
	for j = 0; j <= 18; j += 6 {
		for i = 0; i < 3; i++ {
			k = j + i
			b = (x2[k] - x0[k]) / 2
			a = (x2[k]+x0[k])/2 - x1[k]
			x1[k+3] = (2*a + b) / dt
		}
	}
}

// swiCheckEcliptic implements swi_check_ecliptic from sweph.c.
func (state *State) swiCheckEcliptic(jd float64, flags int32) {
	if state.positions.data.obliquityJ2000.epoch != julianDayJ2000 {
		state.calcEpsilon(julianDayJ2000, flags, &state.positions.data.obliquityJ2000)
	}
	if jd == julianDayJ2000 {
		state.positions.data.obliquity.epoch = state.positions.data.obliquityJ2000.epoch
		state.positions.data.obliquity.angle = state.positions.data.obliquityJ2000.angle
		state.positions.data.obliquity.sine = state.positions.data.obliquityJ2000.sine
		state.positions.data.obliquity.cosine = state.positions.data.obliquityJ2000.cosine
		return
	}
	if state.positions.data.obliquity.epoch != jd || jd == 0 {
		state.calcEpsilon(jd, flags, &state.positions.data.obliquity)
	}
}

// swiCheckNutation implements swi_check_nutation from sweph.c.
//
//	computes nutation, if it is wanted and has not yet been computed.
//
// if speed flag has been turned on since last computation,
// nutation is recomputed
func (state *State) swiCheckNutation(jd float64, flags int32) {
	var (
		speedf1 int32
		speedf2 int32
		t       float64
	)
	speedf1 = state.positions.checkNutationNutflag & 256
	speedf2 = flags & (flagSpeed)
	if flags&flagNoNutation == 0 && (jd != state.positions.data.nutation.epoch || jd == 0 || !(speedf1 != 0) && speedf2 != 0) {
		state.swiNutation(jd, flags, state.positions.data.nutation.angles[:])
		state.positions.data.nutation.epoch = jd
		state.positions.data.nutation.sine = math.Sin(state.positions.data.nutation.angles[1])
		state.positions.data.nutation.cosine = math.Cos(state.positions.data.nutation.angles[1])
		state.positions.checkNutationNutflag = flags
		state.nutMatrix(&state.positions.data.nutation, &state.positions.data.obliquity)
		if flags&flagSpeed != 0 {
			t = jd - 0.0001
			state.swiNutation(t, flags, state.positions.data.nutationWithSpeed.angles[:])
			state.positions.data.nutationWithSpeed.epoch = t
			state.positions.data.nutationWithSpeed.sine = math.Sin(state.positions.data.nutationWithSpeed.angles[1])
			state.positions.data.nutationWithSpeed.cosine = math.Cos(state.positions.data.nutationWithSpeed.angles[1])
			state.nutMatrix(&state.positions.data.nutationWithSpeed, &state.positions.data.obliquity)
		}
	}
}

// plausIflag implements plaus_iflag from sweph.c.
//
//	function
//
// - corrects nonsensical iflags
// - completes incomplete iflags
func (state *State) plausIflag(flags int32, body int32, jd float64, message []byte) int32 {
	var (
		ephemeris    int32 = 0
		jplhorModel  int32 = state.positions.data.astroModels[5]
		jplhoraModel int32 = state.positions.data.astroModels[6]
	)
	if jplhorModel == 0 {
		jplhorModel = 1
	}
	if jplhoraModel == 0 {
		jplhoraModel = 3
	}
	if flags&flagHorizons != 0 {
		flags &= ^(512 * 1024)
	}
	if flags&flagTopocentric != 0 {
		flags = flags & ^(8 | 16*1024)
	}
	if flags&flagBarycentric != 0 {
		flags = flags & ^8
	}
	if flags&flagHeliocentric != 0 {
		flags = flags & ^(16 * 1024)
	}
	if flags&(flagHeliocentric|flagBarycentric) != 0 {
		flags |= flagNoDeflection | flagNoAberration
	}
	if flags&flagJ2000 != 0 {
		flags |= flagNoNutation
	}
	if flags&flagSidereal != 0 {
		flags |= flagNoNutation
		flags = flags & ^(256*1024 | 512*1024)
	}
	if flags&flagTruePosition != 0 {
		flags |= flagNoDeflection | flagNoAberration
	}
	if flags&flagMoshier != 0 {
		ephemeris = 4
	}
	if flags&flagSwiss != 0 {
		ephemeris = 2
	}
	if flags&flagJPL != 0 {
		ephemeris = 1
	}
	if ephemeris == 0 {
		ephemeris = 2
	}
	flags = flags & ^(1|2|4) | ephemeris
	if !(ephemeris&1 != 0) {
		flags = flags & ^(256*1024 | 512*1024)
	}
	if body == 13 || body == 11 || body == 12 || body == 10 || body == 21 || body == 22 {
		flags = flags & ^(256*1024 | 512*1024)
	}
	if body >= 40 && body <= 999 {
		flags = flags & ^(256*1024 | 512*1024)
	}
	if flags&flagHorizons != 0 {
		if state.positions.data.eopDpsiLoaded <= 0 {
			if message != nil {
				switch state.positions.data.eopDpsiLoaded {
				case 0:
					copyText(message, textBytes("you did not call swe_set_jpl_file(); default to SEFLG_JPLHOR_APPROX"))
				case -1:
					copyText(message, textBytes("file eop_1962_today.txt not found; default to SEFLG_JPLHOR_APPROX"))
				case -2:
					copyText(message, textBytes("file eop_1962_today.txt corrupt; default to SEFLG_JPLHOR_APPROX"))
				case -3:
					copyText(message, textBytes("file eop_finals.txt corrupt; default to SEFLG_JPLHOR_APPROX"))
				}
			}
			flags &= ^(256 * 1024)
			flags |= flagHorizonsApprox
		}
	}
	if flags&flagHorizons != 0 {
		flags |= flagICRS
	}
	if flags&flagHorizonsApprox != 0 && jplhoraModel == 2 {
		flags |= flagICRS
	}
	return flags
}

// fixstarFormatSearchName implements fixstar_format_search_name from sweph.c.
//
//	function formats the input search name of a star:
//
// - remove white spaces
// - traditional name to lower case (Bayer designation remains as it is)
func (state *State) fixstarFormatSearchName(star []byte, sstar []byte, message []byte) int32 {
	var (
		sp     []byte
		cmplen uint64
	)
	copyTextN(sstar, star, 40)
	sstar[40] = byte('\000')
	for {
		sp = findByte(sstar, ' ')
		if !(sp != nil) {
			break
		}
		state.swiStrcpy(sp, sp[1:])
	}
	for sp = sstar; int32(sp[0]) != '\000' && int32(sp[0]) != ','; sp = sp[1:] {
		sp[0] = byte(lowerByte(int32(sp[0])))
	}
	cmplen = textLength(sstar)
	if cmplen == 0 {
		if message != nil {
			formatText(message, textBytes("swe_fixstar(): star name empty"))
		}
		return -1
	}
	return 0
}

// fixstarCutString implements fixstar_cut_string from sweph.c.
//
//	function cuts a comma-separated fixed star data record from sefstars.txt
//
// and fills it into a struct fixed_star.
func (state *State) fixstarCutString(srecord []byte, star []byte, stardata *fixedStar, message []byte) int32 {
	var (
		i      int32
		s      [256]byte
		sdeD   []byte
		cpos   [20][]byte
		epoch  float64
		radv   float64
		parall float64
		mag    float64
		raS    float64
		raPm   float64
		dePm   float64
		ra     float64
		de     float64
		raH    float64
		raM    float64
		deD    float64
		deM    float64
		deS    float64
	)
	copyText(s[:], srecord)
	i = state.swiCutstr(s[:], textBytes(","), cpos[:], 20)
	state.swiRightTrim(cpos[0])
	state.swiRightTrim(cpos[1])
	if i < 14 {
		if message != nil {
			if i >= 2 {
				formatText(message, textBytes("data of star '%s,%s' incomplete"), cpos[0], cpos[1])
			} else {
				if textLength(s[:]) > 200 {
					s[200] = byte('\000')
				}
				formatText(message, textBytes("invalid line in fixed stars file: '%s'"), s[:])
			}
		}
		return -1
	}
	if textLength(cpos[0]) > 40 {
		cpos[0][40] = byte('\000')
	}
	if textLength(cpos[1]) > uint64(40-1) {
		cpos[1][40-1] = byte('\000')
	}
	if star != nil {
		copyText(star, cpos[0])
		if textLength(cpos[0])+textLength(cpos[1])+1 < uint64(40-1) {
			formatText(star[textLength(star):], textBytes(",%s"), cpos[1])
		}
	}
	copyText(stardata.name[:], cpos[0])
	copyText(stardata.bayerName[:], cpos[1])
	epoch = parseFloat(cpos[2])
	raH = parseFloat(cpos[3])
	raM = parseFloat(cpos[4])
	raS = parseFloat(cpos[5])
	deD = parseFloat(cpos[6])
	sdeD = cpos[6]
	deM = parseFloat(cpos[7])
	deS = parseFloat(cpos[8])
	raPm = parseFloat(cpos[9])
	dePm = parseFloat(cpos[10])
	radv = parseFloat(cpos[11])
	parall = parseFloat(cpos[12])
	mag = parseFloat(cpos[13])
	ra = (raS/3600.0 + raM/60.0 + raH) * 15.0
	if findByte(sdeD, '-') == nil {
		de = deS/3600.0 + deM/60.0 + deD
	} else {
		de = -deS/3600.0 - deM/60.0 + deD
	}
	if state.positions.data.isOldStarfile == 1 {
		raPm = raPm * 15 / 3600.0
		dePm = dePm / 3600.0
	} else {
		raPm = raPm / 10.0 / 3600.0
		dePm = dePm / 10.0 / 3600.0
		parall /= 1000.0
	}
	if parall > 1 {
		parall = 1 / parall / 3600.0
	} else {
		parall /= 3600
	}
	radv *= 21.095
	ra *= radiansPerDegree
	de *= radiansPerDegree
	raPm *= radiansPerDegree
	dePm *= radiansPerDegree
	raPm /= math.Cos(de)
	parall *= radiansPerDegree
	stardata.epoch = epoch
	stardata.rightAscension = ra
	stardata.declination = de
	stardata.properMotionRA = raPm
	stardata.properMotionDec = dePm
	stardata.parallax = parall
	stardata.radialVelocity = radv
	stardata.magnitude = mag
	return 0
}

// loadAllFixedStars implements load_all_fixed_stars from sweph.c.
//
//	function loads all fixed stars from file sefstars.txt,
//
// into swed.fixed_stars, which is a pointer to an array
// of struct fixed_stars.
// Every star has a record with its Bayer/Flamsteed designation
// as its search key.
// Every star also has a record with its sequential number in
// the file as its search key. (Good for calculating all stars in a loop.)
// If a star has a traditional name, we create a record that has
// this name as its search key.
// The array is sorted in ascending order by search key.
//
// If an error occurs, the function returns value ERR.
// If the stars were loaded at an earlier time the function returns
// value -2, without doing anything and without error string.
// On success, the function returns value OK.
func (state *State) loadAllFixedStars(message []byte) int32 {
	var (
		status        int32 = 0
		nstars        int32 = 0
		line          int32 = 0
		fline         int32 = 0
		nrecs         int32 = 0
		nnamed        int32 = 0
		s             [256]byte
		sp            []byte
		srecord       [256]byte
		fstdata       fixedStar
		lastStarbayer [40 + 1]byte
	)
	lastStarbayer[0] = byte('\000')
	if state.positions.data.nFixstarsRecords > 0 {
		return -2
	}
	if state.positions.data.starFile == nil {
		if state.positions.data.starFile = state.swiFopen(4, textBytes("sefstars.txt"), state.positions.data.ephemerisPath[:], message); state.positions.data.starFile == nil {
			state.positions.data.isOldStarfile = 1
			if state.positions.data.starFile = state.swiFopen(4, textBytes("fixstars.cat"), state.positions.data.ephemerisPath[:], nil); state.positions.data.starFile == nil {
				state.positions.data.isOldStarfile = 0
				return -1
			}
		}
	}
	rewindFile(state.positions.data.starFile)
	state.positions.data.fixedStars = nil
	for readLine(s[:], 256, state.positions.data.starFile) != nil {
		fline++
		if int32(s[0]) == '#' {
			continue
		}
		if int32(s[0]) == '\n' {
			continue
		}
		if int32(s[0]) == '\r' {
			continue
		}
		if int32(s[0]) == '\000' {
			continue
		}
		line++
		copyText(srecord[:], s[:])
		status = state.fixstarCutString(srecord[:], nil, &fstdata, message)
		if status == -1 {
			return -1
		}
		if int32(fstdata.name[0]) != '\000' {
			nrecs++
			nnamed++
			copyText(fstdata.searchKey[:], fstdata.name[:])
			for {
				sp = findByte(fstdata.searchKey[:], ' ')
				if !(sp != nil) {
					break
				}
				state.swiStrcpy(sp, sp[1:])
			}
			for sp = fstdata.searchKey[:]; int32(sp[0]) != '\000'; sp = sp[1:] {
				sp[0] = byte(lowerByte(int32(sp[0])))
			}
			if status = state.saveStarInStruct(nrecs, &fstdata, message); status == -1 {
				return -1
			}
		}
		if compareText(fstdata.bayerName[:], lastStarbayer[:]) == 0 {
			continue
		}
		nstars++
		nrecs++
		formatText(fstdata.searchKey[:], textBytes(",%s"), fstdata.bayerName[:])
		for {
			sp = findByte(fstdata.searchKey[:], ' ')
			if !(sp != nil) {
				break
			}
			state.swiStrcpy(sp, sp[1:])
		}
		copyText(lastStarbayer[:], fstdata.bayerName[:])
		if status = state.saveStarInStruct(nrecs, &fstdata, message); status == -1 {
			return -1
		}
	}
	state.positions.data.nFixstarsReal = nstars
	state.positions.data.nFixstarsNamed = nnamed
	state.positions.data.nFixstarsRecords = nrecs
	sortRecords(state.positions.data.fixedStars, uint64(nrecs), 0, fixedstarNameCompare)
	return status
}

// fixstarCalcFromStruct implements fixstar_calc_from_struct from sweph.c.
//
//	function calculates a fixstar from a star data struct
//
// input:
// struct fixed_star stardata      fixed star data struct
// double tjd        julian daynumber
// int32 iflag       SEFLG_ specifications
// output:
// char *star        star name, Bayer designation
// double xx[6]      position and speed
// char *serr        error return string
func (state *State) fixstarCalcFromStruct(stardata *fixedStar, jd float64, flags int32, star []byte, xx []float64, message []byte) int32 {
	var (
		i              int32
		status         int32 = 0
		epoch          float64
		radv           float64
		parall         float64
		raPm           float64
		dePm           float64
		ra             float64
		de             float64
		t              float64
		daya           [2]float64
		rdist          float64
		x              [6]float64
		xxsv           [6]float64
		observerVector [6]float64
		xobsDt         [6]float64
		xpo            []float64 = nil
		xpoDt          []float64 = nil
		dt             float64   = 1.0000000000000001e-05
		ephemeris      int32
		savedFlags     int32
		obliquityData  *obliquity = &state.positions.data.obliquityJ2000
	)
	savedFlags = flags
	flags |= flagSpeed
	if message != nil {
		message[0] = byte('\000')
	}
	flags = state.plausIflag(flags, -1, jd, message)
	ephemeris = flags & (flagJPL | flagSwiss | flagMoshier)
	if state.swiInitSwedIfStart() == 1 && !(ephemeris&4 != 0) && message != nil {
		copyText(message, textBytes("Please call swe_set_ephe_path() or swe_set_jplfile() before calling swe_fixstar() or swe_fixstar_ut()"))
	}
	if state.positions.data.lastEpheflag != ephemeris {
		state.freePlanets()
		if state.positions.data.jplFileIsOpen != 0 {
			state.swiCloseJplFile()
			state.positions.data.jplFileIsOpen = 0
		}
		for i = 0; i < 7; i++ {
			if state.positions.data.files[i].handle != nil {
				closeTextFile(state.positions.data.files[i].handle)
			}
			state.positions.data.files[i] = ephemerisFile{}
		}
		state.positions.data.lastEpheflag = ephemeris
	}
	if flags&flagSidereal != 0 && !(state.positions.data.ayanaIsSet != 0) {
		state.sweSetSidMode(0, 0, 0)
	}
	state.swiCheckEcliptic(jd, flags)
	state.swiCheckNutation(jd, flags)
	formatText(star, textBytes("%s,%s"), stardata.name[:], stardata.bayerName[:])
	epoch = stardata.epoch
	raPm = stardata.properMotionRA
	dePm = stardata.properMotionDec
	radv = stardata.radialVelocity
	parall = stardata.parallax
	ra = stardata.rightAscension
	de = stardata.declination
	if epoch == 1950 {
		t = jd - 2433282.42345905
	} else {
		t = jd - julianDayJ2000
	}
	x[0] = ra
	x[1] = de
	x[2] = 1
	if parall == 0 {
		rdist = 1000000000
	} else {
		rdist = 1.0 / (parall * degreesPerRadian * 3600) * 206264.8062471
	}
	x[2] = rdist
	x[3] = raPm / daysPerCentury
	x[4] = dePm / daysPerCentury
	x[5] = radv / daysPerCentury
	state.swiPolcartSp(x[:], x[:])
	if epoch == 1950 {
		state.swiFK4FK5(x[:], 2433282.42345905)
		state.swiPrecess(x[:], 2433282.42345905, 0, 1)
		state.swiPrecess(x[3:], 2433282.42345905, 0, 1)
	}
	if epoch != 0 {
		state.swiIcrs2fk5(x[:], flags, 1)
		if state.swiGetDenum(0, flags) >= 403 {
			state.swiBias(x[:], julianDayJ2000, 256, 0)
		}
	}
	if flags&flagBarycentric == 0 && (flags&flagHeliocentric == 0 || flags&flagMoshier == 0) {
		if status = state.mainPlanetBary(jd-dt, 0, ephemeris, flags, 0, state.positions.fixstarCalcFromStructXearthDt[:], state.positions.fixstarCalcFromStructXearthDt[:], state.positions.fixstarCalcFromStructXsunDt[:], nil, message); status != 0 {
			return -1
		}
		if status = state.mainPlanetBary(jd, 0, ephemeris, flags, 1, state.positions.fixstarCalcFromStructXearth[:], state.positions.fixstarCalcFromStructXearth[:], state.positions.fixstarCalcFromStructXsun[:], nil, message); status != 0 {
			return -1
		}
	}
	if flags&flagTopocentric != 0 {
		if state.swiGetObserver(jd-dt, flags|(flagNoNutation), 0, xobsDt[:], message) != 0 {
			return -1
		}
		if state.swiGetObserver(jd, flags|(flagNoNutation), 0, observerVector[:], message) != 0 {
			return -1
		}
		for i = 0; i <= 5; i++ {
			observerVector[i] = observerVector[i] + state.positions.fixstarCalcFromStructXearth[i]
			xobsDt[i] = xobsDt[i] + state.positions.fixstarCalcFromStructXearthDt[i]
		}
	} else if flags&flagBarycentric == 0 && (flags&flagHeliocentric == 0 || flags&flagMoshier == 0) {
		for i = 0; i <= 5; i++ {
			observerVector[i] = state.positions.fixstarCalcFromStructXearth[i]
			xobsDt[i] = state.positions.fixstarCalcFromStructXearthDt[i]
		}
	}
	if flags&flagHeliocentric != 0 && flags&flagMoshier != 0 {
		xpo = nil
		xpoDt = nil
	} else if flags&flagHeliocentric != 0 {
		xpo = state.positions.fixstarCalcFromStructXsun[:]
		xpoDt = state.positions.fixstarCalcFromStructXsunDt[:]
	} else if flags&flagBarycentric != 0 {
		xpo = nil
		xpoDt = nil
	} else {
		xpo = observerVector[:]
		xpoDt = xobsDt[:]
	}
	if xpo == nil {
		for i = 0; i <= 2; i++ {
			x[i] += t * x[i+3]
		}
	} else {
		for i = 0; i <= 2; i++ {
			x[i] += t * x[i+3]
			x[i] -= xpo[i]
			x[i+3] -= xpo[i+3]
		}
	}
	if flags&flagTruePosition == 0 && flags&flagNoDeflection == 0 {
		state.swiDeflectLight(x[:], 0, flags&flagSpeed)
	}
	if flags&flagTruePosition == 0 && flags&flagNoAberration == 0 {
		state.swiAberrLightEx(x[:], xpo, xpoDt, dt, flags&flagSpeed)
	}
	if flags&flagICRS == 0 && (state.swiGetDenum(0, flags) >= 403 || flags&flagBarycentric != 0) {
		state.swiBias(x[:], jd, flags, 0)
	}
	for i = 0; i <= 5; i++ {
		xxsv[i] = x[i]
	}
	if flags&flagJ2000 == 0 {
		state.swiPrecess(x[:], jd, flags, -1)
		if flags&flagSpeed != 0 {
			state.swiPrecessSpeed(x[:], jd, flags, -1)
		}
		obliquityData = &state.positions.data.obliquity
	} else {
		obliquityData = &state.positions.data.obliquityJ2000
	}
	if flags&flagNoNutation == 0 {
		state.swiNutate(x[:], flags, 0)
	}
	if 0 != 0 {
		var r float64 = math.Sqrt(x[0]*x[0] + x[1]*x[1] + x[2]*x[2])
		printText(textBytes("%.17f %.17f %f\n"), x[0]/r, x[1]/r, x[2]/r)
	}
	if flags&flagEquatorial == 0 {
		state.swiCoortrf2(x[:], x[:], obliquityData.sine, obliquityData.cosine)
		if flags&flagSpeed != 0 {
			state.swiCoortrf2(x[3:], x[3:], obliquityData.sine, obliquityData.cosine)
		}
		if flags&flagNoNutation == 0 {
			state.swiCoortrf2(x[:], x[:], state.positions.data.nutation.sine, state.positions.data.nutation.cosine)
			if flags&flagSpeed != 0 {
				state.swiCoortrf2(x[3:], x[3:], state.positions.data.nutation.sine, state.positions.data.nutation.cosine)
			}
		}
	}
	if flags&flagSidereal != 0 {
		if state.positions.data.sidereal.mode&256 != 0 {
			if state.swiTropRa2sidLon(xxsv[:], x[:], xxsv[:], flags) != 0 {
				return -1
			}
			if flags&flagEquatorial != 0 {
				for i = 0; i <= 5; i++ {
					x[i] = xxsv[i]
				}
			}
		} else if state.positions.data.sidereal.mode&512 != 0 {
			if state.swiTropRa2sidLonSosy(xxsv[:], x[:], flags) != 0 {
				return -1
			}
			if flags&flagEquatorial != 0 {
				for i = 0; i <= 5; i++ {
					x[i] = xxsv[i]
				}
			}
		} else {
			state.swiCartpolSp(x[:], x[:])
			if state.swiGetAyanamsaWithSpeed(jd, flags, daya[:], message) == -1 {
				return -1
			}
			x[0] -= daya[0] * radiansPerDegree
			x[3] -= daya[1] * radiansPerDegree
			state.swiPolcartSp(x[:], x[:])
		}
	}
	if flags&flagCartesian == 0 {
		state.swiCartpolSp(x[:], x[:])
	}
	if flags&flagRadians == 0 && flags&flagCartesian == 0 {
		for i = 0; i < 2; i++ {
			x[i] *= degreesPerRadian
			x[i+3] *= degreesPerRadian
		}
	}
	for i = 0; i <= 5; i++ {
		xx[i] = x[i]
	}
	if savedFlags&flagSpeed == 0 {
		for i = 3; i <= 5; i++ {
			xx[i] = 0
		}
	}
	if savedFlags&(flagJPL|flagSwiss|flagMoshier) == 0 {
		flags = flags & ^2
	}
	flags = flags & ^256
	return flags
}

// searchStarInList implements search_star_in_list from sweph.c.
//
//	function searches a star in fixed stars list, i.e. the data loaded from file
//
// sefstars.txt
func (state *State) searchStarInList(sstar []byte, stardata *fixedStar, message []byte) int32 {
	var (
		i            int32
		starNr       int32 = 0
		ndata        int32 = 0
		len          int32
		sp           []byte
		searchkey    [256]byte
		isBayer      int32 = 0
		stardatap    *fixedStar
		stardatabegp []fixedStar
	)
	if int32(sstar[0]) == ',' {
		isBayer = 1
	} else if isDigit(int32(sstar[0])) != 0 {
		starNr = parseInt(sstar)
	} else if sp = findByte(sstar, ','); sp != nil {
		state.swiStrcpy(sstar, sp)
		isBayer = 1
	}
	if starNr > 0 {
		if starNr > state.positions.data.nFixstarsReal {
			if message != nil {
				formatText(message, textBytes("error, swe_fixstar(): sequential fixed star number %d is not available"), starNr)
			}
			return -1
		}
		*stardata = state.positions.data.fixedStars[starNr-1]
		return 0
	} else if isBayer == 0 && findByte(sstar, '%') != nil {
		sp = findByte(sstar, '%')
		stardatabegp = state.positions.data.fixedStars[state.positions.data.nFixstarsReal:]
		ndata = state.positions.data.nFixstarsNamed
		if uint64(sliceOffset(sp, sstar)) != textLength(sstar)-1 {
			if message != nil {
				formatText(message, textBytes("error, swe_fixstar(): invalid search string %s"), sstar)
			}
			return -1
		}
		copyText(searchkey[:], sstar)
		len = int32(textLength(sstar) - 1)
		searchkey[len] = byte('\000')
		for i = 0; i < ndata; i++ {
			if compareTextN(stardatabegp[i].searchKey[:], sstar, len) == 0 {
				*stardata = stardatabegp[i]
				return 0
			}
		}
		if message != nil {
			formatText(message, textBytes("error, swe_fixstar(): star search string %s did not match"), sstar)
		}
		return -1
	} else {
		copyText(searchkey[:], sstar)
		if isBayer != 0 {
			stardatabegp = state.positions.data.fixedStars
			ndata = state.positions.data.nFixstarsReal
		} else {
			stardatabegp = state.positions.data.fixedStars[state.positions.data.nFixstarsReal:]
			ndata = state.positions.data.nFixstarsNamed
		}
		stardatap = searchRecords(searchkey[:], stardatabegp, uint64(ndata), 0, fstarNodeCompare)
		if stardatap == nil {
			if message != nil {
				formatText(message, textBytes("error, swe_fixstar(): could not find star name %s"), sstar)
			}
			return -1
		}
		*stardata = *stardatap
		return 0
	}
}

// getBuiltinStar implements get_builtin_star from sweph.c.
func (state *State) getBuiltinStar(star []byte, sstar []byte, srecord []byte) int32 {
	if compareTextN(star, textBytes("spica"), 5) == 0 || compareTextN(star, textBytes("Spica"), 5) == 0 {
		copyText(srecord, textBytes("Spica,alVir,ICRS,13,25,11.57937,-11,09,40.7501,-42.35,-30.67,1,13.06,0.97,-10,3672"))
		copyText(sstar, textBytes("spica"))
		return 1
	} else if findText(star, textBytes(",zePsc")) != nil || compareTextN(star, textBytes("revati"), 6) == 0 || compareTextN(star, textBytes("Revati"), 6) == 0 {
		copyText(srecord, textBytes("Revati,zePsc,ICRS,01,13,43.88735,+07,34,31.2745,145,-55.69,15,18.76,5.187,06,174"))
		copyText(sstar, textBytes("revati"))
		return 1
	} else if findText(star, textBytes(",deCnc")) != nil || compareTextN(star, textBytes("pushya"), 6) == 0 || compareTextN(star, textBytes("Pushya"), 6) == 0 {
		copyText(srecord, textBytes("Pushya,deCnc,ICRS,08,44,41.09921,+18,09,15.5034,-17.67,-229.26,17.14,24.98,3.94,18,2027"))
		copyText(sstar, textBytes("pushya"))
		return 1
	} else if findText(star, textBytes(",deCnc")) != nil {
		copyText(srecord, textBytes("Pushya,deCnc,ICRS,08,44,41.09921,+18,09,15.5034,-17.67,-229.26,17.14,24.98,3.94,18,2027"))
		copyText(sstar, textBytes("pushya"))
		return 1
	} else if findText(star, textBytes(",laSco")) != nil || compareTextN(star, textBytes("mula"), 6) == 0 || compareTextN(star, textBytes("Mula"), 6) == 0 {
		copyText(srecord, textBytes("Mula,laSco,ICRS,17,33,36.52012,-37,06,13.7648,-8.53,-30.8,-3,5.71,1.62,-37,11673"))
		copyText(sstar, textBytes("mula"))
		return 1
	} else if findText(star, textBytes(",SgrA*")) != nil {
		copyText(srecord, textBytes("Gal. Center,SgrA*,2000,17,45,40.03599,-29,00,28.1699,-2.755718425,-5.547,0.0,0.125,999.99,0,0"))
		copyText(sstar, textBytes(",SgrA*"))
		return 1
	} else if findText(star, textBytes(",GP1958")) != nil {
		copyText(srecord, textBytes("Gal. Pole IAU1958,GP1958,1950,12,49,0.0,27,24,0.0,0.0,0.0,0.0,0.0,0.0,0,0"))
		copyText(sstar, textBytes(",GP1958"))
		return 1
	} else if findText(star, textBytes(",GPol")) != nil {
		copyText(srecord, textBytes("Gal. Pole,GPol,ICRS,12,51,36.7151981,27,06,11.193172,0.0,0.0,0.0,0.0,0.0,0,0"))
		copyText(sstar, textBytes(",GPol"))
		return 1
	} else if findText(star, textBytes(",GPol")) != nil {
		copyText(srecord, textBytes("Gal. Pole,GPol,ICRS,12,51,36.7151981,27,06,11.193172,0.0,0.0,0.0,0.0,0.0,0,0"))
		copyText(sstar, textBytes(",GPol"))
		return 1
	}
	return 0
}

// sweFixstar2 implements swe_fixstar2 from sweph.c.
func (state *State) sweFixstar2(star []byte, jd float64, flags int32, xx []float64, message []byte) int32 {
	var (
		i             int32
		isBuiltinStar int32 = 0
		sstar         [40 + 1]byte
		srecord       [256 + 20]byte
		status        int32
		stardata      fixedStar
	)
found:
	for {
		if message != nil {
			message[0] = byte('\000')
		}
		state.loadAllFixedStars(message)
		status = state.fixstarFormatSearchName(star, sstar[:], message)
		if status == -1 {
			i = 0
			for i = 0; i <= 5; i++ {
				xx[i] = 0
			}
			return status
		}
		if state.positions.data.nFixstarsRecords > 0 && compareText(state.positions.fixstar2SlastStarname[:], sstar[:]) == 0 {
			stardata = state.positions.fixstar2LastStardata
			break found
		}
		if state.getBuiltinStar(star, sstar[:], srecord[:]) != 0 {
			isBuiltinStar = 1
		}
		if isBuiltinStar != 0 {
			status = state.fixstarCutString(srecord[:], star, &stardata, message)
			break found
		}
		status = state.searchStarInList(sstar[:], &stardata, message)
		if status == -1 {
			i = 0
			for i = 0; i <= 5; i++ {
				xx[i] = 0
			}
			return status
		}
		break
	}
	state.positions.fixstar2LastStardata = stardata
	copyText(state.positions.fixstar2SlastStarname[:], sstar[:])
	if status = state.fixstarCalcFromStruct(&stardata, jd, flags, star, xx, message); status == -1 {
		i = 0
		for i = 0; i <= 5; i++ {
			xx[i] = 0
		}
		return status
	}
	return flags
}

// sweFixstar2Ut implements swe_fixstar2_ut from sweph.c.
func (state *State) sweFixstar2Ut(star []byte, jdUT float64, flags int32, xx []float64, message []byte) int32 {
	var (
		deltat    float64
		retflag   int32
		ephemeris int32 = 0
	)
	flags = state.plausIflag(flags, -1, jdUT, message)
	ephemeris = flags & (flagJPL | flagSwiss | flagMoshier)
	if ephemeris == 0 {
		ephemeris = 2
		flags |= flagSwiss
	}
	deltat = state.sweDeltatEx(jdUT, flags, message)
	retflag = state.sweFixstar2(star, jdUT+deltat, flags, xx, message)
	if retflag != -1 && retflag&(1|2|4) != ephemeris {
		deltat = state.sweDeltatEx(jdUT, retflag, nil)
		retflag = state.sweFixstar2(star, jdUT+deltat, flags, xx, nil)
	}
	return retflag
}

// sweFixstar2Mag implements swe_fixstar2_mag from sweph.c.
func (state *State) sweFixstar2Mag(star []byte, mag []float64, message []byte) int32 {
	var (
		sstar    [40 + 1]byte
		status   int32
		stardata fixedStar
	)
found:
	for {
		if message != nil {
			message[0] = byte('\000')
		}
		state.loadAllFixedStars(message)
		status = state.fixstarFormatSearchName(star, sstar[:], message)
		if status == -1 {
			mag[0] = 0
			return status
		}
		if state.positions.data.nFixstarsRecords > 0 && compareText(state.positions.fixstar2MagSlastStarname[:], sstar[:]) == 0 {
			stardata = state.positions.fixstar2MagLastStardata
			break found
		}
		status = state.searchStarInList(sstar[:], &stardata, message)
		if status == -1 {
			mag[0] = 0
			return status
		}
		break
	}
	state.positions.fixstar2MagLastStardata = stardata
	copyText(state.positions.fixstar2MagSlastStarname[:], sstar[:])
	mag[0] = stardata.magnitude
	formatText(star, textBytes("%s,%s"), stardata.name[:], stardata.bayerName[:])
	return 0
}

// sweGetPlanetName implements swe_get_planet_name from sweph.c.
func (state *State) sweGetPlanetName(body int32, s []byte) []byte {
	var (
		i      int32
		status int32
		xp     [6]float64
	)
	state.swiInitSwedIfStart()
	if body == 10000+134340 {
		body = 9
	}
	if body != 0 && body == state.positions.data.iSavedPlanetName {
		copyText(s, state.positions.data.savedPlanetName[:])
		return s
	}
	switch body {
	case 0:
		copyText(s, textBytes("Sun"))
	case 1:
		copyText(s, textBytes("Moon"))
	case 2:
		copyText(s, textBytes("Mercury"))
	case 3:
		copyText(s, textBytes("Venus"))
	case 4:
		copyText(s, textBytes("Mars"))
	case 5:
		copyText(s, textBytes("Jupiter"))
	case 6:
		copyText(s, textBytes("Saturn"))
	case 7:
		copyText(s, textBytes("Uranus"))
	case 8:
		copyText(s, textBytes("Neptune"))
	case 9:
		copyText(s, textBytes("Pluto"))
	case 10:
		copyText(s, textBytes("mean Node"))
	case 11:
		copyText(s, textBytes("true Node"))
	case 12:
		copyText(s, textBytes("mean Apogee"))
	case 13:
		copyText(s, textBytes("osc. Apogee"))
	case 21:
		copyText(s, textBytes("intp. Apogee"))
	case 22:
		copyText(s, textBytes("intp. Perigee"))
	case 14:
		copyText(s, textBytes("Earth"))
	case 15:
		fallthrough
	case 10000 + 2060:
		copyText(s, textBytes("Chiron"))
	case 16:
		fallthrough
	case 10000 + 5145:
		copyText(s, textBytes("Pholus"))
	case 17:
		fallthrough
	case 10000 + 1:
		copyText(s, textBytes("Ceres"))
	case 18:
		fallthrough
	case 10000 + 2:
		copyText(s, textBytes("Pallas"))
	case 19:
		fallthrough
	case 10000 + 3:
		copyText(s, textBytes("Juno"))
	case 20:
		fallthrough
	case 10000 + 4:
		copyText(s, textBytes("Vesta"))
	default:
		if body >= 40 && body <= 999 {
			state.swiGetFictName(body-40, s)
			break
		}
		if body > 9000 || body > 10000 {
			if body == state.positions.data.files[3].bodies[0] {
				copyText(s, state.positions.data.files[3].asteroidName[:])
			} else {
				status = state.sweph(julianDayJ2000, body, 3, 0, nil, 0, xp[:], nil)
				if status != -1 && status != -2 {
					copyText(s, state.positions.data.files[3].asteroidName[:])
				} else if body > 10000 {
					formatText(s, textBytes("%d: not found (asteroid)"), body-10000)
				} else {
					formatText(s, textBytes("%d: not found (planetary moon)"), body)
				}
			}
			if body > 10000 && (int32(s[0]) == '?' || isDigit(int32(s[1])) != 0) {
				var (
					bodyIndex int32 = body - 10000
					fileBody  int32 = 0
					fp        *textFile
					si        [256]byte
					sp        []byte
					sp2       []byte
				)
				if fp = state.swiFopen(-1, textBytes("seasnam.txt"), state.positions.data.ephemerisPath[:], nil); fp != nil {
					for {
						if !(bodyIndex != fileBody) {
							break
						}
						sp = readLine(si[:], 256, fp)
						if !(sp != nil) {
							break
						}
						for int32(sp[0]) == ' ' || int32(sp[0]) == '\t' || int32(sp[0]) == '(' || int32(sp[0]) == '[' || int32(sp[0]) == '{' {
							sp = sp[1:]
						}
						if int32(sp[0]) == '#' || int32(sp[0]) == '\r' || int32(sp[0]) == '\n' || int32(sp[0]) == '\000' {
							continue
						}
						fileBody = parseInt(sp)
						if bodyIndex != fileBody {
							continue
						}
						sp = findAnyText(sp, textBytes(" \t"))
						if sp == nil {
							continue
						}
						for int32(sp[0]) == ' ' || int32(sp[0]) == '\t' {
							sp = sp[1:]
						}
						sp2 = findAnyText(sp, textBytes("#\r\n"))
						if sp2 != nil {
							sp2[0] = byte('\000')
						}
						if int32(sp[0]) == '\000' {
							continue
						}
						state.swiRightTrim(sp)
						copyText(s, sp)
					}
					closeTextFile(fp)
				}
			}
		} else {
			i = body
			formatText(s, textBytes("%d"), i)
		}
	}
	if textLength(s) < 80 {
		state.positions.data.iSavedPlanetName = body
		copyText(state.positions.data.savedPlanetName[:], s)
	}
	return s
}

// sweGetAyanamsaName implements swe_get_ayanamsa_name from sweph.c.
func (state *State) sweGetAyanamsaName(isidmode int32) []byte {
	isidmode %= 256
	if isidmode < 47 {
		return positionsAyanamsaName[isidmode]
	}
	return nil
}

// sweSetTopo implements swe_set_topo from sweph.c.
//
//	set geographic position and altitude of observer
func (state *State) sweSetTopo(geolon float64, geolat float64, geoalt float64) {
	state.swiInitSwedIfStart()
	if state.positions.data.geoposIsSet == 1 && state.positions.data.observer.longitude == geolon && state.positions.data.observer.latitude == geolat && state.positions.data.observer.altitude == geoalt {
		return
	}
	state.positions.data.observer.longitude = geolon
	state.positions.data.observer.latitude = geolat
	state.positions.data.observer.altitude = geoalt
	state.positions.data.geoposIsSet = 1
	state.positions.data.observer.evaluationTime = 0
	state.swiForceAppPosEtc()
}

// swiForceAppPosEtc implements swi_force_app_pos_etc from sweph.c.
//
//	to force new calculation of light-time etc.
func (state *State) swiForceAppPosEtc() {
	var i int32
	for i = 0; i < 18; i++ {
		state.positions.data.planets[i].cachedFlags = -1
	}
	for i = 0; i < 6; i++ {
		state.positions.data.lunarNodes[i].cachedFlags = -1
	}
	for i = 0; i <= 23; i++ {
		state.positions.data.savedPositions[i].epoch = 0
		state.positions.data.savedPositions[i].flags = -1
	}
}

// swiGetObserver implements swi_get_observer from sweph.c.
func (state *State) swiGetObserver(jd float64, flags int32, saveCache int32, observerVector []float64, message []byte) int32 {
	var (
		i     int32
		sidt  float64
		delt  float64
		jdUT  float64
		eps   float64
		nut   float64
		nutlo [2]float64
		f     float64 = 0.003352819697896193
		re    float64 = 6378136.6
		cosfi float64
		sinfi float64
		cc    float64
		ss    float64
		cosl  float64
		sinl  float64
		h     float64
	)
	if !(state.positions.data.geoposIsSet != 0) {
		if message != nil {
			copyText(message, textBytes("geographic position has not been set"))
		}
		return -1
	}
	delt = state.sweDeltatEx(jd, flags, message)
	jdUT = jd - delt
	if state.positions.data.obliquity.epoch == jd && state.positions.data.nutation.epoch == jd {
		eps = state.positions.data.obliquity.angle
		nutlo[1] = state.positions.data.nutation.angles[1]
		nutlo[0] = state.positions.data.nutation.angles[0]
	} else {
		eps = state.swiEpsiln(jd, flags)
		if flags&flagNoNutation == 0 {
			state.swiNutation(jd, flags, nutlo[:])
		}
	}
	if flags&flagNoNutation != 0 {
		nut = 0
	} else {
		eps += nutlo[1]
		nut = nutlo[0]
	}
	sidt = state.sweSidtime0(jdUT, eps*degreesPerRadian, nut*degreesPerRadian)
	sidt *= 15
	cosfi = math.Cos(state.positions.data.observer.latitude * radiansPerDegree)
	sinfi = math.Sin(state.positions.data.observer.latitude * radiansPerDegree)
	cc = 1 / math.Sqrt(cosfi*cosfi+(1-f)*(1-f)*sinfi*sinfi)
	ss = (1 - f) * (1 - f) * cc
	cosl = math.Cos((state.positions.data.observer.longitude + sidt) * radiansPerDegree)
	sinl = math.Sin((state.positions.data.observer.longitude + sidt) * radiansPerDegree)
	h = state.positions.data.observer.altitude
	observerVector[0] = (re*cc + h) * cosfi * cosl
	observerVector[1] = (re*cc + h) * cosfi * sinl
	observerVector[2] = (re*ss + h) * sinfi
	state.swiCartpol(observerVector, observerVector)
	observerVector[3] = 6.3003874867487992
	observerVector[5] = 0
	observerVector[4] = observerVector[5]
	state.swiPolcartSp(observerVector, observerVector)
	for i = 0; i <= 5; i++ {
		observerVector[i] /= 1.49597870700e+11
	}
	if flags&flagNoNutation == 0 {
		state.swiCoortrf2(observerVector, observerVector, -state.positions.data.nutation.sine, state.positions.data.nutation.cosine)
		state.swiCoortrf2(observerVector[3:], observerVector[3:], -state.positions.data.nutation.sine, state.positions.data.nutation.cosine)
		state.swiNutate(observerVector, flags|(flagSpeed), 1)
	}
	state.swiPrecess(observerVector, jd, flags, 1)
	state.swiPrecessSpeed(observerVector, jd, flags, 1)
	if saveCache != 0 {
		for i = 0; i <= 5; i++ {
			state.positions.data.observer.stateVector[i] = observerVector[i]
		}
		state.positions.data.observer.evaluationTime = jd
		state.positions.data.observer.jdUT = jdUT
	}
	return 0
}

// sweTimeEqu implements swe_time_equ from sweph.c.
//
//	Equation of Time
//
// The function returns the difference between
// local apparent and local mean time in days.
// E = LAT - LMT
// Input variable tjd is UT.
func (state *State) sweTimeEqu(jdUT float64, E []float64, message []byte) int32 {
	var (
		retval int32
		t      float64
		dt     float64
		x      [6]float64
		sidt   float64 = state.sweSidtime(jdUT)
		flags  int32   = 2 * 1024
	)
	flags = state.plausIflag(flags, -1, jdUT, message)
	if state.swiInitSwedIfStart() == 1 && flags&flagMoshier == 0 && message != nil {
		copyText(message, textBytes("Please call swe_set_ephe_path() or swe_set_jplfile() before calling swe_time_equ(), swe_lmt_to_lat() or swe_lat_to_lmt()"))
	}
	if state.positions.data.jplFileIsOpen != 0 {
		flags |= flagJPL
	}
	t = jdUT + 0.5
	dt = t - math.Floor(t)
	sidt -= dt * 24
	sidt *= 15
	if retval = state.sweCalcUt(jdUT, 0, flags, x[:], message); retval == -1 {
		E[0] = 0
		return -1
	}
	dt = state.sweDegnorm(sidt - x[0] - 180)
	if dt > 180 {
		dt -= 360
	}
	dt *= 4
	E[0] = dt / 1440.0
	return 0
}

// sweLmtToLat implements swe_lmt_to_lat from sweph.c.
func (state *State) sweLmtToLat(tjdLmt float64, longitude float64, tjdLat []float64, message []byte) int32 {
	var (
		retval  int32
		E       [1]float64
		tjdLmt0 float64
	)
	tjdLmt0 = tjdLmt - longitude/360.0
	retval = state.sweTimeEqu(tjdLmt0, E[:], message)
	tjdLat[0] = tjdLmt + E[0]
	return retval
}

// sweLatToLmt implements swe_lat_to_lmt from sweph.c.
func (state *State) sweLatToLmt(tjdLat float64, longitude float64, tjdLmt []float64, message []byte) int32 {
	var (
		retval  int32
		E       [1]float64
		tjdLmt0 float64
	)
	tjdLmt0 = tjdLat - longitude/360.0
	retval = state.sweTimeEqu(tjdLmt0, E[:], message)
	retval = state.sweTimeEqu(tjdLmt0-E[0], E[:], message)
	retval = state.sweTimeEqu(tjdLmt0-E[0], E[:], message)
	tjdLmt[0] = tjdLat - E[0]
	return retval
}

// openJplFile implements open_jpl_file from sweph.c.
func (state *State) openJplFile(ss []float64, fname []byte, fpath []byte, message []byte) int32 {
	var (
		status int32
		serr2  [256]byte
	)
	status = state.swiOpenJplFile(ss, fname, fpath, message)
	if status != 0 && findText(fname, textBytes("de431.eph")) != nil && message != nil {
		status = state.swiOpenJplFile(ss, textBytes("de406.eph"), fpath, serr2[:])
		if status == 0 {
			copyText(state.positions.data.jplFilename[:], textBytes("de406.eph"))
			if message != nil {
				copyText(serr2[:], textBytes("Error with JPL ephemeris file "))
				if textLength(serr2[:])+textLength(textBytes("de431.eph")) < 256 {
					appendText(serr2[:], textBytes("de431.eph"))
				}
				if textLength(serr2[:])+textLength(message)+2 < 256 {
					formatText(serr2[textLength(serr2[:]):], textBytes(": %s"), message)
				}
				if textLength(serr2[:])+17 < 256 {
					appendText(serr2[:], textBytes(". Defaulting to "))
				}
				if textLength(serr2[:])+textLength(textBytes("de406.eph")) < 256 {
					appendText(serr2[:], textBytes("de406.eph"))
				}
				copyText(message, serr2[:])
			}
		}
	}
	if status == 0 {
		state.positions.data.jplDENumber = state.swiGetJplDenum()
		state.positions.data.jplFileIsOpen = 1
		state.swiSetTidAcc(0, 0, state.positions.data.jplDENumber, message)
	}
	return status
}

// swiFixstarLoadRecord implements swi_fixstar_load_record from sweph.c.
func (state *State) swiFixstarLoadRecord(star []byte, srecord []byte, sname []byte, sbayer []byte, dparams []float64, message []byte) int32 {
	var (
		s        [256 + 20]byte
		sp       []byte
		sp2      []byte
		sstar    [40 + 1]byte
		fstar    [40 + 1]byte
		i        int32
		starNr   int32 = 0
		line     int32 = 0
		fline    int32 = 0
		status   int32 = 0
		isBayer  int32 = 0
		cmplen   uint64
		slen     uint64
		stardata fixedStar
	)
	status = state.fixstarFormatSearchName(star, sstar[:], message)
	if status == -1 {
		return -1
	}
	if int32(sstar[0]) == ',' {
		isBayer = 1
	} else if isDigit(int32(sstar[0])) != 0 {
		starNr = parseInt(sstar[:])
	} else if sp = findByte(sstar[:], ','); sp != nil {
		sp[0] = byte('\000')
	}
	cmplen = textLength(sstar[:])
	if state.positions.data.starFile == nil {
		if state.positions.data.starFile = state.swiFopen(4, textBytes("sefstars.txt"), state.positions.data.ephemerisPath[:], message); state.positions.data.starFile == nil {
			state.positions.data.isOldStarfile = 1
			if state.positions.data.starFile = state.swiFopen(4, textBytes("fixstars.cat"), state.positions.data.ephemerisPath[:], nil); state.positions.data.starFile == nil {
				state.positions.data.isOldStarfile = 0
				return -1
			}
		}
	}
	rewindFile(state.positions.data.starFile)
	for readLine(s[:], 256, state.positions.data.starFile) != nil {
		fline++
		if int32(s[0]) == '#' {
			continue
		}
		line++
		if starNr == line {
			copyText(srecord, s[:])
			status = state.fixstarCutString(srecord, star, &stardata, message)
			if status == -1 {
				return -1
			}
			if dparams != nil {
				dparams[0] = stardata.epoch
				dparams[1] = stardata.rightAscension
				dparams[2] = stardata.declination
				dparams[3] = stardata.properMotionRA
				dparams[4] = stardata.properMotionDec
				dparams[5] = stardata.radialVelocity
				dparams[6] = stardata.parallax
				dparams[7] = stardata.magnitude
			}
			return 0
		} else if starNr > 0 {
			continue
		}
		if sp = findByte(s[:], ','); sp == nil {
			if message != nil {
				formatText(message, textBytes("star file %s damaged at line %d"), textBytes("sefstars.txt"), fline)
			}
			return -1
		}
		if isBayer != 0 {
			if compareTextN(sp, sstar[:], cmplen) == 0 {
				copyText(srecord, s[:])
				status = state.fixstarCutString(srecord, star, &stardata, message)
				if status == -1 {
					return -1
				}
				if dparams != nil {
					dparams[0] = stardata.epoch
					dparams[1] = stardata.rightAscension
					dparams[2] = stardata.declination
					dparams[3] = stardata.properMotionRA
					dparams[4] = stardata.properMotionDec
					dparams[5] = stardata.radialVelocity
					dparams[6] = stardata.parallax
					dparams[7] = stardata.magnitude
				}
				return 0
			} else {
				continue
			}
		}
		sp[0] = byte('\000')
		slen = textLength(s[:])
		if slen > 256 {
			slen = 256
		}
		copy(fstar[:][:slen], s[:][:slen])
		fstar[slen] = byte('\000')
		sp[0] = byte(',')
		for {
			sp = findByte(fstar[:], ' ')
			if !(sp != nil) {
				break
			}
			state.swiStrcpy(sp, sp[1:])
		}
		i = int32(textLength(fstar[:]))
		if i < int32(cmplen) {
			continue
		}
		for sp2 = fstar[:]; int32(sp2[0]) != '\000'; sp2 = sp2[1:] {
			sp2[0] = byte(lowerByte(int32(sp2[0])))
		}
		if compareTextN(fstar[:], sstar[:], cmplen) == 0 {
			copyText(srecord, s[:])
			status = state.fixstarCutString(srecord, star, &stardata, message)
			if status == -1 {
				return -1
			}
			if dparams != nil {
				dparams[0] = stardata.epoch
				dparams[1] = stardata.rightAscension
				dparams[2] = stardata.declination
				dparams[3] = stardata.properMotionRA
				dparams[4] = stardata.properMotionDec
				dparams[5] = stardata.radialVelocity
				dparams[6] = stardata.parallax
				dparams[7] = stardata.magnitude
			}
			return 0
		}
	}
	if message != nil {
		formatText(message, textBytes("star  not found"))
		if textLength(message)+textLength(star) < 256 {
			formatText(message, textBytes("star %s not found"), star)
		}
		return -1
	}
	copyText(srecord, s[:])
	status = state.fixstarCutString(srecord, star, &stardata, message)
	if status == -1 {
		return -1
	}
	if dparams != nil {
		dparams[0] = stardata.epoch
		dparams[1] = stardata.rightAscension
		dparams[2] = stardata.declination
		dparams[3] = stardata.properMotionRA
		dparams[4] = stardata.properMotionDec
		dparams[5] = stardata.radialVelocity
		dparams[6] = stardata.parallax
		dparams[7] = stardata.magnitude
	}
	return 0
}

// swiFixstarCalcFromRecord implements swi_fixstar_calc_from_record from sweph.c.
//
//	function calculates a fixstar from a record from sefstars.txt
//
// input:
// char *srecord     fixed star data record from sefstars.txt
// double tjd        julian daynumber
// int32 iflag       SEFLG_ specifications
// output:
// char *star        star name, Bayer designation
// double xx[6]      position and speed
// char *serr        error return string
func (state *State) swiFixstarCalcFromRecord(srecord []byte, jd float64, flags int32, star []byte, xx []float64, message []byte) int32 {
	var (
		i              int32
		status         int32 = 0
		epoch          float64
		radv           float64
		parall         float64
		raPm           float64
		dePm           float64
		ra             float64
		de             float64
		t              float64
		stardata       fixedStar
		daya           [1]float64
		rdist          float64
		x              [6]float64
		xxsv           [6]float64
		observerVector [6]float64
		xobsDt         [6]float64
		xpo            []float64 = nil
		xpoDt          []float64 = nil
		dt             float64   = 1.0000000000000001e-05
		ephemeris      int32
		savedFlags     int32
		obliquityData  *obliquity = &state.positions.data.obliquityJ2000
	)
	savedFlags = flags
	flags |= flagSpeed
	if message != nil {
		message[0] = byte('\000')
	}
	flags = state.plausIflag(flags, -1, jd, message)
	ephemeris = flags & (flagJPL | flagSwiss | flagMoshier)
	if state.swiInitSwedIfStart() == 1 && !(ephemeris&4 != 0) && message != nil {
		copyText(message, textBytes("Please call swe_set_ephe_path() or swe_set_jplfile() before calling swe_fixstar() or swe_fixstar_ut()"))
	}
	if state.positions.data.lastEpheflag != ephemeris {
		state.freePlanets()
		if state.positions.data.jplFileIsOpen != 0 {
			state.swiCloseJplFile()
			state.positions.data.jplFileIsOpen = 0
		}
		for i = 0; i < 7; i++ {
			if state.positions.data.files[i].handle != nil {
				closeTextFile(state.positions.data.files[i].handle)
			}
			state.positions.data.files[i] = ephemerisFile{}
		}
		state.positions.data.lastEpheflag = ephemeris
	}
	if flags&flagSidereal != 0 && !(state.positions.data.ayanaIsSet != 0) {
		state.sweSetSidMode(0, 0, 0)
	}
	state.swiCheckEcliptic(jd, flags)
	state.swiCheckNutation(jd, flags)
	status = state.fixstarCutString(srecord, star, &stardata, message)
	if status == -1 {
		return -1
	}
	epoch = stardata.epoch
	raPm = stardata.properMotionRA
	dePm = stardata.properMotionDec
	radv = stardata.radialVelocity
	parall = stardata.parallax
	ra = stardata.rightAscension
	de = stardata.declination
	if epoch == 1950 {
		t = jd - 2433282.42345905
	} else {
		t = jd - julianDayJ2000
	}
	x[0] = ra
	x[1] = de
	x[2] = 1
	if parall == 0 {
		rdist = 1000000000
	} else {
		rdist = 1.0 / (parall * degreesPerRadian * 3600) * 206264.8062471
	}
	x[2] = rdist
	x[3] = raPm / daysPerCentury
	x[4] = dePm / daysPerCentury
	x[5] = radv / daysPerCentury
	state.swiPolcartSp(x[:], x[:])
	if epoch == 1950 {
		state.swiFK4FK5(x[:], 2433282.42345905)
		state.swiPrecess(x[:], 2433282.42345905, 0, 1)
		state.swiPrecess(x[3:], 2433282.42345905, 0, 1)
	}
	if epoch != 0 {
		state.swiIcrs2fk5(x[:], flags, 1)
		if state.swiGetDenum(0, flags) >= 403 {
			state.swiBias(x[:], julianDayJ2000, 256, 0)
		}
	}
	if flags&flagBarycentric == 0 && (flags&flagHeliocentric == 0 || flags&flagMoshier == 0) {
		if status = state.mainPlanetBary(jd-dt, 0, ephemeris, flags, 0, state.positions.fixstarCalcFromRecordXearthDt[:], state.positions.fixstarCalcFromRecordXearthDt[:], state.positions.fixstarCalcFromRecordXsunDt[:], nil, message); status != 0 {
			return -1
		}
		if status = state.mainPlanetBary(jd, 0, ephemeris, flags, 1, state.positions.fixstarCalcFromRecordXearth[:], state.positions.fixstarCalcFromRecordXearth[:], state.positions.fixstarCalcFromRecordXsun[:], nil, message); status != 0 {
			return -1
		}
	}
	if flags&flagTopocentric != 0 {
		if state.swiGetObserver(jd-dt, flags|(flagNoNutation), 0, xobsDt[:], message) != 0 {
			return -1
		}
		if state.swiGetObserver(jd, flags|(flagNoNutation), 0, observerVector[:], message) != 0 {
			return -1
		}
		for i = 0; i <= 5; i++ {
			observerVector[i] = observerVector[i] + state.positions.fixstarCalcFromRecordXearth[i]
			xobsDt[i] = xobsDt[i] + state.positions.fixstarCalcFromRecordXearthDt[i]
		}
	} else if flags&flagBarycentric == 0 && (flags&flagHeliocentric == 0 || flags&flagMoshier == 0) {
		for i = 0; i <= 5; i++ {
			observerVector[i] = state.positions.fixstarCalcFromRecordXearth[i]
			xobsDt[i] = state.positions.fixstarCalcFromRecordXearthDt[i]
		}
	}
	if flags&flagHeliocentric != 0 && flags&flagMoshier != 0 {
		xpo = nil
		xpoDt = nil
	} else if flags&flagHeliocentric != 0 {
		xpo = state.positions.fixstarCalcFromRecordXsun[:]
		xpoDt = state.positions.fixstarCalcFromRecordXsunDt[:]
	} else if flags&flagBarycentric != 0 {
		xpo = nil
		xpoDt = nil
	} else {
		xpo = observerVector[:]
		xpoDt = xobsDt[:]
	}
	if xpo == nil {
		for i = 0; i <= 2; i++ {
			x[i] += t * x[i+3]
		}
	} else {
		for i = 0; i <= 2; i++ {
			x[i] += t * x[i+3]
			x[i] -= xpo[i]
			x[i+3] -= xpo[i+3]
		}
	}
	if flags&flagTruePosition == 0 && flags&flagNoDeflection == 0 {
		state.swiDeflectLight(x[:], 0, flags&flagSpeed)
	}
	if flags&flagTruePosition == 0 && flags&flagNoAberration == 0 {
		state.swiAberrLightEx(x[:], xpo, xpoDt, dt, flags&flagSpeed)
	}
	if flags&flagICRS == 0 && (state.swiGetDenum(0, flags) >= 403 || flags&flagBarycentric != 0) {
		state.swiBias(x[:], jd, flags, 0)
	}
	for i = 0; i <= 5; i++ {
		xxsv[i] = x[i]
	}
	if flags&flagJ2000 == 0 {
		state.swiPrecess(x[:], jd, flags, -1)
		if flags&flagSpeed != 0 {
			state.swiPrecessSpeed(x[:], jd, flags, -1)
		}
		obliquityData = &state.positions.data.obliquity
	} else {
		obliquityData = &state.positions.data.obliquityJ2000
	}
	if flags&flagNoNutation == 0 {
		state.swiNutate(x[:], flags, 0)
	}
	if 0 != 0 {
		var r float64 = math.Sqrt(x[0]*x[0] + x[1]*x[1] + x[2]*x[2])
		printText(textBytes("%.17f %.17f %f\n"), x[0]/r, x[1]/r, x[2]/r)
	}
	if flags&flagEquatorial == 0 {
		state.swiCoortrf2(x[:], x[:], obliquityData.sine, obliquityData.cosine)
		if flags&flagSpeed != 0 {
			state.swiCoortrf2(x[3:], x[3:], obliquityData.sine, obliquityData.cosine)
		}
		if flags&flagNoNutation == 0 {
			state.swiCoortrf2(x[:], x[:], state.positions.data.nutation.sine, state.positions.data.nutation.cosine)
			if flags&flagSpeed != 0 {
				state.swiCoortrf2(x[3:], x[3:], state.positions.data.nutation.sine, state.positions.data.nutation.cosine)
			}
		}
	}
	if flags&flagSidereal != 0 {
		if state.positions.data.sidereal.mode&256 != 0 {
			if state.swiTropRa2sidLon(xxsv[:], x[:], xxsv[:], flags) != 0 {
				return -1
			}
			if flags&flagEquatorial != 0 {
				for i = 0; i <= 5; i++ {
					x[i] = xxsv[i]
				}
			}
		} else if state.positions.data.sidereal.mode&512 != 0 {
			if state.swiTropRa2sidLonSosy(xxsv[:], x[:], flags) != 0 {
				return -1
			}
			if flags&flagEquatorial != 0 {
				for i = 0; i <= 5; i++ {
					x[i] = xxsv[i]
				}
			}
		} else {
			state.swiCartpolSp(x[:], x[:])
			if state.swiGetAyanamsaEx(jd, flags, daya[:], message) == -1 {
				return -1
			}
			x[0] -= daya[0] * radiansPerDegree
			state.swiPolcartSp(x[:], x[:])
		}
	}
	if flags&flagCartesian == 0 {
		state.swiCartpolSp(x[:], x[:])
	}
	if flags&flagRadians == 0 && flags&flagCartesian == 0 {
		for i = 0; i < 2; i++ {
			x[i] *= degreesPerRadian
			x[i+3] *= degreesPerRadian
		}
	}
	for i = 0; i <= 5; i++ {
		xx[i] = x[i]
	}
	if savedFlags&flagSpeed == 0 {
		flags = flags & ^256
		for i = 3; i <= 5; i++ {
			xx[i] = 0
		}
	}
	if savedFlags&(flagJPL|flagSwiss|flagMoshier) == 0 {
		flags = flags & ^2
	}
	return flags
}

// sweFixstar implements swe_fixstar from sweph.c.
//
//	Citra
func (state *State) sweFixstar(star []byte, jd float64, flags int32, xx []float64, message []byte) int32 {
	var (
		i       int32
		sstar   [40 + 1]byte
		srecord [256 + 20]byte
		sp      []byte
		status  int32
	)
found:
	for {
		if message != nil {
			message[0] = byte('\000')
		}
		status = state.fixstarFormatSearchName(star, sstar[:], message)
		if status == -1 {
			i = 0
			for i = 0; i <= 5; i++ {
				xx[i] = 0
			}
			return status
		}
		if int32(sstar[0]) == ',' {
		} else if isDigit(int32(sstar[0])) != 0 {
		} else if sp = findByte(sstar[:], ','); sp != nil {
			sp[0] = byte('\000')
		}
		if int32(state.positions.fixstarSlastStardata[0]) != '\000' && compareText(state.positions.fixstarSlastStarname[:], sstar[:]) == 0 {
			copyText(srecord[:], state.positions.fixstarSlastStardata[:])
			break found
		}
		if state.getBuiltinStar(star, sstar[:], srecord[:]) != 0 {
			break found
		}
		if status = state.swiFixstarLoadRecord(star, srecord[:], nil, nil, nil, message); status != 0 {
			i = 0
			for i = 0; i <= 5; i++ {
				xx[i] = 0
			}
			return status
		}
		break
	}
	copyText(state.positions.fixstarSlastStardata[:], srecord[:])
	copyText(state.positions.fixstarSlastStarname[:], sstar[:])
	if status = state.swiFixstarCalcFromRecord(srecord[:], jd, flags, star, xx, message); status == -1 {
		i = 0
		for i = 0; i <= 5; i++ {
			xx[i] = 0
		}
		return status
	}
	return flags
}

// sweFixstarUt implements swe_fixstar_ut from sweph.c.
func (state *State) sweFixstarUt(star []byte, jdUT float64, flags int32, xx []float64, message []byte) int32 {
	var (
		deltat    float64
		retflag   int32
		ephemeris int32 = 0
	)
	flags = state.plausIflag(flags, -1, jdUT, message)
	ephemeris = flags & (flagJPL | flagSwiss | flagMoshier)
	if ephemeris == 0 {
		ephemeris = 2
		flags |= flagSwiss
	}
	deltat = state.sweDeltatEx(jdUT, flags, message)
	retflag = state.sweFixstar(star, jdUT+deltat, flags, xx, message)
	if retflag != -1 && retflag&(1|2|4) != ephemeris {
		deltat = state.sweDeltatEx(jdUT, retflag, nil)
		retflag = state.sweFixstar(star, jdUT+deltat, flags, xx, nil)
	}
	return retflag
}

// sweFixstarMag implements swe_fixstar_mag from sweph.c.
func (state *State) sweFixstarMag(star []byte, mag []float64, message []byte) int32 {
	var (
		sstar    [40 + 1]byte
		srecord  [256 + 20]byte
		sp       []byte
		stardata fixedStar
		status   int32
		dparams  [20]float64
	)
found:
	for {
		if message != nil {
			message[0] = byte('\000')
		}
		status = state.fixstarFormatSearchName(star, sstar[:], message)
		if status == -1 {
			mag[0] = 0
			return status
		}
		if int32(sstar[0]) == ',' {
		} else if isDigit(int32(sstar[0])) != 0 {
		} else if sp = findByte(sstar[:], ','); sp != nil {
			sp[0] = byte('\000')
		}
		if int32(state.positions.fixstarMagSlastStardata[0]) != '\000' && compareText(state.positions.fixstarMagSlastStarname[:], sstar[:]) == 0 {
			copyText(srecord[:], state.positions.fixstarMagSlastStardata[:])
			status = state.fixstarCutString(srecord[:], star, &stardata, message)
			if status == -1 {
				mag[0] = 0
				return status
			}
			dparams[7] = stardata.magnitude
			break found
		}
		if status = state.swiFixstarLoadRecord(star, srecord[:], nil, nil, dparams[:], message); status != 0 {
			mag[0] = 0
			return status
		}
		break
	}
	copyText(state.positions.fixstarMagSlastStardata[:], srecord[:])
	copyText(state.positions.fixstarMagSlastStarname[:], sstar[:])
	mag[0] = dparams[7]
	return 0
}

// sweCalcPctr implements swe_calc_pctr from sweph.c.
func (state *State) sweCalcPctr(jd float64, body int32, iplctr int32, flags int32, xxret []float64, message []byte) int32 {
	var (
		t             float64 = 0
		dt            float64
		daya          [2]float64
		dtsaveForDefl float64 = 0
		xx            [6]float64
		xxctr         [6]float64
		xxctr2        [6]float64
		xx0           [6]float64
		xxsv          [24]float64
		xxsp          [6]float64
		dx            [6]float64
		xreturn       [24]float64
		xs            []float64
		i             int32
		j             int32
		niter         int32
		iflag2        int32
		ephemeris     int32
		status        int32
		obliquityData *obliquity
	)
	if body == iplctr {
		if message != nil {
			formatText(message, textBytes("ipl and iplctr (= %d) must not be identical\n"), body)
		}
		return -1
	}
	flags = state.plausIflag(flags, body, jd, message)
	ephemeris = flags & (flagJPL | flagSwiss | flagMoshier)
	state.sweCalc(jd+state.sweDeltatEx(jd, ephemeris, message), -1, flags, xx[:], message)
	flags &= ^(8 | 16*1024)
	iflag2 = ephemeris
	iflag2 |= 16*1024 | 32 | 128*1024 | 16 | 2*1024 | 4*1024 | 256
	iflag2 |= 1024 | 512
	status = state.sweCalc(jd, iplctr, iflag2, xxctr[:], message)
	if status == -1 {
		return -1
	}
	status = state.sweCalc(jd, body, iflag2, xx[:], message)
	if status == -1 {
		return -1
	}
	for i = 0; i <= 5; i++ {
		xx0[i] = xx[i]
	}
	if flags&flagTruePosition == 0 {
		niter = 1
		if flags&flagSpeed != 0 {
			for i = 0; i <= 2; i++ {
				xxsp[i] = xx[i] - xx[i+3]
				xxsv[i] = xxsp[i]
			}
			for j = 0; j <= niter; j++ {
				for i = 0; i <= 2; i++ {
					dx[i] = xxsp[i]
					dx[i] -= xxctr[i] - xxctr[i+3]
				}
				dt = math.Sqrt(dx[0]*dx[0]+dx[1]*dx[1]+dx[2]*dx[2]) * 1.49597870700e+11 / 2.99792458e+8 / secondsPerDay
				for i = 0; i <= 2; i++ {
					xxsp[i] = xxsv[i] - dt*xx0[i+3]
				}
			}
			for i = 0; i <= 2; i++ {
				xxsp[i] = xxsv[i] - xxsp[i]
			}
		}
		for j = 0; j <= niter; j++ {
			for i = 0; i <= 2; i++ {
				dx[i] = xx[i]
				dx[i] -= xxctr[i]
			}
			dt = math.Sqrt(dx[0]*dx[0]+dx[1]*dx[1]+dx[2]*dx[2]) * 1.49597870700e+11 / 2.99792458e+8 / secondsPerDay
			t = jd - dt
			dtsaveForDefl = dt
			for i = 0; i <= 2; i++ {
				xx[i] = xx0[i] - dt*xx0[i+3]
			}
		}
		if flags&flagSpeed != 0 {
			for i = 0; i <= 2; i++ {
				xxsp[i] = xx0[i] - xx[i] - xxsp[i]
			}
		}
		status = state.sweCalc(t, iplctr, iflag2, xxctr2[:], message)
		status = state.sweCalc(t, body, iflag2, xx[:], message)
	}
	if flags&flagHeliocentric == 0 && flags&flagBarycentric == 0 {
		for i = 0; i <= 5; i++ {
			xx[i] -= xxctr[i]
		}
		if flags&flagTruePosition == 0 {
			if flags&flagSpeed != 0 {
				for i = 3; i <= 5; i++ {
					xx[i] -= xxsp[i-3]
				}
			}
		}
	}
	if flags&flagSpeed == 0 {
		for i = 3; i <= 5; i++ {
			xx[i] = 0
		}
	}
	if flags&flagTruePosition == 0 && flags&flagNoDeflection == 0 {
		state.swiDeflectLight(xx[:], dtsaveForDefl, flags)
	}
	if flags&flagTruePosition == 0 && flags&flagNoAberration == 0 {
		state.swiAberrLight(xx[:], xxctr[:], flags)
		if flags&flagSpeed != 0 {
			for i = 3; i <= 5; i++ {
				xx[i] += xxctr[i] - xxctr2[i]
			}
		}
	}
	if flags&flagSpeed == 0 {
		for i = 3; i <= 5; i++ {
			xx[i] = 0
		}
	}
	if flags&flagICRS == 0 && state.swiGetDenum(body, ephemeris) >= 403 {
		state.swiBias(xx[:], t, flags, 0)
	}
	for i = 0; i <= 5; i++ {
		xxsv[i] = xx[i]
	}
	if flags&flagJ2000 == 0 {
		state.swiPrecess(xx[:], jd, flags, -1)
		if flags&flagSpeed != 0 {
			state.swiPrecessSpeed(xx[:], jd, flags, -1)
		}
		obliquityData = &state.positions.data.obliquity
	} else {
		obliquityData = &state.positions.data.obliquityJ2000
	}
	if flags&flagNoNutation == 0 {
		state.swiNutate(xx[:], flags, 0)
	}
	for i = 0; i <= 5; i++ {
		xreturn[18+i] = xx[i]
	}
	state.swiCoortrf2(xx[:], xx[:], obliquityData.sine, obliquityData.cosine)
	if flags&flagSpeed != 0 {
		state.swiCoortrf2(xx[3:], xx[3:], obliquityData.sine, obliquityData.cosine)
	}
	if flags&flagNoNutation == 0 {
		state.swiCoortrf2(xx[:], xx[:], state.positions.data.nutation.sine, state.positions.data.nutation.cosine)
		if flags&flagSpeed != 0 {
			state.swiCoortrf2(xx[3:], xx[3:], state.positions.data.nutation.sine, state.positions.data.nutation.cosine)
		}
	}
	for i = 0; i <= 5; i++ {
		xreturn[6+i] = xx[i]
	}
	if flags&flagSidereal != 0 {
		if state.positions.data.sidereal.mode&256 != 0 {
			if state.swiTropRa2sidLon(xxsv[:], xreturn[6:], xreturn[18:], flags) != 0 {
				return -1
			}
		} else if state.positions.data.sidereal.mode&512 != 0 {
			if state.swiTropRa2sidLonSosy(xxsv[:], xreturn[6:], flags) != 0 {
				return -1
			}
		} else {
			state.swiCartpolSp(xreturn[6:], xreturn[:])
			for i = 0; i < 24; i++ {
				xxsv[i] = xreturn[i]
			}
			if state.swiGetAyanamsaWithSpeed(jd, flags, daya[:], message) == -1 {
				return -1
			}
			for i = 0; i < 24; i++ {
				xreturn[i] = xxsv[i]
			}
			xreturn[0] -= daya[0] * radiansPerDegree
			xreturn[3] -= daya[1] * radiansPerDegree
			state.swiPolcartSp(xreturn[:], xreturn[6:])
		}
	}
	state.swiCartpolSp(xreturn[18:], xreturn[12:])
	state.swiCartpolSp(xreturn[6:], xreturn[:])
	for i = 0; i < 2; i++ {
		xreturn[i] *= degreesPerRadian
		xreturn[i+3] *= degreesPerRadian
		xreturn[i+12] *= degreesPerRadian
		xreturn[i+15] *= degreesPerRadian
	}
	if flags&flagEquatorial != 0 {
		xs = xreturn[12:]
	} else {
		xs = xreturn[:]
	}
	if flags&flagCartesian != 0 {
		xs = xs[6:]
	}
	for i = 0; i < 6; i++ {
		xxret[i] = xs[i]
	}
	if flags&flagSpeed == 0 {
		for i = 3; i < 6; i++ {
			xxret[i] = 0
		}
	}
	if flags&flagRadians != 0 {
		for i = 0; i < 2; i++ {
			xxret[i] *= radiansPerDegree
		}
		if flags&flagSpeed != 0 {
			for i = 3; i < 5; i++ {
				xxret[i] *= radiansPerDegree
			}
		}
	}
	if status == -1 {
		return -1
	}
	return flags
}

// sweGetCurrentFileData implements swe_get_current_file_data from sweph.c.
func (state *State) sweGetCurrentFileData(fileIndex int32, tfstart []float64, tfend []float64, denum []int32) []byte {
	if fileIndex < 0 || fileIndex > 4 {
		return nil
	}
	var pfp *ephemerisFile = &state.positions.data.files[fileIndex]
	if textLength(pfp.name[:]) == 0 {
		return nil
	}
	tfstart[0] = pfp.start
	tfend[0] = pfp.end
	denum[0] = pfp.deNumber
	return pfp.name[:]
}

// sweSolcross implements swe_solcross from sweph.c.
func (state *State) sweSolcross(x2cross float64, jdEt float64, flag int32, message []byte) float64 {
	var (
		x    [6]float64
		xlp  float64
		dist float64
		jd   float64
		body int32 = 0
	)
	flag |= 256
	if state.sweCalc(jdEt, body, flag, x[:], message) < 0 {
		return jdEt - 1
	}
	xlp = 0.98565326908334239
	dist = state.sweDegnorm(x2cross - x[0])
	for jd = jdEt + dist/xlp; ; {
		if state.sweCalc(jd, body, flag, x[:], message) < 0 {
			return jdEt - 1
		}
		dist = state.sweDifdeg2n(x2cross, x[0])
		jd += dist / x[3]
		if math.Abs(dist) < 2.7777777777777776e-07 {
			break
		}
	}
	return jd
}

// sweSolcrossUt implements swe_solcross_ut from sweph.c.
func (state *State) sweSolcrossUt(x2cross float64, jdUt float64, flag int32, message []byte) float64 {
	var (
		x    [6]float64
		xlp  float64
		dist float64
		jd   float64
		body int32 = 0
	)
	flag |= 256
	if state.sweCalcUt(jdUt, body, flag, x[:], message) < 0 {
		return jdUt - 1
	}
	xlp = 0.98565326908334239
	dist = state.sweDegnorm(x2cross - x[0])
	for jd = jdUt + dist/xlp; ; {
		if state.sweCalcUt(jd, body, flag, x[:], message) < 0 {
			return jdUt - 1
		}
		dist = state.sweDifdeg2n(x2cross, x[0])
		jd += dist / x[3]
		if math.Abs(dist) < 2.7777777777777776e-07 {
			break
		}
	}
	return jd
}

// sweMooncross implements swe_mooncross from sweph.c.
func (state *State) sweMooncross(x2cross float64, jdEt float64, flag int32, message []byte) float64 {
	var (
		x    [6]float64
		xlp  float64
		dist float64
		jd   float64
		body int32 = 1
	)
	flag |= 256
	if state.sweCalc(jdEt, body, flag, x[:], message) < 0 {
		return jdEt - 1
	}
	xlp = 13.177159590043924
	dist = state.sweDegnorm(x2cross - x[0])
	for jd = jdEt + dist/xlp; ; {
		if state.sweCalc(jd, body, flag, x[:], message) < 0 {
			return jdEt - 1
		}
		dist = state.sweDifdeg2n(x2cross, x[0])
		jd += dist / x[3]
		if math.Abs(dist) < 2.7777777777777776e-07 {
			break
		}
	}
	return jd
}

// sweMooncrossUt implements swe_mooncross_ut from sweph.c.
func (state *State) sweMooncrossUt(x2cross float64, jdUt float64, flag int32, message []byte) float64 {
	var (
		x    [6]float64
		xlp  float64
		dist float64
		jd   float64
		body int32 = 1
	)
	flag |= 256
	if state.sweCalcUt(jdUt, body, flag, x[:], message) < 0 {
		return jdUt - 1
	}
	xlp = 13.177159590043924
	dist = state.sweDegnorm(x2cross - x[0])
	for jd = jdUt + dist/xlp; ; {
		if state.sweCalcUt(jd, body, flag, x[:], message) < 0 {
			return jdUt - 1
		}
		dist = state.sweDifdeg2n(x2cross, x[0])
		jd += dist / x[3]
		if math.Abs(dist) < 2.7777777777777776e-07 {
			break
		}
	}
	return jd
}

// sweMooncrossNode implements swe_mooncross_node from sweph.c.
func (state *State) sweMooncrossNode(jdEt float64, flag int32, xlon []float64, xla []float64, message []byte) float64 {
	var (
		x    [6]float64
		xlat float64
		dist float64
		jd   float64
		body int32 = 1
	)
	flag |= 256
	if state.sweCalc(jdEt, body, flag, x[:], message) < 0 {
		return jdEt - 1
	}
	xlat = x[1]
	for jd = jdEt + 1; ; {
		if state.sweCalc(jd, body, flag, x[:], message) < 0 {
			return jdEt - 1
		}
		if x[1] >= 0 && xlat < 0 || x[1] < 0 && xlat > 0 {
			break
		}
		jd += 1
	}
	for dist = x[1]; ; {
		jd -= dist / x[4]
		if state.sweCalc(jd, body, flag, x[:], message) < 0 {
			return jdEt - 1
		}
		dist = x[1]
		if math.Abs(dist) < 2.7777777777777776e-07 {
			xlon[0] = x[0]
			xla[0] = x[1]
			break
		}
	}
	return jd
}

// sweMooncrossNodeUt implements swe_mooncross_node_ut from sweph.c.
func (state *State) sweMooncrossNodeUt(jdUt float64, flag int32, xlon []float64, xla []float64, message []byte) float64 {
	var (
		x    [6]float64
		xlat float64
		dist float64
		jd   float64
		body int32 = 1
	)
	flag |= 256
	if state.sweCalcUt(jdUt, body, flag, x[:], message) < 0 {
		return jdUt - 1
	}
	xlat = x[1]
	for jd = jdUt + 1; ; {
		if state.sweCalcUt(jd, body, flag, x[:], message) < 0 {
			return jdUt - 1
		}
		if x[1] >= 0 && xlat < 0 || x[1] < 0 && xlat > 0 {
			break
		}
		jd += 1
	}
	for dist = x[1]; ; {
		jd -= dist / x[4]
		if state.sweCalcUt(jd, body, flag, x[:], message) < 0 {
			return jdUt - 1
		}
		dist = x[1]
		if math.Abs(dist) < 2.7777777777777776e-07 {
			xlon[0] = x[0]
			xla[0] = x[1]
			break
		}
	}
	return jd
}

// sweHelioCross implements swe_helio_cross from sweph.c.
func (state *State) sweHelioCross(body int32, x2cross float64, jdEt float64, flags int32, dir int32, jdCross []float64, message []byte) int32 {
	var (
		x    [6]float64
		xlp  float64
		dist float64
		jd   float64
		flag int32 = flags | (flagSpeed) | 8
	)
	if body == 0 || body == 1 || body >= 10 && body <= 13 || body >= 21 && body < 23 {
		var snam [256]byte
		state.sweGetPlanetName(body, snam[:])
		if message != nil {
			formatText(message, textBytes("swe_helio_cross: not possible for object %d = %s"), body, snam[:])
		}
		return -1
	}
	if state.sweCalc(jdEt, body, flag, x[:], message) < 0 {
		return -1
	}
	xlp = x[3]
	if body == 15 {
		xlp = 0.01971
	}
	dist = state.sweDegnorm(x2cross - x[0])
	if dir >= 0 {
		jd = jdEt + dist/xlp
	} else {
		dist = 360.0 - dist
		jd = jdEt - dist/xlp
	}
	for {
		if state.sweCalc(jd, body, flag, x[:], message) < 0 {
			return -1
		}
		dist = state.sweDifdeg2n(x2cross, x[0])
		jd += dist / x[3]
		if math.Abs(dist) < 2.7777777777777776e-07 {
			break
		}
	}
	jdCross[0] = jd
	return 0
}

// sweHelioCrossUt implements swe_helio_cross_ut from sweph.c.
func (state *State) sweHelioCrossUt(body int32, x2cross float64, jdUt float64, flags int32, dir int32, jdCross []float64, message []byte) int32 {
	var (
		x    [6]float64
		xlp  float64
		dist float64
		jd   float64
		flag int32 = flags | (flagSpeed) | 8
	)
	if body == 0 || body == 1 || body >= 10 && body <= 13 || body >= 21 && body < 23 {
		var snam [256]byte
		state.sweGetPlanetName(body, snam[:])
		if message != nil {
			formatText(message, textBytes("swe_helio_cross: not possible for object %d = %s"), body, snam[:])
		}
		return -1
	}
	if state.sweCalcUt(jdUt, body, flag, x[:], message) < 0 {
		return -1
	}
	xlp = x[3]
	if body == 15 {
		xlp = 0.01971
	}
	dist = state.sweDegnorm(x2cross - x[0])
	if dir >= 0 {
		jd = jdUt + dist/xlp
	} else {
		dist = 360.0 - dist
		jd = jdUt - dist/xlp
	}
	for {
		if state.sweCalcUt(jd, body, flag, x[:], message) < 0 {
			return -1
		}
		dist = state.sweDifdeg2n(x2cross, x[0])
		jd += dist / x[3]
		if math.Abs(dist) < 2.7777777777777776e-07 {
			break
		}
	}
	jdCross[0] = jd
	return 0
}
