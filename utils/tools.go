package utils

import (
	"crypto/md5"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/Jeffail/gabs/v2"
)

func SimpleHTTPFetch(url string) ([]byte, int64, error) {
	timeStart := time.Now().UnixNano() / int64(time.Millisecond)
	resp, err := http.Get(url)
	if err != nil {
		return nil, 0, err
	}
	if resp.StatusCode != 200 {
		return nil, 0, fmt.Errorf("%d status code recieved", resp.StatusCode)
	}

	stats, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, 0, err
	}
	fetchTime := (time.Now().UnixNano() / int64(time.Millisecond)) - timeStart
	return stats, fetchTime, nil
}

func RandomInt(min int, max int) int {
	rand.Seed(time.Now().UnixNano())
	random := rand.Intn(max-min) + min
	return random
}

func StringToMD5(input string) string {
	return fmt.Sprintf("%x", md5.Sum([]byte(input)))
}

func GabsInt(input *gabs.Container, path string) int {
	output, _ := strconv.Atoi(input.Path(path).String())
	return output
}

func GabsFloat(input *gabs.Container, path string) float64 {
	output, _ := strconv.ParseFloat(input.Path(path).String(), 64)
	return output
}

func GabsString(input *gabs.Container, path string) string {
	output := input.Path(path).String()
	return strings.Trim(output, "\"")
}

func Getenv(key, fallback string) string {
	value := os.Getenv(key)
	if len(value) == 0 {
		return fallback
	}
	return value
}

func FetchStats(router DocsisModem) (ModemStats, error) {
	stats, err := router.ParseStats()
	return stats, err
}

func ResetStats(router DocsisModem) {
	router.ClearStats()
}

type HttpResult struct {
	Index int
	Res   http.Response
	Err   error
}

func BoundedParallelGet(urls []string, concurrencyLimit int) []HttpResult {
	semaphoreChan := make(chan struct{}, concurrencyLimit)
	resultsChan := make(chan *HttpResult)

	defer func() {
		close(semaphoreChan)
		close(resultsChan)
	}()

	for i, url := range urls {
		go func(i int, url string) {
			semaphoreChan <- struct{}{}
			res, err := http.Get(url)
			result := &HttpResult{i, *res, err}
			resultsChan <- result
			<-semaphoreChan
		}(i, url)
	}

	var results []HttpResult
	for {
		result := <-resultsChan
		results = append(results, *result)
		if len(results) == len(urls) {
			break
		}
	}

	return results
}
