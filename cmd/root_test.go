package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func TestRootCommand(t *testing.T) {
	t.Run("root command exists", func(t *testing.T) {
		assert.NotNil(t, rootCmd)
		assert.Equal(t, "bitbucket-runner", rootCmd.Use)
		assert.Contains(t, rootCmd.Short, "CLI tool")
		assert.Contains(t, rootCmd.Long, "bitbucket-runner")
	})

	t.Run("help command works", func(t *testing.T) {
		cmd := &cobra.Command{}
		cmd.AddCommand(rootCmd)

		var output bytes.Buffer
		cmd.SetOut(&output)
		cmd.SetArgs([]string{"bitbucket-runner", "--help"})

		err := cmd.Execute()
		assert.NoError(t, err)

		outputStr := output.String()
		assert.Contains(t, outputStr, "bitbucket-runner")
		assert.Contains(t, outputStr, "Available Commands")
	})

	t.Run("subcommands are registered", func(t *testing.T) {
		commands := rootCmd.Commands()
		commandNames := make([]string, len(commands))
		for i, cmd := range commands {
			commandNames[i] = cmd.Name()
		}

		assert.Contains(t, commandNames, "run")
		assert.Contains(t, commandNames, "list")
	})
}

func TestRunCommand(t *testing.T) {
	t.Run("run command exists", func(t *testing.T) {
		runCommand := rootCmd.Commands()[0] // Assuming run is first
		if runCommand.Name() != "run" {
			// Find run command
			for _, cmd := range rootCmd.Commands() {
				if cmd.Name() == "run" {
					runCommand = cmd
					break
				}
			}
		}

		assert.Equal(t, "run", runCommand.Name())
		assert.Contains(t, runCommand.Short, "Run")
		assert.Contains(t, runCommand.Long, "pipeline")
	})

	t.Run("run command executes", func(t *testing.T) {
		// Create a temporary bitbucket-pipelines.yml file
		tmpDir := t.TempDir()
		tmpfile := filepath.Join(tmpDir, "bitbucket-pipelines.yml")
		
		// Write content to the temporary file
		content := []byte("pipelines:\n  default:\n    - step:\n        script:\n          - echo \"Hello, World!\"\n")
		err := os.WriteFile(tmpfile, content, 0644)
		if err != nil {
			t.Fatal(err)
		}

		// Temporarily move to the directory of the temp file
		oldWd, _ := os.Getwd()
		defer os.Chdir(oldWd)
		os.Chdir(tmpDir)

		// Create a fresh command instance to avoid state pollution
		testCmd := &cobra.Command{
			Use: "test",
		}
		testCmd.AddCommand(rootCmd)

		var output bytes.Buffer
		testCmd.SetOut(&output)
		testCmd.SetArgs([]string{"bitbucket-runner", "run"})

		err = testCmd.Execute()
		assert.NoError(t, err)

		outputStr := output.String()
		assert.Contains(t, outputStr, "Parsed pipeline config")
	})
}

func TestListCommand(t *testing.T) {
	t.Run("list command exists", func(t *testing.T) {
		listCommand := rootCmd.Commands()[1] // Assuming list is second
		if listCommand.Name() != "list" {
			// Find list command
			for _, cmd := range rootCmd.Commands() {
				if cmd.Name() == "list" {
					listCommand = cmd
					break
				}
			}
		}

		assert.Equal(t, "list", listCommand.Name())
		assert.Contains(t, listCommand.Short, "List")
		assert.Contains(t, listCommand.Long, "pipelines")
	})

	t.Run("list command executes", func(t *testing.T) {
		// Create a temporary bitbucket-pipelines.yml file
		tmpDir := t.TempDir()
		tmpfile := filepath.Join(tmpDir, "bitbucket-pipelines.yml")
		
		// Write content to the temporary file
		content := []byte("pipelines:\n  default:\n    - step:\n        script:\n          - echo \"Hello, World!\"\n")
		err := os.WriteFile(tmpfile, content, 0644)
		if err != nil {
			t.Fatal(err)
		}

		// Temporarily move to the directory of the temp file
		oldWd, _ := os.Getwd()
		defer os.Chdir(oldWd)
		os.Chdir(tmpDir)

		// Create a fresh command instance to avoid state pollution
		testCmd := &cobra.Command{
			Use: "test",
		}
		testCmd.AddCommand(rootCmd)

		var output bytes.Buffer
		testCmd.SetOut(&output)
		testCmd.SetArgs([]string{"bitbucket-runner", "list"})

		err = testCmd.Execute()
		assert.NoError(t, err)

		outputStr := output.String()
		assert.Contains(t, outputStr, "Available pipelines:")
		assert.Contains(t, outputStr, "- default")
	})
}

func TestCommandRegistration(t *testing.T) {
	t.Run("all expected commands are registered", func(t *testing.T) {
		expectedCommands := []string{"run", "list", "help", "completion"}
		actualCommands := make(map[string]bool)

		for _, cmd := range rootCmd.Commands() {
			actualCommands[cmd.Name()] = true
		}

		for _, expected := range expectedCommands {
			if expected == "help" || expected == "completion" {
				// These are built-in cobra commands
				continue
			}
			assert.True(t, actualCommands[expected], "Command %s should be registered", expected)
		}
	})
}