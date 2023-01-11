package superhub3

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/msh100/modem-stats/utils"
)

type Modem struct {
	IPAddress string
	Stats     []byte
	FetchTime int64
}

func (sh3 *Modem) ClearStats() {
	sh3.Stats = nil
}

func (sh3 *Modem) Type() string {
	return utils.TypeDocsis
}

func (sh3 *Modem) fetchURL() string {
	if sh3.IPAddress == "" {
		sh3.IPAddress = "192.168.100.1"
	}
	return fmt.Sprintf("http://%s/getRouterStatus", sh3.IPAddress)
}

func (sh3 *Modem) activeChannels() ([]int, []int) {
	downChannelRegex := regexp.MustCompile("\"1.3.6.1.2.1.10.127.1.1.1.1.1.([0-9]+)\"")
	upChannelRegex := regexp.MustCompile("\"1.3.6.1.2.1.10.127.1.1.2.1.1.([0-9]+)\"")

	downChannels := []int{}
	upChannels := []int{}

	for _, value := range downChannelRegex.FindAllStringSubmatch(string(sh3.Stats), -1) {
		channelID, _ := strconv.Atoi(value[1])
		downChannels = append(downChannels, channelID)
	}
	for _, value := range upChannelRegex.FindAllStringSubmatch(string(sh3.Stats), -1) {
		channelID, _ := strconv.Atoi(value[1])
		upChannels = append(upChannels, channelID)
	}

	return downChannels, upChannels
}

func (sh3 *Modem) activeConfigs() (int, int) {
	activeConfigRegex := regexp.MustCompile("\"1.3.6.1.4.1.4491.2.1.21.1.3.1.8.2.[0-9]+.([0-9]+)\":\"1\"")
	configRegex := regexp.MustCompile("\"1.3.6.1.4.1.4491.2.1.21.1.3.1.7.2.[0-9]+.([0-9]+)\":\"([1-2])\"")

	downConfig := 0
	upConfig := 0

	// If the JSON for the stats input has been beautified, this regex wont work
	statsData := strings.ReplaceAll(string(sh3.Stats), "\": \"", "\":\"")

	var upConfigs []int
	var downConfigs []int
	for _, value := range configRegex.FindAllStringSubmatch(statsData, -1) {
		thisConfig, _ := strconv.Atoi(value[1])
		thisDirection, _ := strconv.Atoi(value[2])

		if thisDirection == 1 {
			downConfigs = append(downConfigs, thisConfig)
		} else {
			upConfigs = append(upConfigs, thisConfig)
		}
	}

	for _, value := range activeConfigRegex.FindAllStringSubmatch(statsData, -1) {
		thisConfig, _ := strconv.Atoi(value[1])

		for _, testConfig := range upConfigs {
			if testConfig == thisConfig {
				upConfig = testConfig
				break
			}
		}
		for _, testConfig := range downConfigs {
			if testConfig == thisConfig {
				downConfig = testConfig
				break
			}
		}
	}

	return downConfig, upConfig
}

func (sh3 *Modem) readMIBInt(json map[string]interface{}, mib string, channel int) int {
	MIB := fmt.Sprintf("%s.%d", mib, channel)
	outputInt := -1
	if json[MIB] != nil {
		outputInt, _ = strconv.Atoi(json[MIB].(string))
	}
	return outputInt
}

func (sh3 *Modem) dataAsJSON() map[string]interface{} {
	var snmpData map[string]interface{}
	json.Unmarshal(sh3.Stats, &snmpData)
	return snmpData
}

func (sh3 *Modem) ParseStats() (utils.ModemStats, error) {
	if sh3.Stats == nil {
		var err error
		sh3.Stats, sh3.FetchTime, err = utils.SimpleHTTPFetch(sh3.fetchURL())
		if err != nil {
			return utils.ModemStats{}, err
		}
	}

	downChannelIDs, upChannelIDs := sh3.activeChannels()
	downConfigID, upConfigID := sh3.activeConfigs()

	snmpData := sh3.dataAsJSON()

	var downChannels []utils.ModemChannel
	var upChannels []utils.ModemChannel

	for _, channel := range downChannelIDs {
		channelID := sh3.readMIBInt(snmpData, "1.3.6.1.2.1.10.127.1.1.1.1.1", channel)
		frequency := sh3.readMIBInt(snmpData, "1.3.6.1.2.1.10.127.1.1.1.1.2", channel)
		snr := sh3.readMIBInt(snmpData, "1.3.6.1.2.1.10.127.1.1.4.1.5", channel)
		power := sh3.readMIBInt(snmpData, "1.3.6.1.2.1.10.127.1.1.1.1.6", channel)
		preRSErr := sh3.readMIBInt(snmpData, "1.3.6.1.2.1.10.127.1.1.4.1.3", channel)
		postRSErr := sh3.readMIBInt(snmpData, "1.3.6.1.2.1.10.127.1.1.4.1.4", channel)

		downChannels = append(downChannels, utils.ModemChannel{
			ChannelID:  channelID,
			Channel:    channel,
			Frequency:  frequency,
			Snr:        snr,
			Power:      power,
			Prerserr:   preRSErr,
			Postrserr:  postRSErr,
			Modulation: "QAM256",
			Scheme:     "SC-QAM",
		})
	}

	for _, channel := range upChannelIDs {
		channelID := sh3.readMIBInt(snmpData, "1.3.6.1.2.1.10.127.1.1.2.1.1", channel)
		frequency := sh3.readMIBInt(snmpData, "1.3.6.1.2.1.10.127.1.1.2.1.2", channel)
		power := sh3.readMIBInt(snmpData, "1.3.6.1.4.1.4491.2.1.20.1.2.1.1", channel)

		upChannels = append(upChannels, utils.ModemChannel{
			ChannelID: channelID,
			Channel:   channel,
			Frequency: frequency,
			Power:     power,
		})
	}

	downConfigRate := sh3.readMIBInt(snmpData, "1.3.6.1.4.1.4491.2.1.21.1.2.1.6.2.1", downConfigID)
	downConfigBurst := sh3.readMIBInt(snmpData, "1.3.6.1.4.1.4491.2.1.21.1.2.1.7.2.1", downConfigID)
	upConfigRate := sh3.readMIBInt(snmpData, "1.3.6.1.4.1.4491.2.1.21.1.2.1.6.2.1", upConfigID)
	upConfigBurst := sh3.readMIBInt(snmpData, "1.3.6.1.4.1.4491.2.1.21.1.2.1.7.2.1", upConfigID)

	return utils.ModemStats{
		Configs: []utils.ModemConfig{
			{
				Config:   "upstream",
				Maxrate:  upConfigRate,
				Maxburst: upConfigBurst,
			},
			{
				Config:   "downstream",
				Maxrate:  downConfigRate,
				Maxburst: downConfigBurst,
			},
		},
		UpChannels:   upChannels,
		DownChannels: downChannels,
		FetchTime:    sh3.FetchTime,
		ModemType:    utils.TypeDocsis,
	}, nil
}
