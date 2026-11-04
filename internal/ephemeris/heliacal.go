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

// mymin implements mymin from swehel.c.
func (state *State) mymin(a float64, b float64) float64 {
	if a <= b {
		return a
	}
	return b
}

// mymax implements mymax from swehel.c.
//
// telescope diameter [mm]
func (state *State) mymax(a float64, b float64) float64 {
	if a >= b {
		return a
	}
	return b
}

// Tanh implements Tanh from swehel.c.
//
// ###################################################################
func (state *State) Tanh(x float64) float64 {
	return (math.Exp(x) - math.Exp(-x)) / (math.Exp(x) + math.Exp(-x))
}

// CVA implements CVA from swehel.c.
//
// ' B [nL]
// ' SN [-]
// ' CVA [deg]
func (state *State) CVA(B float64, SN float64, helflag int32) float64 {
	var isScotopic int32 = 0
	if B < 1394 {
		isScotopic = 1
	}
	if helflag&(1<<14) != 0 {
		isScotopic = 0
	}
	if helflag&(1<<15) != 0 {
		isScotopic = 1
	}
	if isScotopic != 0 {
		return state.mymin(900, 380/SN*math.Pow(10, 0.3*math.Pow(B, -0.29))) / 60.0 / 60.0
	} else {
		return 40.0 / SN * math.Pow(10, 8.28*math.Pow(B, -0.29)) / 60.0 / 60.0
	}
}

// PupilDia implements PupilDia from swehel.c.
//
// ' age [year]
// ' B [nL]
// ' PupilDia [mm]
func (state *State) PupilDia(Age float64, B float64) float64 {
	return (0.534 - 0.00211*Age - (0.236-0.00127*Age)*state.Tanh(0.4*math.Log(B)/math.Log(10)-2.2)) * 10
}

// OpticFactor implements OpticFactor from swehel.c.
//
// 'Input
// ' Bback [nL]
// ' kX [-]
// ' Binocular [-]
// ' OpticMag [-]
// ' OpticDia [mm]
// ' OpticTrans [-]
// ' JDNDaysUT [JDN]
// ' Age [Year]
// ' SN [-]
// ' ObjectName
// ' TypeFactor [0=itensity factor 1=background factor]
// 'Output
// ' OpticFactor [-]
func (state *State) OpticFactor(Bback float64, kX float64, dobs []float64, JDNDaysUT float64, ObjectName []byte, TypeFactor int32, helflag int32) float64 {
	var (
		Pst        float64
		CIb        float64
		CIi        float64
		ObjectSize float64
		Fb         float64
		Fe         float64
		Fsc        float64
		Fci        float64
		Fcb        float64
		Ft         float64
		Fp         float64
		Fa         float64
		Fr         float64
		Fm         float64
		Age        float64 = dobs[0]
		SN         float64 = dobs[1]
		SNi        float64
		Binocular  float64 = dobs[2]
		OpticMag   float64 = dobs[3]
		OpticDia   float64 = dobs[4]
		OpticTrans float64 = dobs[5]
		isScotopic int32   = 0
	)
	JDNDaysUT += 0.0
	SNi = SN
	if SNi <= 0.00000001 {
		SNi = 0.00000001
	}
	Pst = state.PupilDia(23, Bback)
	if OpticMag == 1 {
		OpticTrans = 1
		OpticDia = Pst
	}
	CIb = 0.7
	CIi = 0.5
	ObjectSize = 0
	if compareText(ObjectName, textBytes("moon")) == 0 {
	}
	Fb = 1
	if Binocular == 0 {
		Fb = 1.41
	}
	if Bback < 1645 {
		isScotopic = 1
	}
	if helflag&(1<<14) != 0 {
		isScotopic = 0
	}
	if helflag&(1<<15) != 0 {
		isScotopic = 1
	}
	if isScotopic != 0 {
		Fe = math.Pow(10, 0.48*kX)
		Fsc = state.mymin(1, (1-math.Pow(Pst/124.4, 4))/(1-math.Pow(OpticDia/OpticMag/124.4, 4)))
		Fci = math.Pow(10, -0.4*(1-CIi/2.0))
		Fcb = math.Pow(10, -0.4*(1-CIb/2.0))
	} else {
		Fe = math.Pow(10, 0.4*kX)
		Fsc = state.mymin(1, math.Pow(OpticDia/OpticMag/Pst, 2)*(1-math.Exp(-math.Pow(Pst/6.2, 2)))/(1-math.Exp(-math.Pow(OpticDia/OpticMag/6.2, 2))))
		Fci = 1
		Fcb = 1
	}
	Ft = 1 / OpticTrans
	Fp = state.mymax(1, math.Pow(Pst/(OpticMag*state.PupilDia(Age, Bback)), 2))
	Fa = math.Pow(Pst/OpticDia, 2)
	Fr = (1 + 0.03*math.Pow(OpticMag*ObjectSize/state.CVA(Bback, SNi, helflag), 2)) / math.Pow(SNi, 2)
	Fm = math.Pow(OpticMag, 2)
	if TypeFactor == 0 {
		return Fb * Fe * Ft * Fp * Fa * Fr * Fsc * Fci
	} else {
		return Fb * Ft * Fp * Fa * Fm * Fsc * Fcb
	}
}

// DeterObject implements DeterObject from swehel.c.
//
// ###################################################################
func (state *State) DeterObject(ObjectName []byte) int32 {
	var (
		s    [256]byte
		sp   []byte
		body int32
	)
	copyText(s[:], ObjectName)
	for sp = s[:]; int32(sp[0]) != '\000'; sp = sp[1:] {
		sp[0] = byte(lowerByte(sp[0]))
	}
	if compareTextN(s[:], textBytes("sun"), 3) == 0 {
		return 0
	}
	if compareTextN(s[:], textBytes("venus"), 5) == 0 {
		return 3
	}
	if compareTextN(s[:], textBytes("mars"), 4) == 0 {
		return 4
	}
	if compareTextN(s[:], textBytes("mercur"), 6) == 0 {
		return 2
	}
	if compareTextN(s[:], textBytes("jupiter"), 7) == 0 {
		return 5
	}
	if compareTextN(s[:], textBytes("saturn"), 6) == 0 {
		return 6
	}
	if compareTextN(s[:], textBytes("uranus"), 6) == 0 {
		return 7
	}
	if compareTextN(s[:], textBytes("neptun"), 6) == 0 {
		return 8
	}
	if compareTextN(s[:], textBytes("moon"), 4) == 0 {
		return 1
	}
	if body = parseInt(s[:]); body > 0 {
		body += 10000
		return body
	}
	return -1
}

// callSweFixstar implements call_swe_fixstar from swehel.c.
//
//	avoids problems with star name string that may be overwritten by
//	  swe_fixstar()
func (state *State) callSweFixstar(star []byte, jd float64, flags int32, xx []float64, message []byte) int32 {
	var (
		retval int32
		star2  [256]byte
	)
	copyText(star2[:], star)
	retval = state.sweFixstar(star2[:], jd, flags, xx, message)
	return retval
}

// callSweFixstarMag implements call_swe_fixstar_mag from swehel.c.
//
//	avoids problems with star name string that may be overwritten by
//	  swe_fixstar_mag()
func (state *State) callSweFixstarMag(star []byte, mag []float64, message []byte) int32 {
	var (
		retval int32
		star2  [256]byte
	)
	if compareText(star, state.heliacal.callSweFixstarMagStarSave[:]) == 0 {
		mag[0] = state.heliacal.callSweFixstarMagDmag[0]
		return 0
	}
	copyText(state.heliacal.callSweFixstarMagStarSave[:], star)
	copyText(star2[:], star)
	retval = state.sweFixstarMag(star2[:], state.heliacal.callSweFixstarMagDmag[:], message)
	mag[0] = state.heliacal.callSweFixstarMagDmag[0]
	return retval
}

// callSweRiseTrans implements call_swe_rise_trans from swehel.c.
//
//	avoids problems with star name string that may be overwritten by
//	  swe_fixstar()
func (state *State) callSweRiseTrans(jd float64, body int32, star []byte, helflag int32, eventtype int32, location []float64, pressure float64, temperature float64, tret []float64, message []byte) int32 {
	var (
		retval int32
		flags  int32 = helflag & (1 | 2 | 4)
		star2  [256]byte
	)
	copyText(star2[:], star)
	retval = state.sweRiseTrans(jd, body, star2[:], flags, eventtype, location, pressure, temperature, tret, message)
	return retval
}

// calcRiseAndSet implements calc_rise_and_set from swehel.c.
func (state *State) calcRiseAndSet(tjdStart float64, body int32, location []float64, datm []float64, eventflag int32, helflag int32, trise []float64, message []byte) int32 {
	var (
		status    int32 = 0
		i         int32
		sda       float64
		xs        [6]float64
		xx        [6]float64
		xaz       [6]float64
		xaz2      [6]float64
		dfac      float64 = 0.0027378507871321013
		rdi       float64
		rh        float64
		initialJD float64 = tjdStart
		tjdrise   float64
		tjdnoon   float64 = float64(int32(initialJD)) - location[0]/15.0/24.0
		flags     int32   = helflag & (1 | 2 | 4)
		ephemeris int32   = flags
	)
	flags |= flagEquatorial
	if !(helflag&256 != 0) {
		flags |= flagTruePosition | flagNoNutation
	}
	if state.sweCalcUt(initialJD, 0, flags, xs[:], message) == 0 {
		if message != nil {
			copyText(message, textBytes("error in calc_rise_and_set(): calc(sun) failed "))
		}
		return -1
	}
	if state.sweCalcUt(initialJD, body, flags, xx[:], message) == 0 {
		if message != nil {
			copyText(message, textBytes("error in calc_rise_and_set(): calc(sun) failed "))
		}
		return -1
	}
	tjdnoon -= state.sweDegnorm(xs[0]-xx[0])/360.0 + 0
	state.sweAzalt(initialJD, 1, location, datm[0], datm[1], xx[:], xaz[:])
	if eventflag&1 != 0 {
		if xaz[2] > 0 {
			for tjdnoon-initialJD < 0.5 {
				tjdnoon += 1
			}
			for tjdnoon-initialJD > 1.5 {
				tjdnoon -= 1
			}
		} else {
			for tjdnoon-initialJD < 0.0 {
				tjdnoon += 1
			}
			for tjdnoon-initialJD > 1.0 {
				tjdnoon -= 1
			}
		}
	} else if xaz[2] > 0 {
		for initialJD-tjdnoon > 0.5 {
			tjdnoon += 1
		}
		for initialJD-tjdnoon < -0.5 {
			tjdnoon -= 1
		}
	} else {
		for initialJD-tjdnoon > 0.0 {
			tjdnoon += 1
		}
		for initialJD-tjdnoon < -1.0 {
			tjdnoon -= 1
		}
	}
	if state.sweCalcUt(tjdnoon, body, flags, xx[:], message) == -1 {
		if message != nil {
			copyText(message, textBytes("error in calc_rise_and_set(): calc(sun) failed "))
		}
		return -1
	}
	rdi = 0
	if body == 0 {
		rdi = math.Asin(0.004652472637378737/xx[2]) / radiansPerDegree
	} else if body == 1 {
		rdi = math.Asin(1.161112783207883e-05/xx[2]) / radiansPerDegree
	}
	if eventflag&256 != 0 {
		rdi = 0
	}
	rh = -(0.57499999999999996 + rdi)
	sda = math.Acos(-math.Tan(location[1]*radiansPerDegree)*math.Tan(xx[1]*radiansPerDegree)) * degreesPerRadian
	if eventflag&1 != 0 {
		tjdrise = tjdnoon - sda/360.0
	} else {
		tjdrise = tjdnoon + sda/360.0
	}
	flags = ephemeris | 256 | 2*1024
	if body == 1 {
		flags |= flagTopocentric
	}
	if !(helflag&256 != 0) {
		flags |= flagTruePosition | flagNoNutation
	}
	for i = 0; i < 2; i++ {
		if state.sweCalcUt(tjdrise, body, flags, xx[:], message) == -1 {
			return -1
		}
		state.sweAzalt(tjdrise, 1, location, datm[0], datm[1], xx[:], xaz[:])
		xx[0] -= xx[3] * dfac
		xx[1] -= xx[4] * dfac
		state.sweAzalt(tjdrise-dfac, 1, location, datm[0], datm[1], xx[:], xaz2[:])
		tjdrise -= (xaz[1] - rh) / (xaz[1] - xaz2[1]) * dfac
	}
	trise[0] = tjdrise
	return status
}

// myRiseTrans implements my_rise_trans from swehel.c.
//
// fprintf(stderr, "%f\n", ph->trise);
func (state *State) myRiseTrans(jd float64, body int32, starname []byte, eventtype int32, helflag int32, location []float64, datm []float64, tret []float64, message []byte) int32 {
	var status int32 = 0
	if starname != nil && int32(starname[0]) != '\000' {
		body = state.DeterObject(starname)
	}
	if body != -1 && math.Abs(location[1]) < 63 {
		status = state.calcRiseAndSet(jd, body, location, datm, eventtype, helflag, tret, message)
	} else {
		status = state.callSweRiseTrans(jd, body, starname, helflag, eventtype, location, datm[0], datm[1], tret, message)
	}
	return status
}

// RiseSet implements RiseSet from swehel.c.
//
// ###################################################################
// ' JDNDaysUT [Days]
// ' dgeo [array: longitude, latitude, eye height above sea m]
// ' TempE [C]
// ' PresE [mbar]
// ' ObjectName (string)
// ' RSEvent (1=rise, 2=set,3=up transit,4=down transit)
// ' Rim [0=center,1=top]
// ' RiseSet [Day]
func (state *State) RiseSet(JDNDaysUT float64, location []float64, datm []float64, ObjectName []byte, RSEvent int32, helflag int32, Rim int32, tret []float64, message []byte) int32 {
	var (
		eventtype int32 = RSEvent
		Planet    int32
		retval    int32
	)
	if Rim == 0 {
		eventtype |= 256
	}
	Planet = state.DeterObject(ObjectName)
	if Planet != -1 {
		retval = state.myRiseTrans(JDNDaysUT, Planet, textBytes(""), eventtype, helflag, location, datm, tret, message)
	} else {
		retval = state.myRiseTrans(JDNDaysUT, -1, ObjectName, eventtype, helflag, location, datm, tret, message)
	}
	return retval
}

// SunRA implements SunRA from swehel.c.
//
// ###################################################################
// ' JDNDaysUT [Days]
// ' actual [0= approximation, 1=actual]
// ' SunRA [deg]
func (state *State) SunRA(JDNDaysUT float64, helflag int32, message []byte) float64 {
	var (
		imon    [1]int32
		iday    [1]int32
		iyar    [1]int32
		calflag int32 = 1
		dut     [1]float64
	)
	helflag += 0
	message[0] = byte('\000')
	if JDNDaysUT == state.heliacal.sunRATjdlast {
		return state.heliacal.sunRARalast
	}
	state.sweRevjul(JDNDaysUT, calflag, iyar[:], imon[:], iday[:], dut[:])
	state.heliacal.sunRATjdlast = JDNDaysUT
	state.heliacal.sunRARalast = state.sweDegnorm((float64(imon[0]) + float64(iday[0]-1)/30.4 - 3.69) * 30)
	return state.heliacal.sunRARalast
}

// Kelvin implements Kelvin from swehel.c.
//
// ###################################################################
// ' Temp [C]
// ' Kelvin [K]
func (state *State) Kelvin(Temp float64) float64 {
	return Temp + 273.15
}

// TopoAltfromAppAlt implements TopoAltfromAppAlt from swehel.c.
//
// ###################################################################
// ' AppAlt [deg]
// ' TempE [C]
// ' PresE [mbar]
// ' TopoAltitudefromAppAlt [deg]
func (state *State) TopoAltfromAppAlt(AppAlt float64, TempE float64, PresE float64) float64 {
	var (
		R      float64 = 0
		retalt float64 = 0
	)
	if AppAlt >= -3.5 {
		if AppAlt > 17.904104638432 {
			R = 0.97 / math.Tan(AppAlt*radiansPerDegree)
		} else {
			R = (34.46 + 4.23*AppAlt + 0.004*AppAlt*AppAlt) / (1 + 0.505*AppAlt + 0.0845*AppAlt*AppAlt)
		}
		R = (PresE - 80) / 930 / (1 + 0.00008*(R+39)*(TempE-10)) * R
		retalt = AppAlt - R*0.016666666666666666
	} else {
		retalt = AppAlt
	}
	return retalt
}

