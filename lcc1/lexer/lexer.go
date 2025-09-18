package lexer

import (
	"text/scanner"
	"strconv"
	"strings"
)

type TokenType int

const (
	TokType TokenType = iota
	TokReturn
	TokIf
	TokElse
	TokIdent
	TokNumber
	TokLParen
	TokRParen
	TokLCurly
	TokRCurly
	TokLBrack
	TokRBrack
	TokEOF
)

type Token struct {
	Type TokenType
	Value string
}

func contains(set string, c byte) bool {
    for i := 0; i < len(set); i++ {
        if set[i] == c {
            return true
        }
    }
    return false
}

func Lex(code string) []Token {
	var tokens = []Token {}
	var s scanner.Scanner
    s.Init(strings.NewReader(code))
    s.Mode = scanner.ScanIdents | scanner.ScanInts | scanner.ScanChars | scanner.ScanStrings | scanner.SkipComments

    for tok := s.Scan(); tok != scanner.EOF; tok = s.Scan() {
		content := s.TokenText()	

		if content == "int" {
			tokens = append(tokens, Token{Type: TokType, Value: content})
		} else if content == "return" {
			tokens = append(tokens, Token{Type: TokReturn, Value: content})
		} else if content == "if" {
			tokens = append(tokens, Token{Type: TokIf, Value: content})
		} else if content == "else" {
			tokens = append(tokens, Token{Type: TokElse, Value: content})
		} else if _, err := strconv.Atoi(content); err == nil {
			tokens = append(tokens, Token{Type: TokNumber, Value: content})
		} else if content == "(" {
			tokens = append(tokens, Token{Type: TokLParen, Value: content})
		} else if content == ")" {
			tokens = append(tokens, Token{Type: TokRParen, Value: content})
		} else if content == "{" {
			tokens = append(tokens, Token{Type: TokLCurly, Value: content})
		} else if content == "}" {
			tokens = append(tokens, Token{Type: TokRCurly, Value: content})
		} else {
			tokens = append(tokens, Token{Type: TokIdent, Value: content})
		} 
	}
	return tokens
}
