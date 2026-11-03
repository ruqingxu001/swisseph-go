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

// swiMoshplan2 implements swi_moshplan2 from swemplan.c.
func (state *State) swiMoshplan2(jd float64, planetIndex int32, polar []float64) int32 {
	var (
		i                     int32
		j                     int32
		k                     int32
		m                     int32
		hasArgument           int32
		termIndex             int32
		argumentCount         int32
		timePower             int32
		p                     []int8
		longitudeCoefficients []float64
		latitudeCoefficients  []float64
		radiusCoefficients    []float64
		sineCoefficient       float64
		cosineCoefficient     float64
		sineArgument          float64
		cosineArgument        float64
		millennia             float64
		t                     float64
		longitudeSum          float64
		latitudeSum           float64
		radiusSum             float64
		series                *planetSeries = planetsPlanets[planetIndex]
	)
	millennia = (jd - julianDayJ2000) / 3652500.0
	for i = 0; i < 9; i++ {
		if j = int32(series.maxHarmonic[i]); j > 0 {
			radiusSum = (planetsFreqs[i]*millennia - 1.296e6*math.Floor(planetsFreqs[i]*millennia/1.296e6) + planetsPhases[i]) * 4.8481368110953599359e-6
			state.planetsSscc(i, radiusSum, j)
		}
	}
	p = series.arguments
	longitudeCoefficients = series.longitude
	latitudeCoefficients = series.latitude
	radiusCoefficients = series.radius
	longitudeSum = 0.0
	latitudeSum = 0.0
	for radiusSum = 0.0; ; {
		argumentCount = int32(nextValue(&p))
		if argumentCount < 0 {
			break
		}
		if argumentCount == 0 {
			timePower = int32(nextValue(&p))
			cosineCoefficient = nextValue(&longitudeCoefficients)
			for termIndex = 0; termIndex < timePower; termIndex++ {
				cosineCoefficient = cosineCoefficient*millennia + nextValue(&longitudeCoefficients)
			}
			longitudeSum += cosineCoefficient - 1.296e6*math.Floor(cosineCoefficient/1.296e6)
			cosineCoefficient = nextValue(&latitudeCoefficients)
			for termIndex = 0; termIndex < timePower; termIndex++ {
				cosineCoefficient = cosineCoefficient*millennia + nextValue(&latitudeCoefficients)
			}
			latitudeSum += cosineCoefficient
			cosineCoefficient = nextValue(&radiusCoefficients)
			for termIndex = 0; termIndex < timePower; termIndex++ {
				cosineCoefficient = cosineCoefficient*millennia + nextValue(&radiusCoefficients)
			}
			radiusSum += cosineCoefficient
			continue
		}
		hasArgument = 0
		cosineArgument = 0.0
		sineArgument = 0.0
		for termIndex = 0; termIndex < argumentCount; termIndex++ {
			j = int32(nextValue(&p))
			m = int32(nextValue(&p)) - 1
			if j != 0 {
				k = j
				if j < 0 {
					k = -k
				}
				k -= 1
				sineCoefficient = state.planetary.sinMultiples[m][k]
				if j < 0 {
					sineCoefficient = -sineCoefficient
				}
				cosineCoefficient = state.planetary.cosMultiples[m][k]
				if hasArgument == 0 {
					sineArgument = sineCoefficient
					cosineArgument = cosineCoefficient
					hasArgument = 1
				} else {
					t = sineCoefficient*cosineArgument + cosineCoefficient*sineArgument
					cosineArgument = cosineCoefficient*cosineArgument - sineCoefficient*sineArgument
					sineArgument = t
				}
			}
		}
		timePower = int32(nextValue(&p))
		cosineCoefficient = nextValue(&longitudeCoefficients)
		sineCoefficient = nextValue(&longitudeCoefficients)
		for termIndex = 0; termIndex < timePower; termIndex++ {
			cosineCoefficient = cosineCoefficient*millennia + nextValue(&longitudeCoefficients)
			sineCoefficient = sineCoefficient*millennia + nextValue(&longitudeCoefficients)
		}
		longitudeSum += cosineCoefficient*cosineArgument + sineCoefficient*sineArgument
		cosineCoefficient = nextValue(&latitudeCoefficients)
		sineCoefficient = nextValue(&latitudeCoefficients)
		for termIndex = 0; termIndex < timePower; termIndex++ {
			cosineCoefficient = cosineCoefficient*millennia + nextValue(&latitudeCoefficients)
			sineCoefficient = sineCoefficient*millennia + nextValue(&latitudeCoefficients)
		}
		latitudeSum += cosineCoefficient*cosineArgument + sineCoefficient*sineArgument
		cosineCoefficient = nextValue(&radiusCoefficients)
		sineCoefficient = nextValue(&radiusCoefficients)
		for termIndex = 0; termIndex < timePower; termIndex++ {
			cosineCoefficient = cosineCoefficient*millennia + nextValue(&radiusCoefficients)
			sineCoefficient = sineCoefficient*millennia + nextValue(&radiusCoefficients)
		}
		radiusSum += cosineCoefficient*cosineArgument + sineCoefficient*sineArgument
	}
	polar[0] = 4.8481368110953599359e-6 * longitudeSum
	polar[1] = 4.8481368110953599359e-6 * latitudeSum
	polar[2] = 4.8481368110953599359e-6*series.distance*radiusSum + series.distance
	return 0
}