// AppAltfromTopoAlt implements AppAltfromTopoAlt from swehel.c.
//
// ###################################################################
// ' TopoAlt [deg]
// ' TempE [C]
// ' PresE [mbar]
// ' AppAltfromTopoAlt [deg]
// ' call this instead of swe_azalt(), because it is faster (lower precision
// ' is required)
func (state *State) AppAltfromTopoAlt(TopoAlt float64, TempE float64, PresE float64, helflag int32) float64 {
	var (
		i          int32
		nloop      int32   = 2
		newAppAlt  float64 = TopoAlt
		newTopoAlt float64 = 0.0
		oudAppAlt  float64 = newAppAlt
		oudTopoAlt float64 = newTopoAlt
		verschil   float64
		retalt     float64
	)
	if helflag&256 != 0 {
		nloop = 5
	}
	for i = 0; i <= nloop; i++ {
		newTopoAlt = newAppAlt - state.TopoAltfromAppAlt(newAppAlt, TempE, PresE)
		verschil = newAppAlt - oudAppAlt
		oudAppAlt = newTopoAlt - oudTopoAlt - verschil
		if verschil != 0 && oudAppAlt != 0 {
			verschil = newAppAlt - verschil*(TopoAlt+newTopoAlt-newAppAlt)/oudAppAlt
		} else {
			verschil = TopoAlt + newTopoAlt
		}
		oudAppAlt = newAppAlt
		oudTopoAlt = newTopoAlt
		newAppAlt = verschil
	}
	retalt = TopoAlt + newTopoAlt
	if retalt < -3.5 {
		retalt = TopoAlt
	}
	return retalt
}

// HourAngle implements HourAngle from swehel.c.
//
// ###################################################################
// ' TopoAlt [deg]
// ' TopoDecl [deg]
// ' Lat [deg]
// ' HourAngle [hour]
func (state *State) HourAngle(TopoAlt float64, TopoDecl float64, Lat float64) float64 {
	var (
		Alti  float64 = TopoAlt * radiansPerDegree
		decli float64 = TopoDecl * radiansPerDegree
		Lati  float64 = Lat * radiansPerDegree
		ha    float64 = (math.Sin(Alti) - math.Sin(Lati)*math.Sin(decli)) / math.Cos(Lati) / math.Cos(decli)
	)
	if ha < float64(-1) {
		ha = float64(-1)
	}
	if ha > 1 {
		ha = 1
	}
	return math.Acos(ha) / radiansPerDegree / 15.0
}

// ObjectLoc implements ObjectLoc from swehel.c.
//
// ###################################################################
// ' JDNDaysUT [Days]
// ' dgeo [array: longitude, latitude, eye height above sea m]
// ' TempE [C]
// ' PresE [mbar]
// ' ObjectName [-]
// ' Angle (0 = TopoAlt, 1 = Azi, 2=Topo Declination, 3=Topo Rectascension, 4=AppAlt,5=Geo Declination, 6=Geo Rectascension)
// ' ObjectLoc [deg]
func (state *State) ObjectLoc(JDNDaysUT float64, location []float64, datm []float64, ObjectName []byte, Angle int32, helflag int32, dret []float64, message []byte) int32 {
	var (
		x         [6]float64
		xin       [3]float64
		xaz       [3]float64
		tjdTt     float64
		Planet    int32
		ephemeris int32
		flags     int32 = 2 * 1024
	)
	ephemeris = helflag & (1 | 2 | 4)
	flags |= ephemeris
	if !(helflag&256 != 0) {
		flags |= flagTruePosition | flagNoNutation
	}
	if Angle < 5 {
		flags = flags | (flagTopocentric)
	}
	if Angle == 7 {
		Angle = 0
	}
	tjdTt = JDNDaysUT + state.sweDeltatEx(JDNDaysUT, ephemeris, message)
	Planet = state.DeterObject(ObjectName)
	if Planet != -1 {
		if state.sweCalc(tjdTt, Planet, flags, x[:], message) == -1 {
			return -1
		}
	} else if state.callSweFixstar(ObjectName, tjdTt, flags, x[:], message) == -1 {
		return -1
	}
	if Angle == 2 || Angle == 5 {
		dret[0] = x[1]
	} else if Angle == 3 || Angle == 6 {
		dret[0] = x[0]
	} else {
		xin[0] = x[0]
		xin[1] = x[1]
		state.sweAzalt(JDNDaysUT, 1, location, datm[0], datm[1], xin[:], xaz[:])
		if Angle == 0 {
			dret[0] = xaz[1]
		}
		if Angle == 4 {
			dret[0] = state.AppAltfromTopoAlt(xaz[1], datm[0], datm[1], helflag)
		}
		if Angle == 1 {
			xaz[0] += 180
			if xaz[0] >= 360 {
				xaz[0] -= 360
			}
			dret[0] = xaz[0]
		}
	}
	return 0
}

// azaltCart implements azalt_cart from swehel.c.
//
// ###################################################################
// ' JDNDaysUT [Days]
// ' dgeo [array: longitude, latitude, eye height above sea m]
// ' TempE [C]
// ' PresE [mbar]
// ' ObjectName [-]
// ' Angle (0 = TopoAlt, 1 = Azi, 2=Topo Declination, 3=Topo Rectascension, 4=AppAlt,5=Geo Declination, 6=Geo Rectascension)
// ' ObjectLoc [deg]
func (state *State) azaltCart(JDNDaysUT float64, location []float64, datm []float64, ObjectName []byte, helflag int32, dret []float64, message []byte) int32 {
	var (
		x         [6]float64
		xin       [3]float64
		xaz       [3]float64
		tjdTt     float64
		Planet    int32
		ephemeris int32
		flags     int32 = 2 * 1024
	)
	ephemeris = helflag & (1 | 2 | 4)
	flags |= ephemeris
	if !(helflag&256 != 0) {
		flags |= flagTruePosition | flagNoNutation
	}
	flags = flags | (flagTopocentric)
	tjdTt = JDNDaysUT + state.sweDeltatEx(JDNDaysUT, ephemeris, message)
	Planet = state.DeterObject(ObjectName)
	if Planet != -1 {
		if state.sweCalc(tjdTt, Planet, flags, x[:], message) == -1 {
			return -1
		}
	} else if state.callSweFixstar(ObjectName, tjdTt, flags, x[:], message) == -1 {
		return -1
	}
	xin[0] = x[0]
	xin[1] = x[1]
	state.sweAzalt(JDNDaysUT, 1, location, datm[0], datm[1], xin[:], xaz[:])
	dret[0] = xaz[0]
	dret[1] = xaz[1]
	dret[2] = xaz[2]
	xaz[1] = xaz[2]
	xaz[2] = 1
	state.swiPolcart(xaz[:], xaz[:])
	dret[3] = xaz[0]
	dret[4] = xaz[1]
	dret[5] = xaz[2]
	return 0
}

// DistanceAngle implements DistanceAngle from swehel.c.
//
// ###################################################################
// ' LatA [rad]
// ' LongA [rad]
// ' LatB [rad]
// ' LongB [rad]
// ' DistanceAngle [rad]
func (state *State) DistanceAngle(LatA float64, LongA float64, LatB float64, LongB float64) float64 {
	var (
		dlon     float64 = LongB - LongA
		dlat     float64 = LatB - LatA
		sindlat2 float64 = math.Sin(dlat / 2)
		sindlon2 float64 = math.Sin(dlon / 2)
		corde    float64 = sindlat2*sindlat2 + math.Cos(LatA)*math.Cos(LatB)*sindlon2*sindlon2
	)
	if corde > 1 {
		corde = 1
	}
	return 2 * math.Asin(math.Sqrt(corde))
}

// kW implements kW from swehel.c.
//
// ###################################################################
// ' heighteye [m]
// ' TempS [C]
// ' RH [%]
// ' kW [-]
func (state *State) kW(HeightEye float64, TempS float64, RH float64) float64 {
	var WT float64 = 0.031
	WT *= 0.94 * (RH / 100.0) * math.Exp(TempS/15) * math.Exp(float64(-1)*HeightEye/3000.0)
	return WT
}

// kOZ implements kOZ from swehel.c.
//
// ###################################################################
// ' JDNDaysUT [-]
// ' AltS [deg]
// ' lat [deg]
// ' kOZ [-]
func (state *State) kOZ(AltS float64, sunra float64, Lat float64) float64 {
	var (
		CHANGEKO float64
		OZ       float64
		LT       float64
		kOZret   float64
		altslim  float64 = 0
	)
	if AltS == state.heliacal.kOZAltsLast && sunra == state.heliacal.kOZSunraLast {
		return state.heliacal.kOZKozLast
	}
	state.heliacal.kOZAltsLast = AltS
	state.heliacal.kOZSunraLast = sunra
	OZ = 0.031
	LT = Lat * radiansPerDegree
	kOZret = OZ * (3.0 + 0.4*(LT*math.Cos(sunra*radiansPerDegree)-math.Cos(3*LT))) / 3.0
	altslim = -AltS - 12
	if altslim < 0 {
		altslim = 0
	}
	CHANGEKO = (100 - 11.6*state.mymin(6, altslim)) / 100
	if 0 != 0 {
		if state.heliacal.kOZA == 0 {
			printText(textBytes("bsk=%f %f\n"), kOZret, AltS)
		}
		state.heliacal.kOZA = 1
	}
	state.heliacal.kOZKozLast = kOZret * CHANGEKO
	return state.heliacal.kOZKozLast
}

// kR implements kR from swehel.c.
//
// ###################################################################
// ' AltS [deg]
// ' heighteye [m]
// ' kR [-]
func (state *State) kR(AltS float64, HeightEye float64) float64 {
	var (
		CHANGEK float64
		LAMBDA  float64
		val     float64 = -AltS - 12
	)
	if val < 0 {
		val = 0
	}
	if val > 6 {
		val = 6
	}
	CHANGEK = 1 - 0.166667*val
	LAMBDA = 0.55 + (CHANGEK-1)*0.04
	return 0.1066 * math.Exp(float64(-1)*HeightEye/8515.0) * math.Pow(LAMBDA/0.55, float64(-4))
}

// Sgn implements Sgn from swehel.c.
//
//	From Schaefer , Archaeoastronomy, XV, 2000, page 128
func (state *State) Sgn(x float64) int32 {
	if x < 0 {
		return -1
	}
	return 1
}

// ka implements ka from swehel.c.
//
// ###################################################################
// ' JDNDaysUT [-]
// ' AltS [deg]
// ' lat [deg]
// ' heighteye [m]
// ' TempS [C]
// ' RH [%]
// ' VR [km]
// ' ka [-]
func (state *State) ka(AltS float64, sunra float64, Lat float64, HeightEye float64, TempS float64, RH float64, VR float64, message []byte) float64 {
	var (
		CHANGEKA float64
		LAMBDA   float64
		BetaVr   float64
		Betaa    float64
		kaact    float64
		SL       float64 = float64(state.Sgn(Lat))
	)
	if AltS == state.heliacal.kaAltsLast && sunra == state.heliacal.kaSunraLast {
		return state.heliacal.kaKaLast
	}
	state.heliacal.kaAltsLast = AltS
	state.heliacal.kaSunraLast = sunra
	CHANGEKA = 1 - 0.166667*state.mymin(6, state.mymax(-AltS-12, 0))
	LAMBDA = 0.55 + (CHANGEKA-1)*0.04
	if VR != 0 {
		if VR >= 1 {
			BetaVr = 3.912 / VR
			Betaa = BetaVr - (state.kW(HeightEye, TempS, RH)/3000.0+state.kR(AltS, HeightEye)/8515.0)*1000*0.921034037197618
			kaact = Betaa * 3745.0 / 1000 * 1 / 0.921034037197618
			if kaact < 0 {
				if message != nil {
					copyText(message, textBytes("The provided Meteorological range is too long, when taking into acount other atmospheric parameters"))
				}
			}
		} else {
			kaact = VR - state.kW(HeightEye, TempS, RH) - state.kR(AltS, HeightEye) - state.kOZ(AltS, sunra, Lat)
			if kaact < 0 {
				if message != nil {
					copyText(message, textBytes("The provided atmosphic coeefficent (ktot) is too low, when taking into acount other atmospheric parameters"))
				}
			}
		}
	} else {
		if RH <= 0.00000001 {
			RH = 0.00000001
		}
		if RH >= 99.99999999 {
			RH = 99.99999999
		}
		kaact = 0.1 * math.Exp(float64(-1)*HeightEye/3745.0) * math.Pow(1-0.32/math.Log(RH/100.0), 1.33) * (1 + 0.33*SL*math.Sin(sunra*radiansPerDegree))
		kaact = kaact * math.Pow(LAMBDA/0.55, -1.3)
	}
	state.heliacal.kaKaLast = kaact
	return kaact
}

// kt implements kt from swehel.c.
//
// ###################################################################
// ' JDNDaysUT [-]
// ' AltS [deg]
// ' lat [deg]
// ' heighteye [m]
// ' TempS [C]
// ' RH [%]
// ' VR [km]
// ' ExtType [0=ka,1=kW,2=kR,3=kOZ,4=ktot]
// ' kt [-]
func (state *State) kt(AltS float64, sunra float64, Lat float64, HeightEye float64, TempS float64, RH float64, VR float64, ExtType int32, message []byte) float64 {
	var (
		kRact  float64 = 0
		kWact  float64 = 0
		kOZact float64 = 0
		kaact  float64 = 0
	)
	if ExtType == 2 || ExtType == 4 {
		kRact = state.kR(AltS, HeightEye)
	}
	if ExtType == 1 || ExtType == 4 {
		kWact = state.kW(HeightEye, TempS, RH)
	}
	if ExtType == 3 || ExtType == 4 {
		kOZact = state.kOZ(AltS, sunra, Lat)
	}
	if ExtType == 0 || ExtType == 4 {
		kaact = state.ka(AltS, sunra, Lat, HeightEye, TempS, RH, VR, message)
	}
	if kaact < 0 {
		kaact = 0
	}
	return kWact + kRact + kOZact + kaact
}

// Airmass implements Airmass from swehel.c.
//
// ###################################################################
// ' AppAlt0 [deg]
// ' PresS [mbar]
// ' Airmass [??]
func (state *State) Airmass(AppAltO float64, Press float64) float64 {
	var (
		airm float64
		zend float64
	)
	zend = (90 - AppAltO) * radiansPerDegree
	if zend > 1.5707963267948966 {
		zend = 1.5707963267948966
	}
	airm = 1 / (math.Cos(zend) + 0.025*math.Exp(float64(-11)*math.Cos(zend)))
	return Press / 1013 * airm
}

// Xext implements Xext from swehel.c.
//
// ###################################################################
// ' scaleH '[m]
// ' zend [rad]
// ' PresS [mbar]
// ' Xext [-]
func (state *State) Xext(scaleH float64, zend float64, Press float64) float64 {
	return Press / 1013.0 / (math.Cos(zend) + 0.01*math.Sqrt(scaleH/1000.0)*math.Exp(-30.0/math.Sqrt(scaleH/1000.0)*math.Cos(zend)))
}

// Xlay implements Xlay from swehel.c.
//
// ###################################################################
// ' scaleH '[m]
// ' zend [rad]
// ' PresS [mbar]
// ' Xlay [-]
func (state *State) Xlay(scaleH float64, zend float64, Press float64) float64 {
	var a float64 = math.Sin(zend) / (1.0 + scaleH/6378136.6)
	return Press / 1013.0 / math.Sqrt(1.0-a*a)
}

// TempEfromTempS implements TempEfromTempS from swehel.c.
//
// ###################################################################
// ' Meteorological formula
// '###################################################################
// ' TempS [C]
// ' HeightEye [m]
// ' TempEfromTempS [C]
func (state *State) TempEfromTempS(TempS float64, HeightEye float64, Lapse float64) float64 {
	return TempS - Lapse*HeightEye
}

// PresEfromPresS implements PresEfromPresS from swehel.c.
//
// ###################################################################
// ' TempS [C]
// ' PresS [mbar]
// ' HeightEye [m]
// ' PresEfromPresS [mbar]
func (state *State) PresEfromPresS(TempS float64, Press float64, HeightEye float64) float64 {
	return Press * math.Exp(-0.28404373325999999/(state.Kelvin(TempS)+3.25*HeightEye/1000)/8.31441*HeightEye)
}

