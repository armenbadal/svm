package machine

import (
	"encoding/binary"
	"fmt"
	"svm/bytecode"
)

const MemorySize = 1024 * 16

// մեքենայի մոդելը
type Machine struct {
	memory []byte // հիշողություն
	ip     int16  // հրամանների ցուցիչ (հաշվիչ)
	sp     int16  // ստեկի գագաթի ցուցիչ
	fp     int16  // կանչի ակտիվացման կադրի ցուցիչ
}

// ստեղծել նոր մեքենա
func NewMachine() *Machine {
	return &Machine{
		memory: make([]byte, MemorySize),
		ip:     0,
		sp:     0,
		fp:     0,
	}
}

// ծրագիրը բեռնել հիշողության մեջ
func (m *Machine) Load(data []byte) {
	size := int16(len(data))
	copy(m.memory, data)
	m.sp = size + 1 // ստեկի ցուցիչը դնել ծրագրի ավարտից հետո
}

func (m *Machine) Run() {
	for m.step() {
	}
}

// մեքենայի մեկ քայլը
func (m *Machine) step() bool {
	command := m.memory[m.ip]
	m.ip++
	mode := command & 0xC0
	opcode := command & 0x3F
	switch opcode {
	case bytecode.Push:
		m.push(mode)
	case bytecode.Pop:
		m.pop()
	case bytecode.Call:
		m.call()
	case bytecode.Return:
	case bytecode.Input:
	case bytecode.Print:
		value := m.basicPop()
		fmt.Println(value)
	case bytecode.Halt:
		return false
	default:
	}

	return true
}

func (m *Machine) push(mode byte) {
	var value int32
	switch mode {
	case bytecode.Immediate:
		value = m.read(m.ip)
		m.ip += 4
	case bytecode.Indirect:
		raddr := m.readWord(m.ip)
		m.ip += 2
		address := m.resolveRelativeAddress(raddr)
		value = m.read(address)
	}
	m.basicPush(value)
}

func (m *Machine) pop() {
	raddr := m.readWord(m.ip)
	m.ip += 2
	address := m.resolveRelativeAddress(raddr)
	value := m.basicPop()
	m.write(address, value)
}

func (m *Machine) call() {
	// CALL-ի արգումենտը (բացարձակ հասցե)
	address := m.readWord(m.ip)
	m.ip += 2
	// հիշել IP-ը վերադառնալու համար
	m.basicPush(int32(m.ip))
	// հիշել ընթացիկ FP-ը
	m.basicPush(int32(m.fp))
	// փոխել FP-ը
	m.fp = m.sp
	// շարունակել address-ից
	m.ip = int16(address)
}

func (m *Machine) ret() {
	// ֆունկցիայի արժեքը
	value := m.basicPop()
	// վերականգնել ստեկի ցուցիչը
	m.sp = m.fp
	// վերականգնել ակտիվ կադրի ցուցիչը
	m.fp = int16(m.basicPop())
	// հաջորդ հրամանի հասցեն
	m.ip = int16(m.basicPop())
	// ստեկի գագաթին թողնել ֆունկցիայի արժեքը
	m.basicPush(value)
}

func (m *Machine) basicPush(value int32) {
	m.write(m.sp, value)
	m.sp += 4
}

func (m *Machine) basicPop() int32 {
	m.sp -= 4
	return m.read(m.sp)
}

func (m *Machine) resolveRelativeAddress(relative uint16) int16 {
	address := int16(relative<<2) >> 2
	register := relative & 0xC000
	switch register {
	case bytecode.InstructionPointer:
		address += m.ip
	case bytecode.StackPointer:
		address += m.sp
	case bytecode.FramePointer:
		address += m.fp
	}
	return address
}

func (m *Machine) readWord(addr int16) uint16 {
	return binary.LittleEndian.Uint16(m.memory[addr:])
}

func (m *Machine) read(addr int16) int32 {
	return int32(binary.LittleEndian.Uint32(m.memory[addr:]))
}

func (m *Machine) write(addr int16, value int32) {
	binary.LittleEndian.PutUint32(m.memory[addr:], uint32(value))
}
