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

func (comhemc2 comhemc2) fetchURL() string {
	return fmt.Sprintf("http://%s/0.1/gui/#/internetConnectivity/docsis/rf-parameters", comhemc2.IPAddress)
}

/*
curl 'http://192.168.10.1/cgi/json-req' \
  -H 'Cookie: lang=en; session=%7B%22req_id%22%3A173%2C%22sess_id%22%3A1639072317%2C%22basic%22%3Afalse%2C%22user%22%3A%22admin%22%2C%22dataModel%22%3A%7B%22name%22%3A%22Internal%22%2C%22nss%22%3A%5B%7B%22name%22%3A%22gtw%22%2C%22uri%22%3A%22http%3A%2F%2Fsagemcom.com%2Fgateway-data%22%7D%5D%7D%2C%22ha1%22%3A%229f8c33b5fd389ff33df351941149bf42c0e8058dfc5bded0102ab4472538dc95%22%2C%22nonce%22%3A%222585743033%22%7D' \
  --data-raw 'req=%7B%22request%22%3A%7B%22id%22%3A172%2C%22session-id%22%3A1639072317%2C%22priority%22%3Afalse%2C%22actions%22%3A%5B%7B%22id%22%3A0%2C%22method%22%3A%22getValue%22%2C%22xpath%22%3A%22Device%2FDocsis%2FCableModem%2FUpstreams%22%2C%22options%22%3A%7B%22capability-flags%22%3A%7B%22interface%22%3Atrue%7D%7D%7D%2C%7B%22id%22%3A1%2C%22method%22%3A%22getValue%22%2C%22xpath%22%3A%22Device%2FDocsis%2FCableModem%2FDownstreams%22%2C%22options%22%3A%7B%22capability-flags%22%3A%7B%22interface%22%3Atrue%7D%7D%7D%2C%7B%22id%22%3A2%2C%22method%22%3A%22getValue%22%2C%22xpath%22%3A%22Device%2FDocsis%2FCableModem%2FForceScanFreq%22%2C%22options%22%3A%7B%22capability-flags%22%3A%7B%22interface%22%3Atrue%7D%7D%7D%5D%2C%22cnonce%22%3A4186780168%2C%22auth-key%22%3A%2216607082dca5199b3d377047c4ef4b4b%22%7D%7D'
*/

type sagemClient struct {
	host     string
	username string
	password string

	currentNonce string
	serverNonce  string
	sessionID    int
	requestID    int
}

func (sagemClient *sagemClient) apiRequest(actions string) []byte {
	sagemClient.requestID = sagemClient.requestID + 1
	sagemClient.currentNonce = fmt.Sprintf("%d", (randomInt(10000, 50000)))

	credentialHash := fmt.Sprintf("%s:%s:%s",
		sagemClient.username,
		sagemClient.serverNonce,
		stringToMD5(sagemClient.password),
	)

	authKey := stringToMD5(
		fmt.Sprintf("%s:%d:%s:JSON:/cgi/json-req",
			stringToMD5(credentialHash),
			sagemClient.requestID,
			sagemClient.currentNonce,
		),
	)

	APIAddress := fmt.Sprintf("http://%s/cgi/json-req", sagemClient.host)

	payloadObj := gabs.New()
	actionsObj, _ := gabs.ParseJSON([]byte(actions))

	payloadObj.Set(sagemClient.requestID, "request", "id")
	payloadObj.Set(sagemClient.sessionID, "request", "session-id")
	payloadObj.Set("False", "request", "priority")
	payloadObj.Set(actionsObj, "request", "actions")
	payloadObj.Set(sagemClient.currentNonce, "request", "cnonce")
	payloadObj.Set(authKey, "request", "auth-key")

	jsonPayload := []byte(fmt.Sprintf("req=%s", payloadObj.String()))

	req, err := http.NewRequest("POST", APIAddress, bytes.NewBuffer(jsonPayload))

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)

	// Set the server Nonce and session ID from the return values if they're defined
	returnValue, _ := gabs.ParseJSON(body)
	returnID := fmt.Sprintf(
		"%s",
		returnValue.Path("reply.actions.0.callbacks.0.parameters.id").String(),
	)
	returnNonce := fmt.Sprintf(
		"%s",
		returnValue.Path("reply.actions.0.callbacks.0.parameters.nonce").String(),
	)

	if returnID != "null" && returnNonce != "null" {
		sagemClient.sessionID, _ = strconv.Atoi(returnID)
		sagemClient.serverNonce = strings.ReplaceAll(returnNonce, "\"", "")

	}

	return body
}

func (comhemc2 *comhemc2) ParseStats() (routerStats, error) {

	if comhemc2.stats == nil {
		timeStart := time.Now().UnixNano() / int64(time.Millisecond)

		sagemClient := sagemClient{
			host:         comhemc2.IPAddress,
			username:     comhemc2.username,
			password:     comhemc2.password,
			currentNonce: "",
			sessionID:    0,
			requestID:    -1,
		}

		loginRequest := `[{"method":"logIn","parameters":{"user":"%s","persistent":"true","session-options":{"nss":[{"name":"gtw","uri":"http://sagemcom.com/gateway-data"}],"language":"ident","context-flags":{"get-content-name":true,"local-time":true},"capability-depth":2,"capability-flags":{"name":true,"default-value":false,"restriction":true,"description":false},"time-format":"ISO_8601","write-only-string":"_XMO_WRITE_ONLY_","undefined-write-only-string":"_XMO_UNDEFINED_WRITE_ONLY_"}}}]`
		sagemClient.apiRequest(fmt.Sprintf(loginRequest, sagemClient.username))

		channelsXpath := `[{"id":0,"method":"getValue","xpath":"%s","options":{}},{"id":1,"method":"getValue","xpath":"%s","options":{}}]`
		channelsXpath = fmt.Sprintf(channelsXpath, "Device/Docsis/CableModem/Upstreams", "Device/Docsis/CableModem/Downstreams")
		channelDataReq := sagemClient.apiRequest(channelsXpath)

		fetchTime := (time.Now().UnixNano() / int64(time.Millisecond)) - timeStart

		comhemc2.fetchTime = fetchTime
		comhemc2.stats = channelDataReq
	}

	var downChannels []downChannel
	var upChannels []upChannel

	jsonParsed, err := gabs.ParseJSON(comhemc2.stats)
	if err != nil {
		fmt.Println(err)
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
				snr, _ := strconv.Atoi(channelData.Path("SNR").String())
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
					snr:        snr,
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
	}, nil
}
