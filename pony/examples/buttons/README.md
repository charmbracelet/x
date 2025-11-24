# Button Click Example

This example demonstrates mouse click handling in pony using Bubble Tea's View callback feature.

## Features

- **Interactive Buttons**: Click buttons with your mouse
- **Hit Testing**: Automatically detects which element was clicked
- **Hover Detection**: Shows which element is under the cursor
- **Stateless View**: View function remains pure using callbacks

## How It Works

1. **BoundsMap**: During rendering, every element records where it was drawn
2. **View Callback**: The View returns a callback that has access to the BoundsMap via closure
3. **Hit Testing**: When mouse events arrive, the callback uses `HitTest()` to find the clicked element
4. **Event Messages**: The callback returns commands that emit custom messages with element IDs
5. **Update**: The Update function handles these messages to update model state

## Requirements

This example uses Bubble Tea PR #1549 which adds View callback support. The go.mod pins to the specific commit that includes this feature.

## Running

```bash
go run main.go
```

Click the buttons to see mouse interaction in action!

## Architecture

```
View() renders -> returns (screen, boundsMap)
              -> creates View with callback that captures boundsMap
              -> callback does hit testing on mouse events
              -> returns Cmd with element ID
Update() receives element ID message
         -> updates model based on which button was clicked
```

This keeps View() completely pure - no model mutation, just returning a callback closure.
