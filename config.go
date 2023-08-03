package main

import (
	"fmt"
	"os"

	"github.com/mitchellh/go-homedir"
	"gopkg.in/yaml.v3"
)

type config struct {
	Renderer map[string]renderConfig `yaml:"renderers"`
}

func NewConfig(path string) (config, error) {
	c := defaults()
	path, err := homedir.Expand(path)
	if err != nil {
		return c, err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return c, fmt.Errorf("could not read file %s: %s", path, err.Error())
	}

	err = yaml.Unmarshal(data, &c)
	if err != nil {
		return c, fmt.Errorf("could not read data from config file %s: %s", path, err.Error())
	}

	return c, nil
}

func defaults() config {
	return config{}
}
