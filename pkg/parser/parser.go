package parser

import (
	"errors"
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

	Notes  []string
}

type Lyric struct {
	Order []string
}

// Parser implements a source code file parser in ALF.
type Parser struct {
	r io.Reader
}

// New creates and initializes a new `Parser` structure.
func New(r io.Reader) *Parser {
	return &Parser{
		r,
	}
}

// Decode parses and returns an `ALF` structure with the parsed source code.
func (p *Parser) Decode() (ALF, error) {
	_, items := lexer.New(p.r)

	var attrName string
	var alf ALF

	for {
		switch item := <-items; item.Token {
		case lexer.TokenName:
			attrName = item.Literal

		case lexer.TokenText:
			switch attrName {
			case "Title":
				alf.Title = item.Literal
			case "Author":
				alf.Author = item.Literal
			case "Artist":
				alf.Artist = item.Literal
			case "Album":
				alf.Album = item.Literal
			default:
				return ALF{}, errors.New("unknown attribute name")
			}

		case lexer.TokenList:
			<-items // Consume space.
			item = <-items

			if attrName == "Names" {
				alf.Names = append(alf.Names, item.Literal)
			} else {
				alf.Notes = append(alf.Notes, item.Literal)
			}

		case lexer.TokenEOF:
			return alf, nil
		}
	}
}
