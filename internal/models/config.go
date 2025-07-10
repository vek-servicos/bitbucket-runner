package models

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// RunnerConfig represents tool configuration and step type mappings
type RunnerConfig struct {
	Version     string                 `yaml:"version"`
	StepTypes   map[string]StepType   `yaml:"stepTypes"`
	Environment map[string]string     `yaml:"environment"`
	Defaults    DefaultConfig         `yaml:"defaults"`
	Logging     LoggingConfig         `yaml:"logging"`
	Docker      DockerConfig          `yaml:"docker"`
}

// StepType represents configuration for a specific step type
type StepType struct {
	Image       string            `yaml:"image"`
	Environment map[string]string `yaml:"environment"`
	Volumes     []VolumeMount     `yaml:"volumes"`
	Ports       []PortMapping     `yaml:"ports"`
	Timeout     int               `yaml:"timeout"` // in seconds
}

// VolumeMount represents a volume mount configuration
type VolumeMount struct {
	Host      string `yaml:"host"`
	Container string `yaml:"container"`
	ReadOnly  bool   `yaml:"readOnly"`
}

// PortMapping represents a port mapping configuration
type PortMapping struct {
	Host      int    `yaml:"host"`
	Container int    `yaml:"container"`
	Protocol  string `yaml:"protocol"`
}

// DefaultConfig represents default configuration values
type DefaultConfig struct {
	Image       string `yaml:"image"`
	WorkingDir  string `yaml:"workingDir"`
	Timeout     int    `yaml:"timeout"`
	Shell       string `yaml:"shell"`
}

// LoggingConfig represents logging configuration
type LoggingConfig struct {
	Level      string `yaml:"level"`
	Format     string `yaml:"format"`
	OutputFile string `yaml:"outputFile"`
}

// DockerConfig represents Docker-specific configuration
type DockerConfig struct {
	Host       string `yaml:"host"`
	APIVersion string `yaml:"apiVersion"`
	Registry   string `yaml:"registry"`
	PullPolicy string `yaml:"pullPolicy"`
}

// NewDefaultRunnerConfig creates a new RunnerConfig with default values
func NewDefaultRunnerConfig() *RunnerConfig {
	return &RunnerConfig{
		Version: "1.0",
		StepTypes: map[string]StepType{
			"default": {
				Image:   "ubuntu:20.04",
				Timeout: 3600, // 1 hour
			},
		},
		Environment: make(map[string]string),
		Defaults: DefaultConfig{
			Image:      "ubuntu:20.04",
			WorkingDir: "/opt/atlassian/pipelines/agent/build",
			Timeout:    3600,
			Shell:      "/bin/bash",
		},
		Logging: LoggingConfig{
			Level:  "info",
			Format: "text",
		},
		Docker: DockerConfig{
			Host:       "unix:///var/run/docker.sock",
			APIVersion: "1.41",
			PullPolicy: "missing",
		},
	}
}

// LoadFromFile loads runner configuration from a YAML file
func LoadRunnerConfigFromFile(filename string) (*RunnerConfig, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file %s: %w", filename, err)
	}

	var config RunnerConfig
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal config YAML: %w", err)
	}

	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid runner configuration: %w", err)
	}

	return &config, nil
}

// LoadFromDefaultLocations attempts to load configuration from default locations
func LoadRunnerConfigFromDefaultLocations() (*RunnerConfig, error) {
	defaultPaths := []string{
		"bitbucket-runner.yml",
		"bitbucket-runner.yaml",
		".bitbucket-runner.yml",
		".bitbucket-runner.yaml",
		filepath.Join(os.Getenv("HOME"), ".bitbucket-runner.yml"),
		filepath.Join(os.Getenv("HOME"), ".config", "bitbucket-runner", "config.yml"),
	}

	for _, path := range defaultPaths {
		if _, err := os.Stat(path); err == nil {
			return LoadRunnerConfigFromFile(path)
		}
	}

	// Return default configuration if no file found
	return NewDefaultRunnerConfig(), nil
}

// Validate validates the runner configuration
func (rc *RunnerConfig) Validate() error {
	if rc.Version == "" {
		return errors.New("version is required")
	}

	if rc.Defaults.Image == "" {
		return errors.New("default image is required")
	}

	if rc.Defaults.Timeout <= 0 {
		return errors.New("default timeout must be positive")
	}

	for name, stepType := range rc.StepTypes {
		if stepType.Image == "" {
			return fmt.Errorf("step type '%s' must have an image", name)
		}
		if stepType.Timeout <= 0 {
			return fmt.Errorf("step type '%s' timeout must be positive", name)
		}
	}

	return nil
}

// GetStepType returns the configuration for a specific step type
func (rc *RunnerConfig) GetStepType(stepType string) (*StepType, bool) {
	st, exists := rc.StepTypes[stepType]
	return &st, exists
}

// GetDefaultStepType returns the default step type configuration
func (rc *RunnerConfig) GetDefaultStepType() *StepType {
	if st, exists := rc.StepTypes["default"]; exists {
		return &st
	}
	return &StepType{
		Image:   rc.Defaults.Image,
		Timeout: rc.Defaults.Timeout,
	}
}

// SaveToFile saves the runner configuration to a YAML file
func (rc *RunnerConfig) SaveToFile(filename string) error {
	data, err := yaml.Marshal(rc)
	if err != nil {
		return fmt.Errorf("failed to marshal config to YAML: %w", err)
	}

	err = os.WriteFile(filename, data, 0644)
	if err != nil {
		return fmt.Errorf("failed to write config file %s: %w", filename, err)
	}

	return nil
}