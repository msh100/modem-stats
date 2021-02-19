package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
)

func parseStats4(body []byte) (routerStats, error) {
	var errstrings []string

	var arr []string
	json.Unmarshal([]byte(body), &arr)

	downRate, _ := strconv.Atoi(arr[11])
	downBurst, _ := strconv.Atoi(arr[12])
	upRate, _ := strconv.Atoi(arr[15])
	upBurst, _ := strconv.Atoi(arr[16])

	if downRate == 0 || downBurst == 0 || upRate == 0 || upBurst == 0 {
		error := fmt.Errorf("Got nil values for speed config %d,%d,%d,%d", downRate, downBurst, upRate, upBurst)
		errstrings = append(errstrings, error.Error())
	}

	var downChannelsData [][]string
	json.Unmarshal([]byte(arr[20]), &downChannelsData)
	var downChannels []downChannel
	for index, downChannelData := range downChannelsData {
		if len(downChannelData) != 9 {
			error := fmt.Errorf("Abnormal down channel length, expected 9, got %d", len(downChannelData))
			errstrings = append(errstrings, error.Error())
			break
		}

		channelID, _ := strconv.Atoi(downChannelData[0])
		frequency, _ := strconv.Atoi(downChannelData[1])
		snr, _ := strconv.ParseFloat(downChannelData[3], 64)
		snrint := int(snr * 10)
		power, _ := strconv.ParseFloat(downChannelData[2], 64)
		powerint := int(power * 10)
		prerserr, _ := strconv.Atoi(downChannelData[7])
		postrserr, _ := strconv.Atoi(downChannelData[8])

		if channelID < 1 || channelID > 1024 {
			error := fmt.Errorf("Abnormal channel ID, got %d", channelID)
			errstrings = append(errstrings, error.Error())
			break
		}
		if powerint > 2000 || powerint < -2000 {
			error := fmt.Errorf("Power level for 3.0 channel %d is abnormal, got %d", channelID, powerint)
			errstrings = append(errstrings, error.Error())
			break
		}

		downChannels = append(downChannels, downChannel{
			channelID:  channelID,
			channel:    index + 1,
			frequency:  frequency,
			snr:        snrint,
			power:      powerint,
			prerserr:   prerserr,
			postrserr:  postrserr,
			modulation: downChannelData[4],
			scheme:     "SC-QAM",
		})
	}

	var down31ChannelsData [][]string
	json.Unmarshal([]byte(arr[23]), &down31ChannelsData)
	for index, down31ChannelData := range down31ChannelsData {
		if len(down31ChannelData) != 11 {
			error := fmt.Errorf("Abnormal 3.1 down channel length, expected 11, got %d", len(down31ChannelData))
			errstrings = append(errstrings, error.Error())
			break
		}

		channelID, _ := strconv.Atoi(down31ChannelData[0])
		frequency, _ := strconv.Atoi(down31ChannelData[1] + "000000")
		snr, _ := strconv.ParseFloat(down31ChannelData[7], 64)
		snrint := int(snr * 10)
		power, _ := strconv.ParseFloat(down31ChannelData[8], 64)
		powerint := int(power * 10)
		prerserr, _ := strconv.Atoi(down31ChannelData[9])
		postrserr, _ := strconv.Atoi(down31ChannelData[10])

		if channelID < 1 || channelID > 1024 {
			error := fmt.Errorf("Abnormal channel ID, got %d", channelID)
			errstrings = append(errstrings, error.Error())
			break
		}
		if powerint > 2000 || powerint < -2000 {
			error := fmt.Errorf("Power level for 3.1 channel %d is abnormal, got %d", channelID, powerint)
			errstrings = append(errstrings, error.Error())
			break
		}

		downChannels = append(downChannels, downChannel{
			channelID:  channelID,
			channel:    index + 1,
			frequency:  frequency,
			snr:        snrint,
			power:      powerint,
			prerserr:   prerserr,
			postrserr:  postrserr,
			modulation: down31ChannelData[4],
			scheme:     "OFDM",
		})
	}

	var upChannelsData [][]string
	json.Unmarshal([]byte(arr[21]), &upChannelsData)
	var upChannels []upChannel
	for index, upChannelData := range upChannelsData {
		if len(upChannelData) != 10 {
			error := fmt.Errorf("Abnormal up channel length, expected 10, got %d", len(upChannelData))
			errstrings = append(errstrings, error.Error())
			break
		}

		channelID, _ := strconv.Atoi(upChannelData[0])
		frequency, _ := strconv.Atoi(upChannelData[1])
		power, _ := strconv.ParseFloat(upChannelData[2], 64)
		powerint := int(power * 10)

		if channelID < 1 || channelID > 1024 {
			error := fmt.Errorf("Abnormal channel ID, got %d", channelID)
			errstrings = append(errstrings, error.Error())
			break
		}
		if powerint > 2000 || powerint < -2000 {
			error := fmt.Errorf("Power level for up channel %d is abnormal, got %d", channelID, powerint)
			errstrings = append(errstrings, error.Error())
			break
		}

		upChannels = append(upChannels, upChannel{
			channelID: channelID,
			channel:   index + 1,
			frequency: frequency,
			power:     powerint,
		})
	}

	var returnerr error
	returnerr = nil
	if len(errstrings) > 0 {
		returnerr = errors.New(strings.Join(errstrings, "\n"))
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
	}, returnerr
}

func requestURL4(ip string) string {
	return fmt.Sprintf("http://%s/php/ajaxGet_device_networkstatus_data.php", ip)
}
