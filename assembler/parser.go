package assembler

import (
	"fmt"
	"slices"
	"strconv"
	"svm/bytecode"
)

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
	sc        *scanner
	lookahead lexeme

	builder *bytecode.Builder
}

func (p *parser) parse() error {
	p.lookahead = p.sc.scanOne()

	p.parseNewLines()

	for !p.has(xEos) {
		err := p.parseLine()
		if err != nil {
			return err
		}
	}

	return nil
}

// մեկ կամ ավելի նոր տողին նիշեր
func (p *parser) parseNewLines() {
	for p.has(xNewLine) {
		p.lookahead = p.sc.scanOne()
	}
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
		p.parseNewLines()
		return nil
	}

	return p.report("Տողը սկսվում է %s սիմվոլով", p.lookahead)
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
		return p.report("Սպասվում է POP հրահանգը, բայց ստացվել է %s", name)
	}

	if p.has(xLeftBr) {
		register, displacement, err := p.parseIndirect()
		if err != nil {
			return err
		}
		p.builder.AddWithAddress(bytecode.Pop, register, displacement)
		return nil
	}

	return p.report("POP հրահանգը սպասում է անուղղակի հասցեավորում")
}

// վերլուծվում են անցում կատարող բոլոր գործողությունները.
// CALL, JUMP, JZ; Դրանց բոլորի արգումենտը պիտակ է
func (p *parser) parseJump() error {
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
		return 0, 0, p.report("Սպասվում է '+' կամ '-' նշանը")
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
	if err != nil {
		return p.report("'%s' պիտակին պետք է հետևի ':'", name)
	}

	p.builder.SetLabel(name)
	return nil
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
	return fmt.Errorf("ՍԽԱԼ [%d]: %s", p.sc.line, fmt.Sprintf(format, args...))
}
