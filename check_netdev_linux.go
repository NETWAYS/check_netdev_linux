package main

import (
	"bufio"
	"log"
	"os"
	"regexp"
	"strings"
	"github.com/NETWAYS/go-check"
)

const readme = `Read traffic for linux network interfaces and warn on thresholds`

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

	// Parse arguments
	// Not needed right now
	//plugin.ParseArguments()

	netdev_file, err := os.Open("/proc/net/dev")

	if err != nil {
		log.Fatal(err)
	}

	scanner := bufio.NewScanner(netdev_file)

	lines := []string {}

	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	result := ""

	for _, line := range lines[2:] {
		// Ignore the first two lines
		line = strings.Trim(line, " ")
		lineParts := separator.Split(line, -1)

		iface := strings.Trim(lineParts[0], ":")
		rx_bytes := lineParts[1]
		/*
		rx_packets := lineParts[2]
		rx_errs := lineParts[3]
		rx_drop := lineParts[4]
		rx_fifo := lineParts[5]
		rx_frame := lineParts[6]
		rx_compressed := lineParts[7]
		rx_multicast := lineParts[8]
		*/

		tx_bytes := lineParts[9]
		/*
		tx_packets := lineParts[10]
		tx_errs := lineParts[11]
		tx_drop := lineParts[12]
		tx_fifo := lineParts[13]
		tx_colls := lineParts[14]
		tx_carrier := lineParts[15]
		tx_compressed := lineParts[16]
		*/

		result += iface + " rx:" + rx_bytes + " tx:" + tx_bytes + "|"
		result += iface + "rx=valuec;0;0;0;" + rx_bytes + " "
		result += iface + "tx=valuec;0;0;0;" + tx_bytes + " \n"
		//println(idx, " ", iface, "RX: ", rx_bytes, ", TX: ", tx_bytes)
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}

	check.Exit(0, result)
}
