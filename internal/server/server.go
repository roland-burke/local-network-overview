package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"time"

	"github.com/roland-burke/local-network-overview/internal/model"
	"github.com/roland-burke/rollogger"
)

var CurrentState model.AllHostsResponse
var Logger *rollogger.Log

const port = "8080"

func checkSingleAvailability(host model.NetClient) int {
	client := http.Client{
		Timeout: 4 * time.Second,
	}

	res, err := client.Get("http://" + host.AlternativeHost)
	if err != nil {
		Logger.Info("Failed: %s", err.Error())
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

func CheckAvailability() model.AllHostsResponse {
	loadedConfig, err := LoadConfig()

	if err != nil {
		Logger.Warn("Error during config load: " + err.Error())
		return model.AllHostsResponse{
			Status:    1,
			StatusMsg: err.Error(),
		}
	}

	var hosts = loadedConfig.Clients

	var hostStatusList []model.SingleHostStatus

	Logger.Info("Start availability check...")
	for i := 0; i < len(hosts); i++ {
		Logger.Info("Check %s (%s)", hosts[i].HostIp, hosts[i].AlternativeHost)
		available := checkSingleAvailability(hosts[i])

		var availValue = "UNKNOWN"

		if available == 0 {
			availValue = "UP"
		} else if available == 1 {
			availValue = "DOWN"
		} else if available == 2 {
			availValue = "PROBLEM"
		}

		var status = model.SingleHostStatus{
			Client: hosts[i],
			Status: availValue,
		}
		hostStatusList = append(hostStatusList, status)
	}

	return model.AllHostsResponse{
		Status:    0,
		StatusMsg: "okay",
		Data:      hostStatusList[:],
		TimeStamp: time.Now(),
	}
}

func returnCurrentState(w http.ResponseWriter, r *http.Request) {
	json, err := json.Marshal(CurrentState)
	if err != nil {
		fmt.Printf(err.Error())
		w.Write([]byte("Failed to marshal result: " + err.Error()))
	} else {
		w.Write(json)
	}
}

func StartServer() {
	http.Handle("/", http.FileServer(http.Dir("./static")))

	http.HandleFunc("/status", func(w http.ResponseWriter, r *http.Request) {
		returnCurrentState(w, r)
	})

	http.HandleFunc("/status/now", func(w http.ResponseWriter, r *http.Request) {
		CurrentState = CheckAvailability()
		returnCurrentState(w, r)
	})

	var err = http.ListenAndServe(":"+port, nil)
	Logger.Error(err.Error())
}

func LoadConfig() (model.ConfFile, error) {
	// Open our jsonFile
	jsonFile, err := os.Open("conf/config.json")
	// if we os.Open returns an error then handle it
	if err != nil {
		return model.ConfFile{}, errors.New("Cannot open file: " + err.Error())
	}

	// defer the closing of our jsonFile so that we can parse it later on
	defer jsonFile.Close()

	byteValue, err := ioutil.ReadAll(jsonFile)
	if err != nil {
		return model.ConfFile{}, errors.New("Cannot read file: " + err.Error())
	}
	var configFile model.ConfFile

	err = json.Unmarshal(byteValue, &configFile)

	if err != nil {
		return model.ConfFile{}, errors.New("Cannot unmarshal json: " + err.Error())
	}
	Logger.Info("Configured with %d targets and a retry intervall of %ds.", len(configFile.Clients), configFile.RetryIntervall)
	return configFile, nil
}

func ExecuteTimedRequest() {
	CurrentState = CheckAvailability()
}
