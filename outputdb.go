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
	//"fmt"
	"math"
	"database/sql"
)

func dbWriteModeS(db *sql.DB, raw_message []byte, msgdata *MessageInfo, aircraft *Aircraft, prev_aircraft *Aircraft) {
	//fmt.Println(msgdata.callsign)

	if (msgdata.altitude == 0) && (aircraft.latitude == math.MaxFloat64 || aircraft.longitude == math.MaxFloat64) && ((aircraft.callsign == "") || (aircraft.callsign != "" && prev_aircraft.callsign !="")) ||
	(msgdata.msgType==4 || msgdata.msgType==19 || msgdata.msgType==28 || msgdata.msgType==29 || msgdata.msgType==31) {
		return
	}

	if (aircraft.latitude != math.MaxFloat64 && aircraft.longitude != math.MaxFloat64) && ((aircraft.latitude != prev_aircraft.latitude) || (aircraft.longitude != prev_aircraft.longitude)) {
		// we updated our position
    _, err := db.Exec(`INSERT INTO
  		squitters (
  			client_id, icao_addr, message_type, transmission_type, parsed_time, generated_datetime, is_mlat, altitude, callsign, lat, lon
  		) VALUES (
  			?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?
  	)`,
      11,
  		msgdata.icaoAddr,
  		"MSG", // ???
  		msgdata.msgType,
  		msgdata.localTime,
  		msgdata.msgTime,
  		msgdata.mlat,
  		msgdata.altitude,
  		aircraft.callsign,
      aircraft.latitude,
      aircraft.longitude)
  	if err != nil {
  		panic(err.Error())
  	}
	} else {
		// we did not update latlon
    _, err := db.Exec(`INSERT INTO
  		squitters (
  			client_id, icao_addr, message_type, transmission_type, parsed_time, generated_datetime, is_mlat, altitude, callsign
  		) VALUES (
  			?, ?, ?, ?, ?, ?, ?, ?, ?
  	)`,
      11,
  		msgdata.icaoAddr,
  		"MSG", // ???
  		msgdata.msgType,
  		msgdata.localTime,
  		msgdata.msgTime,
  		msgdata.mlat,
  		msgdata.altitude,
  		aircraft.callsign)
  	if err != nil {
  		panic(err.Error())
  	}
	}
}
