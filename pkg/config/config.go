package config

import (
	"gopkg.in/yaml.v2"
	"io/ioutil"
)

type Configuration struct {
	Providers map[string]string `yaml:"providers"`
}

func LoadFromFile(filePath string) (*Configuration, error) {
	bytes, err := ioutil.ReadFile(filePath)
	if nil != err {
		return nil, err
	}
	var cfg Configuration
	err = yaml.Unmarshal(bytes, &cfg)

	return &cfg, err
}
