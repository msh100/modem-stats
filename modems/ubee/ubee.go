package ubee

import (
	"encoding/json"
	"fmt"
	"regexp"

	"github.com/msh100/modem-stats/utils"
)

type Modem struct {
	IPAddress string
	Stats     []byte
	FetchTime int64
}

func (ubee *Modem) ClearStats() {
	ubee.Stats = nil
}

func (ubee *Modem) Type() string {
	return utils.TypeDocsis
}

func (ubee *Modem) fetchURL() string {
	if ubee.IPAddress == "" {
		ubee.IPAddress = "192.168.100.1"
	}
	return fmt.Sprintf("http://%s/htdocs/cm_info_connection.php", ubee.IPAddress)
}

type dsChannel struct {
	Type       int     `json:"ds_type,string"`
	ID         int     `json:"ds_id,string"`
	Frequency  int     `json:"ds_freq,string"`
	Width      int     `json:"ds_width,string"`
	Power      int     `json:"ds_power,string"`
	SNR        float32 `json:"ds_snr,string"`
	Modulation int     `json:"ds_modulation,string"`
	PreRS      int     `json:"ds_correct,string"`
	PostRS     int     `json:"ds_uncorrect,string"`
}

type usChannel struct {
	Status     int `json:"us_status,string"`
	Type       int `json:"us_type,string"`
	ID         int `json:"us_id,string"`
	Frequency  int `json:"us_freq,string"`
	Width      int `json:"us_width,string"`
	Power      int `json:"us_power,string"`
	Modulation int `json:"us_modulation,string"`
}

type resultsStruct struct {
	DownstreamChannels []dsChannel `json:"cm_conn_ds_gourpObj"`
	UpstreamChannels   []usChannel `json:"cm_conn_us_gourpObj"`
}

func (ubee *Modem) ParseStats() (utils.ModemStats, error) {
	if ubee.Stats == nil {
		var err error
		ubee.Stats, ubee.FetchTime, err = utils.SimpleHTTPFetch(ubee.fetchURL())
		if err != nil {
			return utils.ModemStats{}, err
		}
	}

	var re = regexp.MustCompile(`var cm_conn_json = '([^']+)';`)
	match := re.FindAllStringSubmatch(string(ubee.Stats), -1)

	var downChannels []utils.ModemChannel
	var upChannels []utils.ModemChannel

	var results resultsStruct
	json.Unmarshal([]byte(match[0][1]), &results)

	var downModulationMap = []string{
		1: "Unknown",
		2: "OFDM-PLC",
		3: "QAM64",
		4: "QAM256",
		5: "QAM16",
		6: "ALL",
		7: "QAM1024",
		8: "QAM512",
	}

	/*var upModulationMap = []string{
		0: "Unknown",
		1: "TDMA",
		2: "ATDMA",
		3: "SCDMA",
		4: "TDMA AND ATDMA",
	}*/

	var interfaceTypeMap = []string{
		128: "SC-QAM",
		129: "ATDMA",
		277: "OFDM",
		278: "OFDMA",
	}

	for id, downChannelData := range results.DownstreamChannels {
		// For some reason the 3.1 channels are a float that needs to be multiplied by 10
		if downChannelData.Type == 277 {
			downChannelData.SNR = downChannelData.SNR * 10
		}
		snr := int(downChannelData.SNR)

		downChannels = append(downChannels, utils.ModemChannel{
			ChannelID:  downChannelData.ID,
			Channel:    id + 1,
			Frequency:  downChannelData.Frequency,
			Snr:        snr,
			Power:      downChannelData.Power,
			Prerserr:   downChannelData.PreRS,
			Postrserr:  downChannelData.PostRS,
			Modulation: downModulationMap[downChannelData.Modulation],
			Scheme:     interfaceTypeMap[downChannelData.Type],
		})
	}

	for id, upChannelData := range results.UpstreamChannels {
		upChannels = append(upChannels, utils.ModemChannel{
			ChannelID: upChannelData.ID,
			Channel:   id + 1,
			Frequency: upChannelData.Frequency,
			Power:     upChannelData.Power,
		})
	}

	return utils.ModemStats{
		Configs:      []utils.ModemConfig{},
		UpChannels:   upChannels,
		DownChannels: downChannels,
		FetchTime:    ubee.FetchTime,
		ModemType:    utils.TypeDocsis,
	}, nil
}
