package main

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
)

type superhub3 struct {
	IPAddress string
	stats     []byte
	fetchTime int64
}

func (sh3 *superhub3) ClearStats() {
	sh3.stats = nil
}

func (sh3 *superhub3) fetchURL() string {
	return fmt.Sprintf("http://%s/getRouterStatus", sh3.IPAddress)
}

func (sh3 *superhub3) ParseStats() (modemStats, error) {
	if sh3.stats == nil {
		var err error
		sh3.stats, sh3.fetchTime, err = simpleHTTPFetch(sh3.fetchURL())
		if err != nil {
			return modemStats{}, err
		}
	}

	var snmpData map[string]interface{}
	json.Unmarshal(sh3.stats, &snmpData)

	downMIBBase := "1.3.6.1.2.1.10.127.1.1.1.1"
	upMIBBase := "1.3.6.1.2.1.10.127.1.1.2.1"
	snrBase := "1.3.6.1.2.1.10.127.1.1.4.1.5"
	bondBase := "1.3.6.1.2.1.10.127.1.1.4.1"
	powerBase := "1.3.6.1.4.1.4491.2.1.20.1.2.1.1"

	upstreamIDMIBRegex := "^1.3.6.1.4.1.4491.2.1.21.1.3.1.8.2.[0-9]*[13579].([0-9]+)$"
	upstreamIDMIB := regexp.MustCompile(upstreamIDMIBRegex)
	downstreamIDMIBRegex := "^1.3.6.1.4.1.4491.2.1.21.1.3.1.8.2.[0-9]*[02468].([0-9]+)$"
	downstreamIDMIB := regexp.MustCompile(downstreamIDMIBRegex)

	rateMIB := "1.3.6.1.4.1.4491.2.1.21.1.2.1.6.2.1"
	burstMIB := "1.3.6.1.4.1.4491.2.1.21.1.2.1.7.2.1"

	var downChannels []modemChannel
	var upChannels []modemChannel
	var configs []modemConfig

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

			downChannels = append(downChannels, modemChannel{
				channelID:  channelID,
				channel:    channel,
				frequency:  frequency,
				snr:        snr,
				power:      power,
				prerserr:   preRSErr,
				postrserr:  postRSErr,
				modulation: "QAM256",
				scheme:     "SC-QAM",
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

			upChannels = append(upChannels, modemChannel{
				channelID: channelID,
				channel:   channel,
				frequency: frequency,
				power:     power,
			})
			continue
		}

		// Active profile value needs to be set to 1, we don't need to run
		// this logic on every single iteration of SNMP data
		if MIBValue == "1" {
			upstreamIDMatch := upstreamIDMIB.FindStringSubmatch(MIB)
			if len(upstreamIDMatch) > 0 {
				upstreamID := upstreamIDMatch[1]

				maxRateMIB := fmt.Sprintf("%s.%s", rateMIB, upstreamID)
				maxRate, _ := strconv.Atoi(snmpData[maxRateMIB].(string))

				maxBurstMIB := fmt.Sprintf("%s.%s", burstMIB, upstreamID)
				maxBurst, _ := strconv.Atoi(snmpData[maxBurstMIB].(string))

				configs = append(configs, modemConfig{
					config:   "upstream",
					maxrate:  maxRate,
					maxburst: maxBurst,
				})
				continue
			}

			downstreamIDMatch := downstreamIDMIB.FindStringSubmatch(MIB)
			if len(downstreamIDMatch) > 0 {
				downstreamID := downstreamIDMatch[1]

				maxRateMIB := fmt.Sprintf("%s.%s", rateMIB, downstreamID)
				maxRate, _ := strconv.Atoi(snmpData[maxRateMIB].(string))

				maxBurstMIB := fmt.Sprintf("%s.%s", burstMIB, downstreamID)
				maxBurst, _ := strconv.Atoi(snmpData[maxBurstMIB].(string))

				configs = append(configs, modemConfig{
					config:   "downstream",
					maxrate:  maxRate,
					maxburst: maxBurst,
				})
				continue
			}
		}
	}

	return modemStats{
		configs:      configs,
		upChannels:   upChannels,
		downChannels: downChannels,
		fetchTime:    sh3.fetchTime,
	}, nil
}
