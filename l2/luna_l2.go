package main

import (	
	"image"
	"image/color"	
	"os"
	"runtime"
	"time"
	"fmt"	

	"luna_l2/bios"	
	"luna_l2/sound"
	"luna_l2/video"
	"luna_l2/keyboard"

	"gioui.org/app"	
	"gioui.org/f32"
	"gioui.org/op"
	"gioui.org/op/paint"
	"gioui.org/op/clip"
	"gioui.org/io/key"
	"gioui.org/io/event"	
)

type Register struct {
	Address uint16
	Name    string
	Value   uint16
}

var Registers = []Register{
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

var Memory [65535]byte
var Ready bool = false

func setRegister(address uint16, value uint16) {
	for i := range Registers {
		if Registers[i].Address == address {
			Registers[i].Value = value
			return
		}
	}
}

func getRegister(address uint16) uint16 {
	for _, register := range Registers {
		if register.Address == address {
			return register.Value
		}
	}
	return 0x0000
}

func stall(cycles int) {
	clockHz := 1158000
	cycleTime := int(time.Second) / clockHz
	time.Sleep(time.Duration(cycleTime * cycles))
}

func intHandler(code uint16) {
	if code == 0x01 {
		// BIOS print to screen
		// start address in R1
		// Foreground in R2
		// Background in R3
		char := getRegister(0x0001)
		bios.WriteChar(string(rune(char)), uint8(getRegister(0x0002)), uint8(getRegister(0x0003)))
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

		video.MemoryVideo[address] = byte(uint16(word) << 8)
		video.MemoryVideo[address + 1] = byte(uint16(word) & 0xFF)
	} else if code == 0x4 {
		// BIOS configure input mode
		// Mode 1: no type output
		// Mode 2: type output
		// In R1
		if getRegister(0x0001) == uint16(1) {
			bios.TypeOut = true
		} else {
			bios.TypeOut = false
		}
	} else if code == 0x5 {
		if bios.TypeOut == true {
			bios.WriteChar(string(rune(getRegister(0x001b))), uint8(255), uint8(0))
		}
	}
}

func execute() {
	for {
		ProgramCounter := getRegister(0x001a)
		op := Memory[ProgramCounter]

		if ProgramCounter == 0x0000 {
			codesect := uint16(Memory[ProgramCounter])<<8 | uint16(Memory[ProgramCounter+1])
			setRegister(0x001a, codesect)
			continue
		}

		switch op {
		case 0x00:
			return
		case 0x01:
			// MOV
			mode := Memory[ProgramCounter+1]
			dst := Memory[ProgramCounter+2]

			if mode == 0x01 {
				imm := uint16(Memory[ProgramCounter+3])<<8 | uint16(Memory[ProgramCounter+4])
				setRegister(uint16(dst), imm)
				setRegister(0x001a, ProgramCounter+5)
			} else if mode == 0x02 {
				frm := uint16(Memory[ProgramCounter+3])
				setRegister(uint16(dst), uint16(getRegister(frm)))
				setRegister(0x001a, ProgramCounter+4)
			}

			stall(4)
		case 0x02:
			// HLT
			for {
				time.Sleep(time.Second)
			}
			setRegister(0x001a, ProgramCounter+1)
		case 0x03:
			// JMP	
			mode := Memory[ProgramCounter+1]

			if mode == 0x01 {	
				loc := uint16(Memory[ProgramCounter+2])<<8 | uint16(Memory[ProgramCounter+3])
				setRegister(0x001a, loc)
			} else if mode == 0x02 {
				frm := uint16(Memory[ProgramCounter+2])
				loc := getRegister(frm)	
				setRegister(0x001a, loc)
			}
			stall(8)
		case 0x04:
			// INT
			code := uint16(Memory[ProgramCounter+1])<<8 | uint16(Memory[ProgramCounter+2])
			intHandler(code)
			setRegister(0x001a, ProgramCounter+3)
			stall(34)
		case 0x05:
			// JNZ
			// jnz <mode (01 or 02)> <check register> <loc (register or raw addr)>
			mode := Memory[ProgramCounter+1]
			checkRegister := Memory[ProgramCounter+2]
			var loc uint16 = 0
			var not uint16 = 0

			if mode == 0x01 {
				loc = uint16(Memory[ProgramCounter+3])<<8 | uint16(Memory[ProgramCounter+4])	
				not = ProgramCounter + 5
			} else if mode == 0x02 {
				frm := uint16(Memory[ProgramCounter+3])
				loc = getRegister(frm)
				not = ProgramCounter + 4
			}

			if getRegister(uint16(checkRegister)) != 0 {
				setRegister(0x001a, loc)
			} else {
				setRegister(0x001a, not)
			}
			stall(8)
		case 0x06:
			// NOP
			setRegister(0x001a, ProgramCounter+1)
			stall(1)
		case 0x07:
			// CMP
			// Syntax: CMP <to> <r1> <r2>
			to := Memory[ProgramCounter+1]
			first := Memory[ProgramCounter+2]
			second := Memory[ProgramCounter+3]

			if getRegister(uint16(first)) == getRegister(uint16(second)) {
				setRegister(uint16(to), uint16(1))
			} else {
				setRegister(uint16(to), uint16(0))
			}
			setRegister(0x001a, ProgramCounter+4)
			stall(4)
		case 0x08:
			// JZ
			// jz <mode (01 or 02)> <check register> <loc (register or raw addr)>
			mode := Memory[ProgramCounter+1]
			checkRegister := Memory[ProgramCounter+2]
			var loc uint16 = 0
			var not uint16 = 0

			if mode == 0x01 {
				loc = uint16(Memory[ProgramCounter+3])<<8 | uint16(Memory[ProgramCounter+4])
				not = ProgramCounter + 5
			} else if mode == 0x02 {
				frm := uint16(Memory[ProgramCounter+3])
				loc = getRegister(frm)
				not = ProgramCounter + 4
			}

			if getRegister(uint16(checkRegister)) == 0 {
				setRegister(0x001a, loc)
			} else {
				setRegister(0x001a, not)
			}
			stall(8)
		case 0x09:
			// INC
			// inc <register>
			register := uint16(Memory[ProgramCounter+1])
			setRegister(register, getRegister(register)+1)
			setRegister(0x001a, ProgramCounter+2)
			stall(1)
		case 0x0a:
			// DEC
			// dec <register>
			register := uint16(Memory[ProgramCounter+1])
			setRegister(register, getRegister(register)-1)
			setRegister(0x001a, ProgramCounter+2)
			stall(1)
		case 0x0b:
			// PUSH
			// push <mode> <immediate or register>
			mode := Memory[ProgramCounter + 1]
			var value uint16	
			if mode == 0x1 {
				value = uint16(Memory[ProgramCounter + 2]) << 8 | uint16(Memory[ProgramCounter + 3])
				setRegister(0x001a, ProgramCounter + 4)
			} else if mode == 0x2 {
				value = getRegister(uint16(Memory[ProgramCounter + 2]))
				setRegister(0x001a, ProgramCounter + 3)
			}	
			sp := getRegister(0x0019)
			sp += 2
			Memory[sp] = byte(value & 0xFF)
			Memory[sp+1] = byte(value >> 8)	
			setRegister(0x0019, uint16(sp))	
			stall(2)
		case 0x0c:
			// POP
			// pop <register>	
			register := Memory[ProgramCounter+1]
			sp := getRegister(0x0019)
			value := uint16(Memory[sp]) | uint16(Memory[sp+1])<<8
			setRegister(uint16(register), uint16(value))
			sp -= 2	
			setRegister(0x0019, uint16(sp))
			setRegister(0x001a, ProgramCounter+2)	
			stall(2)
		case 0x0d:
			// ADD
			// add <register> <register> <register>
			toregister := Memory[ProgramCounter+1]
			regone := Memory[ProgramCounter+2]
			regtwo := Memory[ProgramCounter+3]
			setRegister(uint16(toregister), getRegister(uint16(regone))+getRegister(uint16(regtwo)))
			setRegister(0x001a, ProgramCounter+4)
			stall(7)
		case 0x0e:
			// SUB
			// SUB <register> <register> <register>
			toregister := Memory[ProgramCounter+1]
			regone := Memory[ProgramCounter+2]
			regtwo := Memory[ProgramCounter+3]
			setRegister(uint16(toregister), getRegister(uint16(regone))-getRegister(uint16(regtwo)))
			setRegister(0x001a, ProgramCounter+4)
			stall(7)
		case 0x0f:
			// MUL
			// mul <register> <register> <register>
			toregister := Memory[ProgramCounter+1]
			regone := Memory[ProgramCounter+2]
			regtwo := Memory[ProgramCounter+3]
			setRegister(uint16(toregister), getRegister(uint16(regone))*getRegister(uint16(regtwo)))
			setRegister(0x001a, ProgramCounter+4)
			stall(70)
		case 0x10:
			// DIV
			// div <register> <register> <register>
			toregister := Memory[ProgramCounter+1]
			regone := Memory[ProgramCounter+2]
			regtwo := Memory[ProgramCounter+3]
			setRegister(uint16(toregister), getRegister(uint16(regone))/getRegister(uint16(regtwo)))
			setRegister(0x001a, ProgramCounter+4)
			stall(140)
		case 0x11:
			// IGT
			// igt <register> <register> <register>
			toregister := Memory[ProgramCounter+1]
			regone := Memory[ProgramCounter+2]
			regtwo := Memory[ProgramCounter+3]
			if getRegister(uint16(regone)) > getRegister(uint16(regtwo)) {
				setRegister(uint16(toregister), uint16(1))
			} else {
				setRegister(uint16(toregister), uint16(0))
			}
			setRegister(0x001a, ProgramCounter + 4)
			stall(4)
		case 0x12:
			// ILT
			// ilt <register> <register> <register>
			toregister := Memory[ProgramCounter+1]
			regone := Memory[ProgramCounter+2]
			regtwo := Memory[ProgramCounter+3]
			if getRegister(uint16(regone)) < getRegister(uint16(regtwo)) {
				setRegister(uint16(toregister), uint16(1))
			} else {
				setRegister(uint16(toregister), uint16(0))
			}
			setRegister(0x001a, ProgramCounter + 4)
			stall(4)
		case 0x13:
			// AND
			// and <register> <register> <register>
			toregister := Memory[ProgramCounter+1]
			regone := Memory[ProgramCounter+2]
			regtwo := Memory[ProgramCounter+3]
			if getRegister(uint16(regone)) != 0 && getRegister(uint16(regtwo)) != 0 {
				setRegister(uint16(toregister), uint16(1))
			} else {
				setRegister(uint16(toregister), uint16(0))
			}
			setRegister(0x001a, ProgramCounter + 4)
			stall(1)
		case 0x14:
			// OR
			// or <register> <register> <register>
			toregister := Memory[ProgramCounter+1]
			regone := Memory[ProgramCounter+2]
			regtwo := Memory[ProgramCounter+3]
			if getRegister(uint16(regone)) != 0 || getRegister(uint16(regtwo)) != 0 {
				setRegister(uint16(toregister), uint16(1))
			} else {
				setRegister(uint16(toregister), uint16(0))
			}
			setRegister(0x001a, ProgramCounter + 4)
			stall(1)
		case 0x15:
			// NOR
			// nor <register> <register> <register>
			toregister := Memory[ProgramCounter+1]
			regone := Memory[ProgramCounter+2]
			regtwo := Memory[ProgramCounter+3]
			if getRegister(uint16(regone)) == 0 && getRegister(uint16(regtwo)) == 0 {
				setRegister(uint16(toregister), uint16(1))
			} else {
				setRegister(uint16(toregister), uint16(0))
			}
			setRegister(0x001a, ProgramCounter + 4)
			stall(3)
		case 0x16:
			// NOT
			// not <register> <register>
			toregister := Memory[ProgramCounter+1]
			regone := Memory[ProgramCounter+2]
			if getRegister(uint16(regone)) != 0 {
				setRegister(uint16(toregister), uint16(1))
			} else {
				setRegister(uint16(toregister), uint16(0))
			}
			setRegister(0x001a, ProgramCounter + 3)
			stall(1)
		case 0x17:
			// XOR
			// xor <register> <register> <register>
			toregister := Memory[ProgramCounter+1]
			regone := Memory[ProgramCounter+2]
			regtwo := Memory[ProgramCounter+3]
			if (getRegister(uint16(regone)) == 0 && getRegister(uint16(regtwo)) != 0) || (getRegister(uint16(regone)) != 0 && getRegister(uint16(regtwo)) == 0) {
				setRegister(uint16(toregister), uint16(1))
			} else {
				setRegister(uint16(toregister), uint16(0))
			}
			setRegister(0x001a, ProgramCounter + 4)
			stall(6)
		case 0x18:
			// LOD
			// lod <addr (register)> <destination register>	
			addr := getRegister(uint16(Memory[ProgramCounter+1]))
			toregister := uint16(Memory[ProgramCounter+2])
			setRegister(toregister, uint16(Memory[addr]))
			setRegister(0x001a, ProgramCounter + 3)
			stall(100)
		case 0x19:
			// STR
			// str <addr (register)> <value (register)>	
			addr := getRegister(uint16(Memory[ProgramCounter+1]))
			value := uint16(Memory[ProgramCounter+2])
			Memory[addr] = byte(getRegister(value))	
			setRegister(0x001a, ProgramCounter + 3)
			stall(100)
		default:
			bios.WriteString("FATAL: Illegal instruction " + fmt.Sprintf("%X", uint16(op)) + " at location " + fmt.Sprintf("%X", getRegister(0x001a)), 255, 0)
			sound.PlaySound("crash")
			return
		}
	}
}

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
			Ready = true
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
	
    					setRegister(0x001b, uint16(rune(char[0])))
    					intHandler(0x05)
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
			app.Size(320, 200),
		)
		if err := WindowManage(w); err != nil {
			fmt.Println("FATAL: Failed to initialize window.", 255, 0)
			os.Exit(1)
		}
	}()
	app.Main()
}

func main() {
	fmt.Printf("\033]0;Luna L2 (Host Console)\007")
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

		bios.WriteLine("Luna L2", 255, 0)
		bios.WriteLine("BIOS: Integrated BIOS", 255, 0)
		bios.WriteLine("Host: " + runtime.GOOS, 255, 0)
		bios.WriteLine("Host CPU: " + runtime.GOARCH, 255, 0)
		bios.WriteLine("Copyright (c) 2025 Luna Microsystems LLC\n", 255, 0)

		if len(os.Args) < 2 {
			bios.WriteLine("FATAL: Disk image not found.", 255, 0)
			sound.PlaySound("crash")
			return
		} 

		filename := os.Args[1]

		data, err := os.ReadFile(filename)
		if err != nil {
			bios.WriteLine("FATAL: Failed to read disk image '"+filename+"'", 255, 0)
			sound.PlaySound("crash")
			return
		}

		if len(data) > 65535 {
			bios.WriteLine("FATAL: Disk image too large (max 64KiB)", 255, 0)
			sound.PlaySound("crash")
			return
		}
		copy(Memory[:], data)
		setRegister(0x0019, uint16(len(data)))
		execute()
	}()	
	InitializeWindow()
}
