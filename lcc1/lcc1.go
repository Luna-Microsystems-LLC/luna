package main

import (	
	"lcc1/lexer"
	"lcc1/parser"
	"lcc1/error"
	"lcc1/codegen"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		error.Error(0, "")
		os.Exit(1)
	}

	var input_files = []string {}
	// var output_file string = ""

	for i := 1; i < len(os.Args); i++ {
		arg := os.Args[i]
		switch arg {
		case "-o":
			// output_file = os.Args[i + 1]
			i++
		default:
			input_files = append(input_files, arg)
		}
	}

	for _, file := range input_files {
		data, err := os.ReadFile(file)
		if err != nil {
			os.Exit(1)
		}
		tokens := lexer.Lex(string(data))
		parser.Parse_entry(tokens)
	}

	code := codegen.Codegen(parser.AbstractSyntaxTree)

	print(code)
}
