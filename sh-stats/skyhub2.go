package main

import (
	"fmt"
	"regexp"
	"strconv"
)

type skyhub2 struct {
	IPAddress string
	stats     []byte
	fetchTime int64
	username  string
	password  string
}

func (sky2 *skyhub2) ClearStats() {
	sky2.stats = nil
}

func (sky2 *skyhub2) fetchURL() string {
	return fmt.Sprintf(
		"http://%s:%s@%s/Now_TV_system.html",
		sky2.username,
		sky2.password,
		sky2.IPAddress,
	)
}

func (sky2 *skyhub2) ParseStats() (modemStats, error) {
	if sky2.stats == nil {
		var err error
		sky2.stats, sky2.fetchTime, err = simpleHTTPFetch(sky2.fetchURL())
		if err != nil {
			return modemStats{}, err
		}
	}

	syncRegex := regexp.MustCompile("Connection Speed \\(Kbps\\)<\\/td><td>([0-9]+)<\\/td><td>([0-9]+)")
	syncRate := syncRegex.FindStringSubmatch(string(sky2.stats))

	downSync, _ := strconv.Atoi(syncRate[1])
	upSync, _ := strconv.Atoi(syncRate[2])

	configs := []modemConfig{
		{
			config:  "upstream",
			maxrate: upSync,
		},
		{
			config:  "downstream",
			maxrate: downSync,
		},
	}

	downChannelRegex := regexp.MustCompile("DS([0-9]*):([0-9]*).([0-9]*)")
	upChannelRegex := regexp.MustCompile("US([0-9]*):([0-9]*).([0-9]*)")

	attenuationRegex := regexp.MustCompile("<td>Line Attenuation \\(dB\\)<\\/td><td>.+?</tr>")
	attenuationData := attenuationRegex.FindStringSubmatch(string(sky2.stats))

	noiseRegex := regexp.MustCompile("<td>Noise Margin \\(dB\\)<\\/td><td>.+?</tr>")
	noiseData := noiseRegex.FindStringSubmatch(string(sky2.stats))

	downAttenuation := downChannelRegex.FindAllStringSubmatch(attenuationData[0], -1)
	downNoise := downChannelRegex.FindAllStringSubmatch(noiseData[0], -1)
	upAttenuation := upChannelRegex.FindAllStringSubmatch(attenuationData[0], -1)
	upNoise := upChannelRegex.FindAllStringSubmatch(noiseData[0], -1)

	var downChannels []modemChannel
	var upChannels []modemChannel

	for _, v := range downAttenuation {
		channelID, _ := strconv.Atoi(v[1])
		attenuation, _ := strconv.Atoi(fmt.Sprintf("%s%s", v[2], v[3]))
		downChannels = append(downChannels, modemChannel{
			channelID:   channelID,
			attenuation: attenuation,
		})
	}
	for _, v := range downNoise {
		channelID, _ := strconv.Atoi(v[1])
		noise, _ := strconv.Atoi(fmt.Sprintf("%s%s", v[2], v[3]))
		downChannels[channelID-1].noise = noise
	}
	for _, v := range upAttenuation {
		channelID, _ := strconv.Atoi(v[1])
		attenuation, _ := strconv.Atoi(fmt.Sprintf("%s%s", v[2], v[3]))
		upChannels = append(upChannels, modemChannel{
			channelID:   channelID,
			attenuation: attenuation,
		})
	}
	for _, v := range upNoise {
		channelID, _ := strconv.Atoi(v[1])
		noise, _ := strconv.Atoi(fmt.Sprintf("%s%s", v[2], v[3]))
		upChannels[channelID].noise = noise
	}

	return modemStats{
		configs:      configs,
		downChannels: downChannels,
		upChannels:   upChannels,
		fetchTime:    sky2.fetchTime,
		modemType:    "VDSL",
	}, nil
}
