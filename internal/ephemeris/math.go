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

// sweDegnorm implements swe_degnorm from swephlib.c.
// sweDegnorm normalizes degrees to [0, 360), snapping roundoff near zero.
//
//	Reduce x modulo 360 degrees
func (state *State) sweDegnorm(angle float64) float64 {
	normalized := math.Mod(angle, 360)
	if math.Abs(normalized) < 1e-13 {
		normalized = 0
	}
	if normalized < 0 {
		normalized += 360
	}
	return normalized
}

// sweRadnorm implements swe_radnorm from swephlib.c.
// sweRadnorm is the radian counterpart of swe_degnorm.
//
//	Reduce x modulo TWOPI degrees
func (state *State) sweRadnorm(angle float64) float64 {
	normalized := math.Mod(angle, 2*math.Pi)
	if math.Abs(normalized) < 1e-13 {
		normalized = 0
	}
	if normalized < 0 {
		normalized += 2 * math.Pi
	}
	return normalized
}

// sweDegMidp implements swe_deg_midp from swephlib.c.
// sweDegMidp finds the midpoint along the shortest arc between two angles.
//
//	Alois fix 11-dec-1999
func (state *State) sweDegMidp(first, second float64) float64 {
	difference := state.sweDifdeg2n(first, second)
	return state.sweDegnorm(second + difference/2)
}

// sweRadMidp implements swe_rad_midp from swephlib.c.
// sweRadMidp retains swe_rad_midp's conversion through degrees.
//
//	arc from x0 to x1
func (state *State) sweRadMidp(first, second float64) float64 {
	return radiansPerDegree * state.sweDegMidp(first*degreesPerRadian, second*degreesPerRadian)
}

// swiMod2PI implements swi_mod2PI from swephlib.c.
// swiMod2PI normalizes radians without the public API's near-zero snap.
//
//	Reduce x modulo 2*PI
func (state *State) swiMod2PI(angle float64) float64 {
	normalized := math.Mod(angle, 2*math.Pi)
	if normalized < 0 {
		normalized += 2 * math.Pi
	}
	return normalized
}

// swiAngnorm implements swi_angnorm from swephlib.c.
// swiAngnorm corrects one revolution; its input is already within [-2pi, 4pi).
//
//	Reduce x modulo 2*PI
func (state *State) swiAngnorm(angle float64) float64 {
	if angle < 0 {
		return angle + 2*math.Pi
	}
	if angle >= 2*math.Pi {
		return angle - 2*math.Pi
	}
	return angle
}

// swiCrossProd implements swi_cross_prod from swephlib.c.
// swiCrossProd writes the three-dimensional cross product, as swi_cross_prod.
func (state *State) swiCrossProd(first, second, output []float64) {
	output[0] = first[1]*second[2] - first[2]*second[1]
	output[1] = first[2]*second[0] - first[0]*second[2]
	output[2] = first[0]*second[1] - first[1]*second[0]
}

// swiEcheb evaluates a Chebyshev series with Clenshaw's recurrence. The Swiss
// file convention stores twice the constant term; retain the final factor 0.5.
func (state *State) swiEcheb(argument float64, coefficients []float64, count int32) float64 {
	twiceArgument := argument * 2
	current, previous, previous2 := 0.0, 0.0, 0.0
	for index := count - 1; index >= 0; index-- {
		previous2 = previous
		previous = current
		current = twiceArgument*previous - previous2 + coefficients[index]
	}
	return (current - previous2) * 0.5
}

// swiEdcheb implements swi_edcheb from swephlib.c.
// swiEdcheb evaluates the derivative of the same Chebyshev series. Keeping the
// upstream recurrence order avoids cancellation changes in planetary speeds.
//
// evaluates derivative of chebyshev series, see echeb
func (state *State) swiEdcheb(argument float64, coefficients []float64, count int32) float64 {
	twiceArgument := argument * 2
	current, beforePrevious := 0.0, 0.0
	previous, previous2 := 0.0, 0.0
	derivativePrevious, derivativePrevious2 := 0.0, 0.0
	for index := count - 1; index >= 1; index-- {
		derivative := coefficients[index]*float64(index+index) + derivativePrevious2
		current = twiceArgument*previous - previous2 + derivative
		beforePrevious = previous2
		previous2 = previous
		previous = current
		derivativePrevious2 = derivativePrevious
		derivativePrevious = derivative
	}
	return (current - beforePrevious) * 0.5
}

// sweCotrans implements swe_cotrans from swephlib.c.
//
// conversion between ecliptical and equatorial polar coordinates.
// for users of SWISSEPH, not used by our routines.
// for ecl. to equ.  eps must be negative.
// for equ. to ecl.  eps must be positive.
// xpo, xpn are arrays of 3 doubles containing position.
// attention: input must be in degrees!
func (state *State) sweCotrans(xpo []float64, xpn []float64, eps float64) {
	var (
		i int32
		x [6]float64
		e float64 = eps * radiansPerDegree
	)
	for i = 0; i <= 1; i++ {
		x[i] = xpo[i]
	}
	x[0] *= radiansPerDegree
	x[1] *= radiansPerDegree
	x[2] = 1
	for i = 3; i <= 5; i++ {
		x[i] = 0
	}
	state.swiPolcart(x[:], x[:])
	state.swiCoortrf(x[:], x[:], e)
	state.swiCartpol(x[:], x[:])
	xpn[0] = x[0] * degreesPerRadian
	xpn[1] = x[1] * degreesPerRadian
	xpn[2] = xpo[2]
}

// sweCotransSp implements swe_cotrans_sp from swephlib.c.
//
// conversion between ecliptical and equatorial polar coordinates
// with speed.
// for users of SWISSEPH, not used by our routines.
// for ecl. to equ.  eps must be negative.
// for equ. to ecl.  eps must be positive.
// xpo, xpn are arrays of 6 doubles containing position and speed.
// attention: input must be in degrees!
func (state *State) sweCotransSp(xpo []float64, xpn []float64, eps float64) {
	var (
		i int32
		x [6]float64
		e float64 = eps * radiansPerDegree
	)
	for i = 0; i <= 5; i++ {
		x[i] = xpo[i]
	}
	x[0] *= radiansPerDegree
	x[1] *= radiansPerDegree
	x[2] = 1
	x[3] *= radiansPerDegree
	x[4] *= radiansPerDegree
	state.swiPolcartSp(x[:], x[:])
	state.swiCoortrf(x[:], x[:], e)
	state.swiCoortrf(x[3:], x[3:], e)
	state.swiCartpolSp(x[:], xpn)
	xpn[0] *= degreesPerRadian
	xpn[1] *= degreesPerRadian
	xpn[2] = xpo[2]
	xpn[3] *= degreesPerRadian
	xpn[4] *= degreesPerRadian
	xpn[5] = xpo[5]
}

// swiCoortrf implements swi_coortrf from swephlib.c.
//
// conversion between ecliptical and equatorial cartesian coordinates
// for ecl. to equ.  eps must be negative
// for equ. to ecl.  eps must be positive
func (state *State) swiCoortrf(xpo []float64, xpn []float64, eps float64) {
	var (
		sineps float64
		coseps float64
		x      [3]float64
	)
	sineps = math.Sin(eps)
	coseps = math.Cos(eps)
	x[0] = xpo[0]
	x[1] = xpo[1]*coseps + xpo[2]*sineps
	x[2] = -xpo[1]*sineps + xpo[2]*coseps
	xpn[0] = x[0]
	xpn[1] = x[1]
	xpn[2] = x[2]
}

// swiCoortrf2 implements swi_coortrf2 from swephlib.c.
func (state *State) swiCoortrf2(xpo []float64, xpn []float64, sineps float64, coseps float64) {
	var x [3]float64
	x[0] = xpo[0]
	x[1] = xpo[1]*coseps + xpo[2]*sineps
	x[2] = -xpo[1]*sineps + xpo[2]*coseps
	xpn[0] = x[0]
	xpn[1] = x[1]
	xpn[2] = x[2]
}

// swiCartpol implements swi_cartpol from swephlib.c.
//
//	conversion of cartesian (x[3]) to polar coordinates (l[3]).
//
// x = l is allowed.
// if |x| = 0, then lon, lat and rad := 0.
func (state *State) swiCartpol(x []float64, l []float64) {
	var (
		rxy float64
		ll  [3]float64
	)
	if x[0] == 0 && x[1] == 0 && x[2] == 0 {
		l[2] = 0
		l[1] = l[2]
		l[0] = l[1]
		return
	}
	rxy = x[0]*x[0] + x[1]*x[1]
	ll[2] = math.Sqrt(rxy + x[2]*x[2])
	rxy = math.Sqrt(rxy)
	ll[0] = math.Atan2(x[1], x[0])
	if ll[0] < 0.0 {
		ll[0] += 6.2831853071795862
	}
	if rxy == 0 {
		if x[2] >= 0 {
			ll[1] = 1.5707963267948966
		} else {
			ll[1] = -1.5707963267948966
		}
	} else {
		ll[1] = math.Atan(x[2] / rxy)
	}
	l[0] = ll[0]
	l[1] = ll[1]
	l[2] = ll[2]
}

// swiPolcart implements swi_polcart from swephlib.c.
//
//	conversion from polar (l[3]) to cartesian coordinates (x[3]).
//
// x = l is allowed.
func (state *State) swiPolcart(l []float64, x []float64) {
	var (
		xx    [3]float64
		cosl1 float64
	)
	cosl1 = math.Cos(l[1])
	xx[0] = l[2] * cosl1 * math.Cos(l[0])
	xx[1] = l[2] * cosl1 * math.Sin(l[0])
	xx[2] = l[2] * math.Sin(l[1])
	x[0] = xx[0]
	x[1] = xx[1]
	x[2] = xx[2]
}

// swiCartpolSp implements swi_cartpol_sp from swephlib.c.
func (state *State) swiCartpolSp(x []float64, l []float64) {
	var (
		i      int32
		xx     [6]float64
		ll     [6]float64
		rxy    float64
		coslon float64
		sinlon float64
		coslat float64
		sinlat float64
	)
	if x[0] == 0 && x[1] == 0 && x[2] == 0 {
		ll[4] = 0
		ll[3] = ll[4]
		ll[1] = ll[3]
		ll[0] = ll[1]
		ll[5] = math.Sqrt(x[3]*x[3] + x[3:][1]*x[3:][1] + x[3:][2]*x[3:][2])
		state.swiCartpol(x[3:], ll[:])
		ll[2] = 0
		for i = 0; i <= 5; i++ {
			l[i] = ll[i]
		}
		return
	}
	if x[3] == 0 && x[4] == 0 && x[5] == 0 {
		l[5] = 0
		l[4] = l[5]
		l[3] = l[4]
		state.swiCartpol(x, l)
		return
	}
	rxy = x[0]*x[0] + x[1]*x[1]
	ll[2] = math.Sqrt(rxy + x[2]*x[2])
	rxy = math.Sqrt(rxy)
	ll[0] = math.Atan2(x[1], x[0])
	if ll[0] < 0.0 {
		ll[0] += 6.2831853071795862
	}
	ll[1] = math.Atan(x[2] / rxy)
	coslon = x[0] / rxy
	sinlon = x[1] / rxy
	coslat = rxy / ll[2]
	sinlat = x[2] / ll[2]
	xx[3] = x[3]*coslon + x[4]*sinlon
	xx[4] = -x[3]*sinlon + x[4]*coslon
	l[3] = xx[4] / rxy
	xx[4] = -sinlat*xx[3] + coslat*x[5]
	xx[5] = coslat*xx[3] + sinlat*x[5]
	l[4] = xx[4] / ll[2]
	l[5] = xx[5]
	l[0] = ll[0]
	l[1] = ll[1]
	l[2] = ll[2]
}

// swiPolcartSp implements swi_polcart_sp from swephlib.c.
func (state *State) swiPolcartSp(l []float64, x []float64) {
	var (
		sinlon float64
		coslon float64
		sinlat float64
		coslat float64
		xx     [6]float64
		rxy    float64
		rxyz   float64
	)
	if l[3] == 0 && l[4] == 0 && l[5] == 0 {
		x[5] = 0
		x[4] = x[5]
		x[3] = x[4]
		state.swiPolcart(l, x)
		return
	}
	coslon = math.Cos(l[0])
	sinlon = math.Sin(l[0])
	coslat = math.Cos(l[1])
	sinlat = math.Sin(l[1])
	xx[0] = l[2] * coslat * coslon
	xx[1] = l[2] * coslat * sinlon
	xx[2] = l[2] * sinlat
	rxyz = l[2]
	rxy = math.Sqrt(xx[0]*xx[0] + xx[1]*xx[1])
	xx[5] = l[5]
	xx[4] = l[4] * rxyz
	x[5] = sinlat*xx[5] + coslat*xx[4]
	xx[3] = coslat*xx[5] - sinlat*xx[4]
	xx[4] = l[3] * rxy
	x[3] = coslon*xx[3] - sinlon*xx[4]
	x[4] = sinlon*xx[3] + coslon*xx[4]
	x[0] = xx[0]
	x[1] = xx[1]
	x[2] = xx[2]
}

// swiDotProdUnit implements swi_dot_prod_unit from swephlib.c.
//
//	return position
func (state *State) swiDotProdUnit(x []float64, y []float64) float64 {
	var (
		dop float64 = x[0]*y[0] + x[1]*y[1] + x[2]*y[2]
		e1  float64 = math.Sqrt(x[0]*x[0] + x[1]*x[1] + x[2]*x[2])
		e2  float64 = math.Sqrt(y[0]*y[0] + y[1]*y[1] + y[2]*y[2])
	)
	dop /= e1
	dop /= e2
	if dop > 1 {
		dop = 1
	}
	if dop < float64(-1) {
		dop = float64(-1)
	}
	return dop
}

// swiLdpPeps implements swi_ldp_peps from swephlib.c.
func (state *State) swiLdpPeps(jd float64, dpre []float64, deps []float64) {
	var (
		i    int32
		npol int32 = 4
		nper int32 = 10
		t    float64
		p    float64
		q    float64
		w    float64
		a    float64
		s    float64
		c    float64
	)
	t = (jd - julianDayJ2000) / daysPerCentury
	p = 0
	q = 0
	for i = 0; i < nper; i++ {
		w = 6.2831853071795862 * t
		a = w / mathPeper[0][i]
		s = math.Sin(a)
		c = math.Cos(a)
		p += c*mathPeper[1][i] + s*mathPeper[3][i]
		q += c*mathPeper[2][i] + s*mathPeper[4][i]
	}
	w = 1
	for i = 0; i < npol; i++ {
		p += mathPepol[i][0] * w
		q += mathPepol[i][1] * w
		w *= t
	}
	p *= 4.8481368110953598e-06
	q *= 4.8481368110953598e-06
	if dpre != nil {
		dpre[0] = p
	}
	if deps != nil {
		deps[0] = q
	}
}

// prePecl implements pre_pecl from swephlib.c.
//
//	precession of the ecliptic
func (state *State) prePecl(jd float64, vec []float64) {
	var (
		i    int32
		npol int32 = 4
		nper int32 = 8
		t    float64
		p    float64
		q    float64
		w    float64
		a    float64
		s    float64
		c    float64
		z    float64
	)
	t = (jd - julianDayJ2000) / daysPerCentury
	p = 0
	q = 0
	for i = 0; i < nper; i++ {
		w = 6.2831853071795862 * t
		a = w / mathPqper[0][i]
		s = math.Sin(a)
		c = math.Cos(a)
		p += c*mathPqper[1][i] + s*mathPqper[3][i]
		q += c*mathPqper[2][i] + s*mathPqper[4][i]
	}
	w = 1
	for i = 0; i < npol; i++ {
		p += mathPqpol[i][0] * w
		q += mathPqpol[i][1] * w
		w *= t
	}
	p *= 4.8481368110953598e-06
	q *= 4.8481368110953598e-06
	z = 1 - p*p - q*q
	if z < 0 {
		z = 0
	} else {
		z = math.Sqrt(z)
	}
	s = math.Sin(0.40909260060058289)
	c = math.Cos(0.40909260060058289)
	vec[0] = p
	vec[1] = -q*c - z*s
	vec[2] = -q*s + z*c
}

