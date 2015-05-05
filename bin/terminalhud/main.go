package main

import (
	"fmt"
	"log"
	"strings"
	"time"

	irsdk "github.com/leonb/irsdk-go"
)

// speed: 120 km/h
// rpm:   8421
// gear:  4
// clutch   [|||||||             ]
// brake    [                    ]
// throttle [|||||||||||||||||||‖]

// Lights up red when you need to shift?

type renderer interface {
	String() string
}

type textRenderer struct {
	label       string
	labelLength int
	data        string
	bgColor     int
}

func (r textRenderer) String() string {
	labelPadding := ""
	if r.labelLength > 0 {
		labelPadding = strings.Repeat(" ", r.labelLength-len(r.label))
	}

	data := ""
	if r.bgColor > 0 {
		data = data + fmt.Sprintf("\033[48;5;%vm", r.bgColor)
	}
	data = data + r.data
	if r.bgColor > 0 {
		data = data + "\033[0m"
	}

	return fmt.Sprintf("%s:%v %v", r.label, labelPadding, data)
}

type graphRenderer struct {
	label       string
	labelLength int
	data        float32
	textColor   int
}

func (r graphRenderer) String() string {
	maxBars := 20
	bars := int(r.data * float32(maxBars))

	graph := "["
	if r.textColor > 0 {
		graph = graph + fmt.Sprintf("\033[38;5;%vm", r.textColor)
	}
	graph = graph + strings.Repeat("|", bars)
	graph = graph + strings.Repeat(" ", maxBars-bars)
	if r.textColor > 0 {
		graph = graph + "\033[0m"
	}
	graph = graph + "]"

	labelPadding := ""
	if r.labelLength > 0 {
		labelPadding = strings.Repeat(" ", r.labelLength-len(r.label))
	}
	return fmt.Sprintf("%s:%v %v", r.label, labelPadding, graph)
}

func main() {
	var conn *irsdk.Connection
	var session *irsdk.SessionData
	var err error

	// Keep trying until connection is made
	conn, _ = irsdk.NewConnection()

	for {
		if conn.IsConnected() == false {
			time.Sleep(time.Second * 1)
			conn.Connect()
			session, err = conn.GetSessionData()
			if err != nil {
				log.Fatal(err)
			}
			continue
		}

		telemetryData, err := conn.GetTelemetryData()
		if err != nil {
			continue
		}

		tSpeed := float32(0.0)
		tRPM := float32(0.0)
		tGear := 0
		tClutch := float32(1.0)
		tBrake := float32(0.0)
		tThrottle := float32(0.0)

		if telemetryData != nil {
			tSpeed = telemetryData.Speed
			tRPM = telemetryData.RPM
			tGear = telemetryData.Gear
			tClutch = telemetryData.Clutch
			tBrake = telemetryData.Brake
			tThrottle = telemetryData.Throttle
		}

		speed := textRenderer{
			label:       "speed",
			labelLength: 5,
			data:        fmt.Sprintf("%.0f km/h", (tSpeed * 3.6)),
		}

		var bgColor int
		// DriverCarSLFirstRPM = 6000
		// DriverCarSLShiftRPM = 6850
		// DriverCarSLBlinkRPM = 7000
		sl4 := session.DriverInfo.DriverCarSLBlinkRPM
		sl3 := session.DriverInfo.DriverCarSLShiftRPM
		sl1 := session.DriverInfo.DriverCarSLFirstRPM
		sl2 := sl1 + ((sl4 - sl1) / 2)
		if tRPM >= sl4 {
			bgColor = 196 // red
		} else if tRPM >= sl3 {
			bgColor = 202 // orange
		} else if tRPM >= sl2 {
			bgColor = 226 // yellow
		} else if tRPM >= sl1 {
			bgColor = 41 // green
		} else {
			bgColor = 0
		}

		rpm := textRenderer{
			label:       "rpm",
			labelLength: 5,
			data:        fmt.Sprintf("%.0f", tRPM),
			bgColor:     bgColor,
		}

		gear := textRenderer{
			label:       "gear",
			labelLength: 5,
		}

		if tGear > 0 {
			gear.data = fmt.Sprintf("%v", tGear)
		} else if tGear < 0 {
			gear.data = "R"
		} else {
			gear.data = "N"
		}

		clutch := graphRenderer{
			label:       "clutch",
			labelLength: 8,
			data:        1.0 - tClutch,
			textColor:   45,
		}

		brake := graphRenderer{
			label:       "brake",
			labelLength: 8,
			data:        tBrake,
			textColor:   196,
		}

		throttle := graphRenderer{
			label:       "throttle",
			labelLength: 8,
			data:        tThrottle,
			textColor:   41,
		}

		fmt.Printf("\033c") // clear entire screen
		fmt.Println(speed)
		fmt.Println(rpm)
		fmt.Println(gear)
		fmt.Println(clutch)
		fmt.Println(brake)
		fmt.Println(throttle)

		fmt.Println(conn.IsConnected())
		fmt.Println(time.Now())
	}
}
