package strings

import "testing"

func TestEnglishJoin(t *testing.T) {
	for i, tc := range []struct {
		words       []string
		lang        Language
		oxfordComma bool
		expected    string
	}{
		{
			words:       []string{"one", "two", "three"},
			lang:        EN,
			oxfordComma: true,
			expected:    "one, two, and three",
		},
		{
			words:       []string{"one", "two", "three", "four"},
			oxfordComma: true,
			expected:    "one, two, three, and four",
		},
		{
			words:       []string{"one", "two"},
			oxfordComma: true,
			expected:    "one and two",
		},
		{
			words:       []string{"one", "two", "three"},
			oxfordComma: false,
			expected:    "one, two and three",
		},
		{
			words:       []string{"one"},
			oxfordComma: true,
			expected:    "one",
		},
	} {
		actual := EnglishJoin(tc.words, tc.oxfordComma)
		if actual != tc.expected {
			t.Errorf("Test #%d:\n  expected: %q\n  got:      %q", i+1, tc.expected, actual)
		}
	}
}

func TestSpokenLanguageJoin(t *testing.T) {
	for i, tc := range []struct {
		words    []string
		lang     Language
		expected string
	}{
		// Test for correct commas and conjunctions in each language.
		{
			words:    []string{"eins", "zwei", "drei"},
			lang:     DE,
			expected: "eins, zwei und drei",
		},
		{
			words:    []string{"en", "to", "tre"},
			lang:     DK,
			expected: "en, to og tre",
		},
		{
			words:    []string{"one", "two", "three"},
			lang:     EN,
			expected: "one, two and three",
		},
		{
			words:    []string{"uno", "dos", "tres"},
			lang:     ES,
			expected: "uno, dos y tres",
		},
		{
			words:    []string{"un", "deux", "trois"},
			lang:     FR,
			expected: "un, deux et trois",
		},
		{
			words:    []string{"uno", "due", "tre"},
			lang:     IT,
			expected: "uno, due e tre",
		},
		{
			words:    []string{"en", "to", "tre"},
			lang:     NO,
			expected: "en, to og tre",
		},
		{
			words:    []string{"um", "dois", "três"},
			lang:     PT,
			expected: "um, dois e três",
		},
		{
			words:    []string{"ett", "två", "tre", "fyra"},
			lang:     SE,
			expected: "ett, två, tre och fyra",
		},

		// Test other things.
		{
			words:    []string{"one", "two", "three", "four"},
			lang:     EN,
			expected: "one, two, three and four",
		},
		{
			words:    []string{"um", "dois"},
			lang:     PT,
			expected: "um e dois",
		},
		{
			words:    []string{"un", "deux"},
			lang:     FR,
			expected: "un et deux",
		},
		{
			words:    []string{"one"},
			lang:     EN,
			expected: "one",
		},
		{
			words:    []string{},
			lang:     EN,
			expected: "",
		},
		{
			words:    []string{"", "", ""},
			lang:     EN,
			expected: "",
		},
		{
			words:    nil,
			lang:     EN,
			expected: "",
		},
	} {
		actual := SpokenLanguageJoin(tc.words, tc.lang)
		if actual != tc.expected {
			t.Errorf("Test #%d:\n  expected: %q\n  got:      %q", i+1, tc.expected, actual)
		}
	}
}
