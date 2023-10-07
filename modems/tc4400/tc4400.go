package tc4400

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/msh100/modem-stats/utils"
)

type Modem struct {
	IPAddress string
	Stats     []byte
	FetchTime int64
	Username  string
	Password  string
}

func (tc4400 *Modem) ClearStats() {
	tc4400.Stats = nil
}

func (tc4400 *Modem) Type() string {
	return utils.TypeDocsis
}

func (tc4400 *Modem) apiAddress() string {
	if tc4400.IPAddress == "" {
		tc4400.IPAddress = "192.168.100.1"
	}
	return fmt.Sprintf("http://%s/cmconnectionstatus.html", tc4400.IPAddress)
}

func (tc4400 *Modem) getStats() ([]byte, error) {
	if tc4400.Stats == nil {
		req, err := http.NewRequest("GET", tc4400.apiAddress(), nil)
		if err != nil {
			return nil, err
		}
		req.SetBasicAuth(tc4400.Username, tc4400.Password)

		timeStart := time.Now().UnixNano() / int64(time.Millisecond)

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()

		if resp.StatusCode == http.StatusOK {
			bodyBytes, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				return nil, err
			}
			tc4400.Stats = bodyBytes
			tc4400.FetchTime = (time.Now().UnixNano() / int64(time.Millisecond)) - timeStart
		} else {
			return nil, fmt.Errorf("Request failed with status: %s", resp.Status)
		}
	}

	return tc4400.Stats, nil
}

func (tc4400 *Modem) ParseStats() (utils.ModemStats, error) {
	modemStats, err := tc4400.getStats()
	if err != nil {
		return utils.ModemStats{}, err
	}

	var downChannels []utils.ModemChannel
	var upChannels []utils.ModemChannel

	r := bytes.NewReader(modemStats)

	doc, err := goquery.NewDocumentFromReader(r)
	if err != nil {
		return utils.ModemStats{}, err
	}

	doc.Find("table").Eq(1).Find("tr").Each(func(i int, rowHtml *goquery.Selection) {
		if i <= 1 {
			return
		}

		var thisChannel utils.ModemChannel
		var lockStatus string

		rowHtml.Find("td").Each(func(j int, cellHtml *goquery.Selection) {
			cellText := strings.TrimSpace(cellHtml.Text())

			switch j {
			case 0:
				thisChannel.Channel = utils.ExtractIntValue(cellText)
			case 1:
				thisChannel.ChannelID = utils.ExtractIntValue(cellText)
			case 2:
				lockStatus = cellText
			case 3:
				thisChannel.Scheme = cellText
			case 5:
				thisChannel.Frequency = utils.ExtractIntValue(cellText)
			case 7:
				thisChannel.Snr = int(utils.ExtractFloatValue(cellText) * 10)
			case 8:
				thisChannel.Power = int(utils.ExtractFloatValue(cellText) * 10)
			// case 9: // TODO: This is broken on OFDM channels
			// 	thisChannel.Modulation = cellText
			case 11:
				thisChannel.Prerserr = utils.ExtractIntValue(cellText)
			case 13:
				thisChannel.Postrserr = utils.ExtractIntValue(cellText)
			}
		})

		if lockStatus == "Locked" {
			downChannels = append(downChannels, thisChannel)
		}
	})

	doc.Find("table").Eq(2).Find("tr").Each(func(i int, rowHtml *goquery.Selection) {
		if i <= 1 {
			return
		}

		var thisChannel utils.ModemChannel
		var lockStatus string

		rowHtml.Find("td").Each(func(j int, cellHtml *goquery.Selection) {
			cellText := strings.TrimSpace(cellHtml.Text())

			switch j {
			case 0:
				thisChannel.Channel = utils.ExtractIntValue(cellText)
			case 1:
				thisChannel.ChannelID = utils.ExtractIntValue(cellText)
			case 2:
				lockStatus = cellText
			case 3:
				thisChannel.Scheme = cellText
			case 5:
				thisChannel.Frequency = utils.ExtractIntValue(cellText)
			case 7:
				thisChannel.Power = int(utils.ExtractFloatValue(cellText) * 10)
				// case 8: // TODO: This is broken on OFDM channels
				// 	thisChannel.Modulation = cellText
			}
		})

		if lockStatus == "Locked" {
			upChannels = append(upChannels, thisChannel)
		}
	})

	return utils.ModemStats{
		DownChannels: downChannels,
		UpChannels:   upChannels,
	}, nil
}
