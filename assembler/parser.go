package assembler

import (
	"bufio"
	"fmt"
	"slices"
	"strconv"
	"svm/bytecode"
)

var operations = map[string]bytecode.Operation{
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

var registers = map[string]bytecode.Register{
	"IP": bytecode.InstructionPointer,
	"SP": bytecode.StackPointer,
	"FP": bytecode.FramePointer,
}

type parser struct {
	sc        *scanner
	lookahead lexeme
	symbols   map[string]int

	builder *bytecode.Builder
}

func createParser(source *bufio.Reader) *parser {
	return &parser{
		sc: &scanner{
			source: source,
			line:   1,
		},
		symbols: make(map[string]int),
		builder: bytecode.NewBuilder(),
	}
}

func (p *parser) parse() error {
	p.lookahead = p.sc.scanOne()

	for p.has(xNewLine) {
		p.match(xNewLine)
	}

	for !p.has(xEos) {
		err := p.parseLine()
		if err != nil {
			return err
		}
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

	if p.has(xNewLine) {
		p.match(xNewLine)
		return nil
	}
	if p.has(xEos) {
		return nil
	}

	return p.report("Տողը սկսվում է %s սիմվոլով", p.lookahead)
}

// պիտակ. IDENT ':'
func (p *parser) parseLabel() error {
	line := p.lookahead.line
	name, err := p.match(xIdent)
	if err != nil {
		return err
	}
	_, err = p.match(xColon)
	if err != nil {
		return p.report("'%s' պիտակին պետք է հետևի ':'", name)
	}

	if place, exists := p.symbols[name]; exists {
		return p.reportAt(line, "'%s' պիտակն արդեն սահմանված է %d տողում", name, place)
	}
	p.symbols[name] = line

	if err := p.builder.SetLabel(name); err != nil {
		return p.reportAt(line, "%v", err)
	}
	return nil
}

// գործողության ընդհանուր վերլուծություն
func (p *parser) parseOperation() error {
	if !p.has(xOperation) {
		return p.report("Սպասվում է հրահանգ, բայց ստացվել է %s", p.lookahead)
	}

	switch p.lookahead.value {
	case "PUSH":
		return p.parsePush()
	case "POP":
		return p.parsePop()
	case "CALL", "JUMP", "JZ":
		return p.parseJumping()
	case "NOP", "HALT", "RET", "ADD", "SUB", "MUL",
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
		return p.report("Սպասվում է PUSH հրահանգը, բայց ստացվել է %s", name)
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
		if err := p.builder.AddWithAddress(bytecode.Push, register, displacement); err != nil {
			return p.report("%v", err)
		}
	} else {
		return p.report("PUSH հրահանգը սպասում է թիվ կամ անուղղակի հասցեավորում")
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
		return p.report("Սպասվում է POP հրահանգը, բայց ստացվել է %s", name)
	}

	if p.has(xLeftBr) {
		register, displacement, err := p.parseIndirect()
		if err != nil {
			return err
		}
		if err := p.builder.AddWithAddress(bytecode.Pop, register, displacement); err != nil {
			return p.report("%v", err)
		}
		return nil
	}

	return p.report("POP հրահանգը սպասում է անուղղակի հասցեավորում")
}

// վերլուծվում են անցում կատարող բոլոր գործողությունները.
// CALL, JUMP, JZ; Դրանց բոլորի արգումենտը պիտակ է
func (p *parser) parseJumping() error {
	name, err := p.match(xOperation)
	if err != nil {
		return err
	}
	if name != "CALL" && name != "JUMP" && name != "JZ" {
		return p.report("Սպասվում է CALL, JUMP կամ JZ, բայց ստացվել է %s", name)
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
	line := p.lookahead.line
	sign := ""
	if p.has(xPlus) {
		p.match(xPlus)
	} else if p.has(xMinus) {
		p.match(xMinus)
		sign = "-"
	}

	nlex, err := p.match(xNumber)
	if err != nil {
		return 0, err
	}
	number, err := strconv.ParseInt(sign+nlex, 10, 32)
	if err != nil {
		return 0, p.reportAt(line, "'%s%s' թիվը 32-բիթանոց միջակայքից դուրս է", sign, nlex)
	}

	return int32(number), nil
}

// անուղղակի հասցեավորում. '[' REGISTER ('+'|'-') NUMBER ']'
func (p *parser) parseIndirect() (bytecode.Register, int16, error) {
	_, err := p.match(xLeftBr)
	if err != nil {
		return 0, 0, err
	}

	regName, err := p.match(xRegister)
	if err != nil {
		return 0, 0, err
	}
	register := registers[regName]

	line := p.lookahead.line
	var sign int64 = 1
	if p.has(xPlus) {
		p.match(xPlus)
	} else if p.has(xMinus) {
		p.match(xMinus)
		sign = -1
	} else {
		return 0, 0, p.report("Սպասվում է '+' կամ '-' նշանը")
	}

	numStr, err := p.match(xNumber)
	if err != nil {
		return 0, 0, err
	}
	number, err := strconv.ParseInt(numStr, 10, 32)
	if err != nil {
		return 0, 0, p.reportAt(line, "Շեղումը թույլատրելի միջակայքից դուրս է")
	}
	displacement := sign * number
	if displacement < -8192 || displacement > 8191 {
		return 0, 0, p.reportAt(line, "Շեղումը պետք է լինի [-8192, 8191] միջակայքում")
	}

	_, err = p.match(xRightBr)
	if err != nil {
		return 0, 0, err
	}

	return register, int16(displacement), nil
}

func (p *parser) match(expected token) (string, error) {
	if p.has(expected) {
		text := p.lookahead.value
		p.lookahead = p.sc.scanOne()
		return text, nil
	}

	return "", p.report("Սպասվում է %s բայց ստացվել է %s", expected, p.lookahead)
}

func (p *parser) has(tokens ...token) bool {
	return slices.Contains(tokens, p.lookahead.kind)
}

func (p *parser) hasValue(values ...string) bool {
	return slices.Contains(values, p.lookahead.value)
}

func (p *parser) report(format string, args ...any) error {
	line := p.lookahead.line
	if line == 0 {
		line = p.sc.line
	}
	return p.reportAt(line, format, args...)
}

func (p *parser) reportAt(line int, format string, args ...any) error {
	return fmt.Errorf("ՍԽԱԼ [%d]: %s", line, fmt.Sprintf(format, args...))
}
