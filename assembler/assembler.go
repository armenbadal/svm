package assembler

import (
	"bufio"
	"fmt"
	"os"
	"slices"
	"strconv"
	"svm/bytecode"
	"unicode"
)

func Assemble(file string) ([]byte, error) {
	// բացել ֆայլը
	input, err := os.Open(file)
	if err != nil {
		return nil, fmt.Errorf("Չհաջողվեց բացել ծրագրի տեքստի ֆայլը։")
	}
	defer input.Close()

	// վերլուծել ծրագիրն ու կառուցել բայթկոդը
	p := &parser{
		source:    bufio.NewReader(input),
		lookahead: lexeme{kind: xUnknown, value: "?"},
		line:      1,
		builder:   bytecode.NewBuilder(),
	}
	err = p.parse()
	if err != nil {
		return nil, err
	}

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
	xNewLine:   "NewLine",
	xColon:     "Colon",
	xLeftBr:    "LeftBracket",
	xRightBr:   "RightBracket",
	xPlus:      "Plus",
	xMinus:     "Minus",
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

func (p *parser) parse() error {
	p.lookahead = p.scanOne()

	p.parseNewLines(optional)

	for !p.has(xEos) {
		err := p.parseLine()
		if err != nil {
			return err
		}
	}

	return nil
}

const (
	mandatory = true
	optional  = false
)

// մեկ կամ ավելի նոր տողին նիշեր
func (p *parser) parseNewLines(first bool) error {
	if first == mandatory && !p.has(xNewLine) {
		return fmt.Errorf("Այստեղ սպասվում է նոր տողի անցման նիշ։")
	}

	for p.has(xNewLine) {
		p.lookahead = p.scanOne()
	}

	return nil
}

// տեքստի մեկ տողի վերլուծությունը
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

	return p.parseNewLines(mandatory)
}

// գործողության ընդհանուր վերլուծություն
func (p *parser) parseOperation() error {
	if !p.has(xOperation) {
		return fmt.Errorf("Սպասվում է հրահանգ, բայց ստացվել է %s", p.lookahead.kind)
	}

	switch p.lookahead.value {
	case "PUSH":
		return p.parsePush()
	case "POP":
		return p.parsePop()
	case "CALL", "JUMP", "JZ":
		return p.parseJump()
	case "HALT", "RET", "ADD", "SUB", "MUL",
		"DIV", "MOD", "NEG", "AND", "OR",
		"NOT", "EQ", "NE", "LT", "LE",
		"GT", "GE", "INPUT", "PRINT":
		return p.parseSimple()
	}

	return nil
}

// PUSH-ը հանդիպում է երկու տեսքով, անմիջական թվային արգումենտով
// և անուղղակի հասցեավորմամբ, օրինակ՝ PUSH [SP+4]
func (p *parser) parsePush() error {
	name, err := p.match(xOperation)
	if err != nil {
		return err
	}
	if name != "PUSH" {
		return fmt.Errorf("Սպասվում է PUSH հրահանգը, բայց ստացվել է %s", name)
	}

	if p.has(xNumber, xPlus, xMinus) {
		number, err := p.parseNumber()
		if err != nil {
			return err
		}
		p.builder.AddWithNumeric(bytecode.Push, int32(number))
	} else if p.has(xLeftBr) {
		register, displacement, err := p.parseIndirect()
		if err != nil {
			return err
		}
		p.builder.AddWithAddress(bytecode.Push, register, displacement)
	}

	return nil
}

// POP-ը հանդիպում է միայն անուղակի հասցեավորմամբ, օրինակ POP [FP-3]
func (p *parser) parsePop() error {
	name, err := p.match(xOperation)
	if err != nil {
		return err
	}
	if name != "POP" {
		return fmt.Errorf("Սպասվում է POP հրահանգը, բայց ստացվել է %s", name)
	}

	if p.has(xLeftBr) {
		register, displacement, err := p.parseIndirect()
		if err != nil {
			return err
		}
		p.builder.AddWithAddress(bytecode.Pop, register, displacement)
		return nil
	}

	return fmt.Errorf("POP հրահանգը սպասում է անուղղակի հասցեավորում")
}

