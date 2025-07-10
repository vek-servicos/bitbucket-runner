package models

import (
	"time"
)

// ExecutionContext tracks execution state and runtime information
type ExecutionContext struct {
	PipelineConfig *PipelineConfig
	CurrentStep    int
	StepResults    []StepResult
	Environment    map[string]string
	WorkingDir     string
	StartTime      time.Time
	EndTime        *time.Time
	Status         ExecutionStatus
	ErrorMessage   string
}

// StepResult represents the result of a single step execution
type StepResult struct {
	StepIndex    int
	StepName     string
	Status       StepStatus
	StartTime    time.Time
	EndTime      *time.Time
	ExitCode     int
	Output       string
	ErrorOutput  string
	Duration     time.Duration
}

// ExecutionStatus represents the overall execution status
type ExecutionStatus string

const (
	ExecutionStatusPending    ExecutionStatus = "pending"
	ExecutionStatusRunning    ExecutionStatus = "running"
	ExecutionStatusCompleted  ExecutionStatus = "completed"
	ExecutionStatusFailed     ExecutionStatus = "failed"
	ExecutionStatusCancelled  ExecutionStatus = "cancelled"
)

// StepStatus represents the status of a single step
type StepStatus string

const (
	StepStatusPending   StepStatus = "pending"
	StepStatusRunning   StepStatus = "running"
	StepStatusCompleted StepStatus = "completed"
	StepStatusFailed    StepStatus = "failed"
	StepStatusSkipped   StepStatus = "skipped"
)

// NewExecutionContext creates a new execution context
func NewExecutionContext(config *PipelineConfig, workingDir string) *ExecutionContext {
	return &ExecutionContext{
		PipelineConfig: config,
		CurrentStep:    0,
		StepResults:    make([]StepResult, 0),
		Environment:    make(map[string]string),
		WorkingDir:     workingDir,
		StartTime:      time.Now(),
		Status:         ExecutionStatusPending,
	}
}

// StartExecution marks the execution as started
func (ec *ExecutionContext) StartExecution() {
	ec.Status = ExecutionStatusRunning
	ec.StartTime = time.Now()
}

// CompleteExecution marks the execution as completed
func (ec *ExecutionContext) CompleteExecution() {
	now := time.Now()
	ec.EndTime = &now
	ec.Status = ExecutionStatusCompleted
}

// FailExecution marks the execution as failed
func (ec *ExecutionContext) FailExecution(errorMsg string) {
	now := time.Now()
	ec.EndTime = &now
	ec.Status = ExecutionStatusFailed
	ec.ErrorMessage = errorMsg
}

// AddStepResult adds a step result to the execution context
func (ec *ExecutionContext) AddStepResult(result StepResult) {
	ec.StepResults = append(ec.StepResults, result)
}

// GetCurrentStep returns the current step being executed
func (ec *ExecutionContext) GetCurrentStep() *Step {
	if ec.PipelineConfig == nil {
		return nil
	}

	// Get default pipeline for now
	pipeline, exists := ec.PipelineConfig.GetDefaultPipeline()
	if !exists || ec.CurrentStep >= len(*pipeline) {
		return nil
	}

	return &(*pipeline)[ec.CurrentStep].Step
}

// NextStep advances to the next step
func (ec *ExecutionContext) NextStep() {
	ec.CurrentStep++
}

// IsCompleted returns true if all steps have been executed
func (ec *ExecutionContext) IsCompleted() bool {
	if ec.PipelineConfig == nil {
		return true
	}

	pipeline, exists := ec.PipelineConfig.GetDefaultPipeline()
	if !exists {
		return true
	}

	return ec.CurrentStep >= len(*pipeline)
}

// GetTotalDuration returns the total execution duration
func (ec *ExecutionContext) GetTotalDuration() time.Duration {
	if ec.EndTime == nil {
		return time.Since(ec.StartTime)
	}
	return ec.EndTime.Sub(ec.StartTime)
}

// SetEnvironmentVariable sets an environment variable
func (ec *ExecutionContext) SetEnvironmentVariable(key, value string) {
	ec.Environment[key] = value
}

// GetEnvironmentVariable gets an environment variable
func (ec *ExecutionContext) GetEnvironmentVariable(key string) (string, bool) {
	value, exists := ec.Environment[key]
	return value, exists
}