// prePequ implements pre_pequ from swephlib.c.
//
//	precession of the equator
func (state *State) prePequ(jd float64, veq []float64) {
	var (
		i    int32
		npol int32 = 4
		nper int32 = 14
		t    float64
		x    float64
		y    float64
		w    float64
		a    float64
		s    float64
		c    float64
	)
	t = (jd - julianDayJ2000) / daysPerCentury
	x = 0
	y = 0
	for i = 0; i < nper; i++ {
		w = 6.2831853071795862 * t
		a = w / mathXyper[0][i]
		s = math.Sin(a)
		c = math.Cos(a)
		x += c*mathXyper[1][i] + s*mathXyper[3][i]
		y += c*mathXyper[2][i] + s*mathXyper[4][i]
	}
	w = 1
	for i = 0; i < npol; i++ {
		x += mathXypol[i][0] * w
		y += mathXypol[i][1] * w
		w *= t
	}
	x *= 4.8481368110953598e-06
	y *= 4.8481368110953598e-06
	veq[0] = x
	veq[1] = y
	w = x*x + y*y
	if w < 1 {
		veq[2] = math.Sqrt(1 - w)
	} else {
		veq[2] = 0
	}
}

// prePmat implements pre_pmat from swephlib.c.
//
//	precession matrix
func (state *State) prePmat(jd float64, rp []float64) {
	var (
		peqr [3]float64
		pecl [3]float64
		v    [3]float64
		w    float64
		eqx  [3]float64
	)
	state.prePequ(jd, peqr[:])
	state.prePecl(jd, pecl[:])
	state.swiCrossProd(peqr[:], pecl[:], v[:])
	w = math.Sqrt(v[0]*v[0] + v[1]*v[1] + v[2]*v[2])
	eqx[0] = v[0] / w
	eqx[1] = v[1] / w
	eqx[2] = v[2] / w
	state.swiCrossProd(peqr[:], eqx[:], v[:])
	rp[0] = eqx[0]
	rp[1] = eqx[1]
	rp[2] = eqx[2]
	rp[3] = v[0]
	rp[4] = v[1]
	rp[5] = v[2]
	rp[6] = peqr[0]
	rp[7] = peqr[1]
	rp[8] = peqr[2]
}

// getOwenT0Icof implements get_owen_t0_icof from swephlib.c.
func (state *State) getOwenT0Icof(jd float64, t0 []float64, icof []int32) {
	var (
		i   int32
		j   int32      = 0
		t0s [5]float64 = [5]float64{-3392455.5, -470455.5, 2451544.5, 5373544.5, 8295544.5}
	)
	t0[0] = t0s[0]
	for i = 1; i < 5; i++ {
		if jd < (t0s[i-1]+t0s[i])/2 {
		} else {
			t0[0] = t0s[i]
			j++
		}
	}
	icof[0] = j
}

// owenPreMatrix implements owen_pre_matrix from swephlib.c.
//
//	precession matrix Owen 1990
func (state *State) owenPreMatrix(jd float64, rp []float64, flags int32) {
	var (
		i       int32
		icof    [1]int32 = [1]int32{0}
		eps0    float64  = 0
		chia    float64  = 0
		psia    float64  = 0
		oma     float64  = 0
		coseps0 float64
		sineps0 float64
		coschia float64
		sinchia float64
		cospsia float64
		sinpsia float64
		cosoma  float64
		sinoma  float64
		k       [10]float64
		tau     [10]float64
		t0      [1]float64
	)
	state.getOwenT0Icof(jd, t0[:], icof[:])
	tau[0] = 0
	tau[1] = (jd - t0[0]) / daysPerCentury / 40.0
	for i = 2; i <= 9; i++ {
		tau[i] = tau[1] * tau[i-1]
	}
	k[0] = 1
	k[1] = tau[1]
	k[2] = 2*tau[2] - 1
	k[3] = 4*tau[3] - 3*tau[1]
	k[4] = 8*tau[4] - 8*tau[2] + 1
	k[5] = 16*tau[5] - 20*tau[3] + 5*tau[1]
	k[6] = 32*tau[6] - 48*tau[4] + 18*tau[2] - 1
	k[7] = 64*tau[7] - 112*tau[5] + 56*tau[3] - 7*tau[1]
	k[8] = 128*tau[8] - 256*tau[6] + 160*tau[4] - 32*tau[2] + 1
	k[9] = 256*tau[9] - 576*tau[7] + 432*tau[5] - 120*tau[3] + 9*tau[1]
	for i = 0; i < 10; i++ {
		psia += k[i] * mathOwenPsiaCoef[icof[0]][i]
		oma += k[i] * mathOwenOmaCoef[icof[0]][i]
		chia += k[i] * mathOwenChiaCoef[icof[0]][i]
	}
	if flags&(flagHorizons|flagHorizonsApprox) != 0 {
		psia += -0.000018560
	}
	eps0 = 23.439291111111114
	eps0 *= radiansPerDegree
	psia *= radiansPerDegree
	chia *= radiansPerDegree
	oma *= radiansPerDegree
	coseps0 = math.Cos(eps0)
	sineps0 = math.Sin(eps0)
	coschia = math.Cos(chia)
	sinchia = math.Sin(chia)
	cospsia = math.Cos(psia)
	sinpsia = math.Sin(psia)
	cosoma = math.Cos(oma)
	sinoma = math.Sin(oma)
	rp[0] = coschia*cospsia + sinchia*cosoma*sinpsia
	rp[1] = (-coschia*sinpsia+sinchia*cosoma*cospsia)*coseps0 + sinchia*sinoma*sineps0
	rp[2] = (-coschia*sinpsia+sinchia*cosoma*cospsia)*sineps0 - sinchia*sinoma*coseps0
	rp[3] = -sinchia*cospsia + coschia*cosoma*sinpsia
	rp[4] = (sinchia*sinpsia+coschia*cosoma*cospsia)*coseps0 + coschia*sinoma*sineps0
	rp[5] = (sinchia*sinpsia+coschia*cosoma*cospsia)*sineps0 - coschia*sinoma*coseps0
	rp[6] = sinoma * sinpsia
	rp[7] = sinoma*cospsia*coseps0 - cosoma*sineps0
	rp[8] = sinoma*cospsia*sineps0 + cosoma*coseps0
}

// epsilnOwen1986 implements epsiln_owen_1986 from swephlib.c.
//
//	for (i = 0; i < 3; i++) {
//	    fprintf(stderr, "(%.17f   %.17f   %.17f)\n", rp[i*3], rp[i*3+1],rp[i*3+2]);
//	  }
func (state *State) epsilnOwen1986(jd float64, eps []float64) {
	var (
		i    int32
		icof [1]int32 = [1]int32{0}
		k    [10]float64
		tau  [10]float64
		t0   [1]float64
	)
	state.getOwenT0Icof(jd, t0[:], icof[:])
	eps[0] = 0
	tau[0] = 0
	tau[1] = (jd - t0[0]) / daysPerCentury / 40.0
	for i = 2; i <= 9; i++ {
		tau[i] = tau[1] * tau[i-1]
	}
	k[0] = 1
	k[1] = tau[1]
	k[2] = 2*tau[2] - 1
	k[3] = 4*tau[3] - 3*tau[1]
	k[4] = 8*tau[4] - 8*tau[2] + 1
	k[5] = 16*tau[5] - 20*tau[3] + 5*tau[1]
	k[6] = 32*tau[6] - 48*tau[4] + 18*tau[2] - 1
	k[7] = 64*tau[7] - 112*tau[5] + 56*tau[3] - 7*tau[1]
	k[8] = 128*tau[8] - 256*tau[6] + 160*tau[4] - 32*tau[2] + 1
	k[9] = 256*tau[9] - 576*tau[7] + 432*tau[5] - 120*tau[3] + 9*tau[1]
	for i = 0; i < 10; i++ {
		eps[0] += k[i] * mathOwenEps0Coef[icof[0]][i]
	}
}

// swiEpsiln implements swi_epsiln from swephlib.c.
func (state *State) swiEpsiln(J float64, flags int32) float64 {
	var (
		T              float64
		eps            [1]float64
		tofs           float64
		dofs           float64
		t0             float64
		t1             float64
		precModel      int32 = state.positions.data.astroModels[1]
		precModelShort int32 = state.positions.data.astroModels[2]
		jplhoraModel   int32 = state.positions.data.astroModels[6]
		isJplhor       int32 = 0
	)
	if precModel == 0 {
		precModel = 9
	}
	if precModelShort == 0 {
		precModelShort = 9
	}
	if jplhoraModel == 0 {
		jplhoraModel = 3
	}
	if flags&flagHorizons != 0 {
		isJplhor = 1
	}
	if flags&flagHorizonsApprox != 0 && jplhoraModel == 3 && J <= 2437684.5 {
		isJplhor = 1
	}
	T = (J - julianDayJ2000) / daysPerCentury
	if isJplhor != 0 {
		if J > 2378131.5 && J < 2525323.5 {
			eps[0] = (((1.813e-3*T-5.9e-4)*T-46.8150)*T + 84381.448) * radiansPerDegree / 3600
		} else {
			state.epsilnOwen1986(J, eps[:])
			eps[0] *= radiansPerDegree
		}
	} else if flags&flagHorizonsApprox != 0 && jplhoraModel == 2 {
		eps[0] = (((1.813e-3*T-5.9e-4)*T-46.8150)*T + 84381.448) * radiansPerDegree / 3600
	} else if precModelShort == 1 && math.Abs(T) <= 2.0 {
		eps[0] = (((1.813e-3*T-5.9e-4)*T-46.8150)*T + 84381.448) * radiansPerDegree / 3600
	} else if precModel == 1 {
		eps[0] = (((1.813e-3*T-5.9e-4)*T-46.8150)*T + 84381.448) * radiansPerDegree / 3600
	} else if precModelShort == 6 && math.Abs(T) <= 2.0 {
		eps[0] = (((1.813e-3*T-5.9e-4)*T-46.84024)*T + 84381.406) * radiansPerDegree / 3600
	} else if precModel == 6 {
		eps[0] = (((1.813e-3*T-5.9e-4)*T-46.84024)*T + 84381.406) * radiansPerDegree / 3600
	} else if precModelShort == 8 && math.Abs(T) <= 75.0 {
		eps[0] = (((((-4.34e-8*T-5.76e-7)*T+2.0034e-3)*T-1.831e-4)*T-46.836769)*T + 84381.406) * radiansPerDegree / 3600.0
	} else if precModel == 11 {
		var Tn float64 = (J - 2396758.0) / daysPerCentury
		eps[0] = (0.0017*Tn*Tn*Tn - 0.0085*Tn*Tn - 46.837*Tn + 84451.68) * radiansPerDegree / 3600.0
	} else if precModel == 8 {
		eps[0] = (((((-4.34e-8*T-5.76e-7)*T+2.0034e-3)*T-1.831e-4)*T-46.836769)*T + 84381.406) * radiansPerDegree / 3600.0
	} else if precModel == 7 {
		eps[0] = ((((((-3e-11*T-2.48e-8)*T-5.23e-7)*T+1.99911e-3)*T-1.667e-4)*T-46.836051)*T + 84381.40880) * radiansPerDegree / 3600.0
	} else if precModel == 5 {
		eps[0] = (((((2.5e-8*T-5.1e-7)*T+1.9989e-3)*T-1.52e-4)*T-46.80927)*T + 84381.412) * radiansPerDegree / 3600.0
	} else if precModel == 4 {
		eps[0] = ((((-1.0e-6*T+2.0e-3)*T-1.74e-4)*T-46.833960)*T + 84381.409) * radiansPerDegree / 3600.0
	} else if precModel == 2 || precModel == 3 {
		T /= 10.0
		eps[0] = (((((((((2.45e-10*T+5.79e-9)*T+2.787e-7)*T+7.12e-7)*T-3.905e-5)*T-2.4967e-3)*T-5.138e-3)*T+1.99925)*T-0.0155)*T-468.093)*T + 84381.448
		eps[0] *= 4.8481368110953598e-06
	} else if precModel == 10 {
		state.epsilnOwen1986(J, eps[:])
		eps[0] *= radiansPerDegree
	} else {
		state.swiLdpPeps(J, nil, eps[:])
		if flags&flagHorizonsApprox != 0 && jplhoraModel != 2 {
			tofs = (J - 2437846.5) / 365.25
			dofs = 35.95
			if tofs < 0 {
				tofs = 0
				dofs = mathDcorEpsJpl[0]
			} else if tofs >= float64(50) {
				tofs = 51
				dofs = mathDcorEpsJpl[51-1]
			} else {
				t0 = float64(int32(tofs))
				t1 = t0 + 1
				dofs = mathDcorEpsJpl[int32(t0)]
				dofs = (tofs-t0)*(mathDcorEpsJpl[int32(t0)]-mathDcorEpsJpl[int32(t1)]) + mathDcorEpsJpl[int32(t0)]
			}
			dofs /= 3600000
			eps[0] += dofs * radiansPerDegree
		}
	}
	return eps[0]
}

// precess1 implements precess_1 from swephlib.c.
func (state *State) precess1(R []float64, J float64, direction int32, precMethod int32) int32 {
	var (
		T     float64
		Z     float64 = 0
		z     float64 = 0
		TH    float64 = 0
		i     int32
		x     [3]float64
		sinth float64
		costh float64
		sinZ  float64
		cosZ  float64
		sinz  float64
		cosz  float64
		A     float64
		B     float64
	)
	if J == julianDayJ2000 {
		return 0
	}
	T = (J - julianDayJ2000) / daysPerCentury
	if precMethod == 1 {
		Z = ((0.017998*T+0.30188)*T + 2306.2181) * T * radiansPerDegree / 3600
		z = ((0.018203*T+1.09468)*T + 2306.2181) * T * radiansPerDegree / 3600
		TH = ((-0.041833*T-0.42665)*T + 2004.3109) * T * radiansPerDegree / 3600
	} else if precMethod == 6 {
		Z = (((((-0.0000002*T-0.0000327)*T+0.0179663)*T+0.3019015)*T+2306.0809506)*T + 2.5976176) * radiansPerDegree / 3600
		z = (((((-0.0000003*T-0.000047)*T+0.0182237)*T+1.0947790)*T+2306.0803226)*T - 2.5976176) * radiansPerDegree / 3600
		TH = ((((-0.0000001*T-0.0000601)*T-0.0418251)*T-0.4269353)*T + 2004.1917476) * T * radiansPerDegree / 3600
	} else if precMethod == 8 {
		T = (J - julianDayJ2000) / daysPerCentury
		Z = (((((-0.0000003173*T-0.000005971)*T+0.01801828)*T+0.2988499)*T+2306.083227)*T + 2.650545) * radiansPerDegree / 3600
		z = (((((-0.0000002904*T-0.000028596)*T+0.01826837)*T+1.0927348)*T+2306.077181)*T - 2.650545) * radiansPerDegree / 3600
		TH = ((((-0.00000011274*T-0.000007089)*T-0.04182264)*T-0.4294934)*T + 2004.191903) * T * radiansPerDegree / 3600
	} else if precMethod == 7 {
		Z = ((((((-0.00000000013*T-0.0000003040)*T-0.000005708)*T+0.01801752)*T+0.3023262)*T+2306.080472)*T + 2.72767) * radiansPerDegree / 3600
		z = ((((((-0.00000000005*T-0.0000002486)*T-0.000028276)*T+0.01826676)*T+1.0956768)*T+2306.076070)*T - 2.72767) * radiansPerDegree / 3600
		TH = ((((((0.000000000009*T+0.00000000036)*T-0.0000001127)*T-0.000007291)*T-0.04182364)*T-0.4266980)*T + 2004.190936) * T * radiansPerDegree / 3600
	} else if precMethod == 11 {
		var (
			mills float64 = 365242.198782
			t1    float64 = 54786.796418999787 / mills
			t2    float64 = (J - 2396758.2035810) / mills
			T     float64 = t2 - t1
			T2    float64 = T * T
			T3    float64 = T2 * T
			Z1    float64 = 23035.5548 + 139.720*t1 + 0.069*t1*t1
		)
		Z = Z1*T + (30.242-0.269*t1)*T2 + 17.996*T3
		z = Z1*T + (109.478-0.387*t1)*T2 + 18.324*T3
		TH = (20051.125-85.294*t1-0.365*t1*t1)*T + (-42.647-0.365*t1)*T2 - 41.802*T3
		Z *= 4.8481368110953598e-06
		z *= 4.8481368110953598e-06
		TH *= 4.8481368110953598e-06
	} else {
		return 0
	}
	sinth = math.Sin(TH)
	costh = math.Cos(TH)
	sinZ = math.Sin(Z)
	cosZ = math.Cos(Z)
	sinz = math.Sin(z)
	cosz = math.Cos(z)
	A = cosZ * costh
	B = sinZ * costh
	if direction < 0 {
		x[0] = (A*cosz-sinZ*sinz)*R[0] - (B*cosz+cosZ*sinz)*R[1] - sinth*cosz*R[2]
		x[1] = (A*sinz+sinZ*cosz)*R[0] - (B*sinz-cosZ*cosz)*R[1] - sinth*sinz*R[2]
		x[2] = cosZ*sinth*R[0] - sinZ*sinth*R[1] + costh*R[2]
	} else {
		x[0] = (A*cosz-sinZ*sinz)*R[0] + (A*sinz+sinZ*cosz)*R[1] + cosZ*sinth*R[2]
		x[1] = -(B*cosz+cosZ*sinz)*R[0] - (B*sinz-cosZ*cosz)*R[1] - sinZ*sinth*R[2]
		x[2] = -sinth*cosz*R[0] - sinth*sinz*R[1] + costh*R[2]
	}
	for i = 0; i < 3; i++ {
		R[i] = x[i]
	}
	return 0
}

