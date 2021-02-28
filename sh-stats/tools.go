package main

import (
	"io/ioutil"
	"net/http"
	"time"
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
