package main

// This is an example program to demonstrate the usage of the package.

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/WulfgarW/sensonetEbus"

	"github.com/eiannone/keyboard"
)

const LOG_FILE = "sensonet.log"
const PAR_FILE = ".sensonet-par.json"
const SYSTEM_FILE = ".sensonet-system.json"
const WITH_SENSONET_EBUS_LOGGING = true // Set this to false if you want no sensonetEbus logging

func readPars(filename string) (*sensonetEbus.ParsStruct, error) {
	b, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	var pars *sensonetEbus.ParsStruct
	err = json.Unmarshal(b, &pars)
	return pars, err
}

func writeSystemInfo(filename string, system *sensonetEbus.VaillantRelData) error {
	b, err := json.MarshalIndent(system, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filename, b, 0o644)
}

func readKey(input chan rune) {
	for {
		char, _, err := keyboard.GetSingleKey()
		if err != nil {
			log.Fatal(err)
		}
		input <- char
	}
}

func printKeyBinding() {
	fmt.Println("#############################################")
	fmt.Println("Choose an action:")
	//	fmt.Println("   1 = Read device and energy data")
	fmt.Println("   4 = Start hotwater boost")
	fmt.Println("   5 = Start zone quick veto")
	fmt.Println("   6 = Start strategy based quick mode")
	fmt.Println("   7 = Stop hotwater boost")
	fmt.Println("   8 = Stop zone quick veto")
	fmt.Println("   9 = Stop strategy based quick mode")
	fmt.Println("   0 = Read current power consumption")
	fmt.Println("   s = Write system info to file")
	fmt.Println("   h = Show key bindings")
	fmt.Println("   q = Quit")
	fmt.Println("#############################################")
	fmt.Println("")
}

// Implementation of log functions for the logger interface of the sensonet library
// (not necessary, if you don't want to use the logger interface)
type SLogger struct {
	logger *log.Logger
}

func NewSLogLogger(logFile *os.File) *SLogger {
	logger := log.New(logFile, "sensonetlogger: ", log.Lshortfile)
	return &SLogger{logger: logger}
}

func (l *SLogger) Printf(msg string, arg ...any) {
	logMsg := fmt.Sprint("Debug: ", msg)
	for _, oneArg := range arg {
		logMsg = fmt.Sprint(logMsg, oneArg)
	}
	l.logger.Println(logMsg)
}

