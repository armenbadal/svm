package assembler

import (
	"bufio"
	"unicode"
)

type scanner struct {
	source *bufio.Reader
	line   int
}

var metasymbols = map[rune]token{
	':':  xColon,
	'[':  xLeftBr,
	']':  xRightBr,
	'+':  xPlus,
	'-':  xMinus,
	'\n': xNewLine,
}

// բառային վերլուծիչ
func (s *scanner) scanOne() lexeme {
	ch := s.readChar()

	// անտեսել բացատները
	if isSpace(ch) {
		s.readCharsWhile(isSpace)
		ch = s.readChar()
	}

	// անտեսել մեկնաբանությունները
	if ch == ';' {
		s.readCharsWhile(func(c rune) bool { return c != '\n' })
		ch = s.readChar()
	}

	// հոսքի վերջը
	if ch == 0 {
		return lexeme{kind: xEos, value: "EOS"}
	}

	// գործողության անուն կամ իդենտիֆիկատոր
	if unicode.IsLetter(ch) {
		s.source.UnreadRune()
		text := s.readCharsWhile(isAlphaNumeric)
		if _, exists := operations[text]; exists {
			return lexeme{kind: xOperation, value: text}
		}
		if _, exists := registers[text]; exists {
			return lexeme{kind: xRegister, value: text}
		}

		return lexeme{kind: xIdent, value: text}
	}

	// ամբողջ թիվ
	if unicode.IsDigit(ch) {
		s.source.UnreadRune()
		text := s.readCharsWhile(unicode.IsDigit)
		return lexeme{kind: xNumber, value: text}
	}

	// այլ սիմվոլներ
	if tok, ok := metasymbols[ch]; ok {
		if tok == xNewLine {
			s.line++
		}
		return lexeme{kind: tok, value: string(ch)}
	}

	return lexeme{kind: xUnknown, value: string(ch)}
}

func isAlphaNumeric(c rune) bool {
	return unicode.IsLetter(c) || unicode.IsDigit(c)
}

func isSpace(c rune) bool {
	return c == ' ' || c == '\t' || c == '\r'
}

func (s *scanner) readCharsWhile(pred func(rune) bool) string {
	var text string
	ch := s.readChar()
	for pred(ch) && ch != 0 {
		text += string(ch)
		ch = s.readChar()
	}
	s.source.UnreadRune()
	return text
}

func (s *scanner) readChar() rune {
	ch, _, err := s.source.ReadRune()
	if err != nil {
		return 0
	}
	return ch
}
