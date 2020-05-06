<div align="center">
  <img
    alt="bag of chippys"
    src="./assets/chippy.jpg"
    height="300px"
  />
</div>
<h1 align="center">Welcome to Chippy 👋</h1>
<p align="center">
  <a href="https://golang.org/dl" target="_blank">
    <img alt="Using go version 1.14" src="https://img.shields.io/badge/go-1.14-9cf.svg" />
  </a>
  <a href="https://goreportcard.com/report/github.com/bradford-hamilton/chippy" target="_blank">
    <img alt="Go Report Card" src="https://goreportcard.com/badge/github.com/bradford-hamilton/chippy/pkg" />
  </a>
  <a href="#" target="_blank">
    <img alt="License: MIT" src="https://img.shields.io/badge/License-MIT-yellow.svg" />
  </a>
</p>

Chippy is a [CHIP-8](https://en.wikipedia.org/wiki/CHIP-8) emulator that runs Chip-8 public domain roms. The Chip 8 actually never was a real system, but more like a virtual machine (VM) developed in the 70’s by Joseph Weisbecker. Games written in the Chip 8 language could easily run on systems that had a Chip 8 interpreter.

> Audio beeps currently not working

Current sources:
- [post by Laurence Muller](http://www.multigesture.net/articles/how-to-write-an-emulator-chip-8-interpreter)
- [CHIP-8 Wiki](https://en.wikipedia.org/wiki/CHIP-8)
- [cowgod's chip-8 technical reference](http://devernay.free.fr/hacks/chip8/C8TECH10.HTM)
- [chip-8 instruction set](http://www.multigesture.net/wp-content/uploads/mirror/goldroad/chip8_instruction_set.shtml)
- [post by Matthew Mikolay](http://mattmik.com/files/chip8/mastering/chip8.html)

## Installation
```
go install github.com/bradford-hamilton/chippy
```
I am still getting a bunch of deprecation warnings when building. If you see those warnings just ignore them.

The screen uses CGO which isn't supported by [go-releaser](https://github.com/goreleaser/goreleaser) :( which means unfortunately I don't have a nice releases section with binaries for multiple systems.

## Usage
### Run
```
chippy run roms/pong.ch8
```

### Version
```
chippy version
```

### Help
```
chippy help
```

Pong

![pong](assets/pong.png)

---

Space Invaders

![space_invaders](assets/space_invaders.png)

---

IBM Logo

![IBM Logo](assets/ibm_logo.png)


### Show your support

Give a ⭐ if this project was helpful in any way!
