all: build

build: ir-syscalls-rpc.exe irsdk

wine: irsdk.exe

irsdk:
	go build ./bin/irsdk
	
irsdk.exe: utils/c_wrapper_windows.go
	CC=i686-w64-mingw32-gcc CGO_ENABLED=1 GOOS=windows GOARCH=386 go build ./bin/irsdk

ir-syscalls-rpc.exe:
	GOOS=windows GOARCH=386 go get github.com/kevinwallace/coprocess
	CC=i686-w64-mingw32-gcc CGO_ENABLED=1 GOOS=windows GOARCH=386 go build ./bin/ir-syscalls-rpc

run: build
	/opt/iracing/bin/wine --bottle default ir-syscalls-rpc.exe

# winetest: wine
# 	/opt/iracing/bin/wine --bottle default irsdk.exe

clean:
	CGO_ENABLED=1 GOOS=windows GOARCH=386 go clean
	go clean
	rm -f ir-syscalls-rpc.exe
	rm -f irsdk
	rm -f irsdk.exe
