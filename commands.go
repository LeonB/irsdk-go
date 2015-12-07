package irsdk

import "C"
import (
	"fmt"
)

func testBroadcastMsg() {
	sdk := Irsdk{}
	err := sdk.BroadcastMsg(
		BroadcastChatComand,
		uint16(ChatCommand_BeginChat),
		0,
		0,
	)
	if err != nil {
		fmt.Println(err)
	}

	chatMacro := 1
	fmt.Printf("Sending chat macro %d\n", chatMacro)
	err = sdk.BroadcastMsg(
		BroadcastChatComand,
		uint16(ChatCommand_Macro),
		uint16(chatMacro),
		0,
	)
	if err != nil {
		fmt.Println(err)
	}
	return
}
