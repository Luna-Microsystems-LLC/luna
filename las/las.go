package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"runtime"
	"os/exec"
	"path/filepath"
	"unicode"
)

var section string = "text"
var input_files []string
var DataBuffer []byte
var TextBuffer []byte
var ExtendedDataBuffer []byte
var current_filename string = ""

func execute(command string) bool {
	shell := "sh"
	flag := "-c"
	if runtime.GOOS == "windows" {
		shell = "cmd"
		flag = "/C"
	}

	cmd := exec.Command(shell, flag, command)
	output, err := cmd.CombinedOutput()
	fmt.Printf(string(output))

	if err != nil {
		return false	
	}
	return true
}

func write(b []byte) {
	switch section {
	case "data":
		DataBuffer = append(DataBuffer, b...)
	case "text":
		TextBuffer = append(TextBuffer, b...)
	case "edata":
		ExtendedDataBuffer = append(ExtendedDataBuffer, b...)
	}
}

func isRegister(word string) byte {
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
		return 0xff
	}
}

var errors = []string{
	"no input files",
	"no such file or directory",
	"invalid register name",
	"invalid operand to instruction",
	"invalid instruction mnemonic",
	"immediate value too large",
	"missing terminating '\"' character",
	"expected string",
}
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
}

func parse(text string) []byte {
	// Check for number
	if _, err := strconv.ParseInt(text, 0, 64); err == nil {
		num, _ := strconv.ParseInt(text, 0, 64)

		high := byte(num >> 8)
		low := byte(num & 0xFF)
		return []byte{high, low}
	}
	if isRegister(text) != 0xff {	
		return []byte{byte(isRegister(text))}
	}
	if string(text[0]) == "\"" {
		if string(text[len(text)-1]) != "\"" {
			error(6, "")
		}
		text = strings.Trim(text, "\"")
		if len(text) > 2 {
			error(5, "'"+text+"'")
		} else if len(text) == 1 {
			text = string(byte(00)) + text
		}
		return []byte(text)
	}
	return append([]byte("LR_"+text), 0x00)
}

func formatString(text string) string {
	var replace = [][2]string {
		{"\\0", "\000"},
		{"\\n", "\n"},
		{"\\r", "\r"},
		{"\\033", "\033"},
	}
	for _, pair := range replace {
		text = strings.ReplaceAll(text, pair[0], pair[1])
	}
	return text
}

func Lex(text string) []string {
	var tokens = []string {}
	var buf = []rune {}

	for _, r := range text {
		switch {
		case r == '\n':	
			if len(buf) > 0 {
				tokens = append(tokens, string(buf))
				buf = buf[:0]
			}
			tokens = append(tokens, "\n")
		case unicode.IsSpace(r):
			if len(buf) > 0 {
				tokens = append(tokens, string(buf))
				buf = buf[:0]
			}
		default:
			buf = append(buf, r)
		}
	}

	if len(buf) > 0 {
		tokens = append(tokens, string(buf))
	}

	return tokens
}

