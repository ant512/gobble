package main

import (
	"encoding/json"
	"io/ioutil"
)

type Config struct {
	Name                string
	CommentsOpenForDays int
	Description         string
	Address             string
	MediaPath           string
	Port                int64
	PostPath            string
	Theme               string
	ThemePath           string
	AkismetAPIKey       string
	RecaptchaPublicKey  string
	RecaptchaPrivateKey string
	HashEmailAddresses  bool
}

var SharedConfig *Config = nil

func LoadConfig(filename string) error {
	file, err := ioutil.ReadFile(filename)

	if err != nil {
		return err
	}

	SharedConfig = new(Config)
	SharedConfig.setDefaults()

	err = json.Unmarshal(file, SharedConfig)

	return err
}

func (c *Config) setDefaults() {
	c.Name = "Gobble"
	c.CommentsOpenForDays = 0
	c.Description = "Blogging Engine"
	c.Port = 8080
	c.PostPath = "./posts"
	c.MediaPath = "./media"
	c.ThemePath = "./themes"
	c.Theme = "grump"
	c.HashEmailAddresses = false
}
