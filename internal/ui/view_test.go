// File view_test.go exercises view-specific helpers.
package ui

import "testing"

func TestClampHeight(t *testing.T) {
	t.Parallel()

	cases := map[string]struct {
		content string
		limit   int
		want    string
	}{
		"zero limit":      {content: "line", limit: 0, want: ""},
		"empty content":   {content: "", limit: 3, want: ""},
		"under the cap":   {content: "a\nb", limit: 3, want: "a\nb"},
		"trim trailing":   {content: "a\nb\nc", limit: 2, want: "a\nb"},
		"trim final":      {content: "a\nb\n", limit: 2, want: "a\nb"},
		"no newline tail": {content: "a\nb", limit: 1, want: "a"},
	}

	for name, tc := range cases {
		tc := tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := clampHeight(tc.content, tc.limit)
			if got != tc.want {
				t.Fatalf("clampHeight(%q, %d) = %q, want %q", tc.content, tc.limit, got, tc.want)
			}
		})
	}
}