func assemble(text string) {
	words := Lex(text)

	for i := 0; i < len(words); i++ {
		words[i] = strings.TrimSuffix(words[i], ",")
	}

	for i := 0; i < len(words); i++ {
		if words[i] == "#define" {
			alias := words[i + 1]
			actual := words[i + 2]
			words = append(words[:i], words[i + 3:]...)
			for j := 0; j < len(words); j++ {
				if words[j] == alias {
					words[j] = actual
				}
			}	
		}
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

			words[i] = strings.TrimSuffix(words[i], ":")
			write(append([]byte("LD_" + words[i]), 0x00))

			tocompile := words[i+1 : end]
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
		case "#", "//", ";":
			for j := i + 1; j < len(words); j++ {
				if words[j] == "\n" {
					i = j
					break
				}
			}
		case "\n":
			continue
		case "mov":
			write([]byte{0x01})

			var mode byte
			if isRegister(words[i+2]) == 0xff {
				mode = 0x01
			} else {
				mode = 0x02
			}
			write([]byte{mode})

			dst := isRegister(words[i+1])
			if dst == 0xff {
				error(2, "'"+words[i+1]+"'")
			}
			write([]byte{dst})

			if mode == 0x02 {
				src := isRegister(words[i+2])
				write([]byte{src})
			} else {
				value := parse(words[i+2])
				write(value)
			}
			i = i + 2
		case "hlt":
			write([]byte{0x02})
		case "jmp":
			write([]byte{0x03})

			if isRegister(words[i+1]) == 0xff {
				write([]byte{0x01})
			} else {
				write([]byte{0x02})
			}

			value := parse(words[i+1])
			write(value)
			i = i + 1
		case "int":
			write([]byte{0x04})
			value := parse(words[i+1])
			if len(value) > 2 {
				error(3, "'"+string(value)+"'")
			}
			write(value)
			i = i + 1
		case "jnz":
			write([]byte{0x05})

			if isRegister(words[i+2]) == 0xff {
				write([]byte{0x01})
			} else {
				write([]byte{0x02})
			}

			register := isRegister(words[i+1])
			if register == 0xff {
				error(2, "'"+words[i+1]+"'")
			}
			write([]byte{register})

			value := parse(words[i+2])
			write(value)
			i = i + 2
		case "nop":
			write([]byte{0x06})
		case "cmp":
			check := isRegister(words[i+1])
			one := isRegister(words[i+2])
			two := isRegister(words[i+3])
			if check == 0xff {
				error(2, "'"+words[i+1]+"'")
			}
			if one == 0xff {
				error(2, "'"+words[i+2]+"'")
			}
			if two == 0xff {
				error(2, "'"+words[i+3]+"'")
			}
			write([]byte{0x07})
			write([]byte{check})
			write([]byte{one})
			write([]byte{two})
			i = i + 3
		case "jz":
			write([]byte{0x08})

			if isRegister(words[i+2]) == 0xff {
				write([]byte{0x01})
			} else {
				write([]byte{0x02})
			}

			register := isRegister(words[i+1])
			if register == 0xff {
				error(2, "'"+words[i+1]+"'")
			}
			write([]byte{register})

			value := parse(words[i+2])
			write(value)
			i = i + 2
		case "inc":
			write([]byte{0x09})
			reg := isRegister(words[i+1])
			if reg == 0xff {
				error(2, "'"+words[i+1]+"'")
			}
			write([]byte{reg})
			i = i + 1
		case "dec":
			write([]byte{0x0a})
			reg := isRegister(words[i+1])
			if reg == 0xff {
				error(2, "'"+words[i+1]+"'")
			}
			write([]byte{reg})
			i = i + 1
		case "push":
			write([]byte{0x0b})
			if isRegister(words[i+1]) == 0xff {
				write([]byte{0x01})
			} else {
				write([]byte{0x02})
			}
			write(parse(words[i+1]))
			i = i + 1
		case "pop":
			write([]byte{0x0c})
			reg := isRegister(words[i+1])
			if reg == 0xff {
				error(2, "'"+words[i+1]+"'")
			}
			write([]byte{reg})
			i = i + 1
		case "add":
			check := isRegister(words[i+1])
			one := isRegister(words[i+2])
			two := isRegister(words[i+3])
			if check == 0xff {
				error(2, "'"+words[i+1]+"'")
			}
			if one == 0xff {
				error(2, "'"+words[i+2]+"'")
			}
			if two == 0xff {
				error(2, "'"+words[i+3]+"'")
			}
			write([]byte{0x0d})
			write([]byte{check})
			write([]byte{one})
			write([]byte{two})
			i = i + 3
		case "sub":
			check := isRegister(words[i+1])
			one := isRegister(words[i+2])
			two := isRegister(words[i+3])
			if check == 0xff {
				error(2, "'"+words[i+1]+"'")
			}
			if one == 0xff {
				error(2, "'"+words[i+2]+"'")
			}
			if two == 0xff {
				error(2, "'"+words[i+3]+"'")
			}
			write([]byte{0x0e})
			write([]byte{check})
			write([]byte{one})
			write([]byte{two})
			i = i + 3
		case "mul":
			check := isRegister(words[i+1])
			one := isRegister(words[i+2])
			two := isRegister(words[i+3])
			if check == 0xff {
				error(2, "'"+words[i+1]+"'")
			}
			if one == 0xff {
				error(2, "'"+words[i+2]+"'")
			}
			if two == 0xff {
				error(2, "'"+words[i+3]+"'")
			}
			write([]byte{0x0f})
			write([]byte{check})
			write([]byte{one})
			write([]byte{two})
			i = i + 3
		case "div":
			check := isRegister(words[i+1])
			one := isRegister(words[i+2])
			two := isRegister(words[i+3])
			if check == 0xff {
				error(2, "'"+words[i+1]+"'")
			}
			if one == 0xff {
				error(2, "'"+words[i+2]+"'")
			}
			if two == 0xff {
				error(2, "'"+words[i+3]+"'")
			}
			write([]byte{0x10})
			write([]byte{check})
			write([]byte{one})
			write([]byte{two})
			i = i + 3
		case "igt":
			check := isRegister(words[i+1])
			one := isRegister(words[i+2])
			two := isRegister(words[i+3])
			if check == 0xff {
				error(2, "'"+words[i+1]+"'")
			}
			if one == 0xff {
				error(2, "'"+words[i+2]+"'")
			}
			if two == 0xff {
				error(2, "'"+words[i+3]+"'")
			}
			write([]byte{0x11})
			write([]byte{check})
			write([]byte{one})
			write([]byte{two})
			i = i + 3
		case "ilt":
			check := isRegister(words[i+1])
			one := isRegister(words[i+2])
			two := isRegister(words[i+3])
			if check == 0xff {
				error(2, "'"+words[i+1]+"'")
			}
			if one == 0xff {
				error(2, "'"+words[i+2]+"'")
			}
			if two == 0xff {
				error(2, "'"+words[i+3]+"'")
			}
			write([]byte{0x12})
			write([]byte{check})
			write([]byte{one})
			write([]byte{two})
			i = i + 3
		case "and":
			check := isRegister(words[i+1])
			one := isRegister(words[i+2])
			two := isRegister(words[i+3])
			if check == 0xff {
				error(2, "'"+words[i+1]+"'")
			}
			if one == 0xff {
				error(2, "'"+words[i+2]+"'")
			}
			if two == 0xff {
				error(2, "'"+words[i+3]+"'")
			}
			write([]byte{0x13})
			write([]byte{check})
			write([]byte{one})
			write([]byte{two})
			i = i + 3
		case "or":
			check := isRegister(words[i+1])
			one := isRegister(words[i+2])
			two := isRegister(words[i+3])
			if check == 0xff {
				error(2, "'"+words[i+1]+"'")
			}
			if one == 0xff {
				error(2, "'"+words[i+2]+"'")
			}
			if two == 0xff {
				error(2, "'"+words[i+3]+"'")
			}
			write([]byte{0x14})
			write([]byte{check})
			write([]byte{one})
			write([]byte{two})
			i = i + 3
		case "nor":
			check := isRegister(words[i+1])
			one := isRegister(words[i+2])
			two := isRegister(words[i+3])
			if check == 0xff {
				error(2, "'"+words[i+1]+"'")
			}
			if one == 0xff {
				error(2, "'"+words[i+2]+"'")
			}
			if two == 0xff {
				error(2, "'"+words[i+3]+"'")
			}
			write([]byte{0x15})
			write([]byte{check})
			write([]byte{one})
			write([]byte{two})
			i = i + 3
		case "not":
			check := isRegister(words[i+1])
			one := isRegister(words[i+2])
			if check == 0xff {
				error(2, "'"+words[i+1]+"'")
			}
			if one == 0xff {
				error(2, "'"+words[i+2]+"'")
			}
			write([]byte{0x16})
			write([]byte{check})
			write([]byte{one})
			i = i + 2
		case "xor":
			check := isRegister(words[i+1])
			one := isRegister(words[i+2])
			two := isRegister(words[i+3])
			if check == 0xff {
				error(2, "'"+words[i+1]+"'")
			}
			if one == 0xff {
				error(2, "'"+words[i+2]+"'")
			}
			if two == 0xff {
				error(2, "'"+words[i+3]+"'")
			}
			write([]byte{0x17})
			write([]byte{check})
			write([]byte{one})
			write([]byte{two})
			i = i + 3
		case "lod":
			check := isRegister(words[i+1])
			one := isRegister(words[i+2])
			if check == 0xff {
				error(2, "'"+words[i+1]+"'")
			}
			if one == 0xff {
				error(2, "'"+words[i+2]+"'")
			}
			write([]byte{0x18})
			write([]byte{check})
			write([]byte{one})
			i = i + 2
		case "str":
			check := isRegister(words[i+1])
			one := isRegister(words[i+2])
			if check == 0xff {
				error(2, "'"+words[i+1]+"'")
			}
			if one == 0xff {
				error(2, "'"+words[i+2]+"'")
			}
			write([]byte{0x19})
			write([]byte{check})
			write([]byte{one})
			i = i + 2
		case "lodw":
			check := isRegister(words[i+1])
			one := isRegister(words[i+2])
			if check == 0xff {
				error(2, "'"+words[i+1]+"'")
			}
			if one == 0xff {
				error(2, "'"+words[i+2]+"'")
			}
			write([]byte{0x20})
			write([]byte{check})
			write([]byte{one})
			i = i + 2
		case "call":
			label := words[i + 1]
			assemble(`
			mov re1, pc
			mov r0, 20
			add re1, re1, r0
			push re1
			jmp	` + label)
			i = i + 1
		case "ret":
			assemble(`jmp re1`)
		case ".ascii":	
			var value string	
			var tokens = []string {}
			
			if string(words[i+1][0]) != "\"" {
				error(7, "'" + words[i+1] + "'")
			}
			if strings.HasSuffix(words[i + 1], "\"") {
				value = strings.Trim(words[i + 1], "\"")
				value = formatString(value)
				write([]byte(value))
				i = i + 1
				continue
			}
			
			ending := 0
			for j := i + 1; j < len(words); j++ {
				tokens = append(tokens, words[j])
				if strings.HasSuffix(words[j], "\"") {
					ending = j
					break
				}
			}
			if ending == 0 {
				error(6, "'" + words[i + 1] + "'")
			}
			
			tokens[0] = strings.TrimPrefix(tokens[0], "\"")
			tokens[len(tokens) - 1] = strings.TrimSuffix(tokens[len(tokens) - 1], "\"")
			value = strings.Join(tokens, " ")
			value = formatString(value)
			write([]byte(value))
			i = ending
		case ".asciz":	
			var value string	
			var tokens = []string {}
			
			if string(words[i+1][0]) != "\"" {
				error(7, "'" + words[i+1] + "'")
			}
			if strings.HasSuffix(words[i + 1], "\"") {
				value = strings.Trim(words[i + 1], "\"")
				value = formatString(value)
				value = value + string("\000")
				write([]byte(value))
				i = i + 1
				continue
			}
			
			ending := 0
			for j := i + 1; j < len(words); j++ {
				tokens = append(tokens, words[j])
				if strings.HasSuffix(words[j], "\"") {
					ending = j
					break
				}
			}
			if ending == 0 {
				error(6, "'" + words[i + 1] + "'")
			}
			
			tokens[0] = strings.TrimPrefix(tokens[0], "\"")
			tokens[len(tokens) - 1] = strings.TrimSuffix(tokens[len(tokens) - 1], "\"")
			value = strings.Join(tokens, " ")
			value = formatString(value)
			value = value + string("\000")
			write([]byte(value))
			i = ending
		default:
			error(4, "'"+words[i]+"'")
		}
	}
}

