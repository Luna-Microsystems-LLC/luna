package main

import (	
	"image"
	"image/color"	
	"os"	
	"time"
	"fmt"
	"strconv"
	"bufio"

	"luna_l2/bios"		
	"luna_l2/video"
	"luna_l2/keyboard"
	"luna_l2/types"

	"gioui.org/app"	
	"gioui.org/f32"
	"gioui.org/op"
	"gioui.org/op/paint"
	"gioui.org/op/clip"
	"gioui.org/io/key"
	"gioui.org/io/event"	
)

// Basic elements of CPU
var Registers = []types.Register {
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
	{0x001b, "RE1", 0},
	{0x001c, "RE2", 0},
	{0x001d, "RE3", 0},
}

var Memory [0x70000000]byte
const (
	MEMSIZE uint32 = 0x70000000
	MEMCAP uint32 = 0x6FFFFFFF
)

// Register controls
func setRegister(address uint32, value uint32) {
	for i := range Registers {
		if Registers[i].Address == address {
			if types.Bits32 == false {
				Registers[i].Value = uint32(uint16(value))
			} else {
				Registers[i].Value = value
			}
		}
	}
}

func getRegister(address uint32) uint32 {
	for _, register := range Registers {
		if register.Address == address {
			return register.Value
		}
	}
	return 0x0000
}

func getRegisterName[T uint32 | byte](address T) string {
	addr := uint32(address)
	for _, register := range Registers {
		if register.Address == addr {
			return register.Name
		}
	}
	return ""
}

func Mapper(address uint32) byte {
	if address < MEMSIZE {
		return Memory[address]
	} else {
		return Memory[MEMCAP]
	}
	return Memory[0x00000000]
}

func MapperIndex(address uint32) uint32 {
	if address < MEMSIZE {
		return address
	} else {
		return MEMCAP
	}
	return 0x00000000	
}

// Meta-code
var LogOn bool = false
var Debug bool = false
var ClockSpeed int64 = 1158000
var Filename string = ""
func Log(text string) {
	if LogOn == true {
		fmt.Println("\033[33m" + fmt.Sprintf("0x%08x: ", getRegister(0x001a)) + text + "\033[0m")
	}
}
func LoadSector(sector int, enforce bool) {
	data, err := os.ReadFile(Filename)
	if err != nil {
		if enforce == false {
			fmt.Println("luna-l2: could not reload block device")
			return
		} else {
			fmt.Println("luna-l2: could not open '" + Filename + "'")
			os.Exit(1)
		}
	}
	start := sector * 512
	if start > len(data) {
		Log("read at address " + fmt.Sprintf("0x%08x", start) + " out of bounds")	
		return
	}
	if start + 512 > len(data) {
		copy(Memory[start:start + 512], data[start:])
	} else {
		copy(Memory[start:start + 512], data[start:start + 512])
	}
}

// CPU code
func stall(cycles int64) { 
	cycleTime := int64(int(time.Second)) / ClockSpeed
	time.Sleep(time.Duration(cycleTime * cycles))
}

