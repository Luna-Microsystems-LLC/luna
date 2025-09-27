package main

import (	
	"lcc1/lexer"
	"lcc1/parser"
	"lcc1/error"	
	"os"
	"fmt"
)

func main() {
	if len(os.Args) < 2 {
		error.Error(0, "")
		os.Exit(1)
	}

	var input_files = []string {}
	var output_file string = ""

	for i := 1; i < len(os.Args); i++ {
		arg := os.Args[i]
		switch arg {
		case "-o":
			output_file = os.Args[i + 1]
			i++
		case "-v":
			fmt.Println("Luna Compiler Collection version 2.0")
			fmt.Println("Target: luna-l2")
			os.Exit(0)
		default:
			input_files = append(input_files, arg)
		}
	}

	if output_file == "" {
		output_file = "a.s"
	}

	for _, file := range input_files {
		data, err := os.ReadFile(file)
		if err != nil {
			os.Exit(1)
		}
		tokens := lexer.Lex(string(data))
		parser.Parse(tokens)
	}

	os.WriteFile(output_file, []byte(".text\n" + parser.Code1 + "\n" + parser.Code2), 0644)
}
