package assembler

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"svm/bytecode"
	"unicode"
)

func Assemble(file string) ([]byte, error) {
	// բացել ֆայլը
	input, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer input.Close()

	// վերլուծել ծրագիրն ու կառուցել բայթկոդը
	p := &parser{
		source:    bufio.NewReader(input),
		lookahead: lexeme{token: xUnknown, value: "?"},
		line:      1,
		builder:   bytecode.NewBuilder(),
	}
	p.parse()
	p.builder.Validate() // լուծել անորոշ հղումները

	return p.builder.Bytes(), nil
}

var operations = map[string]byte{
	"NOP":   bytecode.Nop,
	"PUSH":  bytecode.Push,
	"POP":   bytecode.Pop,
	"CALL":  bytecode.Call,
	"RET":   bytecode.Ret,
	"JUMP":  bytecode.Jump,
	"JZ":    bytecode.Jz,
	"HALT":  bytecode.Halt,
	"ADD":   bytecode.Add,
	"SUB":   bytecode.Sub,
	"MUL":   bytecode.Mul,
	"DIV":   bytecode.Div,
	"MOD":   bytecode.Mod,
	"NEG":   bytecode.Neg,
	"AND":   bytecode.And,
	"OR":    bytecode.Or,
	"NOT":   bytecode.Not,
	"EQ":    bytecode.Eq,
	"NE":    bytecode.Ne,
	"LT":    bytecode.Lt,
	"LE":    bytecode.Le,
	"GT":    bytecode.Gt,
	"GE":    bytecode.Ge,
	"INPUT": bytecode.Input,
	"PRINT": bytecode.Print,
}

var registers = map[string]uint16{
	"IP": bytecode.InstructionPointer,
	"SP": bytecode.StackPointer,
	"FP": bytecode.FramePointer,
}

type parser struct {
	source    *bufio.Reader
	lookahead lexeme
	line      int

	builder *bytecode.Builder
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

	if p.has(xNewLine) {
		p.parseNewLines()
	}
	for !p.has(xEos) {
		p.parseLine()
	}
}

func (p *parser) parseNewLines() {
	for p.has(xNewLine) {
		p.lookahead = p.scanOne()
	}
}

func (p *parser) parseLine() error {
	if p.has(xIdent) {
		err := p.parseLabel()
		if err != nil {
			return err
		}
	}

	if p.has(xOperation) {
		err := p.parseOperation()
		if err != nil {
			return err
		}
	}

	if !p.has(xNewLine) {
		return fmt.Errorf("Տողի վերջում սպասվում է նոր տողի նշան")
	}
	p.parseNewLines()

	return nil
}

func (p *parser) parseOperation() error {
	opName, _ := p.match(xOperation)
	opcode, exists := operations[opName]
	if !exists {
		return fmt.Errorf("Անծանոթ հրահանգ `%s`", opName)
	}

	// երբ արգումենտը թվային հաստատուն է
	if p.has(xNumber) {
		nlex, _ := p.match(xNumber)
		number, _ := strconv.ParseInt(nlex, 10, 32)
		p.builder.AddWithNumeric(opcode, int32(number))
		return nil
	}

	// երբ արգումենտը անուղղակի հասցեավորում է
	if p.has(xLeftBr) {
		_, _ = p.match(xLeftBr)
		regName, err := p.match(xRegister)
		if err != nil {
			return err
		}
		reg := registers[regName]
		var disp int64 = 1
		if p.has(xPlus) {
			p.match(xPlus)
		} else if p.has(xMinus) {
			p.match(xMinus)
			disp = -1
		} else {
			return fmt.Errorf("Սպասվում է '+' կամ '-' նշանը")
		}
		nlex, err := p.match(xNumber)
		if err != nil {
			return err
		}
		number, _ := strconv.ParseInt(nlex, 10, 16)
		_, err = p.match(xRightBr)
		if err != nil {
			return err
		}
		number *= disp

		p.builder.AddWithAddress(opcode, reg, int16(number))
		return nil
	}

	// երբ արգումենտը պիտակ է (իդենտիֆիկատոր)
	if p.has(xIdent) {
		label, _ := p.match(xIdent)
		p.builder.AddWithLabel(opcode, label)
		return nil
	}

	p.builder.AddBasic(opcode)
	return nil
}

func (p *parser) parseLabel() error {
	name, err := p.match(xIdent)
	if err != nil {
		return err
	}
	_, err = p.match(xColon)

	p.builder.SetLabel(name)
	return nil
}

func (p *parser) match(expected int) (string, error) {
	if p.has(expected) {
		text := p.lookahead.value
		p.lookahead = p.scanOne()
		return text, nil
	}

	return "", fmt.Errorf("Expected %d but got %d", expected, p.lookahead.token)
}

func (p *parser) has(token int) bool {
	return p.lookahead.is(token)
}

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
		text := p.readCharsWhile(isAlphaNumeric)
		if _, exists := operations[text]; exists {
			return lexeme{token: xOperation, value: text}
		}
		if _, exists := registers[text]; exists {
			return lexeme{token: xRegister, value: text}
		}

		return lexeme{token: xIdent, value: text}
	}

	// ամբողջ թիվ
	if unicode.IsDigit(ch) {
		p.source.UnreadRune()
		text := p.readCharsWhile(unicode.IsDigit)
		return lexeme{token: xNumber, value: text}
	}

	// այլ սիմվոլներ
	if tok, ok := metasymbols[ch]; ok {
		if tok == xNewLine {
			p.line++
		}
		return lexeme{token: tok, value: string(ch)}
	}

	return lexeme{token: xUnknown, value: string(ch)}
}

func isAlphaNumeric(c rune) bool {
	return unicode.IsLetter(c) || unicode.IsDigit(c)
}

func isSpace(c rune) bool {
	return c == ' ' || c == '\t' || c == '\r'
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