// precess2 implements precess_2 from swephlib.c.
func (state *State) precess2(R []float64, J float64, flags int32, direction int32, precMethod int32) int32 {
	var (
		i       int32
		T       float64
		z       float64
		eps     float64
		sineps  float64
		coseps  float64
		x       [3]float64
		p       []float64
		A       float64
		B       float64
		pA      float64
		W       float64
		pAcof   []float64
		inclcof []float64
		nodecof []float64
	)
	if J == julianDayJ2000 {
		return 0
	}
	if precMethod == 2 {
		pAcof = mathPAcofLaskar[:]
		nodecof = mathNodecofLaskar[:]
		inclcof = mathInclcofLaskar[:]
	} else if precMethod == 5 {
		pAcof = mathPAcofSimon[:]
		nodecof = mathNodecofSimon[:]
		inclcof = mathInclcofSimon[:]
	} else if precMethod == 4 {
		pAcof = mathPAcofWilliams[:]
		nodecof = mathNodecofWilliams[:]
		inclcof = mathInclcofWilliams[:]
	} else {
		pAcof = mathPAcofLaskar[:]
		nodecof = mathNodecofLaskar[:]
		inclcof = mathInclcofLaskar[:]
	}
	T = (J - julianDayJ2000) / daysPerCentury
	if direction == 1 {
		eps = state.swiEpsiln(J, flags)
	} else {
		eps = state.swiEpsiln(julianDayJ2000, flags)
	}
	sineps = math.Sin(eps)
	coseps = math.Cos(eps)
	x[0] = R[0]
	z = coseps*R[1] + sineps*R[2]
	x[2] = -sineps*R[1] + coseps*R[2]
	x[1] = z
	T /= 10.0
	p = pAcof
	pA = nextValue(&p)
	for i = 0; i < 9; i++ {
		pA = pA*T + nextValue(&p)
	}
	pA *= 4.8481368110953598e-06 * T
	p = nodecof
	W = nextValue(&p)
	for i = 0; i < 10; i++ {
		W = W*T + nextValue(&p)
	}
	if direction == 1 {
		z = W + pA
	} else {
		z = W
	}
	B = math.Cos(z)
	A = math.Sin(z)
	z = B*x[0] + A*x[1]
	x[1] = -A*x[0] + B*x[1]
	x[0] = z
	p = inclcof
	z = nextValue(&p)
	for i = 0; i < 10; i++ {
		z = z*T + nextValue(&p)
	}
	if direction == 1 {
		z = -z
	}
	B = math.Cos(z)
	A = math.Sin(z)
	z = B*x[1] + A*x[2]
	x[2] = -A*x[1] + B*x[2]
	x[1] = z
	if direction == 1 {
		z = -W
	} else {
		z = -W - pA
	}
	B = math.Cos(z)
	A = math.Sin(z)
	z = B*x[0] + A*x[1]
	x[1] = -A*x[0] + B*x[1]
	x[0] = z
	if direction == 1 {
		eps = state.swiEpsiln(julianDayJ2000, flags)
	} else {
		eps = state.swiEpsiln(J, flags)
	}
	sineps = math.Sin(eps)
	coseps = math.Cos(eps)
	z = coseps*x[1] - sineps*x[2]
	x[2] = sineps*x[1] + coseps*x[2]
	x[1] = z
	for i = 0; i < 3; i++ {
		R[i] = x[i]
	}
	return 0
}

// precess3 implements precess_3 from swephlib.c.
func (state *State) precess3(R []float64, J float64, direction int32, flags int32, precMeth int32) int32 {
	var (
		x    [3]float64
		pmat [9]float64
		i    int32
		j    int32
	)
	if J == julianDayJ2000 {
		return 0
	}
	if precMeth == 10 {
		state.owenPreMatrix(J, pmat[:], flags)
	} else {
		state.prePmat(J, pmat[:])
	}
	if direction == -1 {
		i = 0
		for j = 0; i <= 2; {
			x[i] = R[0]*pmat[j+0] + R[1]*pmat[j+1] + R[2]*pmat[j+2]
			i++
			j = i * 3
		}
	} else {
		i = 0
		for j = 0; i <= 2; {
			x[i] = R[0]*pmat[i+0] + R[1]*pmat[i+3] + R[2]*pmat[i+6]
			i++
			j = i * 3
		}
	}
	for i = 0; i < 3; i++ {
		R[i] = x[i]
	}
	return 0
}

// swiPrecess implements swi_precess from swephlib.c.
//
//	Subroutine arguments:
//
// R = rectangular equatorial coordinate vector to be precessed.
//
//	The result is written back into the input vector.
//
// J = Julian date
// direction =
//
//	Precess from J to J2000: direction = 1
//	Precess from J2000 to J: direction = -1
//
// Note that if you want to precess from J1 to J2, you would
// first go from J1 to J2000, then call the program again
// to go from J2000 to J2.
func (state *State) swiPrecess(R []float64, J float64, flags int32, direction int32) int32 {
	var (
		T              float64 = (J - julianDayJ2000) / daysPerCentury
		precModel      int32   = state.positions.data.astroModels[1]
		precModelShort int32   = state.positions.data.astroModels[2]
		jplhoraModel   int32   = state.positions.data.astroModels[6]
		isJplhor       int32   = 0
	)
	if precModel == 0 {
		precModel = 9
	}
	if precModelShort == 0 {
		precModelShort = 9
	}
	if jplhoraModel == 0 {
		jplhoraModel = 3
	}
	if flags&flagHorizons != 0 {
		isJplhor = 1
	}
	if flags&flagHorizonsApprox != 0 && jplhoraModel == 3 && J <= 2437684.5 {
		isJplhor = 1
	}
	if isJplhor != 0 {
		if J > 2378131.5 && J < 2525323.5 {
			return state.precess1(R, J, direction, 1)
		} else {
			return state.precess3(R, J, direction, flags, 10)
		}
	} else if precModelShort == 1 && math.Abs(T) <= 2.0 {
		return state.precess1(R, J, direction, 1)
	} else if precModel == 1 {
		return state.precess1(R, J, direction, 1)
	} else if precModelShort == 6 && math.Abs(T) <= 2.0 {
		return state.precess1(R, J, direction, 6)
	} else if precModel == 6 {
		return state.precess1(R, J, direction, 6)
	} else if precModelShort == 8 && math.Abs(T) <= 75.0 {
		return state.precess1(R, J, direction, 8)
	} else if precModel == 8 {
		return state.precess1(R, J, direction, 8)
	} else if precModel == 7 {
		return state.precess1(R, J, direction, 7)
	} else if precModel == 11 {
		return state.precess1(R, J, direction, 11)
	} else if precModel == 2 {
		return state.precess2(R, J, flags, direction, 2)
	} else if precModel == 5 {
		return state.precess2(R, J, flags, direction, 5)
	} else if precModel == 4 || precModel == 3 {
		return state.precess2(R, J, flags, direction, 4)
	} else if precModel == 10 {
		return state.precess3(R, J, direction, flags, 10)
	} else {
		return state.precess3(R, J, direction, flags, 9)
	}
}

// calcNutationIau1980 implements calc_nutation_iau1980 from swephlib.c.
//
// #endif
func (state *State) calcNutationIau1980(J float64, nutlo []float64) int32 {
	var (
		ss       [5][8]float64
		cc       [5][8]float64
		arg      float64
		args     [5]float64
		f        float64
		g        float64
		T        float64
		T2       float64
		MM       float64
		MS       float64
		FF       float64
		DD       float64
		OM       float64
		cu       float64
		su       float64
		cv       float64
		sv       float64
		sw       float64
		s        float64
		C        float64
		D        float64
		i        int32
		j        int32
		k        int32
		k1       int32
		m        int32
		n        int32
		ns       [5]int32
		p        []int16
		nutModel int32 = state.positions.data.astroModels[3]
	)
	if nutModel == 0 {
		nutModel = 4
	}
	T = (J - julianDayJ2000) / daysPerCentury
	T2 = T * T
	OM = -6962890.539*T + 450160.280 + (0.008*T+7.455)*T2
	OM = state.sweDegnorm(OM/3600) * radiansPerDegree
	MS = 129596581.224*T + 1287099.804 - (0.012*T+0.577)*T2
	MS = state.sweDegnorm(MS/3600) * radiansPerDegree
	MM = 1717915922.633*T + 485866.733 + (0.064*T+31.310)*T2
	MM = state.sweDegnorm(MM/3600) * radiansPerDegree
	FF = 1739527263.137*T + 335778.877 + (0.011*T-13.257)*T2
	FF = state.sweDegnorm(FF/3600) * radiansPerDegree
	DD = 1602961601.328*T + 1072261.307 + (0.019*T-6.891)*T2
	DD = state.sweDegnorm(DD/3600) * radiansPerDegree
	args[0] = MM
	ns[0] = 3
	args[1] = MS
	ns[1] = 2
	args[2] = FF
	ns[2] = 4
	args[3] = DD
	ns[3] = 4
	args[4] = OM
	ns[4] = 2
	for k = 0; k <= 4; k++ {
		arg = args[k]
		n = ns[k]
		su = math.Sin(arg)
		cu = math.Cos(arg)
		ss[k][0] = su
		cc[k][0] = cu
		sv = 2.0 * su * cu
		cv = cu*cu - su*su
		ss[k][1] = sv
		cc[k][1] = cv
		for i = 2; i < n; i++ {
			s = su*cv + cu*sv
			cv = cu*cv - su*sv
			sv = s
			ss[k][i] = sv
			cc[k][i] = cv
		}
	}
	C = (-0.01742*T - 17.1996) * ss[4][0]
	D = (0.00089*T + 9.2025) * cc[4][0]
	for p = mathNt[:]; int32(p[0]) != -99; p = p[9:] {
		if nutModel != 2 && (int32(p[0]) == 101 || int32(p[0]) == 102) {
			continue
		}
		k1 = 0
		cv = 0.0
		sv = 0.0
		for m = 0; m < 5; m++ {
			j = int32(p[m])
			if j > 100 {
				j = 0
			}
			if j != 0 {
				k = j
				if j < 0 {
					k = -k
				}
				su = ss[m][k-1]
				if j < 0 {
					su = -su
				}
				cu = cc[m][k-1]
				if k1 == 0 {
					sv = su
					cv = cu
					k1 = 1
				} else {
					sw = su*cv + cu*sv
					cv = cu*cv - su*sv
					sv = sw
				}
			}
		}
		f = float64(p[5]) * 0.0001
		if int32(p[6]) != 0 {
			f += 0.00001 * T * float64(p[6])
		}
		g = float64(p[7]) * 0.0001
		if int32(p[8]) != 0 {
			g += 0.00001 * T * float64(p[8])
		}
		if int32(p[0]) >= 100 {
			f *= 0.1
			g *= 0.1
		}
		if int32(p[0]) != 102 {
			C += f * sv
			D += g * cv
		} else {
			C += f * cv
			D += g * sv
		}
	}
	nutlo[0] = radiansPerDegree * C / 3600.0
	nutlo[1] = radiansPerDegree * D / 3600.0
	return 0
}

// calcNutationIau2000ab implements calc_nutation_iau2000ab from swephlib.c.
func (state *State) calcNutationIau2000ab(J float64, nutlo []float64) int32 {
	var (
		i        int32
		j        int32
		k        int32
		inls     int32
		M        float64
		SM       float64
		F        float64
		D        float64
		OM       float64
		AL       float64
		ALSU     float64
		AF       float64
		AD       float64
		AOM      float64
		APA      float64
		ALME     float64
		ALVE     float64
		ALEA     float64
		ALMA     float64
		ALJU     float64
		ALSA     float64
		ALUR     float64
		ALNE     float64
		darg     float64
		sinarg   float64
		cosarg   float64
		dpsi     float64 = 0
		deps     float64 = 0
		T        float64 = (J - julianDayJ2000) / daysPerCentury
		nutModel int32   = state.positions.data.astroModels[3]
	)
	if nutModel == 0 {
		nutModel = 4
	}
	M = state.sweDegnorm((485868.249036+T*(1717915923.2178+T*(31.8792+T*(0.051635+T*-0.00024470))))/3600.0) * radiansPerDegree
	SM = state.sweDegnorm((1287104.79305+T*(129596581.0481+T*(-0.5532+T*(0.000136+T*-0.00001149))))/3600.0) * radiansPerDegree
	F = state.sweDegnorm((335779.526232+T*(1739527262.8478+T*(-12.7512+T*(-0.001037+T*0.00000417))))/3600.0) * radiansPerDegree
	D = state.sweDegnorm((1072260.70369+T*(1602961601.2090+T*(-6.3706+T*(0.006593+T*-0.00003169))))/3600.0) * radiansPerDegree
	OM = state.sweDegnorm((450160.398036+T*(-6962890.5431+T*(7.4722+T*(0.007702+T*-0.00005939))))/3600.0) * radiansPerDegree
	if nutModel == 4 {
		inls = 77
	} else {
		inls = 678
	}
	for i = inls - 1; i >= 0; i-- {
		j = i * 5
		darg = state.sweRadnorm(float64(mathNls[j+0])*M + float64(mathNls[j+1])*SM + float64(mathNls[j+2])*F + float64(mathNls[j+3])*D + float64(mathNls[j+4])*OM)
		sinarg = math.Sin(darg)
		cosarg = math.Cos(darg)
		k = i * 6
		dpsi += (float64(mathCls[k+0])+float64(mathCls[k+1])*T)*sinarg + float64(mathCls[k+2])*cosarg
		deps += (float64(mathCls[k+3])+float64(mathCls[k+4])*T)*cosarg + float64(mathCls[k+5])*sinarg
	}
	nutlo[0] = dpsi * 2.7777777777777777e-11
	nutlo[1] = deps * 2.7777777777777777e-11
	if nutModel == 3 {
		AL = state.sweRadnorm(2.35555598 + 8328.6914269554*T)
		ALSU = state.sweRadnorm(6.24006013 + 628.301955*T)
		AF = state.sweRadnorm(1.627905234 + 8433.466158131*T)
		AD = state.sweRadnorm(5.198466741 + 7771.3771468121*T)
		AOM = state.sweRadnorm(2.18243920 - 33.757045*T)
		ALME = state.sweRadnorm(4.402608842 + 2608.7903141574*T)
		ALVE = state.sweRadnorm(3.176146697 + 1021.3285546211*T)
		ALEA = state.sweRadnorm(1.753470314 + 628.3075849991*T)
		ALMA = state.sweRadnorm(6.203480913 + 334.0612426700*T)
		ALJU = state.sweRadnorm(0.599546497 + 52.9690962641*T)
		ALSA = state.sweRadnorm(0.874016757 + 21.3299104960*T)
		ALUR = state.sweRadnorm(5.481293871 + 7.4781598567*T)
		ALNE = state.sweRadnorm(5.321159000 + 3.8127774000*T)
		APA = (0.02438175 + 0.00000538691*T) * T
		dpsi = 0
		deps = 0
		for i = 687 - 1; i >= 0; i-- {
			j = i * 14
			darg = state.sweRadnorm(float64(mathNpl[j+0])*AL + float64(mathNpl[j+1])*ALSU + float64(mathNpl[j+2])*AF + float64(mathNpl[j+3])*AD + float64(mathNpl[j+4])*AOM + float64(mathNpl[j+5])*ALME + float64(mathNpl[j+6])*ALVE + float64(mathNpl[j+7])*ALEA + float64(mathNpl[j+8])*ALMA + float64(mathNpl[j+9])*ALJU + float64(mathNpl[j+10])*ALSA + float64(mathNpl[j+11])*ALUR + float64(mathNpl[j+12])*ALNE + float64(mathNpl[j+13])*APA)
			k = i * 4
			sinarg = math.Sin(darg)
			cosarg = math.Cos(darg)
			dpsi += float64(mathIcpl[k+0])*sinarg + float64(mathIcpl[k+1])*cosarg
			deps += float64(mathIcpl[k+2])*sinarg + float64(mathIcpl[k+3])*cosarg
		}
		nutlo[0] += dpsi * 2.7777777777777777e-11
		nutlo[1] += deps * 2.7777777777777777e-11
		dpsi = -8.1*math.Sin(OM) - 0.6*math.Sin(2*F-2*D+2*OM)
		dpsi += T * (47.8*math.Sin(OM) + 3.7*math.Sin(2*F-2*D+2*OM) + 0.6*math.Sin(2*F+2*OM) - 0.6*math.Sin(2*OM))
		deps = T * (-25.6*math.Cos(OM) - 1.6*math.Cos(2*F-2*D+2*OM))
		nutlo[0] += dpsi / 3600000000
		nutlo[1] += deps / 3600000000
	}
	nutlo[0] *= radiansPerDegree
	nutlo[1] *= radiansPerDegree
	return 0
}

