package sensonetEbus

import (
	"fmt"
	"time"
)

// Connection is the SensonetEbus connection
type Connection struct {
	logger           Logger
	ebusdConn        *EbusConnection
	currentQuickmode string
	quickmodeStarted time.Time
	quickmodeStopped time.Time
	relData          VaillantRelData
}

// NewConnection creates a new Sensonet device connection.
func NewConnection(ebusdAddress string, opts ...ConnOption) (*Connection, error) {
	conn := &Connection{}
	conn.currentQuickmode = ""
	conn.quickmodeStarted = time.Now()

	for _, opt := range opts {
		opt(conn)
	}

	var err error
	if conn.logger != nil {
		conn.ebusdConn, err = newEbusConnection(ebusdAddress, withConnLogger(conn.logger))
	} else {
		conn.ebusdConn, err = newEbusConnection(ebusdAddress)
	}
	return conn, err
}

func (c *Connection) debug(fmt string, arg ...any) {
	if c.logger != nil {
		c.logger.Printf(fmt, arg...)
	}
}

func (c *Connection) GetCurrentQuickMode() string {
	return c.currentQuickmode
}

func (c *Connection) GetSystem(refresh bool) (VaillantRelData, error) {
	err := c.ebusdConn.getSystem(&c.relData, refresh)
	c.refreshCurrentQuickMode()
	return c.relData, err
}

func (c *Connection) CheckEbusdConfig() (string, error) {
	details, err := c.ebusdConn.checkEbusdConfig()
	return details, err
}

func (c *Connection) StartZoneQuickVeto(zone int, setpoint float32, duration float32) error {
	if zone < 0 {
		zone = ZONEINDEX_DEFAULT
	} // if parameter "zone" is negative, then the default value is used
	if setpoint < 0.0 {
		setpoint = ZONEVETOSETPOINT_DEFAULT
	} // if parameter "setpoint" is negative, then the default value is used
	if duration < 0.0 {
		duration = ZONEVETODURATION_DEFAULT
	} // if parameter "duration" is negative, then the default value is used

	zonePrefix := fmt.Sprintf("z%01d", zone)
	message := " -c " + c.ebusdConn.controllerForSFMode + " " + zonePrefix + EBUSDREAD_ZONE_QUICKVETOTEMP + fmt.Sprintf(" %2.1f", setpoint)
	err := c.ebusdConn.ebusdWrite(message)
	if err != nil {
		c.debug("could not start zone quick veto. Error: %s", err)
		return err
	}
	// Zone quick veto is started by writing a duration to the controler. A duration of 0.5 hours is set.
	message = " -c " + c.ebusdConn.controllerForSFMode + " " + zonePrefix + EBUSDREAD_ZONE_QUICKVETODURATION + fmt.Sprintf(" %2.1f", duration)
	err = c.ebusdConn.ebusdWrite(message)
	if err != nil {
		c.debug(fmt.Sprintf("could not start zone quick veto. Error: %s", err))
		return err
	}
	c.relData.LastGetSystem = time.Time{} // reset the cache
	return err
}

func (c *Connection) StopZoneQuickVeto(zone int) error {
	if zone < 0 {
		zone = ZONEINDEX_DEFAULT
	} // if parameter "zone" is negative, then the default value is used

	zonePrefix := fmt.Sprintf("z%01d", zone)
	message := " -c " + c.ebusdConn.controllerForSFMode + " " + zonePrefix + EBUSDREAD_ZONE_SFMODE + " auto"
	err := c.ebusdConn.ebusdWrite(message)
	if err != nil {
		c.debug(fmt.Sprintf("could not stop zone quick veto. Error: %s", err))
		return err
	}
	c.relData.LastGetSystem = time.Time{} // reset the cache
	return err
}

func (c *Connection) StartHotWaterBoost() error {
	message := " -c " + c.ebusdConn.controllerForSFMode + " " + EBUSDREAD_HOTWATER_SFMODE + " load"
	err := c.ebusdConn.ebusdWrite(message)
	if err != nil {
		c.debug(fmt.Sprintf("could not start hotwater boost. Error: %s", err))
	}
	c.relData.LastGetSystem = time.Time{} // reset the cache
	return err
}

func (c *Connection) StopHotWaterBoost() error {
	message := " -c " + c.ebusdConn.controllerForSFMode + " " + EBUSDREAD_HOTWATER_SFMODE + " auto"
	err := c.ebusdConn.ebusdWrite(message)
	if err != nil {
		c.debug(fmt.Sprintf("could not start hotwater boost. Error: %s", err))
	}
	c.relData.LastGetSystem = time.Time{} // reset the cache
	return err
}