// Deltam implements Deltam from swehel.c.
//
// ###################################################################
// ' AltO [deg]
// ' JDNDaysUT [-]
// ' AltS [deg]
// ' lat [deg]
// ' heighteye [m]
// ' TempS [C]
// ' PresS [mbar]
// ' RH [%]
// ' VR [km]
// ' Deltam [-]
func (state *State) Deltam(AltO float64, AltS float64, sunra float64, Lat float64, HeightEye float64, datm []float64, helflag int32, message []byte) float64 {
	var (
		zend    float64
		xR      float64
		XW      float64
		Xa      float64
		XOZ     float64
		PresE   float64 = state.PresEfromPresS(datm[1], datm[0], HeightEye)
		TempE   float64 = state.TempEfromTempS(datm[1], HeightEye, 0.0065)
		AppAltO float64 = state.AppAltfromTopoAlt(AltO, TempE, PresE, helflag)
		deltam  float64
	)
	if AltS == state.heliacal.deltamAltsLast && AltO == state.heliacal.deltamAltoLast && sunra == state.heliacal.deltamSunraLast {
		return state.heliacal.deltamDeltamLast
	}
	state.heliacal.deltamAltsLast = AltS
	state.heliacal.deltamAltoLast = AltO
	state.heliacal.deltamSunraLast = sunra
	if 0 == 0 {
		zend = (90 - AppAltO) * radiansPerDegree
		if zend > 1.5707963267948966 {
			zend = 1.5707963267948966
		}
		xR = state.Xext(8515.0, zend, datm[0])
		XW = state.Xext(3000.0, zend, datm[0])
		Xa = state.Xext(3745.0, zend, datm[0])
		XOZ = state.Xlay(20000.0, zend, datm[0])
		deltam = state.kR(AltS, HeightEye)*xR + state.kt(AltS, sunra, Lat, HeightEye, datm[1], datm[2], datm[3], 0, message)*Xa + state.kOZ(AltS, sunra, Lat)*XOZ + state.kW(HeightEye, datm[1], datm[2])*XW
	} else {
		deltam = state.kt(AltS, sunra, Lat, HeightEye, datm[1], datm[2], datm[3], 4, message) * state.Airmass(AppAltO, datm[0])
	}
	state.heliacal.deltamDeltamLast = deltam
	return deltam
}

// Bn implements Bn from swehel.c.
//
// ###################################################################
// ' AltO [deg]
// ' JDNDaysUT [-]
// ' AltS [deg]
// ' lat [deg]
// ' heighteye [m]
// ' TempS [C]
// ' PresS [mbar]
// ' RH [%]
// ' VR [km]
// ' Bn [nL]
func (state *State) Bn(AltO float64, JDNDayUT float64, AltS float64, sunra float64, Lat float64, HeightEye float64, datm []float64, helflag int32, message []byte) float64 {
	var (
		PresE   float64 = state.PresEfromPresS(datm[1], datm[0], HeightEye)
		TempE   float64 = state.TempEfromTempS(datm[1], HeightEye, 0.0065)
		AppAltO float64 = state.AppAltfromTopoAlt(AltO, TempE, PresE, helflag)
		zend    float64
		YearB   float64
		MonthB  float64
		DayB    float64
		Bna     float64
		kX      float64
		Bnb     float64
		B0      float64 = 0.0000000000001
		dut     [1]float64
		iyar    [1]int32
		imon    [1]int32
		iday    [1]int32
	)
	if AppAltO < 10 {
		AppAltO = 10
	}
	zend = (90 - AppAltO) * radiansPerDegree
	state.sweRevjul(JDNDayUT, 1, iyar[:], imon[:], iday[:], dut[:])
	YearB = float64(iyar[0])
	MonthB = float64(imon[0])
	DayB = float64(iday[0])
	Bna = B0 * (1 + 0.3*math.Cos(6.283*(YearB+((DayB-1)/30.4+MonthB-1)/12-1990.33)/11.1))
	kX = state.Deltam(AltO, AltS, sunra, Lat, HeightEye, datm, helflag, message)
	Bnb = Bna * (0.4 + 0.6/math.Sqrt(1-0.96*math.Pow(math.Sin(zend), 2))) * math.Pow(10, -0.4*kX)
	return state.mymax(Bnb, 0) * 980392156862745.12
}

// Magnitude implements Magnitude from swehel.c.
//
// ###################################################################
// ' JDNDaysUT [-]
// ' dgeo [array: longitude, latitude, eye height above sea m]
// ' TempE [C]
// ' PresE [mbar]
// ' ObjectName [-]
// ' Magnitude [-]
func (state *State) Magnitude(JDNDaysUT float64, location []float64, ObjectName []byte, helflag int32, dmag []float64, message []byte) int32 {
	var (
		x         [20]float64
		Planet    int32
		flags     int32
		ephemeris int32
	)
	ephemeris = helflag & (1 | 2 | 4)
	dmag[0] = -99.0
	Planet = state.DeterObject(ObjectName)
	flags = 32*1024 | 2*1024 | ephemeris
	if !(helflag&256 != 0) {
		flags |= flagTruePosition | flagNoNutation
	}
	if Planet != -1 {
		state.sweSetTopo(location[0], location[1], location[2])
		if state.swePhenoUt(JDNDaysUT, Planet, flags, x[:], message) == -1 {
			return -1
		}
		dmag[0] = x[4]
	} else if state.callSweFixstarMag(ObjectName, dmag, message) == -1 {
		return -1
	}
	return 0
}

// MoonsBrightness implements MoonsBrightness from swehel.c.
//
// ###################################################################
// ' dist [km]
// ' phasemoon [-]
// ' MoonsBrightness [-]
func (state *State) MoonsBrightness(dist float64, phasemoon float64) float64 {
	var log10 float64 = 2.302585092994
	return -21.62 + 5*math.Log(dist/6378.1365999999998)/log10 + 0.026*math.Abs(phasemoon) + 0.000000004*math.Pow(phasemoon, 4)
}

// MoonPhase implements MoonPhase from swehel.c.
//
// ###################################################################
// ' AltM [deg]
// ' AziM [deg]
// ' AziS [deg]
// ' MoonPhase [deg]
func (state *State) MoonPhase(AltM float64, AziM float64, AltS float64, AziS float64) float64 {
	var (
		AltMi      float64 = AltM * radiansPerDegree
		AltSi      float64 = AltS * radiansPerDegree
		AziMi      float64 = AziM * radiansPerDegree
		AziSi      float64 = AziS * radiansPerDegree
		MoonAvgPar float64 = 0.95
	)
	return 180 - math.Acos(math.Cos(AziSi-AziMi-MoonAvgPar*radiansPerDegree)*math.Cos(AltMi+MoonAvgPar*radiansPerDegree)*math.Cos(AltSi)+math.Sin(AltSi)*math.Sin(AltMi+MoonAvgPar*radiansPerDegree))/radiansPerDegree
}

// Bm implements Bm from swehel.c.
//
// ###################################################################
// ' Pressure [mbar]
func (state *State) Bm(AltO float64, AziO float64, AltM float64, AziM float64, AltS float64, AziS float64, sunra float64, Lat float64, HeightEye float64, datm []float64, helflag int32, message []byte) float64 {
	var (
		M0           float64 = -11.05
		Bm           float64 = 0
		RM           float64
		kXM          float64
		kX           float64
		C3           float64
		FM           float64
		phasemoon    float64
		MM           float64
		lunarRadius  float64 = 0.0043633231299858239
		objectIsMoon int32   = 0
	)
	if AltO == AltM && AziO == AziM {
		objectIsMoon = 1
	}
	if AltM > -0.26 && !(objectIsMoon != 0) {
		RM = state.DistanceAngle(AltO*radiansPerDegree, AziO*radiansPerDegree, AltM*radiansPerDegree, AziM*radiansPerDegree) / radiansPerDegree
		if RM <= lunarRadius {
			RM = lunarRadius
		}
		kXM = state.Deltam(AltM, AltS, sunra, Lat, HeightEye, datm, helflag, message)
		kX = state.Deltam(AltO, AltS, sunra, Lat, HeightEye, datm, helflag, message)
		C3 = math.Pow(10, -0.4*kXM)
		FM = 62000000.0/RM/RM + math.Pow(10, 6.15-RM/40) + math.Pow(10, 5.36)*(1.06+math.Pow(math.Cos(RM*radiansPerDegree), 2))
		Bm = FM*C3 + 440000*(1-C3)
		phasemoon = state.MoonPhase(AltM, AziM, AltS, AziS)
		MM = state.MoonsBrightness(384410.4978, phasemoon)
		Bm = Bm * math.Pow(10, -0.4*(MM-M0+43.27))
		Bm = Bm * (1 - math.Pow(10, -0.4*kX))
	}
	Bm = state.mymax(Bm, 0) * 980392156862745.12
	return Bm
}

// Btwi implements Btwi from swehel.c.
//
// ###################################################################
// ' Pressure [mbar]
func (state *State) Btwi(AltO float64, AziO float64, AltS float64, AziS float64, sunra float64, Lat float64, HeightEye float64, datm []float64, helflag int32, message []byte) float64 {
	var (
		M0      float64 = -11.05
		MS      float64 = -26.74
		PresE   float64 = state.PresEfromPresS(datm[1], datm[0], HeightEye)
		TempE   float64 = state.TempEfromTempS(datm[1], HeightEye, 0.0065)
		AppAltO float64 = state.AppAltfromTopoAlt(AltO, TempE, PresE, helflag)
		ZendO   float64 = 90 - AppAltO
		RS      float64 = state.DistanceAngle(AltO*radiansPerDegree, AziO*radiansPerDegree, AltS*radiansPerDegree, AziS*radiansPerDegree) / radiansPerDegree
		kX      float64 = state.Deltam(AltO, AltS, sunra, Lat, HeightEye, datm, helflag, message)
		k       float64 = state.kt(AltS, sunra, Lat, HeightEye, datm[1], datm[2], datm[3], 4, message)
		Btwi    float64 = math.Pow(10, -0.4*(MS-M0+32.5-AltS-ZendO/(360*k)))
	)
	Btwi = Btwi * (100 / RS) * (1 - math.Pow(10, -0.4*kX))
	Btwi = state.mymax(Btwi, 0) * 980392156862745.12
	return Btwi
}

// Bday implements Bday from swehel.c.
//
// ###################################################################
// ' Pressure [mbar]
// 2300 REM  Daylight brightness
// 2310 C4=10.0^(-.4*K(I)*XS)
// 2320 FS=6.2E+07*(RS^-2)+(10^(6.15-RS/40))
// 2330 FS=FS+(10^5.36)*(1.06+((COS(RS*RD))^2))
// 2340 BD=10^(-.4*(MS(I)-MO(I)+43.27))
// 2350 BD=BD*(1-10^(-.4*K(I)*X))
// 2360 BD=BD*(FS*C4+440000.0*(1-C4))
func (state *State) Bday(AltO float64, AziO float64, AltS float64, AziS float64, sunra float64, Lat float64, HeightEye float64, datm []float64, helflag int32, message []byte) float64 {
	var (
		M0   float64 = -11.05
		MS   float64 = -26.74
		RS   float64 = state.DistanceAngle(AltO*radiansPerDegree, AziO*radiansPerDegree, AltS*radiansPerDegree, AziS*radiansPerDegree) / radiansPerDegree
		kXS  float64 = state.Deltam(AltS, AltS, sunra, Lat, HeightEye, datm, helflag, message)
		kX   float64 = state.Deltam(AltO, AltS, sunra, Lat, HeightEye, datm, helflag, message)
		C4   float64 = math.Pow(10, -0.4*kXS)
		FS   float64 = 62000000.0/RS/RS + math.Pow(10, 6.15-RS/40) + math.Pow(10, 5.36)*(1.06+math.Pow(math.Cos(RS*radiansPerDegree), 2))
		Bday float64 = FS*C4 + 440000.0*(1-C4)
	)
	Bday = Bday * math.Pow(10, -0.4*(MS-M0+43.27))
	Bday = Bday * (1 - math.Pow(10, -0.4*kX))
	Bday = state.mymax(Bday, 0) * 980392156862745.12
	return Bday
}

// Bcity implements Bcity from swehel.c.
//
// ###################################################################
// ' Value [nL]
// ' PresS [mbar]
// ' Bcity [nL]
func (state *State) Bcity(Value float64, Press float64) float64 {
	var Bcity float64 = Value
	Press += 0.0
	Bcity = state.mymax(Bcity, 0)
	return Bcity
}

// Bsky implements Bsky from swehel.c.
//
// ###################################################################
// ' Pressure [mbar]
func (state *State) Bsky(AltO float64, AziO float64, AltM float64, AziM float64, JDNDaysUT float64, AltS float64, AziS float64, sunra float64, Lat float64, HeightEye float64, datm []float64, helflag int32, message []byte) float64 {
	var Bsky float64 = 0
	if AltS < float64(-3) {
		Bsky += state.Btwi(AltO, AziO, AltS, AziS, sunra, Lat, HeightEye, datm, helflag, message)
	} else if AltS > 4 {
		Bsky += state.Bday(AltO, AziO, AltS, AziS, sunra, Lat, HeightEye, datm, helflag, message)
	} else {
		Bsky += state.mymin(state.Bday(AltO, AziO, AltS, AziS, sunra, Lat, HeightEye, datm, helflag, message), state.Btwi(AltO, AziO, AltS, AziS, sunra, Lat, HeightEye, datm, helflag, message))
		if 0 != 0 {
			if state.heliacal.bskyA == 0 {
				printText(textBytes("bsk=%f\n"), Bsky)
			}
			state.heliacal.bskyA = 1
		}
	}
	if Bsky < 200000000.0 {
		Bsky += state.Bm(AltO, AziO, AltM, AziM, AltS, AziS, sunra, Lat, HeightEye, datm, helflag, message)
	}
	if AltS <= 0 {
		Bsky += state.Bcity(0, datm[0])
	}
	if Bsky < 5000 {
		Bsky = Bsky + state.Bn(AltO, JDNDaysUT, AltS, sunra, Lat, HeightEye, datm, helflag, message)
	}
	return Bsky
}

// defaultHeliacalParameters implements default_heliacal_parameters from swehel.c.
//
// OpticMagn=1 means using eye
func (state *State) defaultHeliacalParameters(datm []float64, location []float64, dobs []float64, helflag int32) {
	var i int32
	if datm[0] <= 0 {
		datm[0] = 1013.25 * math.Pow(1-0.0065*location[2]/288, 5.255)
		if datm[1] == 0 {
			datm[1] = 15 - 0.0065*location[2]
		}
		if datm[2] == 0 {
			datm[2] = 40
		}
	} else {
	}
	if dobs[0] == 0 {
		dobs[0] = 36
	}
	if dobs[1] == 0 {
		dobs[1] = 1
	}
	if !(helflag&512 != 0) {
		for i = 2; i <= 5; i++ {
			dobs[i] = 0
		}
	}
	if dobs[3] == 0 {
		dobs[2] = 1
		dobs[3] = 1
	}
}

// VisLimMagn implements VisLimMagn from swehel.c.
//
// ###################################################################
// ' age [Year]
// ' SN [-]
// ' AltO [deg]
// ' AziO [deg]
// ' AltM [deg]
// ' AziM [deg]
// ' MoonDistance [km]
// ' JDNDaysUT [-]
// ' AltS [deg]
// ' AziS [deg]
// ' lat [deg]
// ' heighteye [m]
// ' TempS [C]
// ' PresS [mbar]
// ' RH [%]
// ' VR [km]
// ' VisLimMagn [-]
func (state *State) VisLimMagn(dobs []float64, AltO float64, AziO float64, AltM float64, AziM float64, JDNDaysUT float64, AltS float64, AziS float64, sunra float64, Lat float64, HeightEye float64, datm []float64, helflag int32, scotopicFlag []int32, message []byte) float64 {
	var (
		C1          float64
		C2          float64
		Th          float64
		kX          float64
		Bsk         float64
		CorrFactor1 float64
		CorrFactor2 float64
		log10       float64 = 2.302585092994
		isScotopic  int32   = 0
	)
	Bsk = state.Bsky(AltO, AziO, AltM, AziM, JDNDaysUT, AltS, AziS, sunra, Lat, HeightEye, datm, helflag, message)
	kX = state.Deltam(AltO, AltS, sunra, Lat, HeightEye, datm, helflag, message)
	if 0 != 0 {
		if state.heliacal.visLimMagnA == 0 {
			printText(textBytes("bsk=%f, kx=%f\n"), Bsk, kX)
		}
		state.heliacal.visLimMagnA = 1
	}
	CorrFactor1 = state.OpticFactor(Bsk, kX, dobs, JDNDaysUT, textBytes(""), 1, helflag)
	CorrFactor2 = state.OpticFactor(Bsk, kX, dobs, JDNDaysUT, textBytes(""), 0, helflag)
	if Bsk < 1645 {
		isScotopic = 1
	}
	if helflag&(1<<14) != 0 {
		isScotopic = 0
	}
	if helflag&(1<<15) != 0 {
		isScotopic = 1
	}
	if isScotopic != 0 {
		C1 = 1.5848931924611e-10
		C2 = 0.012589254117942
		if scotopicFlag != nil {
			scotopicFlag[0] = 1
		}
	} else {
		C1 = 4.4668359215096e-9
		C2 = 1.2589254117942e-6
		if scotopicFlag != nil {
			scotopicFlag[0] = 0
		}
	}
	if scotopicFlag != nil {
		if 1479 > Bsk && 1479 < Bsk {
			scotopicFlag[0] |= 2
		}
	}
	Bsk = Bsk * CorrFactor1
	Th = C1 * math.Pow(1+math.Sqrt(C2*Bsk), 2) * CorrFactor2
	return -16.57 - 2.5*(math.Log(Th)/log10)
}

