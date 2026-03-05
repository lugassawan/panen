package database

import "testing"

func TestBuildINQuery(t *testing.T) {
	tests := []struct {
		name string
		base string
		n    int
		want string
	}{
		{
			name: "single placeholder",
			base: "SELECT * FROM t WHERE id IN",
			n:    1,
			want: "SELECT * FROM t WHERE id IN (?)",
		},
		{
			name: "multiple placeholders",
			base: "SELECT * FROM t WHERE id IN",
			n:    3,
			want: "SELECT * FROM t WHERE id IN (?,?,?)",
		},
		{
			name: "five placeholders",
			base: "DELETE FROM t WHERE col IN",
			n:    5,
			want: "DELETE FROM t WHERE col IN (?,?,?,?,?)",
		},
		{
			name: "zero produces empty parens",
			base: "SELECT * FROM t WHERE id IN",
			n:    0,
			want: "SELECT * FROM t WHERE id IN ()",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := buildINQuery(tt.base, tt.n)
			if got != tt.want {
				t.Errorf("buildINQuery(%q, %d) = %q, want %q", tt.base, tt.n, got, tt.want)
			}
		})
	}
}
