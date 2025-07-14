package bytecode

const (
	Nop byte = iota
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

const (
	StackPointer       uint16 = 0x4000
	FramePointer       uint16 = 0x8000
	InstructionPointer uint16 = 0xC000
)

type Operation = byte
type Integer = int32
type RelativeAddress = uint16

type Instruction struct {
	opcode    Operation
	immediate Integer
	indirect  RelativeAddress
}
