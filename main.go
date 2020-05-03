package main

import (
	"fmt"
	"os"

	"github.com/bradford-hamilton/chippy/internal/chip8"
	"github.com/faiface/pixel/pixelgl"
)

func main() {
	// pixelgl needs access to the main thread so this is the suggested pattern
	pixelgl.Run(runChippy)
}

func runChippy() {
	if len(os.Args) != 2 {
		fmt.Println("incorrect usage. Usage: `chippy path/to/rom`")
		os.Exit(1)
	}
	pathToROM := os.Args[1]

	vm, err := chip8.NewVM(pathToROM)
	if err != nil {
		fmt.Printf("\nerror creating a new chip-8 VM: %v\n", err)
		os.Exit(1)
	}

	go vm.ManageAudio()
	go vm.Run()

	<-vm.Shutdown
}
