package ansi

import "testing"

func TestNotify(t *testing.T) {
	tests := []struct {
		name string
		s    string
		want string
	}{
		{
			name: "basic",
			s:    "Hello, World!",
			want: "\x1b]9;Hello, World!\x07",
		},
		{
			name: "empty",
			s:    "",
			want: "\x1b]9;\x07",
		},
		{
			name: "special characters",
			s:    "Line1\nLine2\tTabbed",
			want: "\x1b]9;Line1\nLine2\tTabbed\x07",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Notify(tt.s); got != tt.want {
				t.Errorf("Notify() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestDesktopNotification(t *testing.T) {
	tests := []struct {
		name     string
		payload  string
		metadata []string
		want     string
	}{
		{
			name:     "basic",
			payload:  "Task Completed",
			metadata: []string{},
			want:     "\x1b]99;;Task Completed\x07",
		},
		{
			name:     "with metadata",
			payload:  "New Message",
			metadata: []string{"i=1", "a=focus"},
			want:     "\x1b]99;i=1:a=focus;New Message\x07",
		},
		{
			name:     "empty payload",
			payload:  "",
			metadata: []string{"i=2"},
			want:     "\x1b]99;i=2;\x07",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := DesktopNotification(tt.payload, tt.metadata...); got != tt.want {
				t.Errorf("DesktopNotification() = %q, want %q", got, tt.want)
			}
		})
	}
}
