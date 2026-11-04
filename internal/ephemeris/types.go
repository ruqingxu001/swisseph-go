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

type ayanamsaEpoch struct {
	t0         float64
	ayanT0     float64
	t0IsUT     int32
	precOffset int32
}

type obliquity struct {
	epoch  float64
	angle  float64
	sine   float64
	cosine float64
}

type planetData struct {
	bodyID                int32
	fileFlags             int32
	coefficientCount      int32
	indexOffset           int32
	segmentCount          int32
	fileStart             float64
	fileEnd               float64
	segmentSpan           float64
	elementEpoch          float64
	prot                  float64
	qrot                  float64
	dprot                 float64
	dqrot                 float64
	coefficientScale      float64
	peri                  float64
	dperi                 float64
	referenceCoefficients []float64
	segmentStart          float64
	segmentEnd            float64
	coefficients          []float64
	evaluationCount       int32
	evaluationTime        float64
	ephemeris             int32
	stateVector           [6]float64
	cachedFlags           int32
	coordinates           [24]float64
}

type nutation struct {
	epoch  float64
	angles [2]float64
	sine   float64
	cosine float64
	matrix [3][3]float64
}

type planetSeries struct {
	maxHarmonic  [9]byte
	maxTimePower byte
	arguments    []int8
	longitude    []float64
	latitude     []float64
	radius       []float64
	distance     float64
}

type ephemerisFile struct {
	name         [256]byte
	version      int32
	asteroidName [50]byte
	deNumber     int32
	handle       *textFile
	start        float64
	end          float64
	flags        int32
	bodyCount    int16
	bodies       [50]int32
}

type physicalConstants struct {
	speedOfLight       float64
	astronomicalUnit   float64
	solarGravity       float64
	earthMoonMassRatio float64
	solarRadius        float64
}

type savedPosition struct {
	body        int32
	epoch       float64
	flags       int32
	coordinates [24]float64
}

type lunarNodeData struct {
	evaluationTime float64
	ephemeris      int32
	stateVector    [6]float64
	cachedFlags    int32
	coordinates    [24]float64
}

type observerData struct {
	longitude      float64
	latitude       float64
	altitude       float64
	evaluationTime float64
	jdUT           float64
	stateVector    [6]float64
}

type siderealData struct {
	mode      int32
	ayanamsa  float64
	epoch     float64
	epochIsUT int32
}

type fixedStar struct {
	searchKey       [40 + 2]byte
	name            [40 + 1]byte
	bayerName       [40 + 1]byte
	catalogNumber   [10]byte
	epoch           float64
	rightAscension  float64
	declination     float64
	properMotionRA  float64
	properMotionDec float64
	radialVelocity  float64
	parallax        float64
	magnitude       float64
}

type nutationInterpolation struct {
	tjdNut0  float64
	tjdNut2  float64
	nutDpsi0 float64
	nutDpsi1 float64
	nutDpsi2 float64
	nutDeps0 float64
	nutDeps1 float64
	nutDeps2 float64
}

type calculationData struct {
	ephePathIsSet      int32
	jplFileIsOpen      int32
	starFile           *textFile
	ephemerisPath      [256]byte
	jplFilename        [256]byte
	jplDENumber        int32
	lastEpheflag       int32
	geoposIsSet        int32
	ayanaIsSet         int32
	isOldStarfile      int32
	eopTjdBeg          float64
	eopTjdBegHorizons  float64
	eopTjdEnd          float64
	eopTjdEndAdd       float64
	eopDpsiLoaded      int32
	tidAcc             float64
	isTidAccManual     int32
	initDtDone         int32
	swedIsInitialised  int32
	deltaTUserdefIsSet int32
	deltaTUserdef      float64
	astG               float64
	astH               float64
	astDiam            float64
	astelem            [256 * 10]byte
	iSavedPlanetName   int32
	savedPlanetName    [80]byte
	dpsi               []float64
	deps               []float64
	timeout            int32
	astroModels        [8]int32
	doInterpolateNut   int32
	interpol           nutationInterpolation
	files              [7]ephemerisFile
	constants          physicalConstants
	planets            [18]planetData
	lunarNodes         [6]planetData
	savedPositions     [23 + 1]savedPosition
	obliquity          obliquity
	obliquityJ2000     obliquity
	nutation           nutation
	nutationJ2000      nutation
	nutationWithSpeed  nutation
	observer           observerData
	sidereal           siderealData
	nFixstarsReal      int32
	nFixstarsNamed     int32
	nFixstarsRecords   int32
	fixedStars         []fixedStar
}

type houseData struct {
	cusp                     [37]float64
	cuspSpeed                [37]float64
	ascendant                float64
	ascendantSpeed           float64
	midheaven                float64
	midheavenSpeed           float64
	armcSpeed                float64
	vertex                   float64
	vertexSpeed              float64
	equatorialAscendant      float64
	equatorialAscendantSpeed float64
	coAscendantKoch          float64
	coAscendantKochSpeed     float64
	coAscendantMunkasey      float64
	coAscendantMunkaseySpeed float64
	polarAscendant           float64
	polarAscendantSpeed      float64
	sunDeclination           float64
	needSpeed                int32
	needCuspSpeed            int32
	needInterpolation        int32
	message                  [256]byte
}

type jplFile struct {
	name                   []byte
	path                   []byte
	handle                 *textFile
	swapBytes              int16
	constantValues         [400]float64
	epochRange             [3]float64
	astronomicalUnit       float64
	earthMoonMassRatio     float64
	deNumber               int32
	constantCount          int32
	coefficientIndex       [39]int32
	constantNames          [6 * 400]byte
	planetVectors          [78]float64
	sunVector              [6]float64
	record                 [1500]float64
	positionPolynomial     [18]float64
	velocityPolynomial     [18]float64
	accelerationPolynomial [18]float64
	jerkPolynomial         [18]float64
	useKilometres          int16
}

type massDistribution struct {
	radiusFraction float64
	massFraction   float64
}

type sarosSeries struct {
	seriesNo   int32
	startEpoch float64
}
