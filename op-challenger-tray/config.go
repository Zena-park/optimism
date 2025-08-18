package main

import (
	"os"

	"gopkg.in/yaml.v3"
)

// saveConfigToFile saves the configuration to a YAML file
func saveConfigToFile(config *ChallengerConfig, filepath string) error {
	data, err := yaml.Marshal(config)
	if err != nil {
		return err
	}
	
	return os.WriteFile(filepath, data, 0644)
}

// loadConfigFromFile loads the configuration from a YAML file
func loadConfigFromFile(filepath string, config *ChallengerConfig) error {
	data, err := os.ReadFile(filepath)
	if err != nil {
		if os.IsNotExist(err) {
			// 파일이 없으면 기본 설정 사용
			return nil
		}
		return err
	}
	
	return yaml.Unmarshal(data, config)
}