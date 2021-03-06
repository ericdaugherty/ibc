package ibc

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

// System Status Constants
const (
	Standby      = 0
	Purging      = 1
	Igniting     = 2
	Heating      = 3
	Circulating  = 4
	Error        = 5
	Initializing = 6
)

// G3 soft errors. TODO: Updated to handle non-G3 soft errors.
var hardErrorsBitMask = [...]int{0x01, 0x10, 0x20, 0x02, 0x04, 0x08}
var hardErrors = [...]string{"Ignition Trials Exceeded", "Roll Out Switch", "Low Water Cutoff", "Module High Current", "Sec/Indoor Sensor", "Low Water Cutoff"}
var softErrors1BitMask = [...]int{0x0001, 0x0004, 0x0008, 0x0010, 0x0020, 0x0040, 0x0080}
var softErrors1 = [...]string{"Flame Sig/Vent Blocked", "Low RPM/Air Flow", "No/Low Water Flow", "Water High Limit", "Vent High Limit", "Interlock 1 Open", "Interlock 2 Open"}
var softErrors2BitMask = [...]int{0x0100, 0x0200, 0x0400, 0x0800, 0x2000, 0x4000, 0x8000, 0x1000}
var softErrors2 = [...]string{"Inlet Pressure Sensor", "Fan Pressure", "No/Low Water Flow", "Low Module Current", "See Error Log/SIM", "Low Water Pressure", "Max deltaT Exceeded", "Reversed Flow"}
var systemErrorBitMask = [...]int{0x01, 0x02, 0x04, 0x08, 0x10, 0x20, 0x40, 0x80}
var systemErrors = [...]string{"CANbus", "CGI Task", "I2C Bus 0", "I2C Bus 1", "BACnet Task", "GPIO Expander", "LCD Module/Bus", "FRAM Module"}

// Boiler represents a specific IBC Boiler to interact with.
type Boiler struct {
	BaseURL string
}

// BoilerStatusData represents the data returned from the ReqBoilerStatusData request.
type BoilerStatusData struct {
	//"rbid": 0
	//"object_no": 3
	Status int `json:"status"`
	// MBH is Thousands of BTUs per hour
	MBH                     int `json:"mbh"`
	SupplyTemp              int `json:"supplyT"`
	ReturnTemp              int `json:"returnT"`
	SecondaryTemp           int `json:"secondaryT"`
	DomesticWaterHeaterTemp int `json:"dhwT"`
	PSIG                    int `json:"psig"`
	Warning                 int `json:"warning"`
}

// BoilerLogData represents the data returned by the ReqBoilerLogData request
type BoilerLogData struct {
	//"rbid": 0
	//"object_no": 6
	PowerOnHrs   int `json:"PowerOnHrs"`
	BurnerOnHrs  int `json:"BurnerOnHrs"`
	Load1OnTime  int `json:"Load1OnTime"`
	Load2OnTime  int `json:"Load2OnTime"`
	Load3OnTime  int `json:"Load3OnTime"`
	Load4OnTime  int `json:"Load4OnTime"`
	RemoteOnTime int `json:"RemoteOnTime"`
	Starts       int `json:"Starts"`
	Trials       int `json:"Trials"`
	Errors       int `json:"Errors"`
	Warnings     int `json:"Warnings"`
	LogEntries   int `json:"LogEntries"`
	Cycles       int `json:"Cycles"`
	BiasCount    int `json:"BiasCount"`
}

// BoilerErrorLogData represents the data for an single entry in the error log.
type BoilerErrorLogData struct {
	//"rbid": 0
	//"object_no": 7
	//log_no:
	Time           string `json:"Time"`
	Date           string `json:"Date"`
	MinErr         int    `json:"MinErr"`
	MajErr         int    `json:"MajErr"`
	SysErr         int    `json:"SysErr"`
	HeatOut        int    `json:"HeatOut"`
	FanRPM         int    `json:"FanRPM"`
	InletTemp      int    `json:"InletTemp"`
	OutletTemp     int    `json:"OutletTemp"`
	BoardTemp      int    `json:"BoardTemp"`
	DiffPressure   int    `json:"DiffPressure"`
	InletTRate     int    `json:"InletTRate"`
	OutletTRate    int    `json:"OutletTRate"`
	InletPressure  int    `json:"InletPressure"`
	OutletPressure int    `json:"OutletPressure"`
	FlameSense     int    `json:"FlameSense"`
	SIMFlame       int    `json:"SIM_Flame"`
	SIMStatus      int    `json:"SIM_Status"`
	FanDutyCycle   int    `json:"FanDutyCycle"`
	BVGauge        int    `json:"BV_Gauge"`
}

