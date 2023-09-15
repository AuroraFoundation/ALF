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
	TokenName
)

//go:generate stringer -trimprefix Token -type Token

func New(r io.Reader) (*Lexer, <-chan Item) {
	lex := Lexer{
		r:     bufio.NewReader(r),
		items: make(chan Item),
	}

	go lex.run()

	return &lex, lex.items
}

func (l *Lexer) run() {
	for {
		switch l.peek() {
		case 0:
			l.items <- Item{Token: TokenEOF}
			return
		case '\n':
			l.stateNewline()
		case '#':
			l.stateComment()
		case ' ':
			l.stateIndent()
		default:
			l.stateName()
		}
	}
}

func (l *Lexer) stateIndent() {
	l.mustAccept(" ")
	l.emit(TokenIndent)
}

func (l *Lexer) stateNewline() {
	l.mustAccept("\n")
	l.emit(TokenNewline)
}

func (l *Lexer) stateComment() {
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
			l.tokenLiteral += string(r)
		}
	}

	l.emit(TokenComment)
}

func (l *Lexer) stateName() {
	const ascii = `AaBbCcDdEeFfGgHhIiJjKkLlMmNnOoPpQqRrSsTtUuVvWwXxYyZz`

	l.mustAccept(ascii + ": ")
	l.emit(TokenName)
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
		l.tokenLiteral += string(l.next())
	}

	return true
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
