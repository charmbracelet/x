package editor

import (
	"reflect"
	"testing"
)

func TestEditor(t *testing.T) {
	filename := "README.md"
	for k, v := range map[string][]string{
		"":             {"nano", filename},
		"nvim":         {"nvim", filename},
		"vim":          {"vim", filename},
		"vscode --foo": {"vscode", "--foo", filename},
		"nvim -a -b":   {"nvim", "-a", "-b", filename},
	} {
		t.Run(k, func(t *testing.T) {
			t.Setenv("EDITOR", k)
			cmd, _ := Cmd("X", "README.md")
			got := cmd.Args
			if !reflect.DeepEqual(got, v) {
				t.Fatalf("expected %v; got %v", v, got)
			}
		})
	}

	t.Run("with line number", func(t *testing.T) {
		for k, v := range map[string][]string{
			"":             {"nano", "+12", filename},
			"nvim":         {"nvim", "+12", filename},
			"vim":          {"vim", "+12", filename},
			"vscode --foo": {"vscode", "--foo", filename},
			"nvim -a -b":   {"nvim", "-a", "-b", "+12", filename},
			"code --foo":   {"code", "--foo", "--goto", filename + ":12"},
		} {
			t.Run(k, func(t *testing.T) {
				t.Setenv("EDITOR", k)
				cmd, _ := Cmd("X", "README.md", OpenAtLine(12))
				got := cmd.Args
				if !reflect.DeepEqual(got, v) {
					t.Fatalf("expected %v; got %v", v, got)
				}
			})
		}
	})

	t.Run("inside snap", func(t *testing.T) {
		t.Setenv("SNAP_REVISION", "10")
		got, err := Cmd("X", "foo")
		if err == nil {
			t.Fatalf("expected an error, got nil")
		}
		if got != nil {
			t.Fatalf("should have returned nil, got %v", got)
		}
	})
}
