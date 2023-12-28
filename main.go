package main

import (
	"crypto/tls"
	"downtime-reporter/core"
	"downtime-reporter/reader"
	"downtime-reporter/transformers"
	"downtime-reporter/writer"
	"fmt"
	"github.com/urfave/cli/v2"
	"log"
	"net/http"
	"os"
	"time"
)

type flags struct {
	prometheus string
	step       time.Duration
	unixTime   bool
	startTime  string
	endTime    string
	query      string
	concise    bool
	fileName   string
}

var flag flags

func main() {
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}

	app := cli.NewApp()
	app.Name = "Downtime Reporter"
	app.Flags = []cli.Flag{
		&cli.StringFlag{
			Name:        "prometheus",
			Value:       "http://localhost:9090",
			Destination: &flag.prometheus,
			Aliases:     []string{"p"},
		},
		&cli.DurationFlag{
			Name:        "step",
			Usage:       "in seconds/minutes",
			Destination: &flag.step,
			Aliases:     []string{"s"},
			Value:       2 * time.Minute,
		},
		&cli.BoolFlag{
			Name:        "unix",
			Usage:       "Show in Unix time",
			Destination: &flag.unixTime,
			Value:       false,
			Aliases:     []string{"u"},
		},
		&cli.StringFlag{
			Name:        "query",
			Usage:       "Prometheus query to run to get downtime. Query should return historical 0 or 1 only (0 for down and 1 for up)",
			Destination: &flag.query,
			Required:    true,
			Aliases:     []string{"q"},
		},
		&cli.StringFlag{
			Name:        "startTime",
			Usage:       "Start Time for query. Layout like 2023-12-31 23:59:59 (yyyy-MM-dd hh:mm:ss)",
			Destination: &flag.startTime,
			Aliases:     []string{"st"},
			Value:       time.Now().Add(-time.Hour).Format(core.DateFormat),
			Action: func(context *cli.Context, s string) error {
				_, err := time.ParseInLocation(core.DateFormat, flag.startTime, time.Local)
				if err != nil {
					return err
				}

				return nil
			},
		},
		&cli.StringFlag{
			Name:        "endTime",
			Usage:       "End Time for query. Layout like 2023-12-31 23:59:59 (yyyy-MM-dd hh:mm:ss)",
			Destination: &flag.endTime,
			Aliases:     []string{"et"},
			Value:       time.Now().Format(core.DateFormat),
			Action: func(context *cli.Context, s string) error {
				_, err := time.ParseInLocation(core.DateFormat, flag.endTime, time.Local)
				if err != nil {
					return err
				}
				return nil
			},
		},
		&cli.BoolFlag{
			Name:        "concise",
			Usage:       "Concise output. Only show downtime period",
			Destination: &flag.concise,
			Value:       false,
			Aliases:     []string{"c"},
		},
		&cli.StringFlag{
			Name:        "fileName",
			Usage:       "File name to write output to",
			Destination: &flag.fileName,
			Value:       "output.csv",
			Aliases:     []string{"f"},
		},
	}
	app.Action = mainAction

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

func mainAction(c *cli.Context) error {
	st, _ := time.ParseInLocation(core.DateFormat, flag.startTime, time.Local)
	et, _ := time.ParseInLocation(core.DateFormat, flag.endTime, time.Local)
	promReader, err := reader.NewPrometheusV2Reader(flag.prometheus, reader.PrometheusQueryParams{
		Query: flag.query,
		Start: &st,
		End:   &et,
		Step:  &flag.step,
	})
	if err != nil {
		return fmt.Errorf("failure initializing prometheus reader: %v", err)
	}

	result, err := promReader.Read()
	if err != nil {
		return fmt.Errorf("failure reading from prometheus: %v", err)
	}

	if !flag.unixTime {
		result = transformers.DateTransformer(result).Transform()
	}

	if flag.concise {
		result = transformers.ConciseTransformer(result).Transform()
	}

	csvWriter := writer.NewCSVWriter(flag.fileName)
	err = csvWriter.Write(result)
	if err != nil {
		return fmt.Errorf("failure writing to csv: %v", err)
	}

	return nil
}

// ./dt-reporter -p https://prometheus.kapitalbank.az -q max(apache_available)' -c -f output.csv -st "2023-06-21 00:00:00" -et "2021-07-21 00:00:00"
