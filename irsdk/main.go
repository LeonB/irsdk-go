package main

// - irsdk dump session
// - irsdk dump telemetry

import (
	"encoding/json"
	"fmt"

	utils "github.com/leonb/irsdk-go/utils"
	"github.com/leonb/irsdk-go"
)

func main() {
	// testBroadcastMsg()
	testTelemetryData()
}

func testTelemetryData() {
	var data []byte
	var err error

	// oldTime := time.Now().Unix()
	changes := 0
	for {
		// newTime := time.Now().Unix()
		// fmt.Println(newTime)

		// if oldTime != newTime {
		// 	oldTime = newTime
		// 	changes = 0
		// 	fmt.Println("number of changes:", changes)
		// }

		// 1% cpu usage
		data, err = utils.Irsdk_waitForDataReady(3000)
		if err != nil {
			fmt.Println(err)
		}

		// fmt.Println(string(utils.Irsdk_getSessionInfoStr()[:]))
		// return

		if data != nil {
			changes++
			// fields := []string{}
			// telemetryData := toTelemetryDataFiltered(data, fields)
			telemetryData := irsdk.ToTelemetryData(data)
			b, err := json.Marshal(telemetryData)
			if err != nil {
				fmt.Printf("Error: %s", err)
				return
			}
			fmt.Println(string(b))
			fmt.Println(changes)
		}

		// utils.Irsdk_shutdown()
		// break
	}

	return
}



