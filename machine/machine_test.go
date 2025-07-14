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
	program := []byte{
		bytecode.Push | bytecode.Immediate,
		0x01, 0x00, 0x00, 0x00,
		bytecode.Print,
		bytecode.Halt,
	}
	m := NewMachine()
	m.Load(program)
	m.Run()
}
