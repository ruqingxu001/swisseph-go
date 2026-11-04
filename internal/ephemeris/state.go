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

type State struct {
	houses    housesState
	jpl       jplState
	positions positionsState
	heliacal  heliacalState
	calendar  calendarState
	lunar     lunarState
	planetary planetaryState
	time      timeState
	eclipse   eclipseState
}

type housesState struct {
	savedSunDeclination float64
}

type jplState struct {
	interpNp     int32
	interpNv     int32
	interpNac    int32
	interpNjk    int32
	interpTwot   float64
	stateIrecsz  int32
	stateNrl     int32
	stateLpt     [3]int32
	stateNcoeffs int32
	file         *jplFile
}

type positionsState struct {
	checkNutationNutflag          int32
	fixstarCalcFromStructXearth   [6]float64
	fixstarCalcFromStructXearthDt [6]float64
	fixstarCalcFromStructXsun     [6]float64
	fixstarCalcFromStructXsunDt   [6]float64
	fixstar2SlastStarname         [256]byte
	fixstar2LastStardata          fixedStar
	fixstar2MagSlastStarname      [256]byte
	fixstar2MagLastStardata       fixedStar
	fixstarCalcFromRecordXearth   [6]float64
	fixstarCalcFromRecordXearthDt [6]float64
	fixstarCalcFromRecordXsun     [6]float64
	fixstarCalcFromRecordXsunDt   [6]float64
	fixstarSlastStardata          [256]byte
	fixstarSlastStarname          [256]byte
	fixstarMagSlastStardata       [256]byte
	fixstarMagSlastStarname       [256]byte
	data                          calculationData
}

type heliacalState struct {
	callSweFixstarMagDmag     [1]float64
	callSweFixstarMagStarSave [256]byte
	sunRATjdlast              float64
	sunRARalast               float64
	kOZKozLast                float64
	kOZAltsLast               float64
	kOZSunraLast              float64
	kOZA                      int32
	kaAltsLast                float64
	kaSunraLast               float64
	kaKaLast                  float64
	deltamAltsLast            float64
	deltamAltoLast            float64
	deltamSunraLast           float64
	deltamDeltamLast          float64
	bskyA                     int32
	visLimMagnA               int32
}

type calendarState struct {
	leapSecondsLoaded int32
	leapSeconds       [100]int32
}

type lunarState struct {
	sinMultiples     [5][8]float64
	cosMultiples     [5][8]float64
	longitude        float64
	latitude         float64
	polar            [3]float64
	meanLongitude    float64
	solarAnomaly     float64
	lunarAnomaly     float64
	elongation       float64
	latitudeArgument float64
	centuries        float64
	centuries2       float64
	centuries3       float64
	centuries4       float64
	f                float64
	g                float64
	venusLongitude   float64
	earthLongitude   float64
	marsLongitude    float64
	jupiterLongitude float64
	saturnLongitude  float64
	cg               float64
	sg               float64
	l1               float64
	l2               float64
	l3               float64
	l4               float64
}

type planetaryState struct {
	sinMultiples [9][24]float64
	cosMultiples [9][24]float64
}

type timeState struct {
	deltaTTable [2028 - 1620 + 1 + 100]float64
	crcTable    [256]uint32
}

type eclipseState struct {
	lapseRate float64
}

// NewState initializes the upstream defaults without opening ephemeris files.
func NewState() *State {
	state := &State{}
	state.houses.savedSunDeclination = 99
	state.calendar.leapSeconds = defaultLeapSeconds
	state.time.deltaTTable = defaultDeltaT
	state.eclipse.lapseRate = 0.0065
	return state
}
