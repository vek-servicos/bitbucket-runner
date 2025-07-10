package cmd

import (
	"bitbucket-runner/internal/parser"
	"fmt"

	"github.com/spf13/cobra"
)

// listCmd represents the list command
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List available pipelines and steps",
	Long: `List all available pipelines and their steps from the
bitbucket-pipelines.yml file in the current directory.`,
	Run: func(cmd *cobra.Command, args []string) {
		config, err := parser.ParsePipelineConfig("bitbucket-pipelines.yml")
		if err != nil {
			fmt.Fprintf(cmd.OutOrStdout(), "Error parsing pipeline config: %v\n", err)
			return
		}

		fmt.Fprintf(cmd.OutOrStdout(), "Available pipelines:\n")
		if config.Pipelines.Default != nil {
			fmt.Fprintf(cmd.OutOrStdout(), "- default\n")
		}
		for name := range config.Pipelines.Branches {
			fmt.Fprintf(cmd.OutOrStdout(), "- %s\n", name)
		}
		for name := range config.Pipelines.PullRequests {
			fmt.Fprintf(cmd.OutOrStdout(), "- %s\n", name)
		}
		for name := range config.Pipelines.Custom {
			fmt.Fprintf(cmd.OutOrStdout(), "- %s\n", name)
		}
		for name := range config.Pipelines.Tags {
			fmt.Fprintf(cmd.OutOrStdout(), "- %s\n", name)
		}
	},

}

func init() {
	rootCmd.AddCommand(listCmd)

	// Here you will define your flags and configuration settings.
	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// listCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// listCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}