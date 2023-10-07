package parser_test

import (
	"strings"
	"testing"

	"github.com/AuroraFoundation/ALF/pkg/parser"
)

func TestParser(t *testing.T) {
	cases := []struct {
		desc string
		test string
		want parser.ALF
	}{
		{
			"comment",
			"# A comment.",
			parser.ALF{},
		},
		{
			"attribute title",
			"Title: Other Test.",
			parser.ALF{
				Title: "Other Test.",
			},
		},
		{
			"attribute with comment",
			"# Test.\nTitle: Test",
			parser.ALF{
				Title: "Test",
			},
		},
		{
			"attribute author",
			"Author: The author...",
			parser.ALF{
				Author: "The author...",
			},
		},
		{
			"attribute artist",
			"Artist: Other Artist.",
			parser.ALF{
				Artist: "Other Artist.",
			},
		},
		{
			"attribute album",
			"Album: An album.",
			parser.ALF{
				Album: "An album.",
			},
		},
		{
			"multiple attributes",
			"Title: The title.\nArtist: Gopher.",
			parser.ALF{
				Title:  "The title.",
				Artist: "Gopher.",
			},
		},
	}

	for _, test := range cases {
		t.Run(test.desc, func(t *testing.T) {
			got := alfFromString(t, test.test)
			assertALF(t, got, test.want)
		})
	}
}

func alfFromString(t *testing.T, s string) parser.ALF {
	t.Helper()

	alf, err := parser.New(strings.NewReader(s)).Decode()
	if err != nil {
		t.Fatal(err)
	}

	return alf
}

func assertALF(t *testing.T, got, want parser.ALF) {
	t.Helper()

	if got != want {
		t.Logf("Got:  %#v", got)
		t.Logf("Want: %#v", want)
		t.Error()
	}
}
