package assembler

import (
	"bufio"
	"bytes"
	"strings"
	"testing"
)

func createParserFor(example string) *parser {
	return createParser(bufio.NewReader(strings.NewReader(example)))
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
	if err := p.parse(); err != nil {
		t.Fatalf("Վերլուծման անսպասելի սխալ։ (%v)", err)
	}
	if err := p.builder.Validate(); err != nil {
		t.Fatalf("Բայթկոդի կառուցման ստուգումը ձախողվեց։ (%v)", err)
	}

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
	example0 := `; syntax error
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
		t.Errorf("Սպասվում է \"%s\" հաղորդագրությունը\n", expected0)
	}
}

func TestParserAcceptsCompleteLinesAtEndOfFile(t *testing.T) {
	for _, source := range []string{
		"NOP",
		"_main: NOP",
		"PUSH -2147483648",
		"PUSH 2147483647",
	} {
		p := createParserFor(source)
		if err := p.parse(); err != nil {
			t.Errorf("%q ծրագիրը չպետք է մերժվեր։ (%v)", source, err)
		}
	}
}

func TestParserRejectsInvalidOperands(t *testing.T) {
	tests := []struct {
		name   string
		source string
	}{
		{name: "push առանց արգումենտի", source: "PUSH\n"},
		{name: "pop առանց արգումենտի", source: "POP\n"},
		{name: "call առանց պիտակի", source: "CALL\n"},
		{name: "մեծ ամբողջ թիվ", source: "PUSH 2147483648\n"},
		{name: "մեծ դրական շեղում", source: "PUSH [FP + 8192]\n"},
		{name: "մեծ բացասական շեղում", source: "PUSH [FP - 8193]\n"},
		{name: "չփակված հասցե", source: "PUSH [FP + 1\n"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := createParserFor(tt.source)
			if err := p.parse(); err == nil {
				t.Fatalf("%q ծրագիրը պետք է մերժվեր", tt.source)
			}
		})
	}
}

func TestParserReportsDuplicateLabelDefinitionLine(t *testing.T) {
	p := createParserFor("main:\nmain:\n")
	err := p.parse()
	if err == nil {
		t.Fatal("Կրկնված պիտակը պետք է մերժվեր")
	}

	expected := "ՍԽԱԼ [2]: 'main' պիտակն արդեն սահմանված է 1 տողում"
	if err.Error() != expected {
		t.Fatalf("Սպասվում էր %q, ստացվել է %q", expected, err)
	}
}
