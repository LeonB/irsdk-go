// -build windows

package irsdk

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/rpc"
	"os"
	"os/exec"
	"time"
	"unsafe"

	mmap "github.com/edsrzf/mmap-go"
	"github.com/kevinwallace/coprocess"
	wineshm "github.com/leonb/wineshm-go"
)

var (
	WineCmd = []string{"/opt/iracing/bin/wine", "--bottle", "default"}
)

const (
	RPC_COMMAND      = "assets/ir-syscalls-rpc.exe"
	DATA_CHANGE_TICK = time.Duration(time.Millisecond * 9)
)

type CWrapper struct {
	sharedMem       []byte
	sharedMemPtr    unsafe.Pointer
	header          *Header
	hDataValidEvent uintptr

	mmapFile *os.File
	client   *rpc.Client

	telemetryReader *BytesReader
}

func (cw *CWrapper) startup() error {
	var err error

	if cw.client == nil {
		cw.client, err = newRpcClient()
		if err != nil {
			return err
		}
	}

	if cw.mmapFile == nil {
		cw.mmapFile, err = cw.getMmapFile()
		if err != nil {
			return err
		}
	}

	if len(cw.sharedMem) == 0 {
		cw.sharedMem, err = cw.getSharedMem()
		if err != nil {
			return err
		}

		cw.telemetryReader = &BytesReader{
			data: cw.sharedMem,
		}
	}

	if cw.sharedMemPtr == nil {
		cw.sharedMemPtr, err = cw.getSharedMemPtr()
		if err != nil {
			return err
		}
	}

	if cw.header == nil {
		cw.header, err = cw.getHeader()
		fmt.Printf("%+v\n", cw.header)
		if err != nil {
			return err
		}
	}

	if cw.hDataValidEvent == 0 {
		cw.hDataValidEvent, err = cw.OpenEvent(DATAVALIDEVENTNAME)
		if err != nil {
			return err
		}
	}

	return nil

}

func (cw *CWrapper) shutdown() error {
	if cw.client != nil {
		cw.client.Close()
	}

	if cw.mmapFile != nil {
		cw.mmapFile.Close()
	}

	// Clean linux specific vars
	cw.client = nil
	cw.mmapFile = nil

	// Clean global vars
	cw.sharedMemPtr = nil
	cw.sharedMem = nil
	cw.header = nil

	return nil
}

func (cw *CWrapper) getMmapFile() (*os.File, error) {
	var err error

	wineshm.WineCmd = WineCmd
	shmFd, err := wineshm.GetWineShm(MEMMAPFILENAME, wineshm.FILE_MAP_READ)
	if err != nil {
		return nil, err
	}

	file := os.NewFile(shmFd, MEMMAPFILENAME)
	return file, err
}

func (cw *CWrapper) getSharedMem() ([]byte, error) {
	sharedMem, err := mmap.Map(cw.mmapFile, mmap.RDONLY, 0)
	if err != nil {
		return nil, err
	}

	return sharedMem, nil
}

func (cw *CWrapper) getSharedMemPtr() (unsafe.Pointer, error) {
	sharedMem, err := cw.getSharedMem()
	if err != nil {
		return nil, err
	}

	return unsafe.Pointer(&sharedMem[0]), nil
}

func (cw *CWrapper) WaitForDataChange(timeout time.Duration) error {
	return cw.WaitForSingleObject(cw.hDataValidEvent, int(timeout/time.Millisecond))
	// or use cw.WaitForDataChangeChannel()?
}

func (cw *CWrapper) WaitForDataChangeChannel(timeout time.Duration) error {
	latest := cw.header.GetLatestVarBufN()
	prevTickCount := cw.header.VarBuf[latest].TickCount

	// Create a ticker and a stop channel
	ticker := time.NewTicker(DATA_CHANGE_TICK)
	defer ticker.Stop()
	stop := make(chan bool)

	// Check every tick if iRacing's tick has changed: when iRacing's tick has
	// changed send a message on the stop channel to make the for loop stop
	// @TODO: this is probably leaking goroutines
	go func() {
		for {
			select {
			case <-ticker.C:
				// Check iRacing tick count
				curTickCount := cw.header.VarBuf[latest].TickCount
				if prevTickCount != curTickCount {
					// tickcount changed: stop it
					stop <- true
				}
			case <-stop:
				ticker.Stop()
			}
		}
	}()

	// After timeout send a message on stop channel
	time.AfterFunc(timeout, func() {
		stop <- true
	})

	// Wait for stopchannel to receive a message
	<-stop
	return nil
}

// Syscalls

func (cw *CWrapper) CloseHandle(handle uintptr) error {
	return nil
}

func (cw *CWrapper) UnmapViewOfFile(lpBaseAddress uintptr) error {
	return nil
}

func (cw *CWrapper) OpenEvent(lpName string) (uintptr, error) {
	var handle uintptr
	args := &OpenEventArgs{lpName}

	err := cw.client.Call("Commands.OpenEvent", args, &handle)
	return handle, err
}

func (cw *CWrapper) WaitForSingleObject(hDataValidEvent uintptr, timeOut int) error {
	retVal := new(int)
	args := &WaitForSingleObjectArgs{hDataValidEvent, timeOut}

	return cw.client.Call("Commands.WaitForSingleObject", args, &retVal)
}

func (cw *CWrapper) RegisterWindowMessageA(lpString string) (uint, error) {
	return 0, nil
}

func (cw *CWrapper) RegisterWindowMessageW(lpString string) (uint, error) {
	return 0, nil
}

func (cw *CWrapper) SendNotifyMessage(msgID uint, wParam uint32, lParam uint32) error {
	return nil
}

func (cw *CWrapper) SendNotifyMessageW(msgID uint, wParam uint32, lParam uint32) error {
	return nil
}

func NewCWrapper() (*CWrapper, error) {
	client, err := newRpcClient()
	if err != nil {
		return nil, err
	}
	return &CWrapper{client: client}, nil
}

func newRpcClient() (*rpc.Client, error) {
	rpcCommand, err := getRpcCommand()
	// Remove the tmp file: if wine is running the rpc wrapper it's in memory
	// and not needed anymore on disk
	defer os.Remove(rpcCommand.Name())
	if err != nil {
		return nil, err
	}

	cmd := exec.Command(WineCmd[0], append(WineCmd[1:], rpcCommand.Name())...)
	client, err := coprocess.NewClient(cmd)
	if err != nil {
		return nil, err
	}

	args := new(string)
	*args = "ping"
	ret := new(bool)
	err = client.Call("Commands.Ping", args, ret)
	if err != nil {
		msg := fmt.Sprintf("Failed to execute rpc client (%v), make sure iRacing and ir-syscalls-rpc.exe are installed", err)
		err = errors.New(msg)
		return nil, err
	}

	return client, nil
}

func getRpcCommand() (*os.File, error) {
	f, err := ioutil.TempFile("", "irsdk-go")
	defer f.Close()
	if err != nil {
		return nil, err
	}

	data, err := Asset(RPC_COMMAND)
	if err != nil {
		return nil, err
	}

	_, err = f.Write(data)
	if err != nil {
		return nil, err
	}

	return f, nil
}
