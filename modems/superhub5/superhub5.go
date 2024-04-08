package superhub5

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"regexp"
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
	ID          int     `json:"channelId"`
	Frequency   int     `json:"frequency"`
	Power       float32 `json:"power"`
	Modulation  string  `json:"modulation"`
	SNR         int     `json:"snr"`
	PreRS       int     `json:"correctedErrors"`
	PostRS      int     `json:"uncorrectedErrors"`
	ChannelType string  `json:"channelType"`
	RxMer       int     `json:"rxMer"`
}

type usChannel struct {
	ID          int     `json:"channelId"`
	Frequency   int     `json:"frequency"`
	Power       float32 `json:"power"`
	Modulation  string  `json:"modulation"`
	ChannelType string  `json:"channelType"`
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
		re := regexp.MustCompile("[0-9]+")
		qamSize := re.FindString(downstream.Modulation)

		powerInt := int(downstream.Power * 10)
		snr := downstream.SNR * 10

		var scheme string
		if downstream.ChannelType == "sc_qam" {
			scheme = "SC-QAM"
		} else if downstream.ChannelType == "ofdm" {
			scheme = "OFDM"
			powerInt = int(downstream.Power)
			snr = downstream.RxMer
		} else {
			fmt.Println("Unknown channel scheme:", downstream.ChannelType)
			continue
		}

		downChannels = append(downChannels, utils.ModemChannel{
			ChannelID:  downstream.ID,
			Channel:    index + 1,
			Frequency:  downstream.Frequency,
			Snr:        snr,
			Power:      powerInt,
			Prerserr:   downstream.PreRS + downstream.PostRS,
			Postrserr:  downstream.PostRS,
			Modulation: "QAM" + qamSize,
			Scheme:     scheme,
		})
	}

	for index, upstream := range results.Upstream.Channels {
		powerInt := int(upstream.Power * 10)

		var scheme string
		if upstream.ChannelType == "atdma" {
			scheme = "ATDMA"
		} else if upstream.ChannelType == "ofdma" {
			scheme = "OFDMA"
			powerInt = int(upstream.Power)
		} else {
			fmt.Println("Unknown channel scheme:", upstream.ChannelType)
			continue
		}

		upChannels = append(upChannels, utils.ModemChannel{
			ChannelID: upstream.ID,
			Channel:   index + 1,
			Frequency: upstream.Frequency,
			Power:     powerInt,
			Scheme:    scheme,
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
