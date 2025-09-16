package main

import (
	"fmt"
	"os"
	"strings"
	"strconv"
)

var section string = "text"
var input_files []string
var DataBuffer string = ""
var TextBuffer string = ""
var ExtendedDataBuffer string = ""
var current_filename string = ""

func write(text string) {
	if section == "data" {
		DataBuffer = DataBuffer + text
	} else if section == "edata" {
		ExtendedDataBuffer = ExtendedDataBuffer + text
	} else if section == "text" {
		TextBuffer = TextBuffer + text
	}
}

func isRegister(word string) int {
	switch word {
	case "r0":
		return 0x00
	case "r1":
		return 0x01
	case "r2":
		return 0x02
	case "r3":
		return 0x03
	case "r4":
		return 0x04
	case "r5":
		return 0x05
	case "r6":
		return 0x06
	case "r7":
		return 0x07
	case "r8":
		return 0x08
	case "r9":
		return 0x09
	case "r10":
		return 0x0a
	case "r11":
		return 0x0b
	case "r12":
		return 0x0c
	case "t1":
		return 0x0d
	case "t2":
		return 0x0e
	case "t3":
		return 0x0f
	case "t4":
		return 0x10
	case "t5":
		return 0x11
	case "t6":
		return 0x12
	case "t7":
		return 0x13
	case "t8":
		return 0x14
	case "t9":
		return 0x15
	case "t10":
		return 0x16
	case "t11":
		return 0x17
	case "t12":
		return 0x18
	case "sp":
		return 0x19
	case "pc":
		return 0x1a
	case "re1":
		return 0x1b
	case "re2":
		return 0x1c
	case "re3":
		return 0x1d
	default:
		return -1
	}
}

var errors = []string {
	"no input files",
	"no such file or directory",
	"invalid register name",
	"invalid operand to instruction",
	"invalid instruction mnemonic",
	"immediate value too large",
	"missing terminating '\"' character",
}
var nocont bool = false
var Errors int
var Warnings int
func error(errno int, args string) {
	label := ""

	if current_filename != "" {
		label = current_filename
	} else {
		label = "lcc"
	}

	fmt.Println("\033[1;39m" + label + ": \033[1;31merror: \033[1;39m" + errors[errno] + " " + args + "\033[0m")
	Errors++
	nocont = true
}

func parse(text string) string {
	// Check for number
	if _, err := strconv.Atoi(text); err == nil {
		num, _ := strconv.Atoi(text)

		high := byte(num >> 8)
		low := byte(num & 0xFF)
		return string([]byte{high, low})
	}
	if isRegister(text) != -1 {
		return string(byte(isRegister(text)))
	}
	if string(text[0]) == "\"" {
		if string(text[len(text) - 1]) != "\"" {
			error(6, "")
		}
		text = strings.Trim(text, "\"")
		if len(text) > 2 {
			error(5, "'" + text + "'")
		}
		return text
	}
	return "LINKER_SYMBOL_" + text	
}

func assemble(text string) {
	words := strings.Fields(text)

	for i := 0; i < len(words); i++ {
		words[i] = strings.TrimSuffix(words[i], ",")
	}

	for i := 0; i < len(words); i++ {
		if strings.HasSuffix(words[i], ":") {	
			end := len(words)
			for j := i + 1; j < len(words); j++ {
				if strings.HasSuffix(words[j], ":") {
					end = j
					break
				}
			}

			tocompile := words[i + 1 : end]
			if len(tocompile) > 0 {	
				assemble(strings.Join(tocompile, " "))
			}

			i = end - 1
			continue
		}
		
		words[i] = strings.ToLower(words[i])
		switch words[i] {
		case ".data":
			section = "data"
		case ".text":
			section = "text"
		case ".edata":
			section = "edata"
		case "mov":
			write(string(byte(0x01)))
	
			if isRegister(words[i + 2]) == -1 {	
				write(string(byte(0x01)))
			} else {
				write(string(byte(0x02)))
			}

			register := isRegister(words[i + 1])
			if register == -1 {
				error(2, "'" + words[i + 1] + "'")
				continue
			}
			write(string(byte(register)))

			value := parse(words[i + 2])
			write(value)
			i = i + 2
		case "hlt":
			write(string(byte(0x02)))
		case "jmp":
			write(string(byte(0x03)))
			if isRegister(words[i + 1]) == -1 {	
				write(string(byte(0x01)))
			} else {
				write(string(byte(0x02)))
			}
			write(parse(words[i + 1]))
			i = i + 1 
		case "int":
			write(string(byte(0x04)))
			value := parse(words[i + 1])
			if len(value) > 2 {
				error(3, "'" + value + "'")
				continue
			}
			write(value)
			i = i + 1
		default:
			error(4, "'" + words[i] + "'")
		}
	}
}

func main() {
	if len(os.Args) < 2 {
		error(0, "")	
		os.Exit(1)
	}

	var output_filename string = ""
	var nolink bool = false

	for i := 1; i < len(os.Args); i++ {
		arg := os.Args[i]

		switch arg {
		case "-v":
			fmt.Println("Luna Compiler Collection v2.0")
			fmt.Println("Target: luna-l2")
			os.Exit(0)
		case "-o":
			output_filename = os.Args[i + 1]
			i++
		case "-c":
			nolink = true
		default:
			input_files = append(input_files, arg)
		}
	}

	if output_filename == "" {
		output_filename = "a.o"
	}

	for _, file := range input_files {
		data, err := os.ReadFile(file)
		if err != nil {
			error(1, "'" + file + "'")
			os.Exit(1)
		}
		current_filename = file
		assemble(string(data))
	}

	var error_str string = ""
	if Warnings > 0 {
		error_str = error_str + fmt.Sprintf("%d", Warnings) + " warning"
		if Warnings > 1 {
			error_str = error_str + "s"
		}
		if Errors > 0 {
			error_str = error_str + " and "
		} else {
			error_str = error_str + " generated."
		}
	}
	if Errors > 0 {
		error_str = error_str + fmt.Sprintf("%d", Errors) + " error"
		if Errors > 1 {
			error_str = error_str + "s"
		}
		error_str = error_str + " generated."
	}

	if Errors > 0 || Warnings > 0 {
		fmt.Println(error_str)
	}

	if nocont == true {
		os.Exit(1)
	}

	buffer := "__data" + DataBuffer + string(byte(00)) + "__text" + TextBuffer + string(byte(00)) + "__edata" + ExtendedDataBuffer
	os.WriteFile(output_filename, []byte(buffer), 0644)

	if nolink == false {

	}
}