// calcNutationWoolard implements calc_nutation_woolard from swephlib.c.
//
//	an incomplete implementation of nutation Woolard 1953
func (state *State) calcNutationWoolard(J float64, nutlo []float64) int32 {
	var (
		deps float64
		dpsi float64
		ls   float64
		ld   float64
		ms   float64
		md   float64
		nm   float64
		t    float64
		t2   float64
		tls  float64
		tnm  float64
		tld  float64
		a    float64
		b    float64
		mjd  float64 = J - 2415020.0
	)
	t = mjd / 36525.
	t2 = t * t
	a = 100.0021358 * t
	b = 360. * (a - float64(int64(a)))
	ls = 279.697 + .000303*t2 + b
	a = 1336.855231 * t
	b = 360. * (a - float64(int64(a)))
	ld = 270.434 - .001133*t2 + b
	a = 99.99736056000026 * t
	b = 360. * (a - float64(int64(a)))
	ms = 358.476 - .00015*t2 + b
	a = 13255523.59 * t
	b = 360. * (a - float64(int64(a)))
	md = 296.105 + .009192*t2 + b
	a = 5.372616667 * t
	b = 360. * (a - float64(int64(a)))
	nm = 259.183 + .002078*t2 - b
	tls = 2 * ls * radiansPerDegree
	nm = nm * radiansPerDegree
	tnm = 2 * nm
	ms = ms * radiansPerDegree
	tld = 2 * ld * radiansPerDegree
	md = md * radiansPerDegree
	dpsi = (-17.2327-.01737*t)*math.Sin(nm) + (-1.2729-.00013*t)*math.Sin(tls) + .2088*math.Sin(tnm) - .2037*math.Sin(tld) + (.1261-.00031*t)*math.Sin(ms) + .0675*math.Sin(md) - (.0497-.00012*t)*math.Sin(tls+ms) - .0342*math.Sin(tld-nm) - .0261*math.Sin(tld+md) + .0214*math.Sin(tls-ms) - .0149*math.Sin(tls-tld+md) + .0124*math.Sin(tls-nm) + .0114*math.Sin(tld-md)
	deps = (9.21+.00091*t)*math.Cos(nm) + (.5522-.00029*t)*math.Cos(tls) - .0904*math.Cos(tnm) + .0884*math.Cos(tld) + .0216*math.Cos(tls+ms) + .0183*math.Cos(tld-nm) + .0113*math.Cos(tld+md) - .0093*math.Cos(tls-ms) - .0066*math.Cos(tls-nm)
	dpsi = dpsi / 3600.0 * radiansPerDegree
	deps = deps / 3600.0 * radiansPerDegree
	nutlo[1] = deps
	nutlo[0] = dpsi
	return 0
}

// bessel implements bessel from swephlib.c.
//
//	convert to radians.
func (state *State) bessel(v []float64, n int32, t float64) float64 {
	var (
		i   int32
		iy  int32
		k   int32
		ans float64
		p   float64
		B   float64
		d   [6]float64
	)
	if t <= 0 {
		ans = v[0]
		return ans
	}
	if t >= float64(n-1) {
		ans = v[n-1]
		return ans
	}
	p = math.Floor(t)
	iy = int32(t)
	ans = v[iy]
	k = iy + 1
	if k >= n {
		return ans
	}
	p = t - p
	ans += p * (v[k] - v[iy])
	if iy-1 < 0 || iy+2 >= n {
		return ans
	}
	k = iy - 2
	for i = 0; i < 5; i++ {
		if k < 0 || k+1 >= n {
			d[i] = 0
		} else {
			d[i] = v[k+1] - v[k]
		}
		k += 1
	}
	for i = 0; i < 4; i++ {
		d[i] = d[i+1] - d[i]
	}
	B = 0.25 * p * (p - 1.0)
	ans += B * (d[1] + d[2])
	if iy+2 >= n {
		return ans
	}
	for i = 0; i < 3; i++ {
		d[i] = d[i+1] - d[i]
	}
	B = 2.0 * B / 3.0
	ans += (p - 0.5) * B * d[1]
	if iy-2 < 0 || iy+3 > n {
		return ans
	}
	for i = 0; i < 2; i++ {
		d[i] = d[i+1] - d[i]
	}
	B = 0.125 * B * (p + 1.0) * (p - 2.0)
	ans += B * (d[0] + d[1])
	return ans
}

// calcNutation implements calc_nutation from swephlib.c.
func (state *State) calcNutation(J float64, flags int32, nutlo []float64) int32 {
	var (
		n            int32
		dpsi         float64
		deps         float64
		J2           float64
		nutModel     int32 = state.positions.data.astroModels[3]
		jplhoraModel int32 = state.positions.data.astroModels[6]
		isJplhor     int32 = 0
	)
	if nutModel == 0 {
		nutModel = 4
	}
	if jplhoraModel == 0 {
		jplhoraModel = 3
	}
	if flags&flagHorizons != 0 {
		isJplhor = 1
	}
	if flags&flagHorizonsApprox != 0 && jplhoraModel == 3 && J <= 2437684.5 {
		isJplhor = 1
	}
	if isJplhor != 0 {
		state.calcNutationIau1980(J, nutlo)
		if flags&flagHorizons != 0 {
			n = int32(state.positions.data.eopTjdEnd - state.positions.data.eopTjdBeg + 0.000001)
			J2 = J
			if J < state.positions.data.eopTjdBegHorizons {
				J2 = state.positions.data.eopTjdBegHorizons
			}
			dpsi = state.bessel(state.positions.data.dpsi, n+1, J2-state.positions.data.eopTjdBeg)
			deps = state.bessel(state.positions.data.deps, n+1, J2-state.positions.data.eopTjdBeg)
			nutlo[0] += dpsi / 3600.0 * radiansPerDegree
			nutlo[1] += deps / 3600.0 * radiansPerDegree
		} else {
			nutlo[0] += 3.1165762676445416e-07
			nutlo[1] += 2.9820889525047561e-08
		}
	} else if nutModel == 1 || nutModel == 2 {
		state.calcNutationIau1980(J, nutlo)
	} else if nutModel == 3 || nutModel == 4 {
		state.calcNutationIau2000ab(J, nutlo)
		if flags&flagHorizonsApprox != 0 && jplhoraModel == 2 {
			nutlo[0] += -2.0253091528350864e-07
			nutlo[1] += -3.3060414542221477e-08
		}
	} else if nutModel == 5 {
		state.calcNutationWoolard(J, nutlo)
	}
	return 0
}

// quadraticIntp implements quadratic_intp from swephlib.c.
func (state *State) quadraticIntp(ym float64, y0 float64, yp float64, x float64) float64 {
	var (
		a float64
		b float64
		c float64
		y float64
	)
	c = y0
	b = (yp - ym) / 2.0
	a = (yp+ym)/2.0 - c
	y = a*x*x + b*x + c
	return y
}

// swiNutation implements swi_nutation from swephlib.c.
func (state *State) swiNutation(jd float64, flags int32, nutlo []float64) int32 {
	var (
		status int32 = 0
		dnut   [2]float64
		dx     float64
	)
	if !(state.positions.data.doInterpolateNut != 0) {
		status = state.calcNutation(jd, flags, nutlo)
	} else if jd < state.positions.data.interpol.tjdNut2 && jd > state.positions.data.interpol.tjdNut0 {
		dx = jd - state.positions.data.interpol.tjdNut0 - 1.0
		nutlo[0] = state.quadraticIntp(state.positions.data.interpol.nutDpsi0, state.positions.data.interpol.nutDpsi1, state.positions.data.interpol.nutDpsi2, dx)
		nutlo[1] = state.quadraticIntp(state.positions.data.interpol.nutDeps0, state.positions.data.interpol.nutDeps1, state.positions.data.interpol.nutDeps2, dx)
	} else {
		state.positions.data.interpol.tjdNut0 = jd - 1.0
		state.positions.data.interpol.tjdNut2 = jd + 1.0
		status = state.calcNutation(state.positions.data.interpol.tjdNut0, flags, dnut[:])
		if status == -1 {
			return -1
		}
		state.positions.data.interpol.nutDpsi0 = dnut[0]
		state.positions.data.interpol.nutDeps0 = dnut[1]
		status = state.calcNutation(state.positions.data.interpol.tjdNut2, flags, dnut[:])
		if status == -1 {
			return -1
		}
		state.positions.data.interpol.nutDpsi2 = dnut[0]
		state.positions.data.interpol.nutDeps2 = dnut[1]
		status = state.calcNutation(jd, flags, nutlo)
		if status == -1 {
			return -1
		}
		state.positions.data.interpol.nutDpsi1 = nutlo[0]
		state.positions.data.interpol.nutDeps1 = nutlo[1]
	}
	return status
}

// swiApproxJplhor implements swi_approx_jplhor from swephlib.c.
func (state *State) swiApproxJplhor(x []float64, jd float64, flags int32, backward int32) {
	var (
		t0           float64
		t1           float64
		t            float64 = (jd - 2437846.5) / 365.25
		dofs         float64 = -52.3
		jplhoraModel int32   = state.positions.data.astroModels[6]
	)
	if jplhoraModel == 0 {
		jplhoraModel = 3
	}
	if flags&flagHorizonsApprox == 0 {
		return
	}
	if jplhoraModel == 2 {
		return
	}
	if t < 0 {
		t = 0
		dofs = mathDcorRaJpl[0]
	} else if t >= float64(50) {
		t = 51
		dofs = mathDcorRaJpl[51-1]
	} else {
		t0 = float64(int32(t))
		t1 = t0 + 1
		dofs = mathDcorRaJpl[int32(t0)]
		dofs = (t-t0)*(mathDcorRaJpl[int32(t0)]-mathDcorRaJpl[int32(t1)]) + mathDcorRaJpl[int32(t0)]
	}
	dofs /= 3600000
	state.swiCartpol(x, x)
	if backward != 0 {
		x[0] -= dofs * radiansPerDegree
	} else {
		x[0] += dofs * radiansPerDegree
	}
	state.swiPolcart(x, x)
}

// swiBias implements swi_bias from swephlib.c.
//
//	GCRS to J2000
func (state *State) swiBias(x []float64, jd float64, flags int32, backward int32) {
	var (
		xx           [6]float64
		rb           [3][3]float64
		i            int32
		biasModel    int32 = state.positions.data.astroModels[4]
		jplhoraModel int32 = state.positions.data.astroModels[6]
	)
	if biasModel == 0 {
		biasModel = 3
	}
	if jplhoraModel == 0 {
		jplhoraModel = 3
	}
	if biasModel == 1 {
		return
	}
	if flags&flagHorizonsApprox != 0 {
		if jplhoraModel == 2 {
			return
		}
		if jplhoraModel == 3 && jd < 2437684.5 {
			return
		}
	}
	if biasModel == 3 {
		rb[0][0] = +0.99999999999999412
		rb[1][0] = -0.00000007078368961
		rb[2][0] = +0.00000008056213978
		rb[0][1] = +0.00000007078368695
		rb[1][1] = +0.99999999999999700
		rb[2][1] = +0.00000003306428553
		rb[0][2] = -0.00000008056214212
		rb[1][2] = -0.00000003306427981
		rb[2][2] = +0.99999999999999634
	} else {
		rb[0][0] = +0.9999999999999942
		rb[1][0] = -0.0000000707827974
		rb[2][0] = +0.0000000805621715
		rb[0][1] = +0.0000000707827948
		rb[1][1] = +0.9999999999999969
		rb[2][1] = +0.0000000330604145
		rb[0][2] = -0.0000000805621738
		rb[1][2] = -0.0000000330604088
		rb[2][2] = +0.9999999999999962
	}
	if backward != 0 {
		state.swiApproxJplhor(x, jd, flags, 1)
		for i = 0; i <= 2; i++ {
			xx[i] = x[0]*rb[i][0] + x[1]*rb[i][1] + x[2]*rb[i][2]
			if flags&flagSpeed != 0 {
				xx[i+3] = x[3]*rb[i][0] + x[4]*rb[i][1] + x[5]*rb[i][2]
			}
		}
	} else {
		for i = 0; i <= 2; i++ {
			xx[i] = x[0]*rb[0][i] + x[1]*rb[1][i] + x[2]*rb[2][i]
			if flags&flagSpeed != 0 {
				xx[i+3] = x[3]*rb[0][i] + x[4]*rb[1][i] + x[5]*rb[2][i]
			}
		}
		state.swiApproxJplhor(xx[:], jd, flags, 0)
	}
	for i = 0; i <= 2; i++ {
		x[i] = xx[i]
	}
	if flags&flagSpeed != 0 {
		for i = 3; i <= 5; i++ {
			x[i] = xx[i]
		}
	}
}

// swiIcrs2fk5 implements swi_icrs2fk5 from swephlib.c.
//
//	GCRS to FK5
func (state *State) swiIcrs2fk5(x []float64, flags int32, backward int32) {
	var (
		xx [6]float64
		rb [3][3]float64
		i  int32
	)
	rb[0][0] = +0.9999999999999928
	rb[0][1] = +0.0000001110223287
	rb[0][2] = +0.0000000441180557
	rb[1][0] = -0.0000001110223330
	rb[1][1] = +0.9999999999999891
	rb[1][2] = +0.0000000964779176
	rb[2][0] = -0.0000000441180450
	rb[2][1] = -0.0000000964779225
	rb[2][2] = +0.9999999999999943
	if backward != 0 {
		for i = 0; i <= 2; i++ {
			xx[i] = x[0]*rb[i][0] + x[1]*rb[i][1] + x[2]*rb[i][2]
			if flags&flagSpeed != 0 {
				xx[i+3] = x[3]*rb[i][0] + x[4]*rb[i][1] + x[5]*rb[i][2]
			}
		}
	} else {
		for i = 0; i <= 2; i++ {
			xx[i] = x[0]*rb[0][i] + x[1]*rb[1][i] + x[2]*rb[2][i]
			if flags&flagSpeed != 0 {
				xx[i+3] = x[3]*rb[0][i] + x[4]*rb[1][i] + x[5]*rb[2][i]
			}
		}
	}
	for i = 0; i <= 5; i++ {
		x[i] = xx[i]
	}
}

