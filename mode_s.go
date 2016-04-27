// simurgh
// Copyright © 2016 Mike Tigas
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as
// published by the Free Software Foundation, either version 3 of the
// License, or (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.
package main

import (
	"encoding/binary"
	"math"
)

func parseModeS(message []byte, known_aircraft *aircraftMap) {
	// https://en.wikipedia.org/wiki/Secondary_surveillance_radar#Mode_S
	// https://github.com/mutability/dump1090/blob/master/mode_s.c
	linkFmt := uint((message[0] & 0xF8) >> 3)

	var aircraft Aircraft
	var aircraft_exists bool
	icaoAddr := uint32(math.MaxUint32)
	altCode := uint16(math.MaxUint16)
	altitude := int32(math.MaxInt32)

	//var msgType string
	//switch linkFmt {
	//case 0:
	//  msgType = "short air-air surveillance (TCAS)"
	//case 4:
	//  msgType = "surveillance, altitude reply"
	//case 5:
	//  msgType = "surveillance, Mode A identity reply"
	//case 11:
	//  msgType = "All-Call reply containing aircraft address"
	//case 16:
	//  msgType = "long air-air surveillance (TCAS)"
	//case 17:
	//  msgType = "extended squitter"
	//case 18:
	//  msgType = "TIS-B"
	//case 19:
	//  msgType = "military extended squitter"
	//case 20:
	//  msgType = "Comm-B including altitude reply"
	//case 21:
	//  msgType = "Comm-B reply including Mode A identity"
	//case 22:
	//  msgType = "military use"
	//case 24:
	//  msgType = "special long msg"
	//default:
	//  msgType = "unknown"
	//}
	//fmt.Printf("UF: %d\n", linkFmt)
	//fmt.Printf("UF: %08s\n", strconv.FormatInt(linkFmt, 2))
	//fmt.Println(msgType)

	if linkFmt == 11 || linkFmt == 17 || linkFmt == 18 {
		icaoAddr = uint32(message[1])*65536 + uint32(message[2])*256 + uint32(message[3])
		//fmt.Printf("ICAO: %06x\n", icaoAddr)
	}

	if icaoAddr != math.MaxUint32 {
		var ptrAircraft *Aircraft
		ptrAircraft, aircraft_exists = (*known_aircraft)[icaoAddr]
		if !aircraft_exists {
			// initialize some values
			aircraft = Aircraft{
				icaoAddr:  icaoAddr,
				oRawLat:   math.MaxUint32,
				oRawLon:   math.MaxUint32,
				eRawLat:   math.MaxUint32,
				eRawLon:   math.MaxUint32,
				latitude:  math.MaxFloat64,
				longitude: math.MaxFloat64,
				altitude:  math.MaxInt32,
				callsign:  ""}
		} else {
			aircraft = (*ptrAircraft)
		}
		aircraft.lastPing = time.Now()
	}
	//fmt.Println(aircraft)
	//fmt.Println(aircraft_exists)

	if linkFmt == 0 || linkFmt == 4 || linkFmt == 16 || linkFmt == 20 {
		// Altitude: 13 bit signal
		altCode = (uint16(message[2])*256 + uint16(message[3])) & 0x1FFF

		if (altCode & 0x0040) > 0 {
			// meters
			// TODO
			fmt.Println("meters")

		} else if (altCode & 0x0010) > 0 {
			// feet, raw integer
			ac := (altCode&0x1F80)>>2 + (altCode&0x0020)>>1 + (altCode & 0x000F)
			altitude = int32((ac * 25) - 1000)
			// TODO
			//fmt.Println("int altitude: ", altitude)

		} else if (altCode & 0x0010) == 0 {
			// feet, Gillham coded
			// TODO
			fmt.Println("gillham")
		}

		if altitude != math.MaxInt32 {
			aircraft.altitude = altitude
		}
	}

	if linkFmt == 17 || linkFmt == 18 {
		decodeExtendedSquitter(message, linkFmt, &aircraft)
	}

	if icaoAddr != math.MaxUint32 {
		(*known_aircraft)[icaoAddr] = &aircraft
	}
	//fmt.Println(aircraft)
}