// BoilerData represents the data returend by the ReqBoilerData request
type BoilerData struct {
	//"rbid": 0
	//"object_no": 11,
	Status          int    `json:"status"`
	Master          int    `json:"master"`
	NetMaster       int    `json:"net_master"`
	Warnings        int    `json:"warnings"`
	Imperial        int    `json:"imperial"`
	OnTime          int    `json:"ontime"`
	BoilerID        int    `json:"boiler_id"`
	DIMTime         int    `json:"dim_time"`
	Configured      int    `json:"configured"`
	ModelNum        int    `json:"model_num"`
	DesignT         int    `json:"designT"`
	Model           string `json:"model"`
	FirmwareVersion string `json:"fwversion"`
	FirmwareDate    string `json:"fwdate"`
	SICCModule      bool   `json:"sicc_module"`
}

// BoilerStandardData represnets the data returned by the ReqBoilerStandardData request.
type BoilerStandardData struct {
	//"rbid": 0
	//"object_no": 13
	Load1Type    int  `json:"Load1Type"`
	Load2Type    int  `json:"Load2Type"`
	Load3Type    int  `json:"Load3Type"`
	Load4Type    int  `json:"Load4Type"`
	Load1Emitter int  `json:"Load1Emitter"`
	Load2Emitter int  `json:"Load2Emitter"`
	Load3Emitter int  `json:"Load3Emitter"`
	Load4Emitter int  `json:"Load4Emitter"`
	SB1Enable    bool `json:"SB1Enable"`
	SB2Enable    bool `json:"SB2Enable"`
	SB3Enable    bool `json:"SB3Enable"`
	SB4Enable    bool `json:"SB4Enable"`
	Occupied     int  `json:"Occupied"`
	Imperial     int  `json:"Imperial"`
}

// BoilerExtDetailData represents the data returend by the ReqReqBoilerExtDetailData request.
type BoilerExtDetailData struct {
	// "rbid": 0
	// "object_no": 19
	BoilerID       int     `json:"BoilerID"`
	Status         string  `json:"Status"`
	Warnings       string  `json:"Warnings"`
	Errors         string  `json:"Errors"`
	MBH            int     `json:"MBH"`
	SupplyTemp     int     `json:"SupplyT"`
	ReturnTemp     int     `json:"ReturnT"`
	TargetTemp     int     `json:"TargetT"`
	StackTemp      int     `json:"StackT"`
	AirTemp        int     `json:"AirT"`
	IndoorTemp     int     `json:"IndoorT"`
	OutdoorTemp    int     `json:"OutdoorT"`
	SecondaryTemp  int     `json:"SecondaryT"`
	TankTemp       int     `json:"TankT"`
	InletPressure  float64 `json:"InletPressure"`
	OutletPressure float64 `json:"OutletPressure"`
	DeltaPressure  float64 `json:"DeltaPressure"`
	Servicing      int     `json:"Servicing"`
	Cycles         int     `json:"Cycles"`
	MajorError     int     `json:"MajorError"`
	MinorError     int     `json:"MinorError"`
	SystemError    int     `json:"SystemError"`
	WarnFlags      int     `json:"WarnFlags"`
	Pumps          int     `json:"Pumps"`
	OpStatus       int     `json:"OpStatus"`
}

// ServicingLoadNumbers returns the load numbers the boiler is currently servicing.
func (bedd BoilerExtDetailData) ServicingLoadNumbers() []int {
	// TODO: Remote (0xFFFF) and Summer Off (0xF000)
	s := bedd.Servicing & 0xF
	return getLoadNumbersFromBits(s)
}

// CirculatingLoadNumbers returns the load numbers the boiler is currently circulating.
func (bedd BoilerExtDetailData) CirculatingLoadNumbers() []int {
	s := bedd.Servicing
	s >>= 4
	s &= 0xF
	return getLoadNumbersFromBits(s)
}

// CallingLoadNumbers returns the load numbers that is currently calling for heat but is not being serviced.
func (bedd BoilerExtDetailData) CallingLoadNumbers() []int {
	s := bedd.Servicing
	s >>= 8
	s &= 0xF
	return getLoadNumbersFromBits(s)
}

