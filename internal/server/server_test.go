package server

import (
	"testing"

	"github.com/roland-burke/local-network-overview/internal/model"
	"github.com/roland-burke/rollogger"
	"github.com/stretchr/testify/assert"
)

func init() {
	// Mute the logger
	Logger = rollogger.Init(rollogger.ERROR_LEVEL, true, true)
}

func TestLoadConfig(t *testing.T) {
	assert := assert.New(t)
	// test
	conf, err := LoadConfig()

	assert.Equal(nil, err)
	assert.Equal(
		model.ConfFile{
			Clients: []model.NetClient{{
				Name: "MyService", Group: "Test1", AlternativeHost: "192.168.178.1", HostIp: "my.test"}, {
				Name: "AnotherService", Group: "Test2", AlternativeHost: "192.168.178.1:8080", HostIp: "my2.test"},
			}, RetryIntervall: 300}, conf)

}
