# Distributed pprof.

A proxy for using the Go pprof tool against multiple servers. Much WIP.

## Install

```bash
go get github.com/aybabtme/dpprof/cmd/dpprofproxy
```

## Usage

Start the proxy:

```bash
dpprofproxy -h http://service.one -h http://service.two -h http://service.three
```

Then invoke pprof commands against the proxy:

```bash
go tool pprof http://127.0.0.1:8000/debug/pprof/profile
```


After a few seconds, you should get:

```
Fetching profile from http://127.0.0.1:8000/debug/pprof/profile
Please wait... (30s)
Entering interactive mode (type "help" for commands)
(pprof) top
2080ms of 2100ms total (99.05%)
Dropped 8 nodes (cum <= 10.50ms)
Showing top 10 nodes out of 44 (cum >= 160ms)
      flat  flat%   sum%        cum   cum%
     870ms 41.43% 41.43%      870ms 41.43%  syscall.Unlink
     370ms 17.62% 59.05%      370ms 17.62%  runtime.mach_semaphore_signal
     260ms 12.38% 71.43%      260ms 12.38%  syscall.Syscall
     180ms  8.57% 80.00%      180ms  8.57%  runtime.memmove
     140ms  6.67% 86.67%      140ms  6.67%  runtime.usleep
     120ms  5.71% 92.38%      120ms  5.71%  runtime.memclr
     110ms  5.24% 97.62%      360ms 17.14%  compress/flate.(*compressor).deflate
      10ms  0.48% 98.10%      160ms  7.62%  compress/flate.(*compressor).fillDeflate
      10ms  0.48% 98.57%       80ms  3.81%  compress/flate.(*compressor).writeBlock
      10ms  0.48% 99.05%      160ms  7.62%  compress/flate.(*huffmanEncoder).assignEncodingAndSize
```

## Testing

A stub service that uses some CPU resources is kept in `cmd/stubservice`.
Start a couple of them on different ports, then start the proxy pointing to them.
You should be able to play with `go tool pprof` and fetch some profiles.


## TODO

- [ ]: singleflight, can't have two cpu pprof requests at the same time
- [ ]: make the threadcreate and block profiles work
