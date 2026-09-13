package assembler

import (
	"bufio"
	"fmt"
	"os"
)

func Assemble(file string) ([]byte, error) {
	// բացել ֆայլը
	input, err := os.Open(file)
	if err != nil {
		return nil, fmt.Errorf("Չհաջողվեց բացել ծրագրի տեքստի ֆայլը։")
	}
	defer input.Close()

	// վերլուծել ծրագիրն ու կառուցել բայթկոդը
	p := createParser(bufio.NewReader(input))
	err = p.parse()
	if err != nil {
		return nil, err
	}

	if err := p.builder.Validate(); err != nil { // լուծել անորոշ հղումները
		return nil, err
	}

	return p.builder.Bytes(), nil
}