func execute() {
	for {
		ProgramCounter := getRegister(0x001a)
		op := Mapper(ProgramCounter)
	
		if ProgramCounter == 0x0000 {
			codesect := uint32(Mapper(ProgramCounter)) << 8 | uint32(Mapper(ProgramCounter + 1))
			setRegister(0x001a, codesect)
			continue
		}

		switch op {
		case 0x00:
			return
		case 0x01:
			// MOV
			mode := Mapper(ProgramCounter + 1)
			dst := Mapper(ProgramCounter + 2)

			if mode == 0x01 {
				var imm uint32 = 0
				var next uint32 = 0
				if types.Bits32 == false {
					imm = uint32(uint16(Memory[ProgramCounter + 3]) << 8 | uint16(Memory[ProgramCounter + 4]))
					next = ProgramCounter + 5
				} else {
					imm = uint32(Memory[ProgramCounter + 3]) << 24 | uint32(Memory[ProgramCounter + 4])	<< 16 | uint32(Memory[ProgramCounter + 5]) << 8 | uint32(Memory[ProgramCounter + 6])
					next = ProgramCounter + 7
				}
				setRegister(uint32(dst), imm)
				setRegister(0x001a, next)
				Log("mov " + getRegisterName(uint32(dst)) + ", " + fmt.Sprintf("0x%08x", imm))
			} else if mode == 0x02 {
				frm := uint32(Memory[ProgramCounter+3])
				setRegister(uint32(dst), uint32(getRegister(frm)))
				setRegister(0x001a, ProgramCounter+4)
				Log("mov " + getRegisterName(uint32(dst)) + ", " + getRegisterName(frm))
			}	
			stall(4)
		case 0x02:
			// HLT
			Log("hlt")
			for {
				time.Sleep(time.Second)
			}
			setRegister(0x001a, ProgramCounter+1)
		case 0x03:
			// JMP	
			mode := Memory[ProgramCounter+1]

			if mode == 0x01 {
				var loc uint32 = 0
				if types.Bits32 == false {
					loc = uint32(uint16(Memory[ProgramCounter + 2]) << 8 | uint16(Memory[ProgramCounter + 3]))
				} else {
					loc = uint32(Memory[ProgramCounter + 2]) << 24 | uint32(Memory[ProgramCounter + 3])	<< 16 | uint32(Memory[ProgramCounter + 4]) << 8 | uint32(Memory[ProgramCounter + 5])
				}	
				setRegister(0x001a, loc)
				Log("jmp " + fmt.Sprintf("0x%08x", loc))
			} else if mode == 0x02 {
				frm := uint32(Memory[ProgramCounter+2])
				loc := getRegister(frm)	
				setRegister(0x001a, loc)
				Log("jmp " + getRegisterName(frm))
			}
			stall(8)
		case 0x04:
			// INT
			var code uint32 = 0
			var next uint32 = 0
			if types.Bits32 == false {
				code = uint32(uint16(Memory[ProgramCounter + 1]) << 8 | uint16(Memory[ProgramCounter + 2]))
				next = ProgramCounter + 3
			} else {
				code = uint32(Memory[ProgramCounter + 1]) << 24 | uint32(Memory[ProgramCounter + 2])	<< 16 | uint32(Memory[ProgramCounter + 3]) << 8 | uint32(Memory[ProgramCounter + 4])
				next = ProgramCounter + 5
			}
			bios.IntHandler(code)
			setRegister(0x001a, next)
			Log("int " + fmt.Sprintf("0x%08x", code))
			stall(34)
		case 0x05:
			// JNZ
			// jnz <mode (01 or 02)> <check register> <loc (register or raw addr)>
			mode := Memory[ProgramCounter+1]
			checkRegister := Memory[ProgramCounter+2]
			var loc uint32 = 0
			var not uint32 = 0

			if mode == 0x01 {	
				if types.Bits32 == false {
					loc = uint32(uint16(Memory[ProgramCounter + 3]) << 8 | uint16(Memory[ProgramCounter + 4]))
					not = ProgramCounter + 5
				} else {
					loc = uint32(Memory[ProgramCounter + 3]) << 24 | uint32(Memory[ProgramCounter + 4])	<< 16 | uint32(Memory[ProgramCounter + 5]) << 8 | uint32(Memory[ProgramCounter + 6])
					not = ProgramCounter + 7
				}	
				Log("jnz " + getRegisterName(uint32(checkRegister)) + ", " + fmt.Sprintf("0x%08x", loc))
			} else if mode == 0x02 {
				frm := uint32(Memory[ProgramCounter+3])
				loc = getRegister(frm)
				not = ProgramCounter + 4
				Log("jnz " + getRegisterName(uint32(checkRegister)) + ", " + getRegisterName(frm))
			}

			if getRegister(uint32(checkRegister)) != 0 {
				setRegister(0x001a, loc)
			} else {
				setRegister(0x001a, not)
			}
			stall(8)
		case 0x06:
			// NOP
			setRegister(0x001a, ProgramCounter+1)
			Log("nop")
			stall(1)
		case 0x07:
			// CMP
			// Syntax: CMP <to> <r1> <r2>
			to := Memory[ProgramCounter+1]
			first := Memory[ProgramCounter+2]
			second := Memory[ProgramCounter+3]
			Log("cmp " + getRegisterName(uint32(to)) + ", " + getRegisterName(first) + ", " + getRegisterName(second))

			if getRegister(uint32(first)) == getRegister(uint32(second)) {
				setRegister(uint32(to), uint32(1))
			} else {
				setRegister(uint32(to), uint32(0))
			}
			setRegister(0x001a, ProgramCounter+4)
			stall(4)
		case 0x08:
			// JZ
			// jz <mode (01 or 02)> <check register> <loc (register or raw addr)>
			mode := Memory[ProgramCounter+1]
			checkRegister := Memory[ProgramCounter+2]
			var loc uint32 = 0
			var not uint32 = 0

			if mode == 0x01 {	
				if types.Bits32 == false {
					loc = uint32(uint16(Mapper(ProgramCounter + 3)) << 8 | uint16(Mapper(ProgramCounter + 4)))
					not = ProgramCounter + 5
				} else {
					loc = uint32(Mapper(ProgramCounter + 3)) << 24 | uint32(Mapper(ProgramCounter + 4))	<< 16 | uint32(Mapper(ProgramCounter + 5)) << 8 | uint32(Mapper(ProgramCounter + 6))
					not = ProgramCounter + 7
				}	
				Log("jz " + getRegisterName(checkRegister) + ", " + fmt.Sprintf("0x%08x", loc))
			} else if mode == 0x02 {
				frm := uint32(Memory[ProgramCounter+3])
				loc = getRegister(frm)
				not = ProgramCounter + 4
				Log("jz " + getRegisterName(checkRegister) + ", " + getRegisterName(frm))
			}

			if getRegister(uint32(checkRegister)) == 0 {
				setRegister(0x001a, loc)
			} else {
				setRegister(0x001a, not)
			}
			stall(8)
		case 0x09:
			// INC
			// inc <register>
			register := uint32(Memory[ProgramCounter+1])
			setRegister(register, getRegister(register)+1)
			setRegister(0x001a, ProgramCounter+2)
			Log("inc " + getRegisterName(register))
			stall(1)
		case 0x0a:
			// DEC
			// dec <register>
			register := uint32(Memory[ProgramCounter+1])
			setRegister(register, getRegister(register)-1)
			setRegister(0x001a, ProgramCounter+2)
			Log("dec " + getRegisterName(register))
			stall(1)
		case 0x0b:
			// PUSH
			// push <mode> <immediate or register>
			mode := Memory[ProgramCounter + 1]
			var value uint32	
			if mode == 0x1 {	
				var next uint32 = 0
				if types.Bits32 == false {
					value = uint32(uint16(Mapper(ProgramCounter + 2)) << 8 | uint16(Mapper(ProgramCounter + 3)))
					next = ProgramCounter + 4
				} else {
					value = uint32(Mapper(ProgramCounter + 2)) << 24 | uint32(Mapper(ProgramCounter + 3))	<< 16 | uint32(Mapper(ProgramCounter + 4)) << 8 | uint32(Mapper(ProgramCounter + 5))
					next = ProgramCounter + 6
				}	
				setRegister(0x001a, next)
				Log("push " + fmt.Sprintf("0x%08x", value))
			} else if mode == 0x2 {
				value = getRegister(uint32(Mapper(ProgramCounter + 2)))
				setRegister(0x001a, ProgramCounter + 3)
				Log("push " + getRegisterName(uint32(Mapper(ProgramCounter + 2))))
			}	
			sp := getRegister(0x0019)
			if types.Bits32 == false {
				sp = video.Clamp(sp - 2, 0, MEMCAP)
			} else {
				sp = video.Clamp(sp - 4, 0, MEMCAP)
			}
			Memory[MapperIndex(sp)] = byte(value & 0xFF)
			Memory[MapperIndex(sp + 1)] = byte(value >> 8)	
			setRegister(0x0019, uint32(sp))	
			stall(2)
		case 0x0c:
			// POP
			// pop <register>	
			register := Mapper(ProgramCounter + 1)
			sp := getRegister(0x0019)
			var value uint32
			if types.Bits32 == false {
				value = uint32(uint16(Mapper(sp)) | uint16(Mapper(sp + 1)) << 8) 
			} else {	
				value = uint32(Mapper(sp)) | uint32(Mapper(sp + 1)) << 8 | uint32(Mapper(sp + 2)) << 16 | uint32(Mapper(sp + 3))
			}
			Log("value: " + fmt.Sprintf("0x%08x", value))
			setRegister(uint32(register), uint32(value))
			if types.Bits32 == false {
				sp = video.Clamp(sp + 2, 0, MEMCAP)
			} else {
				sp = video.Clamp(sp + 4, 0, MEMCAP)
			}
			setRegister(0x0019, uint32(sp))
			setRegister(0x001a, ProgramCounter + 2)
			Log("pop " + getRegisterName(register))
			stall(2)
		case 0x0d:
			// ADD
			// add <register> <register> <register>
			toregister := Memory[ProgramCounter+1]
			regone := Memory[ProgramCounter+2]
			regtwo := Memory[ProgramCounter+3]
			setRegister(uint32(toregister), getRegister(uint32(regone))+getRegister(uint32(regtwo)))
			setRegister(0x001a, ProgramCounter+4)
			Log("add " + getRegisterName(toregister) + ", " + getRegisterName(regone) + ", " + getRegisterName(regtwo))
			stall(7)
		case 0x0e:
			// SUB
			// SUB <register> <register> <register>
			toregister := Memory[ProgramCounter+1]
			regone := Memory[ProgramCounter+2]
			regtwo := Memory[ProgramCounter+3]
			setRegister(uint32(toregister), getRegister(uint32(regone))-getRegister(uint32(regtwo)))
			setRegister(0x001a, ProgramCounter+4)
			Log("sub " + getRegisterName(toregister) + ", " + getRegisterName(regone) + ", " + getRegisterName(regtwo))
			stall(7)
		case 0x0f:
			// MUL
			// mul <register> <register> <register>
			toregister := Memory[ProgramCounter+1]
			regone := Memory[ProgramCounter+2]
			regtwo := Memory[ProgramCounter+3]
			setRegister(uint32(toregister), getRegister(uint32(regone))*getRegister(uint32(regtwo)))
			setRegister(0x001a, ProgramCounter+4)
			Log("mul " + getRegisterName(toregister) + ", " + getRegisterName(regone) + ", " + getRegisterName(regtwo))
			stall(70)
		case 0x10:
			// DIV
			// div <register> <register> <register>
			toregister := Memory[ProgramCounter+1]
			regone := Memory[ProgramCounter+2]
			regtwo := Memory[ProgramCounter+3]
			setRegister(uint32(toregister), getRegister(uint32(regone))/getRegister(uint32(regtwo)))
			setRegister(0x001a, ProgramCounter+4)
			Log("div " + getRegisterName(toregister) + ", " + getRegisterName(regone) + ", " + getRegisterName(regtwo))
			stall(140)
		case 0x11:
			// IGT
			// igt <register> <register> <register>
			toregister := Memory[ProgramCounter+1]
			regone := Memory[ProgramCounter+2]
			regtwo := Memory[ProgramCounter+3]
			if getRegister(uint32(regone)) > getRegister(uint32(regtwo)) {
				setRegister(uint32(toregister), uint32(1))
			} else {
				setRegister(uint32(toregister), uint32(0))
			}
			setRegister(0x001a, ProgramCounter + 4)
			Log("igt " + getRegisterName(toregister) + ", " + getRegisterName(regone) + ", " + getRegisterName(regtwo))
			stall(4)
		case 0x12:
			// ILT
			// ilt <register> <register> <register>
			toregister := Memory[ProgramCounter+1]
			regone := Memory[ProgramCounter+2]
			regtwo := Memory[ProgramCounter+3]
			if getRegister(uint32(regone)) < getRegister(uint32(regtwo)) {
				setRegister(uint32(toregister), uint32(1))
			} else {
				setRegister(uint32(toregister), uint32(0))
			}
			setRegister(0x001a, ProgramCounter + 4)
			Log("ilt " + getRegisterName(toregister) + ", " + getRegisterName(regone) + ", " + getRegisterName(regtwo))
			stall(4)
		case 0x13:
			// AND
			// and <register> <register> <register>
			toregister := Memory[ProgramCounter+1]
			regone := Memory[ProgramCounter+2]
			regtwo := Memory[ProgramCounter+3]
			setRegister(uint32(toregister), getRegister(uint32(regone)) & getRegister(uint32(regtwo)))	
			setRegister(0x001a, ProgramCounter + 4)
			Log("and " + getRegisterName(toregister) + ", " + getRegisterName(regone) + ", " + getRegisterName(regtwo))
			stall(1)
		case 0x14:
			// OR
			// or <register> <register> <register>
			toregister := Memory[ProgramCounter+1]
			regone := Memory[ProgramCounter+2]
			regtwo := Memory[ProgramCounter+3]
			setRegister(uint32(toregister), getRegister(uint32(regone)) | getRegister(uint32(regtwo)))	
			setRegister(0x001a, ProgramCounter + 4)
			Log("or " + getRegisterName(toregister) + ", " + getRegisterName(regone) + ", " + getRegisterName(regtwo))
			stall(1)
		case 0x15:
			// NOR
			// nor <register> <register> <register>
			toregister := Memory[ProgramCounter+1]
			regone := Memory[ProgramCounter+2]
			regtwo := Memory[ProgramCounter+3]
			setRegister(uint32(toregister), ^(getRegister(uint32(regone)) | getRegister(uint32(regtwo))))	
			setRegister(0x001a, ProgramCounter + 4)
			Log("nor " + getRegisterName(toregister) + ", " + getRegisterName(regone) + ", " + getRegisterName(regtwo))
			stall(3)
		case 0x16:
			// NOT
			// not <register> <register>
			toregister := Memory[ProgramCounter+1]
			regone := Memory[ProgramCounter+2]
			setRegister(uint32(uint32(toregister)), ^getRegister(uint32(regone)))	
			setRegister(0x001a, ProgramCounter + 3)
			Log("not " + getRegisterName(toregister) + ", " + getRegisterName(regone))
			stall(1)
		case 0x17:
			// XOR
			// xor <register> <register> <register>
			toregister := Memory[ProgramCounter+1]
			regone := Memory[ProgramCounter+2]
			regtwo := Memory[ProgramCounter+3]
			setRegister(uint32(toregister), getRegister(uint32(regone)) ^ getRegister(uint32(regtwo)))	
			setRegister(0x001a, ProgramCounter + 4)
			Log("xor " + getRegisterName(toregister) + ", " + getRegisterName(regone) + ", " + getRegisterName(regtwo))
			stall(6)
		case 0x18:
			// LOD
			// lod <addr (register)> <destination register>	
			addr := getRegister(uint32(Memory[ProgramCounter+1]))
			toregister := uint32(Memory[ProgramCounter+2])
			setRegister(toregister, uint32(Memory[video.Clamp(addr, 0, MEMCAP)]))
			setRegister(0x001a, ProgramCounter + 3)
			Log("lod " + getRegisterName(uint32(Memory[ProgramCounter + 1])) + ", " + getRegisterName(toregister))
			stall(100)
		case 0x19:
			// STR
			// str <addr (register)> <value (register)>	
			addr := getRegister(uint32(Memory[ProgramCounter+1]))
			value := uint32(Memory[ProgramCounter+2])
			if types.Bits32 == false {
				Memory[video.Clamp(addr, 0, MEMCAP)] = byte(getRegister(value) >> 8)
				Memory[video.Clamp(addr + 1, 0, MEMCAP)] = byte(getRegister(value) & 0xFF)
			} else {
				Memory[video.Clamp(addr, 0, MEMCAP)] = byte(getRegister(value) >> 24)
				Memory[video.Clamp(addr + 1, 0, MEMCAP)] = byte(getRegister(value) >> 16)
				Memory[video.Clamp(addr + 2, 0, MEMCAP)] = byte(getRegister(value) >> 8)
				Memory[video.Clamp(addr + 3, 0, MEMCAP)] = byte(getRegister(value) & 0xFF)
			}	
			setRegister(0x001a, ProgramCounter + 3)
			Log("str " + getRegisterName(uint32(Memory[ProgramCounter + 1])) + ", " + getRegisterName(value))
			stall(100)
		case 0x1a:
			// LODF
			// lodf <addr (register)> <destination register>
			addr := getRegister(uint32(Memory[ProgramCounter+1]))
			toregister := uint32(Memory[ProgramCounter+2])
			if types.Bits32 == false {
				setRegister(toregister, uint32(uint16(Memory[video.Clamp(addr, 0, MEMCAP)]) << 8 | uint16(Memory[video.Clamp(addr + 1, 0, MEMCAP)])))
			} else {
				setRegister(toregister, uint32(Memory[video.Clamp(addr, 0, MEMCAP)]) << 24 | uint32(Memory[video.Clamp(addr + 1, 0, MEMCAP)]) << 16 | uint32(Memory[video.Clamp(addr + 2, 0, MEMCAP)]) << 8 | uint32(Memory[video.Clamp(addr + 3, 0, MEMCAP)]) << 16)
			}
			setRegister(0x001a, ProgramCounter + 3)
			Log("lodw " + getRegisterName(uint32(Memory[ProgramCounter + 1])) + ", " + getRegisterName(toregister))
			stall(100)
		case 0x1b:
			// SET
			// set <00 or 01>
			mode := uint32(Memory[ProgramCounter + 1])
			if mode == 0 {
				types.Bits32 = false
				Log("16 bit mode")
			} else if mode == 1 {
				types.Bits32 = true
				Log("32 bit mode")
			}
			setRegister(0x001a, ProgramCounter + 2)
		default:
			setRegister(0x0001, uint32(op))
			Log("\033[31mIllegal instruction 0x" + fmt.Sprintf("%08x", uint32(op)) + "\033[33m")
			bios.IntHandler(0x7)	
			if Debug == true {
				setRegister(0x001a, ProgramCounter + 1)
			} else {
				return
			}
		}

		if Debug == true {
			bufio.NewReader(os.Stdin).ReadBytes('\n')
		}
	}
}

