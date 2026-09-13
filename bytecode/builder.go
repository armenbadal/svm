package bytecode

import (
	"bytes"
	"fmt"
	"io"
)

type Builder struct {
	instructions []*instruction // հրամանների ցուցակ
	count        int            // հրամանների հաշվիչ

	labels     map[string]int        // պիտակներ, ժամանակավոր
	unresolved []unresolvedReference // ժամանակավորապես անհասցե պիտակներ
	offset     int                   // ընթացիկ շեղումը 0-ից
}

type unresolvedReference struct {
	instruction *instruction
	label       string
}

func NewBuilder() *Builder {
	return &Builder{
		instructions: make([]*instruction, 0),
		labels:       make(map[string]int),
		unresolved:   make([]unresolvedReference, 0),
	}
}

func (b *Builder) Bytes() []byte {
	var buffer bytes.Buffer
	for _, instr := range b.instructions {
		buffer.Write(instr.bytes())
	}
	return buffer.Bytes()
}

func (b *Builder) SetLabel(name string) error {
	if name == "" {
		return fmt.Errorf("պիտակի անունը դատարկ է")
	}
	if _, exists := b.labels[name]; exists {
		return fmt.Errorf("'%s' պիտակն արդեն սահմանված է", name)
	}
	b.labels[name] = b.offset
	return nil
}

func (b *Builder) AddBasic(opcode Operation) {
	instr := &instruction{}
	instr.opcode = byte(opcode) | Basic
	b.addInstruction(instr)
}

func (b *Builder) AddWithNumeric(opcode Operation, number int32) {
	instr := &instruction{}
	instr.opcode = byte(opcode) | Immediate
	instr.immediate = number
	b.addInstruction(instr)
}

func (b *Builder) AddWithAddress(opcode Operation, register Register, displacement int16) error {
	address, err := EncodeRelativeAddress(register, displacement)
	if err != nil {
		return err
	}
	instr := &instruction{}
	instr.opcode = byte(opcode) | Indirect
	instr.indirect = uint16(address)
	b.addInstruction(instr)
	return nil
}

func (b *Builder) AddWithLabel(opcode Operation, label string) {
	instr := &instruction{}
	instr.opcode = byte(opcode) | Indirect
	b.unresolved = append(b.unresolved, unresolvedReference{
		instruction: instr,
		label:       label,
	})
	b.addInstruction(instr)
}

func (b *Builder) addInstruction(instr *instruction) {
	instr.address = b.offset
	b.offset += instr.size()
	b.instructions = append(b.instructions, instr)
	b.count++
}

func (b *Builder) Validate() error {
	if b.offset > 1<<16 {
		return fmt.Errorf("բայթկոդի չափը գերազանցում է 16-բիթանոց հասցեների տիրույթը՝ %d բայթ", b.offset)
	}

	// լրացնել անորոշ հղումները
	for _, reference := range b.unresolved {
		address, exists := b.labels[reference.label]
		if !exists {
			return fmt.Errorf("'%s' պիտակը սահմանված չէ", reference.label)
		}
		if address > 0xFFFF {
			return fmt.Errorf("'%s' պիտակի հասցեն չի տեղավորվում 16 բիթում՝ %d", reference.label, address)
		}
		reference.instruction.indirect = uint16(address)
	}
	return nil
}

func (b *Builder) Dump(writer io.Writer) {
	for _, instr := range b.instructions {
		fmt.Fprintln(writer, instr.String())
	}
}
