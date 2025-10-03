package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"os/exec"
	"runtime"
)

func stderr(str string) {
	fmt.Fprintln(os.Stderr, str)
}

func execute(command string, displayError bool) bool {
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
		if displayError == true {
			stderr("\033[1;39mlcc: \033[1;31merror: \033[1;39compilation command failed.\033[0m")
			os.Exit(1)
		} else {
			return false
		}
	}
	return true
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
			execute("rm -f " + file, false)
		}
	} else {
		for _, file := range files {
			execute("del /f " + file, false)
		}
	}
}

func main() {
	if len(os.Args) < 2 {
		stderr("\033[1;39mlcc: \033[1;31merror: \033[1;39mno input files\033[0m")
		os.Exit(1)
	}


	var nolink bool = false
	var noassemble bool = false
	var input_files = []string {}
	var cleanup = []string {}
	var output_file string = ""

	for i := 1; i < len(os.Args); i++ {
		arg := os.Args[i]
		if i == 0 {
			continue
		}	
		switch arg {
		case "-c":
			nolink = true
		case "-o":
			output_file = os.Args[i + 1]
			i++
		case "-v":
			fmt.Println("Luna Compiler Collection version 2.0")
			fmt.Println("Target: luna-l2")
			os.Exit(0)
		case "-S":
			noassemble = true	
		default:
			input_files = append(input_files, arg)
		}
	}

	if len(input_files) < 1 {
		stderr("\033[1;39mlcc: \033[1;31merror: \033[1;39mno input files\033[0m")
		os.Exit(1)
	}

	if output_file == "" {
		if nolink == true {
			output_file = "a.o"
		} else {
			output_file = "a.bin"
		}
	}

	var assembly_files = []string {}
	var object_files = []string {}

	// First pass: compile high-level languages to assembly
	for _, file := range input_files {
		// name := filepath.Base(file) // add when we make the C compiler
		ext := filepath.Ext(file)
		name, _ := splitFile(file)

		switch ext {
		case ".c":
			success := execute("lcc1 " + file + " -o " + name + ".s", false)
			if success != true {
				continue
			}
			assembly_files = append(assembly_files, name + ".s")
		case ".s":
			assembly_files = append(assembly_files, file)	
		case ".asm":
			assembly_files = append(assembly_files, file)
		case ".S":
			assembly_files = append(assembly_files, file)
		case ".o":
			object_files = append(object_files, file)
		default:
			stderr("\033[1;39mlcc: \033[1;31merror: \033[1;39munknown file type in '" + file + "'\033[0m")
		}
	}

	if noassemble == true {
		os.Exit(0)
	}

	// Second pass: assemble all assembly files	

	for _, file := range assembly_files {
		name, _ := splitFile(file)		
		success := execute("las -c " + file + " -o " + name + ".o", false)

		if success != true {
			os.Exit(1)
		}

		object_files = append(object_files, name + ".o")
		cleanup = append(cleanup, name + ".o")
	}

	if nolink == true {
		os.Exit(0)
	}
	
	// Third pass: link all assembly files to final executable

	success := execute("l2ld " + strings.Join(object_files, " ") + " -o " + output_file, false)
	if success != true {
		cleanupFiles(cleanup)
		stderr("\033[1;39mlcc: \033[1;31merror: \033[1;39mlinker command failed.\033[0m")
		os.Exit(1)
	}
	cleanupFiles(cleanup)
}
