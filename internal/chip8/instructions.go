package chip8

import "math/rand"

func (vm *VM) _0x00E0() {
	vm.gfx = [64 * 32]byte{}
	vm.pc += 2
}

func (vm *VM) _0x00EE() {
	vm.pc = vm.stack[vm.sp] + 2
	vm.sp--
}

func (vm *VM) _0x1000(nnn uint16) {
	vm.pc = nnn
}

func (vm *VM) _0x2000(nnn uint16) {
	vm.sp++
	vm.stack[vm.sp] = vm.pc
	vm.pc = nnn
}

func (vm *VM) _0x3000(x uint16, nn byte) {
	if vm.v[x] == nn {
		vm.pc += 4
	} else {
		vm.pc += 2
	}
}

func (vm *VM) _0x4000(x uint16, nn byte) {
	if vm.v[x] != nn {
		vm.pc += 4
	} else {
		vm.pc += 2
	}
}

func (vm *VM) _0x5000(x, y uint16) {
	if vm.v[x] == vm.v[y] {
		vm.pc += 4
	} else {
		vm.pc += 2
	}
}

func (vm *VM) _0x6000(x uint16, nn byte) {
	vm.v[x] = nn
	vm.pc += 2
}

func (vm *VM) _0x7000(x uint16, nn byte) {
	vm.v[x] += nn
	vm.pc += 2
}

func (vm *VM) _0x0000(x, y uint16) {
	vm.v[x] = vm.v[y]
	vm.pc += 2
}

func (vm *VM) _0x0001(x, y uint16) {
	vm.v[x] |= vm.v[y]
	vm.pc += 2
}

func (vm *VM) _0x0002(x, y uint16) {
	vm.v[x] &= vm.v[y]
	vm.pc += 2
}

func (vm *VM) _0x0003(x, y uint16) {
	vm.v[x] ^= vm.v[y]
	vm.pc += 2
}

// Set VF to 01 if a carry occurs
// Set VF to 00 if a carry does not occur
func (vm *VM) _0x0004(x, y uint16) {
	if vm.v[y] > (0xFF - vm.v[x]) {
		vm.v[0xF] = 1
	} else {
		vm.v[0xF] = 0
	}
	vm.v[x] += vm.v[y]
	vm.pc += 2
}

// Set VF to 00 if a borrow occurs
// Set VF to 01 if a borrow does not occur
func (vm *VM) _0x0005(x, y uint16) {
	if vm.v[y] > vm.v[x] {
		vm.v[0xF] = 0
	} else {
		vm.v[0xF] = 1
	}
	vm.v[x] -= vm.v[y]
	vm.pc += 2
}

// Set register VF to the least significant bit prior to the shift
func (vm *VM) _0x0006(x, y uint16) {
	vm.v[x] = vm.v[y] >> 1
	vm.v[0xF] = vm.v[y] & 0x01
	vm.pc += 2
}

// Set VF to 00 if a borrow occurs
// Set VF to 01 if a borrow does not occur
func (vm *VM) _0x0007_1(x, y uint16) {
	if vm.v[x] > vm.v[y] {
		vm.v[0xF] = 0
	} else {
		vm.v[0xF] = 1
	}
	vm.v[x] = vm.v[y] - vm.v[x]
	vm.pc += 2
}

// Set register VF to the most significant bit prior to the shift
func (vm *VM) _0x000E(x, y uint16) {
	vm.v[x] = vm.v[y] << 1
	vm.v[0xF] = vm.v[y] & 0x80
	vm.pc += 2
}

func (vm *VM) _0x9000(x, y uint16) {
	if vm.v[x] != vm.v[y] {
		vm.pc += 4
	} else {
		vm.pc += 2
	}
}

func (vm *VM) _0xA000(nnn uint16) {
	vm.i = nnn
	vm.pc += 2
}

func (vm *VM) _0xB000(nnn uint16) {
	vm.pc = nnn + uint16(vm.v[0])
	vm.pc += 2
}

func (vm *VM) _0xC000(x uint16, nn byte) {
	vm.v[x] = byte(rand.Float32()*255) & nn
	vm.pc += 2
}

// Set VF to 01 if any set pixels are changed to unset, and 00 otherwise
// Get the starting x and y coordinates of the graphics array.
func (vm *VM) _0xD000(x, y uint16) {
	x = uint16(vm.v[x])
	y = uint16(vm.v[y])
	vm.drawSprite(x, y)
	vm.pc += 2
}

func (vm *VM) _0x009E(x uint16) {
	if vm.keypad[vm.v[x]] == 1 {
		vm.pc += 4
		vm.keypad[vm.v[x]] = 0
	} else {
		vm.pc += 2
	}
}

func (vm *VM) _0x00A1(x uint16) {
	if vm.keypad[vm.v[x]] == 0 {
		vm.pc += 4
	} else {
		vm.keypad[vm.v[x]] = 0
		vm.pc += 2
	}
}

func (vm *VM) _0x0007_2(x uint16) {
	vm.v[x] = vm.delayTimer
	vm.pc += 2
}

func (vm *VM) _0x000A(x uint16) {
	for i, k := range vm.keypad {
		if k != 0 {
			vm.v[x] = byte(i)
			vm.pc += 2
			break
		}
	}
	vm.keypad[vm.v[x]] = 0
}

func (vm *VM) _0x0015(x uint16) {
	vm.delayTimer = vm.v[x]
	vm.pc += 2
}

func (vm *VM) _0x0018(x uint16) {
	vm.soundTimer = vm.v[x]
	vm.pc += 2
}

func (vm *VM) _0x001E(x uint16) {
	vm.i += uint16(vm.v[x])
	vm.pc += 2
}

func (vm *VM) _0x0029(x uint16) {
	vm.i = uint16(vm.v[x]) * 5
	vm.pc += 2
}

func (vm *VM) _0x0033(x uint16) {
	vm.memory[vm.i] = vm.v[x] / 100
	vm.memory[vm.i+1] = (vm.v[x] / 10) % 10
	vm.memory[vm.i+2] = (vm.v[x] % 100) % 10
	vm.pc += 2
}

// i is set to i+x+1 after operation
func (vm *VM) _0x0065(x uint16) {
	for ind := uint16(0); ind <= x; ind++ {
		vm.v[ind] = vm.memory[vm.i+ind]
	}
	vm.pc += 2
}

// i is set to i+x+1 after operation
func (vm *VM) _0x0055(x uint16) {
	for ind := uint16(0); ind <= x; ind++ {
		vm.memory[vm.i+ind] = vm.v[ind]
	}
	vm.pc += 2
}
