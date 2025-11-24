package pony

import (
	"testing"
	"text/template"

	"github.com/charmbracelet/x/exp/golden"
)

func TestTemplateVariables(t *testing.T) {
	const markup = `<text>{{ .Message }}</text>`

	tmpl, err := Parse[map[string]interface{}](markup)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	data := map[string]interface{}{
		"Message": "Hello from template!",
	}

	output := tmpl.Render(data, 80, 24)
	golden.RequireEqual(t, output)
}

func TestTemplateRange(t *testing.T) {
	const markup = `
<vstack>
{{ range .Items }}
	<text>{{ . }}</text>
{{ end }}
</vstack>
`

	tmpl, err := Parse[map[string]interface{}](markup)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	data := map[string]interface{}{
		"Items": []string{"Item 1", "Item 2", "Item 3"},
	}

	output := tmpl.Render(data, 80, 24)
	golden.RequireEqual(t, output)
}

func TestTemplateConditional(t *testing.T) {
	const markup = `
<vstack>
{{ if .ShowMessage }}
	<text>{{ .Message }}</text>
{{ end }}
</vstack>
`

	tmpl, err := Parse[map[string]interface{}](markup)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	t.Run("condition true", func(t *testing.T) {
		data := map[string]interface{}{
			"ShowMessage": true,
			"Message":     "Visible",
		}

		output := tmpl.Render(data, 80, 24)
		golden.RequireEqual(t, output)
	})

	t.Run("condition false", func(t *testing.T) {
		data := map[string]interface{}{
			"ShowMessage": false,
			"Message":     "Visible",
		}

		output := tmpl.Render(data, 80, 24)
		golden.RequireEqual(t, output)
	})
}

func TestTemplateFunctions(t *testing.T) {
	tests := []struct {
		name   string
		markup string
		data   interface{}
	}{
		{
			name:   "upper function",
			markup: `<text>{{ upper .Text }}</text>`,
			data:   map[string]interface{}{"Text": "hello"},
		},
		{
			name:   "lower function",
			markup: `<text>{{ lower .Text }}</text>`,
			data:   map[string]interface{}{"Text": "HELLO"},
		},
		{
			name:   "add function",
			markup: `<text>{{ add .A .B }}</text>`,
			data:   map[string]interface{}{"A": 5, "B": 3},
		},
		{
			name:   "printf function",
			markup: `<text>{{ printf "Count: %d" .Count }}</text>`,
			data:   map[string]interface{}{"Count": 42},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpl, err := Parse[map[string]interface{}](tt.markup)
			if err != nil {
				t.Fatalf("Parse() error = %v", err)
			}

			output := tmpl.Render(tt.data.(map[string]interface{}), 80, 24)
			golden.RequireEqual(t, output)
		})
	}
}

func TestTemplateWithStyle(t *testing.T) {
	const markup = `<text style="fg:{{ .Color }}; bold">{{ .Message }}</text>`

	tmpl, err := Parse[map[string]interface{}](markup)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	data := map[string]interface{}{
		"Color":   "red",
		"Message": "Dynamic styled text",
	}

	output := tmpl.Render(data, 80, 24)
	golden.RequireEqual(t, output)
}

// Test MustParse functions.
func TestMustParse(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Error("MustParse should not panic on valid markup")
		}
	}()

	tmpl := MustParse[any]("<text>Test</text>")
	if tmpl == nil {
		t.Error("MustParse returned nil")
	}
}

func TestMustParsePanic(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("MustParse should panic on invalid template syntax")
		}
	}()

	MustParse[any]("<text>{{ .MissingFunc }}</text>{{ end }}")
}

func TestMustParseWithFuncs(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Error("MustParseWithFuncs should not panic on valid markup")
		}
	}()

	funcs := template.FuncMap{
		"test": func() string { return "test" },
	}

	tmpl := MustParseWithFuncs[any]("<text>{{ test }}</text>", funcs)
	if tmpl == nil {
		t.Error("MustParseWithFuncs returned nil")
	}
}

func TestMustParseWithFuncsPanic(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("MustParseWithFuncs should panic on invalid template syntax")
		}
	}()

	MustParseWithFuncs[any]("<text>{{ end }}</text>", nil)
}

// Test template functions coverage.
func TestTemplateFunctionsCoverage(t *testing.T) {
	funcs := defaultTemplateFuncs()

	// Test sub
	if sub := funcs["sub"].(func(int, int) int); sub(10, 3) != 7 {
		t.Error("sub function failed")
	}

	// Test mul
	if mul := funcs["mul"].(func(int, int) int); mul(3, 4) != 12 {
		t.Error("mul function failed")
	}

	// Test div
	if div := funcs["div"].(func(int, int) int); div(10, 2) != 5 {
		t.Error("div function failed")
	}

	// Test div by zero
	if div := funcs["div"].(func(int, int) int); div(10, 0) != 0 {
		t.Error("div by zero should return 0")
	}

	// Test repeat
	if repeat := funcs["repeat"].(func(string, int) string); repeat("x", 3) != "xxx" {
		t.Error("repeat function failed")
	}

	// Test join
	if join := funcs["join"].(func([]string, string) string); join([]string{"a", "b"}, ",") != "a,b" {
		t.Error("join function failed")
	}

	// Test trim
	if trim := funcs["trim"].(func(string) string); trim("  test  ") != "test" {
		t.Error("trim function failed")
	}

	// Test title
	if title := funcs["title"].(func(string) string); title("hello world") != "Hello World" {
		t.Error("title function failed")
	}

	// Test colorHex
	if colorHex := funcs["colorHex"].(func(string) string); colorHex("#FF0000") != "fg:#FF0000" {
		t.Error("colorHex function failed")
	}

	// Test bgHex
	if bgHex := funcs["bgHex"].(func(string) string); bgHex("#FF0000") != "bg:#FF0000" {
		t.Error("bgHex function failed")
	}
}