// calcDeltat implements calc_deltat from swephlib.c.
func (state *State) calcDeltat(jd float64, flags int32, deltat []float64, message []byte) int32 {
	var (
		ans         float64 = 0
		B           float64
		Y           float64
		Ygreg       float64
		dd          float64
		iy          int32
		status      int32
		deltatModel int32 = state.positions.data.astroModels[0]
		tidAcc      [1]float64
		denum       int32
		denumret    [1]int32
		ephemeris   int32
		otherflag   int32
	)
	if deltatModel == 0 {
		deltatModel = 5
	}
	ephemeris = flags & (flagJPL | flagSwiss | flagMoshier)
	otherflag = flags & ^(1 | 2 | 4)
	if flags == -1 {
		status = state.swiGetTidAcc(jd, 0, 9999, denumret[:], tidAcc[:], message)
	} else {
		denum = state.positions.data.jplDENumber
		if ephemeris&2 != 0 {
			denum = state.positions.data.files[1].deNumber
		}
		if state.swiInitSwedIfStart() == 1 && !(ephemeris&4 != 0) {
			if message != nil {
				copyText(message, textBytes("Please call swe_set_ephe_path() or swe_set_jplfile() before calling swe_deltat_ex()"))
			}
			status = state.swiSetTidAcc(jd, ephemeris, denum, nil)
		} else {
			status = state.swiSetTidAcc(jd, ephemeris, denum, message)
		}
		tidAcc[0] = state.positions.data.tidAcc
	}
	flags = otherflag | status
	Y = 2000.0 + (jd-julianDayJ2000)/365.25
	Ygreg = 2000.0 + (jd-julianDayJ2000)/365.2425
	if deltatModel == 5 && jd < 2435108.5 {
		deltat[0] = state.deltatStephensonEtc2016(jd, tidAcc[0])
		if jd >= 2434108.5 {
			deltat[0] += (1.0 - (2435108.5-jd)/1000.0) * 0.6610218 / secondsPerDay
		}
		return flags
	}
	if deltatModel == 4 && jd < 2317746.13090277789 {
		deltat[0] = state.deltatEspenakMeeus1620(jd, tidAcc[0])
		return flags
	}
	if deltatModel == 3 && Y < 1620 {
		if Y < 1600 {
			deltat[0] = state.deltatStephensonMorrison20041600(jd, tidAcc[0])
			return flags
		} else if Y >= 1600 {
			B = float64(20)
			iy = (1600 - -1000) / 100
			dd = (Y - 1600) / B
			ans = float64(mathDt2[iy]) + dd*(state.time.deltaTTable[0]-float64(mathDt2[iy]))
			ans = state.adjustForTidacc(ans, Ygreg, tidAcc[0], -26.0, 0)
			deltat[0] = ans / secondsPerDay
			return flags
		}
	}
	if deltatModel == 2 && Y < 1620 {
		if Y < 1600 {
			deltat[0] = state.deltatStephensonMorrison19971600(jd, tidAcc[0])
			return flags
		} else if Y >= 1600 {
			B = float64(20)
			iy = (1600 - -500) / 50
			dd = (Y - 1600) / B
			ans = float64(mathDt97[iy]) + dd*(state.time.deltaTTable[0]-float64(mathDt97[iy]))
			ans = state.adjustForTidacc(ans, Ygreg, tidAcc[0], -26.0, 0)
			deltat[0] = ans / secondsPerDay
			return flags
		}
	}
	if deltatModel == 1 && Y < 1620 {
		if Y >= 948.0 {
			B = 0.01 * (Y - 2000.0)
			ans = (23.58*B+100.3)*B + 101.6
		} else {
			B = 0.01*(Y-2000.0) + 3.75
			ans = 35.0*B*B + 40.
		}
		deltat[0] = ans / secondsPerDay
		return flags
	}
	if Y >= 1620 {
		deltat[0] = state.deltatAa(jd, tidAcc[0])
		return flags
	}
	deltat[0] = ans / secondsPerDay
	return flags
}

// sweDeltatEx implements swe_deltat_ex from swephlib.c.
func (state *State) sweDeltatEx(jd float64, flags int32, message []byte) float64 {
	var deltat [1]float64
	if state.positions.data.deltaTUserdefIsSet != 0 {
		return state.positions.data.deltaTUserdef
	}
	if message != nil {
		message[0] = byte('\000')
	}
	state.calcDeltat(jd, flags, deltat[:], message)
	return deltat[0]
}

// sweDeltat implements swe_deltat from swephlib.c.
func (state *State) sweDeltat(jd float64) float64 {
	var flags int32 = state.swiGuessEpheFlag()
	return state.sweDeltatEx(jd, flags, nil)
}

// deltatAa implements deltat_aa from swephlib.c.
//
//	The tabulated values of deltaT, in hundredths of a second,
//
// were taken from The Astronomical Almanac 1997etc., pp. K8-K9.
// Some more recent values are taken from IERS
// http://maia.usno.navy.mil/ser7/deltat.data .
// Bessel's interpolation formula is implemented to obtain fourth
// order interpolated values at intermediate times.
// The values are adjusted depending on the ephemeris used
// and its inherent value of secular tidal acceleration ndot.
// Note by Dieter Jan. 2017:
// Bessel interpolation assumes equidistant sampling points. However the
// sampling points are not equidistant, because they are for first January of
// every year and years can have either 365 or 366 days. The interpolation uses
// a step width of 365.25 days. As a consequence, in three out of four years
// the interpolation does not reproduce the exact values of the sampling points
// on the days they refer to.
func (state *State) deltatAa(jd float64, tidAcc float64) float64 {
	var (
		ans         float64 = 0
		ans2        float64 = 0
		ans3        float64
		p           float64
		B           float64
		B2          float64
		Y           float64
		dd          float64
		d           [6]float64
		i           int32
		iy          int32
		k           int32
		tabsiz      int32 = state.initDt()
		tabend      int32 = 1620 + tabsiz - 1
		deltatModel int32 = state.positions.data.astroModels[0]
	)
	if deltatModel == 0 {
		deltatModel = 5
	}
	Y = 2000.0 + (jd-2451544.5)/365.25
	if Y <= float64(tabend) {
		p = math.Floor(Y)
		iy = int32(p - 1620)
		ans = state.time.deltaTTable[iy]
		k = iy + 1
		if k >= tabsiz {
			ans = state.adjustForTidacc(ans, Y, tidAcc, -26.0, 0)
			return ans / secondsPerDay
		}
		p = Y - p
		ans += p * (state.time.deltaTTable[k] - state.time.deltaTTable[iy])
		if iy-1 < 0 || iy+2 >= tabsiz {
			ans = state.adjustForTidacc(ans, Y, tidAcc, -26.0, 0)
			return ans / secondsPerDay
		}
		k = iy - 2
		for i = 0; i < 5; i++ {
			if k < 0 || k+1 >= tabsiz {
				d[i] = 0
			} else {
				d[i] = state.time.deltaTTable[k+1] - state.time.deltaTTable[k]
			}
			k += 1
		}
		for i = 0; i < 4; i++ {
			d[i] = d[i+1] - d[i]
		}
		B = 0.25 * p * (p - 1.0)
		ans += B * (d[1] + d[2])
		if iy+2 >= tabsiz {
			ans = state.adjustForTidacc(ans, Y, tidAcc, -26.0, 0)
			return ans / secondsPerDay
		}
		for i = 0; i < 3; i++ {
			d[i] = d[i+1] - d[i]
		}
		B = 2.0 * B / 3.0
		ans += (p - 0.5) * B * d[1]
		if iy-2 < 0 || iy+3 > tabsiz {
			ans = state.adjustForTidacc(ans, Y, tidAcc, -26.0, 0)
			return ans / secondsPerDay
		}
		for i = 0; i < 2; i++ {
			d[i] = d[i+1] - d[i]
		}
		B = 0.125 * B * (p + 1.0) * (p - 2.0)
		ans += B * (d[0] + d[1])
		ans = state.adjustForTidacc(ans, Y, tidAcc, -26.0, 0)
		return ans / secondsPerDay
	}
	if deltatModel == 5 {
		B = Y - 2000
		if Y < 2500 {
			ans = B*B*B*121.0/30000000.0 + B*B/1250.0 + B*521.0/3000.0 + 64.0
			B2 = float64(tabend - 2000)
			ans2 = B2*B2*B2*121.0/30000000.0 + B2*B2/1250.0 + B2*521.0/3000.0 + 64.0
		} else {
			B = 0.01 * (Y - 2000)
			ans = B*B*32.5 + 42.5
		}
	} else {
		B = 0.01 * (Y - 1820)
		ans = float64(-20) + 31*B*B
		B2 = 0.01 * float64(tabend-1820)
		ans2 = float64(-20) + 31*B2*B2
	}
	if Y <= float64(tabend+100) {
		ans3 = state.time.deltaTTable[tabsiz-1]
		dd = ans2 - ans3
		ans += dd * (Y - float64(tabend+100)) * 0.01
	}
	return ans / secondsPerDay
}

// deltatLongtermMorrisonStephenson implements deltat_longterm_morrison_stephenson from swephlib.c.
func (state *State) deltatLongtermMorrisonStephenson(jd float64) float64 {
	var (
		Ygreg float64 = 2000.0 + (jd-julianDayJ2000)/365.2425
		u     float64 = (Ygreg - 1820) / 100.0
	)
	return float64(-20) + 32*u*u
}

// deltatStephensonMorrison19971600 implements deltat_stephenson_morrison_1997_1600 from swephlib.c.
func (state *State) deltatStephensonMorrison19971600(jd float64, tidAcc float64) float64 {
	var (
		ans  float64 = 0
		ans2 float64
		ans3 float64
		p    float64
		B    float64
		Y    float64
		dd   float64
		iy   int32
	)
	Y = 2000.0 + (jd-julianDayJ2000)/365.25
	if Y < float64(-500) {
		B = (Y - 1735) * 0.01
		ans = float64(-20) + 35*B*B
		ans = state.adjustForTidacc(ans, Y, tidAcc, -26.0, 0)
		if Y >= float64(-600) {
			ans2 = state.adjustForTidacc(float64(mathDt97[0]), float64(-500), tidAcc, -26.0, 0)
			B = -22.350000000000001
			ans3 = float64(-20) + 35*B*B
			ans3 = state.adjustForTidacc(ans3, Y, tidAcc, -26.0, 0)
			dd = ans3 - ans2
			B = (Y - float64(-600)) * 0.01
			ans = ans - dd*B
		}
	}
	if Y >= float64(-500) && Y < 1600 {
		p = math.Floor(Y)
		iy = int32((p - float64(-500)) / 50.0)
		dd = (Y - float64(-500+50*iy)) / 50.0
		ans = float64(mathDt97[iy]) + float64(int32(mathDt97[iy+1])-int32(mathDt97[iy]))*dd
		ans = state.adjustForTidacc(ans, Y, tidAcc, -26.0, 0)
	}
	ans /= secondsPerDay
	return ans
}

// deltatStephensonMorrison20041600 implements deltat_stephenson_morrison_2004_1600 from swephlib.c.
//
//	Stephenson & Morrison (2004)
func (state *State) deltatStephensonMorrison20041600(jd float64, tidAcc float64) float64 {
	var (
		ans       float64 = 0
		ans2      float64
		ans3      float64
		p         float64
		B         float64
		dd        float64
		initialJD float64
		iy        int32
		Y         float64 = 2000.0 + (jd-julianDayJ2000)/365.2425
	)
	if Y < float64(-1000) {
		ans = state.deltatLongtermMorrisonStephenson(jd)
		ans = state.adjustForTidacc(ans, Y, tidAcc, -26.0, 0)
		if Y >= float64(-1100) {
			ans2 = state.adjustForTidacc(float64(mathDt2[0]), float64(-1000), tidAcc, -26.0, 0)
			initialJD = 1355817.5
			ans3 = state.deltatLongtermMorrisonStephenson(initialJD)
			ans3 = state.adjustForTidacc(ans3, Y, tidAcc, -26.0, 0)
			dd = ans3 - ans2
			B = (Y - float64(-1100)) * 0.01
			ans = ans - dd*B
		}
	}
	if Y >= float64(-1000) && Y < 1600 {
		var Yjul float64 = 2000 + (jd-2451557.5)/365.25
		p = math.Floor(Yjul)
		iy = int32((p - float64(-1000)) / 100)
		dd = (Yjul - float64(-1000+100*iy)) / 100
		ans = float64(mathDt2[iy]) + float64(int32(mathDt2[iy+1])-int32(mathDt2[iy]))*dd
		ans = state.adjustForTidacc(ans, Y, tidAcc, -26.0, 0)
	}
	ans /= secondsPerDay
	return ans
}

// deltatStephensonEtc2016 implements deltat_stephenson_etc_2016 from swephlib.c.
//
//	ybeg=2013, yend=2016
func (state *State) deltatStephensonEtc2016(jd float64, tidAcc float64) float64 {
	var (
		t     float64
		dt    float64
		Ygreg float64
		i     int32
		irec  int32 = -1
	)
	Ygreg = 2000.0 + (jd-julianDayJ2000)/365.2425
	for i = 0; i < 54; i++ {
		if jd < mathDtcf16[i][0] {
			break
		}
		if jd < mathDtcf16[i][1] {
			irec = i
			break
		}
	}
	if irec >= 0 {
		t = (jd - mathDtcf16[irec][0]) / (mathDtcf16[irec][1] - mathDtcf16[irec][0])
		dt = mathDtcf16[irec][2] + mathDtcf16[irec][3]*t + mathDtcf16[irec][4]*t*t + mathDtcf16[irec][5]*t*t*t
	} else if Ygreg < float64(-720) {
		t = (Ygreg - 1825) / 100.0
		dt = float64(-320) + 32.5*t*t
		dt -= 179.7337208
	} else {
		t = (Ygreg - 1825) / 100.0
		dt = float64(-320) + 32.5*t*t
		dt += 269.4790417
	}
	dt = state.adjustForTidacc(dt, Ygreg, tidAcc, -25.85, 1)
	dt /= secondsPerDay
	return dt
}

// deltatEspenakMeeus1620 implements deltat_espenak_meeus_1620 from swephlib.c.
//
//	The parameter adjust_after_1955 must be TRUE here, because the
//
// Stephenson 2016 curve is based on occultation data alone,
// not on IERS data.
// Note, however, the current function deltat_stephenson_etc_2016()
// is called only for dates before 1 Jan 1955.
func (state *State) deltatEspenakMeeus1620(jd float64, tidAcc float64) float64 {
	var (
		ans   float64 = 0
		Ygreg float64
		u     float64
	)
	Ygreg = 2000.0 + (jd-julianDayJ2000)/365.2425
	if Ygreg < float64(-500) {
		ans = state.deltatLongtermMorrisonStephenson(jd)
	} else if Ygreg < 500 {
		u = Ygreg / 100.0
		ans = (((((0.0090316521*u+0.022174192)*u-0.1798452)*u-5.952053)*u+33.78311)*u-1014.41)*u + 10583.6
	} else if Ygreg < 1600 {
		u = (Ygreg - 1000) / 100.0
		ans = (((((0.0083572073*u-0.005050998)*u-0.8503463)*u+0.319781)*u+71.23472)*u-556.01)*u + 1574.2
	} else if Ygreg < 1700 {
		u = Ygreg - 1600
		ans = 120 - 0.9808*u - 0.01532*u*u + u*u*u/7129.0
	} else if Ygreg < 1800 {
		u = Ygreg - 1700
		ans = (((-u/1174000.0+0.00013336)*u-0.0059285)*u+0.1603)*u + 8.83
	} else if Ygreg < 1860 {
		u = Ygreg - 1800
		ans = ((((((0.000000000875*u-0.0000001699)*u+0.0000121272)*u-0.00037436)*u+0.0041116)*u+0.0068612)*u-0.332447)*u + 13.72
	} else if Ygreg < 1900 {
		u = Ygreg - 1860
		ans = ((((u/233174.0-0.0004473624)*u+0.01680668)*u-0.251754)*u+0.5737)*u + 7.62
	} else if Ygreg < 1920 {
		u = Ygreg - 1900
		ans = (((-0.000197*u+0.0061966)*u-0.0598939)*u+1.494119)*u - 2.79
	} else if Ygreg < 1941 {
		u = Ygreg - 1920
		ans = 21.20 + 0.84493*u - 0.076100*u*u + 0.0020936*u*u*u
	} else if Ygreg < 1961 {
		u = Ygreg - 1950
		ans = 29.07 + 0.407*u - u*u/233.0 + u*u*u/2547.0
	} else if Ygreg < 1986 {
		u = Ygreg - 1975
		ans = 45.45 + 1.067*u - u*u/260.0 - u*u*u/718.0
	} else if Ygreg < 2005 {
		u = Ygreg - 2000
		ans = ((((0.00002373599*u+0.000651814)*u+0.0017275)*u-0.060374)*u+0.3345)*u + 63.86
	}
	ans = state.adjustForTidacc(ans, Ygreg, tidAcc, -26.0, 0)
	ans /= secondsPerDay
	return ans
}

