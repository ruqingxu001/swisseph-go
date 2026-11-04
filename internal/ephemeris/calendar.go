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

// sweDateConversion implements swe_date_conversion from swedate.c.
// sweDateConversion checks a civil date by converting it to a Julian day and
// back. As in swe_date_conversion, only 'g' selects the Gregorian calendar.
func (state *State) sweDateConversion(year, month, day int32, hour float64, calendarCode byte, output []float64) int32 {
	calendar := int32(0)
	if calendarCode == 'g' {
		calendar = 1
	}
	jd := state.sweJulday(year, month, day, hour, calendar)
	actualYear, actualMonth, actualDay, _ := reverseJulianDay(jd, calendar)
	output[0] = jd
	if actualYear != year || actualMonth != month || actualDay != day {
		return -1
	}
	return 0
}

// sweJulday converts a civil date to a Julian day. Astronomical year numbering
// includes year zero. The Gregorian correction also handles negative centuries;
// keep the arithmetic order of swedate.c to preserve its rounding at midnight.
func (state *State) sweJulday(year, month, day int32, hour float64, calendar int32) float64 {
	adjustedYear := float64(year)
	if month < 3 {
		adjustedYear--
	}
	shiftedYear := adjustedYear + 4712.0
	shiftedMonth := float64(month) + 1.0
	if shiftedMonth < 4 {
		shiftedMonth += 12.0
	}
	jd := math.Floor(shiftedYear*365.25) + math.Floor(30.6*shiftedMonth+0.000001) + float64(day) + hour/24.0 - 63.5
	if calendar == 1 {
		centuryCorrection := math.Floor(math.Abs(adjustedYear)/100) - math.Floor(math.Abs(adjustedYear)/400)
		if adjustedYear < 0 {
			centuryCorrection = -centuryCorrection
		}
		jd = jd - centuryCorrection + 2
		if adjustedYear < 0 && adjustedYear/100 == math.Floor(adjustedYear/100) && adjustedYear/400 != math.Floor(adjustedYear/400) {
			jd--
		}
	}
	return jd
}

// reverseJulianDay returns a civil date and fractional hours. The intermediate
// epochs follow swe_revjul, including its historical Gregorian correction.
func reverseJulianDay(jd float64, calendar int32) (year, month, day int32, hour float64) {
	shiftedDay := jd + 32082.5
	if calendar == 1 {
		correctedDay := shiftedDay + math.Floor(shiftedDay/daysPerCentury) - math.Floor(shiftedDay/146100.0) - 38.0
		if jd >= 1830691.5 {
			correctedDay++
		}
		shiftedDay = shiftedDay + math.Floor(correctedDay/daysPerCentury) - math.Floor(correctedDay/146100.0) - 38.0
	}
	dayNumber := math.Floor(shiftedDay + 123.0)
	yearNumber := math.Floor((dayNumber - 122.2) / 365.25)
	monthNumber := math.Floor((dayNumber - math.Floor(365.25*yearNumber)) / 30.6001)
	month = int32(monthNumber - 1.0)
	if month > 12 {
		month -= 12
	}
	day = int32(dayNumber - math.Floor(365.25*yearNumber) - math.Floor(30.6001*monthNumber))
	year = int32(yearNumber + math.Floor((monthNumber-2.0)/12.0) - 4800)
	hour = (jd - math.Floor(jd+0.5) + 0.5) * 24.0
	return
}

// sweRevjul writes the swe_revjul outputs into the checked public buffers.
func (state *State) sweRevjul(jd float64, calendar int32, year, month, day []int32, hour []float64) {
	year[0], month[0], day[0], hour[0] = reverseJulianDay(jd, calendar)
}

// sweUtcTimeZone implements swe_utc_time_zone from swedate.c.
// sweUtcTimeZone shifts a civil clock by a zone offset in hours. A leap second
// is temporarily represented as second 59, then restored after shifting the date.
//
//	transform local time to UTC or UTC to local time
//
// input
//
//	iyear ... dsec     date and time
//	d_timezone		timezone offset
//
// output
//
//	iyear_out ... dsec_out
//
// For time zones east of Greenwich, d_timezone is positive.
// For time zones west of Greenwich, d_timezone is negative.
//
// For conversion from local time to utc, use +d_timezone.
// For conversion from utc to local time, use -d_timezone.
func (state *State) sweUtcTimeZone(year, month, day, hour, minute int32, second, zoneHours float64, outYear, outMonth, outDay, outHour, outMinute []int32, outSecond []float64) {
	leapSecond := second >= 60.0
	if leapSecond {
		second--
	}
	hours := float64(hour) + float64(minute)/60.0 + second/3600.0
	jd := state.sweJulday(year, month, day, 0, 1)
	hours -= zoneHours
	if hours < 0 {
		jd--
		hours += 24
	}
	if hours >= 24 {
		jd++
		hours -= 24
	}
	outYear[0], outMonth[0], outDay[0], _ = reverseJulianDay(jd+0.001, 1)
	outHour[0] = int32(hours)
	minutes := (hours - float64(outHour[0])) * 60
	outMinute[0] = int32(minutes)
	outSecond[0] = (minutes - float64(outMinute[0])) * 60
	if leapSecond {
		outSecond[0]++
	}
}

