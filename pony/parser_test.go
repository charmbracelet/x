package pony

import (
	"testing"
)

func TestParse(t *testing.T) {
	tests := []struct {
		name    string
		markup  string
		wantErr bool
	}{
		{
			name:    "simple text",
			markup:  "<text>Hello</text>",
			wantErr: false,
		},
		{
			name:    "vstack with children",
			markup:  "<vstack><text>Line 1</text><text>Line 2</text></vstack>",
			wantErr: false,
		},
		{
			name:    "hstack with gap",
			markup:  `<hstack gap="2"><text>Left</text><text>Right</text></hstack>`,
			wantErr: false,
		},
		{
			name:    "box with border",
			markup:  `<box border="rounded"><text>Content</text></box>`,
			wantErr: false,
		},
		{
			name:    "nested containers",
			markup:  "<vstack><hstack><text>A</text><text>B</text></hstack><text>C</text></vstack>",
			wantErr: false,
		},
		{
			name:    "self-closing spacer",
			markup:  "<spacer />",
			wantErr: false,
		},
		{
			name:    "self-closing divider",
			markup:  "<divider />",
			wantErr: false,
		},
		{
			name:    "multiline markup",
			markup:  "\n<vstack>\n  <text>Hello</text>\n</vstack>\n",
			wantErr: false,
		},
		{
			name:    "invalid xml",
			markup:  "<vstack><text>unclosed",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := parse(tt.markup)
			if (err != nil) != tt.wantErr {
				t.Errorf("parse() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestNodeToElement(t *testing.T) {
	tests := []struct {
		name   string
		markup string
		check  func(Element) bool
	}{
		{
			name:   "text element",
			markup: "<text>Hello</text>",
			check: func(e Element) bool {
				text, ok := e.(*Text)
				return ok && text.Content == "Hello"
			},
		},
		{
			name:   "vstack element",
			markup: "<vstack gap=\"2\"><text>A</text></vstack>",
			check: func(e Element) bool {
				vstack, ok := e.(*VStack)
				return ok && vstack.Gap == 2 && len(vstack.Items) == 1
			},
		},
		{
			name:   "hstack element",
			markup: "<hstack gap=\"1\"><text>A</text><text>B</text></hstack>",
			check: func(e Element) bool {
				hstack, ok := e.(*HStack)
				return ok && hstack.Gap == 1 && len(hstack.Items) == 2
			},
		},
		{
			name:   "box element",
			markup: "<box border=\"rounded\"><text>Content</text></box>",
			check: func(e Element) bool {
				box, ok := e.(*Box)
				return ok && box.Border == "rounded" && box.Child != nil
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			node, err := parse(tt.markup)
			if err != nil {
				t.Fatalf("parse() error = %v", err)
			}

			elem := node.toElement()
			if elem == nil {
				t.Fatal("toElement() returned nil")
			}

			if !tt.check(elem) {
				t.Errorf("element check failed for %T", elem)
			}
		})
	}
}

func TestPropsHelpers(t *testing.T) {
	props := Props{
		"foo": "bar",
		"num": "42",
	}

	if props.Get("foo") != "bar" {
		t.Errorf("Get(foo) = %q, want %q", props.Get("foo"), "bar")
	}

	if props.Get("missing") != "" {
		t.Errorf("Get(missing) = %q, want empty", props.Get("missing"))
	}

	if props.GetOr("missing", "default") != "default" {
		t.Errorf("GetOr(missing) = %q, want %q", props.GetOr("missing", "default"), "default")
	}

	if !props.Has("foo") {
		t.Error("Has(foo) = false, want true")
	}

	if props.Has("missing") {
		t.Error("Has(missing) = true, want false")
	}
}
