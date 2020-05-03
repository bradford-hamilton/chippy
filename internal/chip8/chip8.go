// Package chip8 is a Chip-8 emulator written in Go. Chip-8 used to be implemented on 4k systems like the Telmac 1800 and
// Cosmac VIP where the chip-8 interprete itself occpied the first 512 bytes of memory (up to 0x200). In modern CHIP-8
// implementations (like ours here), where the interpreter is running natively outside the 4K memory space, there is no
// need to avoid the lower 512 bytes of memory (0x000-0x200), and it is common to store font data there.
package chip8

import (
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"time"

	"github.com/bradford-hamilton/chippy/internal/pixel"
	"github.com/faiface/beep/mp3"
	"github.com/faiface/beep/speaker"
)

//		System memory map
// 		+---------------+= 0xFFF (4095) End Chip-8 RAM
// 		|               |
// 		|               |
// 		|               |
// 		|               |
// 		|               |
// 		| 0x200 to 0xFFF|
// 		|     Chip-8    |
// 		| Program / Data|
// 		|     Space     |
// 		|               |
// 		|               |
// 		|               |
// 		+- - - - - - - -+= 0x600 (1536) Start ETI 660 Chip-8 programs
// 		|               |
// 		|               |
// 		|               |
// 		+---------------+= 0x200 (512) Start of most Chip-8 programs
// 		| 0x000 to 0x1FF|
// 		| Reserved for  |
// 		|  interpreter  |
// 		+---------------+= 0x000 (0) Begin Chip-8 RAM. We store font data here instead of storing the interpreter because we don't have that restriction.
//

// VM represents the chip-8 virtual machine
type VM struct {
	// Chip-8 system memory, see memory map above
	memory [4096]byte

	// Opcode under examination
	opcode uint16

	// 8-bit general purpose register, (V0 - VE*)
	v [16]byte

	// index register (0x000 to 0xFFF)
	i uint16

	// Program counter (0x000 to 0xFFF)
	pc uint16

	// Internal stack to store return addresses when calling procedures
	stack [16]uint16

	// Stack pointer is used to store return locations from the program counter register
	sp uint16

	// Represents window pixels. Bytes get flipped on and off inside to guide drawing
	gfx [64 * 32]byte

	// 8-bit delay timer which counts down at 60 hertz, until it reaches 0
	delayTimer byte

	// 8-bit sound timer which counts down at 60 hertz, until it reaches 0
	soundTimer byte

	// Keypad is HEX based: 0x0-0xF
	//  1  2  3  C
	//  4  5  6  D
	//  7  8  9  E
	//  A  0  B  F
	keypad [16]byte

	// Chippy doesn't draw on every cycle, set draw flag when we need to update screen.
	drawFlag bool

	// Embedded pixel window for displaying ROMs
	window *pixel.Window

	// Our "CPU clock"
	Clock *time.Ticker

	// Channel for sending/receiving audio events
	audioChan chan struct{}

	// Channel for sending/receiving a shutdown signal
	Shutdown chan struct{}
}

const keyRepeatDur = time.Second / 5
const refreshRate = 300
const maxRomSize = 0xFFF - 0x200

// NewVM initializes a Window and a VM, loads the font set and the
// ROM into memory, and returns a pointer to the VM or an error
func NewVM(pathToROM string) (*VM, error) {
	window, err := pixel.NewWindow()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	vm := VM{
		memory:    [4096]byte{},
		v:         [16]byte{},
		pc:        0x200,
		stack:     [16]uint16{},
		gfx:       [64 * 32]byte{},
		keypad:    [16]byte{},
		window:    window,
		Clock:     time.NewTicker(time.Second / refreshRate),
		audioChan: make(chan struct{}),
		Shutdown:  make(chan struct{}),
	}
	vm.loadFontSet()
	if err := vm.loadROM(pathToROM); err != nil {
		return nil, err
	}
	return &vm, nil
}

// Run starts the vm and emulates a clock that runs by default at 60MHz
// This can be changed with a flag.
func (vm *VM) Run() {
	for {
		select {
		case <-vm.Clock.C:
			if !vm.window.Closed() {
				vm.emulateCycle()
				vm.drawOrUpdate()
				vm.handleKeyInput()
				vm.delayTimerTick()
				vm.soundTimerTick()
				continue
			}
			break
		case <-vm.Shutdown:
			break
		}
		break
	}
	vm.signalShutdown("Received signal - gracefully shutting down...")
}

