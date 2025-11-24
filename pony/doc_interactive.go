// Package pony provides examples and patterns for building interactive TUI components.
//
// # Building Interactive Components
//
// When creating custom components that render other elements, you must pass through
// your component's ID to the root element you return. This ensures that mouse clicks
// anywhere in your component return your component's ID, not child element IDs.
//
// ## Pattern: Pass Through Component ID
//
//	type MyComponent struct {
//	    pony.BaseElement  // Provides ID() and SetID()
//	    // ... your fields
//	}
//
//	func (c *MyComponent) Render() pony.Element {
//	    // Build your UI
//	    root := pony.NewVStack(
//	        pony.NewText(c.label),
//	        pony.NewBox(pony.NewText(c.value)),
//	    )
//
//	    // CRITICAL: Set your component's ID on the root element
//	    root.SetID(c.ID())
//
//	    return root
//	}
//
// ## Why This Matters
//
// HitTest() prefers elements with explicit IDs over auto-generated ones.
// When multiple elements overlap at a click point:
//
//  1. Without SetID: Returns child element with auto-generated ID like "elem_0x123..."
//  2. With SetID: Returns your component with meaningful ID like "name-input"
//
// This allows you to handle clicks on the component as a whole, not individual children.
//
// ## Example Usage
//
//	input := NewInput("Name:")
//	input.SetID("name-input")
//
//	// In template
//	slots := map[string]pony.Element{
//	    "input": input.Render(),  // VStack with ID "name-input"
//	}
//
//	// In callback
//	view.Callback = func(msg tea.Msg) tea.Cmd {
//	    if click, ok := msg.(tea.MouseClickMsg); ok {
//	        elem := boundsMap.HitTest(click.X, click.Y)
//	        // elem.ID() returns "name-input" - exactly what you want!
//	    }
//	}
//
// See examples/interactive-form for a complete working example.
package pony
