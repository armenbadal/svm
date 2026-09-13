package bytecode

import (
	"encoding/binary"
	"fmt"
	"strings"
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
	strs := []string{fmt.Sprintf("%04x", i.address)}
	for _, b := range i.bytes() {
		strs = append(strs, fmt.Sprintf("%02x", b))
	}
	return strings.Join(strs, " ")
}