// loads the font set into the first 80 bytes of memory
func (vm *VM) loadFontSet() {
	for i := 0; i < 80; i++ {
		vm.memory[i] = pixel.FontSet[i]
	}
}

func (vm *VM) loadROM(path string) error {
	rom, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}
	if len(rom) > maxRomSize {
		panic("error: rom too large. Max size: 3583")
	}
	for i := 0; i < len(rom); i++ {
		// Ensure we write memory starting at the program counter
		vm.memory[0x200+i] = rom[i]
	}
	return nil
}

// emulateCycle runs a full fetch, decode, and execute cycle.
// One opcode is 2 bytes long (ex. 0xA2FO) so we need to fetch two successive bytes (ex. 0xA2 and 0xF0) and merge them
// to get the actual opcode. First we shift current instruction left 8 (ex. from 10100010 -> 1010001000000000)
// Then we OR it with the upcoming byte which gives us a 16 bit chunk containing the combined bytes
func (vm *VM) emulateCycle() {
	vm.opcode = uint16(vm.memory[vm.pc])<<8 | uint16(vm.memory[vm.pc+1])
	vm.drawFlag = false

	if err := vm.parseOpcode(); err != nil {
		fmt.Printf("error parsing opcode: %v", err)
	}
}

func (vm *VM) parseOpcode() error {
	x := (vm.opcode & 0x0F00) >> 8 // Decode Vx register identifier.
	y := (vm.opcode & 0x00F0) >> 4 // Decode Vy register identifier
	nn := byte(vm.opcode & 0x00FF) // load last 8-bits
	nnn := vm.opcode & 0x0FFF      // load last 12-bits

	switch vm.opcode & 0xF000 {
	case 0x0000: // 0NNN -> Execute machine language subroutine at address NNN
		switch vm.opcode & 0x00FF {
		case 0x00E0: // 00E0 -> Clear the screen
			vm.gfx = [64 * 32]byte{}
			vm.pc += 2
		case 0x00EE: // 00EE -> Return from a subroutine.
			vm.pc = vm.stack[vm.sp] + 2
			vm.sp--
		default:
			return vm.unknownOp(vm.opcode & 0x00FF)
		}
	case 0x1000: // 1NNN -> Jump to address NNN
		vm.pc = nnn
	case 0x2000: // 2NNN -> Execute subroutine starting at address NNN
		vm.sp++
		vm.stack[vm.sp] = vm.pc
		vm.pc = nnn
	case 0x3000: // 3XNN -> Skip the following instruction if the value of register VX == NN
		if vm.v[x] == nn {
			vm.pc += 4
		} else {
			vm.pc += 2
		}
	case 0x4000: // 4XNN -> Skip the following instruction if the value of register VX != NN
		if vm.v[x] != nn {
			vm.pc += 4
		} else {
			vm.pc += 2
		}
	case 0x5000: // 5XY0 -> Skip the following instruction if the value of register VX == VY
		if vm.v[x] == vm.v[y] {
			vm.pc += 4
		} else {
			vm.pc += 2
		}
	case 0x6000: // 6XNN -> Store number NN in register VX
		vm.v[x] = nn
		vm.pc += 2
	case 0x7000: // 7XNN -> Add the value NN to register VX
		vm.v[x] += nn
		vm.pc += 2
	case 0x8000:
		switch vm.opcode & 0x000F {
		case 0x0000: // 8XYO -> Store the value of register VY in register VX
			vm.v[x] = vm.v[y]
			vm.pc += 2
		case 0x0001: // 8XY1 -> Set VX to VX OR VY
			vm.v[x] |= vm.v[y]
			vm.pc += 2
		case 0x0002: // 8XY2 -> Set VX to VX AND VY
			vm.v[x] &= vm.v[y]
			vm.pc += 2
		case 0x0003: // 8XY3 -> Set VX to VX XOR VY
			vm.v[x] ^= vm.v[y]
			vm.pc += 2
		case 0x0004: // 8XY4 -> Add the value of register VY to register VX
			// Set VF to 01 if a carry occurs
			// Set VF to 00 if a carry does not occur
			if vm.v[y] > (0xFF - vm.v[x]) {
				vm.v[0xF] = 1
			} else {
				vm.v[0xF] = 0
			}
			vm.v[x] += vm.v[y]
			vm.pc += 2
		case 0x0005: // 8XY5 -> Subtract the value of register VY from register VX
			// Set VF to 00 if a borrow occurs
			// Set VF to 01 if a borrow does not occur
			if vm.v[y] > vm.v[x] {
				vm.v[0xF] = 0
			} else {
				vm.v[0xF] = 1
			}
			vm.v[x] -= vm.v[y]
			vm.pc += 2
		case 0x0006: // 8XY6 -> Store the value of register VY shifted right one bit in register VX
			// Set register VF to the least significant bit prior to the shift
			vm.v[x] = vm.v[y] >> 1
			vm.v[0xF] = vm.v[y] & 0x01
			vm.pc += 2
		case 0x0007: // 8XY7 -> Set register VX to the value of VY minus VX
			// Set VF to 00 if a borrow occurs
			// Set VF to 01 if a borrow does not occur
			if vm.v[x] > vm.v[y] {
				vm.v[0xF] = 0
			} else {
				vm.v[0xF] = 1
			}
			vm.v[x] = vm.v[y] - vm.v[x]
			vm.pc += 2
		case 0x000E: // 8XYE -> Store the value of register VY shifted left one bit in register VX
			// Set register VF to the most significant bit prior to the shift
			vm.v[x] = vm.v[y] << 1
			vm.v[0xF] = vm.v[y] & 0x80
			vm.pc += 2
		default:
			return vm.unknownOp(vm.opcode & 0x000F)
		}
	case 0x9000: // 9XY0 -> Skip the following instruction if the value of VX != value of VY
		if vm.v[x] != vm.v[y] {
			vm.pc += 4
		} else {
			vm.pc += 2
		}
	case 0xA000: // ANNN -> Store memory address NNN in index register
		vm.i = nnn
		vm.pc += 2
	case 0xB000: // BNNN -> Jump to address NNN + V0
		vm.pc = nnn + uint16(vm.v[0])
		vm.pc += 2
	case 0xC000: // CXNN -> Set VX to a random number from 0-255 with a mask of NN
		vm.v[x] = byte(rand.Float32()*255) & nn
		vm.pc += 2
	case 0xD000: // DXYN -> Draw a sprite at position VX, VY with N bytes of sprite data starting at the address stored in index register
		// Set VF to 01 if any set pixels are changed to unset, and 00 otherwise

		// get the starting x and y coordinates of the graphics array.
		x = uint16(vm.v[x])
		y = uint16(vm.v[y])

		height := vm.opcode & 0x000F
		vm.v[0xF] = 0

		var pix uint16
		for yLine := uint16(0); yLine < height; yLine++ {
			pix = uint16(vm.memory[vm.i+yLine])

			for xLine := uint16(0); xLine < 8; xLine++ {
				ind := (x + xLine + ((y + yLine) * 64))
				if ind >= uint16(len(vm.getGraphics())) {
					continue
				}

				if (pix & (0x80 >> xLine)) != 0 {
					if vm.getGraphics()[ind] == 1 {
						vm.v[0xF] = 1
					}
					vm.gfx[ind] ^= 1
				}
			}
		}

		vm.drawFlag = true
		vm.pc += 2
	case 0xE000:
		switch vm.opcode & 0x00FF {
		case 0x009E: // EX9E -> Skip the following instruction if the key corresponding to the hex value currently stored in register VX is pressed
			if vm.keypad[vm.v[x]] == 1 {
				vm.pc += 4
				vm.keypad[vm.v[x]] = 0
			} else {
				vm.pc += 2
			}
		case 0x00A1: // EXA1 -> Skip the following instruction if the key corresponding to the hex value currently stored in register VX is not pressed
			if vm.keypad[vm.v[x]] == 0 {
				vm.pc += 4
			} else {
				vm.keypad[vm.v[x]] = 0
				vm.pc += 2
			}
		default:
			return vm.unknownOp(vm.opcode & 0x00FF)
		}
	case 0xF000:
		switch vm.opcode & 0x00FF {
		case 0x0007: // FX07 -> Store the current value of the delay timer in register VX
			vm.v[x] = vm.delayTimer
			vm.pc += 2
		case 0x000A: // FX0A -> Wait for a keypress and store the result in register VX
			for i, k := range vm.keypad {
				if k != 0 {
					vm.v[x] = byte(i)
					vm.pc += 2
					break
				}
			}
			vm.keypad[vm.v[x]] = 0
		case 0x0015: // FX15 -> Set the delay timer to the value of register VX
			vm.delayTimer = vm.v[x]
			vm.pc += 2
		case 0x0018: // FX18 -> Set the sound timer to the value of register VX
			vm.soundTimer = vm.v[x]
			vm.pc += 2
		case 0x001E: // FX1E -> Add the value stored in register VX to index register
			vm.i += uint16(vm.v[x])
			vm.pc += 2
		case 0x0029: // FX29 -> Set index register to the memory address of the sprite data corresponding to the hexadecimal digit stored in register VX
			vm.i = uint16(vm.v[x]) * 5
			vm.pc += 2
		case 0x0033: // FX33 -> Store the binary-coded decimal equivalent of the value stored in register VX at addresses i, i+1, and i+2
			vm.memory[vm.i] = vm.v[x] / 100
			vm.memory[vm.i+1] = (vm.v[x] / 10) % 10
			vm.memory[vm.i+2] = (vm.v[x] % 100) % 10
			vm.pc += 2
		case 0x0055: // FX55 -> Store the values of registers V0 to VX inclusive in memory starting at address i
			// i is set to i+x+1 after operation
			for ind := uint16(0); ind <= x; ind++ {
				vm.memory[vm.i+ind] = vm.v[ind]
			}
			vm.pc += 2
		case 0x0065: // FX65 -> Fill registers V0 to VX inclusive with the values stored in memory starting at address i
			// i is set to i+x+1 after operation
			for ind := uint16(0); ind <= x; ind++ {
				vm.v[ind] = vm.memory[vm.i+ind]
			}
			vm.pc += 2
		default:
			return vm.unknownOp(vm.opcode & 0x00FF)
		}
	default:
		return vm.unknownOp(vm.opcode & 0x00FF)
	}
	return nil
}