// initDt implements init_dt from swephlib.c.
//
//	Read delta t values from external file.
//
// record structure: year(whitespace)delta_t in 0.01 sec.
func (state *State) initDt() int32 {
	var (
		fp       *textFile
		year     int32
		tabIndex int32
		tabsiz   int32
		i        int32
		s        [256]byte
		sp       []byte
	)
	if !(state.positions.data.initDtDone != 0) {
		state.positions.data.initDtDone = 1
		fp = state.swiFopen(-1, textBytes("swe_deltat.txt"), state.positions.data.ephemerisPath[:], nil)
		if fp == nil {
			fp = state.swiFopen(-1, textBytes("sedeltat.txt"), state.positions.data.ephemerisPath[:], nil)
		}
		if fp == nil {
			return 2028 - 1620 + 1
		}
		for readLine(s[:], 256, fp) != nil {
			for sp = s[:]; findByte(textBytes(" \t"), sp[0]) != nil && int32(sp[0]) != '\000'; {
				sp = sp[1:]
			}
			if int32(sp[0]) == '#' || int32(sp[0]) == '\n' {
				continue
			}
			year = parseInt(s[:])
			tabIndex = year - 1620
			if tabIndex >= 2028-1620+1+100 {
				continue
			}
			for sp = sp[4:]; findByte(textBytes(" \t"), sp[0]) != nil && int32(sp[0]) != '\000'; {
				sp = sp[1:]
			}
			state.time.deltaTTable[tabIndex] = parseFloat(sp)
		}
		closeTextFile(fp)
	}
	tabsiz = 2001 - 1620 + 1
	for i = tabsiz - 1; i < 2028-1620+1+100; i++ {
		if state.time.deltaTTable[i] == 0 {
			break
		} else {
			tabsiz++
		}
	}
	tabsiz--
	return tabsiz
}

// adjustForTidacc implements adjust_for_tidacc from swephlib.c.
func (state *State) adjustForTidacc(ans float64, Y float64, tidAcc float64, tidAcc0 float64, adjustAfter1955 int32) float64 {
	var B float64
	if Y < 1955.0 || adjustAfter1955 != 0 {
		B = Y - 1955.0
		ans += -0.000091 * (tidAcc - tidAcc0) * B * B
	}
	return ans
}

// sweGetTidAcc implements swe_get_tid_acc from swephlib.c.
//
//	returns tidal acceleration used in swe_deltat() and swe_deltat_ex()
func (state *State) sweGetTidAcc() float64 {
	return state.positions.data.tidAcc
}

// sweSetTidAcc implements swe_set_tid_acc from swephlib.c.
//
//	function sets tidal acceleration of the Moon.
//
// t_acc can be either
//   - the value of the tidal acceleration in arcsec/cty^2
//     of the Moon will be set consistent with that ephemeris.
//   - SE_TIDAL_AUTOMATIC,
func (state *State) sweSetTidAcc(tAcc float64) {
	if tAcc == 999999 {
		state.positions.data.tidAcc = -25.80
		state.positions.data.isTidAccManual = 0
		return
	}
	state.positions.data.tidAcc = tAcc
	state.positions.data.isTidAccManual = 1
}

// sweSetDeltaTUserdef implements swe_set_delta_t_userdef from swephlib.c.
func (state *State) sweSetDeltaTUserdef(dt float64) {
	if dt == -1e-10 {
		state.positions.data.deltaTUserdefIsSet = 0
	} else {
		state.positions.data.deltaTUserdefIsSet = 1
		state.positions.data.deltaTUserdef = dt
	}
}

// swiGuessEpheFlag implements swi_guess_ephe_flag from swephlib.c.
func (state *State) swiGuessEpheFlag() int32 {
	var flags int32 = 2
	if state.positions.data.jplFileIsOpen != 0 {
		flags = flagJPL
	} else {
		flags = flagSwiss
	}
	return flags
}

// swiGetTidAcc implements swi_get_tid_acc from swephlib.c.
//
//	if jpl file is open, assume SEFLG_JPLEPH
func (state *State) swiGetTidAcc(jdUT float64, flags int32, denum int32, denumret []int32, tidAcc []float64, message []byte) int32 {
	flags &= flagJPL | flagSwiss | flagMoshier
	if state.positions.data.isTidAccManual != 0 {
		tidAcc[0] = state.positions.data.tidAcc
		return flags
	}
	if denum == 0 {
		if flags&flagMoshier != 0 {
			tidAcc[0] = -25.580
			denumret[0] = 404
			return flags
		}
		if flags&flagJPL != 0 {
			if state.positions.data.jplFileIsOpen != 0 {
				denum = state.positions.data.jplDENumber
			}
		}
		if flags&flagSwiss != 0 {
			if state.positions.data.files[1].handle != nil {
				denum = state.positions.data.files[1].deNumber
			}
		}
	}
	switch denum {
	case 200:
		tidAcc[0] = -23.8946
	case 403:
		tidAcc[0] = -25.580
	case 404:
		tidAcc[0] = -25.580
	case 405:
		tidAcc[0] = -25.826
	case 406:
		tidAcc[0] = -25.826
	case 421:
		tidAcc[0] = -25.85
	case 422:
		tidAcc[0] = -25.85
	case 430:
		tidAcc[0] = -25.82
	case 431:
		tidAcc[0] = -25.80
	case 440:
		tidAcc[0] = -25.936
	case 441:
		tidAcc[0] = -25.936
	default:
		denum = 431
		tidAcc[0] = -25.80
	}
	denumret[0] = denum
	flags &= flagJPL | flagSwiss | flagMoshier
	return flags
}

// swiSetTidAcc implements swi_set_tid_acc from swephlib.c.
func (state *State) swiSetTidAcc(jdUT float64, flags int32, denum int32, message []byte) int32 {
	var (
		status   int32 = flags
		denumret [1]int32
	)
	if state.positions.data.isTidAccManual != 0 {
		return status
	}
	output4 := []float64{state.positions.data.tidAcc}
	result := state.swiGetTidAcc(jdUT, flags, denum, denumret[:], output4, message)
	state.positions.data.tidAcc = output4[0]
	status = result
	return status
}

// sidtimeLongTerm implements sidtime_long_term from swephlib.c.
//
// #define SIDT_LTERM   TRUE
// #if SIDT_LTERM
func (state *State) sidtimeLongTerm(jdUT float64, eps float64, nut float64) float64 {
	var (
		tsid  float64 = 0
		jdTT  float64
		dlon  float64
		xs    [6]float64
		xobl  [6]float64
		dhour float64
		nutlo [2]float64
		dlt   float64 = 0.0057755183314369954
		t     float64
		t2    float64
		t3    float64
	)
	jdTT = jdUT + state.sweDeltatEx(jdUT, -1, nil)
	t = (jdTT - julianDayJ2000) / 365250.0
	t2 = t * t
	t3 = t * t2
	dlon = 100.46645683 + (1295977422.83429*t-2.04411*t2-0.00523*t3)/3600.0
	dlon = state.sweDegnorm(dlon - dlt*360.0/365.2425)
	xs[0] = dlon * radiansPerDegree
	xs[1] = 0
	xs[2] = 1
	xobl[0] = 23.45
	xobl[1] = 23.45
	xobl[1] = state.swiEpsiln(julianDayJ2000+state.sweDeltatEx(julianDayJ2000, -1, nil), 0) * degreesPerRadian
	state.swiPolcart(xs[:], xs[:])
	state.swiCoortrf(xs[:], xs[:], -xobl[1]*radiansPerDegree)
	state.swiPrecess(xs[:], jdTT, 0, -1)
	xobl[1] = state.swiEpsiln(jdTT, 0) * degreesPerRadian
	state.swiNutation(jdTT, 0, nutlo[:])
	xobl[0] = xobl[1] + nutlo[1]*degreesPerRadian
	xobl[2] = nutlo[0] * degreesPerRadian
	state.swiCoortrf(xs[:], xs[:], xobl[1]*radiansPerDegree)
	state.swiCartpol(xs[:], xs[:])
	xs[0] *= degreesPerRadian
	dhour = math.Mod(jdUT-0.5, 1) * 360
	if eps == 0 {
		xs[0] += xobl[2] * math.Cos(xobl[0]*radiansPerDegree)
	} else {
		xs[0] += nut * math.Cos(eps*radiansPerDegree)
	}
	xs[0] = state.sweDegnorm(xs[0] + dhour)
	tsid = xs[0] / 15
	return tsid
}

// sidtimeNonPolynomialPart implements sidtime_non_polynomial_part from swephlib.c.
func (state *State) sidtimeNonPolynomialPart(tt float64) float64 {
	var (
		i    int32
		j    int32
		delm [14]float64
		dadd float64
		darg float64
	)
	delm[0] = state.sweRadnorm(2.35555598 + 8328.6914269554*tt)
	delm[1] = state.sweRadnorm(6.24006013 + 628.301955*tt)
	delm[2] = state.sweRadnorm(1.627905234 + 8433.466158131*tt)
	delm[3] = state.sweRadnorm(5.198466741 + 7771.3771468121*tt)
	delm[4] = state.sweRadnorm(2.18243920 - 33.757045*tt)
	delm[5] = state.sweRadnorm(4.402608842 + 2608.7903141574*tt)
	delm[6] = state.sweRadnorm(3.176146697 + 1021.3285546211*tt)
	delm[7] = state.sweRadnorm(1.753470314 + 628.3075849991*tt)
	delm[8] = state.sweRadnorm(6.203480913 + 334.0612426700*tt)
	delm[9] = state.sweRadnorm(0.599546497 + 52.9690962641*tt)
	delm[10] = state.sweRadnorm(0.874016757 + 21.3299104960*tt)
	delm[11] = state.sweRadnorm(5.481293871 + 7.4781598567*tt)
	delm[12] = state.sweRadnorm(5.321159000 + 3.8127774000*tt)
	delm[13] = (0.02438175 + 0.00000538691*tt) * tt
	dadd = -0.87 * math.Sin(delm[4]) * tt
	for i = 0; i < 33; i++ {
		darg = 0
		for j = 0; j < 14; j++ {
			darg += float64(mathStfarg[i*14+j]) * delm[j]
		}
		dadd += mathStcf[i*2]*math.Sin(darg) + mathStcf[i*2+1]*math.Cos(darg)
	}
	dadd /= 3600000000
	return dadd
}

// sweSidtime0 implements swe_sidtime0 from swephlib.c.
//
//	1 Jan 2050
func (state *State) sweSidtime0(tjd float64, eps float64, nut float64) float64 {
	var (
		jd0            float64
		secs           float64
		eqeq           float64
		jd             float64
		tu             float64
		tt             float64
		msday          float64
		jdrel          float64
		gmst           float64
		dadd           float64
		precModelShort int32 = state.positions.data.astroModels[2]
		sidtModel      int32 = state.positions.data.astroModels[7]
	)
	if precModelShort == 0 {
		precModelShort = 9
	}
	if sidtModel == 0 {
		sidtModel = 4
	}
	state.swiInitSwedIfStart()
	if sidtModel == 4 {
		if tjd <= 2396758.5 || tjd >= 2469807.5 {
			gmst = state.sidtimeLongTerm(tjd, eps, nut)
			if tjd <= 2396758.5 {
				gmst -= 2.5211466666666667e-05
			} else if tjd >= 2469807.5 {
				gmst -= 9.2376399999999999e-05
			}
			if gmst >= 24 {
				gmst -= 24
			}
			if gmst < 0 {
				gmst += 24
			}
			return gmst
		}
	}
	jd = tjd
	jd0 = math.Floor(jd)
	secs = tjd - jd0
	if secs < 0.5 {
		jd0 -= 0.5
		secs += 0.5
	} else {
		jd0 += 0.5
		secs -= 0.5
	}
	secs *= secondsPerDay
	tu = (jd0 - julianDayJ2000) / daysPerCentury
	if sidtModel == 3 || sidtModel == 4 {
		jdrel = tjd - julianDayJ2000
		tt = (tjd + state.sweDeltatEx(tjd, -1, nil) - julianDayJ2000) / daysPerCentury
		gmst = state.sweDegnorm((0.7790572732640 + 1.00273781191135448*jdrel) * 360)
		gmst += (0.014506 + tt*(4612.156534+tt*(1.3915817+tt*(-0.00000044+tt*(-0.000029956+tt*-0.0000000368))))) / 3600.0
		dadd = state.sidtimeNonPolynomialPart(tt)
		gmst = state.sweDegnorm(gmst + dadd)
		gmst = gmst / 15.0 * 3600.0
	} else if sidtModel == 2 {
		tt = (jd0 + state.sweDeltatEx(jd0, -1, nil) - julianDayJ2000) / daysPerCentury
		gmst = (((-0.000000002454*tt-0.00000199708)*tt-0.0000002926)*tt+0.092772110)*tt*tt + 307.4771013*(tt-tu) + 8640184.79447825*tu + 24110.5493771
		msday = 1 + ((((-0.000000012270*tt-0.00000798832)*tt-0.0000008778)*tt+0.185544220)*tt+8640184.79447825)/3155760000
		gmst += msday * secs
	} else {
		gmst = ((-6.2e-6*tu+9.3104e-2)*tu+8640184.812866)*tu + 24110.54841
		msday = 1.0 + ((-1.86e-5*tu+0.186208)*tu+8640184.812866)/3155760000
		gmst += msday * secs
	}
	eqeq = 240.0 * nut * math.Cos(eps*radiansPerDegree)
	gmst = gmst + eqeq
	gmst = gmst - secondsPerDay*math.Floor(gmst/secondsPerDay)
	gmst /= 3600
	return gmst
}

// sweSetInterpolateNut implements swe_set_interpolate_nut from swephlib.c.
func (state *State) sweSetInterpolateNut(doInterpolate int32) {
	if state.positions.data.doInterpolateNut == doInterpolate {
		return
	}
	if doInterpolate != 0 {
		state.positions.data.doInterpolateNut = 1
	} else {
		state.positions.data.doInterpolateNut = 0
	}
	state.positions.data.interpol.tjdNut0 = 0
	state.positions.data.interpol.tjdNut2 = 0
	state.positions.data.interpol.nutDpsi0 = 0
	state.positions.data.interpol.nutDpsi1 = 0
	state.positions.data.interpol.nutDpsi2 = 0
	state.positions.data.interpol.nutDeps0 = 0
	state.positions.data.interpol.nutDeps1 = 0
	state.positions.data.interpol.nutDeps2 = 0
}

