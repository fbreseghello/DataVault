package cli

import (
	"slices"
	"testing"
)

func TestTitleAndText(t *testing.T) {
	for _, c := range []struct {
		args        []string
		title, text string
	}{
		{[]string{"market(milk and bread)"}, "market", "milk and bread"},
		{[]string{"market", "(milk", "and", "bread)"}, "market", "milk and bread"},
		{[]string{"call mom (sunday, 10h)"}, "call mom", "sunday, 10h"},
		{[]string{"deploy"}, "deploy", ""},
		{[]string{"deploy", "the api"}, "deploy", "the api"},
		{[]string{"deploy", "the", "api"}, "deploy", "the api"},
		{[]string{"(only text)"}, "(only text)", ""},
		{nil, "", ""},
	} {
		title, text := titleAndText(c.args)
		if title != c.title || text != c.text {
			t.Errorf("%q: got %q %q, want %q %q", c.args, title, text, c.title, c.text)
		}
	}
}

func TestSplitArgs(t *testing.T) {
	got := splitArgs(`note "buy milk"  'and bread' plain ""`)
	want := []string{"note", "buy milk", "and bread", "plain", ""}
	if !slices.Equal(got, want) {
		t.Fatalf("got %q, want %q", got, want)
	}
}
