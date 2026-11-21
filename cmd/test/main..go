package main

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/docker/docker/api/types/network"
	"gopkg.in/yaml.v3"
)

type test struct {
	NetworkConfig *network.NetworkingConfig
}

var j = `{
  "NetworkConfig": {
    "EndpointsConfig": {
    }
  }
}
`

var y = `
NetworkConfig:
  EndpointsConfig:
    internal: {}
`

func main() {
	reader := bytes.NewReader([]byte(j))
	decoder := json.NewDecoder(reader)
	d := &test{}
	if err := decoder.Decode(d); err != nil {
		panic(err)
	}
	fmt.Printf("%v", d)

	d = &test{}
	yamlReader := bytes.NewReader([]byte(y))
	yamlDecoder := yaml.NewDecoder(yamlReader)
	if err := yamlDecoder.Decode(d); err != nil {
		panic(err)
	}
	fmt.Printf("%v", d)
}
