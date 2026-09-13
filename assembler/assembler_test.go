package assembler

import (
	"bytes"
	"fmt"
	"os"
	"testing"
)

func assembleSource(t *testing.T, source string) ([]byte, error) {
	t.Helper()
	file, err := os.CreateTemp("", "example*.asm")
	if err != nil {
		t.Fatalf("Չկարողացա ստեղծել ֆայլը։ (%v)", err)
	}
	name := file.Name()
	t.Cleanup(func() { os.Remove(name) })

	if _, err := fmt.Fprint(file, source); err != nil {
		file.Close()
		t.Fatalf("Չկարողացա գրել ֆայլը։ (%v)", err)
	}
	if err := file.Close(); err != nil {
		t.Fatalf("Չկարողացա փակել ֆայլը։ (%v)", err)
	}

	return Assemble(name)
}

func TestAssemble(t *testing.T) {
	example0 := `; example 0
	  CALL main
	  HALT
	main:
	  PUSH 1234
	  PRINT
	  RET
	`

	code, err := assembleSource(t, example0)
	if err != nil {
		t.Fatalf("Ասեմբլերի սխալ։ (%v)", err)
	}
	if len(code) == 0 {
		t.Fatal("Ասեմբլերը դատարկ բայթկոդ է վերադարձրել")
	}
}

func TestExample00GoldenBytecode(t *testing.T) {
	code, err := Assemble("../examples/example00.asm")
	if err != nil {
		t.Fatal(err)
	}

	expected := []byte{
		0x83, 0x04, 0x00,
		0x07,
		0x41, 0x09, 0x03, 0x00, 0x00,
		0x09,
		0x41, 0x00, 0x00, 0x00, 0x00,
		0x04,
	}
	if !bytes.Equal(code, expected) {
		t.Fatalf("Սպասվում էր '%v', ստացվել է '%v'", expected, code)
	}
}

func TestExample01IsInvalid(t *testing.T) {
	_, err := Assemble("../examples/example01.asm")
	if err == nil {
		t.Fatal("Սխալ շարահյուսությամբ օրինակը պետք է մերժվեր")
	}

	expected := "ՍԽԱԼ [2]: Տողը սկսվում է NUM<777> սիմվոլով"
	if err.Error() != expected {
		t.Fatalf("Սպասվում էր %q, ստացվել է %q", expected, err)
	}
}

func TestAssembleRejectsUndefinedLabel(t *testing.T) {
	_, err := assembleSource(t, "JUMP missing\n")
	if err == nil {
		t.Fatal("Չսահմանված պիտակը պետք է մերժվեր")
	}
}
