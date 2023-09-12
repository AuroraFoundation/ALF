package lexer_test

import (
	"strings"
	"testing"

	"github.com/AuroraFoundation/ALF/pkg/lexer"
)

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

		got := (<-items).Line
		want := 0

		assertInt(t, got, want)

		got = (<-items).Line
		want = 1

		assertInt(t, got, want)
	})

	t.Run("column", func(t *testing.T) {
		items := itemsFromString(t, "  # 1.")

		got := (<-items).Col
		want := 0

		assertInt(t, got, want)

		got = (<-items).Col
		want = 2

		assertInt(t, got, want)
	})

	t.Run("column with newline", func(t *testing.T) {
		items := itemsFromString(t, "# 1.\n  # 2.")

		// Comment.
		got := (<-items).Col
		want := 0

		assertInt(t, got, want)

		// Indent.
		got = (<-items).Col
		want = 0

		assertInt(t, got, want)

		// Comment.
		got = (<-items).Col
		want = 2

		assertInt(t, got, want)
	})
}

func TestLexer(t *testing.T) {
	t.Run("literal", func(t *testing.T) {
		items := itemsFromString(t, "# Other.")

		got := (<-items).Literal
		want := "# Other."

		if got != want {
			t.Errorf("got %q, want %q", got, want)
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
