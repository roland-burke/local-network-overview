package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"time"

	"github.com/roland-burke/rollogger"
)

var currentState allHostsResponse
var logger *rollogger.Log
var configFile confFile

type singleHostStatus struct {
	Client netClient `json:"client"`
	Status string    `json:"status"`
}

type confFile struct {
	Clients        []netClient `json:"targets"`
	RetryIntervall int         `json:"retryIntervall"`
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

func checkAvailability(hosts []netClient) {
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
	currentState = allHostsResponse{
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
		checkAvailability(configFile.Clients)
		returnCurrentState(w, r)
	})

	var err = http.ListenAndServe(":8081", nil)
	logger.Error(err.Error())
}

func initConfig() confFile {
	// Open our jsonFile
	jsonFile, err := os.Open("conf/config.json")
	// if we os.Open returns an error then handle it
	if err != nil {
		logger.Error("Cannot open file: %s", err.Error())
		os.Exit(1)
	}

	// defer the closing of our jsonFile so that we can parse it later on
	defer jsonFile.Close()

	byteValue, err := ioutil.ReadAll(jsonFile)
	if err != nil {
		logger.Error("Cannot read file: %s", err.Error())
		os.Exit(2)
	}
	var configFile confFile

	err = json.Unmarshal(byteValue, &configFile)

	if err != nil {
		logger.Error("Cannot unmarshal file: %s", err.Error())
		os.Exit(3)
	}
	logger.Info("Configured with %d targets and a retry intervall of %ds.", len(configFile.Clients), configFile.RetryIntervall)
	return configFile
}

func main() {
	logger = rollogger.Init(rollogger.INFO_LEVEL, true, true)

	configFile = initConfig()
	uptimeTicker := time.NewTicker(time.Duration(configFile.RetryIntervall) * time.Second)

	go func() {
		for {
			select {
			case <-uptimeTicker.C:
				go checkAvailability(configFile.Clients)
			}
		}
	}()

	startServer()
}
