package bios
import (
	"luna_l2/video"
	"luna_l2/types"	
	"time"	
	"os"
	"fmt"
)

var TypeOut bool = false
var KeyTrap bool = false
var Registers *[]types.Register
var Memory *[65535]byte

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

func setRegister(address uint16, value uint16) {
	for i := range (*Registers) {
		if (*Registers)[i].Address == address {
			(*Registers)[i].Value = value
			return
		}
	}
}

func getRegister(address uint16) uint16 {
	for _, register := range (*Registers) {
		if register.Address == address {
			return register.Value
		}
	}
	return 0x0000
}

func IntHandler(code uint16) {
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
		time.Sleep(time.Second * time.Duration(timeToSleep))
	} else if code == 0x03 {
		// BIOS write to VRAM
		// address in R1, word in R2
		address := getRegister(0x0001)
		word := getRegister(0x0002)
		video.MemoryVideo[video.Clamp(address, 0, 63999)] = byte(uint16(word) >> 8)
		video.MemoryVideo[video.Clamp(address + 1, 0, 63999)] = byte(uint16(word) & 0xFF)
	} else if code == 0x4 {
		// BIOS configure input mode
		// Mode 1: no type output
		// Mode 2: type output
		// In R1
		if getRegister(0x0001) == uint16(1) {
			TypeOut = true
		} else {
			TypeOut = false
		}
	} else if code == 0x5 {
		// BIOS configure type out
		// Mode in R1
		if TypeOut == true {
			WriteChar(string(rune(getRegister(0x001b))), uint8(255), uint8(0))	
		}
		if KeyTrap == true {
			KeyTrap = false
			setRegister(uint16(0x0001), getRegister(0x001b))
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
		WriteLine("Illegal instruction 0x" + fmt.Sprintf("%04x", getRegister(0x0001)) + " at location 0x" + fmt.Sprintf("%04x", getRegister(0x001a)), 255, 0)
		return
	}
}

func Splash() {
	WriteLine("Luna L2", 255, 0)
	WriteLine("BIOS: Integrated BIOS", 255, 0)	
	WriteLine("Copyright (c) 2025 Luna Microsystems LLC\n", 255, 0)
}

func CheckImage() bool {
	if (*Memory)[0x0000] != 0x4C || (*Memory)[0x0001] != 0x32 || (*Memory)[0x0002] != 0x45 {
		WriteLine("Invalid disk image", 255, 0)	
		return false
	}
	return true
}

func CheckArgs() bool {
	if len(os.Args) < 2 {
		WriteLine("No bootable device", 255, 0)
		return false
	}
	return true
}
