// kafka_config.go
package main

import (
	"gopkg.in/yaml.v3"
	"io/ioutil"
)

type KafkaConfig struct {
	Brokers  []string `yaml:"brokers"`
	Producer struct {
		Topic string `yaml:"topic"`
	} `yaml:"producer"`
	Consumer struct {
		Topic   string `yaml:"topic"`
		GroupID string `yaml:"group_id"`
	} `yaml:"consumer"`
}

func loadKafkaConfig(path string) (*KafkaConfig, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var config KafkaConfig
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return nil, err
	}
	return &config, nil
}
