package bytecode

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
)

type instruction struct {
	address   int    // հասցե
	opcode    byte   // կոդը և տեսակը
	immediate int32  // թվային արգումենտ
	indirect  uint16 // անուղղակի հասցե
}

func (i *instruction) size() int {
	var value int = 1
	switch i.opcode & 0xC0 {
	case Immediate:
		value += 4
	case Indirect:
		value += 2
	}
	return value
}

func (i *instruction) bytes() []byte {
	result := make([]byte, i.size())
	result[0] = i.opcode
	switch i.opcode & 0xC0 {
	case Immediate:
		binary.LittleEndian.PutUint32(result[1:], uint32(i.immediate))
	case Indirect:
		binary.LittleEndian.PutUint16(result[1:], uint16(i.indirect))
	}
	return result
}

func (i instruction) String() string {
	str := fmt.Sprintf("%04x ", i.address)
	for _, b := range i.bytes() {
		str += fmt.Sprintf("%02x ", b)
	}
	return str
}

type Builder struct {
	instructions []*instruction // հրամանների ցուցակ
	count        int            // հրամանների հաշվիչ

	labels     map[string]int          // պիտակներ, ժամանակավոր
	unresolved map[*instruction]string // ժամանակավորապես անհասցե պիտակներ
	offset     int                     // ընթացիկ շեղումը 0-ից
}

func NewBuilder() *Builder {
	return &Builder{
		instructions: make([]*instruction, 0),
		labels:       make(map[string]int),
		unresolved:   make(map[*instruction]string),
	}
}

func (b *Builder) Bytes() []byte {
	var buffer bytes.Buffer
	for _, instr := range b.instructions {
		buffer.Write(instr.bytes())
	}
	return buffer.Bytes()
}

func (b *Builder) SetLabel(name string) {
	if _, exists := b.labels[name]; !exists {
		b.labels[name] = b.offset
	}
}

func (b *Builder) AddBasic(opcode byte) {
	instr := &instruction{}
	instr.opcode = opcode | Basic
	b.addInstruction(instr)
}

func (b *Builder) AddWithNumeric(opcode byte, number int32) {
	instr := &instruction{}
	instr.opcode = opcode | Immediate
	instr.immediate = number
	b.addInstruction(instr)
}

func (b *Builder) AddWithAddress(opcode byte, register uint16, displacement int16) {
	instr := &instruction{}
	instr.opcode = opcode | Indirect
	instr.indirect = register | uint16(displacement)
	b.addInstruction(instr)
}

func (b *Builder) AddWithLabel(opcode byte, label string) {
	instr := &instruction{}
	instr.opcode = opcode | Indirect
	b.unresolved[instr] = label
	b.addInstruction(instr)
}

func (b *Builder) addInstruction(instr *instruction) {
	instr.address = b.offset
	b.offset += instr.size()
	b.instructions = append(b.instructions, instr)
	b.count++
}

func (b *Builder) Validate() bool {
	// լրացնել անորոշ հղումները
	for instr, label := range b.unresolved {
		instr.indirect = uint16(b.labels[label])
	}
	return true
}

// func (b *Builder) PushI(number int32) {}
// func (b *Builder) PushA(raddr int16)  {}
// func (b *Builder) PopA(raddr int16)   {}
// func (b *Builder) Call(name string)   {}
// func (b *Builder) Ret()               {}
// func (b *Builder) Jump(name string)   {}
// func (b *Builder) Jz(name string)     {}
// func (b *Builder) Halt()              {}
// func (b *Builder) Input()             {}
// func (b *Builder) Print()             {}
// func (b *Builder) Add()               {}
// func (b *Builder) Sub()               {}
// func (b *Builder) Mul()               {}
// func (b *Builder) Div()               {}
// func (b *Builder) Mod()               {}
// func (b *Builder) Neg()               {}
// func (b *Builder) And()               {}
// func (b *Builder) Or()                {}
// func (b *Builder) Not()               {}
// func (b *Builder) Eq()                {}
// func (b *Builder) Ne()                {}
// func (b *Builder) Lt()                {}
// func (b *Builder) Le()                {}
// func (b *Builder) Gt()                {}
// func (b *Builder) Ge()                {}

func (b *Builder) Dump(writer io.Writer) {
	for _, instr := range b.instructions {
		fmt.Fprintln(writer, instr.String())
	}
}
