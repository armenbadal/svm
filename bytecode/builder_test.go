package bytecode

import (
	"bytes"
	"os"
	"testing"
)

func TestNewBuilder(t *testing.T) {
	builder := NewBuilder()
	builder.AddBasic(Add)
	builder.AddWithNumeric(Push, 0x7fffffff)
	builder.AddWithNumeric(Push, 0x11111111)
	builder.AddBasic(Sub)
	builder.AddBasic(Mul)
	bc := builder.Bytes()

	expected := []byte{0x0a, 0x41, 0xff, 0xff, 0xff, 0x7f, 0x41, 0x11, 0x11, 0x11, 0x11, 0x0b, 0x0c}
	if !bytes.Equal(expected, bc) {
		t.Errorf("Սպասվում էր '%v', ստացվել է '%v'", expected, bc)
	}
}

func TestLabledInstructions(t *testing.T) {
	builder := NewBuilder()
	builder.SetLabel("start")
	builder.AddWithLabel(Jump, "end")
	builder.AddBasic(Nop)
	builder.AddWithLabel(Jz, "start")
	builder.SetLabel("end")
	builder.AddBasic(Halt)
	builder.Validate()
	builder.Dump(os.Stdout)
}
