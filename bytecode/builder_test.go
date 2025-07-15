package bytecode

import (
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
	builder.Store(os.Stdout)
}
