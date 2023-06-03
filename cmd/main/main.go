package main

import (
	"os"
	"time"

	"github.com/roland-burke/local-network-overview/internal/model"
	"github.com/roland-burke/local-network-overview/internal/server"
	"github.com/roland-burke/rollogger"
)

func main() {
	server.Logger = rollogger.Init(rollogger.INFO_LEVEL, true, true)
	server.CurrentState = model.AllHostsResponse{
		Status:    2,
		StatusMsg: "Check was not performed yet.",
	}

	var loadedConfig, err = server.LoadConfig()
	if err != nil {
		server.Logger.Error(err.Error())
		os.Exit(1)
	}
	uptimeTicker := time.NewTicker(time.Duration(loadedConfig.RetryIntervall) * time.Second)

	go func() {
		for {
			select {
			case <-uptimeTicker.C:
				go server.ExecuteTimedRequest()
			}
		}
	}()

	server.StartServer()
}
