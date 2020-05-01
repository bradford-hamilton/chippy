package chip8

import "github.com/bradford-hamilton/chippy/internal/display"

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
	opcode     uint16      // current opcode
	memory     [4096]byte  // the VM's memory -> see more on this up top
	v          [16]byte    // 8-bit general purpose register, (V0 - VE*)
	i          uint16      // index register (0x000 to 0xFFF)
	pc         uint16      // program counter (0x000 to 0xFFF)
	stack      [16]uint16  // stack for instructions
	sp         uint16      // stack pointer
	gfx        [32][8]byte // represents the screen pixels
	delayTimer byte        // 8-bit delay timer which counts down at 60 hertz, until it reaches 0
	soundTimer byte        // 8-bit sound timer which counts down at 60 hertz, until it reaches 0
	clockSpeed uint16      // "cpu" clock speed
	timerSpeed uint16      // timer speed
	keypad     [16]byte    // HEX based: 0x0-0xF
}

// NewVM returns a pointer to a new VM with the dislpay's font set initialized in the first 80 bytes
func NewVM() *VM {
	var memory [4096]byte
	for i := 0; i < 80; i++ {
		memory[i] = display.FontSet[i]
	}

	return &VM{
		opcode:     0,
		memory:     memory,
		v:          [16]byte{},
		i:          0,
		pc:         0x200, // From the spec -> more detail on why at the top
		stack:      [16]uint16{},
		sp:         0,
		gfx:        [32][8]byte{},
		delayTimer: 0,
		soundTimer: 0,
		clockSpeed: 0,
		timerSpeed: 0,
		keypad:     [16]byte{},
	}
}
