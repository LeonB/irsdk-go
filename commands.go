package irsdk

import "C"
import (
	"fmt"

	utils "github.com/leonb/irsdk-go/utils"
)

func testBroadcastMsg() {
	sdk := utils.Irsdk{}
	err := sdk.BroadcastMsg(
		utils.BroadcastChatComand,
		uint16(utils.ChatCommand_BeginChat),
		0,
		0,
	)
	if err != nil {
		fmt.Println(err)
	}

	chatMacro := 1
	fmt.Printf("Sending chat macro %d\n", chatMacro)
	err = sdk.BroadcastMsg(
		utils.BroadcastChatComand,
		uint16(utils.ChatCommand_Macro),
		uint16(chatMacro),
		0,
	)
	if err != nil {
		fmt.Println(err)
	}
	return
}
