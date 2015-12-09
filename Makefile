all: build

WINE=/opt/iracing/bin/wine --bottle default
WINECC=i686-w64-mingw32-gcc
GO=go
GOWINE=CGO_ENABLED=0 GOOS=windows GOARCH=386 $(GO)

build: irsdk

irsdk: *.go bin/irsdk/*.go bindata.go
	$(GO) build ./bin/irsdk
	
irsdk.exe: c_wrapper_windows.go
	CC=$(WINECC) $(GOWINE) build ./bin/irsdk

assets/ir-syscalls-rpc.exe: bin/ir-syscalls-rpc/main.go
	$(GOWINE) get github.com/kevinwallace/coprocess
	CC=$(WINECC) $(GOWINE) build -o $@ ./bin/ir-syscalls-rpc

terminalhud: assets/ir-syscalls-rpc.exe bin/terminalhud/main.go
	$(GO) build ./bin/terminalhud

run: build
	./irsdk $*

prof: irsdk
	./irsdk profile cpu
	$(GO) tool pprof irsdk cpu.prof

wineprof: irsdk.exe
	$(WINE) irsdk profile cpu
	$(GO) tool pprof irsdk cpu.prof

bindata.go: assets/ir-syscalls-rpc.exe
	go-bindata -pkg=irsdk -o=bindata.go assets/

clean:
	$(GOWINE) clean
	$(GO) clean
	rm -f assets/ir-syscalls-rpc.exe
	rm -f irsdk
	rm -f irsdk.exe
	rm -f terminalhud
	rm -f bindata.go

# vim: syntax=make ts=4 sw=4 sts=4 sr noet
