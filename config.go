package main

import (
	"bytes"
	"fmt"

	"github.com/go-git/go-billy/v5"
	"github.com/mitchellh/go-homedir"
	"gopkg.in/yaml.v3"
)

type config struct {
	Renderer map[string]renderConfig `yaml:"renderers"`
}

func NewConfig(path string, filesys billy.Filesystem) (config, error) {
	c := defaults()
	path, err := homedir.Expand(path)
	if err != nil {
		return c, err
	}

	file, err := filesys.Open(path)
	if err != nil {
		err = fmt.Errorf("Could not open file '%s', error occured: %w", path, err)
		return c, err
	}
	defer file.Close()
	data := new(bytes.Buffer)
	_, err = data.ReadFrom(file)
	if err != nil {
		return c, fmt.Errorf("could not read file %s: %s", path, err.Error())
	}

	err = yaml.Unmarshal(data.Bytes(), &c)
	if err != nil {
		return c, fmt.Errorf("could not read data from config file %s: %s", path, err.Error())
	}

	return c, nil
}

func defaults() config {
	return config{}
}
