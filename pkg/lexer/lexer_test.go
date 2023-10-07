package lexer_test

import (
	"errors"
	"io"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/AuroraFoundation/ALF/pkg/lexer"
)

func TestLexerOverview(t *testing.T) {
	file, err := os.Open("testdata/overview.alf")
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()

	// In theory, if by separating the file into tokens and rebuilding it from
	// these, we could be ensuring almost perfect functioning of the lexer, or
	// at least the analysis of the strings.

	var original strings.Builder
	r := io.TeeReader(file, &original)

	_, items := lexer.New(r)

	var fromLexer strings.Builder
	for {
		item := <-items

		if item.Token == lexer.TokenEOF {
			break
		}

		fromLexer.WriteString(item.Literal)
	}

	if original.String() != fromLexer.String() {
		t.Log("Original:\n", original.String())
		t.Log("Lexer:\n", fromLexer.String())
		t.Error("The strings are different.")
	}
}

func TestLexerItem(t *testing.T) {
	cases := []struct {
		desc string
		test string
		want []lexer.Item
	}{
		{
			"basic attribute-value",
			"Name: Hello",
			[]lexer.Item{
				{lexer.TokenName, "Name", 0, 0},
				{lexer.TokenColon, ":", 0, 4},
				{lexer.TokenWhitespace, " ", 0, 5},
				{lexer.TokenText, "Hello", 0, 6},
			},
		},
		{
			"basic comment",
			"# Comment.",
			[]lexer.Item{
				{lexer.TokenComment, "# Comment.", 0, 0},
			},
		},
		{
			"nested attribute",
			"Attribute:\n\tNestedA:\n\t\tNestedB:",
			[]lexer.Item{
				{lexer.TokenName, "Attribute", 0, 0},
				{lexer.TokenColon, ":", 0, 9},
				{lexer.TokenNewline, "\n", 0, 10},
				{lexer.TokenIndent, "\t", 1, 0},
				{lexer.TokenName, "NestedA", 1, 1},
				{lexer.TokenColon, ":", 1, 8},
				{lexer.TokenNewline, "\n", 1, 9},
				{lexer.TokenIndent, "\t\t", 2, 0},
				{lexer.TokenName, "NestedB", 2, 2},
				{lexer.TokenColon, ":", 2, 9},
			},
		},
		{
			"lists",
			"Notes:\n\t- One Note.\n\t- Other Note.",
			[]lexer.Item{
				{lexer.TokenName, "Notes", 0, 0},
				{lexer.TokenColon, ":", 0, 5},
				{lexer.TokenNewline, "\n", 0, 6},
				{lexer.TokenIndent, "\t", 1, 0},
				{lexer.TokenList, "-", 1, 1},
				{lexer.TokenWhitespace, " ", 1, 2},
				{lexer.TokenText, "One Note.", 1, 3},
				{lexer.TokenNewline, "\n", 1, 12},
				{lexer.TokenIndent, "\t", 2, 0},
				{lexer.TokenList, "-", 2, 1},
				{lexer.TokenWhitespace, " ", 2, 2},
				{lexer.TokenText, "Other Note.", 2, 3},
			},
		},
	}

	for _, test := range cases {
		t.Run(test.desc, func(t *testing.T) {
			items := itemsFromString(t, test.test)

			var got []lexer.Item

			for {
				item := <-items

				if item.Token == lexer.TokenEOF {
					break
				}

				got = append(got, item)
			}

			if !reflect.DeepEqual(got, test.want) {
				t.Log("Got: ", got)
				t.Log("Want:", test.want)
				t.Error()
			}
		})
	}
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

	t.Run("nested", func(t *testing.T) {
		items := itemsFromString(t, "\tNested:")
		want := lexer.TokenName

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

// readerError is a mock for check error messages.
type readerError string

// Read always return a error.
func (r readerError) Read([]byte) (int, error) {
	return 0, errors.New(string(r))
}

func TestLexerError(t *testing.T) {
	t.Run("token", func(t *testing.T) {
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

// Helpers.

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
