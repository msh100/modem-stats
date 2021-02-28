package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"time"
)

func main() {
	var body []byte
	var fetchTime int64
	if localFile := getenv("LOCAL_FILE", ""); localFile != "" {
		timeStart := time.Now().UnixNano() / int64(time.Millisecond)
		file, _ := os.Open(localFile)
		body, _ = ioutil.ReadAll(file)
		fetchTime = (time.Now().UnixNano() / int64(time.Millisecond)) - timeStart
	}

	var routerType string
	superhubType := getenv("SH_VERSION", "")
	if superhubType != "" {
		routerType = fmt.Sprintf("superhub%s", superhubType)
	} else {
		routerType = getenv("ROUTER_TYPE", "superhub3")
	}

	var router router

	switch routerType {
	case "superhub4":
		router = &superhub4{
			IPAddress: getenv("ROUTER_IP", "192.168.100.1"),
			stats:     body,
			fetchTime: fetchTime,
		}
	case "comhemc2":
		router = &comhemc2{
			IPAddress: getenv("ROUTER_IP", "192.168.10.1"),
			stats:     body,
			fetchTime: fetchTime,
			username:  getenv("ROUTER_USER", "admin"),
			password:  getenv("ROUTER_PASS", "admin"),
		}
	default:
		router = &superhub3{
			IPAddress: getenv("ROUTER_IP", "192.168.100.1"),
			stats:     body,
			fetchTime: fetchTime,
		}
	}

	routerStats, err := fetchStats(router)

	if err != nil {
		log.Printf("Error returned by parser: %v", err)
		os.Exit(1)
	} else {
		printForInflux(routerStats)
	}
}

func fetchStats(router router) (routerStats, error) {
	stats, err := router.ParseStats()
	return stats, err
}

func getenv(key, fallback string) string {
	value := os.Getenv(key)
	if len(value) == 0 {
		return fallback
	}
	return value
}
