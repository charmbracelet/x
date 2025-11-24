package pony

import (
	"testing"

	"github.com/charmbracelet/x/exp/golden"
)

// ZStack tests

func TestZStackBasic(t *testing.T) {
	const markup = `
<zstack>
	<box border="rounded" width="20" height="5">
		<text>Background</text>
	</box>
	<box width="10" height="3">
		<text font-weight="bold">Overlay</text>
	</box>
</zstack>
`

	tmpl, err := Parse[any](markup)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	output := tmpl.Render(nil, 40, 10)
	golden.RequireEqual(t, output)
}

func TestZStackAlignment(t *testing.T) {
	tests := []struct {
		name   string
		align  string
		valign string
	}{
		{"top-left", "left", "top"},
		{"top-center", "center", "top"},
		{"top-right", "right", "top"},
		{"middle-left", "left", "middle"},
		{"middle-center", "center", "middle"},
		{"middle-right", "right", "middle"},
		{"bottom-left", "left", "bottom"},
		{"bottom-center", "center", "bottom"},
		{"bottom-right", "right", "bottom"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			children := []Element{
				NewBox(NewText("Background")).Border("rounded").Width(NewFixedConstraint(30)).Height(NewFixedConstraint(10)),
				NewBox(NewText(tt.name)).Border("normal").Width(NewFixedConstraint(15)).Height(NewFixedConstraint(3)),
			}

			elem := NewZStack(children...).
				Alignment(tt.align).
				VerticalAlignment(tt.valign)

			// Use unbounded constraints so ZStack sizes to its children
			size := elem.Layout(Unbounded())
			if size.Width != 30 || size.Height != 10 {
				t.Errorf("Layout() = %v, want {30 10}", size)
			}
		})
	}
}

func TestZStackLayout(t *testing.T) {
	tests := []struct {
		name       string
		children   []Element
		wantWidth  int
		wantHeight int
	}{
		{
			name:       "empty zstack",
			children:   []Element{},
			wantWidth:  0,
			wantHeight: 0,
		},
		{
			name: "single child",
			children: []Element{
				NewText("Hello"),
			},
			wantWidth:  5,
			wantHeight: 1,
		},
		{
			name: "multiple children - max size",
			children: []Element{
				NewText("Short"),
				NewText("Much longer text"),
				NewText("Mid"),
			},
			wantWidth:  16, // longest child
			wantHeight: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			elem := NewZStack(tt.children...)
			size := elem.Layout(Unbounded())

			if size.Width != tt.wantWidth {
				t.Errorf("Layout().Width = %d, want %d", size.Width, tt.wantWidth)
			}
			if size.Height != tt.wantHeight {
				t.Errorf("Layout().Height = %d, want %d", size.Height, tt.wantHeight)
			}
		})
	}
}

// Margin tests

func TestBoxMargin(t *testing.T) {
	const markup = `
<vstack>
	<box border="rounded" margin="2">
		<text>With margin</text>
	</box>
	<box border="rounded">
		<text>No margin</text>
	</box>
</vstack>
`

	tmpl, err := Parse[any](markup)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	output := tmpl.Render(nil, 30, 15)
	golden.RequireEqual(t, output)
}

func TestBoxMarginSides(t *testing.T) {
	const markup = `
<vstack>
	<box border="rounded" margin-left="5" margin-right="2">
		<text>Horizontal margins</text>
	</box>
	<box border="rounded" margin-top="2" margin-bottom="1">
		<text>Vertical margins</text>
	</box>
</vstack>
`

	tmpl, err := Parse[any](markup)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	output := tmpl.Render(nil, 40, 20)
	golden.RequireEqual(t, output)
}

func TestBoxMarginLayout(t *testing.T) {
	child := NewText("Hello")
	childSize := child.Layout(Unbounded())

	tests := []struct {
		name         string
		margin       int
		marginTop    int
		marginRight  int
		marginBottom int
		marginLeft   int
		wantWidth    int
		wantHeight   int
	}{
		{
			name:       "uniform margin",
			margin:     2,
			wantWidth:  childSize.Width + 4,  // 2*2
			wantHeight: childSize.Height + 4, // 2*2
		},
		{
			name:        "horizontal margins",
			marginLeft:  3,
			marginRight: 2,
			wantWidth:   childSize.Width + 5,  // 3+2
			wantHeight:  childSize.Height + 0, // no vertical margin
		},
		{
			name:         "vertical margins",
			marginTop:    1,
			marginBottom: 2,
			wantWidth:    childSize.Width + 0,  // no horizontal margin
			wantHeight:   childSize.Height + 3, // 1+2
		},
		{
			name:         "all sides different",
			marginTop:    1,
			marginRight:  2,
			marginBottom: 3,
			marginLeft:   4,
			wantWidth:    childSize.Width + 6,  // 4+2
			wantHeight:   childSize.Height + 4, // 1+3
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			elem := NewBox(child).
				Border("none").
				Margin(tt.margin).
				MarginTop(tt.marginTop).
				MarginRight(tt.marginRight).
				MarginBottom(tt.marginBottom).
				MarginLeft(tt.marginLeft)

			size := elem.Layout(Unbounded())

			if size.Width != tt.wantWidth {
				t.Errorf("Layout().Width = %d, want %d", size.Width, tt.wantWidth)
			}
			if size.Height != tt.wantHeight {
				t.Errorf("Layout().Height = %d, want %d", size.Height, tt.wantHeight)
			}
		})
	}
}

// Flex tests