// swiMoshplan implements swi_moshplan from swemplan.c.
//
//	Moshier ephemeris.
//
// computes heliocentric cartesian equatorial coordinates of
// equinox 2000
// for earth and a planet
// tjd		julian day
// ipli		internal SWEPH planet number
// xp		array of 6 doubles for planet's position and speed
// xe		                       earth's
// serr		error string
func (state *State) swiMoshplan(jd float64, bodyIndex int32, saveCache int32, xpret []float64, xeret []float64, message []byte) int32 {
	var (
		i           int32
		doEarth     int32 = 0
		dx          [3]float64
		x2          [3]float64
		xxe         [6]float64
		xxp         [6]float64
		xp          []float64
		xe          []float64
		dt          float64
		s           [256]byte
		planetIndex int32       = planetsPnoint2msh[bodyIndex]
		planet      *planetData = &state.positions.data.planets[bodyIndex]
		earthData   *planetData = &state.positions.data.planets[0]
		seps2000    float64     = state.positions.data.obliquityJ2000.sine
		ceps2000    float64     = state.positions.data.obliquityJ2000.cosine
	)
	if saveCache != 0 {
		xp = planet.stateVector[:]
		xe = earthData.stateVector[:]
	} else {
		xp = xxp[:]
		xe = xxe[:]
	}
	if saveCache != 0 || bodyIndex == 0 || xeret != nil {
		doEarth = 1
	}
	if jd < 625000.19999999995 || jd > 2818000.7999999998 {
		if message != nil {
			formatText(s[:], textBytes("jd %f outside Moshier planet range %.2f .. %.2f "), jd, 625000.5, 2818000.5)
			if textLength(message)+textLength(s[:]) < 256 {
				appendText(message, s[:])
			}
		}
		return -1
	}
	if doEarth != 0 {
		if jd == earthData.evaluationTime && earthData.ephemeris == 4 {
			xe = earthData.stateVector[:]
		} else {
			state.swiMoshplan2(jd, planetsPnoint2msh[0], xe)
			state.swiPolcart(xe, xe)
			state.swiCoortrf2(xe, xe, -seps2000, ceps2000)
			state.embofsMosh(jd, xe)
			if saveCache != 0 {
				earthData.evaluationTime = jd
				earthData.cachedFlags = -1
				earthData.ephemeris = 4
			}
			state.swiMoshplan2(jd-0.0001, planetsPnoint2msh[0], x2[:])
			state.swiPolcart(x2[:], x2[:])
			state.swiCoortrf2(x2[:], x2[:], -seps2000, ceps2000)
			state.embofsMosh(jd-0.0001, x2[:])
			for i = 0; i <= 2; i++ {
				dx[i] = (xe[i] - x2[i]) / 0.0001
			}
			for i = 0; i <= 2; i++ {
				xe[i+3] = dx[i]
			}
		}
		if xeret != nil {
			for i = 0; i <= 5; i++ {
				xeret[i] = xe[i]
			}
		}
	}
	if bodyIndex == 0 {
		xp = xe
	} else {
		if jd == planet.evaluationTime && planet.ephemeris == 4 {
			xp = planet.stateVector[:]
		} else {
			state.swiMoshplan2(jd, planetIndex, xp)
			state.swiPolcart(xp, xp)
			state.swiCoortrf2(xp, xp, -seps2000, ceps2000)
			if saveCache != 0 {
				planet.evaluationTime = jd
				planet.cachedFlags = -1
				planet.ephemeris = 4
			}
			dt = 0.0001
			state.swiMoshplan2(jd-dt, planetIndex, x2[:])
			state.swiPolcart(x2[:], x2[:])
			state.swiCoortrf2(x2[:], x2[:], -seps2000, ceps2000)
			for i = 0; i <= 2; i++ {
				dx[i] = (xp[i] - x2[i]) / dt
			}
			for i = 0; i <= 2; i++ {
				xp[i+3] = dx[i]
			}
		}
		if xpret != nil {
			for i = 0; i <= 5; i++ {
				xpret[i] = xp[i]
			}
		}
	}
	return 0
}

