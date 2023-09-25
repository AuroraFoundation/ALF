package lexer_test

import (
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/AuroraFoundation/ALF/pkg/lexer"
)

func TestLexerItem(t *testing.T) {
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
}

func TestLexerToken(t *testing.T) {
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
			"name",
			"Attr: Value",
			lexer.TokenName,
		},
		{
			"eof",
			"",
			lexer.TokenEOF,
		},
		{
			"indent",
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

func TestLexerLiteral(t *testing.T) {
	items := itemsFromString(t, "# Other.")

	got := (<-items).Literal
	want := "# Other."

	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
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

func assertInt(t *testing.T, got, want int) {
	t.Helper()

	if got != want {
		t.Errorf("got %d, want %d", got, want)
	}
}
