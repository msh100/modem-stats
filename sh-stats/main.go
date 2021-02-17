package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"
)

func main() {
	timeStart := time.Now().UnixNano() / int64(time.Millisecond)

	localFile := getenv("LOCAL_FILE", "")
	SHVersion := getenv("SH_VERSION", "3")

	var body []byte
	if localFile != "" {
		file, _ := os.Open(localFile)
		body, _ = ioutil.ReadAll(file)
	} else {
		routerIP := getenv("ROUTER_IP", "192.168.100.1")
		var requestURL string
		if SHVersion == "4" {
			requestURL = requestURL4(routerIP)
		} else {
			requestURL = requestURL3(routerIP)
		}

		resp, err := http.Get(requestURL)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		body, err = ioutil.ReadAll(resp.Body)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	}
	timeTotal := (time.Now().UnixNano() / int64(time.Millisecond)) - timeStart

	var routerStats routerStats
	var err error
	if SHVersion == "4" {
		routerStats, err = parseStats4(body)
	} else {
		routerStats, err = parseStats3(body)
	}

	if err != nil {
		log.Printf("Error returned by parser: %v", err)
		os.Exit(1)
	}

	for _, downChannel := range routerStats.downChannels {
		output := fmt.Sprintf(
			"downstream,channel=%d,id=%d frequency=%d,snr=%d,power=%d,prerserr=%d,postrserr=%d",
			downChannel.channel,
			downChannel.channelID,
			downChannel.frequency,
			downChannel.snr,
			downChannel.power,
			downChannel.prerserr,
			downChannel.postrserr,
		)
		fmt.Println(output)
	}
	for _, upChannel := range routerStats.upChannels {
		output := fmt.Sprintf(
			"upstream,channel=%d,id=%d frequency=%d,power=%d",
			upChannel.channel,
			upChannel.channelID,
			upChannel.frequency,
			upChannel.power,
		)
		fmt.Println(output)
	}
	for _, config := range routerStats.configs {
		output := fmt.Sprintf(
			"config,config=%s maxrate=%d,maxburst=%d",
			config.config,
			config.maxrate,
			config.maxburst,
		)
		fmt.Println(output)
	}

	fmt.Println(fmt.Sprintf("shstatsinfo timems=%d", timeTotal))
}

func getenv(key, fallback string) string {
	value := os.Getenv(key)
	if len(value) == 0 {
		return fallback
	}
	return value
}
