package main

import (
	"regexp"
	"strconv"

	//"strings"
	//"fmt"
	"time"

	"github.com/NETWAYS/go-check"
	"github.com/NETWAYS/go-check/perfdata"
	"github.com/NETWAYS/go-check/result"
)

const readme = `Read traffic for linux network interfaces and warn on thresholds
Normal mode: Detects all network interfaces and checks the link state
Measuring mode: Re-reads the counters after $MeasuringTime seconds to measure
the network traffic.`

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
	measuringTime := plugin.FlagSet.Uint64P("measuringTime", "m", 0, "Measure for n seconds the traffic and report the amount")

	// Parse arguments
	plugin.ParseArguments()

	// Create result struct
	var overall result.Overall

	// Do the real work
	// Get interfaces
	ifaces, err := getInterfacesForCheck(configIface, includeInterfaces, excludeInterfaces)
	if err != nil {
		check.ExitError(err)
	}

	if len(ifaces) == 0 {
		check.ExitRaw(3, "No devices match the expression")
	}

	interfaceData := make([]ifaceData, len(ifaces))

	var numberOfOfflineDevices = 0

	//numberOfMetrics := len(getIfaceStatNames())

	// Get data
	for idx, iface := range ifaces {
		interfaceData[idx].name = iface

		// get state
		err := getInterfaceState(&interfaceData[idx])
		if err != nil {
			check.ExitError(err)
		}
	}

	firstDataPoint := make([]statistics, len(ifaces))
	if *measuringTime != 0 {
		for idx := range ifaces {

			// get numbers
			err = getInfacesStatistics(&interfaceData[idx], &firstDataPoint[idx])
			if err != nil {
				check.ExitError(err)
			}
		}

		time.Sleep(time.Duration(*measuringTime) * time.Second)
	}

	for idx := range ifaces {

		// get numbers
		err = getInfacesStatistics(&interfaceData[idx], nil)
		if err != nil {
			check.ExitError(err)
		}
	}

	if *checkOfflineDevices {
		var onlineResult string
		for idx := range ifaces {
			if *checkOfflineDevices {
				if interfaceData[idx].operstate == Down {
					numberOfOfflineDevices++
				}

				switch interfaceData[idx].operstate {
				case Down:
					{
						onlineResult += interfaceData[idx].name + " is down. "
					}
				case Up:
					{
						onlineResult += interfaceData[idx].name + " is up. "
					}
				case Testing:
					{
						onlineResult += interfaceData[idx].name + " is testing. "
					}
				default:
					{
						onlineResult += interfaceData[idx].name + " is unknown. "
					}
				}

			}
		}
		onlineResult = strconv.Itoa(numberOfOfflineDevices) + " devices are down\n" + onlineResult
		overall.Add(check.Critical, onlineResult)
	}

	// Formulate result with numbers
	metrics := getIfaceStatNames()
	var metricOutput string
	for idx, iface := range interfaceData {
		metricOutput += iface.name + ": "
		for jdx, metric := range metrics {
			metricOutput += metric + " "

			if *measuringTime != 0 {
				diff := (iface.metrics[jdx] - firstDataPoint[idx][jdx]) / *measuringTime
				metricOutput += "Diff: " + strconv.FormatUint(diff, 10) + ", "
			}

			metricOutput += "Total: " + strconv.FormatUint(interfaceData[idx].metrics[jdx], 10) + "; "
		}
	}

	// Perfdata
	pl := new(perfdata.PerfdataList)

	for idx, iface := range interfaceData {
		for jdx, metric := range metrics {

			if *measuringTime != 0 {
				perfdata := new(perfdata.Perfdata)
				diff := (iface.metrics[jdx] - firstDataPoint[idx][jdx]) / *measuringTime
				perfdata.Value = diff
				perfdata.Uom = "B"
				perfdata.Label = iface.name + "-" + metric + "-throughput"
				pl.Add(perfdata)
			}

			perfdata := new(perfdata.Perfdata)
			perfdata.Label = iface.name + "-" + metric + "-total"
			perfdata.Value = interfaceData[idx].metrics[jdx]
			perfdata.Uom = "c"
			pl.Add(perfdata)
		}
	}

	// Perfdata
	overall.Add(check.OK, metricOutput)

	check.ExitRaw(overall.GetStatus(), overall.GetOutput())
}
