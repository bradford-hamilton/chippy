package keypad

// Keypad represents the key interactions with the vm
type Keypad struct {
	eventPump string // TODO this must be sdl EventPump
}

// New returns a pointer to a keypad
func New() *Keypad {
	return new(Keypad)
}
