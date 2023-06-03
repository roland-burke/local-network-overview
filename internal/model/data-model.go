package model

import "time"

type SingleHostStatus struct {
	Client NetClient `json:"client"`
	Status string    `json:"status"`
}

type ConfFile struct {
	Clients        []NetClient `json:"targets"`
	RetryIntervall int         `json:"retryIntervall"`
}

type NetClient struct {
	Name            string `json:"name"`
	Group           string `json:"group"`
	AlternativeHost string `json:"altHost"`
	HostIp          string `json:"host"`
}

type AllHostsResponse struct {
	Status    int                `json:"status"`
	StatusMsg string             `json:"statusMsg"`
	Data      []SingleHostStatus `json:"data"`
	TimeStamp time.Time          `json:"timestamp"`
}