// BoilerFactoryData represents the data returned by the ReqBoilerFactoryData request.
type BoilerFactoryData struct {
	//"rbid": 0
	//"object_no": 20
	InletP     int `json:"InletP"`
	OutletP    int `json:"OutletP"`
	DeltaP     int `json:"DeltaP"`
	FlowRate   int `json:"FlowRate"`
	FanSpeed   int `json:"FanSpeed"`
	FanDuty    int `json:"FanDuty"`
	FanTarget  int `json:"FanTarget"`
	RequiredP  int `json:"RequiredP"`
	FanP       int `json:"FanP"`
	OffsetP    int `json:"OffsetP"`
	VentFactor int `json:"VentFactor"`
	VarDuty    int `json:"VarDuty"`
	Responding int `json:"Responding"`
	Firing     int `json:"Firing"`
	Available  int `json:"Available"`
	FCurrent   int `json:"F_Current"`
	HeatOut    int `json:"HeatOut"`
	FanHeatOut int `json:"FanHeatOut"`
	InletT     int `json:"InletT"`
	OutletT    int `json:"OutletT"`
	StackT     int `json:"StackT"`
	RPMLimit   int `json:"RPMLimit"`
	SICCFlame  int `json:"SICC_Flame"`
}

// GetLoadTypeName returns the name of the specified load type. Pass in the value of Load1Type as the parameter.
func (bsd BoilerStandardData) GetLoadTypeName(loadType int) string {
	return loadName(loadType)
}

// LoadStatusData represents the data returned by the ReqLoadStatusData request.
type LoadStatusData struct {
	// "rbid": 0
	// "object_no": 32
	Load         int `json:"Load"`
	Type         int `json:"Type"`
	HeatOut      int `json:"HeatOut"`
	SupplyTemp   int `json:"SupplyT"`
	ReturnTemp   int `json:"ReturnT"`
	BoilerMax    int `json:"BoilerMax"`
	BoilerDiff   int `json:"BoilerDiff"`
	Cycles       int `json:"Cycles"`
	Priority     int `json:"Priority"`
	Temperature1 int `json:"Temperature1"`
	Temperature2 int `json:"Temperature2"`
	Temperature3 int `json:"Temperature3"`
	Temperature4 int `json:"Temperature4"`
	Temperature5 int `json:"Temperature5"`
	Temperature6 int `json:"Temperature6"`
}

// LoadTypeName returns the name of the LoadType for this Load.
func (lsd LoadStatusData) LoadTypeName() string {
	return loadName(lsd.Type)
}

// Block of constants define Request types.
const (
	ReqMasterBoilerData          = 2
	ReqBoilerStatusData          = 3
	ReqBoilerRunProfileData      = 5
	ReqBoilerLogData             = 6
	ReqBoilerErrorLogData        = 7
	ReqBoilerData                = 11
	ReqBoilerStandardData        = 13
	ReqBoilerSetbackData         = 14
	ReqBoilerAdvSetttingsData    = 15
	ReqBoilerLoadSettingsData    = 16
	ReqBoilerMultiSettingData    = 17
	ReqBoilerCleaningSettingData = 18
	ReqBoilerExtDetailData       = 19
	ReqBoilerFactoryData         = 20
	ReqBoilerFactorySettingsData = 21
	ReqSiteLogData               = 23
	ReqClockData                 = 24
	ReqLoadPairingData           = 25
	ReqBoilerCaptureData         = 26
	ReqBoilerTempSensorData      = 27
	ReqBoilerRestore             = 29
	ReqAlertData                 = 31
	ReqLoadStatusData            = 32
	ReqBoilerSiteData            = 34
	ReqBoilerVersions            = 35
	ReqNetworkBoilerData         = 38
	ReqAdvancedOptionsData       = 42
	ReqBoilerSIMData             = 44
	ReqSlaveMACADDRSData         = 49
	ReqProgSetbackData           = 50
	ReqInternetUpdateData        = 51
	ReqPasswordData              = 99
)

// TempAsF returns the specified temperature in Fahrenheit. By default, all temperatures returned by the API are Celcius * 4.
func (b Boiler) TempAsF(temp int) int {
	t := ((temp * 9) / 5) + (4 * 32)
	return t / 4
}

// TempAsC returns the specified temperature in Celsius. By default, all temperatures returned by the API are Celsius * 4.
func (b Boiler) TempAsC(temp int) float32 {
	return float32(temp) / float32(4)
}

