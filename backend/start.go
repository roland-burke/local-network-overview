package main

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/roland-burke/rollogger"
)

var hosts [6]string
var currentState allHostsResponse
var logger *rollogger.Log

type singleHostStatus struct {
	HostIp string `json:"host"`
	Status string `json:"status"`
}

type allHostsResponse struct {
	Data      []singleHostStatus `json:"data"`
	TimeStamp time.Time          `json:"timestamp"`
}

func checkSingleAvailability(host string) bool {
	returnValue := false
	timeout := 3 * time.Second
	conn, err := net.DialTimeout("tcp", host, timeout)
	if err != nil {
		returnValue = false
		logger.Info("'%s' not reachable (timout: %fs): %s", host, timeout.Seconds(), err.Error())
	} else {
		returnValue = true
		conn.Close()
	}
	return returnValue
}

func checkAvailability() {
	var hostStatusList [6]singleHostStatus

	logger.Info("Start availability check...")
	for i := 0; i < len(hosts); i++ {
		available := checkSingleAvailability(hosts[i])

		var availValue = ""

		if available {
			availValue = "UP"
		} else {
			availValue = "DOWN"
		}

		var status = singleHostStatus{
			HostIp: hosts[i],
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
	hosts[0] = "192.168.178.38:8080"
	hosts[1] = "192.168.178.38:9100"
	hosts[2] = "192.168.178.38:9000"
	hosts[3] = "192.168.178.1:80"
	hosts[4] = "192.168.178.54:80"
	hosts[5] = "192.168.178.20:80"
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
