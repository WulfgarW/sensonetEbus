package sensonetEbus

import (
	"bufio"
	"syscall"

	"errors"
	"fmt"
	"io"
	"net"
	"strconv"
	"strings"
	"time"

	"golang.org/x/exp/slices"
)

const SYSTEM_UPDATE_INTERVAL = 120

type EbusConnection struct {
	logger               Logger
	ebusdAddress         string
	ebusdConn            net.Conn
	ebusdReadBuffer      bufio.Reader
	controllerForSFMode  string
	systemUpdateInterval time.Duration
}

// NewConnection creates a new Sensonet device connection.
func newEbusConnection(ebusdAddress string, opts ...EbusConnOption) (*EbusConnection, error) {
	ebus := &EbusConnection{}
	ebus.ebusdAddress = ebusdAddress
	ebus.systemUpdateInterval = SYSTEM_UPDATE_INTERVAL * time.Second
	for _, opt := range opts {
		opt(ebus)
	}
	err := ebus.connectToEbusd()

	return ebus, err
}

func (c *EbusConnection) debug(fmt string, arg ...any) {
	if c.logger != nil {
		c.logger.Printf(fmt, arg...)
	}
}

func (c *EbusConnection) connectToEbusd() error {
	var err error
	c.ebusdConn, err = net.Dial("tcp", c.ebusdAddress)
	if err != nil {
		return err
	}
	defer c.ebusdConn.Close()
	c.ebusdReadBuffer = *bufio.NewReader(c.ebusdConn)
	scanResult := c.ebusdScanResult()
	if scanResult == "" {
		c.debug("Scan result empty")
		err = fmt.Errorf("empty scan result or error returned from ebusd: %s", scanResult)
		return err
	}
	c.debug(fmt.Sprintf("Scan result= \n%s", scanResult))
	c.controllerForSFMode = c.ebusdFindControllerForSFMode()
	if c.controllerForSFMode == "" {
		c.debug("Find result empty")
		err = fmt.Errorf("empty find %s or error returned from ebusd: %s", EBUSDREAD_HOTWATER_SFMODE, c.controllerForSFMode)
		return err
	}
	c.debug(fmt.Sprintf("Ebus Controller For SFMode= %s\n", c.controllerForSFMode))
	return err
}

func (c *EbusConnection) ebusdScanResult() string {
	var message, messageLine string
	var err error
	fmt.Fprintf(c.ebusdConn, "scan result\n")
	message = ""
	err = nil
	for err == nil {
		messageLine, err = c.ebusdReadBuffer.ReadString('\n')
		message = message + messageLine
		if err != nil || messageLine == "\n" {
			break
		}
	}
	if err != nil {
		c.debug(fmt.Sprintf("error in ebusdScanResult. error: %s", err))
	}
	return message
}

func (c *EbusConnection) ebusdFindControllerForSFMode() string {
	_, err := fmt.Fprintf(c.ebusdConn, "find "+EBUSDREAD_HOTWATER_SFMODE+"\n")
	if err != nil {
		c.debug(fmt.Sprintf("Error sending find command to ebusd: %s", err))
		return ""
	}
	var message string
	message, err = c.ebusdReadBuffer.ReadString('\n')
	if err != nil {
		c.debug(fmt.Sprintf("Error when reading from ebusd: %s", err))
		return ""
	}
	message = strings.TrimSpace(message)
	if message[:min(4, len(message))] == "ERR:" {
		c.debug(fmt.Sprintf("When trying to find controller for SFMode, ebusd answered: %s", message))
	}
	strSlices := strings.SplitAfter(message, " ")
	message = strings.TrimSpace(strSlices[0])
	return message
}

func isNetConnClosedErr(err error) bool {
	switch {
	case
		errors.Is(err, net.ErrClosed),
		errors.Is(err, io.EOF),
		errors.Is(err, syscall.EPIPE):
		return true
	default:
		if strings.Contains(err.Error(), "wsasend") {
			return true
		}
		return false
	}
}

