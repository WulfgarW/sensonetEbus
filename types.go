package sensonetEbus

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

type VaillantRelDataHeatCircuits struct {
	Index                 int
	ActualFlowTempDesired float64
	FlowTemp              float64
	Status                string
}

type VaillantRelData struct {
	//SerialNumber string
	//Timestamp    int64
	//PvMode          int64
	//PvModeActive    int64
	//PvModeTimestamp int64

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
		Index          int // necessary?
		HwcTempDesired float64
		HwcOpMode      string
		HwcStorageTemp float64
		HwcSFMode      string
	}

	Zones []VaillantRelDataZones

	HeatCircuits []VaillantRelDataHeatCircuits
}

/*
const (

)

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

/*
type SystemStatus struct {
	State struct {
		System struct {
			OutdoorTemperature           float64 `json:"outdoorTemperature"`
			OutdoorTemperatureAverage24H float64 `json:"outdoorTemperatureAverage24h"`
			SystemFlowTemperature        float64 `json:"systemFlowTemperature"`
			SystemWaterPressure          float64 `json:"systemWaterPressure"`
			EnergyManagerState           string  `json:"energyManagerState"`
			SystemOff                    bool    `json:"systemOff"`
		} `json:"system"`
		Zones            []StateZone             `json:"zones"`
		Circuits         []StateCircuit          `json:"circuits"`
		Dhw              []StateDhw              `json:"dhw"`
		DomesticHotWater []StateDomesticHotWater `json:"domesticHotWater"`
		// Ventilations
	} `json:"state"`
	Properties struct {
		System struct {
			ControllerType                     string  `json:"controllerType"`
			SystemScheme                       int     `json:"systemScheme"`
			BackupHeaterType                   string  `json:"backupHeaterType"`
			BackupHeaterAllowedFor             string  `json:"backupHeaterAllowedFor"`
			ModuleConfigurationVR71            int     `json:"moduleConfigurationVR71"`
			EnergyProvidePowerCutBehavior      string  `json:"energyProvidePowerCutBehavior"`
			SmartPhotovoltaicBufferOffset      float64 `json:"smartPhotovoltaicBufferOffset"`
			ExternalEnergyManagementActivation bool    `json:"externalEnergyManagementActivation"`
		} `json:"system"`
		Zones            []PropertiesZone             `json:"zones"`
		Circuits         []PropertiesCircuit          `json:"circuits"`
		Dhw              []PropertiesDhw              `json:"dhw"`
		DomesticHotWater []PropertiesDomesticHotWater `json:"domesticHotWater"`
		// Ventilations
	} `json:"properties"`
	Configuration struct {
		System struct {
			ContinuousHeatingStartSetpoint float64 `json:"continuousHeatingStartSetpoint"`
			AlternativePoint               float64 `json:"alternativePoint"`
			HeatingCircuitBivalencePoint   float64 `json:"heatingCircuitBivalencePoint"`
			DhwBivalencePoint              float64 `json:"dhwBivalencePoint"`
			AdaptiveHeatingCurve           bool    `json:"adaptiveHeatingCurve"`
			DhwMaximumLoadingTime          int     `json:"dhwMaximumLoadingTime"`
			DhwHysteresis                  float64 `json:"dhwHysteresis"`
			DhwFlowSetpointOffset          float64 `json:"dhwFlowSetpointOffset"`
			ContinuousHeatingRoomSetpoint  float64 `json:"continuousHeatingRoomSetpoint"`
			HybridControlStrategy          string  `json:"hybridControlStrategy"`
			MaxFlowSetpointHpError         float64 `json:"maxFlowSetpointHpError"`
			DhwMaximumTemperature          float64 `json:"dhwMaximumTemperature"`
			MaximumPreheatingTime          int     `json:"maximumPreheatingTime"`
			ParalellTankLoadingAllowed     bool    `json:"paralellTankLoadingAllowed"`
		} `json:"system"`
		Zones            []ConfigurationZone             `json:"zones"`
		Circuits         []ConfigurationCircuit          `json:"circuits"`
		Dhw              []ConfigurationDhw              `json:"dhw"`
		DomesticHotWater []ConfigurationDomesticHotWater `json:"domesticHotWater"`
		// Ventilations
	} `json:"configuration"`
}

type DhwData struct {
	State         StateDhw
	Properties    PropertiesDhw
	Configuration ConfigurationDhw
}

type DomesticHotWaterData struct {
	State         StateDomesticHotWater
	Properties    PropertiesDomesticHotWater
	Configuration ConfigurationDomesticHotWater
}

type ZoneData struct {
	State         StateZone
	Properties    PropertiesZone
	Configuration ConfigurationZone
}

type SystemAndStatus struct {
	SystemId     string
	SystemStatus SystemStatus
}

type AllSystems struct {
	SystemsAndStatus []SystemAndStatus
}

type StateZone struct {
	Index                                 int     `json:"index"`
	DesiredRoomTemperatureSetpointHeating float64 `json:"desiredRoomTemperatureSetpointHeating"`
	DesiredRoomTemperatureSetpoint        float64 `json:"desiredRoomTemperatureSetpoint"`
	CurrentRoomTemperature                float64 `json:"currentRoomTemperature,omitempty"`
	CurrentRoomHumidity                   float64 `json:"currentRoomHumidity,omitempty"`
	CurrentSpecialFunction                string  `json:"currentSpecialFunction"`
	HeatingState                          string  `json:"heatingState"`
}

type StateCircuit struct {
	Index                         int     `json:"index"`
	CircuitState                  string  `json:"circuitState"`
	CurrentCircuitFlowTemperature float64 `json:"currentCircuitFlowTemperature,omitempty"`
	HeatingCircuitFlowSetpoint    float64 `json:"heatingCircuitFlowSetpoint"`
	CalculatedEnergyManagerState  string  `json:"calculatedEnergyManagerState"`
}

type StateDhw struct {
	Index                  int     `json:"index"`
	CurrentSpecialFunction string  `json:"currentSpecialFunction"`
	CurrentDhwTemperature  float64 `json:"currentDhwTemperature"`
}

type StateDomesticHotWater struct {
	Index                              int     `json:"index"`
	CurrentSpecialFunction             string  `json:"currentSpecialFunction"`
	CurrentDomesticHotWaterTemperature float64 `json:"currentDomesticHotWaterTemperature"`
}

type PropertiesZone struct {
	Index                  int    `json:"index"`
	IsActive               bool   `json:"isActive"`
	ZoneBinding            string `json:"zoneBinding"`
	IsCoolingAllowed       bool   `json:"isCoolingAllowed"`
	AssociatedCircuitIndex int    `json:"associatedCircuitIndex"`
}

type PropertiesCircuit struct {
	Index                    int    `json:"index"`
	MixerCircuitTypeExternal string `json:"mixerCircuitTypeExternal"`
	HeatingCircuitType       string `json:"heatingCircuitType"`
}

type PropertiesDhw struct {
	Index       int     `json:"index"`
	MinSetpoint float64 `json:"minSetpoint"`
	MaxSetpoint float64 `json:"maxSetpoint"`
}

type PropertiesDomesticHotWater struct {
	Index       int     `json:"index"`
	MinSetpoint float64 `json:"minSetpoint"`
	MaxSetpoint float64 `json:"maxSetpoint"`
}

type TimeSlot struct {
	StartTime int `json:"startTime"`
	EndTime   int `json:"endTime"`
}

type Setpoint struct {
	StartTime int     `json:"startTime"`
	EndTime   int     `json:"endTime"`
	Setpoint  float64 `json:"setpoint"`
}

type MetaInfo struct {
	MinSlotsPerDay          int  `json:"minSlotsPerDay"`
	MaxSlotsPerDay          int  `json:"maxSlotsPerDay"`
	SetpointRequiredPerSlot bool `json:"setpointRequiredPerSlot"`
}

type TimeProgram struct {
	MetaInfo  MetaInfo   `json:"metaInfo"`
	Monday    []Setpoint `json:"monday"`
	Tuesday   []Setpoint `json:"tuesday"`
	Wednesday []Setpoint `json:"wednesday"`
	Thursday  []Setpoint `json:"thursday"`
	Friday    []Setpoint `json:"friday"`
	Saturday  []Setpoint `json:"saturday"`
	Sunday    []Setpoint `json:"sunday"`
}

type ConfigurationZone struct {
	Index   int `json:"index"`
	General struct {
		Name                 string    `json:"name"`
		HolidayStartDateTime time.Time `json:"holidayStartDateTime"`
		HolidayEndDateTime   time.Time `json:"holidayEndDateTime"`
		HolidaySetpoint      float64   `json:"holidaySetpoint"`
	} `json:"general"`
	Heating struct {
		OperationModeHeating      string      `json:"operationModeHeating"`
		SetBackTemperature        float64     `json:"setBackTemperature"`
		ManualModeSetpointHeating float64     `json:"manualModeSetpointHeating"`
		TimeProgramHeating        TimeProgram `json:"timeProgramHeating"`
	} `json:"heating"`
}

type ConfigurationCircuit struct {
	Index                                  int     `json:"index"`
	HeatingCurve                           float64 `json:"heatingCurve"`
	HeatingFlowTemperatureMinimumSetpoint  float64 `json:"heatingFlowTemperatureMinimumSetpoint"`
	HeatingFlowTemperatureMaximumSetpoint  float64 `json:"heatingFlowTemperatureMaximumSetpoint"`
	HeatDemandLimitedByOutsideTemperature  float64 `json:"heatDemandLimitedByOutsideTemperature"`
	HeatingCircuitFlowSetpointExcessOffset float64 `json:"heatingCircuitFlowSetpointExcessOffset"`
	SetBackModeEnabled                     bool    `json:"setBackModeEnabled"`
	RoomTemperatureControlMode             string  `json:"roomTemperatureControlMode"`
}

type ConfigurationDhw struct {
	Index                      int         `json:"index"`
	OperationModeDhw           string      `json:"operationModeDhw"`
	TappingSetpoint            float64     `json:"tappingSetpoint"`
	HolidayStartDateTime       time.Time   `json:"holidayStartDateTime"`
	HolidayEndDateTime         time.Time   `json:"holidayEndDateTime"`
	TimeProgramDhw             TimeProgram `json:"timeProgramDhw"`
	TimeProgramCirculationPump TimeProgram `json:"timeProgramCirculationPump"`
}

type ConfigurationDomesticHotWater struct {
	Index                         int         `json:"index"`
	OperationModeDomesticHotWater string      `json:"operationModeDomesticHotWater"`
	TappingSetpoint               float64     `json:"tappingSetpoint"`
	HolidayStartDateTime          time.Time   `json:"holidayStartDateTime"`
	HolidayEndDateTime            time.Time   `json:"holidayEndDateTime"`
	TimeProgramDomesticHotWater   TimeProgram `json:"timeProgramDomesticHotWater"`
	TimeProgramCirculationPump    TimeProgram `json:"timeProgramCirculationPump"`
}

type EnergyData struct {
	ExtraFields struct {
		Timezone string `json:"timezone"`
	} `json:"extra_fields"`
	OperationMode string `json:"operationMode"`
	//	SkipDataUpdate   bool    `json:"skip_data_update"`
	//	DataFrom         any     `json:"data_from"`
	//	DataTo           any     `json:"data_to"`
	StartDate  time.Time `json:"startDate"`
	EndDate    time.Time `json:"endDate"`
	Resolution string    `json:"resolution"`
	EnergyType string    `json:"energyType"`
	//	ValueType        any     `json:"valueType"`
	//	Calculated       any     `json:"calculated"`
	TotalConsumption float64 `json:"totalConsumption"`
	Data             []struct {
		ExtraFields struct {
			Timezone string `json:"timezone"`
		} `json:"extra_fields"`
		StartDate time.Time `json:"startDate"`
		EndDate   time.Time `json:"endDate"`
		Value     float64   `json:"value"`
	} `json:"data"`
}
*/
