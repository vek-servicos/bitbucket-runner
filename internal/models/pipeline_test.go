package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPipelineConfig_Validate(t *testing.T) {
	t.Run("valid configuration", func(t *testing.T) {
		config := &PipelineConfig{
			Pipelines: &Pipelines{
				Default: []StepWrapper{
					{
						Step: Step{
							Name:   "Build",
							Script: []string{"echo 'building'"},
						},
					},
				},
			},
		}

		err := config.Validate()
		assert.NoError(t, err)
	})

	t.Run("no pipelines defined", func(t *testing.T) {
		config := &PipelineConfig{
			Pipelines: &Pipelines{},
		}

		err := config.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no pipelines defined")
	})

	t.Run("pipeline with no steps", func(t *testing.T) {
		config := &PipelineConfig{
			Pipelines: &Pipelines{
				Default: []StepWrapper{},
			},
		}

		err := config.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "has no steps defined")
	})

	t.Run("step with no script", func(t *testing.T) {
		config := &PipelineConfig{
			Pipelines: &Pipelines{
				Default: []StepWrapper{
					{
						Step: Step{
							Name:   "Empty step",
							Script: []string{},
						},
					},
				},
			},
		}

		err := config.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "has no script defined")
	})
}

func TestPipelineConfig_GetDefaultPipeline(t *testing.T) {
	t.Run("default pipeline exists", func(t *testing.T) {
		config := &PipelineConfig{
			Pipelines: &Pipelines{
				Default: []StepWrapper{
					{
						Step: Step{
							Name:   "Default step",
							Script: []string{"echo 'default'"},
						},
					},
				},
			},
		}

		pipeline, exists := config.GetDefaultPipeline()
		assert.True(t, exists)
		assert.NotNil(t, pipeline)
		assert.Len(t, *pipeline, 1)
		assert.Equal(t, "Default step", (*pipeline)[0].Step.Name)
	})

	t.Run("default pipeline does not exist", func(t *testing.T) {
		config := &PipelineConfig{
			Pipelines: &Pipelines{
				Custom: map[string]Pipeline{
					"custom": []StepWrapper{
						{
							Step: Step{
								Script: []string{"echo 'custom'"},
							},
						},
					},
				},
			},
		}

		_, exists := config.GetDefaultPipeline()
		assert.False(t, exists)
	})
}

func TestPipelineConfig_GetBranchPipeline(t *testing.T) {
	t.Run("branch pipeline exists", func(t *testing.T) {
		config := &PipelineConfig{
			Pipelines: &Pipelines{
				Branches: map[string]Pipeline{
					"main": []StepWrapper{
						{
							Step: Step{
								Script: []string{"echo 'branch'"},
							},
						},
					},
				},
			},
		}

		pipeline, exists := config.GetBranchPipeline("main")
		assert.True(t, exists)
		assert.NotNil(t, pipeline)
	})

	t.Run("branch pipeline does not exist", func(t *testing.T) {
		config := &PipelineConfig{
			Pipelines: &Pipelines{
				Default: []StepWrapper{
					{
						Step: Step{
							Script: []string{"echo 'default'"},
						},
					},
				},
			},
		}

		_, exists := config.GetBranchPipeline("main")
		assert.False(t, exists)
	})
}

func TestStep_Validation(t *testing.T) {
	t.Run("step with all fields", func(t *testing.T) {
		step := Step{
			Name:   "Complete step",
			Image:  "ubuntu:20.04",
			Script: []string{"echo 'test'", "ls -la"},
			Services: []string{"postgres", "redis"},
			Caches: []string{"node", "pip"},
			Environment: map[string]string{
				"NODE_ENV": "test",
				"DEBUG":   "true",
			},
		}

		assert.Equal(t, "Complete step", step.Name)
		assert.Equal(t, "ubuntu:20.04", step.Image)
		assert.Len(t, step.Script, 2)
		assert.Contains(t, step.Services, "postgres")
		assert.Contains(t, step.Caches, "node")
		assert.Equal(t, "test", step.Environment["NODE_ENV"])
	})

	t.Run("minimal step", func(t *testing.T) {
		step := Step{
			Script: []string{"echo 'minimal'"},
		}

		assert.Empty(t, step.Name)
		assert.Empty(t, step.Image)
		assert.Len(t, step.Script, 1)
		assert.Empty(t, step.Services)
	})
}