func (c *EbusConnection) ebusdRead(searchString string, notOlderThan int) (string, error) {
	var err error
	var ebusCommand string
	if notOlderThan >= 0 {
		ebusCommand = fmt.Sprintf("read -m %0d ", notOlderThan)
	} else {
		ebusCommand = "read "
	}
	message := EBUSD_ERROR_DUMMY
	readTry := 0
	buf := c.ebusdReadBuffer

	for message[:min(4, len(message))] == "ERR:" && readTry < 3 {
		// err = c.ebusdEmptyReadBuffer()
		_, err = fmt.Fprint(c.ebusdConn, ebusCommand+searchString+"\n")
		if err != nil {
			c.debug(fmt.Sprintf("Error sending read command to ebusd: %s", err))
			if isNetConnClosedErr(err) {
				c.debug("Connection to ebusd is closed. Trying to reopen it.")
				err = c.refreshEbusdConnection()
				if err != nil {
					c.debug("refreshEbusdConnection not successful: %s", err)
					return "", err
				} else {
					_, err = fmt.Fprint(c.ebusdConn, ebusCommand+searchString+"\n")
					if err != nil {
						c.debug(fmt.Sprintf("Error sending read command to ebusd: %s", err))
						return "", err
					}
					buf = c.ebusdReadBuffer
					readTry = 0
				}
			} else {
				return "", err
			}
		}
		time.Sleep(200 * time.Millisecond) // give the ebusd a short time span before reading the answer
		message, err = buf.ReadString('\n')
		if err != nil && readTry > 1 {
			c.debug(fmt.Sprintf("Error when reading from ebusd: %s", err))
			return "", err
		}
		readTry = readTry + 1
		if readTry < 3 && message[:min(4, len(message))] == "ERR:" {
			//c.debug(fmt.Sprintf("Read try no. %d, Command: %s, ebusd answered: %s", readTry, ebusCommand+searchString, message))

			// When ebusd return an ERR: message, it returns an additional '\n'
			_, err = buf.ReadString('\n')
		}
	}
	message = strings.TrimSpace(message)
	if message[:min(4, len(message))] == "ERR:" {
		c.debug(fmt.Sprintf("Command: %s, ebusd answered: %s", ebusCommand+searchString, message))
		if message == EBUSD_ERROR_INVALIDPOSITION {
			// If ebusd returns "ERR: invalid position in decode" the message is set to ""
			message = ""
		}
		// When ebusd return an ERR: message, it returns an additional '\n'
		_, err = buf.ReadString('\n')
	}
	return message, err
}

/*func (c *EbusConnection) ebusdEmptyReadBuffer() error {
	var err error
	buf := c.ebusdReadBuffer

	for buf.Buffered() > 0 {
		fmt.Println("Bytes in Buffer:", buf.Buffered())
		_, err = buf.ReadByte()
		if err != nil {
			if errors.Is(err, io.EOF) {
				err = nil
			} else {
				c.debug(fmt.Sprintf("Error while emptying read buffer to ebusd: %s", err))
				return err
			}
		}
	}
	return nil
}*/

func (c *EbusConnection) ebusdWrite(message string) error {
	var err error
	c.ebusdConn, err = net.Dial("tcp", c.ebusdAddress)
	if err != nil {
		return err
	}
	defer c.ebusdConn.Close()
	c.ebusdReadBuffer = *bufio.NewReader(c.ebusdConn)
	_, err = fmt.Fprint(c.ebusdConn, "write "+message+"\n")
	if err != nil {
		c.debug(fmt.Sprintf("Error writing to ebusd: %s", err))
		return err
	}
	var ebusAnswer string
	ebusAnswer, err = c.ebusdReadBuffer.ReadString('\n')
	if err != nil {
		c.debug(fmt.Sprintf("Error when reading answer after ebusd write: %s", err))
		return err
	}
	c.debug(fmt.Sprintf("Command sent to ebusd: %s", "write "+message))
	c.debug(fmt.Sprintf("ebusd answered: %s", ebusAnswer))
	return err
}

func (c *EbusConnection) refreshEbusdConnection() error {
	var err error
	c.ebusdConn, err = net.Dial("tcp", c.ebusdAddress)
	if err != nil {
		return err
	}
	c.ebusdReadBuffer = *bufio.NewReader(c.ebusdConn)
	return nil
}

