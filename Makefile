all: build

WINE=/opt/iracing/bin/wine --bottle default
WINECC=i686-w64-mingw32-gcc
GO=go
GOWINE=CGO_ENABLED=1 GOOS=windows GOARCH=386 $(GO)

build: irsdk

irsdk: utils/bindata.go
	$(GO) build ./bin/irsdk
	
irsdk.exe: utils/c_wrapper_windows.go
	CC=$(WINECC) $(GOWINE) build ./bin/irsdk

utils/assets/ir-syscalls-rpc.exe: bin/ir-syscalls-rpc/main.go
	$(GOWINE) get github.com/kevinwallace/coprocess
	CC=$(WINECC) $(GOWINE) build -o $@ ./bin/ir-syscalls-rpc

terminalhud: utils/assets/ir-syscalls-rpc.exe bin/terminalhud/main.go
	$(GO) build ./bin/terminalhud

run: build
	./irsdk $*

prof: irsdk
	./irsdk profile cpu
	$(GO) tool pprof irsdk cpu.prof

wineprof: irsdk.exe
	$(WINE) irsdk profile cpu
	$(GO) tool pprof irsdk cpu.prof

utils/bindata.go: utils/assets/ir-syscalls-rpc.exe
	go-bindata -pkg=utils -o=utils/bindata.go utils/assets/

clean:
	$(GOWINE) clean
	$(GO) clean
	rm -f utils/assets/ir-syscalls-rpc.exe
	rm -f irsdk
	rm -f irsdk.exe
	rm -f terminalhud
	rm -f utils/bindata.go

# vim: syntax=make ts=4 sw=4 sts=4 sr noet