func parseTime(timebytes []byte) time.Time {
	// Takes a 6 byte array, which represents a 48bit GPS timestamp
	// http://wiki.modesbeast.com/Radarcape:Firmware_Versions#The_GPS_timestamp
	// and parses it into a Time.time

	upper := []byte{
		timebytes[0]<<2 + timebytes[1]>>6,
		timebytes[1]<<2 + timebytes[2]>>6,
		0, 0, 0, 0}
	lower := []byte{
		timebytes[2] & 0x3F, timebytes[3], timebytes[4], timebytes[5]}

	// the 48bit timestamp is 18bit day seconds | 30bit nanoseconds
	daySeconds := binary.BigEndian.Uint16(upper)
	nanoSeconds := int(binary.BigEndian.Uint32(lower))

	hr := int(daySeconds / 3600)
	min := int(daySeconds / 60 % 60)
	sec := int(daySeconds % 60)

	utcDate := time.Now().UTC()

	return time.Date(
		utcDate.Year(), utcDate.Month(), utcDate.Day(),
		hr, min, sec, nanoSeconds, time.UTC)
}

func decodeExtendedSquitter(message []byte, linkFmt uint,
	aircraft *Aircraft) {

	var callsign string

	//if linkFmt == 18 {
	//  switch (message[0] & 7) {
	//  case 1:
	//    fmt.Println("Non-ICAO")
	//  case 2:
	//    fmt.Println("TIS-B fine")
	//  case 3:
	//    fmt.Println("TIS-B coarse")
	//  case 5:
	//    fmt.Println("TIS-B anon ADS-B relay")
	//  case 6:
	//    fmt.Println("ADS-B rebroadcast")
	//  default:
	//    fmt.Println("Non-ICAO unknown")
	//  }
	//}

	msgType := uint(message[4]) >> 3
	var msgSubType uint
	if msgType == 29 {
		msgSubType = (uint(message[4]) & 6) >> 1
	} else {
		msgSubType = uint(message[4]) & 7
	}

	//fmt.Printf("ext msg: %d\n", msgType)

	raw_latitude := uint32(math.MaxUint32)
	raw_longitude := uint32(math.MaxUint32)
	latitude := float64(math.MaxFloat64)
	longitude := float64(math.MaxFloat64)
	altitude := int32(math.MaxInt32)

	switch msgType {
	case 1, 2, 3, 4:
		// Aircraft ID
		chars1 := uint(message[5])<<16 + uint(message[6])<<8 + uint(message[7])
		chars2 := uint(message[8])<<16 + uint(message[9])<<8 + uint(message[10])

		var fltByte [8]byte

		if chars1 != 0 && chars2 != 0 {
			// Flush the buffered raw bits into the representative 8 char string

			fltByte[3] = aisCharset[chars1&0x3F]
			chars1 >>= 6

			fltByte[2] = aisCharset[chars1&0x3F]
			chars1 >>= 6

			fltByte[1] = aisCharset[chars1&0x3F]
			chars1 >>= 6

			fltByte[0] = aisCharset[chars1&0x3F]

			fltByte[7] = aisCharset[chars2&0x3F]
			chars2 >>= 6

			fltByte[6] = aisCharset[chars2&0x3F]
			chars2 >>= 6

			fltByte[5] = aisCharset[chars2&0x3F]
			chars2 >>= 6

			fltByte[4] = aisCharset[chars2&0x3F]

			callsign = string(fltByte[:8])
			//fmt.Println("Callsign: ", callsign)
		}

	//case 19:
	//  // Airborne Velocity

	case 5, 6, 7, 8:
		// Ground position
		raw_latitude = uint32(message[6])&3<<15 + uint32(message[7])<<7 +
			uint32(message[8])>>1
		raw_longitude = uint32(message[8])&1<<16 + uint32(message[9])<<8 +
			uint32(message[10])

	case 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 20, 21, 22:
		// Airborne position

		ac12Data := (uint(message[5]) << 4) + (uint(message[6])>>4)&0x0FFF
		if msgType != 0 {
			raw_latitude = uint32(message[6])&3<<15 + uint32(message[7])<<7 +
				uint32(message[8])>>1
			raw_longitude = uint32(message[8])&1<<16 + uint32(message[9])<<8 +
				uint32(message[10])
		}
		if msgType != 20 && msgType != 21 && msgType != 22 {
			//altitude :=
			//fmt.Printf("ac12: %#04x\n", ac12Data)
			//fmt.Printf("ac12: %d\n", decodeAC12Field(ac12Data))

			altitude = decodeAC12Field(ac12Data)

		} else {
			// "HAE" ac2-encoded altitude
			// TODO
		}
	}

	if (raw_latitude != math.MaxUint32) && (raw_longitude != math.MaxUint32) {
		tFlag := (byte(message[6]) & 8) == 8
		isOddFrame := (byte(message[6]) & 4) == 4

		if isOddFrame && aircraft.eRawLat != math.MaxUint32 && aircraft.eRawLon != math.MaxUint32 {
			// Odd frame and we have previous even frame data
			latitude, longitude = parseRawLatLon(aircraft.eRawLat, aircraft.eRawLon, raw_latitude, raw_longitude, isOddFrame, tFlag)
			// Reset our buffer
			aircraft.eRawLat = math.MaxUint32
			aircraft.eRawLon = math.MaxUint32
		} else if !isOddFrame && aircraft.oRawLat != math.MaxUint32 && aircraft.oRawLon != math.MaxUint32 {
			// Even frame and we have previous odd frame data
			latitude, longitude = parseRawLatLon(raw_latitude, raw_longitude, aircraft.oRawLat, aircraft.oRawLon, isOddFrame, tFlag)
			// Reset buffer
			aircraft.oRawLat = math.MaxUint32
			aircraft.oRawLon = math.MaxUint32
		} else if isOddFrame {
			aircraft.oRawLat = raw_latitude
			aircraft.oRawLon = raw_longitude
		} else if !isOddFrame {
			aircraft.eRawLat = raw_latitude
			aircraft.eRawLon = raw_longitude
		}
	}

	switch msgSubType {
	case 1:
		break
	case 99999:
		os.Exit(1)
	}

	if callsign != "" {
		aircraft.callsign = callsign
	}
	if altitude != math.MaxInt32 {
		aircraft.altitude = altitude
	}
	if latitude != math.MaxFloat64 && longitude != math.MaxFloat64 {
		aircraft.latitude = latitude
		aircraft.longitude = longitude
		aircraft.lastPos = time.Now()
	}
}

