// This file is part of Simurgh.
// Copyright © 2016 Mike Tigas. All rights reserved.
// This file is licensed under the terms of the GNU Affero General
// Public License, version 3 or later. See the LICENSE.md file.
//
// Some functions in this file are ported from code in mutability/dump1090
// <https://github.com/mutability/dump1090>, under the GNU Public
// License v2.
package main

import (
	"math"
)

func cprNLFunction(lat float64) byte {
	if lat < 0 {
		lat = -lat
	}
	switch {
	case (lat < 10.47047130):
		return 59
	case (lat < 14.82817437):
		return 58
	case (lat < 18.18626357):
		return 57
	case (lat < 21.02939493):
		return 56
	case (lat < 23.54504487):
		return 55
	case (lat < 25.82924707):
		return 54
	case (lat < 27.93898710):
		return 53
	case (lat < 29.91135686):
		return 52
	case (lat < 31.77209708):
		return 51
	case (lat < 33.53993436):
		return 50
	case (lat < 35.22899598):
		return 49
	case (lat < 36.85025108):
		return 48
	case (lat < 38.41241892):
		return 47
	case (lat < 39.92256684):
		return 46
	case (lat < 41.38651832):
		return 45
	case (lat < 42.80914012):
		return 44
	case (lat < 44.19454951):
		return 43
	case (lat < 45.54626723):
		return 42
	case (lat < 46.86733252):
		return 41
	case (lat < 48.16039128):
		return 40
	case (lat < 49.42776439):
		return 39
	case (lat < 50.67150166):
		return 38
	case (lat < 51.89342469):
		return 37
	case (lat < 53.09516153):
		return 36
	case (lat < 54.27817472):
		return 35
	case (lat < 55.44378444):
		return 34
	case (lat < 56.59318756):
		return 33
	case (lat < 57.72747354):
		return 32
	case (lat < 58.84763776):
		return 31
	case (lat < 59.95459277):
		return 30
	case (lat < 61.04917774):
		return 29
	case (lat < 62.13216659):
		return 28
	case (lat < 63.20427479):
		return 27
	case (lat < 64.26616523):
		return 26
	case (lat < 65.31845310):
		return 25
	case (lat < 66.36171008):
		return 24
	case (lat < 67.39646774):
		return 23
	case (lat < 68.42322022):
		return 22
	case (lat < 69.44242631):
		return 21
	case (lat < 70.45451075):
		return 20
	case (lat < 71.45986473):
		return 19
	case (lat < 72.45884545):
		return 18
	case (lat < 73.45177442):
		return 17
	case (lat < 74.43893416):
		return 16
	case (lat < 75.42056257):
		return 15
	case (lat < 76.39684391):
		return 14
	case (lat < 77.36789461):
		return 13
	case (lat < 78.33374083):
		return 12
	case (lat < 79.29428225):
		return 11
	case (lat < 80.24923213):
		return 10
	case (lat < 81.19801349):
		return 9
	case (lat < 82.13956981):
		return 8
	case (lat < 83.07199445):
		return 7
	case (lat < 83.99173563):
		return 6
	case (lat < 84.89166191):
		return 5
	case (lat < 85.75541621):
		return 4
	case (lat < 86.53536998):
		return 3
	case (lat < 87.00000000):
		return 2
	default:
		return 1
	}
}
func cprNFunction(lat float64, fflag bool) byte {
	var t byte
	if fflag {
		t = 1
	} else {
		t = 0
	}

	nl := cprNLFunction(lat) - t
	if nl < 1 {
		nl = 1
	}
	return nl
}
func cprDlonFunction(lat float64, fflag bool, surface bool) float64 {
	var sfc float64
	if surface {
		sfc = 90.0
	} else {
		sfc = 360.0
	}

	return sfc / float64(cprNFunction(lat, fflag))

}

func decodeAC12Field(ac12Data uint) int32 {
	q := (ac12Data & 0x10) == 0x10
	if q {
		n := int32((ac12Data&0x0FE0)>>1) + int32(ac12Data&0x000F)
		return (n * 25) - 1000
	} else {
		/* TODO
		// Make N a 13 bit Gillham coded altitude by inserting M=0 at bit 6
		int n = ((AC12Field & 0x0FC0) << 1) |
						 (AC12Field & 0x003F);
		n = ModeAToModeC(decodeID13Field(n));
		if (n < -12) {
				return INVALID_ALTITUDE;
		}

		return (100 * n);
		*/
		return int32(math.MaxInt32)
	}
}

func greatcircle(lat0, lon0, lat1, lon1 float64) float64 {
	lat0 = lat0 * math.Pi / 180.0
	lon0 = lon0 * math.Pi / 180.0
	lat1 = lat1 * math.Pi / 180.0
	lon1 = lon1 * math.Pi / 180.0

	// avoid NaN
	if math.Abs(lat0-lat1) < 0.0001 && math.Abs(lon0-lon1) < 0.0001 {
		return 0.0
	}

	return 6371e3 * math.Acos(math.Sin(lat0)*math.Sin(lat1)+math.Cos(lat0)*math.Cos(lat1)*math.Cos(math.Abs(lon0-lon1)))
}

func metersInMiles(dist float64) float64 {
	return dist / float64(1609.34721869)
}
