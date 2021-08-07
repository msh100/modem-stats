package comhemc2

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"

	gabs "github.com/Jeffail/gabs/v2"

	"github.com/msh100/modem-stats/utils"
)

type Modem struct {
	IPAddress   string
	Stats       []byte
	FetchTime   int64
	Username    string
	Password    string
	sagemClient *sagemClient
}

type sagemClient struct {
	host     string
	username string
	password string

	currentNonce int
	serverNonce  string
	sessionID    int
	requestID    int
}

func (sagemClient *sagemClient) loginRequest() string {
	request := `[
		{
		  "method": "logIn",
		  "parameters": {
			"user": "%s",
			"persistent": "true",
			"session-options": {
			  "nss": [
				{
				  "name": "gtw",
				  "uri": "http://sagemcom.com/gateway-data"
				}
			  ],
			  "language": "ident",
			  "context-flags": {
				"get-content-name": true,
				"local-time": true
			  },
			  "capability-depth": 2,
			  "capability-flags": {
				"name": true,
				"default-value": false,
				"restriction": true,
				"description": false
			  },
			  "time-format": "ISO_8601",
			  "write-only-string": "_XMO_WRITE_ONLY_",
			  "undefined-write-only-string": "_XMO_UNDEFINED_WRITE_ONLY_"
			}
		  }
		}
	  ]`

	return fmt.Sprintf(request, sagemClient.username)
}

func (sagemClient *sagemClient) apiRequest(actions string) ([]byte, error) {
	actionsObj, _ := gabs.ParseJSON([]byte(actions))
	requestMethod := utils.GabsString(actionsObj, "0.method")
	if sagemClient.sessionID == 0 && requestMethod != "logIn" {
		_, err := sagemClient.apiRequest(sagemClient.loginRequest())
		if err != nil {
			return nil, err
		}
	}

	sagemClient.requestID = sagemClient.requestID + 1
	sagemClient.currentNonce = utils.RandomInt(10000, 50000)

	credentialHash := fmt.Sprintf("%s:%s:%s",
		sagemClient.username,
		sagemClient.serverNonce,
		utils.StringToMD5(sagemClient.password),
	)

	authKey := utils.StringToMD5(
		fmt.Sprintf("%s:%d:%d:JSON:/cgi/json-req",
			utils.StringToMD5(credentialHash),
			sagemClient.requestID,
			sagemClient.currentNonce,
		),
	)

	APIAddress := fmt.Sprintf("http://%s/cgi/json-req", sagemClient.host)

	payloadObj := gabs.New()
	payloadObj.Set(sagemClient.requestID, "request", "id")
	payloadObj.Set(sagemClient.sessionID, "request", "session-id")
	payloadObj.Set("False", "request", "priority")
	payloadObj.Set(actionsObj, "request", "actions")
	if sagemClient.currentNonce > 0 {
		payloadObj.Set(fmt.Sprintf("%d", sagemClient.currentNonce), "request", "cnonce")
	}
	payloadObj.Set(authKey, "request", "auth-key")
	jsonPayload := []byte(fmt.Sprintf("req=%s", payloadObj.String()))

	req, _ := http.NewRequest("POST", APIAddress, bytes.NewBuffer(jsonPayload))
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)

	returnValue, _ := gabs.ParseJSON(body)
	returnID := utils.GabsString(returnValue, "reply.actions.0.callbacks.0.parameters.id")
	returnNonce := utils.GabsString(returnValue, "reply.actions.0.callbacks.0.parameters.nonce")
	errorCode := utils.GabsString(returnValue, "reply.error.code")

	if returnID != "null" && returnNonce != "null" {
		sagemClient.sessionID, _ = strconv.Atoi(returnID)
		sagemClient.serverNonce = returnNonce
	}
	if errorCode == "16777219" { // Session has expired
		sagemClient.sessionID = 0
		sagemClient.currentNonce = 0
		sagemClient.requestID = -1
		sagemClient.serverNonce = ""

		_, err := sagemClient.apiRequest(sagemClient.loginRequest())
		if err != nil {
			return nil, err
		}

		return sagemClient.apiRequest(actions)
	}

	return body, nil
}

