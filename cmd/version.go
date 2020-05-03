package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// versionCmd returns the callers installed chippy version
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Retrieve the currently installed chippy version",
	Long:  "Simply run `chippy version` to get your current chippy version",
	Args:  cobra.MinimumNArgs(0),
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) != 0 {
			fmt.Println("The version command does not take any arguments")
			os.Exit(1)
		}
		fmt.Println(currentReleaseVersion)
	},
}