// Frontend code
var Ready bool = false
func WindowManage(window *app.Window) error {
	var ops op.Ops
	img := image.NewRGBA(image.Rect(0, 0, 320, 200))

	video.InitializePalette()	
	// Init framebuffer
	i := 0
	for y := 0; y < 200; y++ {
		for x := 0; x < 320; x++ {
			img.Set(x, y, video.Palette[uint8(video.MemoryVideo[i])])
			i++
		}
	}

	tex := paint.NewImageOp(img)
	tex.Filter = paint.FilterNearest

	for {
		switch E := window.Event().(type) {
		case app.DestroyEvent:
			os.Exit(0)
		case app.FrameEvent:	
			GTX := app.NewContext(&ops, E)

			paint.Fill(GTX.Ops, color.NRGBA{R: 0, G: 0, B: 0, A: 255})
		
			area := clip.Rect{Max: GTX.Constraints.Max}.Push(GTX.Ops)
			event.Op(GTX.Ops, window)
			for {
				event, ok := GTX.Event(key.Filter{Name: ""})

				if !ok {
					break
				}
				switch event := event.(type) {
				case key.Event:
					if event.State == key.Press {
						char := string(event.Name)

						if event.Name == "Space" {
							char = string(byte(0x20))
						} else if event.Name == "⏎" {
							char = string(byte(0x0a))
						} else if event.Name == "Shift" {
							if keyboard.Shift == false {
								keyboard.Shift = true
							} else {
								keyboard.Shift = false
							}
							continue
						}

						if keyboard.Shift == false {
							char = keyboard.Lower(char)	
						} else {
							char = keyboard.Upper(char)
						}
	
    					setRegister(0x001b, uint32(rune(char[0])))
    					bios.IntHandler(bios.KeyInterruptCode)
					}
				}
			}
			area.Pop()

			i := 0
			for y := 0; y < 200; y++ {
				for x := 0; x < 320; x++ {
					i = video.Clamp(i, 0, 63999)	
					img.Set(x, y, video.Palette[video.MemoryVideo[i]])
					i++
				}
			}

			tex = paint.NewImageOp(img)
			tex.Filter = paint.FilterNearest

			scaleX := float32(GTX.Constraints.Max.X) / float32(320)
			scaleY := float32(GTX.Constraints.Max.Y) / float32(200)

			scale := scaleX
			if scaleY < scaleX {
				scale = scaleY
			}
			defer op.Affine(f32.Affine2D{}.Scale(f32.Pt(0, 0), f32.Pt(scale, scale))).Push(GTX.Ops).Pop()
			tex.Add(GTX.Ops)
			paint.PaintOp{}.Add(GTX.Ops)	
			E.Frame(GTX.Ops)
			Ready = true
			time.Sleep(time.Duration(150) * time.Millisecond)
			window.Invalidate()
		}
	}
	return nil
}

