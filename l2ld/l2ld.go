package main

import (
	"fmt"
	"os"
	"bytes"
	"strings"
)

type binding struct {
	Name string
	Location []byte
}

var DataBuffer []byte
var TextBuffer []byte
var ExtendedDataBuffer []byte
var section string = "text"

var bindings = []binding {}

var errors = []string {
	"no object files specified",
	"file cannot be open()ed, errno=2",
	"multiple definitions of",
	"Undefined symbol for architecture luna-l2:",	
}
func error(errno int, args string) {
	fmt.Fprintln(os.Stderr, "l2ld: " + errors[errno] + " " + args)
	os.Exit(1)
}

func write(content byte) {
	switch section {
	case "data":
		DataBuffer = append(DataBuffer, content)
	case "text":
		TextBuffer = append(TextBuffer, content)
	case "edata":
		ExtendedDataBuffer = append(ExtendedDataBuffer, content)
	}	
}

func checkBinding(name string) ([]byte, bool) {
	for i := range bindings {
		if bindings[i].Name == name {
			return bindings[i].Location, true
		}
	}
	return nil, false
}

func separate(data []byte) {	
	for i := 0; i < len(data); i++ {
		if i + 2 < len(data) {
			bytes := uint32(data[i]) << 16 | uint32(data[i + 1]) << 8 | uint32(data[i + 2])
			switch bytes {
			case 0xC2807D:
				section = "data"
				i += 2
			case 0xC2807E:
				section = "text"
				i += 2
			case 0xC2807F:
				section = "edata"
				i += 2	
			default:
				write(data[i])
			}
		} else {
			write(data[i])
		}
	}	
}

func link() {	
	collect := func(buffer *[]byte, offset int) {	
		data := *buffer
		for i := 0; i < len(data); i++ {
			if bytes.HasPrefix(data[i:], []byte("LD16_")) || bytes.HasPrefix(data[i:], []byte("LD32_")) {
				var Bits32 bool = false
				if bytes.HasPrefix(data[i:], []byte("LD32_")) {
					Bits32 = true
				}

				j := i + 5
				for j < len(data) && data[j] != 0x00 {
					j++
				}
				name := string(data[i + 5:j])

				old := len(data)
				data = append(data[:i], data[j + 1:]...)
				RSF := old - len(data)

				location := (i + offset)

				_, ok := checkBinding(name)
				if ok != false {
					error(2, "`" + name + "'")
				}

				if Bits32 == false {
					H := byte(location >> 8)	
					L := byte(location & 0xFF)
					AH := byte((location - (j - i + 5) - RSF) >> 8)
					AL := byte((location - (j - i + 5) - RSF) & 0xFF)
					bindings = append(bindings, binding{Name: name, Location: []byte{H, L}})
					data = append(bytes.ReplaceAll(data[:i], append([]byte("LR_" + name), 0x00), []byte{AH, AL}), data[i:]...)
				} else {
					HH := byte(location >> 24)
					HL := byte(location >> 16)
					LH := byte(location >> 8)
					LL := byte(location & 0xFF)

					AHH := byte((location - (j - i + 5) - RSF) >> 24)
					AHL := byte((location - (j - i + 5) - RSF) >> 16)
					ALH := byte((location - (j - i + 5) - RSF) >> 8)
					ALL := byte((location - (j - i + 5) - RSF) & 0xFF)
					bindings = append(bindings, binding{Name: name, Location: []byte{HH, HL, LH, LL}})
					data = append(bytes.ReplaceAll(data[:i], append([]byte("LR_" + name), 0x00), []byte{AHH, AHL, ALH, ALL}), data[i:]...)
				}
			} 
		}
		*buffer = data
	}
	collect(&DataBuffer, 2)
	collect(&TextBuffer, 2 + len(DataBuffer))
	collect(&ExtendedDataBuffer, 2 + len(DataBuffer) + len(TextBuffer))

	for _, b := range bindings {
		ref := append([]byte("LR_" + b.Name), 0x00)
		DataBuffer = bytes.ReplaceAll(DataBuffer, ref, b.Location)
		TextBuffer = bytes.ReplaceAll(TextBuffer, ref, b.Location)
		ExtendedDataBuffer = bytes.ReplaceAll(ExtendedDataBuffer, ref, b.Location)
	}
}

func main() {
	if len(os.Args) < 2 {
		error(0, "")
	}

	var input_files []string
	var output_filename string = ""

	for i := 1; i < len(os.Args); i++ {
		arg := os.Args[i]

		switch arg {
		case "-v":
			fmt.Println("@(#)PROGRAM:l2ld PROJECT:l2ld-1.0")
			fmt.Println("BUILD 10:29:40 Sep 18 2025")
			fmt.Println("configured to support archs: luna-l2")	
			os.Exit(0)
		case "-o":
			output_filename = os.Args[i + 1]
			i++
		default:
			input_files = append(input_files, arg)
		}
	}

	if len(input_files) < 1 {
		error(0, "")
	}
	if output_filename == "" {
		output_filename = "a.bin"
	}

	for _, file := range input_files {
		data, err := os.ReadFile(file)
		if err != nil {
			error(1, "path=" + file)
		}
		separate(data)		
	}

	link()
	startloc, found := checkBinding("_start")
	if found == false {
		error(3, "\n  \"_start\", referenced from\n    <initial-undefines>")	
	}
	
	buffer := append([]byte{}, startloc...) 
	buffer = append(buffer, DataBuffer...)
	buffer = append(buffer, TextBuffer...)
	buffer = append(buffer, ExtendedDataBuffer...)

	location := bytes.Index(buffer, []byte("LR_"))
	if location != -1 {
		name := ""
		for i := location; i < len(buffer); i++ {
			if buffer[i] != 0x00 {
				name = name + string(buffer[i])
			} else {
				break
			}
		}
		name = strings.TrimPrefix(name, "LR_")
		error(3, "\n  \"" + name + "\", referenced from\n    <initial-undefines>")
	}
	os.WriteFile(output_filename, []byte(buffer), 0644)
}
