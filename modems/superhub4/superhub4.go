package superhub4

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/msh100/modem-stats/utils"
)

type Modem struct {
	IPAddress string
	Stats     []byte
	FetchTime int64
}

func (sh4 *Modem) ClearStats() {
	sh4.Stats = nil
}

func (sh4 *Modem) Type() string {
	return utils.TypeDocsis
}

func (sh4 *Modem) fetchURL() string {
	if sh4.IPAddress == "" {
		sh4.IPAddress = "192.168.100.1"
	}
	return fmt.Sprintf("http://%s/php/ajaxGet_device_networkstatus_data.php", sh4.IPAddress)
}

func (sh4 *Modem) ParseStats() (utils.ModemStats, error) {
	if sh4.Stats == nil {
		var err error
		sh4.Stats, sh4.FetchTime, err = utils.SimpleHTTPFetch(sh4.fetchURL())
		if err != nil {
			return utils.ModemStats{}, err
		}
	}

	var errstrings []string

	var arr []string
	json.Unmarshal([]byte(sh4.Stats), &arr)

	downRate, _ := strconv.Atoi(arr[11])
	downBurst, _ := strconv.Atoi(arr[12])
	upRate, _ := strconv.Atoi(arr[15])
	upBurst, _ := strconv.Atoi(arr[16])

	if downRate == 0 || downBurst == 0 || upRate == 0 || upBurst == 0 {
		error := fmt.Errorf("got nil values for speed config %d,%d,%d,%d", downRate, downBurst, upRate, upBurst)
		errstrings = append(errstrings, error.Error())
	}

	var downChannelsData [][]string
	json.Unmarshal([]byte(arr[20]), &downChannelsData)
	var downChannels []utils.ModemChannel
	for index, downChannelData := range downChannelsData {
		if len(downChannelData) != 9 {
			error := fmt.Errorf("abnormal down channel length, expected 9, got %d", len(downChannelData))
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
			error := fmt.Errorf("abnormal channel ID, got %d", channelID)
			errstrings = append(errstrings, error.Error())
			break
		}
		if powerint > 2000 || powerint < -2000 {
			error := fmt.Errorf("power level for 3.0 channel %d is abnormal, got %d", channelID, powerint)
			errstrings = append(errstrings, error.Error())
			break
		}

		downChannels = append(downChannels, utils.ModemChannel{
			ChannelID:  channelID,
			Channel:    index + 1,
			Frequency:  frequency,
			Snr:        snrint,
			Power:      powerint,
			Prerserr:   prerserr,
			Postrserr:  postrserr,
			Modulation: downChannelData[4],
			Scheme:     "SC-QAM",
		})
	}

	var down31ChannelsData [][]string
	json.Unmarshal([]byte(arr[23]), &down31ChannelsData)
	for index, down31ChannelData := range down31ChannelsData {
		if len(down31ChannelData) != 11 {
			error := fmt.Errorf("abnormal 3.1 down channel length, expected 11, got %d", len(down31ChannelData))
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
			error := fmt.Errorf("abnormal channel ID, got %d", channelID)
			errstrings = append(errstrings, error.Error())
			break
		}
		if powerint > 2000 || powerint < -2000 {
			error := fmt.Errorf("power level for 3.1 channel %d is abnormal, got %d", channelID, powerint)
			errstrings = append(errstrings, error.Error())
			break
		}

		downChannels = append(downChannels, utils.ModemChannel{
			ChannelID:  channelID,
			Channel:    index + 1,
			Frequency:  frequency,
			Snr:        snrint,
			Power:      powerint,
			Prerserr:   prerserr,
			Postrserr:  postrserr,
			Modulation: down31ChannelData[4],
			Scheme:     "OFDM",
		})
	}

	var upChannelsData [][]string
	json.Unmarshal([]byte(arr[21]), &upChannelsData)
	var upChannels []utils.ModemChannel
	for index, upChannelData := range upChannelsData {
		if len(upChannelData) != 10 {
			error := fmt.Errorf("abnormal up channel length, expected 10, got %d", len(upChannelData))
			errstrings = append(errstrings, error.Error())
			break
		}

		channelID, _ := strconv.Atoi(upChannelData[0])
		frequency, _ := strconv.Atoi(upChannelData[1])
		power, _ := strconv.ParseFloat(upChannelData[2], 64)
		powerint := int(power * 10)

		if channelID < 1 || channelID > 1024 {
			error := fmt.Errorf("abnormal channel ID, got %d", channelID)
			errstrings = append(errstrings, error.Error())
			break
		}
		if powerint > 2000 || powerint < -2000 {
			error := fmt.Errorf("power level for up channel %d is abnormal, got %d", channelID, powerint)
			errstrings = append(errstrings, error.Error())
			break
		}

		upChannels = append(upChannels, utils.ModemChannel{
			ChannelID: channelID,
			Channel:   index + 1,
			Frequency: frequency,
			Power:     powerint,
		})
	}

	var returnerr error
	returnerr = nil
	if len(errstrings) > 0 {
		returnerr = errors.New(strings.Join(errstrings, "\n"))
	}

	return utils.ModemStats{
		Configs: []utils.ModemConfig{
			{
				Config:   "downstream",
				Maxrate:  downRate,
				Maxburst: downBurst,
			},
			{
				Config:   "upstream",
				Maxrate:  upRate,
				Maxburst: upBurst,
			},
		},
		UpChannels:   upChannels,
		DownChannels: downChannels,
		FetchTime:    sh4.FetchTime,
	}, returnerr
}
