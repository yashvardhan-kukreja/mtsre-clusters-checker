package cmd

import (
	"fmt"
	"github.com/mt-sre/mtsre-clusters-checker/pkg/cmd/scan"
	"github.com/spf13/cobra"
	"os"
)

func init() {
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

	rootCmd.AddCommand(scan.NewCommand())
}

var rootCmd = &cobra.Command{
	Use:   "clusters-checker",
	Short: "Stay up-to-date with the stale OSD clusters your organization might be which might be racking up some bills for you!",
	Long: `clusters-checker helps you perform an Openshift-wide check to scan if any of your organization's clusters are running since more than 24hr and potentially worth deleting, thereby saving you the costs.
It currently supports firing messages at a slack channel associated with your workspace with the results of the scan.`,

	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Hello, World! Welcome to OSD Clusters checker.")
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
