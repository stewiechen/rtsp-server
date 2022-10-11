package conf

import (
	"encoding/json"
	"io/ioutil"
)

type Config struct {
	Port        int `json:"port"`
	FrameBuffer int `json:"frame_buffer"`
}

func NewConfig(filename string) *Config {
	bys, err := ioutil.ReadFile(filename)

	var config Config
	err = json.Unmarshal(bys, &config)
	if err != nil {
		return &Config{
			Port:        8554,
			FrameBuffer: 60,
		}
	}

	if config.Port == 0 {
		config.Port = 8554
	}

	if config.FrameBuffer == 0 {
		config.FrameBuffer = 60
	}

	return &config
}