func (c *EbusConnection) getSystem(relData *VaillantRelData, reset bool) error {
	var err error
	var findResult string
	if !reset && time.Now().Before(relData.LastGetSystem.Add(c.systemUpdateInterval)) {
		// Use relData that are already present instead of reading current data from ebusd
		return nil
	}
	c.ebusdConn, err = net.Dial("tcp", c.ebusdAddress)
	if err != nil {
		c.debug(fmt.Sprintf("Error in net.Dial(). Error: %s\n", err))
		return err
	}
	defer c.ebusdConn.Close()
	c.ebusdReadBuffer = *bufio.NewReader(c.ebusdConn)

	// Getting Data for Hotwater
	findResult, err = c.ebusdRead(EBUSDREAD_HOTWATER_OPMODE, -1)
	if err != nil {
		c.debug(fmt.Sprintf("Error when reading '%s' from ebusd: %s. Leaving getSystem()", EBUSDREAD_HOTWATER_OPMODE, err))
		return err
	} else {
		if slices.Contains([]string{"off", "auto", "day"}, findResult) {
			relData.Hotwater.HwcOpMode = findResult
		} else {
			c.debug(fmt.Sprintf("Value '%s' returnd from ebusd for %s invalid and therefore ignored", findResult, EBUSDREAD_HOTWATER_OPMODE))
		}
	}
	findResult, err = c.ebusdRead(EBUSDREAD_HOTWATER_TEMPDESIRED, -1)
	if err != nil {
		c.debug(fmt.Sprintf("Error when reading '%s' from ebusd: %s. Leaving getSystem()", EBUSDREAD_HOTWATER_TEMPDESIRED, err))
		return err
	} else {
		convertedValue, err := convertToFloat(findResult, 0.0, 75.0)
		if err != nil {
			c.debug(fmt.Sprintf("Value '%s' returned from ebusd for %s invalid and therefore ignored. Error: %s", findResult, EBUSDREAD_HOTWATER_TEMPDESIRED, err))
		} else {
			relData.Hotwater.HwcTempDesired = convertedValue
		}
	}
	findResult, err = c.ebusdRead(EBUSDREAD_HOTWATER_STORAGETEMP, 60)
	if err != nil {
		c.debug(fmt.Sprintf("Error when reading '%s' from ebusd: %s. Leaving getSystem()", EBUSDREAD_HOTWATER_STORAGETEMP, err))
		return err
	} else {
		convertedValue, err := convertToFloat(findResult, 0.0, 75.0)
		if err != nil {
			c.debug(fmt.Sprintf("Value '%s' returned from ebusd for %s invalid and therefore ignored. Error: %s", findResult, EBUSDREAD_HOTWATER_STORAGETEMP, err))
		} else {
			relData.Hotwater.HwcStorageTemp = convertedValue
		}
	}
	findResult, err = c.ebusdRead(EBUSDREAD_HOTWATER_SFMODE, 0)
	if err != nil {
		c.debug(fmt.Sprintf("Error when reading '%s' from ebusd: %s. Leaving getSystem()", EBUSDREAD_HOTWATER_SFMODE, err))
		return err
	} else {
		if slices.Contains([]string{HWC_SFMODE_BOOST, HWC_SFMODE_NORMAL}, findResult) {
			relData.Hotwater.HwcSFMode = findResult
		} else {
			c.debug(fmt.Sprintf("Value '%s' returnd from ebusd for %s invalid and therefore ignored", findResult, EBUSDREAD_HOTWATER_SFMODE))
		}
	}

	// Getting General Status Data
	findResult, err = c.ebusdRead(EBUSDREAD_STATUS_TIME, -1)
	if err != nil {
		c.debug(fmt.Sprintf("Error when reading '%s' from ebusd: %s. Leaving getSystem()", EBUSDREAD_STATUS_TIME, err))
		return err
	} else {
		relData.Status.Time = findResult
	}
	findResult, err = c.ebusdRead(EBUSDREAD_STATUS_OUTSIDETEMPERATURE, -1)
	if err != nil {
		c.debug(fmt.Sprintf("Error when reading '%s' from ebusd: %s. Leaving getSystem()", EBUSDREAD_STATUS_OUTSIDETEMPERATURE, err))
		return err
	} else {
		relData.Status.OutsideTemperature, _ = strconv.ParseFloat(findResult, 64)
	}
	findResult, err = c.ebusdRead(EBUSDREAD_STATUS_SYSTEMFLOWTEMPERATUE, -1)
	if err != nil {
		c.debug(fmt.Sprintf("Error when reading '%s' from ebusd: %s. Leaving getSystem()", EBUSDREAD_STATUS_SYSTEMFLOWTEMPERATUE, err))
		return err
	} else {
		relData.Status.SystemFlowTemperature, _ = strconv.ParseFloat(findResult, 64)
	}
	findResult, err = c.ebusdRead(EBUSDREAD_STATUS_WATERPRESSURE, -1)
	if err != nil {
		c.debug(fmt.Sprintf("Error when reading '%s' from ebusd: %s. Leaving getSystem()", EBUSDREAD_STATUS_WATERPRESSURE, err))
		return err
	} else {
		convertedValue, err := convertToFloat(findResult, 0.0, 5.0)
		if err != nil {
			c.debug(fmt.Sprintf("Value '%s' returned from ebusd for %s invalid and therefore ignored. Error: %s", findResult, EBUSDREAD_STATUS_WATERPRESSURE, err))
		} else {
			relData.Status.WaterPressure = convertedValue
		}
	}
	findResult, err = c.ebusdRead(EBUSDREAD_STATUS_CURRENTCONSUMEDPOWER, 60)
	if err != nil {
		c.debug(fmt.Sprintf("Error when reading '%s' from ebusd: %s. Leaving getSystem()", EBUSDREAD_STATUS_WATERPRESSURE, err))
		return err
	} else {
		convertedValue, err := convertToFloat(findResult, 0.0, 30.0)
		if err != nil {
			c.debug(fmt.Sprintf("Value '%s' returned from ebusd for %s invalid and therefore ignored. Error: %s", findResult, EBUSDREAD_STATUS_CURRENTCONSUMEDPOWER, err))
		} else {
			relData.Status.CurrentConsumedPower = convertedValue
		}
	}
	findResult, err = c.ebusdRead(EBUSDREAD_STATUS_STATUS01, -1)
	if err != nil {
		c.debug(fmt.Sprintf("Error when reading '%s' from ebusd: %s. Leaving getSystem()", EBUSDREAD_STATUS_STATUS01, err))
		return err
	} else {
		relData.Status.Status01 = findResult
	}
	findResult, err = c.ebusdRead(EBUSDREAD_STATUS_STATE, -1)
	if err != nil {
		c.debug(fmt.Sprintf("Error when reading '%s' from ebusd: %s. Leaving getSystem()", EBUSDREAD_STATUS_STATE, err))
		return err
	} else {
		relData.Status.State = findResult
	}

	// Getting Zone Data
	if len(relData.Zones) == 0 {
		relData.Zones = make([]VaillantRelDataZones, NUMBER_OF_ZONES_TO_READ)
	}
	for i := 0; i < NUMBER_OF_ZONES_TO_READ && err == nil; i++ {
		err = c.getZoneDataFromEbus(&relData.Zones[i], i+1)
	}

	// Set timestamp lastGetSystemAt and return nil error
	relData.LastGetSystem = time.Now()
	return nil
}