// tolowerStringStar implements tolower_string_star from swehel.c.
func (state *State) tolowerStringStar(str []byte) []byte {
	var sp []byte
	for sp = str; int32(sp[0]) != '\000' && int32(sp[0]) != ','; sp = sp[1:] {
		sp[0] = byte(lowerByte(sp[0]))
	}
	return str
}

// sweVisLimitMag implements swe_vis_limit_mag from swehel.c.
//
//	Limiting magnitude in dark skies
//
// for information about input parameters, see function swe_heliacal_ut().
//
// function returns:
// -1   Error
// -2   Object is below horizon
//
//	0   OK, photopic vision
//	|1  OK, scotopic vision
//	|2  OK, near limit photopic/scotopic
func (state *State) sweVisLimitMag(tjdut float64, location []float64, datm []float64, dobs []float64, ObjectName []byte, helflag int32, dret []float64, message []byte) int32 {
	var (
		retval       int32 = 0
		i            int32
		scotopicFlag [1]int32 = [1]int32{0}
		AltO         [1]float64
		AziO         [1]float64
		AltM         [1]float64
		AziM         [1]float64
		AltS         [1]float64
		AziS         [1]float64
		sunra        float64
	)
	for i = 0; i < 7; i++ {
		dret[i] = 0
	}
	state.tolowerStringStar(ObjectName)
	if state.DeterObject(ObjectName) == 0 {
		if message != nil {
			copyText(message, textBytes("it makes no sense to call swe_vis_limit_mag() for the Sun"))
		}
		return -1
	}
	state.swiSetTidAcc(tjdut, helflag, 0, message)
	sunra = state.SunRA(tjdut, helflag, message)
	state.defaultHeliacalParameters(datm, location, dobs, helflag)
	state.sweSetTopo(location[0], location[1], location[2])
	if state.ObjectLoc(tjdut, location, datm, ObjectName, 0, helflag, AltO[:], message) == -1 {
		return -1
	}
	if AltO[0] < 0 {
		if message != nil {
			copyText(message, textBytes("object is below local horizon"))
		}
		dret[0] = float64(-100)
		return -2
	}
	if state.ObjectLoc(tjdut, location, datm, ObjectName, 1, helflag, AziO[:], message) == -1 {
		return -1
	}
	if helflag&(1<<12) != 0 {
		AltS[0] = float64(-90)
		AziS[0] = 0
	} else {
		if state.ObjectLoc(tjdut, location, datm, textBytes("sun"), 0, helflag, AltS[:], message) == -1 {
			return -1
		}
		if state.ObjectLoc(tjdut, location, datm, textBytes("sun"), 1, helflag, AziS[:], message) == -1 {
			return -1
		}
	}
	if compareTextN(ObjectName, textBytes("moon"), 4) == 0 || helflag&(1<<12) != 0 || helflag&(1<<13) != 0 {
		AltM[0] = float64(-90)
		AziM[0] = 0
	} else {
		if state.ObjectLoc(tjdut, location, datm, textBytes("moon"), 0, helflag, AltM[:], message) == -1 {
			return -1
		}
		if state.ObjectLoc(tjdut, location, datm, textBytes("moon"), 1, helflag, AziM[:], message) == -1 {
			return -1
		}
	}
	dret[0] = state.VisLimMagn(dobs, AltO[0], AziO[0], AltM[0], AziM[0], tjdut, AltS[0], AziS[0], sunra, location[1], location[2], datm, helflag, scotopicFlag[:], message)
	dret[1] = AltO[0]
	dret[2] = AziO[0]
	dret[3] = AltS[0]
	dret[4] = AziS[0]
	dret[5] = AltM[0]
	dret[6] = AziM[0]
	if state.Magnitude(tjdut, location, ObjectName, helflag, dret[7:], message) == -1 {
		return -1
	}
	retval = scotopicFlag[0]
	return retval
}

// TopoArcVisionis implements TopoArcVisionis from swehel.c.
//
// ###################################################################
// ' Magn [-]
// ' age [Year]
// ' SN [-]
// ' AltO [deg]
// ' AziO [deg]
// ' AltM [deg]
// ' AziM [deg]
// ' MoonDistance [km]
// ' JDNDaysUT [-]
// ' AziS [deg]
// ' lat [deg]
// ' heighteye [m]
// ' Temperature [C]
// ' Pressure [mbar]
// ' RH [%]
// ' VR [km]
// ' TopoArcVisionis [deg]
func (state *State) TopoArcVisionis(Magn float64, dobs []float64, AltO float64, AziO float64, AltM float64, AziM float64, JDNDaysUT float64, AziS float64, sunra float64, Lat float64, HeightEye float64, datm []float64, helflag int32, dret []float64, message []byte) int32 {
	var (
		Xm    float64
		Ym    float64
		AltSi float64
		AziSi float64
		xR    float64 = 0
		Xl    float64 = 45
		Yl    float64
		Yr    float64
	)
	Yl = Magn - state.VisLimMagn(dobs, AltO, AziO, AltM, AziM, JDNDaysUT, AltO-Xl, AziS, sunra, Lat, HeightEye, datm, helflag, nil, message)
	Yr = Magn - state.VisLimMagn(dobs, AltO, AziO, AltM, AziM, JDNDaysUT, AltO-xR, AziS, sunra, Lat, HeightEye, datm, helflag, nil, message)
	if Yl*Yr <= 0 {
		for math.Abs(xR-Xl) > 0.001 {
			Xm = (xR + Xl) / 2.0
			AltSi = AltO - Xm
			AziSi = AziS
			Ym = Magn - state.VisLimMagn(dobs, AltO, AziO, AltM, AziM, JDNDaysUT, AltSi, AziSi, sunra, Lat, HeightEye, datm, helflag, nil, message)
			if Yl*Ym > 0 {
				Xl = Xm
				Yl = Ym
			} else {
				xR = Xm
				Yr = Ym
			}
		}
		Xm = (xR + Xl) / 2.0
	} else {
		Xm = 99
	}
	if Xm < AltO {
		Xm = AltO
	}
	dret[0] = Xm
	return 0
}

// sweTopoArcusVisionis implements swe_topo_arcus_visionis from swehel.c.
func (state *State) sweTopoArcusVisionis(tjdut float64, location []float64, datm []float64, dobs []float64, helflag int32, mag float64, aziObj float64, altObj float64, aziSun float64, aziMoon float64, altMoon float64, dret []float64, message []byte) int32 {
	var sunra float64
	state.swiSetTidAcc(tjdut, helflag, 0, message)
	sunra = state.SunRA(tjdut, helflag, message)
	if message != nil && int32(message[0]) != '\000' {
		return -1
	}
	state.defaultHeliacalParameters(datm, location, dobs, helflag)
	return state.TopoArcVisionis(mag, dobs, altObj, aziObj, altMoon, aziMoon, tjdut, aziSun, sunra, location[1], location[2], datm, helflag, dret, message)
}

// HeliacalAngle implements HeliacalAngle from swehel.c.
func (state *State) internalHeliacalAngle(Magn float64, dobs []float64, AziO float64, AltM float64, AziM float64, JDNDaysUT float64, AziS float64, location []float64, datm []float64, helflag int32, dangret []float64, message []byte) int32 {
	var (
		x         float64
		minx      float64
		maxx      float64
		xmin      float64
		ymin      float64
		Xl        float64
		xR        float64
		Yr        [1]float64
		Yl        [1]float64
		Xm        float64
		Ym        [1]float64
		xmd       float64
		ymd       [1]float64
		Arc       [1]float64
		DELTAx    float64
		sunra     float64 = state.SunRA(JDNDaysUT, helflag, message)
		Lat       float64 = location[1]
		HeightEye float64 = location[2]
	)
	if 0 == 1 {
		dangret[0] = 0.0
		dangret[1] = 0.0 + Magn*2.492 + 13.447
		dangret[2] = -(Magn*2.492 + 13.447)
		return 0
	}
	minx = 2
	maxx = 20
	xmin = 0
	ymin = 10000
	for x = minx; x <= maxx; x++ {
		if state.TopoArcVisionis(Magn, dobs, x, AziO, AltM, AziM, JDNDaysUT, AziS, sunra, Lat, HeightEye, datm, helflag, Arc[:], message) == -1 {
			return -1
		}
		if Arc[0] < ymin {
			ymin = Arc[0]
			xmin = x
		}
	}
	Xl = xmin - 1
	xR = xmin + 1
	if state.TopoArcVisionis(Magn, dobs, xR, AziO, AltM, AziM, JDNDaysUT, AziS, sunra, Lat, HeightEye, datm, helflag, Yr[:], message) == -1 {
		return -1
	}
	if state.TopoArcVisionis(Magn, dobs, Xl, AziO, AltM, AziM, JDNDaysUT, AziS, sunra, Lat, HeightEye, datm, helflag, Yl[:], message) == -1 {
		return -1
	}
	for math.Abs(xR-Xl) > 0.1 {
		Xm = (xR + Xl) / 2.0
		DELTAx = 0.025
		xmd = Xm + DELTAx
		if state.TopoArcVisionis(Magn, dobs, Xm, AziO, AltM, AziM, JDNDaysUT, AziS, sunra, Lat, HeightEye, datm, helflag, Ym[:], message) == -1 {
			return -1
		}
		if state.TopoArcVisionis(Magn, dobs, xmd, AziO, AltM, AziM, JDNDaysUT, AziS, sunra, Lat, HeightEye, datm, helflag, ymd[:], message) == -1 {
			return -1
		}
		if Ym[0] >= ymd[0] {
			Xl = Xm
			Yl[0] = Ym[0]
		} else {
			xR = Xm
			Yr[0] = Ym[0]
		}
	}
	Xm = (xR + Xl) / 2.0
	Ym[0] = (Yr[0] + Yl[0]) / 2.0
	dangret[1] = Ym[0]
	dangret[2] = Xm - Ym[0]
	dangret[0] = Xm
	return 0
}

// sweHeliacalAngle implements swe_heliacal_angle from swehel.c.
func (state *State) sweHeliacalAngle(tjdut float64, location []float64, datm []float64, dobs []float64, helflag int32, mag float64, aziObj float64, aziSun float64, aziMoon float64, altMoon float64, dret []float64, message []byte) int32 {
	if location[2] < -500.0 || location[2] > 25000.0 {
		if message != nil {
			formatText(message, textBytes("location for heliacal events must be between %.0f and %.0f m above sea"), -500.0, 25000.0)
		}
		return -1
	}
	state.swiSetTidAcc(tjdut, helflag, 0, message)
	state.defaultHeliacalParameters(datm, location, dobs, helflag)
	return state.internalHeliacalAngle(mag, dobs, aziObj, altMoon, aziMoon, tjdut, aziSun, location, datm, helflag, dret, message)
}

// WidthMoon implements WidthMoon from swehel.c.
//
// ###################################################################
// ' AltO [deg]
// ' AziO [deg]
// ' AltS [deg]
// ' AziS [deg]
// ' parallax [deg]
// ' WidthMoon [deg]
func (state *State) WidthMoon(AltO float64, AziO float64, AltS float64, AziS float64, parallax float64) float64 {
	var GeoAltO float64 = AltO + parallax
	return 0.27245 * parallax * (1 + math.Sin(GeoAltO*radiansPerDegree)*math.Sin(parallax*radiansPerDegree)) * (1 - math.Cos((AltS-GeoAltO)*radiansPerDegree)*math.Cos((AziS-AziO)*radiansPerDegree))
}

// LengthMoon implements LengthMoon from swehel.c.
//
// ###################################################################
// ' W [deg]
// ' LengthMoon [deg]
func (state *State) LengthMoon(W float64, Diamoon float64) float64 {
	var (
		Wi float64
		D  float64
	)
	if Diamoon == 0 {
		Diamoon = 0.51803333333333335
	}
	Wi = W * 60
	D = Diamoon * 60
	return (D - 0.3*(D+Wi)/2.0/Wi) / 60.0
}

// qYallop implements qYallop from swehel.c.
//
// ###################################################################
// ' W [deg]
// ' GeoARCVact [deg]
// ' q [-]
func (state *State) qYallop(W float64, GeoARCVact float64) float64 {
	var Wi float64 = W * 60
	return (GeoARCVact - (11.8371 - 6.3226*Wi + 0.7319*Wi*Wi - 0.1018*Wi*Wi*Wi)) / 10
}

// crossing implements crossing from swehel.c.
//
// ###################################################################
// 'A (0,p)
// 'B (1,q)
// 'C (0,r)
// 'D (1,s)
func (state *State) crossing(A float64, B float64, C float64, D float64) float64 {
	return (C - A) / (B - A - (D - C))
}

// DeterTAV implements DeterTAV from swehel.c.
//
// ###################################################################
func (state *State) DeterTAV(dobs []float64, JDNDaysUT float64, location []float64, datm []float64, ObjectName []byte, helflag int32, dret []float64, message []byte) int32 {
	var (
		Magn  [1]float64
		AltO  [1]float64
		AziS  [1]float64
		AziO  [1]float64
		AziM  [1]float64
		AltM  [1]float64
		sunra float64 = state.SunRA(JDNDaysUT, helflag, message)
	)
	if state.Magnitude(JDNDaysUT, location, ObjectName, helflag, Magn[:], message) == -1 {
		return -1
	}
	if state.ObjectLoc(JDNDaysUT, location, datm, ObjectName, 0, helflag, AltO[:], message) == -1 {
		return -1
	}
	if state.ObjectLoc(JDNDaysUT, location, datm, ObjectName, 1, helflag, AziO[:], message) == -1 {
		return -1
	}
	if compareTextN(ObjectName, textBytes("moon"), 4) == 0 {
		AltM[0] = float64(-90)
		AziM[0] = 0
	} else {
		if state.ObjectLoc(JDNDaysUT, location, datm, textBytes("moon"), 0, helflag, AltM[:], message) == -1 {
			return -1
		}
		if state.ObjectLoc(JDNDaysUT, location, datm, textBytes("moon"), 1, helflag, AziM[:], message) == -1 {
			return -1
		}
	}
	if state.ObjectLoc(JDNDaysUT, location, datm, textBytes("sun"), 1, helflag, AziS[:], message) == -1 {
		return -1
	}
	if state.TopoArcVisionis(Magn[0], dobs, AltO[0], AziO[0], AltM[0], AziM[0], JDNDaysUT, AziS[0], sunra, location[1], location[2], datm, helflag, dret, message) == -1 {
		return -1
	}
	return 0
}

// x2min implements x2min from swehel.c.
//
// ###################################################################
// ' A y-value at x=1
// ' B y-value at x=0
// ' C y-value at x=-1
// ' x2min minimum for the quadratic function
func (state *State) x2min(A float64, B float64, C float64) float64 {
	var term float64 = A + C - 2*B
	if term == 0 {
		return 0
	}
	return -(A - C) / 2.0 / term
}

// funct2 implements funct2 from swehel.c.
//
// ###################################################################
// ' A y-value at x=1
// ' B y-value at x=0
// ' C y-value at x=-1
// ' x
// ' y is y-value of quadratic function
func (state *State) funct2(A float64, B float64, C float64, x float64) float64 {
	return (A+C-2*B)/2.0*x*x + (A-C)/2.0*x + B
}

