package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
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
							b, err := conn.GetSessionDataBytes()
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
						filename := "disktelemetry.ibt"
						f, err := os.Open(filename)
						if err != nil {
							fmt.Fprintln(os.Stdout, err)
							return
						}

						// Start profiling
						pf, err := os.Create("cpu.prof")
						if err != nil {
							fmt.Fprintln(os.Stdout, err)
							return
						}

						pprof.StartCPUProfile(pf)
						defer pprof.StopCPUProfile()

						data, err := ioutil.ReadAll(f)
						if err != nil {
							fmt.Fprintln(os.Stdout, err)
							return
						}

						tr := irsdk.NewTelemetryReader(data)

						// @TODO: profile GetAllDataPoints()
						datapoints, err := tr.GetAllDataPoints()
						if err != nil {
							fmt.Fprintln(os.Stderr, err)
						}

						// first := 0
						// fmt.Printf("%+v\n", datapoints[first])

						last := len(datapoints) - 1
						fmt.Printf("%+v\n", datapoints[last].RPM)
						fmt.Printf("%+v\n", datapoints[last].IsOnTrack)
						fmt.Printf("%+v\n", datapoints[last].CarIdxEstTime)

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

						// conn, err := irsdk.NewConnection()
						// if err != nil {
						// 	fmt.Println(os.Stderr, err)
						// 	return
						// }
					},
				},
				{
					Name:  "varheaders",
					Usage: "dump varheaders from disk telemetry",
					Flags: dumpFlags,
					Action: func(c *cli.Context) {

						type jsonVar struct {
							Type       string
							Name       string
							Desc       string
							Unit       string
							MemMapData bool
							DiskData   bool
						}

						// varNames := make(map[string]*struct{}, 0)
						jsonVars := make(map[string]*jsonVar, 0)

						// Get connection varHeaders
						filename := "memmap.bin"
						f, err := os.Open(filename)
						if err != nil {
							fmt.Fprintln(os.Stdout, err)
							return
						}

						data, err := ioutil.ReadAll(f)
						if err != nil {
							fmt.Fprintln(os.Stdout, err)
							return
						}

						tr := irsdk.NewTelemetryReader(data)

						connectionVarHeaders, err := tr.GetVarHeaders()
						if err != nil {
							fmt.Fprintln(os.Stderr, err)
						}

						for _, varHeader := range connectionVarHeaders {
							if _, ok := jsonVars[varHeader.Name]; ok {
								jsonVars[varHeader.Name].MemMapData = true
								continue
							}

							jsonVar := &jsonVar{
								Type:       varHeader.Type.String(),
								Name:       varHeader.Name,
								Desc:       varHeader.Desc,
								Unit:       varHeader.Unit,
								MemMapData: true,
								DiskData:   false,
							}
							jsonVars[varHeader.Name] = jsonVar
						}

						// Get disk varHeaders
						filename = "disktelemetry.ibt"
						f, err = os.Open(filename)
						if err != nil {
							fmt.Fprintln(os.Stdout, err)
							return
						}

						data, err = ioutil.ReadAll(f)
						if err != nil {
							fmt.Fprintln(os.Stdout, err)
							return
						}

						tr = irsdk.NewTelemetryReader(data)

						diskVarHeaders, err := tr.GetVarHeaders()
						if err != nil {
							fmt.Fprintln(os.Stderr, err)
						}

						for _, varHeader := range diskVarHeaders {
							if _, ok := jsonVars[varHeader.Name]; ok {
								jsonVars[varHeader.Name].DiskData = true
								continue
							}

							jsonVar := &jsonVar{
								Type:       varHeader.Type.String(),
								Name:       varHeader.Name,
								Desc:       varHeader.Desc,
								Unit:       varHeader.Unit,
								MemMapData: false,
								DiskData:   true,
							}
							jsonVars[varHeader.Name] = jsonVar
						}

						dumpData := make([]*jsonVar, 0)
						for _, jsonVar := range jsonVars {
							dumpData = append(dumpData, jsonVar)
						}

						format := c.String("format")
						switch format {
						case "raw", "struct":
							fmt.Printf("%+v\n", dumpData)
						case "json":
							jsonData, err := json.MarshalIndent(dumpData, "", "  ") // convert to JSON
							if err != nil {
								fmt.Fprintln(os.Stderr, err)
							}
							fmt.Println(string(jsonData))
						default:
							err := fmt.Sprintf("Unknow format: %v", format)
							fmt.Fprintln(os.Stdout, err)
							return
						}

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
						duration := 10
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
