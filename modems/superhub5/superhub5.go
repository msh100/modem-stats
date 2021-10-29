package superhub5

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"regexp"
	"strconv"
	"time"

	jsonpatch "github.com/evanphx/json-patch"
	"github.com/msh100/modem-stats/utils"
)

type Modem struct {
	IPAddress string
	Stats     []byte
	FetchTime int64
}

func (sh5 *Modem) ClearStats() {
	sh5.Stats = nil
}

func (sh5 *Modem) Type() string {
	return utils.TypeDocsis
}

func (sh5 *Modem) apiAddress() string {
	if sh5.IPAddress == "" {
		sh5.IPAddress = "192.168.100.1" // TODO: Is this a reasonable default?
	}
	return fmt.Sprintf("http://%s/rest/v1/cablemodem", sh5.IPAddress)
}

type dsChannel struct {
	ID         int     `json:"channelId"`
	Frequency  int     `json:"frequency"`
	Power      float32 `json:"power"`
	Modulation string  `json:"modulation"`
	SNR        int     `json:"snr"`
	PreRS      int     `json:"correctedErrors"`
	PostRS     int     `json:"uncorrectedErrors"`
}

type usChannel struct {
	ID         int     `json:"channelId"`
	Frequency  int     `json:"frequency"`
	Power      float32 `json:"power"`
	Modulation string  `json:"modulation"`
}

type serviceFlow struct {
	ServiceFlow struct {
		ID        int    `json:"serviceFlowId"`
		Direction string `json:"direction"`
		MaxRate   int    `json:"maxTrafficRate"`
		MaxBurst  int    `json:"maxTrafficBurst"`
	} `json:"serviceFlow"`
}

type resultsStruct struct {
	Downstream struct {
		Channels []dsChannel `json:"channels"`
	} `json:"downstream"`
	Upstream struct {
		Channels []usChannel `json:"channels"`
	} `json:"upstream"`
	ServiceFlows []serviceFlow `json:"serviceFlows"`
}

func (sh5 *Modem) ParseStats() (utils.ModemStats, error) {
	if sh5.Stats == nil {
		sh5.Stats = []byte("{}")
		queries := []string{
			sh5.apiAddress() + "/downstream",
			sh5.apiAddress() + "/upstream",
			sh5.apiAddress() + "/serviceflows",
		}

		timeStart := time.Now().UnixNano() / int64(time.Millisecond)
		statsData := utils.BoundedParallelGet(queries, 3)
		sh5.FetchTime = (time.Now().UnixNano() / int64(time.Millisecond)) - timeStart

		for _, query := range statsData {
			stats, err := ioutil.ReadAll(query.Res.Body)
			if err != nil {
				return utils.ModemStats{}, err
			}

			sh5.Stats, err = jsonpatch.MergeMergePatches(sh5.Stats, stats)
			if err != nil {
				return utils.ModemStats{}, err
			}
		}
	}

	var upChannels []utils.ModemChannel
	var downChannels []utils.ModemChannel
	var modemConfigs []utils.ModemConfig

	var results resultsStruct
	json.Unmarshal(sh5.Stats, &results)

	for index, downstream := range results.Downstream.Channels {
		// Note: I have yet to see a D3.1 channel on a Superhub 5 and therefore
		// am not certain this logic even works. It's likely this is broken and
		// needs fixing.
		//
		// As this comes from a map that doesn't guarentee to be in the format
		// of QAM 256, this will probably break.
		re := regexp.MustCompile("[0-9]+")
		qamSize := re.FindString(downstream.Modulation)
		qamSizeInt, err := strconv.Atoi(qamSize)
		if err != nil {
			panic(err)
		}

		scheme := "SC-QAM"
		if qamSizeInt > 256 {
			scheme = "OFDM"
		}

		downChannels = append(downChannels, utils.ModemChannel{
			ChannelID:  downstream.ID,
			Channel:    index + 1,
			Frequency:  downstream.Frequency,
			Snr:        downstream.SNR * 10,
			Power:      int(downstream.Power * 10),
			Prerserr:   downstream.PreRS + downstream.PostRS,
			Postrserr:  downstream.PostRS,
			Modulation: "QAM" + qamSize,
			Scheme:     scheme,
		})
	}

	for index, upstream := range results.Upstream.Channels {
		upChannels = append(upChannels, utils.ModemChannel{
			ChannelID: upstream.ID,
			Channel:   index + 1,
			Frequency: upstream.Frequency,
			Power:     int(upstream.Power * 10),
		})
	}

	for _, modemConfig := range results.ServiceFlows {
		modemConfigs = append(modemConfigs, utils.ModemConfig{
			Config:   modemConfig.ServiceFlow.Direction,
			Maxrate:  modemConfig.ServiceFlow.MaxRate,
			Maxburst: modemConfig.ServiceFlow.MaxBurst,
		})
	}

	return utils.ModemStats{
		Configs:      modemConfigs,
		UpChannels:   upChannels,
		DownChannels: downChannels,
		FetchTime:    sh5.FetchTime,
	}, nil
}
