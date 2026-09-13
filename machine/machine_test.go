package machine

import (
	"bytes"
	"svm/bytecode"
	"testing"
)

func assertPanics(t *testing.T, action func()) {
	t.Helper()
	defer func() {
		if recover() == nil {
			t.Fatal("Գործողությունը պետք է panic առաջացներ")
		}
	}()
	action()
}

func TestNewMachine(t *testing.T) {
	m := NewMachine()
	if m == nil {
		t.Errorf("Machine ստեղծելը ձախողվեց")
	}
}

func TestBasicPushPop(t *testing.T) {
	m := NewMachine()

	m.basicPush(4)
	v := m.basicPop()
	if v != 4 {
		t.Errorf("Սպասվում է 4, բայց ստացվել է %d", v)
	}

	m.basicPush(-2)
	v = m.basicPop()
	if v != -2 {
		t.Errorf("Սպասվում է -2, բայց ստացվել է %d", v)
	}
}

func TestRun(t *testing.T) {
	builder := bytecode.NewBuilder()
	builder.AddWithLabel(bytecode.Call, "answer")
	builder.AddBasic(bytecode.Halt)
	if err := builder.SetLabel("answer"); err != nil {
		t.Fatal(err)
	}
	builder.AddWithNumeric(bytecode.Push, 777)
	builder.AddBasic(bytecode.Ret)
	if err := builder.Validate(); err != nil {
		t.Fatalf("Պիտակների լուծումը ձախողվեց։ (%v)", err)
	}

	program := builder.Bytes()
	m := NewMachine()
	if err := m.Load(program); err != nil {
		t.Fatal(err)
	}
	m.Run()
	if got := m.basicPop(); got != 777 {
		t.Fatalf("Սպասվում էր վերադարձվող 777 արժեքը, ստացվել է %d", got)
	}
}

func TestLoadRejectsInvalidProgramSize(t *testing.T) {
	m := NewMachine()
	if err := m.Load(nil); err == nil {
		t.Fatal("Դատարկ ծրագիրը պետք է մերժվեր")
	}
	if err := m.Load(make([]byte, MemorySize+1)); err == nil {
		t.Fatal("Հիշողությունից մեծ ծրագիրը պետք է մերժվեր")
	}
}

func TestLoadResetsMachine(t *testing.T) {
	m := NewMachine()
	first := []byte{byte(bytecode.Push) | bytecode.Immediate, 5, 0, 0, 0, byte(bytecode.Halt)}
	if err := m.Load(first); err != nil {
		t.Fatal(err)
	}
	m.Run()

	second := []byte{byte(bytecode.Halt)}
	if err := m.Load(second); err != nil {
		t.Fatal(err)
	}
	if m.ip != 0 || m.fp != 0 || m.sp != int16(len(second)) {
		t.Fatalf("Load-ից հետո ռեգիստրները չեն վերակայվել՝ IP=%d SP=%d FP=%d", m.ip, m.sp, m.fp)
	}
	if !bytes.Equal(m.memory[:len(first)], []byte{byte(bytecode.Halt), 0, 0, 0, 0, 0}) {
		t.Fatal("Load-ը չի մաքրել նախորդ ծրագրի հիշողությունը")
	}
}

func TestRunRequiresLoadedProgram(t *testing.T) {
	assertPanics(t, func() {
		NewMachine().Run()
	})
}

func TestMachineRejectsMalformedPrograms(t *testing.T) {
	programs := map[string][]byte{
		"անհայտ opcode":               {0x3f},
		"PUSH-ի սխալ եղանակ":          {byte(bytecode.Push), byte(bytecode.Halt)},
		"կտրված immediate":            {byte(bytecode.Push) | bytecode.Immediate, 1},
		"կտրված հասցե":                {byte(bytecode.Pop) | bytecode.Indirect, 1},
		"անվավեր հարաբերական ռեգիստր": {byte(bytecode.Push) | bytecode.Indirect, 0, 0, byte(bytecode.Halt)},
		"ստեկի թերալցում":             {byte(bytecode.Add), byte(bytecode.Halt)},
		"անվավեր վերադարձ":            {byte(bytecode.Ret)},
		"ծրագրից դուրս անցում":        {byte(bytecode.Jump) | bytecode.Indirect, 3, 0},
		"զրոյի վրա մնացորդ": {
			byte(bytecode.Push) | bytecode.Immediate, 1, 0, 0, 0,
			byte(bytecode.Push) | bytecode.Immediate, 0, 0, 0, 0,
			byte(bytecode.Mod),
			byte(bytecode.Halt),
		},
	}

	for name, program := range programs {
		t.Run(name, func(t *testing.T) {
			m := NewMachine()
			if err := m.Load(program); err != nil {
				t.Fatal(err)
			}
			assertPanics(t, m.Run)
		})
	}
}

func TestNestedCallsReturnValue(t *testing.T) {
	builder := bytecode.NewBuilder()
	builder.AddWithLabel(bytecode.Call, "outer")
	builder.AddBasic(bytecode.Halt)
	if err := builder.SetLabel("outer"); err != nil {
		t.Fatal(err)
	}
	builder.AddWithLabel(bytecode.Call, "inner")
	builder.AddBasic(bytecode.Ret)
	if err := builder.SetLabel("inner"); err != nil {
		t.Fatal(err)
	}
	builder.AddWithNumeric(bytecode.Push, 42)
	builder.AddBasic(bytecode.Ret)
	if err := builder.Validate(); err != nil {
		t.Fatal(err)
	}

	m := NewMachine()
	if err := m.Load(builder.Bytes()); err != nil {
		t.Fatal(err)
	}
	m.Run()
	if got := m.basicPop(); got != 42 {
		t.Fatalf("Սպասվում էր վերադարձվող 42 արժեքը, ստացվել է %d", got)
	}
}

