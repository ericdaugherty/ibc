package main

import (
	"bytes"
	"context"
	"fmt"
	"html/template"
	"io"
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/ericdaugherty/ibc"
	flags "github.com/jessevdk/go-flags"
	gomail "gopkg.in/gomail.v2"
)

var statusTemplateHTML = `<div>
<h1>Boiler Status</h1>
Boiler Model:  {{.boilerData.Model}}<br/>
Firmware:      {{.boilerData.FirmwareVersion}} {{.boilerData.FirmwareDate}}<br/>
Boiler Status: {{.extDetail.Status}}<br/>
Boiler Status: {{.boilerData.Status}}<br/>
Errors:        {{.extDetail.Errors}}<br/>
Warnings:      {{.extDetail.Warnings}}<br/>
Supply Temp:   {{TempAsF .extDetail.SupplyTemp}}F<br/>
Return Temp:   {{TempAsF .extDetail.ReturnTemp}}F<br/>
DWH Tank Temp: {{TempAsF .extDetail.TankTemp}}F<br/>
Cycles:        {{.extDetail.Cycles}}<br/>
Servicing:     {{range $index, $element := .extDetail.ServicingLoadNumbers}}{{if $index}},{{end}}{{$element}}{{end}}<br/>
Calling:       {{range $index, $element := .extDetail.CallingLoadNumbers}}{{if $index}},{{end}}{{$element}}{{end}}<br/>
Circulating:   {{range $index, $element := .extDetail.CirculatingLoadNumbers}}{{if $index}},{{end}}{{$element}}{{end}}<br/>
</div>`

var loadStatusTemplateHTML = `<div>
<h2>Load {{.LoadNum}} Status</h2>
Load Type: {{.lsd.LoadTypeName}}<br/>
Heat Output: {{.lsd.HeatOut}} MBtu<br/>
Load Cycles: {{.lsd.Cycles}}<br/>
</div>`

var b ibc.Boiler
var lastDateRecorded int
var lastEmailSent time.Time

var opts struct {
	BoilerURL         string   `short:"u" long:"url" description:"URL of the Boiler, ex -u \"http://192.168.10.2/\"" required:"true"`
	DailyLogFile      string   `short:"o" long:"csvOutputFile" description:"Path to csv of daily cycles." required:"true"`
	EmailFrom         string   `short:"f" long:"emailFrom" description:"The email address to use for the FROM setting." required:"true"`
	EmailTo           []string `short:"t" long:"emailTo" description:"The email address to use for the TO setting. Can specify multiple." required:"true"`
	EmailSubject      string   `long:"emailSubject" description:"The email address to use for the FROM setting." default:"IBC Boiler Alert"`
	EmailServer       string   `short:"s" long:"emailServer" description:"The SMTP Server to use to send the email." required:"true"`
	EmailPort         int      `long:"emailServerPort" description:"The port to use to connect to the SMTP Server" default:"587"`
	EmailUser         string   `short:"l" long:"emailUser" description:"The SMTP Username to use, if needed."`
	EmailPass         string   `short:"p" long:"emailPass" description:"The SMTP Password to use, if needed."`
	EmailMuteDuration int      `short:"m" long:"emailMuteMinutes" description:"The amount of time to wait between sending emails." default:"60"`
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
		}
	}()

	monitor(ctx)
}

func monitor(ctx context.Context) {

	ticker := time.NewTicker(5 * time.Minute)
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
	boilerData, err := b.GetBoilerData()
	if err != nil {
		fmt.Println("Error retrieving data: ", err)
		return
	}
	if (boilerData.Status != ibc.Standby &&
		boilerData.Status != ibc.Heating &&
		boilerData.Status != ibc.Circulating) ||
		boilerData.Warnings > 0 {
		if time.Now().After(lastEmailSent.Add(time.Duration(opts.EmailMuteDuration) * time.Minute)) {
			emailStatus()
		}
	}
}

func emailStatus() {
	emailBuf := new(bytes.Buffer)
	emailBuf.WriteString("<body>")

	boilerData, err := b.GetBoilerData()
	if err != nil {
		log.Println("Error retrieving data: ", err)
		return
	}
	extDetail, err := b.GetBoilerExtDetailData()
	if err != nil {
		log.Println("Error retrieving data: ", err)
		return
	}

	tmplOpts := make(map[string]interface{})
	tmplOpts["boilerData"] = boilerData
	tmplOpts["extDetail"] = extDetail
	executeTemplate(statusTemplateHTML, tmplOpts, emailBuf)

	lsdSlice, err := b.GetLoadStatusData()
	if err != nil {
		log.Println("Error retrieving data: ", err)
		return
	}
	for _, lsd := range lsdSlice {
		tmplOpts = make(map[string]interface{})
		tmplOpts["LoadNum"] = lsd.Load + 1
		tmplOpts["lsd"] = lsd
		executeTemplate(loadStatusTemplateHTML, tmplOpts, emailBuf)
	}

	emailBuf.WriteString("</body>")
	emailResult(emailBuf.String())
}

func executeTemplate(templateBody string, data interface{}, w io.Writer) {
	funcMap := template.FuncMap{
		"TempAsF": b.TempAsF,
	}
	tmpl := template.New("").Funcs(funcMap)
	tmpl = template.Must(tmpl.Parse(templateBody))

	err := tmpl.Execute(w, data)
	if err != nil {
		log.Println(err)
	}
}

func emailResult(body string) {
	m := gomail.NewMessage()
	m.SetHeader("From", opts.EmailFrom)
	for _, to := range opts.EmailTo {
		m.SetHeader("To", to)
	}
	m.SetHeader("Subject", opts.EmailSubject)
	m.SetBody("text/html", body)

	d := gomail.NewDialer(opts.EmailServer, opts.EmailPort, opts.EmailUser, opts.EmailPass)

	if err := d.DialAndSend(m); err != nil {
		log.Println(err)
		return
	}
	lastEmailSent = time.Now()
}
