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

import "net"
import "fmt"
import "bufio"

func main() {

  fmt.Println("Launching server...")

  // listen on all interfaces
  ln, _ := net.Listen("tcp", ":8081")

  // accept connection on port
  conn, _ := ln.Accept()

  // run loop forever (or until ctrl-c)
  for {
    // will listen for messages, delimited by <esc> (0x1A)
    message, _ := bufio.NewReader(conn).ReadBytes('\n')

    if len(message) == 0 {
      break
    }

    // output message received
    fmt.Print("Msg: ")
    for i:= 0; i < len(message); i++ {
      fmt.Printf("%x", message[i])
    }
    fmt.Print("\n")

    if message[0] != 0x1a {
      fmt.Print("err")
      break
    }

    if message[1] == 0x31 {
      fmt.Print("Type 1 Mode-AC")
    } else if message[1] == 0x32 {
      fmt.Print("Type 2 Mode-S short")
    } else if message[1] == 0x33 {
      fmt.Print("Type 3 Mode-S long")
    } else if message[1] == 0x34 {
      fmt.Print("Status Signal")
    }



    fmt.Print("\n\n")

  }


}