func (state *State) planetsSscc(k int32, arg float64, n int32) {
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
	state.planetary.sinMultiples[k][0] = su
	state.planetary.cosMultiples[k][0] = cu
	sv = 2.0 * su * cu
	cv = cu*cu - su*su
	state.planetary.sinMultiples[k][1] = sv
	state.planetary.cosMultiples[k][1] = cv
	for i = 2; i < n; i++ {
		s = su*cv + cu*sv
		cv = cu*cv - su*sv
		sv = s
		state.planetary.sinMultiples[k][i] = sv
		state.planetary.cosMultiples[k][i] = cv
	}
}

// embofsMosh implements embofs_mosh from swemplan.c.
//
//	Adjust position from Earth-Moon barycenter to Earth
//
// J = Julian day number
// xemb = rectangular equatorial coordinates of Earth
func (state *State) embofsMosh(jd float64, xemb []float64) {
	var (
		T    float64
		M    float64
		a    float64
		L    float64
		B    float64
		p    float64
		smp  float64
		cmp  float64
		s2mp float64
		c2mp float64
		s2d  float64
		c2d  float64
		sf   float64
		cf   float64
		s2f  float64
		sx   float64
		cx   float64
		xyz  [6]float64
		seps float64 = state.positions.data.obliquity.sine
		ceps float64 = state.positions.data.obliquity.cosine
		i    int32
	)
	T = (jd - 2415020.0) / daysPerCentury
	a = state.sweDegnorm(((1.44e-5*T+0.009192)*T+477198.8491)*T + 296.104608)
	a *= radiansPerDegree
	smp = math.Sin(a)
	cmp = math.Cos(a)
	s2mp = 2.0 * smp * cmp
	c2mp = cmp*cmp - smp*smp
	a = state.sweDegnorm(((1.9e-6*T-0.001436)*T+445267.1142)*T + 350.737486)
	a = 0.034906585039886591 * a
	s2d = math.Sin(a)
	c2d = math.Cos(a)
	a = state.sweDegnorm(((-3.e-7*T-0.003211)*T+483202.0251)*T + 11.250889)
	a *= radiansPerDegree
	sf = math.Sin(a)
	cf = math.Cos(a)
	s2f = 2.0 * sf * cf
	sx = s2d*cmp - c2d*smp
	cx = c2d*cmp + s2d*smp
	L = ((1.9e-6*T-0.001133)*T+481267.8831)*T + 270.434164
	M = state.sweDegnorm(((-3.3e-6*T-1.50e-4)*T+35999.0498)*T + 358.475833)
	L = L + 6.288750*smp + 1.274018*sx + 0.658309*s2d + 0.213616*s2mp - 0.185596*math.Sin(radiansPerDegree*M) - 0.114336*s2f
	a = smp * cf
	sx = cmp * sf
	B = 5.128189*sf + 0.280606*(a+sx) + 0.277693*(a-sx) + 0.173238*(s2d*cf-c2d*sf)
	B *= radiansPerDegree
	p = 0.950724 + 0.051818*cmp + 0.009531*cx + 0.007843*c2d + 0.002824*c2mp
	p *= radiansPerDegree
	L = state.sweDegnorm(L)
	L *= radiansPerDegree
	a = 4.263523e-5 / math.Sin(p)
	xyz[0] = L
	xyz[1] = B
	xyz[2] = a
	state.swiPolcart(xyz[:], xyz[:])
	state.swiCoortrf2(xyz[:], xyz[:], -seps, ceps)
	state.swiPrecess(xyz[:], jd, 0, 1)
	for i = 0; i <= 2; i++ {
		xemb[i] -= xyz[i] / 82.30055985272827
	}
}

