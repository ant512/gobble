package main

import (
	"encoding/json"
	"io/ioutil"
)

type Config struct {
	Name string
	Description string
	Address string
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
	config.Name = "Gobble"
	config.Description = "Blogging Engine"
	config.Address = "http://simianzombie.com"
	config.Port = 8080
	config.PostPath = "./posts"
	config.Theme = "grump"

	err = json.Unmarshal(file, config)

	return config, err
}