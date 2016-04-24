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
import "os"
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
  linkFmt := uint((message[0] & 0xF8) >> 3)

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

 if linkFmt == 11 || linkFmt == 17 || linkFmt == 18 {
    icaoAddr := uint(message[1])*65536+uint(message[2])*256+uint(message[3])
    fmt.Printf("ICAO: %06x\n", icaoAddr)
  }

  if linkFmt == 0 || linkFmt == 4 || linkFmt == 16 || linkFmt == 20 {
    // Altitude: 13 bit signal
    altCode := (uint(message[2])*256 + uint(message[3])) & 0x1FFF;
    alt := int(-9999)

    if (altCode & 0x0040) > 0 {
      // meters

    } else if (altCode & 0x0010) > 0 {
      // feet, raw integer
      ac := (altCode & 0x1F80) >> 2 + (altCode & 0x0020) >> 1 + (altCode & 0x000F)
      alt = int((ac * 25) - 1000)

    } else if (altCode & 0x0010) == 0 {
      // feet, Gillham coded
      alt = parseGillhamAlt(altCode)
    }
    fmt.Printf("Alt: %d\n", alt)
  }

  if linkFmt == 17 || linkFmt == 18 {
    decodeExtendedSquitter(message, linkFmt)
  }
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

func parseGillhamAlt(gAlt uint) int {
  onehundreds  := uint(0)
  fivehundreds := uint(0)

  if (gAlt & 0xFFFF8889) > 0 || (gAlt & 0x000000F0) == 0 {
    return -9999;
  }

  if (gAlt & 0x0010) > 0 {onehundreds ^= 0x007;} // C1
  if (gAlt & 0x0020) > 0 {onehundreds ^= 0x003;} // C2
  if (gAlt & 0x0040) > 0 {onehundreds ^= 0x001;} // C4

  // Remove 7s from onehundreds (Make 7->5, snd 5->7).
  if ((onehundreds & 5) == 5) {onehundreds ^= 2;}

  // Check for invalid codes, only 1 to 5 are valid
  if (onehundreds > 5) {return -9999;}

//if (gAlt & 0x0001) > 0 {fivehundreds ^= 0x1FF;} // D1 never used for altitude
  if (gAlt & 0x0002) > 0 {fivehundreds ^= 0x0FF;} // D2
  if (gAlt & 0x0004) > 0 {fivehundreds ^= 0x07F;} // D4

  if (gAlt & 0x1000) > 0 {fivehundreds ^= 0x03F;} // A1
  if (gAlt & 0x2000) > 0 {fivehundreds ^= 0x01F;} // A2
  if (gAlt & 0x4000) > 0 {fivehundreds ^= 0x00F;} // A4

  if (gAlt & 0x0100) > 0 {fivehundreds ^= 0x007;} // B1
  if (gAlt & 0x0200) > 0 {fivehundreds ^= 0x003;} // B2
  if (gAlt & 0x0400) > 0 {fivehundreds ^= 0x001;} // B4

  // Correct order of onehundreds.
  if (fivehundreds & 1) > 0 {onehundreds = 6 - onehundreds;}

  return ((int(fivehundreds) * 5) + int(onehundreds) - 13);
}

func decodeExtendedSquitter(message []byte, linkFmt uint) {
  const ais_charset = "@ABCDEFGHIJKLMNOPQRSTUVWXYZ[\\]^_ !\"#$%&'()*+,-./0123456789:;<=>?"


  if linkFmt == 18 {
    switch (message[0] & 7) {
    case 1:
      fmt.Println("Non-ICAO")
    case 2:
      fmt.Println("TIS-B fine")
    case 3:
      fmt.Println("TIS-B coarse")
    case 5:
      fmt.Println("TIS-B anon ADS-B relay")
    case 6:
      fmt.Println("ADS-B rebroadcast")
    default:
      fmt.Println("Non-ICAO unknown")
    }
  }

  msgType := uint(message[4]) >> 3
  var msgSubType uint
  if msgType == 29 {
    msgSubType = (uint(message[4]) & 6) >> 1
  } else {
    msgSubType = uint(message[4]) & 7
  }

  switch msgType {
  case 1,2,3,4:
    // Aircraft ID
    chars1 := uint(message[5]) << 16 + uint(message[6]) << 8 + uint(message[7])
    chars2 := uint(message[8]) << 16 + uint(message[9]) << 8 + uint(message[10])

    var fltByte [8]byte

    if chars1 != 0 && chars2 != 0 {
      fltByte[3] = ais_charset[chars1 & 0x3F]; chars1 >>= 6
      fltByte[2] = ais_charset[chars1 & 0x3F]; chars1 >>= 6
      fltByte[1] = ais_charset[chars1 & 0x3F]; chars1 >>= 6
      fltByte[0] = ais_charset[chars1 & 0x3F]
      fltByte[7] = ais_charset[chars2 & 0x3F]; chars2 >>= 6
      fltByte[6] = ais_charset[chars2 & 0x3F]; chars2 >>= 6
      fltByte[5] = ais_charset[chars2 & 0x3F]; chars2 >>= 6
      fltByte[4] = ais_charset[chars2 & 0x3F]

      callsign := string(fltByte[:8])
      fmt.Println("Callsign: ", callsign)
    }
  }

  switch msgSubType {
  case 1:
    break
  }




}
