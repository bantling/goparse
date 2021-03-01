package lexer

import (
	"io"
	"strings"

	"github.com/bantling/goiter"
)

// LexType is the type of a lexical token
type LexType uint

// LexType constants
const (
	InvalidLexType LexType = iota
	Comment
	Identifier
	String
	CharacterRange
	Repetition
	OptionAST
	OptionEOL
	OptionIndent
	OptionOutdent
	OptionPreEOL
	OptionPreIndent
	OptionPreOutdent
	Hat
	OpenParens
	CloseParens
	Bar
	Comma
	Equals
	DoubleEquals
	SemiColon
	EOF
)

var (
	// map of valid options strings
	optionStrings = []string{":AST", ":EOL", ":INDENT", ":OUTDENT", ":PREEOL", ":PREINDENT", ":PREOUTDENT"}

	// map of useless ASCII control characters
	uselessChars = map[rune]bool{
		'\x00': true,
		'\x01': true,
		'\x02': true,
		'\x03': true,
		'\x04': true,
		'\x05': true,
		'\x06': true,
		'\x07': true,
		'\x08': true,
		// '\x09' is tab
		// '\x0A' is newline
		'\x0B': true,
		'\x0C': true,
		// '\x0D' is return carriage
		'\x0E': true,
		'\x0F': true,
		'\x10': true,
		'\x11': true,
		'\x12': true,
		'\x13': true,
		'\x14': true,
		'\x15': true,
		'\x16': true,
		'\x17': true,
		'\x18': true,
		'\x19': true,
		'\x1A': true,
		'\x1B': true,
		'\x1C': true,
		'\x1D': true,
		'\x1E': true,
		'\x1F': true,
		// \x7F is DEL
		'\x7F': true,
	}
)

// String is a formatted string for a LexType
func (t LexType) String() string {
	return optionStrings[uint(t)-uint(OptionAST)]
}

// Error message constants
const (
	ErrUnexpectedEOF               = "Unexpected EOF"
	ErrInvalidComment              = "A comment either be on one line after a //, or all chars between /* and */"
	ErrUnexpectedChar              = "Unexpected character"
	ErrInvalidUnicodeEscape        = `A unicode escape must be \uXXXX or \U+XXXX where X is a hex character`
	ErrInvalidStringEscape         = `The only valid string escape sequences are \\, \t, \r, \n, \uXXXX, \U+XXXX, \', and \"`
	ErrInvalidCharacterRangeEscape = `The only valid character range escape sequences are \\, \t, \r, \n, \uXXXX, \U+XXXX, and \]`
	ErrCharacterRangeEmpty         = "A character range cannot be empty"
	ErrCharacterRangeOutOfOrder    = "A character range must be in order, where begin character <= last character"
	ErrRepetitionForm              = "A repetition must be of one of the following forms: {N} or {N,} or {,N} or {N,M}; where N and M are integers, when M present N <= M, when using form {N} N must be > 0"
	ErrInvalidOption               = "The only valid options are :AST, :EOL, :INDENT, and :OUTDENT"
)

// Token is a single lexical token
type Token struct {
	typ               LexType
	token             string        // string form of token
	formattedToken    string        // formatted token
	charRangeInverted bool          // inverted character range
	charRange         map[rune]bool // character range
	n, m              int           // repetitions

}

// Type is the lexical token type
func (l Token) Type() LexType {
	return l.typ
}

// Token returns unformatted token
func (l Token) Token() string {
	return l.token
}

// String is the fmt.Stringer method that returns formatted token
func (l Token) String() string {
	return l.formattedToken
}

// InvertedRange returns true if the character range is inverted
// Only applicable if Type() returns CharacterRange
func (l Token) InvertedRange() bool {
	return l.charRangeInverted
}

// Range returns the character range
// Only applicable if Type() returns CharacterRange
func (l Token) Range() map[rune]bool {
	return l.charRange
}

// Repetitions returns n, m reptition values
// Returns n, n if specified as {N}
// Returns n, -1 if specified as {N,}
// Returns 0, n if specified as {,N}
// Returns n, m if specified as {N,M}
// Only applicable if Type() returns Repetition
func (l Token) Repetitions() (n, m int) {
	return l.n, l.m
}

// Lexer is the lexical analyzer that returns lexical tokens from input
type Lexer struct {
	iter       *goiter.Iter
	lineNumber int
}

