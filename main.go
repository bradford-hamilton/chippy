package main

import (
	"fmt"

	"github.com/bradford-hamilton/chippy/internal/chip8"
)

func main() {
	// make this a CLI
	vm := chip8.NewVM()
	fmt.Println(vm)

	// Setup the graphics (window size, display mode, etc)
	// setupGraphics()

	// setup input system (bind callbacks)
	// setupInput()

	// Initialize registers and memory once (clear the memory, registers and screen)
	// chip8.initialize()

	// Copy the program into the memory
	// chip8.loadGame()

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

// Because the system does not draw every cycle, we should set a draw flag when we need to update our screen.
// Only two opcodes should set this flag:

// 0x00E0 – Clears the screen
// 0xDXYN – Draws a sprite on the screen
