package assembler

import (
	"bufio"
	"fmt"
	"slices"
	"unicode"
)

func Assemble(file string) ([]byte, error) {
	return nil, nil
}

type parser struct {
	source    *bufio.Reader
	lookahead lexeme
}

const (
	xUnknown = iota
	xIdent
	xOperation
	xRegister
	xNumber
	xNewLine
	xColon
	xLeftBr
	xRightBr
	xPlus
	xMinus
	xEos
)

type lexeme struct {
	token int
	value string
}

func (l *lexeme) is(exp int) bool {
	return l.token == exp
}

func (p *parser) parse() {
	p.lookahead = p.scanOne()

	if p.lookahead.is(xNewLine) {
		p.parseNewLines()
	}
	for !p.lookahead.is(xEos) {
		p.parseLine()
	}
}

func (p *parser) parseNewLines() {
	for p.lookahead.is(xNewLine) {
		p.lookahead = p.scanOne()
	}
}

// Line = [Ident ':'] [OPERATION [Argument]].
// Argument = Number | '[' Register ('+'|'-') Number ']'.
// Register = IP | SP | FP.
func (p *parser) parseLine() error {
	label, err := p.parseLabel()
	if err != nil {
		return err
	}
	if len(label) != 0 {
		// store label address in table
	}

	if !p.lookahead.is(xOperation) {
		return nil
	}

	return nil
}

func (p *parser) parseOperation() error {
	if !p.lookahead.is(xOperation) {
		return nil
	}

	oper, _ := p.match(xOperation)

	switch p.lookahead.token {
	case xNumber:
		p.lookahead = p.scanOne()
	case xLeftBr:
		p.lookahead = p.scanOne()
		if !p.lookahead.is(xRegister) {
			// error
		}
		p.lookahead = p.scanOne()
	case xNewLine:
	}
	_ = oper

	return nil
}

func (p *parser) parseLabel() (string, error) {
	if !p.lookahead.is(xIdent) {
		return "", nil
	}

	name, err := p.match(xIdent)
	if err != nil {
		return "", err
	}
	_, err = p.match(xColon)

	return name, nil
}

func (p *parser) match(expected int) (string, error) {
	if p.lookahead.token == expected {
		text := p.lookahead.value
		p.lookahead = p.scanOne()
		return text, nil
	}

	return "", fmt.Errorf("Expected %d but got %d", expected, p.lookahead.token)
}

var operations = []string{
	"NOP", "PUSH", "POP", "CALL", "RETURN", "JUMP", "JZ", "HALT",
	"ADD", "SUB", "MUL", "DIV", "MOD", "NEG", "AND", "OR", "NOT",
	"EQ", "NE", "LT", "LE", "GT", "GE",
}
var registers = []string{"IP", "SP", "FP"}
var metasymbols = map[rune]int{
	':':  xColon,
	'[':  xLeftBr,
	']':  xRightBr,
	'+':  xPlus,
	'-':  xMinus,
	'\n': xNewLine,
}

func (p *parser) scanOne() lexeme {
	ch := p.readChar()

	// հոսքի վերջը
	if ch == 0 {
		return lexeme{token: xEos, value: "EOS"}
	}

	// անտեսել բացատները
	isSpace := func(c rune) bool { return c == ' ' || c == '\t' || c == '\r' }
	if isSpace(ch) {
		p.readCharsWhile(isSpace)
		ch = p.readChar()
	}

	// անտեսել մեկնաբանությունները
	if ch == ';' {
		p.readCharsWhile(func(c rune) bool { return c != '\n' })
		ch = p.readChar()
	}

	// գործողության անուն կամ իդենտիֆիկատոր
	if unicode.IsLetter(ch) {
		p.source.UnreadRune()
		text := p.readCharsWhile(func(c rune) bool { return unicode.IsLetter(c) || unicode.IsDigit(c) })
		switch {
		case slices.Contains(operations, text):
			return lexeme{token: xOperation, value: text}
		case slices.Contains(registers, text):
			return lexeme{token: xRegister, value: text}
		default:
			return lexeme{token: xIdent, value: text}
		}
	}

	// ամբողջ թիվ
	if unicode.IsDigit(ch) {
		p.source.UnreadRune()
		text := p.readCharsWhile(unicode.IsDigit)
		return lexeme{token: xNumber, value: text}
	}

	// այլ սիմվոլներ
	if tok, ok := metasymbols[ch]; ok {
		return lexeme{token: tok, value: string(ch)}
	}

	return lexeme{token: xUnknown, value: string(ch)}
}

func (p *parser) readCharsWhile(pred func(rune) bool) string {
	var text string
	ch := p.readChar()
	for pred(ch) && ch != 0 {
		text += string(ch)
		ch = p.readChar()
	}
	p.source.UnreadRune()
	return text
}

func (p *parser) readChar() rune {
	ch, _, err := p.source.ReadRune()
	if err != nil {
		return 0
	}
	return ch
}
