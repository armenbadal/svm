package bytecode

import (
	"bytes"
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
	if err := builder.SetLabel("start"); err != nil {
		t.Fatal(err)
	}
	builder.AddWithLabel(Jump, "end")
	builder.AddBasic(Nop)
	builder.AddWithLabel(Jz, "start")
	if err := builder.SetLabel("end"); err != nil {
		t.Fatal(err)
	}
	builder.AddBasic(Halt)
	if err := builder.Validate(); err != nil {
		t.Fatalf("Պիտակների լուծումը ձախողվեց։ (%v)", err)
	}

	expected := []byte{
		byte(Jump) | Indirect, 7, 0,
		byte(Nop),
		byte(Jz) | Indirect, 0, 0,
		byte(Halt),
	}
	if got := builder.Bytes(); !bytes.Equal(got, expected) {
		t.Fatalf("Սպասվում էր '%v', ստացվել է '%v'", expected, got)
	}
}

func TestOperationCodesAreStable(t *testing.T) {
	operations := []Operation{
		Nop, Push, Pop, Call, Ret, Jump, Jz, Halt, Input, Print,
		Add, Sub, Mul, Div, Mod, Neg, And, Or, Not, Eq, Ne, Lt,
		Le, Gt, Ge,
	}
	for expected, operation := range operations {
		if byte(operation) != byte(expected) {
			t.Fatalf("%d-րդ գործողության կոդը փոխվել է՝ %d", expected, operation)
		}
	}
}

func TestRelativeAddressRoundTrip(t *testing.T) {
	registers := []Register{StackPointer, FramePointer, InstructionPointer}
	displacements := []int16{-8192, -3, 0, 1, 8191}

	for _, register := range registers {
		for _, displacement := range displacements {
			encoded, err := EncodeRelativeAddress(register, displacement)
			if err != nil {
				t.Fatalf("Չհաջողվեց կոդավորել 0x%04x ռեգիստրը և %d շեղումը։ (%v)", register, displacement, err)
			}
			gotRegister, gotDisplacement := DecodeRelativeAddress(encoded)
			if gotRegister != register || gotDisplacement != displacement {
				t.Fatalf(
					"Սպասվում էր (0x%04x, %d), ստացվել է (0x%04x, %d)",
					register,
					displacement,
					gotRegister,
					gotDisplacement,
				)
			}
		}
	}
}

func TestRelativeAddressRejectsInvalidValues(t *testing.T) {
	if _, err := EncodeRelativeAddress(Register(0), 0); err == nil {
		t.Fatal("Անվավեր ռեգիստրը պետք է մերժվեր")
	}
	if _, err := EncodeRelativeAddress(FramePointer, 8192); err == nil {
		t.Fatal("Մեծ դրական շեղումը պետք է մերժվեր")
	}
	if _, err := EncodeRelativeAddress(FramePointer, -8193); err == nil {
		t.Fatal("Մեծ բացասական շեղումը պետք է մերժվեր")
	}
}

func TestBuilderEncodesNegativeDisplacement(t *testing.T) {
	builder := NewBuilder()
	if err := builder.AddWithAddress(Push, FramePointer, -3); err != nil {
		t.Fatal(err)
	}

	expected := []byte{byte(Push) | Indirect, 0xfd, 0xbf}
	if got := builder.Bytes(); !bytes.Equal(got, expected) {
		t.Fatalf("Սպասվում էր '%v', ստացվել է '%v'", expected, got)
	}
}

func TestBuilderRejectsInvalidLabels(t *testing.T) {
	builder := NewBuilder()
	if err := builder.SetLabel("main"); err != nil {
		t.Fatal(err)
	}
	if err := builder.SetLabel("main"); err == nil {
		t.Fatal("Կրկնված պիտակը պետք է մերժվեր")
	}

	builder.AddWithLabel(Jump, "missing")
	if err := builder.Validate(); err == nil {
		t.Fatal("Չսահմանված պիտակը պետք է մերժվեր")
	}
}
