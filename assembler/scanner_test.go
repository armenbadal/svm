package assembler

import (
	"bufio"
	"strings"
	"testing"
	"unicode"
)

func TestReadChar(t *testing.T) {
	s := &scanner{
		source: bufio.NewReader(strings.NewReader(" a + 3\n")),
	}

	if ch := s.readChar(); ch != ' ' {
		t.Errorf("Expected ' ' got '%c' (%v)", ch, ch)
	}

	if ch := s.readChar(); ch != 'a' {
		t.Errorf("Expected 'a' got '%c' (%v)", ch, ch)
	}

	s.source.UnreadRune()
	if ch := s.readChar(); ch != 'a' {
		t.Errorf("Expected 'a' got '%c' (%v)", ch, ch)
	}
}

func TestReadCharsWhile(t *testing.T) {
	s := &scanner{
		source: bufio.NewReader(strings.NewReader("; example\nPUSH 777\nHALT ; end\n")),
	}

	if text := s.readCharsWhile(func(c rune) bool { return c != '\n' }); text != "; example" {
		t.Errorf("Expected '; example' got '%s'", text)
	}

	if ch := s.readChar(); ch != '\n' {
		t.Errorf("Expected 'new-line' got (%v)", ch)
	}
	if text := s.readCharsWhile(unicode.IsLetter); text != "PUSH" {
		t.Errorf("Expected 'PUSH' got '%s'", text)
	}
	if ch := s.readChar(); ch != ' ' {
		t.Errorf("Expected ' ' got (%v)", ch)
	}
}

func TestScanOne(t *testing.T) {
	s := &scanner{
		source: bufio.NewReader(strings.NewReader("PUSH [FP - 12]\n")),
		line:   1,
	}

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
		lex := s.scanOne()
		if !(lex.kind == expected0[i].kind && lex.value == expected0[i].value) {
			t.Errorf("Expected %v, got %v", expected0[i], lex)
		}
		if lex.kind == xEos {
			break
		}
		i += 1
	}
}

func TestScanOneTracksTokenLines(t *testing.T) {
	s := &scanner{
		source: bufio.NewReader(strings.NewReader("NOP\n; comment\nPUSH 1")),
		line:   1,
	}

	expected := []struct {
		kind token
		line int
	}{
		{kind: xOperation, line: 1},
		{kind: xNewLine, line: 1},
		{kind: xNewLine, line: 2},
		{kind: xOperation, line: 3},
		{kind: xNumber, line: 3},
		{kind: xEos, line: 3},
	}

	for i, want := range expected {
		got := s.scanOne()
		if got.kind != want.kind || got.line != want.line {
			t.Fatalf("%d-րդ լեքսեմի համար սպասվում էր %v՝ %d տողում, ստացվել է %v՝ %d տողում", i, want.kind, want.line, got.kind, got.line)
		}
	}
}
