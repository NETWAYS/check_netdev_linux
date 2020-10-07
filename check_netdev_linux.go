package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"
	"errors"

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

type ifaceData struct {
	//name [64]rune
	name string

	rx_bytes uint64
	/*
	rx_packets uint
	rx_errs uint
	rx_drop uint
	rx_fifo uint
	rx_frame uint
	rx_compressed uint
	rx_multicast uint
	*/

	tx_bytes uint64
	/*
	tx_packets uint
	tx_errs uint
	tx_drop uint
	tx_fifo uint
	tx_frame uint
	tx_compressed uint
	tx_multicast uint
	*/
	operstate string
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

func getInterfacState(ifaceName *string, data *ifaceData) int {
	basePath := "/sys/class/net/" + *ifaceName
	data.name = *ifaceName
	var err error

	operstate_file, err := os.Open(basePath + "/operstate")
	if err != nil {
		log.Fatal(err)
	}


	return 0
}

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
		/*
		rx_packets := lineParts[2]
		rx_errs := lineParts[3]
		rx_drop := lineParts[4]
		rx_fifo := lineParts[5]
		rx_frame := lineParts[6]
		rx_compressed := lineParts[7]
		rx_multicast := lineParts[8]
		*/

		devs[idx].tx_bytes, err = strconv.ParseUint(lineParts[9], 10, 64)
		//println(devs[idx].tx_bytes)
		/*
		tx_packets := lineParts[10]
		tx_errs := lineParts[11]
		tx_drop := lineParts[12]
		tx_fifo := lineParts[13]
		tx_colls := lineParts[14]
		tx_carrier := lineParts[15]
		tx_compressed := lineParts[16]
		*/

		//println(idx, " ", iface, "RX: ", rx_bytes, ", TX: ", tx_bytes)
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}

	return devs, nil
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

	for _, iface := range networkInterfaces {
		inclmatched, err := regexp.MatchString(*includeInterfaces, iface)
		if err != nil {
			log.Fatal(err)
		}
		if inclmatched != true { continue }

		exclmatched, err := regexp.MatchString(*excludeInterfaces, iface)
		if err != nil {
			log.Fatal(err)
		}
		if exclmatched == true { continue }

		result = append(result, iface)
	}
	return result, nil
}

func main() {
	defer check.CatchPanic()

	plugin := check.NewConfig()
	plugin.Name = "check_netdev_linux"
	plugin.Readme = readme
	plugin.Version = "0.1"
	plugin.Timeout = 30


	configIface := plugin.FlagSet.StringP("interface", "I", "", "Single Interface to check (exclusive to incldRgxIntrfc and excldRgxIntrfc)")
	includeInterfaces := plugin.FlagSet.StringP("incldRgxIntrfc", "i", "*", "Regex to select interfaces (Default: all)")
	excludeInterfaces := plugin.FlagSet.StringP("excldRgxIntrfc", "e", "*", "Regex to ignore interfaces (Default: nothing)")
	timeout := plugin.FlagSet.IntP("timeout", "t", 0, "time frame for measuring traffic (Default: 0 - Don't measure)")

	ifaces, err := getInterfacesForCheck(configIface, includeInterfaces, excludeInterfaces)
	if err != nil {
		log.Fatal(err)
	}

	if len(ifaces) == 0 {
		check.Exit(3, "No devices match the expression")
	}

	

	devs, err := readNetdev()
	if err != nil {
		log.Fatal(err)
	}

	// Parse arguments
	// Not needed right now
	//plugin.ParseArguments()
	result := ""

	for _, device := range devs {
		result += device.name + " rx:" + fmt.Sprint(device.rx_bytes) + " tx:" + fmt.Sprint(device.tx_bytes) + "|"
		result += device.name + "rx=valuec;0;0;0;" + fmt.Sprint(device.rx_bytes) + " "
		result += device.name + "tx=valuec;0;0;0;" + fmt.Sprint(device.tx_bytes) + " "
		result += "\n"
	}

	check.Exit(0, result)
}
