package model

import "time"

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
	Status    int                `json:"status"`
	StatusMsg string             `json:"statusMsg"`
	Data      []singleHostStatus `json:"data"`
	TimeStamp time.Time          `json:"timestamp"`
}
