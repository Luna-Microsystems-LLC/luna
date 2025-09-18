package parser

import (
	"lcc1/lexer"
	"lcc1/error"
)

type NodeType int

const (
	NodeFunction NodeType = iota
	NodeName
	NodeDeclare
	NodeReturn
)

type Node struct {
	Type NodeType
	Value string
	Children []Node
}
var AbstractSyntaxTree = []Node {}

func ASTAdd(node Node) {
	AbstractSyntaxTree = append(AbstractSyntaxTree, node)
}

func Parse_entry(tokens []lexer.Token) {
	i := 0
	expect := func(toktype lexer.TokenType) string {
		token := tokens[i]
		if token.Type == toktype {
			i++
			return token.Value
		} else {
			error.Error(1, "'" + tokens[i].Value + "'")
			return ""
		}
	}	
	expect(lexer.TokType)
	name := expect(lexer.TokIdent)
	expect(lexer.TokLParen)
	expect(lexer.TokRParen)
	expect(lexer.TokLCurly)
	expect(lexer.TokRCurly)
	ASTAdd(Node{Type: NodeFunction, Value: name})	
} 
