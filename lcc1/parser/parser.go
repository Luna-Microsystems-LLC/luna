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

func expect(toktype lexer.TokenType, iterator *int, tokens []lexer.Token) string {
	i := *iterator
	token := tokens[i]
	var value string
	if token.Type == toktype {
		i++
		value = token.Value
	} else {
		error.Error(1, "'" + tokens[i].Value + "'")
		return ""
	}
	*iterator = i
	return value
}

func Parse_entry(tokens []lexer.Token) {
	i := 0
	// var node Node		
	expect(lexer.TokType, &i, tokens)
	name := expect(lexer.TokIdent, &i, tokens)
	expect(lexer.TokLParen, &i, tokens)
	expect(lexer.TokRParen, &i, tokens)
	expect(lexer.TokLCurly, &i, tokens)
	closer := -1
	for j := i; j < len(tokens); j++ {
		if tokens[j].Type == lexer.TokRCurly {
			closer = j - 1
			break
		} else {
			
		}
	}
	if closer == -1 {
		error.Error(2, "closing '}'")
	}
	expect(lexer.TokRCurly, &i, tokens)
	ASTAdd(Node{Type: NodeFunction, Value: name})	
} 
