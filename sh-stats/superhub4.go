package main

import (
	"encoding/json"
	"fmt"
	"strconv"
)

func parseStats4(body []byte) routerStats {
	var arr []string
	json.Unmarshal([]byte(body), &arr)

	downRate, _ := strconv.Atoi(arr[11])
	downBurst, _ := strconv.Atoi(arr[12])
	upRate, _ := strconv.Atoi(arr[15])
	upBurst, _ := strconv.Atoi(arr[16])

	var downChannelsData [][]string
	json.Unmarshal([]byte(arr[20]), &downChannelsData)
	var downChannels []downChannel
	for index, downChannelData := range downChannelsData {
		channelID, _ := strconv.Atoi(downChannelData[0])
		frequency, _ := strconv.Atoi(downChannelData[1])
		snr, _ := strconv.ParseFloat(downChannelData[3], 64)
		snrint := int(snr * 10)
		power, _ := strconv.ParseFloat(downChannelData[2], 64)
		powerint := int(power * 10)
		prerserr, _ := strconv.Atoi(downChannelData[7])
		postrserr, _ := strconv.Atoi(downChannelData[8])

		downChannels = append(downChannels, downChannel{
			channelID: channelID,
			channel:   index + 1,
			frequency: frequency,
			snr:       snrint,
			power:     powerint,
			prerserr:  prerserr,  // TODO: Verify?
			postrserr: postrserr, // TODO: Verify?
		})
	}

	var down31ChannelsData [][]string
	json.Unmarshal([]byte(arr[23]), &down31ChannelsData)
	for index, down31ChannelData := range down31ChannelsData {
		channelID, _ := strconv.Atoi(down31ChannelData[0])
		frequency, _ := strconv.Atoi(down31ChannelData[1] + "000000")
		snr, _ := strconv.ParseFloat(down31ChannelData[7], 64)
		snrint := int(snr * 10)
		power, _ := strconv.ParseFloat(down31ChannelData[8], 64)
		powerint := int(power * 10)
		prerserr, _ := strconv.Atoi(down31ChannelData[9])
		postrserr, _ := strconv.Atoi(down31ChannelData[10])

		downChannels = append(downChannels, downChannel{
			channelID: channelID,
			channel:   index + 1,
			frequency: frequency,
			snr:       snrint,
			power:     powerint,
			prerserr:  prerserr,
			postrserr: postrserr,
		})
	}

	var upChannelsData [][]string
	json.Unmarshal([]byte(arr[21]), &upChannelsData)
	var upChannels []upChannel
	for index, upChannelData := range upChannelsData {
		channelID, _ := strconv.Atoi(upChannelData[0])
		frequency, _ := strconv.Atoi(upChannelData[1])
		power, _ := strconv.ParseFloat(upChannelData[2], 64)
		powerint := int(power * 10)

		upChannels = append(upChannels, upChannel{
			channelID: channelID,
			channel:   index + 1,
			frequency: frequency,
			power:     powerint,
		})
	}

	return routerStats{
		configs: []config{
			{
				config:   "downstream",
				maxrate:  downRate,
				maxburst: downBurst,
			},
			{
				config:   "upstream",
				maxrate:  upRate,
				maxburst: upBurst,
			},
		},
		upChannels:   upChannels,
		downChannels: downChannels,
	}
}

func requestURL4(ip string) string {
	return fmt.Sprintf("http://%s/php/ajaxGet_device_networkstatus_data.php", ip)
}
