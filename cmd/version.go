package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// versionCmd returns the callers installed chippy version
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Retrieve the currently installed chippy version",
	Long:  "Run `chippy version` to get your current chippy version",
	Args:  cobra.NoArgs,
	Run:   runVersion,
}

func runVersion(cmd *cobra.Command, args []string) {
	fmt.Println(currentReleaseVersion)
}