func (c *Connection) refreshCurrentQuickMode() {
	newQuickMode := ""
	if c.relData.Hotwater.HwcSFMode == HWC_SFMODE_BOOST {
		newQuickMode = QUICKMODE_HOTWATER
	}
	for _, zone := range c.relData.Zones {
		if zone.SFMode == ZONE_SFMODE_BOOST {
			newQuickMode = QUICKMODE_HEATING
			break
		}
	}
	if newQuickMode != c.currentQuickmode {
		if newQuickMode == "" {
			if c.currentQuickmode == QUICKMODE_NOTHING && time.Now().Before(c.quickmodeStarted.Add(10*time.Minute)) {
				c.debug("Idle mode active for less then 10 minutes. Keeping the idle mode")
			} else {
				c.debug(fmt.Sprintf("Old quickmode: \"%s\"   New quickmode: \"%s\"", c.currentQuickmode, newQuickMode))
				c.currentQuickmode = newQuickMode
				c.quickmodeStopped = time.Now()
			}
		}
		if newQuickMode != "" {
			c.debug(fmt.Sprintf("Old quickmode: \"%s\"   New quickmode: \"%s\"", c.currentQuickmode, newQuickMode))
			c.currentQuickmode = newQuickMode
			c.quickmodeStarted = time.Now()
		}
	}
}

func (c *Connection) StartStrategybased(strategy int, heatingPar *HeatingParStruct) (string, error) {
	err := c.ebusdConn.getSystem(&c.relData, true)
	if err != nil {
		err = fmt.Errorf("could not read current status information in StartStrategybased(): %s", err)
		return "", err
	}
	c.refreshCurrentQuickMode()

	// Extracting correct Zones element
	zoneData := GetZoneData(c.relData.Zones, heatingPar.ZoneIndex)
	if c.currentQuickmode != "" {
		c.debug(fmt.Sprint("System is already in quick mode:", c.currentQuickmode))
		c.debug("Is there any need to change that?")
		c.debug(fmt.Sprint("Special Function of Dhw: ", c.relData.Hotwater.HwcSFMode))
		c.debug(fmt.Sprint("Special Function of Heating Zone: ", zoneData.SFMode))
		return QUICKMODE_ERROR_ALREADYON, err
	}
	whichQuickMode := c.WhichQuickMode(strategy, heatingPar.ZoneIndex)
	c.debug(fmt.Sprint("whichQuickMode=", whichQuickMode))

	switch whichQuickMode {
	case 1:
		err = c.StartHotWaterBoost()
		if err == nil {
			c.currentQuickmode = QUICKMODE_HOTWATER
			c.quickmodeStarted = time.Now()
			c.debug("Starting hotwater boost")
		}
	case 2:
		err = c.StartZoneQuickVeto(heatingPar.ZoneIndex, heatingPar.VetoSetpoint, heatingPar.VetoDuration)
		if err == nil {
			c.currentQuickmode = QUICKMODE_HEATING
			c.quickmodeStarted = time.Now()
			c.debug("Starting zone quick veto")
		}
	default:
		if c.currentQuickmode == QUICKMODE_HOTWATER {
			// if hotwater boost active, then stop it
			err = c.StopHotWaterBoost()
			if err == nil {
				c.debug("Stopping hotwater boost")
			}
		}
		if c.currentQuickmode == QUICKMODE_HEATING {
			// if zone quick veto active, then stop it
			err = c.StopZoneQuickVeto(heatingPar.ZoneIndex)
			if err == nil {
				c.debug("Stopping zone quick veto")
			}
		}
		c.currentQuickmode = QUICKMODE_NOTHING
		c.quickmodeStarted = time.Now()
		c.debug("Enable called but no quick mode possible. Starting idle mode")
	}

	c.relData.LastGetSystem = time.Time{} // reset the cache
	return c.currentQuickmode, err
}