// strcpyVBsafe implements strcpy_VBsafe from swehel.c.
//
// ###################################################################
// ' A y-value at x=1
// ' B y-value at x=0
// ' C y-value at x=-1
// ' x
// ' y is y-value of quadratic function
func (state *State) strcpyVBsafe(sout []byte, sin []byte) {
	var (
		sp  []byte
		sp2 []byte
		iw  int32 = 0
	)
	sp = sin
	for sp2 = sout; (isAlphaNumeric(int32(sp[0])) != 0 || int32(sp[0]) == ' ' || int32(sp[0]) == '-' || int32(sp[0]) == ',') && iw < 30; {
		sp2[0] = sp[0]
		sp = sp[1:]
		sp2 = sp2[1:]
		iw++
	}
	sp2[0] = byte('\000')
}

// sweHeliacalPhenoUt implements swe_heliacal_pheno_ut from swehel.c.
//
// ###################################################################
// ' JDNDaysUT [JDN]
// ' for information about input parameters, see function swe_heliacal_ut().
// '
// ' output values:
// '0=AltO [deg]		topocentric altitude of object (unrefracted)
// '1=AppAltO [deg]        apparent altitude of object (refracted)
// '2=GeoAltO [deg]        geocentric altitude of object
// '3=AziO [deg]           azimuth of object
// '4=AltS [deg]           topocentric altitude of Sun
// '5=AziS [deg]           azimuth of Sun
// '6=TAVact [deg]         actual topocentric arcus visionis
// '7=ARCVact [deg]        actual (geocentric) arcus visionis
// '8=DAZact [deg]         actual difference between object's and sun's azimuth
// '9=ARCLact [deg]        actual longitude difference between object and sun
// '10=kact [-]            extinction coefficient
// '11=minTAV [deg]        smallest topocentric arcus visionis
// '12=TfistVR [JDN]       first time object is visible, according to VR
// '13=TbVR [JDN]          optimum time the object is visible, according to VR
// '14=TlastVR [JDN]       last time object is visible, according to VR
// '15=TbYallop[JDN]       best time the object is visible, according to Yallop
// '16=WMoon [deg]         cresent width of moon
// '17=qYal [-]            q-test value of Yallop
// '18=qCrit [-]           q-test criterion of Yallop
// '19=ParO [deg]          parallax of object
// '20 Magn [-]            magnitude of object
// '21=RiseO [JDN]         rise/set time of object
// '22=RiseS [JDN]         rise/set time of sun
// '23=Lag [JDN]           rise/set time of object minus rise/set time of sun
// '24=TvisVR [JDN]        visibility duration
// '25=LMoon [deg]         cresent length of moon
// '26=CVAact [deg]
// '27=Illum [%] 'new
// '28=CVAact [deg] 'new
// '29=MSk [-]
func (state *State) sweHeliacalPhenoUt(JDNDaysUT float64, location []float64, datm []float64, dobs []float64, ObjectNameIn []byte, TypeEvent int32, helflag int32, darr []float64, message []byte) int32 {
	var (
		AziS          [1]float64
		AltS          [1]float64
		AltS2         [1]float64
		AziO          [1]float64
		AltO          [1]float64
		AltO2         [1]float64
		GeoAltO       [1]float64
		AppAltO       float64
		DAZact        float64
		TAVact        float64
		ParO          float64
		MagnO         [1]float64
		ARCVact       float64
		ARCLact       float64
		kact          float64
		WMoon         float64
		LMoon         float64 = 0
		qYal          float64
		qCrit         float64
		RiseSetO      [1]float64
		RiseSetS      [1]float64
		Lag           float64
		TbYallop      float64
		TfirstVR      float64
		TlastVR       float64
		TbVR          float64
		MinTAV        float64 = 0
		MinTAVact     [1]float64
		Ta            float64
		Tc            float64
		TimeStep      float64
		TimePointer   float64
		MinTAVoud     float64 = 0
		DeltaAltoud   float64 = 0
		DeltaAlt      float64
		TvisVR        float64
		crosspoint    float64
		OldestMinTAV  float64
		extrax        float64
		illum         float64
		elong         float64
		attr          [30]float64
		TimeCheck     float64
		LocalminCheck [1]float64
		retval        int32 = 0
		RS            int32
		Planet        int32
		noriseO       int32 = 0
		ObjectName    [256]byte
		sunra         float64
		flags         int32 = helflag & (1 | 2 | 4)
	)
outputHeliacalPheno:
	for {
		if location[2] < -500.0 || location[2] > 25000.0 {
			if message != nil {
				formatText(message, textBytes("location for heliacal events must be between %.0f and %.0f m above sea"), -500.0, 25000.0)
			}
			return -1
		}
		state.swiSetTidAcc(JDNDaysUT, helflag, 0, message)
		sunra = state.SunRA(JDNDaysUT, helflag, message)
		state.strcpyVBsafe(ObjectName[:], ObjectNameIn)
		state.tolowerStringStar(ObjectName[:])
		state.defaultHeliacalParameters(datm, location, dobs, helflag)
		state.sweSetTopo(location[0], location[1], location[2])
		retval = state.ObjectLoc(JDNDaysUT, location, datm, textBytes("sun"), 1, helflag, AziS[:], message)
		if retval == 0 {
			retval = state.ObjectLoc(JDNDaysUT, location, datm, textBytes("sun"), 0, helflag, AltS[:], message)
		}
		if retval == 0 {
			retval = state.ObjectLoc(JDNDaysUT, location, datm, ObjectName[:], 1, helflag, AziO[:], message)
		}
		if retval == 0 {
			retval = state.ObjectLoc(JDNDaysUT, location, datm, ObjectName[:], 0, helflag, AltO[:], message)
		}
		if retval == 0 {
			retval = state.ObjectLoc(JDNDaysUT, location, datm, ObjectName[:], 7, helflag, GeoAltO[:], message)
		}
		if retval == -1 {
			return -1
		}
		AppAltO = state.AppAltfromTopoAlt(AltO[0], datm[1], datm[0], helflag)
		DAZact = AziS[0] - AziO[0]
		TAVact = AltO[0] - AltS[0]
		ParO = GeoAltO[0] - AltO[0]
		if state.Magnitude(JDNDaysUT, location, ObjectName[:], helflag, MagnO[:], message) == -1 {
			return -1
		}
		ARCVact = TAVact + ParO
		ARCLact = math.Acos(math.Cos(ARCVact*radiansPerDegree)*math.Cos(DAZact*radiansPerDegree)) / radiansPerDegree
		Planet = state.DeterObject(ObjectName[:])
		if Planet == -1 {
			elong = ARCLact
			illum = 100
		} else {
			retval = state.swePhenoUt(JDNDaysUT, Planet, flags|(flagEquatorial|flagTopocentric), attr[:], message)
			if retval == -1 {
				return -1
			}
			elong = attr[2]
			illum = attr[1] * 100
		}
		kact = state.kt(AltS[0], sunra, location[1], location[2], datm[1], datm[2], datm[3], 4, message)
		if 0 != 0 {
			darr[26] = state.kR(AltS[0], location[2])
			darr[27] = state.kW(location[2], datm[1], datm[2])
			darr[28] = state.kOZ(AltS[0], sunra, location[1])
			darr[29] = state.ka(AltS[0], sunra, location[1], location[2], datm[1], datm[2], datm[3], message)
			darr[30] = darr[26] + darr[27] + darr[28] + darr[29]
		}
		WMoon = 0
		qYal = 0
		qCrit = 0
		LMoon = 0
		if Planet == 1 {
			WMoon = state.WidthMoon(AltO[0], AziO[0], AltS[0], AziS[0], ParO)
			LMoon = state.LengthMoon(WMoon, 0)
			qYal = state.qYallop(WMoon, ARCVact)
			if qYal > 0.216 {
				qCrit = 1
			}
			if qYal < 0.216 && qYal > -0.014 {
				qCrit = 2
			}
			if qYal < -0.014 && qYal > -0.16 {
				qCrit = 3
			}
			if qYal < -0.16 && qYal > -0.232 {
				qCrit = 4
			}
			if qYal < -0.232 && qYal > -0.293 {
				qCrit = 5
			}
			if qYal < -0.293 {
				qCrit = 6
			}
		}
		RS = 2
		if TypeEvent == 1 || TypeEvent == 4 {
			RS = 1
		}
		retval = state.RiseSet(JDNDaysUT-0.16666666666666666, location, datm, textBytes("sun"), RS, helflag, 0, RiseSetS[:], message)
		if retval == -1 {
			return -1
		}
		retval = state.RiseSet(JDNDaysUT-0.16666666666666666, location, datm, ObjectName[:], RS, helflag, 0, RiseSetO[:], message)
		if retval == -1 {
			return -1
		}
		TbYallop = 99999999.0
		if retval == -2 {
			Lag = 0
			noriseO = 1
		} else {
			Lag = RiseSetO[0] - RiseSetS[0]
			if Planet == 1 {
				TbYallop = (RiseSetO[0]*4 + RiseSetS[0]*5) / 9.0
			}
		}
		if (TypeEvent == 3 || TypeEvent == 4) && (Planet == -1 || Planet >= 4) {
			TfirstVR = 99999999.0
			TbVR = 99999999.0
			TlastVR = 99999999.0
			TvisVR = 0
			MinTAV = 0
			break outputHeliacalPheno
		}
		MinTAVact[0] = 199
		DeltaAlt = 0
		OldestMinTAV = 0
		Ta = 0
		Tc = 0
		TbVR = 0
		TimeStep = -0.00069444444444444436
		if RS == 2 {
			TimeStep = -TimeStep
		}
		TimePointer = RiseSetS[0] - TimeStep
		for first := true; first || math.Abs(TimePointer-RiseSetS[0]) <= 0.16666666666666666 && Ta == 0 && !(TbVR != 0 && (TypeEvent == 3 || TypeEvent == 4) && (compareTextN(ObjectName[:], textBytes("moon"), 4) != 0 && compareTextN(ObjectName[:], textBytes("venus"), 5) != 0 && compareTextN(ObjectName[:], textBytes("mercury"), 7) != 0)); first = false {
			TimePointer = TimePointer + TimeStep
			OldestMinTAV = MinTAVoud
			MinTAVoud = MinTAVact[0]
			DeltaAltoud = DeltaAlt
			retval = state.ObjectLoc(TimePointer, location, datm, textBytes("sun"), 0, helflag, AltS2[:], message)
			if retval == 0 {
				retval = state.ObjectLoc(TimePointer, location, datm, ObjectName[:], 0, helflag, AltO2[:], message)
			}
			if retval != 0 {
				return -1
			}
			DeltaAlt = AltO2[0] - AltS2[0]
			if state.DeterTAV(dobs, TimePointer, location, datm, ObjectName[:], helflag, MinTAVact[:], message) == -1 {
				return -1
			}
			if MinTAVoud < MinTAVact[0] && TbVR == 0 {
				TimeCheck = TimePointer + float64(state.Sgn(TimeStep)*8)/24.0/60.0
				if RiseSetO[0] != 0 {
					if TimeStep > 0 {
						TimeCheck = state.mymin(TimeCheck, RiseSetO[0])
					} else {
						TimeCheck = state.mymax(TimeCheck, RiseSetO[0])
					}
				}
				if state.DeterTAV(dobs, TimeCheck, location, datm, ObjectName[:], helflag, LocalminCheck[:], message) == -1 {
					return -1
				}
				if LocalminCheck[0] > MinTAVact[0] {
					extrax = state.x2min(MinTAVact[0], MinTAVoud, OldestMinTAV)
					TbVR = TimePointer - (1-extrax)*TimeStep
					MinTAV = state.funct2(MinTAVact[0], MinTAVoud, OldestMinTAV, extrax)
				}
			}
			if DeltaAlt > MinTAVact[0] && Tc == 0 && TbVR == 0 {
				crosspoint = state.crossing(DeltaAltoud, DeltaAlt, MinTAVoud, MinTAVact[0])
				Tc = TimePointer - TimeStep*(1-crosspoint)
			}
			if DeltaAlt < MinTAVact[0] && Ta == 0 && Tc != 0 {
				crosspoint = state.crossing(DeltaAltoud, DeltaAlt, MinTAVoud, MinTAVact[0])
				Ta = TimePointer - TimeStep*(1-crosspoint)
			}
		}
		if RS == 2 {
			TfirstVR = Tc
			TlastVR = Ta
		} else {
			TfirstVR = Ta
			TlastVR = Tc
		}
		if TfirstVR == 0 && TlastVR == 0 {
			if RS == 1 {
				TfirstVR = TbVR - 0.000001
			} else {
				TlastVR = TbVR + 0.000001
			}
		}
		if !(noriseO != 0) {
			if RS == 1 {
				TfirstVR = state.mymax(TfirstVR, RiseSetO[0])
			} else {
				TlastVR = state.mymin(TlastVR, RiseSetO[0])
			}
		}
		TvisVR = 99999999.0
		if TlastVR != 0 && TfirstVR != 0 {
			TvisVR = TlastVR - TfirstVR
		}
		if TlastVR == 0 {
			TlastVR = 99999999.0
		}
		if TbVR == 0 {
			TbVR = 99999999.0
		}
		if TfirstVR == 0 {
			TfirstVR = 99999999.0
		}
		break
	}
	darr[0] = AltO[0]
	darr[1] = AppAltO
	darr[2] = GeoAltO[0]
	darr[3] = AziO[0]
	darr[4] = AltS[0]
	darr[5] = AziS[0]
	darr[6] = TAVact
	darr[7] = ARCVact
	darr[8] = DAZact
	darr[9] = ARCLact
	darr[10] = kact
	darr[11] = MinTAV
	darr[12] = TfirstVR
	darr[13] = TbVR
	darr[14] = TlastVR
	darr[15] = TbYallop
	darr[16] = WMoon
	darr[17] = qYal
	darr[18] = qCrit
	darr[19] = ParO
	darr[20] = MagnO[0]
	darr[21] = RiseSetO[0]
	darr[22] = RiseSetS[0]
	darr[23] = Lag
	darr[24] = TvisVR
	darr[25] = LMoon
	darr[26] = elong
	darr[27] = illum
	return 0
}

// getSynodicPeriod implements get_synodic_period from swehel.c.
func (state *State) getSynodicPeriod(Planet int32) float64 {
	switch Planet {
	case 1:
		return 29.530588853
	case 2:
		return 115.8775
	case 3:
		return 583.9214
	case 4:
		return 779.9361
	case 5:
		return 398.8840
	case 6:
		return 378.0919
	case 7:
		return 369.6560
	case 8:
		return 367.4867
	case 9:
		return 366.7207
	}
	return 366
}

