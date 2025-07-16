package machine

import (
	"svm/bytecode"
	"testing"
)

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
	builder.AddWithLabel(bytecode.Call, "print777")
	builder.AddBasic(bytecode.Halt)
	builder.SetLabel("print777")
	builder.AddWithNumeric(bytecode.Push, 777)
	builder.AddBasic(bytecode.Print)
	builder.AddBasic(bytecode.Ret)
	builder.Validate()

	program := builder.Bytes()
	m := NewMachine()
	m.Load(program)
	m.Run()
}
