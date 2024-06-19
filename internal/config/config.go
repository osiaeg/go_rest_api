package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v2"
)

type Config struct {
	Server struct {
		Host string `yaml:"host"`
		Port string `yaml:"port"`
	} `yaml:"server"`
	Database struct {
		Url string `yaml:"url"`
	} `yaml:"db"`
}

func Parse(env string) *Config {
	var cfg Config
	var path string

	if env == "local" {
		path = "../../configs/local.yml"
		// path = "configs/local.yml"

	} else if env == "docker" {
		path = "configs/docker.yml"
	}

	f, err := os.Open(path)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(2)
	}
	defer f.Close()

	decoder := yaml.NewDecoder(f)
	err = decoder.Decode(&cfg)
	if err != nil {
		fmt.Println(err)
		os.Exit(2)
	}

	return &cfg
}
