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

// fsizer implements fsizer from swejpl.c.
func (state *State) fsizer(message []byte) int32 {
	var (
		constantCount      [1]int32
		earthMoonMassRatio [1]float64
		deNumber           [1]int32
		astronomicalUnit   [1]float64
		epochRange         [3]float64
		i                  int32
		kmx                int32
		khi                int32
		nd                 int32
		recordSize         int32
		lpt                [3]int32
		ttl                [6 * 14 * 3]byte
		readCount          uint64
	)
	if state.jpl.file.handle = state.swiFopen(0, state.jpl.file.name, state.jpl.file.path, message); state.jpl.file.handle == nil {
		return -2
	}
	readCount = readBinary(ttl[:], 1, 252, state.jpl.file.handle)
	if readCount != 252 {
		return -2
	}
	readCount = readBinary(state.jpl.file.constantNames[:], 1, 6*400, state.jpl.file.handle)
	if readCount != uint64(6*400) {
		return -2
	}
	readCount = readBinary(epochRange[:], 8, 3, state.jpl.file.handle)
	if readCount != 3 {
		return -2
	}
	if epochRange[2] < 1 || epochRange[2] > 200 {
		state.jpl.file.swapBytes = 1
	} else {
		state.jpl.file.swapBytes = 0
	}
	for i = 0; i < 3; i++ {
		state.jpl.file.epochRange[i] = epochRange[i]
	}
	if state.jpl.file.swapBytes != 0 {
		state.reorder(state.jpl.file.epochRange[:], 8, 3)
	}
	if state.jpl.file.epochRange[0] < float64(-5583942) || state.jpl.file.epochRange[1] > 9025909 || state.jpl.file.epochRange[2] < 1 || state.jpl.file.epochRange[2] > 200 {
		if message != nil {
			copyText(message, textBytes("alleged ephemeris file has invalid format."))
			if textLength(message)+textLength(state.jpl.file.name)+3 < 256 {
				formatText(message, textBytes("alleged ephemeris file (%s) has invalid format."), state.jpl.file.name)
			}
		}
		return -2
	}
	readCount = readBinary(constantCount[:], 4, 1, state.jpl.file.handle)
	if readCount != 1 {
		return -2
	}
	if state.jpl.file.swapBytes != 0 {
		state.reorder(constantCount[:], 4, 1)
	}
	readCount = readBinary(astronomicalUnit[:], 8, 1, state.jpl.file.handle)
	if readCount != 1 {
		return -2
	}
	if state.jpl.file.swapBytes != 0 {
		state.reorder(astronomicalUnit[:], 8, 1)
	}
	readCount = readBinary(earthMoonMassRatio[:], 8, 1, state.jpl.file.handle)
	if readCount != 1 {
		return -2
	}
	if state.jpl.file.swapBytes != 0 {
		state.reorder(earthMoonMassRatio[:], 8, 1)
	}
	readCount = readBinary(state.jpl.file.coefficientIndex[:], 4, 36, state.jpl.file.handle)
	if readCount != 36 {
		return -2
	}
	if state.jpl.file.swapBytes != 0 {
		state.reorder(state.jpl.file.coefficientIndex[:], 4, 36)
	}
	readCount = readBinary(deNumber[:], 4, 1, state.jpl.file.handle)
	if readCount != 1 {
		return -2
	}
	if state.jpl.file.swapBytes != 0 {
		state.reorder(deNumber[:], 4, 1)
	}
	readCount = readBinary(lpt[:], 4, 3, state.jpl.file.handle)
	if readCount != 3 {
		return -2
	}
	if state.jpl.file.swapBytes != 0 {
		state.reorder(lpt[:], 4, 3)
	}
	for i = 0; i < 3; i++ {
		state.jpl.file.coefficientIndex[i+36] = lpt[i]
	}
	rewindFile(state.jpl.file.handle)
	kmx = 0
	khi = 0
	for i = 0; i < 13; i++ {
		if state.jpl.file.coefficientIndex[i*3] > kmx {
			kmx = state.jpl.file.coefficientIndex[i*3]
			khi = i + 1
		}
	}
	if khi == 12 {
		nd = 2
	} else {
		nd = 3
	}
	recordSize = int32((int64(state.jpl.file.coefficientIndex[khi*3-3]+nd*state.jpl.file.coefficientIndex[khi*3-2]*state.jpl.file.coefficientIndex[khi*3-1]) - 1) * 2)
	if recordSize == 1546 {
		recordSize = 1652
	}
	if recordSize < 1000 || recordSize > 5000 {
		if message != nil {
			formatText(message, textBytes("JPL ephemeris file does not provide valid ksize (%d)"), recordSize)
		}
		return -2
	}
	return recordSize
}

