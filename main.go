package main

import (
	"fmt"
	"os"

	"github.com/bradford-hamilton/chippy/internal/chip8"
)

func main() {
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

	fmt.Println(vm)

	// Setup the graphics (window size, display mode, etc)
	// display.setupGraphics() ?

	// setup input system (bind callbacks)
	// keypad.setupInput() ?

	// emulation loop sudo code:
	//	   for {
	//          chip8.emulateCycle() // Fetch Opcode, Decode Opcode, Execute Opcode, Update timers
	//
	//          if drawFlag {
	//                drawGraphics()
	//          }
	//
	//          chip8.setKeys() // If we press or release a key, we should store this state in the part that emulates the keypad
	//	   }
}
