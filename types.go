package sensonetEbus

import "time"

const (
	STRATEGY_NONE                  = 0
	STRATEGY_HOTWATER              = 1
	STRATEGY_HEATING               = 2
	STRATEGY_HOTWATER_THEN_HEATING = 3

	OPERATIONMODE_AUTO        string = "auto"
	QUICKMODE_HOTWATER        string = "Hotwater Boost"
	QUICKMODE_HEATING         string = "Heating Quick Veto"
	QUICKMODE_NOTHING         string = "Charger running idle"
	QUICKMODE_ERROR_ALREADYON string = "Error. A quickmode is already running"

	EBUSDREAD_STATUS_TIME                 = "vdatetime"
	EBUSDREAD_STATUS_OUTSIDETEMPERATURE   = "outsidetemp"
	EBUSDREAD_STATUS_SYSTEMFLOWTEMPERATUE = "SystemFlowTemp"
	EBUSDREAD_STATUS_WATERPRESSURE        = "WaterPressure"
	EBUSDREAD_STATUS_CURRENTCONSUMEDPOWER = "CurrentConsumedPower"
	EBUSDREAD_STATUS_STATUS01             = "Status01"
	EBUSDREAD_STATUS_STATE                = "State"
	EBUSDREAD_HOTWATER_OPMODE             = "HwcOpMode"
	EBUSDREAD_HOTWATER_TEMPDESIRED        = "HwcTempDesired"
	EBUSDREAD_HOTWATER_STORAGETEMP        = "HwcStorageTemp"
	EBUSDREAD_HOTWATER_SFMODE             = "HwcSFMode"
	EBUSDREAD_ZONE_NAME                   = "Shortname"             //To be added by the zone prefix
	EBUSDREAD_ZONE_ACTUALROOMTEMPDESIRED  = "ActualRoomTempDesired" //To be added by the zone prefix
	EBUSDREAD_ZONE_OPMODE                 = "OpMode"                //To be added by the zone prefix
	EBUSDREAD_ZONE_SFMODE                 = "SFMode"                //To be added by the zone prefix
	EBUSDREAD_ZONE_ROOMTEMP               = "RoomTemp"              //To be added by the zone prefix
	EBUSDREAD_ZONE_QUICKVETOTEMP          = "QuickVetoTemp"         //To be added by the zone prefix
	EBUSDREAD_ZONE_QUICKVETOENDDATE       = "QuickVetoEndDate"      //To be added by the zone prefix
	EBUSDREAD_ZONE_QUICKVETOENDTIME       = "QuickVetoEndTime"      //To be added by the zone prefix
	EBUSDREAD_ZONE_QUICKVETODURATION      = "QuickVetoDuration"     //To be added by the zone prefix

	HWC_SFMODE_BOOST   = "load"
	HWC_SFMODE_NORMAL  = "auto"
	ZONE_SFMODE_BOOST  = "veto"
	ZONE_SFMODE_NORMAL = "auto"
	//HOTWATERINDEX_DEFAULT                = 255
	ZONEINDEX_DEFAULT        = 0
	ZONEVETOSETPOINT_DEFAULT = 20.0
	ZONEVETODURATION_DEFAULT = 0.5

	//eBusd errors
	EBUSD_ERROR_ELEMENTNOTFOUND = "ERR: element not found"
	EBUSD_ERROR_NOSIGNAL        = "ERR: no signal"
	EBUSD_ERROR_DUMMY           = "ERR: dummy"
)

type ParsStruct struct {
	EbusdAddress string `json:"ebusdaddress"`
}

// Types fpr Vaillant data

const NUMBER_OF_ZONES_TO_READ = 3

type VaillantRelDataZones struct {
	Index                 int
	Name                  string
	ActualRoomTempDesired float64
	OpMode                string
	SFMode                string
	QuickVetoTemp         float64
	QuickVetoEndTime      string
	QuickVetoEndDate      string
	InsideTemperature     float64
	RoomTemp              float64
}

/* not used yet
type VaillantRelDataHeatCircuits struct {
	Index                 int
	ActualFlowTempDesired float64
	FlowTemp              float64
	Status                string
}*/

type VaillantRelData struct {
	//SerialNumber string
	//Timestamp    int64
	//PvMode          int64
	//PvModeActive    int64
	//PvModeTimestamp int64
	LastGetSystem time.Time

	Status struct {
		Time                  string
		SensorData1           string
		SensorData2           string
		OutsideTemperature    float64
		SystemFlowTemperature float64
		WaterPressure         float64
		//ControllerForSFMode   string
		CurrentConsumedPower float64
		Status01             string
		State                string
	}

	Hotwater struct {
		//Index          int //
		HwcTempDesired float64
		HwcOpMode      string
		HwcStorageTemp float64
		HwcSFMode      string
	}

	Zones []VaillantRelDataZones

	//	HeatCircuits []VaillantRelDataHeatCircuits
}

/*
const (

	RESOLUTION_HOUR  = "HOUR"
	RESOLUTION_DAY   = "DAY"
	RESOLUTION_MONTH = "MONTH"

)
*/
type Logger interface {
	Printf(msg string, arg ...any)
}

type HeatingParStruct struct {
	ZoneIndex    int
	VetoSetpoint float32
	VetoDuration float32
}
