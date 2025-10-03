package parser

import (
	"lcc1/lexer"
	"lcc1/error"
	"strings"
	"fmt"
)

var level int = 0
var Code1 string = ""
var Code2 string = ""


var IDCounter = 1

const (
	NUMBER int = iota
	STRING
)
type Variable_Static struct {
	Name string
	Type int
	Value any
}

type Variable_Dynamic struct {
	Name string
	Type int
	Value any
	Location uint16
	Length uint16
}

var Location uint16 = 1000

var Variables = []Variable_Dynamic {}

func Write(text string, spaced bool) {
	if spaced == false {
		Code2 = Code2 + text + "\n"
	} else {
		Code2 = Code2 + "    " + text + "\n"
	}
}

func WritePre(text string, spaced bool) {
	if spaced == false {
		Code1 = Code1 + text + "\n"
	} else {
		Code1 = Code1 + "    " + text + "\n"
	}
}

func CreateStatic(variable Variable_Static) {
	WritePre(variable.Name + ":\n    .asciz \"" + variable.Value.(string) + "\"", false)	
}

func CreateDynamic(Name string, Type int, Value any, Length uint16) uint16 {
	// Add to entry
	// Assume memory section starts at 1000 for now.
	Variables = append(Variables, Variable_Dynamic{Name: Name, Type: Type, Value: Value, Length: Length, Location: Location})
	oldloc := Location
	Location += Length
	return oldloc
}

func LookupVariable(Name string, Enforce bool) Variable_Dynamic {
	for _, variable := range Variables {
		if variable.Name == Name {
			return variable
		}
	}
	if Enforce == true {
		error.Error(4, "'" + Name + "'")
		return Variable_Dynamic{Name: "__ZERO", Type: NUMBER, Value: 0, Length: 0, Location: 0}
	} else {
		return Variable_Dynamic{Name: "__ZERO", Type: NUMBER, Value: 0, Length: 0, Location: 0}
	}
}

func StringParse(tokens []lexer.Token, start int) (string, int) {
	// Start would be the first token
	var str string = ""
	var loc int = 0
	if strings.HasSuffix(tokens[start].Value, "\"") {
		tokens[start].Value = strings.Trim(tokens[start].Value, "\"")
		str = tokens[start].Value
		loc = start
	} else {
		var strtokens = []string { tokens[start].Value }
		for k := start + 1; k < len(tokens); k++ {
			strtokens = append(strtokens, tokens[k].Value)
			if strings.HasSuffix(tokens[k].Value, "\"") {
				start = k
				break
			}
		}
		str = strings.Join(strtokens, " ")
		str = strings.Trim(str,  "\"")
		loc = start
	}
	
	return str, loc
}

func Parse(tokens []lexer.Token) {
	i := 0
	expect := func(toktype lexer.TokenType) string {
		var value string
		if i >= len(tokens) {
			error.Error(1, "'<EOF>'")
		}
		if tokens[i].Type == toktype {
			value = tokens[i].Value
			i++
		} else {
			error.Error(1, "'" + tokens[i].Value + "'")
			return ""
		}
		return value
	}
	peek := func(lookahead int) lexer.Token {	
		return tokens[i + lookahead]
	}
	
	for {
		if i >= len(tokens) {
			break
		}
		switch level {
		case 0:
			_type := expect(lexer.TokType)
			name := expect(lexer.TokIdent)
			
			var rtype int
			if _type == "int" {
				rtype = NUMBER
			}

			if LookupVariable(name, false).Name != "__ZERO" {
				// print(LookupVariable(name, false).Name)
				error.Error(3, "'" + name + "'")
			}

			CreateDynamic(name, rtype, 0, 2)

			if peek(0).Type == lexer.TokLParen {
				expect(lexer.TokLParen)
				expect(lexer.TokRParen)
				expect(lexer.TokLCurly)

				var Children = []lexer.Token {}
				ending := -1
				for j := i; j < len(tokens); j++ {
					if tokens[j].Type == lexer.TokRCurly {
						ending = j
						break
					} else {	
						Children = append(Children, tokens[j])
					}
				}
				if ending == -1 {
					error.Error(2, "'}'")
				} else {
					i = ending
				}
			
				expect(lexer.TokRCurly)

				if name == "main" {
					name = "_start"
				}
				Write(name + ":", false)
				if len(Children) > 0 {
					level = 1
					Parse(Children)
					level = 0
				}
				if name != "_start" {
					Write("ret", true)
				}
				i++
				if i < len(tokens) {
					print(tokens[i].Value)
					Parse(tokens)
				}
			} else if peek(0).Type == lexer.TokEqual {
				
			} else {
				error.Error(1, "'" + peek(0).Value + "'")
			}
		case 1:	
			// Variable reassignment / function call
			var type_ lexer.TokenType = peek(0).Type
			switch type_ {
			case lexer.TokIdent:
				name := expect(lexer.TokIdent)
				if peek(0).Type == lexer.TokLParen {	
					expect(lexer.TokLParen)
					var expComma bool = false
					for j := i; j < len(tokens); j++ {
						if tokens[j].Type == lexer.TokRParen {
							i = j
							break
						} else {
							if expComma == true {
								if tokens[j].Type != lexer.TokComma {
									error.Error(2, "','")
								} else {
									expComma = false
									continue
								}
							}
							if strings.HasPrefix(tokens[j].Value, "\"") {
								str, end := StringParse(tokens, j)
								j = end
								CreateStatic(Variable_Static{Name: "var_" + fmt.Sprintf("%d", IDCounter), Type: STRING, Value: str})
								Write("push var_" + fmt.Sprintf("%d", IDCounter), true)
								IDCounter++
								expComma = true
							} else {
								Write("push " + tokens[j].Value, true)
								expComma = true
							}
						}
					}

					expect(lexer.TokRParen)
					expect(lexer.TokSemi)
					Write("call " + name, true)
				} 
			case lexer.TokReturn:
				expect(lexer.TokReturn)
				name := expect(lexer.TokIdent)
				expect(lexer.TokSemi)
				LookupVariable(name, true)
				Write("mov t7, " + name, true)
			case lexer.TokSemi:
				expect(lexer.TokSemi)
			default:
				error.Error(1, "'" + tokens[i].Value + "'")
			}	
		}
	}
} 
