package assembler

import (
	"bufio"
	"bytes"
	"strings"
	"svm/bytecode"
	"testing"
)

func createParserFor(example string) *parser {
	return &parser{
		sc: &scanner{
			source: bufio.NewReader(strings.NewReader(example)),
			line:   1,
		},
		builder: bytecode.NewBuilder(),
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