// swiPleph implements swi_pleph from swejpl.c.
//
//	This subroutine reads the jpl planetary ephemeris
//	and gives the position and velocity of the point 'ntarg'
//	with respect to 'ncent'.
//	calling sequence parameters:
//	  et = d.p. julian ephemeris date at which interpolation
//	       is wanted.
//	  ** note the entry dpleph for a doubly-dimensioned time **
//	     the reason for this option is discussed in the
//	     subroutine state
//	ntarg = integer number of 'target' point.
//	ncent = integer number of center point.
//	       the numbering convention for 'ntarg' and 'ncent' is:
//	           0 = mercury           7 = neptune
//	           1 = venus             8 = pluto
//	           2 = earth             9 = moon
//	           3 = mars             10 = sun
//	           4 = jupiter          11 = solar-system barycenter
//	           5 = saturn           12 = earth-moon barycenter
//	           6 = uranus           13 = nutations (longitude and obliq)
//	                                14 = librations, if on eph file
//	        (if nutations are wanted, set ntarg = 13. for librations,
//	         set ntarg = 14. set ncent=0.)
//	 rrd = output 6-word d.p. array containing position and velocity
//	       of point 'ntarg' relative to 'ncent'. the units are au and
//	       au/day. for librations the units are radians and radians
//	       per day. in the case of nutations the first four words of
//	       rrd will be set to nutations and rates, having units of
//	       radians and radians/day.
//	       The option is available to have the units in km and km/sec.
//	       For this, set do_km=TRUE (default FALSE).
func (state *State) swiPleph(jd float64, target int32, centre int32, result []float64, message []byte) int32 {
	var (
		i      int32
		status int32
		list   [12]int32
		pv     []float64 = state.jpl.file.planetVectors[:]
		pvsun  []float64 = state.jpl.file.sunVector[:]
	)
	for i = 0; i < 6; i++ {
		result[i] = 0.0
	}
	if target == centre {
		return 0
	}
	for i = 0; i < 12; i++ {
		list[i] = 0
	}
	if target == 13 {
		if state.jpl.file.coefficientIndex[34] > 0 {
			list[10] = 2
			return state.state(jd, list[:], 0, pv, pvsun, result, message)
		} else {
			if message != nil {
				formatText(message, textBytes("No nutations on the JPL ephemeris file;"))
			}
			return -2
		}
	}
	if target == 14 {
		if state.jpl.file.coefficientIndex[37] > 0 {
			list[11] = 2
			if status = state.state(jd, list[:], 0, pv, pvsun, result, message); status != 0 {
				return status
			}
			for i = 0; i < 6; i++ {
				result[i] = pv[i+60]
			}
			return 0
		} else {
			if message != nil {
				formatText(message, textBytes("No librations on the ephemeris file;"))
			}
			return -2
		}
	}
	if target < 10 {
		list[target] = 2
	}
	if target == 9 {
		list[2] = 2
	}
	if target == 2 {
		list[9] = 2
	}
	if target == 12 {
		list[2] = 2
	}
	if centre < 10 {
		list[centre] = 2
	}
	if centre == 9 {
		list[2] = 2
	}
	if centre == 2 {
		list[9] = 2
	}
	if centre == 12 {
		list[2] = 2
	}
	if status = state.state(jd, list[:], 1, pv, pvsun, result, message); status != 0 {
		return status
	}
	if target == 10 || centre == 10 {
		for i = 0; i < 6; i++ {
			pv[i+6*10] = pvsun[i]
		}
	}
	if target == 11 || centre == 11 {
		for i = 0; i < 6; i++ {
			pv[i+6*11] = 0.
		}
	}
	if target == 12 || centre == 12 {
		for i = 0; i < 6; i++ {
			pv[i+6*12] = pv[i+6*2]
		}
	}
	if target == 2 && centre == 9 || target == 9 && centre == 2 {
		for i = 0; i < 6; i++ {
			pv[i+6*2] = 0.
		}
	} else {
		if list[2] == 2 {
			for i = 0; i < 6; i++ {
				pv[i+6*2] -= pv[i+6*9] / (state.jpl.file.earthMoonMassRatio + 1.)
			}
		}
		if list[9] == 2 {
			for i = 0; i < 6; i++ {
				pv[i+6*9] += pv[i+6*2]
			}
		}
	}
	for i = 0; i < 6; i++ {
		result[i] = pv[i+target*6] - pv[i+centre*6]
	}
	return 0
}

