package parser

import (
	"errors"
	"io"

	"github.com/AuroraFoundation/ALF/pkg/lexer"
)

type ALF struct {
	Title  string
	Author string
	Names  []string
	Artist string
	Album  string
	Notes  []string
}

type Parser struct {
	r io.Reader
}

func New(r io.Reader) *Parser {
	return &Parser{
		r,
	}
}

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
