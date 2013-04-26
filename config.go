package main

import (
	"encoding/json"
	"io/ioutil"
)

type Config struct {
	Name                string
	Description         string
	Address             string
	MediaPath           string
	Port                int64
	PostPath            string
	Theme               string
	AkismetAPIKey       string
	RecaptchaPublicKey  string
	RecaptchaPrivateKey string
}

func LoadConfig(filename string) (*Config, error) {
	file, err := ioutil.ReadFile(filename)

	if err != nil {
		return nil, err
	}

	config := new(Config)
	config.setDefaults()

	err = json.Unmarshal(file, config)

	return config, err
}

func (c *Config) setDefaults() {
	c.Name = "Gobble"
	c.Description = "Blogging Engine"
	c.Port = 8080
	c.PostPath = "./posts"
	c.MediaPath = "./media"
	c.Theme = "grump"
}
