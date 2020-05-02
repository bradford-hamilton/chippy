package chip8

import (
	"fmt"
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

// EmulateCycle will handle fetch, decode, and execute for a chip-8 VM
func (vm *VM) EmulateCycle() {
	// One opcode is 2 bytes long, ex. 0xA2FO we will need to fetch two successive bytes (ex. 0xA2 and 0xF0) and merge them to
	// get the actual opcode. First we shift current instruction (ex. 10100010) left 8 which would look like 1010001000000000.
	// Then OR it with the upcoming byte which gives us a 16 bit chunk containing the combined bytes

	// TODO: comeback and think about endianness
	vm.opcode = uint16(vm.memory[vm.pc])<<8 | uint16(vm.memory[vm.pc+1])
	vm.drawFlag = false
	vm.parseOpcode()
}

func (vm *VM) parseOpcode() {
	switch vm.opcode & 0xF000 {
	// 0NNN -> Execute machine language subroutine at address NNN
	case 0x0000:
		switch vm.opcode & 0x00FF {
		case 0x00E0:
			// 00E0 -> Clear the screen
		case 0x00EE:
			// 00EE -> Return from a subroutine.
		default:
			// vm.pc += 2
			// default: I don't think it's an error here but double check
		}
	case 0x1000:
		// 1NNN -> Jump to address NNN
	case 0x2000:
		// 2NNN -> Execute subroutine starting at address NNN
	case 0x3000:
		// 3XNN -> Skip the following instruction if the value of register VX == NN
	case 0x4000:
		// 4XNN -> Skip the following instruction if the value of register VX != NN
	case 0x5000:
		// 5XY0 -> Skip the following instruction if the value of register VX == VY
	case 0x6000:
		// 6XNN -> Store number NN in register VX
	case 0x7000:
		// 7XNN -> Add the value NN to register VX
	case 0x8000:
		switch vm.opcode & 0x000F {
		case 0x0000:
			// 8XYO -> Store the value of register VY in register VX
		case 0x0001:
			// 8XY1 -> Set VX to VX OR VY
		case 0x0002:
			// 8XY2 -> Set VX to VX AND VY
		case 0x0003:
			// 8XY3 -> Set VX to VX XOR VY
		case 0x0004:
			// 8XY4 -> Add the value of register VY to register VX
			// Set VF to 01 if a carry occurs
			// Set VF to 00 if a carry does not occur
		case 0x0005:
			// 8XY5 -> Subtract the value of register VY from register VX
			// Set VF to 00 if a borrow occurs
			// Set VF to 01 if a borrow does not occur
		case 0x0006:
			// 8XY6 -> Store the value of register VY shifted right one bit in register VX
			// Set register VF to the least significant bit prior to the shift
		case 0x0007:
			// 8XY7 -> Set register VX to the value of VY minus VX
			// Set VF to 00 if a borrow occurs
			// Set VF to 01 if a borrow does not occur
		case 0x000E:
			// 8XYE -> Store the value of register VY shifted left one bit in register VX
			// Set register VF to the most significant bit prior to the shift
		default:
			fmt.Printf("unknown opcode: %x\n", vm.opcode&0x000F)
		}
	case 0x9000:
		// 9XY0 -> Skip the following instruction if the value of VX != value of VY
	case 0xA000:
		// ANNN	-> Store memory address NNN in register I
	case 0xB000:
		// BNNN	-> Jump to address NNN + V0
	case 0xC000:
		// CXNN	-> Set VX to a random number with a mask of NN
	case 0xD000:
		// DXYN	-> Draw a sprite at position VX, VY with N bytes of sprite data starting at the address stored in I
		// Set VF to 01 if any set pixels are changed to unset, and 00 otherwise
	case 0xE000:
		switch vm.opcode & 0x00FF {
		case 0x009E:
			// EX9E	-> Skip the following instruction if the key corresponding to the hex value currently stored in register VX is pressed
		case 0x00A1:
			// EXA1	-> Skip the following instruction if the key corresponding to the hex value currently stored in register VX is not pressed
		default:
			fmt.Printf("unknown opcode: %x\n", vm.opcode&0x00FF)
		}
	case 0xF000:
		switch vm.opcode & 0x00FF {
		case 0x0007:
			// FX07 -> Store the current value of the delay timer in register VX
		case 0x000A:
			// FX0A -> Wait for a keypress and store the result in register VX
		case 0x0015:
			// FX15 -> Set the delay timer to the value of register VX
		case 0x0018:
			// FX18 -> Set the sound timer to the value of register VX
		case 0x001E:
			// FX1E -> Add the value stored in register VX to register I
		case 0x0029:
			// FX29 -> Set I to the memory address of the sprite data corresponding to the hexadecimal digit stored in register VX
		case 0x0033:
			// FX33 -> Store the binary-coded decimal equivalent of the value stored in register VX at addresses I, I+1, and I+2
		case 0x0055:
			// FX55 -> Store the values of registers V0 to VX inclusive in memory starting at address I
			// I is set to I + X + 1 after operation
		case 0x0065:
			// FX65 -> Fill registers V0 to VX inclusive with the values stored in memory starting at address I
			// I is set to I + X + 1 after operation
		default:
			fmt.Printf("unknown opcode: %x\n", vm.opcode&0x00FF)
		}
	default:
		fmt.Printf("unknown opcode: %x\n", vm.opcode&0x00FF)
	}
}

func (vm *VM) debug() {
	fmt.Printf(`
opcode: %x
pc: %d
sp: %d
i: %d

Registers:
V0: %d
V1: %d
V2: %d
V3: %d
V4: %d
V5: %d
V6: %d
V7: %d
V8: %d
V9: %d
VA: %d
VB: %d
VC: %d
VD: %d
VE: %d
VF: %d`,
		vm.opcode, vm.pc, vm.sp, vm.i,
		vm.v[0], vm.v[1], vm.v[2], vm.v[3],
		vm.v[4], vm.v[5], vm.v[6], vm.v[7],
		vm.v[8], vm.v[9], vm.v[10], vm.v[11],
		vm.v[12], vm.v[13], vm.v[14], vm.v[15],
	)
}