func TestFlexGrow(t *testing.T) {
	const markup = `
<hstack>
	<box border="rounded" width="10">
		<text>Fixed</text>
	</box>
	<flex grow="1">
		<box border="rounded">
			<text>Flex 1</text>
		</box>
	</flex>
	<flex grow="2">
		<box border="rounded">
			<text>Flex 2</text>
		</box>
	</flex>
</hstack>
`

	tmpl, err := Parse[any](markup)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	output := tmpl.Render(nil, 60, 5)
	golden.RequireEqual(t, output)
}

func TestFlexVStack(t *testing.T) {
	const markup = `
<vstack>
	<text>Header</text>
	<flex grow="1">
		<box border="rounded">
			<text>Growing content</text>
		</box>
	</flex>
	<text>Footer</text>
</vstack>
`

	tmpl, err := Parse[any](markup)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	output := tmpl.Render(nil, 30, 15)
	golden.RequireEqual(t, output)
}

func TestFlexLayout(t *testing.T) {
	tests := []struct {
		name   string
		grow   int
		shrink int
		basis  int
	}{
		{"no grow", 0, 1, 0},
		{"grow 1", 1, 1, 0},
		{"grow 2", 2, 1, 0},
		{"with basis", 1, 1, 20},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			child := NewText("Content")
			flex := NewFlex(child).
				Grow(tt.grow).
				Shrink(tt.shrink).
				Basis(tt.basis)

			size := flex.Layout(Unbounded())
			if size.Width == 0 && size.Height == 0 && tt.basis == 0 {
				t.Error("Flex layout should not return zero size without basis")
			}
		})
	}
}

func TestGetFlexGrow(t *testing.T) {
	tests := []struct {
		name string
		elem Element
		want int
	}{
		{
			name: "flex element",
			elem: NewFlex(NewText("x")).Grow(2),
			want: 2,
		},
		{
			name: "flexible spacer",
			elem: NewSpacer(),
			want: 1,
		},
		{
			name: "fixed spacer",
			elem: NewFixedSpacer(10),
			want: 0,
		},
		{
			name: "regular element",
			elem: NewText("x"),
			want: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetFlexGrow(tt.elem)
			if got != tt.want {
				t.Errorf("GetFlexGrow() = %d, want %d", got, tt.want)
			}
		})
	}
}

// Positioned tests

func TestPositionedBasic(t *testing.T) {
	const markup = `
<zstack>
	<box border="rounded" width="40" height="10">
		<text>Background</text>
	</box>
	<positioned x="5" y="2">
		<text font-weight="bold">At (5,2)</text>
	</positioned>
	<positioned x="20" y="5">
		<text font-weight="bold">At (20,5)</text>
	</positioned>
</zstack>
`

	tmpl, err := Parse[any](markup)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	output := tmpl.Render(nil, 50, 15)
	golden.RequireEqual(t, output)
}

func TestPositionedCorners(t *testing.T) {
	const markup = `
<zstack>
	<box border="rounded" width="40" height="12">
		<text>Main Content</text>
	</box>
	<positioned x="1" y="1">
		<text font-weight="bold">TL</text>
	</positioned>
	<positioned right="1" y="1">
		<text font-weight="bold">TR</text>
	</positioned>
	<positioned x="1" bottom="1">
		<text font-weight="bold">BL</text>
	</positioned>
	<positioned right="1" bottom="1">
		<text font-weight="bold">BR</text>
	</positioned>
</zstack>
`

	tmpl, err := Parse[any](markup)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	output := tmpl.Render(nil, 50, 15)
	golden.RequireEqual(t, output)
}

func TestPositionedLayout(t *testing.T) {
	child := NewText("Positioned")
	elem := NewPositioned(child, 10, 5)

	// Positioned elements return zero size in layout (out of flow)
	size := elem.Layout(Unbounded())
	if size.Width != 0 || size.Height != 0 {
		t.Errorf("Positioned Layout() = %v, want {0 0}", size)
	}
}

// Combined features test

func TestAdvancedLayoutCombined(t *testing.T) {
	const markup = `
<zstack>
	<vstack>
		<box border="rounded" margin="1">
			<text>Main content</text>
		</box>
		<flex grow="1">
			<box border="normal" margin-left="2" margin-right="2">
				<text>Flexible growing section</text>
			</box>
		</flex>
		<box margin="1">
			<text>Footer</text>
		</box>
	</vstack>
	<positioned right="2" y="2">
		<box border="thick" padding="1">
			<text font-weight="bold" foreground-color="cyan">Overlay</text>
		</box>
	</positioned>
</zstack>
`

	tmpl, err := Parse[any](markup)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	output := tmpl.Render(nil, 60, 20)
	golden.RequireEqual(t, output)
}

// Test ZStack chaining methods.
func TestZStackWithMethods(t *testing.T) {
	zstack := NewZStack(NewText("a"), NewText("b"))

	zstack.Alignment(AlignmentTrailing)
	zstack.VerticalAlignment(AlignmentBottom)
	zstack.Width(NewFixedConstraint(10))
	zstack.Height(NewFixedConstraint(5))

	// Test Children
	children := zstack.Children()
	if len(children) != 2 {
		t.Error("ZStack Children should return items")
	}
}

// Test Positioned chaining methods.
func TestPositionedWithMethods(t *testing.T) {
	pos := NewPositioned(NewText("test"), 5, 5)

	pos.Right(10)
	pos.Bottom(15)
	pos.Width(NewFixedConstraint(20))
	pos.Height(NewFixedConstraint(25))

	// Test Children
	children := pos.Children()
	if len(children) != 1 {
		t.Error("Positioned Children should return child")
	}

	// Test with nil child
	posNil := NewPositioned(nil, 0, 0)
	if posNil.Children() != nil {
		t.Error("Positioned Children with nil child should return nil")
	}
}
