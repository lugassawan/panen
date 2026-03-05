package database

import "strings"

// buildINQuery appends a parenthesized list of n "?" placeholders to the base
// query (e.g. "SELECT ... WHERE id IN") and returns the complete SQL string.
// n must be >= 1; callers must guard against empty slices before calling.
func buildINQuery(base string, n int) string {
	var b strings.Builder
	b.WriteString(base)
	b.WriteString(" (")
	for i := range n {
		if i > 0 {
			b.WriteByte(',')
		}
		b.WriteByte('?')
	}
	b.WriteByte(')')
	return b.String()
}
