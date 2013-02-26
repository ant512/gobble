package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
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

	err = json.Unmarshal(file, config)

	log.Println(err)

	return config, err
}