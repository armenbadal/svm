package bytecode

const (
	Nop byte = iota
	Push
	Pop
	Call
	Return
	Jump
	Jz
	Halt
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
	InstructionPointer uint16 = 0x0000
	StackPointer       uint16 = 0x4000
	FramePointer       uint16 = 0x8000
)

type Operation = byte
type Integer = int32
type RelativeAddress = uint16

type Instruction struct {
	opcode    Operation
	immediate Integer
	indirect  RelativeAddress
}