func TestReturnRequiresValue(t *testing.T) {
	builder := bytecode.NewBuilder()
	builder.AddWithLabel(bytecode.Call, "empty")
	builder.AddBasic(bytecode.Halt)
	if err := builder.SetLabel("empty"); err != nil {
		t.Fatal(err)
	}
	builder.AddBasic(bytecode.Ret)
	if err := builder.Validate(); err != nil {
		t.Fatal(err)
	}

	m := NewMachine()
	if err := m.Load(builder.Bytes()); err != nil {
		t.Fatal(err)
	}
	assertPanics(t, m.Run)
}

func TestResolveNegativeRelativeAddress(t *testing.T) {
	m := NewMachine()
	m.sp = 100
	m.fp = 200
	m.ip = 300

	tests := []struct {
		register bytecode.Register
		want     int16
	}{
		{register: bytecode.StackPointer, want: 97},
		{register: bytecode.FramePointer, want: 197},
		{register: bytecode.InstructionPointer, want: 297},
	}

	for _, tt := range tests {
		address, err := bytecode.EncodeRelativeAddress(tt.register, -3)
		if err != nil {
			t.Fatal(err)
		}
		if got := m.resolveAddress(uint16(address)); got != tt.want {
			t.Fatalf("0x%04x ռեգիստրի համար սպասվում էր %d, ստացվել է %d", tt.register, tt.want, got)
		}
	}
}

func TestFunctionArguments(t *testing.T) {
	builder := bytecode.NewBuilder()
	pushFromFrame := func(displacement int16) {
		t.Helper()
		if err := builder.AddWithAddress(bytecode.Push, bytecode.FramePointer, displacement); err != nil {
			t.Fatal(err)
		}
	}

	builder.AddWithNumeric(bytecode.Push, 10)
	builder.AddWithNumeric(bytecode.Push, 20)
	builder.AddWithLabel(bytecode.Call, "max")
	builder.AddBasic(bytecode.Halt)
	if err := builder.SetLabel("max"); err != nil {
		t.Fatal(err)
	}
	pushFromFrame(-16)
	pushFromFrame(-12)
	builder.AddBasic(bytecode.Gt)
	builder.AddWithLabel(bytecode.Jz, "second")
	pushFromFrame(-16)
	builder.AddWithLabel(bytecode.Jump, "end")
	if err := builder.SetLabel("second"); err != nil {
		t.Fatal(err)
	}
	pushFromFrame(-12)
	if err := builder.SetLabel("end"); err != nil {
		t.Fatal(err)
	}
	builder.AddBasic(bytecode.Ret)
	if err := builder.Validate(); err != nil {
		t.Fatal(err)
	}

	m := NewMachine()
	if err := m.Load(builder.Bytes()); err != nil {
		t.Fatal(err)
	}
	m.Run()
	if got := m.basicPop(); got != 20 {
		t.Fatalf("Սպասվում էր առավելագույն 20 արժեքը, ստացվել է %d", got)
	}
}

func TestRecursiveFactorial(t *testing.T) {
	builder := bytecode.NewBuilder()
	pushArgument := func() {
		t.Helper()
		if err := builder.AddWithAddress(bytecode.Push, bytecode.FramePointer, -12); err != nil {
			t.Fatal(err)
		}
	}

	builder.AddWithNumeric(bytecode.Push, 5)
	builder.AddWithLabel(bytecode.Call, "factorial")
	builder.AddBasic(bytecode.Halt)
	if err := builder.SetLabel("factorial"); err != nil {
		t.Fatal(err)
	}
	pushArgument()
	builder.AddWithNumeric(bytecode.Push, 1)
	builder.AddBasic(bytecode.Le)
	builder.AddWithLabel(bytecode.Jz, "recurse")
	builder.AddWithNumeric(bytecode.Push, 1)
	builder.AddBasic(bytecode.Ret)
	if err := builder.SetLabel("recurse"); err != nil {
		t.Fatal(err)
	}
	pushArgument()
	builder.AddWithNumeric(bytecode.Push, 1)
	builder.AddBasic(bytecode.Sub)
	builder.AddWithLabel(bytecode.Call, "factorial")
	pushArgument()
	builder.AddBasic(bytecode.Mul)
	builder.AddBasic(bytecode.Ret)
	if err := builder.Validate(); err != nil {
		t.Fatal(err)
	}

	m := NewMachine()
	if err := m.Load(builder.Bytes()); err != nil {
		t.Fatal(err)
	}
	m.Run()
	if got := m.basicPop(); got != 120 {
		t.Fatalf("Սպասվում էր 5! = 120 արժեքը, ստացվել է %d", got)
	}
}

func TestStackOverflow(t *testing.T) {
	m := NewMachine()
	m.sp = MemorySize - 3
	assertPanics(t, func() {
		m.basicPush(1)
	})
}
