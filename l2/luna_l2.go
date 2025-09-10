package main

import (
	"bufio"
	"bytes"
	_ "embed"
	"fmt"
	"io"
	"os"
	"runtime"
	"time"
	"github.com/faiface/beep"
	"github.com/faiface/beep/mp3"
	"github.com/faiface/beep/speaker"
	"image"
	"image/color"
	"gioui.org/app"
    "gioui.org/op"
    "gioui.org/op/paint"	
	"gioui.org/f32"
	"luna_l2/font"
)

//go:embed sounds/crash.mp3
var crashSoundData []byte

type Register struct {
	Address uint16
	Name string
	Value uint16
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
}

var Memory [65535]byte
var MemoryVideo [64000]byte

var cursorX, cursorY int

func PushChar(x, y int, ch rune, fg, bg byte) {

	glyph := font.Font[0x00]
	if ch < 0 || int(ch) >= len(font.Font) {
		glyph = font.Font[0x20]
	} else {
		glyph = font.Font[ch]
	}


	for row := 0; row < 8; row++ {
		line := glyph[row]
		for col := 0; col < 8; col++ {
			mask := byte(1 << (7 - col))
			var color byte
			if line&mask != 0 {
				color = fg
			} else {
				color = bg
			}
			px := (y+row)*320 + (x+col)
			if px >= 0 && px < len(MemoryVideo) {
				MemoryVideo[px] = color
			}
		}
	}
}

func PrintChar(ch rune, fg, bg byte) {
	x := cursorX * 8
	y := cursorY * 8
	PushChar(x, y, ch, fg, bg)

	cursorX++
	if cursorX >= 320/8 {
		cursorX = 0
		cursorY++
	}
	if cursorY >= 200/8 {
		cursorY = 0
	}
}

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
		char := getRegister(0x0001)

		PrintChar(rune(char), byte(255), byte(000))	
	} else if code == 0x02 {
		timeToSleep := getRegister(0x0001)

		time.Sleep(time.Second * time.Duration(timeToSleep))
	} else if code == 0x03 {
		reader := bufio.NewScanner(os.Stdin)
		if reader.Scan() {
			line := reader.Text()
			line = line
			// copyMemory()
		}
	}
}

func stall(cycles int) {
	clockHz := 1158000
	cycleTime := int(time.Second) / clockHz
	time.Sleep(time.Duration(cycleTime * cycles))
}

func execute() {
	for {
		ProgramCounter := getRegister(0x001a)
		op := Memory[ProgramCounter]

		if ProgramCounter == 0x0000 {
			codesect := uint16(Memory[ProgramCounter]) << 8 | uint16(Memory[ProgramCounter + 1])
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
				setRegister(0x001a, getRegister(frm))
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
				print(string(rune(loc)))
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
			// push <register>
			register := Memory[ProgramCounter+1]
			value := getRegister(uint16(register))
			sp := getRegister(0x0019)
			sp += 2
			Memory[sp] = byte(value & 0xFF)
			Memory[sp+1] = byte(value >> 8)
			setRegister(0x0019, uint16(sp))
			setRegister(0x001a, ProgramCounter+2)
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
			toregister := Memory[ProgramCounter + 1]
			regone := Memory[ProgramCounter + 2]
			regtwo := Memory[ProgramCounter + 3]
			setRegister(uint16(toregister), getRegister(uint16(regone)) + getRegister(uint16(regtwo)))
			setRegister(0x001a, ProgramCounter + 4)
			stall(7)
		case 0x0e:
			// SUB
			// SUB <register> <register> <register>
			toregister := Memory[ProgramCounter + 1]
			regone := Memory[ProgramCounter + 2]
			regtwo := Memory[ProgramCounter + 3]
			setRegister(uint16(toregister), getRegister(uint16(regone)) - getRegister(uint16(regtwo)))
			setRegister(0x001a, ProgramCounter + 4)
			stall(7)
		case 0x0f:
			// MUL
			// mul <register> <register> <register>
			toregister := Memory[ProgramCounter + 1]
			regone := Memory[ProgramCounter + 2]
			regtwo := Memory[ProgramCounter + 3]
			setRegister(uint16(toregister), getRegister(uint16(regone)) * getRegister(uint16(regtwo)))
			setRegister(0x001a, ProgramCounter + 4)
			stall(70)
		case 0x10:
			// DIV
			// div <register> <register> <register>
			toregister := Memory[ProgramCounter + 1]
			regone := Memory[ProgramCounter + 2]
			regtwo := Memory[ProgramCounter + 3]
			setRegister(uint16(toregister), getRegister(uint16(regone)) / getRegister(uint16(regtwo)))
			setRegister(0x001a, ProgramCounter + 4)
			stall(140)
		default:
			fmt.Printf("FATAL: Illegal instruction 0x%02x at instruction 0x%02x\n", op, ProgramCounter)
			playSound("crash")
			return
		}
	}
}


