package templates

import (
	"testing"
)

func TestFormatSecs(t *testing.T) {
	tests := []struct {
		input          uint32
		expectedOutput string
	}{
		{0, "0s"},
		{5, "5s"},
		{59, "59s"},
		{60, "1m"},
		{75, "1m 15s"},
		{3599, "59m 59s"},
		{3600, "1h"},
		{3723, "1h 2m 3s"},
		{7265, "2h 1m 5s"},
	}

	for _, tt := range tests {
		if got, want := formatSecs(tt.input), tt.expectedOutput; got != want {
			t.Errorf("input=%v, got=%v, want=%v", tt.input, got, want)
		}
	}
}
