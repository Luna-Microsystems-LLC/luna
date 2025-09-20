package codegen

import (
	"lcc1/parser"
)

var Code string = ""

func Codegen(AST []parser.Node) {	
	write := func(text string) {
		Code = Code + text + "\n"
	}

	for i := 0; i < len(AST); i++ {
		node := AST[i]
		switch node.Type {
		case parser.NodeFunction:
			fname := node.Value
			write(fname + ":\n")
			for _, child_node := range node.Children {
				Codegen([]parser.Node{child_node})
			}
		}	
	}
}
