package linux_net

import (
	"os"
	"strings"
	"errors"
	"io/ioutil"
	"regexp"
	"strconv"
)

type ifaceMetrics struct {
	rx_bytes uint64
	rx_errs uint64
	rx_drop uint64
	rx_packets uint
	/*
	rx_fifo uint
	rx_frame uint
	rx_compressed uint
	rx_multicast uint
	*/

	tx_bytes uint64
	tx_errs uint64
	tx_drop uint64
	tx_packets uint
	/*
	tx_fifo uint
	tx_frame uint
	tx_compressed uint
	tx_multicast uint
	*/
}

type ifaceData struct {
	name string
	operstate string
	metrics ifaceMetrics
}

func getIfaceStatNames() []string {
	return []string{
		"rx_bytes",
		"rx_errors",
		"rx_dropped",
		"rx_packets",
		"tx_bytes",
		"tx_errors",
		"tx_dropped",
		"tx_packets",
	}
}

func ifaceMetricMaps() map[string]uint64 {
	return map[string]uint64{
		"rx_bytes"		: 0,
		"rx_errors"		: 0,
		"rx_dropped"	: 0,
		"rx_packets"	: 0,
		"tx_bytes"		: 0,
		"tx_errors"		: 0,
		"tx_dropped"	: 0,
		"tx_packets"	: 0,
	}
}

func getLinkStateOptions() map[string]int {
	// https://elixir.bootlin.com/linux/latest/source/net/core/net-sysfs.c
	return map[string]int {
		"up" : 0,
		"testing" : 1,
		"lowerlayerdown" : 2,
		"down" : 2,
		"unknown" : 3,
		// dormant int = ?
		}
}


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


func getInterfacesForCheck(configIface *string , includeInterfaces *string , excludeInterfaces *string ) ([]string, error) {
	networkInterfaces, err := getInterfaces()
	if (err != nil) {
		return []string{}, err
	}
	if strings.Compare(*configIface,  "") != 0 {
		// interface set, ignore regex
		for _, iface := range networkInterfaces {
			if strings.Compare(iface, *configIface) == 0 {
				return  []string{iface}, nil
			}
		}
		return []string{""}, errors.New("No suitable Interface")
	}

	var result []string

	//fmt.Print("includePattern: ", *includeInterfaces, "\n")
	//fmt.Print("excludePattern: ", *excludeInterfaces, "\n")

	for _, iface := range networkInterfaces {
		//fmt.Print("Interface: ", iface, "\n")
		if strings.Compare(iface, "lo") == 0 { continue }
		inclmatched, err := regexp.MatchString(*includeInterfaces, iface)
		//fmt.Print("InclMatch: ", inclmatched, "\n")
		if err != nil {
			return nil, err
		}
		if inclmatched != true { continue }

		if *excludeInterfaces != "" {
			exclmatched, err := regexp.MatchString(*excludeInterfaces, iface)
			//fmt.Print("ExclMatch: ", exclmatched, "\n")
			if err != nil {
				return nil, err
			}
			if exclmatched == true { continue }
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
func getInterfaceState(ifaceName *string) (string, error) {
	basePath := "/sys/class/net/" + *ifaceName

	bytes, err := ioutil.ReadFile(basePath + "/operstate")
	if err != nil {
		return "", err
	}
	state:= string(bytes)
	return state, nil
}


// Get interfaces statistics
// @result: ifaceStats, err
func getInfacesStatistics(ifaceName *string) (map[string]int, error) {
	basePath := "/sys/class/net/" + *ifaceName + "/statistics"
	results := make(map[string] int)

	for _, stat := range getIfaceStatNames() {
		numberBytes, err := ioutil.ReadFile(basePath + "/" + stat)
		if err != nil {
			return results, err
		}
		numberString := string(numberBytes)
		results[stat], err = strconv.Atoi(numberString[:len(numberString)-1])
		if err != nil {
			return results, err
		}
	}

	return results, nil
}
