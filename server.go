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

import "bytes"
import "encoding/binary"
import "net"
import "fmt"
import "bufio"
import "time"

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

    parsePrevMessage := false

    switch current_message[0] {
    case 0x31, 0x32, 0x33, 0x34:
      parsePrevMessage = true
    }

    if buffered_message == nil {
      buffered_message = current_message
    } else {
      buffered_message = append(buffered_message, current_message...)
    }

    if !parsePrevMessage {
      continue
    } else {
      message := buffered_message
      buffered_message = nil

      msgType := message[0]
      var msgLen int

      switch msgType {
        case 0x31:
          fmt.Print("Type 1 Mode-AC")
          msgLen = 10 // 2 + 8 header
        case 0x32:
          fmt.Print("Type 2 Mode-S short")
          msgLen = 15 // 7 + 8 header
        case 0x33:
          fmt.Print("Type 3 Mode-S long")
          msgLen = 22 // 14
        case 0x34:
          fmt.Print("Status Signal")
          msgLen = 10 // ??
        default:
          msgLen = 8 // shortest possible msg w/header & timetstamp
      }

      // Message wasn't long enough to contain the full header (maybe
      // input stream error), so skip
      if len(message) < msgLen {
        continue
      }

      // output message received
      fmt.Print("\t")
      for i:= 0; i < len(message); i++ {
        fmt.Printf("%02x", message[i])
      }
      fmt.Print("\n")

      //if message[0] != 0x1a {
      //  fmt.Print("err")
      //  break
      //}

      //timestamp := parseTime(message[2:8])
      //fmt.Print(timestamp)

      //fmt.Print("\n\n")
    }

  }

}

func parseTime(timebytes []byte) time.Time {
  for i:= 0; i < len(timebytes); i++ {
    fmt.Printf("%02x", timebytes[i])
  }
  iTime := binary.LittleEndian.Uint16(timebytes)
  fmt.Print("\n")
  fmt.Println(iTime)


  data := []byte{0x02, 0x18, 0x5f, 0x10, 0x5f, 0x1c, 0x00, 0x00}
  buf := bytes.NewBuffer(data)
  var iTimeb int
  binary.Read(buf, binary.BigEndian, &iTimeb)
  fmt.Print("\n")
  fmt.Println(iTimeb)

  return time.Now()
}


//func (b *bufio.Reader) ReadBeastBytes(delim byte) (line [], err error) {
//	// Use ReadSlice to look for array,
//	// accumulating full buffers.
//	var frag []byte
//	var full [][]byte
//	var err error
//	for {
//		var e error
//		frag, e = b.ReadSlice(delim)
//		if e == nil { // got final fragment
//			break
//		}
//		if e != ErrBufferFull { // unexpected error
//			err = e
//			break
//		}
//
//		// Make a copy of the buffer.
//		buf := make([]byte, len(frag))
//		copy(buf, frag)
//		full = append(full, buf)
//	}
//
//	// Allocate new buffer to hold the full pieces and the fragment.
//	n := 0
//	for i := range full {
//		n += len(full[i])
//	}
//	n += len(frag)
//
//	// Copy full pieces and fragment in.
//	buf := make([]byte, n)
//	n = 0
//	for i := range full {
//		n += copy(buf[n:], full[i])
//	}
//	copy(buf[n:], frag)
//	return buf, err
//}
