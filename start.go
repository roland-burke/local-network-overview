package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/roland-burke/rollogger"
)

const amountHosts = 3

var hosts [amountHosts]netClient
var currentState allHostsResponse
var logger *rollogger.Log

type singleHostStatus struct {
	Client netClient `json:"client"`
	Status string    `json:"status"`
}

type netClient struct {
	Name            string `json:"name"`
	Group           string `json:"group"`
	AlternativeHost string `json:"altHost"`
	HostIp          string `json:"host"`
}

type allHostsResponse struct {
	Data      []singleHostStatus `json:"data"`
	TimeStamp time.Time          `json:"timestamp"`
}

func checkSingleAvailability(host netClient) int {
	client := http.Client{
		Timeout: 4 * time.Second,
	}

	res, err := client.Get("http://" + host.AlternativeHost)
	if err != nil {
		logger.Error(err.Error())
		// Connection not established
		return 1
	}

	if res.StatusCode >= 200 && res.StatusCode < 300 {
		// All fine
		return 0
	}

	// Service reachabled, but no 2xx response
	return 2
}

func checkAvailability() {
	var hostStatusList [amountHosts]singleHostStatus

	logger.Info("Start availability check...")
	for i := 0; i < len(hosts); i++ {
		available := checkSingleAvailability(hosts[i])

		var availValue = "UNKNOWN"

		if available == 0 {
			availValue = "UP"
		} else if available == 1 {
			availValue = "DOWN"
		} else if available == 2 {
			availValue = "PROBLEM"
		}

		var status = singleHostStatus{
			Client: hosts[i],
			Status: availValue,
		}

		hostStatusList[i] = status
	}
	currentState = allHostsResponse{
		Data:      hostStatusList[:],
		TimeStamp: time.Now(),
	}
}

func fillHosts() {
	hosts[0] = netClient{
		Group:           "Fritz",
		Name:            "Fritz Box",
		HostIp:          "fritz.box",
		AlternativeHost: "192.168.178.1"}

	hosts[1] = netClient{
		Group:           "Home-Pi",
		Name:            "Homematic",
		HostIp:          "homematic.pi",
		AlternativeHost: "192.168.178.20:8080"}

	hosts[2] = netClient{
		Group:           "Home-Pi",
		Name:            "Grafana",
		HostIp:          "grafana.pi",
		AlternativeHost: "192.168.178.20:3000"}
}

func returnCurrentState(w http.ResponseWriter, r *http.Request) {
	json, err := json.Marshal(currentState)
	if err != nil {
		fmt.Printf(err.Error())
		w.Write([]byte("Failed to marshal result: " + err.Error()))
	} else {
		w.Write(json)
	}
}

func startServer() {
	http.Handle("/", http.FileServer(http.Dir("./static")))

	http.HandleFunc("/status", func(w http.ResponseWriter, r *http.Request) {
		returnCurrentState(w, r)
	})

	http.HandleFunc("/status/now", func(w http.ResponseWriter, r *http.Request) {
		checkAvailability()
		returnCurrentState(w, r)
	})

	var err = http.ListenAndServe(":8081", nil)
	logger.Error(err.Error())
}

func main() {
	logger = rollogger.Init(rollogger.INFO_LEVEL, true, true)
	uptimeTicker := time.NewTicker(5 * time.Minute)
	fillHosts()

	go func() {
		for {
			select {
			case <-uptimeTicker.C:
				go checkAvailability()
			}
		}
	}()

	startServer()
}
