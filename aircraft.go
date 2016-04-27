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
	"math"
)

type Aircraft struct {
	icaoAddr uint32

	callsign string

	eRawLat uint32
	eRawLon uint32
	oRawLat uint32
	oRawLon uint32

	latitude  float64
	longitude float64
	altitude  int32

	lastPing time.Time
	lastPos  time.Time
}
type aircraftList []*Aircraft
type aircraftMap map[uint32]*Aircraft

func (a aircraftList) Len() int {
	return len(a)
}
func (a aircraftList) Swap(i, j int) {
	a[i], a[j] = a[j], a[i]
}
func (a aircraftList) Less(i, j int) bool {
	if *sortMode == sortModeLastPos {
		// t1 later than t2 means that t1 is more recent
		return a[i].lastPos.After(a[j].lastPos)

	} else if *sortMode == sortModeDistance {
		if a[i].latitude != math.MaxFloat64 && a[j].latitude != math.MaxFloat64 {
			return sortByDistance(a, i, j)
		} else if a[i].latitude != math.MaxFloat64 && a[j].latitude == math.MaxFloat64 {
			return true
		} else if a[i].latitude == math.MaxFloat64 && a[j].latitude != math.MaxFloat64 {
			return false
		} else {
			return sortByCallsign(a, i, j)
		}
	} else if *sortMode == sortModeCallsign {
		return sortByCallsign(a, i, j)
	} else {
		// ?
		//return a[i].lastPos > a[j].lastPos
	}
	return false
}

func sortAircraftByDistance(a aircraftList, i, j int) bool {
	dist_i := greatcircle(a[i].latitude, a[i].longitude,
		*baseLat, *baseLon)
	dist_j := greatcircle(a[j].latitude, a[j].longitude,
		*baseLat, *baseLon)
	return dist_i < dist_j
}
func sortAircraftByDistance(a aircraftList, i, j int) bool {
	if a[i].callsign != "" && a[j].callsign != "" {
		return a[i].callsign < a[j].callsign
	} else if a[i].callsign != "" && a[j].callsign == "" {
		return true
	} else if a[i].callsign == "" && a[j].callsign != "" {
		return false
	} else {
		hexi := fmt.Sprintf("%06x", a[i].icaoAddr)
		hexj := fmt.Sprintf("%06x", a[j].icaoAddr)
		return hexi < hexj
	}
}
