package cmd

import (
	"fmt"
	"os"

	"github.com/bradford-hamilton/chippy/internal/chip8"
	"github.com/spf13/cobra"
)

// runCmd runs the chippy virtual machine and waits for a shutdown signal to exit
var runCmd = &cobra.Command{
	Use:   "run `path/to/rom`",
	Short: "run the chippy emulator",
	Args:  cobra.MinimumNArgs(1),
	Run:   runChippy,
}

func runChippy(cmd *cobra.Command, args []string) {
	if len(args) != 1 {
		fmt.Println("The run command takes one argument: a `path/to/rom`")
		os.Exit(1)
	}
	pathToROM := os.Args[2]

	vm, err := chip8.NewVM(pathToROM, refreshRate)
	if err != nil {
		fmt.Printf("\nerror creating a new chip-8 VM: %v\n", err)
		os.Exit(1)
	}

	go vm.ManageAudio()
	go vm.Run()

	<-vm.ShutdownC
}
