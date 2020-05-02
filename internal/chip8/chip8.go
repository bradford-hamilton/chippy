package chip8

import (
	"io/ioutil"

	"github.com/bradford-hamilton/chippy/internal/display"
)

// system memory map
// 0x000-0x1FF - Chip 8 interpreter (contains font set in emu, more on that below)
// 0x050-0x0A0 - Used for the built in 4x5 pixel font set (0-F)
// 0x200-0xFFF - Program ROM and work RAM

// Chip-8 used to be implemented on 4k systems like the Telmac 1800 and Cosmac VIP where the chip-8 interpreter
// itself occpied the first 512 bytes of memory (up to 0x200). In modern CHIP-8 implementations (like ours here), where
// the interpreter is running natively outside the 4K memory space, there is no need to avoid the lower 512 bytes of
// memory (0x000-0x200), and it is common to store font data there.

// VM represents the chip-8 virtual machine
type VM struct {
	opcode     uint16        // current opcode
	memory     [4096]byte    // the VM's memory -> see more on this up top
	v          [16]byte      // 8-bit general purpose register, (V0 - VE*)
	i          uint16        // index register (0x000 to 0xFFF)
	pc         uint16        // program counter (0x000 to 0xFFF)
	stack      [16]uint16    // stack for instructions
	sp         uint16        // stack pointer
	gfx        [64 * 32]byte // represents the screen pixels
	delayTimer byte          // 8-bit delay timer which counts down at 60 hertz, until it reaches 0
	soundTimer byte          // 8-bit sound timer which counts down at 60 hertz, until it reaches 0
	clockSpeed uint16        // "cpu" clock speed
	timerSpeed uint16        // timer speed
	key        [16]byte      // HEX based: 0x0-0xF
	drawFlag   bool          // system doesn't draw every cycle, set a draw flag when we need to update our screen.
}

// NewVM handles initializing a VM, loading the font set into memory, and loading the ROM into memory
func NewVM(pathToROM string) (*VM, error) {
	vm := VM{
		memory: [4096]byte{},
		v:      [16]byte{},
		pc:     0x200,
		stack:  [16]uint16{},
		gfx:    [64 * 32]byte{}, // or [32][8]byte{}? Decide later when I understand the display better how I want to set this up
		key:    [16]byte{},
	}
	vm.loadFontSet()
	if err := vm.loadROM(pathToROM); err != nil {
		return nil, err
	}
	return &vm, nil
}

// load the dislpay's font set into the first 80 bytes of the vm's memory
func (vm *VM) loadFontSet() {
	for i := 0; i < 80; i++ {
		vm.memory[i] = display.FontSet[i]
	}
}

func (vm *VM) loadROM(path string) error {
	rom, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}
	if len(rom) >= 3585 {
		panic("error: rom too large. Max size: 3584")
	}
	for i := 0; i < len(rom); i++ {
		vm.memory[0x50+i] = rom[i]
	}
	return nil
}
