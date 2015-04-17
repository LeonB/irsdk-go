package irsdk

import "C"
import (
	"fmt"

	utils "github.com/leonb/irsdk-go/utils"
)

func testBroadcastMsg() {
	err := utils.Irsdk_broadcastMsg(
		utils.Irsdk_BroadcastChatComand,
		uint16(utils.Irsdk_ChatCommand_BeginChat),
		0,
		0,
	)
	if err != nil {
		fmt.Println(err)
	}

	chatMacro := 1
	fmt.Printf("Sending chat macro %d\n", chatMacro)
	err = utils.Irsdk_broadcastMsg(
		utils.Irsdk_BroadcastChatComand,
		uint16(utils.Irsdk_ChatCommand_Macro),
		uint16(chatMacro),
		0,
	)
	if err != nil {
		fmt.Println(err)
	}
	return
}
