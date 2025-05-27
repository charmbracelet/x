// Package main is a open example.
package main

import (
	"context"
	"fmt"

	"github.com/charmbracelet/x/exp/open"
)

func main() {
	fmt.Println(open.Open(context.Background(), "https://charm.sh"))
}
