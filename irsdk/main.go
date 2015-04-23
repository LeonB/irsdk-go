package main

// - irsdk dump session
// - irsdk dump telemetry
// - irsdk dump sessionStuct
// - irsdk dump telemetryStruct

import (
	"fmt"
	"os"

	"github.com/codegangsta/cli"
	irsdk "github.com/leonb/irsdk-go"
)

// dictionaryFlag only accepts a list of predefined flag values
type dictionaryFlag struct {
	cli.StringFlag
	Values []string
}

func main() {
	app := cli.NewApp()
	app.Name = "irsdk"
	app.Usage = "some simple commands to check if the iRacing go sdk is working"
	app.Version = "0.0.1"
	dumpFlags := []cli.Flag{
		cli.StringFlag{
			Name:  "format",
			Value: "raw",
			Usage: "format to dump the data in",
		},
	}
	app.Commands = []cli.Command{
		{
			Name:    "dump",
			Aliases: []string{"d"},
			Usage:   "data dump commands",
			Subcommands: []cli.Command{
				{
					Name:  "session",
					Usage: "dump session data",
					Flags: dumpFlags,
					Action: func(c *cli.Context) {
						conn, err := irsdk.NewConnection()
						if err != nil {
							fmt.Fprintln(os.Stdout, err)
						}

						format := c.String("format")
						switch format {
						case "raw":
							b, err := conn.GetRawSessionData()
							if err != nil {
								fmt.Fprintln(app.Writer, err)
								return
							}
							fmt.Println(string(b[:]))
						case "struct":
							session, err := conn.GetSessionData()
							if err != nil {
								fmt.Fprintln(app.Writer, err)
								return
							}
							fmt.Printf("%+v\n", session)
						default:
							err := fmt.Sprintf("Unknow format: %v", format)
							fmt.Fprintln(os.Stdout, err)
						}
					},
				},
				{
					Name:  "telemetry",
					Usage: "dump telemetry data",
					Flags: dumpFlags,
					Action: func(c *cli.Context) {
						conn, err := irsdk.NewConnection()
						if err != nil {
							fmt.Println(os.Stderr, err)
							return
						}

						format := c.String("format")
						switch format {
						case "raw":
							b, err := conn.GetRawTelemetryData()
							if err != nil {
								fmt.Fprintln(os.Stderr, err)
								return
							}
							fmt.Println(string(b[:]))
						case "struct":
							telemetryData, err := conn.GetTelemetryData()
							if err != nil {
								fmt.Fprintln(os.Stderr, err)
								return
							}
							fmt.Printf("%+v\n", telemetryData)
						default:
							err := fmt.Sprintf("Unknow format: %v", format)
							fmt.Fprintln(os.Stdout, err)
						}
					},
				},
				{
					Name:  "memorymap",
					Usage: "dump memorymap",
					Flags: dumpFlags,
					Action: func(c *cli.Context) {
						err := "Not yet implemented"
						fmt.Fprintln(os.Stderr, err)
					},
				},
			},
		},
	}

	app.Run(os.Args)
}
