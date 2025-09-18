package codegen

import (
	"lcc1/parser"
)

func Codegen(AST []parser.Node) string {
	var code string = ""

	write := func(text string) {
		code = code + text + "\n"
	}

	for i := 0; i < len(AST); i++ {
		node := AST[i]
		switch node.Type {
		case parser.NodeFunction:
			fname := node.Value
			write(fname + ":\n")
			for _, child_node := range node.Children {
				Codegen(child_node)
			}
		}	
	}

	return code
}
