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
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program. If not, see <http://www.gnu.org/licenses/>.
package main

import (
	"bufio"
	"flag"
	"fmt"
	"net"
	//"os"
	"reflect"
	"time"
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
)

var magicTimestampMLAT = []byte{0xFF, 0x00, 0x4D, 0x4C, 0x41, 0x54}

const (
	aisCharset       = "@ABCDEFGHIJKLMNOPQRSTUVWXYZ[\\]^_ !\"#$%&'()*+,-./0123456789:;<=>?"
	sortModeLastPos  = uint(0)
	sortModeDistance = uint(1)
	sortModeCallsign = uint(2)
)

var (
	sourceAddr = flag.String("server", "127.0.0.1:30005", "\":port\" or \"ip:port\" of the BEAST server to connect to")
	baseLat    = flag.Float64("baseLat", 40.77725, "latitude used for distance calculation")
	baseLon    = flag.Float64("baseLon", -73.872611, "longitude for distance calculation")
	sortMode   = flag.Uint("sortMode", sortModeDistance, "0: sort by time, 1: sort by distance, 3: sort by air")
)

func main() {
	flag.Parse()

	// test: http://www.lll.lu/~edward/edward/adsb/DecodingADSBposition.html
	// parseRawLatLon(uint32(92095), uint32(39846), uint32(88385), uint32(125818), true, false)
	// test: http://wiki.modesbeast.com/Radarcape:Firmware_Versions#The_GPS_timestamp
	//timestamp := parseTime([]byte{0x24, 0x4b, 0xbb, 0x9a, 0xc9, 0xf0})
	//fmt.Println(timestamp)
	//os.Exit(1)

	fmt.Println("Launching server...")

	known_aircraft := make(aircraftMap)

	conns := serverConn(*sourceAddr)

	db, err := sql.Open("mysql", "root@/dump1090")
	if err != nil {
		fmt.Println("Error connecting to DB server:")
		fmt.Println(err)
	}
	err = db.Ping()
	if err != nil {
		fmt.Println("Error connecting to DB server:")
		fmt.Println(err)
	}


	quit := make(chan struct{})
	ticker := time.NewTicker(500 * time.Millisecond)
	go func() {
		for {
			select {
			case <-ticker.C:
				printAircraftTable(&known_aircraft)
			case <-quit:
				ticker.Stop()
				return
			}
		}
	}()



	for {
		go handleConnection(<-conns, &known_aircraft, db)
	}
}

func serverConn(sourceAddr string) chan net.Conn {
	ch := make(chan net.Conn)

	//i := 0
	go func() {
			conn, err := net.Dial("tcp", sourceAddr)
			if err != nil {
				fmt.Println("Error connecting to server:")
				fmt.Println(err)
			}
			//i++
			//fmt.Printf("%d: %v <-> %v\n", i, conn.LocalAddr(), conn.RemoteAddr())
			ch <- conn
	}()
	return ch
}

func handleConnection(conn net.Conn, known_aircraft *aircraftMap, db *sql.DB) {
	reader := bufio.NewReader(conn)

	var buffered_message []byte
	// listen to this connection forever
	for {
		current_message, err := reader.ReadBytes(0x1A)
		if err != nil {
			fmt.Println("ERR:", err)
		}
		//fmt.Println(current_message)
		// Note `message` does not include 0x1A start byte b/c ReadBytes behavior

		//if len(current_message) == 0 {
			//break
		//	continue
		//}

		// Add to our own buffer (or create it)
		if buffered_message == nil {
			buffered_message = current_message
		} else {
			buffered_message = append(buffered_message, current_message...)
		}

		// Are we on a *real* "message start" boundary? Then we're done
		// with our buffer.
		parseBuffer := false
		if current_message[0] == 0x31 || current_message[0] == 0x32 ||
			current_message[0] == 0x33 || current_message[0] == 0x34 {
			parseBuffer = true
		}

		if !parseBuffer {
			continue
		}
		message := buffered_message
		buffered_message = nil

		msgType := message[0]
		var msgLen int

		switch msgType {
		case 0x31:
			//fmt.Print("Type 1 Mode-AC")
			//msgLen = 10 // 2 + 8 header
			continue // not supported yet
		case 0x32:
			//fmt.Print("Type 2 Mode-S short")
			//msgLen = 15 // 7 + 8 header
			continue // later
		case 0x33:
			//fmt.Print("Type 3 Mode-S long")
			msgLen = 22 // 14
		case 0x34:
			//fmt.Print("Status Signal")
			//msgLen = 10 // ??
			continue // not supported
		default:
			continue
			//msgLen = 8 // shortest possible msg w/header & timetstamp
		}

		// Message wasn't long enough to contain the full header (maybe
		// input stream error), so skip
		if len(message) < msgLen {
			continue
		}

		//fmt.Println()
		var timestamp time.Time
		isMlat := reflect.DeepEqual(message[1:7], magicTimestampMLAT)
		if isMlat {
			//fmt.Println("FROM MLAT")
			//otimestamp := parseTime(message[1:7])
			//fmt.Println(otimestamp)
			//timestamp = time.Now()
		} else {
			timestamp = parseTime(message[1:7])
			//fmt.Println(timestamp)
		}
		switch msgType {
		//case 0x31:
			//fmt.Println("Type 1 Mode-AC")
		case 0x32:
			//fmt.Println("Type 2 Mode-S short")
		case 0x33:
			//fmt.Println("Type 3 Mode-S long")
		//case 0x34:
			//fmt.Println("Status Signal")
		}

		//sigLevel := message[7]
		//fmt.Printf("Signal: %#02x (%d)\n", sigLevel, sigLevel)

		msgContent := message[8 : len(message)-1]
		////fmt.Printf("%d byte frame\n", len(msgContent))
		//for i:= 0; i < len(msgContent); i++ {
		//	fmt.Printf("%02x", msgContent[i])
		//}
		//fmt.Println()

		parseModeS(msgContent, isMlat, timestamp, known_aircraft, db)
		//fmt.Println()

		//printAircraftTable(known_aircraft)
	}
}
