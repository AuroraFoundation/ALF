package lexer_test

import (
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/AuroraFoundation/ALF/pkg/lexer"
)

func TestLexerItem(t *testing.T) {
	t.Run("basic attribute-value", func(t *testing.T) {
		test := "Name: Hello"
		want := []lexer.Item{
			{lexer.TokenName, "Name", 0, 0},
			{lexer.TokenColon, ":", 0, 4},
			{lexer.TokenWhitespace, " ", 0, 5},
			{lexer.TokenText, "Hello", 0, 6},
			{lexer.TokenEOF, "", 0, 11},
		}

		var got []lexer.Item

		items := itemsFromString(t, test)

		for {
			item := <-items

			got = append(got, item)

			if item.Token == lexer.TokenEOF {
				break
			}
		}

		if !reflect.DeepEqual(got, want) {
			t.Log("Got: ", got)
			t.Log("Want:", want)
			t.Error()
		}
	})

	t.Run("basic comment", func(t *testing.T) {
		items := itemsFromString(t, "# Comment.")

		want := lexer.Item{
			Token:   lexer.TokenComment,
			Literal: "# Comment.",
			Line:    0,
			Col:     0,
		}
		got := <-items

		if got != want {
			t.Errorf("got %s, want %s", got, want)
		}
	})
}

func TestLexerLineStartToken(t *testing.T) {
	cases := []struct {
		desc   string
		source string
		want   lexer.Token
	}{
		{
			"comment",
			"# Comment",
			lexer.TokenComment,
		},
		{
			"attribute",
			"Attr:",
			lexer.TokenName,
		},
		{
			"eof",
			"",
			lexer.TokenEOF,
		},
		{
			"indent with spaces",
			"  ",
			lexer.TokenIndent,
		},
		{
			"indent with tabs",
			"\t",
			lexer.TokenIndent,
		},
	}

	for _, tt := range cases {
		t.Run(tt.desc, func(t *testing.T) {
			items := itemsFromString(t, tt.source)
			got := (<-items).Token
			assertToken(t, got, tt.want)
		})
	}
}

func TestLexerMiddleToken(t *testing.T) {
	t.Run("multiline string", func(t *testing.T) {
		items := itemsFromString(t, "\tThis is a multiline string.")
		want := lexer.TokenText

		// Consume indent token.
		<-items

		got := (<-items).Token

		assertToken(t, got, want)
	})

	t.Run("lists", func(t *testing.T) {
		items := itemsFromString(t, "\t- This is a list.")
		want := lexer.TokenList

		// Consume indent token.
		<-items

		got := (<-items).Token

		assertToken(t, got, want)
	})
}

func TestLexerLocation(t *testing.T) {
	t.Run("line", func(t *testing.T) {
		items := itemsFromString(t, "# ABC.\n# XYZ.")

		// Literal "# ABC.".
		got := (<-items).Line
		want := 0
		assertInt(t, got, want)

		// Literal "\n".
		// Here the newline character is treated as a token, which
		// takes effect AFTER it has been processed.
		got = (<-items).Line
		want = 0
		assertInt(t, got, want)

		// Literal "# XYZ.".
		got = (<-items).Line
		want = 1
		assertInt(t, got, want)
	})

	t.Run("column", func(t *testing.T) {
		items := itemsFromString(t, "  # 1.")

		// Literal "  " (indentation).
		got := (<-items).Col
		want := 0
		assertInt(t, got, want)

		// Literal "# 1.".
		got = (<-items).Col
		want = 2
		assertInt(t, got, want)
	})

	t.Run("column with newline", func(t *testing.T) {
		items := itemsFromString(t, "# 1.\n  # 2.")

		// Literal "# 1.".
		got := (<-items).Col
		want := 0
		assertInt(t, got, want)

		// Literal "\n".
		// Here the newline character is treated as a token, which
		// takes effect AFTER it has been processed.
		got = (<-items).Col
		want = 4
		assertInt(t, got, want)

		// Literal "  " (indentation).
		got = (<-items).Col
		want = 0
		assertInt(t, got, want)

		// Literal "# 2.".
		got = (<-items).Col
		want = 2
		assertInt(t, got, want)
	})
}

type readerError string

func (r readerError) Read([]byte) (int, error) {
	return 0, errors.New(string(r))
}

func TestLexerError(t *testing.T) {
	t.Run("basic", func(t *testing.T) {
		re := readerError("basic")

		_, items := lexer.New(re)

		want := lexer.TokenError
		got := (<-items).Token

		assertToken(t, got, want)
	})

	t.Run("message", func(t *testing.T) {
		want := "message with error"
		re := readerError(want)

		lex, items := lexer.New(re)
		if item := <-items; item.Token != lexer.TokenError {
			t.Fatalf("Unexpected Token: %v", item)
		}

		if lex.Error().Error() != want {
			t.Errorf("got %q, want %q", lex.Error().Error(), want)
		}
	})
}

func itemsFromString(t *testing.T, s string) <-chan lexer.Item {
	t.Helper()
	_, items := lexer.New(strings.NewReader(s))
	return items
}

func assertToken(t *testing.T, got, want lexer.Token) {
	t.Helper()

	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func assertString(t *testing.T, got, want string) {
	t.Helper()

	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func assertInt(t *testing.T, got, want int) {
	t.Helper()

	if got != want {
		t.Errorf("got %d, want %d", got, want)
	}
}
