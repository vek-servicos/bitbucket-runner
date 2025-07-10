package parser

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPipelineParser_ParseYAML(t *testing.T) {
	parser := NewPipelineParser()

	t.Run("valid pipeline configuration", func(t *testing.T) {
		yamlData := `
image: ubuntu:20.04
pipelines:
  default:
    - step:
        name: Build and test
        script:
          - echo "Building..."
          - echo "Testing..."
`
		config, err := parser.ParseYAML([]byte(yamlData))
		require.NoError(t, err)
		require.NotNil(t, config)

		assert.Equal(t, "ubuntu:20.04", config.Image)
		assert.NotNil(t, config.Pipelines)
		assert.Greater(t, len(config.Pipelines.Default), 0)
		assert.Len(t, config.Pipelines.Default, 1)
		assert.Equal(t, "Build and test", config.Pipelines.Default[0].Step.Name)
		assert.Len(t, config.Pipelines.Default[0].Step.Script, 2)
	})

	t.Run("pipeline with multiple steps", func(t *testing.T) {
		yamlData := `
pipelines:
  default:
    - step:
        name: Build
        script:
          - echo "Building..."
    - step:
        name: Test
        script:
          - echo "Testing..."
        services:
          - postgres
`
		config, err := parser.ParseYAML([]byte(yamlData))
		require.NoError(t, err)
		require.NotNil(t, config)

		assert.Len(t, config.Pipelines.Default, 2)
		assert.Equal(t, "Build", config.Pipelines.Default[0].Step.Name)
		assert.Equal(t, "Test", config.Pipelines.Default[1].Step.Name)
		assert.Contains(t, config.Pipelines.Default[1].Step.Services, "postgres")
	})

	t.Run("pipeline with definitions", func(t *testing.T) {
		yamlData := `
pipelines:
  default:
    - step:
        script:
          - echo "test"
definitions:
  services:
    postgres:
      image: postgres:13
      environment:
        POSTGRES_DB: test
  caches:
    node:
      key: node-cache
      paths:
        - node_modules/
`
		config, err := parser.ParseYAML([]byte(yamlData))
		require.NoError(t, err)
		require.NotNil(t, config)

		assert.NotNil(t, config.Definitions)
		assert.Contains(t, config.Definitions.Services, "postgres")
		assert.Equal(t, "postgres:13", config.Definitions.Services["postgres"].Image)
		assert.Contains(t, config.Definitions.Caches, "node")
	})

	t.Run("invalid YAML", func(t *testing.T) {
		invalidYAML := `
pipelines:
  default:
    - step
      invalid yaml
`
		_, err := parser.ParseYAML([]byte(invalidYAML))
		assert.Error(t, err)
	})

	t.Run("empty pipelines", func(t *testing.T) {
		yamlData := `
image: ubuntu:20.04
pipelines: {}
`
		_, err := parser.ParseYAML([]byte(yamlData))
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no pipelines defined")
	})

	t.Run("step without script", func(t *testing.T) {
		yamlData := `
pipelines:
  default:
    - step:
        name: Empty step
`
		_, err := parser.ParseYAML([]byte(yamlData))
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no script defined")
	})
}

func TestPipelineParser_ParseFile(t *testing.T) {
	parser := NewPipelineParser()

	t.Run("parse existing file", func(t *testing.T) {
		// Create temporary file
		tmpDir := t.TempDir()
		filePath := filepath.Join(tmpDir, "bitbucket-pipelines.yml")

		yamlContent := `
image: ubuntu:20.04
pipelines:
  default:
    - step:
        name: "Test Step"
        script:
          - echo "Hello World"
`
		err := os.WriteFile(filePath, []byte(yamlContent), 0644)
		require.NoError(t, err)

		config, err := parser.ParseFile(filePath)
		require.NoError(t, err)
		require.NotNil(t, config)

		assert.Equal(t, "ubuntu:20.04", config.Image)
		assert.NotNil(t, config.Pipelines)
		assert.Greater(t, len(config.Pipelines.Default), 0)
	})

	t.Run("file not found", func(t *testing.T) {
		_, err := parser.ParseFile("nonexistent-file.yml")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to read file")
	})
}

func TestPipelineParser_ParseDefault(t *testing.T) {
	parser := NewPipelineParser()

	t.Run("parse default file in current directory", func(t *testing.T) {
		// Save current directory
		originalDir, err := os.Getwd()
		require.NoError(t, err)
		defer os.Chdir(originalDir)

		// Create temporary directory and change to it
		tmpDir := t.TempDir()
		err = os.Chdir(tmpDir)
		require.NoError(t, err)

		yamlContent := `image: ubuntu:20.04
pipelines:
  default:
    - step:
        name: test
        image: node:14
        script:
          - npm test`

		err = os.WriteFile("bitbucket-pipelines.yml", []byte(yamlContent), 0644)
		require.NoError(t, err)

		parser := NewPipelineParser()
		config, err := parser.ParseDefault()
		require.NoError(t, err)
		require.NotNil(t, config)

		assert.Len(t, config.Pipelines.Default, 1)
		assert.Equal(t, "test", config.Pipelines.Default[0].Step.Name)
		assert.Equal(t, "node:14", config.Pipelines.Default[0].Step.Image)
		assert.Len(t, config.Pipelines.Default[0].Step.Script, 1)
	})

	t.Run("default file not found", func(t *testing.T) {
		// Save current directory
		originalDir, err := os.Getwd()
		require.NoError(t, err)
		defer os.Chdir(originalDir)

		// Create temporary directory and change to it
		tmpDir := t.TempDir()
		err = os.Chdir(tmpDir)
		require.NoError(t, err)

		_, err = parser.ParseDefault()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "bitbucket-pipelines.yml not found")
	})
}