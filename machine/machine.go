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

	programLength int // բեռնված ծրագրի չափը
	stackBase     int16
	loaded        bool
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
func (m *Machine) Load(data []byte) error {
	if len(data) == 0 {
		return fmt.Errorf("ծրագիրը դատարկ է")
	}
	if len(data) > len(m.memory) {
		return fmt.Errorf("ծրագրի չափը գերազանցում է հիշողության չափը՝ %d > %d", len(data), len(m.memory))
	}

	clear(m.memory)
	m.programLength = len(data)
	copy(m.memory, data)
	m.ip = 0
	m.sp = int16(m.programLength) // ստեկի ցուցիչը դնել ծրագրի ավարտից անմիջապես հետո
	m.fp = 0
	m.stackBase = m.sp
	m.loaded = true
	return nil
}

func (m *Machine) Run() {
	if !m.loaded {
		panic("Ծրագիր բեռնված չէ")
	}
	for m.step() {
	}
}

// մեքենայի մեկ քայլը
func (m *Machine) step() bool {
	if m.ip < 0 || int(m.ip) >= m.programLength {
		panic("Կատարման հասցեն ծրագրի տիրույթից դուրս է")
	}
	command := m.memory[m.ip]
	m.ip++
	mode := command & 0xC0
	opcode := bytecode.Operation(command & 0x3F)
	m.validateInstruction(opcode, mode)
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
		m.binary(func(a, b int32) int32 {
			if b == 0 {
				panic("Բաժանում զրոյի վրա")
			}
			return a / b
		})
	case bytecode.Mod:
		m.binary(func(a, b int32) int32 {
			if b == 0 {
				panic("Մնացորդի հաշվում զրոյի վրա բաժանելիս")
			}
			return a % b
		})
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

func (m *Machine) validateInstruction(opcode bytecode.Operation, mode byte) {
	var expectedMode byte
	switch opcode {
	case bytecode.Push:
		if mode != bytecode.Immediate && mode != bytecode.Indirect {
			panic("PUSH հրահանգի հասցեավորման անվավեր եղանակ")
		}
		return
	case bytecode.Pop, bytecode.Call, bytecode.Jump, bytecode.Jz:
		expectedMode = bytecode.Indirect
	case bytecode.Nop, bytecode.Ret, bytecode.Halt, bytecode.Input, bytecode.Print,
		bytecode.Add, bytecode.Sub, bytecode.Mul, bytecode.Div, bytecode.Mod,
		bytecode.Neg, bytecode.And, bytecode.Or, bytecode.Not, bytecode.Eq,
		bytecode.Ne, bytecode.Lt, bytecode.Le, bytecode.Gt, bytecode.Ge:
		expectedMode = bytecode.Basic
	default:
		panic("Սխալ (անծանոթ) գործողության կոդ։")
	}

	if mode != expectedMode {
		panic("Հրամանի հասցեավորման անվավեր եղանակ")
	}
}

func (m *Machine) push(mode byte) {
	var value int32
	switch mode {
	case bytecode.Immediate: // անմիջական արժեք
		value = m.readCode(m.ip)
		m.ip += 4
	case bytecode.Indirect: // անուղղակի արժեք
		// հարաբերական հասցեն
		raddr := m.readCodeWord(m.ip)
		m.ip += 2
		// բացարձակ հասցեի հաշվելը
		address := m.resolveAddress(raddr)
		// ստեկում գրելու արժեքը
		value = m.read(address)
	}
	m.basicPush(value)
}

func (m *Machine) pop() {
	// POP-ի հարաբերական հասցեն
	raddr := m.readCodeWord(m.ip)
	m.ip += 2
	// հաշվել բացարձակ հասցեն
	address := m.resolveAddress(raddr)
	// վերցնել ստեկի գագաթի արժեքն ...
	value := m.basicPop()
	// ... ու գրել որոշված հասցեում
	m.write(address, value)
}

func (m *Machine) call() {
	// CALL-ի արգումենտը (բացարձակ հասցե)
	address := m.readCodeWord(m.ip)
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
	if m.fp < m.stackBase+8 || m.fp > m.sp {
		panic("Ֆունկցիայի կանչի կադրը վնասված է")
	}
	if m.sp-m.fp < 4 {
		panic("Ֆունկցիան վերադարձվող արժեք չունի")
	}
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
	address := m.readCodeWord(m.ip)
	// շարունակել address-ից
	m.ip = int16(address)
}

func (m *Machine) jz() {
	// JUMP-ի արգումենտը (բացարձակ հասցե)
	address := m.readCodeWord(m.ip)
	m.ip += 2
	// ստեկի գագաթի արժեքը որպես պայման
	m.requireOperands(1)
	value := m.basicPop()
	if value == 0 {
		m.ip = int16(address)
	}
}

func (m *Machine) input() {
	// կարդալ նշանով ամբողջ թիվ
	var value int32
	if _, err := fmt.Scanf("%d", &value); err != nil {
		panic("Չհաջողվեց կարդալ ամբողջ թիվ")
	}
	// գրել ստեկում
	m.basicPush(value)
}

func (m *Machine) print() {
	// վերցնել ստեկի գագաթի արժեքը
	m.requireOperands(1)
	value := m.basicPop()
	// ... արտածել այն
	fmt.Println(value)
}

// բացասում
func (m *Machine) negation() {
	m.requireOperands(1)
	value := m.basicPop()
	m.basicPush(-value)
}

// բիթային ժխտում
func (m *Machine) not() {
	m.requireOperands(1)
	value := m.basicPop()
	m.basicPush(^value)
}

// բինար թվաբանական կամ բիթային գործողություն
func (m *Machine) binary(op func(int32, int32) int32) {
	m.requireOperands(2)
	right := m.basicPop()
	left := m.basicPop()
	result := op(left, right)
	m.basicPush(result)
}

// համեմատման գործողություն
func (m *Machine) comparison(op func(int32, int32) bool) {
	m.requireOperands(2)
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
	if int(m.sp)+4 > len(m.memory) {
		panic("Ստեկի գերլցում")
	}
	m.write(m.sp, value)
	m.sp += 4
}

// տարրական ստեկային գործողություն pop
func (m *Machine) basicPop() int32 {
	if m.sp-4 < m.stackBase {
		panic("Դատարկ ստեկից արժեք վերցնելու փորձ")
	}
	m.sp -= 4
	return m.read(m.sp)
}

func (m *Machine) requireOperands(count int16) {
	base := m.stackBase
	if m.fp > base {
		base = m.fp
	}
	if m.sp-base < count*4 {
		panic("Ստեկում բավարար արժեքներ չկան")
	}
}

func (m *Machine) resolveAddress(relative uint16) int16 {
	register, displacement := bytecode.DecodeRelativeAddress(bytecode.RelativeAddress(relative))
	address := displacement
	switch register {
	case bytecode.InstructionPointer:
		address += m.ip
	case bytecode.StackPointer:
		address += m.sp
	case bytecode.FramePointer:
		address += m.fp
	default:
		panic("Հարաբերական հասցեի անվավեր ռեգիստր")
	}
	return address
}

func (m *Machine) readWord(addr int16) uint16 {
	address := int(addr)
	if address < 0 || address+2 > len(m.memory) {
		panic("Հասցեն հիշողության տիրույթից դուրս է")
	}

	return binary.LittleEndian.Uint16(m.memory[address:])
}

func (m *Machine) read(addr int16) int32 {
	address := int(addr)
	if address < 0 || address+4 > len(m.memory) {
		panic("Հասցեն հիշողության տիրույթից դուրս է")
	}

	return int32(binary.LittleEndian.Uint32(m.memory[address:]))
}

func (m *Machine) write(addr int16, value int32) {
	address := int(addr)
	if address < 0 || address+4 > len(m.memory) {
		panic("Հասցեն հիշողության տիրույթից դուրս է")
	}

	binary.LittleEndian.PutUint32(m.memory[address:], uint32(value))
}

func (m *Machine) readCodeWord(addr int16) uint16 {
	if addr < 0 || int(addr)+2 > m.programLength {
		panic("Հրամանի արգումենտը ծրագրի տիրույթից դուրս է")
	}
	return m.readWord(addr)
}

func (m *Machine) readCode(addr int16) int32 {
	if addr < 0 || int(addr)+4 > m.programLength {
		panic("Հրամանի արգումենտը ծրագրի տիրույթից դուրս է")
	}
	return m.read(addr)
}
