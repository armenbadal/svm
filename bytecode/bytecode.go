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

var Codes = []byte{
	Nop,
	Push,
	Pop,
	Call,
	Ret,
	Jump,
	Jz,
	Halt,
	Input,
	Print,
	Add,
	Sub,
	Mul,
	Div,
	Mod,
	Neg,
	And,
	Or,
	Not,
	Eq,
	Ne,
	Lt,
	Le,
	Gt,
	Ge,
}

var Mnemonics = map[byte]string{
	Nop:   "NOP",
	Push:  "PUSH",
	Pop:   "POP",
	Call:  "CALL",
	Ret:   "RET",
	Jump:  "JUMP",
	Jz:    "JZ",
	Halt:  "HALT",
	Add:   "ADD",
	Sub:   "SUB",
	Mul:   "MUL",
	Div:   "DIV",
	Mod:   "MOD",
	Neg:   "NEG",
	And:   "AND",
	Or:    "OR",
	Not:   "NOT",
	Eq:    "EQ",
	Ne:    "NE",
	Lt:    "LT",
	Le:    "LE",
	Gt:    "GT",
	Ge:    "GE",
	Input: "INPUT",
	Print: "PRINT",
}

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
