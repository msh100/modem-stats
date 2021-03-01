package main

import (
	"crypto/md5"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/Jeffail/gabs/v2"
)

func simpleHTTPFetch(url string) ([]byte, int64, error) {
	timeStart := time.Now().UnixNano() / int64(time.Millisecond)
	resp, err := http.Get(url)
	if err != nil {
		return nil, 0, err
	}

	stats, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, 0, err
	}
	fetchTime := (time.Now().UnixNano() / int64(time.Millisecond)) - timeStart
	return stats, fetchTime, nil
}

func randomInt(min int, max int) int {
	rand.Seed(time.Now().UnixNano())
	random := rand.Intn(max-min) + min
	return random
}

func stringToMD5(input string) string {
	return fmt.Sprintf("%x", md5.Sum([]byte(input)))
}

func gabsInt(input *gabs.Container, path string) int {
	output, _ := strconv.Atoi(input.Path(path).String())
	return output
}

func gabsFloat(input *gabs.Container, path string) float64 {
	output, _ := strconv.ParseFloat(input.Path(path).String(), 64)
	return output
}

func gabsString(input *gabs.Container, path string) string {
	output := input.Path(path).String()
	return strings.Trim(output, "\"")
}
