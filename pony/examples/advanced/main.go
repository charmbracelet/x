package main

import (
	"fmt"

	"github.com/charmbracelet/x/pony"
)

func main() {
	// Example demonstrating advanced layout features:
	// - ZStack for layering
	// - Margin support
	// - Flex-grow for flexible sizing
	// - Positioned elements for absolute positioning

	const markup = `
<zstack>
	<!-- Main content with vertical layout -->
	<vstack width="60" height="20">
		<!-- Header with margin -->
		<box border="rounded" margin="1">
			<text style="bold; fg:cyan" align="center">Advanced Layout Demo</text>
		</box>
		
		<!-- Flexible content area that grows to fill space -->
		<flex grow="1">
			<box border="normal" margin-left="2" margin-right="2">
				<vstack gap="1">
					<text style="fg:green">✓ ZStack for layered layouts</text>
					<text style="fg:green">✓ Margin support (all sides)</text>
					<text style="fg:green">✓ Flex-grow for flexible sizing</text>
					<text style="fg:green">✓ Absolute positioning</text>
				</vstack>
			</box>
		</flex>
		
		<!-- Footer with margin -->
		<box margin-top="1" margin-bottom="1">
			<text style="fg:gray" align="center">Press any key to exit</text>
		</box>
	</vstack>
	
	<!-- Overlay notification in top-right corner -->
	<positioned right="2" y="2">
		<box border="thick" padding="1" border-style="fg:yellow; bold">
			<text style="bold; fg:yellow">NEW!</text>
		</box>
	</positioned>
	
	<!-- Bottom-right status indicator -->
	<positioned right="3" bottom="2">
		<box border="rounded" padding="1">
			<text style="fg:cyan">v1.0</text>
		</box>
	</positioned>
</zstack>
`

	// Parse the template
	tmpl, err := pony.Parse[any](markup)
	if err != nil {
		panic(err)
	}

	// Render at 80x24 (standard terminal size)
	output := tmpl.Render(nil, 80, 24)
	fmt.Print(output)
}
