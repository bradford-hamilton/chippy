package main

import (
	"fmt"
	"os"

	"github.com/bradford-hamilton/chippy/internal/chip8"
	"github.com/bradford-hamilton/chippy/internal/pixel"
	"github.com/faiface/pixel/pixelgl"
)

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

	w, err := pixel.NewWindow()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	vm, err := chip8.NewVM(pathToROM, w)
	if err != nil {
		fmt.Printf("\nerror creating a new chip-8 VM: %v\n", err)
		os.Exit(1)
	}

	go vm.ManageAudio()

	// Emulate a clock speed of 60MHz
	for {
		select {
		case <-vm.Clock.C:
			if !vm.Window.Closed() {
				vm.EmulateCycle()

				if vm.DrawFlag() {
					vm.Window.DrawGraphics(vm.GetGraphics())
				} else {
					vm.Window.UpdateInput()
				}

				vm.HandleKeyInput()

				if vm.DelayTimer > 0 {
					vm.DelayTimer--
				}
				if vm.SoundTimer > 0 {
					if vm.SoundTimer == 1 {
						// Ensure we don't block if the go routine is not ready
						select {
						case vm.BeepChan <- struct{}{}:
						default:
						}
					}
					vm.SoundTimer--
				}
				continue
			}

			fmt.Println("exit signal detected, gracefully shutting down...")
			goto end
		}
	}
end:
}
