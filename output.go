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
	"fmt"
	"math"
	"sort"
	"time"
)

func durationSecondsElapsed(since time.Duration) string {
	sec := uint8(since.Seconds())
	if sec == math.MaxUint8 {
		return "-"
	} else {
		return fmt.Sprintf("%4d", sec)
	}
}

func printAircraftTable(known_aircraft *aircraftMap) {
	fmt.Println("ICAO  \tCallsign\tLocation\t\tAlt\tDistance")

	sortedAircraft := make(aircraftList, 0, len(*known_aircraft))

	for _, aircraft := range *known_aircraft {
		sortedAircraft = append(sortedAircraft, aircraft)
	}

	sort.Sort(sortedAircraft)

	for _, aircraft := range sortedAircraft {
		if time.Since(aircraft.lastPos) > (time.Duration(45) * time.Second) {
			continue
		}
		stale := (time.Since(aircraft.lastPos) > (time.Duration(10) * time.Second))
		extraStale := (time.Since(aircraft.lastPos) > (time.Duration(20) * time.Second))

		aircraftHasLocation := (aircraft.latitude != math.MaxFloat64 &&
			aircraft.longitude != math.MaxFloat64)
		aircraftHasAltitude := aircraft.altitude != math.MaxInt32

		if !aircraftHasLocation {
			continue
		}

		if aircraft.callsign != "" || aircraftHasLocation || aircraftHasAltitude {
			var sLatLon string
			var sAlt string

			if aircraftHasLocation {
				sLatLon = fmt.Sprintf("%f,%f", aircraft.latitude, aircraft.longitude)
			} else {
				sLatLon = "---.------,---.------"
			}
			if aircraftHasAltitude {
				sAlt = fmt.Sprintf("%d", aircraft.altitude)
			} else {
				sAlt = "-----"
			}

			distance := greatcircle(aircraft.latitude, aircraft.longitude,
				*baseLat, *baseLon)

			//tPing := time.Since(aircraft.lastPing)
			tPos := time.Since(aircraft.lastPos)

			if !stale && !extraStale {
				fmt.Printf("%06x\t%8s\t%s\t%s\t%3.2f\n",
					aircraft.icaoAddr, aircraft.callsign,
					sLatLon, sAlt, metersInMiles(distance))
			} else if stale && !extraStale {
				fmt.Printf("%06x\t%8s\t%s?\t%s\t%3.2f?\n",
					aircraft.icaoAddr, aircraft.callsign,
					sLatLon, sAlt, metersInMiles(distance))
			} else if extraStale {
				fmt.Printf("%06x\t%8s\t%s?\t%s\t%3.2f?\t%s…\n",
					aircraft.icaoAddr, aircraft.callsign,
					sLatLon, sAlt, metersInMiles(distance),
					durationSecondsElapsed(tPos))
			}
		}
	}
	fmt.Println()
}
