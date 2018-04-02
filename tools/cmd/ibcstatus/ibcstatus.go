package main

import (
	"fmt"
	"io"
	"os"

	"github.com/alecthomas/template"
	"github.com/ericdaugherty/ibc"
	flags "github.com/jessevdk/go-flags"
)

var statusTemplateConsole = `Boiler Model:  {{.boilerData.Model}}
Firmware:      {{.boilerData.FirmwareVersion}} {{.boilerData.FirmwareDate}}
Boiler Status: {{.extDetail.Status}}
Errors:        {{.extDetail.Errors}}
Warnings:      {{.extDetail.Warnings}}
Supply Temp:   {{TempAsF .extDetail.SupplyTemp}}F
Return Temp:   {{TempAsF .extDetail.ReturnTemp}}F
DWH Tank Temp: {{TempAsF .extDetail.TankTemp}}F
Cycles:        {{.extDetail.Cycles}}
Servicing:     {{range $index, $element := .extDetail.ServicingLoadNumbers}}{{if $index}},{{end}}{{$element}}{{end}}
Calling:       {{range $index, $element := .extDetail.CallingLoadNumbers}}{{if $index}},{{end}}{{$element}}{{end}}
Circulating:   {{range $index, $element := .extDetail.CirculatingLoadNumbers}}{{if $index}},{{end}}{{$element}}{{end}}

`

var loadStatusTemplateConsole = `Load Number: {{.LoadNum}}
Load Type: {{.lsd.LoadTypeName}}
Heat Output: {{.lsd.HeatOut}} MBtu
Load Cycles: {{.lsd.Cycles}}

`

var b ibc.Boiler

var opts struct {
	BoilerURL string `short:"u" long:"url" description:"URL of the Boiler, ex -u \"http://192.168.10.2/\"" required:"true"`
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

	showStatus(b)
}

func showStatus(b ibc.Boiler) {

	boilerData, err := b.GetBoilerData()
	if err != nil {
		fmt.Println("Error retrieving data: ", err)
		return
	}
	extDetail, err := b.GetBoilerExtDetailData()
	if err != nil {
		fmt.Println("Error retrieving data: ", err)
		return
	}

	tmplOpts := make(map[string]interface{})
	tmplOpts["boilerData"] = boilerData
	tmplOpts["extDetail"] = extDetail
	executeTemplate(statusTemplateConsole, tmplOpts, os.Stdout)

	lsdSlice, err := b.GetLoadStatusData()
	if err != nil {
		fmt.Println("Error retrieving data: ", err)
		return
	}
	for _, lsd := range lsdSlice {
		tmplOpts = make(map[string]interface{})
		tmplOpts["LoadNum"] = lsd.Load + 1
		tmplOpts["lsd"] = lsd
		executeTemplate(loadStatusTemplateConsole, tmplOpts, os.Stdout)
	}
}

func executeTemplate(templateBody string, data interface{}, w io.Writer) {
	funcMap := template.FuncMap{
		"TempAsF": b.TempAsF,
	}
	tmpl := template.New("").Funcs(funcMap)
	tmpl = template.Must(tmpl.Parse(templateBody))

	err := tmpl.Execute(w, data)
	if err != nil {
		panic(err)
	}
}