func (c *EbusConnection) getZoneDataFromEbus(zoneData *VaillantRelDataZones, heatingZone int) error {
	zonePrefix := fmt.Sprintf("z%01d", heatingZone)
	zoneData.Index = heatingZone
	findResult, err := c.ebusdRead(zonePrefix+EBUSDREAD_ZONE_OPMODE, -1)
	if err != nil {
		c.debug(fmt.Sprintf("Error when reading '%s' from ebusd: %s. Leaving getSystem()", EBUSDREAD_ZONE_OPMODE, err))
		return err
	} else {
		if slices.Contains([]string{"off", "auto", "day"}, findResult) {
			zoneData.OpMode = findResult
		} else {
			c.debug(fmt.Sprintf("Value '%s' returned from ebusd for %s invalid and therefore ignored", findResult, zonePrefix+EBUSDREAD_ZONE_OPMODE))
		}
	}
	findResult, err = c.ebusdRead(zonePrefix+EBUSDREAD_ZONE_SFMODE, 0)
	if err != nil {
		c.debug(fmt.Sprintf("Error when reading '%s' from ebusd: %s. Leaving getSystem()", EBUSDREAD_ZONE_SFMODE, err))
		return err
	} else {
		if slices.Contains([]string{"auto", "veto"}, findResult) {
			zoneData.SFMode = findResult
		} else {
			c.debug(fmt.Sprintf("Value '%s' returned from ebusd for %s invalid and therefore ignored", findResult, zonePrefix+EBUSDREAD_ZONE_SFMODE))
		}
	}
	findResult, err = c.ebusdRead(zonePrefix+EBUSDREAD_ZONE_ACTUALROOMTEMPDESIRED, -1)
	if err != nil {
		c.debug(fmt.Sprintf("Error when reading '%s' from ebusd: %s. Leaving getSystem()", EBUSDREAD_ZONE_ACTUALROOMTEMPDESIRED, err))
		return err
	} else {
		convertedValue, err := convertToFloat(findResult, 0.0, 50.0)
		if err != nil {
			c.debug(fmt.Sprintf("Value '%s' returned from ebusd for %s invalid and therefore ignored. Error: %s", findResult, zonePrefix+EBUSDREAD_ZONE_ACTUALROOMTEMPDESIRED, err))
		} else {
			zoneData.ActualRoomTempDesired = convertedValue
		}
	}
	findResult, err = c.ebusdRead(zonePrefix+EBUSDREAD_ZONE_ROOMTEMP, 180)
	if err != nil {
		c.debug(fmt.Sprintf("Error when reading '%s' from ebusd: %s. Leaving getSystem()", EBUSDREAD_ZONE_ROOMTEMP, err))
		return err
	} else {
		convertedValue, err := convertToFloat(findResult, 0.0, 50.0)
		if err != nil {
			c.debug(fmt.Sprintf("Value '%s' returned from ebusd for %s invalid and therefore ignored. Error: %s", findResult, zonePrefix+EBUSDREAD_ZONE_ROOMTEMP, err))
		} else {
			zoneData.RoomTemp = convertedValue
		}
	}
	findResult, err = c.ebusdRead(zonePrefix+EBUSDREAD_ZONE_QUICKVETOTEMP, 0)
	if err != nil {
		c.debug(fmt.Sprintf("Error when reading '%s' from ebusd: %s. Leaving getSystem()", EBUSDREAD_ZONE_QUICKVETOTEMP, err))
		return err
	} else {
		convertedValue, err := convertToFloat(findResult, 0.0, 50.0)
		if err != nil {
			c.debug(fmt.Sprintf("Value '%s' returned from ebusd for %s invalid and therefore ignored. Error: %s", findResult, zonePrefix+EBUSDREAD_ZONE_QUICKVETOTEMP, err))
		} else {
			zoneData.QuickVetoTemp = convertedValue
		}
	}
	findResult, err = c.ebusdRead(zonePrefix+EBUSDREAD_ZONE_QUICKVETOENDDATE, -1)
	if err != nil {
		c.debug(fmt.Sprintf("Error when reading '%s' from ebusd: %s. Leaving getSystem()", EBUSDREAD_ZONE_QUICKVETOENDDATE, err))
		return err
	} else {
		zoneData.QuickVetoEndDate = findResult
	}
	findResult, err = c.ebusdRead(zonePrefix+EBUSDREAD_ZONE_QUICKVETOENDTIME, -1)
	if err != nil {
		c.debug(fmt.Sprintf("Error when reading '%s' from ebusd: %s. Leaving getSystem()", EBUSDREAD_ZONE_QUICKVETOENDTIME, err))
		return err
	} else {
		zoneData.QuickVetoEndTime = findResult
	}
	findResult, err = c.ebusdRead(zonePrefix+EBUSDREAD_ZONE_SHORTNAME, -1)
	if err != nil {
		c.debug(fmt.Sprintf("Error when reading '%s' from ebusd: %s. Leaving getSystem()", EBUSDREAD_ZONE_SHORTNAME, err))
		return err
	} else {
		zoneData.ShortName = findResult
	}
	findResult, err = c.ebusdRead(zonePrefix+EBUSDREAD_ZONE_NAME1, -1)
	if err != nil {
		c.debug(fmt.Sprintf("Error when reading '%s' from ebusd: %s. Leaving getSystem()", EBUSDREAD_ZONE_NAME1, err))
		return err
	} else {
		zoneData.Name1 = findResult
	}
	findResult, err = c.ebusdRead(zonePrefix+EBUSDREAD_ZONE_NAME2, -1)
	if err != nil {
		c.debug(fmt.Sprintf("Error when reading '%s' from ebusd: %s. Leaving getSystem()", EBUSDREAD_ZONE_NAME2, err))
		return err
	} else {
		zoneData.Name2 = findResult
	}

	return err
}

