package assembler

import (
	"fmt"
	"os"
	"testing"
)

func TestAssemble(t *testing.T) {
	example0 := `; example 0
	  CALL main
	  HALT
	main:
	  PUSH 1234
	  PRINT
	  RET
	`

	file, err := os.CreateTemp("", "example*.asm")
	if err != nil {
		t.Fatalf("Չկարողացա ստեղծել ֆայլը։ (%v)", err)
	}

	defer file.Close()
	defer os.Remove(file.Name())

	fmt.Fprint(file, example0)

	bytes, err := Assemble(file.Name())
	if err != nil {
		t.Errorf("Ասեմբլերի սխալ։ (%v)", err)
	}

	fmt.Printf("-> %v\n", bytes)
}