type requestObject struct {
	ObjectNum     int `json:"object_no"`
	ObjectRequest int `json:"object_request"`
	BoilerNum     int `json:"boiler_no"`
	LoadNum       int `json:"load_no,omitempty"`
	ObjectIndex   int `json:"object_index"`
}

var loadNames = [...]string{"Off", "DHW", "Reset Heating", "Set Point", "External Control", "Manual Control", "Zone Of"}

// GetData queries the boiler and returns a map representing the response.
func (b Boiler) GetData(requestNumber int) (interface{}, error) {
	reqObj := requestObject{ObjectNum: 100, ObjectRequest: requestNumber, BoilerNum: 0}
	var respObj interface{}
	return respObj, b.getData(reqObj, &respObj)
}

// GetDataForLoad queries the boiler about data for a specific load and returns a map representing the response.
func (b Boiler) GetDataForLoad(requestNumber int, loadNumber int) (interface{}, error) {
	reqObj := requestObject{ObjectNum: 100, ObjectRequest: requestNumber, BoilerNum: 0, LoadNum: loadNumber}
	var respObj interface{}
	return respObj, b.getData(reqObj, &respObj)
}

// GetBoilerStatusData returns the BoilerStatusData for the current boiler.
func (b Boiler) GetBoilerStatusData() (BoilerStatusData, error) {
	reqObj := requestObject{ObjectNum: 100, ObjectRequest: ReqBoilerStatusData, BoilerNum: 0, LoadNum: 0}
	var respObj = BoilerStatusData{}
	return respObj, b.getData(reqObj, &respObj)
}

// GetBoilerLogData returns the BoilerStatusData for the current boiler.
func (b Boiler) GetBoilerLogData() (BoilerLogData, error) {
	reqObj := requestObject{ObjectNum: 100, ObjectRequest: ReqBoilerLogData, BoilerNum: 0, LoadNum: 0}
	var respObj = BoilerLogData{}
	return respObj, b.getData(reqObj, &respObj)
}

// GetBoilerErrLogData returns the BoilerErrorLogData for the specified logEntryNumber.
func (b Boiler) GetBoilerErrLogData(logEntryNumber int) (BoilerErrorLogData, error) {
	reqObj := requestObject{ObjectNum: 100, ObjectRequest: ReqBoilerErrorLogData, BoilerNum: 0, ObjectIndex: logEntryNumber}
	var respObj = BoilerErrorLogData{}
	return respObj, b.getData(reqObj, &respObj)
}

// GetBoilerData returns the BoilerStatusData for the current boiler.
func (b Boiler) GetBoilerData() (BoilerData, error) {
	reqObj := requestObject{ObjectNum: 100, ObjectRequest: ReqBoilerData, BoilerNum: 0, LoadNum: 0}
	var respObj = BoilerData{}
	return respObj, b.getData(reqObj, &respObj)
}

// GetBoilerStandardData returns the BoilerStandardData response for the current boiler.
func (b Boiler) GetBoilerStandardData() (BoilerStandardData, error) {
	reqObj := requestObject{ObjectNum: 100, ObjectRequest: ReqBoilerStandardData, BoilerNum: 0, LoadNum: 0}
	var respObj = BoilerStandardData{}
	return respObj, b.getData(reqObj, &respObj)
}

// GetBoilerExtDetailData returns the BoilerExtDetailData response for the current boiler.
func (b Boiler) GetBoilerExtDetailData() (BoilerExtDetailData, error) {
	reqObj := requestObject{ObjectNum: 100, ObjectRequest: ReqBoilerExtDetailData, BoilerNum: 0, LoadNum: 0}
	var respObj = BoilerExtDetailData{}
	return respObj, b.getData(reqObj, &respObj)
}

// GetBoilerFactoryData returns the BoilerFactoryData response for the current boiler.
func (b Boiler) GetBoilerFactoryData() (BoilerFactoryData, error) {
	reqObj := requestObject{ObjectNum: 100, ObjectRequest: ReqBoilerFactoryData, BoilerNum: 0, LoadNum: 0}
	var respObj = BoilerFactoryData{}
	return respObj, b.getData(reqObj, &respObj)
}

