# Simurgh

A Mode-S "BEAST" TCP decoder server.

Mostly an experiment to understand the raw binary "BEAST" signal format,
and an excuse to learn Go. Don't use this for anything real.

© 2016 Mike Tigas. Licensed under the [GNU Affero General Public License](LICENSE). Some portions based on [dump1090-mutability](https://github.com/mutability/dump1090), licensed under the [GNU Public License v2](https://github.com/mutability/dump1090/blob/master/LICENSE).

## Building / Installation

You need [Go](https://golang.org/). Download & set it up [using the instructions here](https://golang.org/doc/install);
minimally, you need to install the Go tools (and the tools in your `$PATH`)
and ensure that you have a `$GOPATH` set.

Then:

```
go get github.com/mtigas/simurgh
```

Now you'll have a binary at `$GOPATH/bin/simurgh`.

Make sure `$GOPATH/bin` is on your `$PATH`, or copy `$GOPATH/bin/simurgh` to
somewhere like `/usr/local/bin/`.

## Usage

1. You need a `dump1090` server running.

2. Launch the simurgh listener by running `simurgh`. There are also some flags
   you can use:

   ```
   -baseLat float
       latitude used for distance calculation (default 40.77725)
   -baseLon float
       longitude for distance calculation (default -73.872611)
   -bind string
       ":port" or "ip:port" to bind the server to (default "127.0.0.1:8081")
   -sortMode uint
       0: sort by time, 1: sort by distance, 3: sort by air (default 1)
   ```

   i.e. `simurgh --baseLat 40.68931 --baseLon "-74.04464"` if you're
   receiving data from the Statue of Liberty(???)

3. Given that `dump1090` is running on the same machine as this program,

   ```
   nc 127.0.0.1 30005 | nc 127.0.0.1 8081
   ```

   will pipe the appropriate network data in and you should see some basic
   aircraft output, not unlike dump1090’s "interactive mode".

## Further Reading

* [dump1090](https://github.com/mutability/dump1090) is one of several applications that accepts BEAST input (port `30004` or `31004`) and generates BEAST output (`30005`).
* [Information about the data format](http://wiki.modesbeast.com/Mode-S_Beast:Data_Output_Formats) (see "Binary Format")
  * dump1090's [mode s decoder](https://github.com/mutability/dump1090/blob/master/mode_s.c)
  * Wikipedia: [Aviation transponder interrogation modes](https://en.wikipedia.org/wiki/Aviation_transponder_interrogation_modes)
  * Wikipedia: [Secondary surveillance radar#Mode S](https://en.wikipedia.org/wiki/Secondary_surveillance_radar#Mode_S)
