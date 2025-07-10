package models

import (
	"errors"
	"fmt"
)

// PipelineConfig represents the parsed bitbucket-pipelines.yml structure
type PipelineConfig struct {
	Image    string                 `yaml:"image,omitempty"`
	Clone    *CloneConfig          `yaml:"clone,omitempty"`
	Pipelines *Pipelines           `yaml:"pipelines"`
	Definitions *Definitions       `yaml:"definitions,omitempty"`
	Options  *Options              `yaml:"options,omitempty"`
}

// Pipelines represents the pipelines section
type Pipelines struct {
	Default      Pipeline                    `yaml:"default,omitempty"`
	Branches     map[string]Pipeline         `yaml:"branches,omitempty"`
	PullRequests map[string]Pipeline         `yaml:"pull-requests,omitempty"`
	Custom       map[string]Pipeline         `yaml:"custom,omitempty"`
	Tags         map[string]Pipeline         `yaml:"tags,omitempty"`
}

// Pipeline represents a single pipeline configuration
type Pipeline []StepWrapper

// StepWrapper wraps a Step to handle the YAML structure
type StepWrapper struct {
	Step Step `yaml:"step"`
}

// Step represents a single step in a pipeline
type Step struct {
	Name         string            `yaml:"name,omitempty"`
	Image        string            `yaml:"image,omitempty"`
	Script       []string          `yaml:"script"`
	Services     []string          `yaml:"services,omitempty"`
	Artifacts    *Artifacts        `yaml:"artifacts,omitempty"`
	Caches       []string          `yaml:"caches,omitempty"`
	AfterScript  []string          `yaml:"after-script,omitempty"`
	Condition    *Condition        `yaml:"condition,omitempty"`
	Environment  map[string]string `yaml:"environment,omitempty"`
}

// CloneConfig represents clone configuration
type CloneConfig struct {
	Enabled bool   `yaml:"enabled,omitempty"`
	Depth   int    `yaml:"depth,omitempty"`
	Lfs     bool   `yaml:"lfs,omitempty"`
}

// Definitions represents pipeline definitions
type Definitions struct {
	Services map[string]Service `yaml:"services,omitempty"`
	Caches   map[string]Cache   `yaml:"caches,omitempty"`
}

// Service represents a service definition
type Service struct {
	Image       string            `yaml:"image"`
	Environment map[string]string `yaml:"environment,omitempty"`
	Ports       []string          `yaml:"ports,omitempty"`
}

// Cache represents a cache definition
type Cache struct {
	Key   string   `yaml:"key,omitempty"`
	Paths []string `yaml:"paths,omitempty"`
	Path  string   `yaml:",omitempty"` // For simple string caches
}

// UnmarshalYAML implements custom unmarshaling for Cache
func (c *Cache) UnmarshalYAML(unmarshal func(interface{}) error) error {
	// Try to unmarshal as a string first
	var str string
	if err := unmarshal(&str); err == nil {
		c.Path = str
		return nil
	}
	
	// If that fails, try to unmarshal as a struct
	type cacheAlias Cache
	var cache cacheAlias
	if err := unmarshal(&cache); err != nil {
		return err
	}
	
	c.Key = cache.Key
	c.Paths = cache.Paths
	c.Path = cache.Path
	return nil
}

// Artifacts represents artifacts configuration
type Artifacts struct {
	Paths []string `yaml:"paths"`
}

// Condition represents step execution condition
type Condition struct {
	Changesets *Changesets `yaml:"changesets,omitempty"`
}

// Changesets represents changeset conditions
type Changesets struct {
	IncludePaths []string `yaml:"includePaths,omitempty"`
}

// Options represents pipeline options
type Options struct {
	Docker bool `yaml:"docker,omitempty"`
	Size   string `yaml:"size,omitempty"`
}

// Validate validates the pipeline configuration
func (pc *PipelineConfig) Validate() error {
	if pc.Pipelines == nil {
		return errors.New("no pipelines defined")
	}

	// Validate default pipeline
	if len(pc.Pipelines.Default) > 0 {
		if err := pc.validatePipeline("default", pc.Pipelines.Default); err != nil {
			return err
		}
	} else if pc.Pipelines.Default != nil {
		// Check if default pipeline exists but is empty
		return errors.New("pipeline 'default' has no steps defined")
	}

	// Validate branch pipelines
	for name, pipeline := range pc.Pipelines.Branches {
		if err := pc.validatePipeline(fmt.Sprintf("branches.%s", name), pipeline); err != nil {
			return err
		}
	}

	// Validate pull request pipelines
	for name, pipeline := range pc.Pipelines.PullRequests {
		if err := pc.validatePipeline(fmt.Sprintf("pull-requests.%s", name), pipeline); err != nil {
			return err
		}
	}

	// Validate custom pipelines
	for name, pipeline := range pc.Pipelines.Custom {
		if err := pc.validatePipeline(fmt.Sprintf("custom.%s", name), pipeline); err != nil {
			return err
		}
	}

	// Validate tag pipelines
	for name, pipeline := range pc.Pipelines.Tags {
		if err := pc.validatePipeline(fmt.Sprintf("tags.%s", name), pipeline); err != nil {
			return err
		}
	}

	// Check if at least one pipeline is defined
	if len(pc.Pipelines.Default) == 0 && len(pc.Pipelines.Branches) == 0 && 
	   len(pc.Pipelines.PullRequests) == 0 && len(pc.Pipelines.Custom) == 0 && 
	   len(pc.Pipelines.Tags) == 0 {
		return errors.New("no pipelines defined")
	}

	return nil
}

func (pc *PipelineConfig) validatePipeline(name string, pipeline Pipeline) error {
	if len(pipeline) == 0 {
		return fmt.Errorf("pipeline '%s' has no steps defined", name)
	}

	for i, stepWrapper := range pipeline {
		if len(stepWrapper.Step.Script) == 0 {
			return fmt.Errorf("step %d in pipeline '%s' has no script defined", i+1, name)
		}
	}

	return nil
}

// GetDefaultPipeline returns the default pipeline if it exists
func (pc *PipelineConfig) GetDefaultPipeline() (*Pipeline, bool) {
	if pc.Pipelines == nil || len(pc.Pipelines.Default) == 0 {
		return nil, false
	}
	return &pc.Pipelines.Default, true
}

// GetBranchPipeline returns the pipeline for a specific branch
func (pc *PipelineConfig) GetBranchPipeline(branch string) (*Pipeline, bool) {
	if pc.Pipelines == nil || pc.Pipelines.Branches == nil {
		return nil, false
	}
	pipeline, exists := pc.Pipelines.Branches[branch]
	if !exists {
		return nil, false
	}
	return &pipeline, true
}