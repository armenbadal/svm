package assembler

import (
	"bufio"
	"bytes"
	"strings"
	"svm/bytecode"
	"testing"
	"unicode"
)

func TestReadChar(t *testing.T) {
	p := &parser{
		source: bufio.NewReader(strings.NewReader(" a + 3\n")),
	}

	if ch := p.readChar(); ch != ' ' {
		t.Errorf("Expected ' ' got '%c' (%v)", ch, ch)
	}

	if ch := p.readChar(); ch != 'a' {
		t.Errorf("Expected 'a' got '%c' (%v)", ch, ch)
	}

	p.source.UnreadRune()
	if ch := p.readChar(); ch != 'a' {
		t.Errorf("Expected 'a' got '%c' (%v)", ch, ch)
	}
}

func TestReadCharsWhile(t *testing.T) {
	p := &parser{
		source: bufio.NewReader(strings.NewReader("; example\nPUSH 777\nHALT ; end\n")),
	}

	if text := p.readCharsWhile(func(c rune) bool { return c != '\n' }); text != "; example" {
		t.Errorf("Expected '; example' got '%s'", text)
	}

	if ch := p.readChar(); ch != '\n' {
		t.Errorf("Expected 'new-line' got (%v)", ch)
	}
	if text := p.readCharsWhile(unicode.IsLetter); text != "PUSH" {
		t.Errorf("Expected 'PUSH' got '%s'", text)
	}
	if ch := p.readChar(); ch != ' ' {
		t.Errorf("Expected ' ' got (%v)", ch)
	}
}

func createParserFor(example string) *parser {
	return &parser{
		source:  bufio.NewReader(strings.NewReader(example)),
		builder: bytecode.NewBuilder(),
	}
}

func TestScanOne(t *testing.T) {
	p := createParserFor("PUSH [FP - 12]\n")

	expected0 := []lexeme{
		{kind: xOperation, value: "PUSH"},
		{kind: xLeftBr, value: "["},
		{kind: xRegister, value: "FP"},
		{kind: xMinus, value: "-"},
		{kind: xNumber, value: "12"},
		{kind: xRightBr, value: "]"},
		{kind: xNewLine, value: "\n"},
		{kind: xEos, value: "EOS"},
	}

	i := 0
	for {
		lex := p.scanOne()
		if !(lex.kind == expected0[i].kind && lex.value == expected0[i].value) {
			t.Errorf("Expected %v, got %v", expected0[i], lex)
		}
		if lex.kind == xEos {
			break
		}
		i += 1
	}
}

func TestParse(t *testing.T) {
	example0 := `


	; example 0
	  CALL main
	  HALT
	
	main:
	  PUSH 0 ; local
	  PUSH 345
	  POP [FP + 1]
	  PUSH [FP + 1]
	  PRINT
      RET
	
	`

	p := createParserFor(example0)
	p.parse()
	p.builder.Validate()

	buffer := bytes.NewBufferString("")
	p.builder.Dump(buffer)
	generated := buffer.String()

	expected := "0000 83 04 00\n" +
		"0003 07\n" +
		"0004 41 00 00 00 00\n" +
		"0009 41 59 01 00 00\n" +
		"000e 82 01 80\n" +
		"0011 81 01 80\n" +
		"0014 09\n" +
		"0015 04\n"
	if expected != generated {
		t.Errorf("Ստացված բայթկոդը չի հմապատասխանում սպասվածին։\n|%s|\n\n|%s|", expected, generated)
	}
}

func TestErrorHandling(t *testing.T) {
	example0 := `
		; syntax error
		777
		HALT
	`
	p := createParserFor(example0)
	err := p.parse()
	if err == nil {
		t.Errorf("Սպասվում է վերլուծության սխալ")
	}

	expected0 := "ՍԽԱԼ [2]: Տողը սկսվում է NUM<777> սիմվոլով"
	if expected0 != err.Error() {
		t.Errorf("Սպասվում է \"%s\" հաղորդագրությունը", expected0)
	}
}