// sweSidtime implements swe_sidtime from swephlib.c.
//
//	sidereal time, without eps and nut as parameters.
//
// tjd must be UT !!!
// for more informsation, see comment with swe_sidtime0()
func (state *State) sweSidtime(jdUT float64) float64 {
	var (
		i     int32
		eps   float64
		nutlo [2]float64
		tsid  float64
		tjde  float64
	)
	tjde = jdUT + state.sweDeltatEx(jdUT, -1, nil)
	state.swiInitSwedIfStart()
	eps = state.swiEpsiln(tjde, 0) * degreesPerRadian
	state.swiNutation(tjde, 0, nutlo[:])
	for i = 0; i < 2; i++ {
		nutlo[i] *= degreesPerRadian
	}
	tsid = state.sweSidtime0(jdUT, eps+nutlo[1], nutlo[0])
	return tsid
}

// swiGenFilename implements swi_gen_filename from swephlib.c.
//
//	SWISSEPH
//
// generates name of ephemeris file
// file name looks as follows:
// swephpl.m30, where
//
// "sweph"              	"swiss ephemeris"
// "pl","mo","as"               planet, moon, or asteroid
// "m"  or "_"                  BC or AD
//
// "30"                         start century
// tjd        	= ephemeris file for which julian day
// ipli       	= number of planet
// fname      	= ephemeris file name
func (state *State) swiGenFilename(jd float64, bodyIndex int32, fname []byte) {
	var (
		icty     int32
		ncties   int32 = 6
		calendar int16
		jmon     [1]int32
		jday     [1]int32
		jyear    [1]int32
		sgn      int32
		jut      [1]float64
		sform    []byte
	)
	switch bodyIndex {
	case 1:
		copyText(fname, textBytes("semo"))
	case 0:
		fallthrough
	case 2:
		fallthrough
	case 3:
		fallthrough
	case 4:
		fallthrough
	case 5:
		fallthrough
	case 6:
		fallthrough
	case 7:
		fallthrough
	case 8:
		fallthrough
	case 9:
		fallthrough
	case 10:
		copyText(fname, textBytes("sepl"))
	case 14:
		fallthrough
	case 15:
		fallthrough
	case 16:
		fallthrough
	case 17:
		fallthrough
	case 12:
		fallthrough
	case 13:
		copyText(fname, textBytes("seas"))
	default:
		if bodyIndex > 9000 && bodyIndex < 10000 {
			formatText(fname, textBytes("sat%ssepm%d.%s"), textBytes("/"), bodyIndex, textBytes("se1"))
		} else {
			sform = textBytes("ast%d%sse%05d.%s")
			if bodyIndex-10000 > 99999 {
				sform = textBytes("ast%d%ss%06d.%s")
			}
			formatText(fname, sform, (bodyIndex-10000)/1000, textBytes("/"), bodyIndex-10000, textBytes("se1"))
		}
		return
	}
	if jd >= 2305447.5 {
		calendar = 1
		state.sweRevjul(jd, int32(calendar), jyear[:], jmon[:], jday[:], jut[:])
	} else {
		calendar = 0
		state.sweRevjul(jd, int32(calendar), jyear[:], jmon[:], jday[:], jut[:])
	}
	if jyear[0] < 0 {
		sgn = -1
	} else {
		sgn = 1
	}
	icty = jyear[0] / 100
	if sgn < 0 && jyear[0]%100 != 0 {
		icty -= 1
	}
	for icty%ncties != 0 {
		icty--
	}
	if icty < 0 {
		appendText(fname, textBytes("m"))
	} else {
		appendText(fname, textBytes("_"))
	}
	icty = absInt(icty)
	formatText(fname[textLength(fname):], textBytes("%02d.%s"), icty, textBytes("se1"))
}

// swiCutstr implements swi_cutstr from swephlib.c.
func (state *State) swiCutstr(s []byte, cutlist []byte, cpos [][]byte, nmax int32) int32 {
	var n int32 = 1
	cpos[0] = s
	for int32(s[0]) != '\000' {
		if findByte(cutlist, int32(s[0])) != nil && n < nmax {
			s[0] = byte('\000')
			for int32(s[1]) != '\000' && findByte(cutlist, int32(s[1])) != nil {
				s = s[1:]
			}
			cpos[n] = s[1:]
			n++
		}
		if int32(s[0]) == '\n' || int32(s[0]) == '\r' {
			s[0] = byte('\000')
			break
		}
		s = s[1:]
	}
	if n < nmax {
		cpos[n] = nil
	}
	return n
}

// swiCrc32 implements swi_crc32 from swephlib.c.
//
//	Initialized first time "crc32()" is called. If you prefer, you can
//
// statically initialize it at compile time. [Another exercise.]
func (state *State) swiCrc32(buf []byte, len int32) uint32 {
	var (
		p   []byte
		crc uint32
	)
	if !(state.time.crcTable[1] != 0) {
		state.initCrc32()
	}
	crc = 0xffffffff
	for p = buf; len > 0; {
		crc = crc<<8 ^ state.time.crcTable[crc>>24^uint32(p[0])]
		p = p[1:]
		len--
	}
	return ^crc
}

// initCrc32 implements init_crc32 from swephlib.c.
//
//	AUTODIN II, Ethernet, & FDDI
func (state *State) initCrc32() {
	var (
		j int32
		c uint32
		i uint32
	)
	for i = 0; i < 256; i++ {
		c = i << 24
		for j = 8; j > 0; j-- {
			if c&0x80000000 != 0 {
				c = c<<1 ^ uint32(0x04c11db7)
			} else {
				c <<= 1
			}
		}
		state.time.crcTable[i] = c
	}
}

// sweCsnorm implements swe_csnorm from swephlib.c.
func (state *State) sweCsnorm(p int32) int32 {
	if p < 0 {
		for first := true; first || p < 0; first = false {
			p += 360 * 360000
		}
	} else if p >= 360*360000 {
		for first := true; first || p >= 360*360000; first = false {
			p -= 360 * 360000
		}
	}
	return p
}

// sweDifcsn implements swe_difcsn from swephlib.c.
func (state *State) sweDifcsn(p1 int32, p2 int32) int32 {
	return state.sweCsnorm(p1 - p2)
}

// sweDifdegn implements swe_difdegn from swephlib.c.
func (state *State) sweDifdegn(p1 float64, p2 float64) float64 {
	return state.sweDegnorm(p1 - p2)
}

// sweDifcs2n implements swe_difcs2n from swephlib.c.
func (state *State) sweDifcs2n(p1 int32, p2 int32) int32 {
	var dif int32
	dif = state.sweCsnorm(p1 - p2)
	if dif >= 180*360000 {
		return dif - 360*360000
	}
	return dif
}

// sweDifdeg2n implements swe_difdeg2n from swephlib.c.
func (state *State) sweDifdeg2n(p1 float64, p2 float64) float64 {
	var dif float64
	dif = state.sweDegnorm(p1 - p2)
	if dif >= 180.0 {
		return dif - 360.0
	}
	return dif
}

// sweDifrad2n implements swe_difrad2n from swephlib.c.
func (state *State) sweDifrad2n(p1 float64, p2 float64) float64 {
	var dif float64
	dif = state.sweRadnorm(p1 - p2)
	if dif >= 3.1415926535897931 {
		return dif - 6.2831853071795862
	}
	return dif
}

// sweCsroundsec implements swe_csroundsec from swephlib.c.
func (state *State) sweCsroundsec(x int32) int32 {
	var t int32
	t = int32(int64((x+50)/100) * 100)
	if t > x && t%(30*360000) == 0 {
		t = int32(int64(x/100) * 100)
	}
	return t
}

// sweD2l implements swe_d2l from swephlib.c.
func (state *State) sweD2l(x float64) int32 {
	if x >= 0 {
		return int32(x + 0.5)
	} else {
		return -int32(0.5 - x)
	}
}

// sweDayOfWeek implements swe_day_of_week from swephlib.c.
//
// monday = 0, ... sunday = 6
func (state *State) sweDayOfWeek(jd float64) int32 {
	return (int32(math.Floor(jd-2433282-1.5))%7 + 7) % 7
}

// sweCs2timestr implements swe_cs2timestr from swephlib.c.
func (state *State) sweCs2timestr(t int32, sep int32, suppressZero int32, a []byte) []byte {
	var (
		h int32
		m int32
		s int32
	)
	copyText(a, textBytes("        "))
	a[5] = byte(sep)
	a[2] = a[5]
	t = int32(int64((t+50)/100) % (24 * 3600))
	s = int32(int64(t) % 60)
	m = int32(int64(t/60) % 60)
	h = int32(int64(t/3600) % 100)
	if s == 0 && suppressZero != 0 {
		a[5] = byte('\000')
	} else {
		a[6] = byte(s/10 + '0')
		a[7] = byte(s%10 + '0')
	}
	a[0] = byte(h/10 + '0')
	a[1] = byte(h%10 + '0')
	a[3] = byte(m/10 + '0')
	a[4] = byte(m%10 + '0')
	return a
}

// sweCs2lonlatstr implements swe_cs2lonlatstr from swephlib.c.
//
//	swe_cs2timestr()
func (state *State) sweCs2lonlatstr(t int32, pchar byte, mchar byte, sp []byte) []byte {
	var (
		a  [10]byte
		aa []byte
		h  int32
		m  int32
		s  int32
	)
	copyText(a[:], textBytes("      '  "))
	if t < 0 {
		pchar = mchar
	}
	t = (absInt(t) + 50) / 100
	s = int32(int64(t) % 60)
	m = int32(int64(t/60) % 60)
	h = int32(int64(t/3600) % 1000)
	if s == 0 {
		a[6] = byte('\000')
	} else {
		a[7] = byte(s/10 + '0')
		a[8] = byte(s%10 + '0')
	}
	a[3] = pchar
	if h > 99 {
		a[0] = byte(h/100 + '0')
	}
	if h > 9 {
		a[1] = byte(h%100/10 + '0')
	}
	a[2] = byte(h%10 + '0')
	a[4] = byte(m/10 + '0')
	a[5] = byte(m%10 + '0')
	for aa = a[:]; int32(aa[0]) == ' '; {
		aa = aa[1:]
	}
	copyText(sp, aa)
	return sp
}

// sweCs2degstr implements swe_cs2degstr from swephlib.c.
func (state *State) sweCs2degstr(t int32, a []byte) []byte {
	var (
		h int32
		m int32
		s int32
	)
	t = int32(int64(t/100) % (30 * 3600))
	s = int32(int64(t) % 60)
	m = int32(int64(t/60) % 60)
	h = int32(int64(t/3600) % 100)
	formatText(a, textBytes("%2d%s%02d'%02d"), h, textBytes("°"), m, s)
	return a
}

// splitDegNakshatra implements split_deg_nakshatra from swephlib.c.
func (state *State) splitDegNakshatra(ddeg float64, roundflag int32, ideg []int32, imin []int32, isec []int32, dsecfr []float64, inak []int32) {
	var (
		dadd       float64 = 0
		dnakshsize float64 = 13.33333333333333
		ddeghelp   float64 = math.Mod(ddeg, dnakshsize)
	)
	inak[0] = 1
	if ddeg < 0 {
		inak[0] = -1
		ddeg = 0
	}
	if state.positions.data.sidereal.mode&39 == 39 {
		ddeg = state.sweDegnorm(ddeg + 3.33333333333333)
	}
	if roundflag&4 != 0 {
		dadd = 0.5
	} else if roundflag&2 != 0 {
		dadd = 0.0083333333333333332
	} else if roundflag&1 != 0 {
		dadd = 0.00013888888888888889
	}
	if roundflag&32 != 0 {
		if int32(ddeghelp+dadd)-int32(ddeghelp) > 0 {
			dadd = 0
		}
	} else if roundflag&16 != 0 {
		if ddeghelp+dadd >= dnakshsize {
			dadd = 0
		}
	}
	ddeg += dadd
	inak[0] = int32(ddeg / dnakshsize)
	if inak[0] == 27 {
		inak[0] = 0
	}
	ddeg = math.Mod(ddeg, dnakshsize)
	ideg[0] = int32(ddeg)
	ddeg -= float64(ideg[0])
	imin[0] = int32(ddeg * 60)
	ddeg -= float64(imin[0]) / 60.0
	isec[0] = int32(ddeg * 3600)
	if !(roundflag&(4|2|1) != 0) {
		dsecfr[0] = ddeg*3600 - float64(isec[0])
	} else {
		dsecfr[0] = float64(isec[0])
	}
}

// sweSplitDeg implements swe_split_deg from swephlib.c.
func (state *State) sweSplitDeg(ddeg float64, roundflag int32, ideg []int32, imin []int32, isec []int32, dsecfr []float64, isgn []int32) {
	var dadd float64 = 0
	isgn[0] = 1
	if ddeg < 0 {
		isgn[0] = -1
		ddeg = -ddeg
	} else if roundflag&1024 != 0 {
		state.splitDegNakshatra(ddeg, roundflag, ideg, imin, isec, dsecfr, isgn)
		return
	}
	if roundflag&4 != 0 {
		dadd = 0.5
	} else if roundflag&2 != 0 {
		dadd = 0.0083333333333333332
	} else if roundflag&1 != 0 {
		dadd = 0.00013888888888888889
	}
	if roundflag&32 != 0 {
		if int32(ddeg+dadd)-int32(ddeg) > 0 {
			dadd = 0
		}
	} else if roundflag&16 != 0 {
		if math.Mod(ddeg, 30)+dadd >= 30 {
			dadd = 0
		}
	}
	ddeg += dadd
	if roundflag&8 != 0 {
		isgn[0] = int32(ddeg / 30)
		if isgn[0] == 12 {
			isgn[0] = 0
		}
		ddeg = math.Mod(ddeg, 30)
	}
	ideg[0] = int32(ddeg)
	ddeg -= float64(ideg[0])
	imin[0] = int32(ddeg * 60)
	ddeg -= float64(imin[0]) / 60.0
	isec[0] = int32(ddeg * 3600)
	if !(roundflag&(4|2|1) != 0) {
		dsecfr[0] = ddeg*3600 - float64(isec[0])
	} else {
		dsecfr[0] = float64(isec[0])
	}
}

// swiKepler implements swi_kepler from swephlib.c.
//
//	end split_deg
func (state *State) swiKepler(E float64, M float64, ecce float64) float64 {
	var (
		dE float64 = 1
		E0 float64
		x  float64
	)
	if ecce < 0.4 {
		for dE > 1e-12 {
			E0 = E
			E = M + ecce*math.Sin(E0)
			dE = math.Abs(E - E0)
		}
	} else {
		for dE > 1e-12 {
			E0 = E
			x = (M + ecce*math.Sin(E0) - E0) / (1 - ecce*math.Cos(E0))
			dE = math.Abs(x)
			if dE < 1e-2 {
				E = E0 + x
			} else {
				E = state.swiMod2PI(E0 + x)
				dE = math.Abs(E - E0)
			}
		}
	}
	return E
}

// swiFK4FK5 implements swi_FK4_FK5 from swephlib.c.
func (state *State) swiFK4FK5(xp []float64, jd float64) {
	var correctSpeed int32 = 1
	if xp[0] == 0 && xp[1] == 0 && xp[2] == 0 {
		return
	}
	if xp[3] == 0 {
		correctSpeed = 0
	}
	state.swiCartpolSp(xp, xp)
	xp[0] += (0.035 + 0.085*(jd-2433282.42345905)/36524.2198782) / 3600 * 15 * radiansPerDegree
	if correctSpeed != 0 {
		xp[3] += 1.6924042333443584e-10
	}
	state.swiPolcartSp(xp, xp)
}

// swiFK5FK4 implements swi_FK5_FK4 from swephlib.c.
func (state *State) swiFK5FK4(xp []float64, jd float64) {
	if xp[0] == 0 && xp[1] == 0 && xp[2] == 0 {
		return
	}
	state.swiCartpolSp(xp, xp)
	xp[0] -= (0.035 + 0.085*(jd-2433282.42345905)/36524.2198782) / 3600 * 15 * radiansPerDegree
	xp[3] -= 1.6924042333443584e-10
	state.swiPolcartSp(xp, xp)
}

