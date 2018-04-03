package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/ericdaugherty/ibc"
	flags "github.com/jessevdk/go-flags"
)

var b ibc.Boiler
var lastDateRecorded int

var opts struct {
	BoilerURL    string `short:"u" long:"url" description:"URL of the Boiler, ex -u \"http://192.168.10.2/\"" required:"true"`
	DailyLogFile string `short:"o" long:"csvOutputFile" description:"Path to csv of daily cycles." required:"true"`
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

	b = ibc.Boiler{BaseURL: opts.BoilerURL}

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
		case <-ctx.Done():
			fmt.Println("Done")
		}
	}()

	monitor(ctx)
}

func monitor(ctx context.Context) {

	ticker := time.NewTicker(5 * time.Second)

	t := time.Now()
	recordDailyCycles(t)
	checkErrors()

	log.Println("Monitoring...")
	for {
		select {
		case t = <-ticker.C:
			recordDailyCycles(t)
			checkErrors()
		case <-ctx.Done():
			return
		}
	}
}

func recordDailyCycles(t time.Time) {
	sendWeekly := false
	// Check to see if we should record a new daily log. Record only once after 11:50p each day.
	if t.After(time.Date(t.Year(), t.Month(), t.Day(), 23, 50, 0, 0, t.Location())) && t.YearDay() != lastDateRecorded {
		lastDateRecorded = t.YearDay()
		sendWeekly = t.Weekday() == time.Saturday

		bedd, err := b.GetBoilerExtDetailData()
		if err != nil {
			log.Println(err)
			return
		}

		out := fmt.Sprintf("%s,%d", t.Format("2006-01-02"), bedd.Cycles)
		lsd, err := b.GetLoadStatusData()
		if err != nil {
			log.Println(err)
			return
		}
		loadCycles := make([]int, 4)
		for i := range loadCycles {
			if i < len(lsd) {
				out = fmt.Sprintf("%v,%d", out, lsd[i].Cycles)
			} else {
				out = fmt.Sprintf("%v,0", out)
			}
		}

		f, err := os.OpenFile(opts.DailyLogFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			log.Fatal(err)
		}
		defer func() {
			f.Close()
		}()

		info, err := f.Stat()
		if err != nil {
			log.Fatal(err)
		}
		// Add a header row if the file is new.
		if info.Size() == 0 {
			if _, err := f.Write([]byte("Date,Total,Load 1,Load 2,Load 3,Load 4\n")); err != nil {
				log.Fatal(err)
			}
		}

		// Write the data.
		if _, err := f.Write([]byte(out + "\n")); err != nil {
			log.Fatal(err)
		}
	}
	if sendWeekly {
		sendWeeklySummary()
	}
}

func sendWeeklySummary() {

}

func checkErrors() {

}