// GetLoadStatusDataForLoad returns the LoadStatusData response for the current boiler and specified load.
func (b Boiler) GetLoadStatusDataForLoad(loadNum int) (LoadStatusData, error) {
	reqObj := requestObject{ObjectNum: 100, ObjectRequest: ReqLoadStatusData, BoilerNum: 0, LoadNum: loadNum}
	var respObj = LoadStatusData{}
	return respObj, b.getData(reqObj, &respObj)
}

// GetLoadStatusData returns the LoadStatusData response for the active loads for the current boiler.
func (b Boiler) GetLoadStatusData() ([]LoadStatusData, error) {
	var lsd = make([]LoadStatusData, 0, 4)

	bsd, err := b.GetBoilerStandardData()
	if err != nil {
		return lsd, err
	}

	f := func(loadType, loadNum int) {
		if loadType > 0 {
			reqObj := requestObject{ObjectNum: 100, ObjectRequest: ReqLoadStatusData, BoilerNum: 0, LoadNum: loadNum}
			var respObj = LoadStatusData{}
			b.getData(reqObj, &respObj)
			lsd = append(lsd, respObj)
		}
	}
	f(bsd.Load1Type, 1)
	f(bsd.Load2Type, 2)
	f(bsd.Load3Type, 3)
	f(bsd.Load4Type, 4)

	return lsd, nil
}

func loadName(loadNumber int) string {
	if loadNumber < 0 || loadNumber >= len(loadNames) {
		return "Unknown"
	}
	return loadNames[loadNumber]
}

func getLoadNumbersFromBits(in int) []int {
	var lt = make([]int, 0, 4)
	for i, bit := 1, 1; i <= 4; i++ {
		if in&bit != 0 {
			lt = append(lt, i)
		}
		bit <<= 1
	}
	return lt
}

// GetErrorString returns a descripton of the error code specified. Assumes G3 Boilers
func GetErrorString(minErr int, majErr int, sysErr int) string {
	if sysErr > 0 {
		for i := 0; i < len(systemErrorBitMask); i++ {
			if (sysErr & systemErrorBitMask[i]) > 0 {
				return systemErrors[i]
			}
		}
	}

	softToHardErrors := 0
	hardToSoftErrors := 0
	if (minErr & 0x10) > 0 {
		softToHardErrors |= 0x20
	}
	if (minErr & 0x20) > 0 {
		softToHardErrors |= 0x10
	}
	if (majErr & 0x4) > 0 {
		hardToSoftErrors = 0x4
	}
	// Remove these bits from their associated errors.
	minErr = minErr &^ 0x10
	minErr = minErr &^ 0x20
	majErr = majErr &^ 0x4

	if majErr > 0 || softToHardErrors > 0 {
		errString := "Unknown"
		for i := 0; i < len(hardErrorsBitMask); i++ {
			if majErr&hardErrorsBitMask[i] > 0 {
				errString = hardErrors[i]
				if (hardErrorsBitMask[i] & 0x20) > 0 {
					errString = "Vent High Pressure"
				} else if (hardErrorsBitMask[i] & 0x4) > 0 {
					errString = "Temperature Probe Error"
				}
			}
		}

		if softToHardErrors > 0 {
			if (softToHardErrors & 0x10) > 0 {
				errString = "Vent High Limit"
			}

			if (softToHardErrors & 0x20) > 0 {
				errString = "Water High Limit"
			}
		}
		return errString
	}

	if minErr > 0 || hardToSoftErrors > 0 {
		errString := "Unknown"
		for i := 0; i < len(softErrors1BitMask); i++ {
			if minErr&softErrors1BitMask[i] > 0 {
				errString = softErrors1[i]
			}
		}

		if hardToSoftErrors > 0 {
			errString = "Temp. Probe Error"
		}

		for i := 0; i < len(softErrors2BitMask); i++ {
			if minErr&softErrors2BitMask[i] > 0 {
				errString = softErrors2[i]
			}
		}

		return errString
	}

	return "Unknown"
}

func (b Boiler) getData(reqObj requestObject, respObj interface{}) error {

	sep := "/"
	if strings.HasSuffix(b.BaseURL, "/") {
		sep = ""
	}

	url := fmt.Sprintf("%s%scgi-bin/bc2-cgi", b.BaseURL, sep)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}

	jsonBytes, err := json.Marshal(reqObj)
	if err != nil {
		return err
	}

	q := req.URL.Query()
	q.Add("json", string(jsonBytes))
	req.URL.RawQuery = q.Encode()

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	err = json.Unmarshal(body, &respObj)

	return nil
}
