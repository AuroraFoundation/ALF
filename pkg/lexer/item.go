package lexer

import "fmt"

// Item contains a token, its literal text and its location in the source file.
type Item struct {
	Token   Token
	Literal string
	Line    int // Starts from 0.
	Col     int // Starts from 0.
}

func (i Item) String() string {
	return fmt.Sprintf("<Item (%s)[%d:%d] %q>",
		i.Token,
		i.Line,
		i.Col,
		i.Literal,
	)
}

// Token is a unique identifier for each part of the Aurora Lyrics Format (ALF)
// grammar.
type Token int

// Aurora Lyrics Format (ALF) grammar tokens.
const (
	TokenError Token = iota
	TokenEOF
	TokenNewline
	TokenWhitespace
	TokenIndent
	TokenColon
	TokenComment
	TokenName
	TokenText
)
