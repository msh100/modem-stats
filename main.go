package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"time"

	flags "github.com/jessevdk/go-flags"
	"github.com/msh100/modem-stats/modems/comhemc2"
	"github.com/msh100/modem-stats/modems/skyhub2"
	"github.com/msh100/modem-stats/modems/superhub3"
	"github.com/msh100/modem-stats/modems/superhub4"
	"github.com/msh100/modem-stats/modems/ubee"
	"github.com/msh100/modem-stats/outputs"
	"github.com/msh100/modem-stats/utils"
)

var commandLineOpts struct {
	Daemon         bool   `short:"d" long:"daemon" description:"Gather statistics on new line to STDIN"`
	PrometheusPort int    `short:"p" long:"port" description:"Prometheus exporter port (disabled if not defined)"`
	Modem          string `short:"m" long:"modem" description:"Which modem to use" default:"superhub3"`
	ModemIP        string `long:"ip" description:"The modem's IP address"`
	Username       string `long:"username" description:"The modem's username (if applicable)"`
	Password       string `long:"password" description:"The modem's password (if applicable)"`
}

func main() {
	_, err := flags.ParseArgs(&commandLineOpts, os.Args)
	if err != nil {
		log.Fatal("error parsing command line arguments")
		os.Exit(1)
	}

	var body []byte
	var fetchTime int64
	if localFile := utils.Getenv("LOCAL_FILE", ""); localFile != "" {
		timeStart := time.Now().UnixNano() / int64(time.Millisecond)
		file, _ := os.Open(localFile)
		body, _ = ioutil.ReadAll(file)
		fetchTime = (time.Now().UnixNano() / int64(time.Millisecond)) - timeStart
	}

	var routerType string
	superhubType := utils.Getenv("SH_VERSION", "")
	if superhubType != "" {
		routerType = fmt.Sprintf("superhub%s", superhubType)
	} else {
		routerType = utils.Getenv("ROUTER_TYPE", commandLineOpts.Modem)
	}

	var modem utils.DocsisModem

	switch routerType {
	case "superhub4":
		modem = &superhub4.Modem{
			IPAddress: utils.Getenv("ROUTER_IP", commandLineOpts.ModemIP),
			Stats:     body,
			FetchTime: fetchTime,
		}
	case "comhemc2":
		modem = &comhemc2.Modem{
			IPAddress: utils.Getenv("ROUTER_IP", commandLineOpts.ModemIP),
			Stats:     body,
			FetchTime: fetchTime,
			Username:  utils.Getenv("ROUTER_USER", commandLineOpts.Username),
			Password:  utils.Getenv("ROUTER_PASS", commandLineOpts.Password),
		}
	case "skyhub2":
		modem = &skyhub2.Modem{
			IPAddress: utils.Getenv("ROUTER_IP", commandLineOpts.ModemIP),
			Stats:     body,
			FetchTime: fetchTime,
			Username:  utils.Getenv("ROUTER_USER", commandLineOpts.Username),
			Password:  utils.Getenv("ROUTER_PASS", commandLineOpts.Password),
		}
	case "superhub3":
		modem = &superhub3.Modem{
			IPAddress: utils.Getenv("ROUTER_IP", commandLineOpts.ModemIP),
			Stats:     body,
			FetchTime: fetchTime,
		}
	case "ubee":
		modem = &ubee.Modem{
			IPAddress: utils.Getenv("ROUTER_IP", commandLineOpts.ModemIP),
			Stats:     body,
			FetchTime: fetchTime,
		}
	default:
		log.Fatalf("unknown modem: %s", routerType)
	}

	if commandLineOpts.PrometheusPort > 0 {
		outputs.Prometheus(modem, commandLineOpts.PrometheusPort)
	} else {
		for {
			modemStats, err := utils.FetchStats(modem)

			if err != nil {
				log.Printf("Error returned by parser: %v", err)
			} else {
				outputs.PrintForInflux(modemStats)
			}

			if commandLineOpts.Daemon {
				bufio.NewScanner(os.Stdin).Scan()
				utils.ResetStats(modem)
			} else {
				break
			}
		}
	}
}
