package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"time"
)

type downChannel struct {
	channel_id int
	channel    int
	frequency  int
	snr        int
	power      int
	prerserr   int
	postrserr  int
}
type upChannel struct {
	channel_id int
	channel    int
	frequency  int
	power      int
}
type config struct {
	config   string
	maxrate  int
	maxburst int
}

func main() {
	timeStart := time.Now().UnixNano() / int64(time.Millisecond)

	localFile := getenv("LOCAL_FILE", "")
	var body []byte
	if localFile != "" {
		file, _ := os.Open(localFile)
		body, _ = ioutil.ReadAll(file)
	} else {
		routerIP := getenv("ROUTER_IP", "192.168.100.1")
		requestURL := fmt.Sprintf("http://%s/getRouterStatus", routerIP)

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

	var snmpData map[string]interface{}
	json.Unmarshal(body, &snmpData)

	downMIBBase := "1.3.6.1.2.1.10.127.1.1.1.1"
	upMIBBase := "1.3.6.1.2.1.10.127.1.1.2.1"
	snrBase := "1.3.6.1.2.1.10.127.1.1.4.1.5"
	bondBase := "1.3.6.1.2.1.10.127.1.1.4.1"
	powerBase := "1.3.6.1.4.1.4491.2.1.20.1.2.1.1"

	upstreamIdMIBRegex := "^1.3.6.1.4.1.4491.2.1.21.1.3.1.8.2.1.([0-9]+)$"
	upstreamIdMIB := regexp.MustCompile(upstreamIdMIBRegex)
	downstreamIdMIBRegex := "^1.3.6.1.4.1.4491.2.1.21.1.3.1.8.2.2.([0-9]+)$"
	downstreamIdMIB := regexp.MustCompile(downstreamIdMIBRegex)

	rateMIB := "1.3.6.1.4.1.4491.2.1.21.1.2.1.6.2.1"
	burstMIB := "1.3.6.1.4.1.4491.2.1.21.1.2.1.7.2.1"

	var downChannels []downChannel
	var upChannels []upChannel
	var configs []config

	downChannelMatchRegex := fmt.Sprintf("^%s.1.([0-9]+)$", downMIBBase)
	downChannelMatch := regexp.MustCompile(downChannelMatchRegex)
	upChannelMatchRegex := fmt.Sprintf("^%s.1.([0-9]+)$", upMIBBase)
	upChannelMatch := regexp.MustCompile(upChannelMatchRegex)

	for MIB, MIBValue := range snmpData {
		downMatch := downChannelMatch.FindStringSubmatch(MIB)
		if len(downMatch) > 0 {
			channel, _ := strconv.Atoi(downMatch[1])

			channelID, _ := strconv.Atoi(MIBValue.(string))

			frequencyMIB := fmt.Sprintf("%s.2.%d", downMIBBase, channel)
			frequency, _ := strconv.Atoi(snmpData[frequencyMIB].(string))

			snrMIB := fmt.Sprintf("%s.%d", snrBase, channel)
			snr, _ := strconv.Atoi(snmpData[snrMIB].(string))

			powerMIB := fmt.Sprintf("%s.6.%d", downMIBBase, channel)
			power, _ := strconv.Atoi(snmpData[powerMIB].(string))

			preRSErrMIB := fmt.Sprintf("%s.3.%d", bondBase, channel)
			preRSErr, _ := strconv.Atoi(snmpData[preRSErrMIB].(string))

			postRSErrMIB := fmt.Sprintf("%s.4.%d", bondBase, channel)
			postRSErr, _ := strconv.Atoi(snmpData[postRSErrMIB].(string))

			downChannels = append(downChannels, downChannel{
				channel_id: channelID,
				channel:    channel,
				frequency:  frequency,
				snr:        snr,
				power:      power,
				prerserr:   preRSErr,
				postrserr:  postRSErr,
			})
			continue
		}

		upMatch := upChannelMatch.FindStringSubmatch(MIB)
		if len(upMatch) > 0 {
			channel, _ := strconv.Atoi(upMatch[1])

			channelID, _ := strconv.Atoi(MIBValue.(string))

			frequencyMIB := fmt.Sprintf("%s.2.%d", upMIBBase, channel)
			frequency, _ := strconv.Atoi(snmpData[frequencyMIB].(string))

			powerMIB := fmt.Sprintf("%s.%d", powerBase, channel)
			power, _ := strconv.Atoi(snmpData[powerMIB].(string))

			upChannels = append(upChannels, upChannel{
				channel_id: channelID,
				channel:    channel,
				frequency:  frequency,
				power:      power,
			})
			continue
		}

		// Active profile value needs to be set to 1, we don't need to run
		// this logic on every single iteration of SNMP data
		if MIBValue == "1" {
			upstreamIDMatch := upstreamIdMIB.FindStringSubmatch(MIB)
			if len(upstreamIDMatch) > 0 {
				upstreamId := upstreamIDMatch[1]

				maxRateMIB := fmt.Sprintf("%s.%s", rateMIB, upstreamId)
				maxRate, _ := strconv.Atoi(snmpData[maxRateMIB].(string))

				maxBurstMIB := fmt.Sprintf("%s.%s", burstMIB, upstreamId)
				maxBurst, _ := strconv.Atoi(snmpData[maxBurstMIB].(string))

				configs = append(configs, config{
					config:   "upstream",
					maxrate:  maxRate,
					maxburst: maxBurst,
				})
				continue
			}

			downstreamIDMatch := downstreamIdMIB.FindStringSubmatch(MIB)
			if len(downstreamIDMatch) > 0 {
				upstreamId := downstreamIDMatch[1]

				maxRateMIB := fmt.Sprintf("%s.%s", rateMIB, upstreamId)
				maxRate, _ := strconv.Atoi(snmpData[maxRateMIB].(string))

				maxBurstMIB := fmt.Sprintf("%s.%s", burstMIB, upstreamId)
				maxBurst, _ := strconv.Atoi(snmpData[maxBurstMIB].(string))

				configs = append(configs, config{
					config:   "downstream",
					maxrate:  maxRate,
					maxburst: maxBurst,
				})
				continue
			}
		}
	}

	for _, downChannel := range downChannels {
		output := fmt.Sprintf(
			"downstream,channel=%d,id=%d frequency=%d,snr=%d,power=%d,prerserr=%d,postrserr=%d",
			downChannel.channel,
			downChannel.channel_id,
			downChannel.frequency,
			downChannel.snr,
			downChannel.power,
			downChannel.prerserr,
			downChannel.postrserr,
		)
		fmt.Println(output)
	}
	for _, upChannel := range upChannels {
		output := fmt.Sprintf(
			"upstream,channel=%d,id=%d frequency=%d,power=%d",
			upChannel.channel,
			upChannel.channel_id,
			upChannel.frequency,
			upChannel.power,
		)
		fmt.Println(output)
	}
	for _, config := range configs {
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
