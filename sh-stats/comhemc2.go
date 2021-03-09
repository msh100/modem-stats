package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"

	gabs "github.com/Jeffail/gabs/v2"
)

type comhemc2 struct {
	IPAddress string
	stats     []byte
	fetchTime int64
	username  string
	password  string
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
	requestMethod := gabsString(actionsObj, "0.method")

	if sagemClient.sessionID == 0 && requestMethod != "logIn" {
		loginRequest := sagemClient.loginRequest()
		_, err := sagemClient.apiRequest(loginRequest)
		if err != nil {
			return nil, err
		}
	}

	sagemClient.requestID = sagemClient.requestID + 1
	sagemClient.currentNonce = randomInt(10000, 50000)

	credentialHash := fmt.Sprintf("%s:%s:%s",
		sagemClient.username,
		sagemClient.serverNonce,
		stringToMD5(sagemClient.password),
	)

	authKey := stringToMD5(
		fmt.Sprintf("%s:%d:%d:JSON:/cgi/json-req",
			stringToMD5(credentialHash),
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

	req, err := http.NewRequest("POST", APIAddress, bytes.NewBuffer(jsonPayload))
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)

	returnValue, _ := gabs.ParseJSON(body)
	returnID := gabsString(returnValue, "reply.actions.0.callbacks.0.parameters.id")
	returnNonce := gabsString(returnValue, "reply.actions.0.callbacks.0.parameters.nonce")

	if returnID != "null" && returnNonce != "null" {
		sagemClient.sessionID, _ = strconv.Atoi(returnID)
		sagemClient.serverNonce = returnNonce
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

func (comhemc2 *comhemc2) ParseStats() (modemStats, error) {
	if comhemc2.stats == nil {
		timeStart := time.Now().UnixNano() / int64(time.Millisecond)

		sagemClient := sagemClient{
			host:         comhemc2.IPAddress,
			username:     comhemc2.username,
			password:     comhemc2.password,
			currentNonce: 0,
			sessionID:    0,
			requestID:    -1,
		}

		channelDataReq, err := sagemClient.getXpaths([]string{
			"Device/Docsis/CableModem/Upstreams",
			"Device/Docsis/CableModem/Downstreams",
		})
		if err != nil {
			return modemStats{}, err
		}

		fetchTime := (time.Now().UnixNano() / int64(time.Millisecond)) - timeStart

		comhemc2.fetchTime = fetchTime
		comhemc2.stats = channelDataReq
	}

	var downChannels []modemChannel
	var upChannels []modemChannel

	jsonParsed, err := gabs.ParseJSON(comhemc2.stats)
	if err != nil {
		return modemStats{}, err
	}

	reply := jsonParsed.Path("reply")
	for _, action := range reply.S("actions").Children() {
		query := gabsString(action, "callbacks.0.xpath")
		channels := action.Path("callbacks.0.parameters")

		if query == "Device/Docsis/CableModem/Upstreams" {
			for _, channelData := range channels.S("value").Children() {
				upChannels = append(upChannels, modemChannel{
					channelID: gabsInt(channelData, "ChannelID"),
					channel:   gabsInt(channelData, "uid"),
					frequency: gabsInt(channelData, "Frequency"),
					power:     int(gabsFloat(channelData, "PowerLevel") * 10),
				})
			}

		} else if query == "Device/Docsis/CableModem/Downstreams" {
			for _, channelData := range channels.S("value").Children() {
				modulation := gabsString(channelData, "Modulation")
				var scheme string
				scheme = "SC-QAM"
				if modulation == "256-QAM/4K-QAM" {
					scheme = "OFDM"
				}

				downChannels = append(downChannels, modemChannel{
					channelID:  gabsInt(channelData, "ChannelID"),
					channel:    gabsInt(channelData, "uid"),
					frequency:  gabsInt(channelData, "Frequency"),
					snr:        int(gabsFloat(channelData, "SNR") * 10),
					power:      int(gabsFloat(channelData, "PowerLevel") * 10),
					prerserr:   gabsInt(channelData, "CorrectableCodewords"),
					postrserr:  gabsInt(channelData, "UncorrectableCodewords"),
					modulation: modulation,
					scheme:     scheme,
				})
			}
		}

	}

	return modemStats{
		upChannels:   upChannels,
		downChannels: downChannels,
		fetchTime:    comhemc2.fetchTime,
	}, nil
}
