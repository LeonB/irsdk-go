# irsdk in Go

This is a port of the iRacing sdk to Go. The iRacing sdk logic is ported and
slightly modified to make it a bit more go-y.

The initial idea for porting it to go was to make it compatible with the Linux
version of iRacing (which runs under wine / CrossOver). I choose Go because of
the (small) cross-platform standalone binaries without dependencies on a
third-party frameworks (.Net).

This doesn't mean it's Linux only. Go runs fine on Windows and this library
should work fine on Windows (not tested).

## Installation

### Windows

```
go get github.com/LeonB/irsdk-go
```

### Linux

```
GOARCH=386 GOOS=windows go get github.com/LeonB/irsdk-go
```

## Known bugs / pitfalls

- When running it under wine, Go time functions do not work:
  [go-nuts thread](https://groups.google.com/forum/#!topic/golang-nuts/nhJOw71rw7k) /
  [wine bug](https://bugs.winehq.org/show_bug.cgi?id=38272)
