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
	requestMethod := actionsObj.Path("0.method").String()

	if sagemClient.sessionID == 0 && requestMethod != "\"logIn\"" {
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

	// Set the server Nonce and session ID from the return values if they're defined
	returnValue, _ := gabs.ParseJSON(body)
	returnID := returnValue.Path("reply.actions.0.callbacks.0.parameters.id").String()
	returnNonce := returnValue.Path("reply.actions.0.callbacks.0.parameters.nonce").String()

	if returnID != "null" && returnNonce != "null" {
		sagemClient.sessionID, _ = strconv.Atoi(returnID)
		sagemClient.serverNonce = strings.ReplaceAll(returnNonce, "\"", "")
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

func (comhemc2 *comhemc2) ParseStats() (routerStats, error) {
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
			return routerStats{}, err
		}

		fetchTime := (time.Now().UnixNano() / int64(time.Millisecond)) - timeStart

		comhemc2.fetchTime = fetchTime
		comhemc2.stats = channelDataReq
	}

	var downChannels []downChannel
	var upChannels []upChannel

	jsonParsed, err := gabs.ParseJSON(comhemc2.stats)
	if err != nil {
		return routerStats{}, err
	}

	reply := jsonParsed.Path("reply")
	for _, action := range reply.S("actions").Children() {
		query := action.Path("callbacks.0.xpath").String()
		query = strings.ReplaceAll(query, "\"", "")
		channels := action.Path("callbacks.0.parameters")

		if query == "Device/Docsis/CableModem/Upstreams" {
			for _, channelData := range channels.S("value").Children() {
				channelID, _ := strconv.Atoi(channelData.Path("ChannelID").String())
				channel, _ := strconv.Atoi(channelData.Path("uid").String())
				frequency, _ := strconv.Atoi(channelData.Path("Frequency").String())
				power, _ := strconv.ParseFloat(channelData.Path("PowerLevel").String(), 64)
				powerInt := int(power * 10)

				upChannels = append(upChannels, upChannel{
					channelID: channelID,
					channel:   channel,
					frequency: frequency,
					power:     powerInt,
				})
			}

		} else if query == "Device/Docsis/CableModem/Downstreams" {
			for _, channelData := range channels.S("value").Children() {
				channelID, _ := strconv.Atoi(channelData.Path("ChannelID").String())
				channel, _ := strconv.Atoi(channelData.Path("uid").String())
				frequency, _ := strconv.Atoi(channelData.Path("Frequency").String())
				snr, _ := strconv.ParseFloat(channelData.Path("SNR").String(), 64)
				snrInt := int(snr * 10)
				power, _ := strconv.ParseFloat(channelData.Path("PowerLevel").String(), 64)
				powerInt := int(power * 10)
				prerserr, _ := strconv.Atoi(channelData.Path("CorrectableCodewords").String())
				postrserr, _ := strconv.Atoi(channelData.Path("UncorrectableCodewords").String())
				modulation := channelData.Path("Modulation").String()
				modulation = strings.ReplaceAll(modulation, "\"", "")

				var scheme string
				scheme = "SC-QAM"
				if modulation == "256-QAM/4K-QAM" {
					scheme = "OFDM"
				}

				downChannels = append(downChannels, downChannel{
					channelID:  channelID,
					channel:    channel,
					frequency:  frequency,
					snr:        snrInt,
					power:      powerInt,
					prerserr:   prerserr,
					postrserr:  postrserr,
					modulation: modulation,
					scheme:     scheme,
				})
			}
		}

	}

	return routerStats{
		upChannels:   upChannels,
		downChannels: downChannels,
		fetchTime:    comhemc2.fetchTime,
	}, nil
}
