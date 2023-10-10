package parser

import (
	"errors"
	"fmt"
	"io"

	"github.com/AuroraFoundation/ALF/pkg/lexer"
)

// ALF contains a structure equivalent to the Aurora Lyrics Format
// specification.
type ALF struct {
	Title  string
	Author string
	Names  []string
	Artist string
	Album  string

	Lyric Lyric

	Notes []string
}

type Lyric struct {
	Order []string
}

// Parser implements a source code file parser in ALF.
type Parser struct {
	lex   *lexer.Lexer
	items <-chan lexer.Item

	peek lexer.Item
}

// New creates and initializes a new `Parser` structure.
func New(r io.Reader) *Parser {
	p := Parser{}

	p.lex, p.items = lexer.New(r)

	return &p
}

// Decode parses and returns an `ALF` structure with the parsed source code.
func (p *Parser) Decode() (ALF, error) {
	var alf ALF

	for {
		switch item := p.nextItem(); item.Token {
		case lexer.TokenName:
			switch item.Literal {
			case "Title":
				alf.Title = p.parseSimpleAttrVal()
			case "Author":
				alf.Author = p.parseSimpleAttrVal()
			case "Artist":
				alf.Artist = p.parseSimpleAttrVal()
			case "Album":
				alf.Album = p.parseSimpleAttrVal()
			case "Names":
				alf.Names = p.parseList(0)
			case "Notes":
				alf.Notes = p.parseList(0)
			case "Lyric":
				alf.Lyric = p.parseLyric()
			default:
				return ALF{}, errors.New(fmt.Sprintf("unknown attribute name %q", item.Literal))
			}

		case lexer.TokenEOF:
			return alf, nil
		}
	}
}

func (p *Parser) parseLyric() Lyric {
	var lyric Lyric

	// Consume colon.
	p.nextItem()

	indent := p.parseIndent()

	for {
		item := p.peekItem()

		if item.Token == lexer.TokenEOF {
			return lyric
		}

		p.nextItem()

		if item.Token == lexer.TokenNewline {
			if p.parseIndent() < indent {
				return lyric
			}
		}

		if item.Literal == "Order" {
			lyric.Order = p.parseList(indent)
		}
	}
}

func (p *Parser) parseList(n int) []string {
	var list []string

	// Consume colon.
	p.nextItem()

	for {
		item := p.peekItem()

		if item.Token == lexer.TokenEOF {
			return list
		}

		p.nextItem()

		if item.Token == lexer.TokenNewline {
			if p.parseIndent() < n {
				return list
			}

			continue
		}

		if item.Token != lexer.TokenList {
			return list
		}

		// Consume whitespace.
		p.nextItem()

		list = append(list, p.nextItem().Literal)
	}
}

func (p *Parser) parseSimpleAttrVal() string {
	// Consume colon and whitespace.
	p.nextItem()
	p.nextItem()

	return p.nextItem().Literal
}

func (p *Parser) parseIndent() int {
	item := p.peekItem()

	if item.Token != lexer.TokenIndent {
		return 0
	}

	p.nextItem()

	return len(item.Literal)
}

func (p *Parser) nextItem() lexer.Item {
	if p.peek == (lexer.Item{}) {
		return <-p.items
	}

	ret := p.peek
	p.peek = (lexer.Item{})
	return ret
}

func (p *Parser) peekItem() lexer.Item {
	if p.peek == (lexer.Item{}) {
		p.peek = <-p.items
	}

	return p.peek
}
