package presenter

import (
	"context"
	"testing"
)

func TestNewApp(t *testing.T) {
	a := NewApp()
	if a == nil {
		t.Fatal("NewApp returned nil")
	}
}

func TestStartup(t *testing.T) {
	a := NewApp()
	ctx := context.Background()
	a.Startup(ctx)

	if a.ctx != ctx {
		t.Error("Startup did not store the context")
	}
}

func TestGreet(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{name: "simple name", input: "Alice", want: "Hello Alice, welcome to Panen!"},
		{name: "empty string", input: "", want: "Hello , welcome to Panen!"},
		{name: "name with spaces", input: "Bob Smith", want: "Hello Bob Smith, welcome to Panen!"},
		{name: "unicode name", input: "Büdi", want: "Hello Büdi, welcome to Panen!"},
	}

	a := NewApp()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := a.Greet(tt.input)
			if got != tt.want {
				t.Errorf("Greet(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}
