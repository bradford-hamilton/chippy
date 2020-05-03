package main

import (
	"github.com/bradford-hamilton/chippy/cmd"
	"github.com/faiface/pixel/pixelgl"
)

func main() {
	// pixelgl needs access to the main thread
	pixelgl.Run(runChippy)
}

func runChippy() {
	cmd.Execute()
}