func (state *State) swiGetFictName(body int32, snam []byte) []byte {
	if state.readElementsFile(body, 0, nil, nil, nil, nil, nil, nil, nil, nil, snam, nil, nil) == -1 {
		copyText(snam, textBytes("name not found"))
	}
	return snam
}

// swiOscElPlan implements swi_osc_el_plan from swemplan.c.
//
//	computes a planet from osculating elements *
//
// tjd		julian day
// ipl		body number
// ipli 	body number in planetary data structure
// iflag	flags
func (state *State) swiOscElPlan(jd float64, xp []float64, body int32, bodyIndex int32, earthVector []float64, sunVector []float64, message []byte) int32 {
	var (
		pqr       [9]float64
		x         [6]float64
		eps       float64
		K         float64
		fac       float64
		rho       float64
		cose      float64
		sine      float64
		alpha     float64
		beta      float64
		zeta      float64
		sigma     float64
		M2        float64
		Msgn      float64
		M180Or0   float64
		initialJD [1]float64
		tequ      [1]float64
		mano      [1]float64
		sema      [1]float64
		ecce      [1]float64
		parg      [1]float64
		node      [1]float64
		incl      [1]float64
		dmot      float64
		cosnode   float64
		sinnode   float64
		cosincl   float64
		sinincl   float64
		cosparg   float64
		sinparg   float64
		M         float64
		E         float64
		earthData *planetData = &state.positions.data.planets[0]
		planet    *planetData = &state.positions.data.planets[bodyIndex]
		fictIfl   [1]int32    = [1]int32{0}
		i         int32
	)
	if state.readElementsFile(body, jd, initialJD[:], tequ[:], mano[:], sema[:], ecce[:], parg[:], node[:], incl[:], nil, fictIfl[:], message) == -1 {
		return -1
	}
	dmot = 0.017202098949975132 / sema[0] / math.Sqrt(sema[0])
	if fictIfl[0]&1 != 0 {
		dmot /= math.Sqrt(332946.050895)
	}
	cosnode = math.Cos(node[0])
	sinnode = math.Sin(node[0])
	cosincl = math.Cos(incl[0])
	sinincl = math.Sin(incl[0])
	cosparg = math.Cos(parg[0])
	sinparg = math.Sin(parg[0])
	pqr[0] = cosparg*cosnode - sinparg*cosincl*sinnode
	pqr[1] = -sinparg*cosnode - cosparg*cosincl*sinnode
	pqr[2] = sinincl * sinnode
	pqr[3] = cosparg*sinnode + sinparg*cosincl*cosnode
	pqr[4] = -sinparg*sinnode + cosparg*cosincl*cosnode
	pqr[5] = -sinincl * cosnode
	pqr[6] = sinparg * sinincl
	pqr[7] = cosparg * sinincl
	pqr[8] = cosincl
	M = state.swiMod2PI(mano[0] + (jd-initialJD[0])*dmot)
	E = M
	if ecce[0] > 0.975 {
		M2 = M * degreesPerRadian
		if M2 > 150 && M2 < 210 {
			M2 -= 180
			M180Or0 = 180
		} else {
			M180Or0 = 0
		}
		if M2 > 330 {
			M2 -= 360
		}
		if M2 < 0 {
			M2 = -M2
			Msgn = float64(-1)
		} else {
			Msgn = 1
		}
		if M2 < 30 {
			M2 *= radiansPerDegree
			alpha = (1 - ecce[0]) / (4*ecce[0] + 0.5)
			beta = M2 / (8*ecce[0] + 1)
			zeta = math.Pow(beta+math.Sqrt(beta*beta+alpha*alpha), float64(0.33333333333333331))
			sigma = zeta - alpha/2
			sigma = sigma - 0.078*sigma*sigma*sigma*sigma*sigma/(1+ecce[0])
			E = Msgn*(M2+ecce[0]*(3*sigma-4*sigma*sigma*sigma)) + M180Or0
		}
	}
	E = state.swiKepler(E, M, ecce[0])
	if fictIfl[0]&1 != 0 {
		K = 0.0000298122353216 / math.Sqrt(sema[0])
	} else {
		K = 0.01720209895 / math.Sqrt(sema[0])
	}
	cose = math.Cos(E)
	sine = math.Sin(E)
	fac = math.Sqrt((1 - ecce[0]) * (1 + ecce[0]))
	rho = 1 - ecce[0]*cose
	x[0] = sema[0] * (cose - ecce[0])
	x[1] = sema[0] * fac * sine
	x[3] = -K * sine / rho
	x[4] = K * fac * cose / rho
	xp[0] = pqr[0]*x[0] + pqr[1]*x[1]
	xp[1] = pqr[3]*x[0] + pqr[4]*x[1]
	xp[2] = pqr[6]*x[0] + pqr[7]*x[1]
	xp[3] = pqr[0]*x[3] + pqr[1]*x[4]
	xp[4] = pqr[3]*x[3] + pqr[4]*x[4]
	xp[5] = pqr[6]*x[3] + pqr[7]*x[4]
	eps = state.swiEpsiln(tequ[0], 0)
	state.swiCoortrf(xp, xp, -eps)
	state.swiCoortrf(xp[3:], xp[3:], -eps)
	if tequ[0] != julianDayJ2000 {
		state.swiPrecess(xp, tequ[0], 0, 1)
		state.swiPrecess(xp[3:], tequ[0], 0, 1)
	}
	if fictIfl[0]&1 != 0 {
		for i = 0; i <= 5; i++ {
			xp[i] += earthVector[i]
		}
	} else {
		for i = 0; i <= 5; i++ {
			xp[i] += sunVector[i]
		}
	}
	if sameSlice(planet.stateVector[:], xp) {
		planet.evaluationTime = jd
		planet.ephemeris = earthData.ephemeris
	}
	return 0
}

