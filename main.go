package main

import (
	"github.com/bradford-hamilton/chippy/cmd"
	"github.com/faiface/pixel/pixelgl"
)

// pixelgl needs access to the main thread
func main() {
	pixelgl.Run(runChippy)
}

func runChippy() {
	cmd.Execute()
}
