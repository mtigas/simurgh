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
   aircraft output, not unlike dump1090’s "interactive mode". Something like:

   ```
   ICAO    Callsign    Location                Alt     Distance
   a64d4d              40.631104,-73.874329    2450    5.36
   a7ad0d  JBU839      40.519043,-73.865479    7050    11.89
   a03765  AAL125      40.630646,-73.726013    9450    12.54
   0c6030  BWA17       40.460999,-73.608704    4875    23.59
   aa4252              40.957535,-74.260986?   21425   25.09?    27…
   ae0449  BALSA76     40.287689,-73.836365    35000   27.59
   ac7b1b  JBU1503     40.431381,-73.503662    13650   29.22
   a099a7              40.865204,-73.195496?   18800   41.96?    22…
   a9f2c6              40.313828,-73.042053    22750   54.23
   acc6d1  AAL1313     40.804138,-72.314026    30850   86.49
   ```

   Output is streamed constantly (TODO: better printing to screen without
   filling console buffers). Aircraft with location data older than 10 seconds
   are marked with a `?`, and a timer eventually appears. Old aircraft (45sec)
   are discarded from the on-screen list.

## Further Reading

* [dump1090](https://github.com/mutability/dump1090) is one of several applications that accepts BEAST input (port `30004` or `31004`) and generates BEAST output (`30005`).
* [Information about the data format](http://wiki.modesbeast.com/Mode-S_Beast:Data_Output_Formats) (see "Binary Format")
  * dump1090's [mode s decoder](https://github.com/mutability/dump1090/blob/master/mode_s.c)
  * Wikipedia: [Aviation transponder interrogation modes](https://en.wikipedia.org/wiki/Aviation_transponder_interrogation_modes)
  * Wikipedia: [Secondary surveillance radar#Mode S](https://en.wikipedia.org/wiki/Secondary_surveillance_radar#Mode_S)
