/*
Luna L2 CPU implementation
Find more details at luna.alexflax.xyz

Luna is an open-source CPU implementation.

Get LASM (Luna assembler) and LCC (Luna C compiler) as well at:
https://github.com/alexfdev0/luna

Free to use, modify, and redistribute under Apache 2.0 license (the "License").
You may not use this file except in compliance with the License.
You may obtain a copy of the license at https://www.apache.org/licenses/LICENSE-2.0
Unless required by applicable law or agreed to in writing, software distributed
under the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR
CONDITIONS OF ANY KIND, either express or implied. See the License for the specific
language governing permissions and limitations under the License.

Specs:
16 bit CPU
65 KB memory
Sound capabilities
Big endian

Supported host OSes:
Linux x86_64
MacOS x86_64
Windows x86_64

Copyright (c) 2025 Luna Microsystems LLC, under the Apache 2.0 license.
*/

package main

import (
	"fmt"
	"os"
	"io/ioutil"
	"time"
	_ "embed"
	"github.com/faiface/beep"
	"github.com/faiface/beep/speaker"
	"github.com/faiface/beep/mp3"
	"bytes"
	"io"
	"runtime"
)

//go:embed sounds/crash.mp3
var crashSoundData []byte

type Register struct {
	Address int16
	Name string
	Value int16
}

var Registers = []Register {
	{0x0000, "R0", 0},
	{0x0001, "R1", 0},
	{0x0002, "R2", 0},
	{0x0003, "R3", 0},
	{0x0004, "R4", 0},
	{0x0005, "R5", 0},
	{0x0006, "R6", 0},
	{0x0007, "R7", 0},
	{0x0008, "R8", 0},
	{0x0009, "R9", 0},
	{0x000a, "R10", 0},
	{0x000b, "R11", 0},
	{0x000c, "R12", 0},
	{0x000d, "T1", 0},
	{0x000e, "T2", 0},
	{0x000f, "T3", 0},
	{0x0010, "T4", 0},
	{0x0011, "T5", 0},
	{0x0012, "T6", 0},
	{0x0013, "T7", 0},
	{0x0014, "T8", 0},
	{0x0015, "T9", 0},
	{0x0016, "T10", 0},
	{0x0017, "T11", 0},
	{0x0018, "T12", 0},
	{0x0019, "SP", 0},
	{0x001a, "PC", 0},
	{0x001b, "PTR", 0},
}

var Memory[65535]byte

func setRegister(address int16, value int16) {
	if address == 0x001a {
		value = int16(uint16(value))
	}
    for i := range Registers {
        if Registers[i].Address == address {
            Registers[i].Value = value
            return
        }
    }
}

func getRegister(address int16) int16 {	
	for _, register := range Registers {
		if register.Address == address {
			return register.Value
		}
	}
	return 0x0000
}

func playSound(soundName string) {
	if soundName == "crash" {
		streamer, format, err := mp3.Decode(io.NopCloser(bytes.NewReader(crashSoundData)))
		if err != nil {
			return
		}
		defer streamer.Close()

		speaker.Init(format.SampleRate, format.SampleRate.N(time.Second / 10))

		done := make(chan bool)
		speaker.Play(beep.Seq(streamer, beep.Callback(func() {
			done <- true
		})))

		<-done
	}
}

func intHandler(code uint16) {
	if code == 0x01 {
		// BIOS print to screen
		// start address in R1
		addr := getRegister(0x0001)

		for i := uint16(addr); i <= 65535; i++ {
			if int16(Memory[i]) != 0x00 {
				fmt.Printf(string(rune(int16(Memory[i]))))	
			} else {
				break
			}
		}
	}
}

func execute() {
	for {
		ProgramCounter := getRegister(0x001a)
		op := Memory[ProgramCounter]
	
		switch op {
		case 0x00:
			return
		case 0x01:
			// MOV
            mode := Memory[ProgramCounter + 1]
            dst := Memory[ProgramCounter + 2]

            if mode == 0x01 {
                imm := int16(Memory[ProgramCounter + 3])<<8 | int16(Memory[ProgramCounter + 4])
                setRegister(int16(dst), imm)
                setRegister(0x001a, ProgramCounter + 5)	
            } else if mode == 0x02 {
				frm := int16(Memory[ProgramCounter + 3])
				setRegister(int16(dst), int16(getRegister(frm)))
				setRegister(0x001a, ProgramCounter + 4)
			}
		case 0x02:
			// HLT
			for {
				time.Sleep(time.Second)
			}
			setRegister(0x001a, ProgramCounter + 1)
		case 0x03:
			// JMP
			mode := Memory[ProgramCounter + 1]

			if mode == 0x01 {
				loc := int16(Memory[ProgramCounter + 2]) << 8 | int16(Memory[ProgramCounter + 3])
				setRegister(0x001a, loc)	
			} else if mode == 0x02 {
				frm := int16(Memory[ProgramCounter + 2])
				setRegister(0x001a, getRegister(frm))
			}
		case 0x04:
			// INT	
			code := uint16(Memory[ProgramCounter + 1]) << 8 | uint16(Memory[ProgramCounter + 2])	
			intHandler(code)
			setRegister(0x001a, ProgramCounter + 3)
		case 0x05:
			// DSTART	
			// Indicates label
			for i := uint16(ProgramCounter + 1); i < 65535; i++ {
				if uint16(Memory[i]) == 0x07 {
					setRegister(0x001a, int16(i + 1))
					break
				}
			}
		case 0x06:
			// DSEP
			setRegister(0x001a, ProgramCounter + 1)
		case 0x07:
			// DEND
			setRegister(0x001a, ProgramCounter + 1)
		case 0x08:
			// JNZ
			// jnz <mode (01 or 02)> <check register> <loc (register or raw addr)> 
			mode := Memory[ProgramCounter + 1]
			checkRegister := Memory[ProgramCounter + 2]
			var loc int16 = 0
			var not int16 = 0

			if mode == 0x01 {
				loc = int16(Memory[ProgramCounter + 3]) << 8 | int16(Memory[ProgramCounter + 4])
				not = ProgramCounter + 5
			} else if mode == 0x02 {
				frm := int16(Memory[ProgramCounter + 3])
				loc = getRegister(frm)
				not = ProgramCounter + 4
			}

			if getRegister(int16(checkRegister)) != 0 {
				setRegister(0x001a, loc)
			} else {
				setRegister(0x001a, not)	
			}
		case 0x09:
			// NOP

		default:
			fmt.Printf("FATAL: Illegal instruction 0x%02x at instruction 0x%02x\n", op, ProgramCounter)
			playSound("crash")
			return
		}
	}
}

func main() {	
	fmt.Println("Luna L2")
	fmt.Println("BIOS: Integrated BIOS")
	fmt.Println("Host: " + runtime.GOOS)
	fmt.Println("Host CPU: " + runtime.GOARCH)
	fmt.Println("Copyright (c) 2025 Luna Microsystems LLC\n\n")

	if len(os.Args) < 2 {
		fmt.Println("FATAL: Disk image not found.")
		playSound("crash")
		for {
			time.Sleep(time.Second)
		}
	}

	filename := os.Args[1]

	data, err := ioutil.ReadFile(filename)
	if err != nil {
		fmt.Println("FATAL: Failed to read disk image.", err)
		playSound("crash")
		for {
			time.Sleep(time.Second)
		}
	}
	copy(Memory[:], data)
	execute()	
	for {
		time.Sleep(time.Second)
	}
}
