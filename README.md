# sensonetEbus

sensonetEbus is a library that provides functions to read data from Vaillant heating systems, especially heat pumps, and to initiate certain routines on these systems.
The communication works via ebus and mainly reads from the sensonet module (VR921). So you need a Vaillant heating system with a VR921 module, an ebus adapter (see https://adapter.ebusd.eu/v5-c6/) and the ebusd (see https://github.com/john30/ebusd). 
(Presumably the library also works with a Vaillant VR940f module instead of a VR921 module.)

## Features
- Reading the system information of the heating system (current temperatures and setpoints for hotwater and heating zones, current power consumption)
- Starting and stopping of hotwater boosts and of zone quick veto sessions
- Starting and stopping of strategy based quick mode sessions

## Custom ebus message definition file 15.ctlv2.csv
At the moment, some ebus message definitions needed for the initiation of a zone quick veto and for the current power consumption of the immersion heater are missing in the "official" ebus configuration files under https://ebus.github.io/. 
You can download the config files from the config path https://ebus.github.io/next/ to a local path, add 76.vwz00.csv from the location https://github.com/WulfgarW/sensonetEbus/tree/master/ebusd-config-files/76.vwz00.csv and substitute 15.ctlv2.csv by the file https://github.com/WulfgarW/sensonetEbus/tree/master/ebusd-config-files/15.ctlv2.csv 

## Getting Started

This project is still in a preliminary state.

