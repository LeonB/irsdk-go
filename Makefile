all: build

WINE=/opt/iracing/bin/wine --bottle default
WINECC=i686-w64-mingw32-gcc
GO=go
GOWINE=CGO_ENABLED=1 GOOS=windows GOARCH=386 $(GO)

build: ir-syscalls-rpc.exe irsdk

irsdk: ir-syscalls-rpc.exe
	$(GO) build ./bin/irsdk
	
irsdk.exe: utils/c_wrapper_windows.go
	CC=$(WINECC) $(GOWINE) build ./bin/irsdk

ir-syscalls-rpc.exe: bin/ir-syscalls-rpc/main.go
	$(GOWINE) get github.com/kevinwallace/coprocess
	CC=$(WINECC) $(GOWINE) build ./bin/ir-syscalls-rpc

terminalhud: ir-syscalls-rpc.exe bin/terminalhud/main.go
	$(GO) build ./bin/terminalhud

run: build
	./irsdk $*

prof: irsdk
	./irsdk profile cpu
	$(GO) tool pprof irsdk cpu.prof

wineprof: irsdk.exe
	$(WINE) irsdk profile cpu
	$(GO) tool pprof irsdk cpu.prof

clean:
	$(GOWINE) clean
	$(GO) clean
	rm -f ir-syscalls-rpc.exe
	rm -f irsdk
	rm -f irsdk.exe
	rm -f terminalhud
