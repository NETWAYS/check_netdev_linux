package main

import (
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"
	"errors"
	"io/ioutil"
	//"fmt"

	"github.com/NETWAYS/go-check"
)

const readme = `Read traffic for linux network interfaces and warn on thresholds
Normal mode: Detects all network interfaces and checks the link state
Measuring mode: Re-reads the counters after $MeasuringTime seconds to measure
the network traffic.
`
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

//type ifaceStats struct {
//	rx_bytes uint64
//	rx_errs uint64
//	rx_drop uint64
//	/*
//	rx_packets uint
//	rx_fifo uint
//	rx_frame uint
//	rx_compressed uint
//	rx_multicast uint
//	*/
//
//	tx_bytes uint64
//	tx_errs uint64
//	tx_drop uint64
//	/*
//	tx_packets uint
//	tx_fifo uint
//	tx_frame uint
//	tx_compressed uint
//	tx_multicast uint
//	*/
//}

type ifaceData struct {
	name string
	operstate string
}


func getIfaceStatNames() []string {
	return []string{
		"rx_bytes",
		"rx_errors",
		"rx_dropped",
		"tx_bytes",
		"tx_errors",
		"tx_dropped",
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


func getInterfaces() []string {
	file, err := os.Open("/sys/class/net")
	if err != nil {
		log.Fatal(err)
	}

	devices, err := file.Readdirnames(0)
	if err != nil {
		log.Fatal(err)
	}

	return devices
}


func getInterfacesForCheck(configIface *string , includeInterfaces *string , excludeInterfaces *string ) ([]string, error) {
	networkInterfaces := getInterfaces()
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
		inclmatched, err := regexp.MatchString(*includeInterfaces, iface)
		//fmt.Print("InclMatch: ", inclmatched, "\n")
		if err != nil {
			log.Fatal(err)
		}
		if inclmatched != true { continue }

		if *excludeInterfaces != "" {
			exclmatched, err := regexp.MatchString(*excludeInterfaces, iface)
			//fmt.Print("ExclMatch: ", exclmatched, "\n")
			if err != nil {
				log.Fatal(err)
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
func getInterfaceState(ifaceName *string) string {
	basePath := "/sys/class/net/" + *ifaceName

	bytes, err := ioutil.ReadFile(basePath + "/operstate")
	if err != nil {
		log.Fatal(err)
	}
	state:= string(bytes)
	return state
}

/*
func readNetdev() ([]ifaceData, error) {
	netdev_file, err := os.Open("/proc/net/dev")

	if err != nil {
		log.Fatal(err)
	}

	scanner := bufio.NewScanner(netdev_file)

	lines := []string {}

	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	numberOfDevices := len(lines) - 2
	devs := make([]ifaceData, numberOfDevices)

	for idx, line := range lines[2:] {
		// Ignore the first two lines
		line = strings.Trim(line, " ")
		lineParts := separator.Split(line, -1)

		devs[idx].name = strings.Trim(lineParts[0], ":")
		devs[idx].rx_bytes, err = strconv.ParseUint(lineParts[1], 10, 64)
		//println(devs[idx].rx_bytes)
		rx_packets := lineParts[2]
		rx_errs := lineParts[3]
		rx_drop := lineParts[4]
		rx_fifo := lineParts[5]
		rx_frame := lineParts[6]
		rx_compressed := lineParts[7]
		rx_multicast := lineParts[8]

		devs[idx].tx_bytes, err = strconv.ParseUint(lineParts[9], 10, 64)
		//println(devs[idx].tx_bytes)
		tx_packets := lineParts[10]
		tx_errs := lineParts[11]
		tx_drop := lineParts[12]
		tx_fifo := lineParts[13]
		tx_colls := lineParts[14]
		tx_carrier := lineParts[15]
		tx_compressed := lineParts[16]

		//println(idx, " ", iface, "RX: ", rx_bytes, ", TX: ", tx_bytes)
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}

	return devs, nil
}
*/

// Get interfaces statistics
// @result: ifaceStats, err
func getInfacesStatistics(ifaceName *string) (map[string]int, error) {
	basePath := "/sys/class/net/" + *ifaceName + "/statistics"
	results := make(map[string] int)

	for _, stat := range getIfaceStatNames() {
		numberBytes, err := ioutil.ReadFile(basePath + "/" + stat)
		if err != nil {
			log.Fatal(err)
		}
		numberString := string(numberBytes)
		results[stat], err = strconv.Atoi(numberString[:len(numberString)-1])
		if err != nil {
			log.Fatal(err)
		}
	}

	return results, nil
}

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

	// Parse arguments
	plugin.ParseArguments()

	ifaces, err := getInterfacesForCheck(configIface, includeInterfaces, excludeInterfaces)
	if err != nil {
		log.Fatal(err)
	}

	if len(ifaces) == 0 {
		check.Exit(3, "No devices match the expression")
	}

	interfaceData := make ([]ifaceData, len(ifaces))
	var result string

	var numberOfOfflineDevices = 0

	numberOfMetrics := len(getIfaceStatNames())

	for idx, iface := range ifaces {
		interfaceData[idx].name = iface
		// get state
		operState := getInterfaceState(&iface)
		operState = operState[:len(operState)-1]

		if strings.Compare(operState, "down") == 0 {
			numberOfOfflineDevices ++
		}

		statistics, err := getInfacesStatistics(&iface)
		if err != nil {
			log.Fatal(err)
		}

		result += interfaceData[idx].name + " is " + operState + ". "
		for key, value := range statistics {
			result += key + ": " + strconv.Itoa(value)
		}
		if idx == (len(ifaces) - 1) {
			result += "\n"
		} else {
			result += ",\n"
		}

	}

	if numberOfOfflineDevices > 0 {
		result = strconv.Itoa(numberOfOfflineDevices) + " devices are down\n" + result
		check.Exit(2, result)
	}
	check.Exit(0, result)
}
