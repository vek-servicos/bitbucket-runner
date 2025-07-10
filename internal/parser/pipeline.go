package parser

import (
	"fmt"
	"io/ioutil"
	"os"

	"bitbucket-runner/internal/models"
	"gopkg.in/yaml.v3"
)

// PipelineParser handles parsing of bitbucket-pipelines.yml files
type PipelineParser struct{}

// NewPipelineParser creates a new instance of PipelineParser
func NewPipelineParser() *PipelineParser {
	return &PipelineParser{}
}

// ParseFile parses a bitbucket-pipelines.yml file and returns a PipelineConfig
func (p *PipelineParser) ParseFile(filename string) (*models.PipelineConfig, error) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s: %w", filename, err)
	}

	return p.ParseYAML(data)
}

// ParseYAML parses YAML data and returns a PipelineConfig
func (p *PipelineParser) ParseYAML(data []byte) (*models.PipelineConfig, error) {
	var config models.PipelineConfig
	err := yaml.Unmarshal(data, &config)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal YAML: %w", err)
	}

	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid pipeline configuration: %w", err)
	}

	return &config, nil
}

// ParseDefault parses the default bitbucket-pipelines.yml file in current directory
func (p *PipelineParser) ParseDefault() (*models.PipelineConfig, error) {
	filename := "bitbucket-pipelines.yml"
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		return nil, fmt.Errorf("bitbucket-pipelines.yml not found in current directory")
	}

	return p.ParseFile(filename)
}