// վերլուծվում են անցում կատարող բոլոր գործողությունները.
// CALL, JUMP, JZ; Դրանց բոլորի արգումենտը պիտակ է
func (p *parser) parseJump() error {
	name, err := p.match(xOperation)
	if err != nil {
		return err
	}
	if name != "CALL" && name != "JUMP" && name != "JZ" {
		return fmt.Errorf("Սպասվում է CALL, JUMP կամ JZ, բայց ստացվել է %s", name)
	}

	label, err := p.match(xIdent)
	if err != nil {
		return err
	}

	p.builder.AddWithLabel(operations[name], label)
	return nil
}

// արգումենտներ չունեցող գործողություններ
func (p *parser) parseSimple() error {
	name, err := p.match(xOperation)
	if err != nil {
		return err
	}
	p.builder.AddBasic(operations[name])
	return nil
}

// ամբողջ թիվ
func (p *parser) parseNumber() (int32, error) {
	var sign int32 = 1
	if p.has(xPlus) {
		p.match(xPlus)
	} else if p.has(xMinus) {
		p.match(xMinus)
		sign = -1
	}

	nlex, err := p.match(xNumber)
	if err != nil {
		return 0, err
	}
	number, _ := strconv.ParseInt(nlex, 10, 32)
	return sign * int32(number), nil
}

// անուղղակի հասցեավորում. '[' REGISTER ('+'|'-') NUMBER ']'
func (p *parser) parseIndirect() (uint16, int16, error) {
	_, err := p.match(xLeftBr)
	if err != nil {
		return 0, 0, err
	}

	regName, err := p.match(xRegister)
	if err != nil {
		return 0, 0, err
	}
	register := registers[regName]

	var displacement int16 = 1
	if p.has(xPlus) {
		p.match(xPlus)
	} else if p.has(xMinus) {
		p.match(xMinus)
		displacement = -1
	} else {
		return 0, 0, fmt.Errorf("Սպասվում է '+' կամ '-' նշանը")
	}

	numStr, err := p.match(xNumber)
	if err != nil {
		return 0, 0, err
	}
	number, _ := strconv.ParseInt(numStr, 10, 16)
	displacement *= int16(number)

	_, err = p.match(xRightBr)
	if err != nil {
		return 0, 0, err
	}

	return register, displacement, nil
}

// պիտակ. IDENT ':'
func (p *parser) parseLabel() error {
	name, err := p.match(xIdent)
	if err != nil {
		return err
	}
	_, err = p.match(xColon)

	p.builder.SetLabel(name)
	return nil
}

func (p *parser) match(expected token) (string, error) {
	if p.has(expected) {
		text := p.lookahead.value
		p.lookahead = p.scanOne()
		return text, nil
	}

	return "", fmt.Errorf("Սպասվում է %s բայց ստացվել է %s", expected, p.lookahead.kind)
}

func (p *parser) has(tokens ...token) bool {
	return slices.Contains(tokens, p.lookahead.kind)
}

func (p *parser) hasValue(values ...string) bool {
	return slices.Contains(values, p.lookahead.value)
}

var metasymbols = map[rune]token{
	':':  xColon,
	'[':  xLeftBr,
	']':  xRightBr,
	'+':  xPlus,
	'-':  xMinus,
	'\n': xNewLine,
}

// բառային վերլուծիչ
func (p *parser) scanOne() lexeme {
	ch := p.readChar()

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

	// հոսքի վերջը
	if ch == 0 {
		return lexeme{kind: xEos, value: "EOS"}
	}

	// գործողության անուն կամ իդենտիֆիկատոր
	if unicode.IsLetter(ch) {
		p.source.UnreadRune()
		text := p.readCharsWhile(isAlphaNumeric)
		if _, exists := operations[text]; exists {
			return lexeme{kind: xOperation, value: text}
		}
		if _, exists := registers[text]; exists {
			return lexeme{kind: xRegister, value: text}
		}

		return lexeme{kind: xIdent, value: text}
	}

	// ամբողջ թիվ
	if unicode.IsDigit(ch) {
		p.source.UnreadRune()
		text := p.readCharsWhile(unicode.IsDigit)
		return lexeme{kind: xNumber, value: text}
	}

	// այլ սիմվոլներ
	if tok, ok := metasymbols[ch]; ok {
		if tok == xNewLine {
			p.line++
		}
		return lexeme{kind: tok, value: string(ch)}
	}

	return lexeme{kind: xUnknown, value: string(ch)}
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