func (sagemClient *sagemClient) getXpaths(xpaths []string) ([]byte, error) {
	xpathTpl := `{"id":%d,"method":"getValue","xpath":"%s","options":{}}`
	outputXpaths := []string{}

	for id, xpath := range xpaths {
		outputXpaths = append(outputXpaths, fmt.Sprintf(xpathTpl, id, xpath))
	}
	xpathReq := fmt.Sprintf("[%s]", strings.Join(outputXpaths, ","))

	return sagemClient.apiRequest(xpathReq)
}

func (comhemc2 *Modem) ClearStats() {
	comhemc2.Stats = nil
}

func (comhemc2 *Modem) Type() string {
	return utils.TypeDocsis
}

func (comhemc2 *Modem) ParseStats() (utils.ModemStats, error) {
	if comhemc2.Stats == nil {
		timeStart := time.Now().UnixNano() / int64(time.Millisecond)

		if comhemc2.sagemClient == nil {
			if comhemc2.Username == "" {
				comhemc2.Username = "admin"
			}
			if comhemc2.Password == "" {
				comhemc2.Password = "admin"
			}
			if comhemc2.IPAddress == "" {
				comhemc2.IPAddress = "192.168.10.1"
			}
			comhemc2.sagemClient = &sagemClient{
				host:         comhemc2.IPAddress,
				username:     comhemc2.Username,
				password:     comhemc2.Password,
				currentNonce: 0,
				sessionID:    0,
				requestID:    -1,
			}
		}

		channelDataReq, err := comhemc2.sagemClient.getXpaths([]string{
			"Device/Docsis/CableModem/Upstreams",
			"Device/Docsis/CableModem/Downstreams",
		})
		if err != nil {
			return utils.ModemStats{}, err
		}

		fetchTime := (time.Now().UnixNano() / int64(time.Millisecond)) - timeStart

		comhemc2.FetchTime = fetchTime
		comhemc2.Stats = channelDataReq
	}

	var downChannels []utils.ModemChannel
	var upChannels []utils.ModemChannel

	jsonParsed, err := gabs.ParseJSON(comhemc2.Stats)
	if err != nil {
		return utils.ModemStats{}, err
	}

	reply := jsonParsed.Path("reply")
	for _, action := range reply.S("actions").Children() {
		query := utils.GabsString(action, "callbacks.0.xpath")
		channels := action.Path("callbacks.0.parameters")

		if query == "Device/Docsis/CableModem/Upstreams" {
			for _, channelData := range channels.S("value").Children() {
				upChannels = append(upChannels, utils.ModemChannel{
					ChannelID: utils.GabsInt(channelData, "ChannelID"),
					Channel:   utils.GabsInt(channelData, "uid"),
					Frequency: utils.GabsInt(channelData, "Frequency"),
					Power:     int(utils.GabsFloat(channelData, "PowerLevel") * 10),
				})
			}

		} else if query == "Device/Docsis/CableModem/Downstreams" {
			for _, channelData := range channels.S("value").Children() {
				modulation := utils.GabsString(channelData, "Modulation")
				var scheme string
				scheme = "SC-QAM"
				if modulation == "256-QAM/4K-QAM" {
					scheme = "OFDM"
				}

				downChannels = append(downChannels, utils.ModemChannel{
					ChannelID:  utils.GabsInt(channelData, "ChannelID"),
					Channel:    utils.GabsInt(channelData, "uid"),
					Frequency:  utils.GabsInt(channelData, "Frequency"),
					Snr:        int(utils.GabsFloat(channelData, "SNR") * 10),
					Power:      int(utils.GabsFloat(channelData, "PowerLevel") * 10),
					Prerserr:   utils.GabsInt(channelData, "CorrectableCodewords"),
					Postrserr:  utils.GabsInt(channelData, "UncorrectableCodewords"),
					Modulation: modulation,
					Scheme:     scheme,
				})
			}
		}

	}

	return utils.ModemStats{
		UpChannels:   upChannels,
		DownChannels: downChannels,
		FetchTime:    comhemc2.FetchTime,
	}, nil
}
