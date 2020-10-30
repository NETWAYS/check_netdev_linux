package main

import (
	"regexp"
	"strconv"
	"strings"
	"github.com/check_netdev_linux/linux_net"
	//"fmt"

	"github.com/NETWAYS/go-check"
)

const readme = `Read traffic for linux network interfaces and warn on thresholds
Normal mode: Detects all network interfaces and checks the link state`
/*
Measuring mode: Re-reads the counters after $MeasuringTime seconds to measure
the network traffic.
*/
/*
type Config struct {
	IgnoreLoopback bool
	WarningValue	int
	CriticalValue	int
	includeInterfacesRegex string
	excludeInterfacesRegex string
}
*/


var (
	separator = regexp.MustCompile(`\s+`)
)


func main() {
	defer check.CatchPanic()

	plugin := check.NewConfig()
	plugin.Name = "check_netdev_linux"
	plugin.Readme = readme
	plugin.Version = "0.1"
	plugin.Timeout = 30


	configIface := plugin.FlagSet.StringP("interface", "I", "", "Single Interface to check (exclusive to incldRgxIntrfc and excldRgxIntrfc)")
	includeInterfaces := plugin.FlagSet.StringP("incldRgxIntrfc", "i", ".*", "Regex to select interfaces (Default: all)")
	excludeInterfaces := plugin.FlagSet.StringP("excldRgxIntrfc", "e", "", "Regex to ignore interfaces (Default: nothing)")
	checkOfflineDevices := plugin.FlagSet.BoolP("checkOffline", "c", false, "Check whether interfaces are online")

	// Parse arguments
	plugin.ParseArguments()

	// Do the real work
	// Get interfaces
	ifaces, err := linux_net.getInterfacesForCheck(configIface, includeInterfaces, excludeInterfaces)
	if err != nil {
		check.ExitError(err)
	}

	if len(ifaces) == 0 {
		check.Exit(3, "No devices match the expression")
	}

	interfaceData := make ([]linux_net.ifaceData, len(ifaces))
	var result string

	var numberOfOfflineDevices = 0

	numberOfMetrics := len(linux_net.getIfaceStatNames())

	for idx, iface := range ifaces {
		interfaceData[idx].name = iface

		// get state
		operState := linux_net.getInterfaceState(&iface)
		operState = operState[:len(operState)-1]

		if strings.Compare(operState, "down") == 0 {
			numberOfOfflineDevices ++
		}

		// get numbers
		statistics, err := linux_net.getInfacesStatistics(&iface)
		if err != nil {
			check.ExitError(err)
		}

		result += interfaceData[idx].name + " is " + operState + ". "

		counter := 1
		for key, value := range statistics {
			result += key + ":" + strconv.Itoa(value)
			if counter != numberOfMetrics {
				result += " "
			}
			counter ++
		}
		if idx == (len(ifaces) - 1) {
			result += "\n"
		} else {
			result += ",\n"
		}

	}

	if (numberOfOfflineDevices > 0 && *checkOfflineDevices) {
		result = strconv.Itoa(numberOfOfflineDevices) + " devices are down\n" + result
		check.Exit(2, result)
	}
	check.Exit(0, result)
}
