package main

import (
	"fmt"
	"os"
	"time"

	"github.com/bradford-hamilton/chippy/internal/chip8"
	"github.com/bradford-hamilton/chippy/internal/pixel"
	"github.com/faiface/pixel/pixelgl"
)

const refreshRate = 60

func main() {
	// pixelgl needs access to the main thread so this pattern is suggested
	// will revisit once things are working
	pixelgl.Run(runMain)
}

func runMain() {
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

	win, err := pixel.NewWindow()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// maybe handle beeps here

	ticker := time.NewTicker(time.Second / refreshRate)
	defer ticker.Stop()

	for range ticker.C {
		if !win.Closed() {
			// fetch, decode, and execute opcode and update timers
			vm.EmulateCycle()

			if vm.DrawFlag() {
				win.DrawGraphics(vm.GetGraphics())
			} else {
				win.UpdateInput()
			}

			win.HandleKeyInput() // If we press or release a key, we should store this state in the part that emulates the keypad

			continue
		}

		fmt.Println("exit signal detected, gracefully shutting down...")
		break
	}
}
