package cmd

import (
	"errors"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// currentReleaseVersion is used to print the version the user currently has downloaded
const currentReleaseVersion = "v0.1.0"

// rootCmd is the base for all commands.
var rootCmd = &cobra.Command{
	Use:   "chippy [command]",
	Short: "chippy is Chip-8 emulator",
	Long:  "chippy is Chip-8 emulator",
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			return errors.New("Requires at least 1 argument")
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Unknown command. Try `chippy help` for more information")
	},
}

func init() {
	rootCmd.AddCommand(runCmd)
	rootCmd.AddCommand(versionCmd)
}

// Execute runs chippy according to the user's command/subcommand/flags
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