// readElementsFile implements read_elements_file from swemplan.c.
//
//	note: input parameter tjd is required for T terms in elements
func (state *State) readElementsFile(body int32, jd float64, initialJD []float64, tequ []float64, mano []float64, sema []float64, ecce []float64, parg []float64, node []float64, incl []float64, pname []byte, fictIfl []int32, message []byte) int32 {
	var (
		i         int32
		iline     int32
		iplan     int32
		status    int32
		ncpos     int32
		fp        *textFile = nil
		s         [256]byte
		sp        []byte
		cpos      [20][]byte
		serri     [256]byte
		elemFound int32   = 0
		tt        float64 = 0
	)
	if fp = state.swiFopen(-1, textBytes("seorbel.txt"), state.positions.data.ephemerisPath[:], message); fp == nil {
		if body >= 15 {
			if message != nil {
				formatText(message, textBytes("error no elements for fictitious body no %7.0f"), float64(body))
			}
			return -1
		}
		if initialJD != nil {
			initialJD[0] = planetsPlanOscuElem[body][0]
		}
		if tequ != nil {
			tequ[0] = planetsPlanOscuElem[body][1]
		}
		if mano != nil {
			mano[0] = planetsPlanOscuElem[body][2] * radiansPerDegree
		}
		if sema != nil {
			sema[0] = planetsPlanOscuElem[body][3]
		}
		if ecce != nil {
			ecce[0] = planetsPlanOscuElem[body][4]
		}
		if parg != nil {
			parg[0] = planetsPlanOscuElem[body][5] * radiansPerDegree
		}
		if node != nil {
			node[0] = planetsPlanOscuElem[body][6] * radiansPerDegree
		}
		if incl != nil {
			incl[0] = planetsPlanOscuElem[body][7] * radiansPerDegree
		}
		if pname != nil {
			copyText(pname, planetsPlanFictNam[body])
		}
		return 0
	}
	iline = 0
	for iplan = -1; readLine(s[:], 256, fp) != nil; {
		iline++
		for sp = s[:]; int32(sp[0]) == ' ' || int32(sp[0]) == '\t'; {
			sp = sp[1:]
		}
		state.swiStrcpy(s[:], sp)
		if int32(s[0]) == '#' {
			continue
		}
		if int32(s[0]) == '\r' {
			continue
		}
		if int32(s[0]) == '\n' {
			continue
		}
		if int32(s[0]) == '\000' {
			continue
		}
		if sp = findByte(s[:], '#'); sp != nil {
			sp[0] = byte('\000')
		}
		ncpos = state.swiCutstr(s[:], textBytes(","), cpos[:], 20)
		formatText(serri[:], textBytes("error in file %s, line %7.0f:"), textBytes("seorbel.txt"), float64(iline))
		if ncpos < 9 {
			if message != nil {
				formatText(message, textBytes("%s nine elements required"), serri[:])
			}
			closeTextFile(fp)
			return -1
		}
		iplan++
		if iplan != body {
			continue
		}
		elemFound = 1
		if initialJD != nil {
			sp = cpos[0]
			for i = 0; i < 5; i++ {
				sp[i] = byte(lowerByte(sp[i]))
			}
			if compareTextN(sp, textBytes("j2000"), 5) == 0 {
				initialJD[0] = julianDayJ2000
			} else if compareTextN(sp, textBytes("b1950"), 5) == 0 {
				initialJD[0] = 2433282.42345905
			} else if compareTextN(sp, textBytes("j1900"), 5) == 0 {
				initialJD[0] = 2415020.0
			} else if int32(sp[0]) == 'j' || int32(sp[0]) == 'b' {
				if message != nil {
					formatText(message, textBytes("%s invalid epoch"), serri[:])
				}
				closeTextFile(fp)
				return -1
			} else {
				initialJD[0] = parseFloat(sp)
			}
			tt = jd - initialJD[0]
		}
		if tequ != nil {
			for sp = cpos[1]; int32(sp[0]) == ' ' || int32(sp[0]) == '\t'; {
				sp = sp[1:]
			}
			for i = 0; i < 5; i++ {
				sp[i] = byte(lowerByte(sp[i]))
			}
			if compareTextN(sp, textBytes("j2000"), 5) == 0 {
				tequ[0] = julianDayJ2000
			} else if compareTextN(sp, textBytes("b1950"), 5) == 0 {
				tequ[0] = 2433282.42345905
			} else if compareTextN(sp, textBytes("j1900"), 5) == 0 {
				tequ[0] = 2415020.0
			} else if compareTextN(sp, textBytes("jdate"), 5) == 0 {
				tequ[0] = jd
			} else if int32(sp[0]) == 'j' || int32(sp[0]) == 'b' {
				if message != nil {
					formatText(message, textBytes("%s invalid equinox"), serri[:])
				}
				closeTextFile(fp)
				return -1
			} else {
				tequ[0] = parseFloat(sp)
			}
		}
		if mano != nil {
			status = state.checkTTerms(tt, cpos[2], mano)
			mano[0] = state.sweDegnorm(mano[0])
			if status == -1 {
				if message != nil {
					formatText(message, textBytes("%s mean anomaly value invalid"), serri[:])
				}
				closeTextFile(fp)
				return -1
			}
			if status == 1 {
				initialJD[0] = jd
			}
			mano[0] *= radiansPerDegree
		}
		if sema != nil {
			status = state.checkTTerms(tt, cpos[3], sema)
			if sema[0] <= 0 || status == -1 {
				if message != nil {
					formatText(message, textBytes("%s semi-axis value invalid"), serri[:])
				}
				closeTextFile(fp)
				return -1
			}
		}
		if ecce != nil {
			status = state.checkTTerms(tt, cpos[4], ecce)
			if ecce[0] >= 1 || ecce[0] < 0 || status == -1 {
				if message != nil {
					formatText(message, textBytes("%s eccentricity invalid (no parabolic or hyperbolic orbits allowed)"), serri[:])
				}
				closeTextFile(fp)
				return -1
			}
		}
		if parg != nil {
			status = state.checkTTerms(tt, cpos[5], parg)
			parg[0] = state.sweDegnorm(parg[0])
			if status == -1 {
				if message != nil {
					formatText(message, textBytes("%s perihelion argument value invalid"), serri[:])
				}
				closeTextFile(fp)
				return -1
			}
			parg[0] *= radiansPerDegree
		}
		if node != nil {
			status = state.checkTTerms(tt, cpos[6], node)
			node[0] = state.sweDegnorm(node[0])
			if status == -1 {
				if message != nil {
					formatText(message, textBytes("%s node value invalid"), serri[:])
				}
				closeTextFile(fp)
				return -1
			}
			node[0] *= radiansPerDegree
		}
		if incl != nil {
			status = state.checkTTerms(tt, cpos[7], incl)
			incl[0] = state.sweDegnorm(incl[0])
			if status == -1 {
				if message != nil {
					formatText(message, textBytes("%s inclination value invalid"), serri[:])
				}
				closeTextFile(fp)
				return -1
			}
			incl[0] *= radiansPerDegree
		}
		if pname != nil {
			for sp = cpos[8]; int32(sp[0]) == ' ' || int32(sp[0]) == '\t'; {
				sp = sp[1:]
			}
			state.swiRightTrim(sp)
			copyText(pname, sp)
		}
		if fictIfl != nil && ncpos > 9 {
			for sp = cpos[9]; int32(sp[0]) != '\000'; sp = sp[1:] {
				sp[0] = byte(lowerByte(sp[0]))
			}
			if findText(cpos[9], textBytes("geo")) != nil {
				fictIfl[0] |= 1
			}
		}
		break
	}
	if !(elemFound != 0) {
		if message != nil {
			formatText(message, textBytes("%s elements for planet %7.0f not found"), serri[:], float64(body))
		}
		closeTextFile(fp)
		return -1
	}
	closeTextFile(fp)
	return 0
}

