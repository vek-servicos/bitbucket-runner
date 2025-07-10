package parser

import (
	"io/ioutil"

	"bitbucket-runner/internal/models"
	"gopkg.in/yaml.v3"
)

// ParsePipelineConfig reads a bitbucket-pipelines.yml file and unmarshals it into a PipelineConfig struct.
func ParsePipelineConfig(filePath string) (*models.PipelineConfig, error) {
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	var config models.PipelineConfig
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return nil, err
	}

	return &config, nil
}