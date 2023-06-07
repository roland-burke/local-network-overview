# Network Overview ![Build Status](https://github.com/roland-burke/network-overview/actions/workflows/default-build-and-publish.yml/badge.svg) [![Coverage Status](https://coveralls.io/repos/github/roland-burke/network-overview/badge.svg?branch=master)](https://coveralls.io/github/roland-burke/network-overview?branch=master) [![Go Report Card](https://goreportcard.com/badge/github.com/roland-burke/network-overview)](https://goreportcard.com/report/github.com/roland-burke/network-overview)

## About
This tool gives a quick overview, of what services are on your network. It can provide a url and an alternative link (IP address and port).

## Configuration
A config.json file have to be mounted inside the container with `-v your/path/config.json:/conf/config.json`. In the json file you can configure the service, that you want to have listed. The refresh intervall (seconds) can also be configured. A configuration change will be considered during runtime.