// NewLexer constructs a Lexer from an io.Reader
func NewLexer(source io.Reader) *Lexer {
	return &Lexer{
		iter:       goiter.OfReaderRunes(source),
		lineNumber: 1,
	}
}

// Next reads next lexical token, choosing longest possible sequence
func (l *Lexer) Next() Token {
	var (
		lastReadCR               bool
		typ                      LexType
		token                    strings.Builder
		formattedToken           strings.Builder
		commentState             int           // 0 = initial /, 1 = single line, 2 = multiline looking for *, 3 = multiline trailing /
		doubleQuotes             bool          // true = double quoted String, false = single quoted String
		rangeState               int           // 0 = initial, 1 = begin, 2 = range, 3 = after end
		rangeInverted            bool          // true if range beegins with ^
		rangeBegin               rune          // begin and end chars of a single range
		rangeChars               map[rune]bool // map of all chars in a range
		repetitionState          bool          // false = N, true = M
		repetitionN, repetitionM int           // value of N and M
		nextChar                 rune
		nextCharText             string
		nextCharEscaped          bool
		result                   Token
	)

	// Handle escape sequences
	// Useful for strings and character ranges
	handleEscapes := func(isString bool) {
		// Assume this is not an escape until we know otherwise
		nextCharEscaped = false

		if nextChar == '\\' {
			// Must be a valid escape or we panic below
			nextCharEscaped = true

			// Read next char
			if !l.iter.Next() {
				panic(ErrUnexpectedEOF)
			}
			nextChar = l.iter.RuneValue()

			doPanic := false

			// Common cases are \, t, r, n, and U
			switch nextChar {
			case '\\':
				nextCharText = "\\\\"
			case 't':
				nextChar = '\t'
				nextCharText = "\\t"
			case 'r':
				nextChar = '\r'
				nextCharText = "\\r"
			case 'n':
				nextChar = '\n'
				nextCharText = "\\n"
			// String cases also include ' and "
			case '\'':
				if isString {
					nextChar = '\''
					nextCharText = "\\'"
				} else {
					doPanic = true
				}
			case '"':
				if isString {
					nextChar = '"'
					nextCharText = "\\\""
				} else {
					doPanic = true
				}
			// Character range cases also include ]
			case ']':
				if !isString {
					nextChar = ']'
					nextCharText = "\\]"
				} else {
					doPanic = true
				}
			// Not valid for any case
			default:
				doPanic = true
			}

			if doPanic {
				if isString {
					panic(ErrInvalidStringEscape)
				}
				panic(ErrInvalidCharacterRangeEscape)
			}
		}
	}

MAIN_LOOP:
	for true {
		// EOF only valid if read after a complete token
		if !l.iter.Next() {
			if typ == InvalidLexType {
				result = Token{
					typ:   EOF,
					token: "",
				}
				break MAIN_LOOP
			}
			panic(ErrUnexpectedEOF)
		}

		nextChar = l.iter.RuneValue()
		nextCharText = string(nextChar)

		switch typ {
		// First character of next token
		case InvalidLexType:
			// Skip whitespace between tokens
			if (nextChar == ' ') ||
				(nextChar == '\t') ||
				(nextChar == '\r') ||
				(nextChar == '\n') {
				// Handle line number counting
				if nextChar == '\r' {
					l.lineNumber++
					lastReadCR = true // May be part of CRLF
				} else if nextChar == '\n' {
					if lastReadCR {
						// CRLF, already incremented line number on CR
						lastReadCR = false
					} else {
						// LF by itself
						l.lineNumber++
					}
				} else {
					// Space or tab, clear CR flag if set
					lastReadCR = false
				}

				continue MAIN_LOOP
			}
			lastReadCR = false

			// Letter is first char of an identifier
			if ((nextChar >= 'A') && (nextChar <= 'Z')) ||
				((nextChar >= 'a') && (nextChar <= 'z')) {
				typ = Identifier
				token.WriteRune(nextChar)
				formattedToken.WriteString(nextCharText)
				continue MAIN_LOOP
			}

			switch nextChar {
			case '/':
				typ = Comment
				commentState = 0 // Read initial /
				continue MAIN_LOOP

			case '"':
				typ = String
				formattedToken.WriteRune(nextChar)
				doubleQuotes = true
				continue MAIN_LOOP

			case '\'':
				typ = String
				formattedToken.WriteRune(nextChar)
				doubleQuotes = false
				continue MAIN_LOOP

			case '[':
				typ = CharacterRange
				token.WriteRune(nextChar)
				formattedToken.WriteRune(nextChar)
				rangeState = 0
				rangeInverted = false
				rangeChars = map[rune]bool{}
				continue MAIN_LOOP

			case '{':
				typ = Repetition
				token.WriteRune(nextChar)
				formattedToken.WriteRune(nextChar)
				repetitionState = false // Start reading N
				repetitionN = -1        // Must have at least one char
				repetitionM = -1        // May not have an M
				continue MAIN_LOOP

			case '?':
				// zero or one repetitions - same as {0,1}
				result = Token{
					typ:            Repetition,
					token:          "?",
					formattedToken: "?",
					n:              0,
					m:              1,
				}
				break MAIN_LOOP

			case '*':
				// zero or more repetitions - same as {0,}
				result = Token{
					typ:            Repetition,
					token:          "*",
					formattedToken: "*",
					n:              0,
					m:              -1,
				}
				break MAIN_LOOP

			case '+':
				// one or more repetitions - same as {1,}
				result = Token{
					typ:            Repetition,
					token:          "+",
					formattedToken: "+",
					n:              1,
					m:              -1,
				}
				break MAIN_LOOP

			case ':':
				typ = OptionAST // choose first for now
				token.WriteRune(nextChar)
				formattedToken.WriteRune(nextChar)
				continue MAIN_LOOP

			case '^':
				result = Token{
					typ:            Hat,
					token:          "^",
					formattedToken: "^",
				}
				break MAIN_LOOP

			case '(':
				result = Token{
					typ:            OpenParens,
					token:          "(",
					formattedToken: "(",
				}
				break MAIN_LOOP

			case ')':
				result = Token{
					typ:            CloseParens,
					token:          ")",
					formattedToken: ")",
				}
				break MAIN_LOOP

			case '|':
				result = Token{
					typ:            Bar,
					token:          "|",
					formattedToken: "|",
				}
				break MAIN_LOOP

			case ',':
				result = Token{
					typ:            Comma,
					token:          ",",
					formattedToken: ",",
				}
				break MAIN_LOOP

			case '=':
				// If next char is also =, then it is DoubleEquals
				if !l.iter.Next() {
					panic(ErrUnexpectedEOF)
				}

				if nextChar = l.iter.RuneValue(); nextChar == '=' {
					result = Token{
						typ:            DoubleEquals,
						token:          "==",
						formattedToken: "==",
					}
					break MAIN_LOOP
				}

				// Char after = is first char of next token
				l.iter.Unread(nextChar)

				result = Token{
					typ:            Equals,
					token:          "=",
					formattedToken: "=",
				}
				break MAIN_LOOP

			case ';':
				result = Token{
					typ:            SemiColon,
					token:          ";",
					formattedToken: ";",
				}
				break MAIN_LOOP
			}

			panic(ErrUnexpectedChar)

		case Identifier:
			if ((nextChar >= 'A') && (nextChar <= 'Z')) ||
				((nextChar >= 'a') && (nextChar <= 'z')) ||
				((nextChar >= '0') && (nextChar <= '9')) ||
				(nextChar == '_') {
				token.WriteRune(nextChar)
				formattedToken.WriteString(nextCharText)
				continue MAIN_LOOP
			}

			// Must be first char of next token
			l.iter.Unread(nextChar)

			// Identifier is what we have before this char
			result = Token{
				typ:            typ,
				token:          token.String(),
				formattedToken: formattedToken.String(),
			}
			break MAIN_LOOP

		case Comment:
			switch commentState {
			case 0:
				// Read /, next char must be / or *
				switch nextChar {
				case '/':
					commentState = 1 // single line
					continue MAIN_LOOP

				case '*':
					commentState = 2 // multi line looking for *
					continue MAIN_LOOP

				default:
					// Unlike mnost languages, only use for / is to start a comment
					panic(ErrInvalidComment)
				}

			case 1:
				// single line
				if (nextChar == '\r') || (nextChar == '\n') {
					// No need to push back eol char, don't need to consume more eol chars
					result = Token{
						typ:            typ,
						token:          token.String(),
						formattedToken: formattedToken.String(),
					}
					break MAIN_LOOP
				}

				token.WriteRune(nextChar)
				formattedToken.WriteString(nextCharText)
				continue MAIN_LOOP

			case 2:
				// multiline looking for *
				if nextChar == '*' {
					commentState = 3

					// Don't add * to data until we know whether or not it is part of */
					continue MAIN_LOOP
				}

				token.WriteRune(nextChar)
				formattedToken.WriteString(nextCharText)
				continue MAIN_LOOP

			default:
				// multiline looking for / after *
				if nextChar == '/' {
					result = Token{
						typ:            typ,
						token:          token.String(),
						formattedToken: formattedToken.String(),
					}
					break MAIN_LOOP
				}

				// Write a * and this char since we know the * is part of comment
				token.WriteRune('*')
				token.WriteRune(nextChar)
				formattedToken.WriteRune('*')
				formattedToken.WriteString(nextCharText)

				// Go back to looking for *
				commentState = 2
				continue MAIN_LOOP
			}

		case String:
			// Escapes can be used in terminals
			handleEscapes(true)

			// Look for terminating quote char
			if (doubleQuotes && (nextChar == '"') && (!nextCharEscaped)) ||
				((!doubleQuotes) && (nextChar == '\'') && (!nextCharEscaped)) {
				// Allow zero length terminals, they mean epsilon
				formattedToken.WriteRune(nextChar)
				result = Token{
					typ:            typ,
					token:          token.String(),
					formattedToken: formattedToken.String(),
				}
				break MAIN_LOOP
			}

			// Part of terminal string
			token.WriteRune(nextChar)
			formattedToken.WriteString(nextCharText)
			continue MAIN_LOOP

		case CharacterRange:
			// Examine the char range and handle dashes according to the JavaScript definition:
			//
			// A dash character can be treated literally or it can denote a range.
			// It is treated literally if it is the first or last character of ClassRanges,
			// the beginning or end limit of a range specification,
			// or immediately follows a range specification.
			//
			// where ClassRanges is the entire set of range(s) contained in square brackets;
			// and a range specification is a sequence of a character, a dash, and a character.
			//
			// Note that if the trange begins with ^-. the dash is literal.

			// Escapes may be used in character ranges
			handleEscapes(false)

			switch rangeState {
			case 0: // First char
				token.WriteString(nextCharText)
				formattedToken.WriteString(nextCharText)

				// If nextChar is ^ and range is already inverted, must be ^^, where second ^ is literal, and is part of range
				if (nextChar == '^') && (!rangeInverted) {
					// Starts with ^, so invert the range
					// Always exclude useless ASCII conntrol characters
					rangeInverted = true
					rangeChars = uselessChars
					continue MAIN_LOOP
				}

				if (nextChar == ']') && (!nextCharEscaped) {
					if rangeInverted {
						// Valid range of not nothing = everything; we already excluded useless ASCII control characters above
						return Token{
							typ:               typ,
							token:             token.String(),
							formattedToken:    formattedToken.String(),
							charRangeInverted: rangeInverted,
							charRange:         rangeChars,
						}
					}

					panic(ErrCharacterRangeEmpty)
				}

				// This may be range begin
				rangeState = 1
				rangeBegin = nextChar
				continue MAIN_LOOP

			case 1: // Possible range begin
				token.WriteString(nextCharText)
				formattedToken.WriteString(nextCharText)

				if (nextChar == ']') && (!nextCharEscaped) {
					// last char in rangeBegin is a literal char
					rangeChars[rangeBegin] = true
					return Token{
						typ:               typ,
						token:             token.String(),
						formattedToken:    formattedToken.String(),
						charRangeInverted: rangeInverted,
						charRange:         rangeChars,
					}
				}

				if nextChar == '-' {
					// Possible range of chars
					rangeState = 2
				} else {
					// Last char is not part of range
					rangeChars[rangeBegin] = true
					// But this one might bee
					rangeBegin = nextChar
				}

				continue MAIN_LOOP

			case 2: // rangeBegin dash nextChar
				if (nextChar == ']') && (!nextCharEscaped) {
					// previous dash was a literal dash at end
					token.WriteString(nextCharText)
					formattedToken.WriteString(nextCharText)
					rangeChars[rangeBegin] = true
					rangeChars['-'] = true
					return Token{
						typ:               typ,
						token:             token.String(),
						formattedToken:    formattedToken.String(),
						charRangeInverted: rangeInverted,
						charRange:         rangeChars,
					}
				}

				token.WriteString(nextCharText)
				formattedToken.WriteString(nextCharText)

				// range from rangeBegin thru nextChar inclusive
				if rangeBegin > nextChar {
					panic(ErrCharacterRangeOutOfOrder)
				}

				for r := rangeBegin; r <= nextChar; r++ {
					rangeChars[r] = true
				}

				rangeState = 3
				continue MAIN_LOOP

			case 3:
				// after range end
				if (nextChar == ']') && (!nextCharEscaped) {
					//					if true {
					//						panic("here")
					//					}
					token.WriteString(nextCharText)
					formattedToken.WriteString(nextCharText)
					return Token{
						typ:               typ,
						token:             token.String(),
						formattedToken:    formattedToken.String(),
						charRangeInverted: rangeInverted,
						charRange:         rangeChars,
					}
				}

				token.WriteString(nextCharText)
				formattedToken.WriteString(nextCharText)

				// Any char after range end is literal, may be start of next range
				rangeState = 1
				rangeBegin = nextChar

				continue MAIN_LOOP
			}

		case Repetition:
			// Read required N and optional ,M before closing brace
			if !repetitionState {
				if (nextChar >= '0') && (nextChar <= '9') {
					if repetitionN == -1 {
						repetitionN = int(nextChar - '0')
					} else {
						repetitionN = repetitionN*10 + int(nextChar-'0')
					}

					token.WriteRune(nextChar)
					formattedToken.WriteString(nextCharText)
					continue MAIN_LOOP
				}

				if nextChar == ',' {
					// Form is {,N}; don't set n = 1 yet, in case we have only a comma, which is invalid
					repetitionState = true // Read M, if we have it
					token.WriteRune(nextChar)
					formattedToken.WriteString(nextCharText)
					continue MAIN_LOOP
				}

				if nextChar == '}' {
					// form {N}
					token.WriteRune(nextChar)
					formattedToken.WriteString(nextCharText)

					if repetitionN < 1 {
						// N must have a value >= 1
						panic(ErrRepetitionForm)
					}

					result = Token{
						typ:            typ,
						token:          token.String(),
						formattedToken: formattedToken.String(),
						n:              repetitionN,
						m:              repetitionN, // M = N
					}
					break MAIN_LOOP
				}

				panic(ErrRepetitionForm)
			} else {
				// Reading M
				if (nextChar >= '0') && (nextChar <= '9') {
					if repetitionM == -1 {
						repetitionM = int(nextChar - '0')
					} else {
						repetitionM = repetitionM*10 + int(nextChar-'0')
					}

					token.WriteRune(nextChar)
					formattedToken.WriteString(nextCharText)
					continue MAIN_LOOP
				}

				if nextChar == '}' {
					// If we never read N, N was initialized to -1
					// If we never read M, M was initialized to -1

					// If both N and M are -1, we read just a comma
					if (repetitionN == -1) && (repetitionM == -1) {
						panic(ErrRepetitionForm)
					}

					// N can be zero, M must be -1 or >= 1
					if repetitionM == 0 {
						panic(ErrRepetitionForm)
					}

					token.WriteRune(nextChar)
					formattedToken.WriteString(nextCharText)

					// If N = -1, must be {,N} - provide 0, M
					if repetitionN == -1 {
						repetitionN = 0
					}

					result = Token{
						typ:            typ,
						token:          token.String(),
						formattedToken: formattedToken.String(),
						n:              repetitionN,
						m:              repetitionM,
					}
					break MAIN_LOOP
				}

				panic(ErrRepetitionForm)
			}

		case OptionAST:
			// Remain at type AST until we have read whole option string
			// Like identifier, negative end: stop on first non-letter char
			if (nextChar >= 'A') && (nextChar <= 'Z') {
				token.WriteRune(nextChar)
				formattedToken.WriteString(nextCharText)
				continue MAIN_LOOP
			}

			// Must be first char of next token
			l.iter.Unread(nextChar)

			// String must match a value optionStrings
			tokenStr := token.String()
			for i, optionStr := range optionStrings {
				if tokenStr == optionStr {
					result = Token{
						typ:            LexType(int(OptionAST) + i),
						token:          token.String(),
						formattedToken: formattedToken.String(),
					}
					break MAIN_LOOP
				}
			}

			panic(ErrInvalidOption)
		}
	}

	return result
}
