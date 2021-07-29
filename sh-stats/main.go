package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"time"

	flags "github.com/jessevdk/go-flags"
)

func main() {
	_, err := flags.ParseArgs(&commandLineOpts, os.Args)
	if err != nil {
		os.Exit(1)
	}

	var body []byte
	var fetchTime int64
	if localFile := getenv("LOCAL_FILE", ""); localFile != "" {
		timeStart := time.Now().UnixNano() / int64(time.Millisecond)
		file, _ := os.Open(localFile)
		body, _ = ioutil.ReadAll(file)
		fetchTime = (time.Now().UnixNano() / int64(time.Millisecond)) - timeStart
	}

	var routerType string
	superhubType := getenv("SH_VERSION", "")
	if superhubType != "" {
		routerType = fmt.Sprintf("superhub%s", superhubType)
	} else {
		routerType = getenv("ROUTER_TYPE", "superhub3")
	}

	var modem docsisModem

	switch routerType {
	case "superhub4":
		modem = &superhub4{
			IPAddress: getenv("ROUTER_IP", "192.168.100.1"),
			stats:     body,
			fetchTime: fetchTime,
		}
	case "comhemc2":
		modem = &comhemc2{
			IPAddress: getenv("ROUTER_IP", "192.168.10.1"),
			stats:     body,
			fetchTime: fetchTime,
			username:  getenv("ROUTER_USER", "admin"),
			password:  getenv("ROUTER_PASS", "admin"),
		}
	case "skyhub2":
		modem = &skyhub2{
			IPAddress: getenv("ROUTER_IP", "192.168.0.1"),
			stats:     body,
			fetchTime: fetchTime,
			username:  getenv("ROUTER_USER", "admin"),
			password:  getenv("ROUTER_PASS", "sky"),
		}
	default:
		modem = &superhub3{
			IPAddress: getenv("ROUTER_IP", "192.168.100.1"),
			stats:     body,
			fetchTime: fetchTime,
		}
	}

	if commandLineOpts.PrometheusPort > 0 {
		modemStatsPrometheus(modem, commandLineOpts.PrometheusPort)
	} else {
		for {
			modemStats, err := fetchStats(modem)

			if err != nil {
				log.Printf("Error returned by parser: %v", err)
			} else {
				printForInflux(modemStats)
			}

			if commandLineOpts.Daemon {
				bufio.NewScanner(os.Stdin).Scan()
				resetStats(modem)
			} else {
				break
			}
		}
	}
}

func fetchStats(router docsisModem) (modemStats, error) {
	stats, err := router.ParseStats()
	return stats, err
}

func resetStats(router docsisModem) {
	router.ClearStats()
}
