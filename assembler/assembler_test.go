package assembler

import (
	"bufio"
	"strings"
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

func TestScanOne(t *testing.T) {
	example0 := "PUSH [FP - 12]\n"
	p := &parser{
		source: bufio.NewReader(strings.NewReader(example0)),
	}

	expected0 := []lexeme{
		{token: xOperation, value: "PUSH"},
		{token: xLeftBr, value: "["},
		{token: xRegister, value: "FP"},
		{token: xMinus, value: "-"},
		{token: xNumber, value: "12"},
		{token: xRightBr, value: "]"},
		{token: xNewLine, value: "\n"},
		{token: xEos, value: "EOS"},
	}

	i := 0
	for {
		lex := p.scanOne()
		if !(lex.token == expected0[i].token && lex.value == expected0[i].value) {
			t.Errorf("Expected %v, got %v", expected0[i], lex)
		}
		if lex.token == xEos {
			break
		}
		i += 1
	}
}