// setAstroModels implements set_astro_models from swephlib.c.
//
//	function for inhouse testing only
func (state *State) setAstroModels(samod []byte) {
	var (
		pmodel []int32 = state.positions.data.astroModels[:]
		sp     []byte
		sp2    []byte
		i      int32 = 0
	)
	state.swiInitSwedIfStart()
	sp = samod
	pmodel[0] = parseInt(sp)
	i++
	for {
		sp2 = findByte(sp, ',')
		if !(sp2 != nil && i < 8) {
			break
		}
		sp = sp2[1:]
		pmodel[i] = parseInt(sp)
		i++
	}
}

// sweSetAstroModels implements swe_set_astro_models from swephlib.c.
//
//	note sid. time (S)!
func (state *State) sweSetAstroModels(samod []byte, flags int32) {
	var (
		dversion float64
		s        [30]byte
		sp       []byte
	)
	state.swiInitSwedIfStart()
	if int32(samod[0]) != '\000' && isDigit(int32(samod[0])) != 0 {
		state.setAstroModels(samod)
	} else if int32(samod[0]) == '\000' || compareTextN(samod, textBytes("SE"), 2) == 0 {
		copyTextN(s[:], samod, 20)
		s[20] = byte('\000')
		if sp = findByte(s[5:], '.'); sp != nil {
			state.swiStrcpy(sp, sp[1:])
		}
		if sp = findByte(s[5:], 'b'); sp != nil {
			state.swiStrcpy(sp, sp[1:])
		}
		dversion = parseFloat(s[2:])
		if dversion == 0 {
			dversion = parseFloat(textBytes("2.10.03"))
		}
		if dversion >= 2.06 {
			state.setAstroModels(textBytes("5,9,9,4,3,0,0,4"))
		} else if dversion >= 2.01 {
			state.setAstroModels(textBytes("4,9,9,4,3,0,0,4"))
		} else if dversion >= 2.00 {
			state.setAstroModels(textBytes("4,9,9,4,3,0,0,4"))
			if state.swiGetDenum(0, flags) == 431 {
				state.sweSetTidAcc(-25.826)
			}
		} else if dversion >= 1.80 {
			state.setAstroModels(textBytes("4,9,9,4,3,0,0,1"))
			state.sweSetTidAcc(-25.826)
		} else if dversion >= 1.78 {
			state.setAstroModels(textBytes("4,9,9,4,2,0,0,2"))
			state.sweSetTidAcc(-25.826)
		} else if dversion >= 1.77 {
			state.setAstroModels(textBytes("4,8,8,4,2,0,0,2"))
			state.sweSetTidAcc(-25.826)
		} else if dversion >= 1.72 {
			state.setAstroModels(textBytes("3,8,8,4,2,0,0,2"))
			state.sweSetTidAcc(-25.7376)
		} else if dversion >= 1.70 {
			state.setAstroModels(textBytes("2,8,8,4,2,0,0,2"))
			state.sweSetTidAcc(-25.7376)
		} else if dversion >= 1.64 {
			state.setAstroModels(textBytes("2,3,1,1,1,0,0,1"))
			state.sweSetTidAcc(-25.7376)
		} else {
			state.setAstroModels(textBytes("1,3,1,1,1,0,0,1"))
			state.sweSetTidAcc(-25.7376)
		}
	}
}

// getPrecessionModel implements get_precession_model from swephlib.c.
//
//	function for inhouse testing only
func (state *State) getPrecessionModel(precmod int32, flags int32, s []byte) {
	if precmod == 0 {
		precmod = 9
	}
	if flags&flagJPL != 0 {
		if flags&flagHorizons != 0 {
			copyText(s, textBytes("IAU 1976 (Lieske) / Owen 1990 before 1799"))
			return
		}
		if flags&flagHorizonsApprox != 0 {
			copyText(s, textBytes("Vondrak 2011 / IAU 1976 (Lieske) before 1962 / Owen 1990 before 1799"))
			return
		}
	}
	switch precmod {
	case 1:
		copyText(s, textBytes("IAU 1976 (Lieske)"))
	case 6:
		copyText(s, textBytes("IAU 2000 (Lieske 1976, Mathews 2002)"))
	case 8:
		copyText(s, textBytes("IAU 2006 (Capitaine & alii)"))
	case 7:
		copyText(s, textBytes("Bretagnon 2003"))
	case 2:
		copyText(s, textBytes("Laskar 1986"))
	case 5:
		copyText(s, textBytes("Simon 1994"))
	case 4:
		copyText(s, textBytes("Williams 1994"))
	case 3:
		copyText(s, textBytes("Williams 1994 / Epsilon Laskar 1986"))
	case 10:
		copyText(s, textBytes("Owen 1990"))
	case 11:
		copyText(s, textBytes("Newcomb 1895"))
	case 9:
		copyText(s, textBytes("Vondrák 2011"))
	default:
	}
}

// getDeltatModel implements get_deltat_model from swephlib.c.
//
//	function for inhouse testing only
func (state *State) getDeltatModel(dtmod int32, s []byte) {
	if dtmod == 0 {
		dtmod = 5
	}
	switch dtmod {
	case 4:
		copyText(s, textBytes("Espenak/Meeus 2006 (before 1633)"))
	case 3:
		copyText(s, textBytes("Stephenson/Morrison 2004 (before 1600)"))
	case 2:
		copyText(s, textBytes("Stephenson 1997 (before 1600)"))
	case 1:
		copyText(s, textBytes("Stephenson/Morrison 1984 (before 1600)"))
	case 5:
		copyText(s, textBytes("Stephenson/Morrison/Hohenkerk 2016 (before 1955)"))
	}
}

// getNutationModel implements get_nutation_model from swephlib.c.
//
//	function for inhouse testing only
func (state *State) getNutationModel(nutmod int32, flags int32, s []byte) {
	var (
		jplhormod  int32 = state.positions.data.astroModels[5]
		jplhoramod int32 = state.positions.data.astroModels[6]
	)
	if jplhormod == 0 {
		jplhormod = 1
	}
	if jplhoramod == 0 {
		jplhoramod = 3
	}
	if nutmod == 0 {
		nutmod = 4
	}
	switch nutmod {
	case 5:
		copyText(s, textBytes("Woolard 1953"))
	case 1:
		copyText(s, textBytes("IAU 1980 (Wahr)"))
	case 2:
		copyText(s, textBytes("Herring 1986"))
	case 3:
		copyText(s, textBytes("IAU 2000A (Mathews)"))
	case 4:
		copyText(s, textBytes("IAU 2000B (Mathews)"))
	}
	if flags&flagJPL != 0 {
		if flags&flagHorizons != 0 {
			copyText(s, textBytes("IAU 1980 (Wahr)"))
		}
		if flags&flagHorizons != 0 {
			appendText(s, textBytes("\n+ daily corrections to dpsi/deps 1962-today"))
			if jplhormod == 1 {
				appendText(s, textBytes("\n  good agreement with JPL Horizons between 1800 and today"))
			} else {
				appendText(s, textBytes("\n  defaults to SEFLG_JPLEPH_APPROX before 1962"))
			}
		} else if flags&flagHorizonsApprox != 0 {
			appendText(s, textBytes("\n+ some corrections, approximating JPL Horizons"))
			if jplhoramod == 1 {
				appendText(s, textBytes(" (SEMOD_JPLHORA_1)"))
			} else if jplhoramod == 2 {
				appendText(s, textBytes(" (SEMOD_JPLHORA_2)"))
			} else {
				appendText(s, textBytes(" (SEMOD_JPLHORA_3)"))
			}
		}
	}
}

// getFrameBiasModel implements get_frame_bias_model from swephlib.c.
//
//	function for inhouse testing only
func (state *State) getFrameBiasModel(biasmod int32, s []byte) {
	if biasmod == 0 {
		biasmod = 3
	}
	switch biasmod {
	case 2:
		copyText(s, textBytes("IAU 2000"))
	case 3:
		copyText(s, textBytes("IAU 2006"))
	case 1:
		copyText(s, textBytes("none"))
	}
}

// getSidtModel implements get_sidt_model from swephlib.c.
//
//	function for inhouse testing only
func (state *State) getSidtModel(sidtmod int32, s []byte) {
	if sidtmod == 0 {
		sidtmod = 4
	}
	switch sidtmod {
	case 1:
		copyText(s, textBytes("IAU 1976"))
	case 2:
		copyText(s, textBytes("IAU 2006 (Capitaine 2003)"))
	case 3:
		copyText(s, textBytes("IERS Convention 2010"))
	case 4:
		copyText(s, textBytes("IERS Convention 2010 + long-term extension by Astrodienst"))
	}
}

// sweGetAstroModels implements swe_get_astro_models from swephlib.c.
//
//	function for inhouse testing only
func (state *State) sweGetAstroModels(samod []byte, sdet []byte, flags int32) {
	var (
		i             int32
		imod          int32
		pmodel        []int32 = state.positions.data.astroModels[:]
		s             [256]byte
		samod0        [256]byte
		listAllModels int32 = 0
	)
	if samod != nil {
		if findByte(samod, '+') != nil {
			listAllModels = 1
		}
		state.sweSetAstroModels(samod, flags)
	}
	samod0[0] = byte('\000')
	for i = 0; i < 8; i++ {
		imod = pmodel[i]
		switch i {
		case 1:
			if imod == 9 {
				imod = 0
			}
		case 2:
			if imod == 9 {
				imod = 0
			}
		case 3:
			if imod == 4 {
				imod = 0
			}
		case 7:
			if imod == 4 {
				imod = 0
			}
		case 4:
			if imod == 3 {
				imod = 0
			}
		case 5:
			if imod == 1 {
				imod = 0
			}
		case 6:
			if imod == 3 {
				imod = 0
			}
		case 0:
			if imod == 5 {
				imod = 0
			}
		}
		formatText(samod0[textLength(samod0[:]):], textBytes("%d,"), imod)
	}
	sdet[0] = byte('\000')
	if sdet != nil {
		formatText(sdet[textLength(sdet):], textBytes("JPL eph. %d; tidal acc. Moon used by SE: %.4f\n"), state.swiGetDenum(0, flags), state.sweGetTidAcc())
		if flags&flagJPL != 0 {
			if flags&flagHorizons != 0 {
				appendText(sdet, textBytes("JPL Horizons method:\n"))
			}
			if flags&flagHorizonsApprox != 0 {
				appendText(sdet, textBytes("JPL Horizons method (approximation):\n"))
			}
		} else if flags&flagSwiss != 0 {
			appendText(sdet, textBytes("Swiss Ephemeris compressed files sepl*/semo*\n"))
		} else {
			appendText(sdet, textBytes("Moshier semi-analytical approximation\n"))
		}
		state.getDeltatModel(pmodel[0], s[:])
		formatText(sdet[textLength(sdet):], textBytes("Delta T (long-term): %s\n"), s[:])
		state.getPrecessionModel(pmodel[1], flags, s[:])
		formatText(sdet[textLength(sdet):], textBytes("Precession: %s\n"), s[:])
		if pmodel[1] != pmodel[2] && !(flags&(flagHorizons|flagHorizonsApprox) != 0) {
			state.getPrecessionModel(pmodel[2], flags, s[:])
			formatText(sdet[textLength(sdet):], textBytes("+ short-term model: %s\n"), s[:])
		}
		state.getNutationModel(pmodel[3], flags, s[:])
		formatText(sdet[textLength(sdet):], textBytes("Nutation: %s\n"), s[:])
		state.getFrameBiasModel(pmodel[4], s[:])
		formatText(sdet[textLength(sdet):], textBytes("Frame bias: %s\n"), s[:])
		state.getSidtModel(pmodel[7], s[:])
		formatText(sdet[textLength(sdet):], textBytes("Sid. time: %s\n"), s[:])
		formatText(sdet[textLength(sdet):], textBytes("swetest parameters:      D P P N B J J S\n"))
		formatText(sdet[textLength(sdet):], textBytes("                    -amod%s"), samod0[:])
		formatText(sdet[textLength(sdet):], textBytes(" -tidacc%f"), state.sweGetTidAcc())
		appendText(sdet, textBytes("\n"))
		if !(listAllModels != 0) {
			formatText(sdet[textLength(sdet):], textBytes("For list of all available astronomical models, add a '+' to the version string\n(swetest parameter -amod%s+ or -amod%s+)\n"), samod, samod0[:])
		} else {
			appendText(sdet, textBytes("DELTA T MODELS (D)\n"))
			for i = 0; i <= 5; i++ {
				if i == 5 {
					continue
				}
				formatText(sdet[textLength(sdet):], textBytes("  (%d)"), i)
				if i == 0 {
					formatText(sdet[textLength(sdet):], textBytes(" (=%d)"), 5)
				}
				state.getDeltatModel(i, s[:])
				formatText(sdet[textLength(sdet):], textBytes(": %s\n"), s[:])
			}
			appendText(sdet, textBytes("PRECESSION MODELS (P P) (long-term/short-term)\n"))
			for i = 0; i <= 11; i++ {
				if i == 9 {
					continue
				}
				formatText(sdet[textLength(sdet):], textBytes("  (%d)"), i)
				if i == 0 {
					formatText(sdet[textLength(sdet):], textBytes(" (=%d)"), 9)
				}
				state.getPrecessionModel(i, flags, s[:])
				formatText(sdet[textLength(sdet):], textBytes(": %s\n"), s[:])
			}
			appendText(sdet, textBytes("NUTATION MODELS (N)\n"))
			for i = 0; i <= 5; i++ {
				if i == 4 {
					continue
				}
				formatText(sdet[textLength(sdet):], textBytes("  (%d)"), i)
				if i == 0 {
					formatText(sdet[textLength(sdet):], textBytes(" (=%d)"), 4)
				}
				state.getNutationModel(i, flags, s[:])
				formatText(sdet[textLength(sdet):], textBytes(": %s\n"), s[:])
			}
			appendText(sdet, textBytes("FRAME BIAS MODELS (B)\n"))
			for i = 0; i <= 3; i++ {
				if i == 3 {
					continue
				}
				formatText(sdet[textLength(sdet):], textBytes("  (%d)"), i)
				if i == 0 {
					formatText(sdet[textLength(sdet):], textBytes(" (=%d)"), 3)
				}
				state.getFrameBiasModel(i, s[:])
				formatText(sdet[textLength(sdet):], textBytes(": %s\n"), s[:])
			}
			appendText(sdet, textBytes("JPL HORIZONS MODELS (J) (with SEFLG_JPLEPH|SEFLG_JPLHOR).\n"))
			appendText(sdet, textBytes("  IAU 1980 (Wahr) + daily corrections to dpsi/deps 1962-today.\n"))
			appendText(sdet, textBytes("  (0 (=1): between 1799 and 1962, dpsi/deps of 20-jan-1962 are used.\n"))
			appendText(sdet, textBytes("           For times beyond the dpsi/deps table, the last tabulated values are used.\n"))
			appendText(sdet, textBytes("           Beyond 1799 and 2201, precession Owen 1990 is used..\n"))
			appendText(sdet, textBytes("  Documentation in swephexp.h under 'methods of JPL Horizons'\n"))
			appendText(sdet, textBytes("JPL HORIZONS APPROXIMATION (J) (with SEFLG_JPLEPH|SEFLG_JPLHORA)\n"))
			appendText(sdet, textBytes("  Documentation in swephexp.h under 'methods of JPL Horizons'\n"))
			appendText(sdet, textBytes("SIDEREAL TIME MODELS (S)\n"))
			for i = 0; i <= 4; i++ {
				if i == 4 {
					continue
				}
				formatText(sdet[textLength(sdet):], textBytes("  (%d)"), i)
				if i == 0 {
					formatText(sdet[textLength(sdet):], textBytes(" (=%d)"), 4)
				}
				state.getSidtModel(i, s[:])
				formatText(sdet[textLength(sdet):], textBytes(": %s\n"), s[:])
			}
		}
	}
}

// swiStrcpy implements swi_strcpy from swephlib.c.
func (state *State) swiStrcpy(to []byte, from []byte) []byte {
	var (
		sp []byte
		s  [256]byte
	)
	if int32(from[0]) == '\000' {
		to[0] = byte('\000')
		return to
	}
	if textLength(from) < 256 {
		copyText(s[:], from)
		copyText(to, s[:])
	} else {
		sp = duplicateText(from)
		if sp == nil {
			copyText(to, from)
		} else {
			copyText(to, sp)
		}
	}
	return to
}