// interp implements interp from swejpl.c.
func (state *State) interp(coefficients []float64, t float64, interval float64, coefficientCount int32, componentCount int32, subintervalCount int32, derivativeOrder int32, result []float64) int32 {
	var (
		pc   []float64 = state.jpl.file.positionPolynomial[:]
		vc   []float64 = state.jpl.file.velocityPolynomial[:]
		ac   []float64 = state.jpl.file.accelerationPolynomial[:]
		jc   []float64 = state.jpl.file.jerkPolynomial[:]
		ncf  int32     = coefficientCount
		ncm  int32     = componentCount
		na   int32     = subintervalCount
		temp float64
		i    int32
		j    int32
		ni   int32
		tc   float64
		dt1  float64
		bma  float64
		bma2 float64
		bma3 float64
	)
	if t >= 0 {
		dt1 = math.Floor(t)
	} else {
		dt1 = -math.Floor(-t)
	}
	temp = float64(na) * t
	ni = int32(temp - dt1)
	tc = (math.Mod(temp, 1.0)+dt1)*2. - 1.
	if tc != pc[1] {
		state.jpl.interpNp = 2
		state.jpl.interpNv = 3
		state.jpl.interpNac = 4
		state.jpl.interpNjk = 5
		pc[1] = tc
		state.jpl.interpTwot = tc + tc
	}
	if state.jpl.interpNp < ncf {
		for i = state.jpl.interpNp; i < ncf; i++ {
			pc[i] = state.jpl.interpTwot*pc[i-1] - pc[i-2]
		}
		state.jpl.interpNp = ncf
	}
	for i = 0; i < ncm; i++ {
		result[i] = 0.
		for j = ncf - 1; j >= 0; j-- {
			result[i] += pc[j] * coefficients[j+(i+ni*ncm)*ncf]
		}
	}
	if derivativeOrder <= 1 {
		return 0
	}
	bma = float64(na+na) / interval
	vc[2] = state.jpl.interpTwot + state.jpl.interpTwot
	if state.jpl.interpNv < ncf {
		for i = state.jpl.interpNv; i < ncf; i++ {
			vc[i] = state.jpl.interpTwot*vc[i-1] + pc[i-1] + pc[i-1] - vc[i-2]
		}
		state.jpl.interpNv = ncf
	}
	for i = 0; i < ncm; i++ {
		result[i+ncm] = 0.
		for j = ncf - 1; j >= 1; j-- {
			result[i+ncm] += vc[j] * coefficients[j+(i+ni*ncm)*ncf]
		}
		result[i+ncm] *= bma
	}
	if derivativeOrder == 2 {
		return 0
	}
	bma2 = bma * bma
	ac[3] = pc[1] * 24.
	if state.jpl.interpNac < ncf {
		state.jpl.interpNac = ncf
		for i = state.jpl.interpNac; i < ncf; i++ {
			ac[i] = state.jpl.interpTwot*ac[i-1] + vc[i-1]*4. - ac[i-2]
		}
	}
	for i = 0; i < ncm; i++ {
		result[i+ncm*2] = 0.
		for j = ncf - 1; j >= 2; j-- {
			result[i+ncm*2] += ac[j] * coefficients[j+(i+ni*ncm)*ncf]
		}
		result[i+ncm*2] *= bma2
	}
	if derivativeOrder == 3 {
		return 0
	}
	bma3 = bma * bma2
	jc[4] = pc[1] * 192.
	if state.jpl.interpNjk < ncf {
		state.jpl.interpNjk = ncf
		for i = state.jpl.interpNjk; i < ncf; i++ {
			jc[i] = state.jpl.interpTwot*jc[i-1] + ac[i-1]*6. - jc[i-2]
		}
	}
	for i = 0; i < ncm; i++ {
		result[i+ncm*3] = 0.
		for j = ncf - 1; j >= 3; j-- {
			result[i+ncm*3] += jc[j] * coefficients[j+(i+ni*ncm)*ncf]
		}
		result[i+ncm*3] *= bma3
	}
	return 0
}

