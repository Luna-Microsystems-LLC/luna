package bios
import (
	"luna_l2/video"
	"luna_l2/types"
	"luna_l2/audio"
	"time"	
	"os"
	"fmt"
)

var TypeOut bool = false
var KeyTrap bool = false
var Registers *[]types.Register
var Memory *[0x70000000]byte
var KeyInterruptCode uint32 = 0x5
const (
	MEMSIZE uint32 = 0x70000000
	MEMCAP uint32 = 0x6FFFFFFF
)

func WriteChar(char string, fg uint8, bg uint8) {
	video.PrintChar(rune(char[0]), byte(fg), byte(bg))
}

func WriteString(str string, fg uint8, bg uint8) {
	for _, r := range str {
		WriteChar(string(r), fg, bg)
	}
}

func WriteLine(str string, fg uint8, bg uint8) {
	WriteString(str + "\n", fg, bg)
}

func setRegister(address uint32, value uint32) {
	for i := range (*Registers) {
		if (*Registers)[i].Address == address {
			if types.Bits32 == false {
				(*Registers)[i].Value = uint32(uint16(value))
			} else {
				(*Registers)[i].Value = value
			}
		}
	}
}

func getRegister(address uint32) uint32 {
	for _, register := range (*Registers) {
		if register.Address == address {
			return register.Value
		}
	}
	return 0x0000
}

func IntHandler(code uint32) {
	if code == 0x01 {
		// BIOS print to screen
		// start address in R1
		// Foreground in R2
		// Background in R3
		char := getRegister(0x0001)
		WriteChar(string(rune(char)), uint8(getRegister(0x0002)), uint8(getRegister(0x0003)))
	} else if code == 0x02 {
		// BIOS sleep
		// seconds in R1
		timeToSleep := getRegister(0x0001)
		time.Sleep(time.Duration(timeToSleep) * time.Millisecond)
	} else if code == 0x03 {
		// BIOS write to VRAM
		// address in R1, word in R2
		address := getRegister(0x0001)
		word := getRegister(0x0002)
		if types.Bits32 == false {
			video.MemoryVideo[video.Clamp(address, 0, MEMCAP)] = byte(uint16(word) >> 8)
			video.MemoryVideo[video.Clamp(address + 1, 0, MEMCAP)] = byte(uint16(word) & 0xFF)
		} else {
			video.MemoryVideo[video.Clamp(address, 0, MEMCAP)] = byte(uint32(word) >> 24)
			video.MemoryVideo[video.Clamp(address + 1, 0, MEMCAP)] = byte(uint32(word) >> 16)
			video.MemoryVideo[video.Clamp(address + 2, 0, MEMCAP)] = byte(uint32(word) >> 8)
			video.MemoryVideo[video.Clamp(address + 3, 0, MEMCAP)] = byte(uint32(word) & 0xFF)
		}
	} else if code == 0x4 {
		// BIOS configure input mode
		// Mode 1: no type output
		// Mode 2: type output
		// In R1
		if getRegister(0x0001) == 1 {
			TypeOut = true
		} else {
			TypeOut = false
		}
	} else if code == 0x5 {
		// BIOS key event	
		if TypeOut == true {
			WriteChar(string(rune(getRegister(0x001b))), uint8(255), uint8(0))	
		}
		if KeyTrap == true {
			KeyTrap = false
			setRegister(0x0001, getRegister(0x001b))
		}
	} else if code == 0x6 {
		// BIOS wait for key
		// Return in R1 via interrupt 5
		KeyTrap = true
		for {
			if KeyTrap == true {
				time.Sleep(500)
			} else {
				break
			}
		}
	} else if code == 0x7 {
		WriteLine("Illegal instruction 0x" + fmt.Sprintf("%08x", getRegister(0x0001)) + " at location 0x" + fmt.Sprintf("%08x", getRegister(0x001a)), 255, 0)
		return
	} else if code == 0x8 {
		// BIOS write to ARAM
		// address in R1, word in R2
		address := getRegister(0x0001)
		word := getRegister(0x0002)
		if types.Bits32 == false {
			audio.MemoryAudio[video.Clamp(address, 0, MEMCAP)] = byte(uint16(word) >> 8)
			audio.MemoryAudio[video.Clamp(address + 1, 0, MEMCAP)] = byte(uint16(word) & 0xFF)
		} else {
			audio.MemoryAudio[video.Clamp(address, 0, MEMCAP)] = byte(uint32(word) >> 24)
			audio.MemoryAudio[video.Clamp(address + 1, 0, MEMCAP)] = byte(uint32(word) >> 16)
			audio.MemoryAudio[video.Clamp(address + 2, 0, MEMCAP)] = byte(uint32(word) >> 8)
			audio.MemoryAudio[video.Clamp(address + 3, 0, MEMCAP)] = byte(uint32(word) & 0xFF)
		}	
	} else if code == 0x9 {
		audio.Play()
	} else if code == 0xa {
		if types.Bits32 == false {
			setRegister(0x0001, 0xffff)
		} else {
			setRegister(0x0001, MEMSIZE)
		}
	}
}

func Splash() {
	WriteLine("Luna L2", 255, 0)
	WriteLine("BIOS: Integrated BIOS", 255, 0)	
	WriteLine("Copyright (c) 2025 Luna Microsystems LLC\n", 255, 0)
}

func CheckArgs() bool {
	if len(os.Args) < 2 {
		WriteLine("No bootable device", 255, 0)
		return false
	}
	return true
}
