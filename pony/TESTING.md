# Testing Guide

pony uses Go's standard testing framework with golden file testing for output verification.

## Running Tests

```bash
# Run all tests
go test ./...

# Run specific test
go test -run TestRender

# Run with verbose output
go test -v

# Run with coverage
go test -cover
```

## Golden File Testing

Many rendering tests use **golden file testing** to verify output. Golden files are stored in `testdata/` and contain the expected rendered output.

### Updating Golden Files

When you intentionally change rendering output, update the golden files:

```bash
go test -update
```

This will regenerate all `testdata/*.golden` files with the new output.

### How It Works

1. **First run**: Run tests with `-update` to create golden files
2. **Subsequent runs**: Tests compare output against golden files
3. **On mismatch**: Test fails with a diff showing what changed
4. **Update when needed**: Use `-update` flag to accept new output

### Example

```go
func TestMyRender(t *testing.T) {
    output := myTemplate.Render(data, 80, 24)
    golden.RequireEqual(t, output)  // Compares against testdata/TestMyRender.golden
}
```

### Benefits

- ✅ Catches unintended rendering changes
- ✅ Easy to review diffs
- ✅ Preserves ANSI codes and formatting
- ✅ Works with subtests

## Test Organization

- **element_test.go** - Element layout and rendering tests
- **parser_test.go** - XML parsing tests
- **style_test.go** - Style parsing and rendering tests
- **template_test.go** - Go template integration tests
- **layout_test.go** - Size constraint tests
- **alignment_test.go** - Text and container alignment tests
- **registry_test.go** - Component registry and built-in component tests
- **slot_test.go** - Slot system tests
- **scrollview_test.go** - Scrolling functionality tests
- **helpers_test.go** - Style and layout helper tests

## Coverage

Current test coverage: **129 test cases**, all passing.

Areas covered:
- ✅ XML parsing
- ✅ Element rendering
- ✅ Layout calculations
- ✅ Style parsing
- ✅ Template execution
- ✅ Component system
- ✅ Slot injection
- ✅ Scrolling
- ✅ Helpers

## Writing New Tests

### For Rendering Output

Use golden testing:

```go
func TestNewFeature(t *testing.T) {
    output := tmpl.Render(data, width, height)
    golden.RequireEqual(t, output)
}
```

### For Logic/Behavior

Use standard assertions:

```go
func TestLayout(t *testing.T) {
    size := elem.Layout(constraints)
    if size.Width != expected {
        t.Errorf("got %d, want %d", size.Width, expected)
    }
}
```

## Tips

- Use subtests for multiple cases: `t.Run(name, func(t *testing.T) { ... })`
- Golden files are stored in `testdata/TestName.golden` or `testdata/TestName/subtest.golden`
- Run `-update` after intentional changes to rendering
- Review golden file diffs in version control carefully
- Keep test names descriptive for clear golden file names
