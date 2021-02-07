package main

import (
	"context"
	"encoding/csv"
	"log"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"time"

	"github.com/ericdaugherty/ibc"
	"github.com/jessevdk/go-flags"
)

var opts struct {
	BoilerURL  string `short:"u" long:"url" description:"URL of the Boiler, ex -u \"http://192.168.10.2/\"" required:"true"`
	OutputFile string `short:"f" long:"file" description:"File name/path of the output CSV file" required:"true"`
	Interval   int    `short:"i" long:"interval" description:"The number of minutes to wait between log outputs." required:"true"`
}
var parser = flags.NewParser(&opts, flags.Default)

func main() {

	// Parse command line flags.
	if _, err := parser.Parse(); err != nil {
		if flagsErr, ok := err.(*flags.Error); ok && flagsErr.Type == flags.ErrHelp {
			os.Exit(0)
		} else {
			os.Exit(1)
		}
	}

	// Open the boiler
	b := ibc.Boiler{BaseURL: opts.BoilerURL}

	// Open the file for writing
	f, err := os.OpenFile(opts.OutputFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		panic(err)
	}
	fi, err := f.Stat()
	if err != nil {
		panic(err)
	}
	w := csv.NewWriter(f)

	// If the file is newly created, write the header
	if fi.Size() == 0 {
		header := []string{
			"time",
			"status",
			"errors",
			"warnings",
			"servicing",
			"airTemp",
			"cycles",
			"indoorTemp",
			"mbh",
			"opStatus",
			"outdoorTemp",
			"pumps",
			"returnTemp",
			"secondaryTemp",
			"servicing",
			"stackTemp",
			"supplyTemp",
			"tamkTemp",
			"targetTemp",
			"deltaPressure",
			"inletPressure",
			"outletPressure",
		}
		w.Write(header)
		w.Flush()
	}

	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	defer func() {
		signal.Stop(c)
		cancel()
	}()

	go func() {
		select {
		case <-c:
			cancel()
		}
	}()

	logData(w, b)
	ticker := time.NewTicker(time.Duration(opts.Interval) * time.Minute)
	for {
		select {
		case <-ticker.C:
			logData(w, b)
		case <-ctx.Done():
			w.Flush()
			f.Close()
			return
		}
	}
}

func logData(w *csv.Writer, b ibc.Boiler) {

	bedd, err := b.GetBoilerExtDetailData()
	if err != nil {
		log.Println(err)
		return
	}

	// lsd, err := b.GetLoadStatusData()
	// if err != nil {
	// 	log.Println(err)
	// 	return
	// }
	servicing := bedd.ServicingLoadNumbers()
	servicingStrings := make([]string, len(servicing))
	for i, s := range servicing {
		servicingStrings[i] = strconv.Itoa(s)
	}

	w.Write([]string{
		time.Now().Format(time.RFC3339),
		bedd.Status,
		bedd.Errors,
		bedd.Warnings,
		strings.Join(servicingStrings, ","),
		strconv.Itoa(bedd.AirTemp),
		strconv.Itoa(bedd.Cycles),
		strconv.Itoa(bedd.IndoorTemp),
		strconv.Itoa(bedd.MBH),
		strconv.Itoa(bedd.OpStatus),
		strconv.Itoa(bedd.OutdoorTemp),
		strconv.Itoa(bedd.Pumps),
		strconv.Itoa(bedd.ReturnTemp),
		strconv.Itoa(bedd.SecondaryTemp),
		strconv.Itoa(bedd.Servicing),
		strconv.Itoa(bedd.StackTemp),
		strconv.Itoa(bedd.SupplyTemp),
		strconv.Itoa(bedd.TankTemp),
		strconv.Itoa(bedd.TargetTemp),
		strconv.FormatFloat(bedd.DeltaPressure, 'f', 2, 64),
		strconv.FormatFloat(bedd.InletPressure, 'f', 2, 64),
		strconv.FormatFloat(bedd.OutletPressure, 'f', 2, 64),
	})
	w.Flush()
}