// state implements state from swejpl.c.
func (state *State) state(et float64, list []int32, doBary int32, pv []float64, pvsun []float64, nut []float64, message []byte) int32 {
	var (
		i     int32
		j     int32
		k     int32
		nseg  int32
		flen  int64
		nb    int64
		buf   []float64 = state.jpl.file.record[:]
		aufac float64
		s     float64
		t     float64
		intv  float64
		ts    [4]float64
		nrecl int32
		ksize int32
		nr    int32
		etMn  float64
		etFr  float64
		ipt   []int32 = state.jpl.file.coefficientIndex[:]
		chTtl [252]byte
		nrd   uint64
	)
	if state.jpl.file.handle == nil {
		ksize = state.fsizer(message)
		nrecl = 4
		if ksize == -2 {
			return -2
		}
		state.jpl.stateIrecsz = nrecl * ksize
		state.jpl.stateNcoeffs = ksize / 2
		nrd = readBinary(chTtl[:], 1, 252, state.jpl.file.handle)
		if nrd != 252 {
			return -2
		}
		nrd = readBinary(state.jpl.file.constantNames[:], 1, 2400, state.jpl.file.handle)
		if nrd != 2400 {
			return -2
		}
		nrd = readBinary(state.jpl.file.epochRange[:], 8, 3, state.jpl.file.handle)
		if nrd != 3 {
			return -2
		}
		if state.jpl.file.swapBytes != 0 {
			state.reorder(state.jpl.file.epochRange[:], 8, 3)
		}
		nrd = readBinary(&state.jpl.file.constantCount, 4, 1, state.jpl.file.handle)
		if nrd != 1 {
			return -2
		}
		if state.jpl.file.swapBytes != 0 {
			state.reorder(&state.jpl.file.constantCount, 4, 1)
		}
		nrd = readBinary(&state.jpl.file.astronomicalUnit, 8, 1, state.jpl.file.handle)
		if nrd != 1 {
			return -2
		}
		if state.jpl.file.swapBytes != 0 {
			state.reorder(&state.jpl.file.astronomicalUnit, 8, 1)
		}
		nrd = readBinary(&state.jpl.file.earthMoonMassRatio, 8, 1, state.jpl.file.handle)
		if nrd != 1 {
			return -2
		}
		if state.jpl.file.swapBytes != 0 {
			state.reorder(&state.jpl.file.earthMoonMassRatio, 8, 1)
		}
		nrd = readBinary(ipt[:], 4, 36, state.jpl.file.handle)
		if nrd != 36 {
			return -2
		}
		if state.jpl.file.swapBytes != 0 {
			state.reorder(ipt[:], 4, 36)
		}
		nrd = readBinary(&state.jpl.file.deNumber, 4, 1, state.jpl.file.handle)
		if nrd != 1 {
			return -2
		}
		if state.jpl.file.swapBytes != 0 {
			state.reorder(&state.jpl.file.deNumber, 4, 1)
		}
		nrd = readBinary(state.jpl.stateLpt[:], 4, 3, state.jpl.file.handle)
		if nrd != 3 {
			return -2
		}
		if state.jpl.file.swapBytes != 0 {
			state.reorder(state.jpl.stateLpt[:], 4, 3)
		}
		seekFile(state.jpl.file.handle, 1*int64(state.jpl.stateIrecsz), 0)
		nrd = readBinary(state.jpl.file.constantValues[:], 8, 400, state.jpl.file.handle)
		if nrd != 400 {
			return -2
		}
		if state.jpl.file.swapBytes != 0 {
			state.reorder(state.jpl.file.constantValues[:], 8, 400)
		}
		for i = 0; i < 3; i++ {
			ipt[i+36] = state.jpl.stateLpt[i]
		}
		state.jpl.stateNrl = 0
		seekFile(state.jpl.file.handle, 0, 2)
		flen = tellFile(state.jpl.file.handle)
		nseg = int32((state.jpl.file.epochRange[1] - state.jpl.file.epochRange[0]) / state.jpl.file.epochRange[2])
		i = 0
		for nb = 0; i < 13; i++ {
			k = 3
			if i == 11 {
				k = 2
			}
			nb += int64(ipt[i*3+1] * ipt[i*3+2] * k * nseg)
		}
		nb += int64(2 * nseg)
		nb *= 8
		nb += int64(2 * ksize * nrecl)
		if flen != nb && flen-nb != int64(ksize*nrecl) {
			if message != nil {
				formatText(message, textBytes("JPL ephemeris file is mutilated; length = %d instead of %d."), uint32(flen), uint32(nb))
				if textLength(message)+textLength(state.jpl.file.name) < uint64(256-1) {
					formatText(message, textBytes("JPL ephemeris file %s is mutilated; length = %d instead of %d."), state.jpl.file.name, uint32(flen), uint32(nb))
				}
			}
			return -2
		}
		seekFile(state.jpl.file.handle, 2*int64(state.jpl.stateIrecsz), 0)
		nrd = readBinary(ts[:], 8, 2, state.jpl.file.handle)
		if nrd != 2 {
			return -2
		}
		if state.jpl.file.swapBytes != 0 {
			state.reorder(ts[:], 8, 2)
		}
		seekFile(state.jpl.file.handle, int64(nseg+2-1)*int64(state.jpl.stateIrecsz), 0)
		nrd = readBinary(ts[2:], 8, 2, state.jpl.file.handle)
		if nrd != 2 {
			return -2
		}
		if state.jpl.file.swapBytes != 0 {
			state.reorder(ts[2:], 8, 2)
		}
		if ts[0] != state.jpl.file.epochRange[0] || ts[3] != state.jpl.file.epochRange[1] {
			if message != nil {
				formatText(message, textBytes("JPL ephemeris file is corrupt; start/end date check failed. %.1f != %.1f || %.1f != %.1f"), ts[0], state.jpl.file.epochRange[0], ts[3], state.jpl.file.epochRange[1])
			}
			return -2
		}
	}
	if list == nil {
		return 0
	}
	s = et - .5
	etMn = math.Floor(s)
	etFr = s - etMn
	etMn += .5
	if et < state.jpl.file.epochRange[0] || et > state.jpl.file.epochRange[1] {
		if message != nil {
			formatText(message, textBytes("jd %f outside JPL eph. range %.2f .. %.2f;"), et, state.jpl.file.epochRange[0], state.jpl.file.epochRange[1])
		}
		return -3
	}
	nr = int32((etMn-state.jpl.file.epochRange[0])/state.jpl.file.epochRange[2]) + 2
	if etMn == state.jpl.file.epochRange[1] {
		nr--
	}
	t = (etMn - (float64(nr-2)*state.jpl.file.epochRange[2] + state.jpl.file.epochRange[0]) + etFr) / state.jpl.file.epochRange[2]
	if nr != state.jpl.stateNrl {
		state.jpl.stateNrl = nr
		if seekFile(state.jpl.file.handle, int64(nr)*int64(state.jpl.stateIrecsz), 0) != 0 {
			if message != nil {
				formatText(message, textBytes("Read error in JPL eph. at %f\n"), et)
			}
			return -2
		}
		for k = 1; k <= state.jpl.stateNcoeffs; k++ {
			if readBinary(buf[k-1:], 8, 1, state.jpl.file.handle) != 1 {
				if message != nil {
					formatText(message, textBytes("Read error in JPL eph. at %f\n"), et)
				}
				return -2
			}
			if state.jpl.file.swapBytes != 0 {
				state.reorder(buf[k-1:], 8, 1)
			}
		}
	}
	if state.jpl.file.useKilometres != 0 {
		intv = state.jpl.file.epochRange[2] * 86400.
		aufac = 1.
	} else {
		intv = state.jpl.file.epochRange[2]
		aufac = 1. / state.jpl.file.astronomicalUnit
	}
	state.interp(buf[ipt[30]-1:], t, intv, ipt[31], 3, ipt[32], 2, pvsun)
	for i = 0; i < 6; i++ {
		pvsun[i] *= aufac
	}
	for i = 0; i < 10; i++ {
		if list[i] > 0 {
			state.interp(buf[ipt[i*3]-1:], t, intv, ipt[i*3+1], 3, ipt[i*3+2], list[i], pv[i*6:])
			for j = 0; j < 6; j++ {
				if i < 9 && !(doBary != 0) {
					pv[j+i*6] = pv[j+i*6]*aufac - pvsun[j]
				} else {
					pv[j+i*6] *= aufac
				}
			}
		}
	}
	if list[10] > 0 && ipt[34] > 0 {
		state.interp(buf[ipt[33]-1:], t, intv, ipt[34], 2, ipt[35], list[10], nut)
	}
	if list[11] > 0 && ipt[37] > 0 {
		state.interp(buf[ipt[36]-1:], t, intv, ipt[37], 3, ipt[38], list[1], pv[60:])
	}
	return 0
}

