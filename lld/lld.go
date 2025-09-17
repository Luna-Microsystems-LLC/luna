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
	fmt.Println("lld: " + errors[errno] + " " + args)
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
		if i + 1 < len(data) {
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
			case 0x4C324F: // File magic number, ignore it
				i += 2
			default:
				write(data[i])
			}
		}
	}	
}

func link() {
	process := func(buffer *[]byte, offset int) {
		data := *buffer
		for i := 0; i < len(data); i++ {
			if bytes.HasPrefix(data[i:], []byte("LD_")) {
				j := i + 3
				for j < len(data) && data[j] != 0x00 {
					j++
				}
				name := string(data[i + 3:j])
				data = append(data[:i], data[j + 1:]...)
			
				location := i + offset
				H := byte(location >> 8)
				L := byte(location & 0xFF)

				_, ok := checkBinding(name)
				if ok != false {
					error(2, "`" + name + "'")
				}
				bindings = append(bindings, binding{Name: name, Location: []byte{H, L}})

				ref := append([]byte("LR_" + name), 0x00)	
				DataBuffer = bytes.ReplaceAll(DataBuffer, ref, []byte{H, L})
				TextBuffer = bytes.ReplaceAll(TextBuffer, ref, []byte{H, L})
				ExtendedDataBuffer = bytes.ReplaceAll(ExtendedDataBuffer, ref, []byte{H, L})
			}
		}
		*buffer = data
	}
	process(&DataBuffer, 3 + 2)
	process(&TextBuffer, 3 + 2 + len(DataBuffer))
	process(&ExtendedDataBuffer, 3 + 2 + len(DataBuffer) + len(TextBuffer))
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
			fmt.Println("@(#)PROGRAM:lld PROJECT:ld-1.0")
			fmt.Println("BUILD 19:33:36 Sep 16 2025")
			fmt.Println("configured to support archs: luna-l2")	
			os.Exit(0)
		case "-o":
			output_filename = os.Args[i + 1]
			i++	
		default:
			input_files = append(input_files, arg)
		}
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

	buffer := append([]byte{}, []byte{0x4c, 0x32, 0x45}...)
	buffer = append(buffer, startloc...) 
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