// moonEventArcVis implements moon_event_arc_vis from swehel.c.
//
// ###################################################################
func (state *State) moonEventArcVis(JDNDaysUTStart float64, location []float64, datm []float64, dobs []float64, TypeEvent int32, helflag int32, dret []float64, message []byte) int32 {
	var (
		x                 [20]float64
		MinTAV            [1]float64
		MinTAVoud         float64
		OldestMinTAV      float64
		phase1            float64
		phase2            float64
		JDNDaysUT         float64
		JDNDaysUTi        float64
		tjdMoonevent      [1]float64
		tjdMooneventStart float64
		DeltaAltoud       float64
		TimeCheck         float64
		LocalminCheck     [1]float64
		AltS              [1]float64
		AltO              [1]float64
		DeltaAlt          float64 = 90
		ObjectName        [30]byte
		flags             int32
		Daystep           int32
		goingup           int32
		Planet            int32
		retval            int32
		avkind            int32 = helflag & (1<<16 | 1<<17 | 1<<18 | 1<<19)
		ephemeris         int32 = helflag & (1 | 2 | 4)
	)
	dret[0] = JDNDaysUTStart
	if avkind == 0 {
		avkind = 1 << 16
	}
	if avkind != 1<<16 {
		if message != nil {
			copyText(message, textBytes("error: in valid AV kind for the moon"))
		}
		return -1
	}
	if TypeEvent == 1 || TypeEvent == 2 {
		if message != nil {
			copyText(message, textBytes("error: the moon has no morning first or evening last"))
		}
		return -1
	}
	copyText(ObjectName[:], textBytes("moon"))
	Planet = 1
	flags = 32*1024 | 2*1024 | ephemeris
	if !(helflag&256 != 0) {
		flags |= flagTruePosition | flagNoNutation
	}
	Daystep = 1
	if TypeEvent == 3 {
		TypeEvent = 2
	} else {
		TypeEvent = 1
		Daystep = -Daystep
	}
	JDNDaysUT = JDNDaysUTStart
	if TypeEvent == 1 {
		JDNDaysUT = JDNDaysUT + 30
	}
	state.swePhenoUt(JDNDaysUT, Planet, flags, x[:], message)
	phase2 = x[0]
	goingup = 0
	for first := true; first || (goingup == 0 || goingup == 1 && phase2 > phase1); first = false {
		JDNDaysUT = JDNDaysUT + float64(Daystep)
		phase1 = phase2
		state.swePhenoUt(JDNDaysUT, Planet, flags, x[:], message)
		phase2 = x[0]
		if phase2 > phase1 {
			goingup = 1
		}
	}
	JDNDaysUT = JDNDaysUT - float64(Daystep)
	JDNDaysUTi = JDNDaysUT
	JDNDaysUT = JDNDaysUT - float64(Daystep)
	MinTAVoud = 199
	for first := true; first || DeltaAltoud < MinTAVoud && math.Abs(JDNDaysUT-JDNDaysUTi) < 15; first = false {
		JDNDaysUT = JDNDaysUT + float64(Daystep)
		if retval = state.RiseSet(JDNDaysUT, location, datm, ObjectName[:], TypeEvent, helflag, 0, tjdMoonevent[:], message); retval != 0 {
			return retval
		}
		tjdMooneventStart = tjdMoonevent[0]
		MinTAV[0] = 199
		OldestMinTAV = MinTAV[0]
		for first := true; first || (MinTAV[0] <= MinTAVoud || LocalminCheck[0] < MinTAV[0]) && math.Abs(tjdMoonevent[0]-tjdMooneventStart) < 0.083333333333333329; first = false {
			OldestMinTAV = MinTAVoud
			MinTAVoud = MinTAV[0]
			DeltaAltoud = DeltaAlt
			tjdMoonevent[0] = tjdMoonevent[0] - 0.00069444444444444447*float64(state.Sgn(float64(Daystep)))
			if state.ObjectLoc(tjdMoonevent[0], location, datm, textBytes("sun"), 0, helflag, AltS[:], message) == -1 {
				return -1
			}
			if state.ObjectLoc(tjdMoonevent[0], location, datm, ObjectName[:], 0, helflag, AltO[:], message) == -1 {
				return -1
			}
			DeltaAlt = AltO[0] - AltS[0]
			if state.DeterTAV(dobs, tjdMoonevent[0], location, datm, ObjectName[:], helflag, MinTAV[:], message) == -1 {
				return -1
			}
			TimeCheck = tjdMoonevent[0] - 0.0055555555555555558*float64(state.Sgn(float64(Daystep)))
			if state.DeterTAV(dobs, TimeCheck, location, datm, ObjectName[:], helflag, LocalminCheck[:], message) == -1 {
				return -1
			}
		}
	}
	if math.Abs(JDNDaysUT-JDNDaysUTi) < 15 {
		tjdMoonevent[0] += (1 - state.x2min(MinTAV[0], MinTAVoud, OldestMinTAV)) * float64(state.Sgn(float64(Daystep))) / 60.0 / 24.0
	} else {
		copyText(message, textBytes("no date found for lunar event"))
		return -1
	}
	dret[0] = tjdMoonevent[0]
	return 0
}

// heliacalUtArcVis implements heliacal_ut_arc_vis from swehel.c.
func (state *State) heliacalUtArcVis(JDNDaysUTStart float64, location []float64, datm []float64, dobs []float64, ObjectName []byte, TypeEventIn int32, helflag int32, dret []float64, serrRet []byte) int32 {
	var (
		x                [6]float64
		xin              [3]float64
		xaz              [3]float64
		dang             [3]float64
		objectmagn       [1]float64 = [1]float64{0}
		maxlength        float64
		DayStep          float64
		JDNDaysUT        float64
		JDNDaysUTfinal   float64
		JDNDaysUTstep    float64
		JDNDaysUTstepoud float64
		JDNarcvisUT      float64
		tjdTt            float64
		tret             [1]float64
		OudeDatum        float64
		JDNDaysUTinp     float64 = JDNDaysUTStart
		JDNDaysUTtijd    float64
		ArcusVis         float64
		ArcusVisDelta    float64
		ArcusVisPto      float64
		ArcusVisDeltaoud float64
		Trise            float64
		sunsangle        float64
		Theliacal        float64
		Tdelta           float64
		Angle            float64
		TimeStep         float64
		TimePointer      float64
		OldestMinTAV     [1]float64
		MinTAVoud        [1]float64
		MinTAVact        [1]float64
		extrax           float64
		TbVR             float64 = 0
		AziS             float64
		AltS             float64
		AziO             float64
		AltO             float64
		DeltaAlt         float64
		direct           float64
		Pressure         float64
		Temperature      float64
		d                float64
		ephemeris        int32
		retval           int32 = 0
		flags            int32
		Planet           int32
		eventtype        int32
		TypeEvent        int32 = TypeEventIn
		doneoneday       int32
		message          [256]byte
	)
	dret[0] = JDNDaysUTStart
	message[0] = byte('\000')
	Planet = state.DeterObject(ObjectName)
	Pressure = datm[0]
	Temperature = datm[1]
	if retval = state.Magnitude(JDNDaysUTStart, location, ObjectName, helflag, objectmagn[:], message[:]); retval == -1 {
		if serrRet != nil && int32(message[0]) != '\000' {
			copyText(serrRet, message[:])
		}
		return retval
	}
	ephemeris = helflag & (1 | 2 | 4)
	flags = 32*1024 | 2*1024 | ephemeris
	if !(helflag&256 != 0) {
		flags |= flagTruePosition | flagNoNutation
	}
	switch Planet {
	case 2:
		DayStep = 1
		maxlength = 100
	case 3:
		DayStep = 64
		maxlength = 384
	case 4:
		DayStep = 128
		maxlength = 640
	case 5:
		DayStep = 64
		maxlength = 384
	case 6:
		DayStep = 64
		maxlength = 256
	default:
		DayStep = 64
		maxlength = 256
	}
	eventtype = TypeEvent
	if eventtype == 2 {
		DayStep = -DayStep
	}
	if eventtype == 4 {
		eventtype = 1
		DayStep = -DayStep
	}
	if eventtype == 3 {
		eventtype = 2
	}
	eventtype |= 256
	JDNDaysUT = JDNDaysUTStart
	JDNDaysUTfinal = JDNDaysUT + maxlength
	JDNDaysUT = JDNDaysUT - 1
	if DayStep < 0 {
		JDNDaysUTtijd = JDNDaysUT
		JDNDaysUT = JDNDaysUTfinal
		JDNDaysUTfinal = JDNDaysUTtijd
	}
	JDNDaysUTstep = JDNDaysUT - DayStep
	doneoneday = 0
	ArcusVisDelta = 199
	ArcusVisPto = -5.55
	for first := true; first || doneoneday == 0 && (JDNDaysUTfinal-JDNDaysUTstep)*float64(state.Sgn(DayStep)) > 0; first = false {
		if math.Abs(DayStep) == 1 {
			doneoneday = 1
		}
		for first := true; first || (ArcusVisDeltaoud > 0 || ArcusVisDelta < 0) && (JDNDaysUTfinal-JDNDaysUTstep)*float64(state.Sgn(DayStep)) > 0; first = false {
			JDNDaysUTstepoud = JDNDaysUTstep
			ArcusVisDeltaoud = ArcusVisDelta
			JDNDaysUTstep = JDNDaysUTstep + DayStep
			if retval = state.myRiseTrans(JDNDaysUTstep, 0, textBytes(""), eventtype, helflag, location, datm, tret[:], message[:]); retval == -1 {
				if serrRet != nil && int32(message[0]) != '\000' {
					copyText(serrRet, message[:])
				}
				return retval
			}
			tjdTt = tret[0] + state.sweDeltatEx(tret[0], ephemeris, message[:])
			if retval = state.sweCalc(tjdTt, 0, flags, x[:], message[:]); retval == -1 {
				if serrRet != nil && int32(message[0]) != '\000' {
					copyText(serrRet, message[:])
				}
				return retval
			}
			xin[0] = x[0]
			xin[1] = x[1]
			state.sweAzalt(tret[0], 1, location, Pressure, Temperature, xin[:], xaz[:])
			Trise = state.HourAngle(xaz[1], x[1], location[1])
			sunsangle = ArcusVisPto
			if helflag&(1<<18) != 0 {
				sunsangle = float64(-7)
			}
			if helflag&(1<<19) != 0 {
				sunsangle = float64(-9)
			}
			Theliacal = state.HourAngle(sunsangle, x[1], location[1])
			Tdelta = Theliacal - Trise
			if TypeEvent == 2 || TypeEvent == 3 {
				Tdelta = -Tdelta
			}
			JDNarcvisUT = tret[0] - Tdelta/24
			tjdTt = JDNarcvisUT + state.sweDeltatEx(JDNarcvisUT, ephemeris, message[:])
			if retval = state.sweCalc(tjdTt, 0, flags, x[:], message[:]); retval == -1 {
				if serrRet != nil && int32(message[0]) != '\000' {
					copyText(serrRet, message[:])
				}
				return retval
			}
			xin[0] = x[0]
			xin[1] = x[1]
			state.sweAzalt(JDNarcvisUT, 1, location, Pressure, Temperature, xin[:], xaz[:])
			AziS = xaz[0] + 180
			if AziS >= 360 {
				AziS = AziS - 360
			}
			AltS = xaz[1]
			if Planet != -1 {
				if retval = state.sweCalc(tjdTt, Planet, flags, x[:], message[:]); retval == -1 {
					if serrRet != nil && int32(message[0]) != '\000' {
						copyText(serrRet, message[:])
					}
					return retval
				}
				if retval = state.Magnitude(JDNarcvisUT, location, ObjectName, helflag, objectmagn[:], message[:]); retval == -1 {
					if serrRet != nil && int32(message[0]) != '\000' {
						copyText(serrRet, message[:])
					}
					return retval
				}
			} else if retval = state.callSweFixstar(ObjectName, tjdTt, flags, x[:], message[:]); retval == -1 {
				if serrRet != nil && int32(message[0]) != '\000' {
					copyText(serrRet, message[:])
				}
				return retval
			}
			xin[0] = x[0]
			xin[1] = x[1]
			state.sweAzalt(JDNarcvisUT, 1, location, Pressure, Temperature, xin[:], xaz[:])
			AziO = xaz[0] + 180
			if AziO >= 360 {
				AziO = AziO - 360
			}
			AltO = xaz[1]
			DeltaAlt = AltO - AltS
			if retval = state.internalHeliacalAngle(objectmagn[0], dobs, AziO, float64(-1), 0, JDNarcvisUT, AziS, location, datm, helflag, dang[:], message[:]); retval == -1 {
				if serrRet != nil && int32(message[0]) != '\000' {
					copyText(serrRet, message[:])
				}
				return retval
			}
			ArcusVis = dang[1]
			ArcusVisPto = dang[2]
			ArcusVisDelta = DeltaAlt - ArcusVis
		}
		if doneoneday == 0 && (JDNDaysUTfinal-JDNDaysUTstep)*float64(state.Sgn(DayStep)) > 0 {
			ArcusVisDelta = ArcusVisDeltaoud
			DayStep = float64(int32(math.Abs(DayStep)/2.0) * state.Sgn(DayStep))
			JDNDaysUTstep = JDNDaysUTstepoud
		}
	}
	d = (JDNDaysUTfinal - JDNDaysUTstep) * float64(state.Sgn(DayStep))
	if d <= 0 || d >= maxlength {
		dret[0] = JDNDaysUTinp
		retval = -2
		formatText(message[:], textBytes("heliacal event not found within maxlength %f\n"), maxlength)
		if serrRet != nil && int32(message[0]) != '\000' {
			copyText(serrRet, message[:])
		}
		return retval
	}
	direct = 0.00069444444444444436
	if DayStep < 0 {
		direct = -direct
	}
	if helflag&(1<<16) != 0 {
		TimeStep = direct
		TbVR = 0
		TimePointer = JDNarcvisUT
		if state.DeterTAV(dobs, TimePointer, location, datm, ObjectName, helflag, OldestMinTAV[:], message[:]) == -1 {
			return -1
		}
		TimePointer = TimePointer + TimeStep
		if state.DeterTAV(dobs, TimePointer, location, datm, ObjectName, helflag, MinTAVoud[:], message[:]) == -1 {
			return -1
		}
		if MinTAVoud[0] > OldestMinTAV[0] {
			TimePointer = JDNarcvisUT
			TimeStep = -TimeStep
			MinTAVact[0] = OldestMinTAV[0]
		} else {
			MinTAVact[0] = MinTAVoud[0]
			MinTAVoud[0] = OldestMinTAV[0]
		}
		for first := true; first || TbVR == 0; first = false {
			TimePointer = TimePointer + TimeStep
			OldestMinTAV[0] = MinTAVoud[0]
			MinTAVoud[0] = MinTAVact[0]
			if state.DeterTAV(dobs, TimePointer, location, datm, ObjectName, helflag, MinTAVact[:], message[:]) == -1 {
				return -1
			}
			if MinTAVoud[0] < MinTAVact[0] {
				extrax = state.x2min(MinTAVact[0], MinTAVoud[0], OldestMinTAV[0])
				TbVR = TimePointer - (1-extrax)*TimeStep
			}
		}
		JDNarcvisUT = TbVR
	}
	if helflag&(1<<17) != 0 {
		for first := true; first || Angle > 0; first = false {
			OudeDatum = JDNarcvisUT
			JDNarcvisUT = JDNarcvisUT - direct
			tjdTt = JDNarcvisUT + state.sweDeltatEx(JDNarcvisUT, ephemeris, message[:])
			if Planet != -1 {
				if retval = state.sweCalc(tjdTt, Planet, flags, x[:], message[:]); retval == -1 {
					if serrRet != nil && int32(message[0]) != '\000' {
						copyText(serrRet, message[:])
					}
					return retval
				}
			} else if retval = state.callSweFixstar(ObjectName, tjdTt, flags, x[:], message[:]); retval == -1 {
				if serrRet != nil && int32(message[0]) != '\000' {
					copyText(serrRet, message[:])
				}
				return retval
			}
			xin[0] = x[0]
			xin[1] = x[1]
			state.sweAzalt(JDNarcvisUT, 1, location, Pressure, Temperature, xin[:], xaz[:])
			Angle = xaz[1]
		}
		JDNarcvisUT = (JDNarcvisUT + OudeDatum) / 2.0
	}
	if JDNarcvisUT < float64(-9999999) || JDNarcvisUT > 9999999 {
		dret[0] = JDNDaysUT
		copyText(message[:], textBytes("no heliacal date found"))
		retval = -1
		if serrRet != nil && int32(message[0]) != '\000' {
			copyText(serrRet, message[:])
		}
		return retval
	}
	dret[0] = JDNarcvisUT
	if serrRet != nil && int32(message[0]) != '\000' {
		copyText(serrRet, message[:])
	}
	return retval
}

// getAscObl implements get_asc_obl from swehel.c.
func (state *State) getAscObl(jd float64, body int32, star []byte, flags int32, location []float64, descObl int32, daop []float64, message []byte) int32 {
	var (
		retval    int32
		ephemeris int32 = flags & (flagJPL | flagSwiss | flagMoshier)
		x         [6]float64
		adp       float64
		s         [256]byte
		star2     [256]byte
	)
	copyText(star2[:], star)
	if body == -1 {
		if retval = state.sweFixstar(star2[:], jd, ephemeris|2*1024, x[:], message); retval == -1 {
			return -1
		}
	} else if retval = state.sweCalc(jd, body, ephemeris|2*1024, x[:], message); retval == -1 {
		return -1
	}
	adp = math.Tan(location[1]*radiansPerDegree) * math.Tan(x[1]*radiansPerDegree)
	if math.Abs(adp) > 1 {
		if star != nil && int32(star[0]) != '\000' {
			copyText(s[:], star)
		} else {
			state.sweGetPlanetName(body, s[:])
		}
		formatText(message, textBytes("%s is circumpolar, cannot calculate heliacal event"), s[:])
		return -2
	}
	adp = math.Asin(adp) / radiansPerDegree
	if descObl != 0 {
		daop[0] = x[0] + adp
	} else {
		daop[0] = x[0] - adp
	}
	daop[0] = state.sweDegnorm(daop[0])
	return 0
}