func (c *Connection) StopStrategybased(heatingPar *HeatingParStruct) (string, error) {
	err := c.ebusdConn.getSystem(&c.relData, true)
	if err != nil {
		err = fmt.Errorf("could not read current status information in StopStrategybased(): %s", err)
		return "", err
	}
	c.refreshCurrentQuickMode()

	// Extracting correct Zones element
	zoneData := GetZoneData(c.relData.Zones, heatingPar.ZoneIndex)
	c.debug(fmt.Sprint("Operationg Mode of Dhw: ", c.relData.Hotwater.HwcSFMode))
	c.debug(fmt.Sprint("Operationg Mode of Heating: ", zoneData.SFMode))
	switch c.currentQuickmode {
	case QUICKMODE_HOTWATER:
		err = c.StopHotWaterBoost()
		if err == nil {
			c.debug(fmt.Sprint("Stopping quick mode", c.currentQuickmode))
		}
	case QUICKMODE_HEATING:
		err = c.StopZoneQuickVeto(heatingPar.ZoneIndex)
		if err == nil {
			c.debug("Stopping zone quick veto")
		}
	case QUICKMODE_NOTHING:
		c.debug("Stopping idle quick mode")
	default:
		c.debug("Nothing to do, no quick mode active")
	}
	c.currentQuickmode = ""
	c.quickmodeStopped = time.Now()

	c.relData.LastGetSystem = time.Time{} // reset the cache
	return c.currentQuickmode, err
}

// This function checks the operation mode of heating and hotwater and the hotwater live temperature
// and returns, which quick mode should be started, when StartStrategybased() is called
func (c *Connection) WhichQuickMode(strategy, heatingZone int) int {
	/*err := c.ebusdConn.getSystem(&c.relData, false)
	if err != nil {
		err = fmt.Errorf("could not read current status information in WhichQuickMode(): %s", err)
		return 0, err
	}*/
	c.debug(fmt.Sprintf("Checking if hot water boost possible. Operation Mode = %s, temperature setpoint= %02.2f, live temperature= %02.2f",
		c.relData.Hotwater.HwcOpMode, c.relData.Hotwater.HwcTempDesired, c.relData.Hotwater.HwcStorageTemp))
	hotWaterBoostPossible := false
	// For strategy=STRATEGY_HOTWATER, a hotwater boost is possible when hotwater storage temperature is less than the temperature setpoint.
	// For other strategy values, a hotwater boost is possible when hotwater storage temperature is less than the temperature setpoint minus 5°C
	addOn := -5.0
	if strategy == STRATEGY_HOTWATER {
		addOn = 0.0
	}
	if c.relData.Hotwater.HwcStorageTemp < c.relData.Hotwater.HwcTempDesired+addOn &&
		c.relData.Hotwater.HwcOpMode == OPERATIONMODE_AUTO {
		hotWaterBoostPossible = true
	}

	heatingQuickVetoPossible := false
	for _, z := range c.relData.Zones {
		if z.Index == heatingZone {
			c.debug(fmt.Sprintf("Checking if heating quick veto possible. Operation Mode = %s", z.OpMode))
			if z.OpMode == OPERATIONMODE_AUTO {
				heatingQuickVetoPossible = true
			}
		}
	}

	whichQuickMode := 0
	switch strategy {
	case STRATEGY_HOTWATER:
		if hotWaterBoostPossible {
			whichQuickMode = 1
		} else {
			c.debug("Strategy = hotwater, but hotwater boost not possible")
		}
	case STRATEGY_HEATING:
		if heatingQuickVetoPossible {
			whichQuickMode = 2
		} else {
			c.debug("Strategy = heating, but heating quick veto not possible")
		}
	case STRATEGY_HOTWATER_THEN_HEATING:
		if hotWaterBoostPossible {
			whichQuickMode = 1
		} else {
			if heatingQuickVetoPossible {
				whichQuickMode = 2
			} else {
				c.debug("PV Use Strategy = hotwater_then_heating, but both not possible")
			}
		}
	}
	return whichQuickMode
}

// Returns the energy data for systemId, deviceUuid and other given criteria
/*func (c *Connection) GetEnergyData(systemId, deviceUuid, operationMode, energyType, resolution string, startDate, endDate time.Time) (EnergyData, error) {
	var energyData EnergyData
	v := url.Values{
		"resolution":    {resolution},
		"operationMode": {operationMode},
		"energyType":    {energyType},
		"startDate":     {startDate.Format("2006-01-02T15:04:05-07:00")},
		"endDate":       {endDate.Format("2006-01-02T15:04:05-07:00")},
	}

	url := API_URL_BASE + fmt.Sprintf(ENERGY_URL, systemId, deviceUuid) + v.Encode()
	req, _ := http.NewRequest("GET", url, nil)
	if err := doJSON(c.client, req, &energyData); err != nil {
		return energyData, err
	}
	return energyData, nil
}*/

// Returns the current power consumption for systemId
func (c *Connection) GetSystemCurrentPower() (float64, error) {
	state, err := c.GetSystem(false)
	if err != nil {
		return -1.0, err
	}
	return state.Status.CurrentConsumedPower, err
}
