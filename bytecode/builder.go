package bytecode

import (
	"encoding/binary"
	"fmt"
	"io"
)

type instruction struct {
	address   int
	opcode    byte
	immediate int32
	indirect  uint32
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
	var s string
	for e := range i.bytes() {
		s += fmt.Sprintf("%02x ", e)
	}
	return s
}

type Builder struct {
	instructions []*instruction
	labels       map[string]int
	count        int
	offset       int
}

func NewBuilder() *Builder {
	return &Builder{
		instructions: make([]*instruction, 0),
	}
}

func (b *Builder) Store(writer io.Writer) {
	// Temporary
	for i := 0; i < len(b.instructions); i++ {
		bs := b.instructions[i].bytes()
		for j := 0; j < len(bs); j++ {
			fmt.Fprintf(writer, "%02x", bs[j])
		}
	}
	fmt.Fprintln(writer)
}

func (b *Builder) SetLabel(name string) {
}

func (b *Builder) AddBasic(opcode byte) {
	instr := &instruction{
		address:   b.offset,
		opcode:    opcode | Basic,
		immediate: 0,
		indirect:  0,
	}
	b.addInstruction(instr)
}

func (b *Builder) AddWithNumeric(opcode byte, number int32) {
	instr := &instruction{
		address:   b.offset,
		opcode:    opcode | Immediate,
		immediate: number,
		indirect:  0,
	}
	b.addInstruction(instr)
}

func (b *Builder) AddWithAddress(opcode byte, register uint16, displacement int16) {
}

func (b *Builder) AddWithLabel(opcode byte, label string) {
}

func (b *Builder) addInstruction(instr *instruction) {
	instr.address = b.offset
	b.offset += instr.size()
	b.instructions = append(b.instructions, instr)
	b.count++
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
