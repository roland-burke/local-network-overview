package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"time"

	"github.com/roland-burke/rollogger"
)

var currentState allHostsResponse
var logger *rollogger.Log

const port = "8080"

func checkSingleAvailability(host netClient) int {
	client := http.Client{
		Timeout: 4 * time.Second,
	}

	res, err := client.Get("http://" + host.AlternativeHost)
	if err != nil {
		logger.Info("Failed: %s", err.Error())
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

func checkAvailability() allHostsResponse {
	loadedConfig, err := loadConfig()

	if err != nil {
		logger.Warn("Error during config load: " + err.Error())
		return allHostsResponse{
			Status:    1,
			StatusMsg: err.Error(),
		}
	}

	var hosts = loadedConfig.Clients

	var hostStatusList []singleHostStatus

	logger.Info("Start availability check...")
	for i := 0; i < len(hosts); i++ {
		logger.Info("Check %s (%s)", hosts[i].HostIp, hosts[i].AlternativeHost)
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
		hostStatusList = append(hostStatusList, status)
	}

	return allHostsResponse{
		Status:    0,
		StatusMsg: "okay",
		Data:      hostStatusList[:],
		TimeStamp: time.Now(),
	}
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
		currentState = checkAvailability()
		returnCurrentState(w, r)
	})

	var err = http.ListenAndServe(":"+port, nil)
	logger.Error(err.Error())
}

func loadConfig() (confFile, error) {
	// Open our jsonFile
	jsonFile, err := os.Open("conf/config.json")
	// if we os.Open returns an error then handle it
	if err != nil {
		return confFile{}, errors.New("Cannot open file: " + err.Error())
	}

	// defer the closing of our jsonFile so that we can parse it later on
	defer jsonFile.Close()

	byteValue, err := ioutil.ReadAll(jsonFile)
	if err != nil {
		return confFile{}, errors.New("Cannot read file: " + err.Error())
	}
	var configFile confFile

	err = json.Unmarshal(byteValue, &configFile)

	if err != nil {
		return confFile{}, errors.New("Cannot unmarshal json: " + err.Error())
	}
	logger.Info("Configured with %d targets and a retry intervall of %ds.", len(configFile.Clients), configFile.RetryIntervall)
	return configFile, nil
}

func executeTimedRequest() {
	currentState = checkAvailability()
}

func main() {
	logger = rollogger.Init(rollogger.INFO_LEVEL, true, true)
	currentState = allHostsResponse{
		Status:    2,
		StatusMsg: "Check was not performed yet.",
	}

	var loadedConfig, err = loadConfig()
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}
	uptimeTicker := time.NewTicker(time.Duration(loadedConfig.RetryIntervall) * time.Second)

	go func() {
		for {
			select {
			case <-uptimeTicker.C:
				go executeTimedRequest()
			}
		}
	}()

	startServer()
}
