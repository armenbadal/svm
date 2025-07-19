package assembler

import (
	"bufio"
	"fmt"
	"os"
	"svm/bytecode"
)

func Assemble(file string) ([]byte, error) {
	// բացել ֆայլը
	input, err := os.Open(file)
	if err != nil {
		return nil, fmt.Errorf("Չհաջողվեց բացել ծրագրի տեքստի ֆայլը։")
	}
	defer input.Close()

	// վերլուծել ծրագիրն ու կառուցել բայթկոդը
	p := &parser{
		sc: &scanner{
			source: bufio.NewReader(input),
			line:   1,
		},
		builder: bytecode.NewBuilder(),
	}
	err = p.parse()
	if err != nil {
		return nil, err
	}

	p.builder.Validate() // լուծել անորոշ հղումները

	return p.builder.Bytes(), nil
}
