package cmd

import (
	"bitbucket-runner/internal/parser"
	"fmt"

	"github.com/spf13/cobra"
)

// runCmd represents the run command
var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Run a Bitbucket pipeline",
	Long: `Run a Bitbucket pipeline by parsing the bitbucket-pipelines.yml file
and executing the defined steps in sequence.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		config, err := parser.ParsePipelineConfig("bitbucket-pipelines.yml")
		if err != nil {
			return fmt.Errorf("Error parsing pipeline config: %w", err)
		}
		// Using cmd.OutOrStdout() to respect output redirection in tests.
		fmt.Fprintf(cmd.OutOrStdout(), "Parsed pipeline config: %+v\n", config)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(runCmd)

	// Here you will define your flags and configuration settings.
	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// runCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// runCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}