func WindowManage(window *app.Window) error {
	var ops op.Ops
	img := image.NewRGBA(image.Rect(0, 0, 320, 200))

	// Pallete
	var Palette [256]color.NRGBA
	for i := 0; i < 256; i++ {
		Palette[i] = color.NRGBA{uint8(i), uint8(i), uint8(i), 255}
	}
	// Init framebuffer
	i := 0
	for y := 0; y < 200; y++ {
		for x := 0; x < 320; x++ {
			img.Set(x, y, Palette[uint8(MemoryVideo[i])])
			i++
		}
	}
	
	tex := paint.NewImageOp(img)

	for {
		switch E := window.Event().(type) {
		case app.DestroyEvent:
			return E.Err
		case app.FrameEvent:
			GTX := app.NewContext(&ops, E)

			scaleX := float32(GTX.Constraints.Max.X) / float32(320)
			scaleY := float32(GTX.Constraints.Max.Y) / float32(200)

			scale := scaleX
			if scaleY < scaleX {
				scale = scaleY
			}
			trans := op.Affine(f32.Affine2D{}.Scale(f32.Pt(0, 0), f32.Pt(scale, scale)))
			defer trans.Push(GTX.Ops).Pop()

			tex.Add(GTX.Ops)
			paint.PaintOp{}.Add(GTX.Ops)
			E.Frame(GTX.Ops)
		}
	}
	return nil
}

func InitializeWindow() {
	go func() {
		w := new(app.Window)
		w.Option (
			app.Title("Luna L2"),
			app.Size(320, 200),
		)
		if err := WindowManage(w); err != nil {
			fmt.Printf("FATAL: Failed to initialize window.")
			os.Exit(1)
		}
	}()
	app.Main()
}

func main() {	
	go func() {
		fmt.Println("Luna L2")
		fmt.Println("BIOS: Integrated BIOS")
		fmt.Println("Host: " + runtime.GOOS)
		fmt.Println("Host CPU: " + runtime.GOARCH)
		fmt.Println("Copyright (c) 2025 Luna Microsystems LLC\n")

		if len(os.Args) < 2 {
			fmt.Println("FATAL: Disk image not found.")
			playSound("crash")
			for {
				time.Sleep(time.Second)
			}
		}

		filename := os.Args[1]

		data, err := os.ReadFile(filename)
		if err != nil {
			fmt.Println("FATAL: Failed to read disk image '" + filename + "'")
			playSound("crash")
			for {
				time.Sleep(time.Second)
			}
		}

		if len(data) > 65535 {
			fmt.Println("FATAL: Disk image too large (max 64KiB)")
			playSound("crash")
			for {
				time.Sleep(time.Second)
			}
		}
		copy(Memory[:], data)
		setRegister(0x0019, uint16(len(data)))	
		execute()
		for {
			time.Sleep(time.Second)
		}
	}()
	InitializeWindow()
}
