package main

import (
	"fmt"
	"os"
	"runtime/pprof"
	"time"

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
			Usage: "format to dump the data in (raw, struct)",
		},
	}
	app.Commands = []cli.Command{
		{
			Name:    "dump",
			Aliases: []string{"d"},
			Usage:   "data dump commands",
			Subcommands: []cli.Command{
				{
					Name:  "header",
					Usage: "dump data header",
					Flags: dumpFlags,
					Action: func(c *cli.Context) {
						conn, err := irsdk.NewConnection()
						if err != nil {
							fmt.Fprintln(os.Stdout, err)
							return
						}

						err = conn.Connect()
						if err != nil {
							fmt.Fprintln(os.Stdout, err)
							return
						}

						format := c.String("format")
						switch format {
						case "raw", "struct":
							header, err := conn.GetHeader()
							if err != nil {
								fmt.Fprintln(app.Writer, err)
								return
							}
							fmt.Printf("%+v\n", header)
						default:
							err := fmt.Sprintf("Unknow format: %v", format)
							fmt.Fprintln(os.Stdout, err)
							return
						}
					},
				},
				{
					Name:  "session",
					Usage: "dump session data",
					Flags: dumpFlags,
					Action: func(c *cli.Context) {
						conn, err := irsdk.NewConnection()
						if err != nil {
							fmt.Fprintln(os.Stdout, err)
							return
						}

						err = conn.Connect()
						if err != nil {
							fmt.Fprintln(os.Stdout, err)
							return
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
							return
						}
					},
				},
				{
					Name:  "telemetry",
					Usage: "dump telemetry data",
					Flags: dumpFlags,
					Action: func(c *cli.Context) {
						filename := "telemetry-test.ibt"
						f, err := os.Open(filename)
						if err != nil {
							fmt.Fprintln(os.Stdout, err)
							return
						}

						tr := irsdk.NewTelemetryReader(f)

						// @TODO: profile GetAllDataPoints()
						datapoints, err := tr.GetAllDataPoints()
						if err != nil {
							fmt.Fprintln(os.Stderr, err)
						}

						for _, dp := range datapoints {
							fmt.Println(dp.OnPitRoad)
						}

						last := len(datapoints) - 1
						fmt.Printf("%+v\n", datapoints[last])

						return
					},
				},
				{
					Name:  "memorymap",
					Usage: "dump memorymap",
					Flags: dumpFlags,
					Action: func(c *cli.Context) {
						err := "Not yet implemented"
						fmt.Fprintln(os.Stderr, err)
						return
					},
				},
			},
		},

		{
			// https://blog.golang.org/profiling-go-programs
			Name:    "profile",
			Aliases: []string{"p"},
			Usage:   "dump profiling data",
			Subcommands: []cli.Command{
				{
					Name:  "cpu",
					Usage: "dump cpu profiling data",
					Flags: dumpFlags,
					Action: func(c *cli.Context) {
						// Start profiling
						f, err := os.Create("cpu.prof")
						if err != nil {
							fmt.Fprintln(os.Stdout, err)
							return
						}
						pprof.StartCPUProfile(f)
						defer pprof.StopCPUProfile()

						conn, err := irsdk.NewConnection()
						if err != nil {
							fmt.Println(os.Stderr, err)
							return
						}

						fps := 60
						duration := 60
						loops := fps * duration
						// conn.SetMaxFPS(fps)

						start := time.Now()
						for i := 0; i < loops; i++ {
							_, err := conn.GetTelemetryData()
							if err != nil {
								fmt.Println(os.Stderr, err)
								// Don't quit, just keep on going
							}
						}
						end := time.Now()

						realDuration := (end.Sub(start))
						realFPS := float32(loops) / float32(realDuration.Seconds())
						fmt.Println(realFPS)
					},
				},
			},
		},
	}

	app.Run(os.Args)
}
