// Package lexer implements a tokenizer or lexer for Aurora Lyrics Language
// (ALF) source files.
package lexer

import (
	"bufio"
	"fmt"
	"io"
	"strings"
)

// Lexer is a token generator for Aurora Lyrics Format (ALF) source files.
type Lexer struct {
	// A buffered `io.Reader` containing the source file.
	reader *bufio.Reader
	// Channel to which the items containing the tokens must be sent.
	items chan Item
	// Current line and column in the file (start at zero).
	line, col int
	// Line where the current token starts.
	startTkLine int
	// Column where the current token starts.
	startTkCol int
	// Contains the number of the last column of the previous line (see
	// `Lexer.backup` and `Lexer.next`).
	lastCol int
	// Reports whether the last rune processed was a newline (see
	// `Lexer.backup` and `Lexer.next`).
	lastIsNewline bool
	// The text as it is in the currently processed token source (see
	// `Lexer.append` and `Lexer.emit`).
	tkLiteral string
}

// Item contains a token, its literal text and its location in the source file.
type Item struct {
	Token   Token
	Literal string
	Line    int // Starts from 0.
	Col     int // Starts from 0.
}

// Token is a unique identifier for each part of the Aurora Lyrics Format (ALF)
// grammar.
type Token int

// Aurora Lyrics Format (ALF) grammar tokens.
const (
	TokenEOF Token = iota + 1
	TokenNewline
	TokenWhitespace
	TokenIndent
	TokenColon
	TokenComment
	TokenName
	TokenText
)

//go:generate stringer -trimprefix Token -type Token

// stateFn represents a state (item) in the machine (lexer), as a comment or
// attribute.
type stateFn func() stateFn

// runeError is a flag indicating that some error has occurred while trying to
// read a rune.
const runeError rune = 0

// New creates and initializes a new `Lexer` and starts collecting tokens
// immediately using concurrency. The `r` argument must be source code in ALF
// format.
//
// The first value returned is a pointer to a `Lexer` structure, provided for
// monitoring purposes (it currently does nothing).
//
// The second value returned is a read-only channel on which `Item` structures
// are sent with the tokens found.
func New(r io.Reader) (*Lexer, <-chan Item) {
	lex := Lexer{
		reader: bufio.NewReader(r),
		items:  make(chan Item),
	}

	// Start the machine in the background.
	go lex.run()

	return &lex, lex.items
}

// run starts the execution of the machine (lexer), due to the concurrent
// nature, it must be called from a goroutine or it will block forever.
func (l *Lexer) run() {
	for state := l.initState(); state != nil; {
		state = state()
	}
}

// initState is the initial state of the machine, either because it has just
// been turned on, or because a line break has been reached.
func (l *Lexer) initState() stateFn {
	switch l.peek() {
	case runeError:
		l.items <- Item{Token: TokenEOF, Line: l.line, Col: l.col}

		return nil
	case '\n':
		return l.stateNewline
	case '#':
		return l.stateComment
	case ' ':
		return l.stateIndent
	default:
		return l.stateName
	}
}

// stateIndent handles spaces at the beginning of each line, better known as
// indentation.
func (l *Lexer) stateIndent() stateFn {
	l.mustAccept(" ")
	l.emit(TokenIndent)

	return l.initState
}

// stateNewline handles all line endings encountered.
//
// Note that a line break is reported as one more character on the previous
// line, but subsequent characters will be reported on a newline; that is, the
// newline character takes effect AFTER it has been processed.
func (l *Lexer) stateNewline() stateFn {
	l.mustAccept("\n")
	l.emit(TokenNewline)

	return l.initState
}

// stateComment handles all comments found, i.e. all text after a "#" until the
// end of the line.
func (l *Lexer) stateComment() stateFn {
	l.mustAccept("#")

Loop:
	for {
		switch char := l.next(); char {
		case runeError:
			break Loop
		case '\n': // End Comment.
			l.backup()

			break Loop
		default:
			l.append(char)
		}
	}

	l.emit(TokenComment)

	return l.initState
}