// checkTTerms implements check_t_terms from swemplan.c.
func (state *State) checkTTerms(t float64, sinp []byte, doutp []float64) int32 {
	var (
		i      int32
		isgn   int32 = 1
		z      int32
		status int32 = 0
		sp     []byte
		tt     [5]float64
		fac    float64
	)
	tt[0] = t / 36525
	tt[1] = tt[0]
	tt[2] = tt[1] * tt[1]
	tt[3] = tt[2] * tt[1]
	tt[4] = tt[3] * tt[1]
	if sp = findAnyText(sinp, textBytes("+-")); sp != nil {
		status = 1
	}
	sp = sinp
	doutp[0] = 0
	fac = 1
	for z = 0; 1 != 0; {
		for int32(sp[0]) != '\000' && findByte(textBytes(" \t"), sp[0]) != nil {
			sp = sp[1:]
		}
		if findByte(textBytes("+-"), sp[0]) != nil || int32(sp[0]) == '\000' {
			if z > 0 {
				doutp[0] += fac
			}
			isgn = 1
			if int32(sp[0]) == '-' {
				isgn = -1
			}
			fac = float64(1 * isgn)
			if int32(sp[0]) == '\000' {
				return status
			}
			sp = sp[1:]
		} else {
			for int32(sp[0]) != '\000' && findByte(textBytes("* \t"), sp[0]) != nil {
				sp = sp[1:]
			}
			if int32(sp[0]) != '\000' && findByte(textBytes("tT"), sp[0]) != nil {
				sp = sp[1:]
				if int32(sp[0]) != '\000' && findByte(textBytes("+-"), sp[0]) != nil {
					fac *= tt[0]
				} else if i = parseInt(sp); i <= 4 && i >= 0 {
					fac *= tt[i]
				}
			} else if parseFloat(sp) != 0 || int32(sp[0]) == '0' {
				fac *= parseFloat(sp)
			}
			for int32(sp[0]) != '\000' && findByte(textBytes("0123456789."), sp[0]) != nil {
				sp = sp[1:]
			}
		}
		z++
	}
	return status
}
