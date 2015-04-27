// +build windows

package main

import (
	"net/rpc"

	"github.com/kevinwallace/coprocess"
	"github.com/leonb/irsdk-go/utils"
)

func main() {
	s := rpc.NewServer()
	s.RegisterName("Commands", &utils.RpcCommands{})
	coprocess.Serve(s)
}
