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
	case bytecode.Nop:
		// դատարկ հրաման, ոչինչ չանել
	case bytecode.Push:
		m.push(mode)
	case bytecode.Pop:
		m.pop()
	case bytecode.Call:
		m.call()
	case bytecode.Ret:
		m.ret()
	case bytecode.Jump:
		m.jump()
	case bytecode.Jz:
		m.jz()
	case bytecode.Input:
		m.input()
	case bytecode.Print:
		m.print()
	case bytecode.Halt:
		return false
	case bytecode.Neg:
		m.negation()
	case bytecode.Not:
		m.not()
	case bytecode.Add:
		m.binary(func(a, b int32) int32 { return a + b })
	case bytecode.Sub:
		m.binary(func(a, b int32) int32 { return a - b })
	case bytecode.Mul:
		m.binary(func(a, b int32) int32 { return a * b })
	case bytecode.Div:
		m.binary(func(a, b int32) int32 { return a / b })
	case bytecode.Mod:
		m.binary(func(a, b int32) int32 { return a % b })
	case bytecode.And:
		m.binary(func(a, b int32) int32 { return a & b })
	case bytecode.Or:
		m.binary(func(a, b int32) int32 { return a | b })
	case bytecode.Eq:
		m.comparison(func(a, b int32) bool { return a == b })
	case bytecode.Ne:
		m.comparison(func(a, b int32) bool { return a != b })
	case bytecode.Lt:
		m.comparison(func(a, b int32) bool { return a < b })
	case bytecode.Le:
		m.comparison(func(a, b int32) bool { return a <= b })
	case bytecode.Gt:
		m.comparison(func(a, b int32) bool { return a > b })
	case bytecode.Ge:
		m.comparison(func(a, b int32) bool { return a >= b })
	default:
		panic("Սխալ (անծանոթ) գործողության կոդ։")
	}

	return true
}

func (m *Machine) push(mode byte) {
	var value int32
	switch mode {
	case bytecode.Immediate: // անմիջական արժեք
		value = m.read(m.ip)
		m.ip += 4
	case bytecode.Indirect: // անուղակի արժեք
		// հարաբերական հասցեն
		raddr := m.readWord(m.ip)
		m.ip += 2
		// բացարձակ հասցեի հաշվելը
		address := m.resolveRelativeAddress(raddr)
		// ստեկում գրելու արժեքը
		value = m.read(address)
	}
	m.basicPush(value)
}

func (m *Machine) pop() {
	// POP-ի հարաբերական հասցեն
	raddr := m.readWord(m.ip)
	m.ip += 2
	// հաշվել բացարձակ հասցեն
	address := m.resolveRelativeAddress(raddr)
	// վերցնել ստեկի գագաթի արժեքն ...
	value := m.basicPop()
	// ... ու գրել որոշված հասցեում
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

func (m *Machine) jump() {
	// JUMP-ի արգումենտը (բացարձակ հասցե)
	address := m.readWord(m.ip)
	// շարունակել address-ից
	m.ip = int16(address)
}

func (m *Machine) jz() {
	// JUMP-ի արգումենտը (բացարձակ հասցե)
	address := m.readWord(m.ip)
	m.ip += 2
	// ստեկի գագաթի արժեքը որպես պայման
	value := m.basicPop()
	if value == 0 {
		m.ip = int16(address)
	}
}

func (m *Machine) input() {
	// կարդալ նշանով ամբողջ թիվ
	var value int32
	fmt.Scanf("%d", &value)
	// գրել ստեկում
	m.basicPush(value)
}

func (m *Machine) print() {
	// վերցնել ստեկի գագաթի արժեքը
	value := m.basicPop()
	// ... արտածել այն
	fmt.Println(value)
}

// բացասում
func (m *Machine) negation() {
	value := m.basicPop()
	m.basicPush(-value)
}

// բիթային ժխտում
func (m *Machine) not() {
	value := m.basicPop()
	m.basicPush(^value)
}

// բինար թվաբանական կամ բիթային գործողություն
func (m *Machine) binary(op func(int32, int32) int32) {
	right := m.basicPop()
	left := m.basicPop()
	result := op(left, right)
	m.basicPush(result)
}

// համեմատման գործողություն
func (m *Machine) comparison(op func(int32, int32) bool) {
	right := m.basicPop()
	left := m.basicPop()
	var result int32
	if op(left, right) {
		result = 1
	}
	m.basicPush(result)
}

// տարրական ստեկային գործողություն push
func (m *Machine) basicPush(value int32) {
	m.write(m.sp, value)
	m.sp += 4
}

// տարրական ստեկային գործողություն pop
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
