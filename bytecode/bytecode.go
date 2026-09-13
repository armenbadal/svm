package bytecode

import "fmt"

type Operation byte

const (
	Nop Operation = iota
	Push
	Pop
	Call
	Ret
	Jump
	Jz
	Halt
	Input
	Print
	Add
	Sub
	Mul
	Div
	Mod
	Neg
	And
	Or
	Not
	Eq
	Ne
	Lt
	Le
	Gt
	Ge
)

const (
	Basic     byte = 0x00
	Immediate byte = 0x40
	Indirect  byte = 0x80
)

type RelativeAddress uint16

type Register uint16

const (
	StackPointer       Register = 0x4000
	FramePointer       Register = 0x8000
	InstructionPointer Register = 0xC000
)

const displacementMask uint16 = 0x3FFF

func EncodeRelativeAddress(register Register, displacement int16) (RelativeAddress, error) {
	if register != StackPointer && register != FramePointer && register != InstructionPointer {
		return 0, fmt.Errorf("անվավեր ռեգիստր՝ 0x%04x", uint16(register))
	}
	if displacement < -8192 || displacement > 8191 {
		return 0, fmt.Errorf("շեղումը պետք է լինի [-8192, 8191] միջակայքում՝ %d", displacement)
	}

	encoded := uint16(register) | (uint16(displacement) & displacementMask)
	return RelativeAddress(encoded), nil
}

func DecodeRelativeAddress(address RelativeAddress) (Register, int16) {
	encoded := uint16(address)
	register := Register(encoded &^ displacementMask)
	displacement := int16(encoded<<2) >> 2
	return register, displacement
}
