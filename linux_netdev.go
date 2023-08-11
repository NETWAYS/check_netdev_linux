package main

import (
	"errors"
	"os"
	"regexp"
	"strconv"
	"strings"
)

// types and constants
const (
	Up             = 0
	Testing        = 1
	Lowerlayerdown = 2
	Down           = 3
	Unknown        = 4
	// dormant = 5
)

// Constants and the string array MUST be kept in sync!
//
//nolint:unused
const (
	rx_bytes int = iota
	rx_errs
	rx_drop
	rx_packets
	/*
		rx_fifo uint
		rx_frame uint
		rx_compressed uint
		rx_multicast uint
	*/

	tx_bytes
	tx_errs
	tx_drop
	tx_packets
	/*
		tx_fifo uint
		tx_frame uint
		tx_compressed uint
		tx_multicast uint
	*/
	metricLength
)

func getIfaceStatNames() []string {
	return []string{
		"rx_bytes",
		"rx_errors",
		"rx_dropped",
		"rx_packets",
		/*
			"rx_fifo",
			"rx_frame"
			"rx_compressed"
			"rx_multicast"
		*/

		"tx_bytes",
		"tx_errors",
		"tx_dropped",
		"tx_packets",
		/*
			"tx_fifo",
			"tx_frame",
			"tx_compressed",
			"tx_multicast",
		*/
	}
}

type statistics [metricLength]uint64

type ifaceData struct {
	name      string
	operstate uint
	metrics   statistics
}

// funcs
func getInterfaces() ([]string, error) {
	file, err := os.Open("/sys/class/net")
	if err != nil {
		return []string{}, err
	}

	devices, err := file.Readdirnames(0)
	if err != nil {
		return []string{}, err
	}

	return devices, nil
}

func getInterfacesForCheck(configIface *string, includeInterfaces *string, excludeInterfaces *string) ([]string, error) {
	networkInterfaces, err := getInterfaces()
	if err != nil {
		return []string{}, err
	}
	if strings.Compare(*configIface, "") != 0 {
		// interface set, ignore regex
		for _, iface := range networkInterfaces {
			if strings.Compare(iface, *configIface) == 0 {
				return []string{iface}, nil
			}
		}
		return []string{""}, errors.New("No suitable Interface")
	}

	var result []string

	//fmt.Print("includePattern: ", *includeInterfaces, "\n")
	//fmt.Print("excludePattern: ", *excludeInterfaces, "\n")

	for _, iface := range networkInterfaces {
		//fmt.Print("Interface: ", iface, "\n")
		if strings.Compare(iface, "lo") == 0 {
			continue
		}
		inclmatched, err := regexp.MatchString(*includeInterfaces, iface)
		//fmt.Print("InclMatch: ", inclmatched, "\n")
		if err != nil {
			return nil, err
		}
		if !inclmatched {
			continue
		}

		if *excludeInterfaces != "" {
			exclmatched, err := regexp.MatchString(*excludeInterfaces, iface)
			//fmt.Print("ExclMatch: ", exclmatched, "\n")
			if err != nil {
				return nil, err
			}
			if exclmatched {
				continue
			}
		}

		result = append(result, iface)
	}
	return result, nil
}

// getInterfaceState receives the name of an interfaces and returns
// an integer result code representing the state of the interface
// @result = 0 => Interface is up
// @result = 2 => Interface is down
// @result = 3 => Interface is unknown or state of the interface is unknown for some reason
func getInterfaceState(data *ifaceData) error {
	basePath := "/sys/class/net/" + data.name

	bytes, err := os.ReadFile(basePath + "/operstate")
	if err != nil {
		return err
	}
	//state:= string(bytes)
	//return state, nil
	switch string(bytes) {
	case "up":
		{
			data.operstate = Up
			return nil
		}
	case "testing":
		{
			data.operstate = Testing
			return nil
		}
	case "down":
		{
			data.operstate = Down
			return nil
		}
	case "lowerlayerdown":
		{
			data.operstate = Lowerlayerdown
			return nil
		}
	default:
		{
			data.operstate = Unknown
			return nil
		}
	}
}

// Get interfaces statistics
// @result: ifaceStats, err
func getInfacesStatistics(data *ifaceData, metricArea *statistics) error {
	basePath := "/sys/class/net/" + data.name + "/statistics"

	var val uint64

	for idx, stat := range getIfaceStatNames() {
		numberBytes, err := os.ReadFile(basePath + "/" + stat)
		if err != nil {
			return err
		}
		numberString := string(numberBytes)
		val, err = strconv.ParseUint(numberString[:len(numberString)-1], 10, 64)
		if err != nil {
			return err
		}
		if metricArea == nil {
			data.metrics[idx] = val
		} else {
			metricArea[idx] = val
		}
	}

	return nil
}