// getAscOblDiff implements get_asc_obl_diff from swehel.c.
func (state *State) getAscOblDiff(jd float64, body int32, star []byte, flags int32, location []float64, descObl int32, isAcronychal int32, dsunpl []float64, message []byte) int32 {
	var (
		retval int32 = 0
		aosun  [1]float64
		aopl   [1]float64
	)
	retval = state.getAscObl(jd, 0, textBytes(""), flags, location, descObl, aosun[:], message)
	if retval != 0 {
		return retval
	}
	if isAcronychal != 0 {
		if descObl == 1 {
			descObl = 0
		} else {
			descObl = 1
		}
	}
	retval = state.getAscObl(jd, body, star, flags, location, descObl, aopl[:], message)
	if retval != 0 {
		return retval
	}
	dsunpl[0] = state.sweDegnorm(aosun[0] - aopl[0])
	if isAcronychal != 0 {
		dsunpl[0] = state.sweDegnorm(dsunpl[0] - 180)
	}
	if dsunpl[0] > 180 {
		dsunpl[0] -= 360
	}
	return 0
}

// findConjunctSun implements find_conjunct_sun from swehel.c.
//
//	Neptune
func (state *State) findConjunctSun(tjdStart float64, body int32, helflag int32, TypeEvent int32, jd []float64, message []byte) int32 {
	var (
		ephemeris  int32 = helflag & (1 | 2 | 4)
		i          int32
		tjdcon     float64
		initialJD  float64
		ds         float64
		dsynperiod float64
		x          [6]float64
		xs         [6]float64
		daspect    float64 = 0
	)
	if body >= 4 && TypeEvent >= 3 {
		daspect = 180
	}
	i = (TypeEvent-1)/2 + body*2
	initialJD = heliacalTcon[i]
	dsynperiod = state.getSynodicPeriod(body)
	tjdcon = initialJD + (math.Floor((tjdStart-initialJD)/dsynperiod)+1)*dsynperiod
	for ds = 100; ds > 0.5; {
		if state.sweCalc(tjdcon, body, ephemeris|256, x[:], message) == -1 {
			return -1
		}
		if state.sweCalc(tjdcon, 0, ephemeris|256, xs[:], message) == -1 {
			return -1
		}
		ds = state.sweDegnorm(x[0] - xs[0] - daspect)
		if ds > 180 {
			ds -= 360
		}
		tjdcon -= ds / (x[3] - xs[3])
	}
	jd[0] = tjdcon
	return 0
}

// getAscOblWithSun implements get_asc_obl_with_sun from swehel.c.
func (state *State) getAscOblWithSun(tjdStart float64, body int32, star []byte, helflag int32, evtyp int32, dperiod float64, location []float64, tjdret []float64, message []byte) int32 {
	var (
		i            int32
		retval       int32
		isAcronychal int32      = 0
		ephemeris    int32      = helflag & (1 | 2 | 4)
		dsunpl       [1]float64 = [1]float64{1}
		dsunplSave   float64
		dsunplTest   [1]float64
		jd           float64
		daystep      float64
		descObl      int32 = 0
		retro        int32 = 0
	)
	if evtyp == 2 || evtyp == 3 {
		descObl = 1
	}
	if evtyp == 1 || evtyp == 2 {
		retro = 1
	}
	if evtyp == 5 {
		descObl = 1
	}
	if evtyp == 5 || evtyp == 6 {
		isAcronychal = 1
		if body != 1 {
			retro = 1
		}
	}
	jd = tjdStart
	dsunplSave = float64(-999999999)
	retval = state.getAscOblDiff(jd, body, star, ephemeris, location, descObl, isAcronychal, dsunpl[:], message)
	if retval != 0 {
		return retval
	}
	daystep = 20
	for i = 0; dsunplSave == float64(-999999999) || math.Abs(dsunpl[0])+math.Abs(dsunplSave) > 180 || retro != 0 && !(dsunplSave < 0 && dsunpl[0] >= 0) || !(retro != 0) && !(dsunplSave >= 0 && dsunpl[0] < 0); {
		i++
		if i > 5000 {
			formatText(message, textBytes("loop in get_asc_obl_with_sun() (1)"))
			return -1
		}
		dsunplSave = dsunpl[0]
		jd += 10.0
		if dperiod > 0 && jd-tjdStart > dperiod {
			return -2
		}
		retval = state.getAscOblDiff(jd, body, star, ephemeris, location, descObl, isAcronychal, dsunpl[:], message)
		if retval != 0 {
			return retval
		}
	}
	tjdStart = jd - daystep
	daystep /= 2.0
	jd = tjdStart + daystep
	retval = state.getAscOblDiff(jd, body, star, ephemeris, location, descObl, isAcronychal, dsunplTest[:], message)
	if retval != 0 {
		return retval
	}
	for i = 0; math.Abs(dsunpl[0]) > 0.00001; {
		i++
		if i > 5000 {
			formatText(message, textBytes("loop in get_asc_obl_with_sun() (2)"))
			return -1
		}
		if dsunplSave*dsunplTest[0] >= 0 {
			dsunplSave = dsunplTest[0]
			tjdStart = jd
		} else {
			dsunpl[0] = dsunplTest[0]
		}
		daystep /= 2.0
		jd = tjdStart + daystep
		retval = state.getAscOblDiff(jd, body, star, ephemeris, location, descObl, isAcronychal, dsunplTest[:], message)
		if retval != 0 {
			return retval
		}
	}
	tjdret[0] = jd
	return 0
}

// getHeliacalDay implements get_heliacal_day from swehel.c.
func (state *State) getHeliacalDay(jd float64, location []float64, datm []float64, dobs []float64, ObjectName []byte, helflag int32, TypeEvent int32, thel []float64, message []byte) int32 {
	var (
		i                   int32
		visibleAtSunsetrise int32
		isRiseOrSet         int32 = 0
		ndays               int32
		retval              int32
		retvalOld           int32
		directDay           float64 = 0
		directTime          float64 = 0
		tfac                float64
		tend                float64
		daystep             float64
		tday                float64
		vdelta              float64
		tret                [1]float64
		darr                [30]float64
		vd                  float64
		dmag                [1]float64
		div                 float64
		body                int32 = state.DeterObject(ObjectName)
	)
	switch TypeEvent {
	case 1:
		isRiseOrSet = 1
		directDay = 1
		directTime = float64(-1)
	case 2:
		isRiseOrSet = 2
		directDay = float64(-1)
		directTime = 1
	case 3:
		isRiseOrSet = 2
		directDay = 1
		directTime = 1
	case 4:
		isRiseOrSet = 1
		directDay = float64(-1)
		directTime = float64(-1)
	}
	tfac = 1
	switch body {
	case 1:
		ndays = 16
		daystep = 1
	case 2:
		ndays = 60
		jd -= 0 * directDay
		daystep = 5
		tfac = 5
	case 3:
		ndays = 300
		jd -= 30 * directDay
		daystep = 5
		if TypeEvent >= 3 {
			daystep = 15
			tfac = 3
		}
	case 4:
		ndays = 400
		daystep = 15
		tfac = 5
	case 6:
		ndays = 300
		daystep = 20
		tfac = 5
	case -1:
		ndays = 300
		if state.callSweFixstarMag(ObjectName, dmag[:], message) == -1 {
			return -1
		}
		daystep = 15
		tfac = 10
		if dmag[0] > 2 {
			daystep = 15
		}
		if dmag[0] < 0 {
			tfac = 3
		}
	default:
		ndays = 300
		daystep = 15
		tfac = 3
	}
	tend = jd + float64(ndays)*directDay
	retvalOld = -2
	tday = jd
	for i = 0; directDay > 0 && tday < tend || directDay < 0 && tday > tend; {
		vdelta = float64(-100)
		if i > 0 {
			tday -= 0.3 * directDay
		}
		if retval = state.myRiseTrans(tday, 0, textBytes(""), isRiseOrSet, helflag, location, datm, tret[:], message); retval == -1 {
			return -1
		}
		if retval == -2 {
			retvalOld = retval
			tday += daystep * directDay
			i++
			continue
		}
		retval = state.sweVisLimitMag(tret[0], location, datm, dobs, ObjectName, helflag, darr[:], message)
		if retval == -1 {
			return -1
		}
		if retvalOld == -2 && retval >= 0 && daystep > 1 {
			retvalOld = retval
			tday -= daystep * directDay
			daystep = 1
			if body >= 4 || body == -1 {
				daystep = 5
			}
			tday += daystep * directDay
			i++
			continue
		}
		retvalOld = retval
		if retval == -2 {
			tday += daystep * directDay
			i++
			continue
		}
		vdelta = darr[0] - darr[7]
		div = 1440.0
		vd = float64(-1)
		for visibleAtSunsetrise = 1; ; {
			if !(retval != -2) {
				break
			}
			vd = darr[0] - darr[7]
			if !(vd < 0) {
				break
			}
			visibleAtSunsetrise = 0
			if vd < -1.0 {
				tret[0] += 5.0 / div * directTime * tfac
			} else if vd < -0.5 {
				tret[0] += 2.0 / div * directTime * tfac
			} else if vd < -0.1 {
				tret[0] += 1.0 / div * directTime * tfac
			} else {
				tret[0] += 1.0 / div * directTime
			}
			retval = state.sweVisLimitMag(tret[0], location, datm, dobs, ObjectName, helflag, darr[:], message)
			if retval == -1 {
				return -1
			}
		}
		if visibleAtSunsetrise != 0 {
			for i = 0; i < 10; i++ {
				if retval = state.sweVisLimitMag(tret[0]+1.0/div*directTime, location, datm, dobs, ObjectName, helflag, darr[:], message); retval >= 0 && darr[0]-darr[7] > vd {
					vd = darr[0] - darr[7]
					tret[0] += 1.0 / div * directTime
				}
			}
		}
		vdelta = darr[0] - darr[7]
		if vdelta > 0 {
			if (body >= 4 || body == -1) && daystep > 1 {
				tday -= daystep * directDay
				daystep = 1
			} else {
				thel[0] = tret[0]
				return 0
			}
		}
		tday += daystep * directDay
		i++
	}
	formatText(message, textBytes("heliacal event does not happen"))
	return -2
}

// timeOptimumVisibility implements time_optimum_visibility from swehel.c.
func (state *State) timeOptimumVisibility(jd float64, location []float64, datm []float64, dobs []float64, ObjectName []byte, helflag int32, tret []float64, message []byte) int32 {
	var (
		retval         int32
		retvalSv       int32
		i              int32
		t1             float64
		t2             float64
		vl1            float64
		vl2            float64
		d              float64
		darr           [10]float64
		photScotOpic   float64
		photScotOpicSv float64
		tHasChanged    int32
	)
	tret[0] = jd
	retval = state.sweVisLimitMag(jd, location, datm, dobs, ObjectName, helflag, darr[:], message)
	if retval == -1 {
		return -1
	}
	retvalSv = retval
	t1 = jd
	t2 = jd
	vl1 = float64(-1)
	vl2 = float64(-1)
	photScotOpicSv = float64(retval & 1)
	i = 0
	for d = 0.0011574074074074073; i < 3; {
		t1 += d
		for tHasChanged = 0; ; {
			retval = state.sweVisLimitMag(t1-d, location, datm, dobs, ObjectName, helflag, darr[:], message)
			if !(retval >= 0 && darr[0] > darr[7] && darr[0]-darr[7] > vl1) {
				break
			}
			t1 -= d
			vl1 = darr[0] - darr[7]
			tHasChanged = 1
			retvalSv = retval
			photScotOpicSv = float64(retval & 1)
		}
		if tHasChanged == 0 {
			t1 -= d
		}
		if retval == -1 {
			return -1
		}
		i++
		d /= 10.0
	}
	i = 0
	for d = 0.0011574074074074073; i < 3; {
		t2 -= d
		for tHasChanged = 0; ; {
			retval = state.sweVisLimitMag(t2+d, location, datm, dobs, ObjectName, helflag, darr[:], message)
			if !(retval >= 0 && darr[0] > darr[7] && darr[0]-darr[7] > vl2) {
				break
			}
			t2 += d
			vl2 = darr[0] - darr[7]
			tHasChanged = 1
			retvalSv = retval
			photScotOpicSv = float64(retval & 1)
		}
		if tHasChanged == 0 {
			t2 += d
		}
		if retval == -1 {
			return -1
		}
		i++
		d /= 10.0
	}
	if vl2 > vl1 {
		jd = t2
	} else {
		jd = t1
	}
	tret[0] = jd
	if retval >= 0 {
		photScotOpic = float64(retval & 1)
		if photScotOpicSv != photScotOpic {
			printText(textBytes("hallo -2\n"))
			return -2
		}
		if retvalSv&2 != 0 {
			printText(textBytes("hallo -2\n"))
			return -2
		}
	}
	return 0
}

// timeLimitInvisible implements time_limit_invisible from swehel.c.
//
//	valid result found but it is close to the scotopic/photopic limit
func (state *State) timeLimitInvisible(jd float64, location []float64, datm []float64, dobs []float64, ObjectName []byte, helflag int32, direct int32, tret []float64, message []byte) int32 {
	var (
		retval         int32
		retvalSv       int32
		i              int32
		ncnt           int32   = 3
		d              float64 = 0
		darr           [10]float64
		photScotOpic   float64
		photScotOpicSv float64
		d0             float64 = 0.0011574074074074073
	)
	tret[0] = jd
	if compareText(ObjectName, textBytes("moon")) == 0 {
		d0 *= 10
		ncnt = 4
	}
	retval = state.sweVisLimitMag(jd+d*float64(direct), location, datm, dobs, ObjectName, helflag, darr[:], message)
	if retval == -1 {
		return -1
	}
	retvalSv = retval
	photScotOpicSv = float64(retval & 1)
	i = 0
	for d = d0; i < ncnt; {
		for {
			retval = state.sweVisLimitMag(jd+d*float64(direct), location, datm, dobs, ObjectName, helflag, darr[:], message)
			if !(retval >= 0 && darr[0] > darr[7]) {
				break
			}
			jd += d * float64(direct)
			retvalSv = retval
			photScotOpicSv = float64(retval & 1)
		}
		i++
		d /= 10.0
	}
	tret[0] = jd
	message[0] = byte('\000')
	if retval >= 0 {
		photScotOpic = float64(retval & 1)
		if photScotOpicSv != photScotOpic {
			return -2
		}
		if retvalSv&2 != 0 {
			return -2
		}
	}
	return 0
}

// getAcronychalDay implements get_acronychal_day from swehel.c.
//
//	valid result found but it is close to the scotopic/photopic limit
func (state *State) getAcronychalDay(tjdValue float64, location []float64, datm []float64, dobs []float64, ObjectName []byte, helflag int32, TypeEvent int32, thel []float64, message []byte) int32 {
	var (
		jd          = [1]float64{tjdValue}
		tret        [1]float64
		tretDark    [1]float64
		darr        [30]float64
		dtret       float64
		retval      int32
		isRiseOrSet int32
		direct      int32
		body        int32 = state.DeterObject(ObjectName)
	)
	helflag |= 1 << 14
	if TypeEvent == 3 || TypeEvent == 5 {
		isRiseOrSet = 1
		direct = -1
	} else {
		isRiseOrSet = 2
		direct = 1
	}
	for dtret = 999; math.Abs(dtret) > 0.00034722222222222224; {
		jd[0] += 0.7 * float64(direct)
		if direct < 0 {
			jd[0] -= 1
		}
		retval = state.myRiseTrans(jd[0], body, ObjectName, isRiseOrSet, helflag, location, datm, jd[:], message)
		if retval == -1 {
			return -1
		}
		retval = state.sweVisLimitMag(jd[0], location, datm, dobs, ObjectName, helflag, darr[:], message)
		if retval == -1 {
			return -1
		}
		for darr[0] < darr[7] {
			jd[0] += 0.0069444444444444441 * float64(-direct)
			retval = state.sweVisLimitMag(jd[0], location, datm, dobs, ObjectName, helflag, darr[:], message)
			if retval == -1 {
				return -1
			}
		}
		retval = state.timeLimitInvisible(jd[0], location, datm, dobs, ObjectName, helflag|1<<12, direct, tretDark[:], message)
		if retval == -1 {
			return -1
		}
		retval = state.timeLimitInvisible(jd[0], location, datm, dobs, ObjectName, helflag|1<<13, direct, tret[:], message)
		if retval == -1 {
			return -1
		}
		dtret = math.Abs(tret[0] - tretDark[0])
	}
	if state.azaltCart(tret[0], location, datm, textBytes("sun"), helflag, darr[:], message) == -1 {
		return -1
	}
	thel[0] = tret[0]
	if darr[1] < float64(-12) {
		formatText(message, textBytes("acronychal rising/setting not available, %f"), darr[1])
		return 0
	} else {
		formatText(message, textBytes("solar altitude, %f"), darr[1])
	}
	return 0
}