func (vm *VM) getGraphics() [64 * 32]byte {
	return vm.gfx
}

func (vm *VM) setKeyDown(index byte) {
	vm.keypad[index] = 1
}

func (vm *VM) unknownOp(opcode uint16) error {
	return fmt.Errorf("unknown opcode: %x", opcode)
}

func (vm *VM) handleKeyInput() {
	for i, key := range vm.window.KeyMap {
		if vm.window.JustReleased(key) && vm.window.KeysDown[i] != nil {
			vm.window.KeysDown[i].Stop()
			vm.window.KeysDown[i] = nil
		} else if vm.window.JustPressed(key) {
			if vm.window.KeysDown[i] == nil {
				vm.window.KeysDown[i] = time.NewTicker(keyRepeatDur)
			}
			vm.setKeyDown(byte(i))
		}

		if vm.window.KeysDown[i] == nil {
			continue
		}

		select {
		case <-vm.window.KeysDown[i].C:
			vm.setKeyDown(byte(i))
		default:
		}
	}
}

// ManageAudio reads and decodes the beep.mp3, initializes the speaker, and plays
// a beep each time an audio event is placed on the channel
func (vm *VM) ManageAudio() {
	f, err := os.Open("assets/beep.mp3")
	if err != nil {
		return
	}

	streamer, format, err := mp3.Decode(f)
	if err != nil {
		return
	}
	defer streamer.Close()

	speaker.Init(
		format.SampleRate,
		format.SampleRate.N(time.Second/10),
	)

	for range vm.audioChan {
		speaker.Play(streamer)
	}
}

func (vm *VM) drawOrUpdate() {
	if vm.drawFlag {
		vm.window.DrawGraphics(vm.getGraphics())
	} else {
		vm.window.UpdateInput()
	}
}

func (vm *VM) delayTimerTick() {
	if vm.delayTimer > 0 {
		vm.delayTimer--
	}
}

func (vm *VM) soundTimerTick() {
	if vm.soundTimer > 0 {
		if vm.soundTimer == 1 {
			vm.audioChan <- struct{}{}
		}
		vm.soundTimer--
	}
}

func (vm *VM) signalShutdown(msg string) {
	fmt.Println(msg)
	close(vm.audioChan)
	vm.Shutdown <- struct{}{}
}

func (vm *VM) debug() {
	fmt.Printf(`
opcode: %x
pc: %d
sp: %d
i: %d
---Registers---
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
		vm.opcode, vm.pc, vm.sp, vm.i, vm.v[0], vm.v[1], vm.v[2], vm.v[3], vm.v[4],
		vm.v[5], vm.v[6], vm.v[7], vm.v[8], vm.v[9], vm.v[10], vm.v[11], vm.v[12],
		vm.v[13], vm.v[14], vm.v[15],
	)
}