func splitFile(path string) (name string, ext string) {
	ext = filepath.Ext(path)
	name = filepath.Base(path)	
	if ext != "" {
		name = name[:len(name)-len(ext)]
	}
	return
}

func cleanupFiles(files []string) {
	if runtime.GOOS != "windows" {
		for _, file := range files {
			execute("rm -f " + file)
		}
	} else {
		for _, file := range files {
			execute("del /f " + file)
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
	var object_files = []string {}

	for i := 1; i < len(os.Args); i++ {
		arg := os.Args[i]

		switch arg {
		case "-v":
			fmt.Println("Luna Compiler Collection version 2.0")
			fmt.Println("Target: luna-l2")
			os.Exit(0)
		case "-o":
			output_filename = os.Args[i+1]
			i++
		case "-c":
			nolink = true
		default:
			input_files = append(input_files, arg)
		}
	}

	if output_filename == "" {
		if nolink == false {
			output_filename = "a.bin"
		} else {
			output_filename = "a.o"
		}
	}

	for _, file := range input_files {
		data, err := os.ReadFile(file)
		if err != nil {
			error(1, "'" + file + "'")
			os.Exit(1)
		}
		current_filename = file
		// Assemble everything
		assemble(string(data))
		// Error checking
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
		if Errors > 0 {
			continue
		}
		// Write everything
		name, _ := splitFile(file)
		buffer := append([]byte{0x4c, 0x32, 0x4f, 0xc2, 0x80, 0x7d}, append(DataBuffer, append([]byte{0xc2, 0x80, 0x7e}, append(TextBuffer, append([]byte{0xc2, 0x80, 0x7f}, ExtendedDataBuffer...)...)...)...)...)
		os.WriteFile(name + ".o", buffer, 0644)
		object_files = append(object_files, name + ".o")
		// Reset
		Errors = 0
		Warnings = 0
		DataBuffer = []byte {}
		TextBuffer = []byte {}
		ExtendedDataBuffer = []byte {}
		section = "text"
	}	

	if nolink == true {
		os.Exit(0)
	}

	success := execute("l2ld " + strings.Join(object_files, " ") + " -o " + output_filename)
	if success != true {
		cleanupFiles(object_files)
		fmt.Println("\033[1;39mlcc: \033[1;31merror: \033[1;39mlinker command failed.\033[0m")
		os.Exit(1)
	}
	cleanupFiles(object_files)
}
