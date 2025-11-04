package tmux

import "testing"

func TestParseOptionLine(t *testing.T) {
	t.Parallel()

	tests := []struct {
		line      string
		wantName  string
		wantValue string
		wantOK    bool
	}{
		{line: "@foo bar", wantName: "@foo", wantValue: "bar", wantOK: true},
		{line: "@foo \"bar baz\"", wantName: "@foo", wantValue: "bar baz", wantOK: true},
		{line: "    ", wantOK: false},
		{line: "# comment", wantOK: false},
	}

	for _, tt := range tests {
		name, value, ok := parseOptionLine(tt.line)
		if ok != tt.wantOK {
			t.Fatalf("parseOptionLine(%q) ok = %v, want %v", tt.line, ok, tt.wantOK)
		}
		if !ok {
			continue
		}
		if name != tt.wantName || value != tt.wantValue {
			t.Fatalf("parseOptionLine(%q) = (%q,%q), want (%q,%q)", tt.line, name, value, tt.wantName, tt.wantValue)
		}
	}
}