// stateName handles all identifiers that can be found at the beginning of the
// line, e.g. the name of an attribute.
func (l *Lexer) stateName() stateFn {
	const ascii = `AaBbCcDdEeFfGgHhIiJjKkLlMmNnOoPpQqRrSsTtUuVvWwXxYyZz`

	l.mustAccept(ascii)
	l.emit(TokenName)

	return l.stateColon
}

// stateColon handles the separator between an attribute and its value, i.e.
// the ":" character.
func (l *Lexer) stateColon() stateFn {
	l.mustAccept(":")
	l.emit(TokenColon)

	if l.peek() == ' ' {
		l.emitWhitespace()
	}

	return l.stateText
}

// stateText handles all literal text, typically the value of an attribute; the
// literal text goes to the end of the line or until a comment is found.
func (l *Lexer) stateText() stateFn {
	l.startTkLine = l.line
	l.startTkCol = l.col

	for {
		switch char := l.next(); char {
		case runeError:
			l.emit(TokenText)

			return l.initState
		case '\n':
			l.backup()
			l.emit(TokenText)

			return l.initState
		default:
			l.append(char)
		}
	}
}

// emitWhitespace traps any blanks that may be present and emits the
// corresponding token.
func (l *Lexer) emitWhitespace() {
	l.mustAccept(" ")
	l.emit(TokenWhitespace)
}

// emit sends a new `Item` to the queue. It also takes care of resetting some
// variables used by `Lexer.accept` and `Lexer.append`.
func (l *Lexer) emit(token Token) {
	literal := l.tkLiteral
	line, col := l.startTkLine, l.startTkCol

	// Reset variables.
	l.tkLiteral = ""
	l.startTkLine, l.startTkCol = 0, 0

	l.items <- Item{
		Token:   token,
		Literal: literal,
		Line:    line,
		Col:     col,
	}
}

// mustAccept is a version of `Lexer.accept` that panics if no match was found.
func (l *Lexer) mustAccept(s string) {
	r := l.peek()

	if !l.accept(s) {
		panic(fmt.Sprintf("%q not in any of %q", r, s))
	}
}

// accept advances in the source by adding the characters to the variable with
// the literal text of the next token, the advance is done by comparing the
// next available rune in the source with the runes of the string provided in
// the argument until no match is found, if no match was found from the
// beginning, false is returned.
func (l *Lexer) accept(pattern string) bool {
	if !strings.ContainsRune(pattern, l.peek()) {
		return false
	}

	// Establish token location (if it had not been established before).
	if l.startTkLine+l.startTkCol == 0 {
		l.startTkLine = l.line
		l.startTkCol = l.col
	}

	for strings.ContainsRune(pattern, l.peek()) {
		l.append(l.next())
	}

	return true
}

// append appends the supplied string to the literal text of the next token.
func (l *Lexer) append(r rune) {
	l.tkLiteral += string(r)
}

// peek returns the next rune from the source, without advancing in this one.
// If an error occurs, the constant `runeError` is returned.
func (l *Lexer) peek() rune {
	r, err := l.reader.Peek(1)
	if err != nil {
		return runeError
	}

	return rune(r[0])
}

// next returns the next rune in the source and advances it (unlike
// `Lexer.peek`), if an error occurs, the constant `runeError` is returned.
func (l *Lexer) next() rune {
	char, _, err := l.reader.ReadRune()
	if err != nil {
		return runeError
	}

	if char == '\n' {
		l.lastIsNewline = true
		l.line++
		l.lastCol = l.col
		l.col = 0
	} else {
		l.lastIsNewline = false
		l.col++
	}

	return char
}

// backup returns the source to the last rune and sets the token location
// variables.
func (l *Lexer) backup() {
	if err := l.reader.UnreadRune(); err != nil {
		panic(err)
	}

	if l.lastIsNewline {
		l.line--
		l.col = l.lastCol
		l.lastCol = 0
	} else {
		l.col--
	}
}
