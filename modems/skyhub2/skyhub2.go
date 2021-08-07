package skyhub2

import (
	"fmt"
	"regexp"
	"strconv"

	"github.com/msh100/modem-stats/utils"
)

type Modem struct {
	IPAddress string
	Stats     []byte
	FetchTime int64
	Username  string
	Password  string
}

func (sky2 *Modem) ClearStats() {
	sky2.Stats = nil
}

func (sky2 *Modem) Type() string {
	return utils.TypeVDSL
}

func (sky2 *Modem) fetchURL() string {
	if sky2.Username == "" {
		sky2.Username = "admin"
	}
	if sky2.Password == "" {
		sky2.Password = "sky"
	}
	if sky2.IPAddress == "" {
		sky2.IPAddress = "192.168.0.1"
	}
	return fmt.Sprintf(
		"http://%s:%s@%s/Now_TV_system.html",
		sky2.Username,
		sky2.Password,
		sky2.IPAddress,
	)
}

func (sky2 *Modem) ParseStats() (utils.ModemStats, error) {
	if sky2.Stats == nil {
		var err error
		sky2.Stats, sky2.FetchTime, err = utils.SimpleHTTPFetch(sky2.fetchURL())
		if err != nil {
			return utils.ModemStats{}, err
		}
	}

	syncRegex := regexp.MustCompile("Connection Speed \\(Kbps\\)<\\/td><td>([0-9]+)<\\/td><td>([0-9]+)")
	syncRate := syncRegex.FindStringSubmatch(string(sky2.Stats))

	downSync, _ := strconv.Atoi(syncRate[1])
	upSync, _ := strconv.Atoi(syncRate[2])

	configs := []utils.ModemConfig{
		{
			Config:  "upstream",
			Maxrate: upSync,
		},
		{
			Config:  "downstream",
			Maxrate: downSync,
		},
	}

	downChannelRegex := regexp.MustCompile("DS([0-9]*):([0-9]*).([0-9]*)")
	upChannelRegex := regexp.MustCompile("US([0-9]*):([0-9]*).([0-9]*)")

	attenuationRegex := regexp.MustCompile("<td>Line Attenuation \\(dB\\)<\\/td><td>.+?</tr>")
	attenuationData := attenuationRegex.FindStringSubmatch(string(sky2.Stats))

	noiseRegex := regexp.MustCompile("<td>Noise Margin \\(dB\\)<\\/td><td>.+?</tr>")
	noiseData := noiseRegex.FindStringSubmatch(string(sky2.Stats))

	downAttenuation := downChannelRegex.FindAllStringSubmatch(attenuationData[0], -1)
	downNoise := downChannelRegex.FindAllStringSubmatch(noiseData[0], -1)
	upAttenuation := upChannelRegex.FindAllStringSubmatch(attenuationData[0], -1)
	upNoise := upChannelRegex.FindAllStringSubmatch(noiseData[0], -1)

	var downChannels []utils.ModemChannel
	var upChannels []utils.ModemChannel

	for _, v := range downAttenuation {
		channelID, _ := strconv.Atoi(v[1])
		attenuation, _ := strconv.Atoi(fmt.Sprintf("%s%s", v[2], v[3]))
		downChannels = append(downChannels, utils.ModemChannel{
			ChannelID:   channelID,
			Attenuation: attenuation,
		})
	}
	for _, v := range downNoise {
		channelID, _ := strconv.Atoi(v[1])
		noise, _ := strconv.Atoi(fmt.Sprintf("%s%s", v[2], v[3]))
		downChannels[channelID-1].Noise = noise
	}
	for _, v := range upAttenuation {
		channelID, _ := strconv.Atoi(v[1])
		attenuation, _ := strconv.Atoi(fmt.Sprintf("%s%s", v[2], v[3]))
		upChannels = append(upChannels, utils.ModemChannel{
			ChannelID:   channelID,
			Attenuation: attenuation,
		})
	}
	for _, v := range upNoise {
		channelID, _ := strconv.Atoi(v[1])
		noise, _ := strconv.Atoi(fmt.Sprintf("%s%s", v[2], v[3]))
		upChannels[channelID].Noise = noise
	}

	return utils.ModemStats{
		Configs:      configs,
		DownChannels: downChannels,
		UpChannels:   upChannels,
		FetchTime:    sky2.FetchTime,
		ModemType:    utils.TypeVDSL,
	}, nil
}