// initLeapsec implements init_leapsec from swedate.c.
//
//	Read additional leap second dates from external file, if given.
func (state *State) initLeapsec() int32 {
	var (
		fp       *textFile
		ndat     int32
		ndatLast int32
		tabsiz   int32 = 0
		i        int32
		s        [256]byte
		sp       []byte
	)
	if !(state.calendar.leapSecondsLoaded != 0) {
		state.calendar.leapSecondsLoaded = 1
		tabsiz = 27
		ndatLast = state.calendar.leapSeconds[27-1]
		if fp = state.swiFopen(-1, textBytes("seleapsec.txt"), state.positions.data.ephemerisPath[:], nil); fp == nil {
			return 27
		}
		defer closeTextFile(fp)
		for readLine(s[:], 256, fp) != nil {
			for sp = s[:]; int32(sp[0]) == ' ' || int32(sp[0]) == '\t'; {
				sp = sp[1:]
			}
			sp = sp[1:]
			if int32(sp[0]) == '#' || int32(sp[0]) == '\n' {
				continue
			}
			ndat = parseInt(s[:])
			if ndat <= ndatLast {
				continue
			}
			if tabsiz >= 100 {
				return tabsiz
			}
			state.calendar.leapSeconds[tabsiz] = ndat
			tabsiz++
		}
		if tabsiz > 27 {
			state.calendar.leapSeconds[tabsiz] = 0
		}
		return tabsiz
	}
	tabsiz = 0
	for i = 0; i < 100; i++ {
		if state.calendar.leapSeconds[i] == 0 {
			break
		} else {
			tabsiz++
		}
	}
	return tabsiz
}

// sweUtcToJd implements swe_utc_to_jd from swedate.c.
func (state *State) sweUtcToJd(iyearValue int32, imonthValue int32, idayValue int32, ihour int32, imin int32, dsec float64, gregflag int32, dret []float64, message []byte) int32 {
	var (
		iday        = [1]int32{idayValue}
		imonth      = [1]int32{imonthValue}
		iyear       = [1]int32{iyearValue}
		tjdUt1      float64
		jdTT        float64
		tjdEt1972   float64
		dhour       float64
		d           [1]float64
		iyear2      [1]int32
		imonth2     [1]int32
		iday2       [1]int32
		i           int32
		j           int32
		ndat        int32
		nleap       int32
		tabsizNleap int32
	)
	tjdUt1 = state.sweJulday(iyear[0], imonth[0], iday[0], 0, gregflag)
	state.sweRevjul(tjdUt1, gregflag, iyear2[:], imonth2[:], iday2[:], d[:])
	if iyear[0] != iyear2[0] || imonth[0] != imonth2[0] || iday[0] != iday2[0] {
		if message != nil {
			formatText(message, textBytes("invalid date: year = %d, month = %d, day = %d"), iyear[0], imonth[0], iday[0])
		}
		return -1
	}
	if ihour < 0 || ihour > 23 || imin < 0 || imin > 59 || dsec < 0 || dsec >= 61 || dsec >= 60 && (imin < 59 || ihour < 23 || tjdUt1 < 2441317.5) {
		if message != nil {
			formatText(message, textBytes("invalid time: %d:%d:%.2f"), ihour, imin, dsec)
		}
		return -1
	}
	dhour = float64(ihour) + float64(imin)/60.0 + dsec/3600.0
	if tjdUt1 < 2441317.5 {
		dret[1] = state.sweJulday(iyear[0], imonth[0], iday[0], dhour, gregflag)
		dret[0] = dret[1] + state.sweDeltatEx(dret[1], -1, nil)
		return 0
	}
	if gregflag == 0 {
		gregflag = 1
		state.sweRevjul(tjdUt1, gregflag, iyear[:], imonth[:], iday[:], d[:])
	}
	tabsizNleap = state.initLeapsec()
	nleap = 10
	ndat = iyear[0]*10000 + imonth[0]*100 + iday[0]
	for i = 0; i < tabsizNleap; i++ {
		if ndat <= state.calendar.leapSeconds[i] {
			break
		}
		nleap++
	}
	d[0] = state.sweDeltatEx(tjdUt1, -1, nil) * secondsPerDay
	if d[0]-float64(nleap)-32.184 >= 1.0 {
		dret[1] = tjdUt1 + dhour/24.0
		dret[0] = dret[1] + state.sweDeltatEx(dret[1], -1, nil)
		return 0
	}
	if dsec >= 60 {
		j = 0
		for i = 0; i < tabsizNleap; i++ {
			if ndat == state.calendar.leapSeconds[i] {
				j = 1
				break
			}
		}
		if j != 1 {
			if message != nil {
				formatText(message, textBytes("invalid time (no leap second!): %d:%d:%.2f"), ihour, imin, dsec)
			}
			return -1
		}
	}
	d[0] = tjdUt1 - 2441317.5
	d[0] += float64(ihour)/24.0 + float64(imin)/1440.0 + dsec/secondsPerDay
	tjdEt1972 = 2441317.5004882407
	jdTT = tjdEt1972 + d[0] + float64(nleap-10)/secondsPerDay
	d[0] = state.sweDeltatEx(jdTT, -1, nil)
	tjdUt1 = jdTT - state.sweDeltatEx(jdTT-d[0], -1, nil)
	tjdUt1 = jdTT - state.sweDeltatEx(tjdUt1, -1, nil)
	dret[0] = jdTT
	dret[1] = tjdUt1
	return 0
}