func InitializeWindow() {
	go func() {
		w := new(app.Window)
		w.Option(
			app.Title("Luna L2"),
			app.Size(640, 400),
		)
		if err := WindowManage(w); err != nil {
			fmt.Println("luna-l2: Failed to initialize window.", 255, 0)
			os.Exit(1)
		}
	}()
	app.Main()
}

func main() {
	bios.Registers = &Registers
	bios.Memory = &Memory
	go func() {
		if Ready == false {	
			for {
				if Ready == true {
					break
				} else {
					time.Sleep(500)
				}
			}
		}

		bios.Splash()

		if bios.CheckArgs() == false {
			return
		}
	
		for i := 1; i < len(os.Args); i++ {
			arg := os.Args[i]
			switch arg {
			case "--speed":
				if i + 1 >= len(os.Args) { fmt.Println("Not enough arguments to --speed"); i++; continue }
				speed, err := strconv.ParseInt(os.Args[i + 1], 0, 64)
				if err != nil {
					fmt.Println("Invalid clock speed")
					i++
					continue
				}
				ClockSpeed = int64(speed)
				i++
			case "--log":
				LogOn = true
			case "--debug":
				Debug = true
				LogOn = true
			default:
				Filename = arg
			}
		}

		if Filename == "" {
			bios.WriteLine("No bootable device", 255, 0)
			return
		}	

		LoadSector(0, true)	
		execute()
	}()	
	InitializeWindow()
}
