package lexer_test

import (
	"testing"

	"github.com/AuroraFoundation/ALF/pkg/lexer"
)

func TestItemString(t *testing.T) {
	want := `<Item (Comment)[3:9] "# A comment.">`
	got := (lexer.Item{
		Token:   lexer.TokenComment,
		Literal: "# A comment.",
		Line:    3,
		Col:     9,
	}).String()

	assertString(t, got, want)
}