// sweJdetToUtc implements swe_jdet_to_utc from swedate.c.
func (state *State) sweJdetToUtc(jdTT float64, gregflag int32, iyear []int32, imonth []int32, iday []int32, ihour []int32, imin []int32, dsec []float64) {
	var (
		i           int32
		second60    int32 = 0
		iyear2      [1]int32
		imonth2     [1]int32
		iday2       [1]int32
		nleap       int32
		ndat        int32
		tabsizNleap int32
		d           [1]float64
		jd          float64
		tjdEt1972   float64
		jdUT        float64
		dret        [10]float64
	)
	tjdEt1972 = 2441317.5004882407
	d[0] = state.sweDeltatEx(jdTT, -1, nil)
	jdUT = jdTT - state.sweDeltatEx(jdTT-d[0], -1, nil)
	jdUT = jdTT - state.sweDeltatEx(jdUT, -1, nil)
	if jdTT < tjdEt1972 {
		state.sweRevjul(jdUT, gregflag, iyear, imonth, iday, d[:])
		ihour[0] = int32(d[0])
		d[0] -= float64(ihour[0])
		d[0] *= 60
		imin[0] = int32(d[0])
		dsec[0] = (d[0] - float64(imin[0])) * 60.0
		return
	}
	tabsizNleap = state.initLeapsec()
	state.sweRevjul(jdUT-1, 1, iyear2[:], imonth2[:], iday2[:], d[:])
	ndat = iyear2[0]*10000 + imonth2[0]*100 + iday2[0]
	nleap = 0
	for i = 0; i < tabsizNleap; i++ {
		if ndat <= state.calendar.leapSeconds[i] {
			break
		}
		nleap++
	}
	if nleap < tabsizNleap {
		i = state.calendar.leapSeconds[nleap]
		iyear2[0] = i / 10000
		imonth2[0] = i % 10000 / 100
		iday2[0] = i % 100
		jd = state.sweJulday(iyear2[0], imonth2[0], iday2[0], 0, 1)
		state.sweRevjul(jd+1, 1, iyear2[:], imonth2[:], iday2[:], d[:])
		state.sweUtcToJd(iyear2[0], imonth2[0], iday2[0], 0, 0, 0, 1, dret[:], nil)
		d[0] = jdTT - dret[0]
		if d[0] >= 0 {
			nleap++
		} else if d[0] < 0 && d[0] > -1.1574074074074073e-05 {
			second60 = 1
		}
	}
	jd = 2441317.5 + (jdTT - tjdEt1972) - (float64(nleap)+float64(second60))/secondsPerDay
	state.sweRevjul(jd, 1, iyear, imonth, iday, d[:])
	ihour[0] = int32(d[0])
	d[0] -= float64(ihour[0])
	d[0] *= 60
	imin[0] = int32(d[0])
	dsec[0] = (d[0]-float64(imin[0]))*60.0 + float64(second60)
	d[0] = state.sweDeltatEx(jdTT, -1, nil)
	d[0] = state.sweDeltatEx(jdTT-d[0], -1, nil)
	if d[0]*secondsPerDay-float64(nleap+10)-32.184 >= 1.0 {
		state.sweRevjul(jdTT-d[0], 1, iyear, imonth, iday, d[:])
		ihour[0] = int32(d[0])
		d[0] -= float64(ihour[0])
		d[0] *= 60
		imin[0] = int32(d[0])
		dsec[0] = (d[0] - float64(imin[0])) * 60.0
	}
	if gregflag == 0 {
		jd = state.sweJulday(iyear[0], imonth[0], iday[0], 0, 1)
		state.sweRevjul(jd, gregflag, iyear, imonth, iday, d[:])
	}
}

// sweJdut1ToUtc implements swe_jdut1_to_utc from swedate.c.
func (state *State) sweJdut1ToUtc(jdUT float64, calendar int32, iyear []int32, imonth []int32, iday []int32, ihour []int32, imin []int32, dsec []float64) {
	var jdTT float64 = jdUT + state.sweDeltatEx(jdUT, -1, nil)
	state.sweJdetToUtc(jdTT, calendar, iyear, imonth, iday, ihour, imin, dsec)
}