// Main program
func main() {
	var logFile *os.File
	var err error
	if LOG_FILE != "" {
		logFile, err = os.OpenFile(LOG_FILE, os.O_CREATE|os.O_WRONLY, 0666)
		if err != nil {
			fmt.Println("Error opening log file. Err %w", err)
			os.Exit(1)
		}
	} else {
		logFile = os.Stderr
	}

	var (
		logger = log.New(logFile, "sensonet: ", log.Lshortfile)
	)

	fmt.Println("Sample program to show how to use the sensonetEbus library functions.")
	fmt.Println("")
	fmt.Println("")
	fmt.Println("First step: Reading parameter file")
	// Read paramters from file
	pars, err := readPars(PAR_FILE)
	if err != nil {
		logger.Fatalln("readPars() ended unsuccessful. Probably no parameter file was found. Error:", err)
	}

	fmt.Println("Second step: Generating new connection to be used for further calls of sensonetEbus library")

	// NewConnection() opens the connection to the myVaillant portal and returns a connection object for further function calls.
	// You can provide an http client (especially one with logging) as optional parameter.
	var conn *sensonetEbus.Connection
	if WITH_SENSONET_EBUS_LOGGING {
		// Implements a logger for the sensonet library
		slogger := NewSLogLogger(logFile)
		conn, err = sensonetEbus.NewConnection(pars.EbusdAddress, sensonetEbus.WithLogger(slogger))
	} else {
		conn, err = sensonetEbus.NewConnection(pars.EbusdAddress)
	}
	if err != nil {
		logger.Fatal(err)
	}

	// CheckEbusConfiguration() tries all elements that are used in the package to check if the configuration of the ebusd supports them
	fmt.Println("Third step: Check eBus configuration")
	details, err := conn.CheckEbusdConfig()
	if err == nil {
		fmt.Println("   CheckEbusdConfig() returned no error")
	} else {
		fmt.Printf("   CheckEbusdConfig() returned: Details: \n%s , (Last) Error: %s \n", details, err)
	}

	var heatingPar sensonetEbus.HeatingParStruct
	heatingPar.ZoneIndex = 1
	heatingPar.VetoSetpoint = 18.0
	heatingPar.VetoDuration = -1.0 // negative value means: use default

	// Create a channel to read, if a key was pressed
	if err := keyboard.Open(); err != nil {
		panic(err)
	}
	input := make(chan rune, 1)
	go readKey(input)

	// Print menu and wait for key strokes"
	fmt.Println()
	printKeyBinding()
	lastPrint := time.Now().Add(-25 * time.Second)

	for {
		select {
		case i := <-input:
			switch {
			/*case i == rune('1'):
			fmt.Println("Getting device data")
			devices, err := ctrl.GetDeviceData(systemId, sensonet.DEVICES_ALL)
			if err != nil {
				logger.Println(err)
			}
			fmt.Printf("   Got %d devices\n ", len(devices))
			fmt.Println("Reading energy data")
			startDate, _ := time.Parse("2006-01-02 15:04:05MST", "2025-01-01 00:00:00CET")
			endDate, _ := time.Parse("2006-01-02 15:04:05MST", "2025-01-10 23:59:59CET")
			for _, dev := range devices {
				for _, data := range dev.Device.Data {
					energyData, err := ctrl.GetEnergyData(systemId, dev.Device.DeviceUUID, data.OperationMode, data.ValueType, sensonet.RESOLUTION_DAY,
						startDate, endDate)
					if err != nil {
						logger.Println(err)
					} else {
						fmt.Printf("   Energy data for %s, %s, %s:\n", dev.Device.ProductName, data.OperationMode, data.ValueType)
						fmt.Printf("      %s bis %s: %.2f kWh\n", energyData.StartDate.Format("02.01.2006 15:04 MST"),
							energyData.EndDate.Format("02.01.2006 15:04 MST"), energyData.TotalConsumption/1000)
					}

				}
			}*/
			case i == rune('4'):
				fmt.Println("Starting hotwater boost")
				logger.Println("Starting hotwater boost")
				err = conn.StartHotWaterBoost()
				if err != nil {
					fmt.Println(" An error occurred. ", err)
					logger.Println(err)
				}
			case i == rune('5'):
				fmt.Println("Starting zone quick veto")
				logger.Println("Starting zone quick veto")
				err = conn.StartZoneQuickVeto(heatingPar.ZoneIndex, heatingPar.VetoSetpoint, heatingPar.VetoDuration)
				if err != nil {
					fmt.Println(" An error occurred. ", err)
					logger.Println(err)
				}
			case i == rune('6'):
				fmt.Println("Starting strategy based session")
				logger.Println("Starting strategy based session")
				result, err := conn.StartStrategybased(sensonetEbus.STRATEGY_HOTWATER_THEN_HEATING, &heatingPar)
				if err != nil {
					fmt.Println(" An error occurred. ", err)
					logger.Println(err)
				} else {
					fmt.Printf("result of function StartStrategybased()=\"%s\"\n", result)
				}
			case i == rune('7'):
				fmt.Println("Stopping hotwater boost")
				logger.Println("Stopping hotwater boost")
				err = conn.StopHotWaterBoost()
				if err != nil {
					fmt.Println(" An error occurred. ", err)
					logger.Println(err)
				}
			case i == rune('8'):
				fmt.Println("Stopping zone quick veto")
				logger.Println("Stopping zone quick veto")
				err = conn.StopZoneQuickVeto(heatingPar.ZoneIndex)
				if err != nil {
					fmt.Println(" An error occurred. ", err)
					logger.Println(err)
				}
			case i == rune('9'):
				fmt.Println("Stopping strategy based session")
				logger.Println("Stopping strategy based session")
				result, err := conn.StopStrategybased(&heatingPar)
				if err != nil {
					fmt.Println(" An error occurred. ", err)
					logger.Println(err)
				} else {
					fmt.Printf("result of function StopStrategybased()=\"%s\"\n", result)
				}
			case i == rune('0'):
				fmt.Println("Getting current power consumption")
				logger.Println("Getting current power consumption")
				result, err := conn.GetSystemCurrentPower()
				if err != nil {
					fmt.Println(" An error occurred. ", err)
					logger.Println(err)
				}
				fmt.Println("Current power consumption of system=", result)
			case i == rune('s'):
				fmt.Println("Writing system info to file")
				logger.Println("Writing system info to file")
				result, err := conn.GetSystem(true)
				if err != nil {
					fmt.Println(" An error occurred. ", err)
					logger.Println(err)
				} else {
					// Store the system device information in a file
					if err := writeSystemInfo(SYSTEM_FILE, &result); err != nil {
						logger.Fatal(err)
					}
				}
			case i == rune('h'):
				printKeyBinding()
			case i == rune('q'):
				_ = keyboard.Close()
				os.Exit(0)
			default:
				fmt.Println("You pressed a key without a function. Press h to get help")
			}
		default:
			// No key pressed. Print some information every 30 seconds
			if time.Now().After(lastPrint.Add(30 * time.Second)) {
				logger.Printf("It is %s. Calling Get system and print results.", time.Now().Format("15:04:05"))
				state, err := conn.GetSystem(false) // 'false' means, that data from cache can be used
				if err != nil {
					logger.Fatal(err)
				}
				b, err := json.MarshalIndent(state, "", "  ")
				fmt.Printf("%s", b)
				fmt.Println("---------------------------------------------------------------------------------------------------------------------")
				fmt.Printf("   OutdoorTemperature: %.1f°C\n", state.Status.OutsideTemperature)
				fmt.Print("   Zones: ")
				for _, z := range state.Zones {
					fmt.Printf("\"%s\":%.1f°C (Setpoint=%.1f°C), ", z.Name1+z.Name2, z.RoomTemp, z.ActualRoomTempDesired)
				}
				fmt.Println("")
				fmt.Printf("   HotWaterTemperature: %.1f°C (Setpoint=%.1f°C)\n", state.Hotwater.HwcStorageTemp, state.Hotwater.HwcTempDesired)
				zoneData := sensonetEbus.GetZoneData(state.Zones, heatingPar.ZoneIndex)
				quickModeExpiresAt := conn.GetQuickModeExpiresAt()
				if quickModeExpiresAt == "" {
					quickModeExpiresAt = "(unknown)"
				}
				fmt.Printf("   Quickmodes: internal: \"%s\" until %s. Heat pump: Dhw: \"%s\"  Zone: \"%s\"\n", conn.GetCurrentQuickMode(), quickModeExpiresAt, state.Hotwater.HwcSFMode, zoneData.SFMode)
				fmt.Println("---------------------------------------------------------------------------------------------------------------------")
				lastPrint = time.Now()

			}
		}
	}

}
