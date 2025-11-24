package pony

import (
	"testing"

	"github.com/charmbracelet/x/exp/golden"
)

func TestSizeConstraint(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		available int
		content   int
		want      int
	}{
		{
			name:      "auto uses content",
			input:     "auto",
			available: 100,
			content:   50,
			want:      50,
		},
		{
			name:      "auto limited by available",
			input:     "auto",
			available: 30,
			content:   50,
			want:      30,
		},
		{
			name:      "50 percent",
			input:     "50%",
			available: 100,
			content:   10,
			want:      50,
		},
		{
			name:      "100 percent",
			input:     "100%",
			available: 80,
			content:   10,
			want:      80,
		},
		{
			name:      "fixed size",
			input:     "20",
			available: 100,
			content:   10,
			want:      20,
		},
		{
			name:      "fixed size larger than available",
			input:     "150",
			available: 100,
			content:   10,
			want:      100,
		},
		{
			name:      "empty string is auto",
			input:     "",
			available: 100,
			content:   30,
			want:      30,
		},
		{
			name:      "max takes available",
			input:     "max",
			available: 100,
			content:   30,
			want:      100,
		},
		{
			name:      "min takes content",
			input:     "min",
			available: 100,
			content:   30,
			want:      30,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sc := parseSizeConstraint(tt.input)
			got := sc.Apply(tt.available, tt.content)
			if got != tt.want {
				t.Errorf("Apply(%d, %d) = %d, want %d (constraint: %s)",
					tt.available, tt.content, got, tt.want, sc.String())
			}
		})
	}
}

func TestBoxWithWidth(t *testing.T) {
	const markup = `
<hstack>
	<box border="normal" width="30%">
		<text>30%</text>
	</box>
	<box border="normal" width="70%">
		<text>70%</text>
	</box>
</hstack>
`

	tmpl, err := Parse[any](markup)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	output := tmpl.Render(nil, 100, 10)
	golden.RequireEqual(t, output)
}

func TestZeroValueIsAuto(t *testing.T) {
	// Test that zero value SizeConstraint behaves like auto
	sc := SizeConstraint{}

	if !sc.IsAuto() {
		t.Error("Zero value SizeConstraint should be auto")
	}

	got := sc.Apply(100, 50)
	if got != 50 {
		t.Errorf("Zero value Apply(100, 50) = %d, want 50 (content size)", got)
	}
}

// Test size constraint methods.
func TestSizeConstraintMethods(t *testing.T) {
	// IsFixed
	sc := NewFixedConstraint(10)
	if !sc.IsFixed() {
		t.Error("IsFixed should return true for fixed constraint")
	}

	// IsPercent
	sc = NewPercentConstraint(50)
	if !sc.IsPercent() {
		t.Error("IsPercent should return true for percent constraint")
	}

	// String - fixed
	sc = NewFixedConstraint(10)
	if sc.String() != "10" {
		t.Errorf("String() = %s, want 10", sc.String())
	}

	// String - percent
	sc = NewPercentConstraint(50)
	if sc.String() != "50%" {
		t.Errorf("String() = %s, want 50%%", sc.String())
	}

	// String - auto
	sc = parseSizeConstraint("auto")
	if sc.String() != "auto" {
		t.Errorf("String() = %s, want auto", sc.String())
	}

	// String - min
	sc = parseSizeConstraint("min")
	if sc.String() != "min" {
		t.Errorf("String() = %s, want min", sc.String())
	}

	// String - max
	sc = parseSizeConstraint("max")
	if sc.String() != "max" {
		t.Errorf("String() = %s, want max", sc.String())
	}
}

// Test VStack chaining methods.
func TestVStackWithMethods(t *testing.T) {
	vstack := NewVStack(NewText("a"), NewText("b"))

	vstack.WithGap(2)
	if vstack.Gap != 2 {
		t.Error("WithGap not set")
	}

	vstack.WithAlign(AlignCenter)
	if vstack.Align != AlignCenter {
		t.Error("WithAlign not set")
	}

	vstack.WithWidth(NewFixedConstraint(10))
	if vstack.Width.IsAuto() {
		t.Error("WithWidth not set")
	}

	vstack.WithHeight(NewFixedConstraint(5))
	if vstack.Height.IsAuto() {
		t.Error("WithHeight not set")
	}
}

// Test HStack constructors and chaining methods.
func TestHStackConstructor(t *testing.T) {
	hstack := NewHStack(NewText("a"), NewText("b"))
	if hstack == nil {
		t.Fatal("NewHStack returned nil")
	}
	if len(hstack.Items) != 2 {
		t.Error("NewHStack items not set")
	}

	hstack.WithGap(2)
	if hstack.Gap != 2 {
		t.Error("WithGap not set")
	}

	hstack.WithValign(AlignMiddle)
	if hstack.Valign != AlignMiddle {
		t.Error("WithValign not set")
	}

	hstack.WithWidth(NewFixedConstraint(10))
	if hstack.Width.IsAuto() {
		t.Error("WithWidth not set")
	}

	hstack.WithHeight(NewFixedConstraint(5))
	if hstack.Height.IsAuto() {
		t.Error("WithHeight not set")
	}

	// Test Children
	children := hstack.Children()
	if len(children) != 2 {
		t.Error("HStack Children should return items")
	}
}
