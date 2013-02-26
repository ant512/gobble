package main

import (
	"encoding/json"
	"io/ioutil"
)

type Config struct {
	Port int64
	PostPath string
	Theme string
}

func LoadConfig(filename string) (*Config, error) {
	file, err := ioutil.ReadFile(filename)

	if err != nil {
		return nil, err
	}

	config := new(Config)

	// Set up defaults
	config.Port = 8080
	config.PostPath = "./posts"
	config.Theme = "simianzombie"

	err = json.Unmarshal(file, config)

	return config, err
}