func parseRawLatLon(evenLat uint32, evenLon uint32, oddLat uint32,
	oddLon uint32, lastOdd bool, tFlag bool) (latitude float64, longitude float64) {
	if evenLat == math.MaxUint32 || oddLat == math.MaxUint32 ||
		oddLat == math.MaxUint32 || oddLon == math.MaxUint32 {
		return math.MaxFloat64, math.MaxFloat64
	}

	//fmt.Printf("Parsing: %d,%d + %d,%d\n", evenLat, evenLon, oddLat, oddLon)

	// http://www.lll.lu/~edward/edward/adsb/DecodingADSBposition.html
	j := int32((float64(59*evenLat-60*oddLat) / 131072.0) + 0.5)
	//fmt.Println("J: ", j)

	const airdlat0 = float64(6.0)
	const airdlat1 = float64(360.0) / float64(59.0)

	rlatEven := airdlat0 * (float64(j%60) + float64(evenLat)/131072.0)
	rlatOdd := airdlat1 * (float64(j%59) + float64(oddLat)/131072.0)
	if rlatEven >= 270 {
		rlatEven -= 360
	}
	if rlatOdd >= 270 {
		rlatOdd -= 360
	}

	//fmt.Println("rlat(0): ", rlatEven)
	//fmt.Println("rlat(1): ", rlatOdd)

	nlEven := cprNLFunction(rlatEven)
	nlOdd := cprNLFunction(rlatOdd)

	if nlEven != nlOdd {
		return math.MaxFloat64, math.MaxFloat64
	}

	//fmt.Println("NL(0): ", nlEven)
	//fmt.Println("NL(1): ", nlOdd)

	var ni int16

	if lastOdd {
		ni = int16(nlOdd) - 1
	} else {
		ni = int16(nlEven) - 1
	}
	if ni < 1 {
		ni = 1
	}
	//fmt.Println("NL(i): ", ni)

	//dlon := 360.0/float64(ni)
	//fmt.Println("dlon(i):", dlon)

	var m int16
	var outLat float64
	var outLon float64
	if tFlag {
		m = int16(math.Floor((float64(int32(evenLon*uint32(cprNLFunction(rlatOdd)-1))-
			int32(oddLon*uint32(cprNLFunction(rlatOdd)))) / 131072.0) + 0.5))
		outLon = cprDlonFunction(rlatOdd, tFlag, false) * (float64(m%ni) + float64(oddLon)/131072.0)
		outLat = rlatOdd

	} else {
		m = int16(math.Floor((float64(int32(evenLon*uint32(cprNLFunction(rlatEven)-1))-
			int32(oddLon*uint32(cprNLFunction(rlatEven)))) / 131072.0) + 0.5))
		outLon = cprDlonFunction(rlatEven, tFlag, false) * (float64(m%ni) + float64(evenLon)/131072.0)
		outLat = rlatEven
	}

	outLon -= math.Floor((outLon+180.0)/360.0) * 360.0

	//fmt.Println("M: ", m)
	//fmt.Println("outLat: ", outLat)
	//fmt.Println("outLon: ", outLon)

	return outLat, outLon
}
