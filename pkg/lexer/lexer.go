package lexer

import (
	"bufio"
	"fmt"
	"io"
	"strings"
)

type Lexer struct {
	items                         chan Item
	r                             *bufio.Reader
	line, col, lastCol            int
	latestLine                    bool
	startTokenCol, startTokenLine int
	tokenLiteral                  string
}

type Item struct {
	Token   Token
	Literal string
	Line    int
	Col     int
}

type Token int

const (
	TokenEOF Token = iota + 1
	TokenComment
	TokenNewline
	TokenIndent
	TokenColon
	TokenWhitespace
	TokenName
	TokenText
)

//go:generate stringer -trimprefix Token -type Token

type stateFn func() stateFn

func New(r io.Reader) (*Lexer, <-chan Item) {
	lex := Lexer{
		r:     bufio.NewReader(r),
		items: make(chan Item),
	}

	go lex.run()

	return &lex, lex.items
}

func (l *Lexer) run() {
	for state := l.initState(); state != nil; {
		state = state()
	}
}

func (l *Lexer) initState() stateFn {
	switch l.peek() {
	case 0:
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

func (l *Lexer) stateIndent() stateFn {
	l.mustAccept(" ")
	l.emit(TokenIndent)
	return l.initState
}

func (l *Lexer) stateNewline() stateFn {
	l.mustAccept("\n")
	l.emit(TokenNewline)
	return l.initState
}

func (l *Lexer) stateComment() stateFn {
	l.mustAccept("#")

Loop:
	for {
		switch r := l.next(); r {
		case 0: // Error.
			break Loop
		case '\n': // End Comment.
			l.backup()
			break Loop
		default:
			l.append(r)
		}
	}

	l.emit(TokenComment)

	return l.initState
}

func (l *Lexer) stateName() stateFn {
	const ascii = `AaBbCcDdEeFfGgHhIiJjKkLlMmNnOoPpQqRrSsTtUuVvWwXxYyZz`

	l.mustAccept(ascii)
	l.emit(TokenName)

	return l.stateColon
}

func (l *Lexer) stateColon() stateFn {
	l.mustAccept(":")
	l.emit(TokenColon)

	if l.peek() == ' ' {
		l.emitWhitespace()
	}

	return l.stateText
}

func (l *Lexer) stateText() stateFn {
	l.startTokenLine = l.line
	l.startTokenCol = l.col

	for {
		switch r := l.next(); r {
		case 0:
			l.emit(TokenText)
			return l.initState
		case '\n':
			l.backup()
			l.emit(TokenText)
			return l.initState
		default:
			l.append(r)
		}
	}

	return l.initState
}

func (l *Lexer) emitWhitespace() {
	l.mustAccept(" ")
	l.emit(TokenWhitespace)
}

func (l *Lexer) emit(token Token) {
	literal := l.tokenLiteral
	line, col := l.startTokenLine, l.startTokenCol

	l.tokenLiteral = ""
	l.startTokenLine, l.startTokenCol = 0, 0

	l.items <- Item{
		Token:   token,
		Literal: literal,
		Line:    line,
		Col:     col,
	}
}

func (l *Lexer) mustAccept(s string) {
	r := l.peek()

	if !l.accept(s) {
		panic(fmt.Sprintf("%q not in any of %q", r, s))
	}
}

func (l *Lexer) accept(s string) bool {
	if !strings.ContainsRune(s, l.peek()) {
		return false
	}

	if l.startTokenLine+l.startTokenCol == 0 {
		l.startTokenLine = l.line
		l.startTokenCol = l.col
	}

	for strings.ContainsRune(s, l.peek()) {
		l.append(l.next())
	}

	return true
}

func (l *Lexer) append(r rune) {
	l.tokenLiteral += string(r)
}

func (l *Lexer) peek() rune {
	r, err := l.r.Peek(1)
	if err != nil {
		return 0
	}

	return rune(r[0])
}

func (l *Lexer) next() rune {
	r, _, err := l.r.ReadRune()
	if err != nil {
		return 0
	}

	if r == '\n' {
		l.lastCol = l.col
		l.line++
		l.col = 0
		l.latestLine = true
	} else {
		l.col++
		l.latestLine = false
	}

	return r
}

func (l *Lexer) backup() {
	if err := l.r.UnreadRune(); err != nil {
		panic(err)
	}

	if l.latestLine {
		l.col = l.lastCol
		l.line--
		l.lastCol = 0
	} else {
		l.col--
	}
}