// getHeliacalDetails implements get_heliacal_details from swehel.c.
func (state *State) getHeliacalDetails(tday float64, location []float64, datm []float64, dobs []float64, ObjectName []byte, TypeEvent int32, helflag int32, dret []float64, message []byte) int32 {
	var (
		i                int32
		retval           int32
		direct           int32
		optimumUndefined int32
		limit1Undefined  int32
		limit2Undefined  int32
	)
	optimumUndefined = 0
	retval = state.timeOptimumVisibility(tday, location, datm, dobs, ObjectName, helflag, dret[1:], message)
	if retval == -1 {
		return -1
	}
	if retval == -2 {
		retval = 0
		optimumUndefined = 1
	}
	direct = 1
	if TypeEvent == 1 || TypeEvent == 4 {
		direct = -1
	}
	limit1Undefined = 0
	retval = state.timeLimitInvisible(tday, location, datm, dobs, ObjectName, helflag, direct, dret[:], message)
	if retval == -1 {
		return -1
	}
	if retval == -2 {
		retval = 0
		limit1Undefined = 1
	}
	direct *= -1
	limit2Undefined = 0
	retval = state.timeLimitInvisible(dret[1], location, datm, dobs, ObjectName, helflag, direct, dret[2:], message)
	if retval == -1 {
		return -1
	}
	if retval == -2 {
		retval = 0
		limit2Undefined = 1
	}
	if TypeEvent == 2 || TypeEvent == 3 {
		tday = dret[2]
		dret[2] = dret[0]
		dret[0] = tday
		i = limit1Undefined
		limit1Undefined = limit2Undefined
		limit2Undefined = i
	}
	if optimumUndefined != 0 || limit1Undefined != 0 || limit2Undefined != 0 {
		formatText(message, textBytes("return values ["))
		if limit1Undefined != 0 {
			appendText(message, textBytes("0,"))
		}
		if optimumUndefined != 0 {
			appendText(message, textBytes("1,"))
		}
		if limit2Undefined != 0 {
			appendText(message, textBytes("2,"))
		}
		appendText(message, textBytes("] are uncertain due to change between photopic and scotopic vision"))
	}
	return 0
}

// heliacalUtVisLim implements heliacal_ut_vis_lim from swehel.c.
func (state *State) heliacalUtVisLim(tjdStart float64, location []float64, datm []float64, dobs []float64, ObjectName []byte, TypeEventIn int32, helflag int32, dret []float64, serrRet []byte) int32 {
	var (
		i         int32
		d         float64
		darr      [10]float64
		direct    float64 = 1
		jd        [1]float64
		tday      [1]float64
		ephemeris int32
		retval    int32 = 0
		helflag2  int32
		flags     int32
		body      int32
		TypeEvent int32 = TypeEventIn
		message   [256]byte
	)
	for i = 0; i < 10; i++ {
		dret[i] = 0
	}
	dret[0] = tjdStart
	message[0] = byte('\000')
	body = state.DeterObject(ObjectName)
	ephemeris = helflag & (1 | 2 | 4)
	flags = 32*1024 | 2*1024 | ephemeris
	if !(helflag&256 != 0) {
		flags |= flagTruePosition | flagNoNutation
	}
	if body == 2 {
		jd[0] = tjdStart - 30
	} else {
		jd[0] = tjdStart - 50
	}
	helflag2 = helflag
	if body == 2 || body == 3 || TypeEvent <= 2 {
		if body == -1 {
			retval = state.getAscOblWithSun(jd[0], body, ObjectName, helflag, TypeEvent, 0, location, jd[:], message[:])
			if retval != 0 {
				if serrRet != nil && int32(message[0]) != '\000' {
					copyText(serrRet, message[:])
				}
				return retval
			}
		} else if retval = state.findConjunctSun(jd[0], body, helflag, TypeEvent, jd[:], message[:]); retval == -1 {
			if serrRet != nil && int32(message[0]) != '\000' {
				copyText(serrRet, message[:])
			}
			return retval
		}
		retval = state.getHeliacalDay(jd[0], location, datm, dobs, ObjectName, helflag2, TypeEvent, tday[:], message[:])
		if retval != 0 {
			if serrRet != nil && int32(message[0]) != '\000' {
				copyText(serrRet, message[:])
			}
			return retval
		}
	} else {
		if 1 != 0 || body == -1 {
			retval = state.getAscOblWithSun(jd[0], body, ObjectName, helflag, TypeEvent, 0, location, jd[:], message[:])
			if retval != 0 {
				if serrRet != nil && int32(message[0]) != '\000' {
					copyText(serrRet, message[:])
				}
				return retval
			}
		} else if retval = state.findConjunctSun(jd[0], body, helflag, TypeEvent, jd[:], message[:]); retval == -1 {
			if serrRet != nil && int32(message[0]) != '\000' {
				copyText(serrRet, message[:])
			}
			return retval
		}
		tday[0] = jd[0]
		retval = state.getAcronychalDay(jd[0], location, datm, dobs, ObjectName, helflag2, TypeEvent, tday[:], message[:])
		if retval != 0 {
			if serrRet != nil && int32(message[0]) != '\000' {
				copyText(serrRet, message[:])
			}
			return retval
		}
	}
	dret[0] = tday[0]
	if !(helflag&1024 != 0) {
		if body == 2 || body == 3 || TypeEvent <= 2 {
			retval = state.getHeliacalDetails(tday[0], location, datm, dobs, ObjectName, TypeEvent, helflag2, dret, message[:])
			if retval == -1 {
				if serrRet != nil && int32(message[0]) != '\000' {
					copyText(serrRet, message[:])
				}
				return retval
			}
		} else if 0 != 0 {
			if TypeEvent == 4 || TypeEvent == 6 {
				direct = float64(-1)
			}
			i = 0
			for d = 0.0011574074074074073; i < 3; {
				for {
					retval = state.sweVisLimitMag(dret[0]+d*direct, location, datm, dobs, ObjectName, helflag, darr[:], message[:])
					if !(retval == -2 || retval >= 0 && darr[0] < darr[7]) {
						break
					}
					dret[0] += d * direct
				}
				i++
				d /= 10.0
			}
			if retval == 0 {
				dret[0] += 1.1574074074074073e-05 * direct
			}
		}
	}
	if serrRet != nil && int32(message[0]) != '\000' {
		copyText(serrRet, message[:])
	}
	return retval
}

// moonEventVisLim implements moon_event_vis_lim from swehel.c.
//
// ###################################################################
func (state *State) moonEventVisLim(tjdstart float64, location []float64, datm []float64, dobs []float64, TypeEvent int32, helflag int32, dret []float64, serrRet []byte) int32 {
	var (
		jd         [1]float64
		trise      [1]float64
		message    [256]byte
		ObjectName [30]byte
		flags      int32
		body       int32
		retval     int32
		helflag2   int32
		direct     int32
		ephemeris  int32 = helflag & (1 | 2 | 4)
	)
	dret[0] = tjdstart
	if TypeEvent == 1 || TypeEvent == 2 {
		if serrRet != nil {
			copyText(serrRet, textBytes("error: the moon has no morning first or evening last"))
		}
		return -1
	}
	copyText(ObjectName[:], textBytes("moon"))
	body = 1
	flags = 32*1024 | 2*1024 | ephemeris
	if !(helflag&256 != 0) {
		flags |= flagTruePosition | flagNoNutation
	}
	helflag2 = helflag
	helflag2 &= ^256
	jd[0] = tjdstart - 30
	if retval = state.findConjunctSun(jd[0], body, helflag, TypeEvent, jd[:], message[:]); retval == -1 {
		return -1
	}
	retval = state.getHeliacalDay(jd[0], location, datm, dobs, ObjectName[:], helflag2, TypeEvent, jd[:], message[:])
	if retval != 0 {
		if serrRet != nil && int32(message[0]) != '\000' {
			copyText(serrRet, message[:])
		}
		return retval
	}
	dret[0] = jd[0]
	retval = state.timeOptimumVisibility(jd[0], location, datm, dobs, ObjectName[:], helflag, jd[:], message[:])
	if retval == -1 {
		if serrRet != nil && int32(message[0]) != '\000' {
			copyText(serrRet, message[:])
		}
		return retval
	}
	dret[1] = jd[0]
	direct = 1
	if TypeEvent == 4 {
		direct = -1
	}
	retval = state.timeLimitInvisible(jd[0], location, datm, dobs, ObjectName[:], helflag, direct, jd[:], message[:])
	if retval == -1 {
		if serrRet != nil && int32(message[0]) != '\000' {
			copyText(serrRet, message[:])
		}
		return retval
	}
	dret[2] = jd[0]
	direct *= -1
	retval = state.timeLimitInvisible(dret[1], location, datm, dobs, ObjectName[:], helflag, direct, jd[:], message[:])
	dret[0] = jd[0]
	if retval == -1 {
		if serrRet != nil && int32(message[0]) != '\000' {
			copyText(serrRet, message[:])
		}
		return retval
	}
	if TypeEvent == 3 {
		if retval = state.myRiseTrans(jd[0], 0, textBytes(""), 2, helflag, location, datm, trise[:], message[:]); retval == -1 {
			return -1
		}
		if trise[0] < dret[1] {
			dret[0] = trise[0]
		}
	} else {
		if retval = state.myRiseTrans(dret[1], 0, textBytes(""), 1, helflag, location, datm, trise[:], message[:]); retval == -1 {
			return -1
		}
		if dret[0] > trise[0] {
			dret[0] = trise[0]
		}
	}
	if TypeEvent == 4 {
		jd[0] = dret[0]
		dret[0] = dret[2]
		dret[2] = jd[0]
	}
	if serrRet != nil && int32(message[0]) != '\000' {
		copyText(serrRet, message[:])
	}
	return retval
}

// MoonEventJDut implements MoonEventJDut from swehel.c.
func (state *State) MoonEventJDut(JDNDaysUTStart float64, location []float64, datm []float64, dobs []float64, TypeEvent int32, helflag int32, dret []float64, message []byte) int32 {
	var avkind int32 = helflag & (1<<16 | 1<<17 | 1<<18 | 1<<19)
	if avkind != 0 {
		return state.moonEventArcVis(JDNDaysUTStart, location, datm, dobs, TypeEvent, helflag, dret, message)
	} else {
		return state.moonEventVisLim(JDNDaysUTStart, location, datm, dobs, TypeEvent, helflag, dret, message)
	}
}

// heliacalUt implements heliacal_ut from swehel.c.
func (state *State) heliacalUt(JDNDaysUTStart float64, location []float64, datm []float64, dobs []float64, ObjectName []byte, TypeEventIn int32, helflag int32, dret []float64, serrRet []byte) int32 {
	var avkind int32 = helflag & (1<<16 | 1<<17 | 1<<18 | 1<<19)
	if avkind != 0 {
		return state.heliacalUtArcVis(JDNDaysUTStart, location, datm, dobs, ObjectName, TypeEventIn, helflag, dret, serrRet)
	} else {
		return state.heliacalUtVisLim(JDNDaysUTStart, location, datm, dobs, ObjectName, TypeEventIn, helflag, dret, serrRet)
	}
}

// sweHeliacalUt implements swe_heliacal_ut from swehel.c.
func (state *State) sweHeliacalUt(JDNDaysUTStart float64, location []float64, datm []float64, dobs []float64, ObjectNameIn []byte, TypeEvent int32, helflag int32, dret []float64, serrRet []byte) int32 {
	var (
		retval                int32
		Planet                int32
		itry                  int32
		ObjectName            [256]byte
		message               [256]byte
		s                     [256]byte
		initialJD             float64 = JDNDaysUTStart
		jd                    float64
		dsynperiod            float64
		tjdmax                float64
		tadd                  float64
		MaxCountSynodicPeriod int32     = 5
		sevent                [7][]byte = [7][]byte{textBytes(""), textBytes("morning first"), textBytes("evening last"), textBytes("evening first"), textBytes("morning last"), textBytes("acronychal rising"), textBytes("acronychal setting")}
	)
	if location[2] < -500.0 || location[2] > 25000.0 {
		if serrRet != nil {
			formatText(serrRet, textBytes("location for heliacal events must be between %.0f and %.0f m above sea\n"), -500.0, 25000.0)
		}
		return -1
	}
	state.swiSetTidAcc(JDNDaysUTStart, helflag, 0, message[:])
	if helflag&128 != 0 {
		MaxCountSynodicPeriod = 1000000
	}
	if serrRet != nil {
		serrRet[0] = byte('\000')
	}
	state.strcpyVBsafe(ObjectName[:], ObjectNameIn)
	state.tolowerStringStar(ObjectName[:])
	state.defaultHeliacalParameters(datm, location, dobs, helflag)
	state.sweSetTopo(location[0], location[1], location[2])
	Planet = state.DeterObject(ObjectName[:])
	if Planet == 0 {
		if serrRet != nil {
			copyText(serrRet, textBytes("the sun has no heliacal rising or setting\n"))
		}
		return -1
	}
	if Planet == 1 {
		if TypeEvent == 1 || TypeEvent == 2 {
			if serrRet != nil {
				formatText(serrRet, textBytes("%s (event type %d) does not exist for the moon\n"), sevent[TypeEvent], TypeEvent)
			}
			return -1
		}
		jd = initialJD
		for retval = state.MoonEventJDut(jd, location, datm, dobs, TypeEvent, helflag, dret, message[:]); retval != -2 && dret[0] < initialJD; {
			jd += 15
			message[0] = byte('\000')
			retval = state.MoonEventJDut(jd, location, datm, dobs, TypeEvent, helflag, dret, message[:])
		}
		if serrRet != nil && int32(message[0]) != '\000' {
			copyText(serrRet, message[:])
		}
		return retval
	}
	if !(helflag&(1<<16|1<<17|1<<18|1<<19) != 0) {
		if Planet == -1 || Planet >= 4 {
			if TypeEvent == 3 || TypeEvent == 4 {
				if serrRet != nil {
					if Planet == -1 {
						copyText(s[:], ObjectName[:])
					} else {
						state.sweGetPlanetName(Planet, s[:])
					}
					formatText(serrRet, textBytes("%s (event type %d) does not exist for %s\n"), sevent[TypeEvent], TypeEvent, s[:])
				}
				return -1
			}
		}
	}
	if helflag&(1<<16|1<<17|1<<18|1<<19) != 0 {
		if Planet == -1 || Planet >= 4 {
			if TypeEvent == 5 {
				TypeEvent = 3
			}
			if TypeEvent == 6 {
				TypeEvent = 4
			}
		}
	} else if 1 != 0 {
		if TypeEvent == 5 || TypeEvent == 6 {
			if serrRet != nil {
				if Planet == -1 {
					copyText(s[:], ObjectName[:])
				} else {
					state.sweGetPlanetName(Planet, s[:])
				}
				formatText(serrRet, textBytes("%s (event type %d) is not provided for %s\n"), sevent[TypeEvent], TypeEvent, s[:])
			}
			return -1
		}
	}
	dsynperiod = state.getSynodicPeriod(Planet)
	tjdmax = initialJD + dsynperiod*float64(MaxCountSynodicPeriod)
	tadd = dsynperiod * 0.6
	if Planet == 2 {
		tadd = 30
	}
	jd = initialJD
	retval = -2
	for itry = 0; jd < tjdmax && retval == -2; {
		message[0] = byte('\000')
		for retval = state.heliacalUt(jd, location, datm, dobs, ObjectName[:], TypeEvent, helflag, dret, message[:]); retval != -2 && dret[0] < initialJD; {
			jd += tadd
			message[0] = byte('\000')
			retval = state.heliacalUt(jd, location, datm, dobs, ObjectName[:], TypeEvent, helflag, dret, message[:])
		}
		itry++
		jd += tadd
	}
	if helflag&(1<<11) != 0 && (retval == -2 || dret[0] > initialJD+dsynperiod*1.5) {
		copyText(message[:], textBytes("no heliacal date found within this synodic period"))
		retval = -2
	} else if retval == -2 {
		formatText(message[:], textBytes("no heliacal date found within %d synodic periods"), MaxCountSynodicPeriod)
		retval = -1
	}
	if serrRet != nil && int32(message[0]) != '\000' {
		copyText(serrRet, message[:])
	}
	return retval
}