// checkEbusdConfig() tries to read all elements that are used in the package to check if the configuration of the ebusd supports them
func (c *EbusConnection) checkEbusdConfig() (string, error) {
	var err error
	var findResult, details string
	c.debug("In checkEbusConfig()")
	errElementNotFound := false
	c.ebusdConn, err = net.Dial("tcp", c.ebusdAddress)
	if err != nil {
		details = details + fmt.Sprintf("Error in net.Dial(). Error: %s\n", err)
		c.debug(fmt.Sprintf("Error in net.Dial(). Error: %s\n", err))
		return details, err
	}
	defer c.ebusdConn.Close()
	c.ebusdReadBuffer = *bufio.NewReader(c.ebusdConn)

	// Getting Data for Hotwater
	for _, what := range []string{EBUSDREAD_HOTWATER_OPMODE, EBUSDREAD_HOTWATER_TEMPDESIRED, EBUSDREAD_HOTWATER_STORAGETEMP, EBUSDREAD_HOTWATER_SFMODE} {
		findResult, err = c.ebusdRead(what, -1)
		if err != nil {
			details += c.setDetailsAndWriteDebugMessage(what, findResult, err)
		}
		if findResult == EBUSD_ERROR_ELEMENTNOTFOUND {
			errElementNotFound = true
		}
	}

	// Getting General Status Data
	for _, what := range []string{EBUSDREAD_STATUS_TIME, EBUSDREAD_STATUS_OUTSIDETEMPERATURE, EBUSDREAD_STATUS_SYSTEMFLOWTEMPERATUE, EBUSDREAD_STATUS_WATERPRESSURE,
		EBUSDREAD_STATUS_CURRENTCONSUMEDPOWER, EBUSDREAD_STATUS_STATUS01, EBUSDREAD_STATUS_STATE} {
		findResult, err = c.ebusdRead(what, -1)
		if err != nil {
			details += c.setDetailsAndWriteDebugMessage(what, findResult, err)
		}
		if findResult == EBUSD_ERROR_ELEMENTNOTFOUND {
			errElementNotFound = true
		}
	}

	// Getting Zone Data
	for i := 0; i < NUMBER_OF_ZONES_TO_READ; i++ {
		zonePrefix := fmt.Sprintf("z%01d", i+1)
		for _, what := range []string{EBUSDREAD_ZONE_OPMODE, EBUSDREAD_ZONE_SFMODE, EBUSDREAD_ZONE_ACTUALROOMTEMPDESIRED, EBUSDREAD_ZONE_ROOMTEMP,
			EBUSDREAD_ZONE_QUICKVETOTEMP, EBUSDREAD_ZONE_QUICKVETOENDTIME, EBUSDREAD_ZONE_QUICKVETOENDDATE, EBUSDREAD_ZONE_QUICKVETODURATION,
			EBUSDREAD_ZONE_NAME1, EBUSDREAD_ZONE_NAME2, EBUSDREAD_ZONE_SHORTNAME} {
			findResult, err = c.ebusdRead(zonePrefix+what, -1)
			if err != nil || findResult[:min(4, len(findResult))] == "ERR:" {
				details += c.setDetailsAndWriteDebugMessage(zonePrefix+what, findResult, err)
			}
			if findResult == EBUSD_ERROR_ELEMENTNOTFOUND {
				errElementNotFound = true
			}
		}
	}
	if errElementNotFound {
		err = fmt.Errorf("Some ebus read commands got %q", EBUSD_ERROR_ELEMENTNOTFOUND)
	}
	c.debug("End of checkEbusConfig()")
	return details, err
}

func convertToFloat(rawResult string, min, max float64) (float64, error) {
	if rawResult == "-" {
		return 0.0, nil
	} else {
		convertedValue, err := strconv.ParseFloat(rawResult, 64)
		if err == nil {
			if convertedValue < min || convertedValue > max {
				err = fmt.Errorf("converted value is not in range [%.2f,%.2f]", min, max)
			}
		}
		return convertedValue, err
	}
}

func (c *EbusConnection) setDetailsAndWriteDebugMessage(what, result string, err error) string {
	if err != nil {
		c.debug(fmt.Sprintf("Value '%s' returnd from ebusd for %s. Error: %s", result, what, err))
		return fmt.Sprintf("Trying to read %s. Error: %s\n", what, err)
	}
	c.debug(fmt.Sprintf("Value '%s' returnd from ebusd for %s.", result, what))
	return fmt.Sprintf("Trying to read %s. Result: %s\n", what, result)
}
