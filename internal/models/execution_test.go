package models

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewExecutionContext(t *testing.T) {
	config := &PipelineConfig{
		Pipelines: &Pipelines{
			Default: []StepWrapper{
				{Step: Step{Script: []string{"echo 'test'"}}},
			},
		},
	}

	ec := NewExecutionContext(config, "/tmp/test")

	assert.NotNil(t, ec)
	assert.Equal(t, config, ec.PipelineConfig)
	assert.Equal(t, 0, ec.CurrentStep)
	assert.Equal(t, "/tmp/test", ec.WorkingDir)
	assert.Equal(t, ExecutionStatusPending, ec.Status)
	assert.Empty(t, ec.StepResults)
	assert.NotNil(t, ec.Environment)
}

func TestExecutionContext_StartExecution(t *testing.T) {
	ec := NewExecutionContext(nil, "")
	initialTime := ec.StartTime

	// Wait a small amount to ensure time difference
	time.Sleep(1 * time.Millisecond)

	ec.StartExecution()

	assert.Equal(t, ExecutionStatusRunning, ec.Status)
	assert.True(t, ec.StartTime.After(initialTime))
}

func TestExecutionContext_CompleteExecution(t *testing.T) {
	ec := NewExecutionContext(nil, "")
	ec.StartExecution()

	assert.Nil(t, ec.EndTime)

	ec.CompleteExecution()

	assert.Equal(t, ExecutionStatusCompleted, ec.Status)
	assert.NotNil(t, ec.EndTime)
	assert.True(t, ec.EndTime.After(ec.StartTime))
}

func TestExecutionContext_FailExecution(t *testing.T) {
	ec := NewExecutionContext(nil, "")
	ec.StartExecution()

	errorMsg := "test error"
	ec.FailExecution(errorMsg)

	assert.Equal(t, ExecutionStatusFailed, ec.Status)
	assert.Equal(t, errorMsg, ec.ErrorMessage)
	assert.NotNil(t, ec.EndTime)
}

func TestExecutionContext_AddStepResult(t *testing.T) {
	ec := NewExecutionContext(nil, "")

	result := StepResult{
		StepIndex: 0,
		StepName:  "Test step",
		Status:    StepStatusCompleted,
		ExitCode:  0,
	}

	ec.AddStepResult(result)

	assert.Len(t, ec.StepResults, 1)
	assert.Equal(t, result, ec.StepResults[0])
}

func TestExecutionContext_GetCurrentStep(t *testing.T) {
	t.Run("valid current step", func(t *testing.T) {
		config := &PipelineConfig{
			Pipelines: &Pipelines{
				Default: []StepWrapper{
					{Step: Step{Name: "Step 1", Script: []string{"echo '1'"}}},
					{Step: Step{Name: "Step 2", Script: []string{"echo '2'"}}},
				},
			},
		}

		ec := NewExecutionContext(config, "")
		step := ec.GetCurrentStep()

		assert.NotNil(t, step)
		assert.Equal(t, "Step 1", step.Name)
	})

	t.Run("no pipeline config", func(t *testing.T) {
		ec := NewExecutionContext(nil, "")
		step := ec.GetCurrentStep()

		assert.Nil(t, step)
	})

	t.Run("step index out of bounds", func(t *testing.T) {
		config := &PipelineConfig{
			Pipelines: &Pipelines{
				Default: []StepWrapper{
					{Step: Step{Script: []string{"echo 'test'"}}},
				},
			},
		}

		ec := NewExecutionContext(config, "")
		ec.CurrentStep = 1 // Out of bounds

		step := ec.GetCurrentStep()
		assert.Nil(t, step)
	})
}

func TestExecutionContext_NextStep(t *testing.T) {
	ec := NewExecutionContext(nil, "")
	initialStep := ec.CurrentStep

	ec.NextStep()

	assert.Equal(t, initialStep+1, ec.CurrentStep)
}

func TestExecutionContext_IsCompleted(t *testing.T) {
	t.Run("not completed", func(t *testing.T) {
		config := &PipelineConfig{
			Pipelines: &Pipelines{
				Default: []StepWrapper{
					{Step: Step{Script: []string{"echo '1'"}}},
					{Step: Step{Script: []string{"echo '2'"}}},
				},
			},
		}

		ec := NewExecutionContext(config, "")
		assert.False(t, ec.IsCompleted())
	})

	t.Run("completed", func(t *testing.T) {
		config := &PipelineConfig{
			Pipelines: &Pipelines{
				Default: []StepWrapper{
					{Step: Step{Script: []string{"echo 'test'"}}},
				},
			},
		}

		ec := NewExecutionContext(config, "")
		ec.CurrentStep = 1 // Beyond last step

		assert.True(t, ec.IsCompleted())
	})

	t.Run("no config", func(t *testing.T) {
		ec := NewExecutionContext(nil, "")
		assert.True(t, ec.IsCompleted())
	})
}

func TestExecutionContext_GetTotalDuration(t *testing.T) {
	t.Run("execution not ended", func(t *testing.T) {
		ec := NewExecutionContext(nil, "")
		ec.StartTime = time.Now().Add(-1 * time.Hour)

		duration := ec.GetTotalDuration()
		assert.True(t, duration > 59*time.Minute) // Should be close to 1 hour
	})

	t.Run("execution ended", func(t *testing.T) {
		ec := NewExecutionContext(nil, "")
		start := time.Now().Add(-1 * time.Hour)
		end := start.Add(30 * time.Minute)
		ec.StartTime = start
		ec.EndTime = &end

		duration := ec.GetTotalDuration()
		assert.Equal(t, 30*time.Minute, duration)
	})
}

func TestExecutionContext_EnvironmentVariables(t *testing.T) {
	ec := NewExecutionContext(nil, "")

	// Test setting environment variable
	ec.SetEnvironmentVariable("TEST_VAR", "test_value")

	// Test getting environment variable
	value, exists := ec.GetEnvironmentVariable("TEST_VAR")
	assert.True(t, exists)
	assert.Equal(t, "test_value", value)

	// Test getting non-existent variable
	_, exists = ec.GetEnvironmentVariable("NON_EXISTENT")
	assert.False(t, exists)
}

func TestStepResult(t *testing.T) {
	result := StepResult{
		StepIndex:   0,
		StepName:    "Test Step",
		Status:      StepStatusCompleted,
		StartTime:   time.Now(),
		ExitCode:    0,
		Output:      "Success output",
		ErrorOutput: "",
		Duration:    30 * time.Second,
	}

	assert.Equal(t, 0, result.StepIndex)
	assert.Equal(t, "Test Step", result.StepName)
	assert.Equal(t, StepStatusCompleted, result.Status)
	assert.Equal(t, 0, result.ExitCode)
	assert.Equal(t, "Success output", result.Output)
	assert.Equal(t, 30*time.Second, result.Duration)
}