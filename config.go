package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
)

type Config struct {
	Name                string
	CommentsOpenForDays int
	Description         string
	Address             string
	MediaPath           string
	Port                int64
	PostPath            string
	CommentPath         string
	Theme               string
	ThemePath           string
	AkismetAPIKey       string
	RecaptchaPublicKey  string
	RecaptchaPrivateKey string
	StaticFilePath      string
	StaticFiles         map[string]string
}

func LoadConfig(filename string) (*Config, error) {
	file, err := ioutil.ReadFile(filename)

	if err != nil {
		msg := fmt.Sprintf("Could not read config file %v: %v", filename, err)
		return nil, errors.New(msg)
	}

	config := new(Config)
	config.setDefaults()

	err = json.Unmarshal(file, config)

	if err != nil {
		msg := fmt.Sprintf("Could not parse config file %v: %v", filename, err)
		return nil, errors.New(msg)
	}

	err = config.validateConfig()

	if err != nil {
		return nil, err
	}

	return config, nil
}

func (c *Config) FullThemePath() string {
	return c.ThemePath + string(filepath.Separator) + c.Theme
}

func (c *Config) validateConfig() error {
	_, err := os.Stat(c.FullThemePath())

	if err != nil {
		msg := fmt.Sprintf("Could not load theme %v", c.FullThemePath)
		return errors.New(msg)
	}

	_, err = os.Stat(c.PostPath)

	if err != nil {
		msg := fmt.Sprintf("Could not load posts from %v", c.PostPath)
		return errors.New(msg)
	}

	_, err = os.Stat(c.MediaPath)

	if err != nil {
		msg := fmt.Sprintf("Could not access media from %v", c.MediaPath)
		return errors.New(msg)
	}

	_, err = os.Stat(c.StaticFilePath)

	if err != nil {
		msg := fmt.Sprintf("Could not access files from %v", c.StaticFilePath)
		return errors.New(msg)
	}

	return nil
}

func (c *Config) setDefaults() {
	c.Name = "Gobble"
	c.CommentsOpenForDays = 0
	c.Description = "Blogging Engine"
	c.Port = 8080
	c.PostPath = "./posts"
	c.CommentPath = "./comments"
	c.MediaPath = "./media"
	c.ThemePath = "./themes"
	c.StaticFilePath = "./files"
	c.Theme = "grump"
}
