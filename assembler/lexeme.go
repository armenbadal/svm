package assembler

import "fmt"

type token int

const (
	xUnknown token = iota
	xIdent
	xRegister
	xNumber
	xOperation
	xNewLine
	xColon
	xLeftBr
	xRightBr
	xPlus
	xMinus
	xEos
)

var tokenNames = map[token]string{
	xUnknown:   "Unknown",
	xIdent:     "Identifier",
	xOperation: "Operation",
	xRegister:  "Register",
	xNumber:    "Number",
	xNewLine:   "\\n",
	xColon:     ":",
	xLeftBr:    "[",
	xRightBr:   "]",
	xPlus:      "+",
	xMinus:     "-",
	xEos:       "Eos",
}

func (t token) String() string {
	if name, ok := tokenNames[t]; ok {
		return name
	}
	return fmt.Sprintf("Token(%d)", t)
}

type lexeme struct {
	kind  token
	value string
}

func (l lexeme) String() string {
	if l.kind == xIdent {
		return fmt.Sprintf("IDENT<%s>", l.value)
	}
	if l.kind == xOperation {
		return fmt.Sprintf("OP<%s>", l.value)
	}
	if l.kind == xRegister {
		return fmt.Sprintf("REG<%s>", l.value)
	}
	if l.kind == xNumber {
		return fmt.Sprintf("NUM<%s>", l.value)
	}
	return tokenNames[l.kind]
}
