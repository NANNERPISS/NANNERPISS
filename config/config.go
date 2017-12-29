package config

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
)

type Config struct {
	DB struct {
		Driver string
		Source string
	}
	TG struct {
		Token string
	}
	WL struct {
		ControlGroup int64
		DataDir      string
		LogChannel   int64
	}
}

func Load(c interface{}) (*Config, error) {
	var configBytes []byte
	var err error

	switch v := c.(type) {
	case io.Reader:
		configBytes, err = ioutil.ReadAll(v)
		if err != nil {
			return nil, err
		}
	case string:
		configBytes, err = ioutil.ReadFile(v)
	default:
		return nil, fmt.Errorf("config: Cannot load config from type '%T'", v)
	}

	config := &Config{}
	err = json.Unmarshal(configBytes, config)
	if err != nil {
		return nil, err
	}

	return config, nil
}
