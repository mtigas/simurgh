# Simurgh

A Mode-S "BEAST" TCP decoder server.

© 2016 Mike Tigas. Licensed under the [GNU Affero General Public License](LICENSE).

---

Mostly an experiment to get a better handle on the raw signal format (and also
get better at Go).

This is somewhat run-able by

1. Having a `dump1090` server running
2. Running `go run server.go` to run this
3. `nc 127.0.0.1 30005 | nc 127.0.0.1 8081` to pipe data from the `dump1090`
   BEAST output port into this

(but this thing is pretty darn incomplete still)

## Notes

* [dump1090](https://github.com/mutability/dump1090) is one of several applications that accepts BEAST input (port `30004` or `31004`) and generates BEAST output (`30005`).
* [Information about the data format](http://wiki.modesbeast.com/Mode-S_Beast:Data_Output_Formats) (see "Binary Format")
  * dump1090's [mode s decoder](https://github.com/mutability/dump1090/blob/master/mode_s.c)
  * Wikipedia: [Aviation transponder interrogation modes](https://en.wikipedia.org/wiki/Aviation_transponder_interrogation_modes)
  * Wikipedia: [Secondary surveillance radar#Mode S](https://en.wikipedia.org/wiki/Secondary_surveillance_radar#Mode_S)
