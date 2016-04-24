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

//import "bytes"
import "encoding/binary"
import "net"
import "fmt"
import "bufio"
import "time"
//import "os"
//import "strconv"
import "reflect"


var MAGIC_MLAT_TIMESTAMP = []byte{0xFF, 0x00, 0x4D, 0x4C, 0x41, 0x54}

func main() {
  fmt.Println("Launching server...")

  // listen on all interfaces
  ln, _ := net.Listen("tcp", ":8081")

  // accept connection on port
  conn, _ := ln.Accept()

  reader := bufio.NewReader(conn)

  var buffered_message []byte

  // run loop forever (or until ctrl-c)
  for {
    current_message, err := reader.ReadBytes(0x1A)
    if err != nil {
      fmt.Println("ERR:", err)
    }
    // Note `message` does not include 0x1A start byte b/c ReadBytes behavior

    if len(current_message) == 0 {
      break
    }

    // Add to our own buffer (or create it)
    if buffered_message == nil {
      buffered_message = current_message
    } else {
      buffered_message = append(buffered_message, current_message...)
    }

    // Are we on a *real* "message start" boundary? Then we're done
    // with our buffer.
    parseBuffer := false
    switch current_message[0] {
    case 0x31, 0x32, 0x33, 0x34:
      parseBuffer = true
    }

    if !parseBuffer {
      continue
    } else {
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
          msgLen = 15 // 7 + 8 header
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

      fmt.Println()
      var timestamp time.Time
      if reflect.DeepEqual(message[1:7], MAGIC_MLAT_TIMESTAMP) {
        fmt.Println("MAGIC TIMESTAMP!")
        otimestamp := parseTime(message[1:7])
        fmt.Println(otimestamp)
        timestamp = time.Now()
      } else {
        timestamp = parseTime(message[1:7])
      }
      fmt.Println(timestamp)
      switch msgType {
        //case 0x31:
        //  fmt.Println("Type 1 Mode-AC")
        case 0x32:
          fmt.Println("Type 2 Mode-S short")
        case 0x33:
          fmt.Println("Type 3 Mode-S long")
        //case 0x34:
        //  fmt.Println("Status Signal")
      }

      sigLevel := message[7]
      fmt.Printf("Signal: %#02x (%d)\n", sigLevel, sigLevel)

      msgContent := message[8:len(message)-1]
      fmt.Printf("%d byte frame\n", len(msgContent))
      for i:= 0; i < len(msgContent); i++ {
        fmt.Printf("%02x", msgContent[i])
      }
      fmt.Println()


      parseModeS(msgContent)
    }

  }

}

func parseModeS(message []byte) {
  // https://en.wikipedia.org/wiki/Secondary_surveillance_radar#Mode_S
  // https://github.com/mutability/dump1090/blob/master/mode_s.c
  linkFmt := int64((message[0] & 0xF8) >> 3)

  var msgType string
  switch linkFmt {
  case 0:
    msgType = "short air-air surveillance (TCAS)"
  case 4:
    msgType = "surveillance, altitude reply"
  case 5:
    msgType = "surveillance, Mode A identity reply"
  case 11:
    msgType = "All-Call reply containing aircraft address"
  case 16:
    msgType = "long air-air surveillance (TCAS)"
  case 17:
    msgType = "extended squitter"
  case 18:
    msgType = "TIS-B"
  case 19:
    msgType = "military extended squitter"
  case 20:
    msgType = "Comm-B including altitude reply"
  case 21:
    msgType = "Comm-B reply including Mode A identity"
  case 22:
    msgType = "military use"
  case 24:
    msgType = "special long msg"
  default:
    msgType = "unknown"
  }
  //fmt.Printf("UF: %d\n", linkFmt)
  //fmt.Printf("UF: %08s\n", strconv.FormatInt(linkFmt, 2))
  fmt.Println(msgType)


  icaoCode := int(message[1])*65536+int(message[2])*256+int(message[3])
  fmt.Printf("ICAO: %06x\n", icaoCode)

  fmt.Println()
}

func parseTime(timebytes []byte) time.Time {
  // Takes a 6 byte array, which represents a 48bit GPS timestamp
  // http://wiki.modesbeast.com/Radarcape:Firmware_Versions#The_GPS_timestamp
  // and parses it into a Time.time

  upper := []byte{
    timebytes[0] << 2 + timebytes[1] >> 6,
    timebytes[1] << 2 + timebytes[2] >> 6}
  lower := []byte{
    timebytes[2]&0x3F, timebytes[3], timebytes[4], timebytes[5]}

  //for i:= 0; i < len(upper); i++ {
  //  fmt.Printf("%#02x, ", upper[i])
  //}
  //fmt.Print("\n")
  //for i:= 0; i < len(lower); i++ {
  //  fmt.Printf("%#02x, ", lower[i])
  //}
  //fmt.Print("\n")

  daySeconds  := binary.BigEndian.Uint16(upper)
  nanoSeconds := int(binary.BigEndian.Uint32(lower))

  hr  := int(daySeconds/3600)
  min := int(daySeconds/60 % 60)
  sec := int(daySeconds%60)

  //fmt.Print("\n")
  //fmt.Println(daySeconds)
  //fmt.Println(nanoSeconds)
  //fmt.Println(hr)
  //fmt.Println(min)
  //fmt.Println(sec)

  localDate := time.Now()

  utcDate := localDate.UTC()

  //var t time.Time
  //if (utcDate != localDate) && (hr == localDate.Hour()) {
  //  t = time.Date(
  //    localDate.Year(), localDate.Month(), localDate.Day(),
  //    hr, min, sec, nanoSeconds, time.Local)
  //} else {
  //  t = time.Date(
  //    utcDate.Year(), utcDate.Month(), utcDate.Day(),
  //    hr, min, sec, nanoSeconds, time.UTC)
  //}
  //return t

  return time.Date(
    utcDate.Year(), utcDate.Month(), utcDate.Day(),
    hr, min, sec, nanoSeconds, time.UTC)
}