// readConstJpl implements read_const_jpl from swejpl.c.
//
//	this entry obtains the constants from the ephemeris file
//	call state to initialize the ephemeris and read in the constants
func (state *State) readConstJpl(ss []float64, message []byte) int32 {
	var (
		i      int32
		status int32
	)
	status = state.state(0.0, nil, 0, nil, nil, nil, message)
	if status != 0 {
		return status
	}
	for i = 0; i < 3; i++ {
		ss[i] = state.jpl.file.epochRange[i]
	}
	return 0
}

// swiCloseJplFile implements swi_close_jpl_file from swejpl.c.
func (state *State) swiCloseJplFile() {
	if state.jpl.file != nil {
		if state.jpl.file.handle != nil {
			closeTextFile(state.jpl.file.handle)
		}
		state.jpl.file = nil
	}
}

// swiOpenJplFile implements swi_open_jpl_file from swejpl.c.
func (state *State) swiOpenJplFile(ss []float64, fname []byte, fpath []byte, message []byte) int32 {
	var status int32 = 0
	if state.jpl.file != nil && state.jpl.file.handle != nil {
		return 0
	}
	state.jpl.file = &jplFile{name: duplicateText(fname), path: duplicateText(fpath)}
	status = state.readConstJpl(ss, message)
	if status != 0 {
		state.swiCloseJplFile()
	} else {
		state.jpl.file.positionPolynomial[0] = 1
		state.jpl.file.positionPolynomial[1] = 2
		state.jpl.file.velocityPolynomial[1] = 1
		state.jpl.file.accelerationPolynomial[2] = 4
		state.jpl.file.jerkPolynomial[3] = 24
	}
	return status
}

// swiGetJplDenum implements swi_get_jpl_denum from swejpl.c.
//
//	intializations for function interpol()
func (state *State) swiGetJplDenum() int32 {
	return state.jpl.file.deNumber
}
